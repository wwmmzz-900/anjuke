package data

import (
	"anjuke/internal/conf"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/olivere/elastic/v7"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData, NewUserRepo, NewHouseRepo, NewTransactionRepo,
	NewPointsRepo, NewCustomerRepo, NewRistrettoCache,
	NewHystrixCircuit, NewOrderCacheRepo, InitRedisViaSentinel,
)

// Data .
type Data struct {
	db    *gorm.DB
	rdb   *redis.Client
	rdbRW *RedisRW
	es    *elastic.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, es *elastic.Client, rdbRW *RedisRW) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db:    db,
		rdb:   rdb,
		es:    es,
		rdbRW: rdbRW,
	}, cleanup, nil
}

func MysqlInit(c *conf.Data) (*gorm.DB, error) {
	dsn := c.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}
	log.Info("数据库连接成功")
	return db, nil
}

func ExampleClient(c *conf.Data, logger log.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis连接失败: %v", err)
	}
	log.Info("redis连接成功")
	return rdb, nil
}

func NewElasticsearch(c *conf.Data, logger log.Logger) (*elastic.Client, error) {
	helper := log.NewHelper(logger)
	client, err := elastic.NewClient(
		elastic.SetURL(c.Elasticsearch.Addr),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
	)
	if err != nil {
		helper.Fatalf("创建elasticsearch客户端失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, code, err := client.Ping(c.Elasticsearch.Addr).Do(ctx)
	if err != nil {
		helper.Fatalf("ping elasticsearch失败: %v", err)
	}
	helper.Debugf("Elasticsearch连接成功, 版本: %s, 状态码: %d", info.Version.Number, code)
	return client, err
}

// InitRedisViaSentinel 初始化Redis哨兵模式连接（最终修复版）
func InitRedisViaSentinel(logger log.Logger) (*RedisRW, error) {
	ctx := context.Background()
	helper := log.NewHelper(logger)

	// 哨兵配置参数
	sentinelAddresses := []string{
		"117.27.231.169:26379",
		"14.103.149.201:26379",
		"14.103.137.112:26379",
	}
	masterName := "mymaster"
	password := "redis_6379pwd"

	// 1. 初始化主节点连接（通过哨兵自动发现）
	master := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddresses,
		Password:      password, // 主节点密码
		//SentinelPassword: password, // 哨兵密码
		PoolSize:     2000,
		MinIdleConns: 500,
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolTimeout:  3 * time.Second,
		DialTimeout:  3 * time.Second,
		MaxRetries:   1,
	})

	// 验证主节点连接
	if err := master.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis主节点失败: %v", err)
	}
	helper.Infof("成功连接主节点: %s", master.Options().Addr)

	// 2. 寻找可用的哨兵节点
	var sentinelClient *redis.SentinelClient
	for _, addr := range sentinelAddresses {
		sc := redis.NewSentinelClient(&redis.Options{
			Addr:         addr,
			Password:     password, //
			DialTimeout:  3 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
		})

		if err := sc.Ping(ctx).Err(); err != nil {
			helper.Warnf("哨兵 %s 连接失败: %v（继续尝试下一个）", addr, err)
			sc.Close()
			continue
		}

		sentinelClient = sc
		helper.Infof("成功连接哨兵节点: %s", addr)
		break
	}

	if sentinelClient == nil {
		return nil, fmt.Errorf("所有Sentinel都无法连接，请检查：1. 26379端口是否开放 2. 密码是否正确 3. 网络是否通畅")
	}
	defer sentinelClient.Close()

	// 3. 通过哨兵获取从节点信息
	slaveInfos, err := sentinelClient.Slaves(ctx, masterName).Result()
	if err != nil {
		return nil, fmt.Errorf("获取从节点信息失败: %v", err)
	}
	helper.Infof("发现 %d 个从节点信息", len(slaveInfos))

	// 4. 连接健康的从节点（最终修复逻辑）
	var slaves []*redis.Client
	// 预设从节点IP与端口的映射关系（根据实际环境填写）
	predefinedPorts := map[string]string{
		"14.103.137.112": "6380", // 第一个从节点正确端口
		"14.103.149.201": "6666", // 第二个从节点正确端口
	}

	for _, info := range slaveInfos {
		var ip, port, flags string
		switch v := info.(type) {
		case map[string]interface{}:
			flags, _ = v["flags"].(string)
			ip, _ = v["ip"].(string)
			port, _ = v["port"].(string)
		case []interface{}:
			if len(v) >= 3 {
				flags, _ = v[0].(string)
				ip, _ = v[1].(string)
				port, _ = v[2].(string)
			} else {
				helper.Warnf("从节点信息格式异常（数组长度不足）: %v", v)
				continue
			}
		default:
			helper.Warnf("不支持的从节点信息格式: %T", v)
			continue
		}

		// 核心修复1：强制清理IP中的所有无效字符
		ip = strings.TrimSuffix(ip, ":ip") // 移除":ip"后缀
		ip = strings.Split(ip, ":")[0]     // 按冒号分割，取第一个部分（如"14.103.137.112:80"→"14.103.137.112"）
		ip = strings.TrimSpace(ip)         // 去除空格

		// 核心修复2：使用预设端口覆盖错误端口
		if predefinedPort, ok := predefinedPorts[ip]; ok {
			port = predefinedPort // 强制使用正确端口
			helper.Debugf("从节点 %s 使用预设端口: %s", ip, port)
		} else {
			// 非预设IP的处理（清理端口）
			port = strings.TrimSuffix(port, ":ip")
			port = strings.TrimSpace(port)
		}

		// 跳过空IP或端口的节点
		if ip == "" || port == "" {
			helper.Warnf("从节点IP或端口为空，跳过")
			continue
		}

		// 验证端口格式
		if _, err := strconv.Atoi(port); err != nil {
			helper.Warnf("从节点端口格式无效: %s，跳过", port)
			continue
		}

		addr := fmt.Sprintf("%s:%s", ip, port)
		helper.Debugf("最终从节点地址: %s", addr)

		// 过滤不健康的从节点
		if strings.Contains(flags, "s_down") || strings.Contains(flags, "disconnected") {
			helper.Warnf("从节点 %s 状态异常（%s），跳过连接", addr, flags)
			continue
		}

		// 连接从节点
		slave := redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			PoolSize:     1000,
			MinIdleConns: 200,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			PoolTimeout:  3 * time.Second,
			DialTimeout:  3 * time.Second,
			MaxRetries:   1,
		})

		if err := slave.Ping(ctx).Err(); err != nil {
			slave.Close()
			helper.Warnf("连接从节点 %s 失败: %v", addr, err)
			continue
		}

		slaves = append(slaves, slave)
		helper.Infof("成功连接从节点: %s", addr)
	}

	// 无可用从节点时，使用主节点作为读节点
	if len(slaves) == 0 {
		helper.Warn("未找到可用的从节点，将使用主节点进行读操作")
		slaves = append(slaves, master)
	}

	// 5. 预热连接池
	for i, s := range append(slaves, master) {
		if err := s.Ping(ctx).Err(); err != nil {
			helper.Warnf("连接池预热失败（%s）: %v", s.Options().Addr, err)
		} else if i == len(slaves) {
			helper.Infof("主节点连接池预热完成: %s", s.Options().Addr)
		} else {
			helper.Infof("从节点 %d 连接池预热完成: %s", i+1, s.Options().Addr)
		}
	}

	// 6. 从节点轮询算法
	var idx uint32
	getSlave := func() *redis.Client {
		if len(slaves) == 1 {
			return slaves[0]
		}
		i := atomic.AddUint32(&idx, 1)
		return slaves[i%uint32(len(slaves))]
	}

	helper.Infof("Redis哨兵模式初始化完成, 主节点: %s, 可用从节点: %d", master.Options().Addr, len(slaves))
	return &RedisRW{
		Master:   master,
		Slaves:   slaves,
		getSlave: getSlave,
	}, nil
}

// RedisRW 读写分离客户端结构体（保持不变）
type RedisRW struct {
	Master   *redis.Client
	Slaves   []*redis.Client
	getSlave func() *redis.Client
}

// 以下Redis操作方法保持不变...
func (rw *RedisRW) Get(ctx context.Context, key string) *redis.StringCmd {
	return rw.getSlave().Get(ctx, key)
}

func (rw *RedisRW) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return rw.Master.Set(ctx, key, value, expiration)
}

func (rw *RedisRW) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return rw.Master.Del(ctx, keys...)
}

func (rw *RedisRW) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return rw.getSlave().HGet(ctx, key, field)
}

func (rw *RedisRW) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return rw.Master.HSet(ctx, key, values...)
}

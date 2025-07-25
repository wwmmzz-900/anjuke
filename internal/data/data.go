package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/wwmmzz-900/anjuke/internal/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo, NewHouseRepo, NewTransactionRepo, NewPointsRepo, NewCustomerRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	db  *gorm.DB
	rdb *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db:  db,
		rdb: rdb,
	}, cleanup, nil
}
func MysqlInit(c *conf.Data, logger log.Logger) (*gorm.DB, error) {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	datas := &Data{}
	dsn := c.Database.Source
	var err error
	
	// 优化GORM配置以支持高并发
	config := &gorm.Config{
		// 禁用默认事务，提高性能
		SkipDefaultTransaction: true,
		// 预编译语句缓存
		PrepareStmt: true,
		// 批量插入大小
		CreateBatchSize: 1000,
	}
	
	datas.db, err = gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 获取底层sql.DB以配置连接池
	sqlDB, err := datas.db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 配置连接池参数以支持高并发
	sqlDB.SetMaxOpenConns(100)        // 最大打开连接数
	sqlDB.SetMaxIdleConns(20)         // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 连接最大空闲时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	fmt.Println("数据库连接成功，连接池已配置")
	return datas.db, nil
}
func ExampleClient(c *conf.Data, logger log.Logger) (*redis.Client, error) {
	data := &Data{}
	
	// 优化Redis配置以支持高并发
	data.rdb = redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           0,
		// 连接池配置
		PoolSize:     50,                    // 连接池大小
		MinIdleConns: 10,                    // 最小空闲连接数
		MaxRetries:   3,                     // 最大重试次数
		DialTimeout:  5 * time.Second,       // 连接超时
		ReadTimeout:  3 * time.Second,       // 读取超时
		WriteTimeout: 3 * time.Second,       // 写入超时
		PoolTimeout:  4 * time.Second,       // 连接池超时
		IdleTimeout:  5 * time.Minute,       // 空闲连接超时
		// 启用连接检查
		IdleCheckFrequency: 1 * time.Minute,
	})
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := data.rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis连接失败: %w", err)
	}
	
	fmt.Println("redis连接成功，连接池已配置")
	return data.rdb, nil
}

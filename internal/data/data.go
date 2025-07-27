package data

import (
	"anjuke/internal/conf"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/smartwalle/alipay/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo, NewHouseRepo, NewTransactionRepo, NewPointsRepo, NewCustomerRepo, NewAlipayClient)

// Data .
type Data struct {
	// TODO wrapped database client
	db     *gorm.DB
	rdb    *redis.Client
	alipay *alipay.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, alipay *alipay.Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db:     db,
		rdb:    rdb,
		alipay: alipay,
	}, cleanup, nil
}
func MysqlInit(c *conf.Data, logger log.Logger) (*gorm.DB, error) {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	datas := &Data{}
	dsn := c.Database.Source
	var err error
	datas.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败")
	}

	fmt.Println("数据库连接成功")
	return datas.db, nil
}
func ExampleClient(c *conf.Data, logger log.Logger) (*redis.Client, error) {
	data := &Data{}
	data.rdb = redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password, // no password set
		DB:       0,                // use default DB
	})
	_, err := data.rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("redis连接失败")
	}
	fmt.Println("redis连接成功")
	return data.rdb, nil
}

// NewAlipayClient 初始化支付宝客户端
func NewAlipayClient(c *conf.Data, logger log.Logger) (*alipay.Client, error) {
	// 创建支付宝客户端实例，传入应用ID、私钥和是否为生产环境
	client, err := alipay.New(c.Alipay.AppId, c.Alipay.PrivateKey, c.Alipay.IsProduction)
	if err != nil {
		return nil, fmt.Errorf("创建支付宝客户端失败: %v", err)
	}

	// 从文件加载应用公钥证书
	err = client.LoadAppPublicCertFromFile("C:\\Users\\29259\\Desktop\\anjuke\\internal\\alipay\\appPublicCert.crt")
	if err != nil {
		return nil, fmt.Errorf("加载应用公钥证书失败: %v", err)
	}

	// 从文件加载支付宝根证书
	err = client.LoadAliPayRootCertFromFile("C:\\Users\\29259\\Desktop\\anjuke\\internal\\alipay\\alipayRootCert.crt")
	if err != nil {
		return nil, fmt.Errorf("加载支付宝根证书失败: %v", err)
	}

	// 从文件加载支付宝公钥证书
	err = client.LoadAliPayPublicCertFromFile("C:\\Users\\29259\\Desktop\\anjuke\\internal\\alipay\\alipayPublicCert.crt")
	if err != nil {
		return nil, fmt.Errorf("加载支付宝公钥证书失败: %v", err)
	}

	// 打印连接成功信息
	log.NewHelper(logger).Info("alipay连接成功")
	return client, nil
}

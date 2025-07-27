package data

import (
	"anjuke/server/internal/conf"
	"anjuke/server/internal/domain"
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProvideShumaiAppCode 用于 wire 注入数脉短信 AppCode
func ProvideShumaiAppCode(conf *conf.Data) string {
	return conf.ShumaiSms.AppCode
}

// NewShumaiSmsSenderWithDeps 创建带有依赖的短信发送器
func NewShumaiSmsSenderWithDeps(appCode string, rdb *redis.Client) *ShumaiSmsSender {
	return NewShumaiSmsSender(appCode, rdb)
}

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	MysqlInit,
	ExampleClient,
	NewMinioClient,
	NewRealNameSDK,
	ProvideShumaiAppCode,
	NewShumaiSmsSenderWithDeps,
	NewUserRepo,
	NewGreeterRepo,
	NewPointsRepo,
	NewSmsRepo,
	wire.Bind(new(domain.MinioRepo), new(*MinioClient)),
)

// Data .
type Data struct {
	// TODO wrapped database client
	db  *gorm.DB
	rdb *redis.Client
	Mc  *MinioClient
}

// NewData .
func NewData(confData *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, mc *MinioClient) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db:  db,
		rdb: rdb,
		Mc:  mc,
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

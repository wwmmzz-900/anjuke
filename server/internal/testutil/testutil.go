package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"anjuke/server/internal/conf"
	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MinioConfig 测试用MinIO配置
type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSsl    bool
	Bucket    string
}

// TestConfig 测试配置
type TestConfig struct {
	Database *conf.Data_Database
	Redis    *conf.Data_Redis
	Minio    *MinioConfig
}

// GetTestConfig 获取测试配置
func GetTestConfig() *TestConfig {
	return &TestConfig{
		Database: &conf.Data_Database{
			Driver: "mysql",
			Source: "root:test@tcp(localhost:3306)/anjuke_test?parseTime=True&loc=Local",
		},
		Redis: &conf.Data_Redis{
			Addr:         "localhost:6379",
			Password:     "",
			ReadTimeout:  &durationpb.Duration{Seconds: 0, Nanos: 200000000},
			WriteTimeout: &durationpb.Duration{Seconds: 0, Nanos: 200000000},
		},
		Minio: &MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			UseSsl:    false,
			Bucket:    "test-bucket",
		},
	}
}

// SetupTestDB 设置测试数据库
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	config := GetTestConfig()

	db, err := gorm.Open(mysql.Open(config.Database.Source), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("连接测试数据库失败，跳过测试: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&domain.UserBase{},
		&domain.RealName{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	// 返回清理函数
	cleanup := func() {
		// 清理测试数据
		db.Exec("DELETE FROM userbase")
		db.Exec("DELETE FROM realname")

		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return db, cleanup
}

// SetupTestRedis 设置测试Redis
func SetupTestRedis(t *testing.T) (*redis.Client, func()) {
	config := GetTestConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis连接失败，跳过测试: %v", err)
	}

	cleanup := func() {
		// 清理测试数据
		rdb.FlushDB(context.Background())
		rdb.Close()
	}

	return rdb, cleanup
}

// MockLogger 创建测试用的日志记录器
func MockLogger() log.Logger {
	return log.NewStdLogger(os.Stdout)
}

// CreateTestUser 创建测试用户
func CreateTestUser(db *gorm.DB) *domain.UserBase {
	user := &domain.UserBase{
		Phone:      "13800138000",
		Name:       "测试用户",
		Password:   "5d41402abc4b2a76b9719d911017c592", // MD5("hello")
		RoleID:     1,
		RealStatus: domain.RealNameUnverified,
		Status:     domain.UserStatusNormal,
	}

	db.Create(user)
	return user
}

// AssertError 断言错误
func AssertError(t *testing.T, err error, expectError bool, msg string) {
	if expectError && err == nil {
		t.Errorf("%s: 期望有错误但没有错误", msg)
	}
	if !expectError && err != nil {
		t.Errorf("%s: 不期望有错误但有错误: %v", msg, err)
	}
}

// AssertEqual 断言相等
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	if expected != actual {
		t.Errorf("%s: 期望 %v, 实际 %v", msg, expected, actual)
	}
}

// AssertNotNil 断言不为空
func AssertNotNil(t *testing.T, value interface{}, msg string) {
	if value == nil {
		t.Errorf("%s: 期望不为空但为空", msg)
	}
}

package data

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DataDB 数据库连接配置
type DataDB struct {
	db  *gorm.DB
	log *log.Helper
}

// DBConfig 数据库配置结构
type DBConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Database     string `json:"database"`
	Charset      string `json:"charset"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxLifetime  int    `json:"max_lifetime"` // 秒
	AutoMigrate  bool   `json:"auto_migrate"` // 是否启用自动迁移，默认false
}

// NewDBData 创建数据库连接
// 功能说明：
//   - 建立MySQL数据库连接
//   - 配置连接池参数
//   - 设置GORM日志级别
//   - 自动迁移数据库表结构
func NewDBData(config *DBConfig, logger log.Logger) (*DataDB, func(), error) {
	logHelper := log.NewHelper(logger)

	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
	)

	// 配置GORM
	gormConfig := &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// 建立数据库连接
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		logHelper.Errorf("连接MySQL数据库失败: %v", err)
		return nil, nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 获取底层sql.DB对象配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		logHelper.Errorf("获取数据库连接池失败: %v", err)
		return nil, nil, fmt.Errorf("获取连接池失败: %v", err)
	}

	// 配置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)                                // 最大空闲连接数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)                                // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second) // 连接最大生存时间

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		logHelper.Errorf("数据库连接测试失败: %v", err)
		return nil, nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	logHelper.Info("MySQL数据库连接成功")

	// 根据配置决定是否执行自动迁移
	if config.AutoMigrate {
		logHelper.Info("启用自动迁移模式")
		if err := autoMigrate(db, logHelper); err != nil {
			logHelper.Errorf("数据库表迁移失败: %v", err)
			return nil, nil, fmt.Errorf("数据库迁移失败: %v", err)
		}
	} else {
		logHelper.Info("自动迁移已禁用，跳过表结构迁移")
	}

	data := &DataDB{
		db:  db,
		log: logHelper,
	}

	// 返回清理函数
	cleanup := func() {
		logHelper.Info("关闭数据库连接")
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return data, cleanup, nil
}

// autoMigrate 自动迁移数据库表结构
// 功能说明：
//   - 自动创建或更新数据库表结构
//   - 确保表结构与模型定义一致
//   - 不会删除已存在的数据
//   - 添加表和字段注释到MySQL数据库
func autoMigrate(db *gorm.DB, logger *log.Helper) error {
	logger.Info("开始数据库表结构迁移")

	// 定义需要迁移的模型和表注释
	type modelWithComment struct {
		model   interface{}
		comment string
	}

	models := []modelWithComment{
		// 预约相关表
		{&AppointmentModel{}, "预约表"},
		{&AppointmentLogModel{}, "预约日志表"},
		{&StoreWorkingHoursModel{}, "门店工作时间表"},
		{&RealtorWorkingHoursModel{}, "经纪人工作时间表"},
		{&RealtorStatusModel{}, "经纪人状态表"},
		{&AppointmentReviewModel{}, "预约评价表"},

		// 公司相关表
		{&CompanyModel{}, "公司表"},
		{&StoreModel{}, "门店表"},
		{&RealtorModel{}, "经纪人表"},
	}

	// 执行自动迁移
	for _, m := range models {
		// 先执行自动迁移创建或更新表结构
		if err := db.AutoMigrate(m.model); err != nil {
			logger.Errorf("迁移表 %T 失败: %v", m.model, err)
			return fmt.Errorf("迁移表失败: %v", err)
		}

		// 获取表名
		tableName := ""
		if tabler, ok := m.model.(interface{ TableName() string }); ok {
			tableName = tabler.TableName()
		} else {
			// 如果模型没有实现TableName方法，则使用GORM默认的表名推断
			stmt := &gorm.Statement{DB: db}
			stmt.Parse(m.model)
			tableName = stmt.Schema.Table
		}

		// 添加表注释
		if m.comment != "" {
			sql := fmt.Sprintf("ALTER TABLE `%s` COMMENT '%s'", tableName, m.comment)
			if err := db.Exec(sql).Error; err != nil {
				logger.Warnf("为表 %s 添加注释失败: %v", tableName, err)
			}
		}

		logger.Infof("成功迁移表: %s (%s)", tableName, m.comment)
	}

	logger.Info("数据库表结构迁移完成")
	return nil
}

// GetDB 获取GORM数据库实例
func (d *DataDB) GetDB() *gorm.DB {
	return d.db
}

// Data 数据结构体，包含所有数据存储相关的依赖
type Data struct {
	db  *gorm.DB
	log *log.Helper
	// 可以在这里添加其他数据存储依赖，如Redis、MinIO等
}

// NewData 创建数据存储实例
func NewData(dbData *DataDB, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)

	data := &Data{
		db:  dbData.GetDB(),
		log: logHelper,
	}

	cleanup := func() {
		logHelper.Info("关闭数据存储连接")
	}

	return data, cleanup, nil
}

// GetDB 获取GORM数据库实例
func (d *Data) GetDB() *gorm.DB {
	return d.db
}

// ProviderSet 依赖注入提供者集合
var ProviderSet = wire.NewSet(
	NewDBData,
	NewData,
)

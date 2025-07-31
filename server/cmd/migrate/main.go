package main

import (
	"flag"
	"fmt"
	"log"

	"anjuke/server/internal/data"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	var (
		host     = flag.String("host", "localhost", "数据库主机地址")
		port     = flag.Int("port", 3306, "数据库端口")
		username = flag.String("username", "root", "数据库用户名")
		password = flag.String("password", "", "数据库密码")
		database = flag.String("database", "anjuke", "数据库名称")
		charset  = flag.String("charset", "utf8mb4", "字符集")
		help     = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Parse()

	if *help {
		fmt.Println("数据库迁移工具")
		fmt.Println("用法:")
		fmt.Println("  go run cmd/migrate/main.go [选项]")
		fmt.Println("")
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  go run cmd/migrate/main.go -host=localhost -port=3306 -username=root -password=123456 -database=anjuke")
		return
	}

	if *password == "" {
		fmt.Print("请输入数据库密码: ")
		fmt.Scanln(password)
	}

	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		*username,
		*password,
		*host,
		*port,
		*database,
		*charset,
	)

	// 建立数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 获取底层sql.DB对象
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}
	defer sqlDB.Close()

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Printf("成功连接到数据库: %s@%s:%d/%s\n", *username, *host, *port, *database)

	// 执行数据库迁移
	fmt.Println("开始执行数据库迁移...")
	if err := executeMigration(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("数据库迁移完成！")
}

// executeMigration 执行数据库迁移
func executeMigration(db *gorm.DB) error {
	fmt.Println("开始数据库表结构迁移")

	// 定义需要迁移的模型和表注释
	type modelWithComment struct {
		model   interface{}
		comment string
	}

	models := []modelWithComment{
		// 预约相关表
		{&data.AppointmentModel{}, "预约表"},
		{&data.AppointmentLogModel{}, "预约日志表"},
		{&data.StoreWorkingHoursModel{}, "门店工作时间表"},
		{&data.RealtorWorkingHoursModel{}, "经纪人工作时间表"},
		{&data.RealtorStatusModel{}, "经纪人状态表"},
		{&data.AppointmentReviewModel{}, "预约评价表"},

		// 公司相关表
		{&data.CompanyModel{}, "公司表"},
		{&data.StoreModel{}, "门店表"},
		{&data.RealtorModel{}, "经纪人表"},
	}

	// 执行自动迁移
	for _, m := range models {
		// 先执行自动迁移创建或更新表结构
		if err := db.AutoMigrate(m.model); err != nil {
			fmt.Printf("❌ 迁移表 %T 失败: %v\n", m.model, err)
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
				fmt.Printf("⚠️  为表 %s 添加注释失败: %v\n", tableName, err)
			}
		}

		fmt.Printf("✓ 成功迁移表: %s (%s)\n", tableName, m.comment)
	}

	fmt.Println("数据库表结构迁移完成")
	return nil
}

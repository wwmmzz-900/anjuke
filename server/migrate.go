package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接配置（从config.yaml中获取）
	dsn := "root:e10adc3949ba59abbe56e057f20f883e@tcp(14.103.149.201:3306)/anjuke?parseTime=True&loc=Local"

	fmt.Println("🔧 开始执行数据库迁移...")

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ 数据库连接测试失败: %v", err)
	}
	fmt.Println("✅ 数据库连接成功")

	// 获取所有迁移文件
	migrationFiles, err := getMigrationFiles("migrations")
	if err != nil {
		log.Fatalf("❌ 获取迁移文件失败: %v", err)
	}

	if len(migrationFiles) == 0 {
		fmt.Println("📄 没有找到迁移文件")
		return
	}

	fmt.Printf("📄 找到 %d 个迁移文件\n", len(migrationFiles))

	// 执行所有迁移文件
	for _, migrationFile := range migrationFiles {
		fmt.Printf("\n🔄 执行迁移文件: %s\n", migrationFile)

		// 读取迁移文件
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			log.Printf("❌ 读取迁移文件失败: %v", err)
			continue
		}

		// 执行SQL语句
		sqlContent := string(content)

		// 分割SQL语句（简单的分割，按分号分割）
		statements := splitSQL(sqlContent)

		for i, stmt := range statements {
			stmt = trimStatement(stmt)
			if stmt == "" {
				continue
			}

			fmt.Printf("  🔄 执行语句 %d/%d...\n", i+1, len(statements))

			_, err := db.Exec(stmt)
			if err != nil {
				fmt.Printf("  ⚠️  语句执行警告: %v\n", err)
				fmt.Printf("     语句: %s\n", stmt[:min(100, len(stmt))])
			} else {
				fmt.Printf("  ✅ 语句执行成功\n")
			}
		}

		fmt.Printf("✅ 迁移文件 %s 执行完成\n", migrationFile)
	}

	fmt.Println("🎉 数据库迁移完成！")

	// 验证表是否创建成功
	fmt.Println("\n🔍 验证表结构...")
	tables := []string{"user_points", "points_records", "checkin_records"}

	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			fmt.Printf("❌ 表 %s 验证失败: %v\n", table, err)
		} else {
			fmt.Printf("✅ 表 %s 验证成功，记录数: %d\n", table, count)
		}
	}
}

// 简单的SQL语句分割函数
func splitSQL(content string) []string {
	var statements []string
	var current string

	lines := []string{}
	for _, line := range []string{content} {
		for _, l := range []rune(line) {
			if l == '\n' {
				lines = append(lines, current)
				current = ""
			} else {
				current += string(l)
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	current = ""
	for _, line := range lines {
		line = trimStatement(line)
		if line == "" || line[0] == '-' { // 跳过注释
			continue
		}

		current += line + "\n"

		if len(line) > 0 && line[len(line)-1] == ';' {
			statements = append(statements, current)
			current = ""
		}
	}

	if current != "" {
		statements = append(statements, current)
	}

	return statements
}

// 清理SQL语句
func trimStatement(stmt string) string {
	// 简单的清理：去除前后空格和换行
	result := ""
	for _, char := range stmt {
		if char == '\n' || char == '\r' || char == '\t' {
			result += " "
		} else {
			result += string(char)
		}
	}

	// 去除前后空格
	start := 0
	end := len(result)

	for start < end && result[start] == ' ' {
		start++
	}

	for end > start && result[end-1] == ' ' {
		end--
	}

	if start >= end {
		return ""
	}

	return result[start:end]
}

// getMigrationFiles 获取所有迁移文件并按名称排序
func getMigrationFiles(dir string) ([]string, error) {
	var migrationFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理 .sql 文件
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sql") {
			migrationFiles = append(migrationFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 按文件名排序，确保按顺序执行
	sort.Strings(migrationFiles)

	return migrationFiles, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

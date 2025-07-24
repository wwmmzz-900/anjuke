package data

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/wwmmzz-900/anjuke/internal/conf"
)

// 测试数据库连接和表结构
func TestDatabaseConnection(t *testing.T) {
	// 创建配置
	c := &conf.Data{
		Database: &conf.Data_Database{
			Driver: "mysql",
			Source: "root:e10adc3949ba59abbe56e057f20f883e@tcp(14.103.149.201:3306)/anjuke?parseTime=True&loc=Local",
		},
	}

	// 初始化数据库连接
	db, err := MysqlInit(c, log.DefaultLogger)
	if err != nil {
		t.Fatalf("数据库连接失败: %v", err)
	}

	// 创建数据访问层
	data := &Data{db: db}
	repo := &houseRepo{data: data}

	// 测试查询houses表
	var count int64
	err = repo.data.db.Table("house").Count(&count).Error
	if err != nil {
		t.Fatalf("查询houses表失败: %v", err)
	}

	fmt.Printf("houses表中共有 %d 条记录\n", count)

	// 检查表结构
	var houses []struct {
		HouseID     int64   `gorm:"column:house_id"`
		Title       string  `gorm:"column:title"`
		Description string  `gorm:"column:description"`
		Price       float64 `gorm:"column:price"`
		Area        float64 `gorm:"column:area"`
		Layout      string  `gorm:"column:layout"`
		ImageURL    string  `gorm:"column:image_url"`
		Status      string  `gorm:"column:status"`
	}

	err = repo.data.db.Table("houses").Limit(1).Scan(&houses).Error
	if err != nil {
		t.Fatalf("检查houses表结构失败: %v", err)
	}

	if len(houses) > 0 {
		fmt.Printf("houses表结构正常，示例记录: %+v\n", houses[0])
	} else {
		fmt.Printf("houses表中没有记录\n")
	}

	// 测试推荐接口
	houseList, total, err := repo.GetRecommendList(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("获取推荐列表失败: %v", err)
	}

	fmt.Printf("成功获取到 %d 条推荐，总数: %d\n", len(houseList), total)
	for i, house := range houseList {
		fmt.Printf("房源 %d: ID=%d, 标题=%s, 价格=%.2f\n", i+1, house.HouseID, house.Title, house.Price)
	}
}

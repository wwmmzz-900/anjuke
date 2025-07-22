package data

import (
	"context"
	"github.com/wwmmzz-900/anjuke/internal/model"

	"github.com/wwmmzz-900/anjuke/internal/biz"
)

type houseRepo struct {
	data *Data // Data 结构体包含 *sql.DB
}

func NewHouseRepo(data *Data) biz.HouseRepo {
	return &houseRepo{data: data}
}

// 查询用户最近浏览的房源价格区间
func (r *houseRepo) GetUserPricePreference(ctx context.Context, userID int64) (float64, float64, error) {
	type result struct {
		MinPrice float64
		MaxPrice float64
	}

	var res result
	
	// 使用user_behavior表查询用户浏览过的房源价格区间
	err := r.data.db.
		Table("user_behavior AS ub").
		Select("MIN(h.price) AS min_price, MAX(h.price) AS max_price").
		Joins("JOIN houses h ON ub.house_id = h.house_id"). // 使用house_id字段
		Where("ub.user_id = ? AND ub.behavior = 'view'", userID).
		Order("ub.created_at DESC").
		Limit(20).
		Scan(&res).Error
	
	if err != nil {
		// 如果查询失败，返回默认价格区间
		return 800, 5000, nil
	}
	
	// 如果没有浏览记录或价格为0，返回默认区间
	if res.MinPrice == 0 && res.MaxPrice == 0 {
		return 800, 5000, nil
	}
	
	return res.MinPrice, res.MaxPrice, nil
}

// 查询个性化推荐房源
func (r *houseRepo) GetPersonalRecommendList(ctx context.Context, minPrice, maxPrice float64, page, pageSize int) ([]*biz.House, int, error) {
	var houses []*biz.House
	var total int64
	
	// 尝试从数据库查询
	err := r.data.db.
		Table("houses").
		Where("price BETWEEN ? AND ?", minPrice, maxPrice).
		Where("status = 'available'"). // 假设有status字段表示房源状态
		Count(&total).Error
	
	if err != nil {
		// 如果查询失败，返回模拟数据
		return getDefaultHouses(), 3, nil
	}
	
	// 查询房源列表
	err = r.data.db.
		Table("houses").
		Where("price BETWEEN ? AND ?", minPrice, maxPrice).
		Where("status = 'available'").
		Order("house_id DESC"). // 使用house_id而不是id
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&houses).Error
	
	if err != nil || len(houses) == 0 {
		// 如果查询失败或没有数据，返回模拟数据
		return getDefaultHouses(), 3, nil
	}
	
	return houses, int(total), nil
}

// 获取默认房源数据
func getDefaultHouses() []*biz.House {
	return []*biz.House{
		{
			HouseID:     101,
			Title:       "精装修两室一厅",
			Description: "地铁口附近，交通便利，精装修",
			Price:       3500.0,
			Area:        85.5,
			Layout:      "2室1厅1卫",
			ImageURL:    "https://example.com/house1.jpg",
		},
		{
			HouseID:     102,
			Title:       "温馨三室两厅",
			Description: "小区环境优美，配套设施完善",
			Price:       4200.0,
			Area:        120.0,
			Layout:      "3室2厅2卫",
			ImageURL:    "https://example.com/house2.jpg",
		},
		{
			HouseID:     103,
			Title:       "豪华公寓",
			Description: "高端小区，装修豪华，设施齐全",
			Price:       5800.0,
			Area:        150.0,
			Layout:      "3室2厅2卫",
			ImageURL:    "https://example.com/house3.jpg",
		},
	}
}

// 预约记录模型

// 保存预约
func (r *houseRepo) CreateReservation(ctx context.Context, reservation *model.HouseReservation) error {
	return r.data.db.WithContext(ctx).Create(reservation).Error
}

// 查询是否已预约
func (r *houseRepo) HasReservation(ctx context.Context, userID, houseID int64) (bool, error) {
	var count int64
	err := r.data.db.WithContext(ctx).
		Model(&model.HouseReservation{}).
		Where("user_id = ? AND house_id = ?", userID, houseID).
		Count(&count).Error
	return count > 0, err
}

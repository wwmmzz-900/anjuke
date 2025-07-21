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
	err := r.data.db.
		Table("user_visit_history AS uvh").
		Select("MIN(h.price) AS min_price, MAX(h.price) AS max_price").
		Joins("JOIN house h ON uvh.house_id = h.id").
		Where("uvh.user_id = ?", userID).
		Order("uvh.visit_time DESC").
		Limit(20).
		Scan(&res).Error
	if err != nil {
		return 0, 0, err
	}
	return res.MinPrice, res.MaxPrice, nil
}

// 查询个性化推荐房源
func (r *houseRepo) GetPersonalRecommendList(ctx context.Context, minPrice, maxPrice float64, page, pageSize int) ([]*biz.House, int, error) {
	var houses []*biz.House
	var total int64 // 修改这里

	// 统计总数
	err := r.data.db.
		Model(&biz.House{}).
		Where("price BETWEEN ? AND ?", minPrice, maxPrice).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询房源

	err = r.data.db.
		Table("house").
		Where("price BETWEEN ? AND ?", minPrice, maxPrice).
		Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&houses).Error
	if err != nil {
		return nil, 0, err
	}
	return houses, int(total), nil // 返回时转为int类型
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

package data

import (
	"anjuke/internal/biz"
	"anjuke/internal/model"
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type OrderRepo struct {
	db  *gorm.DB
	log *log.Helper
}

// NewOrderRepo .
func NewOrderRepo(db *gorm.DB, logger log.Logger) biz.OrderRepo {
	return &OrderRepo{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// 从db获取
func (o *OrderRepo) GetRentalOrderByDB(ctx context.Context, tenantId uint, page, pageSize int) (*[]model.RentalOrder, int64, error) {
	var order []model.RentalOrder
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	query := o.db.WithContext(ctx).Model(&model.RentalOrder{}).Where("tenant_id = ?", tenantId)

	var count int64
	query.Count(&count)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&order).Error; err != nil {
		return nil, 0, err
	}
	return &order, count, nil
}

// GetOrderDetail 从db获取订单详情
func (o *OrderRepo) GetOrderDetail(ctx context.Context, id uint) (*model.RentalOrder, error) {
	var order model.RentalOrder
	if err := o.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (o *OrderRepo) GetRentalOrderByCache(ctx context.Context, tenantId uint, page, pageSize int) (*[]model.RentalOrder, int64, bool, error) {
	//TODO implement me
	panic("implement me")
}

func (o *OrderRepo) SetRentalOrderToCache(ctx context.Context, tenantId uint, page, pageSize int, orders *[]model.RentalOrder, total int64) error {
	//TODO implement me
	panic("implement me")
}

func (o *OrderRepo) IsCircuitBreakerOpen() bool {
	//TODO implement me
	panic("implement me")
}

func (o *OrderRepo) GetDegradedOrderList(ctx context.Context, tenantId uint) (*[]model.RentalOrder, int64, error) {
	//TODO implement me
	panic("implement me")
}

// CreateOrder 创建订单
func (o *OrderRepo) CreateOrder(ctx context.Context, order *model.RentalOrder) error {
	if err := o.db.WithContext(ctx).Create(order).Error; err != nil {
		o.log.Errorf("创建订单失败: %v", err)
		return err
	}
	return nil
}

// CheckHouseAvailable 检查房源在指定时间段内是否可用
func (o *OrderRepo) CheckHouseAvailable(ctx context.Context, houseId uint, rentStart, rentEnd time.Time) (bool, error) {
	var count int64

	// 查询是否存在时间冲突的订单
	err := o.db.WithContext(ctx).Model(&model.RentalOrder{}).
		Where("house_id = ? AND status IN (?, ?)", houseId, model.OrderStatusPending, model.OrderStatusActive).
		Where("rent_start < ? AND rent_end > ?", rentEnd, rentStart).
		Count(&count).Error

	if err != nil {
		o.log.Errorf("检查房源可用性失败: %v", err)
		return false, err
	}

	return count == 0, nil
}

// CancelOrder 取消订单
func (o *OrderRepo) CancelOrder(ctx context.Context, id uint, cancelReason string) error {
	now := time.Now()

	// 更新订单状态为已取消，并记录取消原因和时间
	result := o.db.WithContext(ctx).Model(&model.RentalOrder{}).
		Where("id = ? AND status IN (?, ?)", id, model.OrderStatusPending, model.OrderStatusActive).
		Updates(map[string]interface{}{
			"status":        model.OrderStatusCancelled,
			"cancel_reason": cancelReason,
			"cancelled_at":  &now,
		})

	if result.Error != nil {
		o.log.Errorf("取消订单失败: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("订单不存在或状态不允许取消")
	}

	return nil
}

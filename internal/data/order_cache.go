package data

import (
	"anjuke/internal/biz"
	"anjuke/internal/model"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// 缓存键生成器
func generateUserOrderCacheKey(tenantId uint, page, pageSize int) string {
	rawKey := fmt.Sprintf("user_orders_%d_%d_%d", tenantId, page, pageSize)
	hash := md5.Sum([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// OrderCacheRepo 实现带缓存的订单仓储
type OrderCacheRepo struct {
	RedisRW    *RedisRW       // 包含Redis客户端
	db         *gorm.DB       // 包含DB客户端
	localCache LocalCache     // 本地缓存(可使用ristretto实现)
	circuit    CircuitBreaker // 熔断器
	dbRepo     biz.OrderRepo  // 订单仓储
}

func (o *OrderCacheRepo) CreateOrder(ctx context.Context, order *model.RentalOrder) error {
	// 直接调用底层数据库仓储创建订单
	return o.dbRepo.CreateOrder(ctx, order)
}

func (o *OrderCacheRepo) CheckHouseAvailable(ctx context.Context, houseId uint, rentStart, rentEnd time.Time) (bool, error) {
	// 直接调用底层数据库仓储检查房源可用性
	return o.dbRepo.CheckHouseAvailable(ctx, houseId, rentStart, rentEnd)
}

func (o *OrderCacheRepo) CancelOrder(ctx context.Context, id uint, cancelReason string) error {
	// 1. 先执行取消操作
	err := o.dbRepo.CancelOrder(ctx, id, cancelReason)
	if err != nil {
		return err
	}

	// 2. 清除相关缓存
	go func() {
		asyncCtx := context.Background()

		// 清除订单详情缓存
		detailCacheKey := fmt.Sprintf("order_detail_%d", id)
		_ = o.RedisRW.Del(asyncCtx, detailCacheKey).Err()
		o.localCache.Set(detailCacheKey, nil, time.Nanosecond) // 立即过期

		// 注意：这里可能需要清除相关的订单列表缓存
		// 但由于不知道具体的租客ID，暂时不清除列表缓存
		// 实际项目中可以考虑使用缓存标签或者其他策略
	}()

	return nil
}

func (o *OrderCacheRepo) GetOrderDetail(ctx context.Context, id uint) (*model.RentalOrder, error) {
	cacheKey := fmt.Sprintf("order_detail_%d", id)

	// 1. 查本地缓存
	if val, ok := o.localCache.Get(cacheKey); ok {
		if order, ok := val.(*model.RentalOrder); ok {
			return order, nil
		}
	}

	// 2. 查 Redis
	if val, err := o.RedisRW.Get(ctx, cacheKey).Result(); err == nil && val != "" {
		var order model.RentalOrder
		if json.Unmarshal([]byte(val), &order) == nil {
			o.localCache.Set(cacheKey, &order, time.Minute*5) // 回写本地缓存
			return &order, nil
		}
	}

	// 3. 缓存未命中 → 查数据库
	order, err := o.dbRepo.GetOrderDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	// 4. 写回缓存
	if b, err := json.Marshal(order); err == nil {
		_ = o.RedisRW.Set(ctx, cacheKey, b, time.Hour).Err() // 写 Redis
		o.localCache.Set(cacheKey, order, time.Minute*5)     // 写本地缓存
	}

	return order, nil
}

// NewOrderCacheRepo 创建缓存仓储实例
func NewOrderCacheRepo(rdbRW *RedisRW, db *gorm.DB, localCache LocalCache, circuit CircuitBreaker, orderRepo biz.OrderRepo) biz.OrderRepo {
	return &OrderCacheRepo{
		RedisRW:    rdbRW,
		db:         db,
		localCache: localCache,
		circuit:    circuit,
		dbRepo:     orderRepo,
	}
}

// GetRentalOrderByCache 从缓存查询订单
func (o *OrderCacheRepo) GetRentalOrderByCache(ctx context.Context, tenantId uint, page, pageSize int) (*[]model.RentalOrder, int64, bool, error) {
	cacheKey := generateUserOrderCacheKey(tenantId, page, pageSize)

	// 1. 先查本地缓存(5分钟TTL)
	if val, ok := o.localCache.Get(cacheKey); ok {
		cacheData := val.(CacheData)
		return cacheData.Orders, cacheData.Total, true, nil
	}

	// 2. 本地未命中, 查Redis(2小时TTL)
	redisVal, err := o.RedisRW.Get(ctx, cacheKey).Result()
	if err == nil {
		var cacheData CacheData
		if err := json.Unmarshal([]byte(redisVal), &cacheData); err == nil {
			o.localCache.Set(cacheKey, cacheData, 5*time.Minute)
			return cacheData.Orders, cacheData.Total, true, nil
		}
	}

	// 处理Redis错误(非Key不存在的错误)
	if err != redis.Nil {
		return nil, 0, false, err
	}
	return nil, 0, false, nil
}

// SetRentalOrderToCache 将数据写入缓存
func (o *OrderCacheRepo) SetRentalOrderToCache(ctx context.Context, tenantId uint, page, pageSize int, orders *[]model.RentalOrder, total int64) error {
	cacheKey := generateUserOrderCacheKey(tenantId, page, pageSize)
	cacheData := CacheData{
		Orders: orders,
		Total:  total,
	}

	// 1. 序列化数据
	data, err := json.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %v", err)
	}

	// 2. 写入Redis(2小时过期)
	if err := o.RedisRW.Set(ctx, cacheKey, data, 2*time.Hour).Err(); err != nil {
		return err
	}

	// 3. 写入本地缓存(5分钟过期)
	o.localCache.Set(cacheKey, cacheData, 5*time.Minute)
	return nil
}

// GetRentalOrderByDB 从数据库查询并更新缓存
func (o *OrderCacheRepo) GetRentalOrderByDB(ctx context.Context, tenantId uint, page, pageSize int) (*[]model.RentalOrder, int64, error) {
	// 先查数据库
	orders, total, err := o.dbRepo.GetRentalOrderByDB(ctx, tenantId, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 异步更新缓存：创建独立上下文
	go func(tenantId uint, page, pageSize int, orders *[]model.RentalOrder, total int64) {
		asyncCtx := context.Background()
		_ = o.SetRentalOrderToCache(asyncCtx, tenantId, page, pageSize, orders, total)
	}(tenantId, page, pageSize, orders, total)

	return orders, total, nil
}

// IsCircuitBreakerOpen 检查熔断器状态
func (o *OrderCacheRepo) IsCircuitBreakerOpen() bool {
	return o.circuit.IsOpen()
}

// GetDegradedOrderList 获取降级数据(基础订单信息)
func (o *OrderCacheRepo) GetDegradedOrderList(ctx context.Context, tenantId uint) (*[]model.RentalOrder, int64, error) {
	var degradedOrders []model.RentalOrder
	err := o.db.WithContext(ctx).
		Model(&model.RentalOrder{}).
		Where("tenant_id = ?", tenantId).
		Order("created_at DESC").
		Limit(3).
		Select("id", "order_no", "status").
		Find(&degradedOrders).Error

	if err != nil {
		return nil, 0, err
	}
	return &degradedOrders, int64(len(degradedOrders)), nil
}

// 辅助定义
type CacheData struct {
	Orders *[]model.RentalOrder `json:"order"`
	Total  int64                `json:"total"`
}

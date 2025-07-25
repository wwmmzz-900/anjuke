package biz

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/wwmmzz-900/anjuke/api/house/v3"
	"github.com/wwmmzz-900/anjuke/internal/model"
)

// 常量定义
const (
	// 默认分页大小
	DefaultPageSize = 10
	DefaultPage     = 1
	MaxPageSize     = 100 // 最大分页大小，防止大页面查询
	
	// 个性化推荐默认价格区间
	DefaultMinPrice = 800.0
	DefaultMaxPrice = 1500.0
	
	// 用户行为分析的记录数量
	UserBehaviorAnalysisLimit = 20
	
	// 业务层缓存时间
	RecommendCacheTime = 10 * time.Minute
	
	// 并发控制
	MaxConcurrentRecommendRequests = 50
)

// 错误定义
var (
	ErrInvalidUserID        = fmt.Errorf("无效的用户ID")
	ErrInvalidHouseID       = fmt.Errorf("无效的房源ID")
	ErrHouseAlreadyReserved = fmt.Errorf("您已预约过该房源")
	ErrHouseNotFound        = fmt.Errorf("房源不存在")
)

type House struct {
	HouseID     int64
	Title       string
	Description string
	Price       float64
	Area        float64
	Layout      string
	ImageURL    string
}

type HouseRepo interface {
	GetUserPricePreference(ctx context.Context, userID int64) (float64, float64, error)
	GetPersonalRecommendList(ctx context.Context, minPrice, maxPrice float64, page, pageSize int) ([]*House, int, error)
	GetRecommendList(ctx context.Context, page, pageSize int) ([]*House, int, error)
	CreateReservation(ctx context.Context, reservation *model.HouseReservation) error
	HasReservation(ctx context.Context, userID, houseID int64) (bool, error)
}

type HouseUsecase struct {
	repo HouseRepo
	// 并发控制
	recommendSemaphore chan struct{}
	// 本地缓存（简单的内存缓存）
	localCache struct {
		sync.RWMutex
		data map[string]cacheItem
	}
}

type cacheItem struct {
	data      interface{}
	expiredAt time.Time
}

func NewHouseUsecase(repo HouseRepo) *HouseUsecase {
	uc := &HouseUsecase{
		repo: repo,
		recommendSemaphore: make(chan struct{}, MaxConcurrentRecommendRequests),
	}
	
	// 初始化本地缓存
	uc.localCache.data = make(map[string]cacheItem)
	
	// 启动缓存清理协程
	go uc.cleanupLocalCache()
	
	return uc
}

// 清理过期的本地缓存
func (uc *HouseUsecase) cleanupLocalCache() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		uc.localCache.Lock()
		now := time.Now()
		for key, item := range uc.localCache.data {
			if now.After(item.expiredAt) {
				delete(uc.localCache.data, key)
			}
		}
		uc.localCache.Unlock()
	}
}

// 从本地缓存获取数据
func (uc *HouseUsecase) getFromLocalCache(key string) (interface{}, bool) {
	uc.localCache.RLock()
	defer uc.localCache.RUnlock()
	
	item, exists := uc.localCache.data[key]
	if !exists || time.Now().After(item.expiredAt) {
		return nil, false
	}
	
	return item.data, true
}

// 设置本地缓存
func (uc *HouseUsecase) setLocalCache(key string, data interface{}, duration time.Duration) {
	uc.localCache.Lock()
	defer uc.localCache.Unlock()
	
	uc.localCache.data[key] = cacheItem{
		data:      data,
		expiredAt: time.Now().Add(duration),
	}
}

// 个性化推荐（优化版本）
func (uc *HouseUsecase) PersonalRecommendList(ctx context.Context, userID int64, page, pageSize int) ([]*House, int, error) {
	if userID <= 0 {
		return nil, 0, ErrInvalidUserID
	}
	
	// 参数验证和标准化
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	
	// 生成本地缓存键
	localCacheKey := fmt.Sprintf("personal_recommend_%d_%d_%d", userID, page, pageSize)
	
	// 先检查本地缓存
	if cached, exists := uc.getFromLocalCache(localCacheKey); exists {
		if result, ok := cached.(struct {
			houses []*House
			total  int
		}); ok {
			log.Printf("个性化推荐本地缓存命中: userID=%d, page=%d, pageSize=%d", userID, page, pageSize)
			return result.houses, result.total, nil
		}
	}
	
	// 并发控制
	select {
	case uc.recommendSemaphore <- struct{}{}:
		defer func() { <-uc.recommendSemaphore }()
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	}
	
	// 1. 获取用户偏好的价格区间
	minPrice, maxPrice, err := uc.repo.GetUserPricePreference(ctx, userID)
	if err != nil {
		log.Printf("获取用户偏好失败，使用默认区间: %v", err)
		minPrice, maxPrice = DefaultMinPrice, DefaultMaxPrice
	}
	
	// 如果没有浏览记录，使用默认区间
	if minPrice == 0 && maxPrice == 0 {
		minPrice, maxPrice = DefaultMinPrice, DefaultMaxPrice
	}
	
	// 2. 查询推荐房源
	houses, total, err := uc.repo.GetPersonalRecommendList(ctx, minPrice, maxPrice, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取个性化推荐失败: %w", err)
	}
	
	// 缓存结果到本地缓存
	result := struct {
		houses []*House
		total  int
	}{houses, total}
	uc.setLocalCache(localCacheKey, result, 3*time.Minute) // 本地缓存3分钟
	
	return houses, total, nil
}

// 普通推荐列表（优化版本）
func (uc *HouseUsecase) RecommendList(ctx context.Context, page, pageSize int) ([]*House, int, error) {
	// 参数验证和标准化
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	
	// 生成本地缓存键
	localCacheKey := fmt.Sprintf("recommend_list_%d_%d", page, pageSize)
	
	// 先检查本地缓存
	if cached, exists := uc.getFromLocalCache(localCacheKey); exists {
		if result, ok := cached.(struct {
			houses []*House
			total  int
		}); ok {
			log.Printf("推荐列表本地缓存命中: page=%d, pageSize=%d", page, pageSize)
			return result.houses, result.total, nil
		}
	}
	
	// 并发控制
	select {
	case uc.recommendSemaphore <- struct{}{}:
		defer func() { <-uc.recommendSemaphore }()
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	}
	
	// 查询推荐房源
	houses, total, err := uc.repo.GetRecommendList(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取推荐列表失败: %w", err)
	}
	
	// 缓存结果到本地缓存
	result := struct {
		houses []*House
		total  int
	}{houses, total}
	uc.setLocalCache(localCacheKey, result, 5*time.Minute) // 本地缓存5分钟
	
	return houses, total, nil
}

// 预约看房业务逻辑
func (uc *HouseUsecase) ReserveHouse(ctx context.Context, req *pb.ReserveHouseRequest) error {
	// 参数验证
	if req.UserId <= 0 {
		return ErrInvalidUserID
	}
	if req.HouseId <= 0 {
		return ErrInvalidHouseID
	}
	if req.LandlordId <= 0 {
		return fmt.Errorf("无效的房东ID")
	}
	
	// 1. 校验是否已预约
	has, err := uc.repo.HasReservation(ctx, req.UserId, req.HouseId)
	if err != nil {
		return fmt.Errorf("检查预约状态失败: %w", err)
	}
	if has {
		return ErrHouseAlreadyReserved
	}

	// 2. 构造预约记录
	reservation := &model.HouseReservation{
		LandlordID:  req.LandlordId,
		UserID:      req.UserId,
		UserName:    req.UserName,
		HouseID:     req.HouseId,
		HouseTitle:  req.HouseTitle,
		ReserveTime: req.ReserveTime,
		Status:      model.ReservationStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 3. 保存预约
	if err := uc.repo.CreateReservation(ctx, reservation); err != nil {
		return fmt.Errorf("创建预约失败: %w", err)
	}
	
	return nil
}

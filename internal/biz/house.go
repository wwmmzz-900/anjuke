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

	// 并发控制
	MaxConcurrentRecommendRequests = 50
)

// 错误定义
var (
	ErrInvalidUserID        = fmt.Errorf("无效的用户ID")
	ErrInvalidHouseID       = fmt.Errorf("无效的房源ID")
	ErrHouseAlreadyReserved = fmt.Errorf("您已预约过该房源")
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

type FavoriteHouse struct {
	Id          int64
	UserId      int64
	HouseId     int64
	HouseTitle  string
	HousePrice  float64
	HouseArea   float64
	HouseLayout string
	ImageURL    string
	CreatedAt   time.Time
}

type HouseRepo interface {
	GetUserPricePreference(ctx context.Context, userID int64) (float64, float64, error)
	GetPersonalRecommendList(ctx context.Context, minPrice, maxPrice float64, page, pageSize int) ([]*House, int, error)
	GetRecommendList(ctx context.Context, page, pageSize int) ([]*House, int, error)
	CreateReservation(ctx context.Context, reservation *model.HouseReservation) error
	HasReservation(ctx context.Context, userID, houseID int64) (bool, error)
	
	// 收藏相关接口
	CreateFavorite(ctx context.Context, userID, houseID int64) (*model.Favorite, error)
	DeleteFavorite(ctx context.Context, userID, houseID int64) error
	IsFavorited(ctx context.Context, userID, houseID int64) (bool, error)
	BatchCheckFavoriteStatus(ctx context.Context, userID int64, houseIDs []int64) (map[int64]bool, error)
	GetUserFavoriteList(ctx context.Context, userID int64, page, pageSize int) ([]*FavoriteHouse, int, error)
	GetHouseFavoriteCount(ctx context.Context, houseID int64) (int64, error)
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
		repo:               repo,
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

// 个性化推荐
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
// ============================================================================
// 收藏相关业务逻辑
// ============================================================================

// FavoriteHouse 收藏房源
func (uc *HouseUsecase) FavoriteHouse(ctx context.Context, userID, houseID int64) (*model.Favorite, error) {
	// 参数验证
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	if houseID <= 0 {
		return nil, ErrInvalidHouseID
	}
	
	// 并发控制
	select {
	case uc.recommendSemaphore <- struct{}{}:
		defer func() { <-uc.recommendSemaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	
	// 检查是否已收藏
	isFavorited, err := uc.repo.IsFavorited(ctx, userID, houseID)
	if err != nil {
		return nil, fmt.Errorf("检查收藏状态失败: %w", err)
	}
	
	if isFavorited {
		return nil, fmt.Errorf("房源已收藏")
	}
	
	// 创建收藏记录
	favorite, err := uc.repo.CreateFavorite(ctx, userID, houseID)
	if err != nil {
		return nil, fmt.Errorf("收藏失败: %w", err)
	}
	
	// 清理相关缓存
	uc.clearUserFavoriteCache(userID)
	
	return favorite, nil
}

// UnfavoriteHouse 取消收藏
func (uc *HouseUsecase) UnfavoriteHouse(ctx context.Context, userID, houseID int64) error {
	// 参数验证
	if userID <= 0 {
		return ErrInvalidUserID
	}
	if houseID <= 0 {
		return ErrInvalidHouseID
	}
	
	// 并发控制
	select {
	case uc.recommendSemaphore <- struct{}{}:
		defer func() { <-uc.recommendSemaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}
	
	// 检查是否已收藏
	isFavorited, err := uc.repo.IsFavorited(ctx, userID, houseID)
	if err != nil {
		return fmt.Errorf("检查收藏状态失败: %w", err)
	}
	
	if !isFavorited {
		return fmt.Errorf("房源未收藏")
	}
	
	// 删除收藏记录
	if err := uc.repo.DeleteFavorite(ctx, userID, houseID); err != nil {
		return fmt.Errorf("取消收藏失败: %w", err)
	}
	
	// 清理相关缓存
	uc.clearUserFavoriteCache(userID)
	
	return nil
}

// GetFavoriteList 获取收藏列表
func (uc *HouseUsecase) GetFavoriteList(ctx context.Context, userID int64, page, pageSize int) ([]*FavoriteHouse, int, error) {
	// 参数验证
	if userID <= 0 {
		return nil, 0, ErrInvalidUserID
	}
	
	// 参数标准化
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
	localCacheKey := fmt.Sprintf("favorite_list_%d_%d_%d", userID, page, pageSize)
	
	// 先检查本地缓存
	if cached, exists := uc.getFromLocalCache(localCacheKey); exists {
		if result, ok := cached.(struct {
			houses []*FavoriteHouse
			total  int
		}); ok {
			log.Printf("收藏列表本地缓存命中: userID=%d, page=%d, pageSize=%d", userID, page, pageSize)
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
	
	// 查询收藏列表
	houses, total, err := uc.repo.GetUserFavoriteList(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取收藏列表失败: %w", err)
	}
	
	// 缓存结果到本地缓存
	result := struct {
		houses []*FavoriteHouse
		total  int
	}{houses, total}
	uc.setLocalCache(localCacheKey, result, 3*time.Minute) // 本地缓存3分钟
	
	return houses, total, nil
}

// CheckFavoriteStatus 检查收藏状态
func (uc *HouseUsecase) CheckFavoriteStatus(ctx context.Context, userID int64, houseIDs []int64) (map[int64]bool, error) {
	// 参数验证
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	if len(houseIDs) == 0 {
		return make(map[int64]bool), nil
	}
	
	// 生成本地缓存键
	localCacheKey := fmt.Sprintf("favorite_status_%d_%v", userID, houseIDs)
	
	// 先检查本地缓存
	if cached, exists := uc.getFromLocalCache(localCacheKey); exists {
		if statusMap, ok := cached.(map[int64]bool); ok {
			log.Printf("收藏状态本地缓存命中: userID=%d, houseIDs=%v", userID, houseIDs)
			return statusMap, nil
		}
	}
	
	// 并发控制
	select {
	case uc.recommendSemaphore <- struct{}{}:
		defer func() { <-uc.recommendSemaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	
	// 批量查询收藏状态
	statusMap, err := uc.repo.BatchCheckFavoriteStatus(ctx, userID, houseIDs)
	if err != nil {
		return nil, fmt.Errorf("检查收藏状态失败: %w", err)
	}
	
	// 缓存结果到本地缓存
	uc.setLocalCache(localCacheKey, statusMap, 2*time.Minute) // 本地缓存2分钟
	
	return statusMap, nil
}

// GetHouseFavoriteCount 获取房源收藏数量
func (uc *HouseUsecase) GetHouseFavoriteCount(ctx context.Context, houseID int64) (int64, error) {
	// 参数验证
	if houseID <= 0 {
		return 0, ErrInvalidHouseID
	}
	
	// 生成本地缓存键
	localCacheKey := fmt.Sprintf("favorite_count_%d", houseID)
	
	// 先检查本地缓存
	if cached, exists := uc.getFromLocalCache(localCacheKey); exists {
		if count, ok := cached.(int64); ok {
			log.Printf("收藏数量本地缓存命中: houseID=%d", houseID)
			return count, nil
		}
	}
	
	// 查询收藏数量
	count, err := uc.repo.GetHouseFavoriteCount(ctx, houseID)
	if err != nil {
		return 0, fmt.Errorf("获取收藏数量失败: %w", err)
	}
	
	// 缓存结果到本地缓存
	uc.setLocalCache(localCacheKey, count, 5*time.Minute) // 本地缓存5分钟
	
	return count, nil
}

// clearUserFavoriteCache 清理用户收藏相关的本地缓存
func (uc *HouseUsecase) clearUserFavoriteCache(userID int64) {
	uc.localCache.Lock()
	defer uc.localCache.Unlock()
	
	// 清理收藏列表缓存
	for key := range uc.localCache.data {
		if len(key) > 13 && key[:13] == "favorite_list" {
			// 检查是否是该用户的缓存
			if fmt.Sprintf("favorite_list_%d_", userID) == key[:len(fmt.Sprintf("favorite_list_%d_", userID))] {
				delete(uc.localCache.data, key)
			}
		}
		// 清理收藏状态缓存
		if len(key) > 15 && key[:15] == "favorite_status" {
			// 检查是否是该用户的缓存
			if fmt.Sprintf("favorite_status_%d_", userID) == key[:len(fmt.Sprintf("favorite_status_%d_", userID))] {
				delete(uc.localCache.data, key)
			}
		}
	}
}
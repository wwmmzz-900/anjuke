package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/wwmmzz-900/anjuke/internal/biz"
	"github.com/wwmmzz-900/anjuke/internal/model"
)

// 缓存键常量
const (
	// 房源推荐缓存键前缀
	CacheKeyRecommendList = "house:recommend:list:%d:%d"        // page:pageSize
	CacheKeyPersonalList  = "house:personal:%d:%.2f:%.2f:%d:%d" // userID:minPrice:maxPrice:page:pageSize
	CacheKeyUserPreference = "house:user:preference:%d"         // userID
	CacheKeyHouseImages   = "house:images:%s"                   // houseIDs joined by comma
	CacheKeyHouseDetail   = "house:detail:%d"                   // houseID
	
	// 缓存过期时间
	CacheExpireRecommend   = 15 * time.Minute  // 推荐列表缓存15分钟
	CacheExpirePersonal    = 10 * time.Minute  // 个性化推荐缓存10分钟
	CacheExpirePreference  = 30 * time.Minute  // 用户偏好缓存30分钟
	CacheExpireImages      = 60 * time.Minute  // 图片缓存1小时
	CacheExpireDetail      = 30 * time.Minute  // 房源详情缓存30分钟
	
	// 降级策略
	MaxConcurrentQueries = 100 // 最大并发查询数
	QueryTimeout        = 5 * time.Second // 查询超时时间
)

type houseRepo struct {
	data *Data // Data 结构体包含 *sql.DB
	// 并发控制
	querySemaphore chan struct{}
	// 缓存统计
	cacheStats struct {
		sync.RWMutex
		hits   int64
		misses int64
	}
	// 访问模式跟踪（被动记录）
	accessTracker struct {
		sync.RWMutex
		lastAccess     time.Time
		accessCount    int64
		popularQueries map[string]int64
	}
}

func NewHouseRepo(data *Data) biz.HouseRepo {
	repo := &houseRepo{
		data:           data,
		querySemaphore: make(chan struct{}, MaxConcurrentQueries),
	}
	
	// 初始化访问跟踪器
	repo.accessTracker.popularQueries = make(map[string]int64)
	
	// 启动被动式缓存分析（不主动访问数据库）
	go repo.intelligentCacheManager()
	
	return repo
}

// 智能缓存管理器（被动式缓存管理，不主动访问数据库）
func (r *houseRepo) intelligentCacheManager() {
	// 访问统计和热门查询跟踪
	type accessTracker struct {
		sync.RWMutex
		lastAccess     time.Time
		accessCount    int64
		popularQueries map[string]int64 // 记录热门查询模式
		cacheHitRate   float64          // 缓存命中率
	}
	
	tracker := &accessTracker{
		popularQueries: make(map[string]int64),
	}
	
	// 定期分析访问模式，但不主动查询数据库
	ticker := time.NewTicker(15 * time.Minute) // 每15分钟分析一次
	defer ticker.Stop()
	
	for range ticker.C {
		tracker.Lock()
		
		// 计算缓存命中率
		r.cacheStats.RLock()
		totalRequests := r.cacheStats.hits + r.cacheStats.misses
		if totalRequests > 0 {
			tracker.cacheHitRate = float64(r.cacheStats.hits) / float64(totalRequests)
		}
		r.cacheStats.RUnlock()
		
		// 记录分析结果，但不执行预热
		log.Printf("缓存分析报告 - 命中率: %.2f%%, 最近访问: %v, 访问次数: %d", 
			tracker.cacheHitRate*100, 
			time.Since(tracker.lastAccess),
			tracker.accessCount)
		
		// 清理过期的热门查询记录
		for key, count := range tracker.popularQueries {
			// 热门度衰减
			newCount := count / 2
			if newCount <= 1 {
				delete(tracker.popularQueries, key)
			} else {
				tracker.popularQueries[key] = newCount
			}
		}
		
		// 重置访问计数
		tracker.accessCount = 0
		tracker.Unlock()
	}
}

// 记录访问模式（被动记录，不主动查询）
func (r *houseRepo) recordAccess(queryType string, params map[string]interface{}) {
	r.accessTracker.Lock()
	defer r.accessTracker.Unlock()
	
	// 生成查询键
	queryKey := r.generateQueryKey(queryType, params)
	
	// 记录访问
	r.accessTracker.popularQueries[queryKey]++
	r.accessTracker.lastAccess = time.Now()
	r.accessTracker.accessCount++
	
	// 只记录，不执行任何主动查询
	log.Printf("记录访问模式: %s, 累计访问: %d", queryKey, r.accessTracker.popularQueries[queryKey])
}

// 生成查询键
func (r *houseRepo) generateQueryKey(queryType string, params map[string]interface{}) string {
	key := queryType
	for k, v := range params {
		key += fmt.Sprintf("_%s_%v", k, v)
	}
	return key
}

// 获取动态缓存过期时间（基于访问频率）
func (r *houseRepo) getDynamicCacheExpiration(queryKey string) time.Duration {
	r.accessTracker.RLock()
	count := r.accessTracker.popularQueries[queryKey]
	r.accessTracker.RUnlock()
	
	// 访问频率越高，缓存时间越长
	switch {
	case count > 100:
		return 30 * time.Minute
	case count > 50:
		return 15 * time.Minute
	case count > 10:
		return 10 * time.Minute
	default:
		return 5 * time.Minute
	}
}

// 清理过期缓存（定期清理，不主动查询）
func (r *houseRepo) cleanupStaleCache() {
	// 这里可以添加清理Redis中长时间未访问的缓存键的逻辑
	// 但不会主动查询数据库
	log.Printf("执行缓存清理任务")
}

// 查询用户最近浏览的房源价格区间（带缓存优化）
func (r *houseRepo) GetUserPricePreference(ctx context.Context, userID int64) (float64, float64, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf(CacheKeyUserPreference, userID)
	if cached, err := r.data.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var preference struct {
			MinPrice float64 `json:"min_price"`
			MaxPrice float64 `json:"max_price"`
		}
		if json.Unmarshal([]byte(cached), &preference) == nil {
			r.recordCacheHit()
			return preference.MinPrice, preference.MaxPrice, nil
		}
	}
	
	r.recordCacheMiss()
	
	// 并发控制
	select {
	case r.querySemaphore <- struct{}{}:
		defer func() { <-r.querySemaphore }()
	case <-time.After(QueryTimeout):
		// 超时返回默认值
		return model.DefaultMinPrice, model.FallbackMaxPrice, fmt.Errorf("查询超时，返回默认偏好")
	}

	type result struct {
		MinPrice float64 `json:"min_price"`
		MaxPrice float64 `json:"max_price"`
	}

	var res result

	// 优化查询：添加索引提示和查询超时
	queryCtx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	// 使用user_behavior表查询用户浏览过的房源价格区间
	err := r.data.db.WithContext(queryCtx).
		Table("user_behavior AS ub").
		Select("MIN(h.price) AS min_price, MAX(h.price) AS max_price").
		Joins("JOIN house h ON ub.house_id = h.house_id").
		Where("ub.user_id = ? AND ub.behavior = 'view'", userID).
		Where("ub.created_at >= ?", time.Now().AddDate(0, -3, 0)). // 只查询最近3个月的数据
		Order("ub.created_at DESC").
		Limit(model.MaxRecentViewCount).
		Scan(&res).Error

	if err != nil {
		log.Printf("查询用户偏好失败: %v", err)
		// 如果查询失败，返回默认价格区间
		return model.DefaultMinPrice, model.FallbackMaxPrice, nil
	}

	// 如果没有浏览记录或价格为0，返回默认区间
	if res.MinPrice == 0 && res.MaxPrice == 0 {
		res.MinPrice = model.DefaultMinPrice
		res.MaxPrice = model.FallbackMaxPrice
	}

	// 缓存结果
	if cacheData, err := json.Marshal(res); err == nil {
		r.data.rdb.Set(ctx, cacheKey, cacheData, CacheExpirePreference)
	}

	return res.MinPrice, res.MaxPrice, nil
}

// 查询个性化推荐房源（高并发优化版本）
func (r *houseRepo) GetPersonalRecommendList(ctx context.Context, minPrice, maxPrice float64, page, pageSize int) ([]*biz.House, int, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf(CacheKeyPersonalList, 0, minPrice, maxPrice, page, pageSize) // userID设为0表示通用个性化推荐
	
	// 先尝试从缓存获取
	if cached, err := r.data.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var cacheResult struct {
			Houses []*biz.House `json:"houses"`
			Total  int          `json:"total"`
		}
		if json.Unmarshal([]byte(cached), &cacheResult) == nil {
			r.recordCacheHit()
			log.Printf("个性化推荐缓存命中: minPrice=%.2f, maxPrice=%.2f, page=%d, pageSize=%d", 
				minPrice, maxPrice, page, pageSize)
			return cacheResult.Houses, cacheResult.Total, nil
		}
	}
	
	r.recordCacheMiss()
	
	// 并发控制
	select {
	case r.querySemaphore <- struct{}{}:
		defer func() { <-r.querySemaphore }()
	case <-time.After(QueryTimeout):
		// 超时降级，返回默认数据
		log.Printf("个性化推荐查询超时，返回默认数据")
		return r.getDefaultHouses(), 3, nil
	}

	log.Printf("开始查询个性化推荐: minPrice=%.2f, maxPrice=%.2f, page=%d, pageSize=%d",
		minPrice, maxPrice, page, pageSize)

	// 创建查询超时上下文
	queryCtx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	var total int64
	var houses []*biz.House

	// 优化：使用并发查询总数和列表数据
	var wg sync.WaitGroup
	var countErr, listErr error
	var results []model.House

	// 并发查询总数
	wg.Add(1)
	go func() {
		defer wg.Done()
		countErr = r.data.db.WithContext(queryCtx).
			Table("house").
			Where("price BETWEEN ? AND ? AND status = ?", minPrice, maxPrice, model.HouseStatusActive).
			Count(&total).Error
	}()

	// 并发查询列表数据
	wg.Add(1)
	go func() {
		defer wg.Done()
		listErr = r.data.db.WithContext(queryCtx).
			Table("house").
			Select("house_id, title, description, price, area, layout").
			Where("price BETWEEN ? AND ? AND status = ?", minPrice, maxPrice, model.HouseStatusActive).
			Order("created_at DESC, house_id DESC"). // 优化排序，先按创建时间再按ID
			Limit(pageSize).
			Offset((page - 1) * pageSize).
			Scan(&results).Error
	}()

	wg.Wait()

	// 检查错误
	if countErr != nil {
		log.Printf("查询个性化推荐总数失败: %v", countErr)
		return r.getDefaultHouses(), 3, nil
	}

	if listErr != nil {
		log.Printf("查询个性化推荐列表失败: %v", listErr)
		return r.getDefaultHouses(), 3, nil
	}

	// 如果没有数据，返回默认数据
	if len(results) == 0 {
		log.Printf("未找到符合条件的房源，返回默认数据")
		defaultHouses := r.getDefaultHouses()
		// 缓存默认数据（较短时间）
		r.cacheRecommendResult(cacheKey, defaultHouses, 3, 2*time.Minute)
		return defaultHouses, 3, nil
	}

	// 批量获取房源图片（异步）
	houseIDs := make([]int64, len(results))
	for i, result := range results {
		houseIDs[i] = result.HouseId
	}

	imageMap := r.getHouseImagesWithCache(ctx, houseIDs)

	// 转换为业务层结构体
	houses = make([]*biz.House, 0, len(results))
	for _, result := range results {
		imageURL := imageMap[result.HouseId]

		houses = append(houses, &biz.House{
			HouseID:     result.HouseId,
			Title:       result.Title,
			Description: result.Description,
			Price:       result.Price,
			Area:        float64(result.Area),
			Layout:      result.Layout,
			ImageURL:    imageURL,
		})
	}

	log.Printf("成功查询到 %d 条个性化推荐房源", len(houses))

	// 异步缓存结果
	go r.cacheRecommendResult(cacheKey, houses, int(total), CacheExpirePersonal)

	return houses, int(total), nil
}

// 批量获取房源图片（带缓存优化）
func (r *houseRepo) getHouseImagesWithCache(ctx context.Context, houseIDs []int64) map[int64]string {
	if len(houseIDs) == 0 {
		return make(map[int64]string)
	}

	// 生成缓存键
	houseIDStrs := make([]string, len(houseIDs))
	for i, id := range houseIDs {
		houseIDStrs[i] = strconv.FormatInt(id, 10)
	}
	cacheKey := fmt.Sprintf(CacheKeyHouseImages, fmt.Sprintf("%v", houseIDs))

	// 先尝试从缓存获取
	if cached, err := r.data.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var imageMap map[int64]string
		if json.Unmarshal([]byte(cached), &imageMap) == nil {
			return imageMap
		}
	}

	// 缓存未命中，查询数据库
	imageMap := r.getHouseImages(houseIDs)

	// 异步缓存结果
	go func() {
		if cacheData, err := json.Marshal(imageMap); err == nil {
			r.data.rdb.Set(context.Background(), cacheKey, cacheData, CacheExpireImages)
		}
	}()

	return imageMap
}

// 批量获取房源图片（原始方法，优化查询）
func (r *houseRepo) getHouseImages(houseIDs []int64) map[int64]string {
	if len(houseIDs) == 0 {
		return make(map[int64]string)
	}

	// 使用模型定义
	var images []model.HouseImage
	err := r.data.db.
		Table("house_image").
		Select("house_id, image_url").
		Where("house_id IN ? AND sort_order = 0", houseIDs). // 合并条件，减少查询复杂度
		Scan(&images).Error

	imageMap := make(map[int64]string)
	if err != nil {
		log.Printf("获取房源图片失败: %v", err)
		// 如果获取失败，返回默认图片
		for _, houseID := range houseIDs {
			imageMap[houseID] = "https://example.com/default-house.jpg"
		}
		return imageMap
	}

	log.Printf("成功获取到 %d 张房源图片", len(images))

	// 构建图片映射
	for _, img := range images {
		imageMap[img.HouseID] = img.ImageURL
	}

	// 为没有图片的房源设置默认图片
	for _, houseID := range houseIDs {
		if _, exists := imageMap[houseID]; !exists {
			imageMap[houseID] = "https://example.com/default-house.jpg"
		}
	}

	return imageMap
}

// 获取默认房源数据
func (r *houseRepo) getDefaultHouses() []*biz.House {
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

// 查询推荐房源（高并发优化版本）
func (r *houseRepo) GetRecommendList(ctx context.Context, page, pageSize int) ([]*biz.House, int, error) {
	// 记录访问模式（被动记录）
	r.recordAccessPattern("recommend_list", page, pageSize)
	
	// 生成缓存键
	cacheKey := fmt.Sprintf(CacheKeyRecommendList, page, pageSize)
	
	// 先尝试从缓存获取
	if cached, err := r.data.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var cacheResult struct {
			Houses []*biz.House `json:"houses"`
			Total  int          `json:"total"`
		}
		if json.Unmarshal([]byte(cached), &cacheResult) == nil {
			r.recordCacheHit()
			log.Printf("推荐列表缓存命中: page=%d, pageSize=%d", page, pageSize)
			return cacheResult.Houses, cacheResult.Total, nil
		}
	}
	
	r.recordCacheMiss()
	
	// 并发控制
	select {
	case r.querySemaphore <- struct{}{}:
		defer func() { <-r.querySemaphore }()
	case <-time.After(QueryTimeout):
		// 超时降级，返回默认数据
		log.Printf("推荐列表查询超时，返回默认数据")
		return r.getDefaultHouses(), 3, nil
	}

	log.Printf("开始查询推荐房源: page=%d, pageSize=%d", page, pageSize)
	
	// 记录访问统计
	r.recordAccess("recommend_list", map[string]interface{}{
		"page":     page,
		"pageSize": pageSize,
	})

	// 创建查询超时上下文
	queryCtx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	var total int64
	var houses []*biz.House

	// 优化：使用并发查询总数和列表数据
	var wg sync.WaitGroup
	var countErr, listErr error
	var results []model.House

	// 并发查询总数
	wg.Add(1)
	go func() {
		defer wg.Done()
		countErr = r.data.db.WithContext(queryCtx).
			Table("house").
			Where("status = ?", model.HouseStatusActive).
			Count(&total).Error
	}()

	// 并发查询列表数据
	wg.Add(1)
	go func() {
		defer wg.Done()
		listErr = r.data.db.WithContext(queryCtx).
			Table("house").
			Select("house_id, title, description, price, area, layout, created_at").
			Where("status = ?", model.HouseStatusActive).
			Order("created_at DESC, house_id DESC"). // 优化排序策略
			Limit(pageSize).
			Offset((page - 1) * pageSize).
			Scan(&results).Error
	}()

	wg.Wait()

	// 检查错误
	if countErr != nil {
		log.Printf("查询推荐房源总数失败: %v", countErr)
		return r.getDefaultHouses(), 3, nil
	}

	if listErr != nil {
		log.Printf("查询推荐房源列表失败: %v", listErr)
		return r.getDefaultHouses(), 3, nil
	}

	// 如果没有数据，返回默认数据
	if len(results) == 0 {
		log.Printf("未找到符合条件的房源，返回默认数据")
		defaultHouses := r.getDefaultHouses()
		// 缓存默认数据（较短时间）
		r.cacheRecommendResult(cacheKey, defaultHouses, 3, 2*time.Minute)
		return defaultHouses, 3, nil
	}

	// 批量获取房源图片（带缓存）
	houseIDs := make([]int64, len(results))
	for i, result := range results {
		houseIDs[i] = result.HouseId
	}

	imageMap := r.getHouseImagesWithCache(ctx, houseIDs)

	// 转换为业务层结构体
	houses = make([]*biz.House, 0, len(results))
	for _, result := range results {
		imageURL := imageMap[result.HouseId]

		houses = append(houses, &biz.House{
			HouseID:     result.HouseId,
			Title:       result.Title,
			Description: result.Description,
			Price:       result.Price,
			Area:        float64(result.Area),
			Layout:      result.Layout,
			ImageURL:    imageURL,
		})
	}

	log.Printf("成功查询到 %d 条推荐房源", len(houses))

	// 异步缓存结果
	go r.cacheRecommendResult(cacheKey, houses, int(total), CacheExpireRecommend)

	return houses, int(total), nil
}

// 获取房源的第一张图片
func (r *houseRepo) getHouseFirstImage(houseID int64) string {
	var imageURL string
	err := r.data.db.
		Table("house_image").
		Select("image_url").
		Where("house_id = ?", houseID).
		Order("sort_order ASC").
		Limit(1).
		Scan(&imageURL).Error

	if err != nil || imageURL == "" {
		// 如果没有找到图片，返回默认图片URL
		return "https://example.com/default-house.jpg"
	}

	return imageURL
}

// 批量获取房源图片
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

// 缓存推荐结果
func (r *houseRepo) cacheRecommendResult(cacheKey string, houses []*biz.House, total int, expiration time.Duration) {
	cacheResult := struct {
		Houses []*biz.House `json:"houses"`
		Total  int          `json:"total"`
	}{
		Houses: houses,
		Total:  total,
	}
	
	if cacheData, err := json.Marshal(cacheResult); err == nil {
		if err := r.data.rdb.Set(context.Background(), cacheKey, cacheData, expiration).Err(); err != nil {
			log.Printf("缓存推荐结果失败: %v", err)
		}
	}
}

// 记录缓存命中
func (r *houseRepo) recordCacheHit() {
	r.cacheStats.Lock()
	r.cacheStats.hits++
	r.cacheStats.Unlock()
}

// 记录缓存未命中
func (r *houseRepo) recordCacheMiss() {
	r.cacheStats.Lock()
	r.cacheStats.misses++
	r.cacheStats.Unlock()
}

// 记录访问模式（被动记录，不触发预热）
func (r *houseRepo) recordAccessPattern(queryType string, params ...interface{}) {
	r.accessTracker.Lock()
	defer r.accessTracker.Unlock()
	
	r.accessTracker.lastAccess = time.Now()
	r.accessTracker.accessCount++
	
	// 生成查询模式键
	patternKey := fmt.Sprintf("%s:%v", queryType, params)
	r.accessTracker.popularQueries[patternKey]++
}

// 获取缓存统计
func (r *houseRepo) GetCacheStats() (hits, misses int64) {
	r.cacheStats.RLock()
	defer r.cacheStats.RUnlock()
	return r.cacheStats.hits, r.cacheStats.misses
}

// 清理过期缓存（定期任务）
func (r *houseRepo) cleanupExpiredCache() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// 这里可以添加清理逻辑，Redis会自动处理过期键
		// 可以添加一些统计信息的清理
		log.Printf("缓存清理任务执行")
	}
}

// 智能缓存分析（只分析，不主动查询）
func (r *houseRepo) analyzeCachePatterns() {
	r.accessTracker.RLock()
	defer r.accessTracker.RUnlock()
	
	// 分析热门查询模式
	var hotQueries []string
	for query, count := range r.accessTracker.popularQueries {
		if count > 10 { // 访问次数超过10次的查询
			hotQueries = append(hotQueries, query)
		}
	}
	
	// 只记录分析结果，不执行查询
	if len(hotQueries) > 0 {
		log.Printf("发现热门查询模式: %v", hotQueries)
		log.Printf("建议优化这些查询的缓存策略")
	}
	
	// 计算缓存效率
	r.cacheStats.RLock()
	totalRequests := r.cacheStats.hits + r.cacheStats.misses
	hitRate := float64(0)
	if totalRequests > 0 {
		hitRate = float64(r.cacheStats.hits) / float64(totalRequests)
	}
	r.cacheStats.RUnlock()
	
	log.Printf("当前缓存命中率: %.2f%% (%d/%d)", hitRate*100, r.cacheStats.hits, totalRequests)
}

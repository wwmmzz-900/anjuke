package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/wwmmzz-900/anjuke/internal/biz"
	"github.com/wwmmzz-900/anjuke/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &UserRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// BloggerProfileRepo 博主主页数据访问层实现
type BloggerProfileRepo struct {
	data *Data
	log  *log.Helper
}

// NewBloggerProfileRepo 创建博主主页数据访问层实例
func NewBloggerProfileRepo(data *Data, logger log.Logger) biz.BloggerProfileRepo {
	return &BloggerProfileRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetUserById 根据用户ID获取用户基础信息
func (r *BloggerProfileRepo) GetUserById(ctx context.Context, userId int64) (*model.UserBase, error) {
	var user model.UserBase
	err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userId).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("用户不存在")
	}

	return &user, err
}

// GetUserHouseStats 获取用户房源统计信息
func (r *BloggerProfileRepo) GetUserHouseStats(ctx context.Context, userId int64) (*model.HouseStatistics, error) {
	var stats model.HouseStatistics

	// 总房源数
	err := r.data.db.WithContext(ctx).
		Model(&model.House{}).
		Where("landlord_id = ? AND deleted_at IS NULL", userId).
		Count(&stats.TotalCount).Error
	if err != nil {
		return nil, err
	}

	// 活跃房源数
	err = r.data.db.WithContext(ctx).
		Model(&model.House{}).
		Where("landlord_id = ? AND status = ? AND deleted_at IS NULL", userId, "active").
		Count(&stats.ActiveCount).Error
	if err != nil {
		return nil, err
	}

	// 已租房源数
	err = r.data.db.WithContext(ctx).
		Model(&model.House{}).
		Where("landlord_id = ? AND status = ? AND deleted_at IS NULL", userId, "rented").
		Count(&stats.RentedCount).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetUserInteractStats 获取用户互动统计信息
func (r *BloggerProfileRepo) GetUserInteractStats(ctx context.Context, userId int64) (*model.InteractStatistics, error) {
	var stats model.InteractStatistics

	// 获取用户所有房源ID
	var houseIds []int64
	err := r.data.db.WithContext(ctx).
		Model(&model.House{}).
		Where("landlord_id = ? AND deleted_at IS NULL", userId).
		Pluck("house_id", &houseIds).Error
	if err != nil {
		return nil, err
	}

	if len(houseIds) == 0 {
		return &stats, nil
	}

	// 统计收藏量
	err = r.data.db.WithContext(ctx).
		Model(&model.Favorite{}).
		Where("house_id IN ? AND deleted_at IS NULL", houseIds).
		Count(&stats.TotalFavorites).Error
	if err != nil {
		return nil, err
	}

	// 统计预约量
	err = r.data.db.WithContext(ctx).
		Model(&model.HouseReservation{}).
		Where("house_id IN ?", houseIds).
		Count(&stats.TotalReservations).Error
	if err != nil {
		return nil, err
	}

	// 这里暂时设置浏览量和响应率为默认值，实际项目中需要从其他表获取
	stats.TotalViews = stats.TotalFavorites * 10 // 假设浏览量是收藏量的10倍
	stats.ResponseRate = 0.85                    // 假设响应率为85%

	return &stats, nil
}

// GetUserRecentHouses 获取用户最近发布的房源
func (r *BloggerProfileRepo) GetUserRecentHouses(ctx context.Context, userId int64, limit int) ([]*model.House, error) {
	var houses []*model.House
	err := r.data.db.WithContext(ctx).
		Where("landlord_id = ? AND deleted_at IS NULL", userId).
		Order("created_at DESC").
		Limit(limit).
		Find(&houses).Error

	return houses, err
}

// GetUserHouses 分页获取用户房源列表
func (r *BloggerProfileRepo) GetUserHouses(ctx context.Context, userId int64, pagination *model.PaginationParams, filter *model.HouseFilterParams) ([]*model.House, *model.PaginationResult, error) {
	var houses []*model.House
	var total int64

	// 构建基础查询
	query := r.data.db.WithContext(ctx).
		Where("landlord_id = ? AND deleted_at IS NULL", userId)

	// 应用过滤条件
	query = r.applyHouseFilters(query, filter)

	// 获取总数
	err := query.Model(&model.House{}).Count(&total).Error
	if err != nil {
		return nil, nil, err
	}

	// 分页查询
	offset := pagination.GetOffset()
	orderClause := pagination.GetOrderClause()

	err = query.Order(orderClause).
		Offset(offset).
		Limit(pagination.PageSize).
		Find(&houses).Error

	if err != nil {
		return nil, nil, err
	}

	// 构建分页结果
	paginationResult := &model.PaginationResult{
		Total:    total,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}
	paginationResult.CalculatePages()

	return houses, paginationResult, nil
}

// applyHouseFilters 应用房源过滤条件
func (r *BloggerProfileRepo) applyHouseFilters(query *gorm.DB, filter *model.HouseFilterParams) *gorm.DB {
	if filter == nil {
		return query
	}

	// 状态过滤
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// 价格范围过滤
	if filter.MinPrice > 0 {
		query = query.Where("price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("price <= ?", filter.MaxPrice)
	}

	// 面积范围过滤
	if filter.MinArea > 0 {
		query = query.Where("area >= ?", filter.MinArea)
	}
	if filter.MaxArea > 0 {
		query = query.Where("area <= ?", filter.MaxArea)
	}

	// 户型过滤
	if filter.Layout != "" {
		query = query.Where("layout = ?", filter.Layout)
	}

	// 区域过滤
	if filter.RegionId > 0 {
		query = query.Where("region_id = ?", filter.RegionId)
	}

	// 关键词搜索（标题和地址）
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("title LIKE ? OR address LIKE ?", keyword, keyword)
	}

	return query
}

// CreateAccessLog 创建访问日志记录
func (r *BloggerProfileRepo) CreateAccessLog(ctx context.Context, accessLog *model.BloggerProfileAccessLog) error {
	// 验证访问日志数据
	if !accessLog.IsValidLog() {
		r.log.WithContext(ctx).Errorf("Invalid access log data: bloggerID=%d, visitorIP=%s",
			accessLog.BloggerID, accessLog.VisitorIP)
		return errors.New("无效的访问日志数据")
	}

	// 格式化访问日志数据
	r.formatAccessLog(accessLog)

	// 设置默认值
	if accessLog.StatusCode == 0 {
		accessLog.StatusCode = 200
	}
	if accessLog.DeviceType == "" {
		accessLog.DeviceType = accessLog.GetDeviceTypeFromUserAgent()
	}
	if accessLog.Browser == "" {
		accessLog.Browser = accessLog.GetBrowserFromUserAgent()
	}
	if accessLog.Platform == "" {
		accessLog.Platform = accessLog.GetPlatformFromUserAgent()
	}

	// 创建访问日志记录
	err := r.data.db.WithContext(ctx).Create(accessLog).Error
	if err != nil {
		r.log.WithContext(ctx).Errorf("CreateAccessLog failed: %v, bloggerID=%d, visitorID=%d",
			err, accessLog.BloggerID, accessLog.VisitorID)
		return err
	}

	r.log.WithContext(ctx).Infof("AccessLog created successfully: id=%d, bloggerID=%d, visitorID=%d, ip=%s, device=%s",
		accessLog.ID, accessLog.BloggerID, accessLog.VisitorID, accessLog.VisitorIP, accessLog.DeviceType)

	return nil
}

// formatAccessLog 格式化访问日志数据
func (r *BloggerProfileRepo) formatAccessLog(accessLog *model.BloggerProfileAccessLog) {
	// 清理和格式化IP地址
	accessLog.VisitorIP = r.cleanIPAddress(accessLog.VisitorIP)

	// 清理和格式化User-Agent
	if len(accessLog.UserAgent) > 500 {
		accessLog.UserAgent = accessLog.UserAgent[:500]
	}

	// 清理和格式化Referer
	if len(accessLog.Referer) > 500 {
		accessLog.Referer = accessLog.Referer[:500]
	}

	// 清理和格式化请求路径
	if len(accessLog.RequestPath) > 255 {
		accessLog.RequestPath = accessLog.RequestPath[:255]
	}

	// 清理和格式化会话ID
	if len(accessLog.SessionID) > 100 {
		accessLog.SessionID = accessLog.SessionID[:100]
	}
}

// cleanIPAddress 清理IP地址格式
func (r *BloggerProfileRepo) cleanIPAddress(ip string) string {
	// 移除端口号
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	// 验证IP地址格式
	if net.ParseIP(ip) == nil {
		return "unknown"
	}

	return ip
}

// UpdateAccessStats 更新访问统计数据
func (r *BloggerProfileRepo) UpdateAccessStats(ctx context.Context, bloggerID int64) error {
	r.log.WithContext(ctx).Infof("Updating access stats for bloggerID=%d", bloggerID)

	// 获取当前统计数据
	var stats model.BloggerProfileAccessStats
	err := r.data.db.WithContext(ctx).
		Where("blogger_id = ?", bloggerID).
		First(&stats).Error

	// 如果统计记录不存在，创建新记录
	if errors.Is(err, gorm.ErrRecordNotFound) {
		stats = model.BloggerProfileAccessStats{
			BloggerID: bloggerID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		r.log.WithContext(ctx).Infof("Creating new access stats record for bloggerID=%d", bloggerID)
	} else if err != nil {
		r.log.WithContext(ctx).Errorf("Failed to get existing access stats: %v", err)
		return err
	}

	// 计算各时间段的访问量
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 使用事务确保数据一致性
	tx := r.data.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 统计总访问量
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ?", bloggerID).
		Count(&stats.TotalViews).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to count total views: %v", err)
		return err
	}

	// 统计今日访问量
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ? AND created_at >= ?", bloggerID, today).
		Count(&stats.TodayViews).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to count today views: %v", err)
		return err
	}

	// 统计昨日访问量（用于计算增长率）
	var yesterdayViews int64
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ? AND created_at >= ? AND created_at < ?", bloggerID, yesterday, today).
		Count(&yesterdayViews).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to count yesterday views: %v", err)
		return err
	}

	// 统计本周访问量
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ? AND created_at >= ?", bloggerID, weekStart).
		Count(&stats.WeekViews).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to count week views: %v", err)
		return err
	}

	// 统计本月访问量
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ? AND created_at >= ?", bloggerID, monthStart).
		Count(&stats.MonthViews).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to count month views: %v", err)
		return err
	}

	// 统计独立访客数（基于IP去重）
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ?", bloggerID).
		Distinct("visitor_ip").
		Count(&stats.UniqueVisitors).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to count unique visitors: %v", err)
		return err
	}

	// 计算平均响应时间
	var avgResponseTime sql.NullFloat64
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ? AND response_time > 0", bloggerID).
		Select("AVG(response_time)").
		Scan(&avgResponseTime).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to calculate avg response time: %v", err)
		return err
	}

	if avgResponseTime.Valid {
		stats.AvgResponseTime = avgResponseTime.Float64
	}

	// 更新最后访问时间
	var lastAccessTime time.Time
	err = tx.Model(&model.BloggerProfileAccessLog{}).
		Where("blogger_id = ?", bloggerID).
		Select("MAX(created_at)").
		Scan(&lastAccessTime).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to get last access time: %v", err)
		return err
	}

	if !lastAccessTime.IsZero() {
		stats.LastAccessAt = &lastAccessTime
	}

	// 更新统计记录
	stats.UpdatedAt = time.Now()

	// 使用 UPSERT 操作
	err = tx.Save(&stats).Error
	if err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to save access stats: %v", err)
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		r.log.WithContext(ctx).Errorf("Failed to commit transaction: %v", err)
		return err
	}

	// 计算增长率
	growthRate := stats.GetViewsGrowthRate(yesterdayViews)

	r.log.WithContext(ctx).Infof("AccessStats updated successfully: bloggerID=%d, totalViews=%d, todayViews=%d, growthRate=%.2f%%",
		bloggerID, stats.TotalViews, stats.TodayViews, growthRate*100)

	return nil
}

// GetAccessStats 获取访问统计数据
func (r *BloggerProfileRepo) GetAccessStats(ctx context.Context, bloggerID int64) (*model.BloggerProfileAccessStats, error) {
	r.log.WithContext(ctx).Infof("Getting access stats for bloggerID=%d", bloggerID)

	var stats model.BloggerProfileAccessStats
	err := r.data.db.WithContext(ctx).
		Where("blogger_id = ?", bloggerID).
		First(&stats).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果统计记录不存在，返回空统计数据
		r.log.WithContext(ctx).Infof("No access stats found for bloggerID=%d, returning empty stats", bloggerID)
		return &model.BloggerProfileAccessStats{
			BloggerID: bloggerID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	if err != nil {
		r.log.WithContext(ctx).Errorf("Failed to get access stats: %v", err)
		return nil, err
	}

	r.log.WithContext(ctx).Infof("Retrieved access stats: bloggerID=%d, totalViews=%d, uniqueVisitors=%d",
		bloggerID, stats.TotalViews, stats.UniqueVisitors)

	return &stats, nil
}

// GetAccessStatsWithCache 获取访问统计数据（带缓存）
func (r *BloggerProfileRepo) GetAccessStatsWithCache(ctx context.Context, bloggerID int64) (*model.BloggerProfileAccessStats, error) {
	// 这里可以实现缓存逻辑，暂时直接调用原方法
	// TODO: 实现Redis缓存机制
	return r.GetAccessStats(ctx, bloggerID)
}

// GetAccessStatsSummary 获取访问统计摘要
func (r *BloggerProfileRepo) GetAccessStatsSummary(ctx context.Context, bloggerID int64) (map[string]interface{}, error) {
	stats, err := r.GetAccessStats(ctx, bloggerID)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"blogger_id":        stats.BloggerID,
		"total_views":       stats.TotalViews,
		"today_views":       stats.TodayViews,
		"week_views":        stats.WeekViews,
		"month_views":       stats.MonthViews,
		"unique_visitors":   stats.UniqueVisitors,
		"avg_response_time": stats.AvgResponseTime,
		"last_access_at":    stats.LastAccessAt,
		"updated_at":        stats.UpdatedAt,
		// 格式化显示
		"total_views_formatted":       stats.FormatTotalViews(),
		"unique_visitors_formatted":   stats.FormatUniqueVisitors(),
		"avg_response_time_formatted": fmt.Sprintf("%.2fms", stats.AvgResponseTime),
	}

	return summary, nil
}

// todo:用户添加
//func (u UserRepo) CreateUser(ctx context.Context, user *biz.User) (*biz.User, error) {
//	//TODO implement me
//	if user.Mobile == "" {
//		return nil, fmt.Errorf("手机号不能为空")
//	}
//
//	err := u.data.db.Debug().WithContext(ctx).Create(user).Error
//	if err != nil {
//		return nil, fmt.Errorf("创建用户失败: %v", err)
//	}
//	return user, nil
//}

// todo：根据phone查询用户
//func (u UserRepo) GetUser(ctx context.Context, phone string) (*biz.User, error) {
//	var user biz.User
//	err := u.data.db.Debug().WithContext(ctx).Where("mobile = ?", phone).Limit(1).Find(&user).Error
//
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		return nil, nil // 明确返回nil表示用户不存在
//	}
//	if err != nil {
//		return nil, fmt.Errorf("查询用户失败: %v", err)
//	}
//	return &user, nil
//}

// Package data 实现了积分模块的数据访问层
package data

import (
	"context"
	"fmt"
	"time"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// PointsRepo 实现了 domain.PointsRepo 接口
type PointsRepo struct {
	data *Data
	log  *log.Helper
}

// NewPointsRepo 创建积分仓储
func NewPointsRepo(data *Data, logger log.Logger) domain.PointsRepo {
	return &PointsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetUserPoints 查询用户积分余额
func (r *PointsRepo) GetUserPoints(ctx context.Context, userID uint64) (*domain.UserPoints, error) {
	var userPoints domain.UserPoints

	err := r.data.db.Where("user_id = ?", userID).First(&userPoints).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户积分记录不存在，创建初始记录
			userPoints = domain.UserPoints{
				UserID:      userID,
				TotalPoints: 0,
			}
			if err := r.data.db.Create(&userPoints).Error; err != nil {
				return nil, fmt.Errorf("创建用户积分记录失败: %v", err)
			}
		} else {
			return nil, fmt.Errorf("查询用户积分失败: %v", err)
		}
	}

	return &userPoints, nil
}

// GetPointsHistory 查询积分明细记录
func (r *PointsRepo) GetPointsHistory(ctx context.Context, userID uint64, page, pageSize int32, pointsType string) ([]*domain.PointsRecord, int32, error) {
	var records []*domain.PointsRecord
	var total int64

	query := r.data.db.Model(&domain.PointsRecord{}).Where("user_id = ?", userID)

	// 根据类型筛选
	if pointsType == "earn" {
		query = query.Where("points > 0")
	} else if pointsType == "use" {
		query = query.Where("points < 0")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计积分记录总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("查询积分记录失败: %v", err)
	}

	return records, int32(total), nil
}

// CheckIn 签到获取积分
func (r *PointsRepo) CheckIn(ctx context.Context, userID uint64) (*domain.CheckInResult, error) {
	// 开启事务
	tx := r.data.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取连续签到天数（从昨天开始计算）
	consecutiveDays, err := r.GetConsecutiveCheckInDays(ctx, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 计算签到积分（基础积分 + 连续签到奖励）
	points := domain.CheckInBasePoints
	// 连续签到天数达到7的倍数时给予奖励
	if (consecutiveDays+1)%7 == 0 {
		points += domain.CheckInBonusPoints // 连续7天奖励
	}

	today := time.Now().Format("2006-01-02")

	// 1. 创建签到记录
	checkInRecord := &domain.CheckInRecord{
		UserID:    userID,
		CheckDate: today,
		Points:    int64(points),
	}
	if err := tx.Create(checkInRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建签到记录失败: %v", err)
	}

	// 2. 创建积分记录
	pointsRecord := &domain.PointsRecord{
		UserID:      userID,
		Type:        domain.PointsTypeCheckIn,
		Points:      int64(points),
		Description: fmt.Sprintf("签到获得积分（连续%d天）", consecutiveDays+1),
	}
	if err := tx.Create(pointsRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建积分记录失败: %v", err)
	}

	// 3. 更新用户积分
	var userPoints domain.UserPoints
	err = tx.Where("user_id = ?", userID).First(&userPoints).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的积分记录
			userPoints = domain.UserPoints{
				UserID:      userID,
				TotalPoints: int64(points),
			}
			if err := tx.Create(&userPoints).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建用户积分记录失败: %v", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("查询用户积分失败: %v", err)
		}
	} else {
		// 更新现有积分
		userPoints.TotalPoints += int64(points)
		if err := tx.Save(&userPoints).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新用户积分失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("事务提交失败: %v", err)
	}

	r.log.Infof("用户签到成功: user_id=%d, points=%d, consecutive_days=%d", userID, points, consecutiveDays+1)

	return &domain.CheckInResult{
		PointsEarned:    int64(points),
		TotalPoints:     userPoints.TotalPoints,
		ConsecutiveDays: consecutiveDays + 1,
	}, nil
}

// EarnPointsByConsume 消费获取积分
func (r *PointsRepo) EarnPointsByConsume(ctx context.Context, userID uint64, orderID string, amount int64) (*domain.EarnResult, error) {
	// 计算获得的积分（1元=1积分）
	points := amount / 100 * domain.ConsumePointsRate * 100
	if points <= 0 {
		return nil, fmt.Errorf("消费金额太小，无法获得积分")
	}

	// 开启事务
	tx := r.data.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查是否已经为该订单发放过积分
	var existingRecord domain.PointsRecord
	err := tx.Where("user_id = ? AND order_id = ? AND type = ?", userID, orderID, domain.PointsTypeConsume).First(&existingRecord).Error
	if err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("该订单已发放过积分")
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("检查订单积分记录失败: %v", err)
	}

	// 2. 创建积分记录
	pointsRecord := &domain.PointsRecord{
		UserID:      userID,
		Type:        domain.PointsTypeConsume,
		Points:      points,
		Description: fmt.Sprintf("消费获得积分（订单金额：%.2f元）", float64(points)),
		OrderID:     orderID,
		Amount:      amount,
	}
	if err := tx.Create(pointsRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建积分记录失败: %v", err)
	}

	// 3. 更新用户积分
	var userPoints domain.UserPoints
	err = tx.Where("user_id = ?", userID).First(&userPoints).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的积分记录
			userPoints = domain.UserPoints{
				UserID:      userID,
				TotalPoints: points,
			}
			if err := tx.Create(&userPoints).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建用户积分记录失败: %v", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("查询用户积分失败: %v", err)
		}
	} else {
		// 更新现有积分
		userPoints.TotalPoints += points
		if err := tx.Save(&userPoints).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新用户积分失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("事务提交失败: %v", err)
	}

	r.log.Infof("消费获得积分成功: user_id=%d, order_id=%s, amount=%d, points=%d", userID, orderID, amount, points)

	return &domain.EarnResult{
		PointsEarned: points,
		TotalPoints:  userPoints.TotalPoints,
	}, nil
}

// UsePoints 使用积分抵扣
func (r *PointsRepo) UsePoints(ctx context.Context, userID uint64, points int64, orderID, description string) (*domain.UseResult, error) {
	// 计算抵扣金额（10积分=1元）
	amountDeducted := points / domain.UsePointsRate * 100 // 转换为分

	// 开启事务
	tx := r.data.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查询用户积分并加锁
	var userPoints domain.UserPoints
	err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&userPoints).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户积分记录不存在")
		}
		return nil, fmt.Errorf("查询用户积分失败: %v", err)
	}

	// 2. 检查积分余额
	if userPoints.TotalPoints < points {
		tx.Rollback()
		return nil, fmt.Errorf("积分余额不足，当前总积分: %d", userPoints.TotalPoints)
	}

	// 3. 创建积分使用记录
	if description == "" {
		description = fmt.Sprintf("积分抵扣（抵扣金额：%.2f元）", float64(amountDeducted)/100)
	}
	pointsRecord := &domain.PointsRecord{
		UserID:      userID,
		Type:        domain.PointsTypeUse,
		Points:      -points, // 负数表示消费
		Description: description,
		OrderID:     orderID,
		Amount:      amountDeducted,
	}
	if err := tx.Create(pointsRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建积分记录失败: %v", err)
	}

	// 4. 更新用户积分（直接扣除总积分）
	userPoints.TotalPoints -= points
	if err := tx.Save(&userPoints).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新用户积分失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("事务提交失败: %v", err)
	}

	r.log.Infof("积分使用成功: user_id=%d, points=%d, amount_deducted=%d", userID, points, amountDeducted)

	return &domain.UseResult{
		PointsUsed:     points,
		AmountDeducted: amountDeducted,
		TotalPoints:    userPoints.TotalPoints,
	}, nil
}

// HasCheckedInToday 检查今日是否已签到
func (r *PointsRepo) HasCheckedInToday(ctx context.Context, userID uint64) (bool, error) {
	today := time.Now().Format("2006-01-02")

	var count int64
	err := r.data.db.Model(&domain.CheckInRecord{}).
		Where("user_id = ? AND check_date = ?", userID, today).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("检查签到记录失败: %v", err)
	}

	return count > 0, nil
}

// GetConsecutiveCheckInDays 获取连续签到天数
func (r *PointsRepo) GetConsecutiveCheckInDays(ctx context.Context, userID uint64) (int32, error) {
	// 获取最近的签到记录，按日期倒序
	var records []domain.CheckInRecord
	err := r.data.db.Where("user_id = ?", userID).
		Order("check_date DESC").
		Limit(30). // 最多查询30天
		Find(&records).Error

	if err != nil {
		return 0, fmt.Errorf("查询签到记录失败: %v", err)
	}

	if len(records) == 0 {
		return 0, nil
	}

	// 计算连续签到天数（从昨天开始计算）
	consecutiveDays := int32(0)
	yesterday := time.Now().AddDate(0, 0, -1)

	for i, record := range records {
		checkDate, err := time.Parse("2006-01-02", record.CheckDate)
		if err != nil {
			continue
		}

		// 期望的日期：昨天、前天、大前天...
		expectedDate := yesterday.AddDate(0, 0, -i)
		if checkDate.Format("2006-01-02") == expectedDate.Format("2006-01-02") {
			consecutiveDays++
		} else {
			break
		}
	}

	return consecutiveDays, nil
}

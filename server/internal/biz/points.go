// Package biz 封装了积分模块的核心业务逻辑
package biz

import (
	"context"
	"fmt"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
)

// PointsUsecase 封装了积分相关的业务逻辑
type PointsUsecase struct {
	repo domain.PointsRepo
	log  *log.Helper
}

// NewPointsUsecase 创建积分业务用例
func NewPointsUsecase(repo domain.PointsRepo, logger log.Logger) *PointsUsecase {
	return &PointsUsecase{repo: repo, log: log.NewHelper(logger)}
}

// GetUserPoints 查询用户积分余额
func (uc *PointsUsecase) GetUserPoints(ctx context.Context, userID uint64) (*domain.UserPoints, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	uc.log.WithContext(ctx).Infof("查询用户积分余额: user_id=%d", userID)
	return uc.repo.GetUserPoints(ctx, userID)
}

// GetPointsHistory 查询积分明细记录
func (uc *PointsUsecase) GetPointsHistory(ctx context.Context, userID uint64, page, pageSize int32, pointsType string) ([]*domain.PointsRecord, int32, error) {
	if userID == 0 {
		return nil, 0, fmt.Errorf("用户ID不能为空")
	}

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 验证类型参数
	if pointsType != "" && pointsType != "earn" && pointsType != "use" {
		return nil, 0, fmt.Errorf("无效的积分类型: %s", pointsType)
	}

	uc.log.WithContext(ctx).Infof("查询积分明细: user_id=%d, page=%d, page_size=%d, type=%s", userID, page, pageSize, pointsType)
	return uc.repo.GetPointsHistory(ctx, userID, page, pageSize, pointsType)
}

// CheckIn 签到获取积分
func (uc *PointsUsecase) CheckIn(ctx context.Context, userID uint64) (*domain.CheckInResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// 检查今日是否已签到
	hasCheckedIn, err := uc.repo.HasCheckedInToday(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("检查签到状态失败: %v", err)
	}
	if hasCheckedIn {
		return nil, fmt.Errorf("今日已签到，请明天再来")
	}

	uc.log.WithContext(ctx).Infof("用户签到: user_id=%d", userID)
	return uc.repo.CheckIn(ctx, userID)
}

// EarnPointsByConsume 消费获取积分
func (uc *PointsUsecase) EarnPointsByConsume(ctx context.Context, userID uint64, orderID string, amount int64) (*domain.EarnResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}
	if orderID == "" {
		return nil, fmt.Errorf("订单ID不能为空")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("消费金额必须大于0")
	}

	uc.log.WithContext(ctx).Infof("消费获取积分: user_id=%d, order_id=%s, amount=%d", userID, orderID, amount)
	return uc.repo.EarnPointsByConsume(ctx, userID, orderID, amount)
}

// UsePoints 使用积分抵扣
func (uc *PointsUsecase) UsePoints(ctx context.Context, userID uint64, points int64, orderID, description string) (*domain.UseResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}
	if points <= 0 {
		return nil, fmt.Errorf("使用积分数量必须大于0")
	}
	if points%domain.UsePointsRate != 0 {
		return nil, fmt.Errorf("积分数量必须是%d的倍数", domain.UsePointsRate)
	}

	// 检查用户积分余额
	userPoints, err := uc.repo.GetUserPoints(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户积分失败: %v", err)
	}
	if userPoints.TotalPoints < points {
		return nil, fmt.Errorf("积分余额不足，当前总积分: %d", userPoints.TotalPoints)
	}

	uc.log.WithContext(ctx).Infof("使用积分抵扣: user_id=%d, points=%d, order_id=%s", userID, points, orderID)
	return uc.repo.UsePoints(ctx, userID, points, orderID, description)
}

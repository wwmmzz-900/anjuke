package domain

import (
	"context"
	"time"
)

// PointsRepo 积分数据仓储接口
type PointsRepo interface {
	// GetUserPoints 获取用户积分余额
	GetUserPoints(ctx context.Context, userID uint64) (*UserPoints, error)

	// GetPointsHistory 获取积分历史记录
	GetPointsHistory(ctx context.Context, userID uint64, page, pageSize int32, pointsType string) ([]*PointsRecord, int32, error)

	// HasCheckedInToday 检查用户今日是否已签到
	HasCheckedInToday(ctx context.Context, userID uint64) (bool, error)

	// CheckIn 用户签到
	CheckIn(ctx context.Context, userID uint64) (*CheckInResult, error)

	// EarnPointsByConsume 消费获取积分
	EarnPointsByConsume(ctx context.Context, userID uint64, orderID string, amount int64) (*EarnResult, error)

	// UsePoints 使用积分
	UsePoints(ctx context.Context, userID uint64, points int64, orderID, description string) (*UseResult, error)
}

// UserPoints 用户积分信息
type UserPoints struct {
	UserID      uint64 `json:"user_id"`      // 用户ID
	TotalPoints int64  `json:"total_points"` // 总积分
	Available   int64  `json:"available"`    // 可用积分
	Frozen      int64  `json:"frozen"`       // 冻结积分
	UpdateTime  int64  `json:"update_time"`  // 更新时间戳
}

// PointsRecord 积分记录
type PointsRecord struct {
	ID          uint64    `json:"id"`          // 记录ID
	UserID      uint64    `json:"user_id"`     // 用户ID
	Type        string    `json:"type"`        // 类型: earn(获取), use(使用)
	Points      int64     `json:"points"`      // 积分数量，正数表示增加，负数表示减少
	Balance     int64     `json:"balance"`     // 变更后余额
	Description string    `json:"description"` // 描述
	OrderID     string    `json:"order_id"`    // 关联订单号
	Amount      int64     `json:"amount"`      // 关联金额
	CreateTime  time.Time `json:"create_time"` // 创建时间
	CreatedAt   time.Time `json:"created_at"`  // 创建时间（兼容字段）
}

// CheckInResult 签到结果
type CheckInResult struct {
	IsNewRecord     bool  `json:"is_new_record"`    // 是否是新记录
	Points          int64 `json:"points"`           // 获得积分
	TotalPoints     int64 `json:"total_points"`     // 总积分
	PointsEarned    int64 `json:"points_earned"`    // 本次获得积分
	ConsecutiveDays int32 `json:"consecutive_days"` // 连续签到天数
	Continuous      int   `json:"continuous"`       // 连续签到天数（兼容字段）
}

// EarnResult 获得积分结果
type EarnResult struct {
	Points       int64 `json:"points"`        // 获得积分
	PointsEarned int64 `json:"points_earned"` // 获得积分（兼容字段）
	TotalPoints  int64 `json:"total_points"`  // 总积分
}

// UseResult 使用积分结果
type UseResult struct {
	Points         int64 `json:"points"`          // 使用积分
	PointsUsed     int64 `json:"points_used"`     // 使用积分（兼容字段）
	AmountDeducted int64 `json:"amount_deducted"` // 抵扣金额
	TotalPoints    int64 `json:"total_points"`    // 剩余积分
}

// UsePointsRate 积分使用倍率
const UsePointsRate = 100 // 1元=100积分

// 积分相关常量
const (
	CheckInBasePoints  = 10  // 签到基础积分
	CheckInBonusPoints = 5   // 连续签到奖励积分
	ConsumePointsRate  = 100 // 消费积分倍率
)

// 积分类型常量
const (
	PointsTypeCheckIn = "checkin" // 签到积分
	PointsTypeConsume = "consume" // 消费积分
)

// CheckInRecord 签到记录
type CheckInRecord struct {
	ID              uint64    `json:"id"`
	UserID          uint64    `json:"user_id"`
	CheckInDate     time.Time `json:"checkin_date"`
	Points          int64     `json:"points"`
	ConsecutiveDays int32     `json:"consecutive_days"`
	CreatedAt       time.Time `json:"created_at"`
}

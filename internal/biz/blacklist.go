package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Blacklist 黑名单实体
type Blacklist struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"column:user_id" json:"user_id"`
	Reason    string    `gorm:"column:reason" json:"reason"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (Blacklist) TableName() string {
	return "blacklist"
}

// BlacklistWithUser 黑名单与用户信息的联合查询结果
type BlacklistWithUser struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	UserName  string    `json:"user_name"`
	Phone     string    `json:"phone"`
}

// BlacklistRepo 黑名单仓储接口
type BlacklistRepo interface {
	// 添加用户到黑名单
	AddToBlacklist(ctx context.Context, blacklist *Blacklist) (*Blacklist, error)
	// 从黑名单移除用户
	RemoveFromBlacklist(ctx context.Context, userID int64) error
	// 检查用户是否在黑名单中
	CheckBlacklist(ctx context.Context, userID int64) (*Blacklist, error)
	// 获取黑名单列表（分页）
	GetBlacklistList(ctx context.Context, page, pageSize int32) ([]*BlacklistWithUser, int32, error)
	// 检查用户是否存在
	CheckUserExists(ctx context.Context, userID int64) (bool, error)
}

// BlacklistUsecase 黑名单用例
type BlacklistUsecase struct {
	repo BlacklistRepo
	log  *log.Helper
}

// NewBlacklistUsecase 创建黑名单用例
func NewBlacklistUsecase(repo BlacklistRepo, logger log.Logger) *BlacklistUsecase {
	return &BlacklistUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// AddToBlacklist 添加用户到黑名单
func (uc *BlacklistUsecase) AddToBlacklist(ctx context.Context, userID int64, reason string) (*Blacklist, error) {
	uc.log.WithContext(ctx).Infof("AddToBlacklist: userID=%d, reason=%s", userID, reason)

	// 检查用户是否存在
	exists, err := uc.repo.CheckUserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// 检查用户是否已经在黑名单中
	existing, err := uc.repo.CheckBlacklist(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyBlacklisted
	}

	// 添加到黑名单
	blacklist := &Blacklist{
		UserID:    userID,
		Reason:    reason,
		CreatedAt: time.Now(),
	}

	return uc.repo.AddToBlacklist(ctx, blacklist)
}

// RemoveFromBlacklist 从黑名单移除用户
func (uc *BlacklistUsecase) RemoveFromBlacklist(ctx context.Context, userID int64) error {
	uc.log.WithContext(ctx).Infof("RemoveFromBlacklist: userID=%d", userID)

	// 检查用户是否在黑名单中
	existing, err := uc.repo.CheckBlacklist(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotInBlacklist
	}

	return uc.repo.RemoveFromBlacklist(ctx, userID)
}

// CheckBlacklist 检查用户是否在黑名单中
func (uc *BlacklistUsecase) CheckBlacklist(ctx context.Context, userID int64) (*Blacklist, error) {
	uc.log.WithContext(ctx).Infof("CheckBlacklist: userID=%d", userID)
	return uc.repo.CheckBlacklist(ctx, userID)
}

// GetBlacklistList 获取黑名单列表
func (uc *BlacklistUsecase) GetBlacklistList(ctx context.Context, page, pageSize int32) ([]*BlacklistWithUser, int32, error) {
	uc.log.WithContext(ctx).Infof("GetBlacklistList: page=%d, pageSize=%d", page, pageSize)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	return uc.repo.GetBlacklistList(ctx, page, pageSize)
}

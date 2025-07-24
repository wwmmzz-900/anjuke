package data

import (
	"anjuke/internal/biz"
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type BlacklistRepo struct {
	data *Data
	log  *log.Helper
}

// NewBlacklistRepo 创建黑名单仓储
func NewBlacklistRepo(data *Data, logger log.Logger) biz.BlacklistRepo {
	return &BlacklistRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// AddToBlacklist 添加用户到黑名单
func (r *BlacklistRepo) AddToBlacklist(ctx context.Context, blacklist *biz.Blacklist) (*biz.Blacklist, error) {
	err := r.data.db.WithContext(ctx).Create(blacklist).Error
	if err != nil {
		r.log.WithContext(ctx).Errorf("添加黑名单失败: %v", err)
		return nil, fmt.Errorf("添加黑名单失败: %v", err)
	}
	return blacklist, nil
}

// RemoveFromBlacklist 从黑名单移除用户
func (r *BlacklistRepo) RemoveFromBlacklist(ctx context.Context, userID int64) error {
	err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&biz.Blacklist{}).Error
	if err != nil {
		r.log.WithContext(ctx).Errorf("移除黑名单失败: %v", err)
		return fmt.Errorf("移除黑名单失败: %v", err)
	}
	return nil
}

// CheckBlacklist 检查用户是否在黑名单中
func (r *BlacklistRepo) CheckBlacklist(ctx context.Context, userID int64) (*biz.Blacklist, error) {
	var blacklist biz.Blacklist
	err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).First(&blacklist).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 用户不在黑名单中
		}
		r.log.WithContext(ctx).Errorf("查询黑名单失败: %v", err)
		return nil, fmt.Errorf("查询黑名单失败: %v", err)
	}
	return &blacklist, nil
}

// GetBlacklistList 获取黑名单列表（分页）
func (r *BlacklistRepo) GetBlacklistList(ctx context.Context, page, pageSize int32) ([]*biz.BlacklistWithUser, int32, error) {
	var items []*biz.BlacklistWithUser
	var total int64

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	err := r.data.db.WithContext(ctx).
		Table("blacklist b").
		Joins("LEFT JOIN user_base u ON b.user_id = u.user_id").
		Count(&total).Error
	if err != nil {
		r.log.WithContext(ctx).Errorf("查询黑名单总数失败: %v", err)
		return nil, 0, fmt.Errorf("查询黑名单总数失败: %v", err)
	}

	// 查询列表数据
	err = r.data.db.WithContext(ctx).
		Table("blacklist b").
		Select("b.id, b.user_id, b.reason, b.created_at, u.name as user_name, u.phone").
		Joins("LEFT JOIN user_base u ON b.user_id = u.user_id").
		Order("b.created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Scan(&items).Error

	if err != nil {
		r.log.WithContext(ctx).Errorf("查询黑名单列表失败: %v", err)
		return nil, 0, fmt.Errorf("查询黑名单列表失败: %v", err)
	}

	return items, int32(total), nil
}

// CheckUserExists 检查用户是否存在
func (r *BlacklistRepo) CheckUserExists(ctx context.Context, userID int64) (bool, error) {
	var count int64
	err := r.data.db.WithContext(ctx).
		Table("user_base").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error

	if err != nil {
		r.log.WithContext(ctx).Errorf("检查用户是否存在失败: %v", err)
		return false, fmt.Errorf("检查用户是否存在失败: %v", err)
	}

	return count > 0, nil
}

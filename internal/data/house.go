package data

import (
	"anjuke/internal/biz"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type HouseRepo struct {
	data *Data
	log  *log.Helper
}

func NewHouseRepo(data *Data, logger log.Logger) biz.HouseRepo {
	return &HouseRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// LikeProperty 点赞房源
func (r *HouseRepo) LikeProperty(ctx context.Context, propertyId, userId int64) error {
	// 1. 检查参数
	if propertyId <= 0 || userId <= 0 {
		r.log.Errorf("invalid parameters: propertyId=%d, userId=%d", propertyId, userId)
		return errors.New("invalid parameters")
	}

	// 2. 检查是否已点赞（使用Redis避免重复点赞）
	likedKey := fmt.Sprintf("property:liked:%d:%d", propertyId, userId)
	exists, err := r.data.rdb.Exists(ctx, likedKey).Result()
	if err != nil {
		r.log.Errorf("failed to check like status in redis: %v", err)
		// 继续执行， fallback到数据库
	} else if exists > 0 {
		r.log.Infof("user %d already liked property %d", userId, propertyId)
		return nil // 已经点赞过，无需重复操作
	}

	// 3. 数据库事务
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1 检查是否已存在记录
		var existingLike biz.PropertyLike
		result := tx.Where("property_id = ? AND user_id = ?", propertyId, userId).First(&existingLike)

		// 3.2 如果不存在，则插入新记录
		if result.Error != nil {
			// 只有当错误是记录不存在时，才进行插入
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				like := &biz.PropertyLike{
					PropertyId: propertyId,
					UserId:     userId,
				}
				if err := tx.Create(like).Error; err != nil {
					r.log.Errorf("failed to create like record: %v", err)
					return err
				}
			} else {
				// 其他错误
				r.log.Errorf("failed to check existing like record: %v", result.Error)
				return result.Error
			}
		}

		// 3.3 更新Redis缓存
		// 设置点赞标记，24小时过期
		if err := r.data.rdb.SetEX(ctx, likedKey, 1, 24*time.Hour).Err(); err != nil {
			r.log.Errorf("failed to set like flag in redis: %v", err)
			// 不阻止主流程
		}

		// 3.4 增加点赞数
		countKey := fmt.Sprintf("property:like:count:%d", propertyId)
		if err := r.data.rdb.Incr(ctx, countKey).Err(); err != nil {
			r.log.Errorf("failed to increment like count in redis: %v", err)
			// 不阻止主流程
		}
		return nil
	})
}

// UnlikeProperty 取消点赞房源
func (r *HouseRepo) UnlikeProperty(ctx context.Context, propertyId, userId int64) error {
	// 1. 检查参数
	if propertyId <= 0 || userId <= 0 {
		r.log.Errorf("invalid parameters: propertyId=%d, userId=%d", propertyId, userId)
		return errors.New("invalid parameters")
	}

	// 2. 数据库事务
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 2.1 删除点赞记录
		result := tx.Where("property_id = ? AND user_id = ?", propertyId, userId).Delete(&biz.PropertyLike{})
		if result.Error != nil {
			r.log.Errorf("failed to delete like record: %v", result.Error)
			return result.Error
		}

		// 2.2 更新Redis缓存
		likedKey := fmt.Sprintf("property:liked:%d:%d", propertyId, userId)
		if err := r.data.rdb.Del(ctx, likedKey).Err(); err != nil {
			r.log.Errorf("failed to delete like flag in redis: %v", err)
			// 不阻止主流程
		}

		// 2.3 减少点赞数
		countKey := fmt.Sprintf("property:like:count:%d", propertyId)
		if err := r.data.rdb.Decr(ctx, countKey).Err(); err != nil {
			r.log.Errorf("failed to decrement like count in redis: %v", err)
			// 不阻止主流程
		}

		return nil
	})
}

// IsPropertyLiked 检查用户是否已点赞
func (r *HouseRepo) IsPropertyLiked(ctx context.Context, propertyId, userId int64) (bool, error) {
	// 1. 检查参数
	if propertyId <= 0 || userId <= 0 {
		r.log.Errorf("invalid parameters: propertyId=%d, userId=%d", propertyId, userId)
		return false, errors.New("invalid parameters")
	}

	// 2. 先查Redis
	likedKey := fmt.Sprintf("property:liked:%d:%d", propertyId, userId)
	exists, err := r.data.rdb.Exists(ctx, likedKey).Result()
	if err != nil {
		r.log.Errorf("failed to check like status in redis: %v", err)
		// 继续执行，fallback到数据库
	} else if exists > 0 {
		return true, nil
	}

	// 3. 再查数据库
	var count int64
	result := r.data.db.Debug().WithContext(ctx).Model(&biz.PropertyLike{}).
		Where("property_id = ? AND user_id = ?", propertyId, userId).
		Count(&count)
	if result.Error != nil {
		r.log.Errorf("failed to check like status in database: %v", result.Error)
		return false, result.Error
	}

	// 4. 同步缓存
	liked := count > 0
	if liked {
		if err := r.data.rdb.SetEX(ctx, likedKey, 1, 24*time.Hour).Err(); err != nil {
			r.log.Errorf("failed to set like flag in redis: %v", err)
			// 不阻止主流程
		}
	}

	return liked, nil
}

// GetPropertyLikeCount 获取房源点赞数
func (r *HouseRepo) GetPropertyLikeCount(ctx context.Context, propertyId int64) (int64, error) {
	// 1. 检查参数
	if propertyId <= 0 {
		r.log.Errorf("invalid parameter: propertyId=%d", propertyId)
		return 0, errors.New("invalid parameter")
	}

	// 2. 先查Redis
	countKey := fmt.Sprintf("property:like:count:%d", propertyId)
	countStr, err := r.data.rdb.Get(ctx, countKey).Result()
	if err == nil {
		// 解析字符串为int64
		var count int64
		fmt.Sscanf(countStr, "%d", &count)
		return count, nil
	} else if err != redis.Nil {
		r.log.Errorf("failed to get like count from redis: %v", err)
		// 继续执行，fallback到数据库
	}

	// 3. 再查数据库
	var count int64
	result := r.data.db.WithContext(ctx).Model(&biz.PropertyLike{}).
		Where("property_id = ?", propertyId).
		Count(&count)
	if result.Error != nil {
		r.log.Errorf("failed to get like count from database: %v", result.Error)
		return 0, result.Error
	}

	// 4. 同步缓存
	if err := r.data.rdb.SetEX(ctx, countKey, count, 24*time.Hour).Err(); err != nil {
		r.log.Errorf("failed to set like count in redis: %v", err)
		// 不阻止主流程
	}

	return count, nil
}

// GetUserLikeList 获取用户点赞列表
func (r *HouseRepo) GetUserLikeList(ctx context.Context, userId int64, page, pageSize int) ([]*biz.PropertyLike, int64, error) {
	// 1. 检查参数
	if userId <= 0 {
		r.log.Errorf("invalid parameter: userId=%d", userId)
		return nil, 0, errors.New("invalid parameter")
	}

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 2. 计算偏移量
	offset := (page - 1) * pageSize

	// 3. 查询总数
	var total int64
	result := r.data.db.WithContext(ctx).Model(&biz.PropertyLike{}).
		Where("user_id = ?", userId).
		Count(&total)
	if result.Error != nil {
		r.log.Errorf("failed to get like list total: %v", result.Error)
		return nil, 0, result.Error
	}

	// 4. 查询列表
	var likeList []*biz.PropertyLike
	result = r.data.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&likeList)
	if result.Error != nil {
		r.log.Errorf("failed to get like list: %v", result.Error)
		return nil, 0, result.Error
	}

	return likeList, total, nil
}

// GetHouseDetail 获取房源详情
func (r *HouseRepo) GetHouseDetail(ctx context.Context, houseId int64) (*biz.House, error) {
	// 1. 检查参数
	if houseId <= 0 {
		r.log.Errorf("invalid parameter: houseId=%d", houseId)
		return nil, errors.New("invalid parameter")
	}

	// 2. 查询数据库
	var house biz.House
	result := r.data.db.WithContext(ctx).Where("house_id = ?", houseId).Limit(1).Find(&house)
	if result.Error != nil {
		r.log.Errorf("failed to get house detail: %v", result.Error)
		return nil, result.Error
	}

	return &house, nil
}

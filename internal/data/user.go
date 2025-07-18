package data

import (
	"anjuke/internal/biz"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"time"
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
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		return nil, nil // 明确返回nil表示用户不存在
//	}
//	if err != nil {
//		return nil, fmt.Errorf("查询用户失败: %v", err)
//	}
//	return &user, nil
//}

func (r *UserRepo) BindPhone(ctx context.Context, uid int64, phone string, binding *biz.UserBinding) error {
	// 开启事务
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新用户表手机号
		if err := tx.Model(&biz.UserBase{}).Where("user_id = ?", uid).Update("phone", phone).Error; err != nil {
			return fmt.Errorf("更新用户手机号失败: %w", err)
		}

		// 2. 创建绑定记录
		if err := tx.Create(&binding).Error; err != nil {
			return fmt.Errorf("创建绑定记录失败: %w", err)
		}

		return nil
	})
}

func (r *UserRepo) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {

	var count int64
	err := r.data.db.WithContext(ctx).Model(&biz.UserBase{}).
		Where("phone = ?", phone).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func (r *UserRepo) Store(ctx context.Context, Soures, Phone, code string, expire time.Duration) error {
	key := fmt.Sprintf("verify:%s:%s", Soures, Phone)
	return r.data.rdb.Set(ctx, key, code, expire).Err()
}

func (r *UserRepo) Get(ctx context.Context, Soures, Phone string) (string, error) {
	key := fmt.Sprintf("verify:%s:%s", Soures, Phone)
	result, err := r.data.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("获取验证码失败2" + err.Error())
	}
	fmt.Println(result)
	return result, nil
}

func (r *UserRepo) Delete(ctx context.Context, Soures, Phone string) error {
	key := fmt.Sprintf("verify:%s:%s", Soures, Phone)
	return r.data.rdb.Del(ctx, key).Err()
}

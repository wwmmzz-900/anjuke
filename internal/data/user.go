package data

import (
	"anjuke/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
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
//
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		return nil, nil // 明确返回nil表示用户不存在
//	}
//	if err != nil {
//		return nil, fmt.Errorf("查询用户失败: %v", err)
//	}
//	return &user, nil
//}

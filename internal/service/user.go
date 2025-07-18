package service

import (
	v2 "anjuke/api/user/v2"
	"anjuke/internal/biz"
)

type UserService struct {
	v2.UnimplementedUserServer
	v2uc *biz.UserUsecase
}

func NewUserService(v2uc *biz.UserUsecase) *UserService {
	return &UserService{
		v2uc: v2uc,
	}
}

// todo:用户登录注册一体化
//func (s *UserService) CreateUser(ctx context.Context, req *v2.CreateUserRequest) (*v2.CreateUserReply, error) {
//	user, err := s.v2uc.GetUser(ctx, req.Mobile)
//	if err != nil {
//		return nil, fmt.Errorf("查询失败: %v", err)
//	}
//
//	// 用户不存在时才创建
//	if user == nil || user.Mobile == "" {
//		_, err := s.v2uc.CreateUser(ctx, &biz.User{
//			Mobile:   req.Mobile,
//			NickName: req.NickName,
//			Password: req.Password, // 注意：密码应该加密
//			Birthday: 0,            // 设置默认值
//			Gender:   0,            // 设置默认值
//			Grade:    0,            // 设置默认值
//		})
//		if err != nil {
//			return nil, fmt.Errorf("创建用户失败: %v", err)
//		}
//		return &v2.CreateUserReply{
//			Success: "注册成功",
//		}, nil
//	}
//
//	// 用户已存在，检查密码
//	if user.Password != req.Password { // 注意：实际应该对比加密后的密码
//		return nil, fmt.Errorf("密码错误")
//	}
//	return &v2.CreateUserReply{
//		Success: "登录成功",
//	}, nil
//}

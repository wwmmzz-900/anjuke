package service

import (
	v2 "anjuke/api/user/v2"
	"anjuke/internal/biz"
	"anjuke/internal/pkg/jwt"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type UserService struct {
	v2.UnimplementedUserServer
	v2uc *biz.UserUsecase
	jwt  *jwt.JWT
}

func NewUserService(v2uc *biz.UserUsecase, jwtSecret string) *UserService {
	return &UserService{
		v2uc: v2uc,
		jwt:  jwt.NewJwt(jwtSecret),
	}
}

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

// todo:获取验证码
func (s *UserService) SendSms(ctx context.Context, req *v2.SendSmsRequest) (*v2.SendSmsReply, error) {
	// 1. 生成随机验证码
	code := generateCode(6) // 6位数字验证码
	// 2. 存储验证码 (5分钟有效期)
	if err := s.v2uc.Store(ctx, req.Soures, req.Phone, code, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("验证码存储失败")
	}
	return &v2.SendSmsReply{
		Sms: "验证码存储成功",
	}, nil
}

// todo:生成随机数字验证码
func generateCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	letters := "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// todo：用户手机号绑定
func (s *UserService) BindPhone(ctx context.Context, req *v2.BindPhoneRequest) (*v2.BindPhoneReply, error) {
	// 验证验证码
	get, err := s.v2uc.Get(ctx, "phone", req.Value)
	if err != nil {
		return nil, fmt.Errorf("%s", "获取验证码错误"+err.Error())
	}
	if get != req.Code {
		return nil, fmt.Errorf("验证码错误")
	}
	err = s.v2uc.BindPhone(ctx, req.UserId, req.Value, &biz.UserBinding{
		UserId: req.UserId,
		Type:   req.Type,
		Value:  req.Value,
		Extra:  req.Extra,
	})
	if err != nil {
		return nil, fmt.Errorf("%s", "用户绑定失败"+err.Error())
	}
	_ = s.v2uc.Delete(ctx, "phone", req.Value)
	return &v2.BindPhoneReply{
		Success: "用户绑定成功",
		Userid:  req.UserId,
	}, nil
}

// todo：用户邮箱绑定
func (s *UserService) BindEmail(ctx context.Context, req *v2.BindEmailRequest) (*v2.BindEmailReply, error) {
	// 验证验证码
	get, err := s.v2uc.Get(ctx, "email", req.Value)
	if err != nil {
		return nil, fmt.Errorf("%s", "获取验证码错误"+err.Error())
	}
	fmt.Printf("email:%s", get)
	fmt.Printf("emailreq:%s", req.Code)
	if get != req.Code {
		return nil, fmt.Errorf("验证码错误")
	}
	err = s.v2uc.BindEmail(ctx, req.UserId, req.Value, &biz.UserBinding{
		UserId: req.UserId,
		Type:   req.Type,
		Value:  req.Value,
		Extra:  req.Extra,
	})
	if err != nil {
		return nil, fmt.Errorf("%s", "用户绑定失败"+err.Error())
	}
	_ = s.v2uc.Delete(ctx, "email", req.Value)
	return &v2.BindEmailReply{
		Success: "用户绑定邮箱成功",
		UserId:  req.UserId,
	}, nil
}

// LoginOrRegister 用户登录或注册
func (s *UserService) LoginOrRegister(ctx context.Context, req *v2.LoginOrRegisterRequest) (*v2.LoginOrRegisterReply, error) {
	// 调用业务逻辑层处理登录或注册
	user, isNewUser, err := s.v2uc.LoginOrRegister(ctx, req.Account, req.Password, req.Name, req.Sex)
	if err != nil {
		return nil, fmt.Errorf("登录或注册失败: %v", err)
	}

	// 生成JWT token
	token, err := s.jwt.GenerateToken(uint(user.UserId), user.Name, 24*time.Hour) // 24小时有效期
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %v", err)
	}

	message := "登录成功"
	if isNewUser {
		message = "注册成功"
	}

	return &v2.LoginOrRegisterReply{
		Message:   message,
		Token:     token,
		IsNewUser: isNewUser,
	}, nil
}

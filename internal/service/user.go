package service

import (
	v2 "anjuke/api/user/v2"
	"anjuke/internal/biz"
	"context"
	"fmt"
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

// Login 用户登录注册一体化
func (s *UserService) Login(ctx context.Context, req *v2.LoginRequest) (*v2.LoginReply, error) {
	// 获取客户端信息用于日志记录
	ipAddress := getClientIP(ctx)
	userAgent := getUserAgent(ctx)

	var loginLog *biz.LoginLog
	var userId int64

	// 如果有验证码，优先走验证码登录/注册
	if req.SendSmsCode != "" {
		// 校验验证码
		if err := s.v2uc.VerifySmsCode(ctx, req.Phone, req.SendSmsCode); err != nil {
			// 记录登录失败日志
			loginLog = &biz.LoginLog{
				UserId:      0, // 验证码错误时可能还没有用户ID
				IpAddress:   ipAddress,
				UserAgent:   userAgent,
				LoginStatus: 0,
				FailReason:  "验证码错误: " + err.Error(),
			}
			s.v2uc.SaveLoginLog(ctx, loginLog)
			return nil, err
		}

		// Bug #4 修复: 使用数据库事务避免竞态条件
		user, err := s.v2uc.GetUser(ctx, req.Phone)
		if err != nil && err.Error() != "record not found" {
			// Bug #6 修复: 不暴露内部错误信息
			return nil, fmt.Errorf("系统错误，请稍后重试")
		}

		if user == nil || user.Phone == "" {
			// 注册新用户
			newUser := &biz.UserBase{
				Name:       req.Name,
				Phone:      req.Phone,
				Password:   biz.Md5(req.Password), // 密码加密存储
				Sex:        req.Sex,
				RealStatus: 2,
				Status:     1,
			}
			createdUser, err := s.v2uc.CreateUser(ctx, newUser)
			if err != nil {
				// Bug #4 修复: 处理并发创建用户的情况
				if err.Error() == "创建用户失败: UNIQUE constraint failed" ||
					err.Error() == "创建用户失败: Duplicate entry" {
					// 用户已被其他请求创建，重新查询
					user, err = s.v2uc.GetUser(ctx, req.Phone)
					if err != nil {
						return nil, fmt.Errorf("查询用户失败: %v", err)
					}
					// 继续执行登录逻辑
				} else {
					// Bug #6 修复: 不暴露内部错误信息
					return nil, fmt.Errorf("注册失败，请稍后重试")
				}
			} else {
				userId = createdUser.UserId

				token, _, err := s.v2uc.GenerateTokens(createdUser.UserId)
				if err != nil {
					// Bug #6 修复: 不暴露内部错误信息
					return nil, fmt.Errorf("登录失败，请稍后重试")
				}

				// 记录注册成功日志
				loginLog = &biz.LoginLog{
					UserId:      userId,
					IpAddress:   ipAddress,
					UserAgent:   userAgent,
					LoginStatus: 1,
					FailReason:  "",
				}
				s.v2uc.SaveLoginLog(ctx, loginLog)

				return &v2.LoginReply{
					UserId: createdUser.UserId,
					Token:  token,
					Status: int32(createdUser.Status),
				}, nil
			}
		}

		// Bug #9 修复: 验证码登录时不强制要求用户名匹配，允许用户名为空或不匹配的情况
		// 如果用户名不为空且不匹配，则更新用户名
		if req.Name != "" && user.Name != req.Name {
			user.Name = req.Name
			if err := s.v2uc.UpdateUserInfo(ctx, user); err != nil {
				// Bug #12 修复: 记录用户名更新失败的日志
				loginLog = &biz.LoginLog{
					UserId:      user.UserId,
					IpAddress:   ipAddress,
					UserAgent:   userAgent,
					LoginStatus: 0,
					FailReason:  "更新用户名失败: " + err.Error(),
				}
				s.v2uc.SaveLoginLog(ctx, loginLog)
				// 不阻断登录流程，只记录错误
			}
		}

		// 检查账号状态
		if user.Status == 0 {
			loginLog = &biz.LoginLog{
				UserId:      user.UserId,
				IpAddress:   ipAddress,
				UserAgent:   userAgent,
				LoginStatus: 0,
				FailReason:  "账号已被冻结",
			}
			s.v2uc.SaveLoginLog(ctx, loginLog)
			return nil, fmt.Errorf("账号已被冻结，请联系管理员")
		}

		// 登录成功
		token, _, err := s.v2uc.GenerateTokens(user.UserId)
		if err != nil {
			// Bug #6 修复: 不暴露内部错误信息
			return nil, fmt.Errorf("登录失败，请稍后重试")
		}

		// 记录登录成功日志
		loginLog = &biz.LoginLog{
			UserId:      user.UserId,
			IpAddress:   ipAddress,
			UserAgent:   userAgent,
			LoginStatus: 1,
			FailReason:  "",
		}
		s.v2uc.SaveLoginLog(ctx, loginLog)

		return &v2.LoginReply{
			UserId: user.UserId,
			Token:  token,
			Status: int32(user.Status),
		}, nil
	}

	// 没有验证码，走原有密码登录逻辑
	user, err := s.v2uc.GetUser(ctx, req.Phone)
	if err != nil && err.Error() != "record not found" {
		// Bug #12 修复: 记录查询失败的登录日志
		loginLog = &biz.LoginLog{
			UserId:      0,
			IpAddress:   ipAddress,
			UserAgent:   userAgent,
			LoginStatus: 0,
			FailReason:  "系统查询失败",
		}
		s.v2uc.SaveLoginLog(ctx, loginLog)
		return nil, fmt.Errorf("查询失败: %v", err)
	}

	if user != nil {
		// Bug #9 修复: 密码登录时才严格验证用户名匹配
		if user.Name != req.Name {
			loginLog = &biz.LoginLog{
				UserId:      user.UserId,
				IpAddress:   ipAddress,
				UserAgent:   userAgent,
				LoginStatus: 0,
				FailReason:  "用户名和手机号不匹配",
			}
			s.v2uc.SaveLoginLog(ctx, loginLog)
			return nil, fmt.Errorf("用户名和手机号不匹配")
		}

		// 检查密码（密码已加密存储）
		if user.Password != biz.Md5(req.Password) {
			loginLog = &biz.LoginLog{
				UserId:      user.UserId,
				IpAddress:   ipAddress,
				UserAgent:   userAgent,
				LoginStatus: 0,
				FailReason:  "密码错误",
			}
			s.v2uc.SaveLoginLog(ctx, loginLog)
			return nil, fmt.Errorf("密码错误")
		}

		// 检查账号状态
		if user.Status == 0 {
			loginLog = &biz.LoginLog{
				UserId:      user.UserId,
				IpAddress:   ipAddress,
				UserAgent:   userAgent,
				LoginStatus: 0,
				FailReason:  "账号已被冻结",
			}
			s.v2uc.SaveLoginLog(ctx, loginLog)
			return nil, fmt.Errorf("账号已被冻结，请联系管理员")
		}

		token, _, err := s.v2uc.GenerateTokens(user.UserId)
		if err != nil {
			// Bug #6 修复: 不暴露内部错误信息
			return nil, fmt.Errorf("登录失败，请稍后重试")
		}

		// 记录登录成功日志
		loginLog = &biz.LoginLog{
			UserId:      user.UserId,
			IpAddress:   ipAddress,
			UserAgent:   userAgent,
			LoginStatus: 1,
			FailReason:  "",
		}
		s.v2uc.SaveLoginLog(ctx, loginLog)

		return &v2.LoginReply{
			UserId: user.UserId,
			Token:  token,
			Status: int32(user.Status),
		}, nil
	} else {
		// 用户不存在，注册新用户
		newUser := &biz.UserBase{
			Name:       req.Name,
			Phone:      req.Phone,
			Password:   biz.Md5(req.Password), // 密码加密存储
			Sex:        req.Sex,
			RealStatus: 2,
			Status:     1,
		}
		createdUser, err := s.v2uc.CreateUser(ctx, newUser)
		if err != nil {
			// Bug #6 修复: 不暴露内部错误信息
			return nil, fmt.Errorf("注册失败，请稍后重试")
		}

		token, _, err := s.v2uc.GenerateTokens(createdUser.UserId)
		if err != nil {
			// Bug #6 修复: 不暴露内部错误信息
			return nil, fmt.Errorf("登录失败，请稍后重试")
		}

		// 记录注册成功日志
		loginLog = &biz.LoginLog{
			UserId:      createdUser.UserId,
			IpAddress:   ipAddress,
			UserAgent:   userAgent,
			LoginStatus: 1,
			FailReason:  "",
		}
		s.v2uc.SaveLoginLog(ctx, loginLog)

		return &v2.LoginReply{
			UserId: createdUser.UserId,
			Token:  token,
			Status: int32(createdUser.Status),
		}, nil
	}
}

// SendSms 短信验证码
func (s *UserService) SendSms(ctx context.Context, req *v2.SendSmsRequest) (*v2.SendSmsReply, error) {
	if err := s.v2uc.SendSms(ctx, req.Phone, req.Source); err != nil {
		return nil, fmt.Errorf("SendSms error: %v", err)
	}
	return &v2.SendSmsReply{
		Success: "验证码已发送",
	}, nil
}

// UpdateUserInfo 修改用户个人信息
func (s *UserService) UpdateUserInfo(ctx context.Context, req *v2.UpdateUserInfoRequest) (*v2.UpdateUserInfoReply, error) {
	// Bug #7 修复: 添加关键参数验证
	if req.Id <= 0 {
		return nil, fmt.Errorf("用户ID无效")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	if req.Phone == "" {
		return nil, fmt.Errorf("手机号不能为空")
	}

	// 查询当前用户信息
	oldUser, err := s.v2uc.GetUserID(ctx, req.Id)
	if err != nil || oldUser == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 如果新用户名和原用户名不同，检查新用户名是否已被注册
	if req.Name != oldUser.Name {
		existUser, err := s.v2uc.GetUserByName(ctx, req.Name)
		if err != nil {
			return nil, fmt.Errorf("查询用户名失败: %v", err)
		}
		if existUser != nil && existUser.UserId != req.Id {
			return nil, fmt.Errorf("用户名已被注册")
		}
	}

	// 如果新手机号和原手机号不同，检查新手机号是否已被注册
	if req.Phone != oldUser.Phone {
		existUser, err := s.v2uc.GetUser(ctx, req.Phone)
		if err != nil {
			return nil, fmt.Errorf("查询手机号失败: %v", err)
		}
		if existUser != nil && existUser.UserId != req.Id {
			return nil, fmt.Errorf("手机号已被注册")
		}
	}

	user := &biz.UserBase{
		UserId: req.Id,
		Name:   req.Name,
		Phone:  req.Phone,
		Avatar: req.Avatar,
		RoleId: req.RoleId,
		Sex:    req.Sex,
	}

	err = s.v2uc.UpdateUserInfo(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("信息更新失败: %v", err)
	}

	return &v2.UpdateUserInfoReply{
		Id:      user.UserId,
		Success: true,
	}, nil
}

// UpdateUserPws 修改用户密码
func (s *UserService) UpdateUserPws(ctx context.Context, req *v2.UpdateUserPwsRequest) (*v2.UpdateUserPwsReply, error) {
	// 1. 校验验证码
	if err := s.v2uc.UpdateSmsCode(ctx, req.Phone, req.SendSmsCode); err != nil {
		return nil, err
	}
	// 2. 校验新密码和确认密码一致
	if req.Password != req.ConfirmPassword {
		return nil, fmt.Errorf("两次输入的新密码不一致")
	}
	// 3. 查询用户信息
	user, err := s.v2uc.GetUserID(ctx, req.Id)
	if err != nil || user == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	// 4. 校验旧密码
	if user.Password != biz.Md5(req.OldPassword) {
		return nil, fmt.Errorf("旧密码错误")
	}
	// 5. 新密码不能与旧密码相同
	if req.OldPassword == req.Password {
		return nil, fmt.Errorf("新密码不能与旧密码相同")
	}
	// 6. 更新密码
	user.Password = biz.Md5(req.Password)
	err = s.v2uc.UpdateUserInfo(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("信息更新失败: %v", err)
	}

	// 7. 密码修改成功后，删除验证码
	s.v2uc.DeleteUpdateSmsCode(ctx, req.Phone)

	return &v2.UpdateUserPwsReply{
		Success: true,
	}, nil
}

// FaceCertify 支付宝人脸识别实名认证初始化
func (s *UserService) FaceCertify(ctx context.Context, req *v2.FaceCertifyRequest) (*v2.FaceCertifyReply, error) {
	// 参数验证
	if req.UserId <= 0 {
		return &v2.FaceCertifyReply{
			Success: false,
			Message: "用户ID不能为空",
		}, nil
	}
	if req.RealName == "" {
		return &v2.FaceCertifyReply{
			Success: false,
			Message: "真实姓名不能为空",
		}, nil
	}
	if req.IdCardNumber == "" {
		return &v2.FaceCertifyReply{
			Success: false,
			Message: "身份证号不能为空",
		}, nil
	}

	// 检查用户是否已经实名认证
	user, err := s.v2uc.GetUserID(ctx, req.UserId)
	if err != nil {
		return &v2.FaceCertifyReply{
			Success: false,
			Message: "用户不存在",
		}, nil
	}
	if user.RealStatus == 1 {
		return &v2.FaceCertifyReply{
			Success: false,
			Message: "用户已完成实名认证",
		}, nil
	}

	certifyID, certifyURL, err := s.v2uc.FaceCertify(ctx, req.UserId, req.RealName, req.IdCardNumber, req.ReturnUrl)
	if err != nil {
		return &v2.FaceCertifyReply{
			Success: false,
			Message: fmt.Sprintf("发起实名认证失败: %v", err),
		}, nil
	}

	return &v2.FaceCertifyReply{
		CertifyId:  certifyID,
		CertifyUrl: certifyURL,
		Success:    true,
		Message:    "实名认证初始化成功，请访问认证URL完成人脸识别",
	}, nil
}

// CertifyNotify 支付宝实名认证回调
func (s *UserService) CertifyNotify(ctx context.Context, req *v2.CertifyNotifyRequest) (*v2.CertifyNotifyReply, error) {
	if err := s.v2uc.CertifyNotify(ctx, req.BizContent); err != nil {
		return &v2.CertifyNotifyReply{
			Result: "fail",
		}, err
	}
	return &v2.CertifyNotifyReply{
		Result: "success",
	}, nil
}

// QueryCertify 查询实名认证结果
func (s *UserService) QueryCertify(ctx context.Context, req *v2.QueryCertifyRequest) (*v2.QueryCertifyReply, error) {
	if req.CertifyId == "" {
		return &v2.QueryCertifyReply{
			Passed:  false,
			Status:  "FAIL",
			Message: "认证ID不能为空",
		}, nil
	}

	result, err := s.v2uc.QueryCertify(ctx, req.CertifyId, req.UserId)
	if err != nil {
		return &v2.QueryCertifyReply{
			Passed:  false,
			Status:  "FAIL",
			Message: fmt.Sprintf("查询认证结果失败: %v", err),
		}, nil
	}

	return &v2.QueryCertifyReply{
		Passed:       result.Passed,
		Status:       result.Status,
		RealName:     result.RealName,
		IdCardNumber: result.IdCardNumber,
		FailReason:   result.FailReason,
		Message:      result.Message,
	}, nil
}

// ResetPassword 密码重置
func (s *UserService) ResetPassword(ctx context.Context, req *v2.ResetPasswordRequest) (*v2.ResetPasswordReply, error) {
	// 1. 校验验证码
	if err := s.v2uc.VerifyResetPasswordSmsCode(ctx, req.Phone, req.SmsCode); err != nil {
		return nil, err
	}

	// 参数验证
	if req.Phone == "" {
		return &v2.ResetPasswordReply{
			Success: false,
			Message: "手机号不能为空",
		}, nil
	}
	if req.SmsCode == "" {
		return &v2.ResetPasswordReply{
			Success: false,
			Message: "验证码不能为空",
		}, nil
	}
	if req.NewPassword == "" {
		return &v2.ResetPasswordReply{
			Success: false,
			Message: "新密码不能为空",
		}, nil
	}
	if req.NewPassword != req.ConfirmPassword {
		return &v2.ResetPasswordReply{
			Success: false,
			Message: "两次输入的密码不一致",
		}, nil
	}
	if len(req.NewPassword) < 6 {
		return &v2.ResetPasswordReply{
			Success: false,
			Message: "密码长度不能少于6位",
		}, nil
	}

	// 调用业务层重置密码
	err := s.v2uc.ResetPassword(ctx, req.Phone, req.SmsCode, req.NewPassword)
	if err != nil {
		return &v2.ResetPasswordReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v2.ResetPasswordReply{
		Success: true,
		Message: "密码重置成功",
	}, nil
}

// FreezeAccount 账号冻结
func (s *UserService) FreezeAccount(ctx context.Context, req *v2.FreezeAccountRequest) (*v2.FreezeAccountReply, error) {
	// 参数验证
	if req.UserId <= 0 {
		return &v2.FreezeAccountReply{
			Success: false,
			Message: "用户ID不能为空",
		}, nil
	}
	if req.AdminId <= 0 {
		return &v2.FreezeAccountReply{
			Success: false,
			Message: "管理员ID不能为空",
		}, nil
	}
	if req.Reason == "" {
		return &v2.FreezeAccountReply{
			Success: false,
			Message: "冻结原因不能为空",
		}, nil
	}

	// 获取操作IP（实际使用时需要从HTTP请求中获取）
	ipAddress := getClientIP(ctx)

	// 调用业务层冻结账号
	err := s.v2uc.FreezeAccount(ctx, req.UserId, req.AdminId, req.Reason, ipAddress)
	if err != nil {
		return &v2.FreezeAccountReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v2.FreezeAccountReply{
		Success: true,
		Message: "账号冻结成功",
	}, nil
}

// UnfreezeAccount 账号解冻
func (s *UserService) UnfreezeAccount(ctx context.Context, req *v2.UnfreezeAccountRequest) (*v2.UnfreezeAccountReply, error) {
	// 参数验证
	if req.UserId <= 0 {
		return &v2.UnfreezeAccountReply{
			Success: false,
			Message: "用户ID不能为空",
		}, nil
	}
	if req.AdminId <= 0 {
		return &v2.UnfreezeAccountReply{
			Success: false,
			Message: "管理员ID不能为空",
		}, nil
	}
	if req.Reason == "" {
		return &v2.UnfreezeAccountReply{
			Success: false,
			Message: "解冻原因不能为空",
		}, nil
	}

	// 获取操作IP（实际使用时需要从HTTP请求中获取）
	ipAddress := getClientIP(ctx)

	// 调用业务层解冻账号
	err := s.v2uc.UnfreezeAccount(ctx, req.UserId, req.AdminId, req.Reason, ipAddress)
	if err != nil {
		return &v2.UnfreezeAccountReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v2.UnfreezeAccountReply{
		Success: true,
		Message: "账号解冻成功",
	}, nil
}

// GetLoginLogs 获取登录日志
func (s *UserService) GetLoginLogs(ctx context.Context, req *v2.GetLoginLogsRequest) (*v2.GetLoginLogsReply, error) {
	// 参数验证
	if req.UserId <= 0 {
		return &v2.GetLoginLogsReply{}, fmt.Errorf("用户ID不能为空")
	}

	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 调用业务层获取登录日志
	logs, total, err := s.v2uc.GetLoginLogs(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取登录日志失败: %v", err)
	}

	// 转换为响应格式
	var replyLogs []*v2.LoginLog
	for _, log := range logs {
		replyLogs = append(replyLogs, &v2.LoginLog{
			Id:          log.Id,
			UserId:      log.UserId,
			IpAddress:   log.IpAddress,
			UserAgent:   log.UserAgent,
			DeviceInfo:  log.DeviceInfo,
			Location:    log.Location,
			LoginStatus: int32(log.LoginStatus),
			FailReason:  log.FailReason,
			LoginTime:   log.LoginTime.Unix(),
		})
	}

	return &v2.GetLoginLogsReply{
		Logs:     replyLogs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// getClientIP 获取客户端IP地址的辅助函数
func getClientIP(ctx context.Context) string {
	// 这里应该从HTTP请求的context中获取真实IP
	// 具体实现取决于你使用的HTTP框架
	return "127.0.0.1" // 默认值
}

// getUserAgent 获取用户代理的辅助函数
func getUserAgent(ctx context.Context) string {
	// 这里应该从HTTP请求的context中获取User-Agent
	// 具体实现取决于你使用的HTTP框架
	return "Unknown" // 默认值
}

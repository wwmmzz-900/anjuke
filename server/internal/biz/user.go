// Package biz 封装了应用的核心业务逻辑（Use Cases）。
// 这一层负责编排 domain 层的实体和接口，完成具体的业务功能。
// biz 层依赖 domain 层的接口，但不关心其具体实现。
package biz

import (
	"context"
	"fmt"
	"io"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
)

// UserUsecase 封装了用户相关的业务逻辑。
// 它依赖 domain.UserRepo 和 domain.SmsRepo 接口来完成数据持久化操作。
type UserUsecase struct {
	repo      domain.UserRepo
	smsRepo   domain.SmsRepo
	minioRepo domain.MinioRepo //
	log       *log.Helper
}

// NewUserUsecase .
func NewUserUsecase(repo domain.UserRepo, smsRepo domain.SmsRepo, minioRepo domain.MinioRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{repo: repo, smsRepo: smsRepo, minioRepo: minioRepo, log: log.NewHelper(logger)}
}

// SendSms 是发送短信的业务流程。
// 它通过编排风控检查（由 UserRepo 提供）和短信发送（由 SmsRepo 提供）来完成。
// 这种方式使得业务逻辑清晰，且易于测试和维护。
func (uc *UserUsecase) SendSms(ctx context.Context, phone, deviceID, ip, scene string) (string, error) {
	// 1. 场景验证
	if !uc.isValidScene(scene) {
		return "", fmt.Errorf("不支持的短信场景: %s", scene)
	}

	// 2. 调用 UserRepo 进行风控检查
	err := uc.repo.SmsRiskControl(ctx, phone)
	if err != nil {
		return "", err
	}

	// 3. 风控通过后，调用 SmsRepo 执行发送
	uc.log.WithContext(ctx).Infof("SendSms to: %s, scene: %s", phone, scene)
	err = uc.smsRepo.SendSms(ctx, phone, scene)
	if err != nil {
		return "", err
	}
	return "发送成功", nil
}

// VerifySms 验证短信验证码
func (uc *UserUsecase) VerifySms(ctx context.Context, phone, code, scene string) (bool, error) {
	// 1. 场景验证
	if !uc.isValidScene(scene) {
		return false, fmt.Errorf("不支持的短信场景: %s", scene)
	}

	// 2. 参数验证
	if phone == "" || code == "" {
		return false, fmt.Errorf("参数不能为空")
	}

	// 3. 调用 SmsRepo 进行验证
	uc.log.WithContext(ctx).Infof("VerifySms: phone=%s, scene=%s", phone, scene)
	return uc.smsRepo.VerifySms(ctx, phone, code)
}

// isValidScene 检查短信场景是否有效
func (uc *UserUsecase) isValidScene(scene string) bool {
	validScenes := map[string]bool{
		"register":       true, // 注册
		"login":          true, // 登录
		"reset_password": true, // 重置密码
		"bind_phone":     true, // 绑定手机号
		"change_phone":   true, // 更换手机号
		"real_name":      true, // 实名认证
	}
	return validScenes[scene]
}

// RealName 处理用户实名认证的业务逻辑。
func (uc *UserUsecase) RealName(ctx context.Context, user *domain.RealName) (*domain.RealName, error) {

	// 参数校验
	if user.UserId == 0 {
		return nil, fmt.Errorf("用户id不能为空")
	}
	if user.Name == "" {
		return nil, fmt.Errorf("姓名不能为空")
	}
	if len(user.Name) < 2 || len(user.Name) > 10 {
		return nil, fmt.Errorf("姓名长度必须在2-10个字符之间")
	}
	if user.IdCard == "" {
		return nil, fmt.Errorf("身份证号不能为空")
	}
	if len(user.IdCard) != 18 {
		return nil, fmt.Errorf("身份证号必须是18位")
	}

	// 调用 repo 完成数据持久化
	uc.log.WithContext(ctx).Infof("RealName: %s, UserID: %d", user.Name, user.UserId)
	return uc.repo.RealName(ctx, user)
}

// UpdateUserStatus 处理更新用户状态的业务逻辑。
func (uc *UserUsecase) UpdateUserStatus(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
	uc.log.WithContext(ctx).Infof("UpdateUserStatus: user_id=%d, real_status=%d", user.UserId, user.RealStatus)
	return uc.repo.UpdateUserStatus(ctx, user)
}

func (uc *UserUsecase) UploadToMinioWithProgress(ctx context.Context, fileName string, reader io.Reader, size int64, contentType string, progressCallback func(uploaded, total int64)) (string, error) {
	if repo, ok := uc.minioRepo.(interface {
		SmartUploadWithProgress(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string, progressCallback func(uploaded, total int64)) (string, error)
	}); ok {
		return repo.SmartUploadWithProgress(ctx, fileName, reader, size, contentType, progressCallback)
	}
	// fallback: no progress
	fileInfo, err := uc.minioRepo.SimpleUpload(ctx, fileName, reader)
	if err != nil {
		return "", err
	}
	return fileInfo.FileURL, nil
}

func (uc *UserUsecase) DeleteFromMinio(ctx context.Context, objectName string) error {
	return uc.minioRepo.Delete(ctx, objectName)
}

// CreateUser 创建新用户
func (uc *UserUsecase) CreateUser(ctx context.Context, phone, name, password string) (*domain.UserBase, error) {
	// 参数校验
	if phone == "" || name == "" || password == "" {
		return nil, fmt.Errorf("手机号、昵称和密码不能为空")
	}

	// 检查手机号是否已存在
	exists, err := uc.repo.CheckPhoneExists(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("检查手机号失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("该手机号已注册")
	}

	// 创建用户
	user := &domain.UserBase{
		Phone:      phone,
		Name:       name,
		Password:   password,
		RoleID:     1, // 默认角色
		RealStatus: domain.RealNameUnverified,
		Status:     domain.UserStatusNormal,
	}

	// 调用仓储层创建用户
	uc.log.WithContext(ctx).Infof("创建用户: phone=%s, name=%s", phone, name)
	return uc.repo.CreateUser(ctx, user)
}

// Login 用户登录
func (uc *UserUsecase) Login(ctx context.Context, loginType, mobile, password, code string) (*domain.UserBase, string, error) {
	// 参数校验
	if mobile == "" {
		return nil, "", fmt.Errorf("手机号不能为空")
	}

	var user *domain.UserBase
	var err error

	switch loginType {
	case "password":
		// 密码登录
		if password == "" {
			return nil, "", fmt.Errorf("密码不能为空")
		}
		user, err = uc.repo.GetUserByPhoneAndPassword(ctx, mobile, password)
		if err != nil {
			return nil, "", fmt.Errorf("手机号或密码错误")
		}
	case "sms":
		// 短信验证码登录
		if code == "" {
			return nil, "", fmt.Errorf("验证码不能为空")
		}
		// 验证短信验证码
		verified, err := uc.VerifySms(ctx, mobile, code, "login")
		if err != nil {
			return nil, "", fmt.Errorf("验证码验证失败: %w", err)
		}
		if !verified {
			return nil, "", fmt.Errorf("验证码错误或已过期")
		}
		// 根据手机号获取用户
		user, err = uc.repo.GetUserByPhone(ctx, mobile)
		if err != nil {
			return nil, "", fmt.Errorf("用户不存在")
		}
	default:
		return nil, "", fmt.Errorf("不支持的登录类型: %s", loginType)
	}

	// 检查用户状态
	if user.Status != domain.UserStatusNormal {
		return nil, "", fmt.Errorf("用户账号已被禁用")
	}

	// 生成 JWT token
	token, err := uc.generateJWTToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	uc.log.WithContext(ctx).Infof("用户登录成功: phone=%s, loginType=%s", mobile, loginType)
	return user, token, nil
}

// generateJWTToken 生成JWT token
func (uc *UserUsecase) generateJWTToken(user *domain.UserBase) (string, error) {
	// 这里简化处理，实际项目中应该使用JWT库生成真正的token
	// 可以使用 github.com/golang-jwt/jwt/v5 库
	token := fmt.Sprintf("token_%d_%s", user.UserId, user.Phone)
	return token, nil
}

// GetFileList 获取文件列表
func (uc *UserUsecase) GetFileList(ctx context.Context, page, pageSize int32, keyword string) ([]domain.FileInfo, int32, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算偏移量
	maxKeys := int(pageSize)

	var files []domain.FileInfo
	var err error

	if keyword != "" {
		// 如果有搜索关键词，使用搜索接口
		files, total, err := uc.minioRepo.SearchFiles(ctx, keyword, int32(page), int32(pageSize))
		if err != nil {
			uc.log.WithContext(ctx).Errorf("搜索文件失败: %v", err)
			return nil, 0, fmt.Errorf("搜索文件失败: %w", err)
		}
		return files, total, nil
	} else {
		// 否则列出所有文件
		files, err = uc.minioRepo.ListFiles(ctx, "", maxKeys)
		if err != nil {
			uc.log.WithContext(ctx).Errorf("列出文件失败: %v", err)
			return nil, 0, fmt.Errorf("列出文件失败: %w", err)
		}
	}

	if err != nil {
		uc.log.WithContext(ctx).Errorf("获取文件列表失败: %v", err)
		return nil, 0, fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 简单分页处理（实际应该在MinIO层面实现更高效的分页）
	start := int((page - 1) * pageSize)
	end := int(page * pageSize)

	total := int32(len(files))

	if start >= len(files) {
		return []domain.FileInfo{}, total, nil
	}

	if end > len(files) {
		end = len(files)
	}

	pagedFiles := files[start:end]

	uc.log.WithContext(ctx).Infof("获取文件列表成功: 总数=%d, 当前页=%d, 每页=%d", total, page, pageSize)
	return pagedFiles, total, nil
}

// GetUploadStats 获取上传统计
func (uc *UserUsecase) GetUploadStats(ctx context.Context) (map[string]interface{}, error) {
	// 获取文件统计信息
	stats, err := uc.minioRepo.GetFileStats(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("获取上传统计失败: %v", err)
		// 返回默认统计数据而不是错误
		return map[string]interface{}{
			"totalUploads":   0,
			"successUploads": 0,
			"totalSize":      int64(0),
			"todayUploads":   0,
		}, nil
	}

	uc.log.WithContext(ctx).Infof("获取上传统计成功: %+v", stats)
	return stats, nil
}

// DeleteFile 删除文件
func (uc *UserUsecase) DeleteFile(ctx context.Context, objectName string) error {
	if objectName == "" {
		return fmt.Errorf("文件名不能为空")
	}

	err := uc.minioRepo.Delete(ctx, objectName)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("删除文件失败: %v", err)
		return fmt.Errorf("删除文件失败: %w", err)
	}

	uc.log.WithContext(ctx).Infof("删除文件成功: %s", objectName)
	return nil
}

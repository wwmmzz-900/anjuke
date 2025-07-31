// Package data 实现了 domain 层定义的仓储接口。
// 这一层负责与具体的数据源（如 MySQL, Redis, 第三方 API）进行交互。
// 它依赖 domain 层来获取接口和模型的定义。
package data

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
)

// RealNameSDKInterface 定义实名认证SDK接口，用于测试时的mock
type RealNameSDKInterface interface {
	RealName(name, idCard string) (bool, error)
}

// UserRepo 实现了 domain.UserRepo 接口。
type UserRepo struct {
	data        *Data
	log         *log.Helper
	realNameSDK RealNameSDKInterface
}

// NewUserRepo 是 UserRepo 的构造函数，由 wire 自动调用。
func NewUserRepo(data *Data, realNameSDK *RealNameSDK, logger log.Logger) domain.UserRepo {
	return &UserRepo{
		data:        data,
		log:         log.NewHelper(logger),
		realNameSDK: realNameSDK,
	}
}

// SmsRiskControl 实现了短信风控的检查逻辑，它直接调用了封装在 Data 中的 Redis 操作。
func (u *UserRepo) SmsRiskControl(ctx context.Context, phone string) error {
	return u.data.SmsRiskControl(ctx, phone, "", "")
}

// todo:实名认证
func (u *UserRepo) RealName(ctx context.Context, user *domain.RealName) (*domain.RealName, error) {
	var userBase domain.UserBase
	// 查找用户
	if err := u.data.db.Where("user_id = ?", user.UserId).First(&userBase).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查用户是否已经实名认证
	if userBase.RealStatus == domain.RealNameVerified {
		return nil, fmt.Errorf("用户已完成实名认证，如需重新认证请先取消实名")
	}

	// 调用第三方实名认证
	u.log.Infof("开始调用第三方实名认证接口: 用户ID=%d, 姓名=%s", user.UserId, user.Name)
	ok, err := u.realNameSDK.RealName(user.Name, user.IdCard)
	if err != nil {
		u.log.Errorf("第三方实名认证接口调用失败: 用户ID=%d, 错误=%v", user.UserId, err)
		return nil, err
	}
	if !ok {
		u.log.Warnf("实名认证未通过: 用户ID=%d, 姓名=%s", user.UserId, user.Name)
		return nil, fmt.Errorf("实名认证未通过")
	}

	// 使用事务确保数据一致性
	tx := u.data.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 添加 realname 表数据
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		u.log.Errorf("实名信息保存失败: 用户ID=%d, 错误=%v", user.UserId, err)
		return nil, fmt.Errorf("实名信息保存失败: %v", err)
	}

	// 2. 更新 userbase 的实名状态
	if err := tx.Model(&domain.UserBase{}).
		Where("user_id = ?", user.UserId).
		Updates(map[string]interface{}{
			"real_name":   user.Name,
			"real_status": domain.RealNameVerified,
		}).Error; err != nil {
		tx.Rollback()
		u.log.Errorf("用户实名状态更新失败: 用户ID=%d, 错误=%v", user.UserId, err)
		return nil, fmt.Errorf("用户实名状态更新失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		u.log.Errorf("事务提交失败: 用户ID=%d, 错误=%v", user.UserId, err)
		return nil, fmt.Errorf("事务提交失败: %v", err)
	}

	u.log.Infof("用户实名认证成功: 用户ID=%d, 姓名=%s", user.UserId, user.Name)
	return user, nil
}

// todo:取消实名认证
func (u *UserRepo) UpdateUserStatus(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
	// 检查用户是否存在
	var userBase domain.UserBase
	if err := u.data.db.Where("user_id = ?", user.UserId).First(&userBase).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 如果是取消实名认证，检查用户是否已实名
	if user.RealStatus == domain.RealNameUnverified {
		if userBase.RealStatus != domain.RealNameVerified {
			return nil, fmt.Errorf("用户未实名认证，无需取消")
		}
		u.log.Infof("用户 %d 取消实名认证", user.UserId)
	}

	// 取消实名认证时，同时清空真实姓名
	updates := map[string]interface{}{
		"real_status": user.RealStatus,
	}

	// 如果是取消实名认证，清空真实姓名并删除实名认证记录
	if user.RealStatus == domain.RealNameUnverified {
		updates["real_name"] = ""

		// 删除实名认证记录
		if err := u.data.db.Where("user_id = ?", user.UserId).Delete(&domain.RealName{}).Error; err != nil {
			u.log.Errorf("删除实名认证记录失败: %v", err)
			// 不返回错误，继续执行状态更新
		}
	}

	err := u.data.db.Model(&domain.UserBase{}).
		Where("user_id = ?", user.UserId).
		Updates(updates).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CheckPhoneExists 检查手机号是否已存在
func (u *UserRepo) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	var count int64
	if err := u.data.db.Model(&domain.UserBase{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateUser 创建新用户
func (u *UserRepo) CreateUser(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
	// 对密码进行加密
	hashedPassword, err := u.hashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}
	user.Password = hashedPassword

	// 创建用户
	if err := u.data.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// hashPassword 对密码进行加密
func (u *UserRepo) hashPassword(password string) (string, error) {
	// 这里使用简单的MD5加密，实际应用中应使用更安全的算法如bcrypt
	hasher := md5.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// GetUserByPhoneAndPassword 根据手机号和密码获取用户（密码登录）
func (u *UserRepo) GetUserByPhoneAndPassword(ctx context.Context, phone, password string) (*domain.UserBase, error) {
	// 对密码进行加密
	hashedPassword, err := u.hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	var user domain.UserBase
	if err := u.data.db.Where("phone = ? AND password = ?", phone, hashedPassword).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在或密码错误")
	}
	return &user, nil
}

// GetUserByPhone 根据手机号获取用户（短信登录）
func (u *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*domain.UserBase, error) {
	var user domain.UserBase
	if err := u.data.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return &user, nil
}

// GetUserByID 根据用户ID获取用户
func (u *UserRepo) GetUserByID(ctx context.Context, id uint64) (*domain.UserBase, error) {
	var user domain.UserBase
	if err := u.data.db.Where("user_id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (u *UserRepo) UpdateUser(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
	if err := u.data.db.Model(&domain.UserBase{}).Where("user_id = ?", user.UserId).Updates(user).Error; err != nil {
		return nil, fmt.Errorf("更新用户失败: %v", err)
	}
	return user, nil
}

// DeleteUser 删除用户
func (u *UserRepo) DeleteUser(ctx context.Context, id uint64) error {
	// 使用事务确保数据一致性
	tx := u.data.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除实名认证记录
	if err := tx.Where("user_id = ?", id).Delete(&domain.RealName{}).Error; err != nil {
		tx.Rollback()
		u.log.Errorf("删除用户实名认证记录失败: 用户ID=%d, 错误=%v", id, err)
		return fmt.Errorf("删除用户实名认证记录失败: %v", err)
	}

	// 2. 删除用户基础信息
	if err := tx.Where("user_id = ?", id).Delete(&domain.UserBase{}).Error; err != nil {
		tx.Rollback()
		u.log.Errorf("删除用户失败: 用户ID=%d, 错误=%v", id, err)
		return fmt.Errorf("删除用户失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		u.log.Errorf("删除用户事务提交失败: 用户ID=%d, 错误=%v", id, err)
		return fmt.Errorf("删除用户事务提交失败: %v", err)
	}

	u.log.Infof("用户删除成功: 用户ID=%d", id)
	return nil
}

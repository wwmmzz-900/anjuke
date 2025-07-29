package biz

import "errors"

var (
	// 用户相关错误

	// 黑名单相关错误
	ErrUserAlreadyBlacklisted = errors.New("用户已在黑名单中")
	ErrUserNotInBlacklist     = errors.New("用户不在黑名单中")

	// 权限相关错误
	ErrPermissionNotFound = errors.New("权限不存在")
	ErrInvalidPermission  = errors.New("无效权限")
	ErrInternalError      = errors.New("内部服务器错误")
)

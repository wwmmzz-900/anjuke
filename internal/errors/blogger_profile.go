package errors

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/wwmmzz-900/anjuke/internal/model"
)

// BloggerProfile相关错误定义
var (
	// 用户相关错误
	ErrUserNotFound = errors.New(model.ErrCodeUserNotFound, "USER_NOT_FOUND", "用户不存在")
	ErrUserDisabled = errors.New(model.ErrCodeUserDisabled, "USER_DISABLED", "用户已禁用")
	ErrInvalidUserId = errors.New(model.ErrCodeInvalidUserId, "INVALID_USER_ID", "无效的用户ID")
	
	// 博主主页相关错误
	ErrProfileNotFound = errors.New(model.ErrCodeProfileNotFound, "PROFILE_NOT_FOUND", "博主主页不存在")
	ErrProfileDisabled = errors.New(model.ErrCodeProfileDisabled, "PROFILE_DISABLED", "博主主页已禁用")
	
	// 数据查询相关错误
	ErrStatsQueryFailed = errors.New(model.ErrCodeStatsQueryFailed, "STATS_QUERY_FAILED", "统计数据查询失败")
	ErrHouseQueryFailed = errors.New(model.ErrCodeHouseQueryFailed, "HOUSE_QUERY_FAILED", "房源查询失败")
	
	// 参数验证相关错误
	ErrInvalidPageParam = errors.New(model.ErrCodeInvalidPageParam, "INVALID_PAGE_PARAM", "无效的分页参数")
)

// BloggerProfileError 博主主页业务错误
type BloggerProfileError struct {
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现error接口
func (e *BloggerProfileError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("code: %d, reason: %s, message: %s, details: %s", e.Code, e.Reason, e.Message, e.Details)
	}
	return fmt.Sprintf("code: %d, reason: %s, message: %s", e.Code, e.Reason, e.Message)
}

// NewBloggerProfileError 创建博主主页业务错误
func NewBloggerProfileError(code int, reason, message string) *BloggerProfileError {
	return &BloggerProfileError{
		Code:    code,
		Reason:  reason,
		Message: message,
	}
}

// NewBloggerProfileErrorWithDetails 创建带详细信息的博主主页业务错误
func NewBloggerProfileErrorWithDetails(code int, reason, message, details string) *BloggerProfileError {
	return &BloggerProfileError{
		Code:    code,
		Reason:  reason,
		Message: message,
		Details: details,
	}
}

// IsUserNotFoundError 检查是否为用户不存在错误
func IsUserNotFoundError(err error) bool {
	if e, ok := err.(*BloggerProfileError); ok {
		return e.Code == model.ErrCodeUserNotFound
	}
	return false
}

// IsUserDisabledError 检查是否为用户禁用错误
func IsUserDisabledError(err error) bool {
	if e, ok := err.(*BloggerProfileError); ok {
		return e.Code == model.ErrCodeUserDisabled
	}
	return false
}

// IsInvalidUserIdError 检查是否为无效用户ID错误
func IsInvalidUserIdError(err error) bool {
	if e, ok := err.(*BloggerProfileError); ok {
		return e.Code == model.ErrCodeInvalidUserId
	}
	return false
}

// WrapUserNotFoundError 包装用户不存在错误
func WrapUserNotFoundError(userId int64) error {
	return NewBloggerProfileErrorWithDetails(
		model.ErrCodeUserNotFound,
		"USER_NOT_FOUND",
		"用户不存在",
		fmt.Sprintf("用户ID: %d", userId),
	)
}

// WrapUserDisabledError 包装用户禁用错误
func WrapUserDisabledError(userId int64) error {
	return NewBloggerProfileErrorWithDetails(
		model.ErrCodeUserDisabled,
		"USER_DISABLED",
		"用户已禁用",
		fmt.Sprintf("用户ID: %d", userId),
	)
}

// WrapInvalidUserIdError 包装无效用户ID错误
func WrapInvalidUserIdError(userId int64) error {
	return NewBloggerProfileErrorWithDetails(
		model.ErrCodeInvalidUserId,
		"INVALID_USER_ID",
		"无效的用户ID",
		fmt.Sprintf("用户ID: %d", userId),
	)
}

// WrapStatsQueryError 包装统计查询错误
func WrapStatsQueryError(err error) error {
	return NewBloggerProfileErrorWithDetails(
		model.ErrCodeStatsQueryFailed,
		"STATS_QUERY_FAILED",
		"统计数据查询失败",
		err.Error(),
	)
}

// WrapHouseQueryError 包装房源查询错误
func WrapHouseQueryError(err error) error {
	return NewBloggerProfileErrorWithDetails(
		model.ErrCodeHouseQueryFailed,
		"HOUSE_QUERY_FAILED",
		"房源查询失败",
		err.Error(),
	)
}
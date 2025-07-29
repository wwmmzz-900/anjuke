package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	v1 "anjuke/api/permission/v1"
	"anjuke/internal/biz"
)

// PermissionMiddleware 权限验证中间件
func PermissionMiddleware(permissionUC *biz.PermissionUsecase, logger log.Logger) middleware.Middleware {
	log := log.NewHelper(logger)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 获取传输信息
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			// 获取请求路径
			path := tr.Operation()

			// 跳过不需要权限验证的接口
			if shouldSkipPermissionCheck(path) {
				return handler(ctx, req)
			}

			// 从请求头或上下文获取用户ID（这里简化处理，实际应该从JWT token中获取）
			userIDStr := tr.RequestHeader().Get("X-User-ID")
			if userIDStr == "" {
				log.WithContext(ctx).Warn("缺少用户ID")
				return nil, errors.Unauthorized("UNAUTHORIZED", "缺少用户认证信息")
			}

			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				log.WithContext(ctx).Warnf("无效的用户ID: %s", userIDStr)
				return nil, errors.Unauthorized("UNAUTHORIZED", "无效的用户ID")
			}

			// 获取用户权限
			permissionReq := &v1.GetUserPermissionRequest{UserId: userID}
			permissionReply, err := permissionUC.GetUserPermission(ctx, permissionReq)
			if err != nil {
				log.WithContext(ctx).Errorf("获取用户权限失败: %v", err)
				return nil, errors.Forbidden("PERMISSION_DENIED", "权限验证失败")
			}

			// 检查权限
			requiredPermission := getRequiredPermission(path)
			if requiredPermission != v1.PermissionType_UNKNOWN {
				if !hasPermission(permissionReply.PermissionInfo.Permissions, requiredPermission) {
					log.WithContext(ctx).Warnf("用户 %d 缺少权限 %v 访问 %s", userID, requiredPermission, path)
					return nil, errors.Forbidden("PERMISSION_DENIED", "权限不足")
				}
			}

			// 将用户权限信息添加到上下文中，供后续使用
			ctx = context.WithValue(ctx, "user_id", userID)
			ctx = context.WithValue(ctx, "user_permissions", permissionReply.PermissionInfo.Permissions)
			ctx = context.WithValue(ctx, "user_role", permissionReply.PermissionInfo.Role)

			return handler(ctx, req)
		}
	}
}

// shouldSkipPermissionCheck 判断是否跳过权限检查
func shouldSkipPermissionCheck(path string) bool {
	skipPaths := []string{
		"/api.helloworld.v1.Greeter/SayHello",
		"/api.user.v2.User/CreateUser",                     // 用户注册不需要权限
		"/api.permission.v1.Permission/GetRolePermissions", // 获取角色权限不需要验证
	}

	for _, skipPath := range skipPaths {
		if strings.Contains(path, skipPath) {
			return true
		}
	}
	return false
}

// getRequiredPermission 根据接口路径获取所需权限
func getRequiredPermission(path string) v1.PermissionType {
	// 房源相关接口
	if strings.Contains(path, "House") {
		if strings.Contains(path, "Create") {
			return v1.PermissionType_PUBLISH_HOUSE
		}
		return v1.PermissionType_READ
	}

	// 用户管理相关接口
	if strings.Contains(path, "User") && !strings.Contains(path, "CreateUser") {
		return v1.PermissionType_MANAGE_USER
	}

	// 交易相关接口
	if strings.Contains(path, "Transaction") {
		return v1.PermissionType_MANAGE_TRANSACTION
	}

	// 黑名单相关接口
	if strings.Contains(path, "Blacklist") {
		if strings.Contains(path, "Add") || strings.Contains(path, "Remove") {
			return v1.PermissionType_MANAGE_USER
		}
		return v1.PermissionType_READ
	}

	// 权限管理相关接口
	if strings.Contains(path, "Permission") {
		if strings.Contains(path, "Update") || strings.Contains(path, "Batch") {
			return v1.PermissionType_ADMIN
		}
		return v1.PermissionType_READ
	}

	// 客服相关接口
	if strings.Contains(path, "Customer") {
		return v1.PermissionType_CUSTOMER_SERVICE
	}

	// 默认需要读取权限
	return v1.PermissionType_READ
}

// hasPermission 检查用户是否拥有指定权限
func hasPermission(userPermissions []v1.PermissionType, requiredPermission v1.PermissionType) bool {
	// 超级管理员拥有所有权限
	for _, perm := range userPermissions {
		if perm == v1.PermissionType_ADMIN {
			return true
		}
		if perm == requiredPermission {
			return true
		}
	}
	return false
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(ctx context.Context) int64 {
	if userID, ok := ctx.Value("user_id").(int64); ok {
		return userID
	}
	return 0
}

// GetUserPermissionsFromContext 从上下文获取用户权限
func GetUserPermissionsFromContext(ctx context.Context) []v1.PermissionType {
	if permissions, ok := ctx.Value("user_permissions").([]v1.PermissionType); ok {
		return permissions
	}
	return nil
}

// GetUserRoleFromContext 从上下文获取用户角色
func GetUserRoleFromContext(ctx context.Context) v1.UserRole {
	if role, ok := ctx.Value("user_role").(v1.UserRole); ok {
		return role
	}
	return v1.UserRole_ROLE_GUEST
}

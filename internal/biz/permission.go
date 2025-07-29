package biz

import (
	"context"
	"fmt"
	"time"

	v1 "anjuke/api/permission/v1"
	"github.com/go-kratos/kratos/v2/log"
)

// PermissionRepo 权限仓储接口
type PermissionRepo interface {
	// 保存用户权限
	SaveUserPermission(ctx context.Context, permission *UserPermission) (*UserPermission, error)
	// 获取用户权限
	GetUserPermission(ctx context.Context, userID int64) (*UserPermission, error)
	// 批量更新用户权限
	BatchUpdateUserPermission(ctx context.Context, permissions []*UserPermission) error
	// 获取权限列表
	GetPermissionList(ctx context.Context, page, pageSize int32, roleFilter v1.UserRole) ([]*UserPermission, int32, error)
	// 获取角色默认权限
	GetRolePermissions(ctx context.Context, role v1.UserRole) ([]v1.PermissionType, error)
	// 检查用户是否存在
	CheckUserExists(ctx context.Context, userID int64) (bool, error)
	// 获取用户信息
	GetUserInfo(ctx context.Context, userID int64) (*UserInfo, error)
}

// UserPermission 用户权限实体
type UserPermission struct {
	ID          int64               `json:"id"`
	UserID      int64               `json:"user_id"`
	Permissions []v1.PermissionType `json:"permissions"`
	Role        v1.UserRole         `json:"role"`
	OperatorID  int64               `json:"operator_id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	NickName string `json:"nick_name"`
}

// PermissionUsecase 权限用例
type PermissionUsecase struct {
	repo PermissionRepo
	log  *log.Helper
}

// NewPermissionUsecase 创建权限用例
func NewPermissionUsecase(repo PermissionRepo, logger log.Logger) *PermissionUsecase {
	return &PermissionUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// UpdateUserPermission 更新用户权限
func (uc *PermissionUsecase) UpdateUserPermission(ctx context.Context, req *v1.UpdateUserPermissionRequest) (*v1.UpdateUserPermissionReply, error) {
	// 检查用户是否存在
	exists, err := uc.repo.CheckUserExists(ctx, req.UserId)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("检查用户存在性失败: %v", err)
		return nil, nil
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// 验证权限合法性
	if err := uc.validatePermissions(req.Permissions, req.Role); err != nil {
		return nil, err
	}

	// 创建权限实体
	permission := &UserPermission{
		UserID:      req.UserId,
		Permissions: req.Permissions,
		Role:        req.Role,
		OperatorID:  req.OperatorId,
		UpdatedAt:   time.Now(),
	}

	// 保存权限
	savedPermission, err := uc.repo.SaveUserPermission(ctx, permission)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("保存用户权限失败: %v", err)
		return nil, ErrInternalError
	}

	uc.log.WithContext(ctx).Infof("用户权限更新成功, 用户ID: %d, 操作者ID: %d", req.UserId, req.OperatorId)

	return &v1.UpdateUserPermissionReply{
		Success:      true,
		Message:      "权限更新成功",
		PermissionId: savedPermission.ID,
	}, nil
}

// GetUserPermission 获取用户权限
func (uc *PermissionUsecase) GetUserPermission(ctx context.Context, req *v1.GetUserPermissionRequest) (*v1.GetUserPermissionReply, error) {
	permission, err := uc.repo.GetUserPermission(ctx, req.UserId)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("获取用户权限失败: %v", err)
		return nil, ErrInternalError
	}

	if permission == nil {
		return nil, ErrPermissionNotFound
	}

	// 获取操作者信息
	operatorInfo, _ := uc.repo.GetUserInfo(ctx, permission.OperatorID)
	operatorName := ""
	if operatorInfo != nil {
		operatorName = operatorInfo.Name
		if operatorName == "" {
			operatorName = operatorInfo.NickName
		}
	}

	return &v1.GetUserPermissionReply{
		PermissionInfo: &v1.UserPermissionInfo{
			UserId:       permission.UserID,
			Permissions:  permission.Permissions,
			Role:         permission.Role,
			CreatedAt:    permission.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    permission.UpdatedAt.Format("2006-01-02 15:04:05"),
			OperatorId:   permission.OperatorID,
			OperatorName: operatorName,
		},
	}, nil
}

// BatchUpdateUserPermission 批量更新用户权限
func (uc *PermissionUsecase) BatchUpdateUserPermission(ctx context.Context, req *v1.BatchUpdateUserPermissionRequest) (*v1.BatchUpdateUserPermissionReply, error) {
	var permissions []*UserPermission
	var errorMessages []string
	successCount := int32(0)
	failedCount := int32(0)

	for _, update := range req.Updates {
		// 检查用户是否存在
		exists, err := uc.repo.CheckUserExists(ctx, update.UserId)
		if err != nil || !exists {
			errorMessages = append(errorMessages, fmt.Sprintf("用户ID %d 不存在", update.UserId))
			failedCount++
			continue
		}

		// 验证权限合法性
		if err := uc.validatePermissions(update.Permissions, update.Role); err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("用户ID %d 权限验证失败: %v", update.UserId, err))
			failedCount++
			continue
		}

		permission := &UserPermission{
			UserID:      update.UserId,
			Permissions: update.Permissions,
			Role:        update.Role,
			OperatorID:  update.OperatorId,
			UpdatedAt:   time.Now(),
		}
		permissions = append(permissions, permission)
		successCount++
	}

	// 批量保存
	if len(permissions) > 0 {
		if err := uc.repo.BatchUpdateUserPermission(ctx, permissions); err != nil {
			uc.log.WithContext(ctx).Errorf("批量更新用户权限失败: %v", err)
			return nil, ErrInternalError
		}
	}

	return &v1.BatchUpdateUserPermissionReply{
		Success:       successCount > 0,
		Message:       fmt.Sprintf("批量更新完成，成功: %d, 失败: %d", successCount, failedCount),
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		ErrorMessages: errorMessages,
	}, nil
}

// GetPermissionList 获取权限列表
func (uc *PermissionUsecase) GetPermissionList(ctx context.Context, req *v1.GetPermissionListRequest) (*v1.GetPermissionListReply, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	permissions, total, err := uc.repo.GetPermissionList(ctx, req.Page, req.PageSize, req.RoleFilter)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("获取权限列表失败: %v", err)
		return nil, ErrInternalError
	}

	var items []*v1.PermissionItem
	for _, permission := range permissions {
		// 获取用户信息
		userInfo, _ := uc.repo.GetUserInfo(ctx, permission.UserID)
		userName := ""
		phone := ""
		if userInfo != nil {
			userName = userInfo.Name
			if userName == "" {
				userName = userInfo.NickName
			}
			phone = userInfo.Phone
		}

		// 获取操作者信息
		operatorInfo, _ := uc.repo.GetUserInfo(ctx, permission.OperatorID)
		operatorName := ""
		if operatorInfo != nil {
			operatorName = operatorInfo.Name
			if operatorName == "" {
				operatorName = operatorInfo.NickName
			}
		}

		items = append(items, &v1.PermissionItem{
			Id:           permission.ID,
			UserId:       permission.UserID,
			UserName:     userName,
			Phone:        phone,
			Permissions:  permission.Permissions,
			Role:         permission.Role,
			CreatedAt:    permission.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    permission.UpdatedAt.Format("2006-01-02 15:04:05"),
			OperatorId:   permission.OperatorID,
			OperatorName: operatorName,
		})
	}

	return &v1.GetPermissionListReply{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetRolePermissions 获取角色权限
func (uc *PermissionUsecase) GetRolePermissions(ctx context.Context, req *v1.GetRolePermissionsRequest) (*v1.GetRolePermissionsReply, error) {
	permissions, err := uc.repo.GetRolePermissions(ctx, req.RoleId)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("获取角色权限失败: %v", err)
		return nil, ErrInternalError
	}

	description := uc.getRoleDescription(req.RoleId)

	return &v1.GetRolePermissionsReply{
		Role:        req.RoleId,
		Permissions: permissions,
		Description: description,
	}, nil
}

// validatePermissions 验证权限合法性
func (uc *PermissionUsecase) validatePermissions(permissions []v1.PermissionType, role v1.UserRole) error {
	// 根据角色验证权限
	switch role {
	case v1.UserRole_ROLE_GUEST:
		// 游客只能有读取权限
		for _, perm := range permissions {
			if perm != v1.PermissionType_READ {
				return ErrInvalidPermission
			}
		}
	case v1.UserRole_ROLE_SUPER_ADMIN:
		// 超级管理员可以拥有所有权限
		break
	case v1.UserRole_ROLE_ADMIN:
		// 管理员不能拥有超级管理员权限
		for _, perm := range permissions {
			if perm == v1.PermissionType_ADMIN && role != v1.UserRole_ROLE_SUPER_ADMIN {
				return ErrInvalidPermission
			}
		}
	}

	return nil
}

// getRoleDescription 获取角色描述
func (uc *PermissionUsecase) getRoleDescription(role v1.UserRole) string {
	switch role {
	case v1.UserRole_ROLE_GUEST:
		return "游客用户，只能浏览基本信息"
	case v1.UserRole_ROLE_NORMAL_USER:
		return "普通用户，可以浏览和基本操作"
	case v1.UserRole_ROLE_VIP_USER:
		return "VIP用户，享有更多特权"
	case v1.UserRole_ROLE_LANDLORD:
		return "房东用户，可以发布和管理房源"
	case v1.UserRole_ROLE_AGENT:
		return "经纪人，可以管理房源和用户"
	case v1.UserRole_ROLE_ADMIN:
		return "管理员，拥有大部分管理权限"
	case v1.UserRole_ROLE_SUPER_ADMIN:
		return "超级管理员，拥有所有权限"
	default:
		return "未知角色"
	}
}

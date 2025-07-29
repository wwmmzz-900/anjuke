package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	v1 "anjuke/api/permission/v1"
	"anjuke/internal/biz"
)

// Permission 权限表模型
type Permission struct {
	PermissionID int64  `gorm:"column:permission_id;primaryKey;autoIncrement" json:"permission_id"`
	Name         string `gorm:"column:name;not null" json:"name"`
	Description  string `gorm:"column:description" json:"description"`
}

func (Permission) TableName() string {
	return "permission"
}

// Role 角色表模型
type Role struct {
	RoleID      int64  `gorm:"column:role_id;primaryKey;autoIncrement" json:"role_id"`
	Name        string `gorm:"column:name;not null" json:"name"`
	Description string `gorm:"column:description" json:"description"`
}

func (Role) TableName() string {
	return "role"
}

// RolePermission 角色权限关联表模型
type RolePermission struct {
	ID           int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleID       int64 `gorm:"column:role_id;not null" json:"role_id"`
	PermissionID int64 `gorm:"column:permission_id;not null" json:"permission_id"`
}

func (RolePermission) TableName() string {
	return "role_permission"
}

// UserPermission 用户权限表模型（需要新建）
type UserPermission struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"column:user_id;not null;index" json:"user_id"`
	RoleID      int64     `gorm:"column:role_id;not null" json:"role_id"`
	Permissions string    `gorm:"column:permissions;type:text" json:"permissions"` // JSON存储权限列表
	OperatorID  int64     `gorm:"column:operator_id" json:"operator_id"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserPermission) TableName() string {
	return "user_permission"
}

// User 用户表模型（假设已存在）
type User struct {
	ID       int64  `gorm:"column:id;primaryKey" json:"id"`
	Mobile   string `gorm:"column:mobile" json:"mobile"`
	NickName string `gorm:"column:nick_name" json:"nick_name"`
	Name     string `gorm:"column:name" json:"name"`
}

func (User) TableName() string {
	return "user_base"
}

// permissionRepo 权限仓储实现
type permissionRepo struct {
	data *Data
	log  *log.Helper
}

// NewPermissionRepo 创建权限仓储
func NewPermissionRepo(data *Data, logger log.Logger) biz.PermissionRepo {
	return &permissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// SaveUserPermission 保存用户权限
func (r *permissionRepo) SaveUserPermission(ctx context.Context, permission *biz.UserPermission) (*biz.UserPermission, error) {
	// 将权限列表转换为JSON字符串
	permissionsJSON, err := json.Marshal(permission.Permissions)
	if err != nil {
		return nil, fmt.Errorf("序列化权限失败: %w", err)
	}

	// 转换为数据模型
	userPerm := &UserPermission{
		UserID:      permission.UserID,
		RoleID:      int64(permission.Role),
		Permissions: string(permissionsJSON),
		OperatorID:  permission.OperatorID,
		UpdatedAt:   time.Now(),
	}

	// 检查是否已存在记录
	var existing UserPermission
	err = r.data.db.WithContext(ctx).Where("user_id = ?", permission.UserID).First(&existing).Error
	if err == nil {
		// 更新现有记录
		userPerm.ID = existing.ID
		userPerm.CreatedAt = existing.CreatedAt
		err = r.data.db.WithContext(ctx).Save(userPerm).Error
	} else if err == gorm.ErrRecordNotFound {
		// 创建新记录
		userPerm.CreatedAt = time.Now()
		err = r.data.db.WithContext(ctx).Create(userPerm).Error
	}

	if err != nil {
		return nil, fmt.Errorf("保存用户权限失败: %w", err)
	}

	// 转换回业务模型
	return &biz.UserPermission{
		ID:          userPerm.ID,
		UserID:      userPerm.UserID,
		Permissions: permission.Permissions,
		Role:        permission.Role,
		OperatorID:  userPerm.OperatorID,
		CreatedAt:   userPerm.CreatedAt,
		UpdatedAt:   userPerm.UpdatedAt,
	}, nil
}

// GetUserPermission 获取用户权限
func (r *permissionRepo) GetUserPermission(ctx context.Context, userID int64) (*biz.UserPermission, error) {
	var userPerm UserPermission
	err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).First(&userPerm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取用户权限失败: %w", err)
	}

	// 解析权限JSON
	var permissions []v1.PermissionType
	if userPerm.Permissions != "" {
		err = json.Unmarshal([]byte(userPerm.Permissions), &permissions)
		if err != nil {
			return nil, fmt.Errorf("解析权限JSON失败: %w", err)
		}
	}

	return &biz.UserPermission{
		ID:          userPerm.ID,
		UserID:      userPerm.UserID,
		Permissions: permissions,
		Role:        v1.UserRole(userPerm.RoleID),
		OperatorID:  userPerm.OperatorID,
		CreatedAt:   userPerm.CreatedAt,
		UpdatedAt:   userPerm.UpdatedAt,
	}, nil
}

// BatchUpdateUserPermission 批量更新用户权限
func (r *permissionRepo) BatchUpdateUserPermission(ctx context.Context, permissions []*biz.UserPermission) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, permission := range permissions {
			permissionsJSON, err := json.Marshal(permission.Permissions)
			if err != nil {
				return fmt.Errorf("序列化权限失败: %w", err)
			}

			userPerm := &UserPermission{
				UserID:      permission.UserID,
				RoleID:      int64(permission.Role),
				Permissions: string(permissionsJSON),
				OperatorID:  permission.OperatorID,
				UpdatedAt:   time.Now(),
			}

			// 检查是否已存在
			var existing UserPermission
			err = tx.Where("user_id = ?", permission.UserID).First(&existing).Error
			if err == nil {
				// 更新
				userPerm.ID = existing.ID
				userPerm.CreatedAt = existing.CreatedAt
				err = tx.Save(userPerm).Error
			} else if err == gorm.ErrRecordNotFound {
				// 创建
				userPerm.CreatedAt = time.Now()
				err = tx.Create(userPerm).Error
			}

			if err != nil {
				return fmt.Errorf("批量更新用户权限失败: %w", err)
			}
		}
		return nil
	})
}

// GetPermissionList 获取权限列表
func (r *permissionRepo) GetPermissionList(ctx context.Context, page, pageSize int32, roleFilter v1.UserRole) ([]*biz.UserPermission, int32, error) {
	var userPerms []UserPermission
	var total int64

	query := r.data.db.WithContext(ctx).Model(&UserPermission{})

	// 角色筛选
	if roleFilter != v1.UserRole(1) {
		query = query.Where("role_id = ?", int64(roleFilter))
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取权限总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Offset(int(offset)).Limit(int(pageSize)).Order("updated_at DESC").Find(&userPerms).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取权限列表失败: %w", err)
	}

	// 转换为业务模型
	var result []*biz.UserPermission
	for _, userPerm := range userPerms {
		var permissions []v1.PermissionType
		if userPerm.Permissions != "" {
			json.Unmarshal([]byte(userPerm.Permissions), &permissions)
		}

		result = append(result, &biz.UserPermission{
			ID:          userPerm.ID,
			UserID:      userPerm.UserID,
			Permissions: permissions,
			Role:        v1.UserRole(userPerm.RoleID),
			OperatorID:  userPerm.OperatorID,
			CreatedAt:   userPerm.CreatedAt,
			UpdatedAt:   userPerm.UpdatedAt,
		})
	}

	return result, int32(total), nil
}

// GetRolePermissions 获取角色权限
func (r *permissionRepo) GetRolePermissions(ctx context.Context, role v1.UserRole) ([]v1.PermissionType, error) {
	// 根据角色返回默认权限
	switch role {
	case v1.UserRole_ROLE_GUEST:
		return []v1.PermissionType{v1.PermissionType_READ}, nil
	case v1.UserRole_ROLE_NORMAL_USER:
		return []v1.PermissionType{v1.PermissionType_READ, v1.PermissionType_WRITE}, nil
	case v1.UserRole_ROLE_VIP_USER:
		return []v1.PermissionType{v1.PermissionType_READ, v1.PermissionType_WRITE, v1.PermissionType_PUBLISH_HOUSE}, nil
	case v1.UserRole_ROLE_LANDLORD:
		return []v1.PermissionType{v1.PermissionType_READ, v1.PermissionType_WRITE, v1.PermissionType_PUBLISH_HOUSE}, nil
	case v1.UserRole_ROLE_AGENT:
		return []v1.PermissionType{
			v1.PermissionType_READ, v1.PermissionType_WRITE, v1.PermissionType_PUBLISH_HOUSE,
			v1.PermissionType_MANAGE_USER, v1.PermissionType_CUSTOMER_SERVICE,
		}, nil
	case v1.UserRole_ROLE_ADMIN:
		return []v1.PermissionType{
			v1.PermissionType_READ, v1.PermissionType_WRITE, v1.PermissionType_DELETE,
			v1.PermissionType_PUBLISH_HOUSE, v1.PermissionType_MANAGE_USER,
			v1.PermissionType_MANAGE_TRANSACTION, v1.PermissionType_CUSTOMER_SERVICE,
		}, nil
	case v1.UserRole_ROLE_SUPER_ADMIN:
		return []v1.PermissionType{
			v1.PermissionType_READ, v1.PermissionType_WRITE, v1.PermissionType_DELETE,
			v1.PermissionType_ADMIN, v1.PermissionType_PUBLISH_HOUSE, v1.PermissionType_MANAGE_USER,
			v1.PermissionType_MANAGE_TRANSACTION, v1.PermissionType_CUSTOMER_SERVICE,
		}, nil
	default:
		return []v1.PermissionType{}, nil
	}
}

// CheckUserExists 检查用户是否存在
func (r *permissionRepo) CheckUserExists(ctx context.Context, userID int64) (bool, error) {
	var count int64
	err := r.data.db.WithContext(ctx).Model(&User{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查用户存在性失败: %w", err)
	}
	return count > 0, nil
}

// GetUserInfo 获取用户信息
func (r *permissionRepo) GetUserInfo(ctx context.Context, userID int64) (*biz.UserInfo, error) {
	var user User
	err := r.data.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	return &biz.UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		Phone:    user.Mobile,
		NickName: user.NickName,
	}, nil
}

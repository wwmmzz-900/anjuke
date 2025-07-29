package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "anjuke/api/permission/v1"
	"anjuke/internal/biz"
)

// PermissionService 权限服务
type PermissionService struct {
	v1.UnimplementedPermissionServer

	uc  *biz.PermissionUsecase
	log *log.Helper
}

// NewPermissionService 创建权限服务
func NewPermissionService(uc *biz.PermissionUsecase, logger log.Logger) *PermissionService {
	return &PermissionService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// UpdateUserPermission 更新用户权限
func (s *PermissionService) UpdateUserPermission(ctx context.Context, req *v1.UpdateUserPermissionRequest) (*v1.UpdateUserPermissionReply, error) {
	s.log.WithContext(ctx).Infof("更新用户权限请求: 用户ID=%d, 角色=%v, 操作者ID=%d", req.UserId, req.Role, req.OperatorId)

	reply, err := s.uc.UpdateUserPermission(ctx, req)
	if err != nil {
		s.log.WithContext(ctx).Errorf("更新用户权限失败: %v", err)
		return nil, err
	}

	s.log.WithContext(ctx).Infof("更新用户权限成功: 用户ID=%d", req.UserId)
	return reply, nil
}

// GetUserPermission 获取用户权限
func (s *PermissionService) GetUserPermission(ctx context.Context, req *v1.GetUserPermissionRequest) (*v1.GetUserPermissionReply, error) {
	s.log.WithContext(ctx).Infof("获取用户权限请求: 用户ID=%d", req.UserId)

	reply, err := s.uc.GetUserPermission(ctx, req)
	if err != nil {
		s.log.WithContext(ctx).Errorf("获取用户权限失败: %v", err)
		return nil, err
	}

	return reply, nil
}

// BatchUpdateUserPermission 批量更新用户权限
func (s *PermissionService) BatchUpdateUserPermission(ctx context.Context, req *v1.BatchUpdateUserPermissionRequest) (*v1.BatchUpdateUserPermissionReply, error) {
	s.log.WithContext(ctx).Infof("批量更新用户权限请求: 更新数量=%d", len(req.Updates))

	reply, err := s.uc.BatchUpdateUserPermission(ctx, req)
	if err != nil {
		s.log.WithContext(ctx).Errorf("批量更新用户权限失败: %v", err)
		return nil, err
	}

	s.log.WithContext(ctx).Infof("批量更新用户权限完成: 成功=%d, 失败=%d", reply.SuccessCount, reply.FailedCount)
	return reply, nil
}

// GetPermissionList 获取权限列表
func (s *PermissionService) GetPermissionList(ctx context.Context, req *v1.GetPermissionListRequest) (*v1.GetPermissionListReply, error) {
	s.log.WithContext(ctx).Infof("获取权限列表请求: 页码=%d, 页大小=%d, 角色筛选=%v", req.Page, req.PageSize, req.RoleFilter)

	reply, err := s.uc.GetPermissionList(ctx, req)
	if err != nil {
		s.log.WithContext(ctx).Errorf("获取权限列表失败: %v", err)
		return nil, err
	}

	s.log.WithContext(ctx).Infof("获取权限列表成功: 总数=%d", reply.Total)
	return reply, nil
}

// GetRolePermissions 获取角色权限
func (s *PermissionService) GetRolePermissions(ctx context.Context, req *v1.GetRolePermissionsRequest) (*v1.GetRolePermissionsReply, error) {
	s.log.WithContext(ctx).Infof("获取角色权限请求: 角色=%v", req.RoleId)

	reply, err := s.uc.GetRolePermissions(ctx, req)
	if err != nil {
		s.log.WithContext(ctx).Errorf("获取角色权限失败: %v", err)
		return nil, err
	}

	return reply, nil
}

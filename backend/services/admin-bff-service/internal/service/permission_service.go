package service

import (
	"context"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
)

// PermissionService 权限服务接口
type PermissionService interface {
	// 权限查询
	GetPermission(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	ListPermissions(ctx context.Context, resource string) ([]*model.Permission, error)
	ListPermissionsByResource(ctx context.Context) (map[string][]*model.Permission, error)
}

type permissionService struct {
	permissionRepo repository.PermissionRepository
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(
	permissionRepo repository.PermissionRepository,
) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
	}
}

// GetPermission 获取权限详情
func (s *permissionService) GetPermission(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, ErrRoleNotFound // 复用错误，或创建新的 ErrPermissionNotFound
	}
	return permission, nil
}

// ListPermissions 获取权限列表
func (s *permissionService) ListPermissions(ctx context.Context, resource string) ([]*model.Permission, error) {
	return s.permissionRepo.List(ctx, resource)
}

// ListPermissionsByResource 按资源分组获取权限列表
func (s *permissionService) ListPermissionsByResource(ctx context.Context) (map[string][]*model.Permission, error) {
	// 获取所有权限
	permissions, err := s.permissionRepo.List(ctx, "")
	if err != nil {
		return nil, err
	}

	// 按资源分组
	grouped := make(map[string][]*model.Permission)
	for _, perm := range permissions {
		grouped[perm.Resource] = append(grouped[perm.Resource], perm)
	}

	return grouped, nil
}

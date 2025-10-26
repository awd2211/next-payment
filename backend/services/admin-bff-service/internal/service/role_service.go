package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
)

var (
	ErrRoleNotFound      = errors.New("角色不存在")
	ErrRoleExists        = errors.New("角色已存在")
	ErrSystemRoleProtect = errors.New("系统角色不可删除")
)

// RoleService 角色服务接口
type RoleService interface {
	// 角色管理
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*model.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*model.Role, error)
	ListRoles(ctx context.Context, page, pageSize int) ([]*model.Role, int64, error)
	UpdateRole(ctx context.Context, req *UpdateRoleRequest) (*model.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	AssignRoleToAdmin(ctx context.Context, adminID, roleID uuid.UUID) error
}

type roleService struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	adminRepo      repository.AdminRepository
}

// NewRoleService 创建角色服务实例
func NewRoleService(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	adminRepo repository.AdminRepository,
) RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		adminRepo:      adminRepo,
	}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string
	DisplayName   string
	Description   string
	PermissionIDs []uuid.UUID
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	ID            uuid.UUID
	DisplayName   string
	Description   string
	PermissionIDs []uuid.UUID
}

// CreateRole 创建角色
func (s *roleService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*model.Role, error) {
	// 检查角色名是否已存在
	existingRole, err := s.roleRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existingRole != nil {
		return nil, ErrRoleExists
	}

	// 创建角色
	role := &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		if err := s.roleRepo.AssignPermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return nil, err
		}
	}

	// 重新加载角色以获取权限
	return s.roleRepo.GetByID(ctx, role.ID)
}

// GetRole 获取角色详情
func (s *roleService) GetRole(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

// ListRoles 获取角色列表
func (s *roleService) ListRoles(ctx context.Context, page, pageSize int) ([]*model.Role, int64, error) {
	return s.roleRepo.List(ctx, page, pageSize)
}

// UpdateRole 更新角色
func (s *roleService) UpdateRole(ctx context.Context, req *UpdateRoleRequest) (*model.Role, error) {
	// 获取现有角色
	role, err := s.roleRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}

	// 系统角色不可修改名称
	if role.IsSystem {
		// 只允许修改描述
		if req.Description != "" {
			role.Description = req.Description
		}
	} else {
		// 自定义角色可以修改所有信息
		if req.DisplayName != "" {
			role.DisplayName = req.DisplayName
		}
		if req.Description != "" {
			role.Description = req.Description
		}
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	// 更新权限
	if len(req.PermissionIDs) > 0 {
		if err := s.roleRepo.AssignPermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return nil, err
		}
	}

	// 重新加载角色以获取更新后的权限
	return s.roleRepo.GetByID(ctx, role.ID)
}

// DeleteRole 删除角色
func (s *roleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	// 系统角色不可删除
	if role.IsSystem {
		return ErrSystemRoleProtect
	}

	return s.roleRepo.Delete(ctx, id)
}

// AssignPermissions 为角色分配权限
func (s *roleService) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	// 验证所有权限ID是否有效
	permissions, err := s.permissionRepo.ListByIDs(ctx, permissionIDs)
	if err != nil {
		return err
	}
	if len(permissions) != len(permissionIDs) {
		return errors.New("部分权限ID无效")
	}

	return s.roleRepo.AssignPermissions(ctx, roleID, permissionIDs)
}

// AssignRoleToAdmin 为管理员分配角色
func (s *roleService) AssignRoleToAdmin(ctx context.Context, adminID, roleID uuid.UUID) error {
	// 检查管理员是否存在
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return ErrAdminNotFound
	}

	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	// 添加角色到管理员
	admin.Roles = append(admin.Roles, *role)
	return s.adminRepo.Update(ctx, admin)
}

package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/payment-platform/services/admin-service/internal/model"
	"gorm.io/gorm"
)

// RoleRepository 角色仓储接口
type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
	List(ctx context.Context, page, pageSize int) ([]*model.Role, int64, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓储实例
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// Create 创建角色
func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// GetByID 根据ID获取角色
func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// GetByName 根据名称获取角色
func (r *roleRepository) GetByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "name = ?", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// List 分页查询角色列表
func (r *roleRepository) List(ctx context.Context, page, pageSize int) ([]*model.Role, int64, error) {
	var roles []*model.Role
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Role{})

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Preload("Permissions").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&roles).Error

	return roles, total, err
}

// Update 更新角色
func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete 删除角色
func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除角色-权限关联
		if err := tx.Where("role_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		// 删除管理员-角色关联
		if err := tx.Where("role_id = ?", id).Delete(&model.AdminRole{}).Error; err != nil {
			return err
		}
		// 删除角色
		return tx.Delete(&model.Role{}, "id = ?", id).Error
	})
}

// AssignPermissions 为角色分配权限
func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除现有权限
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}

		// 添加新权限
		for _, permissionID := range permissionIDs {
			rolePermission := &model.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
			}
			if err := tx.Create(rolePermission).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

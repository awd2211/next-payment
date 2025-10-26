package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"gorm.io/gorm"
)

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	GetByCode(ctx context.Context, code string) (*model.Permission, error)
	List(ctx context.Context, resource string) ([]*model.Permission, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓储实例
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// GetByID 根据ID获取权限
func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// GetByCode 根据代码获取权限
func (r *permissionRepository) GetByCode(ctx context.Context, code string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.WithContext(ctx).First(&permission, "code = ?", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// List 查询权限列表
func (r *permissionRepository) List(ctx context.Context, resource string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	query := r.db.WithContext(ctx).Model(&model.Permission{})

	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	err := query.Order("resource ASC, action ASC").Find(&permissions).Error
	return permissions, err
}

// ListByIDs 根据ID列表获取权限
func (r *permissionRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}

package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/payment-platform/services/admin-service/internal/model"
	"gorm.io/gorm"
)

// AdminRepository 管理员仓储接口
type AdminRepository interface {
	Create(ctx context.Context, admin *model.Admin) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Admin, error)
	GetByUsername(ctx context.Context, username string) (*model.Admin, error)
	GetByEmail(ctx context.Context, email string) (*model.Admin, error)
	List(ctx context.Context, page, pageSize int, status, keyword string) ([]*model.Admin, int64, error)
	Update(ctx context.Context, admin *model.Admin) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error
}

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository 创建管理员仓储实例
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

// Create 创建管理员
func (r *adminRepository) Create(ctx context.Context, admin *model.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

// GetByID 根据ID获取管理员
func (r *adminRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		First(&admin, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// GetByUsername 根据用户名获取管理员
func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		First(&admin, "username = ?", username).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// GetByEmail 根据邮箱获取管理员
func (r *adminRepository) GetByEmail(ctx context.Context, email string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&admin, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// List 分页查询管理员列表
func (r *adminRepository) List(ctx context.Context, page, pageSize int, status, keyword string) ([]*model.Admin, int64, error) {
	var admins []*model.Admin
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Admin{})

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 关键词搜索（用户名、邮箱、全名）
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Preload("Roles").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&admins).Error

	return admins, total, err
}

// Update 更新管理员
func (r *adminRepository) Update(ctx context.Context, admin *model.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

// Delete 软删除管理员
func (r *adminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Admin{}, "id = ?", id).Error
}

// UpdateLastLogin 更新最后登录时间和IP
func (r *adminRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	return r.db.WithContext(ctx).
		Model(&model.Admin{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"last_login_ip": ip,
		}).Error
}

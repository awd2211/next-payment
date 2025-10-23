package repository

import (
	"context"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"gorm.io/gorm"
)

// SystemConfigRepository 系统配置仓储接口
type SystemConfigRepository interface {
	Create(ctx context.Context, config *model.SystemConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.SystemConfig, error)
	GetByKey(ctx context.Context, key string) (*model.SystemConfig, error)
	List(ctx context.Context, category string, page, pageSize int) ([]*model.SystemConfig, int64, error)
	Update(ctx context.Context, config *model.SystemConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByCategory(ctx context.Context) (map[string][]*model.SystemConfig, error)
}

type systemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置仓储实例
func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

// Create 创建系统配置
func (r *systemConfigRepository) Create(ctx context.Context, config *model.SystemConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetByID 根据ID获取系统配置
func (r *systemConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.WithContext(ctx).First(&config, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetByKey 根据键获取系统配置
func (r *systemConfigRepository) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.WithContext(ctx).First(&config, "key = ?", key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// List 分页查询系统配置列表
func (r *systemConfigRepository) List(ctx context.Context, category string, page, pageSize int) ([]*model.SystemConfig, int64, error) {
	var configs []*model.SystemConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SystemConfig{})

	// 应用过滤条件
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	err := query.Order("category ASC, key ASC").Find(&configs).Error
	return configs, total, err
}

// Update 更新系统配置
func (r *systemConfigRepository) Update(ctx context.Context, config *model.SystemConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除系统配置
func (r *systemConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.SystemConfig{}, "id = ?", id).Error
}

// ListByCategory 按类别分组获取配置
func (r *systemConfigRepository) ListByCategory(ctx context.Context) (map[string][]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.WithContext(ctx).Order("category ASC, key ASC").Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// 按类别分组
	grouped := make(map[string][]*model.SystemConfig)
	for _, config := range configs {
		grouped[config.Category] = append(grouped[config.Category], config)
	}

	return grouped, nil
}

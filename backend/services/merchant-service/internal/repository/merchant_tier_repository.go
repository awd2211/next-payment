package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// MerchantTierRepository 商户等级仓库接口
type MerchantTierRepository interface {
	// 获取等级配置
	GetByTier(ctx context.Context, tier model.MerchantTier) (*model.MerchantTierConfig, error)
	GetAll(ctx context.Context) ([]*model.MerchantTierConfig, error)

	// 创建/更新配置
	Create(ctx context.Context, config *model.MerchantTierConfig) error
	Update(ctx context.Context, config *model.MerchantTierConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// merchantTierRepository 仓库实现
type merchantTierRepository struct {
	db *gorm.DB
}

// NewMerchantTierRepository 创建商户等级仓库
func NewMerchantTierRepository(db *gorm.DB) MerchantTierRepository {
	return &merchantTierRepository{db: db}
}

// GetByTier 根据等级获取配置
func (r *merchantTierRepository) GetByTier(ctx context.Context, tier model.MerchantTier) (*model.MerchantTierConfig, error) {
	var config model.MerchantTierConfig
	err := r.db.WithContext(ctx).
		Where("tier = ?", tier).
		First(&config).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &config, err
}

// GetAll 获取所有等级配置
func (r *merchantTierRepository) GetAll(ctx context.Context) ([]*model.MerchantTierConfig, error) {
	var configs []*model.MerchantTierConfig
	err := r.db.WithContext(ctx).
		Order("priority ASC").
		Find(&configs).Error

	return configs, err
}

// Create 创建等级配置
func (r *merchantTierRepository) Create(ctx context.Context, config *model.MerchantTierConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新等级配置
func (r *merchantTierRepository) Update(ctx context.Context, config *model.MerchantTierConfig) error {
	return r.db.WithContext(ctx).
		Model(&model.MerchantTierConfig{}).
		Where("id = ?", config.ID).
		Updates(config).Error
}

// Delete 删除等级配置
func (r *merchantTierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&model.MerchantTierConfig{}, "id = ?", id).Error
}

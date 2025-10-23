package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// MerchantFeeConfigRepository 商户费率配置仓储接口
type MerchantFeeConfigRepository interface {
	Create(ctx context.Context, config *model.MerchantFeeConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantFeeConfig, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error)
	GetActiveByMerchant(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string) (*model.MerchantFeeConfig, error)
	Update(ctx context.Context, config *model.MerchantFeeConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type merchantFeeConfigRepository struct {
	db *gorm.DB
}

// NewMerchantFeeConfigRepository 创建商户费率配置仓储实例
func NewMerchantFeeConfigRepository(db *gorm.DB) MerchantFeeConfigRepository {
	return &merchantFeeConfigRepository{db: db}
}

func (r *merchantFeeConfigRepository) Create(ctx context.Context, config *model.MerchantFeeConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *merchantFeeConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantFeeConfig, error) {
	var config model.MerchantFeeConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *merchantFeeConfigRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error) {
	var configs []*model.MerchantFeeConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("priority DESC, created_at DESC").
		Find(&configs).Error
	return configs, err
}

func (r *merchantFeeConfigRepository) GetActiveByMerchant(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string) (*model.MerchantFeeConfig, error) {
	var config model.MerchantFeeConfig
	query := r.db.WithContext(ctx).
		Where("merchant_id = ? AND status = ?", merchantID, "active").
		Where("effective_date <= NOW()").
		Where("(expiry_date IS NULL OR expiry_date > NOW())")

	if channel != "" {
		query = query.Where("(channel = ? OR channel = 'all')", channel)
	}

	if paymentMethod != "" {
		query = query.Where("(payment_method = ? OR payment_method = 'all')", paymentMethod)
	}

	err := query.Order("priority DESC, effective_date DESC").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *merchantFeeConfigRepository) Update(ctx context.Context, config *model.MerchantFeeConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *merchantFeeConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantFeeConfig{}, "id = ?", id).Error
}

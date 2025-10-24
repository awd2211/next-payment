package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-config-service/internal/model"
)

// FeeConfigRepository 费率配置仓储接口
type FeeConfigRepository interface {
	Create(ctx context.Context, config *model.MerchantFeeConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantFeeConfig, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error)
	GetEffectiveConfig(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string, queryTime time.Time) (*model.MerchantFeeConfig, error)
	Update(ctx context.Context, config *model.MerchantFeeConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, status string, limit, offset int) ([]*model.MerchantFeeConfig, int64, error)
}

type feeConfigRepository struct {
	db *gorm.DB
}

// NewFeeConfigRepository 创建费率配置仓储实例
func NewFeeConfigRepository(db *gorm.DB) FeeConfigRepository {
	return &feeConfigRepository{db: db}
}

// Create 创建费率配置
func (r *feeConfigRepository) Create(ctx context.Context, config *model.MerchantFeeConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetByID 根据ID获取费率配置
func (r *feeConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantFeeConfig, error) {
	var config model.MerchantFeeConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByMerchantID 获取商户的所有费率配置
func (r *feeConfigRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error) {
	var configs []*model.MerchantFeeConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("priority DESC, created_at DESC").
		Find(&configs).Error
	return configs, err
}

// GetEffectiveConfig 获取生效的费率配置（根据渠道、支付方式、时间查询）
func (r *feeConfigRepository) GetEffectiveConfig(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string, queryTime time.Time) (*model.MerchantFeeConfig, error) {
	var config model.MerchantFeeConfig

	// 查询条件：
	// 1. merchant_id 匹配
	// 2. status = 'active'
	// 3. effective_date <= queryTime
	// 4. (expiry_date IS NULL OR expiry_date > queryTime)
	// 5. channel 匹配（精确匹配或'all'）
	// 6. payment_method 匹配（精确匹配或'all'）
	// 按 priority DESC 排序，取第一条

	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("status = ?", model.FeeStatusActive).
		Where("effective_date <= ?", queryTime).
		Where("(expiry_date IS NULL OR expiry_date > ?)", queryTime).
		Where("(channel = ? OR channel = ?)", channel, model.ChannelAll).
		Where("(payment_method = ? OR payment_method = ?)", paymentMethod, model.PaymentMethodAll).
		Order("priority DESC, created_at DESC").
		First(&config).Error

	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Update 更新费率配置
func (r *feeConfigRepository) Update(ctx context.Context, config *model.MerchantFeeConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除费率配置（软删除）
func (r *feeConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantFeeConfig{}, "id = ?", id).Error
}

// List 列出费率配置（管理员功能）
func (r *feeConfigRepository) List(ctx context.Context, status string, limit, offset int) ([]*model.MerchantFeeConfig, int64, error) {
	var configs []*model.MerchantFeeConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.MerchantFeeConfig{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&configs).Error
	return configs, total, err
}

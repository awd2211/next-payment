package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-config-service/internal/model"
)

// ChannelConfigRepository 渠道配置仓储接口
type ChannelConfigRepository interface {
	Create(ctx context.Context, config *model.ChannelConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error)
	GetByMerchantAndChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error)
	GetEnabledByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error)
	Update(ctx context.Context, config *model.ChannelConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	EnableChannel(ctx context.Context, id uuid.UUID) error
	DisableChannel(ctx context.Context, id uuid.UUID) error
}

type channelConfigRepository struct {
	db *gorm.DB
}

// NewChannelConfigRepository 创建渠道配置仓储实例
func NewChannelConfigRepository(db *gorm.DB) ChannelConfigRepository {
	return &channelConfigRepository{db: db}
}

// Create 创建渠道配置
func (r *channelConfigRepository) Create(ctx context.Context, config *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetByID 根据ID获取渠道配置
func (r *channelConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error) {
	var config model.ChannelConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByMerchantID 获取商户的所有渠道配置
func (r *channelConfigRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error) {
	var configs []*model.ChannelConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("channel").
		Find(&configs).Error
	return configs, err
}

// GetByMerchantAndChannel 获取商户指定渠道的配置
func (r *channelConfigRepository) GetByMerchantAndChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error) {
	var config model.ChannelConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND channel = ?", merchantID, channel).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetEnabledByMerchant 获取商户所有启用的渠道配置
func (r *channelConfigRepository) GetEnabledByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error) {
	var configs []*model.ChannelConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND is_enabled = ?", merchantID, true).
		Order("channel").
		Find(&configs).Error
	return configs, err
}

// Update 更新渠道配置
func (r *channelConfigRepository) Update(ctx context.Context, config *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除渠道配置（软删除）
func (r *channelConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelConfig{}, "id = ?", id).Error
}

// EnableChannel 启用渠道
func (r *channelConfigRepository) EnableChannel(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.ChannelConfig{}).
		Where("id = ?", id).
		Update("is_enabled", true).Error
}

// DisableChannel 停用渠道
func (r *channelConfigRepository) DisableChannel(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.ChannelConfig{}).
		Where("id = ?", id).
		Update("is_enabled", false).Error
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// ChannelRepository 渠道仓储接口
type ChannelRepository interface {
	Create(ctx context.Context, channel *model.ChannelConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error)
	GetByMerchantAndChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error)
	ListByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error)
	Update(ctx context.Context, channel *model.ChannelConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// channelRepository 渠道仓储实现
type channelRepository struct {
	db *gorm.DB
}

// NewChannelRepository 创建渠道仓储实例
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

// Create 创建渠道配置
func (r *channelRepository) Create(ctx context.Context, channel *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

// GetByID 根据ID获取渠道配置
func (r *channelRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error) {
	var channel model.ChannelConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// GetByMerchantAndChannel 根据商户ID和渠道类型获取配置
func (r *channelRepository) GetByMerchantAndChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error) {
	var config model.ChannelConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND channel = ?", merchantID, channel).
		First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// ListByMerchantID 获取商户的所有渠道配置
func (r *channelRepository) ListByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error) {
	var channels []*model.ChannelConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// Update 更新渠道配置
func (r *channelRepository) Update(ctx context.Context, channel *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

// Delete 删除渠道配置（软删除）
func (r *channelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ChannelConfig{}).Error
}

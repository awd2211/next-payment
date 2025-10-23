package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// WebhookRepository Webhook仓储接口
type WebhookRepository interface {
	Create(ctx context.Context, webhook *model.WebhookConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookConfig, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.WebhookConfig, error)
	Update(ctx context.Context, webhook *model.WebhookConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// webhookRepository Webhook仓储实现
type webhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository 创建Webhook仓储实例
func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

// Create 创建Webhook配置
func (r *webhookRepository) Create(ctx context.Context, webhook *model.WebhookConfig) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

// GetByID 根据ID获取Webhook配置
func (r *webhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookConfig, error) {
	var webhook model.WebhookConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&webhook).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// GetByMerchantID 根据商户ID获取Webhook配置
func (r *webhookRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.WebhookConfig, error) {
	var webhook model.WebhookConfig
	err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&webhook).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &webhook, nil
}

// Update 更新Webhook配置
func (r *webhookRepository) Update(ctx context.Context, webhook *model.WebhookConfig) error {
	return r.db.WithContext(ctx).Save(webhook).Error
}

// Delete 删除Webhook配置
func (r *webhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.WebhookConfig{}).Error
}

package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/payment-platform/services/merchant-service/internal/model"
	"gorm.io/gorm"
)

// APIKeyRepository API密钥仓储接口
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *model.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.APIKey, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, environment string) ([]*model.APIKey, error)
	Update(ctx context.Context, apiKey *model.APIKey) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository 创建API密钥仓储实例
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

// Create 创建API密钥
func (r *apiKeyRepository) Create(ctx context.Context, apiKey *model.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// GetByID 根据ID获取API密钥
func (r *apiKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.APIKey, error) {
	var apiKey model.APIKey
	err := r.db.WithContext(ctx).First(&apiKey, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

// GetByAPIKey 根据API Key获取记录
func (r *apiKeyRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.WithContext(ctx).First(&key, "api_key = ? AND is_active = true", apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

// ListByMerchant 获取商户的所有API密钥
func (r *apiKeyRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, environment string) ([]*model.APIKey, error) {
	var apiKeys []*model.APIKey

	query := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)

	// 环境筛选
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}

	err := query.Order("created_at DESC").Find(&apiKeys).Error
	return apiKeys, err
}

// Update 更新API密钥
func (r *apiKeyRepository) Update(ctx context.Context, apiKey *model.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// UpdateLastUsed 更新最后使用时间
func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", gorm.Expr("NOW()")).Error
}

// Revoke 撤销API密钥（禁用）
func (r *apiKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// Delete 删除API密钥
func (r *apiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.APIKey{}, "id = ?", id).Error
}

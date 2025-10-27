package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-auth-service/internal/model"
)

// APIKeyRepository API密钥仓储接口
type APIKeyRepository interface {
	GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error)
	UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, key *model.APIKey) error
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.APIKey, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByIDAndMerchantID(ctx context.Context, id uuid.UUID, merchantID uuid.UUID) (*model.APIKey, error)
}

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository 创建API密钥仓储
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

// GetByAPIKey 根据API Key查询
func (r *apiKeyRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.WithContext(ctx).
		Where("api_key = ? AND is_active = ?", apiKey, true).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// UpdateLastUsedAt 更新最后使用时间
func (r *apiKeyRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}

// Create 创建新的API Key
func (r *apiKeyRepository) Create(ctx context.Context, key *model.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// GetByMerchantID 获取商户的所有API Key
func (r *apiKeyRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.APIKey, error) {
	var keys []*model.APIKey
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

// Delete 删除API Key（软删除）
func (r *apiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// GetByIDAndMerchantID 验证API Key是否属于指定商户
func (r *apiKeyRepository) GetByIDAndMerchantID(ctx context.Context, id uuid.UUID, merchantID uuid.UUID) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.WithContext(ctx).
		Where("id = ? AND merchant_id = ? AND is_active = ?", id, merchantID, true).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

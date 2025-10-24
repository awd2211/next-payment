package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKey API密钥模型（与merchant-service中的模型保持一致）
type APIKey struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	APIKey      string     `gorm:"type:varchar(64);unique;not null;index"`
	APISecret   string     `gorm:"type:varchar(128);not null"`
	Name        string     `gorm:"type:varchar(100)"`
	Environment string     `gorm:"type:varchar(20);not null;index"`
	IsActive    bool       `gorm:"default:true"`
	LastUsedAt  *time.Time `gorm:"type:timestamptz"`
	ExpiresAt   *time.Time `gorm:"type:timestamptz"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;default:now()"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;default:now()"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// APIKeyRepository API密钥仓储接口
type APIKeyRepository interface {
	// GetByAPIKey 根据API Key查询密钥信息
	GetByAPIKey(ctx context.Context, apiKey string) (*APIKey, error)
	// UpdateLastUsedAt 更新最后使用时间
	UpdateLastUsedAt(ctx context.Context, apiKey string) error
}

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository 创建API密钥仓储实例
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

// GetByAPIKey 根据API Key查询密钥信息
func (r *apiKeyRepository) GetByAPIKey(ctx context.Context, apiKey string) (*APIKey, error) {
	var key APIKey
	err := r.db.WithContext(ctx).
		Where("api_key = ?", apiKey).
		First(&key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to query API key: %w", err)
	}
	return &key, nil
}

// UpdateLastUsedAt 更新最后使用时间
func (r *apiKeyRepository) UpdateLastUsedAt(ctx context.Context, apiKey string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&APIKey{}).
		Where("api_key = ?", apiKey).
		Update("last_used_at", now).Error
	if err != nil {
		return fmt.Errorf("failed to update last_used_at: %w", err)
	}
	return nil
}

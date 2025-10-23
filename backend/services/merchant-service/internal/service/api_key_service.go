package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// APIKeyService API密钥服务接口
type APIKeyService interface {
	Create(ctx context.Context, merchantID uuid.UUID, input *CreateAPIKeyInput) (*model.APIKey, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.APIKey, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, environment string) ([]*model.APIKey, error)
	Update(ctx context.Context, id uuid.UUID, input *UpdateAPIKeyInput) (*model.APIKey, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	Rotate(ctx context.Context, id uuid.UUID) (*model.APIKey, error)
}

type apiKeyService struct {
	apiKeyRepo   repository.APIKeyRepository
	merchantRepo repository.MerchantRepository
}

// NewAPIKeyService 创建API密钥服务实例
func NewAPIKeyService(
	apiKeyRepo repository.APIKeyRepository,
	merchantRepo repository.MerchantRepository,
) APIKeyService {
	return &apiKeyService{
		apiKeyRepo:   apiKeyRepo,
		merchantRepo: merchantRepo,
	}
}

// CreateAPIKeyInput 创建API密钥输入
type CreateAPIKeyInput struct {
	Name        string     `json:"name" binding:"required"`
	Environment string     `json:"environment" binding:"required"` // test, production
	ExpiresAt   *time.Time `json:"expires_at"`                     // 可选的过期时间
}

// UpdateAPIKeyInput 更新API密钥输入
type UpdateAPIKeyInput struct {
	Name      *string    `json:"name"`
	IsActive  *bool      `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Create 创建API密钥
func (s *apiKeyService) Create(ctx context.Context, merchantID uuid.UUID, input *CreateAPIKeyInput) (*model.APIKey, error) {
	// 验证商户是否存在
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取商户失败: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("商户不存在")
	}

	// 验证环境类型
	if input.Environment != model.EnvironmentTest && input.Environment != model.EnvironmentProduction {
		return nil, fmt.Errorf("无效的环境类型")
	}

	// 生产环境需要KYC验证
	if input.Environment == model.EnvironmentProduction {
		if merchant.KYCStatus != model.KYCStatusVerified {
			return nil, fmt.Errorf("生产环境API密钥需要先完成KYC验证")
		}
		if merchant.Status != model.MerchantStatusActive {
			return nil, fmt.Errorf("商户状态必须为active才能创建生产环境API密钥")
		}
	}

	// 检查该商户在该环境下的API Key数量限制
	existing, err := s.apiKeyRepo.ListByMerchant(ctx, merchantID, input.Environment)
	if err != nil {
		return nil, fmt.Errorf("检查现有API密钥失败: %w", err)
	}

	// 限制每个环境最多5个API Key
	if len(existing) >= 5 {
		return nil, fmt.Errorf("该环境下API密钥数量已达上限（5个）")
	}

	// 生成API Key和Secret
	var prefix string
	if input.Environment == model.EnvironmentTest {
		prefix = "pk_test"
	} else {
		prefix = "pk_live"
	}

	apiKey := &model.APIKey{
		MerchantID:  merchantID,
		APIKey:      generateAPIKey(prefix),
		APISecret:   generateAPISecret(),
		Name:        input.Name,
		Environment: input.Environment,
		IsActive:    true,
		ExpiresAt:   input.ExpiresAt,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("创建API密钥失败: %w", err)
	}

	return apiKey, nil
}

// GetByID 根据ID获取API密钥
func (s *apiKeyService) GetByID(ctx context.Context, id uuid.UUID) (*model.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取API密钥失败: %w", err)
	}
	if apiKey == nil {
		return nil, fmt.Errorf("API密钥不存在")
	}
	return apiKey, nil
}

// GetByAPIKey 根据API Key获取记录
func (s *apiKeyService) GetByAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error) {
	key, err := s.apiKeyRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("获取API密钥失败: %w", err)
	}
	if key == nil {
		return nil, fmt.Errorf("API密钥不存在或已失效")
	}

	// 检查是否过期
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API密钥已过期")
	}

	// 更新最后使用时间
	go s.apiKeyRepo.UpdateLastUsed(context.Background(), key.ID)

	return key, nil
}

// ListByMerchant 获取商户的所有API密钥
func (s *apiKeyService) ListByMerchant(ctx context.Context, merchantID uuid.UUID, environment string) ([]*model.APIKey, error) {
	keys, err := s.apiKeyRepo.ListByMerchant(ctx, merchantID, environment)
	if err != nil {
		return nil, fmt.Errorf("获取API密钥列表失败: %w", err)
	}

	// 隐藏API Secret（只在创建时返回）
	for _, key := range keys {
		key.APISecret = "sk_****"
	}

	return keys, nil
}

// Update 更新API密钥
func (s *apiKeyService) Update(ctx context.Context, id uuid.UUID, input *UpdateAPIKeyInput) (*model.APIKey, error) {
	apiKey, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		apiKey.Name = *input.Name
	}
	if input.IsActive != nil {
		apiKey.IsActive = *input.IsActive
	}
	if input.ExpiresAt != nil {
		apiKey.ExpiresAt = input.ExpiresAt
	}

	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("更新API密钥失败: %w", err)
	}

	// 隐藏Secret
	apiKey.APISecret = "sk_****"

	return apiKey, nil
}

// Revoke 撤销API密钥
func (s *apiKeyService) Revoke(ctx context.Context, id uuid.UUID) error {
	// 检查是否存在
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.apiKeyRepo.Revoke(ctx, id); err != nil {
		return fmt.Errorf("撤销API密钥失败: %w", err)
	}

	return nil
}

// Delete 删除API密钥
func (s *apiKeyService) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查是否存在
	apiKey, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 生产环境的活跃API Key不能直接删除，需要先撤销
	if apiKey.Environment == model.EnvironmentProduction && apiKey.IsActive {
		return fmt.Errorf("生产环境的活跃API密钥不能直接删除，请先撤销")
	}

	if err := s.apiKeyRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除API密钥失败: %w", err)
	}

	return nil
}

// Rotate 轮换API密钥（生成新的Secret）
func (s *apiKeyService) Rotate(ctx context.Context, id uuid.UUID) (*model.APIKey, error) {
	apiKey, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 生成新的Secret
	apiKey.APISecret = generateAPISecret()

	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("轮换API密钥失败: %w", err)
	}

	return apiKey, nil
}

// 工具函数

// generateAPIKey 生成API Key
func generateAPIKey(prefix string) string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, base64.URLEncoding.EncodeToString(b)[:43])
}

// generateAPISecret 生成API Secret
func generateAPISecret() string {
	b := make([]byte, 64)
	rand.Read(b)
	return fmt.Sprintf("sk_%s", base64.URLEncoding.EncodeToString(b)[:86])
}

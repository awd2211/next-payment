package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-auth-service/internal/model"
	"payment-platform/merchant-auth-service/internal/repository"
)

// APIKeyService API密钥服务接口
type APIKeyService interface {
	ValidateAPIKey(ctx context.Context, apiKey, signature, payload string) (*model.APIKey, error)
	CreateAPIKey(ctx context.Context, merchantID uuid.UUID, name, environment string) (*model.APIKey, string, error)
	ListAPIKeys(ctx context.Context, merchantID uuid.UUID) ([]*model.APIKey, error)
	DeleteAPIKey(ctx context.Context, merchantID uuid.UUID, keyID uuid.UUID) error
}

type apiKeyService struct {
	repo repository.APIKeyRepository
}

// NewAPIKeyService 创建API密钥服务
func NewAPIKeyService(repo repository.APIKeyRepository) APIKeyService {
	return &apiKeyService{repo: repo}
}

// ValidateAPIKey 验证API Key和签名
func (s *apiKeyService) ValidateAPIKey(ctx context.Context, apiKey, signature, payload string) (*model.APIKey, error) {
	// 1. 查询 API Key
	key, err := s.repo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, errors.New("invalid api key")
	}

	// 2. 检查是否过期
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, errors.New("api key expired")
	}

	// 3. 验证签名
	expectedSignature := computeHMAC(payload, key.APISecret)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, errors.New("invalid signature")
	}

	// 4. 更新最后使用时间（异步）
	go func() {
		ctx := context.Background()
		s.repo.UpdateLastUsedAt(ctx, key.ID)
	}()

	return key, nil
}

// CreateAPIKey 创建新的API Key
func (s *apiKeyService) CreateAPIKey(ctx context.Context, merchantID uuid.UUID, name, environment string) (*model.APIKey, string, error) {
	// 生成随机API Key和Secret
	apiKey := generateRandomKey(32)
	apiSecret := generateRandomKey(64)

	key := &model.APIKey{
		MerchantID:  merchantID,
		APIKey:      apiKey,
		APISecret:   apiSecret,
		Name:        name,
		Environment: environment,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return nil, "", err
	}

	// 返回明文 APISecret（仅此一次）
	return key, apiSecret, nil
}

// ListAPIKeys 获取商户的所有API Key
func (s *apiKeyService) ListAPIKeys(ctx context.Context, merchantID uuid.UUID) ([]*model.APIKey, error) {
	keys, err := s.repo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	// 隐藏 APISecret
	for _, key := range keys {
		key.APISecret = ""
	}

	return keys, nil
}

// DeleteAPIKey 删除API Key
func (s *apiKeyService) DeleteAPIKey(ctx context.Context, merchantID uuid.UUID, keyID uuid.UUID) error {
	// 验证API Key是否属于该商户 (安全检查)
	_, err := s.repo.GetByIDAndMerchantID(ctx, keyID, merchantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("API Key不存在或不属于该商户")
		}
		return fmt.Errorf("验证API Key归属失败: %w", err)
	}

	// 验证通过,执行删除
	return s.repo.Delete(ctx, keyID)
}

// computeHMAC 计算HMAC-SHA256签名
func computeHMAC(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

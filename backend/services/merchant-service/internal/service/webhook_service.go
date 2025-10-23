package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// WebhookService Webhook服务接口
type WebhookService interface {
	CreateWebhook(ctx context.Context, input *CreateWebhookInput) (*model.WebhookConfig, error)
	GetWebhook(ctx context.Context, merchantID uuid.UUID) (*model.WebhookConfig, error)
	UpdateWebhook(ctx context.Context, merchantID uuid.UUID, input *UpdateWebhookInput) (*model.WebhookConfig, error)
	DeleteWebhook(ctx context.Context, merchantID uuid.UUID) error
	RegenerateSecret(ctx context.Context, merchantID uuid.UUID) (*model.WebhookConfig, error)
}

// webhookService Webhook服务实现
type webhookService struct {
	webhookRepo repository.WebhookRepository
}

// NewWebhookService 创建Webhook服务实例
func NewWebhookService(webhookRepo repository.WebhookRepository) WebhookService {
	return &webhookService{
		webhookRepo: webhookRepo,
	}
}

// CreateWebhookInput 创建Webhook配置输入
type CreateWebhookInput struct {
	MerchantID     uuid.UUID `json:"merchant_id" binding:"required"`
	URL            string    `json:"url" binding:"required,url"`
	Events         []string  `json:"events" binding:"required,min=1"`
	MaxRetries     *int      `json:"max_retries"`
	TimeoutSeconds *int      `json:"timeout_seconds"`
}

// UpdateWebhookInput 更新Webhook配置输入
type UpdateWebhookInput struct {
	URL            *string   `json:"url" binding:"omitempty,url"`
	Events         *[]string `json:"events" binding:"omitempty,min=1"`
	IsEnabled      *bool     `json:"is_enabled"`
	MaxRetries     *int      `json:"max_retries"`
	TimeoutSeconds *int      `json:"timeout_seconds"`
}

// CreateWebhook 创建Webhook配置
func (s *webhookService) CreateWebhook(ctx context.Context, input *CreateWebhookInput) (*model.WebhookConfig, error) {
	// 检查是否已存在Webhook配置
	existing, err := s.webhookRepo.GetByMerchantID(ctx, input.MerchantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("该商户已存在Webhook配置，请先删除旧配置")
	}

	// 生成签名密钥
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, err
	}

	// Events转JSON
	eventsJSON, err := json.Marshal(input.Events)
	if err != nil {
		return nil, err
	}

	webhook := &model.WebhookConfig{
		MerchantID:     input.MerchantID,
		URL:            input.URL,
		Events:         string(eventsJSON),
		Secret:         secret,
		IsEnabled:      true,
		MaxRetries:     3,
		TimeoutSeconds: 30,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if input.MaxRetries != nil {
		webhook.MaxRetries = *input.MaxRetries
	}
	if input.TimeoutSeconds != nil {
		webhook.TimeoutSeconds = *input.TimeoutSeconds
	}

	if err := s.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

// GetWebhook 获取Webhook配置
func (s *webhookService) GetWebhook(ctx context.Context, merchantID uuid.UUID) (*model.WebhookConfig, error) {
	webhook, err := s.webhookRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if webhook == nil {
		return nil, errors.New("Webhook配置不存在")
	}
	return webhook, nil
}

// UpdateWebhook 更新Webhook配置
func (s *webhookService) UpdateWebhook(ctx context.Context, merchantID uuid.UUID, input *UpdateWebhookInput) (*model.WebhookConfig, error) {
	webhook, err := s.webhookRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if webhook == nil {
		return nil, errors.New("Webhook配置不存在")
	}

	// 更新字段
	if input.URL != nil {
		webhook.URL = *input.URL
	}
	if input.Events != nil {
		eventsJSON, err := json.Marshal(*input.Events)
		if err != nil {
			return nil, err
		}
		webhook.Events = string(eventsJSON)
	}
	if input.IsEnabled != nil {
		webhook.IsEnabled = *input.IsEnabled
	}
	if input.MaxRetries != nil {
		webhook.MaxRetries = *input.MaxRetries
	}
	if input.TimeoutSeconds != nil {
		webhook.TimeoutSeconds = *input.TimeoutSeconds
	}

	webhook.UpdatedAt = time.Now()

	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

// DeleteWebhook 删除Webhook配置
func (s *webhookService) DeleteWebhook(ctx context.Context, merchantID uuid.UUID) error {
	webhook, err := s.webhookRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return err
	}
	if webhook == nil {
		return errors.New("Webhook配置不存在")
	}

	return s.webhookRepo.Delete(ctx, webhook.ID)
}

// RegenerateSecret 重新生成签名密钥
func (s *webhookService) RegenerateSecret(ctx context.Context, merchantID uuid.UUID) (*model.WebhookConfig, error) {
	webhook, err := s.webhookRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if webhook == nil {
		return nil, errors.New("Webhook配置不存在")
	}

	// 生成新密钥
	newSecret, err := generateWebhookSecret()
	if err != nil {
		return nil, err
	}

	webhook.Secret = newSecret
	webhook.UpdatedAt = time.Now()

	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

// generateWebhookSecret 生成Webhook签名密钥（32字节）
func generateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(bytes), nil
}

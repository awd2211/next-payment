package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"payment-platform/merchant-config-service/internal/model"
	"payment-platform/merchant-config-service/internal/repository"
)

// ChannelConfigService 渠道配置服务接口
type ChannelConfigService interface {
	CreateChannelConfig(ctx context.Context, input *CreateChannelConfigInput) (*model.ChannelConfig, error)
	GetChannelConfig(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error)
	ListMerchantChannels(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error)
	GetMerchantChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error)
	UpdateChannelConfig(ctx context.Context, id uuid.UUID, input *UpdateChannelConfigInput) (*model.ChannelConfig, error)
	DeleteChannelConfig(ctx context.Context, id uuid.UUID) error
	EnableChannel(ctx context.Context, id uuid.UUID) error
	DisableChannel(ctx context.Context, id uuid.UUID) error
}

type channelConfigService struct {
	repo repository.ChannelConfigRepository
}

// NewChannelConfigService 创建渠道配置服务实例
func NewChannelConfigService(repo repository.ChannelConfigRepository) ChannelConfigService {
	return &channelConfigService{repo: repo}
}

// CreateChannelConfigInput 创建渠道配置输入
type CreateChannelConfigInput struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	Channel    string    `json:"channel"`
	Config     string    `json:"config"`     // JSON string
	IsTestMode bool      `json:"is_test_mode"`
}

// UpdateChannelConfigInput 更新渠道配置输入
type UpdateChannelConfigInput struct {
	Config     *string `json:"config"`
	IsEnabled  *bool   `json:"is_enabled"`
	IsTestMode *bool   `json:"is_test_mode"`
}

// CreateChannelConfig 创建渠道配置
func (s *channelConfigService) CreateChannelConfig(ctx context.Context, input *CreateChannelConfigInput) (*model.ChannelConfig, error) {
	if err := s.validateChannelConfig(input); err != nil {
		return nil, err
	}

	// 检查是否已存在该渠道配置
	existing, err := s.repo.GetByMerchantAndChannel(ctx, input.MerchantID, input.Channel)
	if err == nil && existing != nil {
		return nil, errors.New("channel config already exists for this merchant")
	}

	config := &model.ChannelConfig{
		MerchantID: input.MerchantID,
		Channel:    input.Channel,
		Config:     input.Config,
		IsEnabled:  false, // 默认未启用，需要手动启用
		IsTestMode: input.IsTestMode,
	}

	if err := s.repo.Create(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetChannelConfig 获取渠道配置
func (s *channelConfigService) GetChannelConfig(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMerchantChannels 列出商户的所有渠道配置
func (s *channelConfigService) ListMerchantChannels(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error) {
	return s.repo.GetByMerchantID(ctx, merchantID)
}

// GetMerchantChannel 获取商户指定渠道配置
func (s *channelConfigService) GetMerchantChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error) {
	return s.repo.GetByMerchantAndChannel(ctx, merchantID, channel)
}

// UpdateChannelConfig 更新渠道配置
func (s *channelConfigService) UpdateChannelConfig(ctx context.Context, id uuid.UUID, input *UpdateChannelConfigInput) (*model.ChannelConfig, error) {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Config != nil {
		config.Config = *input.Config
	}
	if input.IsEnabled != nil {
		config.IsEnabled = *input.IsEnabled
	}
	if input.IsTestMode != nil {
		config.IsTestMode = *input.IsTestMode
	}

	if err := s.repo.Update(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// DeleteChannelConfig 删除渠道配置
func (s *channelConfigService) DeleteChannelConfig(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// EnableChannel 启用渠道
func (s *channelConfigService) EnableChannel(ctx context.Context, id uuid.UUID) error {
	return s.repo.EnableChannel(ctx, id)
}

// DisableChannel 停用渠道
func (s *channelConfigService) DisableChannel(ctx context.Context, id uuid.UUID) error {
	return s.repo.DisableChannel(ctx, id)
}

// validateChannelConfig 验证渠道配置
func (s *channelConfigService) validateChannelConfig(input *CreateChannelConfigInput) error {
	if input.MerchantID == uuid.Nil {
		return errors.New("merchant_id is required")
	}
	if input.Channel == "" {
		return errors.New("channel is required")
	}
	if input.Config == "" {
		return errors.New("config is required")
	}

	// 验证渠道名称
	validChannels := map[string]bool{
		model.ChannelStripe: true,
		model.ChannelPayPal: true,
		model.ChannelCrypto: true,
		model.ChannelAdyen:  true,
		model.ChannelSquare: true,
	}

	if !validChannels[input.Channel] {
		return errors.New("invalid channel")
	}

	return nil
}

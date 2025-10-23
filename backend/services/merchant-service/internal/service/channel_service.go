package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// ChannelService 渠道服务接口
type ChannelService interface {
	CreateChannel(ctx context.Context, input *CreateChannelInput) (*model.ChannelConfig, error)
	GetChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID) (*model.ChannelConfig, error)
	ListChannels(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error)
	UpdateChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID, input *UpdateChannelInput) (*model.ChannelConfig, error)
	DeleteChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID) error
	ToggleChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID, enabled bool) error
}

// channelService 渠道服务实现
type channelService struct {
	channelRepo repository.ChannelRepository
}

// NewChannelService 创建渠道服务实例
func NewChannelService(channelRepo repository.ChannelRepository) ChannelService {
	return &channelService{
		channelRepo: channelRepo,
	}
}

// CreateChannelInput 创建渠道配置输入
type CreateChannelInput struct {
	MerchantID uuid.UUID              `json:"merchant_id" binding:"required"`
	Channel    string                 `json:"channel" binding:"required,oneof=stripe paypal crypto adyen square"`
	Config     map[string]interface{} `json:"config" binding:"required"`
	IsTestMode bool                   `json:"is_test_mode"`
}

// UpdateChannelInput 更新渠道配置输入
type UpdateChannelInput struct {
	Config     *map[string]interface{} `json:"config"`
	IsEnabled  *bool                   `json:"is_enabled"`
	IsTestMode *bool                   `json:"is_test_mode"`
}

// CreateChannel 创建渠道配置
func (s *channelService) CreateChannel(ctx context.Context, input *CreateChannelInput) (*model.ChannelConfig, error) {
	// 验证渠道类型
	if !isValidChannel(input.Channel) {
		return nil, errors.New("无效的渠道类型")
	}

	// 检查是否已存在该渠道配置
	existing, err := s.channelRepo.GetByMerchantAndChannel(ctx, input.MerchantID, input.Channel)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("该渠道配置已存在")
	}

	// 验证渠道配置
	if err := validateChannelConfig(input.Channel, input.Config); err != nil {
		return nil, err
	}

	// 配置转JSON（后续可以加密敏感字段）
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, err
	}

	channel := &model.ChannelConfig{
		MerchantID: input.MerchantID,
		Channel:    input.Channel,
		Config:     string(configJSON),
		IsEnabled:  false, // 默认禁用，需要手动启用
		IsTestMode: input.IsTestMode,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// GetChannel 获取渠道配置
func (s *channelService) GetChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID) (*model.ChannelConfig, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.New("渠道配置不存在")
	}

	// 验证商户权限
	if channel.MerchantID != merchantID {
		return nil, errors.New("无权访问该渠道配置")
	}

	return channel, nil
}

// ListChannels 获取商户的所有渠道配置
func (s *channelService) ListChannels(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error) {
	return s.channelRepo.ListByMerchantID(ctx, merchantID)
}

// UpdateChannel 更新渠道配置
func (s *channelService) UpdateChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID, input *UpdateChannelInput) (*model.ChannelConfig, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.New("渠道配置不存在")
	}

	// 验证商户权限
	if channel.MerchantID != merchantID {
		return nil, errors.New("无权访问该渠道配置")
	}

	// 更新配置
	if input.Config != nil {
		// 验证渠道配置
		if err := validateChannelConfig(channel.Channel, *input.Config); err != nil {
			return nil, err
		}

		configJSON, err := json.Marshal(*input.Config)
		if err != nil {
			return nil, err
		}
		channel.Config = string(configJSON)
	}

	if input.IsEnabled != nil {
		channel.IsEnabled = *input.IsEnabled
	}

	if input.IsTestMode != nil {
		channel.IsTestMode = *input.IsTestMode
	}

	channel.UpdatedAt = time.Now()

	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// DeleteChannel 删除渠道配置
func (s *channelService) DeleteChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID) error {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return errors.New("渠道配置不存在")
	}

	// 验证商户权限
	if channel.MerchantID != merchantID {
		return errors.New("无权删除该渠道配置")
	}

	return s.channelRepo.Delete(ctx, channelID)
}

// ToggleChannel 启用/禁用渠道
func (s *channelService) ToggleChannel(ctx context.Context, channelID uuid.UUID, merchantID uuid.UUID, enabled bool) error {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return errors.New("渠道配置不存在")
	}

	// 验证商户权限
	if channel.MerchantID != merchantID {
		return errors.New("无权操作该渠道配置")
	}

	channel.IsEnabled = enabled
	channel.UpdatedAt = time.Now()

	return s.channelRepo.Update(ctx, channel)
}

// isValidChannel 验证渠道类型
func isValidChannel(channel string) bool {
	validChannels := []string{
		model.ChannelStripe,
		model.ChannelPayPal,
		model.ChannelCrypto,
		model.ChannelAdyen,
		model.ChannelSquare,
	}
	for _, c := range validChannels {
		if c == channel {
			return true
		}
	}
	return false
}

// validateChannelConfig 验证渠道配置（简单验证，实际应该更严格）
func validateChannelConfig(channel string, config map[string]interface{}) error {
	switch channel {
	case model.ChannelStripe:
		// Stripe需要: api_key, webhook_secret
		if _, ok := config["api_key"]; !ok {
			return errors.New("Stripe配置缺少 api_key")
		}
	case model.ChannelPayPal:
		// PayPal需要: client_id, client_secret
		if _, ok := config["client_id"]; !ok {
			return errors.New("PayPal配置缺少 client_id")
		}
		if _, ok := config["client_secret"]; !ok {
			return errors.New("PayPal配置缺少 client_secret")
		}
	case model.ChannelCrypto:
		// Crypto需要: wallet_address, network
		if _, ok := config["wallet_address"]; !ok {
			return errors.New("Crypto配置缺少 wallet_address")
		}
	}
	return nil
}

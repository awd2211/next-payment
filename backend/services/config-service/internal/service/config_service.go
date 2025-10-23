package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/google/uuid"
	"payment-platform/config-service/internal/model"
	"payment-platform/config-service/internal/repository"
)

// ConfigService 配置服务接口
type ConfigService interface {
	// 配置管理
	CreateConfig(ctx context.Context, input *CreateConfigInput) (*model.Config, error)
	GetConfig(ctx context.Context, serviceName, configKey, environment string) (*model.Config, error)
	ListConfigs(ctx context.Context, query *repository.ConfigQuery) ([]*model.Config, int64, error)
	UpdateConfig(ctx context.Context, id uuid.UUID, input *UpdateConfigInput) (*model.Config, error)
	DeleteConfig(ctx context.Context, id uuid.UUID, deletedBy string) error
	GetConfigHistory(ctx context.Context, configID uuid.UUID, limit int) ([]*model.ConfigHistory, error)

	// 功能开关
	CreateFeatureFlag(ctx context.Context, input *CreateFeatureFlagInput) (*model.FeatureFlag, error)
	GetFeatureFlag(ctx context.Context, flagKey string) (*model.FeatureFlag, error)
	ListFeatureFlags(ctx context.Context, query *repository.FeatureFlagQuery) ([]*model.FeatureFlag, int64, error)
	UpdateFeatureFlag(ctx context.Context, id uuid.UUID, input *UpdateFeatureFlagInput) (*model.FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error
	IsFeatureEnabled(ctx context.Context, flagKey string, context map[string]interface{}) (bool, error)

	// 服务注册
	RegisterService(ctx context.Context, input *RegisterServiceInput) (*model.ServiceRegistry, error)
	GetService(ctx context.Context, serviceName string) (*model.ServiceRegistry, error)
	ListServices(ctx context.Context) ([]*model.ServiceRegistry, error)
	UpdateServiceHeartbeat(ctx context.Context, serviceName string) error
	DeregisterService(ctx context.Context, serviceName string) error
}

type configService struct {
	configRepo     repository.ConfigRepository
	encryptionKey  string
}

// NewConfigService 创建配置服务实例
func NewConfigService(configRepo repository.ConfigRepository) ConfigService {
	return &configService{
		configRepo:    configRepo,
		encryptionKey: "default-encryption-key-change-me",
	}
}

// Input structures

type CreateConfigInput struct {
	ServiceName string `json:"service_name" binding:"required"`
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
	ValueType   string `json:"value_type"`
	Environment string `json:"environment"`
	Description string `json:"description"`
	IsEncrypted bool   `json:"is_encrypted"`
	CreatedBy   string `json:"created_by"`
}

type UpdateConfigInput struct {
	ConfigValue string `json:"config_value"`
	Description string `json:"description"`
	UpdatedBy   string `json:"updated_by"`
}

type CreateFeatureFlagInput struct {
	FlagKey     string                 `json:"flag_key" binding:"required"`
	FlagName    string                 `json:"flag_name" binding:"required"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Environment string                 `json:"environment"`
	Conditions  map[string]interface{} `json:"conditions"`
	Percentage  int                    `json:"percentage"`
	CreatedBy   string                 `json:"created_by"`
}

type UpdateFeatureFlagInput struct {
	FlagName    string                 `json:"flag_name"`
	Description string                 `json:"description"`
	Enabled     *bool                  `json:"enabled"`
	Conditions  map[string]interface{} `json:"conditions"`
	Percentage  int                    `json:"percentage"`
	UpdatedBy   string                 `json:"updated_by"`
}

type RegisterServiceInput struct {
	ServiceName string                 `json:"service_name" binding:"required"`
	ServiceURL  string                 `json:"service_url" binding:"required"`
	ServiceIP   string                 `json:"service_ip"`
	ServicePort int                    `json:"service_port"`
	HealthCheck string                 `json:"health_check"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Config Management

func (s *configService) CreateConfig(ctx context.Context, input *CreateConfigInput) (*model.Config, error) {
	// 检查配置是否已存在
	existing, err := s.configRepo.GetConfig(ctx, input.ServiceName, input.ConfigKey, input.Environment)
	if err != nil {
		return nil, fmt.Errorf("检查配置失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("配置已存在")
	}

	configValue := input.ConfigValue
	if input.IsEncrypted {
		encrypted, err := s.encrypt(configValue)
		if err != nil {
			return nil, fmt.Errorf("加密配置值失败: %w", err)
		}
		configValue = encrypted
	}

	config := &model.Config{
		ServiceName: input.ServiceName,
		ConfigKey:   input.ConfigKey,
		ConfigValue: configValue,
		ValueType:   input.ValueType,
		Environment: input.Environment,
		Description: input.Description,
		IsEncrypted: input.IsEncrypted,
		Version:     1,
		CreatedBy:   input.CreatedBy,
	}

	if err := s.configRepo.CreateConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("创建配置失败: %w", err)
	}

	return config, nil
}

func (s *configService) GetConfig(ctx context.Context, serviceName, configKey, environment string) (*model.Config, error) {
	config, err := s.configRepo.GetConfig(ctx, serviceName, configKey, environment)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("配置不存在")
	}

	// 如果配置是加密的，解密后返回
	if config.IsEncrypted {
		decrypted, err := s.decrypt(config.ConfigValue)
		if err != nil {
			return nil, fmt.Errorf("解密配置值失败: %w", err)
		}
		config.ConfigValue = decrypted
	}

	return config, nil
}

func (s *configService) ListConfigs(ctx context.Context, query *repository.ConfigQuery) ([]*model.Config, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.configRepo.ListConfigs(ctx, query)
}

func (s *configService) UpdateConfig(ctx context.Context, id uuid.UUID, input *UpdateConfigInput) (*model.Config, error) {
	config, err := s.configRepo.GetConfigByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("配置不存在")
	}

	// 记录历史
	history := &model.ConfigHistory{
		ConfigID:    config.ID,
		ServiceName: config.ServiceName,
		ConfigKey:   config.ConfigKey,
		OldValue:    config.ConfigValue,
		NewValue:    input.ConfigValue,
		Version:     config.Version,
		ChangedBy:   input.UpdatedBy,
		ChangeType:  "update",
	}
	if err := s.configRepo.CreateConfigHistory(ctx, history); err != nil {
		return nil, fmt.Errorf("记录配置历史失败: %w", err)
	}

	// 更新配置
	if input.ConfigValue != "" {
		if config.IsEncrypted {
			encrypted, err := s.encrypt(input.ConfigValue)
			if err != nil {
				return nil, fmt.Errorf("加密配置值失败: %w", err)
			}
			config.ConfigValue = encrypted
		} else {
			config.ConfigValue = input.ConfigValue
		}
		config.Version++
	}
	if input.Description != "" {
		config.Description = input.Description
	}
	config.UpdatedBy = input.UpdatedBy

	if err := s.configRepo.UpdateConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("更新配置失败: %w", err)
	}

	return config, nil
}

func (s *configService) DeleteConfig(ctx context.Context, id uuid.UUID, deletedBy string) error {
	config, err := s.configRepo.GetConfigByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}
	if config == nil {
		return fmt.Errorf("配置不存在")
	}

	// 记录历史
	history := &model.ConfigHistory{
		ConfigID:    config.ID,
		ServiceName: config.ServiceName,
		ConfigKey:   config.ConfigKey,
		OldValue:    config.ConfigValue,
		Version:     config.Version,
		ChangedBy:   deletedBy,
		ChangeType:  "delete",
	}
	if err := s.configRepo.CreateConfigHistory(ctx, history); err != nil {
		return fmt.Errorf("记录配置历史失败: %w", err)
	}

	return s.configRepo.DeleteConfig(ctx, id)
}

func (s *configService) GetConfigHistory(ctx context.Context, configID uuid.UUID, limit int) ([]*model.ConfigHistory, error) {
	return s.configRepo.ListConfigHistory(ctx, configID, limit)
}

// Feature Flags

func (s *configService) CreateFeatureFlag(ctx context.Context, input *CreateFeatureFlagInput) (*model.FeatureFlag, error) {
	// 检查是否已存在
	existing, _ := s.configRepo.GetFeatureFlagByKey(ctx, input.FlagKey)
	if existing != nil {
		return nil, fmt.Errorf("功能开关已存在")
	}

	flag := &model.FeatureFlag{
		FlagKey:     input.FlagKey,
		FlagName:    input.FlagName,
		Description: input.Description,
		Enabled:     input.Enabled,
		Environment: input.Environment,
		Conditions:  input.Conditions,
		Percentage:  input.Percentage,
		CreatedBy:   input.CreatedBy,
	}

	if err := s.configRepo.CreateFeatureFlag(ctx, flag); err != nil {
		return nil, fmt.Errorf("创建功能开关失败: %w", err)
	}

	return flag, nil
}

func (s *configService) GetFeatureFlag(ctx context.Context, flagKey string) (*model.FeatureFlag, error) {
	flag, err := s.configRepo.GetFeatureFlagByKey(ctx, flagKey)
	if err != nil {
		return nil, fmt.Errorf("获取功能开关失败: %w", err)
	}
	if flag == nil {
		return nil, fmt.Errorf("功能开关不存在")
	}
	return flag, nil
}

func (s *configService) ListFeatureFlags(ctx context.Context, query *repository.FeatureFlagQuery) ([]*model.FeatureFlag, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.configRepo.ListFeatureFlags(ctx, query)
}

func (s *configService) UpdateFeatureFlag(ctx context.Context, id uuid.UUID, input *UpdateFeatureFlagInput) (*model.FeatureFlag, error) {
	// TODO: Implement update logic
	return nil, nil
}

func (s *configService) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	return s.configRepo.DeleteFeatureFlag(ctx, id)
}

func (s *configService) IsFeatureEnabled(ctx context.Context, flagKey string, context map[string]interface{}) (bool, error) {
	flag, err := s.configRepo.GetFeatureFlagByKey(ctx, flagKey)
	if err != nil {
		return false, err
	}
	if flag == nil {
		return false, nil
	}

	if !flag.Enabled {
		return false, nil
	}

	// TODO: 实现条件判断和百分比控制
	return true, nil
}

// Service Registry

func (s *configService) RegisterService(ctx context.Context, input *RegisterServiceInput) (*model.ServiceRegistry, error) {
	service := &model.ServiceRegistry{
		ServiceName: input.ServiceName,
		ServiceURL:  input.ServiceURL,
		ServiceIP:   input.ServiceIP,
		ServicePort: input.ServicePort,
		Status:      "active",
		HealthCheck: input.HealthCheck,
		Metadata:    input.Metadata,
	}

	if err := s.configRepo.RegisterService(ctx, service); err != nil {
		return nil, fmt.Errorf("注册服务失败: %w", err)
	}

	return service, nil
}

func (s *configService) GetService(ctx context.Context, serviceName string) (*model.ServiceRegistry, error) {
	service, err := s.configRepo.GetService(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("获取服务失败: %w", err)
	}
	if service == nil {
		return nil, fmt.Errorf("服务不存在")
	}
	return service, nil
}

func (s *configService) ListServices(ctx context.Context) ([]*model.ServiceRegistry, error) {
	return s.configRepo.ListServices(ctx)
}

func (s *configService) UpdateServiceHeartbeat(ctx context.Context, serviceName string) error {
	return s.configRepo.UpdateServiceHeartbeat(ctx, serviceName)
}

func (s *configService) DeregisterService(ctx context.Context, serviceName string) error {
	return s.configRepo.DeregisterService(ctx, serviceName)
}

// Encryption helpers

func (s *configService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(s.encryptionKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *configService) decrypt(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(s.encryptionKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

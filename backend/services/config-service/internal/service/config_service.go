package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"payment-platform/config-service/internal/model"
	"payment-platform/config-service/internal/repository"
)

// ConfigService 配置服务接口
type ConfigService interface {
	// 配置管理
	CreateConfig(ctx context.Context, input *CreateConfigInput) (*model.Config, error)
	GetConfig(ctx context.Context, serviceName, configKey, environment string) (*model.Config, error)
	GetConfigByID(ctx context.Context, id uuid.UUID) (*model.Config, error)
	ListConfigs(ctx context.Context, query *repository.ConfigQuery) ([]*model.Config, int64, error)
	UpdateConfig(ctx context.Context, id uuid.UUID, input *UpdateConfigInput) (*model.Config, error)
	DeleteConfig(ctx context.Context, id uuid.UUID, deletedBy string) error
	GetConfigHistory(ctx context.Context, configID uuid.UUID, limit int) ([]*model.ConfigHistory, error)
	RollbackConfig(ctx context.Context, configID uuid.UUID, targetVersion int, rolledBy, reason string) (*model.Config, error)

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
	redisClient    *redis.Client
}

// NewConfigService 创建配置服务实例
func NewConfigService(configRepo repository.ConfigRepository, redisClient *redis.Client) ConfigService {
	return &configService{
		configRepo:    configRepo,
		encryptionKey: "default-encryption-key-change-me",
		redisClient:   redisClient,
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
	ConfigValue  string `json:"config_value"`
	Description  string `json:"description"`
	UpdatedBy    string `json:"updated_by"`
	ChangeReason string `json:"change_reason"` // 变更原因
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
	// 【缓存优化】1. 先查 Redis 缓存
	cacheKey := fmt.Sprintf("config:%s:%s:%s", serviceName, configKey, environment)

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			// 缓存命中，反序列化返回
			var config model.Config
			if err := json.Unmarshal([]byte(cached), &config); err == nil {
				logger.Debug("Config cache hit",
					zap.String("service", serviceName),
					zap.String("key", configKey))
				return &config, nil
			}
		} else if err != redis.Nil {
			// Redis 错误，记录日志但继续查询数据库
			logger.Warn("Redis get failed, fallback to DB",
				zap.Error(err),
				zap.String("cache_key", cacheKey))
		}
	}

	// 【缓存优化】2. 缓存未命中，查询数据库
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

	// 【缓存优化】3. 写入缓存 (10分钟TTL)
	if s.redisClient != nil {
		data, err := json.Marshal(config)
		if err == nil {
			// 设置缓存，失败不影响业务
			if err := s.redisClient.Set(ctx, cacheKey, data, 10*time.Minute).Err(); err != nil {
				logger.Warn("Failed to set config cache",
					zap.Error(err),
					zap.String("cache_key", cacheKey))
			} else {
				logger.Debug("Config cached",
					zap.String("service", serviceName),
					zap.String("key", configKey))
			}
		}
	}

	return config, nil
}

func (s *configService) GetConfigByID(ctx context.Context, id uuid.UUID) (*model.Config, error) {
	config, err := s.configRepo.GetConfigByID(ctx, id)
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
		ConfigID:     config.ID,
		ServiceName:  config.ServiceName,
		ConfigKey:    config.ConfigKey,
		OldValue:     config.ConfigValue,
		NewValue:     input.ConfigValue,
		Version:      config.Version,
		ChangedBy:    input.UpdatedBy,
		ChangeType:   "update",
		ChangeReason: input.ChangeReason,
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

	// 【缓存失效】更新成功后删除缓存
	s.invalidateConfigCache(ctx, config.ServiceName, config.ConfigKey, config.Environment)

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

	if err := s.configRepo.DeleteConfig(ctx, id); err != nil {
		return err
	}

	// 【缓存失效】删除成功后清除缓存
	s.invalidateConfigCache(ctx, config.ServiceName, config.ConfigKey, config.Environment)

	return nil
}

func (s *configService) GetConfigHistory(ctx context.Context, configID uuid.UUID, limit int) ([]*model.ConfigHistory, error) {
	return s.configRepo.ListConfigHistory(ctx, configID, limit)
}

func (s *configService) RollbackConfig(ctx context.Context, configID uuid.UUID, targetVersion int, rolledBy, reason string) (*model.Config, error) {
	// 1. 获取当前配置
	config, err := s.configRepo.GetConfigByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("配置不存在")
	}

	// 2. 查询历史记录，找到目标版本
	histories, err := s.configRepo.ListConfigHistory(ctx, configID, 100)
	if err != nil {
		return nil, fmt.Errorf("查询配置历史失败: %w", err)
	}

	var targetHistory *model.ConfigHistory
	for _, h := range histories {
		if h.Version == targetVersion {
			targetHistory = h
			break
		}
	}

	if targetHistory == nil {
		return nil, fmt.Errorf("未找到版本 %d 的历史记录", targetVersion)
	}

	// 3. 记录回滚操作的历史
	rollbackHistory := &model.ConfigHistory{
		ConfigID:     config.ID,
		ServiceName:  config.ServiceName,
		ConfigKey:    config.ConfigKey,
		OldValue:     config.ConfigValue,
		NewValue:     targetHistory.NewValue, // 回滚到目标版本的值
		Version:      config.Version,
		ChangedBy:    rolledBy,
		ChangeType:   "rollback",
		ChangeReason: fmt.Sprintf("回滚到版本 %d: %s", targetVersion, reason),
	}
	if err := s.configRepo.CreateConfigHistory(ctx, rollbackHistory); err != nil {
		return nil, fmt.Errorf("记录回滚历史失败: %w", err)
	}

	// 4. 更新配置到目标版本的值
	if config.IsEncrypted {
		encrypted, err := s.encrypt(targetHistory.NewValue)
		if err != nil {
			return nil, fmt.Errorf("加密配置值失败: %w", err)
		}
		config.ConfigValue = encrypted
	} else {
		config.ConfigValue = targetHistory.NewValue
	}
	config.Version++
	config.UpdatedBy = rolledBy

	if err := s.configRepo.UpdateConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("回滚配置失败: %w", err)
	}

	// 【缓存失效】回滚成功后清除缓存
	s.invalidateConfigCache(ctx, config.ServiceName, config.ConfigKey, config.Environment)

	return config, nil
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
	// 先查询功能开关
	flag, err := s.configRepo.GetFeatureFlagByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取功能开关失败: %w", err)
	}
	if flag == nil {
		return nil, fmt.Errorf("功能开关不存在")
	}

	// 更新字段
	if input.FlagName != "" {
		flag.FlagName = input.FlagName
	}
	if input.Description != "" {
		flag.Description = input.Description
	}
	if input.Conditions != nil {
		flag.Conditions = input.Conditions
	}
	if input.Percentage > 0 {
		flag.Percentage = input.Percentage
	}
	if input.Enabled != nil {
		flag.Enabled = *input.Enabled
	}
	if input.UpdatedBy != "" {
		flag.UpdatedBy = input.UpdatedBy
	}

	if err := s.configRepo.UpdateFeatureFlag(ctx, flag); err != nil {
		return nil, fmt.Errorf("更新功能开关失败: %w", err)
	}

	return flag, nil
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

	// 1. 百分比控制（灰度发布）
	if flag.Percentage > 0 && flag.Percentage < 100 {
		// 使用用户ID或其他唯一标识计算hash，确保同一用户总是得到相同结果
		var hashKey string
		if userID, ok := context["user_id"].(string); ok && userID != "" {
			hashKey = userID
		} else if merchantID, ok := context["merchant_id"].(string); ok && merchantID != "" {
			hashKey = merchantID
		} else {
			// 如果没有唯一标识，随机决定
			hashKey = flagKey
		}

		// 简单的hash算法：将字符串转为数字
		hash := 0
		for _, c := range hashKey {
			hash = (hash*31 + int(c)) % 100
		}

		// 如果hash值大于百分比，不启用
		if hash >= flag.Percentage {
			return false, nil
		}
	}

	// 2. 条件判断（白名单、黑名单等）
	if flag.Conditions != nil && len(flag.Conditions) > 0 {
		// 白名单检查
		if whitelist, ok := flag.Conditions["whitelist"].([]interface{}); ok {
			isInWhitelist := false
			userID, hasUserID := context["user_id"].(string)
			merchantID, hasMerchantID := context["merchant_id"].(string)

			for _, item := range whitelist {
				itemStr, ok := item.(string)
				if !ok {
					continue
				}
				if (hasUserID && userID == itemStr) || (hasMerchantID && merchantID == itemStr) {
					isInWhitelist = true
					break
				}
			}

			// 如果有白名单但不在白名单中，不启用
			if !isInWhitelist {
				return false, nil
			}
		}

		// 黑名单检查
		if blacklist, ok := flag.Conditions["blacklist"].([]interface{}); ok {
			userID, hasUserID := context["user_id"].(string)
			merchantID, hasMerchantID := context["merchant_id"].(string)

			for _, item := range blacklist {
				itemStr, ok := item.(string)
				if !ok {
					continue
				}
				if (hasUserID && userID == itemStr) || (hasMerchantID && merchantID == itemStr) {
					// 在黑名单中，不启用
					return false, nil
				}
			}
		}

		// 地区限制
		if regions, ok := flag.Conditions["regions"].([]interface{}); ok && len(regions) > 0 {
			region, hasRegion := context["region"].(string)
			if !hasRegion {
				return false, nil
			}

			isInRegion := false
			for _, r := range regions {
				regionStr, ok := r.(string)
				if ok && regionStr == region {
					isInRegion = true
					break
				}
			}

			if !isInRegion {
				return false, nil
			}
		}
	}

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

// 【缓存优化】缓存失效辅助方法
func (s *configService) invalidateConfigCache(ctx context.Context, serviceName, configKey, environment string) {
	if s.redisClient == nil {
		return
	}

	cacheKey := fmt.Sprintf("config:%s:%s:%s", serviceName, configKey, environment)
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		logger.Warn("删除配置缓存失败", zap.String("cache_key", cacheKey), zap.Error(err))
	} else {
		logger.Info("配置缓存已失效", zap.String("cache_key", cacheKey))
	}
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"payment-platform/config-service/internal/model"
	"gorm.io/gorm"
)

// ConfigRepository 配置仓储接口
type ConfigRepository interface {
	// 配置管理
	CreateConfig(ctx context.Context, config *model.Config) error
	GetConfigByID(ctx context.Context, id uuid.UUID) (*model.Config, error)
	GetConfig(ctx context.Context, serviceName, configKey, environment string) (*model.Config, error)
	ListConfigs(ctx context.Context, query *ConfigQuery) ([]*model.Config, int64, error)
	UpdateConfig(ctx context.Context, config *model.Config) error
	DeleteConfig(ctx context.Context, id uuid.UUID) error

	// 配置历史
	CreateConfigHistory(ctx context.Context, history *model.ConfigHistory) error
	ListConfigHistory(ctx context.Context, configID uuid.UUID, limit int) ([]*model.ConfigHistory, error)

	// 功能开关
	CreateFeatureFlag(ctx context.Context, flag *model.FeatureFlag) error
	GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*model.FeatureFlag, error)
	GetFeatureFlagByKey(ctx context.Context, flagKey string) (*model.FeatureFlag, error)
	ListFeatureFlags(ctx context.Context, query *FeatureFlagQuery) ([]*model.FeatureFlag, int64, error)
	UpdateFeatureFlag(ctx context.Context, flag *model.FeatureFlag) error
	DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error

	// 服务注册
	RegisterService(ctx context.Context, service *model.ServiceRegistry) error
	GetService(ctx context.Context, serviceName string) (*model.ServiceRegistry, error)
	ListServices(ctx context.Context) ([]*model.ServiceRegistry, error)
	UpdateServiceHeartbeat(ctx context.Context, serviceName string) error
	DeregisterService(ctx context.Context, serviceName string) error
}

type configRepository struct {
	db *gorm.DB
}

// NewConfigRepository 创建配置仓储实例
func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

// ConfigQuery 配置查询条件
type ConfigQuery struct {
	ServiceName string
	Environment string
	ConfigKey   string
	Page        int
	PageSize    int
}

// FeatureFlagQuery 功能开关查询条件
type FeatureFlagQuery struct {
	Environment string
	Enabled     *bool
	Page        int
	PageSize    int
}

// Config Management

func (r *configRepository) CreateConfig(ctx context.Context, config *model.Config) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *configRepository) GetConfigByID(ctx context.Context, id uuid.UUID) (*model.Config, error) {
	var config model.Config
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &config, err
}

func (r *configRepository) GetConfig(ctx context.Context, serviceName, configKey, environment string) (*model.Config, error) {
	var config model.Config
	err := r.db.WithContext(ctx).
		Where("service_name = ? AND config_key = ? AND environment = ?", serviceName, configKey, environment).
		First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &config, err
}

func (r *configRepository) ListConfigs(ctx context.Context, query *ConfigQuery) ([]*model.Config, int64, error) {
	var configs []*model.Config
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Config{})

	if query.ServiceName != "" {
		db = db.Where("service_name = ?", query.ServiceName)
	}
	if query.Environment != "" {
		db = db.Where("environment = ?", query.Environment)
	}
	if query.ConfigKey != "" {
		db = db.Where("config_key LIKE ?", "%"+query.ConfigKey+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&configs).Error
	return configs, total, err
}

func (r *configRepository) UpdateConfig(ctx context.Context, config *model.Config) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *configRepository) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Config{}, "id = ?", id).Error
}

// Config History

func (r *configRepository) CreateConfigHistory(ctx context.Context, history *model.ConfigHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *configRepository) ListConfigHistory(ctx context.Context, configID uuid.UUID, limit int) ([]*model.ConfigHistory, error) {
	var history []*model.ConfigHistory
	err := r.db.WithContext(ctx).
		Where("config_id = ?", configID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// Feature Flags

func (r *configRepository) CreateFeatureFlag(ctx context.Context, flag *model.FeatureFlag) error {
	return r.db.WithContext(ctx).Create(flag).Error
}

func (r *configRepository) GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*model.FeatureFlag, error) {
	var flag model.FeatureFlag
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&flag).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &flag, err
}

func (r *configRepository) GetFeatureFlagByKey(ctx context.Context, flagKey string) (*model.FeatureFlag, error) {
	var flag model.FeatureFlag
	err := r.db.WithContext(ctx).Where("flag_key = ?", flagKey).First(&flag).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &flag, err
}

func (r *configRepository) ListFeatureFlags(ctx context.Context, query *FeatureFlagQuery) ([]*model.FeatureFlag, int64, error) {
	var flags []*model.FeatureFlag
	var total int64

	db := r.db.WithContext(ctx).Model(&model.FeatureFlag{})

	if query.Environment != "" {
		db = db.Where("environment = ?", query.Environment)
	}
	if query.Enabled != nil {
		db = db.Where("enabled = ?", *query.Enabled)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&flags).Error
	return flags, total, err
}

func (r *configRepository) UpdateFeatureFlag(ctx context.Context, flag *model.FeatureFlag) error {
	return r.db.WithContext(ctx).Save(flag).Error
}

func (r *configRepository) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.FeatureFlag{}, "id = ?", id).Error
}

// Service Registry

func (r *configRepository) RegisterService(ctx context.Context, service *model.ServiceRegistry) error {
	return r.db.WithContext(ctx).Create(service).Error
}

func (r *configRepository) GetService(ctx context.Context, serviceName string) (*model.ServiceRegistry, error) {
	var service model.ServiceRegistry
	err := r.db.WithContext(ctx).
		Where("service_name = ? AND status = ?", serviceName, "active").
		First(&service).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &service, err
}

func (r *configRepository) ListServices(ctx context.Context) ([]*model.ServiceRegistry, error) {
	var services []*model.ServiceRegistry
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("service_name ASC").
		Find(&services).Error
	return services, err
}

func (r *configRepository) UpdateServiceHeartbeat(ctx context.Context, serviceName string) error {
	return r.db.WithContext(ctx).
		Model(&model.ServiceRegistry{}).
		Where("service_name = ?", serviceName).
		Update("last_heartbeat", gorm.Expr("NOW()")).Error
}

func (r *configRepository) DeregisterService(ctx context.Context, serviceName string) error {
	return r.db.WithContext(ctx).
		Model(&model.ServiceRegistry{}).
		Where("service_name = ?", serviceName).
		Update("status", "inactive").Error
}

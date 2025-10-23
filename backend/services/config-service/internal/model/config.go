package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Config 配置项
type Config struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceName string         `gorm:"type:varchar(100);not null;index:idx_config_service" json:"service_name"`
	ConfigKey   string         `gorm:"type:varchar(255);not null;index:idx_config_key" json:"config_key"`
	ConfigValue string         `gorm:"type:text" json:"config_value"`
	ValueType   string         `gorm:"type:varchar(50);default:'string'" json:"value_type"`
	Environment string         `gorm:"type:varchar(50);default:'production'" json:"environment"`
	Description string         `gorm:"type:text" json:"description"`
	IsEncrypted bool           `gorm:"default:false" json:"is_encrypted"`
	Version     int            `gorm:"default:1" json:"version"`
	CreatedBy   string         `gorm:"type:varchar(100)" json:"created_by"`
	UpdatedBy   string         `gorm:"type:varchar(100)" json:"updated_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Config) TableName() string {
	return "configs"
}

// ConfigHistory 配置历史
type ConfigHistory struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ConfigID     uuid.UUID `gorm:"type:uuid;not null;index:idx_config_history_config" json:"config_id"`
	ServiceName  string    `gorm:"type:varchar(100);not null" json:"service_name"`
	ConfigKey    string    `gorm:"type:varchar(255);not null" json:"config_key"`
	OldValue     string    `gorm:"type:text" json:"old_value"`
	NewValue     string    `gorm:"type:text" json:"new_value"`
	Version      int       `gorm:"default:1" json:"version"`
	ChangedBy    string    `gorm:"type:varchar(100)" json:"changed_by"`
	ChangeType   string    `gorm:"type:varchar(50)" json:"change_type"` // create, update, delete, rollback
	ChangeReason string    `gorm:"type:text" json:"change_reason"`      // 变更原因
	CreatedAt    time.Time `json:"created_at"`
}

func (ConfigHistory) TableName() string {
	return "config_histories"
}

// FeatureFlag 功能开关
type FeatureFlag struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FlagKey     string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_feature_flag_key" json:"flag_key"`
	FlagName    string         `gorm:"type:varchar(255);not null" json:"flag_name"`
	Description string         `gorm:"type:text" json:"description"`
	Enabled     bool           `gorm:"default:false" json:"enabled"`
	Environment string         `gorm:"type:varchar(50);default:'production'" json:"environment"`
	Conditions  map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"conditions,omitempty"`
	Percentage  int            `gorm:"default:0" json:"percentage"`
	CreatedBy   string         `gorm:"type:varchar(100)" json:"created_by"`
	UpdatedBy   string         `gorm:"type:varchar(100)" json:"updated_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FeatureFlag) TableName() string {
	return "feature_flags"
}

// ServiceRegistry 服务注册
type ServiceRegistry struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceName string         `gorm:"type:varchar(100);not null;index:idx_service_registry_name" json:"service_name"`
	ServiceURL  string         `gorm:"type:varchar(500);not null" json:"service_url"`
	ServiceIP   string         `gorm:"type:varchar(50)" json:"service_ip"`
	ServicePort int            `gorm:"default:0" json:"service_port"`
	Status      string         `gorm:"type:varchar(50);default:'active'" json:"status"`
	HealthCheck string         `gorm:"type:varchar(500)" json:"health_check"`
	Metadata    map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"metadata,omitempty"`
	LastHeartbeat time.Time    `json:"last_heartbeat"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ServiceRegistry) TableName() string {
	return "service_registries"
}

package model

import (
	"time"

	"github.com/google/uuid"
)

// APIKey API密钥表
type APIKey struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"merchant_id"`
	APIKey      string     `gorm:"type:varchar(64);unique;not null;index" json:"api_key"`         // API Key（公开）
	APISecret   string     `gorm:"type:varchar(128);not null" json:"api_secret,omitempty"`        // API Secret（仅创建时返回）
	Name        string     `gorm:"type:varchar(100)" json:"name"`                                 // 密钥名称
	Environment string     `gorm:"type:varchar(20);not null;index" json:"environment"`            // test, production
	IsActive    bool       `gorm:"default:true" json:"is_active"`                                 // 是否启用
	LastUsedAt  *time.Time `gorm:"type:timestamptz" json:"last_used_at"`                          // 最后使用时间
	ExpiresAt   *time.Time `gorm:"type:timestamptz" json:"expires_at"`                            // 过期时间（null表示永不过期）
	CreatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// API Key环境常量
const (
	EnvironmentTest       = "test"       // 测试环境
	EnvironmentProduction = "production" // 生产环境
)

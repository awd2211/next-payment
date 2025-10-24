package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChannelConfig 支付渠道配置表
type ChannelConfig struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID      `gorm:"type:uuid;not null;index:idx_merchant_channel" json:"merchant_id"`
	Channel    string         `gorm:"type:varchar(50);not null;index:idx_merchant_channel" json:"channel"` // stripe, paypal, crypto
	Config     string         `gorm:"type:jsonb;not null" json:"config"`                                   // 渠道配置（JSON，加密存储敏感信息）
	IsEnabled  bool           `gorm:"default:false" json:"is_enabled"`                                     // 是否启用
	IsTestMode bool           `gorm:"default:true" json:"is_test_mode"`                                    // 是否测试模式
	CreatedAt  time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ChannelConfig) TableName() string {
	return "channel_configs"
}

// 唯一索引：一个商户同一渠道只能有一个配置
// CREATE UNIQUE INDEX idx_merchant_channel ON channel_configs(merchant_id, channel) WHERE deleted_at IS NULL;

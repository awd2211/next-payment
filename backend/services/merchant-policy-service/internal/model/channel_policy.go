package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChannelPolicy 渠道策略表 (商户支持的支付渠道配置)
type ChannelPolicy struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index" json:"merchant_id"`
	Channel    string    `gorm:"type:varchar(50);not null;index" json:"channel"` // stripe, paypal, crypto

	// 状态
	IsEnabled bool `gorm:"default:true" json:"is_enabled"`

	// 优先级 (数字越大优先级越高)
	Priority int `gorm:"default:0" json:"priority"`

	// 渠道特定配置 (JSON格式)
	Config string `gorm:"type:jsonb" json:"config,omitempty"`
	// 示例:
	// {
	//   "stripe": {
	//     "account_id": "acct_xxx",
	//     "webhook_secret": "whsec_xxx",
	//     "statement_descriptor": "MY COMPANY"
	//   },
	//   "paypal": {
	//     "client_id": "xxx",
	//     "mode": "live"
	//   }
	// }

	// 费率覆盖 (可选,覆盖全局费率策略)
	CustomFeePercentage *float64 `gorm:"type:decimal(5,4)" json:"custom_fee_percentage,omitempty"`
	CustomFeeFixed      *int64   `gorm:"type:bigint" json:"custom_fee_fixed,omitempty"`

	// 生效时间
	EffectiveDate time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"`
	ExpiryDate    *time.Time `gorm:"type:timestamptz" json:"expiry_date,omitempty"`

	// 审批流程
	CreatedBy  *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	ApprovedBy *uuid.UUID `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt *time.Time `gorm:"type:timestamptz" json:"approved_at,omitempty"`

	// 扩展信息
	Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"`
	Remarks  string `gorm:"type:text" json:"remarks,omitempty"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ChannelPolicy) TableName() string {
	return "channel_policies"
}

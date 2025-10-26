package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantTier 商户等级表 (预定义等级体系)
type MerchantTier struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TierCode    string    `gorm:"type:varchar(50);unique;not null;index" json:"tier_code"`
	TierName    string    `gorm:"type:varchar(100);not null" json:"tier_name"`
	TierLevel   int       `gorm:"type:integer;not null;index" json:"tier_level"` // 1-5 (1=最低, 5=最高)
	Description string    `gorm:"type:text" json:"description"`

	// 默认策略引用 (关联默认费率和限额策略ID)
	DefaultFeePolicyID   *uuid.UUID `gorm:"type:uuid" json:"default_fee_policy_id,omitempty"`
	DefaultLimitPolicyID *uuid.UUID `gorm:"type:uuid" json:"default_limit_policy_id,omitempty"`

	// 升级条件
	UpgradeRequirements string `gorm:"type:jsonb" json:"upgrade_requirements,omitempty"` // JSON: { "min_monthly_volume": 1000000, "min_transactions": 100 }

	// 功能权限
	AllowedChannels   string `gorm:"type:jsonb" json:"allowed_channels,omitempty"`    // JSON: ["stripe", "paypal", "crypto"]
	AllowedCurrencies string `gorm:"type:jsonb" json:"allowed_currencies,omitempty"`  // JSON: ["USD", "EUR", "CNY"]
	MaxAPICallsPerMin int    `gorm:"type:integer;default:100" json:"max_api_calls_per_min"`

	// 状态
	IsActive bool `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantTier) TableName() string {
	return "merchant_tiers"
}

// 商户等级常量
const (
	TierCodeStarter      = "starter"       // 入门级
	TierCodeBasic        = "basic"         // 基础级
	TierCodeProfessional = "professional"  // 专业级
	TierCodeEnterprise   = "enterprise"    // 企业级
	TierCodePremium      = "premium"       // 高级版
)

// MerchantPolicyBinding 商户策略绑定表 (记录商户当前等级和策略)
type MerchantPolicyBinding struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;unique;not null;index" json:"merchant_id"`
	TierID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tier_id"` // 当前等级

	// 可选的自定义策略覆盖
	CustomFeePolicyID   *uuid.UUID `gorm:"type:uuid" json:"custom_fee_policy_id,omitempty"`
	CustomLimitPolicyID *uuid.UUID `gorm:"type:uuid" json:"custom_limit_policy_id,omitempty"`

	// 生效时间
	EffectiveDate time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"`
	ExpiryDate    *time.Time `gorm:"type:timestamptz" json:"expiry_date,omitempty"` // null表示长期有效

	// 变更记录
	ChangedBy     *uuid.UUID `gorm:"type:uuid" json:"changed_by,omitempty"`
	ChangeReason  string     `gorm:"type:varchar(500)" json:"change_reason,omitempty"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantPolicyBinding) TableName() string {
	return "merchant_policy_bindings"
}

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantFeePolicy 商户费率策略表 (统一管理所有费率配置)
type MerchantFeePolicy struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// 关联关系 (二选一: 要么是商户自定义,要么是等级默认)
	MerchantID *uuid.UUID `gorm:"type:uuid;index" json:"merchant_id,omitempty"` // null表示等级默认策略
	TierID     *uuid.UUID `gorm:"type:uuid;index" json:"tier_id,omitempty"`     // 关联等级(如果是默认策略)

	// 适用范围
	Channel       string `gorm:"type:varchar(50);not null;index" json:"channel"`        // stripe, paypal, crypto, all
	PaymentMethod string `gorm:"type:varchar(50);index" json:"payment_method"`          // card, bank_transfer, wallet, all
	Currency      string `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"` // 币种

	// 费率配置
	FeeType       string  `gorm:"type:varchar(20);not null" json:"fee_type"`         // percentage, fixed, tiered
	FeePercentage float64 `gorm:"type:decimal(5,4);default:0" json:"fee_percentage"` // 费率百分比（如0.029表示2.9%）
	FeeFixed      int64   `gorm:"type:bigint;default:0" json:"fee_fixed"`            // 固定费用（分）
	MinFee        int64   `gorm:"type:bigint;default:0" json:"min_fee"`              // 最小费用（分）
	MaxFee        *int64  `gorm:"type:bigint" json:"max_fee,omitempty"`              // 最大费用（分，null表示无上限）

	// 阶梯费率规则 (当FeeType=tiered时使用)
	TieredRules string `gorm:"type:jsonb" json:"tiered_rules,omitempty"`
	// JSON格式: [
	//   { "min_amount": 0, "max_amount": 100000, "percentage": 0.029 },
	//   { "min_amount": 100000, "max_amount": 1000000, "percentage": 0.025 },
	//   { "min_amount": 1000000, "max_amount": null, "percentage": 0.020 }
	// ]

	// 生效时间
	EffectiveDate time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"`
	ExpiryDate    *time.Time `gorm:"type:timestamptz" json:"expiry_date,omitempty"` // null表示长期有效

	// 优先级 (数字越大优先级越高)
	Priority int `gorm:"default:0;index" json:"priority"`

	// 状态
	Status string `gorm:"type:varchar(20);default:'active';index" json:"status"` // active, inactive, pending

	// 审批流程
	CreatedBy  *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	ApprovedBy *uuid.UUID `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt *time.Time `gorm:"type:timestamptz" json:"approved_at,omitempty"`

	// 扩展信息
	Metadata  string `gorm:"type:jsonb" json:"metadata,omitempty"`
	Remarks   string `gorm:"type:text" json:"remarks,omitempty"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantFeePolicy) TableName() string {
	return "merchant_fee_policies"
}

// 费率类型常量
const (
	FeeTypePercentage = "percentage" // 百分比费率
	FeeTypeFixed      = "fixed"      // 固定费用
	FeeTypeTiered     = "tiered"     // 阶梯费率
)

// 费率状态常量
const (
	FeeStatusActive   = "active"   // 启用
	FeeStatusInactive = "inactive" // 停用
	FeeStatusPending  = "pending"  // 待审批
)

// 渠道常量
const (
	ChannelAll    = "all"    // 所有渠道
	ChannelStripe = "stripe" // Stripe
	ChannelPayPal = "paypal" // PayPal
	ChannelCrypto = "crypto" // 加密货币
	ChannelAdyen  = "adyen"  // Adyen
	ChannelSquare = "square" // Square
)

// 支付方式常量
const (
	PaymentMethodAll          = "all"           // 所有方式
	PaymentMethodCard         = "card"          // 卡支付
	PaymentMethodBankTransfer = "bank_transfer" // 银行转账
	PaymentMethodWallet       = "wallet"        // 钱包支付
)

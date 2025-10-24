package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantFeeConfig 商户费率配置表
type MerchantFeeConfig struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`                   // 商户ID
	Channel       string         `gorm:"type:varchar(50);not null;index" json:"channel"`                // 渠道：stripe, paypal, all
	PaymentMethod string         `gorm:"type:varchar(50)" json:"payment_method"`                        // 支付方式：card, bank_transfer, all
	FeeType       string         `gorm:"type:varchar(20);not null" json:"fee_type"`                     // percentage, fixed, tiered
	FeePercentage float64        `gorm:"type:decimal(5,4);default:0" json:"fee_percentage"`             // 费率百分比（如0.029表示2.9%）
	FeeFixed      int64          `gorm:"type:bigint;default:0" json:"fee_fixed"`                        // 固定费用（分）
	MinFee        int64          `gorm:"type:bigint;default:0" json:"min_fee"`                          // 最小费用（分）
	MaxFee        int64          `gorm:"type:bigint" json:"max_fee"`                                    // 最大费用（分，null表示无上限）
	Currency      string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`       // 币种
	TieredRules   string         `gorm:"type:jsonb" json:"tiered_rules"`                                // 阶梯费率规则（JSON）
	EffectiveDate time.Time      `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"` // 生效日期
	ExpiryDate    *time.Time     `gorm:"type:timestamptz" json:"expiry_date"`                           // 失效日期（null表示长期有效）
	Priority      int            `gorm:"default:0" json:"priority"`                                     // 优先级（数字越大优先级越高）
	Status        string         `gorm:"type:varchar(20);default:'active'" json:"status"`               // active, inactive
	CreatedBy     *uuid.UUID     `gorm:"type:uuid" json:"created_by"`                                   // 创建人
	ApprovedBy    *uuid.UUID     `gorm:"type:uuid" json:"approved_by"`                                  // 审批人
	ApprovedAt    *time.Time     `gorm:"type:timestamptz" json:"approved_at"`                           // 审批时间
	Metadata      string         `gorm:"type:jsonb" json:"metadata"`                                    // 扩展信息
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantFeeConfig) TableName() string {
	return "merchant_fee_configs"
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
)

// 渠道常量
const (
	ChannelAll     = "all"     // 所有渠道
	ChannelStripe  = "stripe"  // Stripe
	ChannelPayPal  = "paypal"  // PayPal
	ChannelCrypto  = "crypto"  // 加密货币
	ChannelAdyen   = "adyen"   // Adyen
	ChannelSquare  = "square"  // Square
)

// 支付方式常量
const (
	PaymentMethodAll          = "all"           // 所有方式
	PaymentMethodCard         = "card"          // 卡支付
	PaymentMethodBankTransfer = "bank_transfer" // 银行转账
	PaymentMethodWallet       = "wallet"        // 钱包支付
)

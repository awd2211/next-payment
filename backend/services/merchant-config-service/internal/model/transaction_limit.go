package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantTransactionLimit 商户交易限额配置表
type MerchantTransactionLimit struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`                   // 商户ID
	LimitType     string         `gorm:"type:varchar(20);not null;index" json:"limit_type"`             // single, daily, monthly
	PaymentMethod string         `gorm:"type:varchar(50)" json:"payment_method"`                        // card, bank_transfer, all
	Channel       string         `gorm:"type:varchar(50)" json:"channel"`                               // stripe, paypal, all
	Currency      string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`       // 币种
	MinAmount     int64          `gorm:"type:bigint;default:0" json:"min_amount"`                       // 最小金额（分）
	MaxAmount     int64          `gorm:"type:bigint" json:"max_amount"`                                 // 最大金额（分）
	MaxCount      int            `gorm:"default:0" json:"max_count"`                                    // 最大笔数（0表示不限制）
	Status        string         `gorm:"type:varchar(20);default:'active'" json:"status"`               // active, inactive
	EffectiveDate time.Time      `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"` // 生效日期
	ExpiryDate    *time.Time     `gorm:"type:timestamptz" json:"expiry_date"`                           // 失效日期
	Metadata      string         `gorm:"type:jsonb" json:"metadata"`                                    // 扩展信息
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantTransactionLimit) TableName() string {
	return "merchant_transaction_limits"
}

// 限额类型常量
const (
	LimitTypeSingle  = "single"  // 单笔限额
	LimitTypeDaily   = "daily"   // 日累计限额
	LimitTypeMonthly = "monthly" // 月累计限额
)

// 限额状态常量
const (
	LimitStatusActive   = "active"   // 启用
	LimitStatusInactive = "inactive" // 停用
)

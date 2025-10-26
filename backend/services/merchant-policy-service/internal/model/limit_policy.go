package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantLimitPolicy 商户限额策略表 (统一管理所有限额配置)
type MerchantLimitPolicy struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// 关联关系 (二选一: 要么是商户自定义,要么是等级默认)
	MerchantID *uuid.UUID `gorm:"type:uuid;index" json:"merchant_id,omitempty"` // null表示等级默认策略
	TierID     *uuid.UUID `gorm:"type:uuid;index" json:"tier_id,omitempty"`     // 关联等级(如果是默认策略)

	// 适用范围
	Channel  string `gorm:"type:varchar(50);not null;index;default:'all'" json:"channel"`  // all, stripe, paypal (按渠道差异化)
	Currency string `gorm:"type:varchar(10);not null;index;default:'USD'" json:"currency"` // USD, EUR, CNY (按币种差异化)

	// 交易限额配置 (单位: 分/cents)
	SingleTransMin int64 `gorm:"type:bigint;default:0" json:"single_trans_min"`         // 单笔最小金额
	SingleTransMax int64 `gorm:"type:bigint;not null" json:"single_trans_max"`          // 单笔最大金额
	DailyLimit     int64 `gorm:"type:bigint;not null" json:"daily_limit"`               // 日限额
	MonthlyLimit   int64 `gorm:"type:bigint;not null" json:"monthly_limit"`             // 月限额
	YearlyLimit    *int64 `gorm:"type:bigint" json:"yearly_limit,omitempty"`            // 年限额 (可选)

	// 特殊限额
	MaxPendingAmount int64  `gorm:"type:bigint;default:0" json:"max_pending_amount"`   // 最大待结算金额
	MaxRefundDaily   int64  `gorm:"type:bigint;default:0" json:"max_refund_daily"`     // 日退款限额
	MaxRefundMonthly int64  `gorm:"type:bigint;default:0" json:"max_refund_monthly"`   // 月退款限额

	// 笔数限制
	MaxTransactionsDaily   *int `gorm:"type:integer" json:"max_transactions_daily,omitempty"`   // 日最大交易笔数
	MaxTransactionsMonthly *int `gorm:"type:integer" json:"max_transactions_monthly,omitempty"` // 月最大交易笔数

	// 生效时间
	EffectiveDate time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"`
	ExpiryDate    *time.Time `gorm:"type:timestamptz" json:"expiry_date,omitempty"` // null表示长期有效

	// 优先级 (数字越大优先级越高, 商户自定义 > 等级默认)
	Priority int `gorm:"default:0;index" json:"priority"`

	// 状态
	Status string `gorm:"type:varchar(20);default:'active';index" json:"status"` // active, inactive, pending

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
func (MerchantLimitPolicy) TableName() string {
	return "merchant_limit_policies"
}

// 限额策略状态常量
const (
	LimitStatusActive   = "active"   // 启用
	LimitStatusInactive = "inactive" // 停用
	LimitStatusPending  = "pending"  // 待审批
)

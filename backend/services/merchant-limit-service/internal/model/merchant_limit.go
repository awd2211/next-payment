package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantTier 商户等级表
type MerchantTier struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TierCode    string    `gorm:"type:varchar(50);unique;not null;index" json:"tier_code"`
	TierName    string    `gorm:"type:varchar(100);not null" json:"tier_name"`
	TierLevel   int       `gorm:"type:integer;not null;index" json:"tier_level"` // 1-5 (1=最低, 5=最高)
	Description string    `gorm:"type:text" json:"description"`

	// 额度配置 (单位: 分/cents)
	DailyLimit          int64 `gorm:"type:bigint;not null" json:"daily_limit"`            // 日限额
	MonthlyLimit        int64 `gorm:"type:bigint;not null" json:"monthly_limit"`          // 月限额
	SingleTransLimit    int64 `gorm:"type:bigint;not null" json:"single_trans_limit"`     // 单笔限额
	MinTransAmount      int64 `gorm:"type:bigint;default:0" json:"min_trans_amount"`      // 最小交易金额
	MaxPendingAmount    int64 `gorm:"type:bigint;default:0" json:"max_pending_amount"`    // 最大待结算金额
	MaxRefundAmount     int64 `gorm:"type:bigint;default:0" json:"max_refund_amount"`     // 单日退款限额

	// 费率配置
	TransactionFeeRate  float64 `gorm:"type:decimal(5,4);not null" json:"transaction_fee_rate"`  // 交易手续费率 (0.0025 = 0.25%)
	WithdrawalFeeRate   float64 `gorm:"type:decimal(5,4);not null" json:"withdrawal_fee_rate"`   // 提现手续费率
	ChargebackFeeRate   float64 `gorm:"type:decimal(5,4);default:0" json:"chargeback_fee_rate"`  // 拒付手续费率

	// 功能权限
	AllowedChannels     string `gorm:"type:jsonb" json:"allowed_channels"`      // 允许的支付渠道 (JSON array)
	AllowedCurrencies   string `gorm:"type:jsonb" json:"allowed_currencies"`    // 允许的币种 (JSON array)
	MaxAPICallsPerMin   int    `gorm:"type:integer;default:100" json:"max_api_calls_per_min"` // API调用频率限制

	// 时间戳
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

// MerchantLimit 商户额度表
type MerchantLimit struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;unique;not null;index" json:"merchant_id"`
	TierID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tier_id"`

	// 当前使用量 (单位: 分/cents)
	DailyUsed        int64     `gorm:"type:bigint;default:0" json:"daily_used"`
	MonthlyUsed      int64     `gorm:"type:bigint;default:0" json:"monthly_used"`
	PendingAmount    int64     `gorm:"type:bigint;default:0" json:"pending_amount"`    // 待结算金额
	RefundedToday    int64     `gorm:"type:bigint;default:0" json:"refunded_today"`    // 今日退款金额

	// 重置时间
	DailyResetAt     time.Time `gorm:"type:timestamptz" json:"daily_reset_at"`
	MonthlyResetAt   time.Time `gorm:"type:timestamptz" json:"monthly_reset_at"`

	// 自定义额度 (可选，覆盖tier配置)
	CustomDailyLimit       *int64 `gorm:"type:bigint" json:"custom_daily_limit,omitempty"`
	CustomMonthlyLimit     *int64 `gorm:"type:bigint" json:"custom_monthly_limit,omitempty"`
	CustomSingleTransLimit *int64 `gorm:"type:bigint" json:"custom_single_trans_limit,omitempty"`

	// 状态
	IsSuspended      bool      `gorm:"default:false" json:"is_suspended"`
	SuspendedAt      *time.Time `gorm:"type:timestamptz" json:"suspended_at,omitempty"`
	SuspendedReason  string    `gorm:"type:varchar(500)" json:"suspended_reason,omitempty"`

	// 时间戳
	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantLimit) TableName() string {
	return "merchant_limits"
}

// LimitUsageLog 额度使用日志表
type LimitUsageLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index" json:"merchant_id"`

	// 关联信息
	PaymentNo  string `gorm:"type:varchar(64);index" json:"payment_no,omitempty"`
	OrderNo    string `gorm:"type:varchar(64);index" json:"order_no,omitempty"`

	// 操作类型
	ActionType string `gorm:"type:varchar(20);not null;index" json:"action_type"` // consume, release, reset

	// 金额信息
	Amount     int64  `gorm:"type:bigint;not null" json:"amount"`
	Currency   string `gorm:"type:varchar(10);not null" json:"currency"`

	// 额度快照
	DailyLimitBefore   int64 `gorm:"type:bigint" json:"daily_limit_before"`
	DailyUsedBefore    int64 `gorm:"type:bigint" json:"daily_used_before"`
	DailyUsedAfter     int64 `gorm:"type:bigint" json:"daily_used_after"`
	MonthlyUsedBefore  int64 `gorm:"type:bigint" json:"monthly_used_before"`
	MonthlyUsedAfter   int64 `gorm:"type:bigint" json:"monthly_used_after"`

	// 结果
	Success      bool   `gorm:"default:true" json:"success"`
	FailureReason string `gorm:"type:varchar(200)" json:"failure_reason,omitempty"`

	// 扩展信息
	Metadata   string `gorm:"type:jsonb" json:"metadata,omitempty"`

	// 时间戳
	CreatedAt  time.Time      `gorm:"type:timestamptz;default:now();index" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (LimitUsageLog) TableName() string {
	return "limit_usage_logs"
}

// 操作类型常量
const (
	ActionTypeConsume = "consume" // 消费额度
	ActionTypeRelease = "release" // 释放额度
	ActionTypeReset   = "reset"   // 重置额度
	ActionTypeAdjust  = "adjust"  // 手动调整
)

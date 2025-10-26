package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantQuota 商户配额表 (实时追踪配额消耗)
type MerchantQuota struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null" json:"merchant_id"`

	// 按币种分别追踪 (支持多币种)
	Currency string `gorm:"type:varchar(10);not null;index" json:"currency"` // USD, EUR, CNY

	// 当前使用量 (单位: 分/cents)
	DailyUsed     int64 `gorm:"type:bigint;default:0" json:"daily_used"`
	MonthlyUsed   int64 `gorm:"type:bigint;default:0" json:"monthly_used"`
	YearlyUsed    int64 `gorm:"type:bigint;default:0" json:"yearly_used"`
	PendingAmount int64 `gorm:"type:bigint;default:0" json:"pending_amount"`    // 待结算金额
	RefundedToday int64 `gorm:"type:bigint;default:0" json:"refunded_today"`    // 今日退款金额
	RefundedMonth int64 `gorm:"type:bigint;default:0" json:"refunded_month"`    // 本月退款金额

	// 笔数统计
	TransactionsToday int `gorm:"type:integer;default:0" json:"transactions_today"`
	TransactionsMonth int `gorm:"type:integer;default:0" json:"transactions_month"`

	// 重置时间
	DailyResetAt   time.Time `gorm:"type:timestamptz;not null" json:"daily_reset_at"`
	MonthlyResetAt time.Time `gorm:"type:timestamptz;not null" json:"monthly_reset_at"`
	YearlyResetAt  time.Time `gorm:"type:timestamptz;not null" json:"yearly_reset_at"`

	// 状态
	IsSuspended     bool       `gorm:"default:false" json:"is_suspended"`
	SuspendedAt     *time.Time `gorm:"type:timestamptz" json:"suspended_at,omitempty"`
	SuspendedReason string     `gorm:"type:varchar(500)" json:"suspended_reason,omitempty"`
	SuspendedBy     *uuid.UUID `gorm:"type:uuid" json:"suspended_by,omitempty"`

	// 最后更新的订单号 (防止重复消耗)
	LastOrderNo    string    `gorm:"type:varchar(64)" json:"last_order_no,omitempty"`
	LastPaymentNo  string    `gorm:"type:varchar(64)" json:"last_payment_no,omitempty"`
	LastUpdatedBy  string    `gorm:"type:varchar(100)" json:"last_updated_by,omitempty"` // API caller identifier

	// 版本号 (乐观锁,防止并发冲突)
	Version int `gorm:"default:0" json:"version"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantQuota) TableName() string {
	return "merchant_quotas"
}

// QuotaUsageLog 配额使用日志表 (审计和追溯)
type QuotaUsageLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index" json:"merchant_id"`

	// 关联信息
	PaymentNo string `gorm:"type:varchar(64);index" json:"payment_no,omitempty"`
	OrderNo   string `gorm:"type:varchar(64);index" json:"order_no,omitempty"`
	RefundNo  string `gorm:"type:varchar(64);index" json:"refund_no,omitempty"`

	// 操作类型
	ActionType string `gorm:"type:varchar(20);not null;index" json:"action_type"` // consume, release, reset, adjust

	// 金额信息
	Amount   int64  `gorm:"type:bigint;not null" json:"amount"`
	Currency string `gorm:"type:varchar(10);not null" json:"currency"`

	// 配额快照 (操作前后的状态)
	DailyLimitBefore  int64 `gorm:"type:bigint" json:"daily_limit_before"`
	DailyUsedBefore   int64 `gorm:"type:bigint" json:"daily_used_before"`
	DailyUsedAfter    int64 `gorm:"type:bigint" json:"daily_used_after"`
	MonthlyUsedBefore int64 `gorm:"type:bigint" json:"monthly_used_before"`
	MonthlyUsedAfter  int64 `gorm:"type:bigint" json:"monthly_used_after"`

	// 结果
	Success       bool   `gorm:"default:true" json:"success"`
	FailureReason string `gorm:"type:varchar(200)" json:"failure_reason,omitempty"`

	// 操作人
	OperatorID   *uuid.UUID `gorm:"type:uuid" json:"operator_id,omitempty"` // 管理员手动调整时记录
	OperatorType string     `gorm:"type:varchar(50)" json:"operator_type,omitempty"` // system, admin, api

	// 扩展信息
	Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"`
	Remarks  string `gorm:"type:text" json:"remarks,omitempty"`

	// 时间戳
	CreatedAt time.Time      `gorm:"type:timestamptz;default:now();index" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (QuotaUsageLog) TableName() string {
	return "quota_usage_logs"
}

// 操作类型常量
const (
	ActionTypeConsume = "consume" // 消费配额
	ActionTypeRelease = "release" // 释放配额 (订单取消/退款)
	ActionTypeReset   = "reset"   // 重置配额 (定时任务)
	ActionTypeAdjust  = "adjust"  // 手动调整 (管理员操作)
)

// 操作人类型常量
const (
	OperatorTypeSystem = "system" // 系统自动操作
	OperatorTypeAdmin  = "admin"  // 管理员手动操作
	OperatorTypeAPI    = "api"    // API调用
)

// QuotaAlert 配额预警表 (监控和告警)
type QuotaAlert struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index" json:"merchant_id"`
	Currency   string    `gorm:"type:varchar(10);not null" json:"currency"`

	// 预警类型
	AlertType string `gorm:"type:varchar(50);not null;index" json:"alert_type"` // daily_80, daily_100, monthly_80, monthly_100, suspended

	// 预警级别
	AlertLevel string `gorm:"type:varchar(20);not null" json:"alert_level"` // warning, critical

	// 预警内容
	Message string `gorm:"type:text;not null" json:"message"`

	// 配额信息
	CurrentUsed  int64   `gorm:"type:bigint;not null" json:"current_used"`
	Limit        int64   `gorm:"type:bigint;not null" json:"limit"`
	UsagePercent float64 `gorm:"type:decimal(5,2);not null" json:"usage_percent"` // 使用率百分比

	// 是否已处理
	IsResolved   bool       `gorm:"default:false" json:"is_resolved"`
	ResolvedAt   *time.Time `gorm:"type:timestamptz" json:"resolved_at,omitempty"`
	ResolvedBy   *uuid.UUID `gorm:"type:uuid" json:"resolved_by,omitempty"`
	ResolveNotes string     `gorm:"type:text" json:"resolve_notes,omitempty"`

	// 通知记录
	NotificationSent bool       `gorm:"default:false" json:"notification_sent"`
	NotificationSentAt *time.Time `gorm:"type:timestamptz" json:"notification_sent_at,omitempty"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now();index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (QuotaAlert) TableName() string {
	return "quota_alerts"
}

// 预警类型常量
const (
	AlertTypeDailyWarning    = "daily_80"    // 日限额达到80%
	AlertTypeDailyCritical   = "daily_100"   // 日限额达到100%
	AlertTypeMonthlyWarning  = "monthly_80"  // 月限额达到80%
	AlertTypeMonthlyCritical = "monthly_100" // 月限额达到100%
	AlertTypeSuspended       = "suspended"   // 商户被暂停
)

// 预警级别常量
const (
	AlertLevelWarning  = "warning"  // 警告级别 (80%)
	AlertLevelCritical = "critical" // 严重级别 (100%)
)

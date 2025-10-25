package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReconciliationTask 对账任务表
type ReconciliationTask struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskNo   string    `gorm:"type:varchar(64);unique;not null;index" json:"task_no"`
	TaskDate time.Time `gorm:"type:date;not null;index" json:"task_date"`
	Channel  string    `gorm:"type:varchar(50);not null;index" json:"channel"`
	TaskType string    `gorm:"type:varchar(20);not null" json:"task_type"`

	// 统计信息
	PlatformCount  int   `gorm:"default:0" json:"platform_count"`
	PlatformAmount int64 `gorm:"default:0" json:"platform_amount"`
	ChannelCount   int   `gorm:"default:0" json:"channel_count"`
	ChannelAmount  int64 `gorm:"default:0" json:"channel_amount"`
	MatchedCount   int   `gorm:"default:0" json:"matched_count"`
	MatchedAmount  int64 `gorm:"default:0" json:"matched_amount"`
	DiffCount      int   `gorm:"default:0" json:"diff_count"`
	DiffAmount     int64 `gorm:"default:0" json:"diff_amount"`

	// 状态信息
	Status       string `gorm:"type:varchar(20);not null;index" json:"status"`
	Progress     int    `gorm:"default:0" json:"progress"`
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// 文件信息
	ChannelFileURL string `gorm:"type:varchar(500)" json:"channel_file_url,omitempty"`
	ReportFileURL  string `gorm:"type:varchar(500)" json:"report_file_url,omitempty"`

	// 时间戳
	StartedAt   *time.Time     `gorm:"type:timestamptz" json:"started_at,omitempty"`
	CompletedAt *time.Time     `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ReconciliationTask) TableName() string {
	return "reconciliation_tasks"
}

// 任务状态常量
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// 任务类型常量
const (
	TaskTypeDaily     = "daily"
	TaskTypeManual    = "manual"
	TaskTypeReconcile = "reconcile"
)

// ReconciliationRecord 对账差异记录表
type ReconciliationRecord struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID uuid.UUID `gorm:"type:uuid;not null;index" json:"task_id"`
	TaskNo string    `gorm:"type:varchar(64);not null;index" json:"task_no"`

	// 订单信息
	PaymentNo       string     `gorm:"type:varchar(64);index" json:"payment_no,omitempty"`
	ChannelTradeNo  string     `gorm:"type:varchar(128);index" json:"channel_trade_no,omitempty"`
	OrderNo         string     `gorm:"type:varchar(64)" json:"order_no,omitempty"`
	MerchantID      *uuid.UUID `gorm:"type:uuid;index" json:"merchant_id,omitempty"`

	// 金额信息
	PlatformAmount int64  `gorm:"type:bigint" json:"platform_amount,omitempty"`
	ChannelAmount  int64  `gorm:"type:bigint" json:"channel_amount,omitempty"`
	DiffAmount     int64  `gorm:"type:bigint" json:"diff_amount,omitempty"`
	Currency       string `gorm:"type:varchar(10)" json:"currency,omitempty"`

	// 状态信息
	PlatformStatus string `gorm:"type:varchar(20)" json:"platform_status,omitempty"`
	ChannelStatus  string `gorm:"type:varchar(20)" json:"channel_status,omitempty"`
	DiffType       string `gorm:"type:varchar(20);not null;index" json:"diff_type"`
	DiffReason     string `gorm:"type:text" json:"diff_reason,omitempty"`

	// 处理信息
	IsResolved     bool       `gorm:"default:false;index" json:"is_resolved"`
	ResolvedBy     *uuid.UUID `gorm:"type:uuid" json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time `gorm:"type:timestamptz" json:"resolved_at,omitempty"`
	ResolutionNote string     `gorm:"type:text" json:"resolution_note,omitempty"`

	// 扩展信息
	Extra string `gorm:"type:jsonb" json:"extra,omitempty"`

	// 时间戳
	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ReconciliationRecord) TableName() string {
	return "reconciliation_records"
}

// 差异类型常量
const (
	DiffTypeMatched       = "matched"        // 完全匹配
	DiffTypePlatformOnly  = "platform_only"  // 仅平台有记录
	DiffTypeChannelOnly   = "channel_only"   // 仅渠道有记录
	DiffTypeAmountDiff    = "amount_diff"    // 金额不一致
	DiffTypeStatusDiff    = "status_diff"    // 状态不一致
)

// ChannelSettlementFile 渠道账单文件表
type ChannelSettlementFile struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FileNo         string    `gorm:"type:varchar(64);unique;not null;index" json:"file_no"`
	Channel        string    `gorm:"type:varchar(50);not null;index" json:"channel"`
	SettlementDate time.Time `gorm:"type:date;not null;index" json:"settlement_date"`
	FileURL        string    `gorm:"type:varchar(500);not null" json:"file_url"`
	FileSize       int64     `gorm:"type:bigint" json:"file_size,omitempty"`
	FileHash       string    `gorm:"type:varchar(64)" json:"file_hash,omitempty"`

	// 统计信息
	RecordCount int    `gorm:"default:0" json:"record_count"`
	TotalAmount int64  `gorm:"default:0" json:"total_amount"`
	Currency    string `gorm:"type:varchar(10)" json:"currency,omitempty"`

	// 状态信息
	Status string `gorm:"type:varchar(20);not null;index" json:"status"`

	// 时间戳
	DownloadedAt *time.Time     `gorm:"type:timestamptz" json:"downloaded_at,omitempty"`
	ParsedAt     *time.Time     `gorm:"type:timestamptz" json:"parsed_at,omitempty"`
	ImportedAt   *time.Time     `gorm:"type:timestamptz" json:"imported_at,omitempty"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ChannelSettlementFile) TableName() string {
	return "channel_settlement_files"
}

// 文件状态常量
const (
	FileStatusPending    = "pending"
	FileStatusDownloaded = "downloaded"
	FileStatusParsed     = "parsed"
	FileStatusImported   = "imported"
)

// 支持的渠道常量
const (
	ChannelStripe  = "stripe"
	ChannelPayPal  = "paypal"
	ChannelAlipay  = "alipay"
	ChannelWechat  = "wechat"
)

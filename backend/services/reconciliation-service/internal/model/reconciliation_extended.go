package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Additional task type for realtime (extends existing types in reconciliation.go)
const (
	TaskTypeRealtime = "realtime" // 实时对账
)

// ReconciliationDifference 对账差异记录 (automation扩展)
type ReconciliationDifference struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	DifferenceType string     `gorm:"type:varchar(50);not null;index" json:"difference_type"`
	OrderNo        string     `gorm:"type:varchar(100);not null;index" json:"order_no"`
	InternalAmount int64      `gorm:"type:bigint" json:"internal_amount"`
	ChannelAmount  int64      `gorm:"type:bigint" json:"channel_amount"`
	AmountDiff     int64      `gorm:"type:bigint" json:"amount_diff"`
	InternalStatus string     `gorm:"type:varchar(50)" json:"internal_status"`
	ChannelStatus  string     `gorm:"type:varchar(50)" json:"channel_status"`
	InternalTime   time.Time  `gorm:"type:timestamptz" json:"internal_time,omitempty"`
	ChannelTime    time.Time  `gorm:"type:timestamptz" json:"channel_time,omitempty"`
	Severity       string     `gorm:"type:varchar(20);not null;index" json:"severity"`
	Description    string     `gorm:"type:text" json:"description"`
	Status         string     `gorm:"type:varchar(20);default:'unresolved';index" json:"status"`
	ResolvedBy     *uuid.UUID `gorm:"type:uuid" json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time `gorm:"type:timestamptz" json:"resolved_at,omitempty"`
	ResolutionNote string     `gorm:"type:text" json:"resolution_note,omitempty"`
	AlertSent      bool       `gorm:"type:boolean;default:false" json:"alert_sent"`
	DetectedAt     time.Time  `gorm:"type:timestamptz;not null" json:"detected_at"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (ReconciliationDifference) TableName() string {
	return "reconciliation_differences"
}

// InternalTransaction 内部交易记录
type InternalTransaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderNo     string    `gorm:"type:varchar(100);not null;unique;index" json:"order_no"`
	Amount      int64     `gorm:"type:bigint;not null" json:"amount"`
	Currency    string    `gorm:"type:varchar(10);not null" json:"currency"`
	Status      string    `gorm:"type:varchar(50);not null" json:"status"`
	CompletedAt time.Time `gorm:"type:timestamptz;index" json:"completed_at"`
	CreatedAt   time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}

// TableName 表名
func (InternalTransaction) TableName() string {
	return "internal_transactions"
}

// ChannelTransaction 渠道交易记录
type ChannelTransaction struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChannelName       string    `gorm:"type:varchar(50);not null;index" json:"channel_name"`
	ChannelTradeNo    string    `gorm:"type:varchar(100);not null;unique;index" json:"channel_trade_no"`
	OrderNo           string    `gorm:"type:varchar(100);index" json:"order_no"`
	Amount            int64     `gorm:"type:bigint;not null" json:"amount"`
	Currency          string    `gorm:"type:varchar(10);not null" json:"currency"`
	Status            string    `gorm:"type:varchar(50);not null" json:"status"`
	CompletedAt       time.Time `gorm:"type:timestamptz;index" json:"completed_at"`
	SettlementFileNo  string    `gorm:"type:varchar(100);index" json:"settlement_file_no,omitempty"`
	CreatedAt         time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}

// TableName 表名
func (ChannelTransaction) TableName() string {
	return "channel_transactions"
}

// ReconciliationReport 对账报告
type ReconciliationReport struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReportDate       time.Time  `gorm:"type:date;not null;unique;index" json:"report_date"`
	TotalTasks       int        `gorm:"type:int;not null" json:"total_tasks"`
	CompletedTasks   int        `gorm:"type:int;not null" json:"completed_tasks"`
	FailedTasks      int        `gorm:"type:int;not null" json:"failed_tasks"`
	TotalInternal    int        `gorm:"type:int;not null" json:"total_internal"`
	TotalChannel     int        `gorm:"type:int;not null" json:"total_channel"`
	TotalMatched     int        `gorm:"type:int;not null" json:"total_matched"`
	TotalDifferences int        `gorm:"type:int;not null" json:"total_differences"`
	CriticalDiffs    int        `gorm:"type:int;default:0" json:"critical_diffs"`
	HighDiffs        int        `gorm:"type:int;default:0" json:"high_diffs"`
	MediumDiffs      int        `gorm:"type:int;default:0" json:"medium_diffs"`
	LowDiffs         int        `gorm:"type:int;default:0" json:"low_diffs"`
	TotalAmountDiff  int64      `gorm:"type:bigint;default:0" json:"total_amount_diff"`
	ReportSent       bool       `gorm:"type:boolean;default:false" json:"report_sent"`
	SentAt           *time.Time `gorm:"type:timestamptz" json:"sent_at,omitempty"`
	CreatedAt        time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (ReconciliationReport) TableName() string {
	return "reconciliation_reports"
}

// ReconciliationResult 对账执行结果
type ReconciliationResult struct {
	TotalInternal    int
	TotalChannel     int
	Matched          int
	TotalDifferences int
	CriticalDiffs    int
	HighDiffs        int
	MediumDiffs      int
	LowDiffs         int
	Differences      []*ReconciliationDifference
}

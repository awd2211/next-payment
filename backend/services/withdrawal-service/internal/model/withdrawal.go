package model

import (
	"time"

	"github.com/google/uuid"
)

// WithdrawalStatus 提现状态
type WithdrawalStatus string

const (
	WithdrawalStatusPending    WithdrawalStatus = "pending"     // 待审核
	WithdrawalStatusApproved   WithdrawalStatus = "approved"    // 已审批
	WithdrawalStatusRejected   WithdrawalStatus = "rejected"    // 已拒绝
	WithdrawalStatusProcessing WithdrawalStatus = "processing"  // 处理中
	WithdrawalStatusCompleted  WithdrawalStatus = "completed"   // 已完成
	WithdrawalStatusFailed     WithdrawalStatus = "failed"      // 失败
	WithdrawalStatusCancelled  WithdrawalStatus = "cancelled"   // 已取消
)

// WithdrawalType 提现类型
type WithdrawalType string

const (
	WithdrawalTypeNormal   WithdrawalType = "normal"    // 普通提现
	WithdrawalTypeUrgent   WithdrawalType = "urgent"    // 加急提现
	WithdrawalTypeScheduled WithdrawalType = "scheduled" // 定时提现
)

// Withdrawal 提现记录
type Withdrawal struct {
	ID              uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WithdrawalNo    string           `gorm:"type:varchar(64);uniqueIndex;not null" json:"withdrawal_no"`
	MerchantID      uuid.UUID        `gorm:"type:uuid;index;not null" json:"merchant_id"`
	Amount          int64            `gorm:"not null" json:"amount"`            // 提现金额（分）
	Fee             int64            `gorm:"not null;default:0" json:"fee"`     // 手续费（分）
	ActualAmount    int64            `gorm:"not null" json:"actual_amount"`     // 实际到账金额（分）
	Type            WithdrawalType   `gorm:"type:varchar(20);not null;default:'normal'" json:"type"`
	Status          WithdrawalStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	BankAccountID   uuid.UUID        `gorm:"type:uuid;not null" json:"bank_account_id"`
	BankName        string           `gorm:"type:varchar(100);not null" json:"bank_name"`
	BankAccountName string           `gorm:"type:varchar(100);not null" json:"bank_account_name"`
	BankAccountNo   string           `gorm:"type:varchar(64);not null" json:"bank_account_no"`
	Remarks         string           `gorm:"type:text" json:"remarks"`
	ApprovalLevel   int              `gorm:"not null;default:0" json:"approval_level"`     // 当前审批级别
	RequiredLevel   int              `gorm:"not null;default:1" json:"required_level"`     // 需要审批级别
	ChannelTradeNo  string           `gorm:"type:varchar(128)" json:"channel_trade_no"`    // 渠道交易号
	FailureReason   string           `gorm:"type:text" json:"failure_reason"`              // 失败原因
	ProcessedAt     *time.Time       `json:"processed_at"`                                 // 处理时间
	CompletedAt     *time.Time       `json:"completed_at"`                                 // 完成时间
	CreatedBy       uuid.UUID        `gorm:"type:uuid;not null" json:"created_by"`         // 创建人
	CreatedAt       time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Withdrawal) TableName() string {
	return "withdrawals"
}

// WithdrawalBankAccount 提现银行账户
type WithdrawalBankAccount struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"merchant_id"`
	BankName        string     `gorm:"type:varchar(100);not null" json:"bank_name"`
	BankCode        string     `gorm:"type:varchar(20);not null" json:"bank_code"`          // 银行代码
	BankBranch      string     `gorm:"type:varchar(200)" json:"bank_branch"`                // 开户行
	AccountName     string     `gorm:"type:varchar(100);not null" json:"account_name"`      // 账户名
	AccountNo       string     `gorm:"type:varchar(64);not null" json:"account_no"`         // 账号
	AccountType     string     `gorm:"type:varchar(20);not null;default:'corporate'" json:"account_type"` // 账户类型：corporate, personal
	IsDefault       bool       `gorm:"not null;default:false" json:"is_default"`            // 是否默认账户
	IsVerified      bool       `gorm:"not null;default:false" json:"is_verified"`           // 是否已验证
	VerificationDoc string     `gorm:"type:varchar(500)" json:"verification_doc"`           // 验证文档
	Status          string     `gorm:"type:varchar(20);not null;default:'active'" json:"status"` // active, inactive, suspended
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (WithdrawalBankAccount) TableName() string {
	return "withdrawal_bank_accounts"
}

// WithdrawalApproval 提现审批记录
type WithdrawalApproval struct {
	ID           uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WithdrawalID uuid.UUID        `gorm:"type:uuid;index;not null" json:"withdrawal_id"`
	ApproverID   uuid.UUID        `gorm:"type:uuid;not null" json:"approver_id"`
	ApproverName string           `gorm:"type:varchar(100);not null" json:"approver_name"`
	Level        int              `gorm:"not null" json:"level"`                            // 审批级别
	Action       string           `gorm:"type:varchar(20);not null" json:"action"`          // approve, reject
	Status       WithdrawalStatus `gorm:"type:varchar(20);not null" json:"status"`
	Comments     string           `gorm:"type:text" json:"comments"`
	ApprovedAt   time.Time        `gorm:"not null" json:"approved_at"`
	CreatedAt    time.Time        `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (WithdrawalApproval) TableName() string {
	return "withdrawal_approvals"
}

// WithdrawalBatch 批量提现
type WithdrawalBatch struct {
	ID            uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BatchNo       string           `gorm:"type:varchar(64);uniqueIndex;not null" json:"batch_no"`
	MerchantID    uuid.UUID        `gorm:"type:uuid;index;not null" json:"merchant_id"`
	TotalCount    int              `gorm:"not null" json:"total_count"`
	TotalAmount   int64            `gorm:"not null" json:"total_amount"`
	TotalFee      int64            `gorm:"not null" json:"total_fee"`
	SuccessCount  int              `gorm:"not null;default:0" json:"success_count"`
	FailureCount  int              `gorm:"not null;default:0" json:"failure_count"`
	Status        WithdrawalStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ProcessedAt   *time.Time       `json:"processed_at"`
	CompletedAt   *time.Time       `json:"completed_at"`
	CreatedBy     uuid.UUID        `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt     time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (WithdrawalBatch) TableName() string {
	return "withdrawal_batches"
}

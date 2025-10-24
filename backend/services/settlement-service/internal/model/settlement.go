package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SettlementStatus 结算状态
type SettlementStatus string

const (
	SettlementStatusPending   SettlementStatus = "pending"   // 待审批
	SettlementStatusApproved  SettlementStatus = "approved"  // 已审批
	SettlementStatusRejected  SettlementStatus = "rejected"  // 已拒绝
	SettlementStatusProcessing SettlementStatus = "processing" // 处理中
	SettlementStatusCompleted SettlementStatus = "completed" // 已完成
	SettlementStatusFailed    SettlementStatus = "failed"    // 失败
)

// SettlementCycle 结算周期
type SettlementCycle string

const (
	SettlementCycleDaily   SettlementCycle = "daily"   // 每日
	SettlementCycleWeekly  SettlementCycle = "weekly"  // 每周
	SettlementCycleMonthly SettlementCycle = "monthly" // 每月
	SettlementCycleManual  SettlementCycle = "manual"  // 手动
)

// Settlement 结算单
type Settlement struct {
	ID              uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SettlementNo    string           `gorm:"type:varchar(64);uniqueIndex;not null" json:"settlement_no"` // 结算单号
	MerchantID      uuid.UUID        `gorm:"type:uuid;index;not null" json:"merchant_id"`                // 商户ID
	Cycle           SettlementCycle  `gorm:"type:varchar(20);not null" json:"cycle"`                     // 结算周期
	StartDate       time.Time        `gorm:"not null" json:"start_date"`                                 // 开始日期
	EndDate         time.Time        `gorm:"not null" json:"end_date"`                                   // 结束日期
	TotalAmount     int64            `gorm:"not null;default:0" json:"total_amount"`                     // 交易总额（分）
	TotalCount      int              `gorm:"not null;default:0" json:"total_count"`                      // 交易笔数
	FeeAmount       int64            `gorm:"not null;default:0" json:"fee_amount"`                       // 手续费（分）
	RefundAmount    int64            `gorm:"not null;default:0" json:"refund_amount"`                    // 退款金额（分）
	RefundCount     int              `gorm:"not null;default:0" json:"refund_count"`                     // 退款笔数
	SettlementAmount int64           `gorm:"not null;default:0" json:"settlement_amount"`                // 结算金额（分）
	Status          SettlementStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // 状态
	WithdrawalNo    string           `gorm:"type:varchar(64);index" json:"withdrawal_no"`                // 提现单号
	ApprovedAt      *time.Time       `json:"approved_at"`                                                // 审批时间
	ApprovedBy      *uuid.UUID       `gorm:"type:uuid" json:"approved_by"`                               // 审批人ID
	ProcessedAt     *time.Time       `json:"processed_at"`                                               // 处理时间
	CompletedAt     *time.Time       `json:"completed_at"`                                               // 完成时间
	Remarks         string           `gorm:"type:text" json:"remarks"`                                   // 备注
	ErrorMessage    string           `gorm:"type:text" json:"error_message"`                             // 错误信息
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Settlement) TableName() string {
	return "settlements"
}

// SettlementItem 结算明细
type SettlementItem struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SettlementID  uuid.UUID `gorm:"type:uuid;index;not null" json:"settlement_id"`      // 结算单ID
	TransactionID uuid.UUID `gorm:"type:uuid;index;not null" json:"transaction_id"`     // 交易ID
	OrderNo       string    `gorm:"type:varchar(64);index;not null" json:"order_no"`    // 订单号
	PaymentNo     string    `gorm:"type:varchar(64);index;not null" json:"payment_no"`  // 支付单号
	Amount        int64     `gorm:"not null" json:"amount"`                              // 交易金额（分）
	Fee           int64     `gorm:"not null;default:0" json:"fee"`                       // 手续费（分）
	SettleAmount  int64     `gorm:"not null" json:"settle_amount"`                       // 结算金额（分）
	TransactionAt time.Time `gorm:"not null" json:"transaction_at"`                     // 交易时间
	CreatedAt     time.Time `json:"created_at"`
}

// TableName 指定表名
func (SettlementItem) TableName() string {
	return "settlement_items"
}

// SettlementApproval 结算审批记录
type SettlementApproval struct {
	ID           uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SettlementID uuid.UUID        `gorm:"type:uuid;index;not null" json:"settlement_id"`             // 结算单ID
	ApproverID   uuid.UUID        `gorm:"type:uuid;not null" json:"approver_id"`                     // 审批人ID
	ApproverName string           `gorm:"type:varchar(100);not null" json:"approver_name"`           // 审批人名称
	Action       string           `gorm:"type:varchar(20);not null" json:"action"`                   // 操作（approve/reject）
	Status       SettlementStatus `gorm:"type:varchar(20);not null" json:"status"`                   // 审批后状态
	Comments     string           `gorm:"type:text" json:"comments"`                                 // 审批意见
	ApprovedAt   time.Time        `gorm:"not null" json:"approved_at"`                               // 审批时间
	CreatedAt    time.Time        `json:"created_at"`
}

// TableName 指定表名
func (SettlementApproval) TableName() string {
	return "settlement_approvals"
}

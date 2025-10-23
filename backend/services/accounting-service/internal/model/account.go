package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Account 账户表
type Account struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`           // 商户ID
	AccountType   string         `gorm:"type:varchar(50);not null" json:"account_type"`         // 账户类型：operating(运营账户)、reserve(准备金)、settlement(结算账户)
	Currency      string         `gorm:"type:varchar(10);not null" json:"currency"`             // 货币类型
	Balance       int64          `gorm:"type:bigint;default:0" json:"balance"`                  // 余额（分）
	FrozenBalance int64          `gorm:"type:bigint;default:0" json:"frozen_balance"`           // 冻结余额（分）
	TotalIn       int64          `gorm:"type:bigint;default:0" json:"total_in"`                 // 累计收入
	TotalOut      int64          `gorm:"type:bigint;default:0" json:"total_out"`                // 累计支出
	Status        string         `gorm:"type:varchar(20);default:'active'" json:"status"`       // 状态：active, frozen, closed
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Account) TableName() string {
	return "accounts"
}

// AccountTransaction 账户交易记录
type AccountTransaction struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AccountID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"account_id"`                     // 账户ID
	MerchantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`                    // 商户ID
	TransactionNo   string         `gorm:"type:varchar(64);unique;not null;index" json:"transaction_no"`   // 交易流水号
	TransactionType string         `gorm:"type:varchar(50);not null" json:"transaction_type"`              // 交易类型：payment_in, refund_out, withdraw, fee, adjustment
	RelatedID       uuid.UUID      `gorm:"type:uuid;index" json:"related_id"`                              // 关联ID（支付ID、退款ID等）
	RelatedNo       string         `gorm:"type:varchar(64);index" json:"related_no"`                       // 关联单号
	Amount          int64          `gorm:"type:bigint;not null" json:"amount"`                             // 交易金额（分，正数为入账，负数为出账）
	BalanceBefore   int64          `gorm:"type:bigint;not null" json:"balance_before"`                     // 交易前余额
	BalanceAfter    int64          `gorm:"type:bigint;not null" json:"balance_after"`                      // 交易后余额
	Currency        string         `gorm:"type:varchar(10);not null" json:"currency"`                      // 货币类型
	Description     string         `gorm:"type:text" json:"description"`                                   // 描述
	Status          string         `gorm:"type:varchar(20);default:'completed'" json:"status"`             // 状态：pending, completed, failed, reversed
	CreatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AccountTransaction) TableName() string {
	return "account_transactions"
}

// Settlement 结算记录
type Settlement struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`        // 商户ID
	SettlementNo  string         `gorm:"type:varchar(64);unique;not null;index" json:"settlement_no"` // 结算单号
	AccountID     uuid.UUID      `gorm:"type:uuid;not null" json:"account_id"`               // 账户ID
	PeriodStart   time.Time      `gorm:"type:timestamptz;not null" json:"period_start"`      // 结算周期开始
	PeriodEnd     time.Time      `gorm:"type:timestamptz;not null" json:"period_end"`        // 结算周期结束
	TotalAmount   int64          `gorm:"type:bigint;not null" json:"total_amount"`           // 结算总额
	FeeAmount     int64          `gorm:"type:bigint;default:0" json:"fee_amount"`            // 手续费
	NetAmount     int64          `gorm:"type:bigint;not null" json:"net_amount"`             // 净额（总额-手续费）
	Currency      string         `gorm:"type:varchar(10);not null" json:"currency"`          // 货币类型
	Status        string         `gorm:"type:varchar(20);not null" json:"status"`            // 状态：pending, processing, completed, failed
	PaymentCount  int            `gorm:"type:integer" json:"payment_count"`                  // 支付笔数
	RefundCount   int            `gorm:"type:integer" json:"refund_count"`                   // 退款笔数
	SettledAt     *time.Time     `gorm:"type:timestamptz" json:"settled_at"`                 // 结算完成时间
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Settlement) TableName() string {
	return "settlements"
}

// DoubleEntry 复式记账
type DoubleEntry struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EntryNo     string         `gorm:"type:varchar(64);unique;not null;index" json:"entry_no"`   // 分录号
	RelatedID   uuid.UUID      `gorm:"type:uuid;index" json:"related_id"`                        // 关联ID
	RelatedNo   string         `gorm:"type:varchar(64);index" json:"related_no"`                 // 关联单号
	EntryType   string         `gorm:"type:varchar(50);not null" json:"entry_type"`              // 分录类型
	DebitAccount  string       `gorm:"type:varchar(100);not null" json:"debit_account"`          // 借方科目
	CreditAccount string       `gorm:"type:varchar(100);not null" json:"credit_account"`         // 贷方科目
	Amount      int64          `gorm:"type:bigint;not null" json:"amount"`                       // 金额
	Currency    string         `gorm:"type:varchar(10);not null" json:"currency"`                // 货币
	Description string         `gorm:"type:text" json:"description"`                             // 描述
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (DoubleEntry) TableName() string {
	return "double_entries"
}

// 账户类型常量
const (
	AccountTypeOperating  = "operating"  // 运营账户
	AccountTypeReserve    = "reserve"    // 准备金账户
	AccountTypeSettlement = "settlement" // 结算账户
)

// 账户状态常量
const (
	AccountStatusActive = "active" // 活跃
	AccountStatusFrozen = "frozen" // 冻结
	AccountStatusClosed = "closed" // 关闭
)

// 交易类型常量
const (
	TransactionTypePaymentIn  = "payment_in"  // 支付入账
	TransactionTypeRefundOut  = "refund_out"  // 退款出账
	TransactionTypeWithdraw   = "withdraw"    // 提现
	TransactionTypeFee        = "fee"         // 手续费
	TransactionTypeAdjustment = "adjustment"  // 调账
)

// 结算状态常量
const (
	SettlementStatusPending    = "pending"    // 待结算
	SettlementStatusProcessing = "processing" // 处理中
	SettlementStatusCompleted  = "completed"  // 已完成
	SettlementStatusFailed     = "failed"     // 失败
)

// Withdrawal 提现记录
type Withdrawal struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`
	WithdrawalNo        string         `gorm:"type:varchar(64);unique;not null;index" json:"withdrawal_no"`    // 提现单号
	AccountID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"account_id"`                     // 提现账户ID
	SettlementAccountID uuid.UUID      `gorm:"type:uuid;not null" json:"settlement_account_id"`                // 结算账户ID（merchant-service的settlement_accounts表）
	Amount              int64          `gorm:"type:bigint;not null" json:"amount"`                             // 提现金额（分）
	Currency            string         `gorm:"type:varchar(10);not null" json:"currency"`                      // 货币类型
	FeeAmount           int64          `gorm:"type:bigint;default:0" json:"fee_amount"`                        // 手续费（分）
	ActualAmount        int64          `gorm:"type:bigint;not null" json:"actual_amount"`                      // 实际到账金额（分）
	Status              string         `gorm:"type:varchar(20);not null" json:"status"`                        // 状态
	RequestReason       string         `gorm:"type:text" json:"request_reason"`                                // 申请理由
	ApprovalNotes       string         `gorm:"type:text" json:"approval_notes"`                                // 审批备注
	ApprovedBy          *uuid.UUID     `gorm:"type:uuid" json:"approved_by"`                                   // 审批人ID
	ApprovedAt          *time.Time     `gorm:"type:timestamptz" json:"approved_at"`                            // 审批时间
	ProcessedBy         *uuid.UUID     `gorm:"type:uuid" json:"processed_by"`                                  // 处理人ID
	ProcessedAt         *time.Time     `gorm:"type:timestamptz" json:"processed_at"`                           // 处理时间
	CompletedAt         *time.Time     `gorm:"type:timestamptz" json:"completed_at"`                           // 完成时间
	FailureReason       string         `gorm:"type:text" json:"failure_reason"`                                // 失败原因
	TransactionID       *uuid.UUID     `gorm:"type:uuid" json:"transaction_id"`                                // 关联的账户交易ID
	CreatedAt           time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Withdrawal) TableName() string {
	return "withdrawals"
}

// 提现状态常量
const (
	WithdrawalStatusPending    = "pending"    // 待审核
	WithdrawalStatusApproved   = "approved"   // 已批准
	WithdrawalStatusRejected   = "rejected"   // 已拒绝
	WithdrawalStatusProcessing = "processing" // 处理中
	WithdrawalStatusCompleted  = "completed"  // 已完成
	WithdrawalStatusFailed     = "failed"     // 失败
	WithdrawalStatusCancelled  = "cancelled"  // 已取消
)

// Invoice 账单/发票
type Invoice struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`
	InvoiceNo       string         `gorm:"type:varchar(64);unique;not null;index" json:"invoice_no"`       // 账单号
	InvoiceType     string         `gorm:"type:varchar(50);not null" json:"invoice_type"`                  // 账单类型：service_fee(服务费)、transaction_fee(交易手续费)、withdrawal_fee(提现手续费)
	PeriodStart     time.Time      `gorm:"type:timestamptz;not null" json:"period_start"`                  // 账期开始
	PeriodEnd       time.Time      `gorm:"type:timestamptz;not null" json:"period_end"`                    // 账期结束
	Currency        string         `gorm:"type:varchar(10);not null" json:"currency"`                      // 货币类型
	SubtotalAmount  int64          `gorm:"type:bigint;not null" json:"subtotal_amount"`                    // 小计金额（分）
	TaxAmount       int64          `gorm:"type:bigint;default:0" json:"tax_amount"`                        // 税额（分）
	TotalAmount     int64          `gorm:"type:bigint;not null" json:"total_amount"`                       // 总金额（分）
	PaidAmount      int64          `gorm:"type:bigint;default:0" json:"paid_amount"`                       // 已支付金额（分）
	OutstandingAmount int64        `gorm:"type:bigint;not null" json:"outstanding_amount"`                 // 未付金额（分）
	Status          string         `gorm:"type:varchar(20);not null" json:"status"`                        // 状态
	DueDate         time.Time      `gorm:"type:timestamptz;not null" json:"due_date"`                      // 到期日
	PaidAt          *time.Time     `gorm:"type:timestamptz" json:"paid_at"`                                // 支付时间
	Notes           string         `gorm:"type:text" json:"notes"`                                         // 备注
	Items           []InvoiceItem  `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`                    // 账单明细
	CreatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceItem 账单明细
type InvoiceItem struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InvoiceID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"invoice_id"`             // 账单ID
	ItemType    string         `gorm:"type:varchar(50);not null" json:"item_type"`             // 明细类型
	Description string         `gorm:"type:text" json:"description"`                           // 描述
	Quantity    int            `gorm:"type:integer;default:1" json:"quantity"`                 // 数量
	UnitPrice   int64          `gorm:"type:bigint;not null" json:"unit_price"`                 // 单价（分）
	Amount      int64          `gorm:"type:bigint;not null" json:"amount"`                     // 金额（分）
	RelatedID   *uuid.UUID     `gorm:"type:uuid" json:"related_id"`                            // 关联ID
	RelatedNo   string         `gorm:"type:varchar(64)" json:"related_no"`                     // 关联单号
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// 账单状态常量
const (
	InvoiceStatusDraft     = "draft"     // 草稿
	InvoiceStatusPending   = "pending"   // 待支付
	InvoiceStatusPaid      = "paid"      // 已支付
	InvoiceStatusPartialPaid = "partial_paid" // 部分支付
	InvoiceStatusOverdue   = "overdue"   // 已逾期
	InvoiceStatusCancelled = "cancelled" // 已取消
	InvoiceStatusVoided    = "voided"    // 已作废
)

// 账单类型常量
const (
	InvoiceTypeServiceFee     = "service_fee"     // 服务费
	InvoiceTypeTransactionFee = "transaction_fee" // 交易手续费
	InvoiceTypeWithdrawalFee  = "withdrawal_fee"  // 提现手续费
	InvoiceTypeMonthly        = "monthly"         // 月度账单
	InvoiceTypeCustom         = "custom"          // 自定义账单
)

// Reconciliation 对账单
type Reconciliation struct {
	ID                 uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReconciliationNo   string               `gorm:"type:varchar(64);unique;not null;index" json:"reconciliation_no"` // 对账单号
	MerchantID         uuid.UUID            `gorm:"type:uuid;not null;index" json:"merchant_id"`
	Channel            string               `gorm:"type:varchar(50);not null" json:"channel"`                         // 渠道：stripe, paypal, bank, etc
	ReconciliationDate time.Time            `gorm:"type:date;not null;index" json:"reconciliation_date"`              // 对账日期
	PeriodStart        time.Time            `gorm:"type:timestamptz;not null" json:"period_start"`                    // 对账周期开始
	PeriodEnd          time.Time            `gorm:"type:timestamptz;not null" json:"period_end"`                      // 对账周期结束
	Currency           string               `gorm:"type:varchar(10);not null" json:"currency"`

	// 内部数据（我方系统）
	InternalCount      int                  `gorm:"type:integer;default:0" json:"internal_count"`              // 内部交易笔数
	InternalAmount     int64                `gorm:"type:bigint;default:0" json:"internal_amount"`              // 内部交易总额（分）

	// 外部数据（渠道方）
	ExternalCount      int                  `gorm:"type:integer;default:0" json:"external_count"`              // 外部交易笔数
	ExternalAmount     int64                `gorm:"type:bigint;default:0" json:"external_amount"`              // 外部交易总额（分）

	// 差异数据
	DiffCount          int                  `gorm:"type:integer;default:0" json:"diff_count"`                  // 差异笔数
	DiffAmount         int64                `gorm:"type:bigint;default:0" json:"diff_amount"`                  // 差异金额（分）
	MismatchedCount    int                  `gorm:"type:integer;default:0" json:"mismatched_count"`            // 不匹配笔数

	Status             string               `gorm:"type:varchar(20);not null" json:"status"`                   // 状态
	ReconciledBy       *uuid.UUID           `gorm:"type:uuid" json:"reconciled_by"`                            // 对账人
	ReconciledAt       *time.Time           `gorm:"type:timestamptz" json:"reconciled_at"`                     // 对账完成时间
	Notes              string               `gorm:"type:text" json:"notes"`                                    // 备注
	Items              []ReconciliationItem `gorm:"foreignKey:ReconciliationID" json:"items,omitempty"`        // 对账明细
	CreatedAt          time.Time            `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt          time.Time            `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt          gorm.DeletedAt       `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Reconciliation) TableName() string {
	return "reconciliations"
}

// ReconciliationItem 对账明细
type ReconciliationItem struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReconciliationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"reconciliation_id"`        // 对账单ID
	TransactionNo    string         `gorm:"type:varchar(64);not null;index" json:"transaction_no"`    // 内部交易号
	ExternalTxNo     string         `gorm:"type:varchar(128)" json:"external_tx_no"`                  // 外部交易号
	ItemType         string         `gorm:"type:varchar(50);not null" json:"item_type"`               // 明细类型
	InternalAmount   int64          `gorm:"type:bigint" json:"internal_amount"`                       // 内部金额（分）
	ExternalAmount   int64          `gorm:"type:bigint" json:"external_amount"`                       // 外部金额（分）
	DiffAmount       int64          `gorm:"type:bigint;default:0" json:"diff_amount"`                 // 差异金额（分）
	Status           string         `gorm:"type:varchar(20);not null" json:"status"`                  // 状态
	Description      string         `gorm:"type:text" json:"description"`                             // 描述/差异说明
	CreatedAt        time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ReconciliationItem) TableName() string {
	return "reconciliation_items"
}

// 对账状态常量
const (
	ReconciliationStatusPending    = "pending"    // 待对账
	ReconciliationStatusProcessing = "processing" // 对账中
	ReconciliationStatusMatched    = "matched"    // 完全匹配
	ReconciliationStatusMismatched = "mismatched" // 存在差异
	ReconciliationStatusCompleted  = "completed"  // 已完成（已处理差异）
	ReconciliationStatusCancelled  = "cancelled"  // 已取消
)

// 对账明细状态常量
const (
	ReconciliationItemStatusMatched    = "matched"    // 匹配
	ReconciliationItemStatusMismatched = "mismatched" // 不匹配
	ReconciliationItemStatusMissing    = "missing"    // 缺失（仅一方有）
	ReconciliationItemStatusDuplicate  = "duplicate"  // 重复
	ReconciliationItemStatusResolved   = "resolved"   // 已解决
)

// 对账明细类型常量
const (
	ReconciliationItemTypePayment = "payment" // 支付
	ReconciliationItemTypeRefund  = "refund"  // 退款
	ReconciliationItemTypeFee     = "fee"     // 手续费
)

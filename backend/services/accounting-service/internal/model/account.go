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

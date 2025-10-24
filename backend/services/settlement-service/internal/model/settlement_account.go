package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SettlementAccount 结算账户表
type SettlementAccount struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`                  // 商户ID
	AccountType        string         `gorm:"type:varchar(50);not null" json:"account_type"`                // bank_account, paypal, crypto_wallet, alipay, wechat
	BankName           string         `gorm:"type:varchar(200)" json:"bank_name"`                           // 银行名称
	BankCode           string         `gorm:"type:varchar(50)" json:"bank_code"`                            // 银行代码
	AccountNumber      string         `gorm:"type:varchar(255);not null" json:"account_number"`             // 账号（加密存储）
	AccountName        string         `gorm:"type:varchar(200);not null" json:"account_name"`               // 账户名
	SwiftCode          string         `gorm:"type:varchar(50)" json:"swift_code"`                           // SWIFT代码
	IBAN               string         `gorm:"type:varchar(50)" json:"iban"`                                 // IBAN
	BankAddress        string         `gorm:"type:varchar(500)" json:"bank_address"`                        // 银行地址
	Currency           string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`      // 币种
	Country            string         `gorm:"type:varchar(50)" json:"country"`                              // 国家
	IsDefault          bool           `gorm:"default:false" json:"is_default"`                              // 是否默认账户
	Status             string         `gorm:"type:varchar(20);default:'pending_verify'" json:"status"`      // pending_verify, verified, rejected, suspended
	VerificationMethod string         `gorm:"type:varchar(50)" json:"verification_method"`                  // micro_deposit, manual, auto
	VerifiedAt         *time.Time     `gorm:"type:timestamptz" json:"verified_at"`                          // 验证时间
	VerificationData   string         `gorm:"type:jsonb" json:"verification_data"`                          // 验证相关数据（JSON）
	RejectReason       string         `gorm:"type:text" json:"reject_reason"`                               // 拒绝原因
	Metadata           string         `gorm:"type:jsonb" json:"metadata"`                                   // 扩展信息
	CreatedAt          time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (SettlementAccount) TableName() string {
	return "settlement_accounts"
}

// 结算账户类型常量
const (
	AccountTypeBankAccount  = "bank_account"  // 银行账户
	AccountTypePayPal       = "paypal"        // PayPal
	AccountTypeCryptoWallet = "crypto_wallet" // 加密货币钱包
	AccountTypeAlipay       = "alipay"        // 支付宝
	AccountTypeWechat       = "wechat"        // 微信支付
)

// 结算账户状态常量
const (
	AccountStatusPendingVerify = "pending_verify" // 待验证
	AccountStatusVerified      = "verified"       // 已验证
	AccountStatusRejected      = "rejected"       // 已拒绝
	AccountStatusSuspended     = "suspended"      // 已暂停
)

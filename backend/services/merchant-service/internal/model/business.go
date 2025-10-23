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

// KYCDocument KYC文档表
type KYCDocument struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`               // 商户ID
	DocumentType string         `gorm:"type:varchar(50);not null;index" json:"document_type"`      // id_card, passport, business_license, tax_certificate, bank_statement
	FileURL      string         `gorm:"type:varchar(500);not null" json:"file_url"`                // 文件URL
	FileHash     string         `gorm:"type:varchar(128)" json:"file_hash"`                        // 文件哈希（用于校验）
	FileSize     int64          `gorm:"default:0" json:"file_size"`                                // 文件大小（字节）
	MimeType     string         `gorm:"type:varchar(100)" json:"mime_type"`                        // MIME类型
	Status       string         `gorm:"type:varchar(20);default:'pending';index" json:"status"`    // pending, approved, rejected
	ReviewedBy   *uuid.UUID     `gorm:"type:uuid" json:"reviewed_by"`                              // 审核人ID
	ReviewNotes  string         `gorm:"type:text" json:"review_notes"`                             // 审核备注
	ReviewedAt   *time.Time     `gorm:"type:timestamptz" json:"reviewed_at"`                       // 审核时间
	ExpiryDate   *time.Time     `gorm:"type:date" json:"expiry_date"`                              // 过期日期（如证件有效期）
	OCRData      string         `gorm:"type:jsonb" json:"ocr_data"`                                // OCR识别数据
	Metadata     string         `gorm:"type:jsonb" json:"metadata"`                                // 扩展信息
	CreatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (KYCDocument) TableName() string {
	return "kyc_documents"
}

// BusinessQualification 业务资质表
type BusinessQualification struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`           // 商户ID
	QualificationType string         `gorm:"type:varchar(50);not null" json:"qualification_type"`   // icp_license, payment_license, food_permit, medical_license
	LicenseNumber     string         `gorm:"type:varchar(100);not null" json:"license_number"`      // 证照编号
	LicenseName       string         `gorm:"type:varchar(200)" json:"license_name"`                 // 证照名称
	IssuedBy          string         `gorm:"type:varchar(200)" json:"issued_by"`                    // 发证机关
	IssuedDate        *time.Time     `gorm:"type:date" json:"issued_date"`                          // 发证日期
	ExpiryDate        *time.Time     `gorm:"type:date" json:"expiry_date"`                          // 到期日期（null表示长期有效）
	FileURL           string         `gorm:"type:varchar(500)" json:"file_url"`                     // 证照文件URL
	Status            string         `gorm:"type:varchar(20);default:'pending'" json:"status"`      // pending, verified, rejected, expired
	VerifiedBy        *uuid.UUID     `gorm:"type:uuid" json:"verified_by"`                          // 验证人
	VerifiedAt        *time.Time     `gorm:"type:timestamptz" json:"verified_at"`                   // 验证时间
	Metadata          string         `gorm:"type:jsonb" json:"metadata"`                            // 扩展信息
	CreatedAt         time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (BusinessQualification) TableName() string {
	return "business_qualifications"
}

// MerchantFeeConfig 商户费率配置表
type MerchantFeeConfig struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`           // 商户ID
	Channel        string         `gorm:"type:varchar(50);not null;index" json:"channel"`        // 渠道：stripe, paypal, all
	PaymentMethod  string         `gorm:"type:varchar(50)" json:"payment_method"`                // 支付方式：card, bank_transfer, all
	FeeType        string         `gorm:"type:varchar(20);not null" json:"fee_type"`             // percentage, fixed, tiered
	FeePercentage  float64        `gorm:"type:decimal(5,4);default:0" json:"fee_percentage"`     // 费率百分比（如0.029表示2.9%）
	FeeFixed       int64          `gorm:"type:bigint;default:0" json:"fee_fixed"`                // 固定费用（分）
	MinFee         int64          `gorm:"type:bigint;default:0" json:"min_fee"`                  // 最小费用（分）
	MaxFee         int64          `gorm:"type:bigint" json:"max_fee"`                            // 最大费用（分，null表示无上限）
	Currency       string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"` // 币种
	TieredRules    string         `gorm:"type:jsonb" json:"tiered_rules"`                        // 阶梯费率规则（JSON）
	EffectiveDate  time.Time      `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"` // 生效日期
	ExpiryDate     *time.Time     `gorm:"type:timestamptz" json:"expiry_date"`                   // 失效日期（null表示长期有效）
	Priority       int            `gorm:"default:0" json:"priority"`                             // 优先级（数字越大优先级越高）
	Status         string         `gorm:"type:varchar(20);default:'active'" json:"status"`       // active, inactive
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by"`                           // 创建人
	ApprovedBy     *uuid.UUID     `gorm:"type:uuid" json:"approved_by"`                          // 审批人
	ApprovedAt     *time.Time     `gorm:"type:timestamptz" json:"approved_at"`                   // 审批时间
	Metadata       string         `gorm:"type:jsonb" json:"metadata"`                            // 扩展信息
	CreatedAt      time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantFeeConfig) TableName() string {
	return "merchant_fee_configs"
}

// MerchantUser 商户子账户/团队成员表
type MerchantUser struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`               // 商户ID
	Email           string         `gorm:"type:varchar(255);not null;index" json:"email"`             // 邮箱
	PasswordHash    string         `gorm:"type:varchar(255)" json:"-"`                                // 密码哈希（仅用于登录，可选）
	Name            string         `gorm:"type:varchar(100);not null" json:"name"`                    // 姓名
	Phone           string         `gorm:"type:varchar(20)" json:"phone"`                             // 电话
	Role            string         `gorm:"type:varchar(50);not null;index" json:"role"`               // admin, finance, developer, support, viewer
	Permissions     string         `gorm:"type:jsonb" json:"permissions"`                             // 权限列表（JSON数组）
	Status          string         `gorm:"type:varchar(20);default:'pending';index" json:"status"`    // pending, active, suspended, deleted
	InvitedBy       *uuid.UUID     `gorm:"type:uuid" json:"invited_by"`                               // 邀请人ID
	InvitedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"invited_at"`          // 邀请时间
	AcceptedAt      *time.Time     `gorm:"type:timestamptz" json:"accepted_at"`                       // 接受邀请时间
	LastLoginAt     *time.Time     `gorm:"type:timestamptz" json:"last_login_at"`                     // 最后登录时间
	LastLoginIP     string         `gorm:"type:varchar(50)" json:"last_login_ip"`                     // 最后登录IP
	TwoFactorEnabled bool          `gorm:"default:false" json:"two_factor_enabled"`                   // 是否启用2FA
	Metadata        string         `gorm:"type:jsonb" json:"metadata"`                                // 扩展信息
	CreatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantUser) TableName() string {
	return "merchant_users"
}

// MerchantTransactionLimit 商户交易限额配置表
type MerchantTransactionLimit struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`           // 商户ID
	LimitType     string         `gorm:"type:varchar(20);not null;index" json:"limit_type"`     // single, daily, monthly
	PaymentMethod string         `gorm:"type:varchar(50)" json:"payment_method"`                // card, bank_transfer, all
	Channel       string         `gorm:"type:varchar(50)" json:"channel"`                       // stripe, paypal, all
	Currency      string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"` // 币种
	MinAmount     int64          `gorm:"type:bigint;default:0" json:"min_amount"`               // 最小金额（分）
	MaxAmount     int64          `gorm:"type:bigint" json:"max_amount"`                         // 最大金额（分）
	MaxCount      int            `gorm:"default:0" json:"max_count"`                            // 最大笔数（0表示不限制）
	Status        string         `gorm:"type:varchar(20);default:'active'" json:"status"`       // active, inactive
	EffectiveDate time.Time      `gorm:"type:timestamptz;not null;default:now()" json:"effective_date"` // 生效日期
	ExpiryDate    *time.Time     `gorm:"type:timestamptz" json:"expiry_date"`                   // 失效日期
	Metadata      string         `gorm:"type:jsonb" json:"metadata"`                            // 扩展信息
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantTransactionLimit) TableName() string {
	return "merchant_transaction_limits"
}

// MerchantContract 商户合同协议表
type MerchantContract struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`          // 商户ID
	ContractType string         `gorm:"type:varchar(50);not null" json:"contract_type"`       // service_agreement, supplemental, amendment
	ContractNo   string         `gorm:"type:varchar(100);unique;not null;index" json:"contract_no"` // 合同编号
	ContractName string         `gorm:"type:varchar(200)" json:"contract_name"`               // 合同名称
	SignedAt     *time.Time     `gorm:"type:timestamptz" json:"signed_at"`                    // 签署时间
	EffectiveDate time.Time     `gorm:"type:timestamptz;not null" json:"effective_date"`      // 生效日期
	ExpiryDate   *time.Time     `gorm:"type:timestamptz" json:"expiry_date"`                  // 到期日期
	FileURL      string         `gorm:"type:varchar(500)" json:"file_url"`                    // 合同文件URL
	FileHash     string         `gorm:"type:varchar(128)" json:"file_hash"`                   // 文件哈希
	Status       string         `gorm:"type:varchar(20);default:'draft'" json:"status"`       // draft, signed, active, expired, terminated
	SignMethod   string         `gorm:"type:varchar(50)" json:"sign_method"`                  // electronic, paper, both
	PartyA       string         `gorm:"type:varchar(200)" json:"party_a"`                     // 甲方（平台）
	PartyB       string         `gorm:"type:varchar(200)" json:"party_b"`                     // 乙方（商户）
	Metadata     string         `gorm:"type:jsonb" json:"metadata"`                           // 扩展信息
	CreatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantContract) TableName() string {
	return "merchant_contracts"
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

// KYC文档类型常量
const (
	DocumentTypeIDCard          = "id_card"          // 身份证
	DocumentTypePassport        = "passport"         // 护照
	DocumentTypeBusinessLicense = "business_license" // 营业执照
	DocumentTypeTaxCertificate  = "tax_certificate"  // 税务登记证
	DocumentTypeBankStatement   = "bank_statement"   // 银行流水
	DocumentTypeAuthorization   = "authorization"    // 授权书
)

// 文档状态常量
const (
	DocumentStatusPending  = "pending"  // 待审核
	DocumentStatusApproved = "approved" // 已通过
	DocumentStatusRejected = "rejected" // 已拒绝
)

// 资质类型常量
const (
	QualificationTypeICP     = "icp_license"     // ICP许可证
	QualificationTypePayment = "payment_license" // 支付牌照
	QualificationTypeFood    = "food_permit"     // 食品经营许可证
	QualificationTypeMedical = "medical_license" // 医疗许可证
)

// 资质状态常量
const (
	QualificationStatusPending  = "pending"  // 待审核
	QualificationStatusVerified = "verified" // 已验证
	QualificationStatusRejected = "rejected" // 已拒绝
	QualificationStatusExpired  = "expired"  // 已过期
)

// 费率类型常量
const (
	FeeTypePercentage = "percentage" // 百分比费率
	FeeTypeFixed      = "fixed"      // 固定费用
	FeeTypeTiered     = "tiered"     // 阶梯费率
)

// 子账户角色常量
const (
	UserRoleAdmin     = "admin"     // 管理员
	UserRoleFinance   = "finance"   // 财务
	UserRoleDeveloper = "developer" // 开发者
	UserRoleSupport   = "support"   // 客服
	UserRoleViewer    = "viewer"    // 只读
)

// 子账户状态常量
const (
	UserStatusPending   = "pending"   // 待接受邀请
	UserStatusActive    = "active"    // 活跃
	UserStatusSuspended = "suspended" // 暂停
	UserStatusDeleted   = "deleted"   // 已删除
)

// 限额类型常量
const (
	LimitTypeSingle  = "single"  // 单笔限额
	LimitTypeDaily   = "daily"   // 日累计限额
	LimitTypeMonthly = "monthly" // 月累计限额
)

// 合同类型常量
const (
	ContractTypeServiceAgreement = "service_agreement" // 服务协议
	ContractTypeSupplemental     = "supplemental"      // 补充协议
	ContractTypeAmendment        = "amendment"         // 修正案
)

// 合同状态常量
const (
	ContractStatusDraft      = "draft"      // 草稿
	ContractStatusSigned     = "signed"     // 已签署
	ContractStatusActive     = "active"     // 生效中
	ContractStatusExpired    = "expired"    // 已过期
	ContractStatusTerminated = "terminated" // 已终止
)

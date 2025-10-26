package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantUser 商户子账户/团队成员表
// Phase 7-8评估决定：保留在merchant-service（属于Merchant聚合根）
type MerchantUser struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`               // 商户ID
	Email            string         `gorm:"type:varchar(255);not null;index" json:"email"`             // 邮箱
	PasswordHash     string         `gorm:"type:varchar(255)" json:"-"`                                // 密码哈希（仅用于登录，可选）
	Name             string         `gorm:"type:varchar(100);not null" json:"name"`                    // 姓名
	Phone            string         `gorm:"type:varchar(20)" json:"phone"`                             // 电话
	Role             string         `gorm:"type:varchar(50);not null;index" json:"role"`               // admin, finance, developer, support, viewer
	Permissions      string         `gorm:"type:jsonb" json:"permissions"`                             // 权限列表（JSON数组）
	Status           string         `gorm:"type:varchar(20);default:'pending';index" json:"status"`    // pending, active, suspended, deleted
	InvitedBy        *uuid.UUID     `gorm:"type:uuid" json:"invited_by"`                               // 邀请人ID
	InvitedAt        time.Time      `gorm:"type:timestamptz;default:now()" json:"invited_at"`          // 邀请时间
	AcceptedAt       *time.Time     `gorm:"type:timestamptz" json:"accepted_at"`                       // 接受邀请时间
	LastLoginAt      *time.Time     `gorm:"type:timestamptz" json:"last_login_at"`                     // 最后登录时间
	LastLoginIP      string         `gorm:"type:varchar(50)" json:"last_login_ip"`                     // 最后登录IP
	TwoFactorEnabled bool           `gorm:"default:false" json:"two_factor_enabled"`                   // 是否启用2FA
	Metadata         string         `gorm:"type:jsonb" json:"metadata"`                                // 扩展信息
	CreatedAt        time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantUser) TableName() string {
	return "merchant_users"
}

// MerchantContract 商户合同协议表
// Phase 7-8评估决定：保留在merchant-service（属于Merchant聚合根）
type MerchantContract struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`                    // 商户ID
	ContractType  string         `gorm:"type:varchar(50);not null" json:"contract_type"`                 // service_agreement, supplemental, amendment
	ContractNo    string         `gorm:"type:varchar(100);unique;not null;index" json:"contract_no"`     // 合同编号
	ContractName  string         `gorm:"type:varchar(200)" json:"contract_name"`                         // 合同名称
	SignedAt      *time.Time     `gorm:"type:timestamptz" json:"signed_at"`                              // 签署时间
	EffectiveDate time.Time      `gorm:"type:timestamptz;not null" json:"effective_date"`                // 生效日期
	ExpiryDate    *time.Time     `gorm:"type:timestamptz" json:"expiry_date"`                            // 到期日期
	FileURL       string         `gorm:"type:varchar(500)" json:"file_url"`                              // 合同文件URL
	FileHash      string         `gorm:"type:varchar(128)" json:"file_hash"`                             // 文件哈希
	Status        string         `gorm:"type:varchar(20);default:'draft'" json:"status"`                 // draft, signed, active, expired, terminated
	SignMethod    string         `gorm:"type:varchar(50)" json:"sign_method"`                            // electronic, paper, both
	PartyA        string         `gorm:"type:varchar(200)" json:"party_a"`                               // 甲方（平台）
	PartyB        string         `gorm:"type:varchar(200)" json:"party_b"`                               // 乙方（商户）
	Metadata      string         `gorm:"type:jsonb" json:"metadata"`                                     // 扩展信息
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MerchantContract) TableName() string {
	return "merchant_contracts"
}

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

// ============================================================================
// 以下模型已迁移到其他服务（Phase 1-6）
// ============================================================================
// - APIKey → merchant-auth-service
// - ChannelConfig → merchant-config-service
// - SettlementAccount → settlement-service
// - KYCDocument → kyc-service
// - BusinessQualification → kyc-service
// - MerchantFeeConfig → merchant-config-service
// - MerchantTransactionLimit → merchant-config-service
// ============================================================================

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Merchant 商户表
type Merchant struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`                      // 商户名称
	Email        string         `gorm:"type:varchar(255);unique;not null;index" json:"email"`        // 邮箱
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`                         // 密码哈希（不返回给前端）
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`                               // 手机号
	CompanyName  string         `gorm:"type:varchar(200)" json:"company_name"`                       // 公司名称
	BusinessType string         `gorm:"type:varchar(50)" json:"business_type"`                       // 业务类型：individual, company
	Country      string         `gorm:"type:varchar(50)" json:"country"`                             // 国家
	Website      string         `gorm:"type:varchar(255)" json:"website"`                            // 网站
	Status       string         `gorm:"type:varchar(20);default:'pending';index" json:"status"`      // pending, active, suspended, rejected
	KYCStatus    string         `gorm:"type:varchar(20);default:'pending';index" json:"kyc_status"`  // pending, verified, rejected
	IsTestMode   bool           `gorm:"default:true" json:"is_test_mode"`                            // 是否测试模式
	Metadata     string         `gorm:"type:jsonb" json:"metadata"`                                  // 扩展元数据（JSON）
	CreatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联（不存储在数据库）
	APIKeys        []APIKey        `gorm:"foreignKey:MerchantID" json:"api_keys,omitempty"`
	ChannelConfigs []ChannelConfig `gorm:"foreignKey:MerchantID" json:"channel_configs,omitempty"`
}

// TableName 指定表名
func (Merchant) TableName() string {
	return "merchants"
}

// APIKey API密钥表
type APIKey struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"merchant_id"`
	APIKey      string     `gorm:"type:varchar(64);unique;not null;index" json:"api_key"`         // API Key（公开）
	APISecret   string     `gorm:"type:varchar(128);not null" json:"api_secret,omitempty"`        // API Secret（仅创建时返回）
	Name        string     `gorm:"type:varchar(100)" json:"name"`                                 // 密钥名称
	Environment string     `gorm:"type:varchar(20);not null;index" json:"environment"`            // test, production
	IsActive    bool       `gorm:"default:true" json:"is_active"`                                 // 是否启用
	LastUsedAt  *time.Time `gorm:"type:timestamptz" json:"last_used_at"`                          // 最后使用时间
	ExpiresAt   *time.Time `gorm:"type:timestamptz" json:"expires_at"`                            // 过期时间（null表示永不过期）
	CreatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// ChannelConfig 支付渠道配置表
type ChannelConfig struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID      `gorm:"type:uuid;not null;index:idx_merchant_channel" json:"merchant_id"`
	Channel    string         `gorm:"type:varchar(50);not null;index:idx_merchant_channel" json:"channel"` // stripe, paypal, crypto
	Config     string         `gorm:"type:jsonb;not null" json:"config"`                                   // 渠道配置（JSON，加密存储敏感信息）
	IsEnabled  bool           `gorm:"default:false" json:"is_enabled"`                                     // 是否启用
	IsTestMode bool           `gorm:"default:true" json:"is_test_mode"`                                    // 是否测试模式
	CreatedAt  time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ChannelConfig) TableName() string {
	return "channel_configs"
}

// 商户状态常量
const (
	MerchantStatusPending   = "pending"   // 待审核
	MerchantStatusActive    = "active"    // 正常
	MerchantStatusSuspended = "suspended" // 暂停
	MerchantStatusRejected  = "rejected"  // 拒绝
)

// KYC状态常量
const (
	KYCStatusPending  = "pending"  // 待审核
	KYCStatusVerified = "verified" // 已验证
	KYCStatusRejected = "rejected" // 拒绝
)

// 业务类型常量
const (
	BusinessTypeIndividual = "individual" // 个人
	BusinessTypeCompany    = "company"    // 公司
)

// API Key环境常量
const (
	EnvironmentTest       = "test"       // 测试环境
	EnvironmentProduction = "production" // 生产环境
)

// 支付渠道常量
const (
	ChannelStripe  = "stripe"  // Stripe
	ChannelPayPal  = "paypal"  // PayPal
	ChannelCrypto  = "crypto"  // 加密货币
	ChannelAdyen   = "adyen"   // Adyen
	ChannelSquare  = "square"  // Square
)

// 唯一索引：一个商户同一渠道只能有一个配置
// CREATE UNIQUE INDEX idx_merchant_channel ON channel_configs(merchant_id, channel) WHERE deleted_at IS NULL;

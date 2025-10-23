package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChannelConfig 渠道配置表
type ChannelConfig struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`        // 商户ID
	Channel     string         `gorm:"type:varchar(50);not null;index" json:"channel"`     // 渠道：stripe, paypal, crypto
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`                     // 是否启用
	Mode        string         `gorm:"type:varchar(20);not null" json:"mode"`              // 模式：test, live
	Config      string         `gorm:"type:jsonb;not null" json:"config"`                  // 配置信息（JSON加密存储）
	FeeRate     float64        `gorm:"type:decimal(10,4);default:0" json:"fee_rate"`       // 费率（百分比）
	FixedFee    int64          `gorm:"type:bigint;default:0" json:"fixed_fee"`             // 固定手续费（分）
	MinAmount   int64          `gorm:"type:bigint;default:0" json:"min_amount"`            // 最小金额（分）
	MaxAmount   int64          `gorm:"type:bigint" json:"max_amount"`                      // 最大金额（分）
	Currencies  string         `gorm:"type:jsonb" json:"currencies"`                       // 支持的货币列表
	Countries   string         `gorm:"type:jsonb" json:"countries"`                        // 支持的国家列表
	Priority    int            `gorm:"default:0" json:"priority"`                          // 优先级
	Extra       string         `gorm:"type:jsonb" json:"extra"`                            // 扩展信息
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ChannelConfig) TableName() string {
	return "channel_configs"
}

// StripeConfig Stripe 配置结构
type StripeConfig struct {
	APIKey          string `json:"api_key"`           // Stripe API 密钥
	WebhookSecret   string `json:"webhook_secret"`    // Webhook 签名密钥
	PublishableKey  string `json:"publishable_key"`   // 可发布密钥（给前端使用）
	StatementDescriptor string `json:"statement_descriptor"` // 账单描述符
	SuccessURL      string `json:"success_url"`       // 支付成功跳转URL
	CancelURL       string `json:"cancel_url"`        // 支付取消跳转URL
	CaptureMethod   string `json:"capture_method"`    // 捕获方式：automatic, manual
}

// PayPalConfig PayPal 配置结构
type PayPalConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Mode         string `json:"mode"` // sandbox, live
	WebhookID    string `json:"webhook_id"`
}

// CryptoConfig 加密货币配置结构
type CryptoConfig struct {
	WalletAddress string   `json:"wallet_address"`    // 钱包地址
	Networks      []string `json:"networks"`          // 支持的网络：ETH, BSC, TRON
	Confirmations int      `json:"confirmations"`     // 确认数
	APIEndpoint   string   `json:"api_endpoint"`      // API 端点
	APIKey        string   `json:"api_key"`           // API 密钥
}

// 渠道类型常量
const (
	ChannelStripe  = "stripe"
	ChannelPayPal  = "paypal"
	ChannelCrypto  = "crypto"
	ChannelAlipay  = "alipay"
	ChannelWechat  = "wechat"
)

// 配置模式常量
const (
	ModeTest = "test"
	ModeLive = "live"
)

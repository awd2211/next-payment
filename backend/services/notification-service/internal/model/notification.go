package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification 通知记录表
type Notification struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID    uuid.UUID      `gorm:"type:uuid;index" json:"merchant_id"`           // 商户ID
	Type          string         `gorm:"type:varchar(50);not null;index" json:"type"`  // 通知类型
	Channel       string         `gorm:"type:varchar(50);not null" json:"channel"`     // 通知渠道
	Recipient     string         `gorm:"type:varchar(255);not null" json:"recipient"`  // 接收者
	Subject       string         `gorm:"type:varchar(500)" json:"subject"`             // 主题
	Content       string         `gorm:"type:text;not null" json:"content"`            // 内容
	TemplateCode  string         `gorm:"type:varchar(100)" json:"template_code"`       // 模板编码
	TemplateData  string         `gorm:"type:jsonb" json:"template_data"`              // 模板数据
	Status        string         `gorm:"type:varchar(20);not null;index" json:"status"` // 状态
	Priority      int            `gorm:"default:0" json:"priority"`                    // 优先级（0-9，9最高）
	RetryCount    int            `gorm:"default:0" json:"retry_count"`                 // 重试次数
	MaxRetry      int            `gorm:"default:3" json:"max_retry"`                   // 最大重试次数
	ErrorMessage  string         `gorm:"type:text" json:"error_message"`               // 错误信息
	Provider      string         `gorm:"type:varchar(50)" json:"provider"`             // 服务提供商
	ProviderMsgID string         `gorm:"type:varchar(200)" json:"provider_msg_id"`     // 提供商消息ID
	Extra         string         `gorm:"type:jsonb" json:"extra"`                      // 扩展信息
	ScheduledAt   *time.Time     `gorm:"type:timestamptz" json:"scheduled_at"`         // 计划发送时间
	SentAt        *time.Time     `gorm:"type:timestamptz" json:"sent_at"`              // 实际发送时间
	CreatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}

// NotificationTemplate 通知模板表
type NotificationTemplate struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID      `gorm:"type:uuid;index" json:"merchant_id"`          // 商户ID（NULL表示系统模板）
	Code        string         `gorm:"type:varchar(100);not null;index" json:"code"` // 模板编码
	Name        string         `gorm:"type:varchar(200);not null" json:"name"`      // 模板名称
	Type        string         `gorm:"type:varchar(50);not null" json:"type"`       // 通知类型
	Channel     string         `gorm:"type:varchar(50);not null" json:"channel"`    // 通知渠道
	Subject     string         `gorm:"type:varchar(500)" json:"subject"`            // 主题模板
	Content     string         `gorm:"type:text;not null" json:"content"`           // 内容模板
	Description string         `gorm:"type:text" json:"description"`                // 描述
	Variables   string         `gorm:"type:jsonb" json:"variables"`                 // 可用变量列表
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`              // 是否启用
	IsSystem    bool           `gorm:"default:false" json:"is_system"`              // 是否系统模板
	Extra       string         `gorm:"type:jsonb" json:"extra"`                     // 扩展信息
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// WebhookEndpoint Webhook 端点配置表
type WebhookEndpoint struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`   // 商户ID
	Name        string         `gorm:"type:varchar(200)" json:"name"`                 // 端点名称
	URL         string         `gorm:"type:varchar(500);not null" json:"url"`         // Webhook URL
	Secret      string         `gorm:"type:varchar(200)" json:"secret"`               // 签名密钥（加密存储）
	Events      string         `gorm:"type:jsonb" json:"events"`                      // 订阅的事件列表
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`                // 是否启用
	Version     string         `gorm:"type:varchar(20);default:'v1'" json:"version"`  // API版本
	Timeout     int            `gorm:"default:30" json:"timeout"`                     // 超时时间（秒）
	MaxRetry    int            `gorm:"default:3" json:"max_retry"`                    // 最大重试次数
	Description string         `gorm:"type:text" json:"description"`                  // 描述
	Extra       string         `gorm:"type:jsonb" json:"extra"`                       // 扩展信息
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (WebhookEndpoint) TableName() string {
	return "webhook_endpoints"
}

// WebhookDelivery Webhook 投递记录表
type WebhookDelivery struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EndpointID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"endpoint_id"`    // 端点ID
	MerchantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"merchant_id"`    // 商户ID
	EventType     string     `gorm:"type:varchar(100);not null" json:"event_type"`   // 事件类型
	EventID       string     `gorm:"type:varchar(200);index" json:"event_id"`        // 事件ID
	Payload       string     `gorm:"type:jsonb;not null" json:"payload"`             // 事件数据
	Status        string     `gorm:"type:varchar(20);not null;index" json:"status"`  // 状态
	HTTPStatus    int        `gorm:"default:0" json:"http_status"`                   // HTTP状态码
	ResponseBody  string     `gorm:"type:text" json:"response_body"`                 // 响应内容
	ErrorMessage  string     `gorm:"type:text" json:"error_message"`                 // 错误信息
	RetryCount    int        `gorm:"default:0" json:"retry_count"`                   // 重试次数
	Duration      int        `gorm:"default:0" json:"duration"`                      // 请求耗时（毫秒）
	NextRetryAt   *time.Time `gorm:"type:timestamptz" json:"next_retry_at"`          // 下次重试时间
	DeliveredAt   *time.Time `gorm:"type:timestamptz" json:"delivered_at"`           // 投递时间
	CreatedAt     time.Time  `gorm:"type:timestamptz;default:now();index" json:"created_at"`
}

// TableName 指定表名
func (WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}

// 通知类型常量
const (
	NotificationTypePayment     = "payment"      // 支付通知
	NotificationTypeRefund      = "refund"       // 退款通知
	NotificationTypeOrder       = "order"        // 订单通知
	NotificationTypeAccount     = "account"      // 账户通知
	NotificationTypeSystem      = "system"       // 系统通知
	NotificationTypeSecurity    = "security"     // 安全通知
	NotificationTypeMarketing   = "marketing"    // 营销通知
	NotificationTypeTransaction = "transaction"  // 交易通知
)

// 通知渠道常量
const (
	ChannelEmail   = "email"   // 邮件
	ChannelSMS     = "sms"     // 短信
	ChannelWebhook = "webhook" // Webhook
	ChannelPush    = "push"    // 推送通知
	ChannelInApp   = "in_app"  // 应用内通知
)

// 通知状态常量
const (
	StatusPending   = "pending"   // 待发送
	StatusSending   = "sending"   // 发送中
	StatusSent      = "sent"      // 已发送
	StatusFailed    = "failed"    // 发送失败
	StatusCancelled = "cancelled" // 已取消
)

// Webhook 投递状态常量
const (
	DeliveryStatusPending   = "pending"   // 待投递
	DeliveryStatusDelivered = "delivered" // 已投递
	DeliveryStatusFailed    = "failed"    // 投递失败
	DeliveryStatusRetrying  = "retrying"  // 重试中
)

// 邮件服务提供商常量
const (
	ProviderSMTP    = "smtp"    // SMTP
	ProviderMailgun = "mailgun" // Mailgun
	ProviderSendGrid = "sendgrid" // SendGrid
	ProviderAWSSES  = "aws_ses" // AWS SES
)

// 短信服务提供商常量
const (
	ProviderTwilio  = "twilio"  // Twilio
	ProviderAliyun  = "aliyun"  // 阿里云
	ProviderTencent = "tencent" // 腾讯云
)

// NotificationPreference 通知偏好设置表
type NotificationPreference struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;index" json:"user_id"`               // 用户ID（可选）
	MerchantID  uuid.UUID      `gorm:"type:uuid;index" json:"merchant_id"`           // 商户ID
	Channel     string         `gorm:"type:varchar(50);not null;index" json:"channel"` // 通知渠道
	EventType   string         `gorm:"type:varchar(100);not null;index" json:"event_type"` // 事件类型
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`               // 是否启用
	Description string         `gorm:"type:text" json:"description"`                 // 描述
	Extra       string         `gorm:"type:jsonb" json:"extra"`                      // 扩展信息
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

// 事件类型常量
const (
	EventTypeMerchantRegistered  = "merchant.registered"   // 商户注册
	EventTypeKYCApproved         = "kyc.approved"          // KYC审核通过
	EventTypeKYCRejected         = "kyc.rejected"          // KYC审核拒绝
	EventTypeMerchantFrozen      = "merchant.frozen"       // 商户冻结
	EventTypePasswordReset       = "password.reset"        // 密码重置
	EventTypePaymentSuccess      = "payment.success"       // 支付成功
	EventTypePaymentFailed       = "payment.failed"        // 支付失败
	EventTypeRefundCompleted     = "refund.completed"      // 退款完成
	EventTypeOrderCreated        = "order.created"         // 订单创建
	EventTypeOrderCancelled      = "order.cancelled"       // 订单取消
	EventTypeSettlementCompleted = "settlement.completed"  // 结算完成
	EventTypeSystemMaintenance   = "system.maintenance"    // 系统维护
)

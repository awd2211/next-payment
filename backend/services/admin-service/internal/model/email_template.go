package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailTemplate 邮件模板表
type EmailTemplate struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string         `gorm:"type:varchar(100);unique;not null;index" json:"code"` // 模板代码（唯一标识）
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`              // 模板名称
	Subject     string         `gorm:"type:varchar(255);not null" json:"subject"`           // 邮件主题（支持变量）
	HTMLContent string         `gorm:"type:text;not null" json:"html_content"`              // HTML内容
	TextContent string         `gorm:"type:text" json:"text_content"`                       // 纯文本内容（可选）
	Description string         `gorm:"type:text" json:"description"`                        // 模板描述
	Category    string         `gorm:"type:varchar(50);not null;index" json:"category"`     // 分类：account, payment, notification
	Variables   string         `gorm:"type:jsonb" json:"variables"`                         // 可用变量列表（JSON数组）
	IsActive    bool           `gorm:"default:true" json:"is_active"`                       // 是否启用
	IsSystem    bool           `gorm:"default:false" json:"is_system"`                      // 是否系统内置（不可删除）
	UpdatedBy   uuid.UUID      `gorm:"type:uuid" json:"updated_by"`                         // 最后更新人
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (EmailTemplate) TableName() string {
	return "email_templates"
}

// EmailTemplateVariable 模板变量定义
type EmailTemplateVariable struct {
	Name        string `json:"name"`        // 变量名，如：Name, Email
	Placeholder string `json:"placeholder"` // 占位符，如：{{.Name}}, {{.Email}}
	Description string `json:"description"` // 变量说明
	Example     string `json:"example"`     // 示例值
	Required    bool   `json:"required"`    // 是否必需
}

// EmailLog 邮件发送日志
type EmailLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TemplateID uuid.UUID `gorm:"type:uuid;index" json:"template_id"`             // 使用的模板ID
	To         string    `gorm:"type:varchar(255);not null;index" json:"to"`     // 收件人
	Subject    string    `gorm:"type:varchar(255);not null" json:"subject"`      // 实际主题
	Status     string    `gorm:"type:varchar(20);not null;index" json:"status"`  // pending, sent, failed
	Provider   string    `gorm:"type:varchar(50)" json:"provider"`               // smtp, mailgun
	ErrorMsg   string    `gorm:"type:text" json:"error_msg"`                     // 错误信息
	SentAt     *time.Time `gorm:"type:timestamptz" json:"sent_at"`               // 发送时间
	CreatedAt  time.Time `gorm:"type:timestamptz;default:now();index" json:"created_at"`
}

// TableName 指定表名
func (EmailLog) TableName() string {
	return "email_logs"
}

// 预定义模板代码常量
const (
	TemplateWelcome          = "welcome"           // 欢迎邮件
	TemplateVerifyEmail      = "verify_email"      // 邮箱验证
	TemplateResetPassword    = "reset_password"    // 重置密码
	TemplatePaymentSuccess   = "payment_success"   // 支付成功
	TemplatePaymentFailed    = "payment_failed"    // 支付失败
	TemplateRefundCompleted  = "refund_completed"  // 退款完成
	TemplateMerchantApproved = "merchant_approved" // 商户审核通过
	TemplateMerchantRejected = "merchant_rejected" // 商户审核拒绝
	TemplateInvoice          = "invoice"           // 账单
)

// 预定义模板分类
const (
	CategoryAccount      = "account"      // 账号相关
	CategoryPayment      = "payment"      // 支付相关
	CategoryNotification = "notification" // 通知相关
	CategorySecurity     = "security"     // 安全相关
)

package model

import (
	"time"

	"github.com/google/uuid"
)

// WebhookNotification Webhook 通知记录
type WebhookNotification struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID `gorm:"type:uuid;not null;index:idx_webhook_merchant" json:"merchant_id"`
	PaymentNo   string    `gorm:"type:varchar(64);not null;index:idx_webhook_payment" json:"payment_no"`
	OrderNo     string    `gorm:"type:varchar(128);not null;index:idx_webhook_order" json:"order_no"`
	Event       string    `gorm:"type:varchar(50);not null" json:"event"` // payment.success, payment.failed, refund.success
	URL         string    `gorm:"type:varchar(500);not null" json:"url"`
	Payload     string    `gorm:"type:jsonb" json:"payload"`               // JSON 格式的通知内容
	Status      string    `gorm:"type:varchar(20);not null;index" json:"status"` // pending, success, failed, retrying
	Attempts    int       `gorm:"default:0" json:"attempts"`               // 尝试次数
	MaxAttempts int       `gorm:"default:5" json:"max_attempts"`           // 最大尝试次数
	StatusCode  int       `gorm:"default:0" json:"status_code"`            // HTTP 状态码
	Response    string    `gorm:"type:text" json:"response"`               // 响应内容
	Error       string    `gorm:"type:text" json:"error"`                  // 错误信息
	NextRetryAt *time.Time `gorm:"index:idx_webhook_retry" json:"next_retry_at"` // 下次重试时间
	SucceededAt *time.Time `json:"succeeded_at"`                           // 成功时间
	FailedAt    *time.Time `json:"failed_at"`                              // 最终失败时间
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// Webhook 通知状态
const (
	WebhookStatusPending  = "pending"  // 待发送
	WebhookStatusSuccess  = "success"  // 成功
	WebhookStatusFailed   = "failed"   // 失败（已达最大重试次数）
	WebhookStatusRetrying = "retrying" // 重试中
)

// Webhook 事件类型
const (
	WebhookEventPaymentSuccess = "payment.success" // 支付成功
	WebhookEventPaymentFailed  = "payment.failed"  // 支付失败
	WebhookEventPaymentExpired = "payment.expired" // 支付过期
	WebhookEventRefundSuccess  = "refund.success"  // 退款成功
	WebhookEventRefundFailed   = "refund.failed"   // 退款失败
)

// TableName 表名
func (WebhookNotification) TableName() string {
	return "webhook_notifications"
}

// CanRetry 是否可以重试
func (w *WebhookNotification) CanRetry() bool {
	return w.Status != WebhookStatusSuccess && w.Attempts < w.MaxAttempts
}

// ShouldRetryNow 是否应该立即重试
func (w *WebhookNotification) ShouldRetryNow() bool {
	if !w.CanRetry() {
		return false
	}

	// 如果没有设置下次重试时间，立即重试
	if w.NextRetryAt == nil {
		return true
	}

	// 如果已经到了重试时间
	return time.Now().After(*w.NextRetryAt)
}

package model

import (
	"time"

	"github.com/google/uuid"
)

// PreAuthPayment 预授权支付记录
type PreAuthPayment struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"merchant_id"`
	OrderNo        string     `gorm:"type:varchar(100);not null;unique;index" json:"order_no"`             // 订单号
	PreAuthNo      string     `gorm:"type:varchar(100);not null;unique;index" json:"pre_auth_no"`          // 预授权单号
	PaymentNo      *string    `gorm:"type:varchar(100);index" json:"payment_no"`                           // 确认后的支付单号
	Amount         int64      `gorm:"type:bigint;not null" json:"amount"`                                  // 预授权金额（分）
	CapturedAmount int64      `gorm:"type:bigint;default:0" json:"captured_amount"`                        // 已确认金额（分）
	Currency       string     `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`             // 币种
	Channel        string     `gorm:"type:varchar(50);not null" json:"channel"`                            // 支付渠道
	ChannelTradeNo string     `gorm:"type:varchar(200);index" json:"channel_trade_no"`                     // 渠道交易号
	Status         string     `gorm:"type:varchar(20);not null;index" json:"status"`                       // pending, authorized, captured, cancelled, expired
	ExpiresAt      time.Time  `gorm:"type:timestamptz;not null;index" json:"expires_at"`                   // 过期时间
	AuthorizedAt   *time.Time `gorm:"type:timestamptz" json:"authorized_at"`                               // 授权时间
	CapturedAt     *time.Time `gorm:"type:timestamptz" json:"captured_at"`                                 // 确认时间
	CancelledAt    *time.Time `gorm:"type:timestamptz" json:"cancelled_at"`                                // 取消时间
	Subject        string     `gorm:"type:varchar(255)" json:"subject"`                                    // 商品标题
	Body           string     `gorm:"type:text" json:"body"`                                               // 商品描述
	Extra          string     `gorm:"type:jsonb" json:"extra"`                                             // 扩展信息
	ClientIP       string     `gorm:"type:varchar(50)" json:"client_ip"`                                   // 客户端IP
	ReturnURL      string     `gorm:"type:varchar(500)" json:"return_url"`                                 // 返回URL
	NotifyURL      string     `gorm:"type:varchar(500)" json:"notify_url"`                                 // 通知URL
	ErrorCode      string     `gorm:"type:varchar(50)" json:"error_code"`                                  // 错误码
	ErrorMessage   string     `gorm:"type:varchar(500)" json:"error_message"`                              // 错误信息
	CreatedAt      time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (PreAuthPayment) TableName() string {
	return "pre_auth_payments"
}

// PreAuthPayment 状态常量
const (
	PreAuthStatusPending    = "pending"    // 待授权
	PreAuthStatusAuthorized = "authorized" // 已授权（待确认）
	PreAuthStatusCaptured   = "captured"   // 已确认（已扣款）
	PreAuthStatusCancelled  = "cancelled"  // 已取消
	PreAuthStatusExpired    = "expired"    // 已过期
)

// IsExpired 检查是否已过期
func (p *PreAuthPayment) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// CanCapture 检查是否可以确认
func (p *PreAuthPayment) CanCapture() bool {
	return p.Status == PreAuthStatusAuthorized && !p.IsExpired()
}

// CanCancel 检查是否可以取消
func (p *PreAuthPayment) CanCancel() bool {
	return (p.Status == PreAuthStatusPending || p.Status == PreAuthStatusAuthorized) && !p.IsExpired()
}

// GetRemainingAmount 获取剩余可确认金额
func (p *PreAuthPayment) GetRemainingAmount() int64 {
	return p.Amount - p.CapturedAmount
}

// PreAuthCaptureRequest 确认预授权请求
type PreAuthCaptureRequest struct {
	PreAuthNo string `json:"pre_auth_no" binding:"required"`
	Amount    *int64 `json:"amount"` // 可选，如果不传则全额确认
}

// PreAuthCancelRequest 取消预授权请求
type PreAuthCancelRequest struct {
	PreAuthNo string `json:"pre_auth_no" binding:"required"`
	Reason    string `json:"reason"` // 可选，取消原因
}

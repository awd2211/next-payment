package adapter

import (
	"context"
)

// PaymentAdapter 支付适配器接口
// 所有支付渠道适配器都需要实现这个接口
type PaymentAdapter interface {
	// GetChannel 获取渠道名称
	GetChannel() string

	// CreatePayment 创建支付
	// 返回：渠道交易号、客户端密钥（用于前端）、错误
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)

	// QueryPayment 查询支付状态
	QueryPayment(ctx context.Context, channelTradeNo string) (*QueryPaymentResponse, error)

	// CancelPayment 取消支付
	CancelPayment(ctx context.Context, channelTradeNo string) error

	// CreateRefund 创建退款
	// 返回：退款交易号、错误
	CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error)

	// QueryRefund 查询退款状态
	QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error)

	// VerifyWebhook 验证 Webhook 签名
	VerifyWebhook(ctx context.Context, signature string, body []byte) (bool, error)

	// ParseWebhook 解析 Webhook 数据
	ParseWebhook(ctx context.Context, body []byte) (*WebhookEvent, error)
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	PaymentNo     string                 `json:"payment_no"`      // 平台支付流水号
	OrderNo       string                 `json:"order_no"`        // 订单号
	Amount        int64                  `json:"amount"`          // 金额（分）
	Currency      string                 `json:"currency"`        // 货币
	CustomerEmail string                 `json:"customer_email"`  // 客户邮箱
	CustomerName  string                 `json:"customer_name"`   // 客户姓名
	Description   string                 `json:"description"`     // 描述
	SuccessURL    string                 `json:"success_url"`     // 成功跳转URL
	CancelURL     string                 `json:"cancel_url"`      // 取消跳转URL
	CallbackURL   string                 `json:"callback_url"`    // 回调URL
	Extra         map[string]interface{} `json:"extra"`           // 扩展信息
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	ChannelTradeNo string                 `json:"channel_trade_no"` // 渠道交易号
	ClientSecret   string                 `json:"client_secret"`    // 客户端密钥（给前端使用）
	PaymentURL     string                 `json:"payment_url"`      // 支付URL（重定向方式）
	QRCodeURL      string                 `json:"qr_code_url"`      // 二维码URL
	Status         string                 `json:"status"`           // 状态
	Extra          map[string]interface{} `json:"extra"`            // 扩展信息
}

// QueryPaymentResponse 查询支付响应
type QueryPaymentResponse struct {
	ChannelTradeNo       string                 `json:"channel_trade_no"`        // 渠道交易号
	Status               string                 `json:"status"`                  // 状态
	Amount               int64                  `json:"amount"`                  // 金额（分）
	Currency             string                 `json:"currency"`                // 货币
	PaymentMethod        string                 `json:"payment_method"`          // 支付方式
	PaymentMethodDetails map[string]interface{} `json:"payment_method_details"`  // 支付方式详情
	PaidAt               *int64                 `json:"paid_at"`                 // 支付时间（Unix时间戳）
	Extra                map[string]interface{} `json:"extra"`                   // 扩展信息
}

// CreateRefundRequest 创建退款请求
type CreateRefundRequest struct {
	RefundNo       string                 `json:"refund_no"`        // 退款流水号
	PaymentNo      string                 `json:"payment_no"`       // 原支付流水号
	ChannelTradeNo string                 `json:"channel_trade_no"` // 原渠道交易号
	Amount         int64                  `json:"amount"`           // 退款金额（分）
	Currency       string                 `json:"currency"`         // 货币
	Reason         string                 `json:"reason"`           // 退款原因
	Extra          map[string]interface{} `json:"extra"`            // 扩展信息
}

// CreateRefundResponse 创建退款响应
type CreateRefundResponse struct {
	RefundNo       string                 `json:"refund_no"`        // 退款流水号
	ChannelRefundNo string                `json:"channel_refund_no"` // 渠道退款号
	Status         string                 `json:"status"`           // 状态
	Extra          map[string]interface{} `json:"extra"`            // 扩展信息
}

// QueryRefundResponse 查询退款响应
type QueryRefundResponse struct {
	RefundNo        string                 `json:"refund_no"`         // 退款流水号
	ChannelRefundNo string                 `json:"channel_refund_no"` // 渠道退款号
	Status          string                 `json:"status"`            // 状态
	Amount          int64                  `json:"amount"`            // 退款金额（分）
	Currency        string                 `json:"currency"`          // 货币
	RefundedAt      *int64                 `json:"refunded_at"`       // 退款时间（Unix时间戳）
	Extra           map[string]interface{} `json:"extra"`             // 扩展信息
}

// WebhookEvent Webhook 事件
type WebhookEvent struct {
	EventID        string                 `json:"event_id"`         // 事件ID
	EventType      string                 `json:"event_type"`       // 事件类型
	ChannelTradeNo string                 `json:"channel_trade_no"` // 渠道交易号
	PaymentNo      string                 `json:"payment_no"`       // 支付流水号
	Status         string                 `json:"status"`           // 状态
	Amount         int64                  `json:"amount"`           // 金额（分）
	Currency       string                 `json:"currency"`         // 货币
	Extra          map[string]interface{} `json:"extra"`            // 扩展信息
	RawData        interface{}            `json:"raw_data"`         // 原始数据
}

// AdapterFactory 适配器工厂
type AdapterFactory struct {
	adapters map[string]PaymentAdapter
}

// NewAdapterFactory 创建适配器工厂
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{
		adapters: make(map[string]PaymentAdapter),
	}
}

// Register 注册适配器
func (f *AdapterFactory) Register(channel string, adapter PaymentAdapter) {
	f.adapters[channel] = adapter
}

// GetAdapter 获取适配器
func (f *AdapterFactory) GetAdapter(channel string) (PaymentAdapter, bool) {
	adapter, ok := f.adapters[channel]
	return adapter, ok
}

// 支付状态常量（统一状态）
const (
	PaymentStatusPending    = "pending"    // 待支付
	PaymentStatusProcessing = "processing" // 处理中
	PaymentStatusSuccess    = "success"    // 成功
	PaymentStatusFailed     = "failed"     // 失败
	PaymentStatusCancelled  = "cancelled"  // 已取消
	PaymentStatusRefunded   = "refunded"   // 已退款
)

// Webhook 事件类型常量
const (
	EventTypePaymentSuccess   = "payment.success"    // 支付成功
	EventTypePaymentFailed    = "payment.failed"     // 支付失败
	EventTypePaymentCancelled = "payment.cancelled"  // 支付取消
	EventTypeRefundSuccess    = "refund.success"     // 退款成功
	EventTypeRefundFailed     = "refund.failed"      // 退款失败
)

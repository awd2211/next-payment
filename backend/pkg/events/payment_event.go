package events

import (
	"encoding/json"
	"time"
)

// PaymentEvent 支付事件
type PaymentEvent struct {
	BaseEvent
	Payload PaymentEventPayload `json:"payload"`
}

// PaymentEventPayload 支付事件载荷
type PaymentEventPayload struct {
	PaymentNo     string                 `json:"payment_no"`     // 支付流水号
	MerchantID    string                 `json:"merchant_id"`    // 商户ID
	OrderNo       string                 `json:"order_no"`       // 订单号
	Amount        int64                  `json:"amount"`         // 金额(分)
	Currency      string                 `json:"currency"`       // 货币
	Channel       string                 `json:"channel"`        // 支付渠道
	Status        string                 `json:"status"`         // 状态
	CustomerEmail string                 `json:"customer_email"` // 客户邮箱
	PaidAt        *time.Time             `json:"paid_at"`        // 支付时间
	Extra         map[string]interface{} `json:"extra"`          // 扩展信息
}

// Event Type Constants
const (
	PaymentCreated   = "payment.created"
	PaymentSuccess   = "payment.success"
	PaymentFailed    = "payment.failed"
	PaymentCancelled = "payment.cancelled"
	PaymentExpired   = "payment.expired"
)

// NewPaymentEvent 创建支付事件
func NewPaymentEvent(eventType string, payload PaymentEventPayload) *PaymentEvent {
	event := &PaymentEvent{
		BaseEvent: *NewBaseEvent(eventType, "payment", payload.PaymentNo),
		Payload:   payload,
	}
	return event
}

// GetEventType 实现Event接口
func (e *PaymentEvent) GetEventType() string {
	return e.EventType
}

// GetAggregateType 实现Event接口
func (e *PaymentEvent) GetAggregateType() string {
	return e.AggregateType
}

// GetAggregateID 实现Event接口
func (e *PaymentEvent) GetAggregateID() string {
	return e.AggregateID
}

// ToJSON 序列化为JSON
func (e *PaymentEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// RefundEvent 退款事件
type RefundEvent struct {
	BaseEvent
	Payload RefundEventPayload `json:"payload"`
}

// RefundEventPayload 退款事件载荷
type RefundEventPayload struct {
	RefundNo   string                 `json:"refund_no"`   // 退款单号
	PaymentNo  string                 `json:"payment_no"`  // 支付流水号
	MerchantID string                 `json:"merchant_id"` // 商户ID
	OrderNo    string                 `json:"order_no"`    // 订单号
	Amount     int64                  `json:"amount"`      // 退款金额
	Currency   string                 `json:"currency"`    // 货币
	Reason     string                 `json:"reason"`      // 退款原因
	Status     string                 `json:"status"`      // 状态
	RefundedAt *time.Time             `json:"refunded_at"` // 退款时间
	Extra      map[string]interface{} `json:"extra"`       // 扩展信息
}

// Refund Event Type Constants
const (
	RefundCreated = "refund.created"
	RefundSuccess = "refund.success"
	RefundFailed  = "refund.failed"
)

// NewRefundEvent 创建退款事件
func NewRefundEvent(eventType string, payload RefundEventPayload) *RefundEvent {
	event := &RefundEvent{
		BaseEvent: *NewBaseEvent(eventType, "refund", payload.RefundNo),
		Payload:   payload,
	}
	return event
}

// GetEventType 实现Event接口
func (e *RefundEvent) GetEventType() string {
	return e.EventType
}

// GetAggregateType 实现Event接口
func (e *RefundEvent) GetAggregateType() string {
	return e.AggregateType
}

// GetAggregateID 实现Event接口
func (e *RefundEvent) GetAggregateID() string {
	return e.AggregateID
}

// ToJSON 序列化为JSON
func (e *RefundEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

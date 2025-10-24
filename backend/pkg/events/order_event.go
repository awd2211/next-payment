package events

import (
	"encoding/json"
	"time"
)

// OrderEvent 订单事件
type OrderEvent struct {
	BaseEvent
	Payload OrderEventPayload `json:"payload"`
}

// OrderEventPayload 订单事件载荷
type OrderEventPayload struct {
	OrderNo       string                 `json:"order_no"`       // 订单号
	MerchantID    string                 `json:"merchant_id"`    // 商户ID
	PaymentNo     string                 `json:"payment_no"`     // 支付流水号
	TotalAmount   int64                  `json:"total_amount"`   // 订单总金额
	Currency      string                 `json:"currency"`       // 货币
	Status        string                 `json:"status"`         // 订单状态
	CustomerEmail string                 `json:"customer_email"` // 客户邮箱
	PaidAt        *time.Time             `json:"paid_at"`        // 支付时间
	Extra         map[string]interface{} `json:"extra"`          // 扩展信息
}

// Order Event Type Constants
const (
	OrderCreated   = "order.created"
	OrderPaid      = "order.paid"
	OrderCancelled = "order.cancelled"
	OrderRefunded  = "order.refunded"
	OrderShipped   = "order.shipped"
	OrderCompleted = "order.completed"
)

// NewOrderEvent 创建订单事件
func NewOrderEvent(eventType string, payload OrderEventPayload) *OrderEvent {
	event := &OrderEvent{
		BaseEvent: *NewBaseEvent(eventType, "order", payload.OrderNo),
		Payload:   payload,
	}
	return event
}

// GetEventType 实现Event接口
func (e *OrderEvent) GetEventType() string {
	return e.EventType
}

// GetAggregateType 实现Event接口
func (e *OrderEvent) GetAggregateType() string {
	return e.AggregateType
}

// GetAggregateID 实现Event接口
func (e *OrderEvent) GetAggregateID() string {
	return e.AggregateID
}

// ToJSON 序列化为JSON
func (e *OrderEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

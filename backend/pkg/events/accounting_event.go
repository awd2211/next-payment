package events

import (
	"encoding/json"
	"time"
)

// AccountingEvent 财务事件
type AccountingEvent struct {
	BaseEvent
	Payload AccountingEventPayload `json:"payload"`
}

// AccountingEventPayload 财务事件载荷
type AccountingEventPayload struct {
	TransactionID string                 `json:"transaction_id"` // 交易ID
	AccountID     string                 `json:"account_id"`     // 账户ID
	MerchantID    string                 `json:"merchant_id"`    // 商户ID
	Type          string                 `json:"type"`           // 类型: credit(贷), debit(借)
	Amount        int64                  `json:"amount"`         // 金额
	Balance       int64                  `json:"balance"`        // 余额
	Currency      string                 `json:"currency"`       // 货币
	Description   string                 `json:"description"`    // 描述
	RelatedID     string                 `json:"related_id"`     // 关联ID (payment_no/refund_no)
	CreatedAt     time.Time              `json:"created_at"`     // 创建时间
	Extra         map[string]interface{} `json:"extra"`          // 扩展信息
}

// Accounting Event Type Constants
const (
	TransactionCreated   = "accounting.transaction.created"
	BalanceUpdated       = "accounting.balance.updated"
	SettlementCalculated = "accounting.settlement.calculated"
)

// NewAccountingEvent 创建财务事件
func NewAccountingEvent(eventType string, payload AccountingEventPayload) *AccountingEvent {
	event := &AccountingEvent{
		BaseEvent: *NewBaseEvent(eventType, "accounting", payload.TransactionID),
		Payload:   payload,
	}
	return event
}

// GetEventType 实现Event接口
func (e *AccountingEvent) GetEventType() string {
	return e.EventType
}

// GetAggregateType 实现Event接口
func (e *AccountingEvent) GetAggregateType() string {
	return e.AggregateType
}

// GetAggregateID 实现Event接口
func (e *AccountingEvent) GetAggregateID() string {
	return e.AggregateID
}

// ToJSON 序列化为JSON
func (e *AccountingEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

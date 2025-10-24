package events

import "time"

// SettlementEventPayload 结算事件载荷
type SettlementEventPayload struct {
	SettlementNo     string                 `json:"settlement_no"`
	MerchantID       string                 `json:"merchant_id"`
	Cycle            string                 `json:"cycle"` // daily, weekly, monthly
	TotalAmount      int64                  `json:"total_amount"`
	FeeAmount        int64                  `json:"fee_amount"`
	SettlementAmount int64                  `json:"settlement_amount"`
	TotalCount       int                    `json:"total_count"`
	Currency         string                 `json:"currency"`
	Status           string                 `json:"status"` // pending, approved, rejected, completed, failed
	StartDate        time.Time              `json:"start_date"`
	EndDate          time.Time              `json:"end_date"`
	ApprovedBy       string                 `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time             `json:"approved_at,omitempty"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	Extra            map[string]interface{} `json:"extra,omitempty"`
}

// Settlement Event Type Constants
const (
	SettlementCreated   = "settlement.created"
	SettlementApproved  = "settlement.approved"
	SettlementRejected  = "settlement.rejected"
	SettlementCompleted = "settlement.completed"
	SettlementFailed    = "settlement.failed"
)

// NewSettlementEvent 创建结算事件
func NewSettlementEvent(eventType string, payload SettlementEventPayload) *SettlementEvent {
	return &SettlementEvent{
		BaseEvent: *NewBaseEvent(eventType, "settlement", payload.SettlementNo),
		Payload:   payload,
	}
}

// SettlementEvent 结算事件
type SettlementEvent struct {
	BaseEvent
	Payload SettlementEventPayload `json:"payload"`
}

// 实现 Event 接口
func (e *SettlementEvent) GetEventID() string       { return e.EventID }
func (e *SettlementEvent) GetEventType() string     { return e.EventType }
func (e *SettlementEvent) GetAggregateID() string   { return e.AggregateID }
func (e *SettlementEvent) GetAggregateType() string { return e.AggregateType }
func (e *SettlementEvent) GetTimestamp() time.Time  { return e.Timestamp }
func (e *SettlementEvent) GetVersion() string       { return e.Version }
func (e *SettlementEvent) GetMetadata() map[string]interface{} {
	return e.Metadata
}

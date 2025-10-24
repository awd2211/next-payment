package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BaseEvent 基础事件结构 (所有业务事件的父类)
type BaseEvent struct {
	EventID       string                 `json:"event_id"`        // 事件唯一ID
	EventType     string                 `json:"event_type"`      // 事件类型: payment.success, order.paid
	AggregateType string                 `json:"aggregate_type"`  // 聚合类型: payment, order, merchant
	AggregateID   string                 `json:"aggregate_id"`    // 聚合ID (业务主键)
	Timestamp     time.Time              `json:"timestamp"`       // 事件时间
	Version       string                 `json:"version"`         // 事件Schema版本
	Metadata      map[string]interface{} `json:"metadata"`        // 元数据 (trace_id, user_id等)
}

// NewBaseEvent 创建基础事件
func NewBaseEvent(eventType, aggregateType, aggregateID string) *BaseEvent {
	return &BaseEvent{
		EventID:       uuid.New().String(),
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Timestamp:     time.Now(),
		Version:       "1.0",
		Metadata:      make(map[string]interface{}),
	}
}

// AddMetadata 添加元数据
func (e *BaseEvent) AddMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// ToJSON 序列化为JSON
func (e *BaseEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Event 事件接口
type Event interface {
	GetEventType() string
	GetAggregateType() string
	GetAggregateID() string
	ToJSON() ([]byte, error)
}

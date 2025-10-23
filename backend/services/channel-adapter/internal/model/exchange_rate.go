package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExchangeRate 汇率历史记录
type ExchangeRate struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BaseCurrency string         `gorm:"type:varchar(10);not null;index:idx_exchange_rate_lookup" json:"base_currency"`
	TargetCurrency string       `gorm:"type:varchar(10);not null;index:idx_exchange_rate_lookup" json:"target_currency"`
	Rate         float64        `gorm:"type:decimal(20,8);not null" json:"rate"`
	Source       string         `gorm:"type:varchar(50);default:'exchangerate-api'" json:"source"`
	ValidFrom    time.Time      `gorm:"type:timestamp;not null;index:idx_exchange_rate_time" json:"valid_from"`
	ValidTo      *time.Time     `gorm:"type:timestamp" json:"valid_to,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ExchangeRate) TableName() string {
	return "exchange_rates"
}

// ExchangeRateSnapshot 汇率快照（批量存储）
// 用于存储某个时间点的多个汇率数据
type ExchangeRateSnapshot struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BaseCurrency string         `gorm:"type:varchar(10);not null;index:idx_snapshot_base" json:"base_currency"`
	Rates        map[string]float64 `gorm:"type:jsonb;serializer:json" json:"rates"`
	Source       string         `gorm:"type:varchar(50);default:'exchangerate-api'" json:"source"`
	SnapshotTime time.Time      `gorm:"type:timestamp;not null;index:idx_snapshot_time" json:"snapshot_time"`
	CreatedAt    time.Time      `json:"created_at"`
}

func (ExchangeRateSnapshot) TableName() string {
	return "exchange_rate_snapshots"
}

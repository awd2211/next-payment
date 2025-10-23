package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentMetrics 支付指标
type PaymentMetrics struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID         uuid.UUID      `gorm:"type:uuid;not null;index:idx_payment_metrics_merchant" json:"merchant_id"`
	Date               time.Time      `gorm:"type:date;not null;index:idx_payment_metrics_date" json:"date"`
	TotalPayments      int            `gorm:"default:0" json:"total_payments"`
	SuccessPayments    int            `gorm:"default:0" json:"success_payments"`
	FailedPayments     int            `gorm:"default:0" json:"failed_payments"`
	TotalAmount        int64          `gorm:"default:0" json:"total_amount"`
	SuccessAmount      int64          `gorm:"default:0" json:"success_amount"`
	TotalRefunds       int            `gorm:"default:0" json:"total_refunds"`
	TotalRefundAmount  int64          `gorm:"default:0" json:"total_refund_amount"`
	Currency           string         `gorm:"type:varchar(10)" json:"currency"`
	SuccessRate        float64        `gorm:"type:decimal(5,2)" json:"success_rate"`
	AverageAmount      int64          `gorm:"default:0" json:"average_amount"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PaymentMetrics) TableName() string {
	return "payment_metrics"
}

// MerchantMetrics 商户指标
type MerchantMetrics struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID        uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_merchant_metrics_unique" json:"merchant_id"`
	Date              time.Time      `gorm:"type:date;not null;uniqueIndex:idx_merchant_metrics_unique" json:"date"`
	TotalOrders       int            `gorm:"default:0" json:"total_orders"`
	CompletedOrders   int            `gorm:"default:0" json:"completed_orders"`
	CancelledOrders   int            `gorm:"default:0" json:"cancelled_orders"`
	TotalRevenue      int64          `gorm:"default:0" json:"total_revenue"`
	TotalFees         int64          `gorm:"default:0" json:"total_fees"`
	NetRevenue        int64          `gorm:"default:0" json:"net_revenue"`
	NewCustomers      int            `gorm:"default:0" json:"new_customers"`
	ReturningCustomers int           `gorm:"default:0" json:"returning_customers"`
	Currency          string         `gorm:"type:varchar(10)" json:"currency"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MerchantMetrics) TableName() string {
	return "merchant_metrics"
}

// ChannelMetrics 渠道指标
type ChannelMetrics struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChannelCode     string         `gorm:"type:varchar(50);not null;index:idx_channel_metrics_channel" json:"channel_code"`
	Date            time.Time      `gorm:"type:date;not null;index:idx_channel_metrics_date" json:"date"`
	TotalTransactions int          `gorm:"default:0" json:"total_transactions"`
	SuccessTransactions int        `gorm:"default:0" json:"success_transactions"`
	FailedTransactions int         `gorm:"default:0" json:"failed_transactions"`
	TotalAmount     int64          `gorm:"default:0" json:"total_amount"`
	SuccessAmount   int64          `gorm:"default:0" json:"success_amount"`
	AverageLatency  int            `gorm:"default:0" json:"average_latency"`
	SuccessRate     float64        `gorm:"type:decimal(5,2)" json:"success_rate"`
	Currency        string         `gorm:"type:varchar(10)" json:"currency"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelMetrics) TableName() string {
	return "channel_metrics"
}

// RealtimeStats 实时统计
type RealtimeStats struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID        *uuid.UUID     `gorm:"type:uuid;index:idx_realtime_stats_merchant" json:"merchant_id,omitempty"`
	StatType          string         `gorm:"type:varchar(50);not null;index:idx_realtime_stats_type" json:"stat_type"`
	StatKey           string         `gorm:"type:varchar(255);not null;index:idx_realtime_stats_key" json:"stat_key"`
	StatValue         int64          `gorm:"default:0" json:"stat_value"`
	AdditionalData    map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"additional_data,omitempty"`
	Period            string         `gorm:"type:varchar(20)" json:"period"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RealtimeStats) TableName() string {
	return "realtime_stats"
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/analytics-service/internal/model"
	"gorm.io/gorm"
)

// AnalyticsRepository 分析仓储接口
type AnalyticsRepository interface {
	// 支付指标
	CreatePaymentMetrics(ctx context.Context, metrics *model.PaymentMetrics) error
	GetPaymentMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.PaymentMetrics, error)
	UpdatePaymentMetrics(ctx context.Context, metrics *model.PaymentMetrics) error

	// 商户指标
	CreateMerchantMetrics(ctx context.Context, metrics *model.MerchantMetrics) error
	GetMerchantMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.MerchantMetrics, error)
	UpdateMerchantMetrics(ctx context.Context, metrics *model.MerchantMetrics) error

	// 渠道指标
	CreateChannelMetrics(ctx context.Context, metrics *model.ChannelMetrics) error
	GetChannelMetrics(ctx context.Context, channelCode string, startDate, endDate time.Time) ([]*model.ChannelMetrics, error)
	UpdateChannelMetrics(ctx context.Context, metrics *model.ChannelMetrics) error

	// 实时统计
	CreateRealtimeStats(ctx context.Context, stats *model.RealtimeStats) error
	GetRealtimeStats(ctx context.Context, query *RealtimeStatsQuery) ([]*model.RealtimeStats, error)
	UpdateRealtimeStats(ctx context.Context, stats *model.RealtimeStats) error
	IncrementRealtimeStats(ctx context.Context, merchantID *uuid.UUID, statType, statKey string, increment int64) error
}

type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository 创建分析仓储实例
func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// RealtimeStatsQuery 实时统计查询条件
type RealtimeStatsQuery struct {
	MerchantID *uuid.UUID
	StatType   string
	StatKey    string
	Period     string
}

// Payment Metrics

func (r *analyticsRepository) CreatePaymentMetrics(ctx context.Context, metrics *model.PaymentMetrics) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

func (r *analyticsRepository) GetPaymentMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.PaymentMetrics, error) {
	var metrics []*model.PaymentMetrics
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND date >= ? AND date <= ?", merchantID, startDate, endDate).
		Order("date DESC").
		Find(&metrics).Error
	return metrics, err
}

func (r *analyticsRepository) UpdatePaymentMetrics(ctx context.Context, metrics *model.PaymentMetrics) error {
	return r.db.WithContext(ctx).Save(metrics).Error
}

// Merchant Metrics

func (r *analyticsRepository) CreateMerchantMetrics(ctx context.Context, metrics *model.MerchantMetrics) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

func (r *analyticsRepository) GetMerchantMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.MerchantMetrics, error) {
	var metrics []*model.MerchantMetrics
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND date >= ? AND date <= ?", merchantID, startDate, endDate).
		Order("date DESC").
		Find(&metrics).Error
	return metrics, err
}

func (r *analyticsRepository) UpdateMerchantMetrics(ctx context.Context, metrics *model.MerchantMetrics) error {
	return r.db.WithContext(ctx).Save(metrics).Error
}

// Channel Metrics

func (r *analyticsRepository) CreateChannelMetrics(ctx context.Context, metrics *model.ChannelMetrics) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

func (r *analyticsRepository) GetChannelMetrics(ctx context.Context, channelCode string, startDate, endDate time.Time) ([]*model.ChannelMetrics, error) {
	var metrics []*model.ChannelMetrics
	err := r.db.WithContext(ctx).
		Where("channel_code = ? AND date >= ? AND date <= ?", channelCode, startDate, endDate).
		Order("date DESC").
		Find(&metrics).Error
	return metrics, err
}

func (r *analyticsRepository) UpdateChannelMetrics(ctx context.Context, metrics *model.ChannelMetrics) error {
	return r.db.WithContext(ctx).Save(metrics).Error
}

// Realtime Stats

func (r *analyticsRepository) CreateRealtimeStats(ctx context.Context, stats *model.RealtimeStats) error {
	return r.db.WithContext(ctx).Create(stats).Error
}

func (r *analyticsRepository) GetRealtimeStats(ctx context.Context, query *RealtimeStatsQuery) ([]*model.RealtimeStats, error) {
	var stats []*model.RealtimeStats
	db := r.db.WithContext(ctx).Model(&model.RealtimeStats{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.StatType != "" {
		db = db.Where("stat_type = ?", query.StatType)
	}
	if query.StatKey != "" {
		db = db.Where("stat_key = ?", query.StatKey)
	}
	if query.Period != "" {
		db = db.Where("period = ?", query.Period)
	}

	err := db.Order("created_at DESC").Find(&stats).Error
	return stats, err
}

func (r *analyticsRepository) UpdateRealtimeStats(ctx context.Context, stats *model.RealtimeStats) error {
	return r.db.WithContext(ctx).Save(stats).Error
}

func (r *analyticsRepository) IncrementRealtimeStats(ctx context.Context, merchantID *uuid.UUID, statType, statKey string, increment int64) error {
	return r.db.WithContext(ctx).
		Model(&model.RealtimeStats{}).
		Where("merchant_id = ? AND stat_type = ? AND stat_key = ?", merchantID, statType, statKey).
		Update("stat_value", gorm.Expr("stat_value + ?", increment)).Error
}

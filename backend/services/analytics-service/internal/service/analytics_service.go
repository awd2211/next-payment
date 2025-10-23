package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"payment-platform/analytics-service/internal/model"
	"payment-platform/analytics-service/internal/repository"
)

// AnalyticsService 分析服务接口
type AnalyticsService interface {
	// 支付指标
	GetPaymentMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.PaymentMetrics, error)
	GetPaymentSummary(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*PaymentSummary, error)
	RecordPayment(ctx context.Context, input *RecordPaymentInput) error

	// 商户指标
	GetMerchantMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.MerchantMetrics, error)
	GetMerchantSummary(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*MerchantSummary, error)
	RecordOrder(ctx context.Context, input *RecordOrderInput) error

	// 渠道指标
	GetChannelMetrics(ctx context.Context, channelCode string, startDate, endDate time.Time) ([]*model.ChannelMetrics, error)
	GetChannelSummary(ctx context.Context, channelCode string, startDate, endDate time.Time) (*ChannelSummary, error)
	RecordChannelTransaction(ctx context.Context, input *RecordChannelTransactionInput) error

	// 实时统计
	GetRealtimeStats(ctx context.Context, query *repository.RealtimeStatsQuery) ([]*model.RealtimeStats, error)
	IncrementStats(ctx context.Context, merchantID *uuid.UUID, statType, statKey string, increment int64) error
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

// NewAnalyticsService 创建分析服务实例
func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
	}
}

// Input structures

type RecordPaymentInput struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	IsRefund   bool      `json:"is_refund"`
}

type RecordOrderInput struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	Amount     int64     `json:"amount"`
	Fee        int64     `json:"fee"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	IsNewCustomer bool   `json:"is_new_customer"`
}

type RecordChannelTransactionInput struct {
	ChannelCode string `json:"channel_code"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	Latency     int    `json:"latency"`
}

// Summary structures

type PaymentSummary struct {
	TotalPayments     int     `json:"total_payments"`
	SuccessPayments   int     `json:"success_payments"`
	FailedPayments    int     `json:"failed_payments"`
	TotalAmount       int64   `json:"total_amount"`
	SuccessAmount     int64   `json:"success_amount"`
	TotalRefunds      int     `json:"total_refunds"`
	TotalRefundAmount int64   `json:"total_refund_amount"`
	SuccessRate       float64 `json:"success_rate"`
	AverageAmount     int64   `json:"average_amount"`
	Currency          string  `json:"currency"`
}

type MerchantSummary struct {
	TotalOrders        int     `json:"total_orders"`
	CompletedOrders    int     `json:"completed_orders"`
	CancelledOrders    int     `json:"cancelled_orders"`
	TotalRevenue       int64   `json:"total_revenue"`
	TotalFees          int64   `json:"total_fees"`
	NetRevenue         int64   `json:"net_revenue"`
	NewCustomers       int     `json:"new_customers"`
	ReturningCustomers int     `json:"returning_customers"`
	CompletionRate     float64 `json:"completion_rate"`
	Currency           string  `json:"currency"`
}

type ChannelSummary struct {
	TotalTransactions   int     `json:"total_transactions"`
	SuccessTransactions int     `json:"success_transactions"`
	FailedTransactions  int     `json:"failed_transactions"`
	TotalAmount         int64   `json:"total_amount"`
	SuccessAmount       int64   `json:"success_amount"`
	AverageLatency      int     `json:"average_latency"`
	SuccessRate         float64 `json:"success_rate"`
	Currency            string  `json:"currency"`
}

// Payment Metrics

func (s *analyticsService) GetPaymentMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.PaymentMetrics, error) {
	return s.analyticsRepo.GetPaymentMetrics(ctx, merchantID, startDate, endDate)
}

func (s *analyticsService) GetPaymentSummary(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*PaymentSummary, error) {
	metrics, err := s.analyticsRepo.GetPaymentMetrics(ctx, merchantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取支付指标失败: %w", err)
	}

	summary := &PaymentSummary{}
	for _, m := range metrics {
		summary.TotalPayments += m.TotalPayments
		summary.SuccessPayments += m.SuccessPayments
		summary.FailedPayments += m.FailedPayments
		summary.TotalAmount += m.TotalAmount
		summary.SuccessAmount += m.SuccessAmount
		summary.TotalRefunds += m.TotalRefunds
		summary.TotalRefundAmount += m.TotalRefundAmount
		summary.Currency = m.Currency
	}

	if summary.TotalPayments > 0 {
		summary.SuccessRate = float64(summary.SuccessPayments) / float64(summary.TotalPayments) * 100
		summary.AverageAmount = summary.TotalAmount / int64(summary.TotalPayments)
	}

	return summary, nil
}

func (s *analyticsService) RecordPayment(ctx context.Context, input *RecordPaymentInput) error {
	// TODO: 实现支付记录逻辑
	// 可以异步处理或批量更新指标
	return nil
}

// Merchant Metrics

func (s *analyticsService) GetMerchantMetrics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]*model.MerchantMetrics, error) {
	return s.analyticsRepo.GetMerchantMetrics(ctx, merchantID, startDate, endDate)
}

func (s *analyticsService) GetMerchantSummary(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*MerchantSummary, error) {
	metrics, err := s.analyticsRepo.GetMerchantMetrics(ctx, merchantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取商户指标失败: %w", err)
	}

	summary := &MerchantSummary{}
	for _, m := range metrics {
		summary.TotalOrders += m.TotalOrders
		summary.CompletedOrders += m.CompletedOrders
		summary.CancelledOrders += m.CancelledOrders
		summary.TotalRevenue += m.TotalRevenue
		summary.TotalFees += m.TotalFees
		summary.NetRevenue += m.NetRevenue
		summary.NewCustomers += m.NewCustomers
		summary.ReturningCustomers += m.ReturningCustomers
		summary.Currency = m.Currency
	}

	if summary.TotalOrders > 0 {
		summary.CompletionRate = float64(summary.CompletedOrders) / float64(summary.TotalOrders) * 100
	}

	return summary, nil
}

func (s *analyticsService) RecordOrder(ctx context.Context, input *RecordOrderInput) error {
	// TODO: 实现订单记录逻辑
	return nil
}

// Channel Metrics

func (s *analyticsService) GetChannelMetrics(ctx context.Context, channelCode string, startDate, endDate time.Time) ([]*model.ChannelMetrics, error) {
	return s.analyticsRepo.GetChannelMetrics(ctx, channelCode, startDate, endDate)
}

func (s *analyticsService) GetChannelSummary(ctx context.Context, channelCode string, startDate, endDate time.Time) (*ChannelSummary, error) {
	metrics, err := s.analyticsRepo.GetChannelMetrics(ctx, channelCode, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取渠道指标失败: %w", err)
	}

	summary := &ChannelSummary{}
	totalLatency := 0
	count := 0

	for _, m := range metrics {
		summary.TotalTransactions += m.TotalTransactions
		summary.SuccessTransactions += m.SuccessTransactions
		summary.FailedTransactions += m.FailedTransactions
		summary.TotalAmount += m.TotalAmount
		summary.SuccessAmount += m.SuccessAmount
		summary.Currency = m.Currency
		totalLatency += m.AverageLatency
		count++
	}

	if count > 0 {
		summary.AverageLatency = totalLatency / count
	}
	if summary.TotalTransactions > 0 {
		summary.SuccessRate = float64(summary.SuccessTransactions) / float64(summary.TotalTransactions) * 100
	}

	return summary, nil
}

func (s *analyticsService) RecordChannelTransaction(ctx context.Context, input *RecordChannelTransactionInput) error {
	// TODO: 实现渠道交易记录逻辑
	return nil
}

// Realtime Stats

func (s *analyticsService) GetRealtimeStats(ctx context.Context, query *repository.RealtimeStatsQuery) ([]*model.RealtimeStats, error) {
	return s.analyticsRepo.GetRealtimeStats(ctx, query)
}

func (s *analyticsService) IncrementStats(ctx context.Context, merchantID *uuid.UUID, statType, statKey string, increment int64) error {
	return s.analyticsRepo.IncrementRealtimeStats(ctx, merchantID, statType, statKey, increment)
}

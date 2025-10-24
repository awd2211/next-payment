package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AnalyticsClient Analytics Service HTTP客户端
type AnalyticsClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAnalyticsClient 创建Analytics客户端实例
func NewAnalyticsClient(baseURL string) *AnalyticsClient {
	return &AnalyticsClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// StatisticsResponse 统计响应
type StatisticsResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *StatisticsData    `json:"data"`
}

// StatisticsData 统计数据
type StatisticsData struct {
	TodayPayments      int     `json:"today_payments"`
	TodayAmount        int64   `json:"today_amount"`
	TodaySuccessRate   float64 `json:"today_success_rate"`
	MonthPayments      int     `json:"month_payments"`
	MonthAmount        int64   `json:"month_amount"`
	MonthSuccessRate   float64 `json:"month_success_rate"`
	PaymentTrend       []TrendData `json:"payment_trend"`
}

// TrendData 趋势数据
type TrendData struct {
	Date        string  `json:"date"`
	Payments    int     `json:"payments"`
	Amount      int64   `json:"amount"`
	SuccessRate float64 `json:"success_rate"`
}

// TransactionSummaryResponse 交易汇总响应
type TransactionSummaryResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    *TransactionSummaryData `json:"data"`
}

// TransactionSummaryData 交易汇总数据
type TransactionSummaryData struct {
	TotalPayments     int             `json:"total_payments"`
	SuccessPayments   int             `json:"success_payments"`
	FailedPayments    int             `json:"failed_payments"`
	TotalAmount       int64           `json:"total_amount"`
	SuccessAmount     int64           `json:"success_amount"`
	SuccessRate       float64         `json:"success_rate"`
	AverageAmount     int64           `json:"average_amount"`
	TotalRefunds      int             `json:"total_refunds"`
	TotalRefundAmount int64           `json:"total_refund_amount"`
	RefundRate        float64         `json:"refund_rate"`
	ChannelBreakdown  []ChannelStats  `json:"channel_breakdown"`
}

// ChannelStats 渠道统计
type ChannelStats struct {
	Channel     string  `json:"channel"`
	Payments    int     `json:"payments"`
	Amount      int64   `json:"amount"`
	SuccessRate float64 `json:"success_rate"`
	Percentage  float64 `json:"percentage"`
}

// GetStatistics 获取统计数据
func (c *AnalyticsClient) GetStatistics(ctx context.Context, merchantID uuid.UUID) (*StatisticsData, error) {
	url := fmt.Sprintf("%s/api/v1/statistics/merchant/%s", c.baseURL, merchantID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result StatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

// GetTransactionSummary 获取交易汇总
func (c *AnalyticsClient) GetTransactionSummary(ctx context.Context, merchantID uuid.UUID, startDate, endDate string) (*TransactionSummaryData, error) {
	url := fmt.Sprintf("%s/api/v1/statistics/merchant/%s/summary?start_date=%s&end_date=%s",
		c.baseURL, merchantID.String(), startDate, endDate)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result TransactionSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

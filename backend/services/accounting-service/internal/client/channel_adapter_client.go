package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/payment-platform/pkg/httpclient"
)

// ChannelAdapterClient Channel Adapter HTTP客户端（用于汇率查询）
type ChannelAdapterClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewChannelAdapterClient 创建Channel Adapter客户端实例（带熔断器）
func NewChannelAdapterClient(baseURL string) *ChannelAdapterClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("channel-adapter-service")

	return &ChannelAdapterClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// ExchangeRateResponse 汇率响应
type ExchangeRateResponse struct {
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	Source         string    `json:"source"`
	ValidFrom      time.Time `json:"valid_from"`
}

// GetExchangeRateAPIResponse API响应封装
type GetExchangeRateAPIResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    ExchangeRateResponse `json:"data"`
}

// GetExchangeRate 获取实时汇率（使用熔断器）
func (c *ChannelAdapterClient) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/exchange-rates/latest?from=%s&to=%s", c.baseURL, fromCurrency, toCurrency)

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求汇率API失败: %w", err)
	}

	// 解析响应
	var result GetExchangeRateAPIResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return 0, fmt.Errorf("解析汇率响应失败: %w", err)
	}

	if result.Code != 0 {
		return 0, fmt.Errorf("汇率API错误: %s", result.Message)
	}

	return result.Data.Rate, nil
}

// ExchangeRateSnapshotResponse 汇率快照响应
type ExchangeRateSnapshotResponse struct {
	BaseCurrency string             `json:"base_currency"`
	Rates        map[string]float64 `json:"rates"`
	Source       string             `json:"source"`
	SnapshotTime time.Time          `json:"snapshot_time"`
}

// GetSnapshotAPIResponse 快照API响应
type GetSnapshotAPIResponse struct {
	Code    int                          `json:"code"`
	Message string                       `json:"message"`
	Data    ExchangeRateSnapshotResponse `json:"data"`
}

// GetExchangeRateSnapshot 获取汇率快照（一次获取多个货币对的汇率）
func (c *ChannelAdapterClient) GetExchangeRateSnapshot(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/api/v1/exchange-rates/snapshot?base=%s", c.baseURL, baseCurrency)

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求汇率快照API失败: %w", err)
	}

	// 解析响应
	var result GetSnapshotAPIResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析汇率快照响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("汇率快照API错误: %s", result.Message)
	}

	return result.Data.Rates, nil
}

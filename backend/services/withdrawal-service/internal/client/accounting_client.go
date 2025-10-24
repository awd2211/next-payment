package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// AccountingClient Accounting Service HTTP客户端
type AccountingClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewAccountingClient 创建Accounting客户端实例（带熔断器）
func NewAccountingClient(baseURL string) *AccountingClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("accounting-service")

	return &AccountingClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// BalanceResponse 余额响应
type BalanceResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *BalanceData `json:"data"`
}

// BalanceData 余额数据
type BalanceData struct {
	MerchantID       string `json:"merchant_id"`
	AvailableBalance int64  `json:"available_balance"`
	FrozenBalance    int64  `json:"frozen_balance"`
	TotalBalance     int64  `json:"total_balance"`
}

// GetAvailableBalance 获取可用余额（使用熔断器）
func (c *AccountingClient) GetAvailableBalance(ctx context.Context, merchantID uuid.UUID) (int64, error) {
	url := fmt.Sprintf("%s/api/v1/balances/merchants/%s/summary", c.baseURL, merchantID.String())

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result BalanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return 0, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return 0, fmt.Errorf("业务错误: %s", result.Message)
	}

	if result.Data == nil {
		return 0, fmt.Errorf("余额数据为空")
	}

	return result.Data.AvailableBalance, nil
}

// DeductBalanceRequest 扣减余额请求
type DeductBalanceRequest struct {
	MerchantID      uuid.UUID `json:"merchant_id"`
	Amount          int64     `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	RelatedNo       string    `json:"related_no"`
	Description     string    `json:"description"`
}

// DeductBalanceResponse 扣减余额响应
type DeductBalanceResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DeductBalance 扣减余额（使用熔断器）
func (c *AccountingClient) DeductBalance(ctx context.Context, deductReq *DeductBalanceRequest) error {
	url := fmt.Sprintf("%s/api/v1/transactions", c.baseURL)

	// 创建请求
	req := &httpclient.Request{
		Method: "POST",
		URL:    url,
		Body:   deductReq,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result DeductBalanceResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("业务错误: %s", result.Message)
	}

	return nil
}

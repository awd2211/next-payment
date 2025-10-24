package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// WithdrawalClient Withdrawal Service HTTP客户端
type WithdrawalClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewWithdrawalClient 创建Withdrawal客户端实例（带熔断器）
func NewWithdrawalClient(baseURL string) *WithdrawalClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("withdrawal-service")

	return &WithdrawalClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// CreateWithdrawalRequest 创建提现请求
type CreateWithdrawalRequest struct {
	MerchantID    uuid.UUID `json:"merchant_id"`
	Amount        int64     `json:"amount"`
	Type          string    `json:"type"` // settlement_auto, settlement_manual
	BankAccountID uuid.UUID `json:"bank_account_id"`
	Remarks       string    `json:"remarks"`
	CreatedBy     uuid.UUID `json:"created_by"`
}

// CreateWithdrawalResponse 创建提现响应
type CreateWithdrawalResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *WithdrawalData `json:"data"`
}

// WithdrawalData 提现数据
type WithdrawalData struct {
	ID           string `json:"id"`
	WithdrawalNo string `json:"withdrawal_no"`
	Status       string `json:"status"`
}

// CreateWithdrawalForSettlement 为结算创建提现（使用熔断器）
func (c *WithdrawalClient) CreateWithdrawalForSettlement(ctx context.Context, withdrawalReq *CreateWithdrawalRequest) (string, error) {
	url := fmt.Sprintf("%s/api/v1/withdrawals", c.baseURL)

	// 创建请求
	req := &httpclient.Request{
		Method: "POST",
		URL:    url,
		Body:   withdrawalReq,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result CreateWithdrawalResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("业务错误: %s", result.Message)
	}

	if result.Data == nil {
		return "", fmt.Errorf("响应数据为空")
	}

	return result.Data.WithdrawalNo, nil
}

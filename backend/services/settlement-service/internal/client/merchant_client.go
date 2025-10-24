package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// MerchantClient Merchant Service HTTP客户端
type MerchantClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewMerchantClient 创建Merchant客户端实例（带熔断器）
func NewMerchantClient(baseURL string) *MerchantClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("merchant-service")

	return &MerchantClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// SettlementAccount 结算账户信息
type SettlementAccount struct {
	ID            uuid.UUID `json:"id"`
	MerchantID    uuid.UUID `json:"merchant_id"`
	AccountType   string    `json:"account_type"`
	BankName      string    `json:"bank_name"`
	BankCode      string    `json:"bank_code"`
	AccountNumber string    `json:"account_number"` // 加密存储的账号
	AccountName   string    `json:"account_name"`
	SwiftCode     string    `json:"swift_code"`
	IBAN          string    `json:"iban"`
	BankAddress   string    `json:"bank_address"`
	Currency      string    `json:"currency"`
	Country       string    `json:"country"`
	IsDefault     bool      `json:"is_default"`
	Status        string    `json:"status"`
}

// GetSettlementAccountsResponse 获取结算账户响应
type GetSettlementAccountsResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    []SettlementAccount `json:"data"`
}

// GetDefaultSettlementAccount 获取商户的默认结算账户（使用熔断器）
func (c *MerchantClient) GetDefaultSettlementAccount(ctx context.Context, merchantID uuid.UUID) (*SettlementAccount, error) {
	url := fmt.Sprintf("%s/api/v1/merchants/%s/settlement-accounts", c.baseURL, merchantID.String())

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
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result GetSettlementAccountsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	// 查找默认账户
	for _, account := range result.Data {
		if account.IsDefault && account.Status == "verified" {
			return &account, nil
		}
	}

	// 如果没有默认账户，返回第一个已验证的账户
	for _, account := range result.Data {
		if account.Status == "verified" {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("商户 %s 没有可用的结算账户", merchantID.String())
}

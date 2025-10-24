package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AccountingClient Accounting Service HTTP客户端
type AccountingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAccountingClient 创建Accounting客户端实例
func NewAccountingClient(baseURL string) *AccountingClient {
	return &AccountingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BalanceResponse 余额响应
type BalanceResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    *BalanceData  `json:"data"`
}

// BalanceData 余额数据
type BalanceData struct {
	MerchantID       string `json:"merchant_id"`
	AvailableBalance int64  `json:"available_balance"`
	FrozenBalance    int64  `json:"frozen_balance"`
	TotalBalance     int64  `json:"total_balance"`
}

// GetAvailableBalance 获取可用余额
func (c *AccountingClient) GetAvailableBalance(ctx context.Context, merchantID uuid.UUID) (int64, error) {
	url := fmt.Sprintf("%s/api/v1/balances/merchants/%s/summary", c.baseURL, merchantID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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

// DeductBalance 扣减余额
func (c *AccountingClient) DeductBalance(ctx context.Context, req *DeductBalanceRequest) error {
	url := fmt.Sprintf("%s/api/v1/transactions", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result DeductBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("业务错误: %s", result.Message)
	}

	return nil
}

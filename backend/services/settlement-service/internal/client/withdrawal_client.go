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

// WithdrawalClient Withdrawal Service HTTP客户端
type WithdrawalClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewWithdrawalClient 创建Withdrawal客户端实例
func NewWithdrawalClient(baseURL string) *WithdrawalClient {
	return &WithdrawalClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateWithdrawalRequest 创建提现请求
type CreateWithdrawalRequest struct {
	MerchantID    uuid.UUID `json:"merchant_id"`
	Amount        int64     `json:"amount"`
	Type          string    `json:"type"`           // settlement_auto, settlement_manual
	BankAccountID uuid.UUID `json:"bank_account_id"`
	Remarks       string    `json:"remarks"`
	CreatedBy     uuid.UUID `json:"created_by"`
}

// CreateWithdrawalResponse 创建提现响应
type CreateWithdrawalResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *WithdrawalData  `json:"data"`
}

// WithdrawalData 提现数据
type WithdrawalData struct {
	ID           string `json:"id"`
	WithdrawalNo string `json:"withdrawal_no"`
	Status       string `json:"status"`
}

// CreateWithdrawalForSettlement 为结算创建提现
func (c *WithdrawalClient) CreateWithdrawalForSettlement(ctx context.Context, req *CreateWithdrawalRequest) (string, error) {
	url := fmt.Sprintf("%s/api/v1/withdrawals", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result CreateWithdrawalResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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

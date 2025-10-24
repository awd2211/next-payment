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

// TransactionListResponse 交易列表响应
type TransactionListResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *TransactionListData `json:"data"`
}

// TransactionListData 交易列表数据
type TransactionListData struct {
	List     []TransactionInfo `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// TransactionInfo 交易信息
type TransactionInfo struct {
	ID            string `json:"id"`
	OrderNo       string `json:"order_no"`
	PaymentNo     string `json:"payment_no"`
	Amount        int64  `json:"amount"`
	Fee           int64  `json:"fee"`
	TransactionAt string `json:"transaction_at"`
}

// GetTransactions 获取交易列表用于结算（使用熔断器）
func (c *AccountingClient) GetTransactions(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]TransactionInfo, error) {
	// 构建URL
	url := fmt.Sprintf("%s/api/v1/transactions?merchant_id=%s&start_date=%s&end_date=%s&status=success&page_size=10000",
		c.baseURL,
		merchantID.String(),
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result TransactionListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	if result.Data == nil {
		return []TransactionInfo{}, nil
	}

	return result.Data.List, nil
}

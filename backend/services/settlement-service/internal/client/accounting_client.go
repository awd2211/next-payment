package client

import (
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

// TransactionListResponse 交易列表响应
type TransactionListResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
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

// GetTransactions 获取交易列表用于结算
func (c *AccountingClient) GetTransactions(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]TransactionInfo, error) {
	url := fmt.Sprintf("%s/api/v1/transactions?merchant_id=%s&start_date=%s&end_date=%s&status=success&page_size=10000",
		c.baseURL,
		merchantID.String(),
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

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

	var result TransactionListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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

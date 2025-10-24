package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// PaymentClient Payment Gateway HTTP客户端
type PaymentClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPaymentClient 创建Payment客户端实例
func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PaymentListResponse 支付列表响应
type PaymentListResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *PaymentListData `json:"data"`
}

// PaymentListData 支付列表数据
type PaymentListData struct {
	List     []PaymentInfo `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// PaymentInfo 支付信息
type PaymentInfo struct {
	ID            string `json:"id"`
	OrderNo       string `json:"order_no"`
	PaymentNo     string `json:"payment_no"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	Channel       string `json:"channel"`
	PayMethod     string `json:"pay_method"`
	CustomerEmail string `json:"customer_email"`
	CreatedAt     string `json:"created_at"`
	PaidAt        string `json:"paid_at,omitempty"`
}

// GetPayments 获取支付列表
func (c *PaymentClient) GetPayments(ctx context.Context, merchantID uuid.UUID, params map[string]string) (*PaymentListData, error) {
	url := fmt.Sprintf("%s/api/v1/payments?merchant_id=%s", c.baseURL, merchantID.String())

	// 添加查询参数
	for key, value := range params {
		if value != "" {
			url += fmt.Sprintf("&%s=%s", key, value)
		}
	}

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

	var result PaymentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

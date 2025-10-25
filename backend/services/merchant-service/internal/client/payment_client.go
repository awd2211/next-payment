package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// PaymentClient Payment Gateway HTTP客户端
type PaymentClient struct {
	*ServiceClient
}

// NewPaymentClient 创建Payment客户端实例（带熔断器）
func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "payment-gateway"),
	}
}

// PaymentListResponse 支付列表响应
type PaymentListResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
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
	// 构建查询路径
	path := fmt.Sprintf("/api/v1/payments?merchant_id=%s", merchantID.String())

	// 添加查询参数
	for key, value := range params {
		if value != "" {
			path += fmt.Sprintf("&%s=%s", key, value)
		}
	}

	// 通过熔断器发送请求
	resp, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result PaymentListResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

// RefundListResponse 退款列表响应
type RefundListResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *RefundListData `json:"data"`
}

// RefundListData 退款列表数据
type RefundListData struct {
	List     []RefundInfo `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// RefundInfo 退款信息
type RefundInfo struct {
	ID           string `json:"id"`
	PaymentNo    string `json:"payment_no"`
	RefundNo     string `json:"refund_no"`
	RefundAmount int64  `json:"refund_amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	Reason       string `json:"reason"`
	CreatedAt    string `json:"created_at"`
	RefundedAt   string `json:"refunded_at,omitempty"`
}

// GetRefunds 获取退款列表
func (c *PaymentClient) GetRefunds(ctx context.Context, merchantID uuid.UUID, params map[string]string) (*RefundListData, error) {
	// 构建查询路径
	path := fmt.Sprintf("/api/v1/refunds?merchant_id=%s", merchantID.String())

	// 添加查询参数
	for key, value := range params {
		if value != "" {
			path += fmt.Sprintf("&%s=%s", key, value)
		}
	}

	// 通过熔断器发送请求
	resp, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result RefundListResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ChannelClient Channel服务客户端
type ChannelClient struct {
	*ServiceClient
}

// NewChannelClient 创建Channel服务客户端（带熔断器）
func NewChannelClient(baseURL string) *ChannelClient {
	return &ChannelClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "channel-adapter"),
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	PaymentNo     string    `json:"payment_no"`
	MerchantID    uuid.UUID `json:"merchant_id"`
	Channel       string    `json:"channel"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	PayMethod     string    `json:"pay_method"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
	Description   string    `json:"description"`
	ReturnURL     string    `json:"return_url"`
	NotifyURL     string    `json:"notify_url"`
	Extra         map[string]interface{} `json:"extra"`
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    *PaymentResult `json:"data"`
}

// PaymentResult 支付结果
type PaymentResult struct {
	ChannelOrderNo string `json:"channel_order_no"` // 渠道订单号
	PaymentURL     string `json:"payment_url"`      // 支付URL（跳转）
	QRCodeURL      string `json:"qr_code_url"`      // 二维码URL
	Status         string `json:"status"`           // 支付状态
	Extra          map[string]interface{} `json:"extra"` // 扩展信息
}

// RefundRequest 退款请求
type RefundRequest struct {
	RefundNo       string `json:"refund_no"`
	PaymentNo      string `json:"payment_no"`
	ChannelOrderNo string `json:"channel_order_no"`
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	Reason         string `json:"reason"`
}

// RefundResponse 退款响应
type RefundResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    *RefundResult `json:"data"`
}

// RefundResult 退款结果
type RefundResult struct {
	ChannelRefundNo string `json:"channel_refund_no"` // 渠道退款单号
	Status          string `json:"status"`            // 退款状态
	RefundedAt      string `json:"refunded_at"`       // 退款时间
}

// QueryPaymentRequest 查询支付请求
type QueryPaymentRequest struct {
	PaymentNo      string `json:"payment_no"`
	ChannelOrderNo string `json:"channel_order_no"`
	Channel        string `json:"channel"`
}

// CreatePayment 创建支付
func (c *ChannelClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*PaymentResult, error) {
	resp, err := c.http.Post(ctx, "/api/v1/channel/payment", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Channel服务失败: %w", err)
	}

	var result CreatePaymentResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("创建支付失败: %s", result.Message)
	}

	return result.Data, nil
}

// CreateRefund 创建退款
func (c *ChannelClient) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResult, error) {
	resp, err := c.http.Post(ctx, "/api/v1/channel/refund", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Channel服务失败: %w", err)
	}

	var result RefundResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("创建退款失败: %s", result.Message)
	}

	return result.Data, nil
}

// QueryPayment 查询支付
func (c *ChannelClient) QueryPayment(ctx context.Context, req *QueryPaymentRequest) (*PaymentResult, error) {
	resp, err := c.http.Post(ctx, "/api/v1/channel/query", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Channel服务失败: %w", err)
	}

	var result CreatePaymentResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("查询支付失败: %s", result.Message)
	}

	return result.Data, nil
}

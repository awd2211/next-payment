package client

import (
	"context"
	"fmt"
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
	PaymentNo     string                 `json:"payment_no"`
	MerchantID    string                 `json:"merchant_id"` // 改为 string 类型
	Channel       string                 `json:"channel"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	PayMethod     string                 `json:"pay_method"`
	CustomerEmail string                 `json:"customer_email"`
	CustomerName  string                 `json:"customer_name"`
	Description   string                 `json:"description"`
	ReturnURL     string                 `json:"return_url"`
	NotifyURL     string                 `json:"notify_url"`
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
	ChannelTradeNo string                 `json:"channel_trade_no"` // 渠道交易号
	ChannelOrderNo string                 `json:"channel_order_no"` // 渠道订单号（兼容旧字段）
	PaymentURL     string                 `json:"payment_url"`      // 支付URL（跳转）
	QRCodeURL      string                 `json:"qr_code_url"`      // 二维码URL
	Status         string                 `json:"status"`           // 支付状态
	Extra          map[string]interface{} `json:"extra"`            // 扩展信息
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

// CancelPayment 取消支付（用于 Saga 补偿）
func (c *ChannelClient) CancelPayment(ctx context.Context, channelTradeNo string) error {
	path := fmt.Sprintf("/api/v1/channel/payment/%s/cancel", channelTradeNo)

	resp, err := c.http.Post(ctx, path, nil, nil)
	if err != nil {
		return fmt.Errorf("调用Channel服务失败: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := resp.ParseResponse(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("取消支付失败: %s", result.Message)
	}

	return nil
}

// ====== 预授权相关接口 ======

// CreatePreAuthRequest 创建预授权请求
type CreatePreAuthRequest struct {
	MerchantID string `json:"merchant_id"`
	OrderNo    string `json:"order_no"`
	PreAuthNo  string `json:"pre_auth_no"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Channel    string `json:"channel"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	ReturnURL  string `json:"return_url"`
	NotifyURL  string `json:"notify_url"`
}

// CreatePreAuthResponse 创建预授权响应
type CreatePreAuthResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *PreAuthResult   `json:"data"`
}

// PreAuthResult 预授权结果
type PreAuthResult struct {
	ChannelTradeNo string `json:"channel_trade_no"` // 渠道交易号
	PaymentURL     string `json:"payment_url"`      // 支付URL
	Status         string `json:"status"`           // 预授权状态
}

// CapturePreAuthRequest 确认预授权请求
type CapturePreAuthRequest struct {
	PreAuthNo      string `json:"pre_auth_no"`
	ChannelTradeNo string `json:"channel_trade_no"`
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
}

// CapturePreAuthResponse 确认预授权响应
type CapturePreAuthResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *CaptureResult     `json:"data"`
}

// CaptureResult 确认结果
type CaptureResult struct {
	PaymentTradeNo string `json:"payment_trade_no"` // 支付交易号
	Status         string `json:"status"`           // 状态
}

// CancelPreAuthRequest 取消预授权请求
type CancelPreAuthRequest struct {
	PreAuthNo      string `json:"pre_auth_no"`
	ChannelTradeNo string `json:"channel_trade_no"`
}

// CancelPreAuthResponse 取消预授权响应
type CancelPreAuthResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CreatePreAuth 创建预授权
func (c *ChannelClient) CreatePreAuth(ctx context.Context, req *CreatePreAuthRequest) (*CreatePreAuthResponse, error) {
	resp, err := c.http.Post(ctx, "/api/v1/channel/pre-auth", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Channel服务创建预授权失败: %w", err)
	}

	var result CreatePreAuthResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CapturePreAuth 确认预授权（扣款）
func (c *ChannelClient) CapturePreAuth(ctx context.Context, req *CapturePreAuthRequest) (*CapturePreAuthResponse, error) {
	resp, err := c.http.Post(ctx, "/api/v1/channel/pre-auth/capture", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Channel服务确认预授权失败: %w", err)
	}

	var result CapturePreAuthResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CancelPreAuth 取消预授权
func (c *ChannelClient) CancelPreAuth(ctx context.Context, req *CancelPreAuthRequest) (*CancelPreAuthResponse, error) {
	resp, err := c.http.Post(ctx, "/api/v1/channel/pre-auth/cancel", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Channel服务取消预授权失败: %w", err)
	}

	var result CancelPreAuthResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

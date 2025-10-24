package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// OrderClient Order服务客户端
type OrderClient struct {
	*ServiceClient
}

// NewOrderClient 创建Order服务客户端（带熔断器）
func NewOrderClient(baseURL string) *OrderClient {
	return &OrderClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "order-service"),
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	MerchantID    uuid.UUID `json:"merchant_id"`
	OrderNo       string    `json:"order_no"`
	PaymentNo     string    `json:"payment_no"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Channel       string    `json:"channel"`
	PayMethod     string    `json:"pay_method"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	CustomerIP    string    `json:"customer_ip"`
	Description   string    `json:"description"`
	Extra         string    `json:"extra"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Order `json:"data"`
}

// Order 订单信息
type Order struct {
	ID            uuid.UUID `json:"id"`
	MerchantID    uuid.UUID `json:"merchant_id"`
	OrderNo       string    `json:"order_no"`
	PaymentNo     string    `json:"payment_no"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Channel       string    `json:"channel"`
	PayMethod     string    `json:"pay_method"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
	Description   string    `json:"description"`
}

// UpdateOrderStatusRequest 更新订单状态请求
type UpdateOrderStatusRequest struct {
	Status          string `json:"status"`
	ChannelOrderNo  string `json:"channel_order_no,omitempty"`
	PaidAt          string `json:"paid_at,omitempty"`
	ErrorCode       string `json:"error_code,omitempty"`
	ErrorMsg        string `json:"error_msg,omitempty"`
}

// CreateOrder 创建订单
func (c *OrderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	resp, err := c.http.Post(ctx, "/api/v1/orders", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Order服务失败: %w", err)
	}

	var result CreateOrderResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("创建订单失败: %s", result.Message)
	}

	return result.Data, nil
}

// UpdateOrderStatus 更新订单状态
func (c *OrderClient) UpdateOrderStatus(ctx context.Context, paymentNo string, req *UpdateOrderStatusRequest) error {
	path := fmt.Sprintf("/api/v1/orders/%s/status", paymentNo)
	resp, err := c.http.Put(ctx, path, req, nil)
	if err != nil {
		return fmt.Errorf("调用Order服务失败: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := resp.ParseResponse(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("更新订单状态失败: %s", result.Message)
	}

	return nil
}

// GetOrder 获取订单
func (c *OrderClient) GetOrder(ctx context.Context, paymentNo string) (*Order, error) {
	path := fmt.Sprintf("/api/v1/orders/%s", paymentNo)
	resp, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Order服务失败: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *Order `json:"data"`
	}
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("获取订单失败: %s", result.Message)
	}

	return result.Data, nil
}

// CancelOrder 取消订单（用于 Saga 补偿）
func (c *OrderClient) CancelOrder(ctx context.Context, orderNo string, reason string) error {
	path := fmt.Sprintf("/api/v1/orders/%s/cancel", orderNo)
	req := map[string]string{
		"reason": reason,
	}

	resp, err := c.http.Post(ctx, path, req, nil)
	if err != nil {
		return fmt.Errorf("调用Order服务失败: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := resp.ParseResponse(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("取消订单失败: %s", result.Message)
	}

	return nil
}

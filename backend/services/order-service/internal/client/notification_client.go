package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// NotificationClient Notification Service HTTP客户端
type NotificationClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewNotificationClient 创建Notification客户端实例（带熔断器）
func NewNotificationClient(baseURL string) *NotificationClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("notification-service")

	return &NotificationClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	MerchantID  uuid.UUID              `json:"merchant_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	Type        string                 `json:"type"` // order_created, order_paid, order_cancelled, order_refunded
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Email       string                 `json:"email,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Priority    string                 `json:"priority,omitempty"` // low, medium, high
}

// SendNotificationResponse 发送通知响应
type SendNotificationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		NotificationID string `json:"notification_id"`
		Status         string `json:"status"`
	} `json:"data"`
}

// SendOrderNotification 发送订单相关通知
func (c *NotificationClient) SendOrderNotification(ctx context.Context, req *SendNotificationRequest) error {
	// 构建URL
	url := fmt.Sprintf("%s/api/v1/notifications/send", c.baseURL)

	// 创建请求
	httpReq := &httpclient.Request{
		Method: "POST",
		URL:    url,
		Body:   req,
		Ctx:    ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}

	// 解析响应
	var result SendNotificationResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("通知服务错误: %s", result.Message)
	}

	return nil
}

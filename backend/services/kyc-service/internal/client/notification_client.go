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
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}
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
	Type        string                 `json:"type"` // kyc_submitted, kyc_approved, kyc_rejected, kyc_pending_info
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

// SendKYCNotification 发送KYC相关通知
func (c *NotificationClient) SendKYCNotification(ctx context.Context, req *SendNotificationRequest) error {
	url := fmt.Sprintf("%s/api/v1/notifications/send", c.baseURL)

	httpReq := &httpclient.Request{
		Method: "POST",
		URL:    url,
		Body:   req,
		Ctx:    ctx,
	}

	resp, err := c.breaker.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}

	var result SendNotificationResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("通知服务错误: %s", result.Message)
	}

	return nil
}

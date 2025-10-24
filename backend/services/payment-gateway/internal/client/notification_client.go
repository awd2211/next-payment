package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// NotificationClient Notification Service HTTP客户端
type NotificationClient struct {
	*ServiceClient
}

// NewNotificationClient 创建Notification客户端实例（带熔断器）
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "notification-service"),
	}
}

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	MerchantID  uuid.UUID              `json:"merchant_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	Type        string                 `json:"type"` // payment_success, payment_failed, refund_success, settlement_complete
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

// SendPaymentNotification 发送支付相关通知
func (c *NotificationClient) SendPaymentNotification(ctx context.Context, req *SendNotificationRequest) error {
	resp, err := c.http.Post(ctx, "/api/v1/notifications/send", req, nil)
	if err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}

	var result SendNotificationResponse
	if err := resp.ParseResponse(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("通知服务错误: %s", result.Message)
	}

	return nil
}

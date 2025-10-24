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

// UnreadCountResponse 未读通知数响应
type UnreadCountResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    *UnreadCount `json:"data"`
}

// UnreadCount 未读通知数
type UnreadCount struct {
	Total  int `json:"total"`
	System int `json:"system"`  // 系统通知
	Trade  int `json:"trade"`   // 交易通知
	Alert  int `json:"alert"`   // 告警通知
}

// GetUnreadCount 获取未读通知数
func (c *NotificationClient) GetUnreadCount(ctx context.Context, merchantID uuid.UUID) (*UnreadCount, error) {
	path := fmt.Sprintf("/api/v1/notifications/merchants/%s/unread/count", merchantID.String())

	resp, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	var result UnreadCountResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

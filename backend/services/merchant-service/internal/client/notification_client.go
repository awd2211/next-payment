package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NotificationClient Notification Service HTTP客户端
type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNotificationClient 创建Notification客户端实例
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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
	url := fmt.Sprintf("%s/api/v1/notifications/merchants/%s/unread/count", c.baseURL, merchantID.String())

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

	var result UnreadCountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}

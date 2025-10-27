package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// NotificationClient Notification Service HTTP客户端
type NotificationClient struct {
	baseURL    string
	httpClient *httpclient.Client
}

// NewNotificationClient 创建Notification客户端实例
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		httpClient: httpclient.NewClient(&httpclient.Config{
			Timeout:       10 * time.Second,
			MaxRetries:    3,
			RetryDelay:    time.Second,
			EnableLogging: false,
		}),
	}
}

// SendQuotaAlertRequest 发送配额预警请求
type SendQuotaAlertRequest struct {
	MerchantID   uuid.UUID              `json:"merchant_id"`
	Type         string                 `json:"type"` // quota_warning, quota_critical
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Email        string                 `json:"email,omitempty"`
	Phone        string                 `json:"phone,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Priority     string                 `json:"priority,omitempty"` // low, medium, high
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

// SendQuotaAlert 发送配额预警通知
func (c *NotificationClient) SendQuotaAlert(ctx context.Context, req *SendQuotaAlertRequest) error {
	url := fmt.Sprintf("%s/api/v1/notifications/send", c.baseURL)

	resp, err := c.httpClient.Post(url, req, nil)
	if err != nil {
		return fmt.Errorf("发送配额预警通知失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned status: %d", resp.StatusCode)
	}

	var result SendNotificationResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("通知服务错误: %s", result.Message)
	}

	return nil
}

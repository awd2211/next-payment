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
	MerchantID uuid.UUID              `json:"merchant_id"`
	Type       string                 `json:"type"`     // approval_required, withdrawal_approved, withdrawal_rejected, withdrawal_completed
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Priority   string                 `json:"priority"` // low, medium, high
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// SendNotificationResponse 发送通知响应
type SendNotificationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SendNotification 发送通知（使用熔断器）
func (c *NotificationClient) SendNotification(ctx context.Context, notifyReq *SendNotificationRequest) error {
	url := fmt.Sprintf("%s/api/v1/notifications/send", c.baseURL)

	// 创建请求
	req := &httpclient.Request{
		Method: "POST",
		URL:    url,
		Body:   notifyReq,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result SendNotificationResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("业务错误: %s", result.Message)
	}

	return nil
}

// SendApprovalNotification 发送审批通知
func (c *NotificationClient) SendApprovalNotification(ctx context.Context, merchantID uuid.UUID, withdrawalNo string, amount int64) error {
	return c.SendNotification(ctx, &SendNotificationRequest{
		MerchantID: merchantID,
		Type:       "approval_required",
		Title:      "提现申请待审批",
		Content:    fmt.Sprintf("您有一笔提现申请需要审批，提现单号：%s，金额：%.2f元", withdrawalNo, float64(amount)/100),
		Priority:   "high",
		Extra: map[string]interface{}{
			"withdrawal_no": withdrawalNo,
			"amount":        amount,
		},
	})
}

// SendWithdrawalStatusNotification 发送提现状态通知
func (c *NotificationClient) SendWithdrawalStatusNotification(ctx context.Context, merchantID uuid.UUID, withdrawalNo, status string, amount int64) error {
	var title, notificationType string
	var priority string

	switch status {
	case "approved":
		title = "提现审批通过"
		notificationType = "withdrawal_approved"
		priority = "medium"
	case "rejected":
		title = "提现审批被拒绝"
		notificationType = "withdrawal_rejected"
		priority = "medium"
	case "completed":
		title = "提现已完成"
		notificationType = "withdrawal_completed"
		priority = "high"
	case "failed":
		title = "提现失败"
		notificationType = "withdrawal_failed"
		priority = "high"
	default:
		title = "提现状态更新"
		notificationType = "withdrawal_status_update"
		priority = "low"
	}

	content := fmt.Sprintf("提现单号：%s，金额：%.2f元，状态：%s", withdrawalNo, float64(amount)/100, title)

	return c.SendNotification(ctx, &SendNotificationRequest{
		MerchantID: merchantID,
		Type:       notificationType,
		Title:      title,
		Content:    content,
		Priority:   priority,
		Extra: map[string]interface{}{
			"withdrawal_no": withdrawalNo,
			"amount":        amount,
			"status":        status,
		},
	})
}

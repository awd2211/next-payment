package client

import (
	"bytes"
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

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	MerchantID uuid.UUID              `json:"merchant_id"`
	Type       string                 `json:"type"`        // approval_required, withdrawal_approved, withdrawal_rejected, withdrawal_completed
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Priority   string                 `json:"priority"`    // low, medium, high
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// SendNotificationResponse 发送通知响应
type SendNotificationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SendNotification 发送通知
func (c *NotificationClient) SendNotification(ctx context.Context, req *SendNotificationRequest) error {
	url := fmt.Sprintf("%s/api/v1/notifications/send", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result SendNotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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

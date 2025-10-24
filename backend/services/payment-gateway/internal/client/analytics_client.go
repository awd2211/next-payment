package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AnalyticsClient Analytics Service HTTP客户端
type AnalyticsClient struct {
	*ServiceClient
}

// NewAnalyticsClient 创建Analytics客户端实例（带熔断器）
func NewAnalyticsClient(baseURL string) *AnalyticsClient {
	return &AnalyticsClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "analytics-service"),
	}
}

// PaymentEventRequest 支付事件请求
type PaymentEventRequest struct {
	EventType    string    `json:"event_type"` // payment_created, payment_success, payment_failed, refund_created, refund_success
	MerchantID   uuid.UUID `json:"merchant_id"`
	PaymentNo    string    `json:"payment_no,omitempty"`
	RefundNo     string    `json:"refund_no,omitempty"`
	OrderNo      string    `json:"order_no,omitempty"`
	Amount       int64     `json:"amount"`
	Currency     string    `json:"currency"`
	Channel      string    `json:"channel"`
	PayMethod    string    `json:"pay_method,omitempty"`
	Status       string    `json:"status"`
	ErrorCode    string    `json:"error_code,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentEventResponse 支付事件响应
type PaymentEventResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		EventID string `json:"event_id"`
		Status  string `json:"status"`
	} `json:"data"`
}

// PushPaymentEvent 推送支付事件
func (c *AnalyticsClient) PushPaymentEvent(ctx context.Context, req *PaymentEventRequest) error {
	resp, err := c.http.Post(ctx, "/api/v1/events/payment", req, nil)
	if err != nil {
		// Analytics 推送失败不应该影响主流程，只记录日志
		return fmt.Errorf("推送分析事件失败（非致命）: %w", err)
	}

	var result PaymentEventResponse
	if err := resp.ParseResponse(&result); err != nil {
		return fmt.Errorf("解析分析事件响应失败: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("分析服务错误: %s", result.Message)
	}

	return nil
}

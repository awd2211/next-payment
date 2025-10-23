package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookProvider Webhook 提供商
type WebhookProvider struct {
	client *http.Client
}

// NewWebhookProvider 创建 Webhook 提供商实例
func NewWebhookProvider() *WebhookProvider {
	return &WebhookProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WebhookRequest Webhook 请求
type WebhookRequest struct {
	URL       string                 `json:"url"`        // Webhook URL
	Secret    string                 `json:"secret"`     // 签名密钥
	EventType string                 `json:"event_type"` // 事件类型
	EventID   string                 `json:"event_id"`   // 事件ID
	Timestamp int64                  `json:"timestamp"`  // 时间戳
	Data      map[string]interface{} `json:"data"`       // 事件数据
	Timeout   int                    `json:"timeout"`    // 超时时间（秒）
}

// WebhookResponse Webhook 响应
type WebhookResponse struct {
	Status       string `json:"status"`        // 状态
	HTTPStatus   int    `json:"http_status"`   // HTTP状态码
	ResponseBody string `json:"response_body"` // 响应内容
	Duration     int64  `json:"duration"`      // 请求耗时（毫秒）
	ErrorMessage string `json:"error_message"` // 错误信息
}

// Send 发送 Webhook
func (p *WebhookProvider) Send(ctx context.Context, req *WebhookRequest) (*WebhookResponse, error) {
	startTime := time.Now()

	// 构造请求体
	payload := map[string]interface{}{
		"event_type": req.EventType,
		"event_id":   req.EventID,
		"timestamp":  req.Timestamp,
		"data":       req.Data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化 Webhook 数据失败: %w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", req.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "PaymentPlatform-Webhook/1.0")

	// 计算签名
	signature := p.calculateSignature(payloadBytes, req.Secret)
	httpReq.Header.Set("X-Webhook-Signature", signature)
	httpReq.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", req.Timestamp))
	httpReq.Header.Set("X-Webhook-Event-Type", req.EventType)
	httpReq.Header.Set("X-Webhook-Event-ID", req.EventID)

	// 设置超时
	if req.Timeout > 0 {
		p.client.Timeout = time.Duration(req.Timeout) * time.Second
	}

	// 发送请求
	resp, err := p.client.Do(httpReq)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		return &WebhookResponse{
			Status:       "failed",
			HTTPStatus:   0,
			Duration:     duration,
			ErrorMessage: err.Error(),
		}, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, _ := io.ReadAll(resp.Body)
	duration := time.Since(startTime).Milliseconds()

	// 构造响应
	response := &WebhookResponse{
		HTTPStatus:   resp.StatusCode,
		ResponseBody: string(respBody),
		Duration:     duration,
	}

	// 判断是否成功
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		response.Status = "delivered"
	} else {
		response.Status = "failed"
		response.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return response, nil
}

// calculateSignature 计算签名
func (p *WebhookProvider) calculateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature 验证签名
func (p *WebhookProvider) VerifySignature(payload []byte, signature, secret string) bool {
	expectedSignature := p.calculateSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

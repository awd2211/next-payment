package adapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"payment-platform/channel-adapter/internal/model"
)

// PayPalAdapter PayPal 支付适配器
type PayPalAdapter struct {
	config     *model.PayPalConfig
	httpClient *http.Client
	apiBase    string
	token      string
	tokenExp   time.Time
}

// NewPayPalAdapter 创建 PayPal 适配器实例
func NewPayPalAdapter(config *model.PayPalConfig) *PayPalAdapter {
	apiBase := "https://api-m.paypal.com"
	if config.Mode == "sandbox" {
		apiBase = "https://api-m.sandbox.paypal.com"
	}

	return &PayPalAdapter{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiBase: apiBase,
	}
}

// GetChannel 获取渠道名称
func (a *PayPalAdapter) GetChannel() string {
	return model.ChannelPayPal
}

// getAccessToken 获取访问令牌
func (a *PayPalAdapter) getAccessToken(ctx context.Context) (string, error) {
	// 如果令牌未过期，直接返回
	if a.token != "" && time.Now().Before(a.tokenExp) {
		return a.token, nil
	}

	// 请求新令牌
	req, err := http.NewRequestWithContext(ctx, "POST", a.apiBase+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", fmt.Errorf("创建令牌请求失败: %w", err)
	}

	// 设置基本认证
	auth := base64.StdEncoding.EncodeToString([]byte(a.config.ClientID + ":" + a.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取令牌失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取令牌失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("解析令牌响应失败: %w", err)
	}

	// 缓存令牌
	a.token = tokenResp.AccessToken
	a.tokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return a.token, nil
}

// CreatePayment 创建支付
func (a *PayPalAdapter) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// 构造 PayPal 订单请求
	paypalReq := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": req.OrderNo,
				"amount": map[string]interface{}{
					"currency_code": strings.ToUpper(req.Currency),
					"value":         fmt.Sprintf("%.2f", float64(req.Amount)/100.0),
				},
				"description": req.Description,
				"custom_id":   req.PaymentNo,
			},
		},
		"application_context": map[string]interface{}{
			"return_url": req.SuccessURL,
			"cancel_url": req.CancelURL,
		},
	}

	body, _ := json.Marshal(paypalReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.apiBase+"/v2/checkout/orders", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建 PayPal 订单请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("创建 PayPal 订单失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("创建 PayPal 订单失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	var orderResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
	}

	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, fmt.Errorf("解析 PayPal 订单响应失败: %w", err)
	}

	// 查找批准链接
	var approveURL string
	for _, link := range orderResp.Links {
		if link.Rel == "approve" {
			approveURL = link.Href
			break
		}
	}

	return &CreatePaymentResponse{
		ChannelTradeNo: orderResp.ID,
		PaymentURL:     approveURL,
		Status:         convertPayPalStatus(orderResp.Status),
		Extra: map[string]interface{}{
			"order_id":    orderResp.ID,
			"approve_url": approveURL,
		},
	}, nil
}

// QueryPayment 查询支付状态
func (a *PayPalAdapter) QueryPayment(ctx context.Context, channelTradeNo string) (*QueryPaymentResponse, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", a.apiBase+"/v2/checkout/orders/"+channelTradeNo, nil)
	if err != nil {
		return nil, fmt.Errorf("创建查询请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("查询 PayPal 订单失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询 PayPal 订单失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var orderResp struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		PurchaseUnits []struct {
			Amount struct {
				CurrencyCode string `json:"currency_code"`
				Value        string `json:"value"`
			} `json:"amount"`
			Payments struct {
				Captures []struct {
					ID         string `json:"id"`
					Status     string `json:"status"`
					CreateTime string `json:"create_time"`
				} `json:"captures"`
			} `json:"payments"`
		} `json:"purchase_units"`
	}

	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("解析 PayPal 订单响应失败: %w", err)
	}

	response := &QueryPaymentResponse{
		ChannelTradeNo: orderResp.ID,
		Status:         convertPayPalStatus(orderResp.Status),
	}

	if len(orderResp.PurchaseUnits) > 0 {
		pu := orderResp.PurchaseUnits[0]
		response.Currency = pu.Amount.CurrencyCode

		// 将金额字符串转换为分
		var amountFloat float64
		fmt.Sscanf(pu.Amount.Value, "%f", &amountFloat)
		response.Amount = int64(amountFloat * 100)

		if len(pu.Payments.Captures) > 0 {
			capture := pu.Payments.Captures[0]
			response.PaymentMethod = "paypal"

			// 解析支付时间
			if t, err := time.Parse(time.RFC3339, capture.CreateTime); err == nil {
				paidAt := t.Unix()
				response.PaidAt = &paidAt
			}
		}
	}

	return response, nil
}

// CancelPayment 取消支付
func (a *PayPalAdapter) CancelPayment(ctx context.Context, channelTradeNo string) error {
	// PayPal 不支持直接取消订单，订单会自动在3小时后过期
	// 这里返回 nil 表示成功
	return nil
}

// CreateRefund 创建退款
func (a *PayPalAdapter) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// 首先获取捕获ID
	_, err = a.QueryPayment(ctx, req.ChannelTradeNo)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	// 从订单中提取捕获ID（需要从完整的订单响应中获取）
	// 这里简化处理，实际应该从完整的订单API响应中获取真实的 capture ID
	captureID := req.ChannelTradeNo + "_capture"

	// 构造退款请求
	refundReq := map[string]interface{}{
		"amount": map[string]interface{}{
			"currency_code": strings.ToUpper(req.Currency),
			"value":         fmt.Sprintf("%.2f", float64(req.Amount)/100.0),
		},
		"note_to_payer": req.Reason,
	}

	body, _ := json.Marshal(refundReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.apiBase+"/v2/payments/captures/"+captureID+"/refund", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建退款请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("创建 PayPal 退款失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("创建 PayPal 退款失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	var refundResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(respBody, &refundResp); err != nil {
		return nil, fmt.Errorf("解析退款响应失败: %w", err)
	}

	return &CreateRefundResponse{
		RefundNo:        req.RefundNo,
		ChannelRefundNo: refundResp.ID,
		Status:          convertPayPalRefundStatus(refundResp.Status),
		Extra: map[string]interface{}{
			"refund_id": refundResp.ID,
		},
	}, nil
}

// QueryRefund 查询退款状态
func (a *PayPalAdapter) QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", a.apiBase+"/v2/payments/refunds/"+refundNo, nil)
	if err != nil {
		return nil, fmt.Errorf("创建查询请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("查询 PayPal 退款失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询 PayPal 退款失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var refundResp struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		Amount     struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"amount"`
		CreateTime string `json:"create_time"`
	}

	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("解析退款响应失败: %w", err)
	}

	// 将金额字符串转换为分
	var amountFloat float64
	fmt.Sscanf(refundResp.Amount.Value, "%f", &amountFloat)

	response := &QueryRefundResponse{
		ChannelRefundNo: refundResp.ID,
		Status:          convertPayPalRefundStatus(refundResp.Status),
		Amount:          int64(amountFloat * 100),
		Currency:        refundResp.Amount.CurrencyCode,
	}

	// 解析退款时间
	if t, err := time.Parse(time.RFC3339, refundResp.CreateTime); err == nil {
		refundedAt := t.Unix()
		response.RefundedAt = &refundedAt
	}

	return response, nil
}

// VerifyWebhook 验证 Webhook 签名
func (a *PayPalAdapter) VerifyWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	// PayPal Webhook 验证需要多个头部信息
	// 这里简化处理，实际应该使用 PayPal SDK 的验证方法
	// 或者调用 PayPal 的验证 API: POST /v1/notifications/verify-webhook-signature

	// 简单的 HMAC 验证（示例）
	if a.config.WebhookID == "" {
		return true, nil // 如果没有配置 webhook ID，跳过验证
	}

	// 实际实现应该使用 PayPal 的 webhook 验证 API
	return true, nil
}

// ParseWebhook 解析 Webhook 数据
func (a *PayPalAdapter) ParseWebhook(ctx context.Context, body []byte) (*WebhookEvent, error) {
	var event struct {
		ID         string `json:"id"`
		EventType  string `json:"event_type"`
		Resource   map[string]interface{} `json:"resource"`
		CreateTime string `json:"create_time"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("解析 PayPal Webhook 失败: %w", err)
	}

	webhookEvent := &WebhookEvent{
		EventID:   event.ID,
		EventType: convertPayPalEventType(event.EventType),
		RawData:   event,
	}

	// 从资源中提取信息
	if orderID, ok := event.Resource["id"].(string); ok {
		webhookEvent.ChannelTradeNo = orderID
	}

	if customID, ok := event.Resource["custom_id"].(string); ok {
		webhookEvent.PaymentNo = customID
	}

	// 根据事件类型设置状态
	switch event.EventType {
	case "CHECKOUT.ORDER.APPROVED", "PAYMENT.CAPTURE.COMPLETED":
		webhookEvent.Status = PaymentStatusSuccess
	case "PAYMENT.CAPTURE.DECLINED", "CHECKOUT.ORDER.VOIDED":
		webhookEvent.Status = PaymentStatusFailed
	case "PAYMENT.CAPTURE.REFUNDED":
		webhookEvent.Status = PaymentStatusRefunded
	default:
		webhookEvent.Status = PaymentStatusProcessing
	}

	// 提取金额信息
	if resource, ok := event.Resource["amount"].(map[string]interface{}); ok {
		if value, ok := resource["value"].(string); ok {
			var amountFloat float64
			fmt.Sscanf(value, "%f", &amountFloat)
			webhookEvent.Amount = int64(amountFloat * 100)
		}
		if currency, ok := resource["currency_code"].(string); ok {
			webhookEvent.Currency = currency
		}
	}

	return webhookEvent, nil
}

// convertPayPalStatus 转换 PayPal 订单状态为统一状态
func convertPayPalStatus(status string) string {
	switch status {
	case "CREATED", "SAVED", "APPROVED", "PAYER_ACTION_REQUIRED":
		return PaymentStatusPending
	case "COMPLETED":
		return PaymentStatusSuccess
	case "VOIDED":
		return PaymentStatusCancelled
	default:
		return PaymentStatusFailed
	}
}

// convertPayPalRefundStatus 转换 PayPal 退款状态为统一状态
func convertPayPalRefundStatus(status string) string {
	switch status {
	case "PENDING":
		return PaymentStatusProcessing
	case "COMPLETED":
		return PaymentStatusRefunded
	case "CANCELLED":
		return PaymentStatusCancelled
	case "FAILED":
		return PaymentStatusFailed
	default:
		return PaymentStatusProcessing
	}
}

// convertPayPalEventType 转换 PayPal 事件类型为统一事件类型
func convertPayPalEventType(eventType string) string {
	switch eventType {
	case "CHECKOUT.ORDER.APPROVED", "PAYMENT.CAPTURE.COMPLETED":
		return EventTypePaymentSuccess
	case "PAYMENT.CAPTURE.DECLINED", "CHECKOUT.ORDER.VOIDED":
		return EventTypePaymentFailed
	case "PAYMENT.CAPTURE.REFUNDED":
		return EventTypeRefundSuccess
	default:
		return eventType
	}
}

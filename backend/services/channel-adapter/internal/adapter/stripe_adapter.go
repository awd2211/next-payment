package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"payment-platform/channel-adapter/internal/model"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/charge"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeAdapter Stripe 支付适配器
type StripeAdapter struct {
	config *model.StripeConfig
}

// NewStripeAdapter 创建 Stripe 适配器实例
func NewStripeAdapter(config *model.StripeConfig) *StripeAdapter {
	// 设置 Stripe API 密钥
	stripe.Key = config.APIKey

	return &StripeAdapter{
		config: config,
	}
}

// GetChannel 获取渠道名称
func (a *StripeAdapter) GetChannel() string {
	return model.ChannelStripe
}

// CreatePayment 创建支付
func (a *StripeAdapter) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 创建 PaymentIntent
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(req.Amount),
		Currency:    stripe.String(req.Currency),
		Description: stripe.String(req.Description),
		Metadata: map[string]string{
			"payment_no": req.PaymentNo,
			"order_no":   req.OrderNo,
		},
	}

	// 设置客户信息
	if req.CustomerEmail != "" {
		params.ReceiptEmail = stripe.String(req.CustomerEmail)
	}

	// 设置账单描述符
	if a.config.StatementDescriptor != "" {
		params.StatementDescriptor = stripe.String(a.config.StatementDescriptor)
	}

	// 设置捕获方式
	if a.config.CaptureMethod != "" {
		params.CaptureMethod = stripe.String(a.config.CaptureMethod)
	}

	// 设置自动支付方式
	params.AutomaticPaymentMethods = &stripe.PaymentIntentAutomaticPaymentMethodsParams{
		Enabled: stripe.Bool(true),
	}

	// 调用 Stripe API 创建支付意图
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("创建 Stripe PaymentIntent 失败: %w", err)
	}

	// 构造响应
	response := &CreatePaymentResponse{
		ChannelTradeNo: pi.ID,
		ClientSecret:   pi.ClientSecret,
		Status:         convertStripeStatus(pi.Status),
		Extra: map[string]interface{}{
			"payment_intent_id": pi.ID,
			"client_secret":     pi.ClientSecret,
		},
	}

	return response, nil
}

// QueryPayment 查询支付状态
func (a *StripeAdapter) QueryPayment(ctx context.Context, channelTradeNo string) (*QueryPaymentResponse, error) {
	// 查询 PaymentIntent
	pi, err := paymentintent.Get(channelTradeNo, nil)
	if err != nil {
		return nil, fmt.Errorf("查询 Stripe PaymentIntent 失败: %w", err)
	}

	// 构造响应
	response := &QueryPaymentResponse{
		ChannelTradeNo: pi.ID,
		Status:         convertStripeStatus(pi.Status),
		Amount:         pi.Amount,
		Currency:       string(pi.Currency),
	}

	// 支付时间
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		created := pi.Created
		response.PaidAt = &created
	}

	// 支付方式详情 - 通过 LatestCharge 获取
	if pi.LatestCharge != nil && pi.LatestCharge.ID != "" {
		ch, err := charge.Get(pi.LatestCharge.ID, nil)
		if err == nil && ch.PaymentMethodDetails != nil {
			response.PaymentMethod = string(ch.PaymentMethodDetails.Type)

			// 根据支付方式类型提取详情
			details := make(map[string]interface{})
			switch ch.PaymentMethodDetails.Type {
			case stripe.ChargePaymentMethodDetailsTypeCard:
				if ch.PaymentMethodDetails.Card != nil {
					details["brand"] = ch.PaymentMethodDetails.Card.Brand
					details["last4"] = ch.PaymentMethodDetails.Card.Last4
					details["exp_month"] = ch.PaymentMethodDetails.Card.ExpMonth
					details["exp_year"] = ch.PaymentMethodDetails.Card.ExpYear
					details["country"] = ch.PaymentMethodDetails.Card.Country
				}
			}
			response.PaymentMethodDetails = details
		}
	}

	return response, nil
}

// CancelPayment 取消支付
func (a *StripeAdapter) CancelPayment(ctx context.Context, channelTradeNo string) error {
	// 取消 PaymentIntent
	params := &stripe.PaymentIntentCancelParams{}
	_, err := paymentintent.Cancel(channelTradeNo, params)
	if err != nil {
		return fmt.Errorf("取消 Stripe PaymentIntent 失败: %w", err)
	}

	return nil
}

// CreateRefund 创建退款
func (a *StripeAdapter) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error) {
	// 创建退款
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.ChannelTradeNo),
		Amount:        stripe.Int64(req.Amount),
		Reason:        stripe.String(string(stripe.RefundReasonRequestedByCustomer)),
		Metadata: map[string]string{
			"refund_no":  req.RefundNo,
			"payment_no": req.PaymentNo,
			"reason":     req.Reason,
		},
	}

	r, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("创建 Stripe Refund 失败: %w", err)
	}

	// 构造响应
	response := &CreateRefundResponse{
		RefundNo:        req.RefundNo,
		ChannelRefundNo: r.ID,
		Status:          convertRefundStatus(r.Status),
		Extra: map[string]interface{}{
			"refund_id": r.ID,
		},
	}

	return response, nil
}

// QueryRefund 查询退款状态
func (a *StripeAdapter) QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error) {
	// 查询退款
	r, err := refund.Get(refundNo, nil)
	if err != nil {
		return nil, fmt.Errorf("查询 Stripe Refund 失败: %w", err)
	}

	// 构造响应
	response := &QueryRefundResponse{
		RefundNo:        r.Metadata["refund_no"],
		ChannelRefundNo: r.ID,
		Status:          convertRefundStatus(r.Status),
		Amount:          r.Amount,
		Currency:        string(r.Currency),
	}

	// 退款时间
	if r.Status == stripe.RefundStatusSucceeded {
		created := r.Created
		response.RefundedAt = &created
	}

	return response, nil
}

// VerifyWebhook 验证 Webhook 签名
func (a *StripeAdapter) VerifyWebhook(ctx context.Context, signature string, body []byte) (bool, error) {
	// 验证 Webhook 签名
	_, err := webhook.ConstructEvent(body, signature, a.config.WebhookSecret)
	if err != nil {
		return false, fmt.Errorf("验证 Stripe Webhook 签名失败: %w", err)
	}

	return true, nil
}

// ParseWebhook 解析 Webhook 数据
func (a *StripeAdapter) ParseWebhook(ctx context.Context, body []byte) (*WebhookEvent, error) {
	// 解析事件
	var event stripe.Event
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("解析 Stripe Webhook 失败: %w", err)
	}

	// 构造 Webhook 事件
	webhookEvent := &WebhookEvent{
		EventID:   event.ID,
		EventType: convertStripeEventType(string(event.Type)),
		RawData:   event,
	}

	// 根据事件类型解析数据
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("解析 PaymentIntent 数据失败: %w", err)
		}
		webhookEvent.ChannelTradeNo = pi.ID
		webhookEvent.PaymentNo = pi.Metadata["payment_no"]
		webhookEvent.Status = PaymentStatusSuccess
		webhookEvent.Amount = pi.Amount
		webhookEvent.Currency = string(pi.Currency)

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("解析 PaymentIntent 数据失败: %w", err)
		}
		webhookEvent.ChannelTradeNo = pi.ID
		webhookEvent.PaymentNo = pi.Metadata["payment_no"]
		webhookEvent.Status = PaymentStatusFailed
		webhookEvent.Amount = pi.Amount
		webhookEvent.Currency = string(pi.Currency)

	case "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("解析 PaymentIntent 数据失败: %w", err)
		}
		webhookEvent.ChannelTradeNo = pi.ID
		webhookEvent.PaymentNo = pi.Metadata["payment_no"]
		webhookEvent.Status = PaymentStatusCancelled
		webhookEvent.Amount = pi.Amount
		webhookEvent.Currency = string(pi.Currency)

	case "charge.refunded":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			return nil, fmt.Errorf("解析 Charge 数据失败: %w", err)
		}
		webhookEvent.ChannelTradeNo = charge.PaymentIntent.ID
		webhookEvent.Status = PaymentStatusRefunded
		webhookEvent.Amount = charge.AmountRefunded
		webhookEvent.Currency = string(charge.Currency)
	}

	return webhookEvent, nil
}

// convertStripeStatus 转换 Stripe 支付状态为统一状态
func convertStripeStatus(status stripe.PaymentIntentStatus) string {
	switch status {
	case stripe.PaymentIntentStatusRequiresPaymentMethod,
		stripe.PaymentIntentStatusRequiresConfirmation,
		stripe.PaymentIntentStatusRequiresAction:
		return PaymentStatusPending
	case stripe.PaymentIntentStatusProcessing:
		return PaymentStatusProcessing
	case stripe.PaymentIntentStatusSucceeded:
		return PaymentStatusSuccess
	case stripe.PaymentIntentStatusCanceled:
		return PaymentStatusCancelled
	default:
		return PaymentStatusFailed
	}
}

// convertRefundStatus 转换 Stripe 退款状态为统一状态
func convertRefundStatus(status stripe.RefundStatus) string {
	switch status {
	case stripe.RefundStatusPending:
		return PaymentStatusProcessing
	case stripe.RefundStatusSucceeded:
		return PaymentStatusRefunded
	case stripe.RefundStatusFailed:
		return PaymentStatusFailed
	case stripe.RefundStatusCanceled:
		return PaymentStatusCancelled
	default:
		return PaymentStatusFailed
	}
}

// convertStripeEventType 转换 Stripe 事件类型为统一事件类型
func convertStripeEventType(eventType string) string {
	switch eventType {
	case "payment_intent.succeeded":
		return EventTypePaymentSuccess
	case "payment_intent.payment_failed":
		return EventTypePaymentFailed
	case "payment_intent.canceled":
		return EventTypePaymentCancelled
	case "charge.refunded":
		return EventTypeRefundSuccess
	default:
		return eventType
	}
}

// ConvertAmountToStripe 将金额（分）转换为 Stripe 金额
// Stripe 对于零小数位货币（如 JPY、KRW），金额单位是最小单位（无需转换）
// 对于两位小数货币（如 USD、EUR），金额单位是分（cents）
// 对于三位小数货币（如 BHD、KWD），金额单位是 1/1000
func ConvertAmountToStripe(amount int64, currency string) int64 {
	// 零小数位货币列表
	zeroDecimalCurrencies := map[string]bool{
		"BIF": true, "CLP": true, "DJF": true, "GNF": true,
		"JPY": true, "KMF": true, "KRW": true, "MGA": true,
		"PYG": true, "RWF": true, "UGX": true, "VND": true,
		"VUV": true, "XAF": true, "XOF": true, "XPF": true,
	}

	// 如果是零小数位货币，直接返回
	if zeroDecimalCurrencies[currency] {
		return amount / 100 // 从分转换为主单位
	}

	// 其他货币，Stripe 使用最小货币单位（分）
	return amount
}

// ConvertAmountFromStripe 将 Stripe 金额转换为系统金额（分）
func ConvertAmountFromStripe(amount int64, currency string) int64 {
	// 零小数位货币列表
	zeroDecimalCurrencies := map[string]bool{
		"BIF": true, "CLP": true, "DJF": true, "GNF": true,
		"JPY": true, "KMF": true, "KRW": true, "MGA": true,
		"PYG": true, "RWF": true, "UGX": true, "VND": true,
		"VUV": true, "XAF": true, "XOF": true, "XPF": true,
	}

	// 如果是零小数位货币，需要乘以100转换为分
	if zeroDecimalCurrencies[currency] {
		return amount * 100
	}

	// 其他货币直接返回
	return amount
}

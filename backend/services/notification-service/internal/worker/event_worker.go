package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/payment-platform/pkg/events"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"payment-platform/notification-service/internal/provider"
	"payment-platform/notification-service/internal/repository"
)

// EventWorker 事件处理worker (消费业务事件并发送通知)
type EventWorker struct {
	repo         repository.NotificationRepository
	emailFactory *provider.EmailProviderFactory
	smsFactory   *provider.SMSProviderFactory
}

// NewEventWorker 创建事件worker
func NewEventWorker(
	repo repository.NotificationRepository,
	emailFactory *provider.EmailProviderFactory,
	smsFactory *provider.SMSProviderFactory,
) *EventWorker {
	return &EventWorker{
		repo:         repo,
		emailFactory: emailFactory,
		smsFactory:   smsFactory,
	}
}

// StartPaymentEventWorker 启动支付事件消费worker
func (w *EventWorker) StartPaymentEventWorker(ctx context.Context, consumer *kafka.Consumer) {
	logger.Info("支付事件Worker启动，订阅topic: payment.events")

	handler := func(ctx context.Context, message []byte) error {
		// 解析通用事件
		var baseEvent events.BaseEvent
		if err := json.Unmarshal(message, &baseEvent); err != nil {
			logger.Error("反序列化支付事件失败", zap.Error(err))
			return err
		}

		logger.Info("收到支付事件",
			zap.String("event_type", baseEvent.EventType),
			zap.String("event_id", baseEvent.EventID),
			zap.String("aggregate_id", baseEvent.AggregateID))

		// 根据事件类型路由处理
		switch baseEvent.EventType {
		case events.PaymentCreated:
			return w.handlePaymentCreated(ctx, message)
		case events.PaymentSuccess:
			return w.handlePaymentSuccess(ctx, message)
		case events.PaymentFailed:
			return w.handlePaymentFailed(ctx, message)
		case events.RefundSuccess:
			return w.handleRefundSuccess(ctx, message)
		case events.RefundFailed:
			return w.handleRefundFailed(ctx, message)
		default:
			logger.Info("未处理的支付事件类型", zap.String("event_type", baseEvent.EventType))
			return nil
		}
	}

	// 开始消费，支持重试
	if err := consumer.ConsumeWithRetry(ctx, handler, 3); err != nil {
		logger.Error("支付事件Worker停止", zap.Error(err))
	}
}

// StartOrderEventWorker 启动订单事件消费worker
func (w *EventWorker) StartOrderEventWorker(ctx context.Context, consumer *kafka.Consumer) {
	logger.Info("订单事件Worker启动，订阅topic: order.events")

	handler := func(ctx context.Context, message []byte) error {
		// 解析通用事件
		var baseEvent events.BaseEvent
		if err := json.Unmarshal(message, &baseEvent); err != nil {
			logger.Error("反序列化订单事件失败", zap.Error(err))
			return err
		}

		logger.Info("收到订单事件",
			zap.String("event_type", baseEvent.EventType),
			zap.String("event_id", baseEvent.EventID),
			zap.String("aggregate_id", baseEvent.AggregateID))

		// 根据事件类型路由处理
		switch baseEvent.EventType {
		case events.OrderCreated:
			return w.handleOrderCreated(ctx, message)
		case events.OrderPaid:
			return w.handleOrderPaid(ctx, message)
		case events.OrderCancelled:
			return w.handleOrderCancelled(ctx, message)
		case events.OrderShipped:
			return w.handleOrderShipped(ctx, message)
		default:
			logger.Info("未处理的订单事件类型", zap.String("event_type", baseEvent.EventType))
			return nil
		}
	}

	// 开始消费，支持重试
	if err := consumer.ConsumeWithRetry(ctx, handler, 3); err != nil {
		logger.Error("订单事件Worker停止", zap.Error(err))
	}
}

// ========== 支付事件处理器 ==========

// handlePaymentCreated 处理支付创建事件
func (w *EventWorker) handlePaymentCreated(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理支付创建事件",
		zap.String("payment_no", event.Payload.PaymentNo),
		zap.String("merchant_id", event.Payload.MerchantID))

	// 发送支付确认邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      event.Payload.CustomerEmail,
		Subject: "支付已创建 - Payment Created",
		Template: "payment_created",
		Data: map[string]interface{}{
			"payment_no": event.Payload.PaymentNo,
			"amount":     float64(event.Payload.Amount) / 100, // 分转元
			"currency":   event.Payload.Currency,
			"order_no":   event.Payload.OrderNo,
		},
	})
}

// handlePaymentSuccess 处理支付成功事件
func (w *EventWorker) handlePaymentSuccess(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理支付成功事件",
		zap.String("payment_no", event.Payload.PaymentNo),
		zap.String("order_no", event.Payload.OrderNo))

	// 发送支付成功邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      event.Payload.CustomerEmail,
		Subject: "支付成功 - Payment Successful",
		Template: "payment_success",
		Data: map[string]interface{}{
			"payment_no": event.Payload.PaymentNo,
			"order_no":   event.Payload.OrderNo,
			"amount":     float64(event.Payload.Amount) / 100,
			"currency":   event.Payload.Currency,
			"paid_at":    event.Payload.PaidAt,
		},
	})
}

// handlePaymentFailed 处理支付失败事件
func (w *EventWorker) handlePaymentFailed(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理支付失败事件",
		zap.String("payment_no", event.Payload.PaymentNo))

	// 发送支付失败邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      event.Payload.CustomerEmail,
		Subject: "支付失败 - Payment Failed",
		Template: "payment_failed",
		Data: map[string]interface{}{
			"payment_no": event.Payload.PaymentNo,
			"order_no":   event.Payload.OrderNo,
			"amount":     float64(event.Payload.Amount) / 100,
			"currency":   event.Payload.Currency,
		},
	})
}

// handleRefundSuccess 处理退款成功事件
func (w *EventWorker) handleRefundSuccess(ctx context.Context, message []byte) error {
	var event events.RefundEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理退款成功事件",
		zap.String("refund_no", event.Payload.RefundNo),
		zap.String("payment_no", event.Payload.PaymentNo))

	// 从Extra中获取customer_email (如果有)
	customerEmail, ok := event.Payload.Extra["customer_email"].(string)
	if !ok || customerEmail == "" {
		logger.Warn("退款成功事件缺少customer_email，跳过邮件发送")
		return nil
	}

	// 发送退款成功邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      customerEmail,
		Subject: "退款成功 - Refund Successful",
		Template: "refund_success",
		Data: map[string]interface{}{
			"refund_no":  event.Payload.RefundNo,
			"payment_no": event.Payload.PaymentNo,
			"amount":     float64(event.Payload.Amount) / 100,
			"currency":   event.Payload.Currency,
		},
	})
}

// handleRefundFailed 处理退款失败事件
func (w *EventWorker) handleRefundFailed(ctx context.Context, message []byte) error {
	var event events.RefundEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理退款失败事件",
		zap.String("refund_no", event.Payload.RefundNo))

	// 发送退款失败邮件给商户 (从Extra获取merchant_email)
	merchantEmail, ok := event.Payload.Extra["merchant_email"].(string)
	if !ok || merchantEmail == "" {
		logger.Warn("退款失败事件缺少merchant_email，跳过邮件发送")
		return nil
	}

	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      merchantEmail,
		Subject: "退款失败 - Refund Failed",
		Template: "refund_failed",
		Data: map[string]interface{}{
			"refund_no":  event.Payload.RefundNo,
			"payment_no": event.Payload.PaymentNo,
			"amount":     float64(event.Payload.Amount) / 100,
			"currency":   event.Payload.Currency,
		},
	})
}

// ========== 订单事件处理器 ==========

// handleOrderCreated 处理订单创建事件
func (w *EventWorker) handleOrderCreated(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理订单创建事件",
		zap.String("order_no", event.Payload.OrderNo))

	// 发送订单确认邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      event.Payload.CustomerEmail,
		Subject: "订单已创建 - Order Created",
		Template: "order_created",
		Data: map[string]interface{}{
			"order_no":     event.Payload.OrderNo,
			"total_amount": float64(event.Payload.TotalAmount) / 100,
			"currency":     event.Payload.Currency,
		},
	})
}

// handleOrderPaid 处理订单支付成功事件
func (w *EventWorker) handleOrderPaid(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理订单支付成功事件",
		zap.String("order_no", event.Payload.OrderNo),
		zap.String("payment_no", event.Payload.PaymentNo))

	// 发送订单支付成功邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      event.Payload.CustomerEmail,
		Subject: "订单支付成功 - Order Paid",
		Template: "order_paid",
		Data: map[string]interface{}{
			"order_no":     event.Payload.OrderNo,
			"payment_no":   event.Payload.PaymentNo,
			"total_amount": float64(event.Payload.TotalAmount) / 100,
			"currency":     event.Payload.Currency,
			"paid_at":      event.Payload.PaidAt,
		},
	})
}

// handleOrderCancelled 处理订单取消事件
func (w *EventWorker) handleOrderCancelled(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理订单取消事件",
		zap.String("order_no", event.Payload.OrderNo))

	// 发送订单取消邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      event.Payload.CustomerEmail,
		Subject: "订单已取消 - Order Cancelled",
		Template: "order_cancelled",
		Data: map[string]interface{}{
			"order_no": event.Payload.OrderNo,
		},
	})
}

// handleOrderShipped 处理订单发货事件
func (w *EventWorker) handleOrderShipped(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Info("处理订单发货事件",
		zap.String("order_no", event.Payload.OrderNo))

	// 发送订单发货邮件给客户
	return w.sendEmailNotification(ctx, &EmailNotificationRequest{
		To:      event.Payload.CustomerEmail,
		Subject: "订单已发货 - Order Shipped",
		Template: "order_shipped",
		Data: map[string]interface{}{
			"order_no": event.Payload.OrderNo,
		},
	})
}

// ========== 辅助方法 ==========

// EmailNotificationRequest 邮件通知请求
type EmailNotificationRequest struct {
	To       string
	Subject  string
	Template string
	Data     map[string]interface{}
}

// sendEmailNotification 发送邮件通知 (简化版)
func (w *EventWorker) sendEmailNotification(ctx context.Context, req *EmailNotificationRequest) error {
	// 获取邮件提供商 (使用默认的smtp或mailgun)
	emailProvider, ok := w.emailFactory.GetProvider("smtp")
	if !ok {
		// 尝试mailgun
		emailProvider, ok = w.emailFactory.GetProvider("mailgun")
		if !ok {
			logger.Warn("未配置邮件提供商，跳过发送")
			return nil
		}
	}

	// 构造简单的HTML内容 (实际应该使用模板引擎)
	htmlBody := w.renderSimpleTemplate(req.Template, req.Data)

	// 发送邮件
	emailReq := &provider.EmailRequest{
		To:       []string{req.To},
		Subject:  req.Subject,
		HTMLBody: htmlBody,
	}

	resp, err := emailProvider.Send(ctx, emailReq)
	if err != nil {
		logger.Error("发送邮件失败",
			zap.Error(err),
			zap.String("to", req.To),
			zap.String("subject", req.Subject))
		return err
	}

	logger.Info("邮件发送成功",
		zap.String("to", req.To),
		zap.String("message_id", resp.MessageID))

	return nil
}

// renderSimpleTemplate 简单模板渲染 (实际应该使用html/template)
func (w *EventWorker) renderSimpleTemplate(template string, data map[string]interface{}) string {
	switch template {
	case "payment_created":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>支付已创建</h2>
				<p>您的支付订单已创建成功：</p>
				<ul>
					<li>支付流水号: %v</li>
					<li>订单号: %v</li>
					<li>金额: %v %v</li>
				</ul>
				<p>请尽快完成支付。</p>
			</body>
			</html>
		`, data["payment_no"], data["order_no"], data["amount"], data["currency"])

	case "payment_success":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>支付成功</h2>
				<p>您的支付已成功完成：</p>
				<ul>
					<li>支付流水号: %v</li>
					<li>订单号: %v</li>
					<li>金额: %v %v</li>
					<li>支付时间: %v</li>
				</ul>
				<p>感谢您的购买！</p>
			</body>
			</html>
		`, data["payment_no"], data["order_no"], data["amount"], data["currency"], data["paid_at"])

	case "payment_failed":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>支付失败</h2>
				<p>很抱歉，您的支付未能成功：</p>
				<ul>
					<li>支付流水号: %v</li>
					<li>订单号: %v</li>
					<li>金额: %v %v</li>
				</ul>
				<p>请重新尝试支付或联系客服。</p>
			</body>
			</html>
		`, data["payment_no"], data["order_no"], data["amount"], data["currency"])

	case "order_created":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>订单已创建</h2>
				<p>您的订单已成功创建：</p>
				<ul>
					<li>订单号: %v</li>
					<li>总金额: %v %v</li>
				</ul>
				<p>请尽快完成支付。</p>
			</body>
			</html>
		`, data["order_no"], data["total_amount"], data["currency"])

	case "order_paid":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>订单支付成功</h2>
				<p>您的订单已成功支付：</p>
				<ul>
					<li>订单号: %v</li>
					<li>支付流水号: %v</li>
					<li>总金额: %v %v</li>
					<li>支付时间: %v</li>
				</ul>
				<p>我们将尽快为您处理订单。</p>
			</body>
			</html>
		`, data["order_no"], data["payment_no"], data["total_amount"], data["currency"], data["paid_at"])

	case "order_cancelled":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>订单已取消</h2>
				<p>您的订单已被取消：</p>
				<ul>
					<li>订单号: %v</li>
				</ul>
			</body>
			</html>
		`, data["order_no"])

	case "order_shipped":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>订单已发货</h2>
				<p>您的订单已发货：</p>
				<ul>
					<li>订单号: %v</li>
				</ul>
				<p>请注意查收。</p>
			</body>
			</html>
		`, data["order_no"])

	case "refund_success":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>退款成功</h2>
				<p>您的退款已成功处理：</p>
				<ul>
					<li>退款流水号: %v</li>
					<li>支付流水号: %v</li>
					<li>退款金额: %v %v</li>
				</ul>
				<p>退款将在1-3个工作日内到账。</p>
			</body>
			</html>
		`, data["refund_no"], data["payment_no"], data["amount"], data["currency"])

	case "refund_failed":
		return fmt.Sprintf(`
			<html>
			<body>
				<h2>退款失败</h2>
				<p>退款处理失败：</p>
				<ul>
					<li>退款流水号: %v</li>
					<li>支付流水号: %v</li>
					<li>退款金额: %v %v</li>
				</ul>
				<p>请联系客服处理。</p>
			</body>
			</html>
		`, data["refund_no"], data["payment_no"], data["amount"], data["currency"])

	default:
		return "<html><body><p>通知内容</p></body></html>"
	}
}

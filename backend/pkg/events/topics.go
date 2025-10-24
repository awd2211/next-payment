package events

// Kafka Topic常量定义
// 命名规范: <domain>.<entity>.events

const (
	// 支付相关Topics
	TopicPaymentEvents = "payment.events"        // 支付事件 (created/success/failed/cancelled)
	TopicRefundEvents  = "payment.refund.events" // 退款事件 (created/success/failed)

	// 订单相关Topics
	TopicOrderEvents = "order.events" // 订单事件 (created/paid/cancelled/completed)

	// 财务相关Topics
	TopicAccountingEvents = "accounting.events" // 财务事件 (transaction.created/balance.updated)

	// 结算相关Topics
	TopicSettlementEvents = "settlement.events" // 结算事件 (created/approved/rejected/completed)

	// 提现相关Topics
	TopicWithdrawalEvents = "withdrawal.events" // 提现事件 (created/approved/success/failed)

	// 商户相关Topics
	TopicMerchantEvents = "merchant.events" // 商户事件 (created/approved/frozen/updated)

	// KYC相关Topics
	TopicKYCEvents = "kyc.events" // KYC事件 (submitted/approved/rejected)

	// 内部通知事件 (临时兼容,未来移除)
	TopicInternalNotification = "payment.internal.notification"

	// 已存在的Topics (notification-service使用)
	TopicEmailNotifications = "email-notifications" // 邮件通知
	TopicSMSNotifications   = "sms-notifications"   // 短信通知
)

// GetTopicByEventType 根据事件类型返回Topic
func GetTopicByEventType(eventType string) string {
	switch {
	case eventType == PaymentCreated || eventType == PaymentSuccess || eventType == PaymentFailed || eventType == PaymentCancelled:
		return TopicPaymentEvents
	case eventType == RefundCreated || eventType == RefundSuccess || eventType == RefundFailed:
		return TopicRefundEvents
	case eventType == OrderCreated || eventType == OrderPaid || eventType == OrderCancelled || eventType == OrderCompleted:
		return TopicOrderEvents
	case eventType == TransactionCreated || eventType == BalanceUpdated || eventType == SettlementCalculated:
		return TopicAccountingEvents
	default:
		return "" // 未知事件类型
	}
}

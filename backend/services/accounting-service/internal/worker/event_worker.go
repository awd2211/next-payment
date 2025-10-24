package worker

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/events"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"payment-platform/accounting-service/internal/model"
	"payment-platform/accounting-service/internal/service"
)

// EventWorker 财务服务事件处理worker (消费支付/退款事件自动记账)
type EventWorker struct {
	accountService   service.AccountService
	eventPublisher   *kafka.EventPublisher
}

// NewEventWorker 创建事件worker
func NewEventWorker(
	accountService service.AccountService,
	eventPublisher *kafka.EventPublisher,
) *EventWorker {
	return &EventWorker{
		accountService: accountService,
		eventPublisher: eventPublisher,
	}
}

// StartPaymentEventWorker 启动支付事件消费worker
func (w *EventWorker) StartPaymentEventWorker(ctx context.Context, consumer *kafka.Consumer) {
	logger.Info("Accounting: 支付事件Worker启动，订阅topic: payment.events")

	handler := func(ctx context.Context, message []byte) error {
		// 解析通用事件
		var baseEvent events.BaseEvent
		if err := json.Unmarshal(message, &baseEvent); err != nil {
			logger.Error("Accounting: 反序列化支付事件失败", zap.Error(err))
			return err
		}

		logger.Info("Accounting: 收到支付事件",
			zap.String("event_type", baseEvent.EventType),
			zap.String("event_id", baseEvent.EventID),
			zap.String("aggregate_id", baseEvent.AggregateID))

		// 根据事件类型路由处理
		switch baseEvent.EventType {
		case events.PaymentSuccess:
			return w.handlePaymentSuccess(ctx, message)
		case events.RefundSuccess:
			return w.handleRefundSuccess(ctx, message)
		default:
			logger.Info("Accounting: 未处理的支付事件类型", zap.String("event_type", baseEvent.EventType))
			return nil
		}
	}

	// 开始消费，支持重试
	if err := consumer.ConsumeWithRetry(ctx, handler, 3); err != nil {
		logger.Error("Accounting: 支付事件Worker停止", zap.Error(err))
	}
}

// ========== 支付事件处理器 (Consumer) ==========

// handlePaymentSuccess 处理支付成功事件 → 自动记账
func (w *EventWorker) handlePaymentSuccess(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Accounting: 解析merchant_id失败", zap.Error(err))
		return err
	}

	logger.Info("Accounting: 处理支付成功事件，自动记账",
		zap.String("payment_no", event.Payload.PaymentNo),
		zap.String("merchant_id", event.Payload.MerchantID),
		zap.Int64("amount", event.Payload.Amount),
		zap.String("currency", event.Payload.Currency))

	// 1. 获取或创建商户账户 (自动创建待结算账户)
	account, err := w.accountService.GetMerchantAccount(ctx, merchantID, "settlement", event.Payload.Currency)
	if err != nil {
		// 如果账户不存在，创建新账户
		createAccountInput := &service.CreateAccountInput{
			MerchantID:  merchantID,
			AccountType: "settlement", // 待结算账户
			Currency:    event.Payload.Currency,
		}
		account, err = w.accountService.CreateAccount(ctx, createAccountInput)
		if err != nil {
			logger.Error("Accounting: 创建商户账户失败", zap.Error(err))
			return err
		}
		logger.Info("Accounting: 自动创建商户账户",
			zap.String("merchant_id", merchantID.String()),
			zap.String("account_type", "settlement"),
			zap.String("currency", event.Payload.Currency))
	}

	// 2. 创建财务交易 (复式记账)
	// 借: 商户待结算账户
	// 贷: 平台收入账户
	input := &service.CreateTransactionInput{
		AccountID:       account.ID, // 使用实际账户ID
		TransactionType: "payment",
		Amount:          event.Payload.Amount,
		RelatedNo:       event.Payload.PaymentNo, // 关联支付单号
		Description:     "支付入账: " + event.Payload.PaymentNo,
		Extra: map[string]interface{}{
			"payment_no": event.Payload.PaymentNo,
			"order_no":   event.Payload.OrderNo,
			"channel":    event.Payload.Channel,
			"merchant_id": event.Payload.MerchantID,
		},
	}

	transaction, err := w.accountService.CreateTransaction(ctx, input)
	if err != nil {
		logger.Error("Accounting: 创建财务交易失败",
			zap.Error(err),
			zap.String("payment_no", event.Payload.PaymentNo))
		return err
	}

	logger.Info("Accounting: 财务交易创建成功",
		zap.String("transaction_no", transaction.TransactionNo),
		zap.String("payment_no", event.Payload.PaymentNo),
		zap.Int64("balance_after", transaction.BalanceAfter))

	// 发布财务事件 (Producer)
	return w.publishAccountingEvent(ctx, events.TransactionCreated, transaction)
}

// handleRefundSuccess 处理退款成功事件 → 自动记账
func (w *EventWorker) handleRefundSuccess(ctx context.Context, message []byte) error {
	var event events.RefundEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Accounting: 解析merchant_id失败", zap.Error(err))
		return err
	}

	logger.Info("Accounting: 处理退款成功事件，自动记账",
		zap.String("refund_no", event.Payload.RefundNo),
		zap.String("payment_no", event.Payload.PaymentNo),
		zap.Int64("amount", event.Payload.Amount),
		zap.String("currency", event.Payload.Currency))

	// 1. 获取或创建商户账户 (自动创建待结算账户)
	account, err := w.accountService.GetMerchantAccount(ctx, merchantID, "settlement", event.Payload.Currency)
	if err != nil {
		// 如果账户不存在，创建新账户
		createAccountInput := &service.CreateAccountInput{
			MerchantID:  merchantID,
			AccountType: "settlement",
			Currency:    event.Payload.Currency,
		}
		account, err = w.accountService.CreateAccount(ctx, createAccountInput)
		if err != nil {
			logger.Error("Accounting: 创建商户账户失败", zap.Error(err))
			return err
		}
		logger.Info("Accounting: 自动创建商户账户 (退款)",
			zap.String("merchant_id", merchantID.String()),
			zap.String("account_type", "settlement"),
			zap.String("currency", event.Payload.Currency))
	}

	// 2. 创建退款财务交易 (反向记账)
	// 借: 平台收入账户
	// 贷: 商户待结算账户
	input := &service.CreateTransactionInput{
		AccountID:       account.ID, // 使用实际账户ID
		TransactionType: "refund",
		Amount:          -event.Payload.Amount, // 负数表示退款
		RelatedNo:       event.Payload.RefundNo, // 关联退款单号
		Description:     "退款出账: " + event.Payload.RefundNo,
		Extra: map[string]interface{}{
			"refund_no":  event.Payload.RefundNo,
			"payment_no": event.Payload.PaymentNo,
			"reason":     event.Payload.Reason,
			"merchant_id": event.Payload.MerchantID,
		},
	}

	transaction, err := w.accountService.CreateTransaction(ctx, input)
	if err != nil {
		logger.Error("Accounting: 创建退款交易失败",
			zap.Error(err),
			zap.String("refund_no", event.Payload.RefundNo))
		return err
	}

	logger.Info("Accounting: 退款交易创建成功",
		zap.String("transaction_no", transaction.TransactionNo),
		zap.String("refund_no", event.Payload.RefundNo),
		zap.Int64("balance_after", transaction.BalanceAfter))

	// 发布财务事件 (Producer)
	return w.publishAccountingEvent(ctx, events.TransactionCreated, transaction)
}

// ========== 事件发布器 (Producer) ==========

// publishAccountingEvent 发布财务事件到Kafka
func (w *EventWorker) publishAccountingEvent(
	ctx context.Context,
	eventType string,
	transaction *model.AccountTransaction,
) error {
	if w.eventPublisher == nil {
		logger.Warn("Accounting: eventPublisher is nil, skipping event publishing")
		return nil
	}

	// 构造财务事件载荷
	payload := events.AccountingEventPayload{
		TransactionID: transaction.ID.String(),
		AccountID:     transaction.AccountID.String(),
		MerchantID:    transaction.MerchantID.String(),
		Type:          w.getTransactionDirection(transaction.Amount), // credit or debit
		Amount:        transaction.Amount,
		Balance:       transaction.BalanceAfter,
		Currency:      transaction.Currency,
		Description:   transaction.Description,
		RelatedID:     transaction.RelatedNo, // payment_no or refund_no
		CreatedAt:     transaction.CreatedAt,
		Extra: map[string]interface{}{
			"transaction_type": transaction.TransactionType,
			"related_id":       transaction.RelatedID.String(),
			"balance_before":   transaction.BalanceBefore,
			"balance_after":    transaction.BalanceAfter,
		},
	}

	// 创建事件
	event := events.NewAccountingEvent(eventType, payload)

	// 异步发布事件 (不阻塞主流程)
	w.eventPublisher.PublishAsync(ctx, "accounting.events", event)

	logger.Info("Accounting: 财务事件已发布",
		zap.String("event_type", eventType),
		zap.String("transaction_no", transaction.TransactionNo))

	return nil
}

// getTransactionDirection 根据金额判断交易方向 (借或贷)
func (w *EventWorker) getTransactionDirection(amount int64) string {
	if amount > 0 {
		return "credit" // 贷记 (入账)
	}
	return "debit" // 借记 (出账)
}

// ========== 辅助方法 ==========
// (TransactionNo由AccountService.CreateTransaction自动生成)

package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/events"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/analytics-service/internal/model"
	"payment-platform/analytics-service/internal/repository"
)

// EventWorker 分析服务事件处理worker (消费所有业务事件进行统计分析)
type EventWorker struct {
	db   *gorm.DB
	repo repository.AnalyticsRepository
}

// NewEventWorker 创建事件worker
func NewEventWorker(db *gorm.DB, repo repository.AnalyticsRepository) *EventWorker {
	return &EventWorker{
		db:   db,
		repo: repo,
	}
}

// StartPaymentEventWorker 启动支付事件消费worker
func (w *EventWorker) StartPaymentEventWorker(ctx context.Context, consumer *kafka.Consumer) {
	logger.Info("Analytics: 支付事件Worker启动，订阅topic: payment.events")

	handler := func(ctx context.Context, message []byte) error {
		// 解析通用事件
		var baseEvent events.BaseEvent
		if err := json.Unmarshal(message, &baseEvent); err != nil {
			logger.Error("Analytics: 反序列化支付事件失败", zap.Error(err))
			return err
		}

		logger.Info("Analytics: 收到支付事件",
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
		case events.PaymentCancelled:
			return w.handlePaymentCancelled(ctx, message)
		case events.RefundSuccess:
			return w.handleRefundSuccess(ctx, message)
		default:
			logger.Info("Analytics: 未处理的支付事件类型", zap.String("event_type", baseEvent.EventType))
			return nil
		}
	}

	// 开始消费，支持重试
	if err := consumer.ConsumeWithRetry(ctx, handler, 3); err != nil {
		logger.Error("Analytics: 支付事件Worker停止", zap.Error(err))
	}
}

// StartOrderEventWorker 启动订单事件消费worker
func (w *EventWorker) StartOrderEventWorker(ctx context.Context, consumer *kafka.Consumer) {
	logger.Info("Analytics: 订单事件Worker启动，订阅topic: order.events")

	handler := func(ctx context.Context, message []byte) error {
		// 解析通用事件
		var baseEvent events.BaseEvent
		if err := json.Unmarshal(message, &baseEvent); err != nil {
			logger.Error("Analytics: 反序列化订单事件失败", zap.Error(err))
			return err
		}

		logger.Info("Analytics: 收到订单事件",
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
		case events.OrderCompleted:
			return w.handleOrderCompleted(ctx, message)
		default:
			logger.Info("Analytics: 未处理的订单事件类型", zap.String("event_type", baseEvent.EventType))
			return nil
		}
	}

	// 开始消费，支持重试
	if err := consumer.ConsumeWithRetry(ctx, handler, 3); err != nil {
		logger.Error("Analytics: 订单事件Worker停止", zap.Error(err))
	}
}

// ========== 支付事件处理器 ==========

// handlePaymentCreated 处理支付创建事件
func (w *EventWorker) handlePaymentCreated(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Analytics: 解析merchant_id失败", zap.Error(err))
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)

	// 更新支付指标 (增加总支付数)
	return w.updatePaymentMetrics(ctx, merchantID, date, event.Payload.Currency, func(metrics *model.PaymentMetrics) {
		metrics.TotalPayments++
	})
}

// handlePaymentSuccess 处理支付成功事件
func (w *EventWorker) handlePaymentSuccess(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Analytics: 解析merchant_id失败", zap.Error(err))
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)

	// 更新支付指标
	err = w.updatePaymentMetrics(ctx, merchantID, date, event.Payload.Currency, func(metrics *model.PaymentMetrics) {
		metrics.SuccessPayments++
		metrics.SuccessAmount += event.Payload.Amount
		metrics.TotalAmount += event.Payload.Amount

		// 计算成功率
		if metrics.TotalPayments > 0 {
			metrics.SuccessRate = float64(metrics.SuccessPayments) / float64(metrics.TotalPayments) * 100
		}

		// 计算平均金额
		if metrics.SuccessPayments > 0 {
			metrics.AverageAmount = metrics.SuccessAmount / int64(metrics.SuccessPayments)
		}
	})

	if err != nil {
		return err
	}

	// 更新渠道指标
	return w.updateChannelMetrics(ctx, event.Payload.Channel, date, event.Payload.Currency, func(metrics *model.ChannelMetrics) {
		metrics.SuccessTransactions++
		metrics.SuccessAmount += event.Payload.Amount
		metrics.TotalAmount += event.Payload.Amount

		// 计算成功率
		if metrics.TotalTransactions > 0 {
			metrics.SuccessRate = float64(metrics.SuccessTransactions) / float64(metrics.TotalTransactions) * 100
		}
	})
}

// handlePaymentFailed 处理支付失败事件
func (w *EventWorker) handlePaymentFailed(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Analytics: 解析merchant_id失败", zap.Error(err))
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)

	// 更新支付指标
	err = w.updatePaymentMetrics(ctx, merchantID, date, event.Payload.Currency, func(metrics *model.PaymentMetrics) {
		metrics.FailedPayments++

		// 重新计算成功率
		if metrics.TotalPayments > 0 {
			metrics.SuccessRate = float64(metrics.SuccessPayments) / float64(metrics.TotalPayments) * 100
		}
	})

	if err != nil {
		return err
	}

	// 更新渠道指标
	return w.updateChannelMetrics(ctx, event.Payload.Channel, date, event.Payload.Currency, func(metrics *model.ChannelMetrics) {
		metrics.FailedTransactions++

		// 重新计算成功率
		if metrics.TotalTransactions > 0 {
			metrics.SuccessRate = float64(metrics.SuccessTransactions) / float64(metrics.TotalTransactions) * 100
		}
	})
}

// handlePaymentCancelled 处理支付取消事件
func (w *EventWorker) handlePaymentCancelled(ctx context.Context, message []byte) error {
	var event events.PaymentEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	// 支付取消暂不更新统计指标
	logger.Info("Analytics: 支付取消事件已记录", zap.String("payment_no", event.Payload.PaymentNo))
	return nil
}

// handleRefundSuccess 处理退款成功事件
func (w *EventWorker) handleRefundSuccess(ctx context.Context, message []byte) error {
	var event events.RefundEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Analytics: 解析merchant_id失败", zap.Error(err))
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)

	// 更新支付指标 (增加退款数和退款金额)
	return w.updatePaymentMetrics(ctx, merchantID, date, event.Payload.Currency, func(metrics *model.PaymentMetrics) {
		metrics.TotalRefunds++
		metrics.TotalRefundAmount += event.Payload.Amount
	})
}

// ========== 订单事件处理器 ==========

// handleOrderCreated 处理订单创建事件
func (w *EventWorker) handleOrderCreated(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Analytics: 解析merchant_id失败", zap.Error(err))
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)

	// 更新商户指标
	return w.updateMerchantMetrics(ctx, merchantID, date, event.Payload.Currency, func(metrics *model.MerchantMetrics) {
		metrics.TotalOrders++
	})
}

// handleOrderPaid 处理订单支付成功事件
func (w *EventWorker) handleOrderPaid(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Analytics: 解析merchant_id失败", zap.Error(err))
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)

	// 从Extra中获取pay_amount
	payAmount := event.Payload.TotalAmount
	if extra, ok := event.Payload.Extra["pay_amount"].(float64); ok {
		payAmount = int64(extra)
	}

	// 更新商户指标
	return w.updateMerchantMetrics(ctx, merchantID, date, event.Payload.Currency, func(metrics *model.MerchantMetrics) {
		metrics.CompletedOrders++
		metrics.TotalRevenue += payAmount

		// 假设费率2%
		fee := payAmount * 2 / 100
		metrics.TotalFees += fee
		metrics.NetRevenue = metrics.TotalRevenue - metrics.TotalFees
	})
}

// handleOrderCancelled 处理订单取消事件
func (w *EventWorker) handleOrderCancelled(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	merchantID, err := uuid.Parse(event.Payload.MerchantID)
	if err != nil {
		logger.Error("Analytics: 解析merchant_id失败", zap.Error(err))
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)

	// 更新商户指标
	return w.updateMerchantMetrics(ctx, merchantID, date, event.Payload.Currency, func(metrics *model.MerchantMetrics) {
		metrics.CancelledOrders++
	})
}

// handleOrderCompleted 处理订单完成事件
func (w *EventWorker) handleOrderCompleted(ctx context.Context, message []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	// 订单完成事件暂不更新统计指标
	logger.Info("Analytics: 订单完成事件已记录", zap.String("order_no", event.Payload.OrderNo))
	return nil
}

// ========== 辅助方法 ==========

// updatePaymentMetrics 更新支付指标 (使用UPSERT模式)
func (w *EventWorker) updatePaymentMetrics(
	ctx context.Context,
	merchantID uuid.UUID,
	date time.Time,
	currency string,
	updateFn func(*model.PaymentMetrics),
) error {
	// 使用事务保证原子性
	return w.db.Transaction(func(tx *gorm.DB) error {
		var metrics model.PaymentMetrics

		// 尝试查找现有记录
		err := tx.Where("merchant_id = ? AND date = ? AND currency = ?", merchantID, date, currency).
			First(&metrics).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			metrics = model.PaymentMetrics{
				MerchantID: merchantID,
				Date:       date,
				Currency:   currency,
			}
		} else if err != nil {
			return err
		}

		// 执行更新函数
		updateFn(&metrics)

		// 保存或更新
		return tx.Save(&metrics).Error
	})
}

// updateMerchantMetrics 更新商户指标 (使用UPSERT模式)
func (w *EventWorker) updateMerchantMetrics(
	ctx context.Context,
	merchantID uuid.UUID,
	date time.Time,
	currency string,
	updateFn func(*model.MerchantMetrics),
) error {
	// 使用事务保证原子性
	return w.db.Transaction(func(tx *gorm.DB) error {
		var metrics model.MerchantMetrics

		// 尝试查找现有记录
		err := tx.Where("merchant_id = ? AND date = ?", merchantID, date).
			First(&metrics).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			metrics = model.MerchantMetrics{
				MerchantID: merchantID,
				Date:       date,
				Currency:   currency,
			}
		} else if err != nil {
			return err
		}

		// 执行更新函数
		updateFn(&metrics)

		// 保存或更新
		return tx.Save(&metrics).Error
	})
}

// updateChannelMetrics 更新渠道指标 (使用UPSERT模式)
func (w *EventWorker) updateChannelMetrics(
	ctx context.Context,
	channelCode string,
	date time.Time,
	currency string,
	updateFn func(*model.ChannelMetrics),
) error {
	// 使用事务保证原子性
	return w.db.Transaction(func(tx *gorm.DB) error {
		var metrics model.ChannelMetrics

		// 尝试查找现有记录
		err := tx.Where("channel_code = ? AND date = ? AND currency = ?", channelCode, date, currency).
			First(&metrics).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			metrics = model.ChannelMetrics{
				ChannelCode: channelCode,
				Date:        date,
				Currency:    currency,
			}
		} else if err != nil {
			return err
		}

		// 在首次创建时也要增加TotalTransactions
		if err == gorm.ErrRecordNotFound {
			metrics.TotalTransactions++
		}

		// 执行更新函数
		updateFn(&metrics)

		// 保存或更新
		return tx.Save(&metrics).Error
	})
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/saga"
	"go.uber.org/zap"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
)

// KafkaProducer Kafka 消息发布接口
type KafkaProducer interface {
	Publish(ctx context.Context, topic string, message string) error
}

// CallbackSagaService 支付回调 Saga 服务（用于协调分布式事务）
type CallbackSagaService struct {
	orchestrator  *saga.SagaOrchestrator
	paymentRepo   repository.PaymentRepository
	orderClient   *client.OrderClient
	kafkaProducer KafkaProducer // Kafka 事件发布接口
}

// NewCallbackSagaService 创建支付回调 Saga 服务
func NewCallbackSagaService(
	orchestrator *saga.SagaOrchestrator,
	paymentRepo repository.PaymentRepository,
	orderClient *client.OrderClient,
	kafkaProducer KafkaProducer,
) *CallbackSagaService {
	return &CallbackSagaService{
		orchestrator:  orchestrator,
		paymentRepo:   paymentRepo,
		orderClient:   orderClient,
		kafkaProducer: kafkaProducer,
	}
}

// CallbackData 回调数据
type CallbackData struct {
	PaymentNo      string
	ChannelOrderNo string
	Status         string
	PaidAt         *time.Time
	FailureReason  string
	RawData        string
}

// ExecuteCallbackSaga 执行支付回调 Saga
func (s *CallbackSagaService) ExecuteCallbackSaga(
	ctx context.Context,
	payment *model.Payment,
	callbackData *CallbackData,
) error {
	// 1. 构建 Saga
	sagaBuilder := s.orchestrator.NewSagaBuilder(payment.PaymentNo, "payment_callback")
	sagaBuilder.SetMetadata(map[string]interface{}{
		"payment_no":       payment.PaymentNo,
		"merchant_id":      payment.MerchantID.String(),
		"channel_order_no": callbackData.ChannelOrderNo,
		"status":           callbackData.Status,
		"amount":           payment.Amount,
		"currency":         payment.Currency,
	})

	// 2. 定义步骤
	stepDefs := []saga.StepDefinition{
		{
			Name: "RecordCallback",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeRecordCallback(ctx, payment, callbackData)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateRecordCallback(ctx, payment)
			},
			MaxRetryCount: 3,
			Timeout:       10 * time.Second,
		},
		{
			Name: "UpdatePaymentStatus",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeUpdatePaymentStatus(ctx, payment, callbackData)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateUpdatePaymentStatus(ctx, payment)
			},
			MaxRetryCount: 3,
			Timeout:       10 * time.Second,
		},
		{
			Name: "UpdateOrderStatus",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeUpdateOrderStatus(ctx, payment, callbackData)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateUpdateOrderStatus(ctx, payment)
			},
			MaxRetryCount: 3,
			Timeout:       30 * time.Second,
		},
		{
			Name: "PublishEvent",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executePublishEvent(ctx, payment, callbackData)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensatePublishEvent(ctx, payment)
			},
			MaxRetryCount: 3,
			Timeout:       10 * time.Second,
		},
	}

	// 添加步骤到构建器
	for _, def := range stepDefs {
		sagaBuilder.AddStepWithTimeout(def.Name, def.Execute, def.Compensate, def.MaxRetryCount, def.Timeout)
	}

	// 3. 构建并执行 Saga
	sagaInstance, err := sagaBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build saga: %w", err)
	}

	logger.Info("payment callback saga created",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("payment_no", payment.PaymentNo))

	// 4. 执行 Saga
	if err := s.orchestrator.Execute(ctx, sagaInstance, stepDefs); err != nil {
		logger.Error("payment callback saga execution failed",
			zap.String("saga_id", sagaInstance.ID.String()),
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		return err
	}

	logger.Info("payment callback saga completed",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("payment_no", payment.PaymentNo))

	return nil
}

// executeRecordCallback 执行记录回调步骤
func (s *CallbackSagaService) executeRecordCallback(ctx context.Context, payment *model.Payment, callbackData *CallbackData) (string, error) {
	logger.Info("executing record callback step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("channel_order_no", callbackData.ChannelOrderNo))

	// 创建回调记录
	callback := &model.PaymentCallback{
		PaymentID:   payment.ID,
		Channel:     payment.Channel,
		Event:       "payment." + callbackData.Status,
		RawData:     callbackData.RawData,
		IsVerified:  true, // 假设已验证签名
		IsProcessed: false,
	}

	// 保存回调记录
	if err := s.paymentRepo.CreateCallback(ctx, callback); err != nil {
		return "", fmt.Errorf("save callback record failed: %w", err)
	}

	// 返回回调记录ID
	result := map[string]interface{}{
		"callback_id":  callback.ID,
		"created_at":   callback.CreatedAt.Format(time.RFC3339),
		"status":       callbackData.Status,
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateRecordCallback 补偿记录回调步骤
func (s *CallbackSagaService) compensateRecordCallback(ctx context.Context, payment *model.Payment) error {
	logger.Info("compensating record callback step",
		zap.String("payment_no", payment.PaymentNo))

	// 删除回调记录（如果需要）
	// 注意：通常回调记录用于审计，不应删除，只需标记为已补偿
	if err := s.paymentRepo.MarkCallbackCompensated(ctx, payment.PaymentNo); err != nil {
		logger.Error("failed to mark callback as compensated",
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		// 不返回错误，允许补偿继续
	}

	return nil
}

// executeUpdatePaymentStatus 执行更新支付状态步骤
func (s *CallbackSagaService) executeUpdatePaymentStatus(ctx context.Context, payment *model.Payment, callbackData *CallbackData) (string, error) {
	logger.Info("executing update payment status step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("new_status", callbackData.Status))

	// 保存原始状态
	originalStatus := payment.Status

	// 更新支付状态
	switch callbackData.Status {
	case "success":
		payment.Status = model.PaymentStatusSuccess
		payment.PaidAt = callbackData.PaidAt
		if payment.PaidAt == nil {
			now := time.Now()
			payment.PaidAt = &now
		}
	case "failed":
		payment.Status = model.PaymentStatusFailed
		payment.ErrorMsg = callbackData.FailureReason
	case "cancelled":
		payment.Status = model.PaymentStatusCancelled
	default:
		return "", fmt.Errorf("unknown callback status: %s", callbackData.Status)
	}

	// 更新渠道订单号（如果有）
	if callbackData.ChannelOrderNo != "" {
		payment.ChannelOrderNo = callbackData.ChannelOrderNo
	}

	// 保存到数据库
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return "", fmt.Errorf("update payment status failed: %w", err)
	}

	// 返回更新结果
	result := map[string]interface{}{
		"original_status": originalStatus,
		"new_status":      payment.Status,
		"paid_at":         payment.PaidAt,
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateUpdatePaymentStatus 补偿更新支付状态步骤
func (s *CallbackSagaService) compensateUpdatePaymentStatus(ctx context.Context, payment *model.Payment) error {
	logger.Info("compensating update payment status step",
		zap.String("payment_no", payment.PaymentNo))

	// 恢复支付状态为处理中（pending）
	payment.Status = model.PaymentStatusPending
	payment.PaidAt = nil
	payment.ErrorMsg = ""

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		logger.Error("failed to restore payment status during compensation",
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		return fmt.Errorf("restore payment status failed: %w", err)
	}

	return nil
}

// executeUpdateOrderStatus 执行更新订单状态步骤
func (s *CallbackSagaService) executeUpdateOrderStatus(ctx context.Context, payment *model.Payment, callbackData *CallbackData) (string, error) {
	logger.Info("executing update order status step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("order_no", payment.OrderNo))

	if s.orderClient == nil {
		return "", fmt.Errorf("order client is nil")
	}

	// 调用 order-service 更新订单状态
	var orderStatus string
	var paidAtStr string
	switch callbackData.Status {
	case "success":
		orderStatus = "paid"
		if payment.PaidAt != nil {
			paidAtStr = payment.PaidAt.Format(time.RFC3339)
		}
	case "failed":
		orderStatus = "payment_failed"
	case "cancelled":
		orderStatus = "cancelled"
	default:
		orderStatus = "unknown"
	}

	updateReq := &client.UpdateOrderStatusRequest{
		Status:         orderStatus,
		ChannelOrderNo: payment.ChannelOrderNo,
		PaidAt:         paidAtStr,
	}

	err := s.orderClient.UpdateOrderStatus(ctx, payment.PaymentNo, updateReq)
	if err != nil {
		return "", fmt.Errorf("update order status failed: %w", err)
	}

	// 返回更新结果
	result := map[string]interface{}{
		"order_no":     payment.OrderNo,
		"order_status": orderStatus,
		"updated_at":   time.Now().Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateUpdateOrderStatus 补偿更新订单状态步骤
func (s *CallbackSagaService) compensateUpdateOrderStatus(ctx context.Context, payment *model.Payment) error {
	logger.Info("compensating update order status step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("order_no", payment.OrderNo))

	if s.orderClient == nil {
		return fmt.Errorf("order client is nil")
	}

	// 恢复订单状态为待支付
	updateReq := &client.UpdateOrderStatusRequest{
		Status: "pending",
	}

	err := s.orderClient.UpdateOrderStatus(ctx, payment.PaymentNo, updateReq)
	if err != nil {
		logger.Error("failed to restore order status during compensation",
			zap.String("payment_no", payment.PaymentNo),
			zap.String("order_no", payment.OrderNo),
			zap.Error(err))
		return fmt.Errorf("restore order status failed: %w", err)
	}

	return nil
}

// executePublishEvent 执行发布事件步骤
func (s *CallbackSagaService) executePublishEvent(ctx context.Context, payment *model.Payment, callbackData *CallbackData) (string, error) {
	logger.Info("executing publish event step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("status", callbackData.Status))

	if s.kafkaProducer == nil {
		logger.Warn("kafka producer is nil, skip event publishing")
		return "{}", nil
	}

	// 构建事件消息
	event := map[string]interface{}{
		"event_type":   "payment.callback",
		"payment_no":   payment.PaymentNo,
		"order_no":     payment.OrderNo,
		"merchant_id":  payment.MerchantID.String(),
		"status":       callbackData.Status,
		"amount":       payment.Amount,
		"currency":     payment.Currency,
		"paid_at":      payment.PaidAt,
		"occurred_at":  time.Now().Format(time.RFC3339),
	}

	eventBytes, _ := json.Marshal(event)

	// 发布到 Kafka
	if err := s.kafkaProducer.Publish(ctx, "payment.events", string(eventBytes)); err != nil {
		// 事件发布失败不影响主流程
		logger.Error("failed to publish payment event",
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		// 不返回错误，允许 Saga 继续
	}

	result := map[string]interface{}{
		"event_published": true,
		"event_type":      "payment.callback",
		"published_at":    time.Now().Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensatePublishEvent 补偿发布事件步骤
func (s *CallbackSagaService) compensatePublishEvent(ctx context.Context, payment *model.Payment) error {
	logger.Info("compensating publish event step",
		zap.String("payment_no", payment.PaymentNo))

	if s.kafkaProducer == nil {
		return nil
	}

	// 发布补偿事件
	compensationEvent := map[string]interface{}{
		"event_type":   "payment.callback.compensated",
		"payment_no":   payment.PaymentNo,
		"order_no":     payment.OrderNo,
		"merchant_id":  payment.MerchantID.String(),
		"reason":       "Saga compensation triggered",
		"occurred_at":  time.Now().Format(time.RFC3339),
	}

	eventBytes, _ := json.Marshal(compensationEvent)

	if err := s.kafkaProducer.Publish(ctx, "payment.events", string(eventBytes)); err != nil {
		logger.Error("failed to publish compensation event",
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		// 不返回错误，允许补偿继续
	}

	return nil
}

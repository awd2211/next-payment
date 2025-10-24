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

// RefundSagaService 退款 Saga 服务（用于协调分布式事务）
type RefundSagaService struct {
	orchestrator  *saga.SagaOrchestrator
	paymentRepo   repository.PaymentRepository
	channelClient *client.ChannelClient
	orderClient   *client.OrderClient
}

// NewRefundSagaService 创建退款 Saga 服务
func NewRefundSagaService(
	orchestrator *saga.SagaOrchestrator,
	paymentRepo repository.PaymentRepository,
	channelClient *client.ChannelClient,
	orderClient *client.OrderClient,
	accountingClient interface{}, // 保留参数兼容性，但不使用
) *RefundSagaService {
	return &RefundSagaService{
		orchestrator:  orchestrator,
		paymentRepo:   paymentRepo,
		channelClient: channelClient,
		orderClient:   orderClient,
	}
}

// ExecuteRefundSaga 执行退款 Saga
func (s *RefundSagaService) ExecuteRefundSaga(
	ctx context.Context,
	refund *model.Refund,
	payment *model.Payment,
) error {
	// 1. 构建 Saga
	sagaBuilder := s.orchestrator.NewSagaBuilder(refund.RefundNo, "refund")
	sagaBuilder.SetMetadata(map[string]interface{}{
		"refund_no":   refund.RefundNo,
		"payment_no":  payment.PaymentNo,
		"merchant_id": payment.MerchantID.String(),
		"amount":      refund.Amount,
		"currency":    payment.Currency,
	})

	// 2. 定义步骤
	stepDefs := []saga.StepDefinition{
		{
			Name: "CallChannelRefund",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeCallChannelRefund(ctx, refund, payment)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateCallChannelRefund(ctx, refund, payment, executeResult)
			},
			MaxRetryCount: 3,
			Timeout:       60 * time.Second,
		},
		{
			Name: "UpdatePaymentStatus",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeUpdatePaymentStatus(ctx, refund, payment)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateUpdatePaymentStatus(ctx, refund, payment)
			},
			MaxRetryCount: 3,
			Timeout:       10 * time.Second,
		},
		{
			Name: "UpdateRefundStatus",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeUpdateRefundStatus(ctx, refund)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateUpdateRefundStatus(ctx, refund)
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

	logger.Info("refund saga created",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("refund_no", refund.RefundNo))

	// 4. 执行 Saga
	if err := s.orchestrator.Execute(ctx, sagaInstance, stepDefs); err != nil {
		logger.Error("refund saga execution failed",
			zap.String("saga_id", sagaInstance.ID.String()),
			zap.String("refund_no", refund.RefundNo),
			zap.Error(err))
		return err
	}

	logger.Info("refund saga completed",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("refund_no", refund.RefundNo))

	return nil
}

// executeCallChannelRefund 执行调用渠道退款步骤
func (s *RefundSagaService) executeCallChannelRefund(ctx context.Context, refund *model.Refund, payment *model.Payment) (string, error) {
	logger.Info("executing call channel refund step",
		zap.String("refund_no", refund.RefundNo),
		zap.String("payment_no", payment.PaymentNo))

	if s.channelClient == nil {
		return "", fmt.Errorf("channel client is nil")
	}

	// 调用渠道退款
	channelResp, err := s.channelClient.CreateRefund(ctx, &client.RefundRequest{
		PaymentNo:      payment.PaymentNo,
		RefundNo:       refund.RefundNo,
		ChannelOrderNo: payment.ChannelOrderNo,
		Amount:         refund.Amount,
		Currency:       payment.Currency,
		Reason:         refund.Reason,
	})

	if err != nil {
		return "", fmt.Errorf("channel refund failed: %w", err)
	}

	// 更新退款记录的渠道退款号
	refund.ChannelRefundNo = channelResp.ChannelRefundNo

	// 返回渠道响应（JSON格式）
	resultBytes, _ := json.Marshal(channelResp)
	return string(resultBytes), nil
}

// compensateCallChannelRefund 补偿调用渠道退款步骤
func (s *RefundSagaService) compensateCallChannelRefund(ctx context.Context, refund *model.Refund, payment *model.Payment, executeResult string) error {
	logger.Info("compensating call channel refund step",
		zap.String("refund_no", refund.RefundNo),
		zap.String("channel_refund_no", refund.ChannelRefundNo))

	// 注意：大部分支付渠道不支持取消退款
	// 退款一旦发起，通常无法撤销
	// 这里记录日志，可能需要人工介入

	if refund.ChannelRefundNo != "" {
		logger.Warn("channel refund cannot be automatically cancelled, manual processing may be required",
			zap.String("refund_no", refund.RefundNo),
			zap.String("channel_refund_no", refund.ChannelRefundNo))
	}

	return nil
}

// executeUpdatePaymentStatus 执行更新支付状态步骤
func (s *RefundSagaService) executeUpdatePaymentStatus(ctx context.Context, refund *model.Refund, payment *model.Payment) (string, error) {
	logger.Info("executing update payment status step",
		zap.String("refund_no", refund.RefundNo),
		zap.String("payment_no", payment.PaymentNo))

	// 更新支付记录的退款状态
	// 注意：当前 Payment 模型没有 RefundedAmount 字段和 PaymentStatusRefunded 状态
	// 这里简化实现，实际可能需要在 Payment 模型中添加这些字段
	originalStatus := payment.Status
	// payment.Status = "refunded" // 简化：暂不修改支付状态

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return "", fmt.Errorf("update payment status failed: %w", err)
	}

	// 返回更新结果
	result := map[string]interface{}{
		"original_status": originalStatus,
		"new_status":      payment.Status,
		"refunded_amount": refund.Amount,
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateUpdatePaymentStatus 补偿更新支付状态步骤
func (s *RefundSagaService) compensateUpdatePaymentStatus(ctx context.Context, refund *model.Refund, payment *model.Payment) error {
	logger.Info("compensating update payment status step",
		zap.String("refund_no", refund.RefundNo),
		zap.String("payment_no", payment.PaymentNo))

	// 恢复支付状态（简化实现）
	payment.Status = model.PaymentStatusSuccess // 恢复为成功状态

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		logger.Error("failed to restore payment status during compensation",
			zap.String("refund_no", refund.RefundNo),
			zap.Error(err))
		return fmt.Errorf("restore payment status failed: %w", err)
	}

	return nil
}

// executeUpdateRefundStatus 执行更新退款状态步骤
func (s *RefundSagaService) executeUpdateRefundStatus(ctx context.Context, refund *model.Refund) (string, error) {
	logger.Info("executing update refund status step",
		zap.String("refund_no", refund.RefundNo))

	// 更新退款状态为成功
	refund.Status = model.RefundStatusSuccess
	now := time.Now()
	refund.RefundedAt = &now

	if err := s.paymentRepo.UpdateRefund(ctx, refund); err != nil {
		return "", fmt.Errorf("update refund status failed: %w", err)
	}

	result := map[string]interface{}{
		"status":      string(refund.Status),
		"refunded_at": now.Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateUpdateRefundStatus 补偿更新退款状态步骤
func (s *RefundSagaService) compensateUpdateRefundStatus(ctx context.Context, refund *model.Refund) error {
	logger.Info("compensating update refund status step",
		zap.String("refund_no", refund.RefundNo))

	// 恢复退款状态为失败
	refund.Status = model.RefundStatusFailed
	refund.ErrorMsg = "Saga 补偿: 分布式事务回滚"
	refund.RefundedAt = nil

	if err := s.paymentRepo.UpdateRefund(ctx, refund); err != nil {
		logger.Error("failed to restore refund status during compensation",
			zap.String("refund_no", refund.RefundNo),
			zap.Error(err))
		return fmt.Errorf("restore refund status failed: %w", err)
	}

	return nil
}

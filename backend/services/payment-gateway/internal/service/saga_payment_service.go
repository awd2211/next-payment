package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/saga"
	"go.uber.org/zap"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
)

// SagaPaymentService Saga 支付服务（用于协调分布式事务）
type SagaPaymentService struct {
	orchestrator   *saga.SagaOrchestrator
	paymentRepo    repository.PaymentRepository
	orderClient    *client.OrderClient
	channelClient  *client.ChannelClient
}

// NewSagaPaymentService 创建 Saga 支付服务
func NewSagaPaymentService(
	orchestrator *saga.SagaOrchestrator,
	paymentRepo repository.PaymentRepository,
	orderClient *client.OrderClient,
	channelClient *client.ChannelClient,
) *SagaPaymentService {
	return &SagaPaymentService{
		orchestrator:  orchestrator,
		paymentRepo:   paymentRepo,
		orderClient:   orderClient,
		channelClient: channelClient,
	}
}

// ExecutePaymentSaga 执行支付 Saga
func (s *SagaPaymentService) ExecutePaymentSaga(
	ctx context.Context,
	payment *model.Payment,
) error {
	// 1. 构建 Saga
	sagaBuilder := s.orchestrator.NewSagaBuilder(payment.PaymentNo, "payment")
	sagaBuilder.SetMetadata(map[string]interface{}{
		"payment_no":  payment.PaymentNo,
		"merchant_id": payment.MerchantID.String(),
		"order_no":    payment.OrderNo,
		"amount":      payment.Amount,
		"currency":    payment.Currency,
	})

	// 2. 定义步骤
	stepDefs := []saga.StepDefinition{
		// 步骤1: 创建订单
		{
			Name: "CreateOrder",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeCreateOrder(ctx, payment)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateCreateOrder(ctx, payment)
			},
			MaxRetryCount: 3,
		},
		// 步骤2: 调用支付渠道
		{
			Name: "CallPaymentChannel",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeCallPaymentChannel(ctx, payment)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateCallPaymentChannel(ctx, payment, executeResult)
			},
			MaxRetryCount: 3,
		},
	}

	// 添加步骤到构建器
	for _, stepDef := range stepDefs {
		sagaBuilder.AddStep(stepDef.Name, stepDef.Execute, stepDef.Compensate, stepDef.MaxRetryCount)
	}

	// 3. 构建并执行 Saga
	sagaInstance, err := sagaBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build saga: %w", err)
	}

	logger.Info("payment saga created",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("payment_no", payment.PaymentNo))

	// 4. 执行 Saga
	if err := s.orchestrator.Execute(ctx, sagaInstance, stepDefs); err != nil {
		logger.Error("payment saga execution failed",
			zap.String("saga_id", sagaInstance.ID.String()),
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		return err
	}

	logger.Info("payment saga completed",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("payment_no", payment.PaymentNo))

	return nil
}

// executeCreateOrder 执行创建订单步骤
func (s *SagaPaymentService) executeCreateOrder(ctx context.Context, payment *model.Payment) (string, error) {
	logger.Info("executing create order step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("order_no", payment.OrderNo))

	if s.orderClient == nil {
		return "", fmt.Errorf("order client is nil")
	}

	resp, err := s.orderClient.CreateOrder(ctx, &client.CreateOrderRequest{
		MerchantID:    payment.MerchantID,
		OrderNo:       payment.OrderNo,
		PaymentNo:     payment.PaymentNo,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Channel:       payment.Channel,
		PayMethod:     payment.PayMethod,
		CustomerEmail: payment.CustomerEmail,
		CustomerName:  payment.CustomerName,
		CustomerPhone: payment.CustomerPhone,
		CustomerIP:    payment.CustomerIP,
		Description:   payment.Description,
		Extra:         payment.Extra,
	})

	if err != nil {
		return "", fmt.Errorf("create order failed: %w", err)
	}

	// 返回订单创建结果（JSON格式）
	resultBytes, _ := json.Marshal(resp)
	return string(resultBytes), nil
}

// compensateCreateOrder 补偿创建订单步骤
func (s *SagaPaymentService) compensateCreateOrder(ctx context.Context, payment *model.Payment) error {
	logger.Info("compensating create order step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("order_no", payment.OrderNo))

	if s.orderClient == nil {
		return fmt.Errorf("order client is nil")
	}

	// 调用 Order Service 取消订单
	err := s.orderClient.CancelOrder(ctx, payment.OrderNo, "支付流程失败，自动取消")
	if err != nil {
		logger.Error("failed to cancel order during compensation",
			zap.String("payment_no", payment.PaymentNo),
			zap.String("order_no", payment.OrderNo),
			zap.Error(err))
		return fmt.Errorf("cancel order failed: %w", err)
	}

	logger.Info("order canceled successfully",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("order_no", payment.OrderNo))

	return nil
}

// executeCallPaymentChannel 执行调用支付渠道步骤
func (s *SagaPaymentService) executeCallPaymentChannel(ctx context.Context, payment *model.Payment) (string, error) {
	logger.Info("executing call payment channel step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("channel", payment.Channel))

	if s.channelClient == nil {
		return "", fmt.Errorf("channel client is nil")
	}

	// 解析扩展信息
	var extraMap map[string]interface{}
	if payment.Extra != "" {
		if err := json.Unmarshal([]byte(payment.Extra), &extraMap); err != nil {
			extraMap = nil
		}
	}

	// 调用 Channel Adapter 发起支付
	resp, err := s.channelClient.CreatePayment(ctx, &client.CreatePaymentRequest{
		PaymentNo:     payment.PaymentNo,
		MerchantID:    payment.MerchantID.String(),
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Channel:       payment.Channel,
		PayMethod:     payment.PayMethod,
		CustomerEmail: payment.CustomerEmail,
		CustomerName:  payment.CustomerName,
		Description:   payment.Description,
		NotifyURL:     payment.NotifyURL,
		ReturnURL:     payment.ReturnURL,
		Extra:         extraMap,
	})

	if err != nil {
		return "", fmt.Errorf("call payment channel failed: %w", err)
	}

	// 更新支付记录的渠道订单号，并将支付URL存储到Extra中
	if resp.ChannelTradeNo != "" {
		payment.ChannelOrderNo = resp.ChannelTradeNo
	} else if resp.ChannelOrderNo != "" {
		payment.ChannelOrderNo = resp.ChannelOrderNo
	}

	// 将 PaymentURL 存储到 Extra 中
	if resp.PaymentURL != "" {
		var extraMap map[string]interface{}
		if payment.Extra != "" {
			if err := json.Unmarshal([]byte(payment.Extra), &extraMap); err != nil {
				// 如果解析失败，创建新的 map
				extraMap = make(map[string]interface{})
			}
		} else {
			extraMap = make(map[string]interface{})
		}
		extraMap["payment_url"] = resp.PaymentURL
		if extraBytes, err := json.Marshal(extraMap); err == nil {
			payment.Extra = string(extraBytes)
		}
		// 如果 Marshal 失败，保持原有的 Extra 不变
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		logger.Error("failed to update payment after channel call",
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		// 不返回错误，因为支付已经发起成功
	}

	// 返回渠道响应结果（JSON格式）
	resultBytes, _ := json.Marshal(resp)
	return string(resultBytes), nil
}

// compensateCallPaymentChannel 补偿调用支付渠道步骤
func (s *SagaPaymentService) compensateCallPaymentChannel(ctx context.Context, payment *model.Payment, executeResult string) error {
	logger.Info("compensating call payment channel step",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("channel", payment.Channel))

	if s.channelClient == nil {
		return fmt.Errorf("channel client is nil")
	}

	// 如果有渠道订单号，调用 Channel Adapter 取消支付
	if payment.ChannelOrderNo != "" {
		err := s.channelClient.CancelPayment(ctx, payment.ChannelOrderNo)
		if err != nil {
			logger.Error("failed to cancel payment in channel during compensation",
				zap.String("payment_no", payment.PaymentNo),
				zap.String("channel_order_no", payment.ChannelOrderNo),
				zap.Error(err))
			// 某些支付渠道可能不支持取消，记录错误但不返回
		} else {
			logger.Info("payment canceled in channel successfully",
				zap.String("payment_no", payment.PaymentNo),
				zap.String("channel_order_no", payment.ChannelOrderNo))
		}
	}

	// 更新支付记录状态为失败
	payment.Status = model.PaymentStatusFailed
	payment.ErrorMsg = "Saga 补偿: 分布式事务回滚"
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		logger.Error("failed to update payment status during compensation",
			zap.String("payment_no", payment.PaymentNo),
			zap.Error(err))
		return fmt.Errorf("update payment status failed: %w", err)
	}

	return nil
}

// CompensatePayment 补偿支付（手动触发）
func (s *SagaPaymentService) CompensatePayment(ctx context.Context, paymentNo string) error {
	// 获取 Saga 实例
	sagaInstance, err := s.orchestrator.GetSagaByBusinessID(ctx, paymentNo)
	if err != nil {
		return fmt.Errorf("failed to get saga: %w", err)
	}

	// 获取支付记录
	payment, err := s.paymentRepo.GetByPaymentNo(ctx, paymentNo)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// 定义步骤（与 ExecutePaymentSaga 相同）
	stepDefs := []saga.StepDefinition{
		{
			Name: "CreateOrder",
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateCreateOrder(ctx, payment)
			},
		},
		{
			Name: "CallPaymentChannel",
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateCallPaymentChannel(ctx, payment, executeResult)
			},
		},
	}

	// 执行补偿
	return s.orchestrator.Compensate(ctx, sagaInstance, stepDefs)
}

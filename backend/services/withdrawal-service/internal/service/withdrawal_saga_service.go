package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/saga"
	"go.uber.org/zap"
	"payment-platform/withdrawal-service/internal/client"
	"payment-platform/withdrawal-service/internal/model"
	"payment-platform/withdrawal-service/internal/repository"
)

// WithdrawalSagaService 提现 Saga 服务（用于协调分布式事务）
type WithdrawalSagaService struct {
	orchestrator       *saga.SagaOrchestrator
	withdrawalRepo     repository.WithdrawalRepository
	accountingClient   *client.AccountingClient
	bankTransferClient *client.BankTransferClient
	notificationClient *client.NotificationClient
}

// NewWithdrawalSagaService 创建提现 Saga 服务
func NewWithdrawalSagaService(
	orchestrator *saga.SagaOrchestrator,
	withdrawalRepo repository.WithdrawalRepository,
	accountingClient *client.AccountingClient,
	bankTransferClient *client.BankTransferClient,
	notificationClient *client.NotificationClient,
) *WithdrawalSagaService {
	return &WithdrawalSagaService{
		orchestrator:       orchestrator,
		withdrawalRepo:     withdrawalRepo,
		accountingClient:   accountingClient,
		bankTransferClient: bankTransferClient,
		notificationClient: notificationClient,
	}
}

// ExecuteWithdrawalSaga 执行提现 Saga
func (s *WithdrawalSagaService) ExecuteWithdrawalSaga(
	ctx context.Context,
	withdrawal *model.Withdrawal,
) error {
	// 1. 构建 Saga
	sagaBuilder := s.orchestrator.NewSagaBuilder(withdrawal.WithdrawalNo, "withdrawal")
	sagaBuilder.SetMetadata(map[string]interface{}{
		"withdrawal_no": withdrawal.WithdrawalNo,
		"merchant_id":   withdrawal.MerchantID.String(),
		"amount":        withdrawal.Amount,
		"actual_amount": withdrawal.ActualAmount,
		"fee":           withdrawal.Fee,
	})

	// 2. 定义步骤（使用带超时的构建方法）

	// 步骤1: 预冻结余额（30秒超时）
	sagaBuilder.AddStepWithTimeout(
		"PreFreezeBalance",
		func(ctx context.Context, executeData string) (string, error) {
			return s.executePreFreezeBalance(ctx, withdrawal)
		},
		func(ctx context.Context, compensateData string, executeResult string) error {
			return s.compensatePreFreezeBalance(ctx, withdrawal)
		},
		3,
		30*time.Second,
	)

	// 步骤2: 执行银行转账（120秒超时，银行接口较慢）
	sagaBuilder.AddStepWithTimeout(
		"ExecuteBankTransfer",
		func(ctx context.Context, executeData string) (string, error) {
			return s.executeBankTransfer(ctx, withdrawal)
		},
		func(ctx context.Context, compensateData string, executeResult string) error {
			return s.compensateBankTransfer(ctx, withdrawal, executeResult)
		},
		3,
		120*time.Second,
	)

	// 步骤3: 扣减余额（30秒超时）
	sagaBuilder.AddStepWithTimeout(
		"DeductBalance",
		func(ctx context.Context, executeData string) (string, error) {
			return s.executeDeductBalance(ctx, withdrawal)
		},
		func(ctx context.Context, compensateData string, executeResult string) error {
			return s.compensateDeductBalance(ctx, withdrawal)
		},
		3,
		30*time.Second,
	)

	// 步骤4: 更新提现状态（10秒超时）
	sagaBuilder.AddStepWithTimeout(
		"UpdateWithdrawalStatus",
		func(ctx context.Context, executeData string) (string, error) {
			return s.executeUpdateStatus(ctx, withdrawal)
		},
		func(ctx context.Context, compensateData string, executeResult string) error {
			return s.compensateUpdateStatus(ctx, withdrawal)
		},
		3,
		10*time.Second,
	)

	// 获取步骤定义（用于执行）
	stepDefs := []saga.StepDefinition{
		{
			Name: "PreFreezeBalance",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executePreFreezeBalance(ctx, withdrawal)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensatePreFreezeBalance(ctx, withdrawal)
			},
			MaxRetryCount: 3,
			Timeout:       30 * time.Second,
		},
		{
			Name: "ExecuteBankTransfer",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeBankTransfer(ctx, withdrawal)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateBankTransfer(ctx, withdrawal, executeResult)
			},
			MaxRetryCount: 3,
			Timeout:       120 * time.Second,
		},
		{
			Name: "DeductBalance",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeDeductBalance(ctx, withdrawal)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateDeductBalance(ctx, withdrawal)
			},
			MaxRetryCount: 3,
			Timeout:       30 * time.Second,
		},
		{
			Name: "UpdateWithdrawalStatus",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeUpdateStatus(ctx, withdrawal)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateUpdateStatus(ctx, withdrawal)
			},
			MaxRetryCount: 3,
			Timeout:       10 * time.Second,
		},
	}

	// 3. 构建并执行 Saga
	sagaInstance, err := sagaBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build saga: %w", err)
	}

	logger.Info("withdrawal saga created",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	// 4. 执行 Saga
	if err := s.orchestrator.Execute(ctx, sagaInstance, stepDefs); err != nil {
		logger.Error("withdrawal saga execution failed",
			zap.String("saga_id", sagaInstance.ID.String()),
			zap.String("withdrawal_no", withdrawal.WithdrawalNo),
			zap.Error(err))
		return err
	}

	logger.Info("withdrawal saga completed",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	return nil
}

// executePreFreezeBalance 执行预冻结余额步骤
func (s *WithdrawalSagaService) executePreFreezeBalance(ctx context.Context, withdrawal *model.Withdrawal) (string, error) {
	logger.Info("executing pre-freeze balance step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo),
		zap.Int64("amount", withdrawal.Amount))

	if s.accountingClient == nil {
		return "", fmt.Errorf("accounting client is nil")
	}

	// 调用 accounting-service 冻结余额
	freezeReq := &client.FreezeBalanceRequest{
		MerchantID:      withdrawal.MerchantID,
		Amount:          withdrawal.Amount,
		TransactionType: "withdrawal_freeze",
		RelatedNo:       withdrawal.WithdrawalNo,
		Description:     fmt.Sprintf("提现冻结: %s", withdrawal.WithdrawalNo),
	}

	err := s.accountingClient.FreezeBalance(ctx, freezeReq)
	if err != nil {
		return "", fmt.Errorf("freeze balance failed: %w", err)
	}

	// 返回冻结结果（JSON格式）
	result := map[string]interface{}{
		"frozen_amount": withdrawal.Amount,
		"frozen_at":     time.Now().Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensatePreFreezeBalance 补偿预冻结余额步骤
func (s *WithdrawalSagaService) compensatePreFreezeBalance(ctx context.Context, withdrawal *model.Withdrawal) error {
	logger.Info("compensating pre-freeze balance step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	if s.accountingClient == nil {
		return fmt.Errorf("accounting client is nil")
	}

	// 调用 accounting-service 解冻余额
	unfreezeReq := &client.UnfreezeBalanceRequest{
		MerchantID:      withdrawal.MerchantID,
		Amount:          withdrawal.Amount,
		TransactionType: "withdrawal_unfreeze",
		RelatedNo:       withdrawal.WithdrawalNo,
		Description:     fmt.Sprintf("提现解冻(补偿): %s", withdrawal.WithdrawalNo),
	}

	err := s.accountingClient.UnfreezeBalance(ctx, unfreezeReq)
	if err != nil {
		logger.Error("failed to unfreeze balance during compensation",
			zap.String("withdrawal_no", withdrawal.WithdrawalNo),
			zap.Error(err))
		return fmt.Errorf("unfreeze balance failed: %w", err)
	}

	logger.Info("balance unfrozen successfully",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	return nil
}

// executeBankTransfer 执行银行转账步骤
func (s *WithdrawalSagaService) executeBankTransfer(ctx context.Context, withdrawal *model.Withdrawal) (string, error) {
	logger.Info("executing bank transfer step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo),
		zap.Int64("actual_amount", withdrawal.ActualAmount))

	if s.bankTransferClient == nil {
		return "", fmt.Errorf("bank transfer client is nil")
	}

	// 调用银行转账接口
	transferReq := &client.TransferRequest{
		OrderNo:         withdrawal.WithdrawalNo,
		BankName:        withdrawal.BankName,
		BankAccountName: withdrawal.BankAccountName,
		BankAccountNo:   withdrawal.BankAccountNo,
		Amount:          withdrawal.ActualAmount,
		Currency:        "CNY",
		Remarks:         withdrawal.Remarks,
	}

	transferResp, err := s.bankTransferClient.Transfer(ctx, transferReq)
	if err != nil {
		return "", fmt.Errorf("bank transfer failed: %w", err)
	}

	// 更新提现记录的渠道订单号
	withdrawal.ChannelTradeNo = transferResp.ChannelTradeNo

	// 返回转账结果（JSON格式）
	resultBytes, _ := json.Marshal(transferResp)
	return string(resultBytes), nil
}

// compensateBankTransfer 补偿银行转账步骤
func (s *WithdrawalSagaService) compensateBankTransfer(ctx context.Context, withdrawal *model.Withdrawal, executeResult string) error {
	logger.Info("compensating bank transfer step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo),
		zap.String("channel_trade_no", withdrawal.ChannelTradeNo))

	if s.bankTransferClient == nil {
		return fmt.Errorf("bank transfer client is nil")
	}

	// 如果银行支持退款，调用退款接口
	// 注意：部分银行可能不支持自动退款，需要人工处理
	if withdrawal.ChannelTradeNo != "" {
		refundReq := &client.RefundTransferRequest{
			OriginalOrderNo: withdrawal.WithdrawalNo,
			ChannelTradeNo:  withdrawal.ChannelTradeNo,
			Amount:          withdrawal.ActualAmount,
			Reason:          "提现流程失败，自动退款",
		}

		err := s.bankTransferClient.RefundTransfer(ctx, refundReq)
		if err != nil {
			logger.Error("failed to refund bank transfer during compensation",
				zap.String("withdrawal_no", withdrawal.WithdrawalNo),
				zap.String("channel_trade_no", withdrawal.ChannelTradeNo),
				zap.Error(err))
			// 银行退款失败，记录错误但不返回（需要人工处理）
			// 这种情况会进入 DLQ
		} else {
			logger.Info("bank transfer refunded successfully",
				zap.String("withdrawal_no", withdrawal.WithdrawalNo))
		}
	}

	return nil
}

// executeDeductBalance 执行扣减余额步骤
func (s *WithdrawalSagaService) executeDeductBalance(ctx context.Context, withdrawal *model.Withdrawal) (string, error) {
	logger.Info("executing deduct balance step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo),
		zap.Int64("amount", withdrawal.Amount))

	if s.accountingClient == nil {
		return "", fmt.Errorf("accounting client is nil")
	}

	// 调用 accounting-service 扣减余额
	deductReq := &client.DeductBalanceRequest{
		MerchantID:      withdrawal.MerchantID,
		Amount:          withdrawal.Amount, // 扣减总金额（包含手续费）
		TransactionType: "withdrawal",
		RelatedNo:       withdrawal.WithdrawalNo,
		Description: fmt.Sprintf("提现: %s, 实际到账: %.2f元, 手续费: %.2f元",
			withdrawal.WithdrawalNo,
			float64(withdrawal.ActualAmount)/100,
			float64(withdrawal.Fee)/100),
	}

	err := s.accountingClient.DeductBalance(ctx, deductReq)
	if err != nil {
		return "", fmt.Errorf("deduct balance failed: %w", err)
	}

	// 返回扣减结果（JSON格式）
	result := map[string]interface{}{
		"deducted_amount": withdrawal.Amount,
		"deducted_at":     time.Now().Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateDeductBalance 补偿扣减余额步骤
func (s *WithdrawalSagaService) compensateDeductBalance(ctx context.Context, withdrawal *model.Withdrawal) error {
	logger.Info("compensating deduct balance step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	if s.accountingClient == nil {
		return fmt.Errorf("accounting client is nil")
	}

	// 调用 accounting-service 退还余额
	refundReq := &client.RefundBalanceRequest{
		MerchantID:      withdrawal.MerchantID,
		Amount:          withdrawal.Amount,
		TransactionType: "withdrawal_refund",
		RelatedNo:       withdrawal.WithdrawalNo,
		Description:     fmt.Sprintf("提现退还(补偿): %s", withdrawal.WithdrawalNo),
	}

	err := s.accountingClient.RefundBalance(ctx, refundReq)
	if err != nil {
		logger.Error("failed to refund balance during compensation",
			zap.String("withdrawal_no", withdrawal.WithdrawalNo),
			zap.Error(err))
		return fmt.Errorf("refund balance failed: %w", err)
	}

	logger.Info("balance refunded successfully",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	return nil
}

// executeUpdateStatus 执行更新提现状态步骤
func (s *WithdrawalSagaService) executeUpdateStatus(ctx context.Context, withdrawal *model.Withdrawal) (string, error) {
	logger.Info("executing update withdrawal status step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	// 更新提现状态为已完成
	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusCompleted
	withdrawal.CompletedAt = &now

	if err := s.withdrawalRepo.Update(ctx, withdrawal); err != nil {
		return "", fmt.Errorf("update withdrawal status failed: %w", err)
	}

	// 发送完成通知
	if s.notificationClient != nil {
		if err := s.notificationClient.SendWithdrawalStatusNotification(ctx, withdrawal.MerchantID, withdrawal.WithdrawalNo, "completed", withdrawal.Amount); err != nil {
			// 通知发送失败不影响主流程
			logger.Error("failed to send completion notification",
				zap.Error(err),
				zap.String("withdrawal_no", withdrawal.WithdrawalNo))
		}
	}

	// 返回更新结果（JSON格式）
	result := map[string]interface{}{
		"status":       string(withdrawal.Status),
		"completed_at": now.Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateUpdateStatus 补偿更新提现状态步骤
func (s *WithdrawalSagaService) compensateUpdateStatus(ctx context.Context, withdrawal *model.Withdrawal) error {
	logger.Info("compensating update withdrawal status step",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	// 恢复提现状态为失败
	withdrawal.Status = model.WithdrawalStatusFailed
	withdrawal.FailureReason = "Saga 补偿: 分布式事务回滚"
	withdrawal.CompletedAt = nil

	if err := s.withdrawalRepo.Update(ctx, withdrawal); err != nil {
		logger.Error("failed to update withdrawal status during compensation",
			zap.String("withdrawal_no", withdrawal.WithdrawalNo),
			zap.Error(err))
		return fmt.Errorf("update withdrawal status failed: %w", err)
	}

	return nil
}

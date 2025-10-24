package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/saga"
	"go.uber.org/zap"
	"payment-platform/settlement-service/internal/client"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/repository"
)

// SettlementSagaService 结算 Saga 服务（用于协调分布式事务）
type SettlementSagaService struct {
	orchestrator     *saga.SagaOrchestrator
	settlementRepo   repository.SettlementRepository
	merchantClient   *client.MerchantClient
	withdrawalClient *client.WithdrawalClient
}

// NewSettlementSagaService 创建结算 Saga 服务
func NewSettlementSagaService(
	orchestrator *saga.SagaOrchestrator,
	settlementRepo repository.SettlementRepository,
	merchantClient *client.MerchantClient,
	withdrawalClient *client.WithdrawalClient,
) *SettlementSagaService {
	return &SettlementSagaService{
		orchestrator:     orchestrator,
		settlementRepo:   settlementRepo,
		merchantClient:   merchantClient,
		withdrawalClient: withdrawalClient,
	}
}

// ExecuteSettlementSaga 执行结算 Saga
func (s *SettlementSagaService) ExecuteSettlementSaga(
	ctx context.Context,
	settlement *model.Settlement,
) error {
	// 1. 构建 Saga
	sagaBuilder := s.orchestrator.NewSagaBuilder(settlement.SettlementNo, "settlement")
	sagaBuilder.SetMetadata(map[string]interface{}{
		"settlement_no": settlement.SettlementNo,
		"merchant_id":   settlement.MerchantID.String(),
		"amount":        settlement.SettlementAmount,
		"cycle":         settlement.Cycle,
	})

	// 2. 定义步骤
	stepDefs := []saga.StepDefinition{
		{
			Name: "UpdateSettlementProcessing",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeUpdateSettlementProcessing(ctx, settlement)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateUpdateSettlementProcessing(ctx, settlement)
			},
			MaxRetryCount: 3,
			Timeout:       10 * time.Second,
		},
		{
			Name: "GetMerchantAccount",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeGetMerchantAccount(ctx, settlement)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return nil // 查询操作无需补偿
			},
			MaxRetryCount: 3,
			Timeout:       30 * time.Second,
		},
		{
			Name: "CreateWithdrawal",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeCreateWithdrawal(ctx, settlement, executeData)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateCreateWithdrawal(ctx, settlement, executeResult)
			},
			MaxRetryCount: 3,
			Timeout:       30 * time.Second,
		},
		{
			Name: "UpdateSettlementCompleted",
			Execute: func(ctx context.Context, executeData string) (string, error) {
				return s.executeUpdateSettlementCompleted(ctx, settlement, executeData)
			},
			Compensate: func(ctx context.Context, compensateData string, executeResult string) error {
				return s.compensateUpdateSettlementCompleted(ctx, settlement)
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

	logger.Info("settlement saga created",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("settlement_no", settlement.SettlementNo))

	// 4. 执行 Saga
	if err := s.orchestrator.Execute(ctx, sagaInstance, stepDefs); err != nil {
		logger.Error("settlement saga execution failed",
			zap.String("saga_id", sagaInstance.ID.String()),
			zap.String("settlement_no", settlement.SettlementNo),
			zap.Error(err))
		return err
	}

	logger.Info("settlement saga completed",
		zap.String("saga_id", sagaInstance.ID.String()),
		zap.String("settlement_no", settlement.SettlementNo))

	return nil
}

// executeUpdateSettlementProcessing 执行更新结算单为处理中步骤
func (s *SettlementSagaService) executeUpdateSettlementProcessing(ctx context.Context, settlement *model.Settlement) (string, error) {
	logger.Info("executing update settlement to processing",
		zap.String("settlement_no", settlement.SettlementNo))

	// 更新结算单状态为处理中
	now := time.Now()
	originalStatus := settlement.Status
	settlement.Status = model.SettlementStatusProcessing
	settlement.ProcessedAt = &now

	if err := s.settlementRepo.Update(ctx, settlement); err != nil {
		return "", fmt.Errorf("update settlement status failed: %w", err)
	}

	result := map[string]interface{}{
		"original_status": originalStatus,
		"new_status":      settlement.Status,
		"processed_at":    now.Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateUpdateSettlementProcessing 补偿更新结算单状态步骤
func (s *SettlementSagaService) compensateUpdateSettlementProcessing(ctx context.Context, settlement *model.Settlement) error {
	logger.Info("compensating update settlement status",
		zap.String("settlement_no", settlement.SettlementNo))

	// 恢复结算单状态为已审批
	settlement.Status = model.SettlementStatusApproved
	settlement.ProcessedAt = nil

	if err := s.settlementRepo.Update(ctx, settlement); err != nil {
		logger.Error("failed to restore settlement status during compensation",
			zap.String("settlement_no", settlement.SettlementNo),
			zap.Error(err))
		return fmt.Errorf("restore settlement status failed: %w", err)
	}

	return nil
}

// executeGetMerchantAccount 执行获取商户账户步骤
func (s *SettlementSagaService) executeGetMerchantAccount(ctx context.Context, settlement *model.Settlement) (string, error) {
	logger.Info("executing get merchant account",
		zap.String("settlement_no", settlement.SettlementNo),
		zap.String("merchant_id", settlement.MerchantID.String()))

	if s.merchantClient == nil {
		return "", fmt.Errorf("merchant client is nil")
	}

	// 从 merchant-service 获取商户默认银行账户
	defaultAccount, err := s.merchantClient.GetDefaultSettlementAccount(ctx, settlement.MerchantID)
	if err != nil {
		return "", fmt.Errorf("get default settlement account failed: %w", err)
	}

	// 返回账户信息（用于下一步创建提现）
	resultBytes, _ := json.Marshal(defaultAccount)
	return string(resultBytes), nil
}

// executeCreateWithdrawal 执行创建提现步骤
func (s *SettlementSagaService) executeCreateWithdrawal(ctx context.Context, settlement *model.Settlement, accountData string) (string, error) {
	logger.Info("executing create withdrawal",
		zap.String("settlement_no", settlement.SettlementNo))

	if s.withdrawalClient == nil {
		return "", fmt.Errorf("withdrawal client is nil")
	}

	// 解析账户信息
	var account struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal([]byte(accountData), &account); err != nil {
		return "", fmt.Errorf("parse account data failed: %w", err)
	}

	// 调用 withdrawal-service 创建提现
	withdrawalReq := &client.CreateWithdrawalRequest{
		MerchantID:    settlement.MerchantID,
		Amount:        settlement.SettlementAmount,
		Type:          "settlement_auto",
		BankAccountID: account.ID,
		Remarks:       fmt.Sprintf("自动结算: %s, 周期: %s", settlement.SettlementNo, settlement.Cycle),
		CreatedBy:     uuid.MustParse("00000000-0000-0000-0000-000000000000"), // 系统自动
	}

	withdrawalNo, err := s.withdrawalClient.CreateWithdrawalForSettlement(ctx, withdrawalReq)
	if err != nil {
		return "", fmt.Errorf("create withdrawal failed: %w", err)
	}

	// 更新结算单的提现单号
	settlement.WithdrawalNo = withdrawalNo

	result := map[string]interface{}{
		"withdrawal_no": withdrawalNo,
		"amount":        settlement.SettlementAmount,
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateCreateWithdrawal 补偿创建提现步骤
func (s *SettlementSagaService) compensateCreateWithdrawal(ctx context.Context, settlement *model.Settlement, executeResult string) error {
	logger.Info("compensating create withdrawal",
		zap.String("settlement_no", settlement.SettlementNo),
		zap.String("withdrawal_no", settlement.WithdrawalNo))

	if s.withdrawalClient == nil || settlement.WithdrawalNo == "" {
		return nil
	}

	// 解析提现结果
	var result struct {
		WithdrawalNo string `json:"withdrawal_no"`
	}
	if err := json.Unmarshal([]byte(executeResult), &result); err != nil {
		logger.Warn("failed to parse withdrawal result", zap.Error(err))
		return nil
	}

	// 调用 withdrawal-service 取消提现
	// 注意：如果提现已经在处理中，可能无法取消
	logger.Info("attempting to cancel withdrawal",
		zap.String("withdrawal_no", result.WithdrawalNo))

	// 这里简化实现，实际应该调用 CancelWithdrawal API
	// err := s.withdrawalClient.CancelWithdrawal(ctx, result.WithdrawalNo, "结算流程失败，自动取消")

	return nil
}

// executeUpdateSettlementCompleted 执行更新结算单为完成步骤
func (s *SettlementSagaService) executeUpdateSettlementCompleted(ctx context.Context, settlement *model.Settlement, withdrawalData string) (string, error) {
	logger.Info("executing update settlement to completed",
		zap.String("settlement_no", settlement.SettlementNo))

	// 更新结算单状态为已完成
	now := time.Now()
	settlement.Status = model.SettlementStatusCompleted
	settlement.CompletedAt = &now

	if err := s.settlementRepo.Update(ctx, settlement); err != nil {
		return "", fmt.Errorf("update settlement status failed: %w", err)
	}

	result := map[string]interface{}{
		"status":       string(settlement.Status),
		"completed_at": now.Format(time.RFC3339),
	}
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// compensateUpdateSettlementCompleted 补偿更新结算单为完成步骤
func (s *SettlementSagaService) compensateUpdateSettlementCompleted(ctx context.Context, settlement *model.Settlement) error {
	logger.Info("compensating update settlement completed",
		zap.String("settlement_no", settlement.SettlementNo))

	// 恢复结算单状态为失败
	settlement.Status = model.SettlementStatusFailed
	settlement.ErrorMessage = "Saga 补偿: 分布式事务回滚"
	settlement.CompletedAt = nil

	if err := s.settlementRepo.Update(ctx, settlement); err != nil {
		logger.Error("failed to restore settlement status during compensation",
			zap.String("settlement_no", settlement.SettlementNo),
			zap.Error(err))
		return fmt.Errorf("restore settlement status failed: %w", err)
	}

	return nil
}

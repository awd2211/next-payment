package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/idempotent"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/withdrawal-service/internal/client"
	"payment-platform/withdrawal-service/internal/model"
	"payment-platform/withdrawal-service/internal/repository"
)

// WithdrawalService 提现服务接口
type WithdrawalService interface {
	CreateWithdrawal(ctx context.Context, input *CreateWithdrawalInput) (*model.Withdrawal, error)
	GetWithdrawal(ctx context.Context, id uuid.UUID) (*WithdrawalDetail, error)
	ListWithdrawals(ctx context.Context, query *ListWithdrawalQuery) (*ListWithdrawalResponse, error)
	ApproveWithdrawal(ctx context.Context, withdrawalID, approverID uuid.UUID, approverName, comments string) error
	RejectWithdrawal(ctx context.Context, withdrawalID, approverID uuid.UUID, approverName, comments string) error
	ExecuteWithdrawal(ctx context.Context, withdrawalID uuid.UUID) error
	CancelWithdrawal(ctx context.Context, withdrawalID uuid.UUID, reason string) error
	GetWithdrawalReport(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*WithdrawalReport, error)

	// Bank Account management
	CreateBankAccount(ctx context.Context, input *CreateBankAccountInput) (*model.WithdrawalBankAccount, error)
	UpdateBankAccount(ctx context.Context, id uuid.UUID, input *UpdateBankAccountInput) error
	GetBankAccount(ctx context.Context, id uuid.UUID) (*model.WithdrawalBankAccount, error)
	ListBankAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.WithdrawalBankAccount, error)
	SetDefaultBankAccount(ctx context.Context, merchantID, accountID uuid.UUID) error
}

type withdrawalService struct {
	db                  *gorm.DB
	withdrawalRepo      repository.WithdrawalRepository
	accountingClient    *client.AccountingClient
	notificationClient  *client.NotificationClient
	bankTransferClient  *client.BankTransferClient
	sagaService         *WithdrawalSagaService // Saga 分布式事务服务
	redisClient         *redis.Client
	idempotentService   idempotent.Service
}

// NewWithdrawalService 创建提现服务
func NewWithdrawalService(
	db *gorm.DB,
	withdrawalRepo repository.WithdrawalRepository,
	accountingClient *client.AccountingClient,
	notificationClient *client.NotificationClient,
	bankTransferClient *client.BankTransferClient,
	redisClient *redis.Client,
) WithdrawalService {
	return &withdrawalService{
		db:                 db,
		withdrawalRepo:     withdrawalRepo,
		accountingClient:   accountingClient,
		notificationClient: notificationClient,
		bankTransferClient: bankTransferClient,
		sagaService:        nil, // 通过 SetSagaService 注入
		redisClient:        redisClient,
		idempotentService:  idempotent.NewService(redisClient),
	}
}

// SetSagaService 设置 Saga 服务（用于依赖注入）
func (s *withdrawalService) SetSagaService(sagaService *WithdrawalSagaService) {
	s.sagaService = sagaService
}

// CreateWithdrawalInput 创建提现输入
type CreateWithdrawalInput struct {
	MerchantID    uuid.UUID
	Amount        int64
	Type          model.WithdrawalType
	BankAccountID uuid.UUID
	Remarks       string
	CreatedBy     uuid.UUID
	RequestNo     string // 请求单号（可选，用于幂等性，通常由前端或上游服务生成）
}

// CreateWithdrawal 创建提现申请
func (s *withdrawalService) CreateWithdrawal(ctx context.Context, input *CreateWithdrawalInput) (*model.Withdrawal, error) {
	// 【幂等性保护】1. 如果提供了RequestNo，使用它作为幂等性键
	var lockAcquired bool
	if input.RequestNo != "" {
		// 幂等性键: withdrawal:{merchant_id}:{request_no}
		idempotentKey := idempotent.GenerateKey("withdrawal", input.MerchantID.String(), input.RequestNo)

		// 【幂等性保护】2. 检查是否已处理过该提现请求
		type WithdrawalIdempotentResult struct {
			WithdrawalNo string `json:"withdrawal_no"`
			WithdrawalID string `json:"withdrawal_id"`
			Status       string `json:"status"`
		}
		var cachedResult WithdrawalIdempotentResult
		exists, err := s.idempotentService.Check(ctx, idempotentKey, &cachedResult)
		if err != nil {
			logger.Warn("幂等性检查失败(Redis不可用)，继续处理", zap.Error(err))
		}
		if exists {
			logger.Info("提现单已存在(幂等性缓存命中)",
				zap.String("withdrawal_no", cachedResult.WithdrawalNo),
				zap.String("request_no", input.RequestNo))

			// 返回已存在的提现单
			withdrawalID, _ := uuid.Parse(cachedResult.WithdrawalID)
			withdrawal, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
			if err != nil {
				return nil, fmt.Errorf("获取已存在的提现单失败: %w", err)
			}
			return withdrawal, nil
		}

		// 【幂等性保护】3. 尝试获取分布式锁（30秒超时）
		lockAcquired, err = s.idempotentService.Try(ctx, idempotentKey, 30*time.Second)
		if err != nil {
			logger.Warn("获取分布式锁失败(Redis不可用)，继续处理", zap.Error(err))
		}
		if !lockAcquired {
			return nil, fmt.Errorf("该提现请求正在处理中，请稍后查询结果")
		}

		// 【幂等性保护】4. 确保释放锁
		defer func() {
			if lockAcquired {
				s.idempotentService.Release(ctx, idempotentKey)
			}
		}()
	}

	// 验证提现金额
	if input.Amount <= 0 {
		return nil, fmt.Errorf("提现金额必须大于0")
	}

	// 获取银行账户信息
	bankAccount, err := s.withdrawalRepo.GetBankAccountByID(ctx, input.BankAccountID)
	if err != nil {
		return nil, fmt.Errorf("银行账户不存在: %w", err)
	}

	if bankAccount.MerchantID != input.MerchantID {
		return nil, fmt.Errorf("银行账户不属于该商户")
	}

	if bankAccount.Status != "active" {
		return nil, fmt.Errorf("银行账户不可用")
	}

	// 查询商户可用余额（调用 accounting-service）
	if s.accountingClient != nil {
		availableBalance, err := s.accountingClient.GetAvailableBalance(ctx, input.MerchantID)
		if err != nil {
			return nil, fmt.Errorf("查询商户余额失败: %w", err)
		}
		if availableBalance < input.Amount {
			return nil, fmt.Errorf("余额不足，可用余额: %.2f元，提现金额: %.2f元",
				float64(availableBalance)/100, float64(input.Amount)/100)
		}
	}

	// 计算手续费
	fee := s.calculateFee(input.Amount, input.Type)
	actualAmount := input.Amount - fee

	// 确定审批级别
	requiredLevel := s.determineApprovalLevel(input.Amount)

	// 生成提现单号
	withdrawalNo := fmt.Sprintf("WD%s%d", input.MerchantID.String()[:8], time.Now().Unix())

	withdrawal := &model.Withdrawal{
		WithdrawalNo:    withdrawalNo,
		MerchantID:      input.MerchantID,
		Amount:          input.Amount,
		Fee:             fee,
		ActualAmount:    actualAmount,
		Type:            input.Type,
		Status:          model.WithdrawalStatusPending,
		BankAccountID:   input.BankAccountID,
		BankName:        bankAccount.BankName,
		BankAccountName: bankAccount.AccountName,
		BankAccountNo:   bankAccount.AccountNo,
		Remarks:         input.Remarks,
		ApprovalLevel:   0,
		RequiredLevel:   requiredLevel,
		CreatedBy:       input.CreatedBy,
	}

	// 创建提现记录
	if err := s.withdrawalRepo.Create(ctx, withdrawal); err != nil {
		return nil, fmt.Errorf("创建提现记录失败: %w", err)
	}

	// 【幂等性保护】5. 缓存成功结果（如果使用了幂等性键）
	if input.RequestNo != "" {
		idempotentKey := idempotent.GenerateKey("withdrawal", input.MerchantID.String(), input.RequestNo)

		cacheResult := map[string]interface{}{
			"withdrawal_no": withdrawal.WithdrawalNo,
			"withdrawal_id": withdrawal.ID.String(),
			"status":        withdrawal.Status,
		}
		if err := s.idempotentService.Store(ctx, idempotentKey, cacheResult, 24*time.Hour); err != nil {
			logger.Warn("缓存提现单结果失败(Redis不可用)，不影响业务", zap.Error(err))
		}
	}

	// 发送审批通知（调用 notification-service）
	if s.notificationClient != nil {
		if err := s.notificationClient.SendApprovalNotification(ctx, input.MerchantID, withdrawalNo, input.Amount); err != nil {
			// 通知发送失败不影响提现创建，仅记录日志
			logger.Error("failed to send approval notification",
				zap.Error(err),
				zap.String("merchant_id", input.MerchantID.String()),
				zap.String("withdrawal_no", withdrawalNo),
				zap.Int64("amount", input.Amount))
		}
	}

	return withdrawal, nil
}

// calculateFee 计算手续费
func (s *withdrawalService) calculateFee(amount int64, withdrawalType model.WithdrawalType) int64 {
	// 基础手续费率
	var feeRate float64
	switch withdrawalType {
	case model.WithdrawalTypeNormal:
		feeRate = 0.001 // 0.1%
	case model.WithdrawalTypeUrgent:
		feeRate = 0.005 // 0.5%
	case model.WithdrawalTypeScheduled:
		feeRate = 0.0005 // 0.05%
	default:
		feeRate = 0.001
	}

	fee := int64(float64(amount) * feeRate)

	// 最低手续费
	minFee := int64(100) // 1元
	if fee < minFee {
		fee = minFee
	}

	// 最高手续费
	maxFee := int64(10000) // 100元
	if fee > maxFee {
		fee = maxFee
	}

	return fee
}

// determineApprovalLevel 确定审批级别
func (s *withdrawalService) determineApprovalLevel(amount int64) int {
	if amount >= 100000000 { // >= 100万元
		return 3 // 需要三级审批
	} else if amount >= 10000000 { // >= 10万元
		return 2 // 需要二级审批
	}
	return 1 // 需要一级审批
}

// WithdrawalDetail 提现详情
type WithdrawalDetail struct {
	Withdrawal *model.Withdrawal          `json:"withdrawal"`
	Approvals  []*model.WithdrawalApproval `json:"approvals"`
}

// GetWithdrawal 获取提现详情
func (s *withdrawalService) GetWithdrawal(ctx context.Context, id uuid.UUID) (*WithdrawalDetail, error) {
	withdrawal, err := s.withdrawalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取提现记录失败: %w", err)
	}

	approvals, err := s.withdrawalRepo.GetApprovals(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取审批记录失败: %w", err)
	}

	return &WithdrawalDetail{
		Withdrawal: withdrawal,
		Approvals:  approvals,
	}, nil
}

// ListWithdrawalQuery 查询参数
type ListWithdrawalQuery struct {
	MerchantID *uuid.UUID
	Status     *model.WithdrawalStatus
	Type       *model.WithdrawalType
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

// ListWithdrawalResponse 列表响应
type ListWithdrawalResponse struct {
	Withdrawals []*model.Withdrawal `json:"withdrawals"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}

// ListWithdrawals 提现列表
func (s *withdrawalService) ListWithdrawals(ctx context.Context, query *ListWithdrawalQuery) (*ListWithdrawalResponse, error) {
	repoQuery := &repository.WithdrawalQuery{
		MerchantID: query.MerchantID,
		Status:     query.Status,
		Type:       query.Type,
		StartDate:  query.StartDate,
		EndDate:    query.EndDate,
		Page:       query.Page,
		PageSize:   query.PageSize,
	}

	withdrawals, total, err := s.withdrawalRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, fmt.Errorf("查询提现列表失败: %w", err)
	}

	return &ListWithdrawalResponse{
		Withdrawals: withdrawals,
		Total:       total,
		Page:        query.Page,
		PageSize:    query.PageSize,
	}, nil
}

// ApproveWithdrawal 审批通过提现
func (s *withdrawalService) ApproveWithdrawal(ctx context.Context, withdrawalID, approverID uuid.UUID, approverName, comments string) error {
	withdrawal, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return fmt.Errorf("获取提现记录失败: %w", err)
	}

	if withdrawal.Status != model.WithdrawalStatusPending {
		return fmt.Errorf("提现状态不是待审批，无法审批")
	}

	// 增加审批级别
	withdrawal.ApprovalLevel++

	now := time.Now()
	approval := &model.WithdrawalApproval{
		WithdrawalID: withdrawalID,
		ApproverID:   approverID,
		ApproverName: approverName,
		Level:        withdrawal.ApprovalLevel,
		Action:       "approve",
		Status:       model.WithdrawalStatusApproved,
		Comments:     comments,
		ApprovedAt:   now,
	}

	// 判断是否达到所需审批级别
	if withdrawal.ApprovalLevel >= withdrawal.RequiredLevel {
		withdrawal.Status = model.WithdrawalStatusApproved
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.withdrawalRepo.Update(ctx, withdrawal); err != nil {
			return fmt.Errorf("更新提现记录失败: %w", err)
		}
		if err := s.withdrawalRepo.CreateApproval(ctx, approval); err != nil {
			return fmt.Errorf("创建审批记录失败: %w", err)
		}
		return nil
	})
}

// RejectWithdrawal 拒绝提现
func (s *withdrawalService) RejectWithdrawal(ctx context.Context, withdrawalID, approverID uuid.UUID, approverName, comments string) error {
	withdrawal, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return fmt.Errorf("获取提现记录失败: %w", err)
	}

	if withdrawal.Status != model.WithdrawalStatusPending {
		return fmt.Errorf("提现状态不是待审批，无法拒绝")
	}

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusRejected

	approval := &model.WithdrawalApproval{
		WithdrawalID: withdrawalID,
		ApproverID:   approverID,
		ApproverName: approverName,
		Level:        withdrawal.ApprovalLevel + 1,
		Action:       "reject",
		Status:       model.WithdrawalStatusRejected,
		Comments:     comments,
		ApprovedAt:   now,
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.withdrawalRepo.Update(ctx, withdrawal); err != nil {
			return fmt.Errorf("更新提现记录失败: %w", err)
		}
		if err := s.withdrawalRepo.CreateApproval(ctx, approval); err != nil {
			return fmt.Errorf("创建审批记录失败: %w", err)
		}
		return nil
	})
}

// ExecuteWithdrawal 执行提现
func (s *withdrawalService) ExecuteWithdrawal(ctx context.Context, withdrawalID uuid.UUID) error {
	withdrawal, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return fmt.Errorf("获取提现记录失败: %w", err)
	}

	if withdrawal.Status != model.WithdrawalStatusApproved {
		return fmt.Errorf("提现状态不是已审批，无法执行")
	}

	// ========== 使用 Saga 分布式事务执行提现（生产级方案）==========
	// 如果启用了 Saga，使用分布式事务保证原子性
	if s.sagaService != nil {
		logger.Info("使用 Saga 分布式事务执行提现",
			zap.String("withdrawal_no", withdrawal.WithdrawalNo),
			zap.String("withdrawal_id", withdrawalID.String()))

		// 执行 Withdrawal Saga (4 步骤):
		// 1. 预冻结余额
		// 2. 银行转账
		// 3. 扣减余额
		// 4. 更新提现状态
		// 任何步骤失败会自动回滚所有已完成的步骤
		err := s.sagaService.ExecuteWithdrawalSaga(ctx, withdrawal)
		if err != nil {
			logger.Error("Withdrawal Saga 执行失败",
				zap.Error(err),
				zap.String("withdrawal_no", withdrawal.WithdrawalNo))

			// Saga 已经自动补偿，这里只需返回错误
			return fmt.Errorf("提现执行失败: %w", err)
		}

		logger.Info("Withdrawal Saga 执行成功",
			zap.String("withdrawal_no", withdrawal.WithdrawalNo))

		// 发送完成通知
		if s.notificationClient != nil {
			if err := s.notificationClient.SendWithdrawalStatusNotification(ctx, withdrawal.MerchantID, withdrawal.WithdrawalNo, "completed", withdrawal.Amount); err != nil {
				logger.Error("failed to send completion notification",
					zap.Error(err),
					zap.String("merchant_id", withdrawal.MerchantID.String()),
					zap.String("withdrawal_no", withdrawal.WithdrawalNo),
					zap.Int64("amount", withdrawal.Amount))
			}
		}

		return nil
	}

	// ========== 旧逻辑（向后兼容，如果未启用 Saga）==========
	logger.Warn("未启用 Saga 服务，使用传统方式执行提现（不推荐）",
		zap.String("withdrawal_no", withdrawal.WithdrawalNo))

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusProcessing
	withdrawal.ProcessedAt = &now

	if err := s.withdrawalRepo.Update(ctx, withdrawal); err != nil {
		return fmt.Errorf("更新提现状态失败: %w", err)
	}

	// 调用银行转账接口
	if s.bankTransferClient != nil {
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
			withdrawal.Status = model.WithdrawalStatusFailed
			withdrawal.FailureReason = err.Error()
			s.withdrawalRepo.Update(ctx, withdrawal)

			// 发送失败通知
			if s.notificationClient != nil {
				s.notificationClient.SendWithdrawalStatusNotification(ctx, withdrawal.MerchantID, withdrawal.WithdrawalNo, "failed", withdrawal.Amount)
			}
			return fmt.Errorf("银行转账失败: %w", err)
		}

		withdrawal.ChannelTradeNo = transferResp.ChannelTradeNo
	}

	// 调用 accounting-service 扣减余额
	if s.accountingClient != nil {
		deductReq := &client.DeductBalanceRequest{
			MerchantID:      withdrawal.MerchantID,
			Amount:          withdrawal.Amount, // 扣减总金额（包含手续费）
			TransactionType: "withdrawal",
			RelatedNo:       withdrawal.WithdrawalNo,
			Description:     fmt.Sprintf("提现: %s, 实际到账: %.2f元, 手续费: %.2f元",
				withdrawal.WithdrawalNo,
				float64(withdrawal.ActualAmount)/100,
				float64(withdrawal.Fee)/100),
		}

		if err := s.accountingClient.DeductBalance(ctx, deductReq); err != nil {
			// ⚠️ 余额扣减失败，但银行转账已完成，数据不一致！
			// 生产环境：应该使用上面的 Saga 方案自动回滚
			withdrawal.Status = model.WithdrawalStatusFailed
			withdrawal.FailureReason = fmt.Sprintf("余额扣减失败: %v (银行转账已完成，需要人工处理)", err)
			s.withdrawalRepo.Update(ctx, withdrawal)
			return fmt.Errorf("余额扣减失败: %w", err)
		}
	}

	// 标记为完成
	withdrawal.Status = model.WithdrawalStatusCompleted
	withdrawal.CompletedAt = &now

	if err := s.withdrawalRepo.Update(ctx, withdrawal); err != nil {
		return fmt.Errorf("更新提现状态失败: %w", err)
	}

	// 发送完成通知
	if s.notificationClient != nil {
		if err := s.notificationClient.SendWithdrawalStatusNotification(ctx, withdrawal.MerchantID, withdrawal.WithdrawalNo, "completed", withdrawal.Amount); err != nil {
			logger.Error("failed to send completion notification",
				zap.Error(err),
				zap.String("merchant_id", withdrawal.MerchantID.String()),
				zap.String("withdrawal_no", withdrawal.WithdrawalNo),
				zap.Int64("amount", withdrawal.Amount))
		}
	}

	return nil
}

// CancelWithdrawal 取消提现
func (s *withdrawalService) CancelWithdrawal(ctx context.Context, withdrawalID uuid.UUID, reason string) error {
	withdrawal, err := s.withdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return fmt.Errorf("获取提现记录失败: %w", err)
	}

	if withdrawal.Status != model.WithdrawalStatusPending && withdrawal.Status != model.WithdrawalStatusApproved {
		return fmt.Errorf("当前状态无法取消")
	}

	withdrawal.Status = model.WithdrawalStatusCancelled
	withdrawal.FailureReason = reason

	return s.withdrawalRepo.Update(ctx, withdrawal)
}

// WithdrawalReport 提现报表
type WithdrawalReport struct {
	TotalAmount      int64 `json:"total_amount"`
	TotalCount       int   `json:"total_count"`
	TotalFee         int64 `json:"total_fee"`
	CompletedCount   int   `json:"completed_count"`
	CompletedAmount  int64 `json:"completed_amount"`
	PendingCount     int   `json:"pending_count"`
	PendingAmount    int64 `json:"pending_amount"`
	RejectedCount    int   `json:"rejected_count"`
	FailedCount      int   `json:"failed_count"`
	AvgAmount        int64 `json:"avg_amount"`
	AvgProcessingTime int64 `json:"avg_processing_time"` // 秒
}

// GetWithdrawalReport 获取提现报表
func (s *withdrawalService) GetWithdrawalReport(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*WithdrawalReport, error) {
	query := &repository.WithdrawalQuery{
		MerchantID: &merchantID,
		StartDate:  &startDate,
		EndDate:    &endDate,
		Page:       1,
		PageSize:   10000, // 获取所有数据用于统计
	}

	withdrawals, _, err := s.withdrawalRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询提现记录失败: %w", err)
	}

	report := &WithdrawalReport{}
	var totalProcessingTime int64

	for _, withdrawal := range withdrawals {
		report.TotalAmount += withdrawal.Amount
		report.TotalCount++
		report.TotalFee += withdrawal.Fee

		switch withdrawal.Status {
		case model.WithdrawalStatusCompleted:
			report.CompletedCount++
			report.CompletedAmount += withdrawal.ActualAmount
			if withdrawal.ProcessedAt != nil && withdrawal.CompletedAt != nil {
				processingTime := withdrawal.CompletedAt.Sub(*withdrawal.ProcessedAt).Seconds()
				totalProcessingTime += int64(processingTime)
			}
		case model.WithdrawalStatusPending:
			report.PendingCount++
			report.PendingAmount += withdrawal.Amount
		case model.WithdrawalStatusRejected:
			report.RejectedCount++
		case model.WithdrawalStatusFailed:
			report.FailedCount++
		}
	}

	if report.TotalCount > 0 {
		report.AvgAmount = report.TotalAmount / int64(report.TotalCount)
	}

	if report.CompletedCount > 0 {
		report.AvgProcessingTime = totalProcessingTime / int64(report.CompletedCount)
	}

	return report, nil
}

// Bank Account Management

// CreateBankAccountInput 创建银行账户输入
type CreateBankAccountInput struct {
	MerchantID      uuid.UUID
	BankName        string
	BankCode        string
	BankBranch      string
	AccountName     string
	AccountNo       string
	AccountType     string
	IsDefault       bool
	VerificationDoc string
}

// CreateBankAccount 创建银行账户（使用事务保证默认账户唯一性）
func (s *withdrawalService) CreateBankAccount(ctx context.Context, input *CreateBankAccountInput) (*model.WithdrawalBankAccount, error) {
	// 准备账户数据
	account := &model.WithdrawalBankAccount{
		MerchantID:      input.MerchantID,
		BankName:        input.BankName,
		BankCode:        input.BankCode,
		BankBranch:      input.BankBranch,
		AccountName:     input.AccountName,
		AccountNo:       input.AccountNo,
		AccountType:     input.AccountType,
		IsDefault:       input.IsDefault,
		IsVerified:      false,
		VerificationDoc: input.VerificationDoc,
		Status:          "active",
	}

	// 在事务中处理默认账户设置和创建
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 如果设置为默认账户，先取消其他默认账户
		if input.IsDefault {
			err := tx.Model(&model.WithdrawalBankAccount{}).
				Where("merchant_id = ? AND is_default = true", input.MerchantID).
				Update("is_default", false).Error
			if err != nil {
				return fmt.Errorf("取消其他默认账户失败: %w", err)
			}
		}

		// 2. 创建新账户
		if err := tx.Create(account).Error; err != nil {
			return fmt.Errorf("创建银行账户失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return account, nil
}

// UpdateBankAccountInput 更新银行账户输入
type UpdateBankAccountInput struct {
	BankBranch      *string
	IsDefault       *bool
	Status          *string
	VerificationDoc *string
	IsVerified      *bool
}

// UpdateBankAccount 更新银行账户
func (s *withdrawalService) UpdateBankAccount(ctx context.Context, id uuid.UUID, input *UpdateBankAccountInput) error {
	account, err := s.withdrawalRepo.GetBankAccountByID(ctx, id)
	if err != nil {
		return fmt.Errorf("银行账户不存在: %w", err)
	}

	if input.BankBranch != nil {
		account.BankBranch = *input.BankBranch
	}
	if input.IsDefault != nil {
		account.IsDefault = *input.IsDefault
		// 如果设置为默认，取消其他默认账户
		if *input.IsDefault {
			accounts, _ := s.withdrawalRepo.ListBankAccounts(ctx, account.MerchantID)
			for _, acc := range accounts {
				if acc.ID != id && acc.IsDefault {
					acc.IsDefault = false
					s.withdrawalRepo.UpdateBankAccount(ctx, acc)
				}
			}
		}
	}
	if input.Status != nil {
		account.Status = *input.Status
	}
	if input.VerificationDoc != nil {
		account.VerificationDoc = *input.VerificationDoc
	}
	if input.IsVerified != nil {
		account.IsVerified = *input.IsVerified
	}

	return s.withdrawalRepo.UpdateBankAccount(ctx, account)
}

// GetBankAccount 获取银行账户
func (s *withdrawalService) GetBankAccount(ctx context.Context, id uuid.UUID) (*model.WithdrawalBankAccount, error) {
	return s.withdrawalRepo.GetBankAccountByID(ctx, id)
}

// ListBankAccounts 银行账户列表
func (s *withdrawalService) ListBankAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.WithdrawalBankAccount, error) {
	return s.withdrawalRepo.ListBankAccounts(ctx, merchantID)
}

// SetDefaultBankAccount 设置默认银行账户
func (s *withdrawalService) SetDefaultBankAccount(ctx context.Context, merchantID, accountID uuid.UUID) error {
	account, err := s.withdrawalRepo.GetBankAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("银行账户不存在: %w", err)
	}

	if account.MerchantID != merchantID {
		return fmt.Errorf("银行账户不属于该商户")
	}

	// 取消其他默认账户
	accounts, _ := s.withdrawalRepo.ListBankAccounts(ctx, merchantID)
	for _, acc := range accounts {
		if acc.IsDefault {
			acc.IsDefault = false
			s.withdrawalRepo.UpdateBankAccount(ctx, acc)
		}
	}

	account.IsDefault = true
	return s.withdrawalRepo.UpdateBankAccount(ctx, account)
}

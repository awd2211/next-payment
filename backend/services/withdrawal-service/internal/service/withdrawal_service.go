package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	db             *gorm.DB
	withdrawalRepo repository.WithdrawalRepository
}

// NewWithdrawalService 创建提现服务
func NewWithdrawalService(db *gorm.DB, withdrawalRepo repository.WithdrawalRepository) WithdrawalService {
	return &withdrawalService{
		db:             db,
		withdrawalRepo: withdrawalRepo,
	}
}

// CreateWithdrawalInput 创建提现输入
type CreateWithdrawalInput struct {
	MerchantID    uuid.UUID
	Amount        int64
	Type          model.WithdrawalType
	BankAccountID uuid.UUID
	Remarks       string
	CreatedBy     uuid.UUID
}

// CreateWithdrawal 创建提现申请
func (s *withdrawalService) CreateWithdrawal(ctx context.Context, input *CreateWithdrawalInput) (*model.Withdrawal, error) {
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

	// TODO: 查询商户可用余额（调用 accounting-service）
	// availableBalance := getAvailableBalance(merchantID)
	// if availableBalance < amount {
	//     return nil, fmt.Errorf("余额不足")
	// }

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

	// TODO: 发送审批通知（调用 notification-service）

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

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusProcessing
	withdrawal.ProcessedAt = &now

	if err := s.withdrawalRepo.Update(ctx, withdrawal); err != nil {
		return fmt.Errorf("更新提现状态失败: %w", err)
	}

	// TODO: 调用银行转账接口
	// channelTradeNo, err := bankTransferService.Transfer(withdrawal)
	// if err != nil {
	//     withdrawal.Status = model.WithdrawalStatusFailed
	//     withdrawal.FailureReason = err.Error()
	//     s.withdrawalRepo.Update(ctx, withdrawal)
	//     return err
	// }
	// withdrawal.ChannelTradeNo = channelTradeNo

	// TODO: 调用 accounting-service 扣减余额

	// 标记为完成
	withdrawal.Status = model.WithdrawalStatusCompleted
	withdrawal.CompletedAt = &now

	return s.withdrawalRepo.Update(ctx, withdrawal)
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

// CreateBankAccount 创建银行账户
func (s *withdrawalService) CreateBankAccount(ctx context.Context, input *CreateBankAccountInput) (*model.WithdrawalBankAccount, error) {
	// 如果设置为默认账户，先取消其他默认账户
	if input.IsDefault {
		accounts, _ := s.withdrawalRepo.ListBankAccounts(ctx, input.MerchantID)
		for _, acc := range accounts {
			if acc.IsDefault {
				acc.IsDefault = false
				s.withdrawalRepo.UpdateBankAccount(ctx, acc)
			}
		}
	}

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

	if err := s.withdrawalRepo.CreateBankAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("创建银行账户失败: %w", err)
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

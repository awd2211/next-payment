package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/settlement-service/internal/client"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/repository"
)

// SettlementService 结算服务接口
type SettlementService interface {
	CreateSettlement(ctx context.Context, input *CreateSettlementInput) (*model.Settlement, error)
	GetSettlement(ctx context.Context, id uuid.UUID) (*SettlementDetail, error)
	ListSettlements(ctx context.Context, query *ListSettlementQuery) (*ListSettlementResponse, error)
	ApproveSettlement(ctx context.Context, settlementID, approverID uuid.UUID, approverName, comments string) error
	RejectSettlement(ctx context.Context, settlementID, approverID uuid.UUID, approverName, comments string) error
	ExecuteSettlement(ctx context.Context, settlementID uuid.UUID) error
	GenerateAutoSettlement(ctx context.Context, merchantID uuid.UUID, cycle model.SettlementCycle) (*model.Settlement, error)
	GetSettlementReport(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*SettlementReport, error)
}

type settlementService struct {
	db                 *gorm.DB
	settlementRepo     repository.SettlementRepository
	accountingClient   *client.AccountingClient
	withdrawalClient   *client.WithdrawalClient
	merchantClient     *client.MerchantClient
	notificationClient *client.NotificationClient
}

// NewSettlementService 创建结算服务
func NewSettlementService(
	db *gorm.DB,
	settlementRepo repository.SettlementRepository,
	accountingClient *client.AccountingClient,
	withdrawalClient *client.WithdrawalClient,
	merchantClient *client.MerchantClient,
	notificationClient *client.NotificationClient,
) SettlementService {
	return &settlementService{
		db:                 db,
		settlementRepo:     settlementRepo,
		accountingClient:   accountingClient,
		withdrawalClient:   withdrawalClient,
		merchantClient:     merchantClient,
		notificationClient: notificationClient,
	}
}

// CreateSettlementInput 创建结算单输入
type CreateSettlementInput struct {
	MerchantID   uuid.UUID
	Cycle        model.SettlementCycle
	StartDate    time.Time
	EndDate      time.Time
	Transactions []TransactionItem
}

// TransactionItem 交易明细
type TransactionItem struct {
	TransactionID uuid.UUID
	OrderNo       string
	PaymentNo     string
	Amount        int64
	Fee           int64
	TransactionAt time.Time
}

// CreateSettlement 创建结算单
func (s *settlementService) CreateSettlement(ctx context.Context, input *CreateSettlementInput) (*model.Settlement, error) {
	// 生成结算单号
	settlementNo := fmt.Sprintf("STL%s%d", input.MerchantID.String()[:8], time.Now().Unix())

	// 计算结算金额
	var totalAmount int64
	var totalFee int64
	var totalCount int
	items := make([]*model.SettlementItem, 0, len(input.Transactions))

	for _, tx := range input.Transactions {
		totalAmount += tx.Amount
		totalFee += tx.Fee
		totalCount++

		item := &model.SettlementItem{
			TransactionID: tx.TransactionID,
			OrderNo:       tx.OrderNo,
			PaymentNo:     tx.PaymentNo,
			Amount:        tx.Amount,
			Fee:           tx.Fee,
			SettleAmount:  tx.Amount - tx.Fee,
			TransactionAt: tx.TransactionAt,
		}
		items = append(items, item)
	}

	settlement := &model.Settlement{
		SettlementNo:     settlementNo,
		MerchantID:       input.MerchantID,
		Cycle:            input.Cycle,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		TotalAmount:      totalAmount,
		TotalCount:       totalCount,
		FeeAmount:        totalFee,
		SettlementAmount: totalAmount - totalFee,
		Status:           model.SettlementStatusPending,
	}

	// 使用事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建结算单
		if err := s.settlementRepo.Create(ctx, settlement); err != nil {
			return fmt.Errorf("创建结算单失败: %w", err)
		}

		// 创建明细
		for _, item := range items {
			item.SettlementID = settlement.ID
		}
		if err := s.settlementRepo.CreateItems(ctx, items); err != nil {
			return fmt.Errorf("创建结算明细失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return settlement, nil
}

// SettlementDetail 结算单详情
type SettlementDetail struct {
	Settlement *model.Settlement          `json:"settlement"`
	Items      []*model.SettlementItem    `json:"items"`
	Approvals  []*model.SettlementApproval `json:"approvals"`
}

// GetSettlement 获取结算单详情
func (s *settlementService) GetSettlement(ctx context.Context, id uuid.UUID) (*SettlementDetail, error) {
	settlement, err := s.settlementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取结算单失败: %w", err)
	}

	items, err := s.settlementRepo.GetItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取结算明细失败: %w", err)
	}

	approvals, err := s.settlementRepo.GetApprovals(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取审批记录失败: %w", err)
	}

	return &SettlementDetail{
		Settlement: settlement,
		Items:      items,
		Approvals:  approvals,
	}, nil
}

// ListSettlementQuery 查询参数
type ListSettlementQuery struct {
	MerchantID *uuid.UUID
	Status     *model.SettlementStatus
	Cycle      *model.SettlementCycle
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

// ListSettlementResponse 列表响应
type ListSettlementResponse struct {
	Settlements []*model.Settlement `json:"settlements"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}

// ListSettlements 结算单列表
func (s *settlementService) ListSettlements(ctx context.Context, query *ListSettlementQuery) (*ListSettlementResponse, error) {
	repoQuery := &repository.SettlementQuery{
		MerchantID: query.MerchantID,
		Status:     query.Status,
		Cycle:      query.Cycle,
		StartDate:  query.StartDate,
		EndDate:    query.EndDate,
		Page:       query.Page,
		PageSize:   query.PageSize,
	}

	settlements, total, err := s.settlementRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, fmt.Errorf("查询结算单失败: %w", err)
	}

	return &ListSettlementResponse{
		Settlements: settlements,
		Total:       total,
		Page:        query.Page,
		PageSize:    query.PageSize,
	}, nil
}

// ApproveSettlement 审批通过结算单
func (s *settlementService) ApproveSettlement(ctx context.Context, settlementID, approverID uuid.UUID, approverName, comments string) error {
	settlement, err := s.settlementRepo.GetByID(ctx, settlementID)
	if err != nil {
		return fmt.Errorf("获取结算单失败: %w", err)
	}

	if settlement.Status != model.SettlementStatusPending {
		return fmt.Errorf("结算单状态不是待审批，无法审批")
	}

	now := time.Now()
	settlement.Status = model.SettlementStatusApproved
	settlement.ApprovedAt = &now
	settlement.ApprovedBy = &approverID

	approval := &model.SettlementApproval{
		SettlementID: settlementID,
		ApproverID:   approverID,
		ApproverName: approverName,
		Action:       "approve",
		Status:       model.SettlementStatusApproved,
		Comments:     comments,
		ApprovedAt:   now,
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.settlementRepo.Update(ctx, settlement); err != nil {
			return fmt.Errorf("更新结算单失败: %w", err)
		}
		if err := s.settlementRepo.CreateApproval(ctx, approval); err != nil {
			return fmt.Errorf("创建审批记录失败: %w", err)
		}
		return nil
	})
}

// RejectSettlement 拒绝结算单
func (s *settlementService) RejectSettlement(ctx context.Context, settlementID, approverID uuid.UUID, approverName, comments string) error {
	settlement, err := s.settlementRepo.GetByID(ctx, settlementID)
	if err != nil {
		return fmt.Errorf("获取结算单失败: %w", err)
	}

	if settlement.Status != model.SettlementStatusPending {
		return fmt.Errorf("结算单状态不是待审批，无法拒绝")
	}

	now := time.Now()
	settlement.Status = model.SettlementStatusRejected
	settlement.ApprovedAt = &now
	settlement.ApprovedBy = &approverID

	approval := &model.SettlementApproval{
		SettlementID: settlementID,
		ApproverID:   approverID,
		ApproverName: approverName,
		Action:       "reject",
		Status:       model.SettlementStatusRejected,
		Comments:     comments,
		ApprovedAt:   now,
	}

	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.settlementRepo.Update(ctx, settlement); err != nil {
			return fmt.Errorf("更新结算单失败: %w", err)
		}
		if err := s.settlementRepo.CreateApproval(ctx, approval); err != nil {
			return fmt.Errorf("创建审批记录失败: %w", err)
		}
		return nil
	})
}

// ExecuteSettlement 执行结算
func (s *settlementService) ExecuteSettlement(ctx context.Context, settlementID uuid.UUID) error {
	settlement, err := s.settlementRepo.GetByID(ctx, settlementID)
	if err != nil {
		return fmt.Errorf("获取结算单失败: %w", err)
	}

	if settlement.Status != model.SettlementStatusApproved {
		return fmt.Errorf("结算单状态不是已审批，无法执行")
	}

	now := time.Now()
	settlement.Status = model.SettlementStatusProcessing
	settlement.ProcessedAt = &now

	if err := s.settlementRepo.Update(ctx, settlement); err != nil {
		return fmt.Errorf("更新结算单状态失败: %w", err)
	}

	// 实际转账逻辑：调用 withdrawal-service 创建提现
	if s.withdrawalClient != nil && s.merchantClient != nil {
		// 从merchant-service获取商户默认银行账户
		defaultAccount, err := s.merchantClient.GetDefaultSettlementAccount(ctx, settlement.MerchantID)
		if err != nil {
			settlement.Status = model.SettlementStatusFailed
			settlement.ErrorMessage = fmt.Sprintf("获取默认结算账户失败: %v", err)
			s.settlementRepo.Update(ctx, settlement)
			return fmt.Errorf("获取默认结算账户失败: %w", err)
		}

		withdrawalReq := &client.CreateWithdrawalRequest{
			MerchantID:    settlement.MerchantID,
			Amount:        settlement.SettlementAmount,
			Type:          "settlement_auto",
			BankAccountID: defaultAccount.ID,
			Remarks:       fmt.Sprintf("自动结算: %s, 周期: %s", settlement.SettlementNo, settlement.Cycle),
			CreatedBy:     uuid.MustParse("00000000-0000-0000-0000-000000000000"), // 系统自动
		}

		withdrawalNo, err := s.withdrawalClient.CreateWithdrawalForSettlement(ctx, withdrawalReq)
		if err != nil {
			settlement.Status = model.SettlementStatusFailed
			settlement.ErrorMessage = fmt.Sprintf("创建提现失败: %v", err)
			s.settlementRepo.Update(ctx, settlement)
			return fmt.Errorf("创建提现失败: %w", err)
		}

		settlement.WithdrawalNo = withdrawalNo
	}

	// 标记为完成
	settlement.Status = model.SettlementStatusCompleted
	settlement.CompletedAt = &now

	if err := s.settlementRepo.Update(ctx, settlement); err != nil {
		return err
	}

	// 发送结算完成通知
	if s.notificationClient != nil {
		go func(sett *model.Settlement) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			notifType := "settlement_complete"
			if sett.Status == model.SettlementStatusFailed {
				notifType = "settlement_failed"
			}

			title := "结算完成"
			content := fmt.Sprintf("结算单号 %s 已完成，结算金额 %.2f 元，已创建提现单 %s",
				sett.SettlementNo, float64(sett.SettlementAmount)/100.0, sett.WithdrawalNo)

			if sett.Status == model.SettlementStatusFailed {
				title = "结算失败"
				content = fmt.Sprintf("结算单号 %s 执行失败：%s", sett.SettlementNo, sett.ErrorMessage)
			}

			s.notificationClient.SendSettlementNotification(notifyCtx, &client.SendNotificationRequest{
				MerchantID: sett.MerchantID,
				Type:       notifType,
				Title:      title,
				Content:    content,
				Priority:   "high",
				Data: map[string]interface{}{
					"settlement_no":    sett.SettlementNo,
					"settlement_amount": sett.SettlementAmount,
					"withdrawal_no":    sett.WithdrawalNo,
					"cycle":            sett.Cycle,
					"status":           sett.Status,
				},
			})
		}(settlement)
	}

	return nil
}

// GenerateAutoSettlement 自动生成结算单
func (s *settlementService) GenerateAutoSettlement(ctx context.Context, merchantID uuid.UUID, cycle model.SettlementCycle) (*model.Settlement, error) {
	// 计算日期范围
	now := time.Now()
	var startDate, endDate time.Time

	switch cycle {
	case model.SettlementCycleDaily:
		startDate = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
		endDate = time.Date(now.Year(), now.Month(), now.Day()-1, 23, 59, 59, 0, now.Location())
	case model.SettlementCycleWeekly:
		weekday := now.Weekday()
		startDate = now.AddDate(0, 0, -int(weekday)-7).Truncate(24 * time.Hour)
		endDate = startDate.AddDate(0, 0, 7).Add(-time.Second)
	case model.SettlementCycleMonthly:
		startDate = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		endDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Add(-time.Second)
	default:
		return nil, fmt.Errorf("不支持的结算周期: %s", cycle)
	}

	// 从 accounting-service 获取交易数据
	var transactions []TransactionItem
	if s.accountingClient != nil {
		txList, err := s.accountingClient.GetTransactions(ctx, merchantID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("获取交易数据失败: %w", err)
		}

		transactions = make([]TransactionItem, 0, len(txList))
		for _, tx := range txList {
			// 解析交易时间
			transactionAt, _ := time.Parse("2006-01-02 15:04:05", tx.TransactionAt)
			if transactionAt.IsZero() {
				transactionAt = now
			}

			// 解析UUID
			txID, err := uuid.Parse(tx.ID)
			if err != nil {
				continue // 跳过无效的交易记录
			}

			transactions = append(transactions, TransactionItem{
				TransactionID: txID,
				OrderNo:       tx.OrderNo,
				PaymentNo:     tx.PaymentNo,
				Amount:        tx.Amount,
				Fee:           tx.Fee,
				TransactionAt: transactionAt,
			})
		}
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("没有可结算的交易数据")
	}

	input := &CreateSettlementInput{
		MerchantID:   merchantID,
		Cycle:        cycle,
		StartDate:    startDate,
		EndDate:      endDate,
		Transactions: transactions,
	}

	return s.CreateSettlement(ctx, input)
}

// SettlementReport 结算报表
type SettlementReport struct {
	TotalAmount       int64 `json:"total_amount"`
	TotalCount        int   `json:"total_count"`
	TotalFee          int64 `json:"total_fee"`
	TotalSettlement   int64 `json:"total_settlement"`
	CompletedCount    int   `json:"completed_count"`
	PendingCount      int   `json:"pending_count"`
	RejectedCount     int   `json:"rejected_count"`
	AvgSettlementAmount int64 `json:"avg_settlement_amount"`
}

// GetSettlementReport 获取结算报表
func (s *settlementService) GetSettlementReport(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*SettlementReport, error) {
	query := &repository.SettlementQuery{
		MerchantID: &merchantID,
		StartDate:  &startDate,
		EndDate:    &endDate,
		Page:       1,
		PageSize:   1000, // 获取所有数据用于统计
	}

	settlements, _, err := s.settlementRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询结算单失败: %w", err)
	}

	report := &SettlementReport{}
	for _, settlement := range settlements {
		report.TotalAmount += settlement.TotalAmount
		report.TotalCount += settlement.TotalCount
		report.TotalFee += settlement.FeeAmount
		report.TotalSettlement += settlement.SettlementAmount

		switch settlement.Status {
		case model.SettlementStatusCompleted:
			report.CompletedCount++
		case model.SettlementStatusPending:
			report.PendingCount++
		case model.SettlementStatusRejected:
			report.RejectedCount++
		}
	}

	if report.CompletedCount > 0 {
		report.AvgSettlementAmount = report.TotalSettlement / int64(report.CompletedCount)
	}

	return report, nil
}

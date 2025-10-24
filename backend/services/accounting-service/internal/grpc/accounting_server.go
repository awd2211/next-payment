package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/accounting"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/accounting-service/internal/repository"
	"payment-platform/accounting-service/internal/service"
)

// AccountingServer gRPC服务实现
type AccountingServer struct {
	pb.UnimplementedAccountingServiceServer
	accountService service.AccountService
}

// NewAccountingServer 创建gRPC服务实例
func NewAccountingServer(accountService service.AccountService) *AccountingServer {
	return &AccountingServer{
		accountService: accountService,
	}
}

// CreateEntry 创建账目记录（对应AccountTransaction）
func (s *AccountingServer) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.EntryResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	// 获取或创建账户
	account, err := s.accountService.GetMerchantAccount(ctx, merchantID, "operating", req.Currency)
	if err != nil {
		// 如果账户不存在，创建一个
		createAccountInput := &service.CreateAccountInput{
			MerchantID:  merchantID,
			AccountType: "operating",
			Currency:    req.Currency,
		}
		account, err = s.accountService.CreateAccount(ctx, createAccountInput)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "创建账户失败: %v", err)
		}
	}

	// 解析RelatedID（如果有）
	var relatedID uuid.UUID
	if req.RelatedId != "" {
		relatedID, err = uuid.Parse(req.RelatedId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的关联ID")
		}
	}

	// 创建交易
	input := &service.CreateTransactionInput{
		AccountID:       account.ID,
		TransactionType: req.EntryType,
		Amount:          req.Amount,
		RelatedID:       relatedID,
		RelatedNo:       req.RelatedType,
		Description:     req.Description,
	}

	transaction, err := s.accountService.CreateTransaction(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建账目记录失败: %v", err)
	}

	return &pb.EntryResponse{
		Entry: &pb.Entry{
			Id:          transaction.ID.String(),
			MerchantId:  transaction.MerchantID.String(),
			RelatedId:   transaction.RelatedID.String(),
			RelatedType: transaction.RelatedNo,
			EntryType:   transaction.TransactionType,
			Amount:      transaction.Amount,
			Currency:    transaction.Currency,
			Balance:     transaction.BalanceAfter,
			Description: transaction.Description,
			CreatedAt:   timestamppb.New(transaction.CreatedAt),
		},
	}, nil
}

// GetEntry 获取账目记录
func (s *AccountingServer) GetEntry(ctx context.Context, req *pb.GetEntryRequest) (*pb.EntryResponse, error) {
	entryID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的记录ID")
	}

	// 通过ID查询交易（需要通过repository直接查询）
	// 这里简化处理，使用ListTransactions查询
	query := &repository.TransactionQuery{
		Page:     1,
		PageSize: 1,
	}

	transactions, _, err := s.accountService.ListTransactions(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	// 查找匹配的交易
	for _, tx := range transactions {
		if tx.ID == entryID {
			return &pb.EntryResponse{
				Entry: &pb.Entry{
					Id:          tx.ID.String(),
					MerchantId:  tx.MerchantID.String(),
					RelatedId:   tx.RelatedID.String(),
					RelatedType: tx.RelatedNo,
					EntryType:   tx.TransactionType,
					Amount:      tx.Amount,
					Currency:    tx.Currency,
					Balance:     tx.BalanceAfter,
					Description: tx.Description,
					CreatedAt:   timestamppb.New(tx.CreatedAt),
				},
			}, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "账目记录不存在")
}

// ListEntries 获取账目记录列表
func (s *AccountingServer) ListEntries(ctx context.Context, req *pb.ListEntriesRequest) (*pb.ListEntriesResponse, error) {
	query := &repository.TransactionQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.MerchantId != "" {
		merchantID, err := uuid.Parse(req.MerchantId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
		}
		query.MerchantID = &merchantID
	}

	if req.EntryType != "" {
		query.TransactionType = req.EntryType
	}

	if req.StartTime != nil {
		startTime := req.StartTime.AsTime()
		query.StartTime = &startTime
	}

	if req.EndTime != nil {
		endTime := req.EndTime.AsTime()
		query.EndTime = &endTime
	}

	transactions, total, err := s.accountService.ListTransactions(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	entries := make([]*pb.Entry, len(transactions))
	for i, tx := range transactions {
		entries[i] = &pb.Entry{
			Id:          tx.ID.String(),
			MerchantId:  tx.MerchantID.String(),
			RelatedId:   tx.RelatedID.String(),
			RelatedType: tx.RelatedNo,
			EntryType:   tx.TransactionType,
			Amount:      tx.Amount,
			Currency:    tx.Currency,
			Balance:     tx.BalanceAfter,
			Description: tx.Description,
			CreatedAt:   timestamppb.New(tx.CreatedAt),
		}
	}

	return &pb.ListEntriesResponse{
		Entries: entries,
		Total:   total,
	}, nil
}

// CreateSettlement 创建结算
func (s *AccountingServer) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.SettlementResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	// 解析日期
	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的开始日期格式")
	}

	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的结束日期格式")
	}

	// 获取账户（假设使用operating账户，USD货币）
	account, err := s.accountService.GetMerchantAccount(ctx, merchantID, "operating", "USD")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取账户失败: %v", err)
	}

	input := &service.CreateSettlementInput{
		MerchantID:  merchantID,
		AccountID:   account.ID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Currency:    "USD",
	}

	settlement, err := s.accountService.CreateSettlement(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建结算失败: %v", err)
	}

	var settledAt *timestamppb.Timestamp
	if settlement.SettledAt != nil {
		settledAt = timestamppb.New(*settlement.SettledAt)
	}

	return &pb.SettlementResponse{
		Settlement: &pb.Settlement{
			Id:          settlement.ID.String(),
			MerchantId:  settlement.MerchantID.String(),
			Amount:      settlement.TotalAmount,
			Currency:    settlement.Currency,
			Status:      settlement.Status,
			PeriodStart: settlement.PeriodStart.Format("2006-01-02"),
			PeriodEnd:   settlement.PeriodEnd.Format("2006-01-02"),
			SettledAt:   settledAt,
			CreatedAt:   timestamppb.New(settlement.CreatedAt),
		},
	}, nil
}

// GetSettlement 获取结算
func (s *AccountingServer) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.SettlementResponse, error) {
	// 假设req.Id是SettlementNo
	settlement, err := s.accountService.GetSettlement(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "结算记录不存在: %v", err)
	}

	var settledAt *timestamppb.Timestamp
	if settlement.SettledAt != nil {
		settledAt = timestamppb.New(*settlement.SettledAt)
	}

	return &pb.SettlementResponse{
		Settlement: &pb.Settlement{
			Id:          settlement.ID.String(),
			MerchantId:  settlement.MerchantID.String(),
			Amount:      settlement.TotalAmount,
			Currency:    settlement.Currency,
			Status:      settlement.Status,
			PeriodStart: settlement.PeriodStart.Format("2006-01-02"),
			PeriodEnd:   settlement.PeriodEnd.Format("2006-01-02"),
			SettledAt:   settledAt,
			CreatedAt:   timestamppb.New(settlement.CreatedAt),
		},
	}, nil
}

// ListSettlements 获取结算列表
func (s *AccountingServer) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	query := &repository.SettlementQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.MerchantId != "" {
		merchantID, err := uuid.Parse(req.MerchantId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
		}
		query.MerchantID = &merchantID
	}

	if req.Status != "" {
		query.Status = req.Status
	}

	settlements, total, err := s.accountService.ListSettlements(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	pbSettlements := make([]*pb.Settlement, len(settlements))
	for i, st := range settlements {
		var settledAt *timestamppb.Timestamp
		if st.SettledAt != nil {
			settledAt = timestamppb.New(*st.SettledAt)
		}

		pbSettlements[i] = &pb.Settlement{
			Id:          st.ID.String(),
			MerchantId:  st.MerchantID.String(),
			Amount:      st.TotalAmount,
			Currency:    st.Currency,
			Status:      st.Status,
			PeriodStart: st.PeriodStart.Format("2006-01-02"),
			PeriodEnd:   st.PeriodEnd.Format("2006-01-02"),
			SettledAt:   settledAt,
			CreatedAt:   timestamppb.New(st.CreatedAt),
		}
	}

	return &pb.ListSettlementsResponse{
		Settlements: pbSettlements,
		Total:       total,
	}, nil
}

// UpdateSettlementStatus 更新结算状态
func (s *AccountingServer) UpdateSettlementStatus(ctx context.Context, req *pb.UpdateSettlementStatusRequest) (*pb.SettlementResponse, error) {
	// 根据状态执行不同的操作
	if req.Status == "processing" {
		err := s.accountService.ProcessSettlement(ctx, req.Id)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "处理结算失败: %v", err)
		}
	}

	// 获取更新后的结算
	settlement, err := s.accountService.GetSettlement(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "结算记录不存在: %v", err)
	}

	var settledAt *timestamppb.Timestamp
	if settlement.SettledAt != nil {
		settledAt = timestamppb.New(*settlement.SettledAt)
	}

	return &pb.SettlementResponse{
		Settlement: &pb.Settlement{
			Id:          settlement.ID.String(),
			MerchantId:  settlement.MerchantID.String(),
			Amount:      settlement.TotalAmount,
			Currency:    settlement.Currency,
			Status:      settlement.Status,
			PeriodStart: settlement.PeriodStart.Format("2006-01-02"),
			PeriodEnd:   settlement.PeriodEnd.Format("2006-01-02"),
			SettledAt:   settledAt,
			CreatedAt:   timestamppb.New(settlement.CreatedAt),
		},
	}, nil
}

// GetMerchantBalance 获取商户余额
func (s *AccountingServer) GetMerchantBalance(ctx context.Context, req *pb.GetMerchantBalanceRequest) (*pb.MerchantBalanceResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	// 获取商户余额汇总
	summary, err := s.accountService.GetMerchantBalanceSummary(ctx, merchantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取余额失败: %v", err)
	}

	// 计算不同类型的余额
	var availableBalance, pendingBalance, frozenBalance int64
	currency := "USD" // 默认货币

	for _, acc := range summary.Accounts {
		if acc.Status == "active" {
			availableBalance += acc.AvailableBalance
		}
		frozenBalance += acc.FrozenBalance
		if acc.Currency != "" {
			currency = acc.Currency
		}
	}

	return &pb.MerchantBalanceResponse{
		MerchantId:       req.MerchantId,
		AvailableBalance: availableBalance,
		PendingBalance:   pendingBalance,
		FrozenBalance:    frozenBalance,
		Currency:         currency,
	}, nil
}

// GenerateBill 生成账单
func (s *AccountingServer) GenerateBill(ctx context.Context, req *pb.GenerateBillRequest) (*pb.BillResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	// 解析账期（格式：YYYY-MM）
	periodStart, err := time.Parse("2006-01", req.Period)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的账期格式，应为YYYY-MM")
	}

	// 计算账期结束时间（下个月的第一天）
	periodEnd := periodStart.AddDate(0, 1, 0)
	dueDate := periodEnd.AddDate(0, 0, 15) // 账期结束后15天为到期日

	// 创建简单的账单（暂不包含明细）
	input := &service.CreateInvoiceInput{
		MerchantID:  merchantID,
		InvoiceType: "monthly",
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Currency:    "USD",
		DueDate:     dueDate,
		Items: []service.InvoiceItemInput{
			{
				ItemType:    "service_fee",
				Description: "月度服务费",
				Quantity:    1,
				UnitPrice:   10000, // 100.00 USD
			},
		},
	}

	invoice, err := s.accountService.CreateInvoice(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成账单失败: %v", err)
	}

	return &pb.BillResponse{
		Bill: &pb.Bill{
			Id:          invoice.ID.String(),
			MerchantId:  invoice.MerchantID.String(),
			Period:      req.Period,
			TotalAmount: invoice.TotalAmount,
			FeeAmount:   invoice.TaxAmount,
			NetAmount:   invoice.TotalAmount - invoice.TaxAmount,
			Currency:    invoice.Currency,
			Status:      invoice.Status,
			CreatedAt:   timestamppb.New(invoice.CreatedAt),
		},
	}, nil
}

// ListBills 获取账单列表
func (s *AccountingServer) ListBills(ctx context.Context, req *pb.ListBillsRequest) (*pb.ListBillsResponse, error) {
	query := &repository.InvoiceQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.MerchantId != "" {
		merchantID, err := uuid.Parse(req.MerchantId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
		}
		query.MerchantID = &merchantID
	}

	if req.Status != "" {
		query.Status = req.Status
	}

	invoices, total, err := s.accountService.ListInvoices(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	bills := make([]*pb.Bill, len(invoices))
	for i, inv := range invoices {
		period := inv.PeriodStart.Format("2006-01")

		bills[i] = &pb.Bill{
			Id:          inv.ID.String(),
			MerchantId:  inv.MerchantID.String(),
			Period:      period,
			TotalAmount: inv.TotalAmount,
			FeeAmount:   inv.TaxAmount,
			NetAmount:   inv.TotalAmount - inv.TaxAmount,
			Currency:    inv.Currency,
			Status:      inv.Status,
			CreatedAt:   timestamppb.New(inv.CreatedAt),
		}
	}

	return &pb.ListBillsResponse{
		Bills: bills,
		Total: total,
	}, nil
}

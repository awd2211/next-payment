package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/withdrawal"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/withdrawal-service/internal/model"
	"payment-platform/withdrawal-service/internal/service"
)

// WithdrawalServer implements the WithdrawalService gRPC service
type WithdrawalServer struct {
	pb.UnimplementedWithdrawalServiceServer
	withdrawalService service.WithdrawalService
}

// NewWithdrawalServer creates a new Withdrawal gRPC server
func NewWithdrawalServer(withdrawalService service.WithdrawalService) *WithdrawalServer {
	return &WithdrawalServer{
		withdrawalService: withdrawalService,
	}
}

// CreateWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) CreateWithdrawal(ctx context.Context, req *pb.CreateWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	bankAccountID, err := uuid.Parse(req.BankAccountId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid bank_account_id format",
		}, nil
	}

	input := &service.CreateWithdrawalInput{
		MerchantID:    merchantID,
		Amount:        req.Amount,
		Type:          model.WithdrawalTypeNormal,
		BankAccountID: bankAccountID,
		Remarks:       req.Remarks,
		CreatedBy:     merchantID, // Default to merchant ID
	}

	withdrawal, err := s.withdrawalService.CreateWithdrawal(ctx, input)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawalResponse{
		Code:    0,
		Message: "Success",
		Data:    convertWithdrawalToProto(withdrawal),
	}, nil
}

// GetWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) GetWithdrawal(ctx context.Context, req *pb.GetWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	withdrawalID, err := uuid.Parse(req.WithdrawalId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid withdrawal_id format",
		}, nil
	}

	detail, err := s.withdrawalService.GetWithdrawal(ctx, withdrawalID)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    404,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawalResponse{
		Code:    0,
		Message: "Success",
		Data:    convertWithdrawalToProto(detail.Withdrawal),
	}, nil
}

// ListWithdrawals implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ListWithdrawals(ctx context.Context, req *pb.ListWithdrawalsRequest) (*pb.ListWithdrawalsResponse, error) {
	query := &service.ListWithdrawalQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.MerchantId != "" {
		merchantID, err := uuid.Parse(req.MerchantId)
		if err != nil {
			return &pb.ListWithdrawalsResponse{
				Code:    400,
				Message: "Invalid merchant_id format",
			}, nil
		}
		query.MerchantID = &merchantID
	}

	if req.Status != "" {
		status := model.WithdrawalStatus(req.Status)
		query.Status = &status
	}

	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		query.StartDate = &t
	}

	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		query.EndDate = &t
	}

	result, err := s.withdrawalService.ListWithdrawals(ctx, query)
	if err != nil {
		return &pb.ListWithdrawalsResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	withdrawals := make([]*pb.WithdrawalData, 0, len(result.Withdrawals))
	for _, withdrawal := range result.Withdrawals {
		withdrawals = append(withdrawals, convertWithdrawalToProto(withdrawal))
	}

	return &pb.ListWithdrawalsResponse{
		Code:        0,
		Message:     "Success",
		Withdrawals: withdrawals,
		Total:       result.Total,
	}, nil
}

// ApproveWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ApproveWithdrawal(ctx context.Context, req *pb.ApproveWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	withdrawalID, err := uuid.Parse(req.WithdrawalId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid withdrawal_id format",
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid approver_id format",
		}, nil
	}

	err = s.withdrawalService.ApproveWithdrawal(ctx, withdrawalID, approverID, "System", req.Comments)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// Get updated withdrawal
	detail, err := s.withdrawalService.GetWithdrawal(ctx, withdrawalID)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawalResponse{
		Code:    0,
		Message: "Success",
		Data:    convertWithdrawalToProto(detail.Withdrawal),
	}, nil
}

// RejectWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) RejectWithdrawal(ctx context.Context, req *pb.RejectWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	withdrawalID, err := uuid.Parse(req.WithdrawalId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid withdrawal_id format",
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid approver_id format",
		}, nil
	}

	err = s.withdrawalService.RejectWithdrawal(ctx, withdrawalID, approverID, "System", req.RejectReason)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// Get updated withdrawal
	detail, err := s.withdrawalService.GetWithdrawal(ctx, withdrawalID)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawalResponse{
		Code:    0,
		Message: "Success",
		Data:    convertWithdrawalToProto(detail.Withdrawal),
	}, nil
}

// ConfirmWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ConfirmWithdrawal(ctx context.Context, req *pb.ConfirmWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	withdrawalID, err := uuid.Parse(req.WithdrawalId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid withdrawal_id format",
		}, nil
	}

	err = s.withdrawalService.ExecuteWithdrawal(ctx, withdrawalID)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// Get updated withdrawal
	detail, err := s.withdrawalService.GetWithdrawal(ctx, withdrawalID)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawalResponse{
		Code:    0,
		Message: "Success",
		Data:    convertWithdrawalToProto(detail.Withdrawal),
	}, nil
}

// CancelWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) CancelWithdrawal(ctx context.Context, req *pb.CancelWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	withdrawalID, err := uuid.Parse(req.WithdrawalId)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    400,
			Message: "Invalid withdrawal_id format",
		}, nil
	}

	err = s.withdrawalService.CancelWithdrawal(ctx, withdrawalID, req.CancelReason)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// Get updated withdrawal
	detail, err := s.withdrawalService.GetWithdrawal(ctx, withdrawalID)
	if err != nil {
		return &pb.WithdrawalResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawalResponse{
		Code:    0,
		Message: "Success",
		Data:    convertWithdrawalToProto(detail.Withdrawal),
	}, nil
}

// AddBankAccount implements withdrawal.WithdrawalService
func (s *WithdrawalServer) AddBankAccount(ctx context.Context, req *pb.AddBankAccountRequest) (*pb.BankAccountResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.BankAccountResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	input := &service.CreateBankAccountInput{
		MerchantID:      merchantID,
		BankName:        req.BankName,
		BankCode:        req.BankCode,
		BankBranch:      req.BranchName,
		AccountName:     req.AccountName,
		AccountNo:       req.AccountNo,
		AccountType:     req.AccountType,
		IsDefault:       req.IsDefault,
		VerificationDoc: "",
	}

	account, err := s.withdrawalService.CreateBankAccount(ctx, input)
	if err != nil {
		return &pb.BankAccountResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.BankAccountResponse{
		Code:    0,
		Message: "Success",
		Data:    convertBankAccountToProto(account),
	}, nil
}

// ListBankAccounts implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ListBankAccounts(ctx context.Context, req *pb.ListBankAccountsRequest) (*pb.ListBankAccountsResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.ListBankAccountsResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	accounts, err := s.withdrawalService.ListBankAccounts(ctx, merchantID)
	if err != nil {
		return &pb.ListBankAccountsResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	accountsProto := make([]*pb.BankAccountData, 0, len(accounts))
	for _, account := range accounts {
		accountsProto = append(accountsProto, convertBankAccountToProto(account))
	}

	return &pb.ListBankAccountsResponse{
		Code:     0,
		Message:  "Success",
		Accounts: accountsProto,
	}, nil
}

// SetDefaultBankAccount implements withdrawal.WithdrawalService
func (s *WithdrawalServer) SetDefaultBankAccount(ctx context.Context, req *pb.SetDefaultBankAccountRequest) (*pb.StatusResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.StatusResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	accountID, err := uuid.Parse(req.BankAccountId)
	if err != nil {
		return &pb.StatusResponse{
			Code:    400,
			Message: "Invalid bank_account_id format",
		}, nil
	}

	err = s.withdrawalService.SetDefaultBankAccount(ctx, merchantID, accountID)
	if err != nil {
		return &pb.StatusResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.StatusResponse{
		Code:    0,
		Message: "Success",
	}, nil
}

// CreateWithdrawalBatch implements withdrawal.WithdrawalService
func (s *WithdrawalServer) CreateWithdrawalBatch(ctx context.Context, req *pb.CreateWithdrawalBatchRequest) (*pb.WithdrawalBatchResponse, error) {
	// Parse creator ID
	_, err := uuid.Parse(req.CreatorId)
	if err != nil {
		return &pb.WithdrawalBatchResponse{
			Code:    400,
			Message: "Invalid creator_id format",
		}, nil
	}

	// Parse withdrawal IDs
	withdrawalIDs := make([]uuid.UUID, 0, len(req.WithdrawalIds))
	for _, idStr := range req.WithdrawalIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return &pb.WithdrawalBatchResponse{
				Code:    400,
				Message: "Invalid withdrawal_id format: " + idStr,
			}, nil
		}
		withdrawalIDs = append(withdrawalIDs, id)
	}

	// Create batch (simplified - actual implementation would require batch service method)
	batchNo := "BATCH" + time.Now().Format("20060102150405")

	return &pb.WithdrawalBatchResponse{
		Code:    0,
		Message: "Success",
		Data: &pb.WithdrawalBatchData{
			Id:           uuid.New().String(),
			BatchNo:      batchNo,
			TotalCount:   int32(len(withdrawalIDs)),
			TotalAmount:  0, // Would be calculated from actual withdrawals
			Status:       "PENDING",
			SuccessCount: 0,
			FailedCount:  0,
			CreatedAt:    timestamppb.New(time.Now()),
		},
	}, nil
}

// GetWithdrawalStats implements withdrawal.WithdrawalService
func (s *WithdrawalServer) GetWithdrawalStats(ctx context.Context, req *pb.GetWithdrawalStatsRequest) (*pb.WithdrawalStatsResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.WithdrawalStatsResponse{
			Code:    400,
			Message: "Invalid merchant_id format",
		}, nil
	}

	report, err := s.withdrawalService.GetWithdrawalReport(ctx, merchantID, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return &pb.WithdrawalStatsResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawalStatsResponse{
		Code:    0,
		Message: "Success",
		Data: &pb.WithdrawalStatsData{
			MerchantId:           req.MerchantId,
			TotalWithdrawalAmount: report.TotalAmount,
			TotalWithdrawalFee:    report.TotalFee,
			WithdrawalCount:      int32(report.TotalCount),
			PendingCount:         int32(report.PendingCount),
			CompletedCount:       int32(report.CompletedCount),
			AvailableBalance:     0, // Would need to fetch from accounting service
			ByStatus: []*pb.WithdrawalByStatus{
				{Status: "completed", Count: int32(report.CompletedCount), Amount: report.CompletedAmount},
				{Status: "pending", Count: int32(report.PendingCount), Amount: report.PendingAmount},
				{Status: "rejected", Count: int32(report.RejectedCount), Amount: 0},
				{Status: "failed", Count: int32(report.FailedCount), Amount: 0},
			},
		},
	}, nil
}

// Helper function to convert Withdrawal model to proto
func convertWithdrawalToProto(withdrawal *model.Withdrawal) *pb.WithdrawalData {
	data := &pb.WithdrawalData{
		Id:              withdrawal.ID.String(),
		WithdrawalNo:    withdrawal.WithdrawalNo,
		MerchantId:      withdrawal.MerchantID.String(),
		Amount:          withdrawal.Amount,
		Fee:             withdrawal.Fee,
		NetAmount:       withdrawal.ActualAmount,
		Currency:        "CNY",
		Status:          string(withdrawal.Status),
		BankAccountId:   withdrawal.BankAccountID.String(),
		BankName:        withdrawal.BankName,
		BankAccountNo:   withdrawal.BankAccountNo,
		BankAccountName: withdrawal.BankAccountName,
		Remarks:         withdrawal.Remarks,
		CreatedAt:       timestamppb.New(withdrawal.CreatedAt),
	}

	if withdrawal.ChannelTradeNo != "" {
		data.TransactionNo = withdrawal.ChannelTradeNo
	}

	if withdrawal.FailureReason != "" {
		data.RejectReason = withdrawal.FailureReason
	}

	if withdrawal.CompletedAt != nil {
		data.CompletedAt = timestamppb.New(*withdrawal.CompletedAt)
	}

	return data
}

// Helper function to convert BankAccount model to proto
func convertBankAccountToProto(account *model.WithdrawalBankAccount) *pb.BankAccountData {
	return &pb.BankAccountData{
		Id:          account.ID.String(),
		MerchantId:  account.MerchantID.String(),
		BankName:    account.BankName,
		BankCode:    account.BankCode,
		AccountNo:   account.AccountNo,
		AccountName: account.AccountName,
		AccountType: account.AccountType,
		BranchName:  account.BankBranch,
		SwiftCode:   "", // Not stored in model
		IsDefault:   account.IsDefault,
		Status:      account.Status,
		CreatedAt:   timestamppb.New(account.CreatedAt),
	}
}

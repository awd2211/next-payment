package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/withdrawal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	return nil, status.Errorf(codes.Unimplemented, "method CreateWithdrawal not implemented")
}

// GetWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) GetWithdrawal(ctx context.Context, req *pb.GetWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWithdrawal not implemented")
}

// ListWithdrawals implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ListWithdrawals(ctx context.Context, req *pb.ListWithdrawalsRequest) (*pb.ListWithdrawalsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWithdrawals not implemented")
}

// ApproveWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ApproveWithdrawal(ctx context.Context, req *pb.ApproveWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApproveWithdrawal not implemented")
}

// RejectWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) RejectWithdrawal(ctx context.Context, req *pb.RejectWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RejectWithdrawal not implemented")
}

// ConfirmWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ConfirmWithdrawal(ctx context.Context, req *pb.ConfirmWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmWithdrawal not implemented")
}

// CancelWithdrawal implements withdrawal.WithdrawalService
func (s *WithdrawalServer) CancelWithdrawal(ctx context.Context, req *pb.CancelWithdrawalRequest) (*pb.WithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelWithdrawal not implemented")
}

// AddBankAccount implements withdrawal.WithdrawalService
func (s *WithdrawalServer) AddBankAccount(ctx context.Context, req *pb.AddBankAccountRequest) (*pb.BankAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddBankAccount not implemented")
}

// ListBankAccounts implements withdrawal.WithdrawalService
func (s *WithdrawalServer) ListBankAccounts(ctx context.Context, req *pb.ListBankAccountsRequest) (*pb.ListBankAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListBankAccounts not implemented")
}

// SetDefaultBankAccount implements withdrawal.WithdrawalService
func (s *WithdrawalServer) SetDefaultBankAccount(ctx context.Context, req *pb.SetDefaultBankAccountRequest) (*pb.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetDefaultBankAccount not implemented")
}

// CreateWithdrawalBatch implements withdrawal.WithdrawalService
func (s *WithdrawalServer) CreateWithdrawalBatch(ctx context.Context, req *pb.CreateWithdrawalBatchRequest) (*pb.WithdrawalBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWithdrawalBatch not implemented")
}

// GetWithdrawalStats implements withdrawal.WithdrawalService
func (s *WithdrawalServer) GetWithdrawalStats(ctx context.Context, req *pb.GetWithdrawalStatsRequest) (*pb.WithdrawalStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWithdrawalStats not implemented")
}

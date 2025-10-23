package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/accounting"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AccountingServer gRPC服务实现
type AccountingServer struct {
	pb.UnimplementedAccountingServiceServer
}

// NewAccountingServer 创建gRPC服务实例
func NewAccountingServer() *AccountingServer {
	return &AccountingServer{}
}

// 所有方法暂时返回未实现
func (s *AccountingServer) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.EntryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) GetEntry(ctx context.Context, req *pb.GetEntryRequest) (*pb.EntryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) ListEntries(ctx context.Context, req *pb.ListEntriesRequest) (*pb.ListEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) UpdateSettlementStatus(ctx context.Context, req *pb.UpdateSettlementStatusRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) GetMerchantBalance(ctx context.Context, req *pb.GetMerchantBalanceRequest) (*pb.MerchantBalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) GenerateBill(ctx context.Context, req *pb.GenerateBillRequest) (*pb.BillResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AccountingServer) ListBills(ctx context.Context, req *pb.ListBillsRequest) (*pb.ListBillsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

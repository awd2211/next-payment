package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/settlement"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"payment-platform/settlement-service/internal/service"
)

// SettlementServer implements the SettlementService gRPC service
type SettlementServer struct {
	pb.UnimplementedSettlementServiceServer
	settlementService *service.SettlementService
}

// NewSettlementServer creates a new Settlement gRPC server
func NewSettlementServer(settlementService *service.SettlementService) *SettlementServer {
	return &SettlementServer{
		settlementService: settlementService,
	}
}

// CreateSettlement implements settlement.SettlementService
func (s *SettlementServer) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSettlement not implemented")
}

// GetSettlement implements settlement.SettlementService
func (s *SettlementServer) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSettlement not implemented")
}

// ListSettlements implements settlement.SettlementService
func (s *SettlementServer) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSettlements not implemented")
}

// ApproveSettlement implements settlement.SettlementService
func (s *SettlementServer) ApproveSettlement(ctx context.Context, req *pb.ApproveSettlementRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApproveSettlement not implemented")
}

// RejectSettlement implements settlement.SettlementService
func (s *SettlementServer) RejectSettlement(ctx context.Context, req *pb.RejectSettlementRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RejectSettlement not implemented")
}

// ConfirmSettlement implements settlement.SettlementService
func (s *SettlementServer) ConfirmSettlement(ctx context.Context, req *pb.ConfirmSettlementRequest) (*pb.SettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmSettlement not implemented")
}

// CalculateSettlement implements settlement.SettlementService
func (s *SettlementServer) CalculateSettlement(ctx context.Context, req *pb.CalculateSettlementRequest) (*pb.CalculateSettlementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CalculateSettlement not implemented")
}

// GetSettlementStats implements settlement.SettlementService
func (s *SettlementServer) GetSettlementStats(ctx context.Context, req *pb.GetSettlementStatsRequest) (*pb.SettlementStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSettlementStats not implemented")
}

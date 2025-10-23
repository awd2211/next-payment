package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/kyc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"payment-platform/kyc-service/internal/service"
)

// KYCServer implements the KYCService gRPC service
type KYCServer struct {
	pb.UnimplementedKYCServiceServer
	kycService *service.KYCService
}

// NewKYCServer creates a new KYC gRPC server
func NewKYCServer(kycService *service.KYCService) *KYCServer {
	return &KYCServer{
		kycService: kycService,
	}
}

// SubmitDocument implements kyc.KYCService
func (s *KYCServer) SubmitDocument(ctx context.Context, req *pb.SubmitDocumentRequest) (*pb.DocumentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitDocument not implemented")
}

// GetDocument implements kyc.KYCService
func (s *KYCServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.DocumentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDocument not implemented")
}

// UpdateDocumentStatus implements kyc.KYCService
func (s *KYCServer) UpdateDocumentStatus(ctx context.Context, req *pb.UpdateDocumentStatusRequest) (*pb.DocumentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDocumentStatus not implemented")
}

// GetMerchantKYCLevel implements kyc.KYCService
func (s *KYCServer) GetMerchantKYCLevel(ctx context.Context, req *pb.GetMerchantKYCLevelRequest) (*pb.KYCLevelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMerchantKYCLevel not implemented")
}

// CreateReview implements kyc.KYCService
func (s *KYCServer) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.ReviewResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateReview not implemented")
}

// GetReview implements kyc.KYCService
func (s *KYCServer) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.ReviewResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReview not implemented")
}

// CompleteReview implements kyc.KYCService
func (s *KYCServer) CompleteReview(ctx context.Context, req *pb.CompleteReviewRequest) (*pb.ReviewResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CompleteReview not implemented")
}

// ListAlerts implements kyc.KYCService
func (s *KYCServer) ListAlerts(ctx context.Context, req *pb.ListAlertsRequest) (*pb.ListAlertsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAlerts not implemented")
}

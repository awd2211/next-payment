package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/analytics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"payment-platform/analytics-service/internal/service"
)

// AnalyticsServer gRPC服务实现
type AnalyticsServer struct {
	pb.UnimplementedAnalyticsServiceServer
	analyticsService service.AnalyticsService
}

// NewAnalyticsServer 创建gRPC服务实例
func NewAnalyticsServer(analyticsService service.AnalyticsService) *AnalyticsServer {
	return &AnalyticsServer{
		analyticsService: analyticsService,
	}
}

// 所有方法暂时返回未实现
func (s *AnalyticsServer) GetPaymentStats(ctx context.Context, req *pb.GetPaymentStatsRequest) (*pb.PaymentStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) GetPaymentTrends(ctx context.Context, req *pb.GetPaymentTrendsRequest) (*pb.PaymentTrendsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) GetChannelStats(ctx context.Context, req *pb.GetChannelStatsRequest) (*pb.ChannelStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) GetMerchantStats(ctx context.Context, req *pb.GetMerchantStatsRequest) (*pb.MerchantStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) GetTopMerchants(ctx context.Context, req *pb.GetTopMerchantsRequest) (*pb.TopMerchantsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) GetRealtimeMetrics(ctx context.Context, req *pb.GetRealtimeMetricsRequest) (*pb.RealtimeMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) GetSystemHealth(ctx context.Context, req *pb.GetSystemHealthRequest) (*pb.SystemHealthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*pb.ReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *AnalyticsServer) ListReports(ctx context.Context, req *pb.ListReportsRequest) (*pb.ListReportsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

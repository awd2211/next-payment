package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/analytics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/analytics-service/internal/repository"
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

// GetPaymentStats 获取支付统计
func (s *AnalyticsServer) GetPaymentStats(ctx context.Context, req *pb.GetPaymentStatsRequest) (*pb.PaymentStatsResponse, error) {
	// 解析商户ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID: %v", err)
	}

	// 解析时间范围
	startTime := req.StartTime.AsTime()
	endTime := req.EndTime.AsTime()

	// 获取支付汇总数据
	summary, err := s.analyticsService.GetPaymentSummary(ctx, merchantID, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取支付统计失败: %v", err)
	}

	// 计算时间段
	period := fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	return &pb.PaymentStatsResponse{
		Stats: &pb.PaymentStats{
			TotalCount:    int64(summary.TotalPayments),
			SuccessCount:  int64(summary.SuccessPayments),
			FailedCount:   int64(summary.FailedPayments),
			TotalAmount:   summary.TotalAmount,
			SuccessRate:   summary.SuccessRate,
			AverageAmount: float64(summary.AverageAmount),
			Currency:      summary.Currency,
			Period:        period,
		},
	}, nil
}

// GetPaymentTrends 获取支付趋势
func (s *AnalyticsServer) GetPaymentTrends(ctx context.Context, req *pb.GetPaymentTrendsRequest) (*pb.PaymentTrendsResponse, error) {
	// 解析商户ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID: %v", err)
	}

	// 解析时间范围
	startTime := req.StartTime.AsTime()
	endTime := req.EndTime.AsTime()

	// 获取支付指标数据（按天）
	metrics, err := s.analyticsService.GetPaymentMetrics(ctx, merchantID, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取支付趋势失败: %v", err)
	}

	// 转换为趋势点
	trends := make([]*pb.TrendPoint, len(metrics))
	for i, m := range metrics {
		trends[i] = &pb.TrendPoint{
			Time:   m.Date.Format("2006-01-02"),
			Count:  int64(m.TotalPayments),
			Amount: m.TotalAmount,
		}
	}

	return &pb.PaymentTrendsResponse{
		Trends: trends,
	}, nil
}

// GetChannelStats 获取渠道统计
func (s *AnalyticsServer) GetChannelStats(ctx context.Context, req *pb.GetChannelStatsRequest) (*pb.ChannelStatsResponse, error) {
	// 解析时间范围
	startTime := req.StartTime.AsTime()
	endTime := req.EndTime.AsTime()

	// 获取所有渠道的指标（这里简化处理，实际应该查询所有渠道）
	// 由于服务层方法需要channelCode，我们需要遍历已知的渠道
	channels := []string{"stripe", "paypal", "alipay", "wechat"}
	var channelStats []*pb.ChannelStat

	for _, channelCode := range channels {
		summary, err := s.analyticsService.GetChannelSummary(ctx, channelCode, startTime, endTime)
		if err != nil {
			// 跳过没有数据的渠道
			continue
		}

		if summary.TotalTransactions > 0 {
			channelStats = append(channelStats, &pb.ChannelStat{
				Channel:     channelCode,
				Count:       int64(summary.TotalTransactions),
				Amount:      summary.TotalAmount,
				SuccessRate: summary.SuccessRate,
			})
		}
	}

	return &pb.ChannelStatsResponse{
		Channels: channelStats,
	}, nil
}

// GetMerchantStats 获取商户统计
func (s *AnalyticsServer) GetMerchantStats(ctx context.Context, req *pb.GetMerchantStatsRequest) (*pb.MerchantStatsResponse, error) {
	// 解析商户ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID: %v", err)
	}

	// 解析时间范围
	startTime := req.StartTime.AsTime()
	endTime := req.EndTime.AsTime()

	// 获取商户汇总数据
	summary, err := s.analyticsService.GetMerchantSummary(ctx, merchantID, startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取商户统计失败: %v", err)
	}

	// 计算转化率（完成订单数 / 总订单数）
	conversionRate := 0.0
	if summary.TotalOrders > 0 {
		conversionRate = float64(summary.CompletedOrders) / float64(summary.TotalOrders) * 100
	}

	return &pb.MerchantStatsResponse{
		Stats: &pb.MerchantStats{
			MerchantId:        req.MerchantId,
			MerchantName:      "", // 需要从merchant服务获取
			TotalTransactions: int64(summary.TotalOrders),
			TotalAmount:       summary.TotalRevenue,
			ActiveUsers:       int64(summary.NewCustomers + summary.ReturningCustomers),
			ConversionRate:    conversionRate,
		},
	}, nil
}

// GetTopMerchants 获取头部商户
func (s *AnalyticsServer) GetTopMerchants(ctx context.Context, req *pb.GetTopMerchantsRequest) (*pb.TopMerchantsResponse, error) {
	// TODO: 需要在repository层实现跨商户的统计查询
	// 这里暂时返回空列表，表示功能待实现
	return &pb.TopMerchantsResponse{
		Merchants: []*pb.TopMerchant{},
	}, nil
}

// GetRealtimeMetrics 获取实时指标
func (s *AnalyticsServer) GetRealtimeMetrics(ctx context.Context, req *pb.GetRealtimeMetricsRequest) (*pb.RealtimeMetricsResponse, error) {
	// 解析时间范围（1m, 5m, 15m）
	timeRange := req.TimeRange
	if timeRange == "" {
		timeRange = "5m" // 默认5分钟
	}

	// 构建查询条件
	query := &repository.RealtimeStatsQuery{
		StatType: "realtime",
		Period:   timeRange,
	}

	stats, err := s.analyticsService.GetRealtimeStats(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取实时指标失败: %v", err)
	}

	// 聚合统计数据
	metrics := &pb.RealtimeMetrics{
		ActiveSessions:         0,
		RequestsPerSecond:      0,
		SuccessfulPaymentsCount: 0,
		FailedPaymentsCount:    0,
		AverageResponseTime:    0,
	}

	for _, stat := range stats {
		switch stat.StatKey {
		case "active_sessions":
			metrics.ActiveSessions = stat.StatValue
		case "requests_per_second":
			metrics.RequestsPerSecond = stat.StatValue
		case "successful_payments":
			metrics.SuccessfulPaymentsCount = stat.StatValue
		case "failed_payments":
			metrics.FailedPaymentsCount = stat.StatValue
		case "avg_response_time":
			metrics.AverageResponseTime = float64(stat.StatValue)
		}
	}

	return &pb.RealtimeMetricsResponse{
		Metrics: metrics,
	}, nil
}

// GetSystemHealth 获取系统健康状态
func (s *AnalyticsServer) GetSystemHealth(ctx context.Context, req *pb.GetSystemHealthRequest) (*pb.SystemHealthResponse, error) {
	// TODO: 需要实现系统健康检查逻辑
	// 这里返回一个默认的健康状态
	services := []*pb.ServiceHealth{
		{
			ServiceName:  "analytics-service",
			Status:       "healthy",
			CpuUsage:     0.0,
			MemoryUsage:  0.0,
			RequestCount: 0,
			ErrorRate:    0.0,
		},
	}

	return &pb.SystemHealthResponse{
		OverallStatus: "healthy",
		Services:      services,
	}, nil
}

// GenerateReport 生成报表
func (s *AnalyticsServer) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*pb.ReportResponse, error) {
	// TODO: 需要实现报表生成功能
	// 这里返回一个模拟的报表响应
	reportID := uuid.New().String()

	report := &pb.Report{
		Id:        reportID,
		Name:      req.Name,
		Type:      req.Type,
		Period:    fmt.Sprintf("%s to %s", req.StartTime.AsTime().Format("2006-01-02"), req.EndTime.AsTime().Format("2006-01-02")),
		Format:    req.Format,
		FileUrl:   "",
		Status:    "generating",
		CreatedAt: timestamppb.New(time.Now()),
	}

	return &pb.ReportResponse{
		Report: report,
	}, nil
}

// ListReports 获取报表列表
func (s *AnalyticsServer) ListReports(ctx context.Context, req *pb.ListReportsRequest) (*pb.ListReportsResponse, error) {
	// TODO: 需要实现报表存储和查询功能
	// 这里返回空列表，表示功能待实现
	return &pb.ListReportsResponse{
		Reports: []*pb.Report{},
		Total:   0,
	}, nil
}

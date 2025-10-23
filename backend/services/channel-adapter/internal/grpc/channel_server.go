package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/channel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/channel-adapter/internal/service"
)

// ChannelServer gRPC服务实现
type ChannelServer struct {
	pb.UnimplementedChannelServiceServer
	channelService service.ChannelService
}

// NewChannelServer 创建gRPC服务实例
func NewChannelServer(channelService service.ChannelService) *ChannelServer {
	return &ChannelServer{
		channelService: channelService,
	}
}

// CreatePayment 创建支付
func (s *ChannelServer) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	input := &service.CreatePaymentRequest{
		MerchantID:    merchantID,
		PaymentNo:     req.PaymentNo,
		Channel:       req.Channel,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CustomerEmail: req.CustomerEmail,
		CustomerName:  req.CustomerName,
		Description:   req.Description,
		SuccessURL:    req.ReturnUrl,
		CallbackURL:   req.NotifyUrl,
	}

	resp, err := s.channelService.CreatePayment(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建支付失败: %v", err)
	}

	return &pb.CreatePaymentResponse{
		ChannelOrderNo: resp.ChannelTradeNo,
		PaymentUrl:     resp.PaymentURL,
		ClientSecret:   resp.ClientSecret,
		Status:         resp.Status,
	}, nil
}

// QueryPayment 查询支付
func (s *ChannelServer) QueryPayment(ctx context.Context, req *pb.QueryPaymentRequest) (*pb.QueryPaymentResponse, error) {
	resp, err := s.channelService.QueryPayment(ctx, req.PaymentNo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询支付失败: %v", err)
	}

	var paidAt *timestamppb.Timestamp
	if resp.PaidAt != nil {
		paidAt = timestamppb.New(*resp.PaidAt)
	}

	return &pb.QueryPaymentResponse{
		ChannelOrderNo: resp.ChannelTradeNo,
		Status:         resp.Status,
		Amount:         resp.Amount,
		Currency:       resp.Currency,
		PaidAt:         paidAt,
		FailureCode:    "",
		FailureMessage: "",
	}, nil
}

// CancelPayment 取消支付
func (s *ChannelServer) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.CancelPaymentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// CreateRefund 创建退款
func (s *ChannelServer) CreateRefund(ctx context.Context, req *pb.CreateRefundRequest) (*pb.CreateRefundResponse, error) {
	input := &service.CreateRefundRequest{
		MerchantID: uuid.Nil, // Service layer will retrieve from payment record
		PaymentNo:  req.PaymentNo,
		RefundNo:   req.RefundNo,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Reason:     req.Reason,
	}

	resp, err := s.channelService.CreateRefund(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建退款失败: %v", err)
	}

	return &pb.CreateRefundResponse{
		ChannelRefundNo: resp.ChannelRefundNo,
		Status:          resp.Status,
		RefundedAt:      nil,
	}, nil
}

// QueryRefund 查询退款
func (s *ChannelServer) QueryRefund(ctx context.Context, req *pb.QueryRefundRequest) (*pb.QueryRefundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// GetChannelConfig 获取渠道配置
func (s *ChannelServer) GetChannelConfig(ctx context.Context, req *pb.GetChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// UpdateChannelConfig 更新渠道配置
func (s *ChannelServer) UpdateChannelConfig(ctx context.Context, req *pb.UpdateChannelConfigRequest) (*pb.ChannelConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

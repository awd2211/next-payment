package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/payment-gateway/internal/service"
)

// PaymentServer gRPC服务实现
type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	paymentService *service.PaymentService
}

// NewPaymentServer 创建gRPC服务实例
func NewPaymentServer(paymentService *service.PaymentService) *PaymentServer {
	return &PaymentServer{
		paymentService: paymentService,
	}
}

// CreatePayment 创建支付
func (s *PaymentServer) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.PaymentResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	customerID, _ := uuid.Parse(req.CustomerId)

	input := &service.CreatePaymentInput{
		MerchantID:    merchantID,
		OrderNo:       req.OrderId,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Channel:       req.Channel,
		PayMethod:     req.PaymentMethod,
		CustomerEmail: "",
		Description:   req.Description,
		ReturnURL:     req.ReturnUrl,
		Extra:         make(map[string]interface{}),
	}

	payment, err := s.paymentService.CreatePayment(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建支付失败: %v", err)
	}

	var paidAt *timestamppb.Timestamp
	if payment.PaidAt != nil {
		paidAt = timestamppb.New(*payment.PaidAt)
	}

	var expiresAt *timestamppb.Timestamp
	if payment.ExpiredAt != nil {
		expiresAt = timestamppb.New(*payment.ExpiredAt)
	}

	return &pb.PaymentResponse{
		Payment: &pb.Payment{
			Id:               payment.ID.String(),
			OrderId:          payment.OrderNo,
			MerchantId:       payment.MerchantID.String(),
			CustomerId:       customerID.String(),
			Amount:           payment.Amount,
			Currency:         payment.Currency,
			Channel:          payment.Channel,
			PaymentMethod:    payment.PayMethod,
			Status:           payment.Status,
			FailureCode:      payment.ErrorCode,
			FailureMessage:   payment.ErrorMsg,
			ChannelPaymentId: payment.ChannelOrderNo,
			ReturnUrl:        payment.ReturnURL,
			PaidAt:           paidAt,
			ExpiresAt:        expiresAt,
			CreatedAt:        timestamppb.New(payment.CreatedAt),
			UpdatedAt:        timestamppb.New(payment.UpdatedAt),
		},
		CheckoutUrl: payment.PaymentURL,
	}, nil
}

// GetPayment 获取支付信息
func (s *PaymentServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.PaymentResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的支付ID")
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	payment, err := s.paymentService.GetPaymentByID(ctx, id, merchantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "支付记录不存在")
	}

	var paidAt *timestamppb.Timestamp
	if payment.PaidAt != nil {
		paidAt = timestamppb.New(*payment.PaidAt)
	}

	var expiresAt *timestamppb.Timestamp
	if payment.ExpiredAt != nil {
		expiresAt = timestamppb.New(*payment.ExpiredAt)
	}

	return &pb.PaymentResponse{
		Payment: &pb.Payment{
			Id:               payment.ID.String(),
			OrderId:          payment.OrderNo,
			MerchantId:       payment.MerchantID.String(),
			Amount:           payment.Amount,
			Currency:         payment.Currency,
			Channel:          payment.Channel,
			PaymentMethod:    payment.PayMethod,
			Status:           payment.Status,
			FailureCode:      payment.ErrorCode,
			FailureMessage:   payment.ErrorMsg,
			ChannelPaymentId: payment.ChannelOrderNo,
			ClientIp:         payment.CustomerIP,
			ReturnUrl:        payment.ReturnURL,
			PaidAt:           paidAt,
			ExpiresAt:        expiresAt,
			CreatedAt:        timestamppb.New(payment.CreatedAt),
			UpdatedAt:        timestamppb.New(payment.UpdatedAt),
		},
	}, nil
}

// ListPayments 支付列表
func (s *PaymentServer) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	// 实现支付列表查询逻辑
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// CancelPayment 取消支付
func (s *PaymentServer) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.PaymentResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的支付ID")
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	payment, err := s.paymentService.CancelPayment(ctx, id, merchantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "取消支付失败: %v", err)
	}

	return &pb.PaymentResponse{
		Payment: &pb.Payment{
			Id:          payment.ID.String(),
			OrderId:     payment.OrderNo,
			MerchantId:  payment.MerchantID.String(),
			Amount:      payment.Amount,
			Currency:    payment.Currency,
			Channel:     payment.Channel,
			Status:      payment.Status,
			CreatedAt:   timestamppb.New(payment.CreatedAt),
			UpdatedAt:   timestamppb.New(payment.UpdatedAt),
		},
	}, nil
}

// CreateRefund 创建退款
func (s *PaymentServer) CreateRefund(ctx context.Context, req *pb.CreateRefundRequest) (*pb.RefundResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	input := &service.CreateRefundInput{
		MerchantID: merchantID,
		PaymentID:  req.PaymentId,
		Amount:     req.Amount,
		Reason:     req.Reason,
	}

	refund, err := s.paymentService.CreateRefund(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建退款失败: %v", err)
	}

	var refundedAt *timestamppb.Timestamp
	if refund.RefundedAt != nil {
		refundedAt = timestamppb.New(*refund.RefundedAt)
	}

	return &pb.RefundResponse{
		Refund: &pb.Refund{
			Id:              refund.ID.String(),
			PaymentId:       refund.PaymentNo,
			MerchantId:      refund.MerchantID.String(),
			Amount:          refund.Amount,
			Currency:        refund.Currency,
			Reason:          refund.Reason,
			Status:          refund.Status,
			ChannelRefundId: refund.ChannelRefundNo,
			RefundedAt:      refundedAt,
			CreatedAt:       timestamppb.New(refund.CreatedAt),
			UpdatedAt:       timestamppb.New(refund.UpdatedAt),
		},
	}, nil
}

// GetRefund 获取退款信息
func (s *PaymentServer) GetRefund(ctx context.Context, req *pb.GetRefundRequest) (*pb.RefundResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的退款ID")
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	refund, err := s.paymentService.GetRefundByID(ctx, id, merchantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "退款记录不存在")
	}

	var refundedAt *timestamppb.Timestamp
	if refund.RefundedAt != nil {
		refundedAt = timestamppb.New(*refund.RefundedAt)
	}

	return &pb.RefundResponse{
		Refund: &pb.Refund{
			Id:         refund.ID.String(),
			PaymentId:  refund.PaymentNo,
			MerchantId: refund.MerchantID.String(),
			Amount:     refund.Amount,
			Currency:   refund.Currency,
			Status:     refund.Status,
			RefundedAt: refundedAt,
			CreatedAt:  timestamppb.New(refund.CreatedAt),
			UpdatedAt:  timestamppb.New(refund.UpdatedAt),
		},
	}, nil
}

// ListRefunds 退款列表
func (s *PaymentServer) ListRefunds(ctx context.Context, req *pb.ListRefundsRequest) (*pb.ListRefundsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// HandleWebhook Webhook处理
func (s *PaymentServer) HandleWebhook(ctx context.Context, req *pb.HandleWebhookRequest) (*pb.HandleWebhookResponse, error) {
	// Webhook 处理逻辑
	err := fmt.Errorf("webhook处理待实现")
	if err != nil {
		return &pb.HandleWebhookResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.HandleWebhookResponse{
		Success: true,
		Message: "处理成功",
	}, nil
}

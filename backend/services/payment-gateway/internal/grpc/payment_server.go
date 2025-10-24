package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/payment-gateway/internal/repository"
	"payment-platform/payment-gateway/internal/service"
)

// PaymentServer gRPC服务实现
type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	paymentService service.PaymentService
}

// NewPaymentServer 创建gRPC服务实例
func NewPaymentServer(paymentService service.PaymentService) *PaymentServer {
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
		CheckoutUrl: payment.ReturnURL, // 使用ReturnURL作为checkout URL
	}, nil
}

// GetPayment 获取支付信息
func (s *PaymentServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.PaymentResponse, error) {
	payment, err := s.paymentService.GetPayment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "支付不存在: %v", err)
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
			CustomerId:       payment.CustomerEmail, // 使用CustomerEmail作为customer_id
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
	// 构建查询条件
	query := &repository.PaymentQuery{
		Channel:  req.Channel,
		Status:   req.Status,
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

	if req.CustomerId != "" {
		query.CustomerEmail = req.CustomerId // 使用customer_id字段查询CustomerEmail
	}

	if req.StartTime != nil {
		startTime := req.StartTime.AsTime()
		query.StartTime = &startTime
	}

	if req.EndTime != nil {
		endTime := req.EndTime.AsTime()
		query.EndTime = &endTime
	}

	payments, total, err := s.paymentService.QueryPayment(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	pbPayments := make([]*pb.Payment, len(payments))
	for i, p := range payments {
		var paidAt *timestamppb.Timestamp
		if p.PaidAt != nil {
			paidAt = timestamppb.New(*p.PaidAt)
		}

		var expiresAt *timestamppb.Timestamp
		if p.ExpiredAt != nil {
			expiresAt = timestamppb.New(*p.ExpiredAt)
		}

		pbPayments[i] = &pb.Payment{
			Id:               p.ID.String(),
			OrderId:          p.OrderNo,
			MerchantId:       p.MerchantID.String(),
			CustomerId:       p.CustomerEmail,
			Amount:           p.Amount,
			Currency:         p.Currency,
			Channel:          p.Channel,
			PaymentMethod:    p.PayMethod,
			Status:           p.Status,
			FailureCode:      p.ErrorCode,
			FailureMessage:   p.ErrorMsg,
			ChannelPaymentId: p.ChannelOrderNo,
			ClientIp:         p.CustomerIP,
			ReturnUrl:        p.ReturnURL,
			PaidAt:           paidAt,
			ExpiresAt:        expiresAt,
			CreatedAt:        timestamppb.New(p.CreatedAt),
			UpdatedAt:        timestamppb.New(p.UpdatedAt),
		}
	}

	return &pb.ListPaymentsResponse{
		Payments: pbPayments,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CancelPayment 取消支付
func (s *PaymentServer) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.PaymentResponse, error) {
	err := s.paymentService.CancelPayment(ctx, req.Id, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "取消支付失败: %v", err)
	}

	// 获取更新后的支付信息
	payment, err := s.paymentService.GetPayment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取支付信息失败: %v", err)
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
			CustomerId:       payment.CustomerEmail,
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

// CreateRefund 创建退款
func (s *PaymentServer) CreateRefund(ctx context.Context, req *pb.CreateRefundRequest) (*pb.RefundResponse, error) {
	input := &service.CreateRefundInput{
		PaymentNo:   req.PaymentId,  // PaymentId在proto中是string类型的payment_no
		Amount:      req.Amount,
		Reason:      req.Reason,
		Description: req.Reason,
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
			PaymentId:       refund.PaymentID.String(),
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
	refund, err := s.paymentService.GetRefund(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "退款不存在: %v", err)
	}

	var refundedAt *timestamppb.Timestamp
	if refund.RefundedAt != nil {
		refundedAt = timestamppb.New(*refund.RefundedAt)
	}

	return &pb.RefundResponse{
		Refund: &pb.Refund{
			Id:              refund.ID.String(),
			PaymentId:       refund.PaymentID.String(),
			MerchantId:      refund.MerchantID.String(),
			Amount:          refund.Amount,
			Currency:        refund.Currency,
			Reason:          refund.Reason,
			Status:          refund.Status,
			FailureCode:     refund.ErrorCode,
			FailureMessage:  refund.ErrorMsg,
			ChannelRefundId: refund.ChannelRefundNo,
			RefundedAt:      refundedAt,
			CreatedAt:       timestamppb.New(refund.CreatedAt),
			UpdatedAt:       timestamppb.New(refund.UpdatedAt),
		},
	}, nil
}

// ListRefunds 退款列表
func (s *PaymentServer) ListRefunds(ctx context.Context, req *pb.ListRefundsRequest) (*pb.ListRefundsResponse, error) {
	// 构建查询条件
	query := &repository.RefundQuery{
		Status:   req.Status,
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

	if req.PaymentId != "" {
		paymentID, err := uuid.Parse(req.PaymentId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的支付ID")
		}
		query.PaymentID = &paymentID
	}

	refunds, total, err := s.paymentService.QueryRefunds(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	pbRefunds := make([]*pb.Refund, len(refunds))
	for i, r := range refunds {
		var refundedAt *timestamppb.Timestamp
		if r.RefundedAt != nil {
			refundedAt = timestamppb.New(*r.RefundedAt)
		}

		pbRefunds[i] = &pb.Refund{
			Id:              r.ID.String(),
			PaymentId:       r.PaymentID.String(),
			MerchantId:      r.MerchantID.String(),
			Amount:          r.Amount,
			Currency:        r.Currency,
			Reason:          r.Reason,
			Status:          r.Status,
			FailureCode:     r.ErrorCode,
			FailureMessage:  r.ErrorMsg,
			ChannelRefundId: r.ChannelRefundNo,
			RefundedAt:      refundedAt,
			CreatedAt:       timestamppb.New(r.CreatedAt),
			UpdatedAt:       timestamppb.New(r.UpdatedAt),
		}
	}

	return &pb.ListRefundsResponse{
		Refunds: pbRefunds,
		Total:   int32(total),
	}, nil
}

// HandleWebhook Webhook处理
func (s *PaymentServer) HandleWebhook(ctx context.Context, req *pb.HandleWebhookRequest) (*pb.HandleWebhookResponse, error) {
	// 将payload和headers转换为map
	data := make(map[string]interface{})
	data["payload"] = req.Payload
	data["headers"] = req.Headers

	// 调用service层的HandleCallback处理webhook
	err := s.paymentService.HandleCallback(ctx, req.Channel, data)
	if err != nil {
		return &pb.HandleWebhookResponse{
			Success: false,
			Message: fmt.Sprintf("Webhook处理失败: %v", err),
		}, nil
	}

	return &pb.HandleWebhookResponse{
		Success: true,
		Message: "Webhook处理成功",
	}, nil
}

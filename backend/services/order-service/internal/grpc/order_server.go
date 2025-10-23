package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/order-service/internal/service"
)

// OrderServer gRPC服务实现
type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	orderService service.OrderService
}

// NewOrderServer 创建gRPC服务实例
func NewOrderServer(orderService service.OrderService) *OrderServer {
	return &OrderServer{
		orderService: orderService,
	}
}

// CreateOrder 创建订单
func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	input := &service.CreateOrderInput{
		MerchantID:    merchantID,
		OrderNo:       req.OrderNo,
		PaymentNo:     req.PaymentNo,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Channel:       req.Channel,
		PayMethod:     req.PaymentMethod,
		CustomerEmail: req.CustomerEmail,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		CustomerIP:    req.ClientIp,
		Description:   req.Description,
	}

	order, err := s.orderService.CreateOrder(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建订单失败: %v", err)
	}

	var paidAt *timestamppb.Timestamp
	if order.PaidAt != nil {
		paidAt = timestamppb.New(*order.PaidAt)
	}

	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:            order.ID.String(),
			MerchantId:    order.MerchantID.String(),
			OrderNo:       order.OrderNo,
			PaymentNo:     order.PaymentNo,
			Amount:        order.Amount,
			Currency:      order.Currency,
			Status:        order.Status,
			Channel:       order.Channel,
			PaymentMethod: order.PayMethod,
			CustomerEmail: order.CustomerEmail,
			CustomerName:  order.CustomerName,
			Description:   order.Description,
			PaidAt:        paidAt,
			CreatedAt:     timestamppb.New(order.CreatedAt),
			UpdatedAt:     timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// GetOrder 获取订单
func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	order, err := s.orderService.GetOrderByNo(ctx, merchantID, req.OrderNo)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}

	var paidAt *timestamppb.Timestamp
	if order.PaidAt != nil {
		paidAt = timestamppb.New(*order.PaidAt)
	}

	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:            order.ID.String(),
			MerchantId:    order.MerchantID.String(),
			OrderNo:       order.OrderNo,
			PaymentNo:     order.PaymentNo,
			Amount:        order.Amount,
			Currency:      order.Currency,
			Status:        order.Status,
			Channel:       order.Channel,
			PaymentMethod: order.PayMethod,
			PaidAt:        paidAt,
			CreatedAt:     timestamppb.New(order.CreatedAt),
			UpdatedAt:     timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商户ID")
	}

	order, err := s.orderService.UpdateOrderStatus(ctx, merchantID, req.PaymentNo, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新订单状态失败: %v", err)
	}

	var paidAt *timestamppb.Timestamp
	if order.PaidAt != nil {
		paidAt = timestamppb.New(*order.PaidAt)
	}

	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:         order.ID.String(),
			MerchantId: order.MerchantID.String(),
			OrderNo:    order.OrderNo,
			PaymentNo:  order.PaymentNo,
			Amount:     order.Amount,
			Currency:   order.Currency,
			Status:     order.Status,
			PaidAt:     paidAt,
			CreatedAt:  timestamppb.New(order.CreatedAt),
			UpdatedAt:  timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// ListOrders 订单列表
func (s *OrderServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

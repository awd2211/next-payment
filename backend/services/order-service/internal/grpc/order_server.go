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

	var customerID uuid.UUID
	if req.CustomerId != "" {
		customerID, err = uuid.Parse(req.CustomerId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的客户ID")
		}
	}

	// 简化的订单创建 - 使用 proto 中定义的字段
	// TODO: 需要更新 proto 定义以支持完整的订单字段
	input := &service.CreateOrderInput{
		MerchantID:    merchantID,
		CustomerID:    customerID,
		CustomerEmail: "", // proto 中没有这个字段
		CustomerName:  "", // proto 中没有这个字段
		CustomerIP:    req.ClientIp,
		Currency:      req.Currency,
		Items:         []service.OrderItemInput{}, // proto 中没有详细的商品信息
		Remark:        req.Description,
		Extra:         make(map[string]interface{}),
	}

	// 将 metadata 转换到 extra
	for k, v := range req.Metadata {
		input.Extra[k] = v
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
			Id:          order.ID.String(),
			MerchantId:  order.MerchantID.String(),
			OrderNo:     order.OrderNo,
			Amount:      order.TotalAmount,
			Currency:    order.Currency,
			Status:      order.Status,
			Description: req.Description,
			PaidAt:      paidAt,
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// GetOrder 获取订单
func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	// TODO: proto 定义与 service 实现不匹配，需要重新设计
	return nil, status.Errorf(codes.Unimplemented, "方法未实现 - proto 定义需要更新")
}

// UpdateOrderStatus 更新订单状态
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderResponse, error) {
	// TODO: proto 定义与 service 实现不匹配，需要重新设计
	return nil, status.Errorf(codes.Unimplemented, "方法未实现 - proto 定义需要更新")
}

// ListOrders 订单列表
func (s *OrderServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

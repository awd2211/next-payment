package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/order-service/internal/model"
	"payment-platform/order-service/internal/repository"
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

	// 完整的订单创建 - 使用 proto 中更新后的字段
	input := &service.CreateOrderInput{
		MerchantID:      merchantID,
		CustomerID:      customerID,
		CustomerEmail:   req.CustomerEmail,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		CustomerIP:      req.ClientIp,
		Currency:        req.Currency,
		Language:        req.Language,
		Items:           convertProtoItemsToService(req.Items),
		ShippingMethod:  req.ShippingMethod,
		ShippingFee:     req.ShippingFee,
		ShippingAddress: convertProtoAddressToModel(req.ShippingAddress),
		BillingAddress:  convertProtoAddressToModel(req.BillingAddress),
		DiscountAmount:  req.DiscountAmount,
		Remark:          req.Remark,
		Extra:           make(map[string]interface{}),
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
	order, err := s.orderService.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "订单不存在: %v", err)
	}

	var paidAt *timestamppb.Timestamp
	if order.PaidAt != nil {
		paidAt = timestamppb.New(*order.PaidAt)
	}

	var cancelledAt *timestamppb.Timestamp
	if order.CancelledAt != nil {
		cancelledAt = timestamppb.New(*order.CancelledAt)
	}

	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:            order.ID.String(),
			OrderNo:       order.OrderNo,
			MerchantId:    order.MerchantID.String(),
			CustomerId:    order.CustomerID.String(),
			Amount:        order.TotalAmount,
			Currency:      order.Currency,
			Status:        order.Status,
			PaymentId:     order.PaymentNo,
			Description:   order.Remark,
			ClientIp:      order.CustomerIP,
			PaidAt:        paidAt,
			CancelledAt:   cancelledAt,
			CreatedAt:     timestamppb.New(order.CreatedAt),
			UpdatedAt:     timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderResponse, error) {
	// 使用system作为操作人
	systemID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	err := s.orderService.UpdateOrderStatus(ctx, req.Id, req.Status, systemID, "system")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新订单状态失败: %v", err)
	}

	// 获取更新后的订单
	order, err := s.orderService.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取订单失败: %v", err)
	}

	var paidAt *timestamppb.Timestamp
	if order.PaidAt != nil {
		paidAt = timestamppb.New(*order.PaidAt)
	}

	var cancelledAt *timestamppb.Timestamp
	if order.CancelledAt != nil {
		cancelledAt = timestamppb.New(*order.CancelledAt)
	}

	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:            order.ID.String(),
			OrderNo:       order.OrderNo,
			MerchantId:    order.MerchantID.String(),
			CustomerId:    order.CustomerID.String(),
			Amount:        order.TotalAmount,
			Currency:      order.Currency,
			Status:        order.Status,
			PaymentId:     order.PaymentNo,
			Description:   order.Remark,
			ClientIp:      order.CustomerIP,
			PaidAt:        paidAt,
			CancelledAt:   cancelledAt,
			CreatedAt:     timestamppb.New(order.CreatedAt),
			UpdatedAt:     timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// ListOrders 订单列表
func (s *OrderServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	// 构建查询条件
	query := &repository.OrderQuery{
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
		customerID, err := uuid.Parse(req.CustomerId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "无效的客户ID")
		}
		query.CustomerID = &customerID
	}

	if req.StartTime != nil {
		startTime := req.StartTime.AsTime()
		query.StartTime = &startTime
	}

	if req.EndTime != nil {
		endTime := req.EndTime.AsTime()
		query.EndTime = &endTime
	}

	orders, total, err := s.orderService.QueryOrders(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	pbOrders := make([]*pb.Order, len(orders))
	for i, o := range orders {
		var paidAt *timestamppb.Timestamp
		if o.PaidAt != nil {
			paidAt = timestamppb.New(*o.PaidAt)
		}

		var cancelledAt *timestamppb.Timestamp
		if o.CancelledAt != nil {
			cancelledAt = timestamppb.New(*o.CancelledAt)
		}

		pbOrders[i] = &pb.Order{
			Id:            o.ID.String(),
			OrderNo:       o.OrderNo,
			MerchantId:    o.MerchantID.String(),
			CustomerId:    o.CustomerID.String(),
			Amount:        o.TotalAmount,
			Currency:      o.Currency,
			Status:        o.Status,
			PaymentId:     o.PaymentNo,
			Description:   o.Remark,
			ClientIp:      o.CustomerIP,
			PaidAt:        paidAt,
			CancelledAt:   cancelledAt,
			CreatedAt:     timestamppb.New(o.CreatedAt),
			UpdatedAt:     timestamppb.New(o.UpdatedAt),
		}
	}

	return &pb.ListOrdersResponse{
		Orders:   pbOrders,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CancelOrder 取消订单
func (s *OrderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	// 使用system作为操作人
	systemID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	err := s.orderService.CancelOrder(ctx, req.Id, req.Reason, systemID, "system")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "取消订单失败: %v", err)
	}

	// 获取更新后的订单
	order, err := s.orderService.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取订单失败: %v", err)
	}

	var paidAt *timestamppb.Timestamp
	if order.PaidAt != nil {
		paidAt = timestamppb.New(*order.PaidAt)
	}

	var cancelledAt *timestamppb.Timestamp
	if order.CancelledAt != nil {
		cancelledAt = timestamppb.New(*order.CancelledAt)
	}

	return &pb.OrderResponse{
		Order: &pb.Order{
			Id:            order.ID.String(),
			OrderNo:       order.OrderNo,
			MerchantId:    order.MerchantID.String(),
			CustomerId:    order.CustomerID.String(),
			Amount:        order.TotalAmount,
			Currency:      order.Currency,
			Status:        order.Status,
			PaymentId:     order.PaymentNo,
			Description:   order.Remark,
			ClientIp:      order.CustomerIP,
			PaidAt:        paidAt,
			CancelledAt:   cancelledAt,
			CreatedAt:     timestamppb.New(order.CreatedAt),
			UpdatedAt:     timestamppb.New(order.UpdatedAt),
		},
	}, nil
}

// GetOrderStats 获取订单统计
func (s *OrderServer) GetOrderStats(ctx context.Context, req *pb.GetOrderStatsRequest) (*pb.OrderStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// ========== 辅助转换函数 ==========

// convertProtoItemsToService 将proto商品列表转换为service层输入
func convertProtoItemsToService(protoItems []*pb.OrderItem) []service.OrderItemInput {
	if len(protoItems) == 0 {
		return []service.OrderItemInput{}
	}

	items := make([]service.OrderItemInput, 0, len(protoItems))
	for _, item := range protoItems {
		serviceItem := service.OrderItemInput{
			ProductID:    item.ProductId,
			ProductName:  item.ProductName,
			ProductSKU:   item.ProductSku,
			ProductImage: item.ProductImage,
			UnitPrice:    item.UnitPrice,
			Quantity:     int(item.Quantity),
			Attributes:   make(map[string]interface{}),
			Extra:        make(map[string]interface{}),
		}

		// 转换attributes
		for k, v := range item.Attributes {
			serviceItem.Attributes[k] = v
		}

		items = append(items, serviceItem)
	}

	return items
}

// convertProtoAddressToModel 将proto地址转换为model层Address
func convertProtoAddressToModel(protoAddr *pb.Address) *model.Address {
	if protoAddr == nil {
		return nil
	}

	return &model.Address{
		Country:    protoAddr.Country,
		Province:   protoAddr.Province,
		City:       protoAddr.City,
		District:   protoAddr.District,
		Street:     protoAddr.Street,
		PostalCode: protoAddr.PostalCode,
		Name:       protoAddr.RecipientName,
		Phone:      protoAddr.RecipientPhone,
	}
}

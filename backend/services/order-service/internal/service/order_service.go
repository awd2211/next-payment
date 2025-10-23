package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/order-service/internal/model"
	"github.com/payment-platform/services/order-service/internal/repository"
)

// OrderService 订单服务接口
type OrderService interface {
	// 订单管理
	CreateOrder(ctx context.Context, input *CreateOrderInput) (*model.Order, error)
	GetOrder(ctx context.Context, orderNo string) (*model.Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	QueryOrders(ctx context.Context, query *repository.OrderQuery) ([]*model.Order, int64, error)
	CancelOrder(ctx context.Context, orderNo string, reason string, operatorID uuid.UUID, operatorType string) error
	UpdateOrderStatus(ctx context.Context, orderNo string, status string, operatorID uuid.UUID, operatorType string) error

	// 支付相关
	PayOrder(ctx context.Context, orderNo string, paymentNo string) error
	RefundOrder(ctx context.Context, orderNo string, amount int64, reason string) error

	// 配送相关
	ShipOrder(ctx context.Context, orderNo string, shippingInfo map[string]interface{}) error
	CompleteOrder(ctx context.Context, orderNo string) error

	// 统计分析
	GetOrderStatistics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time, currency string) ([]*model.OrderStatistics, error)
	GetDailySummary(ctx context.Context, merchantID uuid.UUID, date time.Time) (map[string]interface{}, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

// NewOrderService 创建订单服务实例
func NewOrderService(orderRepo repository.OrderRepository) OrderService {
	return &orderService{
		orderRepo: orderRepo,
	}
}

// CreateOrderInput 创建订单输入
type CreateOrderInput struct {
	MerchantID      uuid.UUID             `json:"merchant_id" binding:"required"`
	CustomerID      uuid.UUID             `json:"customer_id"`
	CustomerEmail   string                `json:"customer_email" binding:"required,email"`
	CustomerName    string                `json:"customer_name" binding:"required"`
	CustomerPhone   string                `json:"customer_phone"`
	CustomerIP      string                `json:"customer_ip"`
	Currency        string                `json:"currency" binding:"required"`
	Language        string                `json:"language"`
	Items           []OrderItemInput      `json:"items" binding:"required,min=1"`
	ShippingMethod  string                `json:"shipping_method"`
	ShippingFee     int64                 `json:"shipping_fee"`
	ShippingAddress *model.Address        `json:"shipping_address"`
	BillingAddress  *model.Address        `json:"billing_address"`
	DiscountAmount  int64                 `json:"discount_amount"`
	Remark          string                `json:"remark"`
	Extra           map[string]interface{} `json:"extra"`
	ExpireMinutes   int                   `json:"expire_minutes"` // 订单过期时间（分钟）
}

// OrderItemInput 订单项输入
type OrderItemInput struct {
	ProductID    string                 `json:"product_id" binding:"required"`
	ProductName  string                 `json:"product_name" binding:"required"`
	ProductSKU   string                 `json:"product_sku"`
	ProductImage string                 `json:"product_image"`
	UnitPrice    int64                  `json:"unit_price" binding:"required,gt=0"`
	Quantity     int                    `json:"quantity" binding:"required,gt=0"`
	Attributes   map[string]interface{} `json:"attributes"`
	Extra        map[string]interface{} `json:"extra"`
}

// CreateOrder 创建订单
func (s *orderService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*model.Order, error) {
	// 生成订单号
	orderNo := s.generateOrderNo()

	// 计算订单金额
	var totalAmount int64 = 0
	for _, item := range input.Items {
		totalAmount += item.UnitPrice * int64(item.Quantity)
	}

	// 加上运费
	totalAmount += input.ShippingFee

	// 减去优惠金额
	payAmount := totalAmount - input.DiscountAmount
	if payAmount < 0 {
		payAmount = 0
	}

	// 序列化地址信息
	var shippingAddressJSON string
	var billingAddressJSON string
	if input.ShippingAddress != nil {
		addrBytes, _ := json.Marshal(input.ShippingAddress)
		shippingAddressJSON = string(addrBytes)
	}
	if input.BillingAddress != nil {
		addrBytes, _ := json.Marshal(input.BillingAddress)
		billingAddressJSON = string(addrBytes)
	}

	// 序列化扩展信息
	var extraJSON string
	if input.Extra != nil {
		extraBytes, _ := json.Marshal(input.Extra)
		extraJSON = string(extraBytes)
	}

	// 计算过期时间
	var expiredAt *time.Time
	if input.ExpireMinutes > 0 {
		t := time.Now().Add(time.Duration(input.ExpireMinutes) * time.Minute)
		expiredAt = &t
	}

	// 创建订单
	order := &model.Order{
		MerchantID:      input.MerchantID,
		OrderNo:         orderNo,
		TotalAmount:     totalAmount,
		PayAmount:       payAmount,
		DiscountAmount:  input.DiscountAmount,
		ShippingFee:     input.ShippingFee,
		Currency:        input.Currency,
		Status:          model.OrderStatusPending,
		PayStatus:       model.PayStatusPending,
		ShippingStatus:  model.ShippingStatusPending,
		CustomerID:      input.CustomerID,
		CustomerEmail:   input.CustomerEmail,
		CustomerName:    input.CustomerName,
		CustomerPhone:   input.CustomerPhone,
		CustomerIP:      input.CustomerIP,
		ShippingMethod:  input.ShippingMethod,
		ShippingAddress: shippingAddressJSON,
		BillingAddress:  billingAddressJSON,
		Remark:          input.Remark,
		Extra:           extraJSON,
		Language:        input.Language,
		ExpiredAt:       expiredAt,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 创建订单项
	var items []*model.OrderItem
	for _, itemInput := range input.Items {
		totalPrice := itemInput.UnitPrice * int64(itemInput.Quantity)

		var attributesJSON string
		if itemInput.Attributes != nil {
			attrBytes, _ := json.Marshal(itemInput.Attributes)
			attributesJSON = string(attrBytes)
		}

		var itemExtraJSON string
		if itemInput.Extra != nil {
			extraBytes, _ := json.Marshal(itemInput.Extra)
			itemExtraJSON = string(extraBytes)
		}

		item := &model.OrderItem{
			OrderID:      order.ID,
			ProductID:    itemInput.ProductID,
			ProductName:  itemInput.ProductName,
			ProductSKU:   itemInput.ProductSKU,
			ProductImage: itemInput.ProductImage,
			UnitPrice:    itemInput.UnitPrice,
			Quantity:     itemInput.Quantity,
			TotalPrice:   totalPrice,
			Attributes:   attributesJSON,
			Extra:        itemExtraJSON,
		}
		items = append(items, item)
	}

	if err := s.orderRepo.CreateItems(ctx, items); err != nil {
		return nil, fmt.Errorf("创建订单项失败: %w", err)
	}

	// 记录订单日志
	s.createOrderLog(ctx, order.ID, model.OrderActionCreate, "", model.OrderStatusPending, uuid.Nil, model.OperatorTypeSystem, "订单创建")

	// 加载订单项
	order.Items = items

	return order, nil
}

// GetOrder 获取订单
func (s *orderService) GetOrder(ctx context.Context, orderNo string) (*model.Order, error) {
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("订单不存在")
	}
	return order, nil
}

// GetOrderByID 根据ID获取订单
func (s *orderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("订单不存在")
	}
	return order, nil
}

// QueryOrders 查询订单列表
func (s *orderService) QueryOrders(ctx context.Context, query *repository.OrderQuery) ([]*model.Order, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	return s.orderRepo.List(ctx, query)
}

// CancelOrder 取消订单
func (s *orderService) CancelOrder(ctx context.Context, orderNo string, reason string, operatorID uuid.UUID, operatorType string) error {
	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return err
	}

	// 只有待支付或已支付的订单可以取消
	if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusPaid {
		return fmt.Errorf("当前状态不允许取消: %s", order.Status)
	}

	// 如果已支付，需要先退款
	if order.PayStatus == model.PayStatusPaid {
		return fmt.Errorf("已支付订单需要先申请退款")
	}

	oldStatus := order.Status
	if err := s.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusCancelled); err != nil {
		return fmt.Errorf("取消订单失败: %w", err)
	}

	// 记录日志
	s.createOrderLog(ctx, order.ID, model.OrderActionCancel, oldStatus, model.OrderStatusCancelled, operatorID, operatorType, reason)

	return nil
}

// UpdateOrderStatus 更新订单状态
func (s *orderService) UpdateOrderStatus(ctx context.Context, orderNo string, status string, operatorID uuid.UUID, operatorType string) error {
	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return err
	}

	oldStatus := order.Status
	if err := s.orderRepo.UpdateStatus(ctx, order.ID, status); err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 记录日志
	s.createOrderLog(ctx, order.ID, model.OrderActionUpdateStatus, oldStatus, status, operatorID, operatorType, fmt.Sprintf("状态从 %s 更新为 %s", oldStatus, status))

	return nil
}

// PayOrder 支付订单
func (s *orderService) PayOrder(ctx context.Context, orderNo string, paymentNo string) error {
	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return err
	}

	if order.Status != model.OrderStatusPending {
		return fmt.Errorf("订单状态不正确，无法支付")
	}

	paidAt := time.Now()
	order.PaymentNo = paymentNo

	if err := s.orderRepo.UpdatePayStatus(ctx, order.ID, model.PayStatusPaid, &paidAt); err != nil {
		return fmt.Errorf("更新支付状态失败: %w", err)
	}

	// 更新订单
	order.PaymentNo = paymentNo
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("更新订单失败: %w", err)
	}

	// 记录日志
	s.createOrderLog(ctx, order.ID, model.OrderActionPay, model.OrderStatusPending, model.OrderStatusPaid, uuid.Nil, model.OperatorTypeSystem, fmt.Sprintf("支付成功，支付流水号: %s", paymentNo))

	return nil
}

// RefundOrder 退款订单
func (s *orderService) RefundOrder(ctx context.Context, orderNo string, amount int64, reason string) error {
	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return err
	}

	if order.PayStatus != model.PayStatusPaid {
		return fmt.Errorf("订单未支付，无法退款")
	}

	// 部分退款还是全额退款
	var newPayStatus string
	if amount >= order.PayAmount {
		newPayStatus = model.PayStatusRefunded
		order.Status = model.OrderStatusRefunded
	} else {
		newPayStatus = model.PayStatusPartialRefunded
	}

	if err := s.orderRepo.UpdatePayStatus(ctx, order.ID, newPayStatus, nil); err != nil {
		return fmt.Errorf("更新支付状态失败: %w", err)
	}

	if newPayStatus == model.PayStatusRefunded {
		s.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusRefunded)
	}

	// 记录日志
	s.createOrderLog(ctx, order.ID, model.OrderActionRefund, order.Status, order.Status, uuid.Nil, model.OperatorTypeSystem, fmt.Sprintf("退款金额: %d, 原因: %s", amount, reason))

	return nil
}

// ShipOrder 发货
func (s *orderService) ShipOrder(ctx context.Context, orderNo string, shippingInfo map[string]interface{}) error {
	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return err
	}

	if order.Status != model.OrderStatusPaid {
		return fmt.Errorf("订单未支付，无法发货")
	}

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusShipped); err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	if err := s.orderRepo.UpdateShippingStatus(ctx, order.ID, model.ShippingStatusShipped); err != nil {
		return fmt.Errorf("更新配送状态失败: %w", err)
	}

	// 记录日志
	infoJSON, _ := json.Marshal(shippingInfo)
	s.createOrderLog(ctx, order.ID, model.OrderActionShip, model.OrderStatusPaid, model.OrderStatusShipped, uuid.Nil, model.OperatorTypeSystem, fmt.Sprintf("订单已发货: %s", string(infoJSON)))

	return nil
}

// CompleteOrder 完成订单
func (s *orderService) CompleteOrder(ctx context.Context, orderNo string) error {
	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return err
	}

	if order.Status != model.OrderStatusShipped {
		return fmt.Errorf("订单未发货，无法完成")
	}

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusCompleted); err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	if err := s.orderRepo.UpdateShippingStatus(ctx, order.ID, model.ShippingStatusDelivered); err != nil {
		return fmt.Errorf("更新配送状态失败: %w", err)
	}

	// 记录日志
	s.createOrderLog(ctx, order.ID, model.OrderActionComplete, model.OrderStatusShipped, model.OrderStatusCompleted, uuid.Nil, model.OperatorTypeSystem, "订单已完成")

	return nil
}

// GetOrderStatistics 获取订单统计
func (s *orderService) GetOrderStatistics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time, currency string) ([]*model.OrderStatistics, error) {
	return s.orderRepo.GetStatistics(ctx, merchantID, startDate, endDate, currency)
}

// GetDailySummary 获取每日汇总
func (s *orderService) GetDailySummary(ctx context.Context, merchantID uuid.UUID, date time.Time) (map[string]interface{}, error) {
	return s.orderRepo.GetDailySummary(ctx, merchantID, date)
}

// 工具函数

// createOrderLog 创建订单日志
func (s *orderService) createOrderLog(ctx context.Context, orderID uuid.UUID, action, oldStatus, newStatus string, operatorID uuid.UUID, operatorType, remark string) {
	log := &model.OrderLog{
		OrderID:      orderID,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		OperatorID:   operatorID,
		OperatorType: operatorType,
		Remark:       remark,
	}
	s.orderRepo.CreateLog(context.Background(), log)
}

// generateOrderNo 生成订单号
func (s *orderService) generateOrderNo() string {
	// 格式：OD + 时间戳 + 随机字符
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("OD%s%s", timestamp, randomStr)
}

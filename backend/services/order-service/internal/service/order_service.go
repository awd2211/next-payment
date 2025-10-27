package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/events"
	"github.com/payment-platform/pkg/idempotent"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/order-service/internal/client"
	"payment-platform/order-service/internal/model"
	"payment-platform/order-service/internal/repository"
)

// OrderService 订单服务接口
type OrderService interface {
	// 订单管理
	CreateOrder(ctx context.Context, input *CreateOrderInput) (*model.Order, error)
	GetOrder(ctx context.Context, orderNo string) (*model.Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetOrderByPaymentNo(ctx context.Context, paymentNo string) (*model.Order, error)
	QueryOrders(ctx context.Context, query *repository.OrderQuery) ([]*model.Order, int64, error)
	BatchGetOrders(ctx context.Context, orderNos []string, merchantID uuid.UUID) (map[string]*model.Order, []string, error)
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
	GetOrderStats(ctx context.Context) (map[string]interface{}, error)
}

type orderService struct {
	db                 *gorm.DB
	orderRepo          repository.OrderRepository
	redisClient        *redis.Client
	idempotentService  idempotent.Service
	notificationClient *client.NotificationClient // 保留作为降级方案
	eventPublisher     *kafka.EventPublisher      // 事件发布器
}

// NewOrderService 创建订单服务实例
func NewOrderService(db *gorm.DB, orderRepo repository.OrderRepository, redisClient *redis.Client, notificationClient *client.NotificationClient, eventPublisher *kafka.EventPublisher) OrderService {
	return &orderService{
		db:                 db,
		orderRepo:          orderRepo,
		redisClient:        redisClient,
		idempotentService:  idempotent.NewService(redisClient),
		notificationClient: notificationClient,
		eventPublisher:     eventPublisher,
	}
}

// CreateOrderInput 创建订单输入
type CreateOrderInput struct {
	MerchantID      uuid.UUID             `json:"merchant_id" binding:"required"`
	OrderNo         string                `json:"order_no"`         // 订单号（可选，用于幂等性）
	PaymentNo       string                `json:"payment_no"`       // 支付流水号（可选，用于幂等性）
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

// CreateOrder 创建订单（使用事务保护订单和订单项的完整性 + 幂等性保护）
func (s *orderService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*model.Order, error) {
	logger.Info("creating order",
		zap.String("merchant_id", input.MerchantID.String()),
		zap.String("customer_email", input.CustomerEmail),
		zap.String("currency", input.Currency),
		zap.Int("item_count", len(input.Items)))

	// 【幂等性保护】1. 如果有 PaymentNo（从 payment-gateway 调用），使用它作为幂等性键
	if input.PaymentNo != "" {
		idempotentKey := idempotent.GenerateKey("order", input.MerchantID.String(), input.PaymentNo)

		// 2. 检查是否已处理
		type OrderIdempotentResult struct {
			OrderNo string `json:"order_no"`
			OrderID string `json:"order_id"`
		}
		var cachedResult OrderIdempotentResult
		exists, err := s.idempotentService.Check(ctx, idempotentKey, &cachedResult)
		if err != nil {
			logger.Warn("order idempotent check failed, continue processing",
				zap.Error(err),
				zap.String("payment_no", input.PaymentNo))
		} else if exists {
			logger.Info("idempotent order request detected, returning cached result",
				zap.String("payment_no", input.PaymentNo),
				zap.String("order_no", cachedResult.OrderNo))

			// 从数据库查询订单
			order, err := s.GetOrder(ctx, cachedResult.OrderNo)
			if err != nil {
				return nil, fmt.Errorf("查询缓存的订单失败: %w", err)
			}
			return order, nil
		}

		// 3. 获取分布式锁
		lockAcquired, err := s.idempotentService.Try(ctx, idempotentKey, 30*time.Second)
		if err != nil {
			logger.Warn("failed to acquire order lock, continue processing",
				zap.Error(err),
				zap.String("payment_no", input.PaymentNo))
		} else if !lockAcquired {
			return nil, fmt.Errorf("该订单正在创建中，请稍后查询")
		}

		// 4. 释放锁
		defer func() {
			if lockAcquired {
				if err := s.idempotentService.Release(ctx, idempotentKey); err != nil {
					logger.Error("failed to release order lock",
						zap.Error(err),
						zap.String("payment_no", input.PaymentNo))
				}
			}
		}()
	}

	// 生成订单号（如果没有提供）
	orderNo := input.OrderNo
	if orderNo == "" {
		orderNo = s.generateOrderNo()
	}

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
		addrBytes, err := json.Marshal(input.ShippingAddress)
		if err != nil {
			logger.Error("failed to marshal shipping address",
				zap.Error(err),
				zap.String("merchant_id", input.MerchantID.String()))
			return nil, fmt.Errorf("序列化收货地址失败: %w", err)
		}
		shippingAddressJSON = string(addrBytes)
	}
	if input.BillingAddress != nil {
		addrBytes, err := json.Marshal(input.BillingAddress)
		if err != nil {
			logger.Error("failed to marshal billing address",
				zap.Error(err),
				zap.String("merchant_id", input.MerchantID.String()))
			return nil, fmt.Errorf("序列化账单地址失败: %w", err)
		}
		billingAddressJSON = string(addrBytes)
	}

	// 序列化扩展信息
	var extraJSON string
	if input.Extra != nil {
		extraBytes, err := json.Marshal(input.Extra)
		if err != nil {
			logger.Error("failed to marshal extra data",
				zap.Error(err),
				zap.String("merchant_id", input.MerchantID.String()))
			return nil, fmt.Errorf("序列化扩展信息失败: %w", err)
		}
		extraJSON = string(extraBytes)
	}

	// 计算过期时间
	var expiredAt *time.Time
	if input.ExpireMinutes > 0 {
		t := time.Now().Add(time.Duration(input.ExpireMinutes) * time.Minute)
		expiredAt = &t
	}

	// 准备订单数据
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

	// 准备订单项数据
	var items []*model.OrderItem
	for _, itemInput := range input.Items {
		totalPrice := itemInput.UnitPrice * int64(itemInput.Quantity)

		var attributesJSON string
		if itemInput.Attributes != nil {
			attrBytes, err := json.Marshal(itemInput.Attributes)
			if err != nil {
				logger.Error("failed to marshal item attributes",
					zap.Error(err),
					zap.String("product_id", itemInput.ProductID))
				return nil, fmt.Errorf("序列化商品属性失败: %w", err)
			}
			attributesJSON = string(attrBytes)
		}

		var itemExtraJSON string
		if itemInput.Extra != nil {
			extraBytes, err := json.Marshal(itemInput.Extra)
			if err != nil {
				logger.Error("failed to marshal item extra data",
					zap.Error(err),
					zap.String("product_id", itemInput.ProductID))
				return nil, fmt.Errorf("序列化商品扩展信息失败: %w", err)
			}
			itemExtraJSON = string(extraBytes)
		}

		item := &model.OrderItem{
			// OrderID 将在事务中设置
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

	// 在事务中创建订单、订单项和日志，确保原子性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建订单
		if err := tx.Create(order).Error; err != nil {
			logger.Error("failed to create order in database",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 2. 设置订单项的 OrderID 并创建
		for _, item := range items {
			item.OrderID = order.ID
			if err := tx.Create(item).Error; err != nil {
				logger.Error("failed to create order item",
					zap.Error(err),
					zap.String("order_no", orderNo),
					zap.String("product_id", item.ProductID))
				return fmt.Errorf("创建订单项失败: %w", err)
			}
		}

		// 3. 创建订单日志
		log := &model.OrderLog{
			OrderID:      order.ID,
			Action:       model.OrderActionCreate,
			OldStatus:    "",
			NewStatus:    model.OrderStatusPending,
			OperatorID:   uuid.Nil,
			OperatorType: model.OperatorTypeSystem,
			Remark:       "订单创建",
		}
		if err := tx.Create(log).Error; err != nil {
			logger.Error("failed to create order log",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("创建订单日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("order creation transaction failed",
			zap.Error(err),
			zap.String("merchant_id", input.MerchantID.String()),
			zap.String("order_no", orderNo))
		return nil, err
	}

	// 加载订单项到订单对象
	order.Items = items

	logger.Info("order created successfully",
		zap.String("order_no", orderNo),
		zap.Int64("total_amount", totalAmount),
		zap.Int64("pay_amount", payAmount),
		zap.String("currency", input.Currency))

	// 发布订单创建事件 (异步,不阻塞主流程)
	s.publishOrderEvent(ctx, events.OrderCreated, order)

	// 【幂等性保护】5. 缓存成功结果（如果有 PaymentNo）
	if input.PaymentNo != "" {
		idempotentKey := idempotent.GenerateKey("order", input.MerchantID.String(), input.PaymentNo)
		cacheResult := map[string]string{
			"order_no": order.OrderNo,
			"order_id": order.ID.String(),
		}
		if err := s.idempotentService.Store(ctx, idempotentKey, cacheResult, 24*time.Hour); err != nil {
			logger.Error("failed to cache order idempotent result",
				zap.Error(err),
				zap.String("payment_no", input.PaymentNo),
				zap.String("order_no", order.OrderNo))
			// 缓存失败不影响订单创建结果
		}
	}

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

// GetOrderByPaymentNo 根据支付流水号获取订单
func (s *orderService) GetOrderByPaymentNo(ctx context.Context, paymentNo string) (*model.Order, error) {
	order, err := s.orderRepo.GetByPaymentNo(ctx, paymentNo)
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

// BatchGetOrders 批量查询订单
// 返回: (成功查询的订单map, 查询失败的orderNo列表, error)
func (s *orderService) BatchGetOrders(ctx context.Context, orderNos []string, merchantID uuid.UUID) (map[string]*model.Order, []string, error) {
	// 验证请求
	if len(orderNos) == 0 {
		return nil, nil, fmt.Errorf("订单号列表不能为空")
	}
	if len(orderNos) > 100 {
		return nil, nil, fmt.Errorf("批量查询最多支持100个订单号")
	}

	// 使用 map 去重
	uniqueOrderNos := make(map[string]bool)
	for _, orderNo := range orderNos {
		if orderNo != "" {
			uniqueOrderNos[orderNo] = true
		}
	}

	results := make(map[string]*model.Order)
	failed := make([]string, 0)

	// 批量查询（使用 WHERE IN 子句）
	var orders []*model.Order
	err := s.db.WithContext(ctx).
		Where("order_no IN ? AND merchant_id = ?", orderNos, merchantID).
		Find(&orders).Error

	if err != nil {
		logger.Error("批量查询订单失败", zap.Error(err))
		return nil, nil, fmt.Errorf("批量查询订单失败: %w", err)
	}

	// 构建结果 map
	for _, order := range orders {
		results[order.OrderNo] = order
		delete(uniqueOrderNos, order.OrderNo)
	}

	// 未找到的订单号记录为 failed
	for orderNo := range uniqueOrderNos {
		failed = append(failed, orderNo)
	}

	logger.Info("批量查询订单完成",
		zap.Int("total_requested", len(orderNos)),
		zap.Int("found", len(results)),
		zap.Int("not_found", len(failed)))

	return results, failed, nil
}

// CancelOrder 取消订单（使用事务保证状态更新和日志记录的原子性）
func (s *orderService) CancelOrder(ctx context.Context, orderNo string, reason string, operatorID uuid.UUID, operatorType string) error {
	logger.Info("cancelling order",
		zap.String("order_no", orderNo),
		zap.String("operator_id", operatorID.String()),
		zap.String("operator_type", operatorType))

	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		logger.Error("failed to get order for cancellation",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	// 只有待支付或已支付的订单可以取消
	if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusPaid {
		logger.Warn("order status does not allow cancellation",
			zap.String("order_no", orderNo),
			zap.String("status", order.Status))
		return fmt.Errorf("当前状态不允许取消: %s", order.Status)
	}

	// 如果已支付，需要先退款
	if order.PayStatus == model.PayStatusPaid {
		logger.Warn("paid order requires refund before cancellation",
			zap.String("order_no", orderNo))
		return fmt.Errorf("已支付订单需要先申请退款")
	}

	oldStatus := order.Status

	// 使用事务保证状态更新和日志记录的原子性
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新订单状态
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).
			Update("status", model.OrderStatusCancelled).Error; err != nil {
			logger.Error("failed to update order status",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 2. 记录日志
		log := &model.OrderLog{
			OrderID:      order.ID,
			Action:       model.OrderActionCancel,
			OldStatus:    oldStatus,
			NewStatus:    model.OrderStatusCancelled,
			OperatorID:   operatorID,
			OperatorType: operatorType,
			Remark:       reason,
		}
		if err := tx.Create(log).Error; err != nil {
			logger.Error("failed to create cancellation log",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("创建日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("order cancellation transaction failed",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	logger.Info("order cancelled successfully",
		zap.String("order_no", orderNo),
		zap.String("old_status", oldStatus))

	return nil
}

// UpdateOrderStatus 更新订单状态（使用事务保证一致性）
func (s *orderService) UpdateOrderStatus(ctx context.Context, orderNo string, status string, operatorID uuid.UUID, operatorType string) error {
	logger.Info("updating order status",
		zap.String("order_no", orderNo),
		zap.String("new_status", status),
		zap.String("operator_id", operatorID.String()))

	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		logger.Error("failed to get order for status update",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	oldStatus := order.Status

	// 使用事务保证状态更新和日志记录的原子性
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新状态
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).
			Update("status", status).Error; err != nil {
			logger.Error("failed to update order status",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 2. 记录日志
		log := &model.OrderLog{
			OrderID:      order.ID,
			Action:       model.OrderActionUpdateStatus,
			OldStatus:    oldStatus,
			NewStatus:    status,
			OperatorID:   operatorID,
			OperatorType: operatorType,
			Remark:       fmt.Sprintf("状态从 %s 更新为 %s", oldStatus, status),
		}
		if err := tx.Create(log).Error; err != nil {
			logger.Error("failed to create status update log",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("创建日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("order status update transaction failed",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	logger.Info("order status updated successfully",
		zap.String("order_no", orderNo),
		zap.String("old_status", oldStatus),
		zap.String("new_status", status))

	// 发送订单状态变化通知（异步）
	if s.notificationClient != nil && oldStatus != status {
		go func(o *model.Order, oldSt, newSt string) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var notifType, title, content string
			switch newSt {
			case model.OrderStatusPaid:
				notifType = "order_paid"
				title = "订单已支付"
				content = fmt.Sprintf("订单 %s 已成功支付，金额 %.2f 元", o.OrderNo, float64(o.TotalAmount)/100.0)
			case model.OrderStatusCancelled:
				notifType = "order_cancelled"
				title = "订单已取消"
				content = fmt.Sprintf("订单 %s 已取消", o.OrderNo)
			case model.OrderStatusRefunded:
				notifType = "order_refunded"
				title = "订单已退款"
				content = fmt.Sprintf("订单 %s 已退款", o.OrderNo)
			case model.OrderStatusShipped:
				notifType = "order_shipped"
				title = "订单已发货"
				content = fmt.Sprintf("订单 %s 已发货", o.OrderNo)
			case model.OrderStatusCompleted:
				notifType = "order_completed"
				title = "订单已完成"
				content = fmt.Sprintf("订单 %s 已完成", o.OrderNo)
			default:
				return // 其他状态不发送通知
			}

			err := s.notificationClient.SendOrderNotification(notifyCtx, &client.SendNotificationRequest{
				MerchantID: o.MerchantID,
				Type:       notifType,
				Title:      title,
				Content:    content,
				Email:      o.CustomerEmail,
				Priority:   "high",
				Data: map[string]interface{}{
					"order_no":     o.OrderNo,
					"order_id":     o.ID.String(),
					"total_amount": o.TotalAmount,
					"currency":     o.Currency,
					"old_status":   oldSt,
					"new_status":   newSt,
				},
			})
			if err != nil {
				logger.Warn("发送订单状态通知失败（非致命）",
					zap.Error(err),
					zap.String("order_no", o.OrderNo),
					zap.String("status", newSt))
			}
		}(order, oldStatus, status)
	}

	return nil
}

// PayOrder 支付订单（使用事务保护状态更新的原子性）
func (s *orderService) PayOrder(ctx context.Context, orderNo string, paymentNo string) error {
	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		return err
	}

	if order.Status != model.OrderStatusPending {
		return fmt.Errorf("订单状态不正确，无法支付")
	}

	paidAt := time.Now()

	// 在事务中更新支付状态、订单状态、支付流水号和日志
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 一次性更新所有订单字段（避免多次UPDATE）
		err := tx.Model(&model.Order{}).
			Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"pay_status":  model.PayStatusPaid,
				"paid_at":     &paidAt,
				"payment_no":  paymentNo,
				"status":      model.OrderStatusPaid,
				"updated_at":  time.Now(),
			}).Error
		if err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 2. 创建订单日志
		log := &model.OrderLog{
			OrderID:      order.ID,
			Action:       model.OrderActionPay,
			OldStatus:    model.OrderStatusPending,
			NewStatus:    model.OrderStatusPaid,
			OperatorID:   uuid.Nil,
			OperatorType: model.OperatorTypeSystem,
			Remark:       fmt.Sprintf("支付成功，支付流水号: %s", paymentNo),
		}
		if err := tx.Create(log).Error; err != nil {
			return fmt.Errorf("创建订单日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 发布订单支付成功事件 (异步)
	order.Status = model.OrderStatusPaid
	order.PayStatus = model.PayStatusPaid
	order.PaidAt = &paidAt
	order.PaymentNo = paymentNo
	s.publishOrderEvent(ctx, events.OrderPaid, order)

	return nil
}

// RefundOrder 退款订单（使用事务保证退款状态更新的原子性）
func (s *orderService) RefundOrder(ctx context.Context, orderNo string, amount int64, reason string) error {
	logger.Info("refunding order",
		zap.String("order_no", orderNo),
		zap.Int64("amount", amount),
		zap.String("reason", reason))

	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		logger.Error("failed to get order for refund",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	if order.PayStatus != model.PayStatusPaid {
		logger.Warn("order not paid, cannot refund",
			zap.String("order_no", orderNo),
			zap.String("pay_status", order.PayStatus))
		return fmt.Errorf("订单未支付，无法退款")
	}

	// 部分退款还是全额退款
	var newPayStatus string
	var newOrderStatus string
	if amount >= order.PayAmount {
		newPayStatus = model.PayStatusRefunded
		newOrderStatus = model.OrderStatusRefunded
	} else {
		newPayStatus = model.PayStatusPartialRefunded
		newOrderStatus = order.Status // 部分退款不改变订单状态
	}

	oldStatus := order.Status

	// 使用事务保证退款状态更新的原子性
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新支付状态
		updates := map[string]interface{}{
			"pay_status": newPayStatus,
			"updated_at": time.Now(),
		}
		if newPayStatus == model.PayStatusRefunded {
			updates["status"] = newOrderStatus
		}

		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).
			Updates(updates).Error; err != nil {
			logger.Error("failed to update order refund status",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("更新退款状态失败: %w", err)
		}

		// 2. 记录日志
		log := &model.OrderLog{
			OrderID:      order.ID,
			Action:       model.OrderActionRefund,
			OldStatus:    oldStatus,
			NewStatus:    newOrderStatus,
			OperatorID:   uuid.Nil,
			OperatorType: model.OperatorTypeSystem,
			Remark:       fmt.Sprintf("退款金额: %d, 原因: %s", amount, reason),
		}
		if err := tx.Create(log).Error; err != nil {
			logger.Error("failed to create refund log",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("创建退款日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("order refund transaction failed",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	logger.Info("order refunded successfully",
		zap.String("order_no", orderNo),
		zap.Int64("refund_amount", amount),
		zap.String("pay_status", newPayStatus))

	return nil
}

// ShipOrder 发货（使用事务保证发货状态更新的原子性）
func (s *orderService) ShipOrder(ctx context.Context, orderNo string, shippingInfo map[string]interface{}) error {
	logger.Info("shipping order",
		zap.String("order_no", orderNo))

	order, err := s.GetOrder(ctx, orderNo)
	if err != nil {
		logger.Error("failed to get order for shipping",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	if order.Status != model.OrderStatusPaid {
		logger.Warn("order not paid, cannot ship",
			zap.String("order_no", orderNo),
			zap.String("status", order.Status))
		return fmt.Errorf("订单未支付，无法发货")
	}

	// 序列化物流信息
	var shippingInfoJSON string
	if shippingInfo != nil {
		infoBytes, err := json.Marshal(shippingInfo)
		if err != nil {
			logger.Error("failed to marshal shipping info",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("序列化物流信息失败: %w", err)
		}
		shippingInfoJSON = string(infoBytes)
	}

	oldStatus := order.Status

	// 使用事务保证发货状态更新的原子性
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新订单状态和物流状态
		updates := map[string]interface{}{
			"status":          model.OrderStatusShipped,
			"shipping_status": model.ShippingStatusShipped,
			"shipping_info":   shippingInfoJSON,
			"updated_at":      time.Now(),
		}

		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).
			Updates(updates).Error; err != nil {
			logger.Error("failed to update order shipping status",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		// 2. 记录日志
		log := &model.OrderLog{
			OrderID:      order.ID,
			Action:       model.OrderActionShip,
			OldStatus:    oldStatus,
			NewStatus:    model.OrderStatusShipped,
			OperatorID:   uuid.Nil,
			OperatorType: model.OperatorTypeSystem,
			Remark:       "订单已发货",
		}
		if err := tx.Create(log).Error; err != nil {
			logger.Error("failed to create shipping log",
				zap.Error(err),
				zap.String("order_no", orderNo))
			return fmt.Errorf("创建发货日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("order shipping transaction failed",
			zap.Error(err),
			zap.String("order_no", orderNo))
		return err
	}

	logger.Info("order shipped successfully",
		zap.String("order_no", orderNo))

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

// GetOrderStats 获取全局订单统计（实时聚合数据库）
func (s *orderService) GetOrderStats(ctx context.Context) (map[string]interface{}, error) {
	var result struct {
		TotalAmount    int64 `gorm:"column:total_amount"`
		TotalCount     int64 `gorm:"column:total_count"`
		PaidCount      int64 `gorm:"column:paid_count"`
		PendingCount   int64 `gorm:"column:pending_count"`
		CancelledCount int64 `gorm:"column:cancelled_count"`
	}

	// 全局订单统计（所有商户）
	err := s.db.WithContext(ctx).
		Model(&model.Order{}).
		Select(`
			COALESCE(SUM(total_amount), 0) as total_amount,
			COUNT(*) as total_count,
			COUNT(CASE WHEN status = ? THEN 1 END) as paid_count,
			COUNT(CASE WHEN status = ? THEN 1 END) as pending_count,
			COUNT(CASE WHEN status = ? THEN 1 END) as cancelled_count
		`, model.OrderStatusPaid, model.OrderStatusPending, model.OrderStatusCancelled).
		Scan(&result).Error

	if err != nil {
		logger.Error("failed to get order stats", zap.Error(err))
		return nil, fmt.Errorf("查询订单统计失败: %w", err)
	}

	// 今日订单统计
	var todayResult struct {
		TodayAmount int64 `gorm:"column:today_amount"`
		TodayCount  int64 `gorm:"column:today_count"`
	}

	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	err = s.db.WithContext(ctx).
		Model(&model.Order{}).
		Select(`
			COALESCE(SUM(total_amount), 0) as today_amount,
			COUNT(*) as today_count
		`).
		Where("created_at >= ?", startOfDay).
		Scan(&todayResult).Error

	if err != nil {
		logger.Error("failed to get today's order stats", zap.Error(err))
		return nil, fmt.Errorf("查询今日订单统计失败: %w", err)
	}

	stats := map[string]interface{}{
		"total_amount":    result.TotalAmount,
		"total_count":     result.TotalCount,
		"paid_count":      result.PaidCount,
		"pending_count":   result.PendingCount,
		"cancelled_count": result.CancelledCount,
		"today_amount":    todayResult.TodayAmount,
		"today_count":     todayResult.TodayCount,
	}

	logger.Info("order stats retrieved",
		zap.Int64("total_count", result.TotalCount),
		zap.Int64("paid_count", result.PaidCount),
		zap.Int64("today_count", todayResult.TodayCount))

	return stats, nil
}

// publishOrderEvent 发布订单事件到Kafka
func (s *orderService) publishOrderEvent(ctx context.Context, eventType string, order *model.Order) {
	if s.eventPublisher == nil {
		logger.Warn("eventPublisher is nil, skipping event publishing",
			zap.String("event_type", eventType),
			zap.String("order_no", order.OrderNo))
		return
	}

	// 构造订单事件载荷
	payload := events.OrderEventPayload{
		OrderNo:       order.OrderNo,
		MerchantID:    order.MerchantID.String(),
		PaymentNo:     order.PaymentNo,
		TotalAmount:   order.TotalAmount,
		Currency:      order.Currency,
		Status:        order.Status,
		CustomerEmail: order.CustomerEmail,
		PaidAt:        order.PaidAt,
		Extra: map[string]interface{}{
			"pay_amount":      order.PayAmount,
			"discount_amount": order.DiscountAmount,
			"shipping_fee":    order.ShippingFee,
		},
	}

	// 创建事件
	event := events.NewOrderEvent(eventType, payload)

	// 异步发布事件 (不阻塞主流程)
	s.eventPublisher.PublishOrderEventAsync(ctx, event)

	logger.Info("order event published",
		zap.String("event_type", eventType),
		zap.String("order_no", order.OrderNo),
		zap.String("status", order.Status))
}

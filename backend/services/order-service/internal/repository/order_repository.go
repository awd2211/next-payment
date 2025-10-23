package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/order-service/internal/model"
	"gorm.io/gorm"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	// 订单管理
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error)
	GetByPaymentNo(ctx context.Context, paymentNo string) (*model.Order, error)
	List(ctx context.Context, query *OrderQuery) ([]*model.Order, int64, error)
	Update(ctx context.Context, order *model.Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdatePayStatus(ctx context.Context, id uuid.UUID, payStatus string, paidAt *time.Time) error
	UpdateShippingStatus(ctx context.Context, id uuid.UUID, shippingStatus string) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 订单项管理
	CreateItems(ctx context.Context, items []*model.OrderItem) error
	GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*model.OrderItem, error)

	// 订单日志
	CreateLog(ctx context.Context, log *model.OrderLog) error
	GetLogsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*model.OrderLog, error)

	// 订单统计
	GetStatistics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time, currency string) ([]*model.OrderStatistics, error)
	UpdateStatistics(ctx context.Context, stat *model.OrderStatistics) error
	GetDailySummary(ctx context.Context, merchantID uuid.UUID, date time.Time) (map[string]interface{}, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// OrderQuery 订单查询参数
type OrderQuery struct {
	MerchantID     *uuid.UUID
	CustomerID     *uuid.UUID
	CustomerEmail  string
	Status         string
	PayStatus      string
	ShippingStatus string
	Currency       string
	StartTime      *time.Time
	EndTime        *time.Time
	MinAmount      *int64
	MaxAmount      *int64
	Keyword        string // 订单号、客户姓名、邮箱
	Page           int
	PageSize       int
}

// Create 创建订单
func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetByID 根据ID获取订单
func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByOrderNo 根据订单号获取订单
func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "order_no = ?", orderNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByPaymentNo 根据支付流水号获取订单
func (r *orderRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "payment_no = ?", paymentNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// List 分页查询订单列表
func (r *orderRepository) List(ctx context.Context, query *OrderQuery) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Order{})

	// 构建查询条件
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.CustomerID != nil {
		db = db.Where("customer_id = ?", *query.CustomerID)
	}
	if query.CustomerEmail != "" {
		db = db.Where("customer_email = ?", query.CustomerEmail)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.PayStatus != "" {
		db = db.Where("pay_status = ?", query.PayStatus)
	}
	if query.ShippingStatus != "" {
		db = db.Where("shipping_status = ?", query.ShippingStatus)
	}
	if query.Currency != "" {
		db = db.Where("currency = ?", query.Currency)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}
	if query.MinAmount != nil {
		db = db.Where("total_amount >= ?", *query.MinAmount)
	}
	if query.MaxAmount != nil {
		db = db.Where("total_amount <= ?", *query.MaxAmount)
	}
	if query.Keyword != "" {
		db = db.Where("order_no LIKE ? OR customer_name LIKE ? OR customer_email LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Items").
		Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&orders).Error

	return orders, total, err
}

// Update 更新订单
func (r *orderRepository) Update(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// UpdateStatus 更新订单状态
func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 根据状态更新相应的时间字段
	switch status {
	case model.OrderStatusCompleted:
		updates["completed_at"] = time.Now()
	case model.OrderStatusCancelled:
		updates["cancelled_at"] = time.Now()
	case model.OrderStatusShipped:
		updates["shipped_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdatePayStatus 更新支付状态
func (r *orderRepository) UpdatePayStatus(ctx context.Context, id uuid.UUID, payStatus string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"pay_status": payStatus,
	}

	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	// 如果支付成功，同时更新订单状态为已支付
	if payStatus == model.PayStatusPaid {
		updates["status"] = model.OrderStatusPaid
		if paidAt == nil {
			updates["paid_at"] = time.Now()
		}
	}

	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateShippingStatus 更新配送状态
func (r *orderRepository) UpdateShippingStatus(ctx context.Context, id uuid.UUID, shippingStatus string) error {
	updates := map[string]interface{}{
		"shipping_status": shippingStatus,
	}

	if shippingStatus == model.ShippingStatusShipped {
		updates["shipped_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete 删除订单（软删除）
func (r *orderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Order{}, "id = ?", id).Error
}

// CreateItems 创建订单项
func (r *orderRepository) CreateItems(ctx context.Context, items []*model.OrderItem) error {
	return r.db.WithContext(ctx).Create(items).Error
}

// GetItemsByOrderID 获取订单的所有订单项
func (r *orderRepository) GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Find(&items).Error
	return items, err
}

// CreateLog 创建订单日志
func (r *orderRepository) CreateLog(ctx context.Context, log *model.OrderLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLogsByOrderID 获取订单的所有日志
func (r *orderRepository) GetLogsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*model.OrderLog, error) {
	var logs []*model.OrderLog
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetStatistics 获取订单统计数据
func (r *orderRepository) GetStatistics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time, currency string) ([]*model.OrderStatistics, error) {
	var stats []*model.OrderStatistics

	db := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate)

	if currency != "" {
		db = db.Where("currency = ?", currency)
	}

	err := db.Order("stat_date DESC").Find(&stats).Error
	return stats, err
}

// UpdateStatistics 更新统计数据
func (r *orderRepository) UpdateStatistics(ctx context.Context, stat *model.OrderStatistics) error {
	return r.db.WithContext(ctx).Save(stat).Error
}

// GetDailySummary 获取每日汇总数据
func (r *orderRepository) GetDailySummary(ctx context.Context, merchantID uuid.UUID, date time.Time) (map[string]interface{}, error) {
	var result struct {
		TotalOrders     int   `json:"total_orders"`
		PaidOrders      int   `json:"paid_orders"`
		CancelledOrders int   `json:"cancelled_orders"`
		TotalAmount     int64 `json:"total_amount"`
		PaidAmount      int64 `json:"paid_amount"`
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Select(`
			COUNT(*) as total_orders,
			COUNT(CASE WHEN pay_status = 'paid' THEN 1 END) as paid_orders,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_orders,
			COALESCE(SUM(total_amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN pay_status = 'paid' THEN pay_amount ELSE 0 END), 0) as paid_amount
		`).
		Where("merchant_id = ?", merchantID).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_orders":     result.TotalOrders,
		"paid_orders":      result.PaidOrders,
		"cancelled_orders": result.CancelledOrders,
		"total_amount":     result.TotalAmount,
		"paid_amount":      result.PaidAmount,
	}, nil
}

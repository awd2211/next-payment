package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/payment-gateway/internal/model"
	"gorm.io/gorm"
)

// PaymentRepository 支付仓储接口
type PaymentRepository interface {
	// 支付管理
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	GetByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error)
	GetByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) (*model.Payment, error)
	GetByChannelOrderNo(ctx context.Context, channelOrderNo string) (*model.Payment, error)
	List(ctx context.Context, query *PaymentQuery) ([]*model.Payment, int64, error)
	Update(ctx context.Context, payment *model.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateNotifyStatus(ctx context.Context, id uuid.UUID, status string, times int) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 退款管理
	CreateRefund(ctx context.Context, refund *model.Refund) error
	GetRefundByID(ctx context.Context, id uuid.UUID) (*model.Refund, error)
	GetRefundByRefundNo(ctx context.Context, refundNo string) (*model.Refund, error)
	ListRefunds(ctx context.Context, query *RefundQuery) ([]*model.Refund, int64, error)
	UpdateRefund(ctx context.Context, refund *model.Refund) error
	UpdateRefundStatus(ctx context.Context, id uuid.UUID, status string) error

	// 回调记录
	CreateCallback(ctx context.Context, callback *model.PaymentCallback) error
	UpdateCallback(ctx context.Context, callback *model.PaymentCallback) error
	GetCallbacksByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*model.PaymentCallback, error)

	// 路由规则
	CreateRoute(ctx context.Context, route *model.PaymentRoute) error
	GetRouteByID(ctx context.Context, id uuid.UUID) (*model.PaymentRoute, error)
	ListActiveRoutes(ctx context.Context) ([]*model.PaymentRoute, error)
	UpdateRoute(ctx context.Context, route *model.PaymentRoute) error
	DeleteRoute(ctx context.Context, id uuid.UUID) error
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建支付仓储实例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// PaymentQuery 支付查询参数
type PaymentQuery struct {
	MerchantID   *uuid.UUID
	Channel      string
	Status       string
	Currency     string
	StartTime    *time.Time
	EndTime      *time.Time
	MinAmount    *int64
	MaxAmount    *int64
	CustomerEmail string
	Keyword      string
	Page         int
	PageSize     int
}

// RefundQuery 退款查询参数
type RefundQuery struct {
	PaymentID  *uuid.UUID
	MerchantID *uuid.UUID
	Status     string
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// Create 创建支付记录
func (r *paymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// GetByID 根据ID获取支付记录
func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetByPaymentNo 根据支付流水号获取
func (r *paymentRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.WithContext(ctx).First(&payment, "payment_no = ?", paymentNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetByOrderNo 根据商户订单号获取
func (r *paymentRepository) GetByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.WithContext(ctx).
		First(&payment, "merchant_id = ? AND order_no = ?", merchantID, orderNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// GetByChannelOrderNo 根据渠道订单号获取
func (r *paymentRepository) GetByChannelOrderNo(ctx context.Context, channelOrderNo string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.WithContext(ctx).First(&payment, "channel_order_no = ?", channelOrderNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// List 分页查询支付列表
func (r *paymentRepository) List(ctx context.Context, query *PaymentQuery) ([]*model.Payment, int64, error) {
	var payments []*model.Payment
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Payment{})

	// 构建查询条件
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Channel != "" {
		db = db.Where("channel = ?", query.Channel)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
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
		db = db.Where("amount >= ?", *query.MinAmount)
	}
	if query.MaxAmount != nil {
		db = db.Where("amount <= ?", *query.MaxAmount)
	}
	if query.CustomerEmail != "" {
		db = db.Where("customer_email = ?", query.CustomerEmail)
	}
	if query.Keyword != "" {
		db = db.Where("order_no LIKE ? OR payment_no LIKE ? OR customer_name LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&payments).Error

	return payments, total, err
}

// Update 更新支付记录
func (r *paymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// UpdateStatus 更新支付状态
func (r *paymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 如果是成功状态，记录支付时间
	if status == model.PaymentStatusSuccess {
		updates["paid_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateNotifyStatus 更新通知状态
func (r *paymentRepository) UpdateNotifyStatus(ctx context.Context, id uuid.UUID, status string, times int) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"notify_status":   status,
			"notify_times":    times,
			"last_notify_at":  now,
		}).Error
}

// Delete 删除支付记录（软删除）
func (r *paymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Payment{}, "id = ?", id).Error
}

// CreateRefund 创建退款记录
func (r *paymentRepository) CreateRefund(ctx context.Context, refund *model.Refund) error {
	return r.db.WithContext(ctx).Create(refund).Error
}

// GetRefundByID 根据ID获取退款记录
func (r *paymentRepository) GetRefundByID(ctx context.Context, id uuid.UUID) (*model.Refund, error) {
	var refund model.Refund
	err := r.db.WithContext(ctx).
		Preload("Payment").
		First(&refund, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

// GetRefundByRefundNo 根据退款单号获取
func (r *paymentRepository) GetRefundByRefundNo(ctx context.Context, refundNo string) (*model.Refund, error) {
	var refund model.Refund
	err := r.db.WithContext(ctx).
		Preload("Payment").
		First(&refund, "refund_no = ?", refundNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

// ListRefunds 分页查询退款列表
func (r *paymentRepository) ListRefunds(ctx context.Context, query *RefundQuery) ([]*model.Refund, int64, error) {
	var refunds []*model.Refund
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Refund{})

	if query.PaymentID != nil {
		db = db.Where("payment_id = ?", *query.PaymentID)
	}
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Payment").
		Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&refunds).Error

	return refunds, total, err
}

// UpdateRefund 更新退款记录
func (r *paymentRepository) UpdateRefund(ctx context.Context, refund *model.Refund) error {
	return r.db.WithContext(ctx).Save(refund).Error
}

// UpdateRefundStatus 更新退款状态
func (r *paymentRepository) UpdateRefundStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == model.RefundStatusSuccess {
		updates["refunded_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.Refund{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// CreateCallback 创建回调记录
func (r *paymentRepository) CreateCallback(ctx context.Context, callback *model.PaymentCallback) error {
	return r.db.WithContext(ctx).Create(callback).Error
}

// UpdateCallback 更新回调记录
func (r *paymentRepository) UpdateCallback(ctx context.Context, callback *model.PaymentCallback) error {
	return r.db.WithContext(ctx).Save(callback).Error
}

// GetCallbacksByPaymentID 获取支付的所有回调记录
func (r *paymentRepository) GetCallbacksByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*model.PaymentCallback, error) {
	var callbacks []*model.PaymentCallback
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&callbacks).Error
	return callbacks, err
}

// CreateRoute 创建路由规则
func (r *paymentRepository) CreateRoute(ctx context.Context, route *model.PaymentRoute) error {
	return r.db.WithContext(ctx).Create(route).Error
}

// GetRouteByID 根据ID获取路由规则
func (r *paymentRepository) GetRouteByID(ctx context.Context, id uuid.UUID) (*model.PaymentRoute, error) {
	var route model.PaymentRoute
	err := r.db.WithContext(ctx).First(&route, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

// ListActiveRoutes 获取所有启用的路由规则
func (r *paymentRepository) ListActiveRoutes(ctx context.Context) ([]*model.PaymentRoute, error) {
	var routes []*model.PaymentRoute
	err := r.db.WithContext(ctx).
		Where("is_enabled = true").
		Order("priority DESC, created_at ASC").
		Find(&routes).Error
	return routes, err
}

// UpdateRoute 更新路由规则
func (r *paymentRepository) UpdateRoute(ctx context.Context, route *model.PaymentRoute) error {
	return r.db.WithContext(ctx).Save(route).Error
}

// DeleteRoute 删除路由规则
func (r *paymentRepository) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.PaymentRoute{}, "id = ?", id).Error
}

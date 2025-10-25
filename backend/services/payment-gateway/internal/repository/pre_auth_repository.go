package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/payment-gateway/internal/model"
)

// PreAuthRepository 预授权仓库接口
type PreAuthRepository interface {
	// 创建和查询
	Create(ctx context.Context, preAuth *model.PreAuthPayment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.PreAuthPayment, error)
	GetByPreAuthNo(ctx context.Context, merchantID uuid.UUID, preAuthNo string) (*model.PreAuthPayment, error)
	GetByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) (*model.PreAuthPayment, error)
	GetByChannelTradeNo(ctx context.Context, channelTradeNo string) (*model.PreAuthPayment, error)

	// 状态更新
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateToAuthorized(ctx context.Context, id uuid.UUID, channelTradeNo string, authorizedAt time.Time) error
	UpdateToCaptured(ctx context.Context, id uuid.UUID, capturedAmount int64, paymentNo string, capturedAt time.Time) error
	UpdateToCancelled(ctx context.Context, id uuid.UUID, cancelledAt time.Time, reason string) error
	UpdateToExpired(ctx context.Context, id uuid.UUID) error

	// 批量操作
	GetExpiredPreAuths(ctx context.Context, limit int) ([]*model.PreAuthPayment, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, status string, offset, limit int) ([]*model.PreAuthPayment, error)
}

// preAuthRepository 仓库实现
type preAuthRepository struct {
	db *gorm.DB
}

// NewPreAuthRepository 创建预授权仓库
func NewPreAuthRepository(db *gorm.DB) PreAuthRepository {
	return &preAuthRepository{db: db}
}

// Create 创建预授权记录
func (r *preAuthRepository) Create(ctx context.Context, preAuth *model.PreAuthPayment) error {
	return r.db.WithContext(ctx).Create(preAuth).Error
}

// GetByID 根据ID获取预授权记录
func (r *preAuthRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.PreAuthPayment, error) {
	var preAuth model.PreAuthPayment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&preAuth).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &preAuth, err
}

// GetByPreAuthNo 根据预授权单号获取
func (r *preAuthRepository) GetByPreAuthNo(ctx context.Context, merchantID uuid.UUID, preAuthNo string) (*model.PreAuthPayment, error) {
	var preAuth model.PreAuthPayment
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND pre_auth_no = ?", merchantID, preAuthNo).
		First(&preAuth).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &preAuth, err
}

// GetByOrderNo 根据订单号获取
func (r *preAuthRepository) GetByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) (*model.PreAuthPayment, error) {
	var preAuth model.PreAuthPayment
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND order_no = ?", merchantID, orderNo).
		First(&preAuth).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &preAuth, err
}

// GetByChannelTradeNo 根据渠道交易号获取
func (r *preAuthRepository) GetByChannelTradeNo(ctx context.Context, channelTradeNo string) (*model.PreAuthPayment, error) {
	var preAuth model.PreAuthPayment
	err := r.db.WithContext(ctx).
		Where("channel_trade_no = ?", channelTradeNo).
		First(&preAuth).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &preAuth, err
}

// UpdateStatus 更新状态
func (r *preAuthRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.PreAuthPayment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// UpdateToAuthorized 更新为已授权状态
func (r *preAuthRepository) UpdateToAuthorized(ctx context.Context, id uuid.UUID, channelTradeNo string, authorizedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.PreAuthPayment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           model.PreAuthStatusAuthorized,
			"channel_trade_no": channelTradeNo,
			"authorized_at":    authorizedAt,
			"updated_at":       time.Now(),
		}).Error
}

// UpdateToCaptured 更新为已确认状态
func (r *preAuthRepository) UpdateToCaptured(ctx context.Context, id uuid.UUID, capturedAmount int64, paymentNo string, capturedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.PreAuthPayment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          model.PreAuthStatusCaptured,
			"captured_amount": gorm.Expr("captured_amount + ?", capturedAmount),
			"payment_no":      paymentNo,
			"captured_at":     capturedAt,
			"updated_at":      time.Now(),
		}).Error
}

// UpdateToCancelled 更新为已取消状态
func (r *preAuthRepository) UpdateToCancelled(ctx context.Context, id uuid.UUID, cancelledAt time.Time, reason string) error {
	return r.db.WithContext(ctx).
		Model(&model.PreAuthPayment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.PreAuthStatusCancelled,
			"cancelled_at":  cancelledAt,
			"error_message": reason,
			"updated_at":    time.Now(),
		}).Error
}

// UpdateToExpired 更新为已过期状态
func (r *preAuthRepository) UpdateToExpired(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.PreAuthPayment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.PreAuthStatusExpired,
			"updated_at": time.Now(),
		}).Error
}

// GetExpiredPreAuths 获取过期的预授权记录
func (r *preAuthRepository) GetExpiredPreAuths(ctx context.Context, limit int) ([]*model.PreAuthPayment, error) {
	var preAuths []*model.PreAuthPayment
	err := r.db.WithContext(ctx).
		Where("status IN (?) AND expires_at < ?",
			[]string{model.PreAuthStatusPending, model.PreAuthStatusAuthorized},
			time.Now()).
		Order("expires_at ASC").
		Limit(limit).
		Find(&preAuths).Error

	return preAuths, err
}

// ListByMerchant 获取商户的预授权列表
func (r *preAuthRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, status string, offset, limit int) ([]*model.PreAuthPayment, error) {
	var preAuths []*model.PreAuthPayment
	query := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&preAuths).Error

	return preAuths, err
}

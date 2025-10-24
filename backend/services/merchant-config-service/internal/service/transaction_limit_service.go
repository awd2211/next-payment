package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-config-service/internal/model"
	"payment-platform/merchant-config-service/internal/repository"
)

// TransactionLimitService 交易限额服务接口
type TransactionLimitService interface {
	CreateLimit(ctx context.Context, input *CreateLimitInput) (*model.MerchantTransactionLimit, error)
	GetLimit(ctx context.Context, id uuid.UUID) (*model.MerchantTransactionLimit, error)
	ListMerchantLimits(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error)
	CheckLimit(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string, amount int64) error
	UpdateLimit(ctx context.Context, id uuid.UUID, input *UpdateLimitInput) (*model.MerchantTransactionLimit, error)
	DeleteLimit(ctx context.Context, id uuid.UUID) error
}

type transactionLimitService struct {
	repo repository.TransactionLimitRepository
}

// NewTransactionLimitService 创建交易限额服务实例
func NewTransactionLimitService(repo repository.TransactionLimitRepository) TransactionLimitService {
	return &transactionLimitService{repo: repo}
}

// CreateLimitInput 创建限额输入
type CreateLimitInput struct {
	MerchantID    uuid.UUID  `json:"merchant_id"`
	LimitType     string     `json:"limit_type"`
	PaymentMethod string     `json:"payment_method"`
	Channel       string     `json:"channel"`
	Currency      string     `json:"currency"`
	MinAmount     int64      `json:"min_amount"`
	MaxAmount     int64      `json:"max_amount"`
	MaxCount      int        `json:"max_count"`
	EffectiveDate time.Time  `json:"effective_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
}

// UpdateLimitInput 更新限额输入
type UpdateLimitInput struct {
	MinAmount  *int64     `json:"min_amount"`
	MaxAmount  *int64     `json:"max_amount"`
	MaxCount   *int       `json:"max_count"`
	ExpiryDate *time.Time `json:"expiry_date"`
	Status     *string    `json:"status"`
}

// CreateLimit 创建交易限额
func (s *transactionLimitService) CreateLimit(ctx context.Context, input *CreateLimitInput) (*model.MerchantTransactionLimit, error) {
	if err := s.validateLimit(input); err != nil {
		return nil, err
	}

	limit := &model.MerchantTransactionLimit{
		MerchantID:    input.MerchantID,
		LimitType:     input.LimitType,
		PaymentMethod: input.PaymentMethod,
		Channel:       input.Channel,
		Currency:      input.Currency,
		MinAmount:     input.MinAmount,
		MaxAmount:     input.MaxAmount,
		MaxCount:      input.MaxCount,
		Status:        model.LimitStatusActive,
		EffectiveDate: input.EffectiveDate,
		ExpiryDate:    input.ExpiryDate,
	}

	if err := s.repo.Create(ctx, limit); err != nil {
		return nil, err
	}

	return limit, nil
}

// GetLimit 获取交易限额
func (s *transactionLimitService) GetLimit(ctx context.Context, id uuid.UUID) (*model.MerchantTransactionLimit, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMerchantLimits 列出商户的所有限额
func (s *transactionLimitService) ListMerchantLimits(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error) {
	return s.repo.GetByMerchantID(ctx, merchantID)
}

// CheckLimit 检查是否超过限额
func (s *transactionLimitService) CheckLimit(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string, amount int64) error {
	// 检查单笔限额
	limits, err := s.repo.GetEffectiveLimits(ctx, merchantID, model.LimitTypeSingle, channel, paymentMethod, time.Now())
	if err != nil {
		return err
	}

	for _, limit := range limits {
		// 检查最小金额
		if limit.MinAmount > 0 && amount < limit.MinAmount {
			return errors.New("amount is below minimum limit")
		}
		// 检查最大金额
		if limit.MaxAmount > 0 && amount > limit.MaxAmount {
			return errors.New("amount exceeds maximum limit")
		}
	}

	// TODO: 实现日累计、月累计限额检查（需要查询 payment 表统计当日/当月交易额）

	return nil
}

// UpdateLimit 更新交易限额
func (s *transactionLimitService) UpdateLimit(ctx context.Context, id uuid.UUID, input *UpdateLimitInput) (*model.MerchantTransactionLimit, error) {
	limit, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.MinAmount != nil {
		limit.MinAmount = *input.MinAmount
	}
	if input.MaxAmount != nil {
		limit.MaxAmount = *input.MaxAmount
	}
	if input.MaxCount != nil {
		limit.MaxCount = *input.MaxCount
	}
	if input.ExpiryDate != nil {
		limit.ExpiryDate = input.ExpiryDate
	}
	if input.Status != nil {
		limit.Status = *input.Status
	}

	if err := s.repo.Update(ctx, limit); err != nil {
		return nil, err
	}

	return limit, nil
}

// DeleteLimit 删除交易限额
func (s *transactionLimitService) DeleteLimit(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// validateLimit 验证限额配置
func (s *transactionLimitService) validateLimit(input *CreateLimitInput) error {
	if input.MerchantID == uuid.Nil {
		return errors.New("merchant_id is required")
	}
	if input.LimitType == "" {
		return errors.New("limit_type is required")
	}
	if input.Currency == "" {
		return errors.New("currency is required")
	}
	if input.MaxAmount > 0 && input.MinAmount > 0 && input.MinAmount >= input.MaxAmount {
		return errors.New("min_amount must be less than max_amount")
	}

	return nil
}

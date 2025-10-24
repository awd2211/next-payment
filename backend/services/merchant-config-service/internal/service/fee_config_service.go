package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-config-service/internal/model"
	"payment-platform/merchant-config-service/internal/repository"
)

// FeeConfigService 费率配置服务接口
type FeeConfigService interface {
	CreateFeeConfig(ctx context.Context, input *CreateFeeConfigInput) (*model.MerchantFeeConfig, error)
	GetFeeConfig(ctx context.Context, id uuid.UUID) (*model.MerchantFeeConfig, error)
	ListMerchantFeeConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error)
	CalculateFee(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string, amount int64) (int64, error)
	UpdateFeeConfig(ctx context.Context, id uuid.UUID, input *UpdateFeeConfigInput) (*model.MerchantFeeConfig, error)
	DeleteFeeConfig(ctx context.Context, id uuid.UUID) error
	ApproveFeeConfig(ctx context.Context, id, approverID uuid.UUID) error
}

type feeConfigService struct {
	repo repository.FeeConfigRepository
}

// NewFeeConfigService 创建费率配置服务实例
func NewFeeConfigService(repo repository.FeeConfigRepository) FeeConfigService {
	return &feeConfigService{repo: repo}
}

// CreateFeeConfigInput 创建费率配置输入
type CreateFeeConfigInput struct {
	MerchantID    uuid.UUID  `json:"merchant_id"`
	Channel       string     `json:"channel"`
	PaymentMethod string     `json:"payment_method"`
	FeeType       string     `json:"fee_type"`
	FeePercentage float64    `json:"fee_percentage"`
	FeeFixed      int64      `json:"fee_fixed"`
	MinFee        int64      `json:"min_fee"`
	MaxFee        int64      `json:"max_fee"`
	Currency      string     `json:"currency"`
	TieredRules   string     `json:"tiered_rules"`
	EffectiveDate time.Time  `json:"effective_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	Priority      int        `json:"priority"`
	CreatedBy     *uuid.UUID `json:"created_by"`
}

// UpdateFeeConfigInput 更新费率配置输入
type UpdateFeeConfigInput struct {
	FeePercentage *float64   `json:"fee_percentage"`
	FeeFixed      *int64     `json:"fee_fixed"`
	MinFee        *int64     `json:"min_fee"`
	MaxFee        *int64     `json:"max_fee"`
	TieredRules   *string    `json:"tiered_rules"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	Priority      *int       `json:"priority"`
	Status        *string    `json:"status"`
}

// CreateFeeConfig 创建费率配置
func (s *feeConfigService) CreateFeeConfig(ctx context.Context, input *CreateFeeConfigInput) (*model.MerchantFeeConfig, error) {
	// 验证
	if err := s.validateFeeConfig(input); err != nil {
		return nil, err
	}

	config := &model.MerchantFeeConfig{
		MerchantID:    input.MerchantID,
		Channel:       input.Channel,
		PaymentMethod: input.PaymentMethod,
		FeeType:       input.FeeType,
		FeePercentage: input.FeePercentage,
		FeeFixed:      input.FeeFixed,
		MinFee:        input.MinFee,
		MaxFee:        input.MaxFee,
		Currency:      input.Currency,
		TieredRules:   input.TieredRules,
		EffectiveDate: input.EffectiveDate,
		ExpiryDate:    input.ExpiryDate,
		Priority:      input.Priority,
		Status:        model.FeeStatusActive,
		CreatedBy:     input.CreatedBy,
	}

	if err := s.repo.Create(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetFeeConfig 获取费率配置
func (s *feeConfigService) GetFeeConfig(ctx context.Context, id uuid.UUID) (*model.MerchantFeeConfig, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMerchantFeeConfigs 列出商户的所有费率配置
func (s *feeConfigService) ListMerchantFeeConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantFeeConfig, error) {
	return s.repo.GetByMerchantID(ctx, merchantID)
}

// CalculateFee 计算手续费
func (s *feeConfigService) CalculateFee(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod string, amount int64) (int64, error) {
	// 获取生效的费率配置
	config, err := s.repo.GetEffectiveConfig(ctx, merchantID, channel, paymentMethod, time.Now())
	if err != nil {
		return 0, errors.New("no effective fee config found")
	}

	var fee int64

	switch config.FeeType {
	case model.FeeTypePercentage:
		// 百分比费率
		fee = int64(float64(amount) * config.FeePercentage)
	case model.FeeTypeFixed:
		// 固定费用
		fee = config.FeeFixed
	case model.FeeTypeTiered:
		// 阶梯费率（TODO: 实现阶梯规则解析）
		fee = int64(float64(amount) * config.FeePercentage)
	default:
		return 0, errors.New("invalid fee type")
	}

	// 应用最小/最大限制
	if config.MinFee > 0 && fee < config.MinFee {
		fee = config.MinFee
	}
	if config.MaxFee > 0 && fee > config.MaxFee {
		fee = config.MaxFee
	}

	return fee, nil
}

// UpdateFeeConfig 更新费率配置
func (s *feeConfigService) UpdateFeeConfig(ctx context.Context, id uuid.UUID, input *UpdateFeeConfigInput) (*model.MerchantFeeConfig, error) {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if input.FeePercentage != nil {
		config.FeePercentage = *input.FeePercentage
	}
	if input.FeeFixed != nil {
		config.FeeFixed = *input.FeeFixed
	}
	if input.MinFee != nil {
		config.MinFee = *input.MinFee
	}
	if input.MaxFee != nil {
		config.MaxFee = *input.MaxFee
	}
	if input.TieredRules != nil {
		config.TieredRules = *input.TieredRules
	}
	if input.ExpiryDate != nil {
		config.ExpiryDate = input.ExpiryDate
	}
	if input.Priority != nil {
		config.Priority = *input.Priority
	}
	if input.Status != nil {
		config.Status = *input.Status
	}

	if err := s.repo.Update(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// DeleteFeeConfig 删除费率配置
func (s *feeConfigService) DeleteFeeConfig(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ApproveFeeConfig 审批费率配置
func (s *feeConfigService) ApproveFeeConfig(ctx context.Context, id, approverID uuid.UUID) error {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	config.ApprovedBy = &approverID
	config.ApprovedAt = &now
	config.Status = model.FeeStatusActive

	return s.repo.Update(ctx, config)
}

// validateFeeConfig 验证费率配置
func (s *feeConfigService) validateFeeConfig(input *CreateFeeConfigInput) error {
	if input.MerchantID == uuid.Nil {
		return errors.New("merchant_id is required")
	}
	if input.Channel == "" {
		return errors.New("channel is required")
	}
	if input.FeeType == "" {
		return errors.New("fee_type is required")
	}
	if input.Currency == "" {
		return errors.New("currency is required")
	}

	// 验证费率类型
	switch input.FeeType {
	case model.FeeTypePercentage:
		if input.FeePercentage <= 0 {
			return errors.New("fee_percentage must be greater than 0")
		}
	case model.FeeTypeFixed:
		if input.FeeFixed <= 0 {
			return errors.New("fee_fixed must be greater than 0")
		}
	case model.FeeTypeTiered:
		if input.TieredRules == "" {
			return errors.New("tiered_rules is required for tiered fee type")
		}
	default:
		return errors.New("invalid fee_type")
	}

	return nil
}

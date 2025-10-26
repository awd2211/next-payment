package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"payment-platform/merchant-policy-service/internal/model"
	"payment-platform/merchant-policy-service/internal/repository"
)

// TierService 商户等级服务接口
type TierService interface {
	CreateTier(ctx context.Context, input *CreateTierInput) (*model.MerchantTier, error)
	UpdateTier(ctx context.Context, id uuid.UUID, input *UpdateTierInput) (*model.MerchantTier, error)
	DeleteTier(ctx context.Context, id uuid.UUID) error
	GetTierByID(ctx context.Context, id uuid.UUID) (*model.MerchantTier, error)
	GetTierByCode(ctx context.Context, tierCode string) (*model.MerchantTier, error)
	ListTiers(ctx context.Context, isActive *bool, page, pageSize int) (*TierListOutput, error)
	GetAllActiveTiers(ctx context.Context) ([]*model.MerchantTier, error)
}

type tierService struct {
	tierRepo repository.TierRepository
}

// NewTierService 创建等级服务实例
func NewTierService(tierRepo repository.TierRepository) TierService {
	return &tierService{
		tierRepo: tierRepo,
	}
}

// CreateTierInput 创建等级输入
type CreateTierInput struct {
	TierCode              string     `json:"tier_code" binding:"required"`
	TierName              string     `json:"tier_name" binding:"required"`
	TierLevel             int        `json:"tier_level" binding:"required,min=1,max=5"`
	Description           string     `json:"description"`
	DefaultFeePolicyID    *uuid.UUID `json:"default_fee_policy_id"`
	DefaultLimitPolicyID  *uuid.UUID `json:"default_limit_policy_id"`
	UpgradeRequirements   string     `json:"upgrade_requirements"`
	AllowedChannels       string     `json:"allowed_channels"`
	AllowedCurrencies     string     `json:"allowed_currencies"`
	MaxAPICallsPerMin     int        `json:"max_api_calls_per_min"`
	IsActive              bool       `json:"is_active"`
}

// UpdateTierInput 更新等级输入
type UpdateTierInput struct {
	TierName             *string    `json:"tier_name"`
	Description          *string    `json:"description"`
	DefaultFeePolicyID   *uuid.UUID `json:"default_fee_policy_id"`
	DefaultLimitPolicyID *uuid.UUID `json:"default_limit_policy_id"`
	UpgradeRequirements  *string    `json:"upgrade_requirements"`
	AllowedChannels      *string    `json:"allowed_channels"`
	AllowedCurrencies    *string    `json:"allowed_currencies"`
	MaxAPICallsPerMin    *int       `json:"max_api_calls_per_min"`
	IsActive             *bool      `json:"is_active"`
}

// TierListOutput 等级列表输出
type TierListOutput struct {
	Tiers      []*model.MerchantTier `json:"tiers"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

func (s *tierService) CreateTier(ctx context.Context, input *CreateTierInput) (*model.MerchantTier, error) {
	// 检查等级代码是否已存在
	existing, err := s.tierRepo.GetByCode(ctx, input.TierCode)
	if err != nil {
		return nil, fmt.Errorf("检查等级代码失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("等级代码 %s 已存在", input.TierCode)
	}

	tier := &model.MerchantTier{
		TierCode:             input.TierCode,
		TierName:             input.TierName,
		TierLevel:            input.TierLevel,
		Description:          input.Description,
		DefaultFeePolicyID:   input.DefaultFeePolicyID,
		DefaultLimitPolicyID: input.DefaultLimitPolicyID,
		UpgradeRequirements:  input.UpgradeRequirements,
		AllowedChannels:      input.AllowedChannels,
		AllowedCurrencies:    input.AllowedCurrencies,
		MaxAPICallsPerMin:    input.MaxAPICallsPerMin,
		IsActive:             input.IsActive,
	}

	if err := s.tierRepo.Create(ctx, tier); err != nil {
		return nil, fmt.Errorf("创建等级失败: %w", err)
	}

	return tier, nil
}

func (s *tierService) UpdateTier(ctx context.Context, id uuid.UUID, input *UpdateTierInput) (*model.MerchantTier, error) {
	tier, err := s.tierRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询等级失败: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("等级不存在")
	}

	// 更新字段
	if input.TierName != nil {
		tier.TierName = *input.TierName
	}
	if input.Description != nil {
		tier.Description = *input.Description
	}
	if input.DefaultFeePolicyID != nil {
		tier.DefaultFeePolicyID = input.DefaultFeePolicyID
	}
	if input.DefaultLimitPolicyID != nil {
		tier.DefaultLimitPolicyID = input.DefaultLimitPolicyID
	}
	if input.UpgradeRequirements != nil {
		tier.UpgradeRequirements = *input.UpgradeRequirements
	}
	if input.AllowedChannels != nil {
		tier.AllowedChannels = *input.AllowedChannels
	}
	if input.AllowedCurrencies != nil {
		tier.AllowedCurrencies = *input.AllowedCurrencies
	}
	if input.MaxAPICallsPerMin != nil {
		tier.MaxAPICallsPerMin = *input.MaxAPICallsPerMin
	}
	if input.IsActive != nil {
		tier.IsActive = *input.IsActive
	}

	if err := s.tierRepo.Update(ctx, tier); err != nil {
		return nil, fmt.Errorf("更新等级失败: %w", err)
	}

	return tier, nil
}

func (s *tierService) DeleteTier(ctx context.Context, id uuid.UUID) error {
	tier, err := s.tierRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询等级失败: %w", err)
	}
	if tier == nil {
		return fmt.Errorf("等级不存在")
	}

	// TODO: 检查是否有商户使用此等级（需要调用 PolicyBindingRepository）

	if err := s.tierRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除等级失败: %w", err)
	}

	return nil
}

func (s *tierService) GetTierByID(ctx context.Context, id uuid.UUID) (*model.MerchantTier, error) {
	tier, err := s.tierRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询等级失败: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("等级不存在")
	}
	return tier, nil
}

func (s *tierService) GetTierByCode(ctx context.Context, tierCode string) (*model.MerchantTier, error) {
	tier, err := s.tierRepo.GetByCode(ctx, tierCode)
	if err != nil {
		return nil, fmt.Errorf("查询等级失败: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("等级代码 %s 不存在", tierCode)
	}
	return tier, nil
}

func (s *tierService) ListTiers(ctx context.Context, isActive *bool, page, pageSize int) (*TierListOutput, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	tiers, total, err := s.tierRepo.List(ctx, isActive, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询等级列表失败: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &TierListOutput{
		Tiers:      tiers,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *tierService) GetAllActiveTiers(ctx context.Context) ([]*model.MerchantTier, error) {
	tiers, err := s.tierRepo.GetAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询活跃等级列表失败: %w", err)
	}
	return tiers, nil
}

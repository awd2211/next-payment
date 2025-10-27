package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-policy-service/internal/model"
	"payment-platform/merchant-policy-service/internal/repository"
)

// PolicyBindingService 策略绑定服务接口
type PolicyBindingService interface {
	// 绑定商户到等级（新商户注册时调用）
	BindMerchantToTier(ctx context.Context, input *BindMerchantInput) (*model.MerchantPolicyBinding, error)

	// 升级/降级商户等级
	ChangeMerchantTier(ctx context.Context, input *ChangeTierInput) (*model.MerchantPolicyBinding, error)

	// 设置商户自定义策略（覆盖等级默认策略）
	SetCustomPolicy(ctx context.Context, input *SetCustomPolicyInput) (*model.MerchantPolicyBinding, error)

	// 获取商户当前策略绑定
	GetMerchantBinding(ctx context.Context, merchantID uuid.UUID) (*MerchantBindingDetail, error)

	// 删除商户策略绑定
	UnbindMerchant(ctx context.Context, merchantID uuid.UUID) error
}

type policyBindingService struct {
	bindingRepo     repository.PolicyBindingRepository
	tierRepo        repository.TierRepository
	feePolicyRepo   repository.FeePolicyRepository
	limitPolicyRepo repository.LimitPolicyRepository
}

// NewPolicyBindingService 创建策略绑定服务实例
func NewPolicyBindingService(
	bindingRepo repository.PolicyBindingRepository,
	tierRepo repository.TierRepository,
	feePolicyRepo repository.FeePolicyRepository,
	limitPolicyRepo repository.LimitPolicyRepository,
) PolicyBindingService {
	return &policyBindingService{
		bindingRepo:     bindingRepo,
		tierRepo:        tierRepo,
		feePolicyRepo:   feePolicyRepo,
		limitPolicyRepo: limitPolicyRepo,
	}
}

// BindMerchantInput 绑定商户输入
type BindMerchantInput struct {
	MerchantID  uuid.UUID  `json:"merchant_id" binding:"required"`
	TierID      uuid.UUID  `json:"tier_id" binding:"required"`
	ChangedBy   *uuid.UUID `json:"changed_by"`
	ChangeReason string    `json:"change_reason"`
}

// ChangeTierInput 变更等级输入
type ChangeTierInput struct {
	MerchantID   uuid.UUID  `json:"merchant_id" binding:"required"`
	NewTierID    uuid.UUID  `json:"new_tier_id" binding:"required"`
	ChangedBy    *uuid.UUID `json:"changed_by"`
	ChangeReason string     `json:"change_reason" binding:"required"`
}

// SetCustomPolicyInput 设置自定义策略输入
type SetCustomPolicyInput struct {
	MerchantID          uuid.UUID  `json:"merchant_id" binding:"required"`
	CustomFeePolicyID   *uuid.UUID `json:"custom_fee_policy_id"`
	CustomLimitPolicyID *uuid.UUID `json:"custom_limit_policy_id"`
	ChangedBy           *uuid.UUID `json:"changed_by"`
	ChangeReason        string     `json:"change_reason"`
}

// MerchantBindingDetail 商户策略绑定详情
type MerchantBindingDetail struct {
	Binding            *model.MerchantPolicyBinding `json:"binding"`
	Tier               *model.MerchantTier          `json:"tier"`
	CustomFeePolicy    *model.MerchantFeePolicy     `json:"custom_fee_policy,omitempty"`
	CustomLimitPolicy  *model.MerchantLimitPolicy   `json:"custom_limit_policy,omitempty"`
}

func (s *policyBindingService) BindMerchantToTier(ctx context.Context, input *BindMerchantInput) (*model.MerchantPolicyBinding, error) {
	// 检查等级是否存在
	tier, err := s.tierRepo.GetByID(ctx, input.TierID)
	if err != nil {
		return nil, fmt.Errorf("查询等级失败: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("等级不存在")
	}
	if !tier.IsActive {
		return nil, fmt.Errorf("等级已停用")
	}

	// 检查商户是否已绑定
	existing, err := s.bindingRepo.GetByMerchantID(ctx, input.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("查询商户绑定失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("商户已绑定等级，请使用变更等级接口")
	}

	// 创建绑定
	now := time.Now()
	binding := &model.MerchantPolicyBinding{
		MerchantID:    input.MerchantID,
		TierID:        input.TierID,
		EffectiveDate: now,
		ChangedBy:     input.ChangedBy,
		ChangeReason:  input.ChangeReason,
	}

	if err := s.bindingRepo.Create(ctx, binding); err != nil {
		return nil, fmt.Errorf("创建商户策略绑定失败: %w", err)
	}

	return binding, nil
}

func (s *policyBindingService) ChangeMerchantTier(ctx context.Context, input *ChangeTierInput) (*model.MerchantPolicyBinding, error) {
	// 检查新等级是否存在
	newTier, err := s.tierRepo.GetByID(ctx, input.NewTierID)
	if err != nil {
		return nil, fmt.Errorf("查询新等级失败: %w", err)
	}
	if newTier == nil {
		return nil, fmt.Errorf("新等级不存在")
	}
	if !newTier.IsActive {
		return nil, fmt.Errorf("新等级已停用")
	}

	// 查询当前绑定
	binding, err := s.bindingRepo.GetByMerchantID(ctx, input.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("查询商户绑定失败: %w", err)
	}
	if binding == nil {
		return nil, fmt.Errorf("商户未绑定等级")
	}

	// 检查是否变更到同一等级
	if binding.TierID == input.NewTierID {
		return nil, fmt.Errorf("商户已在该等级，无需变更")
	}

	// 更新绑定
	binding.TierID = input.NewTierID
	binding.ChangedBy = input.ChangedBy
	binding.ChangeReason = input.ChangeReason

	if err := s.bindingRepo.Update(ctx, binding); err != nil {
		return nil, fmt.Errorf("更新商户策略绑定失败: %w", err)
	}

	return binding, nil
}

func (s *policyBindingService) SetCustomPolicy(ctx context.Context, input *SetCustomPolicyInput) (*model.MerchantPolicyBinding, error) {
	// 查询当前绑定
	binding, err := s.bindingRepo.GetByMerchantID(ctx, input.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("查询商户绑定失败: %w", err)
	}
	if binding == nil {
		return nil, fmt.Errorf("商户未绑定等级")
	}

	// ✅ FIXED: 验证自定义策略ID是否存在
	if input.CustomFeePolicyID != nil {
		feePolicy, err := s.feePolicyRepo.GetByID(ctx, *input.CustomFeePolicyID)
		if err != nil {
			return nil, fmt.Errorf("查询自定义费率策略失败: %w", err)
		}
		if feePolicy == nil {
			return nil, fmt.Errorf("自定义费率策略不存在 (ID: %s)", input.CustomFeePolicyID.String())
		}
		if feePolicy.Status != model.FeeStatusActive {
			return nil, fmt.Errorf("自定义费率策略未启用 (状态: %s)", feePolicy.Status)
		}
	}

	if input.CustomLimitPolicyID != nil {
		limitPolicy, err := s.limitPolicyRepo.GetByID(ctx, *input.CustomLimitPolicyID)
		if err != nil {
			return nil, fmt.Errorf("查询自定义限额策略失败: %w", err)
		}
		if limitPolicy == nil {
			return nil, fmt.Errorf("自定义限额策略不存在 (ID: %s)", input.CustomLimitPolicyID.String())
		}
		if limitPolicy.Status != model.LimitStatusActive {
			return nil, fmt.Errorf("自定义限额策略未启用 (状态: %s)", limitPolicy.Status)
		}
	}

	// 更新自定义策略
	binding.CustomFeePolicyID = input.CustomFeePolicyID
	binding.CustomLimitPolicyID = input.CustomLimitPolicyID
	binding.ChangedBy = input.ChangedBy
	binding.ChangeReason = input.ChangeReason

	if err := s.bindingRepo.Update(ctx, binding); err != nil {
		return nil, fmt.Errorf("更新商户自定义策略失败: %w", err)
	}

	return binding, nil
}

func (s *policyBindingService) GetMerchantBinding(ctx context.Context, merchantID uuid.UUID) (*MerchantBindingDetail, error) {
	// 查询绑定和等级
	binding, tier, err := s.bindingRepo.GetByMerchantIDWithTier(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("查询商户策略绑定失败: %w", err)
	}
	if binding == nil {
		return nil, fmt.Errorf("商户未绑定等级")
	}

	detail := &MerchantBindingDetail{
		Binding: binding,
		Tier:    tier,
	}

	// ✅ FIXED: 查询自定义费率策略和限额策略（如果有）
	if binding.CustomFeePolicyID != nil {
		feePolicy, err := s.feePolicyRepo.GetByID(ctx, *binding.CustomFeePolicyID)
		if err != nil {
			return nil, fmt.Errorf("查询自定义费率策略失败: %w", err)
		}
		detail.CustomFeePolicy = feePolicy
	}

	if binding.CustomLimitPolicyID != nil {
		limitPolicy, err := s.limitPolicyRepo.GetByID(ctx, *binding.CustomLimitPolicyID)
		if err != nil {
			return nil, fmt.Errorf("查询自定义限额策略失败: %w", err)
		}
		detail.CustomLimitPolicy = limitPolicy
	}

	return detail, nil
}

func (s *policyBindingService) UnbindMerchant(ctx context.Context, merchantID uuid.UUID) error {
	binding, err := s.bindingRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("查询商户绑定失败: %w", err)
	}
	if binding == nil {
		return fmt.Errorf("商户未绑定等级")
	}

	if err := s.bindingRepo.Delete(ctx, merchantID); err != nil {
		return fmt.Errorf("删除商户策略绑定失败: %w", err)
	}

	return nil
}

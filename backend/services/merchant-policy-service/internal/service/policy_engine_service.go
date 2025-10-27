package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-policy-service/internal/model"
	"payment-platform/merchant-policy-service/internal/repository"
)

// PolicyEngineService 策略引擎服务接口
type PolicyEngineService interface {
	// 获取商户的有效费率策略（优先级: 商户自定义 > 等级默认）
	GetEffectiveFeePolicy(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string) (*model.MerchantFeePolicy, error)

	// 获取商户的有效限额策略（优先级: 商户自定义 > 等级默认）
	GetEffectiveLimitPolicy(ctx context.Context, merchantID uuid.UUID, channel, currency string) (*model.MerchantLimitPolicy, error)

	// 计算交易费用
	CalculateFee(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string, amount int64) (*FeeCalculationResult, error)

	// 检查交易是否超限
	CheckLimit(ctx context.Context, merchantID uuid.UUID, channel, currency string, amount int64, dailyUsed, monthlyUsed int64) (*LimitCheckResult, error)
}

type policyEngineService struct {
	feePolicyRepo    repository.FeePolicyRepository
	limitPolicyRepo  repository.LimitPolicyRepository
	bindingRepo      repository.PolicyBindingRepository
	tierRepo         repository.TierRepository
}

// NewPolicyEngineService 创建策略引擎服务实例
func NewPolicyEngineService(
	feePolicyRepo repository.FeePolicyRepository,
	limitPolicyRepo repository.LimitPolicyRepository,
	bindingRepo repository.PolicyBindingRepository,
	tierRepo repository.TierRepository,
) PolicyEngineService {
	return &policyEngineService{
		feePolicyRepo:   feePolicyRepo,
		limitPolicyRepo: limitPolicyRepo,
		bindingRepo:     bindingRepo,
		tierRepo:        tierRepo,
	}
}

// TieredRule 阶梯费率规则
type TieredRule struct {
	MinAmount  int64    `json:"min_amount"`            // 最小金额（分）
	MaxAmount  *int64   `json:"max_amount,omitempty"`  // 最大金额（分，null表示无上限）
	Percentage float64  `json:"percentage"`            // 费率百分比
}

// FeeCalculationResult 费用计算结果
type FeeCalculationResult struct {
	FeeAmount        int64                      `json:"fee_amount"`         // 费用金额（分）
	FeePercentage    float64                    `json:"fee_percentage"`     // 费率百分比
	FeeFixed         int64                      `json:"fee_fixed"`          // 固定费用（分）
	FeeType          string                     `json:"fee_type"`           // 费率类型
	AppliedPolicy    *model.MerchantFeePolicy   `json:"applied_policy"`     // 应用的策略
	CalculationNotes string                     `json:"calculation_notes"`  // 计算说明
}

// LimitCheckResult 限额检查结果
type LimitCheckResult struct {
	IsAllowed        bool                        `json:"is_allowed"`         // 是否允许交易
	RejectionReason  string                      `json:"rejection_reason"`   // 拒绝原因
	SingleTransMax   int64                       `json:"single_trans_max"`   // 单笔最大金额
	DailyLimit       int64                       `json:"daily_limit"`        // 日限额
	DailyRemaining   int64                       `json:"daily_remaining"`    // 日剩余额度
	MonthlyLimit     int64                       `json:"monthly_limit"`      // 月限额
	MonthlyRemaining int64                       `json:"monthly_remaining"`  // 月剩余额度
	AppliedPolicy    *model.MerchantLimitPolicy  `json:"applied_policy"`     // 应用的策略
}

func (s *policyEngineService) GetEffectiveFeePolicy(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string) (*model.MerchantFeePolicy, error) {
	now := time.Now()

	// 1. 先查商户自定义策略
	customPolicies, err := s.feePolicyRepo.GetEffectivePolicies(ctx, &merchantID, nil, channel, paymentMethod, currency, now)
	if err != nil {
		return nil, fmt.Errorf("查询商户自定义费率策略失败: %w", err)
	}
	if len(customPolicies) > 0 {
		return customPolicies[0], nil // 返回优先级最高的
	}

	// 2. 查询商户绑定的等级
	binding, tier, err := s.bindingRepo.GetByMerchantIDWithTier(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("查询商户策略绑定失败: %w", err)
	}
	if binding == nil || tier == nil {
		return nil, fmt.Errorf("商户未绑定等级或等级不存在")
	}

	// 3. 查询等级默认策略
	tierPolicies, err := s.feePolicyRepo.GetEffectivePolicies(ctx, nil, &tier.ID, channel, paymentMethod, currency, now)
	if err != nil {
		return nil, fmt.Errorf("查询等级默认费率策略失败: %w", err)
	}
	if len(tierPolicies) > 0 {
		return tierPolicies[0], nil
	}

	return nil, fmt.Errorf("未找到适用的费率策略")
}

func (s *policyEngineService) GetEffectiveLimitPolicy(ctx context.Context, merchantID uuid.UUID, channel, currency string) (*model.MerchantLimitPolicy, error) {
	now := time.Now()

	// 1. 先查商户自定义策略
	customPolicies, err := s.limitPolicyRepo.GetEffectivePolicies(ctx, &merchantID, nil, channel, currency, now)
	if err != nil {
		return nil, fmt.Errorf("查询商户自定义限额策略失败: %w", err)
	}
	if len(customPolicies) > 0 {
		return customPolicies[0], nil
	}

	// 2. 查询商户绑定的等级
	binding, tier, err := s.bindingRepo.GetByMerchantIDWithTier(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("查询商户策略绑定失败: %w", err)
	}
	if binding == nil || tier == nil {
		return nil, fmt.Errorf("商户未绑定等级或等级不存在")
	}

	// 3. 查询等级默认策略
	tierPolicies, err := s.limitPolicyRepo.GetEffectivePolicies(ctx, nil, &tier.ID, channel, currency, now)
	if err != nil {
		return nil, fmt.Errorf("查询等级默认限额策略失败: %w", err)
	}
	if len(tierPolicies) > 0 {
		return tierPolicies[0], nil
	}

	return nil, fmt.Errorf("未找到适用的限额策略")
}

func (s *policyEngineService) CalculateFee(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string, amount int64) (*FeeCalculationResult, error) {
	policy, err := s.GetEffectiveFeePolicy(ctx, merchantID, channel, paymentMethod, currency)
	if err != nil {
		return nil, err
	}

	result := &FeeCalculationResult{
		FeeType:       policy.FeeType,
		FeePercentage: policy.FeePercentage,
		FeeFixed:      policy.FeeFixed,
		AppliedPolicy: policy,
	}

	// 根据费率类型计算费用
	switch policy.FeeType {
	case model.FeeTypePercentage:
		result.FeeAmount = int64(float64(amount) * policy.FeePercentage)
		result.CalculationNotes = fmt.Sprintf("百分比费率: %.2f%%", policy.FeePercentage*100)

	case model.FeeTypeFixed:
		result.FeeAmount = policy.FeeFixed
		result.CalculationNotes = fmt.Sprintf("固定费用: %d 分", policy.FeeFixed)

	case model.FeeTypeTiered:
		// ✅ FIXED: 解析 TieredRules JSON 实现阶梯费率
		if policy.TieredRules == "" {
			return nil, fmt.Errorf("阶梯费率策略缺少TieredRules配置")
		}

		var rules []TieredRule
		if err := json.Unmarshal([]byte(policy.TieredRules), &rules); err != nil {
			return nil, fmt.Errorf("解析阶梯费率规则失败: %w", err)
		}

		// 根据交易金额找到匹配的阶梯
		matchedRule := findMatchingTieredRule(rules, amount)
		if matchedRule == nil {
			return nil, fmt.Errorf("未找到匹配的阶梯费率规则（金额: %d）", amount)
		}

		result.FeeAmount = int64(float64(amount) * matchedRule.Percentage)
		result.FeePercentage = matchedRule.Percentage
		if matchedRule.MaxAmount != nil {
			result.CalculationNotes = fmt.Sprintf("阶梯费率: 金额区间 [%d, %d], 费率 %.2f%%",
				matchedRule.MinAmount, *matchedRule.MaxAmount, matchedRule.Percentage*100)
		} else {
			result.CalculationNotes = fmt.Sprintf("阶梯费率: 金额区间 [%d, +∞), 费率 %.2f%%",
				matchedRule.MinAmount, matchedRule.Percentage*100)
		}

	default:
		return nil, fmt.Errorf("不支持的费率类型: %s", policy.FeeType)
	}

	// 应用最小费用
	if result.FeeAmount < policy.MinFee {
		result.FeeAmount = policy.MinFee
		result.CalculationNotes += fmt.Sprintf(", 应用最小费用: %d 分", policy.MinFee)
	}

	// 应用最大费用
	if policy.MaxFee != nil && result.FeeAmount > *policy.MaxFee {
		result.FeeAmount = *policy.MaxFee
		result.CalculationNotes += fmt.Sprintf(", 应用最大费用: %d 分", *policy.MaxFee)
	}

	return result, nil
}

func (s *policyEngineService) CheckLimit(ctx context.Context, merchantID uuid.UUID, channel, currency string, amount int64, dailyUsed, monthlyUsed int64) (*LimitCheckResult, error) {
	policy, err := s.GetEffectiveLimitPolicy(ctx, merchantID, channel, currency)
	if err != nil {
		return nil, err
	}

	result := &LimitCheckResult{
		IsAllowed:        true,
		SingleTransMax:   policy.SingleTransMax,
		DailyLimit:       policy.DailyLimit,
		MonthlyLimit:     policy.MonthlyLimit,
		DailyRemaining:   policy.DailyLimit - dailyUsed,
		MonthlyRemaining: policy.MonthlyLimit - monthlyUsed,
		AppliedPolicy:    policy,
	}

	// 检查单笔最小金额
	if amount < policy.SingleTransMin {
		result.IsAllowed = false
		result.RejectionReason = fmt.Sprintf("交易金额 %d 低于最小限额 %d", amount, policy.SingleTransMin)
		return result, nil
	}

	// 检查单笔最大金额
	if amount > policy.SingleTransMax {
		result.IsAllowed = false
		result.RejectionReason = fmt.Sprintf("交易金额 %d 超过单笔最大限额 %d", amount, policy.SingleTransMax)
		return result, nil
	}

	// 检查日限额
	if dailyUsed+amount > policy.DailyLimit {
		result.IsAllowed = false
		result.RejectionReason = fmt.Sprintf("超过日限额: 已使用 %d + 当前交易 %d > 限额 %d", dailyUsed, amount, policy.DailyLimit)
		return result, nil
	}

	// 检查月限额
	if monthlyUsed+amount > policy.MonthlyLimit {
		result.IsAllowed = false
		result.RejectionReason = fmt.Sprintf("超过月限额: 已使用 %d + 当前交易 %d > 限额 %d", monthlyUsed, amount, policy.MonthlyLimit)
		return result, nil
	}

	return result, nil
}

// findMatchingTieredRule 根据金额找到匹配的阶梯费率规则
func findMatchingTieredRule(rules []TieredRule, amount int64) *TieredRule {
	for i := range rules {
		rule := &rules[i]
		// 金额必须 >= MinAmount
		if amount < rule.MinAmount {
			continue
		}
		// 如果MaxAmount为null（表示无上限）或金额 <= MaxAmount，则匹配
		if rule.MaxAmount == nil || amount <= *rule.MaxAmount {
			return rule
		}
	}
	return nil
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"payment-platform/risk-service/internal/model"
	"payment-platform/risk-service/internal/repository"
)

// RiskService 风控服务接口
type RiskService interface {
	// 风控规则管理
	CreateRule(ctx context.Context, input *CreateRuleInput) (*model.RiskRule, error)
	GetRule(ctx context.Context, id uuid.UUID) (*model.RiskRule, error)
	ListRules(ctx context.Context, query *repository.RuleQuery) ([]*model.RiskRule, int64, error)
	UpdateRule(ctx context.Context, id uuid.UUID, input *UpdateRuleInput) (*model.RiskRule, error)
	DeleteRule(ctx context.Context, id uuid.UUID) error
	EnableRule(ctx context.Context, id uuid.UUID) error
	DisableRule(ctx context.Context, id uuid.UUID) error

	// 风控检查
	CheckPayment(ctx context.Context, input *PaymentCheckInput) (*model.RiskCheck, error)
	GetCheck(ctx context.Context, id uuid.UUID) (*model.RiskCheck, error)
	ListChecks(ctx context.Context, query *repository.CheckQuery) ([]*model.RiskCheck, int64, error)

	// 黑名单管理
	AddBlacklist(ctx context.Context, input *AddBlacklistInput) (*model.Blacklist, error)
	RemoveBlacklist(ctx context.Context, id uuid.UUID) error
	CheckBlacklist(ctx context.Context, entityType, entityValue string) (bool, *model.Blacklist, error)
	ListBlacklist(ctx context.Context, query *repository.BlacklistQuery) ([]*model.Blacklist, int64, error)
}

type riskService struct {
	riskRepo repository.RiskRepository
}

// NewRiskService 创建风控服务实例
func NewRiskService(riskRepo repository.RiskRepository) RiskService {
	return &riskService{
		riskRepo: riskRepo,
	}
}

// Input structures

type CreateRuleInput struct {
	RuleName    string                 `json:"rule_name" binding:"required"`
	RuleType    string                 `json:"rule_type" binding:"required"`
	Conditions  map[string]interface{} `json:"conditions" binding:"required"`
	Actions     map[string]interface{} `json:"actions" binding:"required"`
	Priority    int                    `json:"priority"`
	Description string                 `json:"description"`
}

type UpdateRuleInput struct {
	RuleName    string                 `json:"rule_name"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     map[string]interface{} `json:"actions"`
	Priority    int                    `json:"priority"`
	Description string                 `json:"description"`
}

type PaymentCheckInput struct {
	MerchantID    uuid.UUID              `json:"merchant_id" binding:"required"`
	RelatedID     uuid.UUID              `json:"related_id" binding:"required"`
	RelatedType   string                 `json:"related_type" binding:"required"`
	Amount        int64                  `json:"amount" binding:"required"`
	Currency      string                 `json:"currency" binding:"required"`
	PayerIP       string                 `json:"payer_ip"`
	PayerEmail    string                 `json:"payer_email"`
	PayerPhone    string                 `json:"payer_phone"`
	DeviceID      string                 `json:"device_id"`
	PaymentMethod string                 `json:"payment_method"`
	Extra         map[string]interface{} `json:"extra"`
}

type AddBlacklistInput struct {
	EntityType  string `json:"entity_type" binding:"required"`
	EntityValue string `json:"entity_value" binding:"required"`
	Reason      string `json:"reason" binding:"required"`
	AddedBy     string `json:"added_by"`
	ExpireAt    *time.Time `json:"expire_at"`
}

// Rule Management

func (s *riskService) CreateRule(ctx context.Context, input *CreateRuleInput) (*model.RiskRule, error) {
	rule := &model.RiskRule{
		RuleName:    input.RuleName,
		RuleType:    input.RuleType,
		Conditions:  input.Conditions,
		Actions:     input.Actions,
		Priority:    input.Priority,
		Status:      model.RuleStatusActive,
		Description: input.Description,
	}

	if err := s.riskRepo.CreateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("创建规则失败: %w", err)
	}

	return rule, nil
}

func (s *riskService) GetRule(ctx context.Context, id uuid.UUID) (*model.RiskRule, error) {
	rule, err := s.riskRepo.GetRuleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取规则失败: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("规则不存在")
	}
	return rule, nil
}

func (s *riskService) ListRules(ctx context.Context, query *repository.RuleQuery) ([]*model.RiskRule, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.riskRepo.ListRules(ctx, query)
}

func (s *riskService) UpdateRule(ctx context.Context, id uuid.UUID, input *UpdateRuleInput) (*model.RiskRule, error) {
	rule, err := s.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.RuleName != "" {
		rule.RuleName = input.RuleName
	}
	if input.Conditions != nil {
		rule.Conditions = input.Conditions
	}
	if input.Actions != nil {
		rule.Actions = input.Actions
	}
	if input.Priority > 0 {
		rule.Priority = input.Priority
	}
	if input.Description != "" {
		rule.Description = input.Description
	}

	if err := s.riskRepo.UpdateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("更新规则失败: %w", err)
	}

	return rule, nil
}

func (s *riskService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return s.riskRepo.DeleteRule(ctx, id)
}

func (s *riskService) EnableRule(ctx context.Context, id uuid.UUID) error {
	rule, err := s.GetRule(ctx, id)
	if err != nil {
		return err
	}
	rule.Status = model.RuleStatusActive
	return s.riskRepo.UpdateRule(ctx, rule)
}

func (s *riskService) DisableRule(ctx context.Context, id uuid.UUID) error {
	rule, err := s.GetRule(ctx, id)
	if err != nil {
		return err
	}
	rule.Status = model.RuleStatusInactive
	return s.riskRepo.UpdateRule(ctx, rule)
}

// Risk Checks

func (s *riskService) CheckPayment(ctx context.Context, input *PaymentCheckInput) (*model.RiskCheck, error) {
	check := &model.RiskCheck{
		MerchantID:  input.MerchantID,
		RelatedID:   input.RelatedID,
		RelatedType: input.RelatedType,
		CheckData: map[string]interface{}{
			"amount":         input.Amount,
			"currency":       input.Currency,
			"payer_ip":       input.PayerIP,
			"payer_email":    input.PayerEmail,
			"payer_phone":    input.PayerPhone,
			"device_id":      input.DeviceID,
			"payment_method": input.PaymentMethod,
		},
		RiskLevel:   model.RiskLevelLow,
		Decision:    model.DecisionPass,
		CheckResult: make(map[string]interface{}),
	}

	// 1. 黑名单检查
	blacklistHit, blacklistReason := s.checkBlacklistRules(ctx, input)
	if blacklistHit {
		check.RiskLevel = model.RiskLevelCritical
		check.Decision = model.DecisionReject
		check.Reason = blacklistReason
		check.CheckResult["blacklist"] = "hit"
		if err := s.riskRepo.CreateCheck(ctx, check); err != nil {
			return nil, fmt.Errorf("创建检查记录失败: %w", err)
		}
		return check, nil
	}
	check.CheckResult["blacklist"] = "pass"

	// 2. 金额风险检查
	amountRisk := s.checkAmountRisk(input.Amount, input.Currency)
	if amountRisk != "" {
		check.RiskLevel = s.upgradeRiskLevel(check.RiskLevel, model.RiskLevelHigh)
		check.Reason = amountRisk
		check.CheckResult["amount_risk"] = "high"
	} else {
		check.CheckResult["amount_risk"] = "normal"
	}

	// 3. 频率检查
	frequencyRisk := s.checkFrequency(ctx, input)
	if frequencyRisk != "" {
		check.RiskLevel = s.upgradeRiskLevel(check.RiskLevel, model.RiskLevelMedium)
		if check.Reason != "" {
			check.Reason += "; " + frequencyRisk
		} else {
			check.Reason = frequencyRisk
		}
		check.CheckResult["frequency_risk"] = "high"
	} else {
		check.CheckResult["frequency_risk"] = "normal"
	}

	// 4. 设备风险检查
	if input.DeviceID != "" {
		deviceRisk := s.checkDeviceRisk(ctx, input.DeviceID)
		if deviceRisk != "" {
			check.RiskLevel = s.upgradeRiskLevel(check.RiskLevel, model.RiskLevelMedium)
			if check.Reason != "" {
				check.Reason += "; " + deviceRisk
			} else {
				check.Reason = deviceRisk
			}
			check.CheckResult["device_risk"] = "suspicious"
		} else {
			check.CheckResult["device_risk"] = "normal"
		}
	}

	// 决策逻辑
	switch check.RiskLevel {
	case model.RiskLevelCritical, model.RiskLevelHigh:
		check.Decision = model.DecisionReview
	case model.RiskLevelMedium:
		check.Decision = model.DecisionReview
	default:
		check.Decision = model.DecisionPass
	}

	if err := s.riskRepo.CreateCheck(ctx, check); err != nil {
		return nil, fmt.Errorf("创建检查记录失败: %w", err)
	}

	return check, nil
}

func (s *riskService) GetCheck(ctx context.Context, id uuid.UUID) (*model.RiskCheck, error) {
	check, err := s.riskRepo.GetCheckByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取检查记录失败: %w", err)
	}
	if check == nil {
		return nil, fmt.Errorf("检查记录不存在")
	}
	return check, nil
}

func (s *riskService) ListChecks(ctx context.Context, query *repository.CheckQuery) ([]*model.RiskCheck, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.riskRepo.ListChecks(ctx, query)
}

// Blacklist Management

func (s *riskService) AddBlacklist(ctx context.Context, input *AddBlacklistInput) (*model.Blacklist, error) {
	// 检查是否已存在
	existing, _ := s.riskRepo.CheckBlacklist(ctx, input.EntityType, input.EntityValue)
	if existing != nil {
		return nil, fmt.Errorf("黑名单记录已存在")
	}

	blacklist := &model.Blacklist{
		EntityType:  input.EntityType,
		EntityValue: input.EntityValue,
		Reason:      input.Reason,
		AddedBy:     input.AddedBy,
		Status:      "active",
		ExpireAt:    input.ExpireAt,
	}

	if err := s.riskRepo.CreateBlacklist(ctx, blacklist); err != nil {
		return nil, fmt.Errorf("添加黑名单失败: %w", err)
	}

	return blacklist, nil
}

func (s *riskService) RemoveBlacklist(ctx context.Context, id uuid.UUID) error {
	blacklist, err := s.riskRepo.GetBlacklistByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取黑名单失败: %w", err)
	}
	if blacklist == nil {
		return fmt.Errorf("黑名单记录不存在")
	}

	blacklist.Status = "removed"
	now := time.Now()
	blacklist.RemovedAt = &now
	return s.riskRepo.DeleteBlacklist(ctx, id)
}

func (s *riskService) CheckBlacklist(ctx context.Context, entityType, entityValue string) (bool, *model.Blacklist, error) {
	blacklist, err := s.riskRepo.CheckBlacklist(ctx, entityType, entityValue)
	if err != nil {
		return false, nil, err
	}
	if blacklist != nil {
		// 检查是否过期
		if blacklist.ExpireAt != nil && blacklist.ExpireAt.Before(time.Now()) {
			return false, nil, nil
		}
		return true, blacklist, nil
	}
	return false, nil, nil
}

func (s *riskService) ListBlacklist(ctx context.Context, query *repository.BlacklistQuery) ([]*model.Blacklist, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.riskRepo.ListBlacklist(ctx, query)
}

// Helper functions

func (s *riskService) checkBlacklistRules(ctx context.Context, input *PaymentCheckInput) (bool, string) {
	checks := []struct {
		entityType  string
		entityValue string
	}{
		{"ip", input.PayerIP},
		{"email", input.PayerEmail},
		{"phone", input.PayerPhone},
		{"device", input.DeviceID},
	}

	for _, check := range checks {
		if check.entityValue == "" {
			continue
		}
		hit, blacklist, _ := s.CheckBlacklist(ctx, check.entityType, check.entityValue)
		if hit {
			return true, fmt.Sprintf("命中黑名单: %s (%s)", check.entityType, blacklist.Reason)
		}
	}

	return false, ""
}

func (s *riskService) checkAmountRisk(amount int64, currency string) string {
	// 大额交易检查 (示例阈值)
	threshold := int64(1000000) // 10000 元
	if amount > threshold {
		return fmt.Sprintf("大额交易: %.2f %s", float64(amount)/100, currency)
	}
	return ""
}

func (s *riskService) checkFrequency(ctx context.Context, input *PaymentCheckInput) string {
	// TODO: 实现频率检查逻辑
	// 可以检查同一商户、同一IP、同一设备在短时间内的交易次数
	return ""
}

func (s *riskService) checkDeviceRisk(ctx context.Context, deviceID string) string {
	// TODO: 实现设备风险检查
	// 可以检查设备是否关联多个账户、是否有异常行为等
	return ""
}

func (s *riskService) upgradeRiskLevel(current, new string) string {
	levels := map[string]int{
		model.RiskLevelLow:      1,
		model.RiskLevelMedium:   2,
		model.RiskLevelHigh:     3,
		model.RiskLevelCritical: 4,
	}

	if levels[new] > levels[current] {
		return new
	}
	return current
}

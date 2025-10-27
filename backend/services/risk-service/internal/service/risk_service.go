package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"payment-platform/risk-service/internal/client"
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

	// 支付反馈（用于风控模型训练）
	ReportPaymentResult(ctx context.Context, input *PaymentFeedbackInput) error
}

type riskService struct {
	riskRepo    repository.RiskRepository
	redisClient *redis.Client
	geoipClient *client.IPAPIClient
}

// NewRiskService 创建风控服务实例
func NewRiskService(riskRepo repository.RiskRepository, redisClient *redis.Client, geoipClient *client.IPAPIClient) RiskService {
	return &riskService{
		riskRepo:    riskRepo,
		redisClient: redisClient,
		geoipClient: geoipClient,
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

type PaymentFeedbackInput struct {
	PaymentNo  string `json:"payment_no" binding:"required"`
	Success    bool   `json:"success"`
	Fraudulent bool   `json:"fraudulent"`
	Notes      string `json:"notes"`
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
	// 【缓存优化】1. 先查 Redis 缓存
	cacheKey := fmt.Sprintf("risk_rule:%s", id.String())

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var rule model.RiskRule
			if err := json.Unmarshal([]byte(cached), &rule); err == nil {
				logger.Info("风控规则缓存命中", zap.String("rule_id", id.String()))
				return &rule, nil
			}
		}
	}

	// 【缓存优化】2. 缓存未命中，查询数据库
	rule, err := s.riskRepo.GetRuleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取规则失败: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("规则不存在")
	}

	// 【缓存优化】3. 写入缓存 (5分钟TTL - 风控规则变更频率较低)
	if s.redisClient != nil {
		data, err := json.Marshal(rule)
		if err == nil {
			if err := s.redisClient.Set(ctx, cacheKey, data, 5*time.Minute).Err(); err != nil {
				logger.Warn("写入风控规则缓存失败", zap.String("rule_id", id.String()), zap.Error(err))
			}
		}
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

	// 【缓存失效】更新成功后删除缓存
	s.invalidateRuleCache(ctx, id)

	return rule, nil
}

func (s *riskService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	if err := s.riskRepo.DeleteRule(ctx, id); err != nil {
		return err
	}

	// 【缓存失效】删除成功后清除缓存
	s.invalidateRuleCache(ctx, id)

	return nil
}

func (s *riskService) EnableRule(ctx context.Context, id uuid.UUID) error {
	rule, err := s.GetRule(ctx, id)
	if err != nil {
		return err
	}
	rule.Status = model.RuleStatusActive
	if err := s.riskRepo.UpdateRule(ctx, rule); err != nil {
		return err
	}

	// 【缓存失效】启用成功后删除缓存
	s.invalidateRuleCache(ctx, id)

	return nil
}

func (s *riskService) DisableRule(ctx context.Context, id uuid.UUID) error {
	rule, err := s.GetRule(ctx, id)
	if err != nil {
		return err
	}
	rule.Status = model.RuleStatusInactive
	if err := s.riskRepo.UpdateRule(ctx, rule); err != nil {
		return err
	}

	// 【缓存失效】禁用成功后删除缓存
	s.invalidateRuleCache(ctx, id)

	return nil
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
		RiskScore:   0, // 初始分数为0
		RiskLevel:   model.RiskLevelLow,
		Decision:    model.DecisionPass,
		CheckResult: make(map[string]interface{}),
	}

	// 风险评分累加器
	riskScore := 0

	// 1. 黑名单检查 (+100分，直接critical)
	blacklistHit, blacklistReason := s.checkBlacklistRules(ctx, input)
	if blacklistHit {
		riskScore += 100
		check.RiskScore = riskScore
		check.RiskLevel = model.RiskLevelCritical
		check.Decision = model.DecisionReject
		check.Reason = blacklistReason
		check.CheckResult["blacklist"] = "hit"
		check.CheckResult["blacklist_score"] = 100
		if err := s.riskRepo.CreateCheck(ctx, check); err != nil {
			return nil, fmt.Errorf("创建检查记录失败: %w", err)
		}
		return check, nil
	}
	check.CheckResult["blacklist"] = "pass"

	// 2. 金额风险检查 (+20分)
	amountRisk := s.checkAmountRisk(input.Amount, input.Currency)
	if amountRisk != "" {
		riskScore += 20
		check.Reason = amountRisk
		check.CheckResult["amount_risk"] = "high"
		check.CheckResult["amount_score"] = 20
	} else {
		check.CheckResult["amount_risk"] = "normal"
	}

	// 3. 频率检查 (+15分)
	frequencyRisk := s.checkFrequency(ctx, input)
	if frequencyRisk != "" {
		riskScore += 15
		if check.Reason != "" {
			check.Reason += "; " + frequencyRisk
		} else {
			check.Reason = frequencyRisk
		}
		check.CheckResult["frequency_risk"] = "high"
		check.CheckResult["frequency_score"] = 15
	} else {
		check.CheckResult["frequency_risk"] = "normal"
	}

	// 4. 设备风险检查 (+10分)
	if input.DeviceID != "" {
		deviceRisk := s.checkDeviceRisk(ctx, input.DeviceID)
		if deviceRisk != "" {
			riskScore += 10
			if check.Reason != "" {
				check.Reason += "; " + deviceRisk
			} else {
				check.Reason = deviceRisk
			}
			check.CheckResult["device_risk"] = "suspicious"
			check.CheckResult["device_score"] = 10
		} else {
			check.CheckResult["device_risk"] = "normal"
		}
		// 记录设备活动（用于后续风控分析）
		go s.recordDeviceActivity(context.Background(), input)
	}

	// 5. IP地理位置检查 (+10分)
	geoRisk, geoInfo := s.checkIPGeolocationWithInfo(ctx, input.PayerIP)
	if geoRisk != "" {
		riskScore += 10
		if check.Reason != "" {
			check.Reason += "; " + geoRisk
		} else {
			check.Reason = geoRisk
		}
		check.CheckResult["geo_risk"] = "suspicious"
		check.CheckResult["geo_score"] = 10
	} else {
		check.CheckResult["geo_risk"] = "normal"
	}
	// 存储 GeoIP 信息供下游使用（如智能路由）
	if geoInfo != nil {
		check.CheckResult["geo_country_code"] = geoInfo.CountryCode
		check.CheckResult["geo_country"] = geoInfo.Country
		check.CheckResult["geo_city"] = geoInfo.City
	}

	// 6. 执行动态规则引擎（可能增加额外分数）
	ruleDecision, _, ruleResults := s.executeRules(ctx, input)
	if ruleDecision != "" {
		check.CheckResult["rules"] = ruleResults
		// 规则引擎的决策优先级最高
		if ruleDecision == "block" || ruleDecision == model.DecisionReject {
			riskScore += 50 // 规则拒绝 +50分
			check.CheckResult["rule_score"] = 50
			check.Decision = model.DecisionReject
			if reason, ok := ruleResults["reason"].(string); ok {
				if check.Reason != "" {
					check.Reason += "; " + reason
				} else {
					check.Reason = reason
				}
			}
		} else if ruleDecision == "review" || ruleDecision == model.DecisionReview {
			riskScore += 25 // 规则建议审核 +25分
			check.CheckResult["rule_score"] = 25
		}
	}

	// 7. 根据总分计算最终风险等级
	check.RiskScore = riskScore
	check.RiskLevel = s.calculateRiskLevel(riskScore)
	check.CheckResult["total_score"] = riskScore

	// 决策逻辑（如果规则引擎没有做决策）
	if check.Decision == "" {
		switch check.RiskLevel {
		case model.RiskLevelCritical, model.RiskLevelHigh:
			check.Decision = model.DecisionReview
		case model.RiskLevelMedium:
			check.Decision = model.DecisionReview
		default:
			check.Decision = model.DecisionPass
		}
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
	// 频率检查配置
	type freqCheck struct {
		value     string
		key       string
		limit     int
		duration  time.Duration
		riskMsg   string
	}

	checks := []freqCheck{
		// IP频率：同一IP每分钟最多10笔交易
		{
			value:    input.PayerIP,
			key:      fmt.Sprintf("risk:freq:ip:%s", input.PayerIP),
			limit:    10,
			duration: time.Minute,
			riskMsg:  "IP交易频率过高",
		},
		// 商户频率：同一商户每分钟最多100笔交易
		{
			value:    input.MerchantID.String(),
			key:      fmt.Sprintf("risk:freq:merchant:%s", input.MerchantID.String()),
			limit:    100,
			duration: time.Minute,
			riskMsg:  "商户交易频率异常",
		},
		// 设备频率：同一设备每小时最多30笔交易
		{
			value:    input.DeviceID,
			key:      fmt.Sprintf("risk:freq:device:%s", input.DeviceID),
			limit:    30,
			duration: time.Hour,
			riskMsg:  "设备交易频率异常",
		},
		// 邮箱频率：同一邮箱每10分钟最多5笔交易
		{
			value:    input.PayerEmail,
			key:      fmt.Sprintf("risk:freq:email:%s", input.PayerEmail),
			limit:    5,
			duration: 10 * time.Minute,
			riskMsg:  "邮箱交易频率过高",
		},
	}

	for _, check := range checks {
		// 跳过空值
		if check.value == "" {
			continue
		}

		// 获取当前计数
		count, err := s.redisClient.Incr(ctx, check.key).Result()
		if err != nil {
			// Redis错误不阻断流程，记录日志
			continue
		}

		// 第一次访问，设置过期时间
		if count == 1 {
			s.redisClient.Expire(ctx, check.key, check.duration)
		}

		// 超过限制
		if count > int64(check.limit) {
			return fmt.Sprintf("%s (%d次/%v)", check.riskMsg, count, check.duration)
		}
	}

	return ""
}

func (s *riskService) checkDeviceRisk(ctx context.Context, deviceID string) string {
	if deviceID == "" {
		return ""
	}

	// 1. 检查设备关联的邮箱数量（使用 Set 存储）
	emailSetKey := fmt.Sprintf("risk:device:emails:%s", deviceID)
	emailCount, err := s.redisClient.SCard(ctx, emailSetKey).Result()
	if err == nil && emailCount > 5 {
		return fmt.Sprintf("设备关联过多账户 (%d个邮箱)", emailCount)
	}

	// 2. 检查设备关联的商户数量
	merchantSetKey := fmt.Sprintf("risk:device:merchants:%s", deviceID)
	merchantCount, err := s.redisClient.SCard(ctx, merchantSetKey).Result()
	if err == nil && merchantCount > 10 {
		return fmt.Sprintf("设备异常：关联过多商户 (%d个)", merchantCount)
	}

	// 3. 检查设备的IP变化频率（24小时内IP数量）
	ipSetKey := fmt.Sprintf("risk:device:ips:%s", deviceID)
	ipCount, err := s.redisClient.SCard(ctx, ipSetKey).Result()
	if err == nil && ipCount > 20 {
		return fmt.Sprintf("设备异常：IP频繁变化 (%d个IP/24h)", ipCount)
	}

	return ""
}

// RecordDeviceActivity 记录设备活动（在风控检查时调用）
func (s *riskService) recordDeviceActivity(ctx context.Context, input *PaymentCheckInput) {
	if input.DeviceID == "" {
		return
	}

	// 记录设备关联的邮箱
	if input.PayerEmail != "" {
		emailSetKey := fmt.Sprintf("risk:device:emails:%s", input.DeviceID)
		s.redisClient.SAdd(ctx, emailSetKey, input.PayerEmail)
		s.redisClient.Expire(ctx, emailSetKey, 30*24*time.Hour) // 30天过期
	}

	// 记录设备关联的商户
	merchantSetKey := fmt.Sprintf("risk:device:merchants:%s", input.DeviceID)
	s.redisClient.SAdd(ctx, merchantSetKey, input.MerchantID.String())
	s.redisClient.Expire(ctx, merchantSetKey, 30*24*time.Hour)

	// 记录设备的IP
	if input.PayerIP != "" {
		ipSetKey := fmt.Sprintf("risk:device:ips:%s", input.DeviceID)
		s.redisClient.SAdd(ctx, ipSetKey, input.PayerIP)
		s.redisClient.Expire(ctx, ipSetKey, 24*time.Hour) // 24小时过期
	}
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

// calculateRiskLevel 根据评分计算风险等级
func (s *riskService) calculateRiskLevel(score int) string {
	switch {
	case score >= 81:
		return model.RiskLevelCritical // 81+分：极高风险
	case score >= 51:
		return model.RiskLevelHigh // 51-80分：高风险
	case score >= 21:
		return model.RiskLevelMedium // 21-50分：中等风险
	default:
		return model.RiskLevelLow // 0-20分：低风险
	}
}

// executeRules 执行动态规则引擎
func (s *riskService) executeRules(ctx context.Context, input *PaymentCheckInput) (string, string, map[string]interface{}) {
	// 获取所有启用的规则，按优先级排序
	query := &repository.RuleQuery{
		Status:   model.RuleStatusActive,
		Page:     1,
		PageSize: 100,
	}
	rules, _, err := s.riskRepo.ListRules(ctx, query)
	if err != nil {
		return "", "", nil
	}

	// 按优先级排序（优先级高的先执行）
	sortedRules := s.sortRulesByPriority(rules)

	ruleResults := make(map[string]interface{})
	var matchedRule *model.RiskRule

	// 遍历规则进行匹配
	for _, rule := range sortedRules {
		if s.matchRule(rule, input) {
			matchedRule = rule
			ruleResults[rule.RuleName] = "matched"
			break // 匹配到第一个符合的规则就停止
		}
	}

	if matchedRule == nil {
		return "", "", ruleResults
	}

	// 执行规则动作
	decision := ""
	riskLevel := ""

	if matchedRule.Actions != nil {
		if d, ok := matchedRule.Actions["decision"].(string); ok {
			decision = d
		}
		if rl, ok := matchedRule.Actions["risk_level"].(string); ok {
			riskLevel = rl
		}
		ruleResults["action"] = matchedRule.Actions
	}

	ruleResults["matched_rule"] = matchedRule.RuleName
	ruleResults["reason"] = fmt.Sprintf("命中规则: %s", matchedRule.RuleName)

	return decision, riskLevel, ruleResults
}

// matchRule 判断规则是否匹配
func (s *riskService) matchRule(rule *model.RiskRule, input *PaymentCheckInput) bool {
	if rule.Conditions == nil {
		return false
	}

	// 金额范围检查
	if minAmount, ok := rule.Conditions["amount_min"].(float64); ok {
		if input.Amount < int64(minAmount) {
			return false
		}
	}
	if maxAmount, ok := rule.Conditions["amount_max"].(float64); ok {
		if input.Amount > int64(maxAmount) {
			return false
		}
	}

	// 货币检查
	if currency, ok := rule.Conditions["currency"].(string); ok {
		if input.Currency != currency {
			return false
		}
	}

	// 支付方式检查
	if payMethod, ok := rule.Conditions["payment_method"].(string); ok {
		if input.PaymentMethod != payMethod {
			return false
		}
	}

	// IP前缀检查（简单实现）
	if ipPrefix, ok := rule.Conditions["ip_prefix"].(string); ok {
		if !s.ipMatchesPrefix(input.PayerIP, ipPrefix) {
			return false
		}
	}

	// 邮箱域名检查
	if emailDomain, ok := rule.Conditions["email_domain"].(string); ok {
		if !s.emailMatchesDomain(input.PayerEmail, emailDomain) {
			return false
		}
	}

	return true
}

// sortRulesByPriority 按优先级排序规则
func (s *riskService) sortRulesByPriority(rules []*model.RiskRule) []*model.RiskRule {
	// 简单的冒泡排序（实际项目中应使用 sort.Slice）
	sorted := make([]*model.RiskRule, len(rules))
	copy(sorted, rules)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority < sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// 辅助函数：IP前缀匹配
func (s *riskService) ipMatchesPrefix(ip, prefix string) bool {
	if ip == "" || prefix == "" {
		return false
	}
	return len(ip) >= len(prefix) && ip[:len(prefix)] == prefix
}

// 辅助函数：邮箱域名匹配
func (s *riskService) emailMatchesDomain(email, domain string) bool {
	if email == "" || domain == "" {
		return false
	}
	parts := []rune(email)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '@' {
			emailDomain := string(parts[i+1:])
			return emailDomain == domain
		}
	}
	return false
}

// checkIPGeolocation 检查IP地理位置风险（向后兼容）
func (s *riskService) checkIPGeolocation(ctx context.Context, ip string) string {
	riskMsg, _ := s.checkIPGeolocationWithInfo(ctx, ip)
	return riskMsg
}

// checkIPGeolocationWithInfo 检查IP地理位置风险并返回GeoIP信息
func (s *riskService) checkIPGeolocationWithInfo(ctx context.Context, ip string) (string, *client.GeoIPInfo) {
	if ip == "" {
		return "", nil
	}

	// 注意：这是一个简化的示例实现
	// 生产环境应该使用专业的GeoIP库，如 MaxMind GeoIP2
	// https://github.com/oschwald/geoip2-golang

	// 高风险IP段示例（仅供演示）
	highRiskPrefixes := []string{
		// Tor 出口节点示例
		"104.200.",
		"185.220.",
		// 其他高风险段
		// 实际应该从数据库或配置文件加载
	}

	for _, prefix := range highRiskPrefixes {
		if s.ipMatchesPrefix(ip, prefix) {
			return fmt.Sprintf("高风险地理位置: IP段 %s", prefix), nil
		}
	}

	// 使用 ipapi.co 进行 GeoIP 查询
	if s.geoipClient != nil {
		geoInfo, err := s.geoipClient.LookupIP(ctx, ip)
		if err == nil && geoInfo != nil {
			// 检查高风险国家
			if client.IsHighRiskCountry(geoInfo.CountryCode) {
				return fmt.Sprintf("高风险国家: %s (%s)", geoInfo.Country, geoInfo.CountryCode), geoInfo
			}

			// 检查IP是否属于已知的高风险段
			if client.IsHighRiskIP(ip) {
				return fmt.Sprintf("高风险IP段: %s (来源: %s, %s)", ip, geoInfo.City, geoInfo.Country), geoInfo
			}

			// 注：ipapi.co 免费版不提供代理/VPN检测
			// 如需此功能，可升级到付费版或集成其他服务

			// 即使没有风险，也返回 GeoIP 信息供下游使用
			return "", geoInfo
		}
		// GeoIP 查询失败不影响整体风控流程，继续后续检查
	}

	return "", nil
}

// isHighRiskCountry 判断是否为高风险国家（示例）
func (s *riskService) isHighRiskCountry(countryCode string) bool {
	// 这应该从配置或数据库加载
	highRiskCountries := map[string]bool{
		// 示例，实际应根据业务需求配置
		// "XX": true,
	}
	return highRiskCountries[countryCode]
}

// 【缓存优化】缓存失效辅助方法
func (s *riskService) invalidateRuleCache(ctx context.Context, id uuid.UUID) {
	if s.redisClient == nil {
		return
	}

	cacheKey := fmt.Sprintf("risk_rule:%s", id.String())
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		logger.Warn("删除风控规则缓存失败", zap.String("cache_key", cacheKey), zap.Error(err))
	} else {
		logger.Info("风控规则缓存已失效", zap.String("cache_key", cacheKey))
	}
}

// ReportPaymentResult 上报支付结果（用于风控模型训练）
func (s *riskService) ReportPaymentResult(ctx context.Context, input *PaymentFeedbackInput) error {
	// 1. 查找关联的风控检查记录
	check, err := s.riskRepo.GetCheckByPaymentNo(ctx, input.PaymentNo)
	if err != nil {
		// 未找到风控检查记录，仍然记录反馈（可能是跳过风控的支付）
		logger.Warn("未找到风控检查记录",
			zap.String("payment_no", input.PaymentNo),
			zap.Error(err))
	}

	// 2. 创建支付反馈记录
	feedback := &model.PaymentFeedback{
		PaymentNo:  input.PaymentNo,
		Success:    input.Success,
		Fraudulent: input.Fraudulent,
		Notes:      input.Notes,
	}

	// 3. 如果找到风控检查记录，关联数据
	if check != nil {
		feedback.CheckID = check.ID
		feedback.RiskScore = check.RiskScore
		feedback.Decision = check.Decision

		// 根据实际结果计算真实风险等级
		if input.Fraudulent {
			feedback.ActualRisk = model.RiskLevelCritical
		} else if !input.Success {
			feedback.ActualRisk = model.RiskLevelHigh
		} else {
			feedback.ActualRisk = model.RiskLevelLow
		}
	}

	// 4. 保存反馈记录到数据库
	if err := s.riskRepo.CreatePaymentFeedback(ctx, feedback); err != nil {
		return fmt.Errorf("保存支付反馈失败: %w", err)
	}

	// 5. 记录日志用于后续分析
	logger.Info("支付反馈已记录",
		zap.String("payment_no", input.PaymentNo),
		zap.Bool("success", input.Success),
		zap.Bool("fraudulent", input.Fraudulent),
		zap.String("actual_risk", feedback.ActualRisk),
		zap.Int("risk_score", feedback.RiskScore))

	// 6. 如果是欺诈交易，自动添加到黑名单（可选逻辑）
	if input.Fraudulent && check != nil {
		go s.autoBlacklistFraud(context.Background(), check)
	}

	return nil
}

// autoBlacklistFraud 自动将欺诈交易的相关信息添加到黑名单
func (s *riskService) autoBlacklistFraud(ctx context.Context, check *model.RiskCheck) {
	// 从检查数据中提取可疑信息
	checkData := check.CheckData

	// 添加邮箱到黑名单
	if email, ok := checkData["payer_email"].(string); ok && email != "" {
		s.AddBlacklist(ctx, &AddBlacklistInput{
			EntityType:  "email",
			EntityValue: email,
			Reason:      "欺诈交易自动拉黑",
			AddedBy:     "system",
		})
		logger.Info("欺诈邮箱已添加到黑名单", zap.String("email", email))
	}

	// 添加IP到黑名单
	if ip, ok := checkData["payer_ip"].(string); ok && ip != "" {
		s.AddBlacklist(ctx, &AddBlacklistInput{
			EntityType:  "ip",
			EntityValue: ip,
			Reason:      "欺诈交易自动拉黑",
			AddedBy:     "system",
		})
		logger.Info("欺诈IP已添加到黑名单", zap.String("ip", ip))
	}

	// 添加设备ID到黑名单
	if deviceID, ok := checkData["device_id"].(string); ok && deviceID != "" {
		s.AddBlacklist(ctx, &AddBlacklistInput{
			EntityType:  "device",
			EntityValue: deviceID,
			Reason:      "欺诈交易自动拉黑",
			AddedBy:     "system",
		})
		logger.Info("欺诈设备已添加到黑名单", zap.String("device_id", deviceID))
	}
}

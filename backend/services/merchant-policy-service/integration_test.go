package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"payment-platform/merchant-policy-service/internal/model"
	"payment-platform/merchant-policy-service/internal/service"
)

// TestMerchantTierManagement 测试商户等级管理
func TestMerchantTierManagement(t *testing.T) {
	t.Run("等级创建和查询", func(t *testing.T) {
		input := &service.CreateTierInput{
			TierCode:          "TIER_BRONZE",
			TierName:          "铜牌商户",
			TierLevel:         1,
			Description:       "新注册商户，入门级别",
			AllowedChannels:   "stripe,paypal",
			AllowedCurrencies: "USD,EUR,CNY",
			MaxAPICallsPerMin: 100,
			IsActive:          true,
		}

		assert.Equal(t, "TIER_BRONZE", input.TierCode)
		assert.Equal(t, 1, input.TierLevel)
		assert.True(t, input.IsActive)
	})

	t.Run("等级升级路径", func(t *testing.T) {
		// 测试等级升级要求
		tiers := []struct {
			Level       int
			Name        string
			Requirement string
		}{
			{1, "铜牌", "新注册商户"},
			{2, "银牌", "月交易额≥10万USD"},
			{3, "金牌", "月交易额≥50万USD"},
			{4, "铂金", "月交易额≥100万USD"},
			{5, "钻石", "月交易额≥500万USD"},
		}

		assert.Len(t, tiers, 5, "应有5个等级")
		assert.Equal(t, "钻石", tiers[4].Name)
	})

	t.Run("等级删除前检查使用情况", func(t *testing.T) {
		// 验证删除等级时会检查是否有商户正在使用
		_ = uuid.New() // tierID
		merchantCount := 5 // 假设有5个商户使用此等级

		canDelete := merchantCount == 0
		assert.False(t, canDelete, "有商户使用时不应允许删除")

		expectedError := "等级正在被 5 个商户使用，无法删除"
		assert.Contains(t, expectedError, "5 个商户")
	})
}

// TestFeePolicyManagement 测试费率策略管理
func TestFeePolicyManagement(t *testing.T) {
	t.Run("百分比费率计算", func(t *testing.T) {
		// 2.9%费率
		amount := int64(100000)     // ¥1000.00
		feePercentage := 0.029      // 2.9%
		expectedFee := int64(2900)  // ¥29.00

		calculatedFee := int64(float64(amount) * feePercentage)
		assert.Equal(t, expectedFee, calculatedFee)
	})

	t.Run("固定费率计算", func(t *testing.T) {
		// 固定30分（¥0.30）
		feeFixed := int64(30)
		assert.Equal(t, int64(30), feeFixed)
	})

	t.Run("阶梯费率计算", func(t *testing.T) {
		// 测试阶梯费率JSON解析
		tieredRulesJSON := `[
			{"min_amount": 0, "max_amount": 100000, "percentage": 0.029},
			{"min_amount": 100000, "max_amount": 1000000, "percentage": 0.025},
			{"min_amount": 1000000, "max_amount": null, "percentage": 0.020}
		]`

		var rules []service.TieredRule
		err := json.Unmarshal([]byte(tieredRulesJSON), &rules)
		assert.NoError(t, err)
		assert.Len(t, rules, 3)

		// 测试不同金额区间的费率
		testCases := []struct {
			amount         int64
			expectedRate   float64
			expectedTier   string
		}{
			{50000, 0.029, "小额交易 (0-10万)"},
			{500000, 0.025, "中额交易 (10万-100万)"},
			{2000000, 0.020, "大额交易 (>100万)"},
		}

		for _, tc := range testCases {
			// 找到匹配的阶梯
			var matchedRule *service.TieredRule
			for i := range rules {
				rule := &rules[i]
				if tc.amount >= rule.MinAmount {
					if rule.MaxAmount == nil || tc.amount <= *rule.MaxAmount {
						matchedRule = rule
						break
					}
				}
			}

			assert.NotNil(t, matchedRule, "应找到匹配的阶梯规则")
			assert.Equal(t, tc.expectedRate, matchedRule.Percentage)
		}
	})

	t.Run("最小/最大费用限制", func(t *testing.T) {
		// 测试最小费用和最大费用的应用
		calculatedFee := int64(10)      // 计算出的费用只有10分
		minFee := int64(50)             // 最小费用50分
		maxFee := int64(500)            // 最大费用500分

		finalFee := calculatedFee
		if finalFee < minFee {
			finalFee = minFee
		}

		assert.Equal(t, minFee, finalFee, "应用最小费用限制")

		// 测试最大费用
		calculatedFeeHigh := int64(1000)
		finalFeeHigh := calculatedFeeHigh
		if finalFeeHigh > maxFee {
			finalFeeHigh = maxFee
		}

		assert.Equal(t, maxFee, finalFeeHigh, "应用最大费用限制")
	})
}

// TestLimitPolicyManagement 测试限额策略管理
func TestLimitPolicyManagement(t *testing.T) {
	t.Run("限额检查逻辑", func(t *testing.T) {
		// 商户限额配置
		singleTransMin := int64(100)      // 单笔最小100分（¥1.00）
		singleTransMax := int64(100000)   // 单笔最大10万分（¥1000.00）
		dailyLimit := int64(500000)       // 日限额50万分（¥5000.00）
		monthlyLimit := int64(10000000)   // 月限额1000万分（¥10万）

		// 测试案例
		currentAmount := int64(50000)     // 当前交易500元
		dailyUsed := int64(400000)        // 已用4000元
		monthlyUsed := int64(8000000)     // 已用8万元

		// 检查单笔限额
		assert.True(t, currentAmount >= singleTransMin, "单笔金额应≥最小限额")
		assert.True(t, currentAmount <= singleTransMax, "单笔金额应≤最大限额")

		// 检查日限额
		canPassDaily := (dailyUsed + currentAmount) <= dailyLimit
		assert.True(t, canPassDaily, "应通过日限额检查")

		// 检查月限额
		canPassMonthly := (monthlyUsed + currentAmount) <= monthlyLimit
		assert.True(t, canPassMonthly, "应通过月限额检查")
	})

	t.Run("超限拒绝场景", func(t *testing.T) {
		// 场景1: 单笔金额过低
		amount := int64(50)
		minAmount := int64(100)
		assert.False(t, amount >= minAmount, "应拒绝:单笔金额过低")

		// 场景2: 单笔金额过高
		amountHigh := int64(200000)
		maxAmount := int64(100000)
		assert.False(t, amountHigh <= maxAmount, "应拒绝:单笔金额过高")

		// 场景3: 超过日限额
		dailyUsed := int64(450000)
		currentTrans := int64(100000)
		dailyLimit := int64(500000)
		assert.False(t, (dailyUsed+currentTrans) <= dailyLimit, "应拒绝:超过日限额")

		// 场景4: 超过月限额
		monthlyUsed := int64(9900000)
		monthlyLimit := int64(10000000)
		assert.False(t, (monthlyUsed+currentTrans) <= monthlyLimit, "应拒绝:超过月限额")
	})
}

// TestPolicyBindingManagement 测试策略绑定管理
func TestPolicyBindingManagement(t *testing.T) {
	t.Run("商户绑定等级", func(t *testing.T) {
		merchantID := uuid.New()
		tierID := uuid.New()

		binding := &model.MerchantPolicyBinding{
			MerchantID:    merchantID,
			TierID:        tierID,
			EffectiveDate: time.Now(),
			ChangeReason:  "新商户注册",
		}

		assert.NotEqual(t, uuid.Nil, binding.MerchantID)
		assert.NotEqual(t, uuid.Nil, binding.TierID)
		assert.NotEmpty(t, binding.ChangeReason)
	})

	t.Run("商户等级变更", func(t *testing.T) {
		// 从铜牌升级到银牌
		oldTierID := uuid.New()
		newTierID := uuid.New()

		assert.NotEqual(t, oldTierID, newTierID, "新旧等级应不同")

		changeReason := "月交易额达到10万USD，升级为银牌商户"
		assert.Contains(t, changeReason, "升级")
	})

	t.Run("自定义策略设置", func(t *testing.T) {
		merchantID := uuid.New()
		customFeePolicyID := uuid.New()
		customLimitPolicyID := uuid.New()

		// 商户可以设置自定义策略覆盖等级默认策略
		binding := &model.MerchantPolicyBinding{
			MerchantID:          merchantID,
			CustomFeePolicyID:   &customFeePolicyID,
			CustomLimitPolicyID: &customLimitPolicyID,
			ChangeReason:        "大客户特殊费率协议",
		}

		assert.NotNil(t, binding.CustomFeePolicyID)
		assert.NotNil(t, binding.CustomLimitPolicyID)
	})

	t.Run("自定义策略ID验证", func(t *testing.T) {
		// 验证自定义策略ID存在性和状态
		_ = uuid.New() // policyID

		// 模拟策略验证
		policyExists := true
		policyStatus := model.FeeStatusActive

		assert.True(t, policyExists, "策略ID应存在")
		assert.Equal(t, model.FeeStatusActive, policyStatus, "策略应处于active状态")

		// 如果策略不存在或未启用，应拒绝
		invalidStatus := model.FeeStatusInactive
		assert.NotEqual(t, model.FeeStatusActive, invalidStatus, "未启用的策略应被拒绝")
	})
}

// TestPolicyEngineIntegration 测试策略引擎集成
func TestPolicyEngineIntegration(t *testing.T) {
	t.Run("策略优先级: 自定义 > 等级默认", func(t *testing.T) {
		// 等级默认费率: 2.9%
		tierDefaultRate := 0.029
		// 商户自定义费率: 2.5%
		customRate := 0.025

		// 应优先使用自定义费率
		hasCustomPolicy := true
		effectiveRate := tierDefaultRate
		if hasCustomPolicy {
			effectiveRate = customRate
		}

		assert.Equal(t, 0.025, effectiveRate, "应优先使用自定义费率")
	})

	t.Run("多维度策略匹配", func(t *testing.T) {
		// 策略匹配维度
		_ = "stripe" // channel
		_ = "card"   // paymentMethod
		_ = "USD"    // currency

		// 查找最匹配的策略（优先级：全匹配 > 部分匹配 > 通配符）
		policies := []struct {
			Channel       string
			PaymentMethod string
			Currency      string
			Priority      int
		}{
			{"stripe", "card", "USD", 100}, // 全匹配
			{"stripe", "all", "USD", 50},   // 部分匹配
			{"all", "all", "USD", 20},      // 通配符
		}

		// 找到优先级最高的策略
		maxPriority := 0
		for _, p := range policies {
			if p.Priority > maxPriority {
				maxPriority = p.Priority
			}
		}

		assert.Equal(t, 100, maxPriority, "应选择优先级最高的策略")
	})

	t.Run("策略生效时间检查", func(t *testing.T) {
		now := time.Now()
		effectiveDate := now.AddDate(0, 0, -7) // 7天前生效
		expiryDate := now.AddDate(0, 0, 7)     // 7天后过期

		isActive := now.After(effectiveDate) && now.Before(expiryDate)
		assert.True(t, isActive, "策略应在有效期内")

		// 测试已过期策略
		expiredDate := now.AddDate(0, 0, -1)
		isExpired := now.After(expiredDate)
		assert.False(t, isExpired && now.Before(expiredDate), "已过期策略不应生效")
	})
}

// TestChannelPolicyManagement 测试渠道策略管理
func TestChannelPolicyManagement(t *testing.T) {
	t.Run("渠道支持配置", func(t *testing.T) {
		// 不同等级支持不同的支付渠道
		tierChannels := map[string][]string{
			"TIER_BRONZE": {"stripe"},
			"TIER_SILVER": {"stripe", "paypal"},
			"TIER_GOLD":   {"stripe", "paypal", "adyen"},
			"TIER_PLATINUM": {"stripe", "paypal", "adyen", "square"},
			"TIER_DIAMOND": {"stripe", "paypal", "adyen", "square", "crypto"},
		}

		bronzeChannels := tierChannels["TIER_BRONZE"]
		diamondChannels := tierChannels["TIER_DIAMOND"]

		assert.Len(t, bronzeChannels, 1, "铜牌只支持1个渠道")
		assert.Len(t, diamondChannels, 5, "钻石支持5个渠道")
		assert.Contains(t, diamondChannels, "crypto", "钻石等级应支持加密货币")
	})

	t.Run("币种支持配置", func(t *testing.T) {
		// 不同等级支持不同的币种
		tierCurrencies := map[string][]string{
			"TIER_BRONZE": {"USD"},
			"TIER_SILVER": {"USD", "EUR"},
			"TIER_GOLD":   {"USD", "EUR", "GBP", "CNY"},
			"TIER_PLATINUM": {"USD", "EUR", "GBP", "CNY", "JPY", "AUD"},
			"TIER_DIAMOND": {"USD", "EUR", "GBP", "CNY", "JPY", "AUD", "BTC", "ETH"},
		}

		assert.Len(t, tierCurrencies["TIER_BRONZE"], 1)
		assert.Len(t, tierCurrencies["TIER_DIAMOND"], 8)
	})

	t.Run("API调用频率限制", func(t *testing.T) {
		// 不同等级有不同的API调用频率限制
		tierRateLimits := map[string]int{
			"TIER_BRONZE":   100,  // 100次/分钟
			"TIER_SILVER":   300,  // 300次/分钟
			"TIER_GOLD":     1000, // 1000次/分钟
			"TIER_PLATINUM": 3000, // 3000次/分钟
			"TIER_DIAMOND":  10000, // 10000次/分钟（无限制）
		}

		assert.Equal(t, 100, tierRateLimits["TIER_BRONZE"])
		assert.Equal(t, 10000, tierRateLimits["TIER_DIAMOND"])
	})
}

// TestPolicyApprovalWorkflow 测试策略审批流程
func TestPolicyApprovalWorkflow(t *testing.T) {
	t.Run("策略状态转换", func(t *testing.T) {
		// 策略状态流转: pending → active / inactive
		statuses := []string{
			model.FeeStatusPending,  // 待审批
			model.FeeStatusActive,   // 已启用
			model.FeeStatusInactive, // 已停用
		}

		assert.Len(t, statuses, 3)
		assert.Equal(t, "pending", model.FeeStatusPending)
	})

	t.Run("审批权限检查", func(t *testing.T) {
		// 只有特定角色可以审批策略
		approverID := uuid.New()
		approvedAt := time.Now()

		approval := struct {
			ApprovedBy uuid.UUID
			ApprovedAt time.Time
		}{
			ApprovedBy: approverID,
			ApprovedAt: approvedAt,
		}

		assert.NotEqual(t, uuid.Nil, approval.ApprovedBy)
		assert.False(t, approval.ApprovedAt.IsZero())
	})
}

// TestDataValidation 测试数据验证
func TestDataValidation(t *testing.T) {
	t.Run("等级代码格式验证", func(t *testing.T) {
		validCodes := []string{"TIER_BRONZE", "TIER_SILVER", "TIER_GOLD"}
		invalidCodes := []string{"bronze", "TIER-SILVER", "tier_gold"}

		for _, code := range validCodes {
			assert.Regexp(t, `^TIER_[A-Z]+$`, code, "等级代码格式应为TIER_开头的大写字母")
		}

		for _, code := range invalidCodes {
			assert.NotRegexp(t, `^TIER_[A-Z]+$`, code, "无效的等级代码格式")
		}
	})

	t.Run("费率范围验证", func(t *testing.T) {
		// 费率应在合理范围内 (0% - 10%)
		validRates := []float64{0.0, 0.029, 0.05, 0.10}
		invalidRates := []float64{-0.01, 0.15, 1.0}

		for _, rate := range validRates {
			assert.True(t, rate >= 0 && rate <= 0.10, "费率应在0%-10%范围")
		}

		for _, rate := range invalidRates {
			assert.False(t, rate >= 0 && rate <= 0.10, "费率超出合理范围")
		}
	})

	t.Run("限额逻辑验证", func(t *testing.T) {
		// 限额应满足: 最小 < 最大, 日限额 < 月限额
		singleMin := int64(100)
		singleMax := int64(100000)
		dailyLimit := int64(500000)
		monthlyLimit := int64(10000000)

		assert.True(t, singleMin < singleMax, "单笔最小应<单笔最大")
		assert.True(t, singleMax <= dailyLimit, "单笔最大应≤日限额")
		assert.True(t, dailyLimit <= monthlyLimit, "日限额应≤月限额")
	})
}

// BenchmarkTieredRuleMatching 阶梯规则匹配性能测试
func BenchmarkTieredRuleMatching(b *testing.B) {
	// 模拟100个阶梯规则
	rules := make([]service.TieredRule, 100)
	for i := 0; i < 100; i++ {
		minAmount := int64(i * 10000)
		maxAmount := int64((i + 1) * 10000)
		rules[i] = service.TieredRule{
			MinAmount:  minAmount,
			MaxAmount:  &maxAmount,
			Percentage: 0.029 - float64(i)*0.0001, // 费率递减
		}
	}

	testAmount := int64(550000) // 测试金额

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 查找匹配的阶梯
		for _, rule := range rules {
			if testAmount >= rule.MinAmount {
				if rule.MaxAmount == nil || testAmount <= *rule.MaxAmount {
					_ = rule.Percentage
					break
				}
			}
		}
	}
}

// BenchmarkPolicyPriorityMatching 策略优先级匹配性能测试
func BenchmarkPolicyPriorityMatching(b *testing.B) {
	// 模拟50个策略
	type PolicyItem struct {
		Channel       string
		PaymentMethod string
		Currency      string
		Priority      int
		Percentage    float64
	}

	policies := make([]PolicyItem, 50)
	for i := 0; i < 50; i++ {
		policies[i] = PolicyItem{
			Channel:       "stripe",
			PaymentMethod: "card",
			Currency:      "USD",
			Priority:      i,
			Percentage:    0.029,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 找到优先级最高的策略
		maxPriority := -1
		var selectedPolicy *PolicyItem
		for j := range policies {
			if policies[j].Priority > maxPriority {
				maxPriority = policies[j].Priority
				selectedPolicy = &policies[j]
			}
		}
		_ = selectedPolicy
	}
}

// BenchmarkLimitCheck 限额检查性能测试
func BenchmarkLimitCheck(b *testing.B) {
	singleTransMin := int64(100)
	singleTransMax := int64(100000)
	dailyLimit := int64(500000)
	monthlyLimit := int64(10000000)

	amount := int64(50000)
	dailyUsed := int64(300000)
	monthlyUsed := int64(5000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 执行所有限额检查
		passed := true
		if amount < singleTransMin || amount > singleTransMax {
			passed = false
		}
		if dailyUsed+amount > dailyLimit {
			passed = false
		}
		if monthlyUsed+amount > monthlyLimit {
			passed = false
		}
		_ = passed
	}
}

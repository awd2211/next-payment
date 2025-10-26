package handler

import (
	"github.com/gin-gonic/gin"
	"payment-platform/merchant-policy-service/internal/service"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	router *gin.Engine,
	authMiddleware gin.HandlerFunc,
	tierService service.TierService,
	policyEngineService service.PolicyEngineService,
	policyBindingService service.PolicyBindingService,
) {
	// 创建 Handler 实例
	tierHandler := NewTierHandler(tierService)
	policyEngineHandler := NewPolicyEngineHandler(policyEngineService)
	policyBindingHandler := NewPolicyBindingHandler(policyBindingService)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 应用认证中间件（所有接口需要认证）
		v1.Use(authMiddleware)

		// Tier (商户等级) 路由
		tiers := v1.Group("/tiers")
		{
			tiers.POST("", tierHandler.CreateTier)                  // 创建等级
			tiers.GET("", tierHandler.ListTiers)                    // 列表查询
			tiers.GET("/active", tierHandler.GetAllActiveTiers)     // 获取所有活跃等级
			tiers.GET("/code/:code", tierHandler.GetTierByCode)     // 根据代码查询
			tiers.GET("/:id", tierHandler.GetTierByID)              // 查询详情
			tiers.PUT("/:id", tierHandler.UpdateTier)               // 更新等级
			tiers.DELETE("/:id", tierHandler.DeleteTier)            // 删除等级
		}

		// PolicyEngine (策略引擎) 路由 - 最核心的接口
		policyEngine := v1.Group("/policy-engine")
		{
			policyEngine.GET("/fee-policy", policyEngineHandler.GetEffectiveFeePolicy)       // 获取有效费率策略
			policyEngine.GET("/limit-policy", policyEngineHandler.GetEffectiveLimitPolicy)   // 获取有效限额策略
			policyEngine.POST("/calculate-fee", policyEngineHandler.CalculateFee)            // 计算费用
			policyEngine.POST("/check-limit", policyEngineHandler.CheckLimit)                // 检查限额
		}

		// PolicyBinding (策略绑定) 路由
		policyBindings := v1.Group("/policy-bindings")
		{
			policyBindings.POST("/bind", policyBindingHandler.BindMerchantToTier)           // 绑定商户到等级
			policyBindings.POST("/change-tier", policyBindingHandler.ChangeMerchantTier)    // 变更商户等级
			policyBindings.POST("/custom-policy", policyBindingHandler.SetCustomPolicy)     // 设置自定义策略
			policyBindings.GET("/:merchant_id", policyBindingHandler.GetMerchantBinding)    // 获取商户绑定
			policyBindings.DELETE("/:merchant_id", policyBindingHandler.UnbindMerchant)     // 解绑商户
		}
	}
}

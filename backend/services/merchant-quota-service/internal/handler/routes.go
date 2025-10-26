package handler

import (
	"github.com/gin-gonic/gin"
	"payment-platform/merchant-quota-service/internal/service"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	router *gin.Engine,
	authMiddleware gin.HandlerFunc,
	quotaService service.QuotaService,
	alertService service.AlertService,
) {
	// 创建 Handler 实例
	quotaHandler := NewQuotaHandler(quotaService)
	alertHandler := NewAlertHandler(alertService)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 应用认证中间件（所有接口需要认证）
		v1.Use(authMiddleware)

		// Quota (配额) 路由
		quotas := v1.Group("/quotas")
		{
			quotas.POST("/initialize", quotaHandler.InitializeQuota)   // 初始化配额
			quotas.POST("/consume", quotaHandler.ConsumeQuota)         // 消耗配额 (交易时调用)
			quotas.POST("/release", quotaHandler.ReleaseQuota)         // 释放配额 (退款时调用)
			quotas.POST("/adjust", quotaHandler.AdjustQuota)           // 调整配额 (管理员操作)
			quotas.POST("/suspend", quotaHandler.SuspendQuota)         // 暂停配额
			quotas.POST("/resume", quotaHandler.ResumeQuota)           // 恢复配额
			quotas.GET("", quotaHandler.GetQuota)                      // 查询配额
			quotas.GET("/list", quotaHandler.ListQuotas)               // 列表查询
		}

		// Alert (预警) 路由
		alerts := v1.Group("/alerts")
		{
			alerts.POST("/check", alertHandler.CheckMerchantQuotaAlert)    // 检查商户预警
			alerts.POST("/:alert_id/resolve", alertHandler.ResolveAlert)   // 标记预警为已处理
			alerts.GET("/active", alertHandler.GetActiveAlerts)            // 获取活跃预警
			alerts.GET("", alertHandler.ListAlerts)                        // 列表查询
		}
	}
}

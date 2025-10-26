package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// RiskBFFHandler Risk Service BFF处理器
type RiskBFFHandler struct {
	riskClient *client.ServiceClient
}

// NewRiskBFFHandler 创建Risk BFF处理器
func NewRiskBFFHandler(riskServiceURL string) *RiskBFFHandler {
	return &RiskBFFHandler{
		riskClient: client.NewServiceClient(riskServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *RiskBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/risk")
	admin.Use(authMiddleware)
	{
		// 风控规则管理
		rules := admin.Group("/rules")
		{
			rules.POST("", h.CreateRule)
			rules.GET("/:id", h.GetRule)
			rules.GET("", h.ListRules)
			rules.PUT("/:id", h.UpdateRule)
			rules.DELETE("/:id", h.DeleteRule)
			rules.POST("/:id/enable", h.EnableRule)
			rules.POST("/:id/disable", h.DisableRule)
		}

		// 黑名单管理
		blacklist := admin.Group("/blacklist")
		{
			blacklist.POST("", h.AddToBlacklist)
			blacklist.GET("/:id", h.GetBlacklistItem)
			blacklist.GET("", h.ListBlacklist)
			blacklist.DELETE("/:id", h.RemoveFromBlacklist)
		}

		// 风控检查记录（监控）
		checks := admin.Group("/checks")
		{
			checks.GET("", h.ListRiskChecks)
			checks.GET("/statistics", h.GetRiskStatistics)
		}
	}
}

// ========== 风控规则管理 ==========

// CreateRule 创建风控规则
func (h *RiskBFFHandler) CreateRule(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.riskClient.Post(c.Request.Context(), "/api/v1/rules", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetRule 获取风控规则
func (h *RiskBFFHandler) GetRule(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/rules/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListRules 获取风控规则列表
func (h *RiskBFFHandler) ListRules(c *gin.Context) {
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if ruleType := c.Query("rule_type"); ruleType != "" {
		queryParams["rule_type"] = ruleType
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/rules", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateRule 更新风控规则
func (h *RiskBFFHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.riskClient.Put(c.Request.Context(), "/api/v1/rules/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteRule 删除风控规则
func (h *RiskBFFHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.riskClient.Delete(c.Request.Context(), "/api/v1/rules/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// EnableRule 启用风控规则
func (h *RiskBFFHandler) EnableRule(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.riskClient.Post(c.Request.Context(), "/api/v1/rules/"+id+"/enable", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DisableRule 禁用风控规则
func (h *RiskBFFHandler) DisableRule(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.riskClient.Post(c.Request.Context(), "/api/v1/rules/"+id+"/disable", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 黑名单管理 ==========

// AddToBlacklist 添加到黑名单
func (h *RiskBFFHandler) AddToBlacklist(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.riskClient.Post(c.Request.Context(), "/api/v1/blacklist", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetBlacklistItem 获取黑名单项
func (h *RiskBFFHandler) GetBlacklistItem(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/blacklist/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListBlacklist 获取黑名单列表
func (h *RiskBFFHandler) ListBlacklist(c *gin.Context) {
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if itemType := c.Query("type"); itemType != "" {
		queryParams["type"] = itemType
	}

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/blacklist", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// RemoveFromBlacklist 从黑名单移除
func (h *RiskBFFHandler) RemoveFromBlacklist(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.riskClient.Delete(c.Request.Context(), "/api/v1/blacklist/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 风控检查记录 ==========

// ListRiskChecks 获取风控检查记录列表
func (h *RiskBFFHandler) ListRiskChecks(c *gin.Context) {
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if riskLevel := c.Query("risk_level"); riskLevel != "" {
		queryParams["risk_level"] = riskLevel
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/checks", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetRiskStatistics 获取风控统计数据
func (h *RiskBFFHandler) GetRiskStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/checks/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Risk Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

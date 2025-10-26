package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// CashierBFFHandler Cashier Service BFF处理器
type CashierBFFHandler struct {
	cashierClient *client.ServiceClient
}

// NewCashierBFFHandler 创建Cashier BFF处理器
func NewCashierBFFHandler(cashierServiceURL string) *CashierBFFHandler {
	return &CashierBFFHandler{
		cashierClient: client.NewServiceClient(cashierServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *CashierBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/cashier")
	admin.Use(authMiddleware)
	{
		// 收银台模板管理
		templates := admin.Group("/templates")
		{
			templates.POST("", h.CreateTemplate)
			templates.GET("/:id", h.GetTemplate)
			templates.GET("", h.ListTemplates)
			templates.PUT("/:id", h.UpdateTemplate)
			templates.DELETE("/:id", h.DeleteTemplate)
			templates.PUT("/:id/activate", h.ActivateTemplate)
			templates.PUT("/:id/deactivate", h.DeactivateTemplate)
		}

		// 样式配置
		styles := admin.Group("/styles")
		{
			styles.POST("", h.CreateStyle)
			styles.GET("/:id", h.GetStyle)
			styles.GET("", h.ListStyles)
			styles.PUT("/:id", h.UpdateStyle)
			styles.DELETE("/:id", h.DeleteStyle)
		}

		// 字段配置
		fields := admin.Group("/fields")
		{
			fields.POST("", h.CreateField)
			fields.GET("/:id", h.GetField)
			fields.GET("", h.ListFields)
			fields.PUT("/:id", h.UpdateField)
			fields.DELETE("/:id", h.DeleteField)
		}
	}
}

// ========== 收银台模板管理 ==========

// CreateTemplate 创建收银台模板
func (h *CashierBFFHandler) CreateTemplate(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.cashierClient.Post(c.Request.Context(), "/api/v1/admin/cashier/templates", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetTemplate 获取收银台模板详情
func (h *CashierBFFHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/admin/cashier/templates/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListTemplates 获取收银台模板列表
func (h *CashierBFFHandler) ListTemplates(c *gin.Context) {
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if templateType := c.Query("type"); templateType != "" {
		queryParams["type"] = templateType
	}

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/admin/cashier/templates", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateTemplate 更新收银台模板
func (h *CashierBFFHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.cashierClient.Put(c.Request.Context(), "/api/v1/admin/cashier/templates/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteTemplate 删除收银台模板
func (h *CashierBFFHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Delete(c.Request.Context(), "/api/v1/admin/cashier/templates/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ActivateTemplate 激活收银台模板
func (h *CashierBFFHandler) ActivateTemplate(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Put(c.Request.Context(), "/api/v1/admin/cashier/templates/"+id+"/activate", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeactivateTemplate 停用收银台模板
func (h *CashierBFFHandler) DeactivateTemplate(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Put(c.Request.Context(), "/api/v1/admin/cashier/templates/"+id+"/deactivate", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 样式配置管理 ==========

// CreateStyle 创建样式配置
func (h *CashierBFFHandler) CreateStyle(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.cashierClient.Post(c.Request.Context(), "/api/v1/admin/cashier/styles", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetStyle 获取样式配置详情
func (h *CashierBFFHandler) GetStyle(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/admin/cashier/styles/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListStyles 获取样式配置列表
func (h *CashierBFFHandler) ListStyles(c *gin.Context) {
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/admin/cashier/styles", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateStyle 更新样式配置
func (h *CashierBFFHandler) UpdateStyle(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.cashierClient.Put(c.Request.Context(), "/api/v1/admin/cashier/styles/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteStyle 删除样式配置
func (h *CashierBFFHandler) DeleteStyle(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Delete(c.Request.Context(), "/api/v1/admin/cashier/styles/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 字段配置管理 ==========

// CreateField 创建字段配置
func (h *CashierBFFHandler) CreateField(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.cashierClient.Post(c.Request.Context(), "/api/v1/admin/cashier/fields", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetField 获取字段配置详情
func (h *CashierBFFHandler) GetField(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/admin/cashier/fields/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListFields 获取字段配置列表
func (h *CashierBFFHandler) ListFields(c *gin.Context) {
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/admin/cashier/fields", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateField 更新字段配置
func (h *CashierBFFHandler) UpdateField(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.cashierClient.Put(c.Request.Context(), "/api/v1/admin/cashier/fields/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteField 删除字段配置
func (h *CashierBFFHandler) DeleteField(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.cashierClient.Delete(c.Request.Context(), "/api/v1/admin/cashier/fields/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Cashier Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

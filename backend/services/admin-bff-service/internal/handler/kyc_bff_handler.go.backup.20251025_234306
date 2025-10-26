package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// KYCBFFHandler KYC Service BFF处理器
type KYCBFFHandler struct {
	kycClient *client.ServiceClient
}

// NewKYCBFFHandler 创建KYC BFF处理器
func NewKYCBFFHandler(kycServiceURL string) *KYCBFFHandler {
	return &KYCBFFHandler{
		kycClient: client.NewServiceClient(kycServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *KYCBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/kyc")
	admin.Use(authMiddleware)
	{
		// 文档审核
		documents := admin.Group("/documents")
		{
			documents.GET("", h.ListAllDocuments)
			documents.GET("/pending", h.ListPendingDocuments)
			documents.GET("/:id", h.GetDocument)
			documents.POST("/:id/approve", h.ApproveDocument)
			documents.POST("/:id/reject", h.RejectDocument)
		}

		// 资质审核
		qualifications := admin.Group("/qualifications")
		{
			qualifications.GET("", h.ListAllQualifications)
			qualifications.GET("/pending", h.ListPendingQualifications)
			qualifications.GET("/:id", h.GetQualification)
			qualifications.POST("/:id/approve", h.ApproveQualification)
			qualifications.POST("/:id/reject", h.RejectQualification)
		}

		// 商户等级管理
		levels := admin.Group("/levels")
		{
			levels.GET("/statistics", h.GetLevelStatistics)
			levels.POST("/:merchant_id/upgrade", h.UpgradeLevel)
			levels.POST("/:merchant_id/downgrade", h.DowngradeLevel)
		}

		// 告警管理
		alerts := admin.Group("/alerts")
		{
			alerts.GET("", h.ListAlerts)
			alerts.POST("/:id/resolve", h.ResolveAlert)
		}
	}
}

// ========== 文档审核 ==========

// ListAllDocuments 获取所有KYC文档
func (h *KYCBFFHandler) ListAllDocuments(c *gin.Context) {
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
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if docType := c.Query("type"); docType != "" {
		queryParams["type"] = docType
	}

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/documents", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListPendingDocuments 获取待审核文档
func (h *KYCBFFHandler) ListPendingDocuments(c *gin.Context) {
	queryParams := make(map[string]string)
	queryParams["status"] = "pending"
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/documents", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetDocument 获取文档详情
func (h *KYCBFFHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/documents/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ApproveDocument 批准文档
func (h *KYCBFFHandler) ApproveDocument(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	// 添加审核人信息（从JWT获取）
	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/documents/"+id+"/approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// RejectDocument 拒绝文档
func (h *KYCBFFHandler) RejectDocument(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 添加审核人信息（从JWT获取）
	adminID := c.GetString("user_id")
	req["rejected_by"] = adminID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/documents/"+id+"/reject", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 资质审核 ==========

// ListAllQualifications 获取所有资质
func (h *KYCBFFHandler) ListAllQualifications(c *gin.Context) {
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
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/qualifications", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListPendingQualifications 获取待审核资质
func (h *KYCBFFHandler) ListPendingQualifications(c *gin.Context) {
	queryParams := make(map[string]string)
	queryParams["status"] = "pending"
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/qualifications", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetQualification 获取资质详情
func (h *KYCBFFHandler) GetQualification(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/qualifications/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ApproveQualification 批准资质
func (h *KYCBFFHandler) ApproveQualification(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	// 添加审核人信息
	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/qualifications/"+id+"/approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// RejectQualification 拒绝资质
func (h *KYCBFFHandler) RejectQualification(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 添加审核人信息
	adminID := c.GetString("user_id")
	req["rejected_by"] = adminID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/qualifications/"+id+"/reject", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 商户等级管理 ==========

// GetLevelStatistics 获取等级统计
func (h *KYCBFFHandler) GetLevelStatistics(c *gin.Context) {
	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/levels/statistics", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpgradeLevel 升级商户等级
func (h *KYCBFFHandler) UpgradeLevel(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	// 添加操作人信息
	adminID := c.GetString("user_id")
	req["operator_id"] = adminID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/levels/"+merchantID+"/upgrade", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DowngradeLevel 降级商户等级
func (h *KYCBFFHandler) DowngradeLevel(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	// 添加操作人信息
	adminID := c.GetString("user_id")
	req["operator_id"] = adminID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/levels/"+merchantID+"/downgrade", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 告警管理 ==========

// ListAlerts 获取告警列表
func (h *KYCBFFHandler) ListAlerts(c *gin.Context) {
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
	if severity := c.Query("severity"); severity != "" {
		queryParams["severity"] = severity
	}

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/alerts", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ResolveAlert 解决告警
func (h *KYCBFFHandler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	// 添加操作人信息
	adminID := c.GetString("user_id")
	req["resolved_by"] = adminID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/alerts/"+id+"/resolve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用KYC Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/service"
)

// BusinessHandler 业务处理器（聚合所有业务功能的API）
type BusinessHandler struct {
	businessService service.BusinessService
}

// NewBusinessHandler 创建业务处理器实例
func NewBusinessHandler(businessService service.BusinessService) *BusinessHandler {
	return &BusinessHandler{
		businessService: businessService,
	}
}

// RegisterRoutes 注册业务路由
func (h *BusinessHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	business := r.Group("")
	business.Use(authMiddleware) // 所有业务API都需要认证
	{
		// 结算账户路由
		accounts := business.Group("/settlement-accounts")
		{
			accounts.POST("", h.CreateSettlementAccount)
			accounts.GET("", h.GetSettlementAccounts)
			accounts.PUT("/:id", h.UpdateSettlementAccount)
			accounts.DELETE("/:id", h.DeleteSettlementAccount)
			accounts.POST("/:id/set-default", h.SetDefaultAccount)
			accounts.POST("/:id/verify", h.VerifySettlementAccount)
		}

		// KYC文档路由
		kyc := business.Group("/kyc-documents")
		{
			kyc.POST("", h.UploadKYCDocument)
			kyc.GET("", h.GetKYCDocuments)
			kyc.POST("/:id/review", h.ReviewKYCDocument)
			kyc.DELETE("/:id", h.DeleteKYCDocument)
		}

		// 费率配置路由
		fees := business.Group("/fee-configs")
		{
			fees.POST("", h.CreateFeeConfig)
			fees.GET("", h.GetFeeConfigs)
			fees.PUT("/:id", h.UpdateFeeConfig)
			fees.DELETE("/:id", h.DeleteFeeConfig)
		}

		// 子账户路由
		users := business.Group("/users")
		{
			users.POST("/invite", h.InviteUser)
			users.GET("", h.GetMerchantUsers)
			users.PUT("/:id", h.UpdateMerchantUser)
			users.DELETE("/:id", h.DeleteMerchantUser)
		}

		// 交易限额路由
		limits := business.Group("/transaction-limits")
		{
			limits.POST("", h.CreateTransactionLimit)
			limits.GET("", h.GetTransactionLimits)
			limits.PUT("/:id", h.UpdateTransactionLimit)
			limits.DELETE("/:id", h.DeleteTransactionLimit)
		}

		// 业务资质路由
		qualifications := business.Group("/qualifications")
		{
			qualifications.POST("", h.CreateQualification)
			qualifications.GET("", h.GetQualifications)
			qualifications.POST("/:id/verify", h.VerifyQualification)
			qualifications.DELETE("/:id", h.DeleteQualification)
		}
	}
}

// ==================== 结算账户 API ====================

// CreateSettlementAccount 创建结算账户
// @Summary 创建结算账户
// @Tags SettlementAccount
// @Accept json
// @Produce json
// @Param request body service.CreateSettlementAccountInput true "结算账户信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/settlement-accounts [post]
func (h *BusinessHandler) CreateSettlementAccount(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateSettlementAccountInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.MerchantID = merchantID.(uuid.UUID)

	account, err := h.businessService.CreateSettlementAccount(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    account,
	})
}

// GetSettlementAccounts 获取结算账户列表
// @Summary 获取结算账户列表
// @Tags SettlementAccount
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/settlement-accounts [get]
func (h *BusinessHandler) GetSettlementAccounts(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	accounts, err := h.businessService.GetSettlementAccounts(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": accounts,
	})
}

// UpdateSettlementAccount 更新结算账户
// @Summary 更新结算账户
// @Tags SettlementAccount
// @Accept json
// @Produce json
// @Param id path string true "账户ID"
// @Param request body service.UpdateSettlementAccountInput true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/settlement-accounts/{id} [put]
func (h *BusinessHandler) UpdateSettlementAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	var req service.UpdateSettlementAccountInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.businessService.UpdateSettlementAccount(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    account,
	})
}

// DeleteSettlementAccount 删除结算账户
// @Summary 删除结算账户
// @Tags SettlementAccount
// @Produce json
// @Param id path string true "账户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/settlement-accounts/{id} [delete]
func (h *BusinessHandler) DeleteSettlementAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	if err := h.businessService.DeleteSettlementAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// SetDefaultAccount 设置默认结算账户
// @Summary 设置默认结算账户
// @Tags SettlementAccount
// @Produce json
// @Param id path string true "账户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/settlement-accounts/{id}/set-default [post]
func (h *BusinessHandler) SetDefaultAccount(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	if err := h.businessService.SetDefaultAccount(c.Request.Context(), merchantID.(uuid.UUID), accountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设置成功",
	})
}

// VerifySettlementAccount 验证结算账户
// @Summary 验证结算账户（管理员）
// @Tags SettlementAccount
// @Accept json
// @Produce json
// @Param id path string true "账户ID"
// @Param request body map[string]string true "status,reason"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/settlement-accounts/{id}/verify [post]
func (h *BusinessHandler) VerifySettlementAccount(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.businessService.VerifySettlementAccount(c.Request.Context(), accountID, req.Status, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "验证成功",
	})
}

// ==================== KYC文档 API ====================

// UploadKYCDocument 上传KYC文档
// @Summary 上传KYC文档
// @Tags KYCDocument
// @Accept json
// @Produce json
// @Param request body service.UploadKYCDocumentInput true "文档信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kyc-documents [post]
func (h *BusinessHandler) UploadKYCDocument(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.UploadKYCDocumentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.MerchantID = merchantID.(uuid.UUID)

	doc, err := h.businessService.UploadKYCDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "上传成功",
		"data":    doc,
	})
}

// GetKYCDocuments 获取KYC文档列表
// @Summary 获取KYC文档列表
// @Tags KYCDocument
// @Produce json
// @Param document_type query string false "文档类型"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kyc-documents [get]
func (h *BusinessHandler) GetKYCDocuments(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	documentType := c.Query("document_type")

	docs, err := h.businessService.GetKYCDocuments(c.Request.Context(), merchantID.(uuid.UUID), documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": docs,
	})
}

// ReviewKYCDocument 审核KYC文档
// @Summary 审核KYC文档（管理员）
// @Tags KYCDocument
// @Accept json
// @Produce json
// @Param id path string true "文档ID"
// @Param request body map[string]string true "status,review_notes"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kyc-documents/{id}/review [post]
func (h *BusinessHandler) ReviewKYCDocument(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Status      string `json:"status" binding:"required"`
		ReviewNotes string `json:"review_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.businessService.ReviewKYCDocument(c.Request.Context(), docID, req.Status, req.ReviewNotes, reviewerID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "审核成功",
	})
}

// DeleteKYCDocument 删除KYC文档
// @Summary 删除KYC文档
// @Tags KYCDocument
// @Produce json
// @Param id path string true "文档ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kyc-documents/{id} [delete]
func (h *BusinessHandler) DeleteKYCDocument(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID"})
		return
	}

	if err := h.businessService.DeleteKYCDocument(c.Request.Context(), docID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// ==================== 费率配置 API ====================

// CreateFeeConfig 创建费率配置
// @Summary 创建费率配置
// @Tags FeeConfig
// @Accept json
// @Produce json
// @Param request body service.CreateFeeConfigInput true "费率配置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fee-configs [post]
func (h *BusinessHandler) CreateFeeConfig(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateFeeConfigInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.MerchantID = merchantID.(uuid.UUID)

	config, err := h.businessService.CreateFeeConfig(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    config,
	})
}

// GetFeeConfigs 获取费率配置列表
// @Summary 获取费率配置列表
// @Tags FeeConfig
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fee-configs [get]
func (h *BusinessHandler) GetFeeConfigs(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	configs, err := h.businessService.GetFeeConfigs(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": configs,
	})
}

// UpdateFeeConfig 更新费率配置
// @Summary 更新费率配置
// @Tags FeeConfig
// @Accept json
// @Produce json
// @Param id path string true "配置ID"
// @Param request body service.UpdateFeeConfigInput true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fee-configs/{id} [put]
func (h *BusinessHandler) UpdateFeeConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	var req service.UpdateFeeConfigInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.businessService.UpdateFeeConfig(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    config,
	})
}

// DeleteFeeConfig 删除费率配置
// @Summary 删除费率配置
// @Tags FeeConfig
// @Produce json
// @Param id path string true "配置ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fee-configs/{id} [delete]
func (h *BusinessHandler) DeleteFeeConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	if err := h.businessService.DeleteFeeConfig(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// ==================== 子账户 API ====================

// InviteUser 邀请子账户
// @Summary 邀请子账户
// @Tags MerchantUser
// @Accept json
// @Produce json
// @Param request body service.InviteUserInput true "邀请信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/invite [post]
func (h *BusinessHandler) InviteUser(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.InviteUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.MerchantID = merchantID.(uuid.UUID)
	req.InvitedBy = merchantID.(uuid.UUID)

	user, err := h.businessService.InviteUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "邀请成功",
		"data":    user,
	})
}

// GetMerchantUsers 获取子账户列表
// @Summary 获取子账户列表
// @Tags MerchantUser
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users [get]
func (h *BusinessHandler) GetMerchantUsers(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	users, err := h.businessService.GetMerchantUsers(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// UpdateMerchantUser 更新子账户
// @Summary 更新子账户
// @Tags MerchantUser
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body service.UpdateMerchantUserInput true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{id} [put]
func (h *BusinessHandler) UpdateMerchantUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req service.UpdateMerchantUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.businessService.UpdateMerchantUser(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    user,
	})
}

// DeleteMerchantUser 删除子账户
// @Summary 删除子账户
// @Tags MerchantUser
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{id} [delete]
func (h *BusinessHandler) DeleteMerchantUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := h.businessService.DeleteMerchantUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// ==================== 交易限额 API ====================

// CreateTransactionLimit 创建交易限额
// @Summary 创建交易限额
// @Tags TransactionLimit
// @Accept json
// @Produce json
// @Param request body service.CreateTransactionLimitInput true "限额配置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/transaction-limits [post]
func (h *BusinessHandler) CreateTransactionLimit(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateTransactionLimitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.MerchantID = merchantID.(uuid.UUID)

	limit, err := h.businessService.CreateTransactionLimit(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    limit,
	})
}

// GetTransactionLimits 获取交易限额列表
// @Summary 获取交易限额列表
// @Tags TransactionLimit
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/transaction-limits [get]
func (h *BusinessHandler) GetTransactionLimits(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	limits, err := h.businessService.GetTransactionLimits(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": limits,
	})
}

// UpdateTransactionLimit 更新交易限额
// @Summary 更新交易限额
// @Tags TransactionLimit
// @Accept json
// @Produce json
// @Param id path string true "限额ID"
// @Param request body service.UpdateTransactionLimitInput true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/transaction-limits/{id} [put]
func (h *BusinessHandler) UpdateTransactionLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的限额ID"})
		return
	}

	var req service.UpdateTransactionLimitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit, err := h.businessService.UpdateTransactionLimit(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    limit,
	})
}

// DeleteTransactionLimit 删除交易限额
// @Summary 删除交易限额
// @Tags TransactionLimit
// @Produce json
// @Param id path string true "限额ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/transaction-limits/{id} [delete]
func (h *BusinessHandler) DeleteTransactionLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的限额ID"})
		return
	}

	if err := h.businessService.DeleteTransactionLimit(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// ==================== 业务资质 API ====================

// CreateQualification 创建业务资质
// @Summary 创建业务资质
// @Tags BusinessQualification
// @Accept json
// @Produce json
// @Param request body service.CreateQualificationInput true "资质信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/qualifications [post]
func (h *BusinessHandler) CreateQualification(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateQualificationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.MerchantID = merchantID.(uuid.UUID)

	qualification, err := h.businessService.CreateQualification(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    qualification,
	})
}

// GetQualifications 获取业务资质列表
// @Summary 获取业务资质列表
// @Tags BusinessQualification
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/qualifications [get]
func (h *BusinessHandler) GetQualifications(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	qualifications, err := h.businessService.GetQualifications(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": qualifications,
	})
}

// VerifyQualification 验证业务资质
// @Summary 验证业务资质（管理员）
// @Tags BusinessQualification
// @Accept json
// @Produce json
// @Param id path string true "资质ID"
// @Param request body map[string]string true "status"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/qualifications/{id}/verify [post]
func (h *BusinessHandler) VerifyQualification(c *gin.Context) {
	qualificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资质ID"})
		return
	}

	verifierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.businessService.VerifyQualification(c.Request.Context(), qualificationID, req.Status, verifierID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "验证成功",
	})
}

// DeleteQualification 删除业务资质
// @Summary 删除业务资质
// @Tags BusinessQualification
// @Produce json
// @Param id path string true "资质ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/qualifications/{id} [delete]
func (h *BusinessHandler) DeleteQualification(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的资质ID"})
		return
	}

	if err := h.businessService.DeleteQualification(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

package handler

import (
	"net/http"
	"strconv"

	"payment-platform/merchant-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
)

// MerchantHandler 商户处理器
type MerchantHandler struct {
	merchantService service.MerchantService
}

// NewMerchantHandler 创建商户处理器实例
func NewMerchantHandler(merchantService service.MerchantService) *MerchantHandler {
	return &MerchantHandler{
		merchantService: merchantService,
	}
}

// Register 商户注册
// @Summary 商户注册
// @Tags Merchant
// @Accept json
// @Produce json
// @Param request body service.RegisterMerchantInput true "注册信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/register [post]
func (h *MerchantHandler) Register(c *gin.Context) {
	var req service.RegisterMerchantInput
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchant, err := h.merchantService.Register(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "注册失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "注册成功，请等待审核",
		"data":    merchant,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Login 商户登录
// @Summary 商户登录
// @Tags Merchant
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/login [post]
func (h *MerchantHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	result, err := h.merchantService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "登录失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "登录成功",
		"data":    result,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Create 创建商户（管理员操作）
// @Summary 创建商户
// @Tags Merchant
// @Accept json
// @Produce json
// @Param request body service.CreateMerchantInput true "商户信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant [post]
func (h *MerchantHandler) Create(c *gin.Context) {
	var req service.CreateMerchantInput
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchant, err := h.merchantService.Create(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建商户失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "创建成功",
		"data":    merchant,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetByID 根据ID获取商户
// @Summary 获取商户详情
// @Tags Merchant
// @Produce json
// @Param id path string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/{id} [get]
func (h *MerchantHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchant, err := h.merchantService.GetByID(c.Request.Context(), id)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取商户失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(merchant).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// List 获取商户列表
// @Summary 获取商户列表
// @Tags Merchant
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态筛选"
// @Param kyc_status query string false "KYC状态筛选"
// @Param keyword query string false "关键词搜索"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant [get]
func (h *MerchantHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	kycStatus := c.Query("kyc_status")
	keyword := c.Query("keyword")

	merchants, total, err := h.merchantService.List(c.Request.Context(), page, pageSize, status, kycStatus, keyword)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取商户列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"list":      merchants,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Update 更新商户
// @Summary 更新商户信息
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path string true "商户ID"
// @Param request body service.UpdateMerchantInput true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/{id} [put]
func (h *MerchantHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req service.UpdateMerchantInput
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchant, err := h.merchantService.Update(c.Request.Context(), id, &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新商户失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "更新成功",
		"data":    merchant,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus 更新商户状态
// @Summary 更新商户状态
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path string true "商户ID"
// @Param request body UpdateStatusRequest true "状态信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/{id}/status [put]
func (h *MerchantHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.merchantService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新商户状态失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "状态更新成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateKYCStatusRequest 更新KYC状态请求
type UpdateKYCStatusRequest struct {
	KYCStatus string `json:"kyc_status" binding:"required"`
}

// UpdateKYCStatus 更新KYC状态
// @Summary 更新KYC状态
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path string true "商户ID"
// @Param request body UpdateKYCStatusRequest true "KYC状态信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/{id}/kyc-status [put]
func (h *MerchantHandler) UpdateKYCStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req UpdateKYCStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.merchantService.UpdateKYCStatus(c.Request.Context(), id, req.KYCStatus); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新KYC状态失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "KYC状态更新成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Delete 删除商户
// @Summary 删除商户
// @Tags Merchant
// @Produce json
// @Param id path string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/{id} [delete]
func (h *MerchantHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.merchantService.Delete(c.Request.Context(), id); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "删除商户失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "删除成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetProfile 获取当前商户信息
// @Summary 获取当前商户信息
// @Tags Merchant
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/profile [get]
func (h *MerchantHandler) GetProfile(c *gin.Context) {
	// 从JWT中获取商户ID
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	merchant, err := h.merchantService.GetByID(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取商户信息失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(merchant).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateProfile 更新当前商户信息
// @Summary 更新当前商户信息
// @Tags Merchant
// @Accept json
// @Produce json
// @Param request body service.UpdateMerchantInput true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/profile [put]
func (h *MerchantHandler) UpdateProfile(c *gin.Context) {
	// 从JWT中获取商户ID
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req service.UpdateMerchantInput
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchant, err := h.merchantService.Update(c.Request.Context(), merchantID.(uuid.UUID), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新商户信息失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "更新成功",
		"data":    merchant,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// MerchantWithPassword 带密码的商户信息响应（仅用于内部接口）
type MerchantWithPassword struct {
	ID           uuid.UUID `json:"id"`
	MerchantNo   string    `json:"merchant_no"`
	Name         string    `json:"merchant_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Status       string    `json:"status"`
	PasswordHash string    `json:"password_hash"` // 仅内部接口返回
}

// GetWithPassword 获取带密码的商户信息（内部接口，供merchant-auth-service调用）
// @Summary 获取带密码的商户信息
// @Tags Merchant
// @Produce json
// @Param id path string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchants/{id}/with-password [get]
func (h *MerchantHandler) GetWithPassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchant, err := h.merchantService.GetByIDWithPassword(c.Request.Context(), id)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取商户信息失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	// 构造包含密码哈希的响应（仅内部接口）
	response := MerchantWithPassword{
		ID:           merchant.ID,
		MerchantNo:   "", // merchant_no字段可能不存在，保持为空
		Name:         merchant.Name,
		Email:        merchant.Email,
		Phone:        merchant.Phone,
		Status:       merchant.Status,
		PasswordHash: merchant.PasswordHash,
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(response).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdatePasswordRequest 更新密码请求
type UpdatePasswordRequest struct {
	PasswordHash string `json:"password_hash" binding:"required"`
}

// UpdatePassword 更新商户密码（内部接口，供merchant-auth-service调用）
// @Summary 更新商户密码
// @Tags Merchant
// @Accept json
// @Produce json
// @Param id path string true "商户ID"
// @Param request body UpdatePasswordRequest true "密码哈希"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchants/{id}/password [put]
func (h *MerchantHandler) UpdatePassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.merchantService.UpdatePasswordHash(c.Request.Context(), id, req.PasswordHash); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新密码失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "密码更新成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetBalance 获取商户余额（临时占位API）
// @Summary 获取商户余额
// @Tags Merchant
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/balance [get]
func (h *MerchantHandler) GetBalance(c *gin.Context) {
	// TODO: 从 accounting-service 获取真实余额数据
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"available_balance": 0,
		"frozen_balance":    0,
		"total_balance":     0,
		"currency":          "USD",
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetStats 获取商户统计（临时占位API）
// @Summary 获取商户统计
// @Tags Merchant
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/stats [get]
func (h *MerchantHandler) GetStats(c *gin.Context) {
	// TODO: 从 analytics-service 和 payment-gateway 聚合真实统计数据
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"total_transactions": 0,
		"total_amount":       0,
		"success_rate":       0,
		"today_transactions": 0,
		"today_amount":       0,
		"this_month_amount":  0,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes 注册路由
func (h *MerchantHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 公开路由（无需认证）
	public := r.Group("/merchant")
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
	}

	// 需要认证的商户路由
	merchant := r.Group("/merchant")
	merchant.Use(authMiddleware) // 启用认证中间件
	{
		merchant.GET("/profile", h.GetProfile)
		merchant.PUT("/profile", h.UpdateProfile)

		// 临时占位API：balance 和 stats（应该从其他服务聚合数据）
		merchant.GET("/balance", h.GetBalance)
		merchant.GET("/stats", h.GetStats)
	}

	// 管理员路由（需要管理员权限）
	admin := r.Group("/merchant")
	admin.Use(authMiddleware) // 管理员也需要认证
	{
		admin.POST("", h.Create)
		admin.GET("", h.List)
		admin.GET("/:id", h.GetByID)
		admin.PUT("/:id", h.Update)
		admin.PUT("/:id/status", h.UpdateStatus)
		admin.PUT("/:id/kyc-status", h.UpdateKYCStatus)
		admin.DELETE("/:id", h.Delete)
	}

	// 内部接口（供其他微服务调用，未来可添加服务间认证）
	internal := r.Group("/merchants")
	{
		internal.GET("/:id/with-password", h.GetWithPassword)
		internal.PUT("/:id/password", h.UpdatePassword)
	}
}

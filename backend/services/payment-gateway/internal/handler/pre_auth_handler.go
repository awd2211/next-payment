package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/payment-gateway/internal/service"
)

// PreAuthHandler 预授权处理器
type PreAuthHandler struct {
	preAuthService service.PreAuthService
}

// NewPreAuthHandler 创建预授权处理器
func NewPreAuthHandler(preAuthService service.PreAuthService) *PreAuthHandler {
	return &PreAuthHandler{
		preAuthService: preAuthService,
	}
}

// CreatePreAuthRequest 创建预授权请求
type CreatePreAuthRequest struct {
	OrderNo   string `json:"order_no" binding:"required"`
	Amount    int64  `json:"amount" binding:"required,min=1"`
	Currency  string `json:"currency" binding:"required"`
	Channel   string `json:"channel" binding:"required"`
	Subject   string `json:"subject" binding:"required"`
	Body      string `json:"body"`
	ClientIP  string `json:"client_ip"`
	ReturnURL string `json:"return_url"`
	NotifyURL string `json:"notify_url"`
}

// CapturePreAuthRequest 确认预授权请求
type CapturePreAuthRequest struct {
	PreAuthNo string `json:"pre_auth_no" binding:"required"`
	Amount    *int64 `json:"amount"` // 可选，不传则全额确认
}

// CancelPreAuthRequest 取消预授权请求
type CancelPreAuthRequest struct {
	PreAuthNo string `json:"pre_auth_no" binding:"required"`
	Reason    string `json:"reason"`
}

// CreatePreAuth 创建预授权
// @Summary 创建预授权
// @Description 创建预授权支付，用于两阶段支付场景（如酒店预订、租车等）
// @Tags 预授权
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param request body CreatePreAuthRequest true "创建预授权请求"
// @Success 200 {object} SuccessResponse{data=model.PreAuthPayment}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/merchant/pre-auth [post]
func (h *PreAuthHandler) CreatePreAuth(c *gin.Context) {
	// 获取商户ID
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse("未授权"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	// 解析请求
	var req CreatePreAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	// 获取客户端IP
	clientIP := req.ClientIP
	if clientIP == "" {
		clientIP = c.ClientIP()
	}

	// 调用服务层
	input := &service.CreatePreAuthInput{
		MerchantID: merchantID,
		OrderNo:    req.OrderNo,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Channel:    req.Channel,
		Subject:    req.Subject,
		Body:       req.Body,
		ClientIP:   clientIP,
		ReturnURL:  req.ReturnURL,
		NotifyURL:  req.NotifyURL,
	}

	preAuth, err := h.preAuthService.CreatePreAuth(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(preAuth))
}

// CapturePreAuth 确认预授权（扣款）
// @Summary 确认预授权
// @Description 确认预授权并扣款，可以部分确认或全额确认
// @Tags 预授权
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param request body CapturePreAuthRequest true "确认预授权请求"
// @Success 200 {object} SuccessResponse{data=model.Payment}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/merchant/pre-auth/capture [post]
func (h *PreAuthHandler) CapturePreAuth(c *gin.Context) {
	// 获取商户ID
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse("未授权"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	// 解析请求
	var req CapturePreAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	// 调用服务层
	payment, err := h.preAuthService.CapturePreAuth(c.Request.Context(), merchantID, req.PreAuthNo, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(payment))
}

// CancelPreAuth 取消预授权
// @Summary 取消预授权
// @Description 取消预授权，释放冻结的金额
// @Tags 预授权
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param request body CancelPreAuthRequest true "取消预授权请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/merchant/pre-auth/cancel [post]
func (h *PreAuthHandler) CancelPreAuth(c *gin.Context) {
	// 获取商户ID
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse("未授权"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	// 解析请求
	var req CancelPreAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	// 调用服务层
	err = h.preAuthService.CancelPreAuth(c.Request.Context(), merchantID, req.PreAuthNo, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("预授权已取消"))
}

// GetPreAuth 查询预授权
// @Summary 查询预授权
// @Description 根据预授权单号查询预授权详情
// @Tags 预授权
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param pre_auth_no path string true "预授权单号"
// @Success 200 {object} SuccessResponse{data=model.PreAuthPayment}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/merchant/pre-auth/{pre_auth_no} [get]
func (h *PreAuthHandler) GetPreAuth(c *gin.Context) {
	// 获取商户ID
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse("未授权"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	// 获取预授权单号
	preAuthNo := c.Param("pre_auth_no")
	if preAuthNo == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("预授权单号不能为空"))
		return
	}

	// 调用服务层
	preAuth, err := h.preAuthService.GetPreAuth(c.Request.Context(), merchantID, preAuthNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	if preAuth == nil {
		c.JSON(http.StatusNotFound, ErrorResponse("预授权不存在"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(preAuth))
}

// ListPreAuths 查询预授权列表
// @Summary 查询预授权列表
// @Description 查询商户的预授权列表，支持分页和状态筛选
// @Tags 预授权
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param status query string false "状态筛选: pending, authorized, captured, cancelled, expired"
// @Param page query int false "页码（默认1）"
// @Param page_size query int false "每页数量（默认20，最大100）"
// @Success 200 {object} SuccessResponse{data=[]model.PreAuthPayment}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/merchant/pre-auths [get]
func (h *PreAuthHandler) ListPreAuths(c *gin.Context) {
	// 获取商户ID
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse("未授权"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	// 获取查询参数
	status := c.Query("status")
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	// 调用服务层
	preAuths, err := h.preAuthService.ListPreAuths(c.Request.Context(), merchantID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(preAuths))
}

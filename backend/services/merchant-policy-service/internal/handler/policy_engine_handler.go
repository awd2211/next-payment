package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-policy-service/internal/service"
)

// PolicyEngineHandler 策略引擎处理器
type PolicyEngineHandler struct {
	policyEngineService service.PolicyEngineService
}

// NewPolicyEngineHandler 创建策略引擎处理器实例
func NewPolicyEngineHandler(policyEngineService service.PolicyEngineService) *PolicyEngineHandler {
	return &PolicyEngineHandler{
		policyEngineService: policyEngineService,
	}
}

// GetEffectiveFeePolicy godoc
// @Summary 获取商户有效费率策略
// @Description 根据商户ID、渠道、支付方式、币种获取有效的费率策略（优先级：商户自定义 > 等级默认）
// @Tags PolicyEngine
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param channel query string false "支付渠道 (stripe, paypal, crypto, all)"
// @Param payment_method query string false "支付方式 (card, bank_transfer, wallet, all)"
// @Param currency query string true "币种 (USD, EUR, CNY)"
// @Success 200 {object} map[string]interface{} "成功返回费率策略"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "未找到适用策略"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-engine/fee-policy [get]
func (h *PolicyEngineHandler) GetEffectiveFeePolicy(c *gin.Context) {
	// 解析参数
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "merchant_id is required"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id format"})
		return
	}

	channel := c.DefaultQuery("channel", "")
	paymentMethod := c.DefaultQuery("payment_method", "")
	currency := c.Query("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency is required"})
		return
	}

	// 调用服务层
	policy, err := h.policyEngineService.GetEffectiveFeePolicy(c.Request.Context(), merchantID, channel, paymentMethod, currency)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    policy,
	})
}

// GetEffectiveLimitPolicy godoc
// @Summary 获取商户有效限额策略
// @Description 根据商户ID、渠道、币种获取有效的限额策略（优先级：商户自定义 > 等级默认）
// @Tags PolicyEngine
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param channel query string false "支付渠道 (stripe, paypal, crypto, all)"
// @Param currency query string true "币种 (USD, EUR, CNY)"
// @Success 200 {object} map[string]interface{} "成功返回限额策略"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "未找到适用策略"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-engine/limit-policy [get]
func (h *PolicyEngineHandler) GetEffectiveLimitPolicy(c *gin.Context) {
	// 解析参数
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "merchant_id is required"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id format"})
		return
	}

	channel := c.DefaultQuery("channel", "")
	currency := c.Query("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency is required"})
		return
	}

	// 调用服务层
	policy, err := h.policyEngineService.GetEffectiveLimitPolicy(c.Request.Context(), merchantID, channel, currency)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    policy,
	})
}

// CalculateFeeRequest 计算费用请求
type CalculateFeeRequest struct {
	MerchantID    string `json:"merchant_id" binding:"required"`
	Channel       string `json:"channel"`
	PaymentMethod string `json:"payment_method"`
	Currency      string `json:"currency" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
}

// CalculateFee godoc
// @Summary 计算交易费用
// @Description 根据商户的费率策略计算交易费用
// @Tags PolicyEngine
// @Accept json
// @Produce json
// @Param body body CalculateFeeRequest true "计算费用请求"
// @Success 200 {object} map[string]interface{} "成功返回费用计算结果"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "未找到适用策略"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-engine/calculate-fee [post]
func (h *PolicyEngineHandler) CalculateFee(c *gin.Context) {
	var req CalculateFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id format"})
		return
	}

	// 调用服务层
	result, err := h.policyEngineService.CalculateFee(
		c.Request.Context(),
		merchantID,
		req.Channel,
		req.PaymentMethod,
		req.Currency,
		req.Amount,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

// CheckLimitRequest 检查限额请求
type CheckLimitRequest struct {
	MerchantID  string `json:"merchant_id" binding:"required"`
	Channel     string `json:"channel"`
	Currency    string `json:"currency" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,min=1"`
	DailyUsed   int64  `json:"daily_used" binding:"min=0"`
	MonthlyUsed int64  `json:"monthly_used" binding:"min=0"`
}

// CheckLimit godoc
// @Summary 检查交易限额
// @Description 检查交易金额是否超过单笔/日/月限额
// @Tags PolicyEngine
// @Accept json
// @Produce json
// @Param body body CheckLimitRequest true "检查限额请求"
// @Success 200 {object} map[string]interface{} "成功返回限额检查结果"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "未找到适用策略"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-engine/check-limit [post]
func (h *PolicyEngineHandler) CheckLimit(c *gin.Context) {
	var req CheckLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id format"})
		return
	}

	// 调用服务层
	result, err := h.policyEngineService.CheckLimit(
		c.Request.Context(),
		merchantID,
		req.Channel,
		req.Currency,
		req.Amount,
		req.DailyUsed,
		req.MonthlyUsed,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

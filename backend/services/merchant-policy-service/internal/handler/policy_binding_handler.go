package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-policy-service/internal/service"
)

// PolicyBindingHandler 策略绑定处理器
type PolicyBindingHandler struct {
	policyBindingService service.PolicyBindingService
}

// NewPolicyBindingHandler 创建策略绑定处理器实例
func NewPolicyBindingHandler(policyBindingService service.PolicyBindingService) *PolicyBindingHandler {
	return &PolicyBindingHandler{
		policyBindingService: policyBindingService,
	}
}

// BindMerchantToTier godoc
// @Summary 绑定商户到等级
// @Description 将商户绑定到指定等级（新商户注册时调用）
// @Tags PolicyBinding
// @Accept json
// @Produce json
// @Param body body service.BindMerchantInput true "绑定商户请求"
// @Success 200 {object} SuccessResponse "成功返回绑定信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-bindings/bind [post]
func (h *PolicyBindingHandler) BindMerchantToTier(c *gin.Context) {
	var input service.BindMerchantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	binding, err := h.policyBindingService.BindMerchantToTier(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "merchant bound to tier successfully",
		Data:    binding,
	})
}

// ChangeMerchantTier godoc
// @Summary 变更商户等级
// @Description 升级或降级商户等级
// @Tags PolicyBinding
// @Accept json
// @Produce json
// @Param body body service.ChangeTierInput true "变更等级请求"
// @Success 200 {object} SuccessResponse "成功返回变更后的绑定信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-bindings/change-tier [post]
func (h *PolicyBindingHandler) ChangeMerchantTier(c *gin.Context) {
	var input service.ChangeTierInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	binding, err := h.policyBindingService.ChangeMerchantTier(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "merchant tier changed successfully",
		Data:    binding,
	})
}

// SetCustomPolicy godoc
// @Summary 设置商户自定义策略
// @Description 为商户设置自定义费率或限额策略，覆盖等级默认策略
// @Tags PolicyBinding
// @Accept json
// @Produce json
// @Param body body service.SetCustomPolicyInput true "设置自定义策略请求"
// @Success 200 {object} SuccessResponse "成功返回更新后的绑定信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-bindings/custom-policy [post]
func (h *PolicyBindingHandler) SetCustomPolicy(c *gin.Context) {
	var input service.SetCustomPolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	binding, err := h.policyBindingService.SetCustomPolicy(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "custom policy set successfully",
		Data:    binding,
	})
}

// GetMerchantBinding godoc
// @Summary 获取商户策略绑定
// @Description 获取商户当前的等级绑定和自定义策略信息
// @Tags PolicyBinding
// @Accept json
// @Produce json
// @Param merchant_id path string true "商户ID"
// @Success 200 {object} SuccessResponse "成功返回绑定详情"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "绑定不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-bindings/{merchant_id} [get]
func (h *PolicyBindingHandler) GetMerchantBinding(c *gin.Context) {
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	detail, err := h.policyBindingService.GetMerchantBinding(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    detail,
	})
}

// UnbindMerchant godoc
// @Summary 解绑商户
// @Description 删除商户的等级绑定（谨慎操作）
// @Tags PolicyBinding
// @Accept json
// @Produce json
// @Param merchant_id path string true "商户ID"
// @Success 200 {object} SuccessResponse "成功解绑"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "绑定不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /policy-bindings/{merchant_id} [delete]
func (h *PolicyBindingHandler) UnbindMerchant(c *gin.Context) {
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	if err := h.policyBindingService.UnbindMerchant(c.Request.Context(), merchantID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "merchant unbound successfully",
		Data:    nil,
	})
}

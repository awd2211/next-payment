package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-policy-service/internal/service"
)

// TierHandler 商户等级处理器
type TierHandler struct {
	tierService service.TierService
}

// NewTierHandler 创建等级处理器实例
func NewTierHandler(tierService service.TierService) *TierHandler {
	return &TierHandler{
		tierService: tierService,
	}
}

// CreateTier godoc
// @Summary 创建商户等级
// @Description 创建新的商户等级（需管理员权限）
// @Tags Tier
// @Accept json
// @Produce json
// @Param body body service.CreateTierInput true "创建等级请求"
// @Success 200 {object} SuccessResponse "成功返回等级信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /tiers [post]
func (h *TierHandler) CreateTier(c *gin.Context) {
	var input service.CreateTierInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	tier, err := h.tierService.CreateTier(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    tier,
	})
}

// UpdateTier godoc
// @Summary 更新商户等级
// @Description 更新指定等级的信息（需管理员权限）
// @Tags Tier
// @Accept json
// @Produce json
// @Param id path string true "等级ID"
// @Param body body service.UpdateTierInput true "更新等级请求"
// @Success 200 {object} SuccessResponse "成功返回更新后的等级信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "等级不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /tiers/{id} [put]
func (h *TierHandler) UpdateTier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tier id"})
		return
	}

	var input service.UpdateTierInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	tier, err := h.tierService.UpdateTier(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    tier,
	})
}

// DeleteTier godoc
// @Summary 删除商户等级
// @Description 删除指定的等级（需管理员权限）
// @Tags Tier
// @Accept json
// @Produce json
// @Param id path string true "等级ID"
// @Success 200 {object} SuccessResponse "成功删除"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "等级不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /tiers/{id} [delete]
func (h *TierHandler) DeleteTier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tier id"})
		return
	}

	if err := h.tierService.DeleteTier(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "tier deleted successfully",
		Data:    nil,
	})
}

// GetTierByID godoc
// @Summary 获取等级详情
// @Description 根据ID获取等级详细信息
// @Tags Tier
// @Accept json
// @Produce json
// @Param id path string true "等级ID"
// @Success 200 {object} SuccessResponse "成功返回等级信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "等级不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /tiers/{id} [get]
func (h *TierHandler) GetTierByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid tier id"})
		return
	}

	tier, err := h.tierService.GetTierByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    tier,
	})
}

// GetTierByCode godoc
// @Summary 根据代码获取等级
// @Description 根据等级代码（如 starter, basic, professional）获取等级信息
// @Tags Tier
// @Accept json
// @Produce json
// @Param code path string true "等级代码"
// @Success 200 {object} SuccessResponse "成功返回等级信息"
// @Failure 404 {object} ErrorResponse "等级不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /tiers/code/{code} [get]
func (h *TierHandler) GetTierByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tier code is required"})
		return
	}

	tier, err := h.tierService.GetTierByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    tier,
	})
}

// ListTiers godoc
// @Summary 获取等级列表
// @Description 获取所有商户等级列表，支持分页和筛选
// @Tags Tier
// @Accept json
// @Produce json
// @Param is_active query boolean false "是否启用"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} ListResponse "成功返回等级列表"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /tiers [get]
func (h *TierHandler) ListTiers(c *gin.Context) {
	var isActive *bool
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		val, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid is_active value"})
			return
		}
		isActive = &val
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	// 验证并限制分页参数（防止DoS攻击）
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制100条/页
	}

	result, err := h.tierService.ListTiers(c.Request.Context(), isActive, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{
		Code:       0,
		Message:    "success",
		Data:       result.Tiers,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}

// GetAllActiveTiers godoc
// @Summary 获取所有活跃等级
// @Description 获取所有处于活跃状态的等级（不分页）
// @Tags Tier
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse "成功返回活跃等级列表"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /tiers/active [get]
func (h *TierHandler) GetAllActiveTiers(c *gin.Context) {
	tiers, err := h.tierService.GetAllActiveTiers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    tiers,
	})
}

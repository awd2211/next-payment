package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/services/risk-service/internal/repository"
	"github.com/payment-platform/services/risk-service/internal/service"
)

// RiskHandler 风控处理器
type RiskHandler struct {
	riskService service.RiskService
}

// NewRiskHandler 创建风控处理器实例
func NewRiskHandler(riskService service.RiskService) *RiskHandler {
	return &RiskHandler{
		riskService: riskService,
	}
}

// RegisterRoutes 注册路由
func (h *RiskHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// 风控规则管理
		rules := v1.Group("/rules")
		{
			rules.POST("", h.CreateRule)
			rules.GET("/:id", h.GetRule)
			rules.GET("", h.ListRules)
			rules.PUT("/:id", h.UpdateRule)
			rules.DELETE("/:id", h.DeleteRule)
			rules.POST("/:id/enable", h.EnableRule)
			rules.POST("/:id/disable", h.DisableRule)
		}

		// 风控检查
		checks := v1.Group("/checks")
		{
			checks.POST("/payment", h.CheckPayment)
			checks.GET("/:id", h.GetCheck)
			checks.GET("", h.ListChecks)
		}

		// 黑名单管理
		blacklist := v1.Group("/blacklist")
		{
			blacklist.POST("", h.AddBlacklist)
			blacklist.DELETE("/:id", h.RemoveBlacklist)
			blacklist.GET("/check", h.CheckBlacklist)
			blacklist.GET("", h.ListBlacklist)
		}
	}
}

// Rule Management

func (h *RiskHandler) CreateRule(c *gin.Context) {
	var input service.CreateRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	rule, err := h.riskService.CreateRule(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(rule))
}

func (h *RiskHandler) GetRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的规则ID"))
		return
	}

	rule, err := h.riskService.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(rule))
}

func (h *RiskHandler) ListRules(c *gin.Context) {
	query := &repository.RuleQuery{
		RuleType: c.Query("rule_type"),
		Status:   c.Query("status"),
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	rules, total, err := h.riskService.ListRules(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     rules,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

func (h *RiskHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的规则ID"))
		return
	}

	var input service.UpdateRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	rule, err := h.riskService.UpdateRule(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(rule))
}

func (h *RiskHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的规则ID"))
		return
	}

	if err := h.riskService.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

func (h *RiskHandler) EnableRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的规则ID"))
		return
	}

	if err := h.riskService.EnableRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

func (h *RiskHandler) DisableRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的规则ID"))
		return
	}

	if err := h.riskService.DisableRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// Risk Checks

func (h *RiskHandler) CheckPayment(c *gin.Context) {
	var input service.PaymentCheckInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	check, err := h.riskService.CheckPayment(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(check))
}

func (h *RiskHandler) GetCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的检查ID"))
		return
	}

	check, err := h.riskService.GetCheck(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(check))
}

func (h *RiskHandler) ListChecks(c *gin.Context) {
	query := &repository.CheckQuery{
		RelatedType: c.Query("related_type"),
		Decision:    c.Query("decision"),
		RiskLevel:   c.Query("risk_level"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query.EndTime = &endTime
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	checks, total, err := h.riskService.ListChecks(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     checks,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// Blacklist Management

func (h *RiskHandler) AddBlacklist(c *gin.Context) {
	var input service.AddBlacklistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	blacklist, err := h.riskService.AddBlacklist(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(blacklist))
}

func (h *RiskHandler) RemoveBlacklist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的黑名单ID"))
		return
	}

	if err := h.riskService.RemoveBlacklist(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

func (h *RiskHandler) CheckBlacklist(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityValue := c.Query("entity_value")

	if entityType == "" || entityValue == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("entity_type和entity_value不能为空"))
		return
	}

	hit, blacklist, err := h.riskService.CheckBlacklist(c.Request.Context(), entityType, entityValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"hit":       hit,
		"blacklist": blacklist,
	}))
}

func (h *RiskHandler) ListBlacklist(c *gin.Context) {
	query := &repository.BlacklistQuery{
		EntityType: c.Query("entity_type"),
		Status:     c.Query("status"),
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	blacklists, total, err := h.riskService.ListBlacklist(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     blacklists,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// Response structures

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func ErrorResponse(message string) Response {
	return Response{
		Code:    -1,
		Message: message,
	}
}

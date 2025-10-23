package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/service"
)

// SettlementHandler 结算处理器
type SettlementHandler struct {
	settlementService service.SettlementService
}

// NewSettlementHandler 创建结算处理器
func NewSettlementHandler(settlementService service.SettlementService) *SettlementHandler {
	return &SettlementHandler{
		settlementService: settlementService,
	}
}

// RegisterRoutes 注册路由
func (h *SettlementHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		settlements := api.Group("/settlements")
		{
			settlements.POST("", h.CreateSettlement)
			settlements.GET("", h.ListSettlements)
			settlements.GET("/:id", h.GetSettlement)
			settlements.POST("/:id/approve", h.ApproveSettlement)
			settlements.POST("/:id/reject", h.RejectSettlement)
			settlements.POST("/:id/execute", h.ExecuteSettlement)
			settlements.GET("/reports", h.GetSettlementReport)
		}
	}
}

// CreateSettlementRequest 创建结算单请求
type CreateSettlementRequest struct {
	MerchantID   string                     `json:"merchant_id" binding:"required"`
	Cycle        model.SettlementCycle      `json:"cycle" binding:"required"`
	StartDate    string                     `json:"start_date" binding:"required"`
	EndDate      string                     `json:"end_date" binding:"required"`
	Transactions []TransactionItemRequest   `json:"transactions" binding:"required"`
}

// TransactionItemRequest 交易明细请求
type TransactionItemRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	OrderNo       string `json:"order_no" binding:"required"`
	PaymentNo     string `json:"payment_no" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
	Fee           int64  `json:"fee" binding:"min=0"`
	TransactionAt string `json:"transaction_at" binding:"required"`
}

// CreateSettlement 创建结算单
// @Summary 创建结算单
// @Tags Settlements
// @Accept json
// @Produce json
// @Param request body CreateSettlementRequest true "创建结算单请求"
// @Success 200 {object} map[string]interface{}
// @Router /settlements [post]
func (h *SettlementHandler) CreateSettlement(c *gin.Context) {
	var req CreateSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期"})
		return
	}

	transactions := make([]service.TransactionItem, 0, len(req.Transactions))
	for _, tx := range req.Transactions {
		txID, err := uuid.Parse(tx.TransactionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的交易ID"})
			return
		}

		txTime, err := time.Parse(time.RFC3339, tx.TransactionAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的交易时间"})
			return
		}

		transactions = append(transactions, service.TransactionItem{
			TransactionID: txID,
			OrderNo:       tx.OrderNo,
			PaymentNo:     tx.PaymentNo,
			Amount:        tx.Amount,
			Fee:           tx.Fee,
			TransactionAt: txTime,
		})
	}

	input := &service.CreateSettlementInput{
		MerchantID:   merchantID,
		Cycle:        req.Cycle,
		StartDate:    startDate,
		EndDate:      endDate,
		Transactions: transactions,
	}

	settlement, err := h.settlementService.CreateSettlement(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": settlement,
	})
}

// GetSettlement 获取结算单详情
// @Summary 获取结算单详情
// @Tags Settlements
// @Accept json
// @Produce json
// @Param id path string true "结算单ID"
// @Success 200 {object} map[string]interface{}
// @Router /settlements/{id} [get]
func (h *SettlementHandler) GetSettlement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结算单ID"})
		return
	}

	detail, err := h.settlementService.GetSettlement(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": detail,
	})
}

// ListSettlements 结算单列表
// @Summary 结算单列表
// @Tags Settlements
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Param status query string false "状态"
// @Param cycle query string false "周期"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /settlements [get]
func (h *SettlementHandler) ListSettlements(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
			return
		}
		merchantID = &id
	}

	var status *model.SettlementStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := model.SettlementStatus(statusStr)
		status = &s
	}

	var cycle *model.SettlementCycle
	if cycleStr := c.Query("cycle"); cycleStr != "" {
		cy := model.SettlementCycle(cycleStr)
		cycle = &cy
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	query := &service.ListSettlementQuery{
		MerchantID: merchantID,
		Status:     status,
		Cycle:      cycle,
		Page:       page,
		PageSize:   pageSize,
	}

	resp, err := h.settlementService.ListSettlements(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// ApproveSettlementRequest 审批结算单请求
type ApproveSettlementRequest struct {
	ApproverID   string `json:"approver_id" binding:"required"`
	ApproverName string `json:"approver_name" binding:"required"`
	Comments     string `json:"comments"`
}

// ApproveSettlement 审批通过结算单
// @Summary 审批通过结算单
// @Tags Settlements
// @Accept json
// @Produce json
// @Param id path string true "结算单ID"
// @Param request body ApproveSettlementRequest true "审批请求"
// @Success 200 {object} map[string]interface{}
// @Router /settlements/{id}/approve [post]
func (h *SettlementHandler) ApproveSettlement(c *gin.Context) {
	idStr := c.Param("id")
	settlementID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结算单ID"})
		return
	}

	var req ApproveSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审批人ID"})
		return
	}

	err = h.settlementService.ApproveSettlement(c.Request.Context(), settlementID, approverID, req.ApproverName, req.Comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "审批通过",
	})
}

// RejectSettlement 拒绝结算单
// @Summary 拒绝结算单
// @Tags Settlements
// @Accept json
// @Produce json
// @Param id path string true "结算单ID"
// @Param request body ApproveSettlementRequest true "拒绝请求"
// @Success 200 {object} map[string]interface{}
// @Router /settlements/{id}/reject [post]
func (h *SettlementHandler) RejectSettlement(c *gin.Context) {
	idStr := c.Param("id")
	settlementID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结算单ID"})
		return
	}

	var req ApproveSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审批人ID"})
		return
	}

	err = h.settlementService.RejectSettlement(c.Request.Context(), settlementID, approverID, req.ApproverName, req.Comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "已拒绝",
	})
}

// ExecuteSettlement 执行结算
// @Summary 执行结算
// @Tags Settlements
// @Accept json
// @Produce json
// @Param id path string true "结算单ID"
// @Success 200 {object} map[string]interface{}
// @Router /settlements/{id}/execute [post]
func (h *SettlementHandler) ExecuteSettlement(c *gin.Context) {
	idStr := c.Param("id")
	settlementID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结算单ID"})
		return
	}

	err = h.settlementService.ExecuteSettlement(c.Request.Context(), settlementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "结算执行成功",
	})
}

// GetSettlementReport 获取结算报表
// @Summary 获取结算报表
// @Tags Settlements
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param start_date query string true "开始日期"
// @Param end_date query string true "结束日期"
// @Success 200 {object} map[string]interface{}
// @Router /settlements/reports [get]
func (h *SettlementHandler) GetSettlementReport(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商户ID不能为空"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期"})
		return
	}

	report, err := h.settlementService.GetSettlementReport(c.Request.Context(), merchantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": report,
	})
}

package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payment-platform/dispute-service/internal/service"
)

// DisputeHandler 拒付HTTP处理器
type DisputeHandler struct {
	service service.DisputeService
}

// NewDisputeHandler 创建处理器实例
func NewDisputeHandler(service service.DisputeService) *DisputeHandler {
	return &DisputeHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *DisputeHandler) RegisterRoutes(router *gin.RouterGroup) {
	disputes := router.Group("/disputes")
	{
		// Dispute management
		disputes.POST("", h.CreateDispute)
		disputes.GET("", h.ListDisputes)
		disputes.GET("/:dispute_id", h.GetDisputeDetails)
		disputes.PUT("/:dispute_id/status", h.UpdateStatus)
		disputes.POST("/:dispute_id/assign", h.AssignDispute)

		// Evidence management
		disputes.POST("/:dispute_id/evidence", h.UploadEvidence)
		disputes.GET("/:dispute_id/evidence", h.ListEvidence)
		disputes.DELETE("/evidence/:evidence_id", h.DeleteEvidence)

		// Stripe operations
		disputes.POST("/:dispute_id/submit", h.SubmitToStripe)
		disputes.POST("/sync/:channel_dispute_id", h.SyncFromStripe)

		// Statistics
		disputes.GET("/statistics", h.GetStatistics)
	}
}

// CreateDispute 创建拒付
func (h *DisputeHandler) CreateDispute(c *gin.Context) {
	var req service.CreateDisputeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	dispute, err := h.service.CreateDispute(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("CREATE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(dispute))
}

// GetDisputeDetails 获取拒付详情
func (h *DisputeHandler) GetDisputeDetails(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("dispute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DISPUTE_ID", "Invalid dispute ID format"))
		return
	}

	details, err := h.service.GetDisputeByID(c.Request.Context(), disputeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(details))
}

// ListDisputes 查询拒付列表
func (h *DisputeHandler) ListDisputes(c *gin.Context) {
	filters := &service.DisputeFilters{}

	// Parse filters
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant_id format"))
			return
		}
		filters.MerchantID = &merchantID
	}

	if assignedToStr := c.Query("assigned_to"); assignedToStr != "" {
		assignedTo, err := uuid.Parse(assignedToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_ASSIGNED_TO", "Invalid assigned_to format"))
			return
		}
		filters.AssignedTo = &assignedTo
	}

	if evidenceSubmittedStr := c.Query("evidence_submitted"); evidenceSubmittedStr != "" {
		evidenceSubmitted, err := strconv.ParseBool(evidenceSubmittedStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_PARAM", "Invalid evidence_submitted format"))
			return
		}
		filters.EvidenceSubmitted = &evidenceSubmitted
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid start_date format"))
			return
		}
		filters.StartDate = &startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid end_date format"))
			return
		}
		filters.EndDate = &endDate
	}

	filters.Channel = c.Query("channel")
	filters.Status = c.Query("status")
	filters.Reason = c.Query("reason")
	filters.PaymentNo = c.Query("payment_no")

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.service.ListDisputes(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("LIST_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

// UpdateStatus 更新拒付状态
func (h *DisputeHandler) UpdateStatus(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("dispute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DISPUTE_ID", "Invalid dispute ID format"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	if err := h.service.UpdateDisputeStatus(c.Request.Context(), disputeID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("UPDATE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Status updated successfully"}))
}

// AssignDispute 分配拒付
func (h *DisputeHandler) AssignDispute(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("dispute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DISPUTE_ID", "Invalid dispute ID format"))
		return
	}

	var req struct {
		AssignedTo string `json:"assigned_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	assignedTo, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_USER_ID", "Invalid assigned_to format"))
		return
	}

	if err := h.service.AssignDispute(c.Request.Context(), disputeID, assignedTo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("ASSIGN_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Dispute assigned successfully"}))
}

// UploadEvidence 上传证据
func (h *DisputeHandler) UploadEvidence(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("dispute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DISPUTE_ID", "Invalid dispute ID format"))
		return
	}

	var req service.UploadEvidenceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	req.DisputeID = disputeID

	evidence, err := h.service.UploadEvidence(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("UPLOAD_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(evidence))
}

// ListEvidence 查询证据列表
func (h *DisputeHandler) ListEvidence(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("dispute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DISPUTE_ID", "Invalid dispute ID format"))
		return
	}

	evidence, err := h.service.ListEvidence(c.Request.Context(), disputeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("LIST_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(evidence))
}

// DeleteEvidence 删除证据
func (h *DisputeHandler) DeleteEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("evidence_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_EVIDENCE_ID", "Invalid evidence ID format"))
		return
	}

	if err := h.service.DeleteEvidence(c.Request.Context(), evidenceID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("DELETE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Evidence deleted successfully"}))
}

// SubmitToStripe 提交证据到Stripe
func (h *DisputeHandler) SubmitToStripe(c *gin.Context) {
	disputeID, err := uuid.Parse(c.Param("dispute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DISPUTE_ID", "Invalid dispute ID format"))
		return
	}

	if err := h.service.SubmitToStripe(c.Request.Context(), disputeID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("SUBMIT_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Evidence submitted to Stripe successfully"}))
}

// SyncFromStripe 从Stripe同步拒付数据
func (h *DisputeHandler) SyncFromStripe(c *gin.Context) {
	channelDisputeID := c.Param("channel_dispute_id")
	if channelDisputeID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_PARAM", "channel_dispute_id is required"))
		return
	}

	dispute, err := h.service.SyncFromStripe(c.Request.Context(), channelDisputeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("SYNC_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(dispute))
}

// GetStatistics 获取拒付统计信息
func (h *DisputeHandler) GetStatistics(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant_id format"))
			return
		}
		merchantID = &id
	}

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		sd, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid start_date format"))
			return
		}
		startDate = &sd
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		ed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid end_date format"))
			return
		}
		endDate = &ed
	}

	stats, err := h.service.GetStatistics(c.Request.Context(), merchantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_STATS_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(stats))
}

// Response helpers

type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    "SUCCESS",
		Message: "操作成功",
		Data:    data,
	}
}

func ErrorResponse(code, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

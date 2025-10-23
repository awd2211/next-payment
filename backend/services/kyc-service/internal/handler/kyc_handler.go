package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/kyc-service/internal/model"
	"payment-platform/kyc-service/internal/service"
)

// KYCHandler KYC处理器
type KYCHandler struct {
	kycService service.KYCService
}

// NewKYCHandler 创建KYC处理器
func NewKYCHandler(kycService service.KYCService) *KYCHandler {
	return &KYCHandler{
		kycService: kycService,
	}
}

// RegisterRoutes 注册路由
func (h *KYCHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		documents := api.Group("/documents")
		{
			documents.POST("", h.SubmitDocument)
			documents.GET("", h.ListDocuments)
			documents.GET("/:id", h.GetDocument)
			documents.POST("/:id/approve", h.ApproveDocument)
			documents.POST("/:id/reject", h.RejectDocument)
		}

		qualifications := api.Group("/qualifications")
		{
			qualifications.POST("", h.SubmitQualification)
			qualifications.GET("", h.ListQualifications)
			qualifications.GET("/merchant/:merchant_id", h.GetQualification)
			qualifications.POST("/:id/approve", h.ApproveQualification)
			qualifications.POST("/:id/reject", h.RejectQualification)
		}

		levels := api.Group("/levels")
		{
			levels.GET("/:merchant_id", h.GetMerchantLevel)
			levels.GET("/:merchant_id/eligibility", h.CheckMerchantEligibility)
		}

		alerts := api.Group("/alerts")
		{
			alerts.GET("", h.ListAlerts)
			alerts.POST("/:id/resolve", h.ResolveAlert)
		}

		api.GET("/statistics", h.GetKYCStatistics)
	}
}

// Document Handlers

// SubmitDocumentRequest 提交文档请求
type SubmitDocumentRequest struct {
	MerchantID     string               `json:"merchant_id" binding:"required"`
	DocumentType   model.DocumentType   `json:"document_type" binding:"required"`
	DocumentNumber string               `json:"document_number"`
	DocumentURL    string               `json:"document_url" binding:"required"`
	FrontImageURL  string               `json:"front_image_url"`
	BackImageURL   string               `json:"back_image_url"`
	IssueDate      string               `json:"issue_date"`
	ExpiryDate     string               `json:"expiry_date"`
	IssuingCountry string               `json:"issuing_country"`
}

// SubmitDocument 提交KYC文档
// @Summary 提交KYC文档
// @Tags Documents
// @Accept json
// @Produce json
// @Param request body SubmitDocumentRequest true "提交文档请求"
// @Success 200 {object} map[string]interface{}
// @Router /documents [post]
func (h *KYCHandler) SubmitDocument(c *gin.Context) {
	var req SubmitDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	input := &service.SubmitDocumentInput{
		MerchantID:     merchantID,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		DocumentURL:    req.DocumentURL,
		FrontImageURL:  req.FrontImageURL,
		BackImageURL:   req.BackImageURL,
		IssuingCountry: req.IssuingCountry,
	}

	if req.IssueDate != "" {
		issueDate, err := time.Parse("2006-01-02", req.IssueDate)
		if err == nil {
			input.IssueDate = &issueDate
		}
	}

	if req.ExpiryDate != "" {
		expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate)
		if err == nil {
			input.ExpiryDate = &expiryDate
		}
	}

	document, err := h.kycService.SubmitDocument(c.Request.Context(), input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "提交文档失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(document).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetDocument 获取文档详情
// @Summary 获取文档详情
// @Tags Documents
// @Accept json
// @Produce json
// @Param id path string true "文档ID"
// @Success 200 {object} map[string]interface{}
// @Router /documents/{id} [get]
func (h *KYCHandler) GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的文档ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	document, err := h.kycService.GetDocument(c.Request.Context(), id)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取文档失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(document).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListDocuments 文档列表
// @Summary 文档列表
// @Tags Documents
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Param document_type query string false "文档类型"
// @Param status query string false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /documents [get]
func (h *KYCHandler) ListDocuments(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		merchantID = &id
	}

	var documentType *model.DocumentType
	if docTypeStr := c.Query("document_type"); docTypeStr != "" {
		dt := model.DocumentType(docTypeStr)
		documentType = &dt
	}

	var status *model.KYCStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := model.KYCStatus(statusStr)
		status = &s
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

	query := &service.ListDocumentQuery{
		MerchantID:   merchantID,
		DocumentType: documentType,
		Status:       status,
		Page:         page,
		PageSize:     pageSize,
	}

	result, err := h.kycService.ListDocuments(c.Request.Context(), query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取文档列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(result).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ReviewRequest 审核请求
type ReviewRequest struct {
	ReviewerID   string `json:"reviewer_id" binding:"required"`
	ReviewerName string `json:"reviewer_name" binding:"required"`
	Comments     string `json:"comments"`
	Reason       string `json:"reason"`
}

// ApproveDocument 审批通过文档
// @Summary 审批通过文档
// @Tags Documents
// @Accept json
// @Produce json
// @Param id path string true "文档ID"
// @Param request body ReviewRequest true "审批请求"
// @Success 200 {object} map[string]interface{}
// @Router /documents/{id}/approve [post]
func (h *KYCHandler) ApproveDocument(c *gin.Context) {
	idStr := c.Param("id")
	documentID, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的文档ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	reviewerID, err := uuid.Parse(req.ReviewerID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的审核人ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.kycService.ApproveDocument(c.Request.Context(), documentID, reviewerID, req.ReviewerName, req.Comments)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "审批失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "审批通过"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RejectDocument 拒绝文档
// @Summary 拒绝文档
// @Tags Documents
// @Accept json
// @Produce json
// @Param id path string true "文档ID"
// @Param request body ReviewRequest true "拒绝请求"
// @Success 200 {object} map[string]interface{}
// @Router /documents/{id}/reject [post]
func (h *KYCHandler) RejectDocument(c *gin.Context) {
	idStr := c.Param("id")
	documentID, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的文档ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	reviewerID, err := uuid.Parse(req.ReviewerID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的审核人ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.kycService.RejectDocument(c.Request.Context(), documentID, reviewerID, req.ReviewerName, req.Reason)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "拒绝失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "已拒绝"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Qualification Handlers

// SubmitQualificationRequest 提交企业资质请求
type SubmitQualificationRequest struct {
	MerchantID                    string `json:"merchant_id" binding:"required"`
	CompanyName                   string `json:"company_name" binding:"required"`
	BusinessLicenseNo             string `json:"business_license_no" binding:"required"`
	BusinessLicenseURL            string `json:"business_license_url" binding:"required"`
	LegalPersonName               string `json:"legal_person_name" binding:"required"`
	LegalPersonIDCard             string `json:"legal_person_id_card" binding:"required"`
	LegalPersonIDCardFrontURL     string `json:"legal_person_id_card_front_url"`
	LegalPersonIDCardBackURL      string `json:"legal_person_id_card_back_url"`
	RegisteredAddress             string `json:"registered_address"`
	RegisteredCapital             int64  `json:"registered_capital"`
	EstablishedDate               string `json:"established_date"`
	BusinessScope                 string `json:"business_scope"`
	Industry                      string `json:"industry"`
	TaxRegistrationNo             string `json:"tax_registration_no"`
	TaxRegistrationURL            string `json:"tax_registration_url"`
	OrganizationCode              string `json:"organization_code"`
}

// SubmitQualification 提交企业资质
// @Summary 提交企业资质
// @Tags Qualifications
// @Accept json
// @Produce json
// @Param request body SubmitQualificationRequest true "提交企业资质请求"
// @Success 200 {object} map[string]interface{}
// @Router /qualifications [post]
func (h *KYCHandler) SubmitQualification(c *gin.Context) {
	var req SubmitQualificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	input := &service.SubmitQualificationInput{
		MerchantID:                    merchantID,
		CompanyName:                   req.CompanyName,
		BusinessLicenseNo:             req.BusinessLicenseNo,
		BusinessLicenseURL:            req.BusinessLicenseURL,
		LegalPersonName:               req.LegalPersonName,
		LegalPersonIDCard:             req.LegalPersonIDCard,
		LegalPersonIDCardFrontURL:     req.LegalPersonIDCardFrontURL,
		LegalPersonIDCardBackURL:      req.LegalPersonIDCardBackURL,
		RegisteredAddress:             req.RegisteredAddress,
		RegisteredCapital:             req.RegisteredCapital,
		BusinessScope:                 req.BusinessScope,
		Industry:                      req.Industry,
		TaxRegistrationNo:             req.TaxRegistrationNo,
		TaxRegistrationURL:            req.TaxRegistrationURL,
		OrganizationCode:              req.OrganizationCode,
	}

	if req.EstablishedDate != "" {
		establishedDate, err := time.Parse("2006-01-02", req.EstablishedDate)
		if err == nil {
			input.EstablishedDate = &establishedDate
		}
	}

	qualification, err := h.kycService.SubmitQualification(c.Request.Context(), input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "提交企业资质失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(qualification).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetQualification 获取企业资质
// @Summary 获取企业资质
// @Tags Qualifications
// @Accept json
// @Produce json
// @Param merchant_id path string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /qualifications/merchant/{merchant_id} [get]
func (h *KYCHandler) GetQualification(c *gin.Context) {
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	qualification, err := h.kycService.GetQualification(c.Request.Context(), merchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取企业资质失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(qualification).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListQualifications 企业资质列表
// @Summary 企业资质列表
// @Tags Qualifications
// @Accept json
// @Produce json
// @Param status query string false "状态"
// @Param industry query string false "行业"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /qualifications [get]
func (h *KYCHandler) ListQualifications(c *gin.Context) {
	var status *model.KYCStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := model.KYCStatus(statusStr)
		status = &s
	}

	var industry *string
	if industryStr := c.Query("industry"); industryStr != "" {
		industry = &industryStr
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

	query := &service.ListQualificationQuery{
		Status:   status,
		Industry: industry,
		Page:     page,
		PageSize: pageSize,
	}

	result, err := h.kycService.ListQualifications(c.Request.Context(), query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取企业资质列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(result).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ApproveQualification 审批通过企业资质
// @Summary 审批通过企业资质
// @Tags Qualifications
// @Accept json
// @Produce json
// @Param id path string true "资质ID"
// @Param request body ReviewRequest true "审批请求"
// @Success 200 {object} map[string]interface{}
// @Router /qualifications/{id}/approve [post]
func (h *KYCHandler) ApproveQualification(c *gin.Context) {
	idStr := c.Param("id")
	qualID, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的资质ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	reviewerID, err := uuid.Parse(req.ReviewerID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的审核人ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.kycService.ApproveQualification(c.Request.Context(), qualID, reviewerID, req.ReviewerName, req.Comments)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "审批失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "审批通过"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RejectQualification 拒绝企业资质
// @Summary 拒绝企业资质
// @Tags Qualifications
// @Accept json
// @Produce json
// @Param id path string true "资质ID"
// @Param request body ReviewRequest true "拒绝请求"
// @Success 200 {object} map[string]interface{}
// @Router /qualifications/{id}/reject [post]
func (h *KYCHandler) RejectQualification(c *gin.Context) {
	idStr := c.Param("id")
	qualID, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的资质ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	reviewerID, err := uuid.Parse(req.ReviewerID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的审核人ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.kycService.RejectQualification(c.Request.Context(), qualID, reviewerID, req.ReviewerName, req.Reason)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "拒绝失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "已拒绝"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Level Handlers

// GetMerchantLevel 获取商户KYC级别
// @Summary 获取商户KYC级别
// @Tags Levels
// @Accept json
// @Produce json
// @Param merchant_id path string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /levels/{merchant_id} [get]
func (h *KYCHandler) GetMerchantLevel(c *gin.Context) {
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	level, err := h.kycService.GetMerchantLevel(c.Request.Context(), merchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取KYC级别失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(level).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CheckMerchantEligibility 检查商户资格
// @Summary 检查商户资格
// @Tags Levels
// @Accept json
// @Produce json
// @Param merchant_id path string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /levels/{merchant_id}/eligibility [get]
func (h *KYCHandler) CheckMerchantEligibility(c *gin.Context) {
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	eligibility, err := h.kycService.CheckMerchantEligibility(c.Request.Context(), merchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "检查资格失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(eligibility).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Alert Handlers

// ListAlerts 预警列表
// @Summary 预警列表
// @Tags Alerts
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Param alert_type query string false "预警类型"
// @Param severity query string false "严重程度"
// @Param status query string false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /alerts [get]
func (h *KYCHandler) ListAlerts(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		merchantID = &id
	}

	var alertType *string
	if alertTypeStr := c.Query("alert_type"); alertTypeStr != "" {
		alertType = &alertTypeStr
	}

	var severity *string
	if severityStr := c.Query("severity"); severityStr != "" {
		severity = &severityStr
	}

	var status *string
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
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

	query := &service.ListAlertQuery{
		MerchantID: merchantID,
		AlertType:  alertType,
		Severity:   severity,
		Status:     status,
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := h.kycService.ListAlerts(c.Request.Context(), query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取预警列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(result).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ResolveAlertRequest 处理预警请求
type ResolveAlertRequest struct {
	ResolverID string `json:"resolver_id" binding:"required"`
}

// ResolveAlert 处理预警
// @Summary 处理预警
// @Tags Alerts
// @Accept json
// @Produce json
// @Param id path string true "预警ID"
// @Param request body ResolveAlertRequest true "处理请求"
// @Success 200 {object} map[string]interface{}
// @Router /alerts/{id}/resolve [post]
func (h *KYCHandler) ResolveAlert(c *gin.Context) {
	idStr := c.Param("id")
	alertID, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的预警ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req ResolveAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resolverID, err := uuid.Parse(req.ResolverID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的处理人ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.kycService.ResolveAlert(c.Request.Context(), alertID, resolverID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "处理预警失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "预警已处理"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Statistics Handler

// GetKYCStatistics 获取KYC统计
// @Summary 获取KYC统计
// @Tags Statistics
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /statistics [get]
func (h *KYCHandler) GetKYCStatistics(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		merchantID = &id
	}

	stats, err := h.kycService.GetKYCStatistics(c.Request.Context(), merchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取KYC统计失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(stats).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

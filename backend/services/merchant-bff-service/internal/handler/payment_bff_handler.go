package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type PaymentBFFHandler struct {
	paymentClient *client.ServiceClient
}

func NewPaymentBFFHandler(paymentServiceURL string) *PaymentBFFHandler {
	return &PaymentBFFHandler{
		paymentClient: client.NewServiceClient(paymentServiceURL),
	}
}

func (h *PaymentBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/merchant/payments")
	merchant.Use(authMiddleware)
	{
		merchant.GET("", h.ListPayments)
		merchant.GET("/:payment_no", h.GetPayment)
		merchant.POST("/:payment_no/refund", h.CreateRefund)
		merchant.GET("/statistics", h.GetStatistics)
	}
}

func (h *PaymentBFFHandler) ListPayments(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"page":        c.DefaultQuery("page", "1"),
		"page_size":   c.DefaultQuery("page_size", "10"),
		"status":      c.Query("status"),
		"channel":     c.Query("channel"),
		"start_time":  c.Query("start_time"),
		"end_time":    c.Query("end_time"),
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/payments", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetPayment(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	paymentNo := c.Param("payment_no")
	if paymentNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "支付单号不能为空"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/payments/"+paymentNo, queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) CreateRefund(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	paymentNo := c.Param("payment_no")
	if paymentNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "支付单号不能为空"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	requestBody["merchant_id"] = merchantID
	requestBody["payment_no"] = paymentNo

	result, statusCode, err := h.paymentClient.Post(c.Request.Context(), "/api/v1/refunds", requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetStatistics(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"start_time":  c.Query("start_time"),
		"end_time":    c.Query("end_time"),
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/payments/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

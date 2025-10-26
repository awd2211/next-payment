package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type OrderBFFHandler struct {
	orderClient *client.ServiceClient
}

func NewOrderBFFHandler(orderServiceURL string) *OrderBFFHandler {
	return &OrderBFFHandler{
		orderClient: client.NewServiceClient(orderServiceURL),
	}
}

func (h *OrderBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/merchant/orders")
	merchant.Use(authMiddleware)
	{
		merchant.GET("", h.ListOrders)
		merchant.GET("/:order_no", h.GetOrder)
		merchant.GET("/statistics", h.GetStatistics)
	}
}

func (h *OrderBFFHandler) ListOrders(c *gin.Context) {
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
		"start_time":  c.Query("start_time"),
		"end_time":    c.Query("end_time"),
	}

	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *OrderBFFHandler) GetOrder(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	orderNo := c.Param("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单号不能为空"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/"+orderNo, queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *OrderBFFHandler) GetStatistics(c *gin.Context) {
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

	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

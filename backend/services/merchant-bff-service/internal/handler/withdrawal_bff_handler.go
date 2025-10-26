package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type WithdrawalBFFHandler struct {
	withdrawalClient *client.ServiceClient
}

func NewWithdrawalBFFHandler(withdrawalServiceURL string) *WithdrawalBFFHandler {
	return &WithdrawalBFFHandler{
		withdrawalClient: client.NewServiceClient(withdrawalServiceURL),
	}
}

func (h *WithdrawalBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/merchant/withdrawals")
	merchant.Use(authMiddleware)
	{
		merchant.POST("", h.CreateWithdrawal)
		merchant.GET("", h.ListWithdrawals)
		merchant.GET("/:withdrawal_no", h.GetWithdrawal)
		merchant.POST("/:withdrawal_no/cancel", h.CancelWithdrawal)
	}
}

func (h *WithdrawalBFFHandler) CreateWithdrawal(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	requestBody["merchant_id"] = merchantID

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/withdrawals", requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) ListWithdrawals(c *gin.Context) {
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

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/withdrawals", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) GetWithdrawal(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	withdrawalNo := c.Param("withdrawal_no")
	if withdrawalNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "提现单号不能为空"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo, queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) CancelWithdrawal(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	withdrawalNo := c.Param("withdrawal_no")
	if withdrawalNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "提现单号不能为空"})
		return
	}

	requestBody := map[string]interface{}{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo+"/cancel", requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

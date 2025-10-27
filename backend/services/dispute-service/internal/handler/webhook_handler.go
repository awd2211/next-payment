package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/logger"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
	"go.uber.org/zap"

	"payment-platform/dispute-service/internal/service"
)

// WebhookHandler Stripe webhook处理器
type WebhookHandler struct {
	disputeService service.DisputeService
	webhookSecret  string
}

// NewWebhookHandler 创建webhook处理器
func NewWebhookHandler(disputeService service.DisputeService, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		disputeService: disputeService,
		webhookSecret:  webhookSecret,
	}
}

// RegisterRoutes 注册webhook路由
func (h *WebhookHandler) RegisterRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/stripe/disputes", h.HandleStripeDispute)
	}
}

// HandleStripeDispute 处理Stripe dispute webhook
// @Summary 接收Stripe争议webhook
// @Description 自动接收Stripe dispute.created, dispute.updated等事件
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param Stripe-Signature header string true "Stripe签名"
// @Success 200 {object} map[string]string
// @Router /webhooks/stripe/disputes [post]
func (h *WebhookHandler) HandleStripeDispute(c *gin.Context) {
	const MaxBodyBytes = int64(65536) // 64KB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	// 验证webhook签名
	signature := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, signature, h.webhookSecret)
	if err != nil {
		logger.Error("Failed to verify webhook signature", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	logger.Info("Received Stripe webhook",
		zap.String("event_type", string(event.Type)),
		zap.String("event_id", event.ID))

	// 处理不同的事件类型
	switch event.Type {
	case "charge.dispute.created":
		h.handleDisputeCreated(c, event)
	case "charge.dispute.updated":
		h.handleDisputeUpdated(c, event)
	case "charge.dispute.closed":
		h.handleDisputeClosed(c, event)
	default:
		logger.Info("Unhandled webhook event type",
			zap.String("event_type", string(event.Type)))
		c.JSON(http.StatusOK, gin.H{"received": true, "handled": false})
		return
	}
}

// handleDisputeCreated 处理争议创建事件
func (h *WebhookHandler) handleDisputeCreated(c *gin.Context, event stripe.Event) {
	// 提取dispute ID从event data
	disputeID := event.GetObjectValue("id")
	if disputeID == "" {
		logger.Error("Missing dispute ID in webhook event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing dispute ID"})
		return
	}
	logger.Info("Processing dispute.created event", zap.String("dispute_id", disputeID))

	// 调用SyncFromStripe同步争议数据
	dispute, err := h.disputeService.SyncFromStripe(c.Request.Context(), disputeID)
	if err != nil {
		logger.Error("Failed to sync dispute from Stripe",
			zap.String("dispute_id", disputeID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync dispute"})
		return
	}

	logger.Info("Successfully synced dispute from Stripe",
		zap.String("dispute_id", disputeID),
		zap.String("local_dispute_id", dispute.ID.String()))

	c.JSON(http.StatusOK, gin.H{
		"received":         true,
		"handled":          true,
		"dispute_id":       disputeID,
		"local_dispute_id": dispute.ID.String(),
	})
}

// handleDisputeUpdated 处理争议更新事件
func (h *WebhookHandler) handleDisputeUpdated(c *gin.Context, event stripe.Event) {
	// 提取dispute ID
	disputeID := event.GetObjectValue("id")
	if disputeID == "" {
		logger.Error("Missing dispute ID in webhook event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing dispute ID"})
		return
	}
	logger.Info("Processing dispute.updated event", zap.String("dispute_id", disputeID))

	// 重新同步以更新本地状态
	dispute, err := h.disputeService.SyncFromStripe(c.Request.Context(), disputeID)
	if err != nil {
		logger.Error("Failed to sync updated dispute",
			zap.String("dispute_id", disputeID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync dispute"})
		return
	}

	logger.Info("Successfully synced updated dispute",
		zap.String("dispute_id", disputeID),
		zap.String("status", dispute.Status))

	c.JSON(http.StatusOK, gin.H{
		"received": true,
		"handled":  true,
		"status":   dispute.Status,
	})
}

// handleDisputeClosed 处理争议关闭事件
func (h *WebhookHandler) handleDisputeClosed(c *gin.Context, event stripe.Event) {
	// 提取dispute ID和结果
	disputeID := event.GetObjectValue("id")
	status := event.GetObjectValue("status")
	if disputeID == "" {
		logger.Error("Missing dispute ID in webhook event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing dispute ID"})
		return
	}

	logger.Info("Processing dispute.closed event",
		zap.String("dispute_id", disputeID),
		zap.String("status", status))

	// 同步最终状态
	dispute, err := h.disputeService.SyncFromStripe(c.Request.Context(), disputeID)
	if err != nil {
		logger.Error("Failed to sync closed dispute",
			zap.String("dispute_id", disputeID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync dispute"})
		return
	}

	logger.Info("Successfully synced closed dispute",
		zap.String("dispute_id", disputeID),
		zap.String("final_status", dispute.Status))

	c.JSON(http.StatusOK, gin.H{
		"received":     true,
		"handled":      true,
		"final_status": dispute.Status,
	})
}

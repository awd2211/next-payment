package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"payment-platform/notification-service/internal/model"
	"payment-platform/notification-service/internal/provider"
	"payment-platform/notification-service/internal/repository"
)

// WebhookRetryWorker Webhook重试工作器
type WebhookRetryWorker struct {
	repo            repository.NotificationRepository
	webhookProvider *provider.WebhookProvider
	logger          *zap.Logger
	ticker          *time.Ticker
	stopChan        chan struct{}
}

// NewWebhookRetryWorker 创建Webhook重试工作器
func NewWebhookRetryWorker(
	repo repository.NotificationRepository,
	webhookProvider *provider.WebhookProvider,
	logger *zap.Logger,
) *WebhookRetryWorker {
	return &WebhookRetryWorker{
		repo:            repo,
		webhookProvider: webhookProvider,
		logger:          logger,
		ticker:          time.NewTicker(5 * time.Minute), // 每5分钟检查一次
		stopChan:        make(chan struct{}),
	}
}

// Start 启动重试工作器
func (w *WebhookRetryWorker) Start(ctx context.Context) {
	w.logger.Info("Starting webhook retry worker")

	go func() {
		// 立即执行一次
		w.processRetries(ctx)

		for {
			select {
			case <-w.ticker.C:
				w.processRetries(ctx)
			case <-w.stopChan:
				w.logger.Info("Webhook retry worker stopped")
				return
			}
		}
	}()
}

// Stop 停止重试工作器
func (w *WebhookRetryWorker) Stop() {
	w.ticker.Stop()
	close(w.stopChan)
}

// processRetries 处理需要重试的Webhook
func (w *WebhookRetryWorker) processRetries(ctx context.Context) {
	w.logger.Info("Processing webhook retries")

	// 查询需要重试的Webhook通知（状态为failed，且重试次数 < 3）
	// 注意：这需要在repository中添加查询方法
	notifications, err := w.findFailedWebhooks(ctx)
	if err != nil {
		w.logger.Error("Failed to find failed webhooks", zap.Error(err))
		return
	}

	if len(notifications) == 0 {
		w.logger.Debug("No webhooks need retry")
		return
	}

	w.logger.Info("Found webhooks to retry", zap.Int("count", len(notifications)))

	for _, notification := range notifications {
		// 检查是否超过最大重试次数
		if notification.RetryCount >= 3 {
			w.logger.Warn("Webhook exceeded max retries",
				zap.String("notification_id", notification.ID.String()),
				zap.Int("retry_count", notification.RetryCount))
			continue
		}

		// 执行重试
		w.retryWebhook(ctx, notification)
	}
}

// findFailedWebhooks 查询失败的Webhook通知
func (w *WebhookRetryWorker) findFailedWebhooks(ctx context.Context) ([]*model.Notification, error) {
	// 查询条件：
	// 1. Status = "failed"
	// 2. Channel = "webhook"
	// 3. RetryCount < 3
	// 4. UpdatedAt > 现在 - 24小时（避免重试过期的通知）

	// 简化实现：手动查询失败的webhook通知
	// 生产环境应该在repository中添加专门的查询方法
	return w.repo.GetFailedNotifications(ctx, "webhook", 3)
}

// retryWebhook 重试单个Webhook
func (w *WebhookRetryWorker) retryWebhook(ctx context.Context, notification *model.Notification) {
	w.logger.Info("Retrying webhook",
		zap.String("notification_id", notification.ID.String()),
		zap.Int("retry_count", notification.RetryCount))

	// 解析Webhook配置
	var webhookConfig struct {
		URL    string                 `json:"url"`
		Secret string                 `json:"secret"`
		Data   map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal([]byte(notification.Content), &webhookConfig); err != nil {
		w.logger.Error("Failed to parse webhook config",
			zap.String("notification_id", notification.ID.String()),
			zap.Error(err))
		return
	}

	// 发送Webhook
	req := &provider.WebhookRequest{
		URL:       webhookConfig.URL,
		Secret:    webhookConfig.Secret,
		EventType: "webhook_retry",
		EventID:   notification.ID.String(),
		Data:      webhookConfig.Data,
		Timestamp: time.Now().Unix(),
	}

	response, err := w.webhookProvider.Send(ctx, req)
	if err != nil {
		w.logger.Error("Failed to send webhook",
			zap.String("notification_id", notification.ID.String()),
			zap.Error(err))
		return
	}

	// 更新通知状态
	notification.RetryCount++

	if response.Status == "delivered" {
		notification.Status = model.StatusSent
		now := time.Now()
		notification.SentAt = &now
		notification.ErrorMessage = ""

		w.logger.Info("Webhook retry succeeded",
			zap.String("notification_id", notification.ID.String()),
			zap.Int("retry_count", notification.RetryCount))
	} else {
		notification.Status = model.StatusFailed
		notification.ErrorMessage = response.ErrorMessage

		w.logger.Warn("Webhook retry failed",
			zap.String("notification_id", notification.ID.String()),
			zap.Int("retry_count", notification.RetryCount),
			zap.String("error", response.ErrorMessage))
	}

	// 保存更新
	if err := w.repo.Update(ctx, notification); err != nil {
		w.logger.Error("Failed to update notification",
			zap.String("notification_id", notification.ID.String()),
			zap.Error(err))
	}
}

// RetryWebhookByID 手动重试指定的Webhook（供API调用）
func (w *WebhookRetryWorker) RetryWebhookByID(ctx context.Context, notificationID uuid.UUID) error {
	// 查询通知
	notification, err := w.repo.GetByID(ctx, notificationID)
	if err != nil {
		return err
	}

	// 检查是否为Webhook通知
	if notification.Channel != "webhook" {
		w.logger.Warn("Notification is not webhook type",
			zap.String("notification_id", notificationID.String()),
			zap.String("channel", notification.Channel))
		return nil
	}

	// 执行重试
	w.retryWebhook(ctx, notification)

	return nil
}

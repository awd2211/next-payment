package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/webhook"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
)

// WebhookNotificationService Webhook 通知服务接口
type WebhookNotificationService interface {
	// 发送 Webhook 通知
	SendPaymentNotification(ctx context.Context, payment *model.Payment, event string, notifyURL string, secret string) error

	// 重试失败的通知
	RetryFailedNotifications(ctx context.Context) (int, error)

	// 查询通知记录
	GetNotificationsByPayment(ctx context.Context, merchantID uuid.UUID, paymentNo string) ([]*model.WebhookNotification, error)
	GetFailedNotifications(ctx context.Context, merchantID uuid.UUID, limit int, offset int) ([]*model.WebhookNotification, int64, error)
}

type webhookNotificationService struct {
	repo                 repository.WebhookNotificationRepository
	retrier              *webhook.WebhookRetrier
	merchantConfigClient client.MerchantConfigClient
}

// NewWebhookNotificationService 创建 Webhook 通知服务
func NewWebhookNotificationService(
	repo repository.WebhookNotificationRepository,
	redisClient *redis.Client,
	merchantConfigClient client.MerchantConfigClient,
) WebhookNotificationService {
	return &webhookNotificationService{
		repo:                 repo,
		retrier:              webhook.NewWebhookRetrier(webhook.DefaultRetryConfig(), redisClient),
		merchantConfigClient: merchantConfigClient,
	}
}

// SendPaymentNotification 发送支付通知
func (s *webhookNotificationService) SendPaymentNotification(
	ctx context.Context,
	payment *model.Payment,
	event string,
	notifyURL string,
	secret string,
) error {
	// 构造 payload
	payload := &webhook.WebhookPayload{
		Event:     event,
		PaymentNo: payment.PaymentNo,
		OrderNo:   payment.OrderNo,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Status:    payment.Status,
		Timestamp: time.Now().Unix(),
		Extra: map[string]interface{}{
			"channel":          payment.Channel,
			"channel_order_no": payment.ChannelOrderNo,
			"paid_at":          payment.PaidAt,
		},
	}

	// 序列化 payload
	payloadBytes, _ := json.Marshal(payload)

	// 创建通知记录
	notification := &model.WebhookNotification{
		MerchantID:  payment.MerchantID,
		PaymentNo:   payment.PaymentNo,
		OrderNo:     payment.OrderNo,
		Event:       event,
		URL:         notifyURL,
		Payload:     string(payloadBytes),
		Status:      model.WebhookStatusPending,
		Attempts:    0,
		MaxAttempts: 5,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("创建通知记录失败: %w", err)
	}

	// 异步发送（不阻塞主流程）
	go s.sendAsync(context.Background(), notification, payload, notifyURL, secret)

	return nil
}

// sendAsync 异步发送通知
func (s *webhookNotificationService) sendAsync(
	ctx context.Context,
	notification *model.WebhookNotification,
	payload *webhook.WebhookPayload,
	notifyURL string,
	secret string,
) {
	req := &webhook.WebhookRequest{
		URL:        notifyURL,
		Secret:     secret,
		Payload:    payload,
		MerchantID: notification.MerchantID.String(),
	}

	// 发送（带重试）
	resp, err := s.retrier.Send(ctx, req)

	// 更新通知记录
	notification.Attempts = resp.Attempt
	notification.StatusCode = resp.StatusCode
	notification.Response = resp.Body

	if err != nil {
		notification.Status = model.WebhookStatusFailed
		notification.Error = err.Error()
		now := time.Now()
		notification.FailedAt = &now

		logger.Error("Webhook 通知最终失败",
			zap.Error(err),
			zap.String("merchant_id", notification.MerchantID.String()),
			zap.String("payment_no", notification.PaymentNo),
			zap.Int("attempts", notification.Attempts))
	} else {
		notification.Status = model.WebhookStatusSuccess
		now := time.Now()
		notification.SucceededAt = &now

		logger.Info("Webhook 通知成功",
			zap.String("merchant_id", notification.MerchantID.String()),
			zap.String("payment_no", notification.PaymentNo),
			zap.Int("attempts", notification.Attempts))
	}

	// 保存更新
	if err := s.repo.Update(ctx, notification); err != nil {
		logger.Error("更新通知记录失败",
			zap.Error(err),
			zap.String("notification_id", notification.ID.String()))
	}
}

// RetryFailedNotifications 重试失败的通知（后台任务调用）
func (s *webhookNotificationService) RetryFailedNotifications(ctx context.Context) (int, error) {
	// 查询待重试的通知（最多 100 条）
	notifications, err := s.repo.GetPendingRetries(ctx, 100)
	if err != nil {
		return 0, fmt.Errorf("查询待重试通知失败: %w", err)
	}

	if len(notifications) == 0 {
		return 0, nil
	}

	logger.Info("开始重试失败的 Webhook 通知",
		zap.Int("count", len(notifications)))

	successCount := 0

	for _, notification := range notifications {
		// 解析 payload
		var payload webhook.WebhookPayload
		if err := json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
			logger.Error("解析 payload 失败",
				zap.Error(err),
				zap.String("notification_id", notification.ID.String()))
			continue
		}

		// 从merchant-config-service获取商户的webhook密钥
		secret, err := s.merchantConfigClient.GetWebhookSecret(ctx, notification.MerchantID)
		if err != nil {
			logger.Error("获取商户webhook密钥失败",
				zap.Error(err),
				zap.String("merchant_id", notification.MerchantID.String()),
				zap.String("notification_id", notification.ID.String()))
			// 获取密钥失败,标记为失败并继续下一个
			notification.Status = model.WebhookStatusFailed
			s.repo.Update(ctx, notification)
			continue
		}

		// 更新状态为重试中
		notification.Status = model.WebhookStatusRetrying
		notification.Attempts++
		if err := s.repo.Update(ctx, notification); err != nil {
			logger.Error("更新通知状态失败",
				zap.Error(err),
				zap.String("notification_id", notification.ID.String()))
			continue
		}

		// 发送请求
		req := &webhook.WebhookRequest{
			URL:        notification.URL,
			Secret:     secret,
			Payload:    &payload,
			MerchantID: notification.MerchantID.String(),
		}

		resp, err := s.retrier.Send(ctx, req)

		// 更新记录
		notification.Attempts = resp.Attempt
		notification.StatusCode = resp.StatusCode
		notification.Response = resp.Body

		if err != nil {
			// 检查是否已达最大重试次数
			if notification.Attempts >= notification.MaxAttempts {
				notification.Status = model.WebhookStatusFailed
				now := time.Now()
				notification.FailedAt = &now
				notification.Error = err.Error()

				logger.Warn("Webhook 通知已达最大重试次数",
					zap.String("merchant_id", notification.MerchantID.String()),
					zap.String("payment_no", notification.PaymentNo),
					zap.Int("attempts", notification.Attempts))
			} else {
				// 计算下次重试时间（指数退避）
				nextRetryAt := s.calculateNextRetry(notification.Attempts)
				notification.NextRetryAt = &nextRetryAt
				notification.Status = model.WebhookStatusPending
				notification.Error = err.Error()

				logger.Info("Webhook 通知重试失败，将稍后重试",
					zap.String("merchant_id", notification.MerchantID.String()),
					zap.String("payment_no", notification.PaymentNo),
					zap.Int("attempts", notification.Attempts),
					zap.Time("next_retry_at", nextRetryAt))
			}
		} else {
			notification.Status = model.WebhookStatusSuccess
			now := time.Now()
			notification.SucceededAt = &now
			successCount++

			logger.Info("Webhook 通知重试成功",
				zap.String("merchant_id", notification.MerchantID.String()),
				zap.String("payment_no", notification.PaymentNo),
				zap.Int("attempts", notification.Attempts))
		}

		// 保存更新
		if err := s.repo.Update(ctx, notification); err != nil {
			logger.Error("更新通知记录失败",
				zap.Error(err),
				zap.String("notification_id", notification.ID.String()))
		}
	}

	logger.Info("Webhook 通知重试完成",
		zap.Int("total", len(notifications)),
		zap.Int("success", successCount),
		zap.Int("failed", len(notifications)-successCount))

	return successCount, nil
}

// calculateNextRetry 计算下次重试时间（指数退避）
func (s *webhookNotificationService) calculateNextRetry(attempts int) time.Time {
	// 1秒, 2秒, 4秒, 8秒, 16秒, 32秒, 1分钟, 2分钟, 5分钟, 10分钟, 30分钟, 1小时
	delays := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		32 * time.Second,
		1 * time.Minute,
		2 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
	}

	index := attempts - 1
	if index >= len(delays) {
		index = len(delays) - 1
	}

	return time.Now().Add(delays[index])
}

// GetNotificationsByPayment 查询支付的通知记录
func (s *webhookNotificationService) GetNotificationsByPayment(
	ctx context.Context,
	merchantID uuid.UUID,
	paymentNo string,
) ([]*model.WebhookNotification, error) {
	return s.repo.GetByPaymentNo(ctx, merchantID, paymentNo)
}

// GetFailedNotifications 查询失败的通知
func (s *webhookNotificationService) GetFailedNotifications(
	ctx context.Context,
	merchantID uuid.UUID,
	limit int,
	offset int,
) ([]*model.WebhookNotification, int64, error) {
	return s.repo.GetFailedNotifications(ctx, merchantID, limit, offset)
}

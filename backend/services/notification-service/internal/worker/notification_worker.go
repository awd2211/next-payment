package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"payment-platform/notification-service/internal/model"
	"payment-platform/notification-service/internal/provider"
	"payment-platform/notification-service/internal/repository"
	"strings"
	"time"
)

// NotificationWorker 通知处理worker
type NotificationWorker struct {
	repo          repository.NotificationRepository
	emailFactory  *provider.EmailProviderFactory
	smsFactory    *provider.SMSProviderFactory
}

// NewNotificationWorker 创建通知worker
func NewNotificationWorker(
	repo repository.NotificationRepository,
	emailFactory *provider.EmailProviderFactory,
	smsFactory *provider.SMSProviderFactory,
) *NotificationWorker {
	return &NotificationWorker{
		repo:         repo,
		emailFactory: emailFactory,
		smsFactory:   smsFactory,
	}
}

// NotificationMessage Kafka消息结构
type NotificationMessage struct {
	NotificationID uuid.UUID `json:"notification_id"`
	Channel        string    `json:"channel"` // email, sms
}

// StartEmailWorker 启动邮件处理worker
func (w *NotificationWorker) StartEmailWorker(ctx context.Context, consumer *kafka.Consumer) {
	logger.Info("邮件Worker启动...")

	handler := func(ctx context.Context, message []byte) error {
		var msg NotificationMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Error(fmt.Sprintf("反序列化邮件消息失败: %v", err))
			return err
		}

		logger.Info(fmt.Sprintf("处理邮件通知: %s", msg.NotificationID))

		// 获取通知记录
		notification, err := w.repo.GetByID(ctx, msg.NotificationID)
		if err != nil {
			logger.Error(fmt.Sprintf("获取通知记录失败: %v", err))
			return err
		}

		if notification == nil {
			logger.Error(fmt.Sprintf("通知记录不存在: %s", msg.NotificationID))
			return fmt.Errorf("通知记录不存在")
		}

		// 检查状态
		if notification.Status != model.StatusPending {
			logger.Info(fmt.Sprintf("通知状态不是pending，跳过: %s (status=%s)", msg.NotificationID, notification.Status))
			return nil
		}

		// 更新状态为发送中
		w.repo.UpdateStatus(ctx, notification.ID, model.StatusSending)

		// 发送邮件
		err = w.sendEmail(ctx, notification)
		if err != nil {
			// 更新状态为失败
			notification.Status = model.StatusFailed
			notification.ErrorMessage = err.Error()
			notification.RetryCount++
			w.repo.Update(ctx, notification)
			logger.Error(fmt.Sprintf("发送邮件失败: %v", err))
			return err
		}

		// 更新状态为已发送
		notification.Status = model.StatusSent
		now := time.Now()
		notification.SentAt = &now
		w.repo.Update(ctx, notification)

		logger.Info(fmt.Sprintf("邮件发送成功: %s", msg.NotificationID))
		return nil
	}

	// 开始消费，支持重试
	if err := consumer.ConsumeWithRetry(ctx, handler, 3); err != nil {
		logger.Error(fmt.Sprintf("邮件Worker停止: %v", err))
	}
}

// StartSMSWorker 启动短信处理worker
func (w *NotificationWorker) StartSMSWorker(ctx context.Context, consumer *kafka.Consumer) {
	logger.Info("短信Worker启动...")

	handler := func(ctx context.Context, message []byte) error {
		var msg NotificationMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Error(fmt.Sprintf("反序列化短信消息失败: %v", err))
			return err
		}

		logger.Info(fmt.Sprintf("处理短信通知: %s", msg.NotificationID))

		// 获取通知记录
		notification, err := w.repo.GetByID(ctx, msg.NotificationID)
		if err != nil {
			logger.Error(fmt.Sprintf("获取通知记录失败: %v", err))
			return err
		}

		if notification == nil {
			logger.Error(fmt.Sprintf("通知记录不存在: %s", msg.NotificationID))
			return fmt.Errorf("通知记录不存在")
		}

		// 检查状态
		if notification.Status != model.StatusPending {
			logger.Info(fmt.Sprintf("通知状态不是pending，跳过: %s (status=%s)", msg.NotificationID, notification.Status))
			return nil
		}

		// 更新状态为发送中
		w.repo.UpdateStatus(ctx, notification.ID, model.StatusSending)

		// 发送短信
		err = w.sendSMS(ctx, notification)
		if err != nil {
			// 更新状态为失败
			notification.Status = model.StatusFailed
			notification.ErrorMessage = err.Error()
			notification.RetryCount++
			w.repo.Update(ctx, notification)
			logger.Error(fmt.Sprintf("发送短信失败: %v", err))
			return err
		}

		// 更新状态为已发送
		notification.Status = model.StatusSent
		now := time.Now()
		notification.SentAt = &now
		w.repo.Update(ctx, notification)

		logger.Info(fmt.Sprintf("短信发送成功: %s", msg.NotificationID))
		return nil
	}

	// 开始消费，支持重试
	if err := consumer.ConsumeWithRetry(ctx, handler, 3); err != nil {
		logger.Error(fmt.Sprintf("短信Worker停止: %v", err))
	}
}

// sendEmail 实际发送邮件
func (w *NotificationWorker) sendEmail(ctx context.Context, notification *model.Notification) error {
	// 获取邮件提供商
	emailProvider, ok := w.emailFactory.GetProvider(notification.Provider)
	if !ok {
		return fmt.Errorf("不支持的邮件提供商: %s", notification.Provider)
	}

	// 发送邮件
	to := strings.Split(notification.Recipient, ",")
	emailReq := &provider.EmailRequest{
		To:       to,
		Subject:  notification.Subject,
		HTMLBody: notification.Content,
	}

	resp, err := emailProvider.Send(ctx, emailReq)
	if err != nil {
		return err
	}

	// 更新消息ID
	notification.ProviderMsgID = resp.MessageID
	return nil
}

// sendSMS 实际发送短信
func (w *NotificationWorker) sendSMS(ctx context.Context, notification *model.Notification) error {
	// 获取短信提供商
	smsProvider, ok := w.smsFactory.GetProvider(notification.Provider)
	if !ok {
		return fmt.Errorf("不支持的短信提供商: %s", notification.Provider)
	}

	// 发送短信
	smsReq := &provider.SMSRequest{
		To:      notification.Recipient,
		Content: notification.Content,
	}

	resp, err := smsProvider.Send(ctx, smsReq)
	if err != nil {
		return err
	}

	// 更新消息ID
	notification.ProviderMsgID = resp.MessageID
	return nil
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
)

// TimeoutService 超时处理服务
type TimeoutService struct {
	db                 *gorm.DB
	paymentRepo        repository.PaymentRepository
	orderClient        *client.OrderClient
	channelClient      *client.ChannelClient
	notificationClient *client.NotificationClient
}

// NewTimeoutService 创建超时处理服务
func NewTimeoutService(
	db *gorm.DB,
	paymentRepo repository.PaymentRepository,
	orderClient *client.OrderClient,
	channelClient *client.ChannelClient,
	notificationClient *client.NotificationClient,
) *TimeoutService {
	return &TimeoutService{
		db:                 db,
		paymentRepo:        paymentRepo,
		orderClient:        orderClient,
		channelClient:      channelClient,
		notificationClient: notificationClient,
	}
}

// ScanExpiredPayments 扫描并处理过期支付
func (s *TimeoutService) ScanExpiredPayments(ctx context.Context) error {
	logger.Info("开始扫描过期支付...")

	// 查询所有过期且仍处于pending状态的支付
	var expiredPayments []model.Payment
	now := time.Now()

	err := s.db.WithContext(ctx).
		Where("status = ? AND expired_at < ? AND expired_at IS NOT NULL", "pending", now).
		Limit(100). // 每次处理100条,避免一次性处理过多
		Find(&expiredPayments).Error

	if err != nil {
		logger.Error("查询过期支付失败", zap.Error(err))
		return fmt.Errorf("查询过期支付失败: %w", err)
	}

	if len(expiredPayments) == 0 {
		logger.Info("没有发现过期支付")
		return nil
	}

	logger.Info(fmt.Sprintf("发现 %d 笔过期支付，开始处理", len(expiredPayments)))

	successCount := 0
	failedCount := 0

	// 逐个处理过期支付
	for _, payment := range expiredPayments {
		if err := s.processExpiredPayment(ctx, &payment); err != nil {
			logger.Error("处理过期支付失败",
				zap.String("payment_no", payment.PaymentNo),
				zap.Error(err))
			failedCount++
		} else {
			successCount++
		}
	}

	logger.Info("过期支付处理完成",
		zap.Int("total", len(expiredPayments)),
		zap.Int("success", successCount),
		zap.Int("failed", failedCount))

	return nil
}

// processExpiredPayment 处理单个过期支付
func (s *TimeoutService) processExpiredPayment(ctx context.Context, payment *model.Payment) error {
	logger.Info("处理过期支付",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("order_no", payment.OrderNo),
		zap.Time("expired_at", *payment.ExpiredAt))

	// 使用数据库事务确保原子性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 尝试向渠道发起取消请求（如果有渠道订单号）
		if payment.ChannelOrderNo != "" {
			cancelErr := s.cancelChannelPayment(ctx, payment)
			if cancelErr != nil {
				logger.Warn("取消渠道支付失败",
					zap.String("payment_no", payment.PaymentNo),
					zap.String("channel_order_no", payment.ChannelOrderNo),
					zap.Error(cancelErr))
				// 继续执行,即使渠道取消失败也要更新本地状态
			}
		}

		// 2. 更新支付状态为expired
		payment.Status = "expired"
		payment.ErrorCode = "PAYMENT_EXPIRED"
		payment.ErrorMsg = "支付已超时"
		payment.UpdatedAt = time.Now()

		if err := tx.Save(payment).Error; err != nil {
			return fmt.Errorf("更新支付状态失败: %w", err)
		}

		// 3. 更新订单状态
		if s.orderClient != nil {
			updateOrderReq := &client.UpdateOrderStatusRequest{
				Status:    "expired",
				ErrorCode: "PAYMENT_EXPIRED",
				ErrorMsg:  "支付超时自动取消",
			}

			if err := s.orderClient.UpdateOrderStatus(ctx, payment.PaymentNo, updateOrderReq); err != nil {
				logger.Warn("更新订单状态失败",
					zap.String("order_no", payment.OrderNo),
					zap.Error(err))
				// 继续执行,不因订单更新失败而回滚
			}
		}

		// 4. 发送超时通知给商户
		s.sendTimeoutNotification(ctx, payment)

		logger.Info("过期支付处理成功",
			zap.String("payment_no", payment.PaymentNo),
			zap.String("new_status", payment.Status))

		return nil
	})
}

// cancelChannelPayment 向支付渠道发起取消请求
func (s *TimeoutService) cancelChannelPayment(ctx context.Context, payment *model.Payment) error {
	if s.channelClient == nil {
		return fmt.Errorf("渠道客户端未初始化")
	}

	// 渠道取消支付通常需要调用渠道适配器的专用API
	// 这里简化处理：记录日志，实际可以调用channel-adapter的取消接口
	logger.Warn("渠道支付取消暂未实现",
		zap.String("channel", payment.Channel),
		zap.String("channel_order_no", payment.ChannelOrderNo))

	return nil
}

// sendTimeoutNotification 发送超时通知给商户
func (s *TimeoutService) sendTimeoutNotification(ctx context.Context, payment *model.Payment) {
	if s.notificationClient == nil {
		return
	}

	// 异步发送通知,不阻塞主流程
	go func() {
		notifyCtx := context.Background()

		notifyReq := &client.SendNotificationRequest{
			MerchantID: payment.MerchantID,
			Type:       "payment_expired",
			Title:      "支付超时通知",
			Content:    fmt.Sprintf("支付单 %s 已超时", payment.PaymentNo),
			Data: map[string]interface{}{
				"payment_no": payment.PaymentNo,
				"order_no":   payment.OrderNo,
				"amount":     payment.Amount,
				"currency":   payment.Currency,
				"status":     "expired",
				"expired_at": payment.ExpiredAt,
			},
			Priority: "high",
		}

		if err := s.notificationClient.SendPaymentNotification(notifyCtx, notifyReq); err != nil {
			logger.Error("发送超时通知失败",
				zap.String("payment_no", payment.PaymentNo),
				zap.Error(err))
		} else {
			logger.Info("超时通知发送成功", zap.String("payment_no", payment.PaymentNo))
		}
	}()
}

// TimeoutWorker 超时扫描工作器
type TimeoutWorker struct {
	timeoutService *TimeoutService
	interval       time.Duration
	stopCh         chan struct{}
}

// NewTimeoutWorker 创建超时扫描工作器
func NewTimeoutWorker(timeoutService *TimeoutService, interval time.Duration) *TimeoutWorker {
	return &TimeoutWorker{
		timeoutService: timeoutService,
		interval:       interval,
		stopCh:         make(chan struct{}),
	}
}

// Start 启动超时扫描工作器
func (w *TimeoutWorker) Start(ctx context.Context) {
	logger.Info("超时扫描工作器已启动", zap.Duration("interval", w.interval))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.timeoutService.ScanExpiredPayments(ctx); err != nil {
				logger.Error("超时扫描失败", zap.Error(err))
			}

		case <-w.stopCh:
			logger.Info("超时扫描工作器已停止")
			return

		case <-ctx.Done():
			logger.Info("超时扫描工作器收到退出信号")
			return
		}
	}
}

// Stop 停止超时扫描工作器
func (w *TimeoutWorker) Stop() {
	close(w.stopCh)
}

// CancelExpiredPayment 手动取消过期支付（提供给handler调用）
func (s *TimeoutService) CancelExpiredPayment(ctx context.Context, paymentNo string, merchantID uuid.UUID) error {
	// 查询支付记录
	payment, err := s.paymentRepo.GetByPaymentNo(ctx, paymentNo)
	if err != nil {
		return fmt.Errorf("查询支付记录失败: %w", err)
	}

	if payment == nil {
		return fmt.Errorf("支付记录不存在")
	}

	// 验证商户ID
	if payment.MerchantID != merchantID {
		return fmt.Errorf("无权操作此支付记录")
	}

	// 检查支付状态
	if payment.Status != "pending" {
		return fmt.Errorf("支付状态不允许取消，当前状态: %s", payment.Status)
	}

	// 检查是否已过期
	if payment.ExpiredAt != nil && time.Now().After(*payment.ExpiredAt) {
		// 已过期，直接处理
		return s.processExpiredPayment(ctx, payment)
	}

	return fmt.Errorf("支付尚未过期")
}

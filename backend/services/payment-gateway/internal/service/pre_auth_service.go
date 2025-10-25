package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
)

// PreAuthService 预授权服务接口
type PreAuthService interface {
	// 创建预授权
	CreatePreAuth(ctx context.Context, input *CreatePreAuthInput) (*model.PreAuthPayment, error)

	// 确认预授权（扣款）
	CapturePreAuth(ctx context.Context, merchantID uuid.UUID, preAuthNo string, amount *int64) (*model.Payment, error)

	// 取消预授权
	CancelPreAuth(ctx context.Context, merchantID uuid.UUID, preAuthNo string, reason string) error

	// 查询预授权
	GetPreAuth(ctx context.Context, merchantID uuid.UUID, preAuthNo string) (*model.PreAuthPayment, error)
	ListPreAuths(ctx context.Context, merchantID uuid.UUID, status string, page, pageSize int) ([]*model.PreAuthPayment, error)

	// 定时任务：扫描并过期超时的预授权
	ScanAndExpirePreAuths(ctx context.Context) (int, error)
}

// CreatePreAuthInput 创建预授权输入
type CreatePreAuthInput struct {
	MerchantID uuid.UUID
	OrderNo    string
	Amount     int64
	Currency   string
	Channel    string
	Subject    string
	Body       string
	ClientIP   string
	ReturnURL  string
	NotifyURL  string
	ExpiresIn  time.Duration // 预授权有效期，默认7天
}

// preAuthService 预授权服务实现
type preAuthService struct {
	db              *gorm.DB
	preAuthRepo     repository.PreAuthRepository
	paymentRepo     repository.PaymentRepository
	orderClient     *client.OrderClient
	channelClient   *client.ChannelClient
	riskClient      *client.RiskClient
	paymentService  PaymentService
	redisClient     *redis.Client
}

// NewPreAuthService 创建预授权服务
func NewPreAuthService(
	db *gorm.DB,
	preAuthRepo repository.PreAuthRepository,
	paymentRepo repository.PaymentRepository,
	orderClient *client.OrderClient,
	channelClient *client.ChannelClient,
	riskClient *client.RiskClient,
	paymentService PaymentService,
	redisClient *redis.Client,
) PreAuthService {
	return &preAuthService{
		db:             db,
		preAuthRepo:    preAuthRepo,
		paymentRepo:    paymentRepo,
		orderClient:    orderClient,
		channelClient:  channelClient,
		riskClient:     riskClient,
		paymentService: paymentService,
		redisClient:    redisClient,
	}
}

// CreatePreAuth 创建预授权
func (s *preAuthService) CreatePreAuth(ctx context.Context, input *CreatePreAuthInput) (*model.PreAuthPayment, error) {
	// 1. 检查订单是否已存在预授权
	existing, err := s.preAuthRepo.GetByOrderNo(ctx, input.MerchantID, input.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("查询订单预授权失败: %w", err)
	}
	if existing != nil {
		return existing, nil // 幂等性：返回已存在的预授权
	}

	// 2. 生成预授权单号
	preAuthNo := generatePreAuthNo()

	// 3. 设置过期时间（默认7天）
	expiresIn := input.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 7 * 24 * time.Hour
	}
	expiresAt := time.Now().Add(expiresIn)

	// 4. 风控检查
	if s.riskClient != nil {
		riskResult, err := s.riskClient.CheckRisk(ctx, &client.RiskCheckRequest{
			MerchantID: input.MerchantID,
			PaymentNo:  preAuthNo,
			Amount:     input.Amount,
			Currency:   input.Currency,
			Channel:    input.Channel,
			CustomerIP: input.ClientIP,
		})

		if err != nil {
			logger.Warn("风控检查失败，继续处理", zap.Error(err))
		} else if riskResult != nil {
			if riskResult.RiskLevel == "high" || riskResult.Decision == "reject" {
				reasons := ""
				if len(riskResult.Reasons) > 0 {
					reasons = riskResult.Reasons[0]
				}
				return nil, fmt.Errorf("风控拒绝: %s", reasons)
			}
		}
	}

	// 5. 调用渠道适配器创建预授权
	channelResp, err := s.channelClient.CreatePreAuth(ctx, &client.CreatePreAuthRequest{
		MerchantID: input.MerchantID.String(),
		OrderNo:    input.OrderNo,
		PreAuthNo:  preAuthNo,
		Amount:     input.Amount,
		Currency:   input.Currency,
		Channel:    input.Channel,
		Subject:    input.Subject,
		Body:       input.Body,
		ReturnURL:  input.ReturnURL,
		NotifyURL:  input.NotifyURL,
	})

	if err != nil {
		return nil, fmt.Errorf("调用渠道创建预授权失败: %w", err)
	}

	if channelResp.Code != 0 {
		return nil, fmt.Errorf("渠道返回错误: %s", channelResp.Message)
	}

	// 6. 创建预授权记录
	preAuth := &model.PreAuthPayment{
		MerchantID:     input.MerchantID,
		OrderNo:        input.OrderNo,
		PreAuthNo:      preAuthNo,
		Amount:         input.Amount,
		CapturedAmount: 0,
		Currency:       input.Currency,
		Channel:        input.Channel,
		ChannelTradeNo: channelResp.Data.ChannelTradeNo,
		Status:         model.PreAuthStatusPending,
		ExpiresAt:      expiresAt,
		Subject:        input.Subject,
		Body:           input.Body,
		ClientIP:       input.ClientIP,
		ReturnURL:      input.ReturnURL,
		NotifyURL:      input.NotifyURL,
	}

	err = s.preAuthRepo.Create(ctx, preAuth)
	if err != nil {
		return nil, fmt.Errorf("创建预授权记录失败: %w", err)
	}

	logger.Info("预授权创建成功",
		zap.String("pre_auth_no", preAuthNo),
		zap.String("order_no", input.OrderNo),
		zap.Int64("amount", input.Amount))

	return preAuth, nil
}

// CapturePreAuth 确认预授权（扣款）
func (s *preAuthService) CapturePreAuth(ctx context.Context, merchantID uuid.UUID, preAuthNo string, amount *int64) (*model.Payment, error) {
	// 1. 查询预授权记录
	preAuth, err := s.preAuthRepo.GetByPreAuthNo(ctx, merchantID, preAuthNo)
	if err != nil {
		return nil, fmt.Errorf("查询预授权失败: %w", err)
	}
	if preAuth == nil {
		return nil, fmt.Errorf("预授权不存在")
	}

	// 2. 检查状态
	if !preAuth.CanCapture() {
		return nil, fmt.Errorf("预授权状态不允许确认: status=%s, expired=%v", preAuth.Status, preAuth.IsExpired())
	}

	// 3. 确定确认金额
	captureAmount := preAuth.Amount
	if amount != nil {
		captureAmount = *amount
		// 检查金额是否超过剩余可确认金额
		if captureAmount > preAuth.GetRemainingAmount() {
			return nil, fmt.Errorf("确认金额超过剩余可确认金额: requested=%d, remaining=%d",
				captureAmount, preAuth.GetRemainingAmount())
		}
	}

	// 4. 调用渠道适配器确认预授权
	channelResp, err := s.channelClient.CapturePreAuth(ctx, &client.CapturePreAuthRequest{
		PreAuthNo:      preAuthNo,
		ChannelTradeNo: preAuth.ChannelTradeNo,
		Amount:         captureAmount,
		Currency:       preAuth.Currency,
	})

	if err != nil {
		return nil, fmt.Errorf("调用渠道确认预授权失败: %w", err)
	}

	if channelResp.Code != 0 {
		return nil, fmt.Errorf("渠道返回错误: %s", channelResp.Message)
	}

	// 5. 生成支付单号
	paymentNo := generatePaymentNo()

	// 6. 开始事务：创建支付记录 + 更新预授权记录
	var payment *model.Payment
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建支付记录
		payment = &model.Payment{
			MerchantID:     merchantID,
			OrderNo:        preAuth.OrderNo,
			PaymentNo:      paymentNo,
			Amount:         captureAmount,
			Currency:       preAuth.Currency,
			Channel:        preAuth.Channel,
			ChannelOrderNo: channelResp.Data.PaymentTradeNo,
			Status:         model.PaymentStatusSuccess,
			Description:    fmt.Sprintf("%s (预授权确认)", preAuth.Subject),
			CustomerIP:     preAuth.ClientIP,
			ReturnURL:      preAuth.ReturnURL,
			NotifyURL:      preAuth.NotifyURL,
			PaidAt:         timePtr(time.Now()),
			Extra:          fmt.Sprintf(`{"pre_auth_no": "%s", "type": "pre_auth_capture"}`, preAuthNo),
		}

		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("创建支付记录失败: %w", err)
		}

		// 更新预授权记录
		if err := s.preAuthRepo.UpdateToCaptured(ctx, preAuth.ID, captureAmount, paymentNo, time.Now()); err != nil {
			return fmt.Errorf("更新预授权记录失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 7. 通知订单服务
	if s.orderClient != nil {
		go func() {
			_ = s.orderClient.UpdateOrderStatus(context.Background(), paymentNo, &client.UpdateOrderStatusRequest{
				Status:         "paid",
				ChannelOrderNo: channelResp.Data.PaymentTradeNo,
				PaidAt:         time.Now().Format(time.RFC3339),
			})
		}()
	}

	logger.Info("预授权确认成功",
		zap.String("pre_auth_no", preAuthNo),
		zap.String("payment_no", paymentNo),
		zap.Int64("amount", captureAmount))

	return payment, nil
}

// CancelPreAuth 取消预授权
func (s *preAuthService) CancelPreAuth(ctx context.Context, merchantID uuid.UUID, preAuthNo string, reason string) error {
	// 1. 查询预授权记录
	preAuth, err := s.preAuthRepo.GetByPreAuthNo(ctx, merchantID, preAuthNo)
	if err != nil {
		return fmt.Errorf("查询预授权失败: %w", err)
	}
	if preAuth == nil {
		return fmt.Errorf("预授权不存在")
	}

	// 2. 检查状态
	if !preAuth.CanCancel() {
		return fmt.Errorf("预授权状态不允许取消: status=%s", preAuth.Status)
	}

	// 3. 调用渠道适配器取消预授权
	if preAuth.ChannelTradeNo != "" {
		channelResp, err := s.channelClient.CancelPreAuth(ctx, &client.CancelPreAuthRequest{
			PreAuthNo:      preAuthNo,
			ChannelTradeNo: preAuth.ChannelTradeNo,
		})

		if err != nil {
			logger.Warn("调用渠道取消预授权失败，继续更新本地状态", zap.Error(err))
		} else if channelResp.Code != 0 {
			logger.Warn("渠道返回取消失败，继续更新本地状态", zap.String("message", channelResp.Message))
		}
	}

	// 4. 更新预授权记录
	err = s.preAuthRepo.UpdateToCancelled(ctx, preAuth.ID, time.Now(), reason)
	if err != nil {
		return fmt.Errorf("更新预授权记录失败: %w", err)
	}

	logger.Info("预授权取消成功",
		zap.String("pre_auth_no", preAuthNo),
		zap.String("reason", reason))

	return nil
}

// GetPreAuth 查询预授权
func (s *preAuthService) GetPreAuth(ctx context.Context, merchantID uuid.UUID, preAuthNo string) (*model.PreAuthPayment, error) {
	return s.preAuthRepo.GetByPreAuthNo(ctx, merchantID, preAuthNo)
}

// ListPreAuths 获取预授权列表
func (s *preAuthService) ListPreAuths(ctx context.Context, merchantID uuid.UUID, status string, page, pageSize int) ([]*model.PreAuthPayment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.preAuthRepo.ListByMerchant(ctx, merchantID, status, offset, pageSize)
}

// ScanAndExpirePreAuths 扫描并过期超时的预授权
func (s *preAuthService) ScanAndExpirePreAuths(ctx context.Context) (int, error) {
	preAuths, err := s.preAuthRepo.GetExpiredPreAuths(ctx, 100)
	if err != nil {
		return 0, fmt.Errorf("查询过期预授权失败: %w", err)
	}

	expiredCount := 0
	for _, preAuth := range preAuths {
		// 取消渠道的预授权
		if preAuth.ChannelTradeNo != "" {
			_, _ = s.channelClient.CancelPreAuth(ctx, &client.CancelPreAuthRequest{
				PreAuthNo:      preAuth.PreAuthNo,
				ChannelTradeNo: preAuth.ChannelTradeNo,
			})
		}

		// 更新状态为已过期
		err := s.preAuthRepo.UpdateToExpired(ctx, preAuth.ID)
		if err != nil {
			logger.Error("更新预授权为过期状态失败",
				zap.String("pre_auth_no", preAuth.PreAuthNo),
				zap.Error(err))
			continue
		}

		expiredCount++
		logger.Info("预授权已自动过期",
			zap.String("pre_auth_no", preAuth.PreAuthNo),
			zap.Time("expires_at", preAuth.ExpiresAt))
	}

	if expiredCount > 0 {
		logger.Info("预授权过期扫描完成",
			zap.Int("total", len(preAuths)),
			zap.Int("expired", expiredCount))
	}

	return expiredCount, nil
}

// generatePreAuthNo 生成预授权单号
func generatePreAuthNo() string {
	return fmt.Sprintf("PA%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8])
}

// generatePaymentNo 生成支付单号
func generatePaymentNo() string {
	return fmt.Sprintf("PAY%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8])
}

// timePtr 返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}

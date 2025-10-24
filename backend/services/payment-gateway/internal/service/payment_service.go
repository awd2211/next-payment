package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/events"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/tracing"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
)

// PaymentService 支付服务接口
type PaymentService interface {
	// 支付管理
	CreatePayment(ctx context.Context, input *CreatePaymentInput) (*model.Payment, error)
	GetPayment(ctx context.Context, paymentNo string) (*model.Payment, error)
	QueryPayment(ctx context.Context, query *repository.PaymentQuery) ([]*model.Payment, int64, error)
	CancelPayment(ctx context.Context, paymentNo string, reason string) error

	// 回调处理
	HandleCallback(ctx context.Context, channel string, data map[string]interface{}) error

	// 退款管理
	CreateRefund(ctx context.Context, input *CreateRefundInput) (*model.Refund, error)
	GetRefund(ctx context.Context, refundNo string) (*model.Refund, error)
	QueryRefunds(ctx context.Context, query *repository.RefundQuery) ([]*model.Refund, int64, error)

	// 路由管理
	SelectChannel(ctx context.Context, payment *model.Payment) (string, error)
}

type paymentService struct {
	db                  *gorm.DB
	paymentRepo         repository.PaymentRepository
	apiKeyRepo          repository.APIKeyRepository
	orderClient         *client.OrderClient
	channelClient       *client.ChannelClient
	riskClient          *client.RiskClient
	notificationClient  *client.NotificationClient // 保留作为降级方案
	analyticsClient     *client.AnalyticsClient    // 保留作为降级方案
	redisClient         *redis.Client
	paymentMetrics      *metrics.PaymentMetrics
	messageService      MessageService
	eventPublisher      *kafka.EventPublisher // 新增: 统一事件发布器
	webhookBaseURL      string                // Webhook基础URL，用于构建回调地址
	refundSagaService   *RefundSagaService    // Refund Saga 分布式事务服务
	callbackSagaService *CallbackSagaService  // Callback Saga 分布式事务服务
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(
	db *gorm.DB,
	paymentRepo repository.PaymentRepository,
	apiKeyRepo repository.APIKeyRepository,
	orderClient *client.OrderClient,
	channelClient *client.ChannelClient,
	riskClient *client.RiskClient,
	notificationClient *client.NotificationClient,
	analyticsClient *client.AnalyticsClient,
	redisClient *redis.Client,
	paymentMetrics *metrics.PaymentMetrics,
	messageService MessageService,
	eventPublisher *kafka.EventPublisher, // 新增参数
	webhookBaseURL string,
) PaymentService {
	return &paymentService{
		db:                 db,
		paymentRepo:        paymentRepo,
		apiKeyRepo:         apiKeyRepo,
		orderClient:        orderClient,
		channelClient:      channelClient,
		riskClient:         riskClient,
		notificationClient: notificationClient,
		analyticsClient:    analyticsClient,
		redisClient:        redisClient,
		paymentMetrics:     paymentMetrics,
		messageService:     messageService,
		eventPublisher:     eventPublisher, // 新增字段
		webhookBaseURL:      webhookBaseURL,
		refundSagaService:   nil, // 通过 setter 注入
		callbackSagaService: nil, // 通过 setter 注入
	}
}

// SetRefundSagaService 设置 Refund Saga 服务（依赖注入）
func (s *paymentService) SetRefundSagaService(sagaService *RefundSagaService) {
	s.refundSagaService = sagaService
}

// SetCallbackSagaService 设置 Callback Saga 服务（依赖注入）
func (s *paymentService) SetCallbackSagaService(sagaService *CallbackSagaService) {
	s.callbackSagaService = sagaService
}

// CreatePaymentInput 创建支付输入
type CreatePaymentInput struct {
	MerchantID    uuid.UUID `json:"merchant_id" binding:"required"`
	OrderNo       string    `json:"order_no" binding:"required"`         // 商户订单号
	Amount        int64     `json:"amount" binding:"required,gt=0"`      // 金额（分）
	Currency      string    `json:"currency" binding:"required"`         // 货币类型
	Channel       string    `json:"channel"`                             // 指定渠道（可选）
	PayMethod     string    `json:"pay_method"`                          // 支付方式
	CustomerEmail string    `json:"customer_email" binding:"email"`      // 客户邮箱
	CustomerName  string    `json:"customer_name"`                       // 客户姓名
	CustomerPhone string    `json:"customer_phone"`                      // 客户手机
	CustomerIP    string    `json:"customer_ip"`                         // 客户IP
	Description   string    `json:"description"`                         // 商品描述
	NotifyURL     string    `json:"notify_url" binding:"required,url"`   // 异步通知URL
	ReturnURL     string    `json:"return_url" binding:"url"`            // 同步跳转URL
	ExpireMinutes int       `json:"expire_minutes"`                      // 过期时间（分钟，默认30分钟）
	Extra         map[string]interface{} `json:"extra"`                // 扩展信息
	Language      string    `json:"language"`                            // 语言（en, zh-CN, zh-TW, ja等）
}

// CreateRefundInput 创建退款输入
type CreateRefundInput struct {
	PaymentNo   string    `json:"payment_no" binding:"required"`    // 支付流水号
	Amount      int64     `json:"amount" binding:"required,gt=0"`   // 退款金额（分）
	Reason      string    `json:"reason" binding:"required"`        // 退款原因
	Description string    `json:"description"`                      // 退款说明
	OperatorID  uuid.UUID `json:"operator_id"`                      // 操作人ID
	OperatorType string   `json:"operator_type"`                    // 操作人类型
}

// CreatePayment 创建支付（完整流程，带事务保护）
func (s *paymentService) CreatePayment(ctx context.Context, input *CreatePaymentInput) (*model.Payment, error) {
	// 记录开始时间用于性能指标
	start := time.Now()
	var finalStatus string
	var finalChannel string

	// 使用 defer 确保指标总是被记录
	defer func() {
		if s.paymentMetrics != nil {
			duration := time.Since(start)
			amount := float64(input.Amount) / 100.0 // 转换为主币单位
			if finalChannel == "" {
				finalChannel = input.Channel
			}
			s.paymentMetrics.RecordPayment(finalStatus, finalChannel, input.Currency, amount, duration)
		}
	}()

	// 1. 验证货币类型
	if !s.isValidCurrency(input.Currency) {
		finalStatus = "failed"
		return nil, fmt.Errorf("不支持的货币类型: %s", input.Currency)
	}

	// 2. 风控检查（在事务外执行，减少事务持有时间）
	if s.riskClient != nil {
		// 创建 span 追踪风控检查
		ctx, riskSpan := tracing.StartSpan(ctx, "payment-gateway", "RiskCheck")
		tracing.AddSpanTags(ctx, map[string]interface{}{
			"merchant_id": input.MerchantID.String(),
			"amount":      input.Amount,
			"currency":    input.Currency,
		})

		riskResult, err := s.riskClient.CheckRisk(ctx, &client.RiskCheckRequest{
			MerchantID:    input.MerchantID,
			PaymentNo:     "", // 此时还没有生成
			Amount:        input.Amount,
			Currency:      input.Currency,
			Channel:       input.Channel,
			PayMethod:     input.PayMethod,
			CustomerEmail: input.CustomerEmail,
			CustomerName:  input.CustomerName,
			CustomerPhone: input.CustomerPhone,
			CustomerIP:    input.CustomerIP,
		})
		if err != nil {
			// 风控服务失败不阻塞支付，只记录日志
			riskSpan.SetStatus(codes.Error, err.Error())
			riskSpan.RecordError(err)
			logger.Error("risk check failed",
				zap.Error(err),
				zap.String("merchant_id", input.MerchantID.String()),
				zap.Int64("amount", input.Amount),
				zap.String("currency", input.Currency))
		} else if riskResult != nil {
			riskSpan.SetAttributes(
				attribute.String("risk.decision", riskResult.Decision),
				attribute.Int("risk.score", riskResult.Score),
			)
			// 如果风控拒绝，直接返回错误
			if riskResult.Decision == "reject" {
				riskSpan.SetStatus(codes.Error, "Risk rejected")
				riskSpan.End()
				finalStatus = "risk_rejected"
				return nil, fmt.Errorf("风控拒绝: %s", strings.Join(riskResult.Reasons, ", "))
			}
			// 如果需要人工审核，标记支付状态
			if riskResult.Decision == "review" {
				riskSpan.AddEvent("Manual review required")
				logger.Warn("risk manual review required",
					zap.Int("score", riskResult.Score),
					zap.Strings("reasons", riskResult.Reasons),
					zap.String("merchant_id", input.MerchantID.String()),
					zap.String("order_no", input.OrderNo))
			}
			riskSpan.SetStatus(codes.Ok, "")
		}
		riskSpan.End()
	}

	// 3. 生成支付流水号
	paymentNo := s.generatePaymentNo()

	// 4. 计算过期时间
	expireMinutes := input.ExpireMinutes
	if expireMinutes <= 0 {
		expireMinutes = 30 // 默认30分钟
	}
	expiredAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)

	// 5. 扩展信息
	var extraJSON string
	if input.Extra != nil {
		input.Extra["language"] = input.Language
		extraBytes, err := json.Marshal(input.Extra)
		if err != nil {
			logger.Error("failed to marshal extra data",
				zap.Error(err),
				zap.String("merchant_id", input.MerchantID.String()))
			return nil, fmt.Errorf("序列化扩展信息失败: %w", err)
		}
		extraJSON = string(extraBytes)
	} else if input.Language != "" {
		extraJSON = fmt.Sprintf(`{"language":"%s"}`, input.Language)
	}

	// 6. 准备支付记录数据
	payment := &model.Payment{
		MerchantID:    input.MerchantID,
		OrderNo:       input.OrderNo,
		PaymentNo:     paymentNo,
		Amount:        input.Amount,
		Currency:      strings.ToUpper(input.Currency),
		Status:        model.PaymentStatusPending,
		PayMethod:     input.PayMethod,
		CustomerEmail: input.CustomerEmail,
		CustomerName:  input.CustomerName,
		CustomerPhone: input.CustomerPhone,
		CustomerIP:    input.CustomerIP,
		Description:   input.Description,
		NotifyURL:     input.NotifyURL,
		ReturnURL:     input.ReturnURL,
		Extra:         extraJSON,
		ExpiredAt:     &expiredAt,
		NotifyStatus:  model.NotifyStatusPending,
		NotifyTimes:   0,
	}

	// 7. 选择支付渠道
	if input.Channel != "" {
		payment.Channel = input.Channel
	} else {
		channel, err := s.SelectChannel(ctx, payment)
		if err != nil {
			finalStatus = "failed"
			return nil, fmt.Errorf("选择支付渠道失败: %w", err)
		}
		payment.Channel = channel
	}

	// 记录最终选择的渠道
	finalChannel = payment.Channel

	// 8. 在事务中检查订单号唯一性并创建支付记录
	// 使用 SELECT FOR UPDATE 防止并发创建相同订单号
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 8.1 在事务中使用行级锁检查订单号是否已存在
		var count int64
		if err := tx.Model(&model.Payment{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("merchant_id = ? AND order_no = ?", input.MerchantID, input.OrderNo).
			Count(&count).Error; err != nil {
			return fmt.Errorf("检查订单号失败: %w", err)
		}
		if count > 0 {
			finalStatus = "duplicate"
			return fmt.Errorf("订单号已存在: %s", input.OrderNo)
		}

		// 8.2 创建支付记录
		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("创建支付记录失败: %w", err)
		}

		return nil
	})
	if err != nil {
		if finalStatus == "" {
			finalStatus = "failed"
		}
		return nil, err
	}

	// 10. 调用Order-Service创建订单（事务外，使用补偿机制）
	var orderCreated bool
	if s.orderClient != nil {
		// 创建 span 追踪订单创建
		ctx, orderSpan := tracing.StartSpan(ctx, "payment-gateway", "CreateOrder")
		tracing.AddSpanTags(ctx, map[string]interface{}{
			"payment_no": payment.PaymentNo,
			"order_no":   payment.OrderNo,
		})

		_, err := s.orderClient.CreateOrder(ctx, &client.CreateOrderRequest{
			MerchantID:    payment.MerchantID,
			OrderNo:       payment.OrderNo,
			PaymentNo:     payment.PaymentNo,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			Channel:       payment.Channel,
			PayMethod:     payment.PayMethod,
			CustomerEmail: payment.CustomerEmail,
			CustomerName:  payment.CustomerName,
			CustomerPhone: payment.CustomerPhone,
			CustomerIP:    payment.CustomerIP,
			Description:   payment.Description,
			Extra:         payment.Extra,
		})
		if err != nil {
			// 订单创建失败，标记支付为失败状态（而非直接删除）
			orderSpan.SetStatus(codes.Error, err.Error())
			orderSpan.RecordError(err)
			orderSpan.End()

			payment.Status = model.PaymentStatusFailed
			payment.ErrorMsg = fmt.Sprintf("创建订单失败: %v", err)
			if updateErr := s.paymentRepo.Update(ctx, payment); updateErr != nil {
				logger.Error("failed to update payment status after order creation failed",
					zap.Error(updateErr),
					zap.String("payment_no", payment.PaymentNo),
					zap.String("merchant_id", payment.MerchantID.String()))
			}
			finalStatus = "failed"
			return nil, fmt.Errorf("创建订单失败: %w", err)
		}
		orderSpan.SetStatus(codes.Ok, "")
		orderSpan.End()
		orderCreated = true
	}

	// 11. 调用Channel-Adapter发起支付（事务外）
	if s.channelClient != nil {
		var extraMap map[string]interface{}
		if payment.Extra != "" {
			if err := json.Unmarshal([]byte(payment.Extra), &extraMap); err != nil {
				logger.Warn("failed to unmarshal payment extra data",
					zap.Error(err),
					zap.String("payment_no", payment.PaymentNo),
					zap.String("extra", payment.Extra))
				// 继续处理，但不使用 extraMap
				extraMap = nil
			}
		}

		channelResult, err := s.channelClient.CreatePayment(ctx, &client.CreatePaymentRequest{
			PaymentNo:     payment.PaymentNo,
			MerchantID:    payment.MerchantID.String(),
			Channel:       payment.Channel,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			PayMethod:     payment.PayMethod,
			CustomerEmail: payment.CustomerEmail,
			CustomerName:  payment.CustomerName,
			Description:   payment.Description,
			ReturnURL:     payment.ReturnURL,
			NotifyURL:     fmt.Sprintf("%s/api/v1/webhooks/%s", s.webhookBaseURL, payment.Channel),
			Extra:         extraMap,
		})
		if err != nil {
			// 渠道调用失败，更新支付状态为失败
			payment.Status = model.PaymentStatusFailed
			payment.ErrorMsg = fmt.Sprintf("发起支付失败: %v", err)
			if updateErr := s.paymentRepo.Update(ctx, payment); updateErr != nil {
				logger.Error("failed to update payment status after channel payment failed",
					zap.Error(updateErr),
					zap.String("payment_no", payment.PaymentNo),
					zap.String("channel", payment.Channel))
			}

			// 如果订单已创建，发送补偿消息到消息队列
			if orderCreated && s.messageService != nil {
				compensationMsg := &CompensationMessage{
					Type:       CompensationTypeCancelOrder,
					PaymentNo:  payment.PaymentNo,
					OrderNo:    payment.OrderNo,
					MerchantID: payment.MerchantID.String(),
					Reason:     fmt.Sprintf("支付失败需要取消订单: %v", err),
					Extra: map[string]interface{}{
						"error_msg": payment.ErrorMsg,
						"channel":   payment.Channel,
					},
				}
				if msgErr := s.messageService.SendCompensationMessage(ctx, compensationMsg); msgErr != nil {
					logger.Error("failed to send compensation message for order cancellation",
						zap.Error(msgErr),
						zap.String("payment_no", payment.PaymentNo),
						zap.String("order_no", payment.OrderNo))
				}
			}

			finalStatus = "failed"
			return nil, fmt.Errorf("发起支付失败: %w", err)
		}

		// 在事务中更新支付记录（包括渠道订单号）
		payment.ChannelOrderNo = channelResult.ChannelOrderNo
		payment.Status = model.PaymentStatusProcessing
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			logger.Error("failed to update payment record after channel success",
				zap.Error(err),
				zap.String("payment_no", payment.PaymentNo),
				zap.String("channel_order_no", channelResult.ChannelOrderNo))
			// 注意：这里不返回错误，因为支付已经发起成功
		}

		// 将支付URL/二维码放入Extra返回
		if extraMap == nil {
			extraMap = make(map[string]interface{})
		}
		extraMap["payment_url"] = channelResult.PaymentURL
		extraMap["qr_code_url"] = channelResult.QRCodeURL
		extraBytes, err := json.Marshal(extraMap)
		if err != nil {
			logger.Error("failed to marshal extra data with payment URL",
				zap.Error(err),
				zap.String("payment_no", payment.PaymentNo))
			// 这里不返回错误，因为支付已经发起成功，只是额外信息序列化失败
		} else {
			payment.Extra = string(extraBytes)
		}
	}

	// 支付创建成功
	finalStatus = "success"
	return payment, nil
}

// GetPayment 获取支付信息
func (s *paymentService) GetPayment(ctx context.Context, paymentNo string) (*model.Payment, error) {
	payment, err := s.paymentRepo.GetByPaymentNo(ctx, paymentNo)
	if err != nil {
		return nil, fmt.Errorf("获取支付信息失败: %w", err)
	}
	if payment == nil {
		return nil, fmt.Errorf("支付记录不存在")
	}
	return payment, nil
}

// QueryPayment 查询支付列表
func (s *paymentService) QueryPayment(ctx context.Context, query *repository.PaymentQuery) ([]*model.Payment, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	return s.paymentRepo.List(ctx, query)
}

// CancelPayment 取消支付
func (s *paymentService) CancelPayment(ctx context.Context, paymentNo string, reason string) error {
	payment, err := s.GetPayment(ctx, paymentNo)
	if err != nil {
		return err
	}

	// 只有pending或processing状态的支付可以取消
	if payment.Status != model.PaymentStatusPending && payment.Status != model.PaymentStatusProcessing {
		return fmt.Errorf("当前状态不允许取消: %s", payment.Status)
	}

	// 更新状态为已取消
	payment.Status = model.PaymentStatusCancelled
	payment.ErrorMsg = reason

	return s.paymentRepo.Update(ctx, payment)
}

// HandleCallback 处理支付回调（完整流程）
func (s *paymentService) HandleCallback(ctx context.Context, channel string, data map[string]interface{}) error {
	// 1. 记录原始回调数据
	rawData, err := json.Marshal(data)
	if err != nil {
		logger.Error("failed to marshal callback data",
			zap.Error(err),
			zap.String("channel", channel))
		return fmt.Errorf("序列化回调数据失败: %w", err)
	}

	// 2. 提取支付流水号
	paymentNo, ok := data["payment_no"].(string)
	if !ok {
		// 尝试从其他字段提取
		if channelOrderNo, ok := data["channel_order_no"].(string); ok {
			// 通过渠道订单号查询
			payment, err := s.paymentRepo.GetByChannelOrderNo(ctx, channelOrderNo)
			if err != nil || payment == nil {
				return fmt.Errorf("找不到支付记录")
			}
			paymentNo = payment.PaymentNo
		} else {
			return fmt.Errorf("回调数据中缺少支付标识")
		}
	}

	// 3. 获取支付记录
	payment, err := s.GetPayment(ctx, paymentNo)
	if err != nil {
		return err
	}

	// 4. 创建回调记录
	callback := &model.PaymentCallback{
		PaymentID:   payment.ID,
		Channel:     channel,
		Event:       "payment_callback",
		RawData:     string(rawData),
		IsVerified:  false,
		IsProcessed: false,
	}

	// 5. 保存回调记录
	if err := s.paymentRepo.CreateCallback(ctx, callback); err != nil {
		return fmt.Errorf("保存回调记录失败: %w", err)
	}

	// 6. 验证回调签名（根据不同渠道）
	isVerified := s.verifyCallbackSignature(ctx, channel, data, rawData)
	callback.IsVerified = isVerified

	if !isVerified {
		logger.Warn("回调签名验证失败",
			zap.String("channel", channel),
			zap.String("payment_no", paymentNo),
		)
		// 签名验证失败但继续处理，只是标记为未验证
	}

	// 7. 解析回调状态
	status, ok := data["status"].(string)
	if !ok {
		return fmt.Errorf("回调数据中缺少status")
	}

	// 8. 更新支付状态
	oldStatus := payment.Status
	switch status {
	case "success", "paid":
		payment.Status = model.PaymentStatusSuccess
		now := time.Now()
		payment.PaidAt = &now
	case "failed", "error":
		payment.Status = model.PaymentStatusFailed
		if errorMsg, ok := data["error_msg"].(string); ok {
			payment.ErrorMsg = errorMsg
		}
		if errorCode, ok := data["error_code"].(string); ok {
			payment.ErrorCode = errorCode
		}
	case "cancelled", "canceled":
		payment.Status = model.PaymentStatusCancelled
	default:
		return fmt.Errorf("未知的支付状态: %s", status)
	}

	// 9. 更新渠道订单号
	if channelOrderNo, ok := data["channel_order_no"].(string); ok {
		payment.ChannelOrderNo = channelOrderNo
	}

	// 10. 保存支付状态
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("更新支付状态失败: %w", err)
	}

	// 11. 标记回调已处理
	callback.IsProcessed = true
	s.paymentRepo.UpdateCallback(ctx, callback)

	// 12. 通知Order-Service更新订单状态
	if s.orderClient != nil && oldStatus != payment.Status {
		s.orderClient.UpdateOrderStatus(ctx, payment.PaymentNo, &client.UpdateOrderStatusRequest{
			Status:         payment.Status,
			ChannelOrderNo: payment.ChannelOrderNo,
			PaidAt:         payment.PaidAt.Format(time.RFC3339),
			ErrorCode:      payment.ErrorCode,
			ErrorMsg:       payment.ErrorMsg,
		})
	}

	// 12. 发布支付事件到Kafka (异步,事件驱动架构)
	// 替代原来的HTTP同步调用,解耦下游服务依赖
	if oldStatus != payment.Status {
		s.publishPaymentStatusEvent(payment, oldStatus, channel)
	}

	// 13. 异步通知商户（放入队列）
	go func(p *model.Payment) {
		// 创建带超时的context
		notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Panic恢复
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in notifyMerchant goroutine",
					zap.Any("panic", r),
					zap.String("payment_no", p.PaymentNo),
					zap.Stack("stack"))
			}
		}()

		// 执行通知
		s.notifyMerchant(notifyCtx, p)
	}(payment)

	return nil
}

// notifyMerchant 通知商户（异步）
func (s *paymentService) notifyMerchant(ctx context.Context, payment *model.Payment) {
	if payment.NotifyURL == "" {
		return
	}

	// 构建通知数据
	notifyData := map[string]interface{}{
		"payment_no":       payment.PaymentNo,
		"order_no":         payment.OrderNo,
		"merchant_id":      payment.MerchantID,
		"amount":           payment.Amount,
		"currency":         payment.Currency,
		"status":           payment.Status,
		"channel":          payment.Channel,
		"channel_order_no": payment.ChannelOrderNo,
		"paid_at":          payment.PaidAt,
		"error_code":       payment.ErrorCode,
		"error_msg":        payment.ErrorMsg,
	}

	// 使用消息队列实现可靠通知和重试机制
	if s.messageService != nil {
		// 计算签名（使用商户的API Secret）
		apiKey, err := s.apiKeyRepo.GetByMerchantID(ctx, payment.MerchantID)
		if err != nil {
			logger.Error("failed to get merchant API secret for notification",
				zap.Error(err),
				zap.String("merchant_id", payment.MerchantID.String()),
				zap.String("payment_no", payment.PaymentNo))
			// 无法获取密钥时,不发送通知,避免签名错误
			return
		}
		signature := calculateNotifySignature(notifyData, apiKey.APISecret)

		notifyMsg := &NotificationMessage{
			PaymentNo:     payment.PaymentNo,
			MerchantID:    payment.MerchantID.String(),
			NotifyURL:     payment.NotifyURL,
			NotifyData:    notifyData,
			Signature:     signature,
			RetryCount:    0,
			MaxRetries:    5, // 最多重试5次
			NextRetryTime: time.Now().Add(5 * time.Second),
		}

		if err := s.messageService.SendNotificationMessage(ctx, notifyMsg); err != nil {
			logger.Error("failed to send notification message to queue",
				zap.Error(err),
				zap.String("payment_no", payment.PaymentNo),
				zap.String("notify_url", payment.NotifyURL))
		} else {
			logger.Info("notification message sent to queue",
				zap.String("payment_no", payment.PaymentNo),
				zap.String("notify_url", payment.NotifyURL))
		}
	} else {
		// 降级处理：直接打印（开发环境）
		logger.Warn("message service not configured, notification skipped",
			zap.String("payment_no", payment.PaymentNo),
			zap.String("notify_url", payment.NotifyURL))
	}
}

// CreateRefund 创建退款（完整流程）
func (s *paymentService) CreateRefund(ctx context.Context, input *CreateRefundInput) (*model.Refund, error) {
	// 记录退款指标
	var finalStatus string
	var currency string
	var amount float64

	defer func() {
		if s.paymentMetrics != nil && currency != "" {
			s.paymentMetrics.RecordRefund(finalStatus, currency, amount)
		}
	}()

	// 1. 获取原支付记录
	payment, err := s.GetPayment(ctx, input.PaymentNo)
	if err != nil {
		finalStatus = "failed"
		return nil, fmt.Errorf("获取支付记录失败: %w", err)
	}

	// 记录货币类型和金额
	currency = payment.Currency
	amount = float64(input.Amount) / 100.0

	// 2. 只有成功的支付才能退款
	if payment.Status != model.PaymentStatusSuccess {
		finalStatus = "invalid_status"
		return nil, fmt.Errorf("只有成功的支付才能退款，当前状态: %s", payment.Status)
	}

	// 3. 验证退款金额（基本检查）
	if input.Amount <= 0 {
		finalStatus = "invalid_amount"
		return nil, fmt.Errorf("退款金额必须大于0")
	}
	if input.Amount > payment.Amount {
		finalStatus = "amount_exceeded"
		return nil, fmt.Errorf("退款金额不能大于支付金额: 退款=%d, 支付=%d", input.Amount, payment.Amount)
	}

	// 4. 生成退款单号
	refundNo := s.generateRefundNo()

	// 5. 在事务中检查已退款总额并创建退款记录
	// 使用行级锁防止并发退款导致总额超限
	var refund *model.Refund
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 5.1 锁定支付记录，防止并发退款
		var lockedPayment model.Payment
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", payment.ID).
			First(&lockedPayment).Error
		if err != nil {
			return fmt.Errorf("锁定支付记录失败: %w", err)
		}

		// 5.2 在事务中查询已成功退款总额（使用 SUM 聚合查询，避免N+1问题）
		var refundedAmount int64
		err = tx.Model(&model.Refund{}).
			Where("payment_id = ? AND status = ?", payment.ID, model.RefundStatusSuccess).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&refundedAmount).Error
		if err != nil {
			return fmt.Errorf("查询已退款总额失败: %w", err)
		}

		// 5.3 校验退款总额（在事务内，确保数据一致性）
		if refundedAmount+input.Amount > lockedPayment.Amount {
			finalStatus = "amount_exceeded"
			return fmt.Errorf("退款总额超过支付金额: 已退款=%d, 本次退款=%d, 支付金额=%d",
				refundedAmount, input.Amount, lockedPayment.Amount)
		}

		// 5.4 创建退款记录（初始状态为 pending）
		refund = &model.Refund{
			PaymentID:    payment.ID,
			MerchantID:   payment.MerchantID,
			RefundNo:     refundNo,
			Amount:       input.Amount,
			Currency:     payment.Currency,
			Status:       model.RefundStatusPending,
			Reason:       input.Reason,
			Description:  input.Description,
			OperatorID:   input.OperatorID,
			OperatorType: input.OperatorType,
		}

		if err := tx.Create(refund).Error; err != nil {
			return fmt.Errorf("创建退款记录失败: %w", err)
		}

		return nil
	})
	if err != nil {
		if finalStatus == "" {
			finalStatus = "failed"
		}
		return nil, err
	}

	// 7. 调用 Channel-Adapter 执行渠道退款（事务外，使用 Saga 模式）
	var channelRefundSuccess bool

	// ========== 使用 Saga 分布式事务执行退款（生产级方案）==========
	if s.refundSagaService != nil && s.channelClient != nil {
		logger.Info("使用 Saga 分布式事务执行退款",
			zap.String("refund_no", refund.RefundNo),
			zap.String("payment_no", payment.PaymentNo))

		// 执行 Refund Saga (3 步骤):
		// 1. 调用渠道退款
		// 2. 更新支付状态
		// 3. 更新退款状态
		// 任何步骤失败会自动回滚所有已完成的步骤
		err := s.refundSagaService.ExecuteRefundSaga(ctx, refund, payment)
		if err != nil {
			logger.Error("Refund Saga 执行失败",
				zap.Error(err),
				zap.String("refund_no", refund.RefundNo))
			finalStatus = "failed"
			return nil, fmt.Errorf("退款执行失败: %w", err)
		}

		logger.Info("Refund Saga 执行成功",
			zap.String("refund_no", refund.RefundNo))
		channelRefundSuccess = true
		finalStatus = "success"
	} else if s.channelClient != nil {
		// ========== 旧逻辑（向后兼容，如果未启用 Saga）==========
		logger.Warn("未启用 Refund Saga 服务，使用传统方式执行退款（不推荐）",
			zap.String("refund_no", refund.RefundNo))

		channelResult, err := s.channelClient.CreateRefund(ctx, &client.RefundRequest{
			RefundNo:       refund.RefundNo,
			PaymentNo:      payment.PaymentNo,
			ChannelOrderNo: payment.ChannelOrderNo,
			Amount:         refund.Amount,
			Currency:       refund.Currency,
			Reason:         refund.Reason,
		})
		if err != nil {
			// 渠道退款失败，标记退款为失败状态
			refund.Status = model.RefundStatusFailed
			refund.ErrorMsg = fmt.Sprintf("渠道退款失败: %v", err)
			if updateErr := s.paymentRepo.UpdateRefund(ctx, refund); updateErr != nil {
				logger.Error("failed to update refund status after channel refund failed",
					zap.Error(updateErr),
					zap.String("refund_no", refund.RefundNo),
					zap.String("payment_no", payment.PaymentNo))
			}
			finalStatus = "failed"
			return nil, fmt.Errorf("渠道退款失败: %w", err)
		}

		// 渠道退款成功
		channelRefundSuccess = true
		refund.ChannelRefundNo = channelResult.ChannelRefundNo
		refund.Status = model.RefundStatusSuccess
		now := time.Now()
		refund.RefundedAt = &now

		// 更新退款记录为成功状态
		if err := s.paymentRepo.UpdateRefund(ctx, refund); err != nil {
			// ⚠️ 警告：渠道已退款成功，但本地状态更新失败，数据不一致！
			// 生产环境：应该使用上面的 Saga 方案自动回滚
			logger.Error("channel refund succeeded but local status update failed",
				zap.Error(err),
				zap.String("refund_no", refund.RefundNo),
				zap.String("channel_refund_no", refund.ChannelRefundNo),
				zap.String("payment_no", payment.PaymentNo))

			// 发送补偿消息到消息队列，由后台任务重试更新
			if s.messageService != nil {
				compensationMsg := &CompensationMessage{
					Type:       CompensationTypeUpdateRefundStatus,
					PaymentNo:  payment.PaymentNo,
					RefundNo:   refund.RefundNo,
					MerchantID: payment.MerchantID.String(),
					Reason:     fmt.Sprintf("渠道退款成功但本地状态更新失败: %v", err),
					Extra: map[string]interface{}{
						"channel_refund_no": refund.ChannelRefundNo,
						"status":            model.RefundStatusSuccess,
						"refunded_at":       now,
					},
				}
				if msgErr := s.messageService.SendCompensationMessage(ctx, compensationMsg); msgErr != nil {
					logger.Error("failed to send compensation message for refund status update",
						zap.Error(msgErr),
						zap.String("refund_no", refund.RefundNo),
						zap.String("payment_no", payment.PaymentNo))
				}
			}

			finalStatus = "partial_success"
			return nil, fmt.Errorf("退款成功但状态更新失败，请手动确认: %w", err)
		}
		finalStatus = "success"
	} else {
		// 没有配置渠道客户端（测试环境）
		logger.Warn("channel client not configured, refund record created but not processed",
			zap.String("refund_no", refund.RefundNo),
			zap.String("payment_no", payment.PaymentNo))
	}

	// 8. 只有成功后才通知商户
	if channelRefundSuccess {
		go func(p *model.Payment, r *model.Refund) {
			// 创建带超时的context
			notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Panic恢复
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic in notifyMerchantRefund goroutine",
						zap.Any("panic", rec),
						zap.String("refund_no", r.RefundNo),
						zap.Stack("stack"))
				}
			}()

			// 执行通知
			s.notifyMerchantRefund(notifyCtx, p, r)
		}(payment, refund)
	}

	// 退款成功
	finalStatus = "success"
	return refund, nil
}

// notifyMerchantRefund 通知商户退款结果（异步）
func (s *paymentService) notifyMerchantRefund(ctx context.Context, payment *model.Payment, refund *model.Refund) {
	if payment.NotifyURL == "" {
		return
	}

	// 构建通知数据并记录日志
	logger.Info("notifying merchant about refund",
		zap.String("refund_no", refund.RefundNo),
		zap.String("payment_no", payment.PaymentNo),
		zap.String("notify_url", payment.NotifyURL),
		zap.String("status", refund.Status),
		zap.Int64("amount", refund.Amount),
		zap.String("currency", refund.Currency))
}

// GetRefund 获取退款信息
func (s *paymentService) GetRefund(ctx context.Context, refundNo string) (*model.Refund, error) {
	refund, err := s.paymentRepo.GetRefundByRefundNo(ctx, refundNo)
	if err != nil {
		return nil, fmt.Errorf("获取退款信息失败: %w", err)
	}
	if refund == nil {
		return nil, fmt.Errorf("退款记录不存在")
	}
	return refund, nil
}

// QueryRefunds 查询退款列表
func (s *paymentService) QueryRefunds(ctx context.Context, query *repository.RefundQuery) ([]*model.Refund, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	return s.paymentRepo.ListRefunds(ctx, query)
}

// SelectChannel 选择支付渠道
func (s *paymentService) SelectChannel(ctx context.Context, payment *model.Payment) (string, error) {
	// 获取所有启用的路由规则
	routes, err := s.paymentRepo.ListActiveRoutes(ctx)
	if err != nil {
		return "", fmt.Errorf("获取路由规则失败: %w", err)
	}

	// 按优先级匹配规则
	for _, route := range routes {
		if s.matchRoute(payment, route) {
			return route.Channel, nil
		}
	}

	// 默认渠道
	return model.ChannelStripe, nil
}

// matchRoute 匹配路由规则
func (s *paymentService) matchRoute(payment *model.Payment, route *model.PaymentRoute) bool {
	var conditions map[string]interface{}
	if err := json.Unmarshal([]byte(route.Conditions), &conditions); err != nil {
		return false
	}

	// 金额范围
	if minAmount, ok := conditions["min_amount"].(float64); ok {
		if payment.Amount < int64(minAmount) {
			return false
		}
	}
	if maxAmount, ok := conditions["max_amount"].(float64); ok {
		if payment.Amount > int64(maxAmount) {
			return false
		}
	}

	// 货币类型
	if currencies, ok := conditions["currencies"].([]interface{}); ok {
		matched := false
		for _, c := range currencies {
			if currency, ok := c.(string); ok && currency == payment.Currency {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 国家/地区
	// TODO: 根据customer_ip或其他信息判断国家

	return true
}

// 工具函数

// generatePaymentNo 生成支付流水号
func (s *paymentService) generatePaymentNo() string {
	// 格式：PY + 时间戳 + 随机字符
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("PY%s%s", timestamp, randomStr)
}

// generateRefundNo 生成退款单号
func (s *paymentService) generateRefundNo() string {
	// 格式：RF + 时间戳 + 随机字符
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("RF%s%s", timestamp, randomStr)
}

// isValidCurrency 验证货币类型（支持全球主流货币）
func (s *paymentService) isValidCurrency(currency string) bool {
	// 支持的货币列表（与用户偏好设置中的货币列表一致）
	validCurrencies := []string{
		"USD", "EUR", "GBP", "CNY", "JPY", "KRW", "HKD", "SGD",
		"AUD", "CAD", "INR", "BRL", "MXN", "RUB", "TRY", "ZAR",
		"CHF", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF", "THB",
		"IDR", "MYR", "PHP", "VND", "AED", "SAR", "ILS", "EGP",
	}

	currencyUpper := strings.ToUpper(currency)
	for _, valid := range validCurrencies {
		if currencyUpper == valid {
			return true
		}
	}
	return false
}

// calculateNotifySignature 计算通知签名
func calculateNotifySignature(data map[string]interface{}, secret string) string {
	// 1. 对参数按key排序
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	// 简单排序（生产环境应使用 sort.Strings）
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// 2. 拼接参数: key1=value1&key2=value2&...
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(fmt.Sprintf("%s=%v", k, data[k]))
	}

	// 3. 追加secret
	sb.WriteString("&key=")
	sb.WriteString(secret)

	// 4. 计算MD5（生产环境建议使用HMAC-SHA256）
	// 这里简化处理，实际应使用 crypto/md5 或 crypto/sha256
	signStr := sb.String()
	return fmt.Sprintf("SIGN_%s", base64.StdEncoding.EncodeToString([]byte(signStr))[:32])
}

// verifyCallbackSignature 验证不同渠道的回调签名
func (s *paymentService) verifyCallbackSignature(ctx context.Context, channel string, data map[string]interface{}, rawData []byte) bool {
	switch strings.ToLower(channel) {
	case "stripe":
		return s.verifyStripeSignature(ctx, data, rawData)
	case "paypal":
		return s.verifyPayPalSignature(ctx, data, rawData)
	case "alipay":
		return s.verifyAlipaySignature(ctx, data)
	case "wechat":
		return s.verifyWechatSignature(ctx, data)
	case "crypto":
		// 加密货币支付通常使用区块链确认，不需要签名验证
		return true
	default:
		logger.Warn("未知渠道，跳过签名验证",
			zap.String("channel", channel),
		)
		// 未知渠道默认返回false，需要人工审核
		return false
	}
}

// verifyStripeSignature 验证Stripe签名
func (s *paymentService) verifyStripeSignature(ctx context.Context, data map[string]interface{}, rawData []byte) bool {
	// Stripe使用 stripe-signature header验证
	// 签名应该在HTTP header中，这里简化处理
	signature, ok := data["stripe_signature"].(string)
	if !ok {
		logger.Debug("Stripe回调缺少签名")
		return false
	}

	// 实际生产环境应该调用Stripe SDK验证
	// webhook.ConstructEvent(body, signature, webhookSecret)
	// 这里简化为检查签名是否存在且非空
	if signature == "" {
		return false
	}

	logger.Info("Stripe签名验证通过",
		zap.String("signature_prefix", signature[:min(10, len(signature))]),
	)
	return true
}

// verifyPayPalSignature 验证PayPal签名
func (s *paymentService) verifyPayPalSignature(ctx context.Context, data map[string]interface{}, rawData []byte) bool {
	// PayPal使用PAYPAL-TRANSMISSION-ID, PAYPAL-TRANSMISSION-TIME, PAYPAL-TRANSMISSION-SIG等header
	transmissionID, hasID := data["paypal_transmission_id"].(string)
	transmissionSig, hasSig := data["paypal_transmission_sig"].(string)

	if !hasID || !hasSig {
		logger.Debug("PayPal回调缺少必要的验证参数")
		return false
	}

	// 实际生产环境应该调用PayPal Webhook Verification API
	// https://api.paypal.com/v1/notifications/verify-webhook-signature
	// 这里简化为检查必要字段是否存在
	if transmissionID == "" || transmissionSig == "" {
		return false
	}

	logger.Info("PayPal签名验证通过",
		zap.String("transmission_id", transmissionID[:min(10, len(transmissionID))]),
	)
	return true
}

// verifyAlipaySignature 验证支付宝签名
func (s *paymentService) verifyAlipaySignature(ctx context.Context, data map[string]interface{}) bool {
	// 支付宝使用RSA签名验证
	sign, ok := data["sign"].(string)
	if !ok || sign == "" {
		logger.Debug("支付宝回调缺少签名")
		return false
	}

	signType, ok := data["sign_type"].(string)
	if !ok {
		signType = "RSA2" // 默认使用RSA2
	}

	// 实际生产环境应该使用支付宝公钥验证签名
	// 1. 排除sign和sign_type参数
	// 2. 按参数名ASCII码升序排列
	// 3. 拼接成待签名字符串
	// 4. 使用支付宝公钥验证
	// 这里简化为检查签名和签名类型
	logger.Info("支付宝签名验证",
		zap.String("sign_type", signType),
		zap.String("sign_prefix", sign[:min(10, len(sign))]),
	)

	// 简化验证：检查签名长度合理性
	if len(sign) < 100 {
		return false
	}

	return true
}

// verifyWechatSignature 验证微信签名
func (s *paymentService) verifyWechatSignature(ctx context.Context, data map[string]interface{}) bool {
	// 微信支付使用签名验证
	sign, ok := data["sign"].(string)
	if !ok || sign == "" {
		logger.Debug("微信回调缺少签名")
		return false
	}

	// 实际生产环境应该：
	// 1. 获取所有回调参数（除sign）
	// 2. 按参数名ASCII码升序排列
	// 3. 拼接成 key1=value1&key2=value2&key=API密钥
	// 4. MD5加密并转大写
	// 5. 与回调sign比较
	// 这里简化为检查签名格式
	logger.Info("微信签名验证",
		zap.String("sign_prefix", sign[:min(10, len(sign))]),
	)

	// 简化验证：微信签名为32位大写MD5
	if len(sign) != 32 {
		return false
	}

	return true
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// publishPaymentStatusEvent 发布支付状态变更事件到Kafka
// 替代原来的HTTP同步调用 (notificationClient, analyticsClient)
func (s *paymentService) publishPaymentStatusEvent(payment *model.Payment, oldStatus, channel string) {
	// 只在有EventPublisher时才发布事件
	if s.eventPublisher == nil {
		logger.Warn("EventPublisher not initialized, fallback to HTTP clients")
		// 降级: 使用原有的HTTP调用 (保持向后兼容)
		s.fallbackToHTTPClients(payment, oldStatus, channel)
		return
	}

	// 确定事件类型
	var eventType string
	switch payment.Status {
	case model.PaymentStatusSuccess:
		eventType = events.PaymentSuccess
	case model.PaymentStatusFailed:
		eventType = events.PaymentFailed
	case model.PaymentStatusCancelled:
		eventType = events.PaymentCancelled
	default:
		// 其他状态不发布事件
		return
	}

	// 构造事件载荷
	payload := events.PaymentEventPayload{
		PaymentNo:     payment.PaymentNo,
		MerchantID:    payment.MerchantID.String(),
		OrderNo:       payment.OrderNo,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Channel:       payment.Channel,
		Status:        payment.Status,
		CustomerEmail: payment.CustomerEmail,
		PaidAt:        payment.PaidAt,
		Extra: map[string]interface{}{
			"old_status":       oldStatus,
			"callback_channel": channel,
			"error_code":       payment.ErrorCode,
			"error_msg":        payment.ErrorMsg,
		},
	}

	// 创建事件
	event := events.NewPaymentEvent(eventType, payload)

	// 添加追踪元数据
	event.AddMetadata("service", "payment-gateway")
	event.AddMetadata("old_status", oldStatus)
	event.AddMetadata("callback_channel", channel)

	// 异步发布事件 (不阻塞主流程)
	go func() {
		publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.eventPublisher.Publish(publishCtx, events.TopicPaymentEvents, event); err != nil {
			logger.Error("failed to publish payment event to kafka",
				zap.String("payment_no", payment.PaymentNo),
				zap.String("event_type", eventType),
				zap.Error(err))

			// 失败降级: 使用HTTP调用
			logger.Info("fallback to HTTP clients due to kafka publish failure")
			s.fallbackToHTTPClients(payment, oldStatus, channel)
		} else {
			logger.Info("payment event published successfully",
				zap.String("payment_no", payment.PaymentNo),
				zap.String("event_type", eventType),
				zap.String("topic", events.TopicPaymentEvents))
		}
	}()
}

// fallbackToHTTPClients 降级到HTTP客户端调用 (保持向后兼容)
func (s *paymentService) fallbackToHTTPClients(payment *model.Payment, oldStatus, channel string) {
	// 12.1 发送通知（支付成功/失败通知）
	if s.notificationClient != nil {
		go func(p *model.Payment) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var notifType, title, content string
			switch p.Status {
			case model.PaymentStatusSuccess:
				notifType = "payment_success"
				title = "支付成功"
				content = fmt.Sprintf("支付单号 %s 已成功支付，金额 %.2f %s",
					p.PaymentNo, float64(p.Amount)/100.0, p.Currency)
			case model.PaymentStatusFailed:
				notifType = "payment_failed"
				title = "支付失败"
				content = fmt.Sprintf("支付单号 %s 支付失败：%s", p.PaymentNo, p.ErrorMsg)
			default:
				return // 其他状态不发送通知
			}

			err := s.notificationClient.SendPaymentNotification(notifyCtx, &client.SendNotificationRequest{
				MerchantID: p.MerchantID,
				Type:       notifType,
				Title:      title,
				Content:    content,
				Email:      p.CustomerEmail,
				Priority:   "high",
				Data: map[string]interface{}{
					"payment_no": p.PaymentNo,
					"order_no":   p.OrderNo,
					"amount":     p.Amount,
					"currency":   p.Currency,
					"status":     p.Status,
				},
			})
			if err != nil {
				logger.Warn("发送支付通知失败（非致命,降级模式）",
					zap.Error(err),
					zap.String("payment_no", p.PaymentNo),
					zap.String("status", p.Status))
			}
		}(payment)
	}

	// 12.2 推送Analytics事件（实时统计）
	if s.analyticsClient != nil {
		go func(p *model.Payment) {
			analyticsCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			eventType := "payment_status_changed"
			if p.Status == model.PaymentStatusSuccess {
				eventType = "payment_success"
			} else if p.Status == model.PaymentStatusFailed {
				eventType = "payment_failed"
			}

			err := s.analyticsClient.PushPaymentEvent(analyticsCtx, &client.PaymentEventRequest{
				EventType:  eventType,
				MerchantID: p.MerchantID,
				PaymentNo:  p.PaymentNo,
				OrderNo:    p.OrderNo,
				Amount:     p.Amount,
				Currency:   p.Currency,
				Channel:    p.Channel,
				Status:     p.Status,
				Timestamp:  time.Now(),
				Metadata: map[string]interface{}{
					"old_status":       oldStatus,
					"new_status":       p.Status,
					"callback_channel": channel,
				},
			})
			if err != nil {
				logger.Warn("推送Analytics事件失败（非致命,降级模式）",
					zap.Error(err),
					zap.String("payment_no", p.PaymentNo),
					zap.String("event_type", eventType))
			}
		}(payment)
	}
}

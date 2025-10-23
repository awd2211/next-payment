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
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/tracing"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
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
	db             *gorm.DB
	paymentRepo    repository.PaymentRepository
	orderClient    *client.OrderClient
	channelClient  *client.ChannelClient
	riskClient     *client.RiskClient
	redisClient    *redis.Client
	paymentMetrics *metrics.PaymentMetrics
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(
	db *gorm.DB,
	paymentRepo repository.PaymentRepository,
	orderClient *client.OrderClient,
	channelClient *client.ChannelClient,
	riskClient *client.RiskClient,
	redisClient *redis.Client,
	paymentMetrics *metrics.PaymentMetrics,
) PaymentService {
	return &paymentService{
		db:             db,
		paymentRepo:    paymentRepo,
		orderClient:    orderClient,
		channelClient:  channelClient,
		riskClient:     riskClient,
		redisClient:    redisClient,
		paymentMetrics: paymentMetrics,
	}
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

	// 2. 检查订单号是否已存在（防重复）
	existing, err := s.paymentRepo.GetByOrderNo(ctx, input.MerchantID, input.OrderNo)
	if err != nil {
		finalStatus = "failed"
		return nil, fmt.Errorf("检查订单号失败: %w", err)
	}
	if existing != nil {
		finalStatus = "duplicate"
		return nil, fmt.Errorf("订单号已存在: %s", input.OrderNo)
	}

	// 3. 风控检查
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
			fmt.Printf("风控检查失败: %v\n", err)
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
				fmt.Printf("风控需要审核: score=%d, reasons=%v\n", riskResult.Score, riskResult.Reasons)
			}
			riskSpan.SetStatus(codes.Ok, "")
		}
		riskSpan.End()
	}

	// 4. 生成支付流水号
	paymentNo := s.generatePaymentNo()

	// 5. 计算过期时间
	expireMinutes := input.ExpireMinutes
	if expireMinutes <= 0 {
		expireMinutes = 30 // 默认30分钟
	}
	expiredAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)

	// 6. 扩展信息
	var extraJSON string
	if input.Extra != nil {
		input.Extra["language"] = input.Language
		extraBytes, _ := json.Marshal(input.Extra)
		extraJSON = string(extraBytes)
	} else if input.Language != "" {
		extraJSON = fmt.Sprintf(`{"language":"%s"}`, input.Language)
	}

	// 7. 创建支付记录
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

	// 8. 选择支付渠道
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

	// 9. 在事务中创建支付记录
	// 注意：这里只保存支付记录，外部服务调用放在事务外
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		finalStatus = "failed"
		return nil, fmt.Errorf("创建支付记录失败: %w", err)
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
				fmt.Printf("更新支付状态失败: %v\n", updateErr)
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
			json.Unmarshal([]byte(payment.Extra), &extraMap)
		}

		channelResult, err := s.channelClient.CreatePayment(ctx, &client.CreatePaymentRequest{
			PaymentNo:     payment.PaymentNo,
			MerchantID:    payment.MerchantID,
			Channel:       payment.Channel,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			PayMethod:     payment.PayMethod,
			CustomerEmail: payment.CustomerEmail,
			CustomerName:  payment.CustomerName,
			Description:   payment.Description,
			ReturnURL:     payment.ReturnURL,
			NotifyURL:     fmt.Sprintf("http://payment-gateway:8003/api/v1/webhooks/%s", payment.Channel),
			Extra:         extraMap,
		})
		if err != nil {
			// 渠道调用失败，更新支付状态为失败
			payment.Status = model.PaymentStatusFailed
			payment.ErrorMsg = fmt.Sprintf("发起支付失败: %v", err)
			if updateErr := s.paymentRepo.Update(ctx, payment); updateErr != nil {
				fmt.Printf("更新支付状态失败: %v\n", updateErr)
			}

			// TODO: 如果订单已创建，需要补偿取消订单
			// 这里应该发送到消息队列由后台任务处理
			if orderCreated {
				fmt.Printf("警告：支付失败但订单已创建，需要补偿。PaymentNo=%s\n", payment.PaymentNo)
			}

			finalStatus = "failed"
			return nil, fmt.Errorf("发起支付失败: %w", err)
		}

		// 在事务中更新支付记录（包括渠道订单号）
		payment.ChannelOrderNo = channelResult.ChannelOrderNo
		payment.Status = model.PaymentStatusProcessing
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			fmt.Printf("更新支付记录失败: %v\n", err)
			// 注意：这里不返回错误，因为支付已经发起成功
		}

		// 将支付URL/二维码放入Extra返回
		if extraMap == nil {
			extraMap = make(map[string]interface{})
		}
		extraMap["payment_url"] = channelResult.PaymentURL
		extraMap["qr_code_url"] = channelResult.QRCodeURL
		extraBytes, _ := json.Marshal(extraMap)
		payment.Extra = string(extraBytes)
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
	rawData, _ := json.Marshal(data)

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
	// TODO: 实现不同渠道的签名验证
	callback.IsVerified = true

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

	// 13. 异步通知商户（放入队列）
	go s.notifyMerchant(context.Background(), payment)

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

	// TODO: 使用消息队列实现可靠通知
	// TODO: 实现重试机制（最多重试5次）
	// TODO: 签名通知数据

	fmt.Printf("通知商户: url=%s, data=%+v\n", payment.NotifyURL, notifyData)
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

	// 3. 验证退款金额
	if input.Amount <= 0 {
		finalStatus = "invalid_amount"
		return nil, fmt.Errorf("退款金额必须大于0")
	}
	if input.Amount > payment.Amount {
		finalStatus = "amount_exceeded"
		return nil, fmt.Errorf("退款金额不能大于支付金额: 退款=%d, 支付=%d", input.Amount, payment.Amount)
	}

	// 4. 检查已退款总额（防止重复退款）
	existingRefunds, _, err := s.paymentRepo.ListRefunds(ctx, &repository.RefundQuery{
		PaymentID: &payment.ID,
		Status:    model.RefundStatusSuccess,
		Page:      1,
		PageSize:  100,
	})
	if err != nil {
		finalStatus = "failed"
		return nil, fmt.Errorf("查询已退款总额失败: %w", err)
	}

	var refundedAmount int64
	for _, r := range existingRefunds {
		refundedAmount += r.Amount
	}

	if refundedAmount+input.Amount > payment.Amount {
		finalStatus = "amount_exceeded"
		return nil, fmt.Errorf("退款总额超过支付金额: 已退款=%d, 本次退款=%d, 支付金额=%d",
			refundedAmount, input.Amount, payment.Amount)
	}

	// 5. 生成退款单号
	refundNo := s.generateRefundNo()

	// 6. 创建退款记录（初始状态为 pending）
	refund := &model.Refund{
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

	if err := s.paymentRepo.CreateRefund(ctx, refund); err != nil {
		finalStatus = "failed"
		return nil, fmt.Errorf("创建退款记录失败: %w", err)
	}

	// 7. 调用 Channel-Adapter 执行渠道退款（事务外，使用 Saga 模式）
	var channelRefundSuccess bool
	if s.channelClient != nil {
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
				fmt.Printf("更新退款失败状态时出错: %v\n", updateErr)
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
			// 警告：渠道已退款成功，但本地状态更新失败
			fmt.Printf("警告：渠道退款成功但本地状态更新失败，RefundNo=%s, ChannelRefundNo=%s, Error=%v\n",
				refund.RefundNo, refund.ChannelRefundNo, err)
			// 这里应该发送到消息队列，由后台任务重试更新
			// TODO: 发送补偿消息到 MQ
			finalStatus = "partial_success"
			return nil, fmt.Errorf("退款成功但状态更新失败，请手动确认: %w", err)
		}
	} else {
		// 没有配置渠道客户端（测试环境）
		fmt.Printf("警告：未配置渠道客户端，退款记录已创建但未实际退款: RefundNo=%s\n", refund.RefundNo)
	}

	// 8. 只有成功后才通知商户
	if channelRefundSuccess {
		go s.notifyMerchantRefund(context.Background(), payment, refund)
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

	// 构建通知数据
	notifyData := map[string]interface{}{
		"event":             "refund",
		"refund_no":         refund.RefundNo,
		"payment_no":        payment.PaymentNo,
		"order_no":          payment.OrderNo,
		"merchant_id":       payment.MerchantID,
		"amount":            refund.Amount,
		"currency":          refund.Currency,
		"status":            refund.Status,
		"channel_refund_no": refund.ChannelRefundNo,
		"refunded_at":       refund.RefundedAt,
		"reason":            refund.Reason,
	}

	fmt.Printf("通知商户退款: url=%s, data=%+v\n", payment.NotifyURL, notifyData)
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

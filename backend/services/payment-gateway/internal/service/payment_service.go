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
	"github.com/payment-platform/services/payment-gateway/internal/model"
	"github.com/payment-platform/services/payment-gateway/internal/repository"
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
	paymentRepo repository.PaymentRepository
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(paymentRepo repository.PaymentRepository) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
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

// CreatePayment 创建支付
func (s *paymentService) CreatePayment(ctx context.Context, input *CreatePaymentInput) (*model.Payment, error) {
	// 验证货币类型
	if !s.isValidCurrency(input.Currency) {
		return nil, fmt.Errorf("不支持的货币类型: %s", input.Currency)
	}

	// 检查订单号是否已存在
	existing, err := s.paymentRepo.GetByOrderNo(ctx, input.MerchantID, input.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("检查订单号失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("订单号已存在: %s", input.OrderNo)
	}

	// 生成支付流水号
	paymentNo := s.generatePaymentNo()

	// 计算过期时间
	expireMinutes := input.ExpireMinutes
	if expireMinutes <= 0 {
		expireMinutes = 30 // 默认30分钟
	}
	expiredAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)

	// 扩展信息
	var extraJSON string
	if input.Extra != nil {
		input.Extra["language"] = input.Language // 保存用户语言偏好
		extraBytes, _ := json.Marshal(input.Extra)
		extraJSON = string(extraBytes)
	} else if input.Language != "" {
		extraJSON = fmt.Sprintf(`{"language":"%s"}`, input.Language)
	}

	// 创建支付记录
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

	// 选择支付渠道
	if input.Channel != "" {
		payment.Channel = input.Channel
	} else {
		channel, err := s.SelectChannel(ctx, payment)
		if err != nil {
			return nil, fmt.Errorf("选择支付渠道失败: %w", err)
		}
		payment.Channel = channel
	}

	// 创建支付记录
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("创建支付记录失败: %w", err)
	}

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

// HandleCallback 处理支付回调
func (s *paymentService) HandleCallback(ctx context.Context, channel string, data map[string]interface{}) error {
	// 记录回调
	rawData, _ := json.Marshal(data)

	// 从data中提取paymentNo或其他标识
	paymentNo, ok := data["payment_no"].(string)
	if !ok {
		return fmt.Errorf("回调数据中缺少payment_no")
	}

	payment, err := s.GetPayment(ctx, paymentNo)
	if err != nil {
		return err
	}

	callback := &model.PaymentCallback{
		PaymentID:   payment.ID,
		Channel:     channel,
		Event:       "payment_callback",
		RawData:     string(rawData),
		IsVerified:  false,
		IsProcessed: false,
	}

	// 保存回调记录
	if err := s.paymentRepo.CreateCallback(ctx, callback); err != nil {
		return fmt.Errorf("保存回调记录失败: %w", err)
	}

	// TODO: 验证回调签名
	// TODO: 根据渠道处理回调逻辑
	// TODO: 更新支付状态

	return nil
}

// CreateRefund 创建退款
func (s *paymentService) CreateRefund(ctx context.Context, input *CreateRefundInput) (*model.Refund, error) {
	// 获取原支付记录
	payment, err := s.GetPayment(ctx, input.PaymentNo)
	if err != nil {
		return nil, err
	}

	// 只有成功的支付才能退款
	if payment.Status != model.PaymentStatusSuccess {
		return nil, fmt.Errorf("只有成功的支付才能退款")
	}

	// 验证退款金额
	if input.Amount > payment.Amount {
		return nil, fmt.Errorf("退款金额不能大于支付金额")
	}

	// 检查已退款总额
	// TODO: 计算已退款总额，确保不超过支付金额

	// 生成退款单号
	refundNo := s.generateRefundNo()

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
		return nil, fmt.Errorf("创建退款记录失败: %w", err)
	}

	// TODO: 调用渠道API执行退款

	return refund, nil
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

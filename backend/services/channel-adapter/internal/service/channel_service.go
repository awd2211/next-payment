package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/channel-adapter/internal/adapter"
	"github.com/payment-platform/services/channel-adapter/internal/model"
	"github.com/payment-platform/services/channel-adapter/internal/repository"
)

// ChannelService 渠道服务接口
type ChannelService interface {
	// 支付操作
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)
	QueryPayment(ctx context.Context, paymentNo string) (*QueryPaymentResponse, error)
	CancelPayment(ctx context.Context, paymentNo string) error

	// 退款操作
	CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error)
	QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error)

	// Webhook 处理
	HandleWebhook(ctx context.Context, channel string, signature string, body []byte, headers map[string]string) error
	ProcessPendingWebhooks(ctx context.Context) error

	// 渠道配置管理
	GetChannelConfig(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error)
	ListChannelConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error)
}

type channelService struct {
	repo           repository.ChannelRepository
	adapterFactory *adapter.AdapterFactory
}

// NewChannelService 创建渠道服务实例
func NewChannelService(repo repository.ChannelRepository, factory *adapter.AdapterFactory) ChannelService {
	return &channelService{
		repo:           repo,
		adapterFactory: factory,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	MerchantID    uuid.UUID              `json:"merchant_id"`
	Channel       string                 `json:"channel"`
	PaymentNo     string                 `json:"payment_no"`
	OrderNo       string                 `json:"order_no"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	CustomerEmail string                 `json:"customer_email"`
	CustomerName  string                 `json:"customer_name"`
	Description   string                 `json:"description"`
	SuccessURL    string                 `json:"success_url"`
	CancelURL     string                 `json:"cancel_url"`
	CallbackURL   string                 `json:"callback_url"`
	Extra         map[string]interface{} `json:"extra"`
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	PaymentNo      string                 `json:"payment_no"`
	ChannelTradeNo string                 `json:"channel_trade_no"`
	ClientSecret   string                 `json:"client_secret,omitempty"`
	PaymentURL     string                 `json:"payment_url,omitempty"`
	QRCodeURL      string                 `json:"qr_code_url,omitempty"`
	Status         string                 `json:"status"`
	Extra          map[string]interface{} `json:"extra"`
}

// QueryPaymentResponse 查询支付响应
type QueryPaymentResponse struct {
	PaymentNo            string                 `json:"payment_no"`
	ChannelTradeNo       string                 `json:"channel_trade_no"`
	Status               string                 `json:"status"`
	Amount               int64                  `json:"amount"`
	Currency             string                 `json:"currency"`
	PaymentMethod        string                 `json:"payment_method"`
	PaymentMethodDetails map[string]interface{} `json:"payment_method_details"`
	PaidAt               *time.Time             `json:"paid_at"`
}

// CreateRefundRequest 创建退款请求
type CreateRefundRequest struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	RefundNo   string    `json:"refund_no"`
	PaymentNo  string    `json:"payment_no"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Reason     string    `json:"reason"`
}

// CreateRefundResponse 创建退款响应
type CreateRefundResponse struct {
	RefundNo        string `json:"refund_no"`
	ChannelRefundNo string `json:"channel_refund_no"`
	Status          string `json:"status"`
}

// QueryRefundResponse 查询退款响应
type QueryRefundResponse struct {
	RefundNo        string     `json:"refund_no"`
	ChannelRefundNo string     `json:"channel_refund_no"`
	Status          string     `json:"status"`
	Amount          int64      `json:"amount"`
	Currency        string     `json:"currency"`
	RefundedAt      *time.Time `json:"refunded_at"`
}

// CreatePayment 创建支付
func (s *channelService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 获取渠道配置
	config, err := s.repo.GetConfig(ctx, req.MerchantID, req.Channel)
	if err != nil {
		return nil, fmt.Errorf("获取渠道配置失败: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("渠道配置不存在或未启用")
	}

	// 获取适配器
	adpt, ok := s.adapterFactory.GetAdapter(req.Channel)
	if !ok {
		return nil, fmt.Errorf("不支持的支付渠道: %s", req.Channel)
	}

	// 创建支付请求
	adapterReq := &adapter.CreatePaymentRequest{
		PaymentNo:     req.PaymentNo,
		OrderNo:       req.OrderNo,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CustomerEmail: req.CustomerEmail,
		CustomerName:  req.CustomerName,
		Description:   req.Description,
		SuccessURL:    req.SuccessURL,
		CancelURL:     req.CancelURL,
		CallbackURL:   req.CallbackURL,
		Extra:         req.Extra,
	}

	// 调用适配器创建支付
	adapterResp, err := adpt.CreatePayment(ctx, adapterReq)
	if err != nil {
		// 记录失败的交易
		s.createFailedTransaction(ctx, req, "", err.Error())
		return nil, fmt.Errorf("创建支付失败: %w", err)
	}

	// 保存交易记录
	tx := &model.Transaction{
		MerchantID:     req.MerchantID,
		OrderNo:        req.OrderNo,
		PaymentNo:      req.PaymentNo,
		Channel:        req.Channel,
		ChannelTradeNo: adapterResp.ChannelTradeNo,
		TransactionType: model.TransactionTypePayment,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         adapterResp.Status,
		CustomerEmail:  req.CustomerEmail,
		CustomerName:   req.CustomerName,
	}

	// 序列化请求和响应数据
	if reqData, err := json.Marshal(req); err == nil {
		tx.RequestData = string(reqData)
	}
	if respData, err := json.Marshal(adapterResp); err == nil {
		tx.ResponseData = string(respData)
	}

	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		// 记录日志，但不影响支付创建
		fmt.Printf("保存交易记录失败: %v\n", err)
	}

	// 构造响应
	response := &CreatePaymentResponse{
		PaymentNo:      req.PaymentNo,
		ChannelTradeNo: adapterResp.ChannelTradeNo,
		ClientSecret:   adapterResp.ClientSecret,
		PaymentURL:     adapterResp.PaymentURL,
		QRCodeURL:      adapterResp.QRCodeURL,
		Status:         adapterResp.Status,
		Extra:          adapterResp.Extra,
	}

	return response, nil
}

// QueryPayment 查询支付状态
func (s *channelService) QueryPayment(ctx context.Context, paymentNo string) (*QueryPaymentResponse, error) {
	// 获取交易记录
	tx, err := s.repo.GetTransaction(ctx, paymentNo)
	if err != nil {
		return nil, fmt.Errorf("获取交易记录失败: %w", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("交易记录不存在")
	}

	// 获取适配器
	adpt, ok := s.adapterFactory.GetAdapter(tx.Channel)
	if !ok {
		return nil, fmt.Errorf("不支持的支付渠道: %s", tx.Channel)
	}

	// 查询支付状态
	adapterResp, err := adpt.QueryPayment(ctx, tx.ChannelTradeNo)
	if err != nil {
		return nil, fmt.Errorf("查询支付状态失败: %w", err)
	}

	// 更新交易记录
	if adapterResp.Status != tx.Status {
		tx.Status = adapterResp.Status
		if adapterResp.PaidAt != nil {
			paidAt := time.Unix(*adapterResp.PaidAt, 0)
			tx.ProcessedAt = &paidAt
		}
		s.repo.UpdateTransaction(ctx, tx)
	}

	// 构造响应
	var paidAt *time.Time
	if adapterResp.PaidAt != nil {
		t := time.Unix(*adapterResp.PaidAt, 0)
		paidAt = &t
	}

	response := &QueryPaymentResponse{
		PaymentNo:            paymentNo,
		ChannelTradeNo:       adapterResp.ChannelTradeNo,
		Status:               adapterResp.Status,
		Amount:               adapterResp.Amount,
		Currency:             adapterResp.Currency,
		PaymentMethod:        adapterResp.PaymentMethod,
		PaymentMethodDetails: adapterResp.PaymentMethodDetails,
		PaidAt:               paidAt,
	}

	return response, nil
}

// CancelPayment 取消支付
func (s *channelService) CancelPayment(ctx context.Context, paymentNo string) error {
	// 获取交易记录
	tx, err := s.repo.GetTransaction(ctx, paymentNo)
	if err != nil {
		return fmt.Errorf("获取交易记录失败: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("交易记录不存在")
	}

	// 获取适配器
	adpt, ok := s.adapterFactory.GetAdapter(tx.Channel)
	if !ok {
		return fmt.Errorf("不支持的支付渠道: %s", tx.Channel)
	}

	// 取消支付
	if err := adpt.CancelPayment(ctx, tx.ChannelTradeNo); err != nil {
		return fmt.Errorf("取消支付失败: %w", err)
	}

	// 更新交易状态
	tx.Status = model.TransactionStatusCancelled
	return s.repo.UpdateTransaction(ctx, tx)
}

// CreateRefund 创建退款
func (s *channelService) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*CreateRefundResponse, error) {
	// 获取原交易记录
	tx, err := s.repo.GetTransaction(ctx, req.PaymentNo)
	if err != nil {
		return nil, fmt.Errorf("获取原交易记录失败: %w", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("原交易记录不存在")
	}

	// 获取适配器
	adpt, ok := s.adapterFactory.GetAdapter(tx.Channel)
	if !ok {
		return nil, fmt.Errorf("不支持的支付渠道: %s", tx.Channel)
	}

	// 创建退款请求
	adapterReq := &adapter.CreateRefundRequest{
		RefundNo:       req.RefundNo,
		PaymentNo:      req.PaymentNo,
		ChannelTradeNo: tx.ChannelTradeNo,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Reason:         req.Reason,
	}

	// 调用适配器创建退款
	adapterResp, err := adpt.CreateRefund(ctx, adapterReq)
	if err != nil {
		return nil, fmt.Errorf("创建退款失败: %w", err)
	}

	// 保存退款交易记录
	refundTx := &model.Transaction{
		MerchantID:      req.MerchantID,
		OrderNo:         tx.OrderNo,
		PaymentNo:       req.RefundNo, // 使用退款号作为支付流水号
		Channel:         tx.Channel,
		ChannelTradeNo:  adapterResp.ChannelRefundNo,
		TransactionType: model.TransactionTypeRefund,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          adapterResp.Status,
	}

	if err := s.repo.CreateTransaction(ctx, refundTx); err != nil {
		fmt.Printf("保存退款交易记录失败: %v\n", err)
	}

	// 构造响应
	response := &CreateRefundResponse{
		RefundNo:        req.RefundNo,
		ChannelRefundNo: adapterResp.ChannelRefundNo,
		Status:          adapterResp.Status,
	}

	return response, nil
}

// QueryRefund 查询退款状态
func (s *channelService) QueryRefund(ctx context.Context, refundNo string) (*QueryRefundResponse, error) {
	// 获取退款交易记录
	tx, err := s.repo.GetTransaction(ctx, refundNo)
	if err != nil {
		return nil, fmt.Errorf("获取退款记录失败: %w", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("退款记录不存在")
	}

	// 获取适配器
	adpt, ok := s.adapterFactory.GetAdapter(tx.Channel)
	if !ok {
		return nil, fmt.Errorf("不支持的支付渠道: %s", tx.Channel)
	}

	// 查询退款状态
	adapterResp, err := adpt.QueryRefund(ctx, tx.ChannelTradeNo)
	if err != nil {
		return nil, fmt.Errorf("查询退款状态失败: %w", err)
	}

	// 构造响应
	var refundedAt *time.Time
	if adapterResp.RefundedAt != nil {
		t := time.Unix(*adapterResp.RefundedAt, 0)
		refundedAt = &t
	}

	response := &QueryRefundResponse{
		RefundNo:        refundNo,
		ChannelRefundNo: adapterResp.ChannelRefundNo,
		Status:          adapterResp.Status,
		Amount:          adapterResp.Amount,
		Currency:        adapterResp.Currency,
		RefundedAt:      refundedAt,
	}

	return response, nil
}

// HandleWebhook 处理 Webhook 回调
func (s *channelService) HandleWebhook(ctx context.Context, channel string, signature string, body []byte, headers map[string]string) error {
	// 获取适配器
	adpt, ok := s.adapterFactory.GetAdapter(channel)
	if !ok {
		return fmt.Errorf("不支持的支付渠道: %s", channel)
	}

	// 验证签名
	verified, err := adpt.VerifyWebhook(ctx, signature, body)
	if err != nil || !verified {
		return fmt.Errorf("Webhook 签名验证失败: %w", err)
	}

	// 解析 Webhook 数据
	event, err := adpt.ParseWebhook(ctx, body)
	if err != nil {
		return fmt.Errorf("解析 Webhook 数据失败: %w", err)
	}

	// 保存 Webhook 日志
	headersJSON, _ := json.Marshal(headers)
	log := &model.WebhookLog{
		Channel:        channel,
		EventID:        event.EventID,
		EventType:      event.EventType,
		PaymentNo:      event.PaymentNo,
		Signature:      signature,
		IsVerified:     true,
		IsProcessed:    false,
		RequestBody:    string(body),
		RequestHeaders: string(headersJSON),
	}

	// 检查是否已处理过该事件（幂等性）
	existingLog, _ := s.repo.GetWebhookLog(ctx, event.EventID)
	if existingLog != nil && existingLog.IsProcessed {
		return nil // 已处理，直接返回
	}

	if err := s.repo.CreateWebhookLog(ctx, log); err != nil {
		fmt.Printf("保存 Webhook 日志失败: %v\n", err)
	}

	// 处理 Webhook 事件
	if err := s.processWebhookEvent(ctx, event); err != nil {
		log.ProcessResult = err.Error()
		s.repo.UpdateWebhookLog(ctx, log)
		return err
	}

	// 标记为已处理
	now := time.Now()
	log.IsProcessed = true
	log.ProcessedAt = &now
	log.ProcessResult = "success"
	s.repo.UpdateWebhookLog(ctx, log)

	return nil
}

// processWebhookEvent 处理 Webhook 事件
func (s *channelService) processWebhookEvent(ctx context.Context, event *adapter.WebhookEvent) error {
	// 获取交易记录
	tx, err := s.repo.GetTransactionByChannelTradeNo(ctx, event.ChannelTradeNo)
	if err != nil {
		return fmt.Errorf("获取交易记录失败: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("交易记录不存在")
	}

	// 更新交易状态
	tx.Status = event.Status
	webhookData, _ := json.Marshal(event)
	tx.WebhookData = string(webhookData)
	now := time.Now()
	tx.ProcessedAt = &now

	return s.repo.UpdateTransaction(ctx, tx)
}

// ProcessPendingWebhooks 处理待处理的 Webhook
func (s *channelService) ProcessPendingWebhooks(ctx context.Context) error {
	// 获取未处理的 Webhook 列表
	logs, err := s.repo.ListUnprocessedWebhooks(ctx, 100)
	if err != nil {
		return err
	}

	// 逐个处理
	for _, log := range logs {
		// 解析事件数据
		var event adapter.WebhookEvent
		if err := json.Unmarshal([]byte(log.RequestBody), &event); err != nil {
			continue
		}

		// 处理事件
		if err := s.processWebhookEvent(ctx, &event); err != nil {
			log.RetryCount++
			log.ProcessResult = err.Error()
		} else {
			now := time.Now()
			log.IsProcessed = true
			log.ProcessedAt = &now
			log.ProcessResult = "success"
		}

		s.repo.UpdateWebhookLog(ctx, log)
	}

	return nil
}

// GetChannelConfig 获取渠道配置
func (s *channelService) GetChannelConfig(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error) {
	return s.repo.GetConfig(ctx, merchantID, channel)
}

// ListChannelConfigs 列出渠道配置
func (s *channelService) ListChannelConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error) {
	return s.repo.ListConfigs(ctx, merchantID)
}

// createFailedTransaction 创建失败的交易记录
func (s *channelService) createFailedTransaction(ctx context.Context, req *CreatePaymentRequest, channelTradeNo, errorMsg string) {
	tx := &model.Transaction{
		MerchantID:     req.MerchantID,
		OrderNo:        req.OrderNo,
		PaymentNo:      req.PaymentNo,
		Channel:        req.Channel,
		ChannelTradeNo: channelTradeNo,
		TransactionType: model.TransactionTypePayment,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         model.TransactionStatusFailed,
		CustomerEmail:  req.CustomerEmail,
		CustomerName:   req.CustomerName,
		ErrorMessage:   errorMsg,
	}
	s.repo.CreateTransaction(ctx, tx)
}

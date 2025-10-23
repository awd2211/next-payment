package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/notification-service/internal/model"
	"github.com/payment-platform/services/notification-service/internal/provider"
	"github.com/payment-platform/services/notification-service/internal/repository"
)

// NotificationService 通知服务接口
type NotificationService interface {
	// 邮件通知
	SendEmail(ctx context.Context, req *SendEmailRequest) error
	// 短信通知
	SendSMS(ctx context.Context, req *SendSMSRequest) error
	// Webhook 通知
	SendWebhook(ctx context.Context, req *SendWebhookRequest) error

	// 使用模板发送
	SendEmailByTemplate(ctx context.Context, req *SendEmailByTemplateRequest) error

	// 查询通知
	GetNotification(ctx context.Context, id uuid.UUID) (*model.Notification, error)
	ListNotifications(ctx context.Context, query *repository.NotificationQuery) ([]*model.Notification, int64, error)

	// 模板管理
	CreateTemplate(ctx context.Context, template *model.NotificationTemplate) error
	GetTemplate(ctx context.Context, code string, merchantID *uuid.UUID) (*model.NotificationTemplate, error)
	ListTemplates(ctx context.Context, merchantID *uuid.UUID) ([]*model.NotificationTemplate, error)

	// Webhook 端点管理
	CreateWebhookEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error
	ListWebhookEndpoints(ctx context.Context, merchantID uuid.UUID) ([]*model.WebhookEndpoint, error)
	UpdateWebhookEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error
	DeleteWebhookEndpoint(ctx context.Context, id uuid.UUID) error

	// 后台处理
	ProcessPendingNotifications(ctx context.Context) error
	ProcessPendingWebhookDeliveries(ctx context.Context) error
}

type notificationService struct {
	repo             repository.NotificationRepository
	emailFactory     *provider.EmailProviderFactory
	smsFactory       *provider.SMSProviderFactory
	webhookProvider  *provider.WebhookProvider
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(
	repo repository.NotificationRepository,
	emailFactory *provider.EmailProviderFactory,
	smsFactory *provider.SMSProviderFactory,
	webhookProvider *provider.WebhookProvider,
) NotificationService {
	return &notificationService{
		repo:            repo,
		emailFactory:    emailFactory,
		smsFactory:      smsFactory,
		webhookProvider: webhookProvider,
	}
}

// SendEmailRequest 发送邮件请求
type SendEmailRequest struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	To         []string  `json:"to"`
	Subject    string    `json:"subject"`
	HTMLBody   string    `json:"html_body"`
	TextBody   string    `json:"text_body"`
	Provider   string    `json:"provider"` // smtp, mailgun
	Priority   int       `json:"priority"`
}

// SendSMSRequest 发送短信请求
type SendSMSRequest struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	To         string    `json:"to"`
	Content    string    `json:"content"`
	Provider   string    `json:"provider"` // twilio, aliyun
	Priority   int       `json:"priority"`
}

// SendWebhookRequest 发送 Webhook 请求
type SendWebhookRequest struct {
	MerchantID uuid.UUID              `json:"merchant_id"`
	EventType  string                 `json:"event_type"`
	EventID    string                 `json:"event_id"`
	Data       map[string]interface{} `json:"data"`
}

// SendEmailByTemplateRequest 使用模板发送邮件请求
type SendEmailByTemplateRequest struct {
	MerchantID   uuid.UUID              `json:"merchant_id"`
	To           []string               `json:"to"`
	TemplateCode string                 `json:"template_code"`
	TemplateData map[string]interface{} `json:"template_data"`
	Provider     string                 `json:"provider"`
	Priority     int                    `json:"priority"`
}

// SendEmail 发送邮件
func (s *notificationService) SendEmail(ctx context.Context, req *SendEmailRequest) error {
	// 获取邮件提供商
	emailProvider, ok := s.emailFactory.GetProvider(req.Provider)
	if !ok {
		return fmt.Errorf("不支持的邮件提供商: %s", req.Provider)
	}

	// 创建通知记录
	notification := &model.Notification{
		MerchantID: req.MerchantID,
		Type:       model.NotificationTypeSystem,
		Channel:    model.ChannelEmail,
		Recipient:  strings.Join(req.To, ","),
		Subject:    req.Subject,
		Content:    req.HTMLBody,
		Status:     model.StatusPending,
		Provider:   req.Provider,
		Priority:   req.Priority,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("创建通知记录失败: %w", err)
	}

	// 更新状态为发送中
	s.repo.UpdateStatus(ctx, notification.ID, model.StatusSending)

	// 发送邮件
	emailReq := &provider.EmailRequest{
		To:       req.To,
		Subject:  req.Subject,
		HTMLBody: req.HTMLBody,
		TextBody: req.TextBody,
	}

	resp, err := emailProvider.Send(ctx, emailReq)
	if err != nil {
		// 更新状态为失败
		notification.Status = model.StatusFailed
		notification.ErrorMessage = err.Error()
		s.repo.Update(ctx, notification)
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	// 更新状态为已发送
	notification.Status = model.StatusSent
	notification.ProviderMsgID = resp.MessageID
	now := time.Now()
	notification.SentAt = &now
	s.repo.Update(ctx, notification)

	return nil
}

// SendSMS 发送短信
func (s *notificationService) SendSMS(ctx context.Context, req *SendSMSRequest) error {
	// 获取短信提供商
	smsProvider, ok := s.smsFactory.GetProvider(req.Provider)
	if !ok {
		return fmt.Errorf("不支持的短信提供商: %s", req.Provider)
	}

	// 创建通知记录
	notification := &model.Notification{
		MerchantID: req.MerchantID,
		Type:       model.NotificationTypeSystem,
		Channel:    model.ChannelSMS,
		Recipient:  req.To,
		Content:    req.Content,
		Status:     model.StatusPending,
		Provider:   req.Provider,
		Priority:   req.Priority,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("创建通知记录失败: %w", err)
	}

	// 更新状态为发送中
	s.repo.UpdateStatus(ctx, notification.ID, model.StatusSending)

	// 发送短信
	smsReq := &provider.SMSRequest{
		To:      req.To,
		Content: req.Content,
	}

	resp, err := smsProvider.Send(ctx, smsReq)
	if err != nil {
		// 更新状态为失败
		notification.Status = model.StatusFailed
		notification.ErrorMessage = err.Error()
		s.repo.Update(ctx, notification)
		return fmt.Errorf("发送短信失败: %w", err)
	}

	// 更新状态为已发送
	notification.Status = model.StatusSent
	notification.ProviderMsgID = resp.MessageID
	now := time.Now()
	notification.SentAt = &now
	s.repo.Update(ctx, notification)

	return nil
}

// SendWebhook 发送 Webhook
func (s *notificationService) SendWebhook(ctx context.Context, req *SendWebhookRequest) error {
	// 获取商户的 Webhook 端点列表
	endpoints, err := s.repo.ListEndpoints(ctx, req.MerchantID)
	if err != nil {
		return fmt.Errorf("获取 Webhook 端点失败: %w", err)
	}

	// 遍历端点，发送 Webhook
	for _, endpoint := range endpoints {
		if !endpoint.IsEnabled {
			continue
		}

		// 检查端点是否订阅了该事件
		var events []string
		if endpoint.Events != "" {
			json.Unmarshal([]byte(endpoint.Events), &events)
		}

		subscribed := false
		for _, event := range events {
			if event == req.EventType || event == "*" {
				subscribed = true
				break
			}
		}

		if !subscribed {
			continue
		}

		// 创建投递记录
		delivery := &model.WebhookDelivery{
			EndpointID: endpoint.ID,
			MerchantID: req.MerchantID,
			EventType:  req.EventType,
			EventID:    req.EventID,
			Status:     model.DeliveryStatusPending,
		}

		payloadBytes, _ := json.Marshal(req.Data)
		delivery.Payload = string(payloadBytes)

		if err := s.repo.CreateDelivery(ctx, delivery); err != nil {
			fmt.Printf("创建投递记录失败: %v\n", err)
			continue
		}

		// 异步发送 Webhook
		go s.deliverWebhook(context.Background(), delivery, endpoint)
	}

	return nil
}

// deliverWebhook 投递 Webhook
func (s *notificationService) deliverWebhook(ctx context.Context, delivery *model.WebhookDelivery, endpoint *model.WebhookEndpoint) {
	// 解析数据
	var data map[string]interface{}
	json.Unmarshal([]byte(delivery.Payload), &data)

	// 构造 Webhook 请求
	webhookReq := &provider.WebhookRequest{
		URL:       endpoint.URL,
		Secret:    endpoint.Secret,
		EventType: delivery.EventType,
		EventID:   delivery.EventID,
		Timestamp: time.Now().Unix(),
		Data:      data,
		Timeout:   endpoint.Timeout,
	}

	// 发送 Webhook
	resp, err := s.webhookProvider.Send(ctx, webhookReq)

	// 更新投递记录
	now := time.Now()
	delivery.DeliveredAt = &now

	if err != nil {
		delivery.Status = model.DeliveryStatusFailed
		delivery.ErrorMessage = err.Error()
		delivery.RetryCount++

		// 计算下次重试时间（指数退避）
		if delivery.RetryCount < endpoint.MaxRetry {
			retryDelay := time.Duration(1<<uint(delivery.RetryCount)) * time.Minute
			nextRetry := time.Now().Add(retryDelay)
			delivery.NextRetryAt = &nextRetry
			delivery.Status = model.DeliveryStatusRetrying
		}
	} else {
		delivery.Status = resp.Status
		delivery.HTTPStatus = resp.HTTPStatus
		delivery.ResponseBody = resp.ResponseBody
		delivery.Duration = int(resp.Duration)
		delivery.ErrorMessage = resp.ErrorMessage
	}

	s.repo.UpdateDelivery(ctx, delivery)
}

// SendEmailByTemplate 使用模板发送邮件
func (s *notificationService) SendEmailByTemplate(ctx context.Context, req *SendEmailByTemplateRequest) error {
	// 获取模板
	template, err := s.repo.GetTemplate(ctx, req.TemplateCode, &req.MerchantID)
	if err != nil {
		return fmt.Errorf("获取模板失败: %w", err)
	}
	if template == nil {
		return fmt.Errorf("模板不存在: %s", req.TemplateCode)
	}

	// 渲染模板
	subject := s.renderTemplate(template.Subject, req.TemplateData)
	content := s.renderTemplate(template.Content, req.TemplateData)

	// 发送邮件
	return s.SendEmail(ctx, &SendEmailRequest{
		MerchantID: req.MerchantID,
		To:         req.To,
		Subject:    subject,
		HTMLBody:   content,
		Provider:   req.Provider,
		Priority:   req.Priority,
	})
}

// renderTemplate 渲染模板
func (s *notificationService) renderTemplate(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// GetNotification 获取通知
func (s *notificationService) GetNotification(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

// ListNotifications 列出通知
func (s *notificationService) ListNotifications(ctx context.Context, query *repository.NotificationQuery) ([]*model.Notification, int64, error) {
	return s.repo.List(ctx, query)
}

// CreateTemplate 创建模板
func (s *notificationService) CreateTemplate(ctx context.Context, template *model.NotificationTemplate) error {
	return s.repo.CreateTemplate(ctx, template)
}

// GetTemplate 获取模板
func (s *notificationService) GetTemplate(ctx context.Context, code string, merchantID *uuid.UUID) (*model.NotificationTemplate, error) {
	return s.repo.GetTemplate(ctx, code, merchantID)
}

// ListTemplates 列出模板
func (s *notificationService) ListTemplates(ctx context.Context, merchantID *uuid.UUID) ([]*model.NotificationTemplate, error) {
	return s.repo.ListTemplates(ctx, merchantID)
}

// CreateWebhookEndpoint 创建 Webhook 端点
func (s *notificationService) CreateWebhookEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error {
	return s.repo.CreateEndpoint(ctx, endpoint)
}

// ListWebhookEndpoints 列出 Webhook 端点
func (s *notificationService) ListWebhookEndpoints(ctx context.Context, merchantID uuid.UUID) ([]*model.WebhookEndpoint, error) {
	return s.repo.ListEndpoints(ctx, merchantID)
}

// UpdateWebhookEndpoint 更新 Webhook 端点
func (s *notificationService) UpdateWebhookEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error {
	return s.repo.UpdateEndpoint(ctx, endpoint)
}

// DeleteWebhookEndpoint 删除 Webhook 端点
func (s *notificationService) DeleteWebhookEndpoint(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteEndpoint(ctx, id)
}

// ProcessPendingNotifications 处理待发送的通知
func (s *notificationService) ProcessPendingNotifications(ctx context.Context) error {
	// 获取待处理的通知（每次处理100条）
	notifications, err := s.repo.ListPendingNotifications(ctx, 100)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		// 根据渠道类型处理
		switch notification.Channel {
		case model.ChannelEmail:
			s.processEmailNotification(ctx, notification)
		case model.ChannelSMS:
			s.processSMSNotification(ctx, notification)
		}
	}

	return nil
}

// processEmailNotification 处理邮件通知
func (s *notificationService) processEmailNotification(ctx context.Context, notification *model.Notification) {
	// 获取邮件提供商
	emailProvider, ok := s.emailFactory.GetProvider(notification.Provider)
	if !ok {
		notification.Status = model.StatusFailed
		notification.ErrorMessage = fmt.Sprintf("不支持的邮件提供商: %s", notification.Provider)
		s.repo.Update(ctx, notification)
		return
	}

	// 更新状态为发送中
	s.repo.UpdateStatus(ctx, notification.ID, model.StatusSending)

	// 发送邮件
	to := strings.Split(notification.Recipient, ",")
	emailReq := &provider.EmailRequest{
		To:       to,
		Subject:  notification.Subject,
		HTMLBody: notification.Content,
	}

	resp, err := emailProvider.Send(ctx, emailReq)
	if err != nil {
		notification.Status = model.StatusFailed
		notification.ErrorMessage = err.Error()
		notification.RetryCount++
	} else {
		notification.Status = model.StatusSent
		notification.ProviderMsgID = resp.MessageID
		now := time.Now()
		notification.SentAt = &now
	}

	s.repo.Update(ctx, notification)
}

// processSMSNotification 处理短信通知
func (s *notificationService) processSMSNotification(ctx context.Context, notification *model.Notification) {
	// 获取短信提供商
	smsProvider, ok := s.smsFactory.GetProvider(notification.Provider)
	if !ok {
		notification.Status = model.StatusFailed
		notification.ErrorMessage = fmt.Sprintf("不支持的短信提供商: %s", notification.Provider)
		s.repo.Update(ctx, notification)
		return
	}

	// 更新状态为发送中
	s.repo.UpdateStatus(ctx, notification.ID, model.StatusSending)

	// 发送短信
	smsReq := &provider.SMSRequest{
		To:      notification.Recipient,
		Content: notification.Content,
	}

	resp, err := smsProvider.Send(ctx, smsReq)
	if err != nil {
		notification.Status = model.StatusFailed
		notification.ErrorMessage = err.Error()
		notification.RetryCount++
	} else {
		notification.Status = model.StatusSent
		notification.ProviderMsgID = resp.MessageID
		now := time.Now()
		notification.SentAt = &now
	}

	s.repo.Update(ctx, notification)
}

// ProcessPendingWebhookDeliveries 处理待投递的 Webhook
func (s *notificationService) ProcessPendingWebhookDeliveries(ctx context.Context) error {
	// 获取待投递的记录（每次处理100条）
	deliveries, err := s.repo.ListPendingDeliveries(ctx, 100)
	if err != nil {
		return err
	}

	for _, delivery := range deliveries {
		// 获取端点配置
		endpoint, err := s.repo.GetEndpoint(ctx, delivery.EndpointID)
		if err != nil || endpoint == nil {
			continue
		}

		// 投递 Webhook
		go s.deliverWebhook(ctx, delivery, endpoint)
	}

	return nil
}

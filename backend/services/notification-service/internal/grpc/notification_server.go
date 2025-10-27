package grpc

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	pb "github.com/payment-platform/proto/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment-platform/notification-service/internal/model"
	"payment-platform/notification-service/internal/repository"
	"payment-platform/notification-service/internal/service"
)

// NotificationServer gRPC服务实现
type NotificationServer struct {
	pb.UnimplementedNotificationServiceServer
	notificationService service.NotificationService
}

// NewNotificationServer 创建gRPC服务实例
func NewNotificationServer(notificationService service.NotificationService) *NotificationServer {
	return &NotificationServer{
		notificationService: notificationService,
	}
}

// SendEmail 发送邮件
func (s *NotificationServer) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	// 参数验证
	if req.To == "" {
		return &pb.SendEmailResponse{
			Success: false,
			Error:   "收件人不能为空",
		}, nil
	}

	if req.Subject == "" {
		return &pb.SendEmailResponse{
			Success: false,
			Error:   "邮件主题不能为空",
		}, nil
	}

	if req.Content == "" {
		return &pb.SendEmailResponse{
			Success: false,
			Error:   "邮件内容不能为空",
		}, nil
	}

	// 如果提供了模板ID，使用模板发送
	if req.TemplateId != "" {
		templateData := make(map[string]interface{})
		if req.TemplateData != nil {
			templateData = req.TemplateData.AsMap()
		}

		// 需要从模板ID解析出merchant_id，这里简化处理，使用默认商户
		merchantID := uuid.New() // 实际应从上下文或请求中获取

		err := s.notificationService.SendEmailByTemplate(ctx, &service.SendEmailByTemplateRequest{
			MerchantID:   merchantID,
			To:           []string{req.To},
			TemplateCode: req.TemplateId,
			TemplateData: templateData,
			Provider:     model.ProviderSMTP,
			Priority:     0,
		})

		if err != nil {
			return &pb.SendEmailResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		return &pb.SendEmailResponse{
			MessageId: uuid.New().String(),
			Success:   true,
		}, nil
	}

	// 普通邮件发送
	merchantID := uuid.New() // 实际应从上下文或请求中获取
	recipients := []string{req.To}
	if len(req.Cc) > 0 {
		recipients = append(recipients, req.Cc...)
	}
	if len(req.Bcc) > 0 {
		recipients = append(recipients, req.Bcc...)
	}

	err := s.notificationService.SendEmail(ctx, &service.SendEmailRequest{
		MerchantID: merchantID,
		To:         recipients,
		Subject:    req.Subject,
		HTMLBody:   req.Content,
		TextBody:   req.Content,
		Provider:   model.ProviderSMTP,
		Priority:   0,
	})

	if err != nil {
		return &pb.SendEmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SendEmailResponse{
		MessageId: uuid.New().String(),
		Success:   true,
	}, nil
}

// GetEmailTemplate 获取邮件模板
func (s *NotificationServer) GetEmailTemplate(ctx context.Context, req *pb.GetEmailTemplateRequest) (*pb.EmailTemplateResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "模板ID不能为空")
	}

	// 使用模板code查询（简化处理，实际可能需要UUID）
	template, err := s.notificationService.GetTemplate(ctx, req.Id, nil)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "模板不存在: %v", err)
	}

	if template == nil {
		return nil, status.Errorf(codes.NotFound, "模板不存在")
	}

	return &pb.EmailTemplateResponse{
		Template: &pb.EmailTemplate{
			Id:        template.ID.String(),
			Name:      template.Name,
			Subject:   template.Subject,
			Content:   template.Content,
			Language:  "zh-CN", // 默认语言
			CreatedAt: timestamppb.New(template.CreatedAt),
		},
	}, nil
}

// UpdateEmailTemplate 更新邮件模板
func (s *NotificationServer) UpdateEmailTemplate(ctx context.Context, req *pb.UpdateEmailTemplateRequest) (*pb.EmailTemplateResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "模板ID不能为空")
	}

	// 解析模板ID
	templateID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的模板ID")
	}

	// 构造更新对象
	template := &model.NotificationTemplate{
		ID:      templateID,
		Subject: req.Subject,
		Content: req.Content,
	}

	err = s.notificationService.UpdateTemplate(ctx, template)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新模板失败: %v", err)
	}

	// 重新获取更新后的模板
	updatedTemplate, err := s.notificationService.GetTemplate(ctx, req.Id, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取更新后的模板失败: %v", err)
	}

	return &pb.EmailTemplateResponse{
		Template: &pb.EmailTemplate{
			Id:        updatedTemplate.ID.String(),
			Name:      updatedTemplate.Name,
			Subject:   updatedTemplate.Subject,
			Content:   updatedTemplate.Content,
			Language:  "zh-CN",
			CreatedAt: timestamppb.New(updatedTemplate.CreatedAt),
		},
	}, nil
}

// SendWebhook 发送Webhook
func (s *NotificationServer) SendWebhook(ctx context.Context, req *pb.SendWebhookRequest) (*pb.SendWebhookResponse, error) {
	if req.Url == "" {
		return &pb.SendWebhookResponse{
			Success: false,
			Error:   "Webhook URL不能为空",
		}, nil
	}

	if req.Payload == nil {
		return &pb.SendWebhookResponse{
			Success: false,
			Error:   "Payload不能为空",
		}, nil
	}

	// 转换payload
	payloadData := req.Payload.AsMap()

	// 构造webhook请求
	merchantID := uuid.New() // 实际应从上下文中获取
	eventType := req.RelatedType
	if eventType == "" {
		eventType = "webhook.custom"
	}

	eventID := req.RelatedId
	if eventID == "" {
		eventID = uuid.New().String()
	}

	webhookReq := &service.SendWebhookRequest{
		MerchantID: merchantID,
		EventType:  eventType,
		EventID:    eventID,
		Data:       payloadData,
	}

	err := s.notificationService.SendWebhook(ctx, webhookReq)
	if err != nil {
		return &pb.SendWebhookResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SendWebhookResponse{
		Success:    true,
		StatusCode: 200,
	}, nil
}

// ListWebhookLogs 列出Webhook投递日志
func (s *NotificationServer) ListWebhookLogs(ctx context.Context, req *pb.ListWebhookLogsRequest) (*pb.ListWebhookLogsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 构建查询条件
	query := &repository.DeliveryQuery{
		EventType: req.RelatedId, // 使用EventType字段
		Status:    req.Status,
		Page:      page,
		PageSize:  pageSize,
	}

	deliveries, total, err := s.notificationService.ListWebhookDeliveries(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	// 转换为proto格式
	logs := make([]*pb.WebhookLog, len(deliveries))
	for i, d := range deliveries {
		logs[i] = &pb.WebhookLog{
			Id:         d.ID.String(),
			Url:        "",              // 需要关联endpoint获取URL
			Method:     "POST",
			Payload:    d.Payload,
			StatusCode: int32(d.HTTPStatus),
			Response:   d.ResponseBody,
			RetryCount: int32(d.RetryCount),
			Status:     d.Status,
			CreatedAt:  timestamppb.New(d.CreatedAt),
		}
	}

	return &pb.ListWebhookLogsResponse{
		Logs:  logs,
		Total: total,
	}, nil
}

// RetryWebhook 重试Webhook
func (s *NotificationServer) RetryWebhook(ctx context.Context, req *pb.RetryWebhookRequest) (*pb.RetryWebhookResponse, error) {
	if req.WebhookLogId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Webhook日志ID不能为空")
	}

	// 解析notification ID
	notificationID, err := uuid.Parse(req.WebhookLogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的Webhook日志ID")
	}

	// 调用服务层重试方法
	err = s.notificationService.RetryWebhookDelivery(ctx, notificationID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "重试失败: %v", err)
	}

	return &pb.RetryWebhookResponse{
		Success: true,
	}, nil
}

// SendSMS 发送短信
func (s *NotificationServer) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	if req.Phone == "" {
		return &pb.SendSMSResponse{
			Success: false,
			Error:   "手机号不能为空",
		}, nil
	}

	if req.Content == "" && req.TemplateId == "" {
		return &pb.SendSMSResponse{
			Success: false,
			Error:   "短信内容或模板ID不能为空",
		}, nil
	}

	merchantID := uuid.New() // 实际应从上下文中获取

	// 如果使用模板
	if req.TemplateId != "" {
		// 使用模板发送暂不支持，降级为普通短信
		content := req.Content
		if content == "" {
			content = "您有新的通知，请查收。"
		}

		err := s.notificationService.SendSMS(ctx, &service.SendSMSRequest{
			MerchantID: merchantID,
			To:         req.Phone,
			Content:    content,
			Provider:   model.ProviderTwilio,
			Priority:   0,
		})

		if err != nil {
			return &pb.SendSMSResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		return &pb.SendSMSResponse{
			MessageId: uuid.New().String(),
			Success:   true,
		}, nil
	}

	// 普通短信发送
	err := s.notificationService.SendSMS(ctx, &service.SendSMSRequest{
		MerchantID: merchantID,
		To:         req.Phone,
		Content:    req.Content,
		Provider:   model.ProviderTwilio,
		Priority:   0,
	})

	if err != nil {
		return &pb.SendSMSResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SendSMSResponse{
		MessageId: uuid.New().String(),
		Success:   true,
	}, nil
}

// ListNotifications 列出通知记录
func (s *NotificationServer) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 构建查询条件
	query := &repository.NotificationQuery{
		Type:     req.Type,
		Status:   req.Status,
		Page:     page,
		PageSize: pageSize,
	}
	// Note: Recipient field doesn't exist in NotificationQuery
	// If needed, should be added to repository query struct

	notifications, total, err := s.notificationService.ListNotifications(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	// 转换为proto格式
	pbNotifications := make([]*pb.Notification, len(notifications))
	for i, n := range notifications {
		var sentAt *timestamppb.Timestamp
		if n.SentAt != nil {
			sentAt = timestamppb.New(*n.SentAt)
		}

		pbNotifications[i] = &pb.Notification{
			Id:        n.ID.String(),
			Type:      n.Type,
			Recipient: n.Recipient,
			Content:   n.Content,
			Status:    n.Status,
			SentAt:    sentAt,
			CreatedAt: timestamppb.New(n.CreatedAt),
		}
	}

	return &pb.ListNotificationsResponse{
		Notifications: pbNotifications,
		Total:         total,
	}, nil
}

// convertMapToJSON 将map转换为JSON字符串
func convertMapToJSON(m map[string]interface{}) string {
	if m == nil {
		return "{}"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// parseJSONToMap 将JSON字符串转换为map
func parseJSONToMap(s string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return make(map[string]interface{})
	}
	return result
}

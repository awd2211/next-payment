package grpc

import (
	"context"

	pb "github.com/payment-platform/proto/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NotificationServer gRPC服务实现
type NotificationServer struct {
	pb.UnimplementedNotificationServiceServer
}

// NewNotificationServer 创建gRPC服务实例
func NewNotificationServer() *NotificationServer {
	return &NotificationServer{}
}

// 所有方法暂时返回未实现
func (s *NotificationServer) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *NotificationServer) GetEmailTemplate(ctx context.Context, req *pb.GetEmailTemplateRequest) (*pb.EmailTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *NotificationServer) UpdateEmailTemplate(ctx context.Context, req *pb.UpdateEmailTemplateRequest) (*pb.EmailTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *NotificationServer) SendWebhook(ctx context.Context, req *pb.SendWebhookRequest) (*pb.SendWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *NotificationServer) ListWebhookLogs(ctx context.Context, req *pb.ListWebhookLogsRequest) (*pb.ListWebhookLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *NotificationServer) RetryWebhook(ctx context.Context, req *pb.RetryWebhookRequest) (*pb.RetryWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *NotificationServer) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

func (s *NotificationServer) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

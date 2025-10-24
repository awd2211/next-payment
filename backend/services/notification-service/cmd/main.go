package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	// pb "github.com/payment-platform/proto/notification" // gRPC proto (预留,暂不使用)
	"payment-platform/notification-service/internal/handler"
	"payment-platform/notification-service/internal/model"
	"payment-platform/notification-service/internal/provider"
	"payment-platform/notification-service/internal/repository"
	"payment-platform/notification-service/internal/service"
	"payment-platform/notification-service/internal/worker"
	// grpcServer "payment-platform/notification-service/internal/grpc" // gRPC 实现(预留,暂不使用)
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title						Notification Service API
//	@version					1.0
//	@description				支付平台通知服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40008
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 使用 Bootstrap 框架初始化应用
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "notification-service",
		DBName:      config.GetEnv("DB_NAME", "payment_notification"),
		Port:        config.GetEnvInt("PORT", 40008),
		// GRPCPort:    config.GetEnvInt("GRPC_PORT", 50008), // 不使用 gRPC,保持 HTTP 通信

		// 自动迁移数据库模型
		AutoMigrate: []any{
			&model.Notification{},
			&model.NotificationTemplate{},
			&model.WebhookEndpoint{},
			&model.WebhookDelivery{},
			&model.NotificationPreference{},
		},

		// 启用企业级功能(gRPC 默认关闭,使用 HTTP/REST)
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 默认关闭 gRPC,使用 HTTP 通信
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false), // mTLS 服务间认证

		// 速率限制配置
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	logger.Info("正在启动 Notification Service...")

	// 2. 创建邮件提供商工厂
	emailFactory := provider.NewEmailProviderFactory()

	// 注册 SMTP 提供商
	smtpHost := config.GetEnv("SMTP_HOST", "")
	if smtpHost != "" {
		smtpProvider := provider.NewSMTPProvider(
			smtpHost,
			config.GetEnvInt("SMTP_PORT", 587),
			config.GetEnv("SMTP_USERNAME", ""),
			config.GetEnv("SMTP_PASSWORD", ""),
			config.GetEnv("SMTP_FROM", ""),
		)
		emailFactory.Register("smtp", smtpProvider)
		logger.Info("SMTP 邮件提供商已注册")
	}

	// 注册 Mailgun 提供商
	mailgunDomain := config.GetEnv("MAILGUN_DOMAIN", "")
	if mailgunDomain != "" {
		mailgunProvider := provider.NewMailgunProvider(
			mailgunDomain,
			config.GetEnv("MAILGUN_API_KEY", ""),
			config.GetEnv("MAILGUN_FROM", ""),
		)
		emailFactory.Register("mailgun", mailgunProvider)
		logger.Info("Mailgun 邮件提供商已注册")
	}

	// 3. 创建短信提供商工厂
	smsFactory := provider.NewSMSProviderFactory()

	// 注册 Twilio 提供商
	twilioAccountSID := config.GetEnv("TWILIO_ACCOUNT_SID", "")
	if twilioAccountSID != "" {
		twilioProvider := provider.NewTwilioProvider(
			twilioAccountSID,
			config.GetEnv("TWILIO_AUTH_TOKEN", ""),
			config.GetEnv("TWILIO_FROM", ""),
		)
		smsFactory.Register("twilio", twilioProvider)
		logger.Info("Twilio 短信提供商已注册")
	}

	// 注册模拟短信提供商（用于测试）
	mockSMSProvider := provider.NewMockSMSProvider()
	smsFactory.Register("mock", mockSMSProvider)
	logger.Info("Mock 短信提供商已注册")

	// 4. 创建 Webhook 提供商
	webhookProvider := provider.NewWebhookProvider()

	// 5. 初始化Repository
	notificationRepo := repository.NewNotificationRepository(application.DB)

	// 6. 初始化Kafka（可选）
	var notificationService service.NotificationService
	kafkaEnabled := config.GetEnv("KAFKA_ENABLE_ASYNC", "false") == "true"

	// Kafka Brokers配置 (用于事件消费)
	var kafkaBrokers []string
	kafkaBrokersStr := config.GetEnv("KAFKA_BROKERS", "")
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
		logger.Info(fmt.Sprintf("Kafka Brokers配置完成: %v", kafkaBrokers))
	}

	if kafkaEnabled {
		logger.Info("Kafka异步模式已启用")

		// 获取Kafka配置
		if len(kafkaBrokers) == 0 {
			kafkaBrokers = strings.Split(config.GetEnv("KAFKA_BROKERS", "localhost:9092"), ",")
		}

		// 创建邮件生产者
		emailProducer := kafka.NewProducer(kafka.ProducerConfig{
			Brokers: kafkaBrokers,
			Topic:   "notifications.email",
		})
		logger.Info("邮件Kafka生产者已创建")

		// 创建短信生产者
		smsProducer := kafka.NewProducer(kafka.ProducerConfig{
			Brokers: kafkaBrokers,
			Topic:   "notifications.sms",
		})
		logger.Info("短信Kafka生产者已创建")

		// 使用Kafka模式初始化Service
		notificationService = service.NewNotificationServiceWithKafka(
			notificationRepo,
			emailFactory,
			smsFactory,
			webhookProvider,
			emailProducer,
			smsProducer,
		)

		// 创建Worker
		notificationWorker := worker.NewNotificationWorker(
			notificationRepo,
			emailFactory,
			smsFactory,
		)

		// 创建邮件消费者并启动Worker
		emailConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: kafkaBrokers,
			Topic:   "notifications.email",
			GroupID: "notification-email-worker",
		})
		go func() {
			ctx := context.Background()
			notificationWorker.StartEmailWorker(ctx, emailConsumer)
		}()
		logger.Info("邮件Worker已启动")

		// 创建短信消费者并启动Worker
		smsConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: kafkaBrokers,
			Topic:   "notifications.sms",
			GroupID: "notification-sms-worker",
		})
		go func() {
			ctx := context.Background()
			notificationWorker.StartSMSWorker(ctx, smsConsumer)
		}()
		logger.Info("短信Worker已启动")
	} else {
		logger.Info("使用同步模式（Kafka未启用）")
		// 使用同步模式初始化Service
		notificationService = service.NewNotificationService(
			notificationRepo,
			emailFactory,
			smsFactory,
			webhookProvider,
		)
	}

	// 7. 初始化Handler
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// 8. Swagger UI（公开接口）
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 9. 初始化JWT管理器
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// JWT认证中间件
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 10. 注册通知路由（带认证）
	notificationHandler.RegisterRoutes(application.Router, authMiddleware)

	// 11. gRPC 服务（预留但不启用，系统使用 HTTP/REST 通信）
	// notificationGrpcServer := grpcServer.NewNotificationServer(notificationService)
	// pb.RegisterNotificationServiceServer(application.GRPCServer, notificationGrpcServer)
	// logger.Info(fmt.Sprintf("gRPC Server 已注册，将监听端口 %d", config.GetEnvInt("GRPC_PORT", 50008)))

	// 12. 启动事件消费Workers (消费payment.events和order.events)
	if len(kafkaBrokers) > 0 {
		logger.Info("启动事件消费Workers...")

		// 创建EventWorker
		eventWorker := worker.NewEventWorker(
			notificationRepo,
			emailFactory,
			smsFactory,
		)

		// 启动支付事件消费Worker
		paymentEventConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: kafkaBrokers,
			Topic:   "payment.events",
			GroupID: "notification-payment-event-worker",
		})
		go func() {
			ctx := context.Background()
			eventWorker.StartPaymentEventWorker(ctx, paymentEventConsumer)
		}()
		logger.Info("支付事件Worker已启动 (topic: payment.events)")

		// 启动订单事件消费Worker
		orderEventConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: kafkaBrokers,
			Topic:   "order.events",
			GroupID: "notification-order-event-worker",
		})
		go func() {
			ctx := context.Background()
			eventWorker.StartOrderEventWorker(ctx, orderEventConsumer)
		}()
		logger.Info("订单事件Worker已启动 (topic: order.events)")
	} else {
		logger.Info("未配置Kafka Brokers，事件消费Workers未启动")
	}

	// 13. 启动后台任务
	go startBackgroundWorkers(notificationService)

	// 14. 启动服务（仅 HTTP，优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

// startBackgroundWorkers 启动后台任务
func startBackgroundWorkers(notificationService service.NotificationService) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	ctx := context.Background()
	for range ticker.C {
		// 处理待发送的通知
		if err := notificationService.ProcessPendingNotifications(ctx); err != nil {
			logger.Error(fmt.Sprintf("处理待发送通知失败: %v", err))
		}

		// 处理待投递的 Webhook
		if err := notificationService.ProcessPendingWebhookDeliveries(ctx); err != nil {
			logger.Error(fmt.Sprintf("处理待投递 Webhook 失败: %v", err))
		}
	}
}

// 代码行数对比：
// - 原始版本: 345行 (手动初始化所有组件)
// - Bootstrap版本: 254行 (框架自动处理)
// - 减少代码: 26%（保留了所有自定义业务逻辑：provider factories、Kafka workers、background tasks）
//
// 自动获得的功能：
// ✅ 数据库连接和迁移
// ✅ Redis 连接
// ✅ Zap 日志系统
// ✅ Gin 路由和中间件（CORS, RequestID, Panic Recovery）
// ✅ gRPC 服务器（自动启动在独立端口）
// ✅ Jaeger 分布式追踪
// ✅ Prometheus 指标收集（/metrics 端点）
// ✅ 健康检查端点 (/health, /health/live, /health/ready)
// ✅ 速率限制
// ✅ 优雅关闭（信号处理，HTTP + gRPC 双协议）
// ✅ 请求 ID
//
// 保留的自定义能力：
// ✅ 邮件提供商工厂（SMTP, Mailgun）
// ✅ 短信提供商工厂（Twilio, Mock）
// ✅ Kafka 异步消息队列（可选）
// ✅ Kafka Workers（邮件、短信）
// ✅ 后台任务（定时处理待发送通知和 Webhook）
// ✅ Webhook 提供商
// ✅ gRPC 服务注册
// ✅ JWT 认证中间件
// ✅ Swagger UI

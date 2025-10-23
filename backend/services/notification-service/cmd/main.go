package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/middleware"
	"github.com/payment-platform/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"payment-platform/notification-service/internal/handler"
	"payment-platform/notification-service/internal/model"
	"payment-platform/notification-service/internal/provider"
	"payment-platform/notification-service/internal/repository"
	"payment-platform/notification-service/internal/service"
	"payment-platform/notification-service/internal/worker"
	grpcServer "payment-platform/notification-service/internal/grpc"
	pb "github.com/payment-platform/proto/notification"
	pkggrpc "github.com/payment-platform/pkg/grpc"
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
	// 初始化日志
	env := config.GetEnv("ENV", "development")
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("正在启动 Notification Service...")

	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_notification"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		TimeZone: config.GetEnv("DB_TIMEZONE", "UTC"),
	}

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		logger.Fatal("数据库连接失败")
		log.Fatalf("Error: %v", err)
	}
	logger.Info("数据库连接成功")

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		&model.Notification{},
		&model.NotificationTemplate{},
		&model.WebhookEndpoint{},
		&model.WebhookDelivery{},
		&model.NotificationPreference{},
	); err != nil {
		logger.Fatal("数据库迁移失败")
		log.Fatalf("Error: %v", err)
	}
	logger.Info("数据库迁移完成")

	// 初始化Redis
	redisConfig := db.RedisConfig{
		Host:     config.GetEnv("REDIS_HOST", "localhost"),
		Port:     config.GetEnvInt("REDIS_PORT", 6379),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       config.GetEnvInt("REDIS_DB", 0),
	}

	redisClient, err := db.NewRedisClient(redisConfig)
	if err != nil {
		logger.Fatal("Redis连接失败")
		log.Fatalf("Error: %v", err)
	}
	logger.Info("Redis连接成功")

	// 初始化 Prometheus 指标
	httpMetrics := metrics.NewHTTPMetrics("notification_service")
	logger.Info("Prometheus 指标初始化完成")

	// 初始化 Jaeger 分布式追踪
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	samplingRate := float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0
	tracerShutdown, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "notification-service",
		ServiceVersion: "1.0.0",
		Environment:    env,
		JaegerEndpoint: jaegerEndpoint,
		SamplingRate:   samplingRate,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Jaeger 初始化失败: %v", err))
	} else {
		logger.Info("Jaeger 追踪初始化完成")
		defer tracerShutdown(context.Background())
	}

	// 创建邮件提供商工厂
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

	// 创建短信提供商工厂
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

	// 创建 Webhook 提供商
	webhookProvider := provider.NewWebhookProvider()

	// 初始化Repository
	notificationRepo := repository.NewNotificationRepository(database)

	// 初始化Kafka（可选）
	var notificationService service.NotificationService
	kafkaEnabled := config.GetEnv("KAFKA_ENABLE_ASYNC", "false") == "true"

	if kafkaEnabled {
		logger.Info("Kafka异步模式已启用")

		// 获取Kafka配置
		kafkaBrokers := strings.Split(config.GetEnv("KAFKA_BROKERS", "localhost:9092"), ",")

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

	// 初始化Handler
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(tracing.TracingMiddleware("notification-service"))
	r.Use(middleware.Logger(logger.Log))
	r.Use(metrics.PrometheusMiddleware(httpMetrics))

	// 限流中间件
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(rateLimiter.RateLimit())

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查（公开接口）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "notification-service",
			"time":    time.Now().Unix(),
		})
	})

	// Swagger UI（公开接口）
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 初始化JWT管理器
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// JWT认证中间件
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 注册通知路由（带认证）
	notificationHandler.RegisterRoutes(r, authMiddleware)

	// 启动 gRPC 服务器（独立 goroutine）
	grpcPort := config.GetEnvInt("GRPC_PORT", 50008)
	gRPCServer := pkggrpc.NewSimpleServer()
	notificationGrpcServer := grpcServer.NewNotificationServer(notificationService)
	pb.RegisterNotificationServiceServer(gRPCServer, notificationGrpcServer)

	go func() {
		logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
		if err := pkggrpc.StartServer(gRPCServer, grpcPort); err != nil {
			logger.Fatal(fmt.Sprintf("gRPC Server 启动失败: %v", err))
		}
	}()

	// 启动后台任务
	go startBackgroundWorkers(notificationService)

	// 启动 HTTP 服务器
	port := config.GetEnvInt("PORT", 40008)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Notification Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务启动失败")
		log.Fatalf("Error: %v", err)
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

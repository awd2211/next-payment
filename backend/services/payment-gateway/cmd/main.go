package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"github.com/payment-platform/pkg/health"
	"github.com/payment-platform/pkg/idempotency"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/middleware"
	"github.com/payment-platform/pkg/saga"
	"github.com/payment-platform/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/handler"
	localMiddleware "payment-platform/payment-gateway/internal/middleware"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
	"payment-platform/payment-gateway/internal/service"
	grpcServer "payment-platform/payment-gateway/internal/grpc"
	pb "github.com/payment-platform/proto/payment"
	pkggrpc "github.com/payment-platform/pkg/grpc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title						Payment Gateway API
//	@version					1.0
//	@description				支付平台支付网关服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40003
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

	logger.Info("正在启动 Payment Gateway Service...")

	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_gateway"),
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
		&model.Payment{},
		&model.Refund{},
		&model.PaymentCallback{},
		&model.PaymentRoute{},
		&saga.Saga{},
		&saga.SagaStep{},
	); err != nil {
		logger.Fatal("数据库迁移失败")
		log.Fatalf("Error: %v", err)
	}
	logger.Info("数据库迁移完成（包含 Saga 表）")

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
	httpMetrics := metrics.NewHTTPMetrics("payment_gateway")
	paymentMetrics := metrics.NewPaymentMetrics("payment_gateway")
	logger.Info("Prometheus 指标初始化完成")

	// 初始化 Jaeger 分布式追踪
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	samplingRate := float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0 // 默认 100% 采样
	tracerShutdown, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "payment-gateway",
		ServiceVersion: "1.0.0",
		Environment:    env,
		JaegerEndpoint: jaegerEndpoint,
		SamplingRate:   samplingRate,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Jaeger 初始化失败: %v", err))
	} else {
		logger.Info("Jaeger 追踪初始化完成")
		defer func() {
			if err := tracerShutdown(context.Background()); err != nil {
				logger.Error(fmt.Sprintf("Jaeger shutdown 失败: %v", err))
			}
		}()
	}

	// 初始化Repository
	paymentRepo := repository.NewPaymentRepository(database)

	// 初始化微服务客户端
	orderServiceURL := config.GetEnv("ORDER_SERVICE_URL", "http://localhost:8004")
	channelServiceURL := config.GetEnv("CHANNEL_SERVICE_URL", "http://localhost:8005")
	riskServiceURL := config.GetEnv("RISK_SERVICE_URL", "http://localhost:8006")

	orderClient := client.NewOrderClient(orderServiceURL)
	channelClient := client.NewChannelClient(channelServiceURL)
	riskClient := client.NewRiskClient(riskServiceURL)

	logger.Info(fmt.Sprintf("Order Service URL: %s", orderServiceURL))
	logger.Info(fmt.Sprintf("Channel Service URL: %s", channelServiceURL))
	logger.Info(fmt.Sprintf("Risk Service URL: %s", riskServiceURL))

	// 初始化Kafka Brokers（可选，如果未配置则为nil）
	var kafkaBrokers []string
	kafkaBrokersStr := config.GetEnv("KAFKA_BROKERS", "")
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
		logger.Info(fmt.Sprintf("Kafka Brokers配置完成: %v", kafkaBrokers))
	} else {
		logger.Info("未配置Kafka，将使用降级模式（打印日志）")
	}

	// 初始化MessageService
	messageService := service.NewMessageService(kafkaBrokers)

	// 初始化 Saga Orchestrator（分布式事务补偿）
	sagaOrchestrator := saga.NewSagaOrchestrator(database, redisClient)
	logger.Info("Saga Orchestrator 初始化完成")

	// 初始化 Saga Payment Service（支付流程 Saga 编排）
	// 注意：Saga 功能暂时可选，未来会集成到 paymentService 中使用
	_ = service.NewSagaPaymentService(
		sagaOrchestrator,
		paymentRepo,
		orderClient,
		channelClient,
	)
	logger.Info("Saga Payment Service 初始化完成（功能已准备就绪）")

	// 初始化Service
	paymentService := service.NewPaymentService(
		database, // 添加 db 参数，用于事务支持
		paymentRepo,
		orderClient,
		channelClient,
		riskClient,
		redisClient,
		paymentMetrics, // 添加 Prometheus 指标
		messageService, // 添加消息服务
	)

	// 初始化Handler
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// 初始化API Key仓储
	apiKeyRepo := repository.NewAPIKeyRepository(database)

	// 初始化签名验证中间件
	signatureMiddleware := localMiddleware.NewSignatureMiddleware(
		func(apiKey string) (*localMiddleware.APIKeyData, error) {
			// 从数据库查询API Key
			ctx := context.Background()
			key, err := apiKeyRepo.GetByAPIKey(ctx, apiKey)
			if err != nil {
				return nil, err
			}

			// 转换为中间件需要的数据结构
			return &localMiddleware.APIKeyData{
				Secret:       key.APISecret,
				MerchantID:   key.MerchantID,
				IsActive:     key.IsActive,
				ExpiresAt:    key.ExpiresAt,
				Environment:  key.Environment,
				IPWhitelist:  key.IPWhitelist,  // IP白名单
				ShouldRotate: key.ShouldRotate(), // 轮换提醒
			}, nil
		},
		redisClient,
	)

	// 设置API Key更新器（用于更新last_used_at）
	signatureMiddleware.SetAPIKeyUpdater(apiKeyRepo)

	// 初始化健康检查器
	healthChecker := health.NewHealthChecker()

	// 注册数据库健康检查
	healthChecker.Register(health.NewDBChecker("database", database))

	// 注册Redis健康检查
	healthChecker.Register(health.NewRedisChecker("redis", redisClient))

	// 注册下游服务健康检查
	if orderServiceURL != "" {
		healthChecker.Register(health.NewServiceHealthChecker("order-service", orderServiceURL))
	}
	if channelServiceURL != "" {
		healthChecker.Register(health.NewServiceHealthChecker("channel-adapter", channelServiceURL))
	}
	if riskServiceURL != "" {
		healthChecker.Register(health.NewServiceHealthChecker("risk-service", riskServiceURL))
	}

	// 创建健康检查处理器
	healthHandler := health.NewGinHandler(healthChecker)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(tracing.TracingMiddleware("payment-gateway"))     // Jaeger 分布式追踪
	r.Use(middleware.Logger(logger.Log))
	r.Use(metrics.PrometheusMiddleware(httpMetrics)) // Prometheus HTTP 指标收集

	// 限流中间件
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(rateLimiter.RateLimit())

	// 幂等性中间件（针对创建操作）
	idempotencyManager := idempotency.NewIdempotencyManager(redisClient, "payment-gateway", 24*time.Hour)
	r.Use(middleware.IdempotencyMiddleware(idempotencyManager))

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点
	r.GET("/health", healthHandler.Handle)           // 完整健康检查
	r.GET("/health/live", healthHandler.HandleLiveness)    // 存活探针（Kubernetes）
	r.GET("/health/ready", healthHandler.HandleReadiness)  // 就绪探针（Kubernetes）

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册支付路由
	// 公开路由（Webhook回调，不需要签名验证）
	webhooks := r.Group("/api/v1/webhooks")
	{
		webhooks.POST("/stripe", paymentHandler.HandleStripeWebhook)
		webhooks.POST("/paypal", paymentHandler.HandlePayPalWebhook)
	}

	// 需要签名验证的路由
	api := r.Group("/api/v1")
	api.Use(signatureMiddleware.Verify())
	{
		// 支付管理
		payments := api.Group("/payments")
		{
			payments.POST("", paymentHandler.CreatePayment)
			payments.GET("/:paymentNo", paymentHandler.GetPayment)
			payments.GET("", paymentHandler.QueryPayments)
			payments.POST("/:paymentNo/cancel", paymentHandler.CancelPayment)
		}

		// 退款管理
		refunds := api.Group("/refunds")
		{
			refunds.POST("", paymentHandler.CreateRefund)
			refunds.GET("/:refundNo", paymentHandler.GetRefund)
			refunds.GET("", paymentHandler.QueryRefunds)
		}
	}

	// 启动 gRPC 服务器（独立 goroutine）
	grpcPort := config.GetEnvInt("GRPC_PORT", 50003)
	gRPCServer := pkggrpc.NewSimpleServer()
	paymentGrpcServer := grpcServer.NewPaymentServer(paymentService)
	pb.RegisterPaymentServiceServer(gRPCServer, paymentGrpcServer)

	go func() {
		logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
		if err := pkggrpc.StartServer(gRPCServer, grpcPort); err != nil {
			logger.Fatal(fmt.Sprintf("gRPC Server 启动失败: %v", err))
		}
	}()

	// 启动 HTTP 服务器
	port := config.GetEnvInt("PORT", 40003)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Payment Gateway Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务启动失败")
		log.Fatalf("Error: %v", err)
	}
}

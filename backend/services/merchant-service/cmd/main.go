package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	pkggrpc "github.com/payment-platform/pkg/grpc"
	"github.com/payment-platform/pkg/health"
	"github.com/payment-platform/pkg/idempotency"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/middleware"
	"github.com/payment-platform/pkg/tracing"
	pb "github.com/payment-platform/proto/merchant"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"payment-platform/merchant-service/internal/client"
	"payment-platform/merchant-service/internal/grpc"
	"payment-platform/merchant-service/internal/handler"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
	"payment-platform/merchant-service/internal/service"
)

//	@title						Merchant Service API
//	@version					1.0
//	@description				支付平台商户管理服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40002
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

	logger.Info("正在启动 Merchant Service...")

	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_merchant"),
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
		&model.Merchant{},
		&model.APIKey{},
		&model.ChannelConfig{},
		// 新增业务模型
		&model.SettlementAccount{},
		&model.KYCDocument{},
		&model.BusinessQualification{},
		&model.MerchantFeeConfig{},
		&model.MerchantUser{},
		&model.MerchantTransactionLimit{},
		&model.MerchantContract{},
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
	httpMetrics := metrics.NewHTTPMetrics("merchant_service")
	logger.Info("Prometheus 指标初始化完成")

	// 初始化 Jaeger 分布式追踪
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	samplingRate := float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0
	tracerShutdown, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "merchant-service",
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

	// 初始化JWT Manager
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// 初始化Repository
	merchantRepo := repository.NewMerchantRepository(database)
	apiKeyRepo := repository.NewAPIKeyRepository(database)
	channelRepo := repository.NewChannelRepository(database)

	// 新增业务Repository
	settlementAccountRepo := repository.NewSettlementAccountRepository(database)
	kycDocRepo := repository.NewKYCDocumentRepository(database)
	feeConfigRepo := repository.NewMerchantFeeConfigRepository(database)
	merchantUserRepo := repository.NewMerchantUserRepository(database)
	transactionLimitRepo := repository.NewMerchantTransactionLimitRepository(database)
	qualificationRepo := repository.NewBusinessQualificationRepository(database)

	// 初始化Service
	merchantService := service.NewMerchantService(database, merchantRepo, apiKeyRepo, jwtManager)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, merchantRepo)
	channelService := service.NewChannelService(channelRepo)

	// 新增业务Service
	businessService := service.NewBusinessService(
		settlementAccountRepo,
		kycDocRepo,
		feeConfigRepo,
		merchantUserRepo,
		transactionLimitRepo,
		qualificationRepo,
		merchantRepo,
	)

	// 初始化HTTP客户端（用于Dashboard聚合）
	analyticsServiceURL := config.GetEnv("ANALYTICS_SERVICE_URL", "http://localhost:40009")
	accountingServiceURL := config.GetEnv("ACCOUNTING_SERVICE_URL", "http://localhost:40007")
	riskServiceURL := config.GetEnv("RISK_SERVICE_URL", "http://localhost:40006")
	notificationServiceURL := config.GetEnv("NOTIFICATION_SERVICE_URL", "http://localhost:40008")
	paymentServiceURL := config.GetEnv("PAYMENT_SERVICE_URL", "http://localhost:40003")

	analyticsClient := client.NewAnalyticsClient(analyticsServiceURL)
	accountingClient := client.NewAccountingClient(accountingServiceURL)
	riskClient := client.NewRiskClient(riskServiceURL)
	notificationClient := client.NewNotificationClient(notificationServiceURL)
	paymentClient := client.NewPaymentClient(paymentServiceURL)

	logger.Info("HTTP客户端初始化完成")

	// Dashboard聚合服务
	dashboardService := service.NewDashboardService(
		analyticsClient,
		accountingClient,
		riskClient,
		notificationClient,
		paymentClient,
	)

	// 初始化gRPC Server（并行启动）
	grpcPort := config.GetEnvInt("GRPC_PORT", 50002)
	grpcServer := pkggrpc.NewSimpleServer()
	merchantGrpcServer := grpc.NewMerchantServer(merchantService, apiKeyService, channelService)
	pb.RegisterMerchantServiceServer(grpcServer, merchantGrpcServer)

	// 在后台启动gRPC服务器
	go func() {
		logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
		if err := pkggrpc.StartServer(grpcServer, grpcPort); err != nil {
			logger.Fatal("gRPC服务启动失败")
			log.Fatalf("Error: %v", err)
		}
	}()

	// 初始化HTTP Handler
	merchantHandler := handler.NewMerchantHandler(merchantService)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService)
	channelHandler := handler.NewChannelHandler(channelService)
	businessHandler := handler.NewBusinessHandler(businessService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	// 初始化健康检查器
	healthChecker := health.NewHealthChecker()
	healthChecker.Register(health.NewDBChecker("database", database))
	healthChecker.Register(health.NewRedisChecker("redis", redisClient))
	healthHandler := health.NewGinHandler(healthChecker)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(tracing.TracingMiddleware("merchant-service"))
	r.Use(middleware.Logger(logger.Log))
	r.Use(metrics.PrometheusMiddleware(httpMetrics))

	// 限流中间件
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(rateLimiter.RateLimit())

	// 幂等性中间件（针对创建操作）
	idempotencyManager := idempotency.NewIdempotencyManager(redisClient, "merchant-service", 24*time.Hour)
	r.Use(middleware.IdempotencyMiddleware(idempotencyManager))

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点
	r.GET("/health", healthHandler.Handle)                     // 完整健康检查
	r.GET("/health/live", healthHandler.HandleLiveness)        // 存活探针
	r.GET("/health/ready", healthHandler.HandleReadiness)      // 就绪探针

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由
	api := r.Group("/api/v1")
	{
		// 认证中间件
		authMiddleware := middleware.AuthMiddleware(jwtManager)

		// 商户路由（不需要authMiddleware，内部处理）
		merchantHandler.RegisterRoutes(api)

		// API密钥路由（不需要authMiddleware，内部处理）
		apiKeyHandler.RegisterRoutes(api)

		// 渠道配置路由
		channelHandler.RegisterRoutes(api, authMiddleware)

		// 业务路由（结算账户、KYC、费率、子账户、限额、资质）
		businessHandler.RegisterRoutes(api, authMiddleware)

		// Dashboard聚合查询路由
		dashboardHandler.RegisterRoutes(api, authMiddleware)
	}

	// 启动服务器
	port := config.GetEnvInt("PORT", 40002)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Merchant Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务启动失败")
		log.Fatalf("Error: %v", err)
	}
}

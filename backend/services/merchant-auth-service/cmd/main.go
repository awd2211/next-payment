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
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/middleware"
	"github.com/payment-platform/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"payment-platform/merchant-auth-service/internal/client"
	"payment-platform/merchant-auth-service/internal/handler"
	"payment-platform/merchant-auth-service/internal/model"
	"payment-platform/merchant-auth-service/internal/repository"
	"payment-platform/merchant-auth-service/internal/service"
	grpcServer "payment-platform/merchant-auth-service/internal/grpc"
	pb "github.com/payment-platform/proto/merchant_auth"
	pkggrpc "github.com/payment-platform/pkg/grpc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title						Merchant Auth Service API
//	@version					1.0
//	@description				支付平台商户认证服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40011
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

	logger.Info("正在启动 Merchant Auth Service...")

	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_merchant_auth"),
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
		&model.TwoFactorAuth{},
		&model.LoginActivity{},
		&model.SecuritySettings{},
		&model.PasswordHistory{},
		&model.Session{},
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
	httpMetrics := metrics.NewHTTPMetrics("merchant_auth_service")
	logger.Info("Prometheus 指标初始化完成")

	// 初始化 Jaeger 分布式追踪
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	samplingRate := float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0
	tracerShutdown, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "merchant-auth-service",
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

	// 初始化 Merchant Service 客户端
	merchantServiceURL := config.GetEnv("MERCHANT_SERVICE_URL", "http://localhost:8002")
	merchantClient := client.NewMerchantClient(merchantServiceURL)
	logger.Info(fmt.Sprintf("Merchant Service 客户端初始化成功: %s", merchantServiceURL))

	// 初始化Repository
	securityRepo := repository.NewSecurityRepository(database)

	// 初始化Service
	securityService := service.NewSecurityService(securityRepo, merchantClient)

	// 初始化Handler
	securityHandler := handler.NewSecurityHandler(securityService)

	// 初始化JWT Manager
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(tracing.TracingMiddleware("merchant-auth-service"))
	r.Use(middleware.Logger(logger.Log))
	r.Use(metrics.PrometheusMiddleware(httpMetrics))

	// 限流中间件
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(rateLimiter.RateLimit())

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "merchant-auth-service",
			"time":    time.Now().Unix(),
		})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由组
	api := r.Group("/api/v1")

	// 注册安全路由（需要认证）
	securityHandler.RegisterRoutes(api, authMiddleware)

	// 启动定时任务：清理过期会话
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			logger.Info("开始清理过期会话...")
			ctx := context.Background()
			if err := securityService.CleanExpiredSessions(ctx); err != nil {
				logger.Error(fmt.Sprintf("清理过期会话失败: %v", err))
			} else {
				logger.Info("过期会话清理完成")
			}
		}
	}()

	// 启动 gRPC 服务器（独立 goroutine）
	grpcPort := config.GetEnvInt("GRPC_PORT", 50011)
	gRPCServer := pkggrpc.NewSimpleServer()
	authGrpcServer := grpcServer.NewMerchantAuthServer(securityService)
	pb.RegisterMerchantAuthServiceServer(gRPCServer, authGrpcServer)

	go func() {
		logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
		if err := pkggrpc.StartServer(gRPCServer, grpcPort); err != nil {
			logger.Fatal(fmt.Sprintf("gRPC Server 启动失败: %v", err))
		}
	}()

	// 启动 HTTP 服务器
	port := config.GetEnvInt("PORT", 40011)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Merchant Auth Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil{
		logger.Fatal("服务启动失败")
		log.Fatalf("Error: %v", err)
	}
}

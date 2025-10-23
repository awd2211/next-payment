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
	"github.com/payment-platform/pkg/email"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/middleware"
	"github.com/payment-platform/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"payment-platform/admin-service/internal/handler"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
	"payment-platform/admin-service/internal/service"
	grpcServer "payment-platform/admin-service/internal/grpc"
	pb "github.com/payment-platform/proto/admin"
	pkggrpc "github.com/payment-platform/pkg/grpc"

	_ "payment-platform/admin-service/api-docs" // Import generated swagger docs
)

//	@title						Admin Service API
//	@version					1.0
//	@description				支付平台管理后台服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.url				http://www.swagger.io/support
//	@contact.email				support@swagger.io
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40001
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer JWT token

func main() {
	// 初始化日志
	env := config.GetEnv("ENV", "development")
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("正在启动 Admin Service...")

	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_admin"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		TimeZone: config.GetEnv("DB_TIMEZONE", "UTC"),
	}

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		&model.Admin{},
		&model.Role{},
		&model.Permission{},
		&model.AdminRole{},
		&model.RolePermission{},
		&model.AuditLog{},
		&model.SystemConfig{},
		&model.MerchantReview{},
		&model.ApprovalFlow{},
	); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
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
		logger.Fatal("Redis连接失败", zap.Error(err))
	}
	logger.Info("Redis连接成功")

	// 初始化 Prometheus 指标
	httpMetrics := metrics.NewHTTPMetrics("admin_service")
	logger.Info("Prometheus 指标初始化完成")

	// 初始化 Jaeger 分布式追踪
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	samplingRate := float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0
	tracerShutdown, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "admin-service",
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

	// 初始化邮件客户端
	// 初始化邮件客户端
	emailClient, err := email.NewClient(&email.Config{
		Provider:     "smtp",
		SMTPHost:     config.GetEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     config.GetEnvInt("SMTP_PORT", 587),
		SMTPUsername: config.GetEnv("SMTP_USERNAME", ""),
		SMTPPassword: config.GetEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     config.GetEnv("SMTP_FROM", "noreply@payment-platform.com"),
		SMTPFromName: config.GetEnv("SMTP_FROM_NAME", "Payment Platform"),
	})
	if err != nil {
		logger.Warn("SMTP 邮件客户端初始化失败，邮件功能将不可用", zap.Error(err))
	}

	// 初始化Repository
	adminRepo := repository.NewAdminRepository(database)
	roleRepo := repository.NewRoleRepository(database)
	permissionRepo := repository.NewPermissionRepository(database)
	auditLogRepo := repository.NewAuditLogRepository(database)
	systemConfigRepo := repository.NewSystemConfigRepository(database)
	securityRepo := repository.NewSecurityRepository(database)
	preferencesRepo := repository.NewPreferencesRepository(database)
	emailTemplateRepo := repository.NewEmailTemplateRepository(database)

	// 初始化Service
	adminService := service.NewAdminService(adminRepo, roleRepo, jwtManager)
	roleService := service.NewRoleService(roleRepo, permissionRepo, adminRepo)
	permissionService := service.NewPermissionService(permissionRepo)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	systemConfigService := service.NewSystemConfigService(systemConfigRepo)
	securityService := service.NewSecurityService(securityRepo, adminRepo)
	preferencesService := service.NewPreferencesService(preferencesRepo)
	emailTemplateService := service.NewEmailTemplateService(emailTemplateRepo, emailClient)

	// 初始化Handler
	adminHandler := handler.NewAdminHandler(adminService)
	roleHandler := handler.NewRoleHandler(roleService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
	auditLogHandler := handler.NewAuditLogHandler(auditLogService)
	systemConfigHandler := handler.NewSystemConfigHandler(systemConfigService)
	securityHandler := handler.NewSecurityHandler(securityService)
	preferencesHandler := handler.NewPreferencesHandler(preferencesService)
	emailTemplateHandler := handler.NewEmailTemplateHandler(emailTemplateService)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(tracing.TracingMiddleware("admin-service"))
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
			"service": "admin-service",
			"time":    time.Now().Unix(),
		})
	})

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由
	api := r.Group("/api/v1")
	{
		// 认证中间件
		authMiddleware := middleware.AuthMiddleware(jwtManager)

		// 注册所有路由
		adminHandler.RegisterRoutes(api, authMiddleware)
		roleHandler.RegisterRoutes(api, authMiddleware)
		permissionHandler.RegisterRoutes(api, authMiddleware)
		auditLogHandler.RegisterRoutes(api, authMiddleware)
		systemConfigHandler.RegisterRoutes(api, authMiddleware)
		securityHandler.RegisterRoutes(api.Group("/security", authMiddleware))
		preferencesHandler.RegisterRoutes(api.Group("/preferences", authMiddleware))
		emailTemplateHandler.RegisterRoutes(api, authMiddleware)
	}

	// 启动 gRPC 服务器（独立 goroutine）
	grpcPort := config.GetEnvInt("GRPC_PORT", 50001)
	gRPCServer := pkggrpc.NewSimpleServer()
	adminGrpcServer := grpcServer.NewAdminServer(adminService)
	pb.RegisterAdminServiceServer(gRPCServer, adminGrpcServer)

	go func() {
		logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
		if err := pkggrpc.StartServer(gRPCServer, grpcPort); err != nil {
			logger.Fatal("gRPC Server 启动失败", zap.Error(err))
		}
	}()

	// 启动 HTTP 服务器
	port := config.GetEnvInt("PORT", 40001)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Admin Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务启动失败", zap.Error(err))
	}
}

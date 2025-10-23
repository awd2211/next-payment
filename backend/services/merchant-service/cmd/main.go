package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
		&model.WebhookConfig{},
		&model.ChannelConfig{},
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

	// 初始化JWT Manager
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// 初始化Repository
	merchantRepo := repository.NewMerchantRepository(database)
	apiKeyRepo := repository.NewAPIKeyRepository(database)
	webhookRepo := repository.NewWebhookRepository(database)
	channelRepo := repository.NewChannelRepository(database)
	securityRepo := repository.NewSecurityRepository(database)

	// 初始化Service
	merchantService := service.NewMerchantService(merchantRepo, apiKeyRepo, securityRepo, jwtManager)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, merchantRepo)
	webhookService := service.NewWebhookService(webhookRepo)
	channelService := service.NewChannelService(channelRepo)
	securityService := service.NewSecurityService(securityRepo, merchantRepo)
	// authService := service.NewAuthService(merchantRepo, securityRepo, jwtManager) // 后续可以用于增强的登录功能

	// 初始化Handler
	merchantHandler := handler.NewMerchantHandler(merchantService)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService)
	webhookHandler := handler.NewWebhookHandler(webhookService)
	channelHandler := handler.NewChannelHandler(channelService)
	securityHandler := handler.NewSecurityHandler(securityService)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger.Log))

	// 限流中间件
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(rateLimiter.RateLimit())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "merchant-service",
			"time":    time.Now().Unix(),
		})
	})

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

		// Webhook配置路由
		webhookHandler.RegisterRoutes(api, authMiddleware)

		// 渠道配置路由
		channelHandler.RegisterRoutes(api, authMiddleware)

		// 安全路由
		securityHandler.RegisterRoutes(api, authMiddleware)
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

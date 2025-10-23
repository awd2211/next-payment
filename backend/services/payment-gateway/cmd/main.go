package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/handler"
	localMiddleware "payment-platform/payment-gateway/internal/middleware"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
	"payment-platform/payment-gateway/internal/service"
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

	// 初始化Service
	paymentService := service.NewPaymentService(
		paymentRepo,
		orderClient,
		channelClient,
		riskClient,
		redisClient,
	)

	// 初始化Handler
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// 初始化签名验证中间件
	signatureMiddleware := localMiddleware.NewSignatureMiddleware(func(apiKey string) (string, error) {
		// TODO: 从数据库或缓存中获取API Secret
		// 暂时返回测试密钥
		return "test-secret-key", nil
	})

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
			"service": "payment-gateway",
			"time":    time.Now().Unix(),
		})
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	})

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

	// 启动服务器
	port := config.GetEnvInt("PORT", 40003)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Payment Gateway Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务启动失败")
		log.Fatalf("Error: %v", err)
	}
}

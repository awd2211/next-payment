package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"

	"payment-platform/dispute-service/internal/client"
	"payment-platform/dispute-service/internal/handler"
	"payment-platform/dispute-service/internal/model"
	"payment-platform/dispute-service/internal/repository"
	"payment-platform/dispute-service/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "payment-platform/dispute-service/docs" // Swagger文档
)

//	@title						Dispute Service API
//	@version					1.0
//	@description				争议处理服务 - 处理支付争议、退单和Chargeback
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40021
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// Use Bootstrap framework for service initialization
	var configClient *configclient.Client
	if config.GetEnv("ENABLE_CONFIG_CLIENT", "false") == "true" {
		clientCfg := configclient.ClientConfig{ServiceName: "dispute-service", Environment: config.GetEnv("ENV", "production"), ConfigURL: config.GetEnv("CONFIG_SERVICE_URL", "http://localhost:40010"), RefreshRate: 30 * time.Second}
		if config.GetEnvBool("CONFIG_CLIENT_MTLS", false) { clientCfg.EnableMTLS = true; clientCfg.TLSCertFile = config.GetEnv("TLS_CERT_FILE", ""); clientCfg.TLSKeyFile = config.GetEnv("TLS_KEY_FILE", ""); clientCfg.TLSCAFile = config.GetEnv("TLS_CA_FILE", "") }
		client, _ := configclient.NewClient(clientCfg)
		if client != nil { configClient = client; defer configClient.Stop() }
	}
	getConfig := func(key, defaultValue string) string { if configClient != nil { if val := configClient.Get(key); val != "" { return val } }; return config.GetEnv(key, defaultValue) }

	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "dispute-service",
		DBName:      config.GetEnv("DB_NAME", "payment_dispute"),
		Port:        config.GetEnvInt("PORT", 40021),
		AutoMigrate: []any{
			&model.Dispute{},
			&model.DisputeEvidence{},
			&model.DisputeTimeline{},
		},

		// Feature flags
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // HTTP-only service
		EnableHealthCheck: true,
		EnableRateLimit:   true,

		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to bootstrap service: %v", err)
	}

	// Initialize repository
	disputeRepo := repository.NewDisputeRepository(application.DB)

	// Initialize Stripe client
	stripeAPIKey := config.GetEnv("STRIPE_API_KEY", "")
	stripeClient := client.NewStripeDisputeClient(stripeAPIKey)

	// Initialize Payment client
	paymentServiceURL := config.GetEnv("PAYMENT_SERVICE_URL", "http://localhost:40003")
	paymentClient := client.NewPaymentClient(paymentServiceURL)

	// Create service
	disputeService := service.NewDisputeService(disputeRepo, stripeClient, paymentClient)

	// Create handlers
	disputeHandler := handler.NewDisputeHandler(disputeService)

	// Create webhook handler
	stripeWebhookSecret := config.GetEnv("STRIPE_WEBHOOK_SECRET", "")
	if stripeWebhookSecret == "" {
		logger.Warn("STRIPE_WEBHOOK_SECRET not set, webhook signature verification disabled")
	}
	webhookHandler := handler.NewWebhookHandler(disputeService, stripeWebhookSecret)

	// JWT 认证中间件
	// ⚠️ 安全要求: JWT_SECRET必须在生产环境中设置，不能使用默认值
	jwtSecret := getConfig("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required and cannot be empty")
	}
	if len(jwtSecret) < 32 {
		logger.Fatal("JWT_SECRET must be at least 32 characters for security",
			zap.Int("current_length", len(jwtSecret)),
			zap.Int("minimum_length", 32))
	}
	logger.Info("JWT_SECRET validation passed", zap.Int("length", len(jwtSecret)))
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// Register Swagger documentation (public access)
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register routes
	api := application.Router.Group("/api/v1")
	disputeHandler.RegisterRoutes(api)
	webhookHandler.RegisterRoutes(api)

	logger.Info("Swagger documentation enabled", zap.String("url", "http://localhost:40021/swagger/index.html"))

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

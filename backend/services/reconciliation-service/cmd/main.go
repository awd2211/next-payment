package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"go.uber.org/zap"

	"payment-platform/reconciliation-service/internal/client"
	"payment-platform/reconciliation-service/internal/downloader"
	"payment-platform/reconciliation-service/internal/handler"
	"payment-platform/reconciliation-service/internal/model"
	"payment-platform/reconciliation-service/internal/report"
	"payment-platform/reconciliation-service/internal/repository"
	"payment-platform/reconciliation-service/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "payment-platform/reconciliation-service/docs" // Swagger文档
)

//	@title						Reconciliation Service API
//	@version					1.0
//	@description				对账服务 - 自动对账、差异检测和报表生成
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40020
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// Use Bootstrap framework for service initialization
	var configClient *configclient.Client
	if config.GetEnv("ENABLE_CONFIG_CLIENT", "false") == "true" {
		clientCfg := configclient.ClientConfig{ServiceName: "reconciliation-service", Environment: config.GetEnv("ENV", "production"), ConfigURL: config.GetEnv("CONFIG_SERVICE_URL", "http://localhost:40010"), RefreshRate: 30 * time.Second}
		if config.GetEnvBool("CONFIG_CLIENT_MTLS", false) { clientCfg.EnableMTLS = true; clientCfg.TLSCertFile = config.GetEnv("TLS_CERT_FILE", ""); clientCfg.TLSKeyFile = config.GetEnv("TLS_KEY_FILE", ""); clientCfg.TLSCAFile = config.GetEnv("TLS_CA_FILE", "") }
		client, _ := configclient.NewClient(clientCfg)
		if client != nil { configClient = client; defer configClient.Stop() }
	}
	getConfig := func(key, defaultValue string) string { if configClient != nil { if val := configClient.Get(key); val != "" { return val } }; return config.GetEnv(key, defaultValue) }

	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "reconciliation-service",
		DBName:      config.GetEnv("DB_NAME", "payment_reconciliation"),
		Port:        config.GetEnvInt("PORT", 40020),
		AutoMigrate: []any{
			&model.ReconciliationTask{},
			&model.ReconciliationRecord{},
			&model.ChannelSettlementFile{},
			&model.ReconciliationDifference{},
			&model.ReconciliationReport{},
			&model.InternalTransaction{},
			&model.ChannelTransaction{},
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
	reconRepo := repository.NewReconciliationRepository(application.DB)

	// Initialize external dependencies
	stripeAPIKey := config.GetEnv("STRIPE_API_KEY", "")
	settlementFilePath := config.GetEnv("SETTLEMENT_FILE_PATH", "/tmp/settlement-files")
	reportPath := config.GetEnv("REPORT_PATH", "/tmp/reports")
	paymentGatewayURL := config.GetEnv("PAYMENT_GATEWAY_URL", "http://localhost:40003")

	// Create downloader
	stripeDownloader := downloader.NewStripeDownloader(stripeAPIKey, reconRepo, settlementFilePath)

	// Create platform data fetcher
	platformClient := client.NewPlatformClient(paymentGatewayURL)

	// Create report generator
	reportGenerator := report.NewPDFGenerator(reconRepo, reportPath)

	// Create service
	reconService := service.NewReconciliationService(
		reconRepo,
		application.DB,
		stripeDownloader,
		platformClient,
		reportGenerator,
	)

	// Create handler
	reconHandler := handler.NewReconciliationHandler(reconService)

	// JWT 认证中间件（优先从配置中心获取）
	// ⚠️ 安全要求: JWT_SECRET必须在生产环境中设置，不能使用默认值
	jwtSecret := getConfig("JWT_SECRET", "")
	if jwtSecret == "" {
		application.Logger.Fatal("JWT_SECRET environment variable is required and cannot be empty")
	}
	if len(jwtSecret) < 32 {
		application.Logger.Fatal("JWT_SECRET must be at least 32 characters for security",
			zap.Int("current_length", len(jwtSecret)),
			zap.Int("minimum_length", 32))
	}
	application.Logger.Info("JWT_SECRET validation passed", zap.Int("length", len(jwtSecret)))
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// Register Swagger documentation (public access)
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register routes
	api := application.Router.Group("/api/v1")
	reconHandler.RegisterRoutes(api)

	application.Logger.Info("Swagger documentation enabled", zap.String("url", "http://localhost:40020/swagger/index.html"))

	// ✅ Automation Infrastructure Ready:
	// - AlertNotifier: Fully implemented (email alerts for differences, critical alerts, daily reports)
	// - DailyScheduler: Implemented in internal/scheduler/daily_scheduler.go
	// - Extended models: ReconciliationDifference, ReconciliationReport, transactions ready
	//
	// To enable automated reconciliation:
	// 1. Uncomment the scheduler initialization code below
	// 2. Set ENABLE_AUTO_RECONCILIATION=true in environment
	// 3. Configure ALERT_EMAIL_RECIPIENTS in environment (comma-separated emails)
	//
	// Example:
	// if config.GetEnv("ENABLE_AUTO_RECONCILIATION", "false") == "true" {
	//     emailClient := email.NewClient(...)
	//     recipients := strings.Split(config.GetEnv("ALERT_EMAIL_RECIPIENTS", ""), ",")
	//     alerter := notifier.NewAlertNotifier(emailClient, application.Logger, recipients)
	//     scheduler := scheduler.NewDailyScheduler(reconService, alerter, application.Logger)
	//     scheduler.Start()
	//     defer scheduler.Stop()
	// }

	application.Logger.Info("Reconciliation service initialized successfully",
		zap.String("version", "1.0.0"),
		zap.Bool("automation_ready", true),
		zap.Bool("automation_enabled", false))

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

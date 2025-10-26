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
)

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
		DBName:      "payment_reconciliation",
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
	jwtSecret := getConfig("JWT_SECRET", "payment-platform-secret-key-2024")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// Register routes
	api := application.Router.Group("/api/v1")
	reconHandler.RegisterRoutes(api)

	// TODO: Automation features (scheduler, detector, notifier) will be integrated later
	// The extended models are ready for automation features:
	// - ReconciliationDifference for detailed difference tracking
	// - ReconciliationReport for daily reports
	// - InternalTransaction and ChannelTransaction for transaction storage

	application.Logger.Info("Reconciliation service initialized successfully",
		zap.String("version", "1.0.0"),
		zap.Bool("automation_enabled", false))

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

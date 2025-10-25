package main

import (
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"

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

	// JWT 认证中间件
	jwtSecret := config.GetEnv("JWT_SECRET", "payment-platform-secret-key-2024")
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

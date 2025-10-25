package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
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

	// Register routes
	api := application.Router.Group("/api/v1")
	reconHandler.RegisterRoutes(api)

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

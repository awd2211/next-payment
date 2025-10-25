package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"

	"payment-platform/dispute-service/internal/client"
	"payment-platform/dispute-service/internal/handler"
	"payment-platform/dispute-service/internal/model"
	"payment-platform/dispute-service/internal/repository"
	"payment-platform/dispute-service/internal/service"
)

func main() {
	// Use Bootstrap framework for service initialization
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "dispute-service",
		DBName:      "payment_dispute",
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

	// Create service
	disputeService := service.NewDisputeService(disputeRepo, stripeClient)

	// Create handler
	disputeHandler := handler.NewDisputeHandler(disputeService)

	// Register routes
	api := application.Router.Group("/api/v1")
	disputeHandler.RegisterRoutes(api)

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

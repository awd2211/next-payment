package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"

	"payment-platform/merchant-limit-service/internal/handler"
	"payment-platform/merchant-limit-service/internal/model"
	"payment-platform/merchant-limit-service/internal/repository"
	"payment-platform/merchant-limit-service/internal/service"
)

func main() {
	// Use Bootstrap framework for service initialization
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "merchant-limit-service",
		DBName:      "payment_merchant_limit",
		Port:        config.GetEnvInt("PORT", 40022),
		AutoMigrate: []any{
			&model.MerchantTier{},
			&model.MerchantLimit{},
			&model.LimitUsageLog{},
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
	limitRepo := repository.NewLimitRepository(application.DB)

	// Create service
	limitService := service.NewLimitService(limitRepo, application.DB)

	// Create handler
	limitHandler := handler.NewLimitHandler(limitService)

	// Register routes
	api := application.Router.Group("/api/v1")
	limitHandler.RegisterRoutes(api)

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

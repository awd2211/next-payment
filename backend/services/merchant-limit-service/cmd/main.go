package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
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

	// JWT 认证中间件
	jwtSecret := config.GetEnv("JWT_SECRET", "payment-platform-secret-key-2024")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// Register routes
	api := application.Router.Group("/api/v1")
	limitHandler.RegisterRoutes(api)

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

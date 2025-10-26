package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"

	"payment-platform/dispute-service/internal/client"
	"payment-platform/dispute-service/internal/handler"
	"payment-platform/dispute-service/internal/model"
	"payment-platform/dispute-service/internal/repository"
	"payment-platform/dispute-service/internal/service"
)

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

	// JWT 认证中间件
	jwtSecret := getConfig("JWT_SECRET", "payment-platform-secret-key-2024")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// Register routes
	api := application.Router.Group("/api/v1")
	disputeHandler.RegisterRoutes(api)

	// Start service with graceful shutdown
	application.RunWithGracefulShutdown()
}

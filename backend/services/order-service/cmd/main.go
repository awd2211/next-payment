package main

import (
	"fmt"
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"payment-platform/order-service/internal/client"
	"payment-platform/order-service/internal/handler"
	"payment-platform/order-service/internal/model"
	"payment-platform/order-service/internal/repository"
	"payment-platform/order-service/internal/service"
	"github.com/payment-platform/pkg/idempotency"
	"github.com/payment-platform/pkg/middleware"
)

func main() {
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "order-service",
		DBName:      config.GetEnv("DB_NAME", "payment_order"),
		Port:        config.GetEnvInt("PORT", 40004),
		AutoMigrate: []any{
			&model.Order{},
			&model.OrderItem{},
			&model.OrderLog{},
			&model.OrderStatistics{},
		},
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false,
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	logger.Info("正在启动 Order Service...")

	// 初始化 HTTP 客户端
	notificationServiceURL := config.GetEnv("NOTIFICATION_SERVICE_URL", "http://localhost:40008")
	notificationClient := client.NewNotificationClient(notificationServiceURL)

	repo := repository.NewOrderRepository(application.DB)
	svc := service.NewOrderService(application.DB, repo, notificationClient)
	handler := handler.NewOrderHandler(svc)

	idempotencyManager := idempotency.NewIdempotencyManager(application.Redis, "order-service", 24*time.Hour)
	application.Router.Use(middleware.IdempotencyMiddleware(idempotencyManager))

	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	handler.RegisterRoutes(application.Router)

	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

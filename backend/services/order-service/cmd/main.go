package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/kafka"
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

	_ "payment-platform/order-service/api-docs" // Import generated swagger docs
)

//	@title						Order Service API
//	@version					1.0
//	@description				支付平台订单服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40004
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

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
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false), // mTLS 服务间认证
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	logger.Info("正在启动 Order Service...")

	// 初始化 Kafka Brokers
	var kafkaBrokers []string
	kafkaBrokersStr := config.GetEnv("KAFKA_BROKERS", "")
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
		logger.Info(fmt.Sprintf("Kafka Brokers配置完成: %v", kafkaBrokers))
	} else {
		logger.Info("未配置Kafka，将使用降级模式")
	}

	// 初始化 EventPublisher
	eventPublisher := kafka.NewEventPublisher(kafkaBrokers)
	logger.Info("EventPublisher 初始化完成")

	// 初始化 HTTP 客户端 (保留作为降级方案)
	notificationServiceURL := config.GetEnv("NOTIFICATION_SERVICE_URL", "http://localhost:40008")
	notificationClient := client.NewNotificationClient(notificationServiceURL)

	repo := repository.NewOrderRepository(application.DB)
	svc := service.NewOrderService(application.DB, repo, notificationClient, eventPublisher)
	handler := handler.NewOrderHandler(svc)

	idempotencyManager := idempotency.NewIdempotencyManager(application.Redis, "order-service", 24*time.Hour)
	application.Router.Use(middleware.IdempotencyMiddleware(idempotencyManager))

	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	handler.RegisterRoutes(application.Router)

	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

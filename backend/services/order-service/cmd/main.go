package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
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
	// 1. 初始化配置客户端（可选，失败不影响启动）
	var configClient *configclient.Client
	enableConfigClient := config.GetEnv("ENABLE_CONFIG_CLIENT", "false") == "true"

	if enableConfigClient {
		// 检查是否启用 mTLS
		enableConfigMTLS := config.GetEnvBool("CONFIG_CLIENT_MTLS", false)

		clientCfg := configclient.ClientConfig{
			ServiceName: "order-service",
			Environment: config.GetEnv("ENV", "production"),
			ConfigURL:   config.GetEnv("CONFIG_SERVICE_URL", "http://localhost:40010"),
			RefreshRate: 30 * time.Second,
		}

		// 如果启用 mTLS,添加证书配置
		if enableConfigMTLS {
			clientCfg.EnableMTLS = true
			clientCfg.TLSCertFile = config.GetEnv("TLS_CERT_FILE", "")
			clientCfg.TLSKeyFile = config.GetEnv("TLS_KEY_FILE", "")
			clientCfg.TLSCAFile = config.GetEnv("TLS_CA_FILE", "")
		}

		client, err := configclient.NewClient(clientCfg)
		if err != nil {
			logger.Warn("配置客户端初始化失败，将使用环境变量", zap.Error(err))
		} else {
			configClient = client
			defer configClient.Stop()
			logger.Info("配置中心客户端初始化成功")
		}
	}

	// 定义配置获取函数：优先从配置中心获取，失败则使用环境变量
	getConfig := func(key, defaultValue string) string {
		if configClient != nil {
			if val := configClient.Get(key); val != "" {
				return val
			}
		}
		return config.GetEnv(key, defaultValue)
	}

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

	// 初始化 Kafka Brokers（优先从配置中心获取）
	var kafkaBrokers []string
	kafkaBrokersStr := getConfig("KAFKA_BROKERS", "localhost:40092")
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
		logger.Info(fmt.Sprintf("Kafka Brokers配置完成: %v", kafkaBrokers))
	} else {
		logger.Info("未配置Kafka，将使用降级模式")
	}

	// 初始化 EventPublisher
	eventPublisher := kafka.NewEventPublisher(kafkaBrokers)
	logger.Info("EventPublisher 初始化完成")

	// 初始化 HTTP 客户端 (保留作为降级方案，优先从配置中心获取URL)
	notificationServiceURL := getConfig("NOTIFICATION_SERVICE_URL", "http://localhost:40008")
	notificationClient := client.NewNotificationClient(notificationServiceURL)
	logger.Info(fmt.Sprintf("通知服务客户端初始化: %s", notificationServiceURL))

	repo := repository.NewOrderRepository(application.DB)
	svc := service.NewOrderService(application.DB, repo, application.Redis, notificationClient, eventPublisher)
	handler := handler.NewOrderHandler(svc)

	idempotencyManager := idempotency.NewIdempotencyManager(application.Redis, "order-service", 24*time.Hour)
	application.Router.Use(middleware.IdempotencyMiddleware(idempotencyManager))

	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	handler.RegisterRoutes(application.Router)

	// JWT 认证中间件（优先从配置中心获取）
	// ⚠️ 安全要求: JWT_SECRET必须在生产环境中设置，不能使用默认值
	jwtSecret := getConfig("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("JWT_SECRET environment variable is required and cannot be empty")
	}
	if len(jwtSecret) < 32 {
		logger.Fatal("JWT_SECRET must be at least 32 characters for security",
			zap.Int("current_length", len(jwtSecret)),
			zap.Int("minimum_length", 32))
	}
	logger.Info("JWT_SECRET validation passed", zap.Int("length", len(jwtSecret)))
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// 启动服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

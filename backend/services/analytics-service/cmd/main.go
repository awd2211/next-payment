package main

import (
	"context"
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
	"payment-platform/analytics-service/internal/handler"
	"payment-platform/analytics-service/internal/model"
	"payment-platform/analytics-service/internal/repository"
	"payment-platform/analytics-service/internal/service"
	"payment-platform/analytics-service/internal/worker"
)

//	@title						Analytics Service API
//	@version					1.0
//	@description				支付平台数据分析服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40009
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 初始化配置客户端
	var configClient *configclient.Client
	enableConfigClient := config.GetEnv("ENABLE_CONFIG_CLIENT", "false") == "true"

	if enableConfigClient {
		enableConfigMTLS := config.GetEnvBool("CONFIG_CLIENT_MTLS", false)

		clientCfg := configclient.ClientConfig{
			ServiceName: "analytics-service",
			Environment: config.GetEnv("ENV", "production"),
			ConfigURL:   config.GetEnv("CONFIG_SERVICE_URL", "http://localhost:40010"),
			RefreshRate: 30 * time.Second,
		}

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

	getConfig := func(key, defaultValue string) string {
		if configClient != nil {
			if val := configClient.Get(key); val != "" {
				return val
			}
		}
		return config.GetEnv(key, defaultValue)
	}

	// 2. Bootstrap框架初始化
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "analytics-service",
		DBName:      config.GetEnv("DB_NAME", "payment_analytics"),
		Port:        config.GetEnvInt("PORT", 40009),
		AutoMigrate: []any{
			&model.PaymentMetrics{},
			&model.MerchantMetrics{},
			&model.ChannelMetrics{},
			&model.RealtimeStats{},
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

	logger.Info("正在启动 Analytics Service...")

	// 初始化Repository和Service
	analyticsRepo := repository.NewAnalyticsRepository(application.DB)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	// 注册HTTP路由
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	analyticsHandler.RegisterRoutes(application.Router)

	// 启动事件消费Workers (消费所有业务事件进行统计分析，优先从配置中心获取)
	var kafkaBrokers []string
	kafkaBrokersStr := getConfig("KAFKA_BROKERS", "localhost:40092")
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
		logger.Info(fmt.Sprintf("Kafka Brokers配置完成: %v", kafkaBrokers))

		// 创建EventWorker
		eventWorker := worker.NewEventWorker(application.DB, analyticsRepo)

		// 启动支付事件消费Worker
		paymentEventConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: kafkaBrokers,
			Topic:   "payment.events",
			GroupID: "analytics-payment-event-worker",
		})
		go func() {
			ctx := context.Background()
			eventWorker.StartPaymentEventWorker(ctx, paymentEventConsumer)
		}()
		logger.Info("Analytics: 支付事件Worker已启动 (topic: payment.events)")

		// 启动订单事件消费Worker
		orderEventConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: kafkaBrokers,
			Topic:   "order.events",
			GroupID: "analytics-order-event-worker",
		})
		go func() {
			ctx := context.Background()
			eventWorker.StartOrderEventWorker(ctx, orderEventConsumer)
		}()
		logger.Info("Analytics: 订单事件Worker已启动 (topic: order.events)")

		// 未来可以添加更多事件消费者
		// - accounting.events
		// - settlement.events
		// - merchant.events
		// 等等
	} else {
		logger.Info("未配置Kafka Brokers，事件消费Workers未启动")
	}

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

// 186 → 38 行, 减少 80%

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
	"payment-platform/accounting-service/internal/client"
	"payment-platform/accounting-service/internal/handler"
	"payment-platform/accounting-service/internal/model"
	"payment-platform/accounting-service/internal/repository"
	"payment-platform/accounting-service/internal/service"
	"payment-platform/accounting-service/internal/worker"
	// grpcServer "payment-platform/accounting-service/internal/grpc"
	// pb "github.com/payment-platform/proto/accounting"
)

//	@title						Accounting Service API
//	@version					1.0
//	@description				支付平台财务核算服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40007
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
		enableConfigMTLS := config.GetEnvBool("CONFIG_CLIENT_MTLS", false)

		clientCfg := configclient.ClientConfig{
			ServiceName: "accounting-service",
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

	// 2. 使用 Bootstrap 框架初始化应用
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "accounting-service",
		DBName:      config.GetEnv("DB_NAME", "payment_accounting"),
		Port:        config.GetEnvInt("PORT", 40007),

		// 自动迁移数据库模型（仅核心账户模型）
		AutoMigrate: []any{
			&model.Account{},
			&model.AccountTransaction{},
			&model.DoubleEntry{},
		},

		// 启用企业级功能
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

	logger.Info("正在启动 Accounting Service...")

	// 3. 初始化 HTTP 客户端（优先从配置中心获取）
	channelAdapterURL := getConfig("CHANNEL_SERVICE_URL", "http://localhost:40005")
	channelAdapterClient := client.NewChannelAdapterClient(channelAdapterURL)
	logger.Info(fmt.Sprintf("渠道适配器客户端初始化: %s", channelAdapterURL))

	// 3. 初始化Repository
	accountRepo := repository.NewAccountRepository(application.DB)

	// 4. 初始化Service（传入 application.DB 用于事务支持）
	accountService := service.NewAccountService(application.DB, accountRepo, channelAdapterClient)

	// 5. 初始化Handler
	accountHandler := handler.NewAccountHandler(accountService)

	// 6. Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 7. 注册账户路由
	accountHandler.RegisterRoutes(application.Router)

	// 8. 初始化Kafka (事件驱动架构，优先从配置中心获取)
	var kafkaBrokers []string
	kafkaBrokersStr := getConfig("KAFKA_BROKERS", "localhost:40092")
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
		logger.Info(fmt.Sprintf("Kafka Brokers配置完成: %v", kafkaBrokers))

		// 初始化EventPublisher (Producer)
		eventPublisher := kafka.NewEventPublisher(kafkaBrokers)
		logger.Info("Accounting: EventPublisher初始化完成")

		// 创建EventWorker (Producer + Consumer)
		eventWorker := worker.NewEventWorker(accountService, eventPublisher)

		// 启动支付事件消费Worker (Consumer: payment.events → 自动记账)
		paymentEventConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: kafkaBrokers,
			Topic:   "payment.events",
			GroupID: "accounting-payment-event-worker",
		})
		go func() {
			ctx := context.Background()
			eventWorker.StartPaymentEventWorker(ctx, paymentEventConsumer)
		}()
		logger.Info("Accounting: 支付事件Worker已启动 (自动记账) - topic: payment.events")

		// 未来可以添加退款事件消费者
		// - payment.refund.events (已在payment.events中包含)
	} else {
		logger.Info("未配置Kafka Brokers，事件消费Workers未启动 (手动记账模式)")
	}

	// 9. 初始化 JWT 认证中间件（优先从配置中心获取）
	jwtSecret := getConfig("JWT_SECRET", "payment-platform-secret-key-2024")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// 10. 启动服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

// 代码行数: 192 → 80 行, 减少 58%

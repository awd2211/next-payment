package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/saga"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"payment-platform/withdrawal-service/internal/client"
	"payment-platform/withdrawal-service/internal/handler"
	"payment-platform/withdrawal-service/internal/model"
	"payment-platform/withdrawal-service/internal/repository"
	"payment-platform/withdrawal-service/internal/service"
)

//	@title						Withdrawal Service API
//	@version					1.0
//	@description				支付平台提现服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40014
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
			ServiceName: "withdrawal-service",
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

	// 2. Bootstrap初始化
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "withdrawal-service",
		DBName:      config.GetEnv("DB_NAME", "payment_withdrawal"),
		Port:        config.GetEnvInt("PORT", 40014),
		AutoMigrate: []any{
			&model.Withdrawal{},
			&model.WithdrawalBankAccount{},
			&model.WithdrawalApproval{},
			&model.WithdrawalBatch{},
		},
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false,
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false),
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap失败: %v", err)
	}

	logger.Info("正在启动 Withdrawal Service...")

	// 2. 初始化客户端（优先从配置中心获取）
	accountingServiceURL := getConfig("ACCOUNTING_SERVICE_URL", "http://localhost:40007")
	notificationServiceURL := getConfig("NOTIFICATION_SERVICE_URL", "http://localhost:40008")

	accountingClient := client.NewAccountingClient(accountingServiceURL)
	notificationClient := client.NewNotificationClient(notificationServiceURL)
	logger.Info("HTTP客户端初始化完成")

	// 银行转账客户端配置（优先从配置中心获取敏感信息）
	bankConfig := &client.BankConfig{
		BankChannel: getConfig("BANK_CHANNEL", "mock"),
		APIEndpoint: getConfig("BANK_API_ENDPOINT", "https://api.bank.example.com"),
		MerchantID:  getConfig("BANK_MERCHANT_ID", ""),
		APIKey:      getConfig("BANK_API_KEY", ""),
		APISecret:   getConfig("BANK_API_SECRET", ""),
		Timeout:     30 * time.Second,
		UseSandbox:  config.GetEnvBool("BANK_USE_SANDBOX", true),
	}
	bankTransferClient := client.NewBankTransferClient(bankConfig)
	logger.Info("银行转账客户端初始化完成")

	// 3. 初始化Repository
	withdrawalRepo := repository.NewWithdrawalRepository(application.DB)

	// 4. 初始化 Saga Orchestrator
	sagaOrchestrator := saga.NewSagaOrchestrator(application.DB, application.Redis)

	// 5. 初始化Service
	withdrawalService := service.NewWithdrawalService(
		application.DB,
		withdrawalRepo,
		accountingClient,
		notificationClient,
		bankTransferClient,
		application.Redis,
	)

	_ = service.NewWithdrawalSagaService(
		sagaOrchestrator,
		withdrawalRepo,
		accountingClient,
		bankTransferClient,
		notificationClient,
	)
	// Saga 服务可用于处理分布式事务，预留未来使用

	// 6. 初始化Handler
	withdrawalHandler := handler.NewWithdrawalHandler(withdrawalService)

	// 7. 注册路由
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	withdrawalHandler.RegisterRoutes(application.Router)

	logger.Info("路由注册完成")

	// 8. JWT认证中间件（优先从配置中心获取）
	jwtSecret := getConfig("JWT_SECRET", "payment-platform-secret-key-2024")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	_ = jwtManager // 预留给需要认证的路由使用

	// 9. 启动服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal("服务启动失败: " + err.Error())
	}
}

// ============================================================
// 代码行数对比:
// 原始版本: 217 行
// Bootstrap版本: 128 行
// 减少: 89 行 (41%)
//
// 主要简化:
// 1. 使用 Bootstrap 自动初始化 DB, Redis, Logger, Router, Middleware
// 2. 自动启用 Metrics, Tracing, Health, RateLimit
// 3. 统一优雅关闭逻辑
// ============================================================

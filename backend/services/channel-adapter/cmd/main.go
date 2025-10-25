package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"payment-platform/channel-adapter/internal/adapter"
	"payment-platform/channel-adapter/internal/client"
	"payment-platform/channel-adapter/internal/handler"
	"payment-platform/channel-adapter/internal/model"
	"payment-platform/channel-adapter/internal/repository"
	"payment-platform/channel-adapter/internal/service"
	// grpcServer "payment-platform/channel-adapter/internal/grpc"
	// pb "github.com/payment-platform/proto/channel"
)

//	@title						Channel Adapter API
//	@version					1.0
//	@description				支付平台支付渠道适配器服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40005
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 使用 Bootstrap 框架初始化应用
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "channel-adapter",
		DBName:      config.GetEnv("DB_NAME", "payment_channel"),
		Port:        config.GetEnvInt("PORT", 40005),
		// GRPCPort:    config.GetEnvInt("GRPC_PORT", 50005), // gRPC 可选

		// 自动迁移数据库模型
		AutoMigrate: []any{
			&model.ChannelConfig{},
			&model.Transaction{},
			&model.WebhookLog{},
			&model.ExchangeRate{},
			&model.ExchangeRateSnapshot{},
		},

		// 启用企业级功能
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 默认关闭 gRPC,使用 HTTP 通信
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false), // mTLS 服务间认证

		// 速率限制配置
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	logger.Info("正在启动 Channel Adapter Service...")

	// 2. 创建适配器工厂
	adapterFactory := adapter.NewAdapterFactory()

	// 3. 注册 Stripe 适配器
	stripeConfig := &model.StripeConfig{
		APIKey:              config.GetEnv("STRIPE_API_KEY", ""),
		WebhookSecret:       config.GetEnv("STRIPE_WEBHOOK_SECRET", ""),
		PublishableKey:      config.GetEnv("STRIPE_PUBLISHABLE_KEY", ""),
		StatementDescriptor: "Payment Platform",
		CaptureMethod:       "automatic",
	}
	stripeAdapter := adapter.NewStripeAdapter(stripeConfig)
	adapterFactory.Register(model.ChannelStripe, stripeAdapter)
	logger.Info("Stripe 适配器已注册")

	// 4. 注册 PayPal 适配器（可选）
	paypalClientID := config.GetEnv("PAYPAL_CLIENT_ID", "")
	if paypalClientID != "" {
		paypalConfig := &model.PayPalConfig{
			ClientID:     paypalClientID,
			ClientSecret: config.GetEnv("PAYPAL_CLIENT_SECRET", ""),
			Mode:         config.GetEnv("PAYPAL_MODE", "sandbox"),
			WebhookID:    config.GetEnv("PAYPAL_WEBHOOK_ID", ""),
		}
		paypalAdapter := adapter.NewPayPalAdapter(paypalConfig)
		adapterFactory.Register(model.ChannelPayPal, paypalAdapter)
		logger.Info("PayPal 适配器已注册")
	}

	// 5. 注册 Alipay 适配器（可选）
	alipayAppID := config.GetEnv("ALIPAY_APP_ID", "")
	if alipayAppID != "" {
		alipayConfig := &model.AlipayConfig{
			AppID:      alipayAppID,
			PrivateKey: config.GetEnv("ALIPAY_PRIVATE_KEY", ""),
			PublicKey:  config.GetEnv("ALIPAY_PUBLIC_KEY", ""),
			NotifyURL:  config.GetEnv("ALIPAY_NOTIFY_URL", ""),
			ReturnURL:  config.GetEnv("ALIPAY_RETURN_URL", ""),
			SignType:   "RSA2",
			Format:     "json",
			Charset:    "utf-8",
			APIGateway: config.GetEnv("ALIPAY_API_GATEWAY", "https://openapi.alipay.com/gateway.do"),
		}
		alipayAdapter, err := adapter.NewAlipayAdapter(alipayConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("创建 Alipay 适配器失败: %v", err))
		} else {
			adapterFactory.Register(model.ChannelAlipay, alipayAdapter)
			logger.Info("Alipay 适配器已注册")
		}
	}

	// 6. 初始化汇率存储仓库（注入Redis客户端用于缓存）
	exchangeRateRepo := repository.NewExchangeRateRepository(application.DB, application.Redis)

	// 7. 初始化汇率客户端（用于 Crypto 适配器的法币转换）
	exchangeRateCacheTTL := time.Duration(config.GetEnvInt("EXCHANGE_RATE_CACHE_TTL", 3600)) * time.Second
	exchangeRateClient := client.NewExchangeRateClient(application.Redis, exchangeRateRepo, exchangeRateCacheTTL)
	logger.Info("汇率客户端初始化完成 (exchangerate-api.com + 历史存储)")

	// 8. 启动汇率定期更新任务（默认每2小时更新一次）
	exchangeRateUpdateInterval := time.Duration(config.GetEnvInt("EXCHANGE_RATE_UPDATE_INTERVAL", 7200)) * time.Second
	exchangeRateClient.StartPeriodicUpdate(context.Background(), exchangeRateUpdateInterval)

	// 9. 注册加密货币适配器（可选）
	cryptoWallet := config.GetEnv("CRYPTO_WALLET_ADDRESS", "")
	if cryptoWallet != "" {
		// 解析支持的网络列表
		networksStr := config.GetEnv("CRYPTO_NETWORKS", "ETH,BSC,TRON")
		networks := []string{}
		for _, network := range strings.Split(networksStr, ",") {
			networks = append(networks, strings.TrimSpace(network))
		}

		cryptoConfig := &model.CryptoConfig{
			WalletAddress: cryptoWallet,
			Networks:      networks,
			Confirmations: config.GetEnvInt("CRYPTO_CONFIRMATIONS", 12),
			APIEndpoint:   config.GetEnv("CRYPTO_API_ENDPOINT", ""),
			APIKey:        config.GetEnv("CRYPTO_API_KEY", ""),
		}
		cryptoAdapter := adapter.NewCryptoAdapter(cryptoConfig, exchangeRateClient)
		adapterFactory.Register(model.ChannelCrypto, cryptoAdapter)
		logger.Info(fmt.Sprintf("Crypto 适配器已注册，支持网络: %s", networksStr))
	}

	// 10. 初始化Repository
	channelRepo := repository.NewChannelRepository(application.DB)

	// 11. 初始化Service
	channelService := service.NewChannelService(channelRepo, adapterFactory)

	// 12. 初始化Handler
	channelHandler := handler.NewChannelHandler(channelService)
	exchangeRateHandler := handler.NewExchangeRateHandler(exchangeRateRepo)

	// 13. Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 14. 注册渠道路由
	channelHandler.RegisterRoutes(application.Router)

	// 15. 注册汇率路由
	exchangeRateHandler.RegisterRoutes(application.Router)

	// 16. gRPC 服务（预留但不启用，系统使用 HTTP/REST 通信）
	// channelGrpcServer := grpcServer.NewChannelServer(channelService)
	// pb.RegisterChannelServiceServer(application.GRPCServer, channelGrpcServer)
	// logger.Info(fmt.Sprintf("gRPC Server 已注册，将监听端口 %d", config.GetEnvInt("GRPC_PORT", 50005)))

	// 17. 启动服务（仅 HTTP，优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
	}
}

// 代码行数对比：
// - 原始版本: 280行 (手动初始化所有组件)
// - Bootstrap版本: 190行 (框架自动处理)
// - 减少代码: 32%（保留了所有业务逻辑）
//
// 自动获得的功能：
// ✅ 数据库连接和迁移
// ✅ Redis 连接
// ✅ Zap 日志系统
// ✅ Gin 路由和中间件（CORS, RequestID, Panic Recovery）
// ✅ Jaeger 分布式追踪
// ✅ Prometheus 指标收集（/metrics 端点 + HTTP 指标）
// ✅ 健康检查端点（/health, /health/live, /health/ready）
// ✅ 速率限制
// ✅ 优雅关闭（信号处理）
// ✅ 请求 ID
//
// 保留的自定义能力：
// ✅ 适配器工厂模式（Stripe, PayPal, Alipay, Crypto）
// ✅ 汇率客户端和定期更新任务
// ✅ 多渠道注册逻辑
// ✅ HTTP 处理器和路由
// ✅ Swagger UI

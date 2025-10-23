package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/db"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/metrics"
	"github.com/payment-platform/pkg/middleware"
	"github.com/payment-platform/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"payment-platform/channel-adapter/internal/adapter"
	"payment-platform/channel-adapter/internal/client"
	"payment-platform/channel-adapter/internal/handler"
	"payment-platform/channel-adapter/internal/model"
	"payment-platform/channel-adapter/internal/repository"
	"payment-platform/channel-adapter/internal/service"
	grpcServer "payment-platform/channel-adapter/internal/grpc"
	pb "github.com/payment-platform/proto/channel"
	pkggrpc "github.com/payment-platform/pkg/grpc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	// 初始化日志
	env := config.GetEnv("ENV", "development")
	if err := logger.InitLogger(env); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("正在启动 Channel Adapter Service...")

	// 初始化数据库
	dbConfig := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "payment_channel"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		TimeZone: config.GetEnv("DB_TIMEZONE", "UTC"),
	}

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		logger.Fatal("数据库连接失败")
		log.Fatalf("Error: %v", err)
	}
	logger.Info("数据库连接成功")

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		&model.ChannelConfig{},
		&model.Transaction{},
		&model.WebhookLog{},
	); err != nil {
		logger.Fatal("数据库迁移失败")
		log.Fatalf("Error: %v", err)
	}
	logger.Info("数据库迁移完成")

	// 初始化Redis
	redisConfig := db.RedisConfig{
		Host:     config.GetEnv("REDIS_HOST", "localhost"),
		Port:     config.GetEnvInt("REDIS_PORT", 6379),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       config.GetEnvInt("REDIS_DB", 0),
	}

	redisClient, err := db.NewRedisClient(redisConfig)
	if err != nil {
		logger.Fatal("Redis连接失败")
		log.Fatalf("Error: %v", err)
	}
	logger.Info("Redis连接成功")

	// 初始化 Prometheus 指标
	httpMetrics := metrics.NewHTTPMetrics("channel_adapter")
	logger.Info("Prometheus 指标初始化完成")

	// 初始化 Jaeger 分布式追踪
	jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	samplingRate := float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0
	tracerShutdown, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "channel-adapter",
		ServiceVersion: "1.0.0",
		Environment:    env,
		JaegerEndpoint: jaegerEndpoint,
		SamplingRate:   samplingRate,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Jaeger 初始化失败: %v", err))
	} else {
		logger.Info("Jaeger 追踪初始化完成")
		defer tracerShutdown(context.Background())
	}

	// 创建适配器工厂
	adapterFactory := adapter.NewAdapterFactory()

	// 注册 Stripe 适配器
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

	// 注册 PayPal 适配器
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

	// 注册 Alipay 适配器
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

	// 初始化汇率客户端（用于 Crypto 适配器的法币转换）
	exchangeRateCacheTTL := time.Duration(config.GetEnvInt("EXCHANGE_RATE_CACHE_TTL", 3600)) * time.Second
	exchangeRateClient := client.NewExchangeRateClient(redisClient, exchangeRateCacheTTL)
	logger.Info("汇率客户端初始化完成 (exchangerate-api.com)")

	// 注册加密货币适配器
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

	// 初始化Repository
	channelRepo := repository.NewChannelRepository(database)

	// 初始化Service
	channelService := service.NewChannelService(channelRepo, adapterFactory)

	// 初始化Handler
	channelHandler := handler.NewChannelHandler(channelService)

	// 初始化Gin
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(tracing.TracingMiddleware("channel-adapter"))
	r.Use(middleware.Logger(logger.Log))
	r.Use(metrics.PrometheusMiddleware(httpMetrics)) // Prometheus HTTP 指标收集

	// 限流中间件
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(rateLimiter.RateLimit())

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "channel-adapter",
			"time":    time.Now().Unix(),
		})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册渠道路由
	channelHandler.RegisterRoutes(r)

	// 启动 gRPC 服务器（独立 goroutine）
	grpcPort := config.GetEnvInt("GRPC_PORT", 50005)
	gRPCServer := pkggrpc.NewSimpleServer()
	channelGrpcServer := grpcServer.NewChannelServer(channelService)
	pb.RegisterChannelServiceServer(gRPCServer, channelGrpcServer)

	go func() {
		logger.Info(fmt.Sprintf("gRPC Server 正在监听端口 %d", grpcPort))
		if err := pkggrpc.StartServer(gRPCServer, grpcPort); err != nil {
			logger.Fatal(fmt.Sprintf("gRPC Server 启动失败: %v", err))
		}
	}()

	// 启动 HTTP 服务器
	port := config.GetEnvInt("PORT", 40005)
	addr := fmt.Sprintf(":%d", port)
	logger.Info(fmt.Sprintf("Channel Adapter Service 正在监听 %s", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务启动失败")
		log.Fatalf("Error: %v", err)
	}
}

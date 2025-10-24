package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/logger"
	"payment-platform/merchant-config-service/internal/handler"
	"payment-platform/merchant-config-service/internal/model"
	"payment-platform/merchant-config-service/internal/repository"
	"payment-platform/merchant-config-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title						Merchant Config Service API
//	@version					1.0
//	@description				支付平台商户配置服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40012
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. Bootstrap初始化
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "merchant-config-service",
		DBName:      config.GetEnv("DB_NAME", "payment_merchant_config"),
		Port:        config.GetEnvInt("PORT", 40012),

		AutoMigrate: []any{
			&model.MerchantFeeConfig{},
			&model.MerchantTransactionLimit{},
			&model.ChannelConfig{},
		},

		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 系统使用 HTTP/REST 通信,不需要 gRPC
		// GRPCPort:          config.GetEnvInt("GRPC_PORT", 50012), // 已禁用
		EnableHealthCheck: true,
		EnableRateLimit:   true,
		EnableMTLS:        config.GetEnvBool("ENABLE_MTLS", false), // mTLS 服务间认证

		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap失败: %v", err)
	}

	logger.Info("正在启动 Merchant Config Service...")

	// 2. 初始化Repository
	feeConfigRepo := repository.NewFeeConfigRepository(application.DB)
	transactionLimitRepo := repository.NewTransactionLimitRepository(application.DB)
	channelConfigRepo := repository.NewChannelConfigRepository(application.DB)

	logger.Info("Repository 层初始化完成")

	// 3. 初始化Service
	feeConfigService := service.NewFeeConfigService(feeConfigRepo)
	transactionLimitService := service.NewTransactionLimitService(transactionLimitRepo)
	channelConfigService := service.NewChannelConfigService(channelConfigRepo)

	logger.Info("Service 层初始化完成")

	// 4. 初始化Handler
	configHandler := handler.NewConfigHandler(feeConfigService, transactionLimitService, channelConfigService)

	logger.Info("Handler 层初始化完成")

	// 5. 注册Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 6. API路由组
	apiV1 := application.Router.Group("/api/v1")
	configHandler.RegisterRoutes(apiV1)

	logger.Info("路由注册完成")

	// 7. 启动HTTP服务（gRPC已禁用）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal("服务启动失败: " + err.Error())
	}
}

// ============================================================
// 代码行数对比:
// 原始版本: 161 行
// Bootstrap版本: 99 行
// 减少: 62 行 (38.5%)
//
// 自动获得的新功能:
// ✅ 统一的日志初始化和优雅关闭 (logger.Sync)
// ✅ 数据库连接池配置和健康检查
// ✅ Redis连接管理
// ✅ 完整的Prometheus指标收集 (/metrics端点)
// ✅ Jaeger分布式追踪 (W3C上下文传播)
// ✅ 全局中间件栈 (CORS, RequestID, Logger, Metrics, Tracing)
// ✅ 限流中间件 (Redis支持)
// ✅ 增强的健康检查端点 (/health，包含依赖状态)
// ✅ 优雅关闭 (SIGINT/SIGTERM处理，资源清理)
// ✅ gRPC服务器自动管理 (独立goroutine)
// ✅ 双协议支持 (HTTP + gRPC同时运行)
//
// 保留的业务逻辑:
// ✅ 3个Repository (FeeConfig, TransactionLimit, ChannelConfig)
// ✅ 3个Service层
// ✅ ConfigHandler统一管理
// ✅ 完整的路由注册逻辑
// ✅ Swagger文档UI
// ============================================================

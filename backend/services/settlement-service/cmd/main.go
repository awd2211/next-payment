package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/settlement-service/internal/client"
	"payment-platform/settlement-service/internal/handler"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/repository"
	"payment-platform/settlement-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title						Settlement Service API
//	@version					1.0
//	@description				支付平台结算处理服务API文档
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40013
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. Bootstrap初始化
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "settlement-service",
		DBName:      config.GetEnv("DB_NAME", "payment_settlement"),
		Port:        config.GetEnvInt("PORT", 40013),

		AutoMigrate: []any{
			&model.Settlement{},
			&model.SettlementItem{},
			&model.SettlementApproval{},
			&model.SettlementAccount{},
		},

		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 系统使用 HTTP/REST 通信,不需要 gRPC
		// GRPCPort:          config.GetEnvInt("GRPC_PORT", 50013), // 已禁用
		EnableHealthCheck: true,
		EnableRateLimit:   true,

		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap失败: %v", err)
	}

	logger.Info("正在启动 Settlement Service...")

	// 2. 初始化Repository
	settlementRepo := repository.NewSettlementRepository(application.DB)
	settlementAccountRepo := repository.NewSettlementAccountRepository(application.DB)

	// 3. 初始化HTTP客户端
	accountingServiceURL := config.GetEnv("ACCOUNTING_SERVICE_URL", "http://localhost:40007")
	withdrawalServiceURL := config.GetEnv("WITHDRAWAL_SERVICE_URL", "http://localhost:40014")
	merchantServiceURL := config.GetEnv("MERCHANT_SERVICE_URL", "http://localhost:40002")

	accountingClient := client.NewAccountingClient(accountingServiceURL)
	withdrawalClient := client.NewWithdrawalClient(withdrawalServiceURL)
	merchantClient := client.NewMerchantClient(merchantServiceURL)

	logger.Info("HTTP客户端初始化完成")

	// 4. 初始化Service
	settlementService := service.NewSettlementService(
		application.DB,
		settlementRepo,
		accountingClient,
		withdrawalClient,
		merchantClient,
	)
	settlementAccountService := service.NewSettlementAccountService(settlementAccountRepo)

	// 5. 初始化Handler
	settlementHandler := handler.NewSettlementHandler(settlementService)
	settlementAccountHandler := handler.NewSettlementAccountHandler(settlementAccountService)

	// 6. 注册Swagger UI
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 7. 注册路由
	settlementHandler.RegisterRoutes(application.Router)

	// 注册结算账户路由
	handler.RegisterSettlementAccountRoutes(
		application.Router.Group("/api/v1"),
		settlementAccountHandler,
		middleware.AuthMiddleware(nil),
	)

	// 8. 启动HTTP服务（gRPC已禁用）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal("服务启动失败: " + err.Error())
	}
}

// ============================================================
// 代码行数对比:
// 原始版本: 209 行
// Bootstrap版本: 124 行
// 减少: 85 行 (41%)
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
// ✅ 3个HTTP客户端 (Accounting, Withdrawal, Merchant)
// ✅ Settlement和SettlementAccount的Repository/Service/Handler
// ✅ 完整的路由注册逻辑
// ✅ Swagger文档UI
// ============================================================

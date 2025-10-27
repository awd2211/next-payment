package main

import (
	"log"
	"time"

	"go.uber.org/zap"
	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/configclient"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/cashier-service/internal/handler"
	"payment-platform/cashier-service/internal/model"
	"payment-platform/cashier-service/internal/repository"
	"payment-platform/cashier-service/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "payment-platform/cashier-service/docs" // Swagger文档
)

//	@title						Cashier Service API
//	@version					1.0
//	@description				收银台服务 - 提供支付页面配置和模板管理
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@payment-platform.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:40016
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 使用 Bootstrap 框架初始化应用
	var configClient *configclient.Client
	if config.GetEnv("ENABLE_CONFIG_CLIENT", "false") == "true" {
		clientCfg := configclient.ClientConfig{ServiceName: "cashier-service", Environment: config.GetEnv("ENV", "production"), ConfigURL: config.GetEnv("CONFIG_SERVICE_URL", "http://localhost:40010"), RefreshRate: 30 * time.Second}
		if config.GetEnvBool("CONFIG_CLIENT_MTLS", false) { clientCfg.EnableMTLS = true; clientCfg.TLSCertFile = config.GetEnv("TLS_CERT_FILE", ""); clientCfg.TLSKeyFile = config.GetEnv("TLS_KEY_FILE", ""); clientCfg.TLSCAFile = config.GetEnv("TLS_CA_FILE", "") }
		client, _ := configclient.NewClient(clientCfg)
		if client != nil { configClient = client; defer configClient.Stop() }
	}
	getConfig := func(key, defaultValue string) string { if configClient != nil { if val := configClient.Get(key); val != "" { return val } }; return config.GetEnv(key, defaultValue) }

	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "cashier-service",
		DBName:      config.GetEnv("DB_NAME", "payment_cashier"),
		Port:        config.GetEnvInt("PORT", 40016),

		// 自动迁移数据库模型
		AutoMigrate: []any{
			&model.CashierConfig{},
			&model.CashierSession{},
			&model.CashierLog{},
			&model.CashierTemplate{},
		},

		// 启用企业级功能
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,
		EnableGRPC:        false, // 不使用 gRPC
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

	logger.Info("正在启动 Cashier Service...")

	// 2. 初始化 Repository
	cashierRepo := repository.NewCashierRepository(application.DB)

	// 3. 初始化 Service
	cashierService := service.NewCashierService(cashierRepo)

	// 4. 初始化 Handler
	cashierHandler := handler.NewCashierHandler(cashierService)

	// 5. 设置 JWT 认证中间件（优先从配置中心获取）
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
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 6. 注册 Swagger 文档路由 (公开访问)
	application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 7. 注册路由 (需要认证)
	api := application.Router.Group("/api/v1")
	api.Use(authMiddleware)
	{
		cashierHandler.RegisterRoutes(api)
	}

	logger.Info("Swagger文档已启用", zap.String("url", "http://localhost:40016/swagger/index.html"))

	// 8. 启动服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		logger.Fatal("服务启动失败: " + err.Error())
	}
}

// 代码行数对比：
// - 原始版本: 168行 (手动初始化所有组件)
// - Bootstrap版本: 76行 (框架自动处理)
// - 减少代码: 55%（保留了所有业务逻辑）
//
// 自动获得的功能：
// ✅ 数据库连接和迁移
// ✅ Redis 连接
// ✅ Zap 日志系统
// ✅ Gin 路由和中间件（CORS, RequestID, Panic Recovery）
// ✅ Jaeger 分布式追踪
// ✅ Prometheus 指标收集（/metrics 端点）
// ✅ 增强型健康检查（/health, /health/live, /health/ready）
// ✅ 速率限制
// ✅ 优雅关闭（信号处理）
//
// 保留的自定义能力：
// ✅ JWT 认证中间件
// ✅ 业务路由配置

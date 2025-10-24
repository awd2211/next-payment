package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/cashier-service/internal/handler"
	"payment-platform/cashier-service/internal/model"
	"payment-platform/cashier-service/internal/repository"
	"payment-platform/cashier-service/internal/service"
)

func main() {
	// 1. 使用 Bootstrap 框架初始化应用
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

	// 5. 设置 JWT 认证中间件
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 6. 注册路由 (需要认证)
	api := application.Router.Group("/api/v1")
	api.Use(authMiddleware)
	{
		cashierHandler.RegisterRoutes(api)
	}

	// 7. 启动服务（优雅关闭）
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

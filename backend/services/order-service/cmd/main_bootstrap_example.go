package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/yourproject/order-service/internal/handler"
	"github.com/yourproject/order-service/internal/model"
	"github.com/yourproject/order-service/internal/repository"
	"github.com/yourproject/order-service/internal/service"
)

// 这是使用 Bootstrap 框架的简化版本
// 对比原始 main.go (190行)，这个版本只需要 ~60行

func main() {
	// 1. 使用 Bootstrap 框架初始化应用（替代 ~100行初始化代码）
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "order-service",
		DBName:      "payment_order",
		Port:        40004,

		// 自动迁移数据库模型
		AutoMigrate: []any{
			&model.Order{},
			&model.OrderItem{},
			&model.OrderLog{},
		},

		// 启用所有企业级功能
		EnableTracing:     true,  // Jaeger 分布式追踪
		EnableMetrics:     true,  // Prometheus 指标
		EnableRedis:       true,  // Redis 缓存
		EnableHealthCheck: true,  // K8s 健康检查
		EnableRateLimit:   true,  // 速率限制

		// 速率限制配置
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap 失败: %v", err)
	}

	// 2. 初始化业务层（三层架构）
	// Repository 层
	orderRepo := repository.NewOrderRepository(application.DB)
	orderItemRepo := repository.NewOrderItemRepository(application.DB)
	orderLogRepo := repository.NewOrderLogRepository(application.DB)

	// Service 层
	orderService := service.NewOrderService(
		application.DB,
		orderRepo,
		orderItemRepo,
		orderLogRepo,
	)

	// Handler 层
	orderHandler := handler.NewOrderHandler(orderService)

	// 3. 注册业务路由
	api := application.Router.Group("/api/v1")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:order_no", orderHandler.GetOrder)
			orders.GET("", orderHandler.ListOrders)
			orders.PUT("/:order_no/status", orderHandler.UpdateOrderStatus)
			orders.PUT("/:order_no/cancel", orderHandler.CancelOrder)
			orders.PUT("/:order_no/ship", orderHandler.ShipOrder)
			orders.PUT("/:order_no/refund", orderHandler.RefundOrder)
		}
	}

	// 4. 注册 Webhook 路由（无需速率限制）
	webhooks := application.Router.Group("/webhooks")
	{
		webhooks.POST("/payment", orderHandler.HandlePaymentWebhook)
	}

	// 5. 启动服务（优雅关闭）
	// 替代原来的手动启动代码
	if err := application.RunWithGracefulShutdown(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// 代码行数对比：
// - 原始版本: ~190行 (手动初始化所有组件)
// - Bootstrap版本: ~60行 (框架自动处理)
// - 减少代码: 68%
//
// 自动获得的功能：
// ✅ 数据库连接和迁移
// ✅ Redis 连接
// ✅ Zap 日志系统
// ✅ Gin 路由和中间件
// ✅ Jaeger 分布式追踪
// ✅ Prometheus 指标收集
// ✅ 健康检查端点 (/health, /health/live, /health/ready)
// ✅ 速率限制
// ✅ 优雅关闭（信号处理）
// ✅ 请求 ID
// ✅ CORS 配置
// ✅ Panic 恢复
//
// 保留的自定义能力：
// ✅ 自定义业务路由
// ✅ 自定义中间件
// ✅ 自定义健康检查
// ✅ 访问所有底层对象 (application.DB, application.Redis, application.Router)

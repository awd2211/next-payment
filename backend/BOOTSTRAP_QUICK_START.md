# Bootstrap框架快速入门指南

适用于新服务开发或现有服务迁移

---

## 🚀 30秒快速开始

### 最小化示例

```go
package main

import (
	"github.com/payment-platform/pkg/app"
	"payment-platform/your-service/internal/model"
)

func main() {
	application, _ := app.Bootstrap(app.ServiceConfig{
		ServiceName: "your-service",
		DBName:      "payment_your_db",
		Port:        40XXX,
		AutoMigrate: []any{&model.YourModel{}},
	})
	
	// 注册路由
	// yourHandler.RegisterRoutes(application.Router)
	
	application.RunWithGracefulShutdown()
}
```

**就这么简单!** 你已经获得了11项企业级功能。

---

## 📋 完整配置示例

```go
package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/payment-platform/pkg/auth"
	"github.com/payment-platform/pkg/config"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/your-service/internal/handler"
	"payment-platform/your-service/internal/model"
	"payment-platform/your-service/internal/repository"
	"payment-platform/your-service/internal/service"
)

func main() {
	// 1. Bootstrap初始化
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "your-service",
		DBName:      config.GetEnv("DB_NAME", "payment_your_db"),
		Port:        config.GetEnvInt("PORT", 40XXX),

		// 数据库模型自动迁移
		AutoMigrate: []any{
			&model.YourModel1{},
			&model.YourModel2{},
		},

		// 功能开关
		EnableTracing:     true,  // Jaeger追踪
		EnableMetrics:     true,  // Prometheus指标
		EnableRedis:       true,  // Redis连接
		EnableGRPC:        false, // gRPC服务器
		EnableHealthCheck: true,  // 健康检查
		EnableRateLimit:   true,  // 速率限制

		// gRPC配置(如果EnableGRPC=true)
		GRPCPort: config.GetEnvInt("GRPC_PORT", 50XXX),

		// 速率限制配置
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap失败: %v", err)
	}

	logger.Info("正在启动 Your Service...")

	// 2. 初始化Repository
	yourRepo := repository.NewYourRepository(application.DB)

	// 3. 初始化Service
	yourService := service.NewYourService(yourRepo)

	// 4. 初始化Handler
	yourHandler := handler.NewYourHandler(yourService)

	// 5. (可选) 自定义中间件
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// 6. 注册路由
	api := application.Router.Group("/api/v1")
	api.Use(authMiddleware) // 应用中间件
	{
		yourHandler.RegisterRoutes(api)
	}

	// 7. (可选) 注册gRPC服务
	if application.GRPCServer != nil {
		// pb.RegisterYourServiceServer(application.GRPCServer, grpcServer)
	}

	// 8. 启动服务
	if application.GRPCServer != nil {
		// HTTP + gRPC双协议
		if err := application.RunDualProtocol(); err != nil {
			logger.Fatal("服务启动失败: " + err.Error())
		}
	} else {
		// 仅HTTP
		if err := application.RunWithGracefulShutdown(); err != nil {
			logger.Fatal("服务启动失败: " + err.Error())
		}
	}
}
```

---

## 🎁 自动获得的功能

### 1. 基础设施 (3项)
- ✅ **Zap日志系统** - 结构化日志,自动Sync()
- ✅ **PostgreSQL连接池** - 健康检查 + 自动迁移
- ✅ **Redis客户端** - 集中管理 + 连接验证

### 2. 可观测性 (3项)
- ✅ **Prometheus指标** - `/metrics`端点 + HTTP指标
- ✅ **Jaeger追踪** - 分布式追踪 + W3C context传播
- ✅ **健康检查** - `/health`, `/health/live`, `/health/ready`

### 3. 中间件 (3项)
- ✅ **CORS** - 跨域请求处理
- ✅ **RequestID** - 请求追踪ID
- ✅ **速率限制** - Redis支持的限流器

### 4. 运维 (2项)
- ✅ **优雅关闭** - SIGINT/SIGTERM处理 + 资源清理
- ✅ **gRPC支持** - 可选的双协议(HTTP+gRPC)

---

## 📦 ServiceConfig 配置项

| 配置项 | 类型 | 必填 | 默认值 | 说明 |
|-------|------|------|--------|------|
| ServiceName | string | ✅ | - | 服务名称(用于日志和追踪) |
| DBName | string | ✅ | - | PostgreSQL数据库名 |
| Port | int | ✅ | - | HTTP端口 |
| AutoMigrate | []any | ❌ | nil | 自动迁移的GORM模型 |
| EnableTracing | bool | ❌ | false | 启用Jaeger追踪 |
| EnableMetrics | bool | ❌ | false | 启用Prometheus指标 |
| EnableRedis | bool | ❌ | false | 启用Redis连接 |
| EnableGRPC | bool | ❌ | false | 启用gRPC服务器 |
| EnableHealthCheck | bool | ❌ | false | 启用增强健康检查 |
| EnableRateLimit | bool | ❌ | false | 启用速率限制 |
| GRPCPort | int | ❌ | 0 | gRPC端口(如EnableGRPC=true) |
| RateLimitRequests | int | ❌ | 100 | 速率限制请求数 |
| RateLimitWindow | time.Duration | ❌ | 1min | 速率限制时间窗口 |

---

## 🔧 常用模式

### 模式1: 纯HTTP服务
```go
application, _ := app.Bootstrap(app.ServiceConfig{
	ServiceName: "api-service",
	DBName:      "payment_api",
	Port:        40001,
	EnableTracing:     true,
	EnableMetrics:     true,
	EnableRedis:       true,
	EnableHealthCheck: true,
	EnableRateLimit:   true,
})
// 注册HTTP路由
application.RunWithGracefulShutdown()
```

### 模式2: HTTP + gRPC双协议
```go
application, _ := app.Bootstrap(app.ServiceConfig{
	ServiceName: "dual-service",
	DBName:      "payment_dual",
	Port:        40002,
	GRPCPort:    50002,
	EnableGRPC:  true, // 启用gRPC
	// 其他配置...
})
// 注册HTTP路由
// 注册gRPC服务
application.RunDualProtocol() // 同时启动HTTP和gRPC
```

### 模式3: 添加自定义中间件
```go
application, _ := app.Bootstrap(/* config */)

// JWT认证
authMiddleware := middleware.AuthMiddleware(jwtManager)
api := application.Router.Group("/api/v1")
api.Use(authMiddleware)

// 幂等性
idempotencyManager := idempotency.NewIdempotencyManager(application.Redis, "service", 24*time.Hour)
application.Router.Use(middleware.IdempotencyMiddleware(idempotencyManager))

// 自定义中间件
application.Router.Use(yourCustomMiddleware)
```

### 模式4: 多客户端集成
```go
application, _ := app.Bootstrap(/* config */)

// 初始化多个HTTP客户端
client1URL := config.GetEnv("SERVICE1_URL", "http://localhost:40001")
client2URL := config.GetEnv("SERVICE2_URL", "http://localhost:40002")

client1 := client.NewService1Client(client1URL)
client2 := client.NewService2Client(client2URL)

// 依赖注入到Service层
yourService := service.NewYourService(yourRepo, client1, client2)
```

---

## 🌍 环境变量

Bootstrap框架使用这些环境变量:

### 数据库
```bash
DB_HOST=localhost
DB_PORT=40432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_your_db
DB_SSL_MODE=disable
DB_TIMEZONE=UTC
```

### Redis
```bash
REDIS_HOST=localhost
REDIS_PORT=40379
REDIS_PASSWORD=
REDIS_DB=0
```

### 服务端口
```bash
PORT=40XXX           # HTTP端口
GRPC_PORT=50XXX      # gRPC端口(如果启用)
```

### 可观测性
```bash
ENV=development                                      # 环境(development/production)
JAEGER_ENDPOINT=http://localhost:14268/api/traces  # Jaeger endpoint
JAEGER_SAMPLING_RATE=100                            # 采样率(0-100, 生产建议10-20)
```

---

## 📊 可用的端点

Bootstrap自动创建这些端点:

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 基础健康检查 |
| `/health/live` | GET | Liveness探针(Kubernetes) |
| `/health/ready` | GET | Readiness探针(Kubernetes) |
| `/metrics` | GET | Prometheus指标 |

---

## 🔍 调试和监控

### 查看健康状态
```bash
curl http://localhost:40XXX/health
```

响应示例:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-24T08:00:00Z",
  "checks": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

### 查看Prometheus指标
```bash
curl http://localhost:40XXX/metrics
```

### 查看Jaeger追踪
访问: http://localhost:40686

### 查看日志
日志自动输出到stdout,格式为JSON:
```json
{
  "level": "info",
  "ts": "2025-10-24T08:00:00.000Z",
  "msg": "Server started",
  "port": 40001
}
```

---

## 🐛 故障排查

### 问题: 服务启动失败
**检查**:
1. 数据库连接是否正常?
2. Redis连接是否正常?
3. 端口是否被占用?

**解决**:
```bash
# 检查数据库
psql -h localhost -p 40432 -U postgres

# 检查Redis
redis-cli -h localhost -p 40379 ping

# 检查端口
lsof -i :40XXX
```

### 问题: AutoMigrate失败
**检查**: 模型定义是否正确? GORM标签是否完整?

**解决**: 查看日志中的具体错误信息

### 问题: gRPC服务无法启动
**检查**: `EnableGRPC`是否为`true`? `GRPCPort`是否配置?

### 问题: 性能下降
**检查**:
1. 速率限制是否过低?
2. 数据库连接池是否足够?
3. Redis延迟是否正常?

**优化**:
- 调整`RateLimitRequests`
- 配置数据库连接池大小
- 使用Redis集群

---

## 📚 参考示例

### 简单服务: cashier-service
[backend/services/cashier-service/cmd/main.go](backend/services/cashier-service/cmd/main.go)
- 无客户端依赖
- JWT认证
- 96行代码

### 复杂服务: payment-gateway
[backend/services/payment-gateway/cmd/main.go](backend/services/payment-gateway/cmd/main.go)
- 3个HTTP客户端
- Saga分布式事务
- 自定义签名验证中间件
- 239行代码

### 双协议服务: kyc-service
[backend/services/kyc-service/cmd/main.go](backend/services/kyc-service/cmd/main.go)
- HTTP + gRPC
- Swagger UI
- 119行代码

### 多客户端服务: settlement-service
[backend/services/settlement-service/cmd/main.go](backend/services/settlement-service/cmd/main.go)
- 3个HTTP客户端
- 完整依赖注入
- 144行代码

---

## 💡 最佳实践

### 1. 配置管理
- ✅ 使用环境变量(不要硬编码)
- ✅ 提供合理的默认值
- ✅ 使用`config.GetEnv()`助手函数

### 2. 错误处理
```go
application, err := app.Bootstrap(config)
if err != nil {
	log.Fatalf("Bootstrap失败: %v", err)
}
```

### 3. 资源清理
Bootstrap自动处理资源清理,包括:
- 数据库连接关闭
- Redis连接关闭
- HTTP服务器优雅关闭
- gRPC服务器优雅关闭

### 4. 日志记录
```go
logger.Info("服务启动")
logger.Error("错误信息", zap.Error(err))
logger.Warn("警告信息", zap.String("key", "value"))
```

### 5. 中间件顺序
```go
// 1. 全局中间件(Bootstrap自动添加)
// 2. 自定义全局中间件
application.Router.Use(yourGlobalMiddleware)
// 3. 路由组中间件
api := application.Router.Group("/api/v1")
api.Use(authMiddleware)
```

---

## 🎓 进阶主题

### 自定义健康检查
```go
import "github.com/payment-platform/pkg/health"

healthChecker := health.NewHealthChecker()
healthChecker.Register(health.NewDBChecker("database", application.DB))
healthChecker.Register(health.NewRedisChecker("redis", application.Redis))
healthChecker.Register(health.NewServiceHealthChecker("downstream", "http://localhost:40002"))

healthHandler := health.NewGinHandler(healthChecker)
application.Router.GET("/health", healthHandler.Handle)
```

### 自定义Prometheus指标
```go
import "github.com/payment-platform/pkg/metrics"

paymentMetrics := metrics.NewPaymentMetrics("your_service")
// 在业务代码中记录指标
paymentMetrics.RecordPayment(status, channel, currency, amount, duration)
```

### 后台任务
```go
// 在main函数中启动后台任务
go func() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// 执行定时任务
	}
}()

// 优雅关闭会自动处理goroutine清理
```

---

## ✅ 检查清单

创建新服务时的检查清单:

- [ ] 选择合适的端口号(40XXX for HTTP, 50XXX for gRPC)
- [ ] 定义所有GORM模型
- [ ] 配置AutoMigrate模型列表
- [ ] 确定需要的功能(Tracing, Metrics, Redis, gRPC, etc.)
- [ ] 实现Repository层
- [ ] 实现Service层
- [ ] 实现Handler层
- [ ] 注册路由
- [ ] (可选) 注册gRPC服务
- [ ] 测试编译
- [ ] 测试运行
- [ ] 验证健康检查
- [ ] 验证Prometheus指标
- [ ] 添加注释说明

---

## 🆘 获取帮助

- **Bootstrap源码**: [backend/pkg/app/bootstrap.go](backend/pkg/app/bootstrap.go)
- **示例服务**: 查看已迁移的11个服务
- **完整文档**: [BOOTSTRAP_MIGRATION_COMPLETE.md](BOOTSTRAP_MIGRATION_COMPLETE.md)

---

**快速开始,专注业务!** Bootstrap框架让你在几分钟内创建生产就绪的微服务。

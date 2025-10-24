# Bootstrap 框架使用指南

`pkg/app/bootstrap.go` 提供了统一的微服务启动框架，用于减少重复代码并标准化所有服务的初始化流程。

## 功能特性

### 核心功能
- ✅ **数据库连接** - 自动连接 PostgreSQL 并执行迁移
- ✅ **Redis 连接** - 可选的 Redis 客户端（用于缓存、会话、速率限制）
- ✅ **日志系统** - Zap 结构化日志，支持开发/生产模式
- ✅ **HTTP 路由** - Gin 框架，自动配置常用中间件
- ✅ **Swagger 文档** - 自动生成 API 文档（需要 swag）

### 可选功能
- ✅ **分布式追踪** - Jaeger/OpenTelemetry 集成
- ✅ **Prometheus 指标** - HTTP 请求指标自动收集
- ✅ **健康检查** - K8s 就绪/存活探针支持
- ✅ **速率限制** - 基于 Redis 的分布式速率限制
- ✅ **优雅关闭** - 信号处理和资源清理

## 快速开始

### 最小化示例

```go
package main

import (
	"github.com/payment-platform/pkg/app"
	"log"
)

func main() {
	// 创建应用实例
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "my-service",
		DBName:      "my_database",
		Port:        8080,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 注册业务路由
	application.Router.GET("/api/v1/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	// 启动服务（简单模式）
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
```

### 完整功能示例

```go
package main

import (
	"github.com/payment-platform/pkg/app"
	"github.com/yourproject/internal/model"
	"github.com/yourproject/internal/repository"
	"github.com/yourproject/internal/service"
	"github.com/yourproject/internal/handler"
	"log"
	"time"
)

func main() {
	// 创建应用实例（启用所有功能）
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "payment-gateway",
		DBName:      "payment_gateway",
		Port:        40003,
		GRPCPort:    50003, // gRPC 端口（如果启用）

		// 数据库自动迁移
		AutoMigrate: []any{
			&model.Payment{},
			&model.Refund{},
			&model.Transaction{},
		},

		// 启用所有可选功能
		EnableTracing:     true,  // Jaeger 分布式追踪
		EnableMetrics:     true,  // Prometheus 指标收集
		EnableRedis:       true,  // Redis 缓存
		EnableHealthCheck: true,  // 增强健康检查
		EnableRateLimit:   true,  // 速率限制

		// 速率限制配置
		RateLimitRequests: 100,        // 100 请求
		RateLimitWindow:   time.Minute, // 每分钟
	})
	if err != nil {
		log.Fatal(err)
	}

	// 初始化业务层（三层架构）
	// 1. Repository 层
	paymentRepo := repository.NewPaymentRepository(application.DB)
	refundRepo := repository.NewRefundRepository(application.DB)

	// 2. Service 层
	paymentService := service.NewPaymentService(paymentRepo, application.Redis)
	refundService := service.NewRefundService(refundRepo, paymentService)

	// 3. Handler 层
	paymentHandler := handler.NewPaymentHandler(paymentService)
	refundHandler := handler.NewRefundHandler(refundService)

	// 注册业务路由
	api := application.Router.Group("/api/v1")
	{
		// 支付相关路由
		payments := api.Group("/payments")
		{
			payments.POST("", paymentHandler.CreatePayment)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.GET("", paymentHandler.ListPayments)
		}

		// 退款相关路由
		refunds := api.Group("/refunds")
		{
			refunds.POST("", refundHandler.CreateRefund)
			refunds.GET("/:id", refundHandler.GetRefund)
		}
	}

	// 注册自定义健康检查（可选）
	if application.HealthChecker != nil {
		// 添加下游服务健康检查
		application.HealthChecker.Register(
			health.NewServiceHealthChecker("order-service", "http://localhost:40004/health"),
		)
	}

	// 启动服务（优雅关闭模式）
	if err := application.RunWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}
```

## 配置选项详解

### ServiceConfig 字段说明

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `ServiceName` | string | ✅ | - | 服务名称，用于日志、追踪、指标命名空间 |
| `DBName` | string | ✅ | - | PostgreSQL 数据库名 |
| `Port` | int | ✅ | - | HTTP 服务端口 |
| `GRPCPort` | int | ❌ | 0 | gRPC 服务端口（未启用则忽略） |
| `AutoMigrate` | []any | ❌ | nil | 需要自动迁移的 GORM 模型 |
| `EnableTracing` | bool | ❌ | true | 是否启用 Jaeger 分布式追踪 |
| `EnableMetrics` | bool | ❌ | true | 是否启用 Prometheus 指标 |
| `EnableRedis` | bool | ❌ | false | 是否连接 Redis |
| `EnableHealthCheck` | bool | ❌ | true | 是否启用增强健康检查 |
| `EnableRateLimit` | bool | ❌ | false | 是否启用速率限制（需要 Redis） |
| `RateLimitRequests` | int | ❌ | 100 | 速率限制请求数 |
| `RateLimitWindow` | time.Duration | ❌ | 1分钟 | 速率限制时间窗口 |

### 环境变量配置

Bootstrap 框架从环境变量读取配置，支持以下变量：

#### 通用配置
```bash
ENV=production              # 环境：development/production
```

#### 数据库配置
```bash
DB_HOST=localhost           # 数据库主机
DB_PORT=5432                # 数据库端口
DB_USER=postgres            # 数据库用户
DB_PASSWORD=your_password   # 数据库密码（生产环境禁止使用默认值）
DB_NAME=auto_from_config    # 数据库名（会被 ServiceConfig.DBName 覆盖）
DB_SSL_MODE=disable         # SSL 模式：disable/require
DB_TIMEZONE=UTC             # 时区
```

#### Redis 配置（EnableRedis=true 时）
```bash
REDIS_HOST=localhost        # Redis 主机
REDIS_PORT=6379             # Redis 端口
REDIS_PASSWORD=             # Redis 密码（可选）
REDIS_DB=0                  # Redis 数据库索引
```

#### 追踪配置（EnableTracing=true 时）
```bash
JAEGER_ENDPOINT=http://localhost:14268/api/traces  # Jaeger Collector 端点
JAEGER_SAMPLING_RATE=100    # 采样率 0-100（生产建议 10-20）
```

## App 实例方法

### Run() - 简单启动
```go
// 启动服务（阻塞，无优雅关闭）
if err := application.Run(); err != nil {
    log.Fatal(err)
}
```

### RunWithGracefulShutdown() - 优雅关闭
```go
// 启动服务并监听 SIGINT/SIGTERM 信号
// Ctrl+C 或 kill 信号会触发优雅关闭
if err := application.RunWithGracefulShutdown(); err != nil {
    log.Fatal(err)
}
```

**优雅关闭流程**:
1. 收到信号（SIGINT/SIGTERM）
2. 停止接受新请求
3. 等待现有请求处理完成（最多30秒）
4. 关闭数据库连接
5. 关闭 Redis 连接
6. 刷新日志缓冲区
7. 退出进程

### Shutdown(ctx) - 手动关闭
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := application.Shutdown(ctx); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

## 内置端点

Bootstrap 自动注册以下端点：

### 健康检查端点（EnableHealthCheck=true）

| 端点 | 用途 | 说明 |
|------|------|------|
| `GET /health` | 完整健康检查 | 检查数据库、Redis、下游服务 |
| `GET /health/live` | K8s 存活探针 | 检查服务进程是否运行 |
| `GET /health/ready` | K8s 就绪探针 | 检查服务是否准备接受流量 |

**响应示例**:
```json
{
  "status": "healthy",
  "checks": {
    "database": { "status": "healthy", "latency_ms": 2 },
    "redis": { "status": "healthy", "latency_ms": 1 }
  },
  "timestamp": 1698765432
}
```

### Prometheus 指标端点（EnableMetrics=true）

| 端点 | 用途 |
|------|------|
| `GET /metrics` | Prometheus 抓取端点 |

**收集的指标**:
- `http_requests_total` - HTTP 请求总数（按方法、路径、状态码）
- `http_request_duration_seconds` - HTTP 请求延迟（直方图）
- `http_request_size_bytes` - HTTP 请求大小
- `http_response_size_bytes` - HTTP 响应大小

## 中间件顺序

Bootstrap 按以下顺序注册中间件（确保正确性）：

```
1. gin.Recovery()              # Panic 恢复
2. middleware.RequestID()      # 生成请求 ID
3. middleware.CORS()           # 跨域配置
4. tracing.TracingMiddleware() # 分布式追踪（可选）
5. middleware.Logger()         # 访问日志
6. metrics.PrometheusMiddleware() # 指标收集（可选）
7. middleware.RateLimit()      # 速率限制（可选）
```

## 最佳实践

### 1. 生产环境安全检查

```go
// ❌ 错误：生产环境使用默认密码
DB_PASSWORD=postgres  # 启动会失败

// ✅ 正确：使用强密码
DB_PASSWORD=your_secure_password_here
```

### 2. 使用优雅关闭

```go
// ✅ 推荐：生产环境必须使用优雅关闭
application.RunWithGracefulShutdown()

// ❌ 不推荐：简单关闭可能导致请求丢失
application.Run()
```

### 3. 自定义健康检查

```go
// 添加下游服务健康检查
if application.HealthChecker != nil {
    application.HealthChecker.Register(
        health.NewServiceHealthChecker("order-service", "http://order:8004/health"),
    )
    application.HealthChecker.Register(
        health.NewServiceHealthChecker("payment-provider", "https://api.stripe.com"),
    )
}
```

### 4. 合理配置速率限制

```go
ServiceConfig{
    EnableRateLimit:   true,
    RateLimitRequests: 1000,        // 根据服务容量调整
    RateLimitWindow:   time.Minute,
}
```

### 5. 生产环境追踪采样率

```bash
# 开发环境：100% 采样
JAEGER_SAMPLING_RATE=100

# 生产环境：10-20% 采样（降低开销）
JAEGER_SAMPLING_RATE=10
```

## Docker Compose 示例

```yaml
version: '3.8'

services:
  my-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}  # 从 .env 读取
      - DB_NAME=my_database
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
      - JAEGER_SAMPLING_RATE=20
    depends_on:
      - postgres
      - redis
      - jaeger
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health/live"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: my_database
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redisdata:/data

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # Collector

volumes:
  pgdata:
  redisdata:
```

## Kubernetes 示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-service
  template:
    metadata:
      labels:
        app: my-service
    spec:
      containers:
      - name: my-service
        image: my-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: DB_HOST
          value: "postgres-service"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        - name: REDIS_HOST
          value: "redis-service"
        - name: JAEGER_ENDPOINT
          value: "http://jaeger-collector:14268/api/traces"

        # 健康检查配置
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3

        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## 故障排查

### 问题：数据库连接失败

```
错误: 数据库连接失败: dial tcp: connect: connection refused
```

**解决方案**:
1. 检查 `DB_HOST` 和 `DB_PORT` 是否正确
2. 确认 PostgreSQL 服务已启动
3. 检查网络连通性：`telnet $DB_HOST $DB_PORT`

### 问题：Redis 连接失败

```
错误: Redis 连接失败: dial tcp: i/o timeout
```

**解决方案**:
1. 确认 `EnableRedis=true`
2. 检查 `REDIS_HOST` 和 `REDIS_PORT`
3. 测试连接：`redis-cli -h $REDIS_HOST -p $REDIS_PORT ping`

### 问题：健康检查失败

```
GET /health/ready => 503 Service Unavailable
```

**解决方案**:
1. 查看具体失败的检查项：`curl http://localhost:8080/health`
2. 修复对应的依赖服务
3. 等待健康检查恢复

## 性能调优

### 数据库连接池

```go
// Bootstrap 会自动配置合理的连接池
// 如需自定义，可在 Bootstrap 后调整：
sqlDB, _ := application.DB.DB()
sqlDB.SetMaxOpenConns(25)        // 最大打开连接数
sqlDB.SetMaxIdleConns(5)         // 最大空闲连接数
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期
```

### HTTP 服务器超时

```go
// RunWithGracefulShutdown 内部使用默认超时
// 如需自定义，使用标准 http.Server：
srv := &http.Server{
    Addr:           ":8080",
    Handler:        application.Router,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
}
```

## 扩展性

Bootstrap 框架设计为可扩展的：

### 添加自定义中间件

```go
// 在业务路由之前添加
application.Router.Use(YourCustomMiddleware())

// 或针对特定路由组
api := application.Router.Group("/api/v1")
api.Use(AuthMiddleware())
```

### 注册额外端点

```go
// Swagger 文档
application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// Webhook 端点
application.Router.POST("/webhooks/stripe", webhookHandler.HandleStripe)
```

## 迁移指南

### 从手动初始化迁移到 Bootstrap

**迁移前** (main.go ~150行):
```go
// 初始化日志
logger.InitLogger("production")

// 连接数据库
db, _ := gorm.Open(...)

// 初始化 Redis
redis := redis.NewClient(...)

// 初始化 Gin
router := gin.New()
router.Use(middleware.CORS())
router.Use(middleware.Logger())
// ... 更多配置
```

**迁移后** (main.go ~10行):
```go
app, _ := app.Bootstrap(app.ServiceConfig{
    ServiceName: "my-service",
    DBName:      "my_db",
    Port:        8080,
    EnableRedis: true,
})
```

**节省代码行数**: ~140行 → 统一框架管理

## 总结

Bootstrap 框架提供了：
- ✅ **标准化** - 所有服务使用相同的初始化流程
- ✅ **简化** - 减少90%的样板代码
- ✅ **企业级** - 内置追踪、指标、健康检查
- ✅ **生产就绪** - 优雅关闭、安全检查、资源管理
- ✅ **可扩展** - 保留自定义能力

建议所有新服务使用 Bootstrap 框架，现有服务逐步迁移。

# Bootstrap 框架迁移指南

本文档指导如何将所有微服务从手动初始化迁移到 `pkg/app` Bootstrap 框架。

## 目录

- [迁移收益](#迁移收益)
- [迁移前后对比](#迁移前后对比)
- [通用迁移步骤](#通用迁移步骤)
- [迁移示例](#迁移示例)
- [特殊场景处理](#特殊场景处理)
- [迁移清单](#迁移清单)

---

## 迁移收益

### 代码减少
- **平均减少 26% 的样板代码**（notification-service: 345行 → 254行）
- 消除重复的初始化逻辑
- 统一配置模式

### 自动获得的企业级功能

✅ **基础设施**
- 数据库连接池（PostgreSQL）
- Redis 连接和缓存
- 结构化日志（Zap）
- Gin HTTP 路由器

✅ **可观测性**
- Jaeger 分布式追踪（W3C Trace Context）
- Prometheus 指标收集（`/metrics`）
- 增强型健康检查（`/health`, `/health/live`, `/health/ready`）

✅ **安全性与稳定性**
- 速率限制（基于 Redis）
- CORS 支持
- 请求 ID 传播
- Panic 恢复

✅ **优雅关闭**
- SIGINT/SIGTERM 信号处理
- 资源清理（数据库、Redis、gRPC）
- HTTP 和 gRPC 双协议支持

---

## 迁移前后对比

### 手动初始化（旧模式）

```go
func main() {
    // 1. 初始化日志
    env := config.GetEnv("ENV", "development")
    if err := logger.InitLogger(env); err != nil {
        log.Fatalf("初始化日志失败: %v", err)
    }
    defer logger.Sync()

    // 2. 初始化数据库
    dbConfig := db.Config{
        Host:     config.GetEnv("DB_HOST", "localhost"),
        Port:     config.GetEnvInt("DB_PORT", 5432),
        User:     config.GetEnv("DB_USER", "postgres"),
        Password: config.GetEnv("DB_PASSWORD", "postgres"),
        DBName:   config.GetEnv("DB_NAME", "payment_admin"),
        SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
        TimeZone: config.GetEnv("DB_TIMEZONE", "UTC"),
    }
    database, err := db.NewPostgresDB(dbConfig)
    if err != nil {
        logger.Fatal("数据库连接失败", zap.Error(err))
    }

    // 3. 自动迁移
    if err := database.AutoMigrate(&model.Admin{}, ...); err != nil {
        logger.Fatal("数据库迁移失败", zap.Error(err))
    }

    // 4. 初始化 Redis
    redisConfig := db.RedisConfig{
        Host:     config.GetEnv("REDIS_HOST", "localhost"),
        Port:     config.GetEnvInt("REDIS_PORT", 6379),
        Password: config.GetEnv("REDIS_PASSWORD", ""),
        DB:       config.GetEnvInt("REDIS_DB", 0),
    }
    redisClient, err := db.NewRedisClient(redisConfig)
    if err != nil {
        logger.Fatal("Redis连接失败", zap.Error(err))
    }

    // 5. 初始化 Prometheus 指标
    httpMetrics := metrics.NewHTTPMetrics("admin_service")

    // 6. 初始化 Jaeger 追踪
    jaegerEndpoint := config.GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
    samplingRate := float64(config.GetEnvInt("JAEGER_SAMPLING_RATE", 100)) / 100.0
    tracerShutdown, err := tracing.InitTracer(tracing.Config{
        ServiceName:    "admin-service",
        ServiceVersion: "1.0.0",
        Environment:    env,
        JaegerEndpoint: jaegerEndpoint,
        SamplingRate:   samplingRate,
    })
    if err != nil {
        logger.Error(fmt.Sprintf("Jaeger 初始化失败: %v", err))
    } else {
        defer tracerShutdown(context.Background())
    }

    // 7. 创建 Gin 路由器
    router := gin.Default()
    router.Use(middleware.CORSMiddleware())
    router.Use(middleware.RequestIDMiddleware())
    router.Use(middleware.LoggerMiddleware())
    router.Use(middleware.TracingMiddleware("admin-service"))
    router.Use(middleware.MetricsMiddleware())

    // 8. 健康检查端点
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 9. Prometheus 指标端点
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // 10. 初始化业务层 repositories, services, handlers
    // ...

    // 11. 注册路由
    // ...

    // 12. 启动 HTTP 服务器
    port := config.GetEnvInt("PORT", 8001)
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: router,
    }

    // 13. 优雅关闭（手动实现）
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-quit
        logger.Info("正在关闭服务...")
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := srv.Shutdown(ctx); err != nil {
            logger.Fatal("服务关闭失败", zap.Error(err))
        }
    }()

    logger.Info(fmt.Sprintf("服务启动在端口 %d", port))
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Fatal("服务启动失败", zap.Error(err))
    }
}
```

**问题**:
- 大量重复代码（每个服务都要写一遍）
- 容易遗漏配置（例如忘记添加某个中间件）
- 缺乏统一标准
- 维护成本高

---

### Bootstrap 框架（新模式）

```go
func main() {
    // 1. 一键初始化所有基础设施
    application, err := app.Bootstrap(app.ServiceConfig{
        ServiceName: "admin-service",
        DBName:      "payment_admin",
        Port:        40001,

        // 自动迁移模型
        AutoMigrate: []any{
            &model.Admin{},
            &model.Role{},
            &model.Permission{},
        },

        // 功能开关（可选，有合理默认值）
        EnableTracing:     true,  // Jaeger 追踪
        EnableMetrics:     true,  // Prometheus 指标
        EnableRedis:       true,  // Redis 连接
        EnableGRPC:        false, // gRPC 默认关闭（系统使用 HTTP/REST）
        EnableHealthCheck: true,  // 健康检查
        EnableRateLimit:   true,  // 速率限制

        // 速率限制配置
        RateLimitRequests: 100,
        RateLimitWindow:   time.Minute,
    })
    if err != nil {
        log.Fatalf("Bootstrap 失败: %v", err)
    }

    logger.Info("正在启动 Admin Service...")

    // 2. 初始化业务层（使用 application.DB, application.RedisClient）
    adminRepo := repository.NewAdminRepository(application.DB)
    roleRepo := repository.NewRoleRepository(application.DB)
    // ...

    adminService := service.NewAdminService(adminRepo, roleRepo, application.RedisClient)
    // ...

    // 3. 创建 handler 并注册路由
    adminHandler := handler.NewAdminHandler(adminService)

    jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
    jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
    authMiddleware := middleware.AuthMiddleware(jwtManager)

    adminHandler.RegisterRoutes(application.Router, authMiddleware)

    // 4. （可选）启用 gRPC（默认不需要）
    // if application.GRPCServer != nil {
    //     adminGrpcServer := grpcServer.NewAdminServer(adminService)
    //     pb.RegisterAdminServiceServer(application.GRPCServer, adminGrpcServer)
    // }

    // 5. 启动服务（自动处理优雅关闭）
    if err := application.RunWithGracefulShutdown(); err != nil {
        logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
    }
}
```

**优势**:
- 代码简洁清晰（专注业务逻辑）
- 统一配置标准
- 自动获得所有企业级功能
- 易于维护和升级

---

## 通用迁移步骤

### 1. 导入 Bootstrap 包

```go
import (
    "github.com/payment-platform/pkg/app"
    "github.com/payment-platform/pkg/config"
    "github.com/payment-platform/pkg/logger"
    "github.com/payment-platform/pkg/auth"
    "github.com/payment-platform/pkg/middleware"
    // ... 其他业务包
)
```

### 2. 删除手动初始化代码

移除以下代码块：
- ❌ `logger.InitLogger()` 和 `defer logger.Sync()`
- ❌ `db.NewPostgresDB()` 和数据库配置
- ❌ `database.AutoMigrate()`
- ❌ `db.NewRedisClient()` 和 Redis 配置
- ❌ `metrics.NewHTTPMetrics()`
- ❌ `tracing.InitTracer()` 和 `defer tracerShutdown()`
- ❌ `gin.Default()` 和中间件注册
- ❌ 手动健康检查和 `/metrics` 端点
- ❌ 手动 HTTP 服务器创建和优雅关闭逻辑

### 3. 添加 Bootstrap 配置

```go
application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName: "service-name",      // 服务名称
    DBName:      "database_name",     // 数据库名
    Port:        40001,               // HTTP 端口

    // 自动迁移模型
    AutoMigrate: []any{
        &model.ModelA{},
        &model.ModelB{},
    },

    // 功能开关
    EnableTracing:     true,
    EnableMetrics:     true,
    EnableRedis:       true,
    EnableGRPC:        false,  // 默认关闭，系统使用 HTTP
    EnableHealthCheck: true,
    EnableRateLimit:   true,

    RateLimitRequests: 100,
    RateLimitWindow:   time.Minute,
})
if err != nil {
    log.Fatalf("Bootstrap 失败: %v", err)
}
```

### 4. 使用 application 提供的资源

```go
// 使用 application.DB 替代 database
adminRepo := repository.NewAdminRepository(application.DB)

// 使用 application.RedisClient 替代 redisClient
service := service.NewService(repo, application.RedisClient)

// 使用 application.Router 替代 router
handler.RegisterRoutes(application.Router, authMiddleware)

// （可选）使用 application.GRPCServer
if application.GRPCServer != nil {
    pb.RegisterServiceServer(application.GRPCServer, grpcImpl)
}
```

### 5. 替换启动方法

```go
// ❌ 旧方式
// srv := &http.Server{...}
// srv.ListenAndServe()

// ✅ 新方式（仅 HTTP）
if err := application.RunWithGracefulShutdown(); err != nil {
    logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
}

// 或者（HTTP + gRPC 双协议）
// if err := application.RunDualProtocol(); err != nil {
//     logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
// }
```

---

## 迁移示例

### 示例 1: Admin Service（标准服务）

**迁移前** (admin-service/cmd/main.go - 约 280 行):
```go
func main() {
    // 50+ 行：日志、数据库、Redis 初始化
    // 30+ 行：指标、追踪初始化
    // 20+ 行：Gin 路由器和中间件
    // 10+ 行：健康检查和指标端点
    // ...
    // 30+ 行：优雅关闭逻辑
}
```

**迁移后** (约 200 行，减少 28%):
```go
func main() {
    application, err := app.Bootstrap(app.ServiceConfig{
        ServiceName: "admin-service",
        DBName:      "payment_admin",
        Port:        40001,
        AutoMigrate: []any{
            &model.Admin{},
            &model.Role{},
            &model.Permission{},
            &model.AdminRole{},
            &model.RolePermission{},
            &model.AuditLog{},
            &model.SystemConfig{},
            &model.MerchantReview{},
            &model.ApprovalFlow{},
        },
        EnableTracing:     true,
        EnableMetrics:     true,
        EnableRedis:       true,
        EnableGRPC:        false,
        EnableHealthCheck: true,
        EnableRateLimit:   true,
        RateLimitRequests: 100,
        RateLimitWindow:   time.Minute,
    })
    if err != nil {
        log.Fatalf("Bootstrap 失败: %v", err)
    }

    logger.Info("正在启动 Admin Service...")

    // 初始化邮件客户端（业务特定）
    emailClient, err := email.NewClient(&email.Config{
        Provider:     "smtp",
        SMTPHost:     config.GetEnv("SMTP_HOST", "smtp.gmail.com"),
        SMTPPort:     config.GetEnvInt("SMTP_PORT", 587),
        SMTPUsername: config.GetEnv("SMTP_USERNAME", ""),
        SMTPPassword: config.GetEnv("SMTP_PASSWORD", ""),
        SMTPFrom:     config.GetEnv("SMTP_FROM", "noreply@payment-platform.com"),
    })
    if err != nil {
        logger.Warn("SMTP 邮件客户端初始化失败", zap.Error(err))
    }

    // 初始化 Repository
    adminRepo := repository.NewAdminRepository(application.DB)
    roleRepo := repository.NewRoleRepository(application.DB)
    permissionRepo := repository.NewPermissionRepository(application.DB)
    auditLogRepo := repository.NewAuditLogRepository(application.DB)
    systemConfigRepo := repository.NewSystemConfigRepository(application.DB)
    merchantReviewRepo := repository.NewMerchantReviewRepository(application.DB)

    // 初始化 Service
    jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
    jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

    adminService := service.NewAdminService(adminRepo, roleRepo, jwtManager, application.RedisClient, emailClient)
    roleService := service.NewRoleService(roleRepo, permissionRepo)
    permissionService := service.NewPermissionService(permissionRepo)
    auditLogService := service.NewAuditLogService(auditLogRepo)
    systemConfigService := service.NewSystemConfigService(systemConfigRepo, application.RedisClient)
    merchantReviewService := service.NewMerchantReviewService(merchantReviewRepo, auditLogRepo, emailClient)

    // 初始化 Handler
    adminHandler := handler.NewAdminHandler(adminService, auditLogService)
    roleHandler := handler.NewRoleHandler(roleService)
    permissionHandler := handler.NewPermissionHandler(permissionService)
    auditLogHandler := handler.NewAuditLogHandler(auditLogService)
    systemConfigHandler := handler.NewSystemConfigHandler(systemConfigService)
    merchantReviewHandler := handler.NewMerchantReviewHandler(merchantReviewService)

    // Swagger UI（公开接口）
    application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // JWT 认证中间件
    authMiddleware := middleware.AuthMiddleware(jwtManager)

    // 注册路由（带认证）
    adminHandler.RegisterRoutes(application.Router, authMiddleware)
    roleHandler.RegisterRoutes(application.Router, authMiddleware)
    permissionHandler.RegisterRoutes(application.Router, authMiddleware)
    auditLogHandler.RegisterRoutes(application.Router, authMiddleware)
    systemConfigHandler.RegisterRoutes(application.Router, authMiddleware)
    merchantReviewHandler.RegisterRoutes(application.Router, authMiddleware)

    // 启动服务（仅 HTTP，优雅关闭）
    if err := application.RunWithGracefulShutdown(); err != nil {
        logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
    }
}
```

---

### 示例 2: Payment Gateway（复杂服务，带自定义中间件）

```go
func main() {
    application, err := app.Bootstrap(app.ServiceConfig{
        ServiceName: "payment-gateway",
        DBName:      "payment_gateway",
        Port:        40003,
        AutoMigrate: []any{
            &model.Payment{},
            &model.PaymentEvent{},
            &model.Refund{},
            &model.ApiKey{},
        },
        EnableTracing:     true,
        EnableMetrics:     true,
        EnableRedis:       true,
        EnableGRPC:        false,
        EnableHealthCheck: true,
        EnableRateLimit:   true,
        RateLimitRequests: 200,
        RateLimitWindow:   time.Minute,
    })
    if err != nil {
        log.Fatalf("Bootstrap 失败: %v", err)
    }

    logger.Info("正在启动 Payment Gateway...")

    // 初始化 HTTP 客户端（用于调用其他服务）
    orderServiceURL := config.GetEnv("ORDER_SERVICE_URL", "http://localhost:40004")
    channelServiceURL := config.GetEnv("CHANNEL_SERVICE_URL", "http://localhost:40005")
    riskServiceURL := config.GetEnv("RISK_SERVICE_URL", "http://localhost:40006")

    orderClient := client.NewOrderClient(orderServiceURL)
    channelClient := client.NewChannelClient(channelServiceURL)
    riskClient := client.NewRiskClient(riskServiceURL)

    // 初始化 Repository
    paymentRepo := repository.NewPaymentRepository(application.DB)
    refundRepo := repository.NewRefundRepository(application.DB)
    apiKeyRepo := repository.NewApiKeyRepository(application.DB)

    // 初始化 Service
    paymentService := service.NewPaymentService(
        paymentRepo,
        orderClient,
        channelClient,
        riskClient,
        application.RedisClient,
    )
    refundService := service.NewRefundService(refundRepo, paymentRepo, channelClient)
    apiKeyService := service.NewApiKeyService(apiKeyRepo, application.RedisClient)

    // 初始化 Handler
    paymentHandler := handler.NewPaymentHandler(paymentService)
    refundHandler := handler.NewRefundHandler(refundService)
    webhookHandler := handler.NewWebhookHandler(paymentService)

    // Swagger UI
    application.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // 公开路由（Webhook，不需要签名验证）
    public := application.Router.Group("/api/v1/webhooks")
    webhookHandler.RegisterWebhookRoutes(public)

    // 受保护路由（需要签名验证）
    secretFetcher := func(merchantID string) (string, error) {
        return apiKeyService.GetSecretByMerchantID(context.Background(), merchantID)
    }
    signatureMiddleware := localMiddleware.NewSignatureMiddleware(secretFetcher)

    protected := application.Router.Group("/api/v1")
    protected.Use(signatureMiddleware.Verify())
    paymentHandler.RegisterRoutes(protected)
    refundHandler.RegisterRoutes(protected)

    // 启动服务
    if err := application.RunWithGracefulShutdown(); err != nil {
        logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
    }
}
```

---

## 特殊场景处理

### 场景 1: 服务需要 Kafka

```go
application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName:   "notification-service",
    DBName:        "payment_notification",
    Port:          40008,
    EnableRedis:   true,
    // ... 其他配置
})
if err != nil {
    log.Fatalf("Bootstrap 失败: %v", err)
}

// Kafka 由业务层自行管理（Bootstrap 不包含 Kafka）
kafkaEnabled := config.GetEnv("KAFKA_ENABLE_ASYNC", "false") == "true"
if kafkaEnabled {
    kafkaBrokers := strings.Split(config.GetEnv("KAFKA_BROKERS", "localhost:9092"), ",")

    emailProducer := kafka.NewProducer(kafka.ProducerConfig{
        Brokers: kafkaBrokers,
        Topic:   "notifications.email",
    })

    // 创建 Kafka workers...
}
```

### 场景 2: 服务需要外部 HTTP 客户端

```go
application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName: "channel-adapter",
    DBName:      "payment_channel",
    Port:        40005,
})
if err != nil {
    log.Fatalf("Bootstrap 失败: %v", err)
}

// 创建外部服务客户端（使用 pkg/httpclient 或直接创建）
exchangeRateClient := client.NewExchangeRateClient(
    config.GetEnv("EXCHANGE_RATE_API_URL", "https://api.exchangerate.host"),
)

// Stripe、PayPal 等适配器初始化
stripeConfig := &model.StripeConfig{
    APIKey:        config.GetEnv("STRIPE_API_KEY", ""),
    WebhookSecret: config.GetEnv("STRIPE_WEBHOOK_SECRET", ""),
}
stripeAdapter := adapter.NewStripeAdapter(stripeConfig)

adapterFactory := adapter.NewAdapterFactory()
adapterFactory.Register(model.ChannelStripe, stripeAdapter)
```

### 场景 3: 服务需要后台任务

```go
application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName: "settlement-service",
    DBName:      "payment_settlement",
    Port:        40013,
})
if err != nil {
    log.Fatalf("Bootstrap 失败: %v", err)
}

// 初始化业务层
settlementService := service.NewSettlementService(/* ... */)

// 启动后台任务
go startBackgroundWorkers(settlementService)

// 启动服务
if err := application.RunWithGracefulShutdown(); err != nil {
    logger.Fatal(fmt.Sprintf("服务启动失败: %v", err))
}

func startBackgroundWorkers(service service.SettlementService) {
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()

    ctx := context.Background()
    for range ticker.C {
        if err := service.ProcessPendingSettlements(ctx); err != nil {
            logger.Error(fmt.Sprintf("处理待结算失败: %v", err))
        }
    }
}
```

### 场景 4: 服务需要自定义中间件

```go
application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName: "payment-gateway",
    DBName:      "payment_gateway",
    Port:        40003,
})
if err != nil {
    log.Fatalf("Bootstrap 失败: %v", err)
}

// 添加自定义中间件（在 Bootstrap 默认中间件之后）
secretFetcher := func(merchantID string) (string, error) {
    return apiKeyService.GetSecretByMerchantID(context.Background(), merchantID)
}
signatureMiddleware := localMiddleware.NewSignatureMiddleware(secretFetcher)

// 受保护路由
protected := application.Router.Group("/api/v1")
protected.Use(signatureMiddleware.Verify())
paymentHandler.RegisterRoutes(protected)
```

---

## 迁移清单

### Phase 1: 核心服务（优先级最高）

- [ ] admin-service (40001)
- [ ] merchant-service (40002)
- [ ] config-service (40010)

### Phase 2: 支付核心（依赖 Phase 1）

- [ ] payment-gateway (40003) - 最复杂，需要自定义签名中间件
- [ ] order-service (40004)
- [ ] channel-adapter (40005)
- [ ] risk-service (40006)

### Phase 3: 辅助服务

- [ ] accounting-service (40007)
- [ ] analytics-service (40009)
- [ ] merchant-auth-service (40011)
- [ ] settlement-service (40013)
- [ ] withdrawal-service (40014)
- [ ] kyc-service (40015)
- [ ] cashier-service (未指定端口)

### 已完成

- [x] notification-service (40008) - 参考实现

---

## 环境变量

Bootstrap 框架使用以下环境变量（可选，有默认值）：

```bash
# 基础配置
ENV=development                              # development/production
PORT=40001                                   # HTTP 端口
GRPC_PORT=50001                              # gRPC 端口（如果启用）

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_xxx                          # 由 Bootstrap 配置指定
DB_SSL_MODE=disable
DB_TIMEZONE=UTC

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Jaeger 配置
JAEGER_ENDPOINT=http://localhost:14268/api/traces
JAEGER_SAMPLING_RATE=100                     # 0-100 (生产环境建议 10-20)

# 业务配置（各服务自定义）
JWT_SECRET=your-secret-key
STRIPE_API_KEY=sk_test_xxx
KAFKA_BROKERS=localhost:9092
```

---

## 常见问题

### Q1: 迁移后服务无法启动？

**A**: 检查以下内容：
1. `go.mod` 是否包含 `github.com/payment-platform/pkg` 依赖
2. 运行 `go mod tidy` 更新依赖
3. 检查环境变量是否正确设置
4. 查看日志中的详细错误信息

### Q2: 如何禁用某个功能（如 Redis）？

**A**: 在 Bootstrap 配置中设置 `EnableRedis: false`。但注意：
- 速率限制依赖 Redis（需同时设置 `EnableRateLimit: false`）
- 某些业务逻辑可能依赖 Redis（需修改代码）

### Q3: 迁移后如何测试？

**A**:
1. 编译测试: `go build ./cmd/main.go`
2. 运行服务: `./main` 或 `go run cmd/main.go`
3. 检查健康端点: `curl http://localhost:40001/health`
4. 检查指标端点: `curl http://localhost:40001/metrics`
5. 测试业务 API（使用 Postman 或 curl）

### Q4: 迁移后性能会变差吗？

**A**: 不会。Bootstrap 框架仅封装初始化逻辑，运行时性能与手动初始化完全相同。实际上，统一配置可以更容易地优化性能（如调整连接池大小）。

### Q5: 可以逐步迁移吗？

**A**: 可以！建议按 Phase 1 → Phase 2 → Phase 3 的顺序逐个迁移。已迁移和未迁移的服务可以并存。

---

## 参考资料

- Bootstrap 框架源码: `backend/pkg/app/bootstrap.go`
- 参考实现: `backend/services/notification-service/cmd/main.go`
- 单元测试: `backend/pkg/app/bootstrap_test.go`

---

## 迁移支持

如有问题，请查看：
1. 本文档的"特殊场景处理"章节
2. notification-service 的完整实现
3. Bootstrap 框架的代码注释

**祝迁移顺利！** 🚀

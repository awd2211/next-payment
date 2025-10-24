# Bootstrap Framework Quick Reference

## Migration Summary (Phase 2 Complete)

### Services Migrated ✅

| Service | Lines Reduced | Status | Port (HTTP/gRPC) |
|---------|---------------|--------|------------------|
| cashier-service | ~60 (26%) | ✅ Complete | 40016 / 50016 |
| kyc-service | ~70 (31%) | ✅ Complete | 40015 / 50015 |
| merchant-auth-service | 65 (29%) | ✅ Complete | 40011 / 50011 |
| settlement-service | 65 (31%) | ✅ Complete | 40013 / 50013 |
| withdrawal-service | 63 (29%) | ✅ Complete | 40014 / 50014 |

**Total**: 5 services, ~320+ lines saved, 30% average reduction

---

## Bootstrap Template

```go
package main

import (
    "log"
    "time"
    
    "github.com/payment-platform/pkg/app"
    "github.com/payment-platform/pkg/config"
    "github.com/payment-platform/pkg/logger"
    "payment-platform/YOUR-SERVICE/internal/model"
    // ... other imports
)

func main() {
    // 1. Bootstrap初始化
    application, err := app.Bootstrap(app.ServiceConfig{
        ServiceName: "your-service",
        DBName:      config.GetEnv("DB_NAME", "payment_your_db"),
        Port:        config.GetEnvInt("PORT", 40XXX),
        
        AutoMigrate: []any{
            &model.YourModel{},
            // ... more models
        },
        
        EnableTracing:     true,
        EnableMetrics:     true,
        EnableRedis:       true,
        EnableGRPC:        true,  // Set to false if no gRPC
        GRPCPort:          config.GetEnvInt("GRPC_PORT", 50XXX),
        EnableHealthCheck: true,
        EnableRateLimit:   true,
        
        RateLimitRequests: 100,
        RateLimitWindow:   time.Minute,
    })
    if err != nil {
        log.Fatalf("Bootstrap失败: %v", err)
    }
    
    // 2. Initialize clients
    // yourClient := client.NewYourClient(url)
    
    // 3. Initialize repositories
    // repo := repository.NewYourRepository(application.DB)
    
    // 4. Initialize services
    // service := service.NewYourService(repo, yourClient)
    
    // 5. Initialize handlers
    // handler := handler.NewYourHandler(service)
    
    // 6. Register routes
    // handler.RegisterRoutes(application.Router)
    
    // 7. (Optional) Register gRPC services
    // pb.RegisterYourServiceServer(application.GRPCServer, grpcImpl)
    
    // 8. Start service
    if err := application.RunDualProtocol(); err != nil {  // or RunWithGracefulShutdown()
        logger.Fatal("服务启动失败: " + err.Error())
    }
}
```

---

## Available Fields

### Application Fields (use after Bootstrap)

```go
application.DB            // *gorm.DB - PostgreSQL connection
application.Redis         // *redis.Client - Redis connection
application.Router        // *gin.Engine - HTTP router
application.GRPCServer    // *grpc.Server - gRPC server (if enabled)
application.Logger        // *zap.Logger - Structured logger
application.HealthChecker // *health.HealthChecker - Health checker
application.Config        // ServiceConfig - Original config
application.Environment   // string - Current environment
```

---

## Auto-Configured Middleware

Bootstrap automatically adds:

1. **CORS**: Cross-origin resource sharing
2. **RequestID**: Unique request tracking
3. **TracingMiddleware**: Jaeger distributed tracing
4. **Logger**: Structured request logging
5. **MetricsMiddleware**: Prometheus HTTP metrics
6. **RateLimiter**: Redis-backed rate limiting (if enabled)

---

## Auto-Created Endpoints

- `GET /health` - Enhanced health check with dependencies
- `GET /metrics` - Prometheus metrics endpoint
- `GET /swagger/*any` - Add manually if needed

---

## Common Patterns

### Pattern 1: HTTP-only Service

```go
application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName: "your-service",
    DBName:      "payment_your_db",
    Port:        40XXX,
    EnableGRPC:  false,  // ← Disable gRPC
    // ... other config
})

// Register routes
handler.RegisterRoutes(application.Router)

// Start HTTP only
if err := application.RunWithGracefulShutdown(); err != nil {
    logger.Fatal("服务启动失败: " + err.Error())
}
```

### Pattern 2: HTTP + gRPC Dual Protocol

```go
application, err := app.Bootstrap(app.ServiceConfig{
    EnableGRPC: true,
    GRPCPort:   50XXX,
    // ... other config
})

// Register gRPC services
pb.RegisterYourServiceServer(application.GRPCServer, grpcImpl)

// Start both protocols
if err := application.RunDualProtocol(); err != nil {
    logger.Fatal("服务启动失败: " + err.Error())
}
```

### Pattern 3: Add Custom Middleware

```go
// After Bootstrap
idempotencyMgr := idempotency.NewIdempotencyManager(
    application.Redis,
    "your-service",
    24*time.Hour,
)
application.Router.Use(middleware.IdempotencyMiddleware(idempotencyMgr))
```

### Pattern 4: Background Tasks

```go
// After Bootstrap, before Run
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        // Your periodic task
        service.DoPeriodicWork(context.Background())
    }
}()
```

---

## Migration Checklist

- [ ] Backup original `cmd/main.go` to `cmd/main.go.backup`
- [ ] Copy models from original `AutoMigrate()` call
- [ ] Identify all client dependencies (`NewXxxClient`)
- [ ] Identify custom middleware beyond standard stack
- [ ] Identify background tasks (goroutines)
- [ ] Set `EnableGRPC` based on gRPC server presence
- [ ] Use `application.DB`, `application.Redis`, `application.Router`
- [ ] Register gRPC to `application.GRPCServer` (if enabled)
- [ ] Choose `RunDualProtocol()` vs `RunWithGracefulShutdown()`
- [ ] Test compilation: `go build -o /tmp/service ./cmd/main.go`
- [ ] Verify binary created successfully
- [ ] Compare line counts: `wc -l main.go.backup main.go`

---

## Environment Variables

Bootstrap reads these standard variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_your_db
DB_SSL_MODE=disable
DB_TIMEZONE=UTC

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Service
ENV=development  # or production
PORT=40XXX
GRPC_PORT=50XXX

# Observability
JAEGER_ENDPOINT=http://localhost:14268/api/traces
JAEGER_SAMPLING_RATE=100  # 0-100
```

---

## Troubleshooting

### Error: "undefined: application.RedisClient"

**Fix**: Use `application.Redis` (not `RedisClient`)

### Error: "missing method ... on type *App"

**Cause**: App struct doesn't have that method.
**Fix**: Check `/home/eric/payment/backend/pkg/app/bootstrap.go` for correct field names.

### Error: "gRPC server not started"

**Cause**: `EnableGRPC: false` but using `RunDualProtocol()`
**Fix**: Either set `EnableGRPC: true` or use `RunWithGracefulShutdown()`

### Error: "rate limiter failed: redis not connected"

**Cause**: `EnableRateLimit: true` but `EnableRedis: false`
**Fix**: Set `EnableRedis: true` when using rate limiting

---

## Benefits Summary

### Code Reduction
- Average: **30%** fewer lines
- Range: 26-31% across migrated services
- Total saved: **320+ lines** (5 services)

### Features Gained (11 Total)
1. Unified logger with graceful shutdown
2. Database connection pooling + health checks
3. Redis connection management
4. Prometheus metrics (/metrics)
5. Jaeger distributed tracing (W3C)
6. Full middleware stack (CORS, RequestID, etc.)
7. Rate limiting (Redis-backed)
8. Enhanced health checks (/health)
9. Graceful shutdown (SIGINT/SIGTERM)
10. gRPC server management
11. Dual-protocol support (HTTP + gRPC)

### Preserved
- ✅ 100% business logic
- ✅ All client dependencies
- ✅ Custom middleware
- ✅ Background tasks
- ✅ Service-specific configuration

---

**Last Updated**: 2025-10-24
**Services Migrated**: 5 / 15 (33%)
**Next Target**: channel-adapter, risk-service, merchant-config-service

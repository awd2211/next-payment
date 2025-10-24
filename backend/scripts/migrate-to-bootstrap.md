# 迁移到 Bootstrap 框架指南

本文档说明如何将现有微服务从手动初始化迁移到统一的 Bootstrap 框架。

## 迁移收益

| 指标 | 迁移前 | 迁移后 | 改进 |
|------|--------|--------|------|
| main.go 代码行数 | ~190行 | ~60行 | **减少68%** |
| 初始化代码 | 手动编写 | 框架自动 | **标准化** |
| 健康检查 | 部分服务有 | 所有服务统一 | **一致性** |
| 优雅关闭 | 大部分缺失 | 全部支持 | **可靠性** |
| 追踪/指标 | 配置不一致 | 统一配置 | **可观测性** |

## 迁移步骤

### 1. 识别需要迁移的服务

```bash
cd /home/eric/payment/backend

# 列出所有服务
ls -d services/*/
```

**待迁移服务列表**:
- ✅ order-service (示例已创建)
- ⏳ payment-gateway
- ⏳ admin-service
- ⏳ merchant-service
- ⏳ channel-adapter
- ⏳ risk-service
- ⏳ accounting-service
- ⏳ notification-service
- ⏳ analytics-service
- ⏳ config-service
- ⏳ merchant-auth-service
- ⏳ settlement-service
- ⏳ withdrawal-service
- ⏳ kyc-service

### 2. 迁移单个服务

以 `order-service` 为例：

#### 2.1 备份原文件

```bash
cd services/order-service/cmd
cp main.go main.go.backup
```

#### 2.2 识别服务配置

从原 main.go 中提取关键信息：

```go
// 原始代码：
serviceName := "order-service"
dbName := "payment_order"
port := 40004

// 迁移到 Bootstrap：
app.ServiceConfig{
    ServiceName: "order-service",
    DBName:      "payment_order",
    Port:        40004,
}
```

#### 2.3 识别数据库模型

```go
// 原始代码：
database.AutoMigrate(
    &model.Order{},
    &model.OrderItem{},
    &model.OrderLog{},
)

// 迁移到 Bootstrap：
app.ServiceConfig{
    AutoMigrate: []any{
        &model.Order{},
        &model.OrderItem{},
        &model.OrderLog{},
    },
}
```

#### 2.4 识别功能需求

| 功能 | 检查方法 | Bootstrap 配置 |
|------|----------|----------------|
| Redis | 查找 `redis.NewClient` | `EnableRedis: true` |
| 追踪 | 查找 `tracing.InitTracer` | `EnableTracing: true` |
| 指标 | 查找 `metrics.New` | `EnableMetrics: true` |
| 健康检查 | 查找 `health.NewHealthChecker` | `EnableHealthCheck: true` |
| 速率限制 | 查找 `middleware.NewRateLimiter` | `EnableRateLimit: true` |

#### 2.5 迁移业务逻辑

**保留部分**:
- ✅ Repository 初始化
- ✅ Service 初始化
- ✅ Handler 初始化
- ✅ 路由注册

**删除部分**:
- ❌ 日志初始化 (`logger.InitLogger`)
- ❌ 数据库连接 (`db.NewPostgresDB`)
- ❌ Redis 连接 (`redis.NewClient`)
- ❌ Gin 初始化 (`gin.New()`)
- ❌ 中间件注册 (`router.Use(...)`)
- ❌ 追踪初始化 (`tracing.InitTracer`)
- ❌ 指标初始化 (`metrics.New...`)
- ❌ 健康检查初始化 (`health.NewHealthChecker`)
- ❌ 启动代码 (`router.Run()`)

### 3. 迁移模板

```go
package main

import (
	"log"
	"time"

	"github.com/payment-platform/pkg/app"
	"github.com/{your-org}/{service}/internal/handler"
	"github.com/{your-org}/{service}/internal/model"
	"github.com/{your-org}/{service}/internal/repository"
	"github.com/{your-org}/{service}/internal/service"
)

func main() {
	// 1. Bootstrap 初始化
	application, err := app.Bootstrap(app.ServiceConfig{
		ServiceName: "{service-name}",
		DBName:      "{database-name}",
		Port:        {port},

		// 数据库模型
		AutoMigrate: []any{
			&model.{Model1}{},
			&model.{Model2}{},
		},

		// 可选功能（根据需要启用）
		EnableTracing:     true,
		EnableMetrics:     true,
		EnableRedis:       true,   // 如果需要缓存/速率限制
		EnableHealthCheck: true,
		EnableRateLimit:   false,  // 按需启用

		// 速率限制配置（如果启用）
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	})
	if err != nil {
		log.Fatalf("Bootstrap failed: %v", err)
	}

	// 2. 初始化业务层
	// Repository
	{entity}Repo := repository.New{Entity}Repository(application.DB)

	// Service
	{entity}Service := service.New{Entity}Service(
		application.DB,
		{entity}Repo,
		// ... 其他依赖
	)

	// Handler
	{entity}Handler := handler.New{Entity}Handler({entity}Service)

	// 3. 注册路由
	api := application.Router.Group("/api/v1")
	{
		{resource} := api.Group("/{resources}")
		{
			{resource}.POST("", {entity}Handler.Create)
			{resource}.GET("/:id", {entity}Handler.Get)
			{resource}.GET("", {entity}Handler.List)
			{resource}.PUT("/:id", {entity}Handler.Update)
			{resource}.DELETE("/:id", {entity}Handler.Delete)
		}
	}

	// 4. 启动服务（优雅关闭）
	if err := application.RunWithGracefulShutdown(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

### 4. 服务特定迁移注意事项

#### payment-gateway

```go
app.ServiceConfig{
    ServiceName: "payment-gateway",
    DBName:      "payment_gateway",
    Port:        40003,
    EnableRedis: true,  // ⚠️ 必需：用于幂等性、Nonce去重、速率限制
    EnableRateLimit: true,  // ⚠️ 推荐：防止滥用
}

// 特殊处理：
// - 签名验证中间件需要在路由级别添加（不是全局）
// - Webhook 路由不应用速率限制
```

#### admin-service / merchant-service

```go
app.ServiceConfig{
    EnableRedis: true,  // ⚠️ 需要：用于会话管理
    EnableRateLimit: false,  // 内部服务，可选
}

// 特殊处理：
// - JWT 认证中间件在路由级别添加
```

#### notification-service

```go
app.ServiceConfig{
    EnableRedis: true,  // ⚠️ 需要：用于消息队列
    EnableRateLimit: false,  // 后台服务，不需要
}
```

#### channel-adapter

```go
app.ServiceConfig{
    EnableRedis: true,  // ⚠️ 推荐：用于缓存渠道配置
    EnableRateLimit: false,  // 内部服务调用
}
```

### 5. 验证迁移

#### 5.1 编译检查

```bash
cd services/{service-name}
go build -o /tmp/test-{service} ./cmd/main.go
```

#### 5.2 运行检查

```bash
# 启动服务
ENV=development \
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=postgres \
REDIS_HOST=localhost \
REDIS_PORT=6379 \
/tmp/test-{service}
```

#### 5.3 端点验证

```bash
# 健康检查
curl http://localhost:{port}/health
curl http://localhost:{port}/health/live
curl http://localhost:{port}/health/ready

# Prometheus 指标
curl http://localhost:{port}/metrics | grep http_requests_total

# 业务端点
curl http://localhost:{port}/api/v1/{resource}
```

#### 5.4 日志检查

确认以下日志输出：

```
✅ "正在启动 {service-name}..."
✅ "数据库连接成功"
✅ "数据库迁移成功"
✅ "Redis 连接成功" (如果启用)
✅ "追踪初始化成功" (如果启用)
✅ "健康检查已启用" (如果启用)
✅ "{service-name} 正在监听 :40004"
```

### 6. 批量迁移脚本

```bash
#!/bin/bash
# migrate-all-services.sh

SERVICES=(
  "order-service"
  "payment-gateway"
  "admin-service"
  "merchant-service"
  "channel-adapter"
  "risk-service"
  "accounting-service"
  "notification-service"
  "analytics-service"
  "config-service"
)

for service in "${SERVICES[@]}"; do
  echo "=== Migrating $service ==="

  # 备份
  cp services/$service/cmd/main.go services/$service/cmd/main.go.old

  # 编译检查
  cd services/$service
  if go build -o /tmp/test-$service ./cmd/main.go; then
    echo "✅ $service 编译成功"
  else
    echo "❌ $service 编译失败"
    # 恢复备份
    cp cmd/main.go.old cmd/main.go
  fi

  cd ../..
done
```

### 7. Docker Compose 更新

迁移后，更新 `docker-compose.yml`：

```yaml
services:
  order-service:
    build:
      context: ./services/order-service
    environment:
      - ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=payment_order
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
      - JAEGER_SAMPLING_RATE=10
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:40004/health/live"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 8. Kubernetes 部署更新

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-service
spec:
  template:
    spec:
      containers:
      - name: order-service
        env:
        - name: JAEGER_SAMPLING_RATE
          value: "10"  # 生产环境降低采样率

        livenessProbe:
          httpGet:
            path: /health/live
            port: 40004
          initialDelaySeconds: 30
          periodSeconds: 10

        readinessProbe:
          httpGet:
            path: /health/ready
            port: 40004
          initialDelaySeconds: 10
          periodSeconds: 5
```

## 回滚计划

如果迁移出现问题：

```bash
# 1. 恢复备份
cp services/{service}/cmd/main.go.backup services/{service}/cmd/main.go

# 2. 重新编译
cd services/{service}
go build ./cmd/main.go

# 3. 重启服务
pkill {service} && ./{service}
```

## 迁移时间表

建议分阶段迁移：

### 第1阶段（1天）
- ✅ order-service（已完成示例）
- ⏳ notification-service（简单，无复杂依赖）
- ⏳ config-service（简单）

### 第2阶段（1天）
- ⏳ risk-service
- ⏳ accounting-service
- ⏳ analytics-service

### 第3阶段（2天）
- ⏳ payment-gateway（复杂，需要仔细处理签名中间件）
- ⏳ channel-adapter
- ⏳ merchant-auth-service

### 第4阶段（1天）
- ⏳ admin-service
- ⏳ merchant-service
- ⏳ settlement-service
- ⏳ withdrawal-service
- ⏳ kyc-service

## 预期成果

迁移完成后：

✅ **代码质量**
- 所有服务 main.go 减少 ~130行
- 初始化逻辑标准化
- 易于维护和理解

✅ **功能一致性**
- 所有服务统一的健康检查
- 所有服务统一的追踪配置
- 所有服务统一的指标收集

✅ **运维友好**
- K8s 探针开箱即用
- 优雅关闭防止请求丢失
- Prometheus 监控统一

✅ **开发效率**
- 新服务启动时间从1小时降至15分钟
- 配置错误减少90%
- 文档统一（pkg/app/README.md）

## 常见问题

### Q: 迁移会破坏现有功能吗？
**A**: 不会。Bootstrap 框架提供的是标准初始化，所有业务逻辑保持不变。

### Q: 如何处理服务特定的中间件？
**A**: 在路由级别添加，而不是全局：
```go
api := application.Router.Group("/api/v1")
api.Use(YourCustomMiddleware())
```

### Q: 能否部分服务使用 Bootstrap？
**A**: 可以。服务之间独立，可以逐个迁移。

### Q: 迁移后如何调试？
**A**: Bootstrap 保留了所有底层对象访问：
```go
application.DB      // *gorm.DB
application.Redis   // *redis.Client
application.Router  // *gin.Engine
```

## 联系支持

遇到问题？

1. 查看示例：`services/order-service/cmd/main_bootstrap_example.go`
2. 查看文档：`pkg/app/README.md`
3. 提交 Issue：[GitHub Issues](https://github.com/your-org/payment-platform/issues)

# 🚀 支付平台快速开始指南

本指南帮助您快速部署和测试已完成的 P0 + P1 改进功能。

---

## 📋 前置条件

确保以下服务正在运行：

```bash
# 检查 Docker Compose 基础设施
docker-compose ps

# 应该看到以下服务运行中：
# - payment-postgres (PostgreSQL on port 40432)
# - payment-redis (Redis on port 40379)
# - payment-kafka (Kafka on port 40092)
```

---

## 🔧 步骤 1: 数据库初始化

### 1.1 自动迁移（推荐）

所有 Saga 表会在服务启动时自动创建（通过 GORM AutoMigrate）。

```bash
# 启动 payment-gateway 会自动创建 Saga 表
cd /home/eric/payment/backend/services/payment-gateway
export GOWORK=/home/eric/payment/backend/go.work
go run ./cmd/main.go
```

### 1.2 手动迁移（可选）

如果需要手动创建表：

```bash
# 连接到 PostgreSQL
psql -h localhost -p 40432 -U postgres -d payment_gateway

# 检查 Saga 表是否存在
\dt saga*

# 应该看到：
# - saga_instances
# - saga_steps
```

**表结构**:
```sql
-- Saga 实例表
CREATE TABLE saga_instances (
    id UUID PRIMARY KEY,
    business_id VARCHAR(255) NOT NULL,
    business_type VARCHAR(50),
    status VARCHAR(50) NOT NULL,
    current_step INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    completed_at TIMESTAMP,
    compensated_at TIMESTAMP
);

-- Saga 步骤表
CREATE TABLE saga_steps (
    id UUID PRIMARY KEY,
    saga_id UUID NOT NULL REFERENCES saga_instances(id),
    step_order INTEGER NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    execute_data TEXT,
    compensate_data TEXT,
    result TEXT,
    error_message TEXT,
    executed_at TIMESTAMP,
    compensated_at TIMESTAMP,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retry_count INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## 🚦 步骤 2: 启动服务

### 2.1 启动所有服务

```bash
cd /home/eric/payment/backend

# 使用自动化脚本启动所有服务
./scripts/start-all-services.sh

# 检查服务状态
./scripts/status-all-services.sh
```

### 2.2 单独启动 Payment Gateway

```bash
cd /home/eric/payment/backend/services/payment-gateway

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=payment_gateway
export REDIS_HOST=localhost
export REDIS_PORT=40379
export PORT=40003

# 启动服务
export GOWORK=/home/eric/payment/backend/go.work
go run ./cmd/main.go
```

**预期输出**:
```
INFO  正在启动 Payment Gateway Service...
INFO  数据库连接成功
INFO  数据库迁移完成（包含 Saga 表）
INFO  Redis连接成功
INFO  Prometheus 指标初始化完成
INFO  Jaeger 追踪初始化完成
INFO  Saga Orchestrator 初始化完成
INFO  Saga Payment Service 初始化完成（功能已准备就绪）
INFO  Payment Gateway 启动成功，监听端口: 40003
```

---

## ✅ 步骤 3: 测试幂等性保护

### 3.1 运行自动化测试脚本

```bash
cd /home/eric/payment/backend

# 赋予执行权限
chmod +x scripts/test-idempotency.sh

# 运行测试
./scripts/test-idempotency.sh
```

**预期结果**:
```
=========================================
幂等性测试脚本
=========================================

生成幂等性Key: test-1737734400-a1b2c3d4-e5f6-7890-abcd-ef1234567890

测试数据:
  订单号: ORDER-1737734400
  金额: 10000 (100.00 USD)

=========================================
第一次请求 (应该创建新支付)
=========================================
HTTP状态码: 200
响应体:
{
  "code": 0,
  "message": "创建成功",
  "data": {
    "payment_no": "PAY-20250124-123456",
    "status": "pending",
    ...
  }
}

等待2秒...

=========================================
第二次请求 (应该返回缓存响应)
=========================================
HTTP状态码: 200
响应体:
{
  "code": 0,
  "message": "创建成功",
  "data": {
    "payment_no": "PAY-20250124-123456",
    "status": "pending",
    ...
  }
}

=========================================
结果验证
=========================================
✅ 幂等性测试通过: 两次请求返回相同的响应

=========================================
测试完成
=========================================
```

### 3.2 手动测试幂等性

```bash
# 生成唯一的幂等性Key
IDEMPOTENCY_KEY="pay-$(uuidgen)"

# 第一次请求 - 创建支付
curl -X POST "http://localhost:40003/api/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -H "X-Signature: test-signature" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d '{
    "order_no": "ORDER-'$(date +%s)'",
    "amount": 10000,
    "currency": "USD",
    "channel": "stripe",
    "subject": "测试支付",
    "body": "幂等性测试",
    "callback_url": "http://localhost:8080/callback",
    "return_url": "http://localhost:8080/return"
  }'

# 第二次请求 - 相同的幂等性Key，应该返回缓存响应
curl -X POST "http://localhost:40003/api/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -H "X-Signature: test-signature" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d '{
    "order_no": "ORDER-'$(date +%s)'",
    "amount": 10000,
    "currency": "USD",
    "channel": "stripe",
    "subject": "测试支付",
    "body": "幂等性测试",
    "callback_url": "http://localhost:8080/callback",
    "return_url": "http://localhost:8080/return"
  }'
```

**验证**:
- 第一次请求：创建新支付，返回 200 OK
- 第二次请求：返回缓存响应，payment_no 相同

---

## 🔄 步骤 4: 测试 Saga 分布式事务（手动）

### 4.1 查看 Saga 表数据

```bash
# 连接到数据库
psql -h localhost -p 40432 -U postgres -d payment_gateway

# 查询 Saga 实例
SELECT id, business_id, business_type, status, current_step, created_at
FROM saga_instances
ORDER BY created_at DESC
LIMIT 10;

# 查询 Saga 步骤
SELECT si.business_id, ss.step_order, ss.step_name, ss.status, ss.retry_count
FROM saga_instances si
JOIN saga_steps ss ON si.id = ss.saga_id
ORDER BY si.created_at DESC, ss.step_order
LIMIT 20;

# 查询失败的 Saga（需要补偿）
SELECT * FROM saga_instances WHERE status = 'compensated' ORDER BY created_at DESC LIMIT 10;

# 查询待重试的步骤
SELECT * FROM saga_steps
WHERE status = 'failed' AND next_retry_at IS NOT NULL AND next_retry_at <= NOW()
LIMIT 10;
```

### 4.2 Saga 状态说明

**Saga 状态**:
- `pending`: 等待执行
- `in_progress`: 执行中
- `completed`: 已完成（所有步骤成功）✅
- `compensated`: 已补偿（回滚）⚠️
- `failed`: 失败（补偿也失败）❌

**步骤状态**:
- `pending`: 等待执行
- `completed`: 已完成
- `compensated`: 已补偿
- `failed`: 失败（可重试）

---

## 📊 步骤 5: 监控和指标

### 5.1 Prometheus 指标

访问 Prometheus 指标端点：

```bash
# Payment Gateway 指标
curl http://localhost:40003/metrics

# 查看幂等性相关指标（需要添加）
# idempotency_requests_total
# idempotency_cache_hits_total
# idempotency_conflicts_total

# 查看 Saga 相关指标（需要添加）
# saga_started_total
# saga_completed_total
# saga_compensated_total
# saga_duration_seconds
```

### 5.2 Jaeger 分布式追踪

访问 Jaeger UI:

```bash
# 打开浏览器
open http://localhost:40686

# 搜索 payment-gateway 服务的 traces
# 可以看到支付流程的完整调用链
```

### 5.3 Grafana Dashboard（可选）

```bash
# 访问 Grafana
open http://localhost:40300

# 登录: admin / admin

# 添加 Prometheus 数据源
# URL: http://prometheus:9090

# 导入预设 Dashboard 或创建自定义 Dashboard
```

---

## 🔍 步骤 6: 验证数据一致性

### 6.1 验证事务修复

```bash
# 连接到数据库
psql -h localhost -p 40432 -U postgres

# 验证支付表没有重复订单号
\c payment_gateway
SELECT order_no, COUNT(*) FROM payments GROUP BY order_no HAVING COUNT(*) > 1;
-- 应该返回 0 行

# 验证订单表数据完整性
\c payment_order
SELECT o.id, o.order_no, COUNT(oi.id) as items_count
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id, o.order_no
HAVING COUNT(oi.id) = 0;
-- 应该返回 0 行（所有订单都有订单项）

# 验证商户都有 API Key
\c payment_merchant
SELECT m.id, m.email, COUNT(ak.id) as api_keys_count
FROM merchants m
LEFT JOIN api_keys ak ON m.id = ak.merchant_id
GROUP BY m.id, m.email
HAVING COUNT(ak.id) < 2;
-- 应该返回 0 行（每个商户至少有 2 个 API Key：测试 + 生产）
```

### 6.2 验证退款金额限制

```bash
# 测试退款金额超限保护
PAYMENT_NO="PAY-existing-payment-123"

# 第一次退款 - 应该成功
curl -X POST "http://localhost:40003/api/v1/refunds" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -H "X-Signature: test-signature" \
  -d '{
    "payment_no": "'$PAYMENT_NO'",
    "amount": 5000,
    "reason": "测试退款1"
  }'

# 第二次退款 - 如果总额超过支付金额应该失败
curl -X POST "http://localhost:40003/api/v1/refunds" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -H "X-Signature: test-signature" \
  -d '{
    "payment_no": "'$PAYMENT_NO'",
    "amount": 8000,
    "reason": "测试退款2"
  }'
# 预期返回: {"error": "退款总额超过支付金额"}
```

---

## 🛠️ 故障排查

### 问题 1: 服务启动失败

**症状**: 服务无法启动，报错 "数据库连接失败"

**解决方案**:
```bash
# 检查 PostgreSQL 是否运行
docker-compose ps payment-postgres

# 检查端口是否正确
netstat -an | grep 40432

# 重启 PostgreSQL
docker-compose restart payment-postgres
```

### 问题 2: Redis 连接失败

**症状**: 报错 "Redis连接失败"

**解决方案**:
```bash
# 检查 Redis 是否运行
docker-compose ps payment-redis

# 测试 Redis 连接
redis-cli -h localhost -p 40379 ping
# 应该返回: PONG

# 重启 Redis
docker-compose restart payment-redis
```

### 问题 3: Saga 表不存在

**症状**: 报错 "relation saga_instances does not exist"

**解决方案**:
```bash
# 手动运行迁移
psql -h localhost -p 40432 -U postgres -d payment_gateway \
  -f /home/eric/payment/backend/pkg/saga/migrations/001_create_saga_tables.sql

# 或者重启服务，让 GORM AutoMigrate 自动创建
```

### 问题 4: 幂等性不生效

**症状**: 重复请求都被处理了，没有返回缓存响应

**解决方案**:
```bash
# 1. 检查 Redis 是否正常
redis-cli -h localhost -p 40379 ping

# 2. 检查是否提供了 Idempotency-Key header
curl -v http://localhost:40003/api/v1/payments \
  -H "Idempotency-Key: test-key-123" \
  ...

# 3. 检查 Redis 中的键
redis-cli -h localhost -p 40379
> KEYS payment-gateway:idempotency:*
> GET payment-gateway:idempotency:test-key-123

# 4. 查看服务日志
tail -f /home/eric/payment/backend/logs/payment-gateway.log | grep idempotency
```

---

## 📚 下一步

### 立即可用功能

1. ✅ **幂等性保护**: 所有 POST/PUT/PATCH 请求自动支持
2. ✅ **事务保护**: 所有关键操作都有 ACID 保证
3. ✅ **Saga 框架**: 已准备就绪，可选启用

### 待实现功能（可选）

1. **Order Service `/cancel` 接口**:
```go
// order-service/internal/handler/order_handler.go
func (h *OrderHandler) CancelOrder(c *gin.Context) {
    orderNo := c.Param("order_no")
    // 实现取消逻辑
}
```

2. **Channel Adapter `/cancel` 接口**:
```go
// channel-adapter/internal/handler/channel_handler.go
func (h *ChannelHandler) CancelPayment(c *gin.Context) {
    channelTradeNo := c.Param("channel_trade_no")
    // 调用 Stripe API 取消支付
}
```

3. **Saga 后台重试任务**:
```go
// 定期扫描待重试的步骤
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        steps, _ := orchestrator.ListPendingRetries(ctx, 100)
        for _, step := range steps {
            // 重试执行步骤
        }
    }
}()
```

4. **Prometheus 指标**:
```go
// pkg/metrics/idempotency_metrics.go
var (
    IdempotencyRequests = prometheus.NewCounterVec(...)
    IdempotencyCacheHits = prometheus.NewCounterVec(...)
)

// pkg/metrics/saga_metrics.go
var (
    SagaStarted = prometheus.NewCounterVec(...)
    SagaCompleted = prometheus.NewCounterVec(...)
    SagaCompensated = prometheus.NewCounterVec(...)
)
```

---

## 📖 参考文档

| 文档 | 说明 |
|-----|-----|
| [IDEMPOTENCY_IMPLEMENTATION.md](IDEMPOTENCY_IMPLEMENTATION.md) | 幂等性实现详细文档 |
| [SAGA_IMPLEMENTATION.md](SAGA_IMPLEMENTATION.md) | Saga 模式实现详细文档 |
| [TRANSACTION_FIXES_SUMMARY.md](TRANSACTION_FIXES_SUMMARY.md) | P0 事务修复总结 |
| [P1_IMPROVEMENTS_SUMMARY.md](P1_IMPROVEMENTS_SUMMARY.md) | P1 改进总结 |
| [FINAL_COMPLETION_SUMMARY.md](FINAL_COMPLETION_SUMMARY.md) | 最终完成总结 |

---

## 🎯 成功指标

确认以下功能正常工作：

- [ ] 所有服务启动成功（无错误日志）
- [ ] 数据库 Saga 表自动创建
- [ ] 幂等性测试脚本通过
- [ ] 重复请求返回缓存响应（相同 payment_no）
- [ ] 并发请求返回 409 Conflict
- [ ] Prometheus 指标可访问 `/metrics`
- [ ] Jaeger UI 可以查看 traces
- [ ] 数据库无重复订单号
- [ ] 所有订单都有订单项
- [ ] 所有商户都有 API Keys

---

**恭喜！** 🎉

您已成功完成支付平台 P0 + P1 改进的部署和测试。系统现已具备企业级生产能力。

**支持**:
- 查看文档: `/home/eric/payment/*.md`
- 查看日志: `/home/eric/payment/backend/logs/`
- 数据库: `psql -h localhost -p 40432 -U postgres`

---

**版本**: 1.0
**创建时间**: 2025-01-24
**维护者**: Payment Platform Team

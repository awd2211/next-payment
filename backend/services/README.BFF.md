# BFF Services - 快速开始指南

## 📋 概览

BFF (Backend for Frontend) 服务为 Admin Portal 和 Merchant Portal 提供统一的 API 网关，集成企业级安全特性。

**两个 BFF 服务**:
- **Admin BFF** (40001): 管理员门户 API 网关 - Zero-Trust 安全架构
- **Merchant BFF** (40023): 商户门户 API 网关 - 租户隔离架构

---

## 🚀 快速开始

### 方法 1: 使用启动脚本（推荐）

```bash
# 1. 进入 backend 目录
cd /home/eric/payment/backend

# 2. 设置必需的环境变量
export JWT_SECRET="payment-platform-secret-key-2024"

# 3. 启动两个 BFF 服务
./scripts/start-bff-services.sh

# 4. 验证服务状态
# Admin BFF:    http://localhost:40001/health
# Merchant BFF: http://localhost:40023/health
```

### 方法 2: 手动启动

```bash
# 编译 Admin BFF
cd services/admin-bff-service
GOWORK=../../go.work go build -o /tmp/admin-bff-service ./cmd/main.go

# 启动 Admin BFF
PORT=40001 \
DB_NAME=payment_admin \
JWT_SECRET="your-secret" \
/tmp/admin-bff-service &

# 编译 Merchant BFF
cd ../merchant-bff-service
GOWORK=../../go.work go build -o /tmp/merchant-bff-service ./cmd/main.go

# 启动 Merchant BFF
PORT=40023 \
JWT_SECRET="your-secret" \
/tmp/merchant-bff-service &
```

### 方法 3: Docker Compose

```bash
# 1. 启动基础设施
docker-compose -f docker-compose.yml up -d postgres redis kafka

# 2. 启动 BFF 服务
docker-compose -f docker-compose.bff.yml up -d

# 3. 查看日志
docker-compose -f docker-compose.bff.yml logs -f
```

---

## 🔧 环境变量

### Admin BFF 必需变量

```bash
# 服务配置
export PORT=40001
export JWT_SECRET="your-jwt-secret-key"

# 数据库配置（必需，用于审计日志）
export DB_HOST=localhost
export DB_PORT=40432
export DB_NAME=payment_admin
export DB_USER=postgres
export DB_PASSWORD=postgres

# 后端服务 URLs（18 个）
export CONFIG_SERVICE_URL=http://localhost:40010
export RISK_SERVICE_URL=http://localhost:40006
export KYC_SERVICE_URL=http://localhost:40015
# ... (见 .env.example 完整列表)
```

### Merchant BFF 必需变量

```bash
# 服务配置
export PORT=40023
export JWT_SECRET="your-jwt-secret-key"

# Merchant BFF 不需要数据库

# 后端服务 URLs（15 个）
export PAYMENT_GATEWAY_URL=http://localhost:40003
export ORDER_SERVICE_URL=http://localhost:40004
export SETTLEMENT_SERVICE_URL=http://localhost:40013
# ... (见 .env.example 完整列表)
```

### 可选变量

```bash
# Redis（可选，用于速率限制）
export REDIS_HOST=localhost
export REDIS_PORT=40379

# 可观测性
export JAEGER_ENDPOINT=http://localhost:14268/api/traces
export JAEGER_SAMPLING_RATE=10  # 10% 采样
export LOG_LEVEL=info

# SMTP（仅 Admin BFF）
export SMTP_HOST=smtp.gmail.com
export SMTP_PORT=587
export SMTP_USERNAME=your-email@gmail.com
export SMTP_PASSWORD=your-app-password
```

---

## 📊 服务端点

### Admin BFF (40001)

| 端点 | 说明 |
|------|------|
| http://localhost:40001/swagger/index.html | Swagger UI API 文档 |
| http://localhost:40001/health | 健康检查 |
| http://localhost:40001/health/live | 存活探针 |
| http://localhost:40001/health/ready | 就绪探针 |
| http://localhost:40001/metrics | Prometheus 指标 |

**聚合的微服务** (18):
- config, risk, kyc, merchant, analytics, limit
- channel, cashier, order, accounting, dispute
- merchant-auth, merchant-config, notification
- payment, reconciliation, settlement, withdrawal

### Merchant BFF (40023)

| 端点 | 说明 |
|------|------|
| http://localhost:40023/swagger/index.html | Swagger UI API 文档 |
| http://localhost:40023/health | 健康检查 |
| http://localhost:40023/health/live | 存活探针 |
| http://localhost:40023/health/ready | 就绪探针 |
| http://localhost:40023/metrics | Prometheus 指标 |

**聚合的微服务** (15):
- payment, order, settlement, withdrawal, accounting
- analytics, kyc, merchant-auth, merchant-config
- merchant-limit, notification, risk, dispute
- reconciliation, cashier

---

## 🔒 安全特性

### Admin BFF - 8 层安全栈

```
1. Structured Logging     → JSON 格式日志
2. Rate Limiting           → 60 req/min (normal), 5 req/min (sensitive)
3. JWT Authentication      → Token 验证
4. RBAC Permission Check   → 6 种角色权限控制
5. Require Reason          → 敏感操作需理由
6. 2FA Verification        → 财务操作二次验证
7. Business Logic          → 业务逻辑执行
8. Data Masking + Audit    → 数据脱敏 + 审计日志
```

**RBAC 角色**:
- `super_admin`: 完全访问权限
- `operator`: 商户/订单管理，KYC 审核
- `finance`: 会计、结算、提现
- `risk_manager`: 风控、争议处理
- `support`: 只读访问（客服）
- `auditor`: 审计日志、数据分析

**2FA 保护的操作**:
- 支付操作（查询、退款、取消）
- 结算操作（批准、发放）
- 提现操作（批准、处理）
- 争议操作（创建、处理）

### Merchant BFF - 5 层安全栈

```
1. Structured Logging     → JSON 格式日志
2. Rate Limiting           → 300 req/min (relaxed), 60 req/min (financial)
3. JWT Authentication      → 商户 Token 验证
4. Tenant Isolation        → 强制 merchant_id 注入
5. Data Masking            → PII 数据脱敏
```

**租户隔离**:
```go
// merchant_id 自动从 JWT 提取
// 强制注入到所有后端服务调用
// 商户无法跨租户访问数据
```

**速率限制**:
- Relaxed: 300 req/min（一般操作）- 5x 宽松
- Normal: 60 req/min（财务操作）

---

## 🧪 测试

### 1. 健康检查

```bash
# Admin BFF
curl http://localhost:40001/health

# 预期响应:
# {"status":"healthy","timestamp":"2025-10-26T12:00:00Z"}

# Merchant BFF
curl http://localhost:40023/health
```

### 2. 运行安全测试

```bash
cd /home/eric/payment/backend
./scripts/test-bff-security.sh
```

**测试项**:
- [x] 服务可用性
- [x] JWT 认证（缺少 Token）
- [x] JWT 认证（无效 Token）
- [x] 速率限制
- [x] 数据脱敏（需手动验证）
- [x] Prometheus 指标
- [x] 健康检查端点

### 3. 手动测试 RBAC

```bash
# 1. 管理员登录
curl -X POST http://localhost:40001/api/v1/admins/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"SecurePass123!"}'

# 2. 获取 Token
export ADMIN_TOKEN="eyJhbGc..."

# 3. 尝试访问需要特定权限的端点
curl -X POST http://localhost:40001/api/v1/admin/settlements/123/approve \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason":"All compliance checks passed"}'
```

### 4. 手动测试 2FA

```bash
# 1. 启用 2FA
curl -X POST http://localhost:40001/api/v1/admins/2fa/enable \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# 2. 获取 TOTP Secret
# 使用 Google Authenticator 扫描二维码

# 3. 访问敏感操作（需要 2FA）
curl -X GET http://localhost:40001/api/v1/admin/payments \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-2FA-Code: 123456"
```

### 5. 手动测试租户隔离

```bash
# 1. 商户 A 登录
export MERCHANT_A_TOKEN=$(curl -X POST http://localhost:40023/api/v1/merchant/login \
  -d '{"email":"merchantA@example.com","password":"pass"}' | jq -r '.data.token')

# 2. 查询订单（只能看到自己的）
curl -X GET http://localhost:40023/api/v1/merchant/orders \
  -H "Authorization: Bearer $MERCHANT_A_TOKEN"

# 3. 尝试传递其他商户 ID（会被忽略）
curl -X GET "http://localhost:40023/api/v1/merchant/orders?merchant_id=other-merchant" \
  -H "Authorization: Bearer $MERCHANT_A_TOKEN"
# 预期: 依然只返回商户 A 的订单
```

---

## 📈 监控

### Prometheus 指标

```bash
# 查看所有指标
curl http://localhost:40001/metrics
curl http://localhost:40023/metrics

# 关键指标:
# - http_requests_total              : 总请求数
# - http_request_duration_seconds    : 请求延迟
# - http_request_size_bytes          : 请求大小
# - http_response_size_bytes         : 响应大小
# - process_resident_memory_bytes    : 内存使用
# - process_cpu_seconds_total        : CPU 使用
# - go_goroutines                    : Goroutine 数量
```

### Grafana Dashboard

```bash
# 访问 Grafana
open http://localhost:40300

# 导入 Dashboard
# 文件: /home/eric/payment/monitoring/grafana/dashboards/bff-services-dashboard.json
```

### 日志查看

```bash
# 实时日志
tail -f logs/bff/admin-bff.log
tail -f logs/bff/merchant-bff.log

# 查找错误
grep "ERROR" logs/bff/*.log

# 查找限流事件
grep "429" logs/bff/*.log

# 查找 2FA 失败
grep "2FA" logs/bff/admin-bff.log
```

---

## 🛠️ 故障排查

### 问题 1: Admin BFF 启动失败

**错误**: `database connection failed`

**原因**: PostgreSQL 未运行或 payment_admin 数据库不存在

**解决**:
```bash
# 启动 PostgreSQL
docker-compose up -d postgres

# 创建数据库
docker exec -it payment-postgres psql -U postgres -c "CREATE DATABASE payment_admin;"
```

### 问题 2: 速率限制不生效

**原因**: Redis 未运行，降级为内存存储

**解决**:
```bash
# 启动 Redis
docker-compose up -d redis

# 验证连接
redis-cli -h localhost -p 40379 ping
```

### 问题 3: 2FA 总是失败

**原因**: 服务器时间不同步或 Secret 错误

**解决**:
```bash
# 检查服务器时间
date

# 同步时间（Linux）
sudo ntpdate pool.ntp.org

# 重新生成 2FA Secret
curl -X POST http://localhost:40001/api/v1/admins/2fa/regenerate \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### 问题 4: Swagger UI 空白

**原因**: Swagger 文档未生成

**解决**:
```bash
cd services/admin-bff-service
swag init -g cmd/main.go
go build ./cmd/main.go
```

---

## 📊 监控和告警

### 快速启动监控

```bash
# 一键启动 Prometheus + Grafana 监控
./scripts/start-bff-monitoring.sh

# 访问监控界面
# Prometheus: http://localhost:40090
# Grafana:    http://localhost:40300 (admin/admin)
```

### 监控指标

**Prometheus 采集**:
- Admin BFF:    http://localhost:40001/metrics (10s 间隔)
- Merchant BFF: http://localhost:40023/metrics (15s 间隔)

**关键指标**:
```promql
# 服务可用性
up{job=~"admin-bff|merchant-bff"}

# 请求速率
job:http_requests:rate5m

# P95 延迟
job:http_request_duration:p95

# 错误率
job:http_errors:rate5m

# 安全事件
job:security_events:rate5m
```

### 告警规则 (21 Total)

**Critical** (6):
- BFFServiceDown - 服务宕机 (1 min)
- BFFHighErrorRate - 错误率 >5% (5 min)
- BFFExtremelyHighLatency - P95 >3s (5 min)
- BFFMemoryExhaustion - 内存 >90% (5 min)
- BFFHighRateLimitViolations - 限流滥用 >10/s (5 min)
- BFFCriticalSecurityEvents - 安全事件 >50/min (5 min)

**Warning** (11):
- BFFHighLatency, BFFHighMemoryUsage, BFFHighCPUUsage
- BFFMediumRateLimitViolations, BFFAuthFailures, BFFPermissionDenied
- BFF2FAFailures, BFFHighGoroutines, BFFSlowRequests
- BFFHighRequestSize, BFFHighResponseSize

**Info** (4):
- BFFServiceRestarted, BFFLowTraffic
- BFFUnusualErrorPattern, BFFFileDescriptorWarning

### Grafana Dashboard

**15 监控面板**:
1. Service Status (服务状态)
2. Request Rate (请求速率)
3. Error Rate (错误率)
4. P95/P99 Latency (延迟)
5. Rate Limit Violations (限流违规)
6. Authentication Failures (认证失败)
7. HTTP Status Distribution (状态码分布)
8. Memory Usage (内存使用)
9. CPU Usage (CPU 使用)
10. Active Goroutines (协程数)
11. Request by Endpoint (Top 10 端点)
12. 2FA Failures (2FA 失败 - Admin BFF)
13. Tenant Metrics (租户指标 - Merchant BFF)
14. Request Size (请求大小)
15. Response Size (响应大小)

**导入 Dashboard**:
1. 访问 http://localhost:40300
2. 登录: admin / admin
3. 导航: Dashboards → Import
4. 上传: `monitoring/grafana/dashboards/bff-services-dashboard.json`

---

## 📚 相关文档

**核心文档**:
- **[BFF_SECURITY_COMPLETE_SUMMARY.md](../../BFF_SECURITY_COMPLETE_SUMMARY.md)** - 架构总览
- **[ADVANCED_SECURITY_COMPLETE.md](admin-bff-service/ADVANCED_SECURITY_COMPLETE.md)** - Admin BFF 详细文档
- **[MERCHANT_BFF_SECURITY.md](merchant-bff-service/MERCHANT_BFF_SECURITY.md)** - Merchant BFF 详细文档
- **[BFF_IMPLEMENTATION_COMPLETE.md](../../BFF_IMPLEMENTATION_COMPLETE.md)** - 实施报告

**监控文档**:
- **[BFF_MONITORING_COMPLETE.md](../../BFF_MONITORING_COMPLETE.md)** - 监控实施完整报告
- **[Prometheus README](../../monitoring/prometheus/README.md)** - Prometheus 配置指南
- **[Grafana README](../../monitoring/grafana/README.md)** - Grafana Dashboard 指南
- **[Prometheus Alerts](../../backend/deployments/prometheus/alerts/bff-alerts.yml)** - 21 条告警规则
- **[Recording Rules](../../backend/deployments/prometheus/rules/bff-recording-rules.yml)** - 25 条预计算规则
- **[Grafana Dashboard](../../monitoring/grafana/dashboards/bff-services-dashboard.json)** - 15 面板监控

---

## 🚦 停止服务

```bash
# 使用脚本停止
./scripts/stop-bff-services.sh

# 或手动停止
pkill -f admin-bff-service
pkill -f merchant-bff-service

# Docker 方式
docker-compose -f docker-compose.bff.yml down
```

---

## 💡 最佳实践

### 生产环境

1. **使用强 JWT Secret**:
```bash
export JWT_SECRET=$(openssl rand -base64 32)
```

2. **启用 HTTPS**:
- 使用 Nginx/Traefik 作为反向代理
- 配置 SSL/TLS 证书

3. **配置数据库连接池**:
```bash
export DB_MAX_OPEN_CONNS=100
export DB_MAX_IDLE_CONNS=10
```

4. **调整日志采样率**:
```bash
export JAEGER_SAMPLING_RATE=10  # 10% 采样（生产环境推荐）
```

5. **设置资源限制**:
```yaml
# docker-compose.bff.yml
deploy:
  resources:
    limits:
      cpus: '1.0'
      memory: 512M
```

---

**最后更新**: 2025-10-26
**维护团队**: Payment Platform Team
**支持**: https://github.com/your-org/payment-platform/issues

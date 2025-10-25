# 基础设施端口配置

本文档说明了整个支付平台使用的所有基础设施端口配置。

## Docker Compose 服务端口映射

| 服务 | 外部端口 | 内部端口 | 说明 |
|------|---------|---------|------|
| **数据存储** |
| PostgreSQL | 40432 | 5432 | 主数据库 |
| Redis | 40379 | 6379 | 缓存和会话存储 |
| **消息队列** |
| Kafka | 40092 | 9092 | Kafka Broker (内部通信) |
| Kafka External | 40093 | 9093 | Kafka Broker (外部访问) |
| **API 网关** |
| Kong Proxy | 40080 | 8000 | API 网关入口 |
| Kong Admin | 40081 | 8001 | Kong 管理 API |
| **监控和追踪** |
| Prometheus | 40090 | 9090 | 指标收集 |
| Grafana | 40300 | 3000 | 可视化仪表板 |
| Jaeger UI | 40686 | 16686 | 分布式追踪 UI |
| **日志管理** |
| Elasticsearch | 40920 | 9200 | 日志存储 |
| Kibana | 40561 | 5601 | 日志查询 UI |
| Logstash | 40514 | 5014 | 日志收集 |

## 微服务端口分配

| 服务名 | 端口 | 数据库 |
|--------|------|--------|
| config-service | 40010 | payment_config |
| admin-service | 40001 | payment_admin |
| merchant-auth-service | 40011 | payment_merchant_auth |
| merchant-service | 40002 | payment_merchant |
| merchant-config-service | 40012 | payment_merchant_config |
| merchant-limit-service | 40022 | payment_merchant_limit |
| risk-service | 40006 | payment_risk |
| channel-adapter | 40005 | payment_channel |
| order-service | 40004 | payment_order |
| payment-gateway | 40003 | payment_gateway |
| accounting-service | 40007 | payment_accounting |
| analytics-service | 40009 | payment_analytics |
| notification-service | 40008 | payment_notification |
| settlement-service | 40013 | payment_settlement |
| withdrawal-service | 40014 | payment_withdrawal |
| kyc-service | 40015 | payment_kyc |
| cashier-service | 40016 | payment_cashier |
| reconciliation-service | 40020 | payment_reconciliation |
| dispute-service | 40021 | payment_dispute |

## 前端应用端口

| 应用 | 端口 | 说明 |
|------|------|------|
| admin-portal | 5173 | 管理员后台 |
| merchant-portal | 5174 | 商户门户 |
| website | 5175 | 官方网站 |

## 代码中的默认配置

### Bootstrap 框架 (pkg/app/bootstrap.go)

```go
// 数据库配置
Port: config.GetEnvInt("DB_PORT", 40432)  // PostgreSQL

// Redis 配置
Port: config.GetEnvInt("REDIS_PORT", 40379)  // Redis
```

### 各服务中的配置

```go
// Kafka Broker (使用 Kafka 的服务)
kafkaBrokersStr := config.GetEnv("KAFKA_BROKERS", "localhost:40092")

// JWT 密钥 (所有需要认证的服务)
jwtSecret := config.GetEnv("JWT_SECRET", "payment-platform-secret-key-2024")
```

### 启动脚本 (scripts/start-all-mtls.sh)

```bash
# 统一的 JWT 密钥（所有微服务共享，支持跨服务认证）
export JWT_SECRET="payment-platform-secret-key-2024"
```

## 配置优先级

1. **环境变量** - 最高优先级，可以覆盖代码中的默认值
2. **代码默认值** - 次优先级，确保无需配置即可运行
3. **启动脚本导出** - 为批量启动提供统一配置

## 跨服务认证

所有微服务使用统一的 JWT 密钥 `payment-platform-secret-key-2024`，这样：

- ✅ 管理员 (admin-service) 签发的 token 可以访问商户服务 (merchant-service)
- ✅ 商户 (merchant-service) 签发的 token 可以访问支付网关 (payment-gateway)
- ✅ 所有服务之间可以互相验证 JWT token

## 验证配置

运行以下命令验证所有配置：

```bash
cd /home/eric/payment/backend

# 检查 Bootstrap 配置
grep -E "DB_PORT.*40432|REDIS_PORT.*40379" pkg/app/bootstrap.go

# 检查 JWT 密钥
find services -name "main.go" -path "*/cmd/*" | xargs grep "JWT_SECRET"

# 检查 Kafka 配置
find services -name "main.go" -path "*/cmd/*" | xargs grep "KAFKA_BROKERS"
```

## 更新日期

2025-10-25 - 初始版本，统一所有基础设施端口配置

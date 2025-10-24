# 环境变量配置文档

本文档列出了所有微服务支持的环境变量及其默认值。

## 通用配置

所有服务都支持以下环境变量：

### 基础配置
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `ENV` | `development` | 运行环境：development, production |
| `PORT` | 各服务不同 | HTTP服务端口 |
| `GRPC_PORT` | 各服务不同 | gRPC服务端口（如果启用） |

### 数据库配置 (PostgreSQL)
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DB_HOST` | `localhost` | 数据库主机地址 |
| `DB_PORT` | `5432` | 数据库端口 |
| `DB_USER` | `postgres` | 数据库用户名 |
| `DB_PASSWORD` | `postgres` | 数据库密码 |
| `DB_NAME` | 各服务不同 | 数据库名称 |
| `DB_SSL_MODE` | `disable` | SSL模式：disable, require, verify-full |
| `DB_TIMEZONE` | `UTC` | 数据库时区 |

**连接池配置**（hardcoded in pkg/db/postgres.go）:
- MaxIdleConns: 10
- MaxOpenConns: 100
- ConnMaxLifetime: 1 hour

### Redis配置
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `REDIS_HOST` | `localhost` | Redis主机地址 |
| `REDIS_PORT` | `6379` | Redis端口 |
| `REDIS_PASSWORD` | `""` | Redis密码（空表示无密码） |
| `REDIS_DB` | `0` | Redis数据库索引 |
| `REDIS_POOL_SIZE` | `10` | 连接池大小（高并发建议50-100） |
| `REDIS_MIN_IDLE_CONNS` | `5` | 最小空闲连接数 |

**超时配置**（可通过 RedisConfig 结构体配置）:
- DialTimeout: 5 seconds
- ReadTimeout: 3 seconds
- WriteTimeout: 3 seconds

### 可观测性配置

#### Jaeger Tracing
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `JAEGER_ENDPOINT` | `http://localhost:14268/api/traces` | Jaeger Collector端点 |
| `JAEGER_SAMPLING_RATE` | `100` | 采样率 0-100（生产环境建议10-20） |

#### Prometheus Metrics
- Metrics端点: `/metrics` (自动启用)
- 端口: 与HTTP服务相同

### Kafka配置（可选）
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `KAFKA_BROKERS` | `""` | Kafka brokers列表，逗号分隔 |

如果未配置，服务将使用降级模式（日志输出）。

---

## 服务特定配置

### Payment Gateway (端口 40003)

**数据库**:
- `DB_NAME`: `payment_gateway`

**服务间调用**:
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `ORDER_SERVICE_URL` | `http://localhost:8004` | Order Service地址 |
| `CHANNEL_SERVICE_URL` | `http://localhost:8005` | Channel Adapter地址 |
| `RISK_SERVICE_URL` | `http://localhost:8006` | Risk Service地址 |

**Webhook配置**:
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `WEBHOOK_BASE_URL` | `http://payment-gateway:40003` | Webhook回调基础URL |

**说明**: 生产环境应配置为公网可访问的域名，例如 `https://api.yourcompany.com`

**限流配置**（hardcoded）:
- 100 requests/minute per IP/User

**幂等性配置**（hardcoded）:
- TTL: 24 hours

---

### Order Service (端口 40004)

**数据库**:
- `DB_NAME`: `payment_order`

**限流配置**:
- 100 requests/minute

---

### Channel Adapter (端口 40005)

**数据库**:
- `DB_NAME`: `payment_channel`

**Stripe配置**:
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `STRIPE_API_KEY` | `""` | Stripe Secret Key |
| `STRIPE_WEBHOOK_SECRET` | `""` | Stripe Webhook签名密钥 |

---

### Admin Service (端口 40001)

**数据库**:
- `DB_NAME`: `payment_admin`

**JWT配置**:
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `JWT_SECRET` | `""` | JWT签名密钥（生产必须配置） |
| `JWT_EXPIRY_HOURS` | `24` | JWT过期时间（小时） |

---

### Merchant Service (端口 40002)

**数据库**:
- `DB_NAME`: `payment_merchant`

---

### Risk Service (端口 40006)

**数据库**:
- `DB_NAME`: `payment_risk`

---

### Accounting Service (端口 40007)

**数据库**:
- `DB_NAME`: `payment_accounting`

---

### Notification Service (端口 40008)

**数据库**:
- `DB_NAME`: `payment_notify`

**邮件配置**:
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `SMTP_HOST` | `""` | SMTP服务器地址 |
| `SMTP_PORT` | `587` | SMTP端口 |
| `SMTP_USER` | `""` | SMTP用户名 |
| `SMTP_PASSWORD` | `""` | SMTP密码 |
| `EMAIL_FROM` | `noreply@payment.com` | 发件人地址 |

---

### Analytics Service (端口 40009)

**数据库**:
- `DB_NAME`: `payment_analytics`

---

### Config Service (端口 40010)

**数据库**:
- `DB_NAME`: `payment_config`

---

## 生产环境建议配置

### 高并发场景
```bash
# Redis连接池
REDIS_POOL_SIZE=100
REDIS_MIN_IDLE_CONNS=20

# Jaeger采样率（降低性能开销）
JAEGER_SAMPLING_RATE=10

# 数据库（需要修改pkg/db/postgres.go）
# SetMaxOpenConns(200)
# SetMaxIdleConns(20)
```

### 安全配置
```bash
# 生产环境必须配置
JWT_SECRET=<strong-random-secret>
DB_PASSWORD=<strong-password>
REDIS_PASSWORD=<strong-password>
STRIPE_API_KEY=sk_live_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# SSL配置
DB_SSL_MODE=verify-full

# Webhook公网地址
WEBHOOK_BASE_URL=https://api.yourcompany.com
```

### Docker Compose示例
```yaml
services:
  payment-gateway:
    environment:
      - ENV=production
      - PORT=40003
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=payment_gateway
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_POOL_SIZE=50
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
      - JAEGER_SAMPLING_RATE=20
      - WEBHOOK_BASE_URL=https://api.yourcompany.com
      - ORDER_SERVICE_URL=http://order-service:40004
      - CHANNEL_SERVICE_URL=http://channel-adapter:40005
      - RISK_SERVICE_URL=http://risk-service:40006
```

---

## 配置优先级

1. 环境变量（最高优先级）
2. 代码默认值

**注意**: 某些配置（如数据库连接池、限流参数）目前硬编码在代码中，未来版本将支持环境变量配置。

---

## 配置验证

启动服务时，建议检查日志确认配置已正确加载：

```bash
# 查看服务日志
docker-compose logs payment-gateway | grep -i "config\|连接成功"

# 检查健康状态
curl http://localhost:40003/health
```

---

## 故障排查

### Redis连接失败
- 检查 `REDIS_HOST` 和 `REDIS_PORT`
- 检查网络连通性: `telnet redis-host 6379`
- 检查密码配置: `REDIS_PASSWORD`

### 数据库连接失败
- 检查 `DB_*` 环境变量
- 验证数据库是否已创建
- 检查 SSL 配置

### Webhook回调失败
- 确认 `WEBHOOK_BASE_URL` 可从公网访问
- 检查防火墙和负载均衡配置
- 验证Stripe配置的webhook endpoint

---

**最后更新**: 2025-10-24

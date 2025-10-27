# 🐳 Docker 部署完整指南

## 📋 目录

- [快速开始](#快速开始)
- [架构概览](#架构概览)
- [系统要求](#系统要求)
- [部署步骤](#部署步骤)
- [配置说明](#配置说明)
- [监控与运维](#监控与运维)
- [故障排查](#故障排查)
- [安全最佳实践](#安全最佳实践)

---

## 🚀 快速开始

### 一键启动完整系统

```bash
# 1. 克隆代码仓库
cd /home/eric/payment

# 2. 生成所有 Dockerfile（如果还没有）
cd backend && ./scripts/generate-dockerfiles.sh

# 3. 生成 docker-compose.services.yml（如果还没有）
./scripts/generate-docker-compose-services.sh

# 4. 启动基础设施（PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger）
cd .. && docker-compose up -d

# 5. 等待基础设施就绪（约30秒）
docker-compose ps

# 6. 启动所有微服务（17个）
docker-compose -f docker-compose.services.yml up -d

# 7. 启动 BFF 服务（Admin + Merchant）
docker-compose -f docker-compose.bff.yml up -d

# 8. 查看所有服务状态
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### 验证部署

```bash
# 健康检查
curl http://localhost:40003/health  # Payment Gateway
curl http://localhost:40004/health  # Order Service
curl http://localhost:40001/health  # Admin BFF
curl http://localhost:40023/health  # Merchant BFF

# 查看 Prometheus 监控
open http://localhost:40090

# 查看 Grafana 仪表板
open http://localhost:40300  # admin/admin

# 查看 Jaeger 追踪
open http://localhost:50686
```

---

## 🏗️ 架构概览

### 服务分层

```
┌─────────────────────────────────────────────────────────────┐
│                       外部访问层                              │
│  Kong Gateway (40080) + Admin Portal + Merchant Portal      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                       BFF 聚合层                              │
│  Admin BFF (40001) + Merchant BFF (40023)                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     核心业务服务层 (17个)                      │
│  Payment Gateway, Order, Channel Adapter, Risk, Accounting  │
│  Notification, Analytics, Config, Merchant Auth, Settlement │
│  Withdrawal, KYC, Cashier, Reconciliation, Dispute          │
│  Merchant Policy, Merchant Quota                            │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                       基础设施层                              │
│  PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger     │
└─────────────────────────────────────────────────────────────┘
```

### 网络拓扑

```
Docker Network: payment-network (172.28.0.0/16)

内网域名格式: <service-name>.payment-network

示例:
- payment-gateway.payment-network:40003
- order-service.payment-network:40004
- postgres.payment-network:5432
- redis.payment-network:6379
- kafka.payment-network:9092
```

### 服务端口映射

| 服务类型 | 服务名称 | 内网端口 | 外网端口 | 数据库 |
|---------|---------|---------|---------|--------|
| **BFF** | admin-bff-service | 40001 | 40001 | payment_admin |
| **BFF** | merchant-bff-service | 40023 | 40023 | payment_merchant |
| **核心** | payment-gateway | 40003 | 40003 | payment_gateway |
| **核心** | order-service | 40004 | 40004 | payment_order |
| **核心** | channel-adapter | 40005 | 40005 | payment_channel |
| **核心** | risk-service | 40006 | 40006 | payment_risk |
| **核心** | accounting-service | 40007 | 40007 | payment_accounting |
| **核心** | notification-service | 40008 | 40008 | payment_notification |
| **核心** | analytics-service | 40009 | 40009 | payment_analytics |
| **核心** | config-service | 40010 | 40010 | payment_config |
| **核心** | merchant-auth-service | 40011 | 40011 | payment_merchant_auth |
| **核心** | settlement-service | 40013 | 40013 | payment_settlement |
| **核心** | withdrawal-service | 40014 | 40014 | payment_withdrawal |
| **核心** | kyc-service | 40015 | 40015 | payment_kyc |
| **核心** | cashier-service | 40016 | 40016 | payment_cashier |
| **核心** | reconciliation-service | 40020 | 40020 | payment_reconciliation |
| **核心** | dispute-service | 40021 | 40021 | payment_dispute |
| **核心** | merchant-policy-service | 40022 | 40022 | payment_merchant_policy |
| **核心** | merchant-quota-service | 40024 | 40024 | payment_merchant_quota |

| 基础设施 | 服务名称 | 外网端口 | 用途 |
|---------|---------|---------|------|
| **数据库** | PostgreSQL | 40432 | 主数据库 |
| **缓存** | Redis | 40379 | 分布式缓存 |
| **消息队列** | Kafka | 40092 | 事件流 |
| **监控** | Prometheus | 40090 | 指标收集 |
| **可视化** | Grafana | 40300 | 监控仪表板 |
| **追踪** | Jaeger UI | 50686 | 分布式追踪 |
| **API网关** | Kong Gateway | 40080 | 统一入口 |

---

## 💻 系统要求

### 最低配置（开发环境）

- **CPU**: 4 核
- **内存**: 8 GB
- **磁盘**: 50 GB 可用空间
- **操作系统**: Linux (推荐 Ubuntu 20.04+), macOS, Windows (WSL2)
- **Docker**: 24.0+
- **Docker Compose**: 2.20+

### 推荐配置（生产环境）

- **CPU**: 16 核
- **内存**: 32 GB
- **磁盘**: 500 GB SSD (数据库/日志持久化)
- **网络**: 1 Gbps+
- **操作系统**: Linux (Ubuntu 22.04 LTS / Rocky Linux 9)

### 检查系统资源

```bash
# CPU 核心数
lscpu | grep "^CPU(s):"

# 内存
free -h

# 磁盘空间
df -h

# Docker 版本
docker --version
docker-compose --version
```

---

## 📦 部署步骤

### 步骤 1: 准备环境

```bash
# 安装 Docker（Ubuntu）
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 重新登录以应用 docker 组权限
exit
# 重新登录后验证
docker ps
```

### 步骤 2: 生成 mTLS 证书（必需）

```bash
cd /home/eric/payment/backend/certs

# 生成 CA 证书
./generate-ca-cert.sh

# 为每个服务生成证书（19个服务）
for service in payment-gateway order-service channel-adapter risk-service \
               accounting-service notification-service analytics-service \
               config-service merchant-auth-service settlement-service \
               withdrawal-service kyc-service cashier-service \
               reconciliation-service dispute-service merchant-policy-service \
               merchant-quota-service admin-bff-service merchant-bff-service; do
    ./generate-service-cert.sh $service
done

# 验证证书
ls -lh services/*/
```

### 步骤 3: 配置环境变量

```bash
# 创建 .env 文件
cd /home/eric/payment
cat > .env << 'EOF'
# 数据库配置
DB_PASSWORD=your-strong-password-here

# Redis 配置
REDIS_PASSWORD=your-redis-password

# JWT 密钥（生产环境必须修改！）
JWT_SECRET=your-super-secret-jwt-key-256-bits-minimum

# Stripe 配置（如果使用）
STRIPE_API_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# SMTP 配置（邮件通知）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@payment-platform.com
EOF

# 设置权限（防止泄露）
chmod 600 .env
```

### 步骤 4: 启动基础设施

```bash
# 启动基础设施容器
docker-compose up -d

# 查看日志
docker-compose logs -f postgres redis kafka

# 等待健康检查通过（约30秒）
docker-compose ps

# 应该看到所有服务 Status 为 "healthy" 或 "running"
```

### 步骤 5: 初始化数据库

```bash
# 进入 backend 目录
cd backend

# 运行初始化脚本（创建19个数据库）
./scripts/init-db.sh

# 验证数据库创建
docker exec -it payment-postgres psql -U postgres -c "\l"
```

### 步骤 6: 构建所有服务镜像

```bash
# 方式1: 使用自动化脚本（推荐）
cd backend
./scripts/build-all-docker-images.sh

# 方式2: 使用 docker-compose build
cd ..
docker-compose -f docker-compose.services.yml build
docker-compose -f docker-compose.bff.yml build

# 查看构建的镜像
docker images | grep payment-platform
```

### 步骤 7: 启动所有微服务

```bash
# 启动17个核心服务
docker-compose -f docker-compose.services.yml up -d

# 启动2个 BFF 服务
docker-compose -f docker-compose.bff.yml up -d

# 查看所有服务状态
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep payment-
```

### 步骤 8: 验证部署

```bash
# 健康检查脚本
cat > check-health.sh << 'EOF'
#!/bin/bash
SERVICES=(
  "payment-gateway:40003"
  "order-service:40004"
  "channel-adapter:40005"
  "risk-service:40006"
  "accounting-service:40007"
  "notification-service:40008"
  "analytics-service:40009"
  "config-service:40010"
  "merchant-auth-service:40011"
  "settlement-service:40013"
  "withdrawal-service:40014"
  "kyc-service:40015"
  "cashier-service:40016"
  "reconciliation-service:40020"
  "dispute-service:40021"
  "merchant-policy-service:40022"
  "merchant-quota-service:40024"
  "admin-bff-service:40001"
  "merchant-bff-service:40023"
)

for svc in "${SERVICES[@]}"; do
  IFS=':' read -r name port <<< "$svc"
  if curl -sf http://localhost:$port/health > /dev/null; then
    echo "✅ $name is healthy"
  else
    echo "❌ $name is unhealthy"
  fi
done
EOF

chmod +x check-health.sh
./check-health.sh
```

---

## ⚙️ 配置说明

### mTLS 配置

所有服务间通信使用 mTLS 加密：

```yaml
# 环境变量配置
ENABLE_MTLS=true
ENABLE_HTTPS=true
TLS_CERT_FILE=/app/certs/services/{service-name}/{service-name}.crt
TLS_KEY_FILE=/app/certs/services/{service-name}/{service-name}.key
TLS_CLIENT_CERT=/app/certs/services/{service-name}/{service-name}.crt
TLS_CLIENT_KEY=/app/certs/services/{service-name}/{service-name}.key
TLS_CA_FILE=/app/certs/ca/ca-cert.pem
```

### 服务间通信 URL

**内网域名格式**: `https://<service-name>.payment-network:<port>`

示例：
```bash
# Payment Gateway 调用 Order Service
ORDER_SERVICE_URL=https://order-service.payment-network:40004

# Payment Gateway 调用 Risk Service
RISK_SERVICE_URL=https://risk-service.payment-network:40006
```

### 资源限制

每个服务的默认资源配额：

```yaml
deploy:
  resources:
    limits:
      cpus: '1.0'          # 最多1个CPU核心
      memory: 512M         # 最多512MB内存
    reservations:
      cpus: '0.5'          # 预留0.5个CPU核心
      memory: 256M         # 预留256MB内存
```

### 日志配置

日志自动轮转：

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"        # 单个日志文件最大10MB
    max-file: "3"          # 保留最近3个日志文件
```

---

## 📊 监控与运维

### Prometheus 监控

访问: http://localhost:40090

**常用查询:**

```promql
# Payment Gateway 请求速率
rate(http_requests_total{service="payment-gateway"}[5m])

# Payment 成功率
sum(rate(payment_gateway_payment_total{status="success"}[5m]))
/ sum(rate(payment_gateway_payment_total[5m]))

# P95 延迟
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 服务内存使用
container_memory_usage_bytes{name=~"payment-.*"}
```

### Grafana 仪表板

访问: http://localhost:40300 (admin/admin)

**预配置仪表板:**
- 服务健康概览
- 支付流程监控
- 数据库性能
- Kafka 消息队列
- 容器资源使用

### Jaeger 分布式追踪

访问: http://localhost:50686

**使用场景:**
- 追踪支付完整流程（Gateway → Order → Channel → Risk → Accounting）
- 定位性能瓶颈
- 分析服务依赖关系
- 错误链路分析

### 日志查看

```bash
# 查看特定服务日志
docker-compose -f docker-compose.services.yml logs -f payment-gateway

# 查看所有服务日志
docker-compose -f docker-compose.services.yml logs -f

# 查看最近100行日志
docker logs --tail 100 payment-payment-gateway

# 查看实时日志（带时间戳）
docker logs -f --timestamps payment-payment-gateway
```

### 性能调优

```bash
# 扩展服务实例（水平扩展）
docker-compose -f docker-compose.services.yml up -d --scale payment-gateway=3

# 查看资源使用
docker stats

# 查看容器详细信息
docker inspect payment-payment-gateway
```

---

## 🔧 故障排查

### 常见问题

#### 1. 服务无法启动

**症状**: 容器状态为 "Restarting" 或 "Exited"

```bash
# 查看容器日志
docker logs payment-payment-gateway

# 查看退出原因
docker inspect payment-payment-gateway --format='{{.State.ExitCode}}'

# 常见原因:
# - 数据库连接失败（检查 DB_HOST）
# - 证书文件缺失（检查 /app/certs）
# - 端口冲突（lsof -i :40003）
```

#### 2. 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
docker ps | grep postgres

# 测试数据库连接
docker exec -it payment-postgres psql -U postgres -c "SELECT 1"

# 检查网络连接
docker exec payment-payment-gateway ping postgres.payment-network
```

#### 3. 服务间通信失败

```bash
# 检查 mTLS 证书
docker exec payment-payment-gateway ls -la /app/certs/services/payment-gateway/

# 验证证书有效性
docker exec payment-payment-gateway openssl x509 -in /app/certs/services/payment-gateway/payment-gateway.crt -text -noout

# 测试 HTTPS 连接
docker exec payment-payment-gateway curl -v --cacert /app/certs/ca/ca-cert.pem \
  --cert /app/certs/services/payment-gateway/payment-gateway.crt \
  --key /app/certs/services/payment-gateway/payment-gateway.key \
  https://order-service.payment-network:40004/health
```

#### 4. 内存/CPU 不足

```bash
# 查看资源使用
docker stats --no-stream

# 增加资源限制（修改 docker-compose.yml）
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 1024M

# 重启服务
docker-compose -f docker-compose.services.yml up -d payment-gateway
```

### 调试技巧

```bash
# 进入容器内部
docker exec -it payment-payment-gateway sh

# 查看环境变量
docker exec payment-payment-gateway env | grep -E "DB_|REDIS_|KAFKA_"

# 查看进程
docker exec payment-payment-gateway ps aux

# 查看网络配置
docker network inspect payment-network

# 查看卷挂载
docker inspect payment-payment-gateway --format='{{json .Mounts}}' | jq
```

---

## 🔒 安全最佳实践

### 1. 密钥管理

```bash
# ❌ 错误：硬编码密钥
JWT_SECRET=default-secret-key

# ✅ 正确：使用强密钥
JWT_SECRET=$(openssl rand -base64 32)

# 使用 Docker Secrets（生产环境）
echo "your-strong-password" | docker secret create db_password -
```

### 2. 网络隔离

```yaml
# 仅暴露必要的端口
ports:
  - "40003:40003"  # 仅外部访问的服务

# 其他服务不暴露端口，仅内网访问
expose:
  - "40004"
```

### 3. 最小权限原则

```dockerfile
# ✅ 非 root 用户运行
USER appuser

# ✅ 只读文件系统
volumes:
  - ./backend/certs:/app/certs:ro  # 只读挂载
```

### 4. 镜像安全

```bash
# 扫描镜像漏洞
docker scan payment-platform/payment-gateway:latest

# 使用 Alpine 基础镜像（最小化）
FROM alpine:3.19
```

### 5. 日志安全

```yaml
# 避免记录敏感信息
logging:
  options:
    labels: "com.payment.security=high"
    env: "ENV,SERVICE_NAME"  # 仅记录非敏感环境变量
```

---

## 📚 附录

### A. 完整命令速查

```bash
# 启动所有服务
docker-compose up -d
docker-compose -f docker-compose.services.yml up -d
docker-compose -f docker-compose.bff.yml up -d

# 停止所有服务
docker-compose -f docker-compose.bff.yml down
docker-compose -f docker-compose.services.yml down
docker-compose down

# 查看状态
docker-compose ps
docker ps

# 查看日志
docker-compose logs -f [service-name]

# 重启服务
docker-compose restart [service-name]

# 重建服务
docker-compose up -d --build [service-name]

# 清理
docker-compose down -v  # 删除卷
docker system prune -a  # 清理所有未使用资源
```

### B. 目录结构

```
/home/eric/payment/
├── docker-compose.yml              # 基础设施
├── docker-compose.services.yml     # 17个微服务
├── docker-compose.bff.yml          # 2个BFF服务
├── .env                            # 环境变量
├── backend/
│   ├── services/                   # 19个服务源码
│   │   ├── payment-gateway/
│   │   │   ├── Dockerfile
│   │   │   ├── .dockerignore
│   │   │   └── ...
│   │   └── ...
│   ├── certs/                      # mTLS证书
│   │   ├── ca/
│   │   └── services/
│   ├── scripts/
│   │   ├── generate-dockerfiles.sh
│   │   ├── generate-docker-compose-services.sh
│   │   └── build-all-docker-images.sh
│   └── logs/                       # 日志目录
└── frontend/                       # 前端应用
    ├── admin-portal/
    ├── merchant-portal/
    └── website/
```

### C. 联系与支持

- **文档**: [README.md](README.md)
- **架构**: [CLAUDE.md](CLAUDE.md)
- **问题反馈**: GitHub Issues

---

**🎉 部署完成！祝您使用愉快！**

# 本地开发环境指南

本文档介绍如何在本地开发机器上运行支付平台微服务，同时使用 Docker 运行基础设施组件。

## 🎯 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                      开发环境架构                                 │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────────┐         ┌──────────────────────────┐
│   本地开发机器 (localhost)  │         │   Docker 容器环境         │
│                          │         │                          │
│  微服务 (使用 Air 热加载)   │────────▶│  基础设施组件              │
│  • Admin Service:40000   │         │  • PostgreSQL:40432      │
│  • Merchant:40001        │         │  • Redis:40379           │
│  • Payment Gateway:40002 │         │  • Kafka:40092           │
│  • Channel Adapter:40003 │         │  • Prometheus:40090      │
│  • Order:40004           │         │  • Grafana:40300         │
│  • Accounting:40005      │         │  • Jaeger:40686          │
│  • Risk:40006            │         │  • cAdvisor:40180        │
│  • Notification:40007    │         │                          │
│  • Analytics:40008       │         │  监控 Exporters:          │
│  • Config:40009          │         │  • Postgres:40187        │
│                          │         │  • Redis:40121           │
│  前端开发服务器            │         │  • Kafka:40308           │
│                          │         │  • Node:40100            │
└──────────────────────────┘         └──────────────────────────┘
```

---

## 🚀 快速开始

### 1. 启动 Docker 基础设施

```bash
# 进入项目目录
cd /home/eric/payment

# 启动所有基础设施组件
docker-compose up -d

# 查看容器状态
docker-compose ps

# 查看容器日志
docker-compose logs -f
```

### 2. 等待数据库初始化

首次启动时，PostgreSQL 会自动执行初始化脚本：

```bash
# 检查 PostgreSQL 是否就绪
docker-compose logs postgres | grep "database system is ready"

# 或使用 healthcheck
docker-compose ps postgres
```

### 3. 启动微服务（本地）

使用提供的开发脚本或手动启动：

#### 方式 A: 使用开发脚本（推荐）

```bash
# 启动所有微服务
./scripts/dev-with-air.sh

# 停止所有微服务
./scripts/stop-services.sh
```

#### 方式 B: 手动启动单个服务

```bash
# 进入服务目录
cd backend/services/admin-service

# 使用 Air 启动（热加载）
air

# 或直接使用 Go 启动
go run cmd/main.go
```

### 4. 验证服务状态

```bash
# 检查所有微服务健康状态
for port in {40000..40009}; do
  echo "检查端口 $port..."
  curl -s http://localhost:$port/health || echo "端口 $port 未响应"
done

# 检查 Prometheus 监控
curl http://localhost:40090/api/v1/targets

# 访问 Grafana
open http://localhost:40300  # 用户名/密码: admin/admin
```

---

## 📋 环境变量配置

### 本地微服务配置

创建或编辑 `.env` 文件（基于 `.env.example`）：

```bash
# 数据库连接（连接到 Docker 中的 PostgreSQL）
DATABASE_URL=postgres://postgres:postgres@localhost:40432/payment_platform?sslmode=disable

# Redis 连接（连接到 Docker 中的 Redis）
REDIS_HOST=localhost
REDIS_PORT=40379

# Kafka 连接（连接到 Docker 中的 Kafka）
KAFKA_BROKERS=localhost:40092

# JWT 密钥
JWT_SECRET=dev-secret-key-change-in-production

# SMTP 配置（可选）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Stripe 配置（可选）
STRIPE_API_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
```

**重要提示**：
- 本地微服务使用 `localhost:40xxx` 访问 Docker 容器
- Docker 内部服务间通信使用服务名（如 `postgres:5432`）

---

## 🔧 开发工作流

### 日常开发流程

```bash
# 1. 启动 Docker 基础设施（仅需一次）
docker-compose up -d

# 2. 启动你正在开发的微服务
cd backend/services/payment-gateway
air  # 自动热加载

# 3. 进行代码修改
# Air 会自动检测文件变化并重新编译运行

# 4. 运行测试
go test ./...

# 5. 查看日志
# Air 日志会直接显示在终端
# Docker 日志可以通过 docker-compose logs 查看
```

### 调试技巧

#### 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试模式
cd backend/services/payment-gateway
dlv debug cmd/main.go --headless --listen=:2345 --api-version=2

# 在 VS Code 中连接调试器
# 使用 launch.json 配置远程调试
```

#### 查看实时日志

```bash
# 查看特定服务日志
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs -f kafka

# 查看所有基础设施日志
docker-compose logs -f
```

#### 数据库操作

```bash
# 连接到 PostgreSQL
psql postgres://postgres:postgres@localhost:40432/payment_platform

# 或使用 Docker exec
docker exec -it payment-postgres psql -U postgres -d payment_platform

# 运行数据库迁移
cd backend/services/admin-service
go run cmd/migrate.go
```

#### Redis 操作

```bash
# 使用 redis-cli
docker exec -it payment-redis redis-cli

# 常用命令
> KEYS *              # 查看所有键
> GET key_name        # 获取值
> FLUSHALL           # 清空所有数据（慎用）
```

#### Kafka 操作

```bash
# 列出所有 topics
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list

# 创建 topic
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 \
  --create --topic payment-events --partitions 3 --replication-factor 1

# 查看 topic 详情
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 \
  --describe --topic payment-events

# 消费消息（测试）
docker exec payment-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 --topic payment-events --from-beginning
```

---

## 📊 监控和调试

### Prometheus 监控

访问 Prometheus UI：http://localhost:40090

#### 常用查询

```promql
# 查看所有服务状态
up

# 查看本地微服务状态
up{tier="backend"}

# HTTP 请求速率
rate(http_requests_total[5m])

# 数据库连接数
pg_stat_database_numbackends

# Redis 内存使用
redis_memory_used_bytes
```

#### 检查监控目标

```bash
# 查看 Prometheus 是否能抓取到本地微服务
curl http://localhost:40090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

### Grafana 可视化

访问 Grafana：http://localhost:40300
- 默认用户名：`admin`
- 默认密码：`admin`

#### 导入推荐的 Dashboard

1. 点击 "+" → "Import"
2. 输入以下 Dashboard ID：
   - `1860` - Node Exporter Full（主机监控）
   - `893` - Docker & System Monitoring（容器监控）
   - `763` - Redis Dashboard
   - `9628` - PostgreSQL Database
   - `7589` - Kafka Exporter Overview

### Jaeger 分布式追踪

访问 Jaeger UI：http://localhost:40686

查看跨服务调用链路和性能瓶颈。

### cAdvisor 容器监控

访问 cAdvisor：http://localhost:40180

实时查看 Docker 容器的资源使用情况。

---

## 🛠️ 故障排查

### 问题 1: 微服务无法连接数据库

**症状**：
```
Error: could not connect to database: dial tcp [::1]:40432: connect: connection refused
```

**解决方案**：
```bash
# 1. 检查 PostgreSQL 容器是否运行
docker-compose ps postgres

# 2. 检查端口映射
netstat -an | grep 40432

# 3. 测试数据库连接
psql postgres://postgres:postgres@localhost:40432/payment_platform -c "SELECT 1"

# 4. 查看 PostgreSQL 日志
docker-compose logs postgres
```

### 问题 2: Air 热加载不工作

**症状**：修改代码后服务没有自动重启

**解决方案**：
```bash
# 1. 检查 .air.toml 配置
cat backend/services/your-service/.air.toml

# 2. 手动删除 tmp 目录
rm -rf backend/services/your-service/tmp

# 3. 重启 Air
cd backend/services/your-service
air
```

### 问题 3: Prometheus 无法抓取本地微服务

**症状**：Prometheus Targets 页面显示 `host.docker.internal` 不可达

**解决方案**：
```bash
# 1. 检查微服务是否在运行
curl http://localhost:40000/metrics

# 2. 测试 Docker 容器是否能访问主机
docker exec payment-prometheus ping host.docker.internal

# 3. 如果 host.docker.internal 不可用（Linux 系统）
# 编辑 docker-compose.yml，确认有 extra_hosts 配置：
# extra_hosts:
#   - "host.docker.internal:host-gateway"

# 4. 重启 Prometheus
docker-compose restart prometheus
```

### 问题 4: Kafka 连接失败

**症状**：
```
Error: kafka: client has run out of available brokers to talk to
```

**解决方案**：
```bash
# 1. 检查 Kafka 和 Zookeeper 状态
docker-compose ps kafka zookeeper

# 2. 查看 Kafka 日志
docker-compose logs kafka | tail -50

# 3. 重启 Kafka（注意顺序）
docker-compose restart zookeeper
sleep 10
docker-compose restart kafka

# 4. 测试连接
docker exec payment-kafka kafka-broker-api-versions \
  --bootstrap-server localhost:9092
```

### 问题 5: 端口被占用

**症状**：
```
Error: bind: address already in use
```

**解决方案**：
```bash
# 1. 查找占用端口的进程
lsof -i :40000  # 或其他端口号

# 2. 停止占用进程
kill -9 <PID>

# 3. 或使用 stop-services.sh 脚本
./scripts/stop-services.sh
```

---

## 📦 依赖管理

### Go 模块

```bash
# 安装依赖
go mod download

# 更新依赖
go get -u ./...
go mod tidy

# 查看依赖树
go mod graph

# 清理缓存
go clean -modcache
```

### Docker 镜像更新

```bash
# 拉取最新镜像
docker-compose pull

# 重新构建并启动
docker-compose up -d --build

# 清理未使用的镜像
docker image prune -a
```

---

## 🧪 测试

### 单元测试

```bash
# 运行所有测试
cd backend
go test ./...

# 运行特定服务的测试
cd backend/services/payment-gateway
go test ./... -v

# 运行带覆盖率的测试
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 集成测试

```bash
# 确保 Docker 环境运行
docker-compose up -d

# 运行集成测试
cd backend/services/payment-gateway
go test ./tests/integration/... -v

# 使用测试数据库
export DATABASE_URL=postgres://postgres:postgres@localhost:40432/payment_platform_test
go test ./tests/integration/...
```

### API 测试

```bash
# 使用 curl 测试 API
curl -X POST http://localhost:40000/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 使用 httpie（更友好）
http POST http://localhost:40000/api/v1/login \
  username=admin password=admin123

# 或使用 Postman Collection（如果有）
```

---

## 🔒 安全最佳实践

### 开发环境

1. **不要提交敏感信息**
   - `.env` 文件应该在 `.gitignore` 中
   - 使用 `.env.example` 作为模板
   - 永远不要提交真实的 API 密钥

2. **使用强密码**
   - 即使在开发环境也要避免使用 `admin/admin`
   - 定期更换 JWT_SECRET

3. **限制网络访问**
   - 开发环境只监听 localhost
   - 不要将开发端口暴露到公网

### 数据保护

```bash
# 定期备份开发数据库
docker exec payment-postgres pg_dump -U postgres payment_platform > backup.sql

# 恢复数据库
docker exec -i payment-postgres psql -U postgres payment_platform < backup.sql
```

---

## 📚 相关文档

- [Air 开发指南](./AIR_DEVELOPMENT.md)
- [监控系统配置](./MONITORING_SETUP.md)
- [端口映射表](./PORT_MAPPING.md)
- [配置检查清单](./CONFIGURATION_CHECKLIST.md)

---

## 🎓 开发技巧

### VS Code 配置

创建 `.vscode/launch.json` 用于调试：

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Payment Gateway",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/services/payment-gateway/cmd/main.go",
      "env": {
        "ENV": "development",
        "DATABASE_URL": "postgres://postgres:postgres@localhost:40432/payment_platform?sslmode=disable",
        "REDIS_HOST": "localhost",
        "REDIS_PORT": "40379",
        "KAFKA_BROKERS": "localhost:40092",
        "PORT": "40002"
      }
    }
  ]
}
```

### Git Hooks

推荐使用 pre-commit hook 确保代码质量：

```bash
# .git/hooks/pre-commit
#!/bin/bash
set -e

echo "Running go fmt..."
go fmt ./...

echo "Running go vet..."
go vet ./...

echo "Running tests..."
go test ./... -short
```

---

## 📝 常用命令速查表

```bash
# Docker 相关
docker-compose up -d              # 启动所有基础设施
docker-compose down               # 停止并删除容器
docker-compose restart <service>  # 重启特定服务
docker-compose logs -f <service>  # 查看服务日志
docker-compose ps                 # 查看服务状态

# 微服务相关
./scripts/dev-with-air.sh        # 启动所有微服务
./scripts/stop-services.sh       # 停止所有微服务
air                              # 在服务目录下启动单个服务
go run cmd/main.go               # 直接运行服务

# 数据库相关
psql postgres://...              # 连接数据库
docker exec -it payment-postgres psql -U postgres  # 进入 PG
go run cmd/migrate.go            # 运行迁移

# 测试相关
go test ./...                    # 运行所有测试
go test -v -run TestName         # 运行特定测试
go test -cover                   # 带覆盖率测试

# 监控相关
open http://localhost:40090      # Prometheus
open http://localhost:40300      # Grafana
open http://localhost:40686      # Jaeger
open http://localhost:40180      # cAdvisor
```

---

**更新时间**: 2025-10-23
**版本**: 1.0
**状态**: ✅ 混合开发环境配置完成

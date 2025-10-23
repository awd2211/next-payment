# 开发环境快速启动指南

## 🎯 开发环境架构说明

您的支付平台采用**混合开发架构**：

```
┌─────────────────────┐        ┌──────────────────────┐
│   本地开发机器       │  访问   │   Docker 容器         │
│   (localhost)       │ ─────▶ │   (基础设施)          │
├─────────────────────┤        ├──────────────────────┤
│ Go 微服务 (Air)     │        │ PostgreSQL  :40432   │
│ • 端口: 40000-40009 │        │ Redis       :40379   │
│                     │        │ Kafka       :40092   │
│ 前端开发服务器       │        │ Prometheus  :40090   │
│                     │        │ Grafana     :40300   │
│                     │        │ Jaeger      :40686   │
│                     │        │ cAdvisor    :40180   │
└─────────────────────┘        └──────────────────────┘
```

---

## 📦 第一步：启动 Docker 基础设施

```bash
# 进入项目目录
cd /home/eric/payment

# 启动所有基础设施容器
docker-compose up -d

# 等待所有容器启动（约 30-60 秒）
docker-compose ps

# 预期输出：所有容器状态为 "Up" 或 "Up (healthy)"
```

**包含的服务**：
- ✅ PostgreSQL (端口 40432)
- ✅ Redis (端口 40379)
- ✅ Kafka + Zookeeper (端口 40092)
- ✅ Prometheus (端口 40090)
- ✅ Grafana (端口 40300，默认 admin/admin)
- ✅ Jaeger (端口 40686)
- ✅ cAdvisor (端口 40180)
- ✅ PostgreSQL Exporter (端口 40187)
- ✅ Redis Exporter (端口 40121)
- ✅ Kafka Exporter (端口 40308)
- ✅ Node Exporter (端口 40100)

---

## 🔧 第二步：配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，确保以下配置正确：
# DATABASE_URL=postgres://postgres:postgres@localhost:40432/payment_platform?sslmode=disable
# REDIS_HOST=localhost
# REDIS_PORT=40379
# KAFKA_BROKERS=localhost:40092
```

**重要**：本地微服务连接 Docker 容器时使用 `localhost:40xxx`，而不是 Docker 内部服务名。

---

## 🚀 第三步：启动微服务

### 方式 A：启动所有微服务（推荐）

```bash
# 使用提供的脚本一键启动所有 10 个微服务
./scripts/dev-with-air.sh
```

### 方式 B：启动单个微服务（开发调试）

```bash
# 进入要开发的服务目录
cd backend/services/payment-gateway

# 使用 Air 启动（自动热加载）
air

# 或直接运行
go run cmd/main.go
```

---

## ✅ 第四步：验证环境

### 1. 检查基础设施

```bash
# PostgreSQL
psql postgres://postgres:postgres@localhost:40432/payment_platform -c "SELECT 1"

# Redis
docker exec payment-redis redis-cli ping

# Kafka
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### 2. 检查微服务

```bash
# 快速测试所有微服务
for port in {40000..40009}; do
  echo "测试端口 $port..."
  curl -s http://localhost:$port/health && echo " ✅ OK" || echo " ❌ 失败"
done
```

### 3. 检查监控

```bash
# 访问 Prometheus 查看所有目标
open http://localhost:40090/targets

# 访问 Grafana
open http://localhost:40300
# 用户名：admin
# 密码：admin

# 访问 Jaeger
open http://localhost:40686

# 访问 cAdvisor
open http://localhost:40180
```

---

## 📝 常用命令

### Docker 管理

```bash
# 查看所有容器状态
docker-compose ps

# 查看容器日志
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs -f kafka

# 重启某个服务
docker-compose restart postgres

# 停止所有容器
docker-compose down

# 停止并删除数据卷（慎用！会清空数据）
docker-compose down -v
```

### 微服务管理

```bash
# 启动所有微服务
./scripts/dev-with-air.sh

# 停止所有微服务
./scripts/stop-services.sh

# 查看运行中的微服务进程
ps aux | grep "air\|go run"
```

### 数据库操作

```bash
# 连接 PostgreSQL
psql postgres://postgres:postgres@localhost:40432/payment_platform

# 运行数据库迁移（在各服务目录下）
cd backend/services/admin-service
go run cmd/migrate.go

# 备份数据库
docker exec payment-postgres pg_dump -U postgres payment_platform > backup.sql

# 恢复数据库
docker exec -i payment-postgres psql -U postgres payment_platform < backup.sql
```

### 监控查询

```bash
# Prometheus 查询示例
curl 'http://localhost:40090/api/v1/query?query=up'

# 查看所有 targets 状态
curl 'http://localhost:40090/api/v1/targets' | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

---

## 🐛 常见问题

### 问题 1：端口被占用

```bash
# 查找占用端口的进程
lsof -i :40000  # 或其他端口

# 杀死进程
kill -9 <PID>
```

### 问题 2：Prometheus 显示微服务 DOWN

**原因**：微服务没有在本地运行

**解决**：
```bash
# 启动微服务
./scripts/dev-with-air.sh

# 或单独启动某个服务
cd backend/services/payment-gateway
air
```

### 问题 3：数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
docker-compose ps postgres

# 查看日志
docker-compose logs postgres

# 重启 PostgreSQL
docker-compose restart postgres

# 测试连接
psql postgres://postgres:postgres@localhost:40432/payment_platform -c "SELECT version()"
```

### 问题 4：Kafka 连接超时

```bash
# Kafka 启动较慢，确保 Zookeeper 先启动
docker-compose restart zookeeper
sleep 10
docker-compose restart kafka

# 查看 Kafka 日志
docker-compose logs kafka | tail -50
```

### 问题 5：host.docker.internal 不可用（Linux）

**症状**：Prometheus 无法通过 `host.docker.internal` 访问本地服务

**解决**：
```bash
# 检查 docker-compose.yml 中 prometheus 服务的 extra_hosts 配置
# 应该包含：
# extra_hosts:
#   - "host.docker.internal:host-gateway"

# 重启 Prometheus
docker-compose restart prometheus

# 验证
docker exec payment-prometheus ping -c 1 host.docker.internal
```

---

## 📚 详细文档

- **[本地开发环境完整指南](./docs/LOCAL_DEVELOPMENT.md)** - 详细的开发流程和最佳实践
- **[监控系统配置](./docs/MONITORING_SETUP.md)** - Prometheus、Grafana 完整配置
- **[Air 热加载指南](./docs/AIR_DEVELOPMENT.md)** - Air 工具使用说明
- **[端口映射表](./docs/PORT_MAPPING.md)** - 所有端口分配详情
- **[配置检查清单](./docs/CONFIGURATION_CHECKLIST.md)** - 配置验证清单

---

## 🎓 开发工作流程

```bash
# 1. 每天开始工作
cd /home/eric/payment
docker-compose up -d                    # 启动基础设施
./scripts/dev-with-air.sh              # 启动微服务

# 2. 开发过程
# - 编辑代码，Air 会自动检测并重新编译
# - 运行测试：go test ./...
# - 查看日志：终端输出 + docker-compose logs

# 3. 结束工作
./scripts/stop-services.sh             # 停止微服务
docker-compose down                     # 停止基础设施（可选，可以保持运行）
```

---

## 🔐 安全提醒

- ⚠️ `.env` 文件包含敏感信息，**不要提交到 Git**
- ⚠️ 修改默认密码（PostgreSQL、Grafana 等）
- ⚠️ 开发端口仅监听 localhost，不要暴露到公网
- ⚠️ 定期备份开发数据库

---

## 📊 监控面板地址

| 工具 | 地址 | 用户名/密码 | 用途 |
|------|------|-----------|------|
| Prometheus | http://localhost:40090 | 无需认证 | 指标查询和告警 |
| Grafana | http://localhost:40300 | admin/admin | 数据可视化 |
| Jaeger | http://localhost:40686 | 无需认证 | 分布式追踪 |
| cAdvisor | http://localhost:40180 | 无需认证 | 容器监控 |

---

## 🚀 下一步

1. ✅ 启动 Docker 基础设施：`docker-compose up -d`
2. ✅ 配置环境变量：编辑 `.env` 文件
3. ✅ 启动微服务：`./scripts/dev-with-air.sh`
4. ✅ 访问 Grafana 设置监控面板：http://localhost:40300
5. 📖 阅读详细文档：[docs/LOCAL_DEVELOPMENT.md](./docs/LOCAL_DEVELOPMENT.md)

---

**版本**: 1.0
**更新时间**: 2025-10-23
**环境**: 本地开发 + Docker 混合架构
**状态**: ✅ 就绪

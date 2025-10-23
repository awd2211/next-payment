# 支付平台开发环境配置总结

## ✅ 已完成的配置

### 🏗️ 架构调整

根据您的需求，我们将开发环境配置为**混合架构**：

- **Docker 容器运行**：基础设施组件（PostgreSQL、Redis、Kafka、监控工具）
- **本地机器运行**：所有 Go 微服务（使用 Air 热加载）+ 前端开发服务器
- **稍后打包**：生产环境再将微服务容器化

### 📝 主要变更

#### 1. Docker Compose 配置更新

**文件**: `docker-compose.yml`

**变更内容**:
- ✅ 移除了所有 10 个微服务容器定义
- ✅ 保留基础设施组件：
  - PostgreSQL (端口 40432)
  - Redis (端口 40379)
  - Kafka + Zookeeper (端口 40092)
  - Prometheus (端口 40090)
  - Grafana (端口 40300)
  - Jaeger (端口 40686)
  - cAdvisor (端口 40180)
  - PostgreSQL Exporter (端口 40187)
  - Redis Exporter (端口 40121)
  - Kafka Exporter (端口 40308)
  - Node Exporter (端口 40100)
- ✅ Prometheus 添加 `extra_hosts` 配置以支持 `host.docker.internal`
- ✅ 移除了 API Gateway (Traefik) - 开发环境不需要

#### 2. Prometheus 监控配置更新

**文件**: `backend/deployments/prometheus/prometheus.yml`

**变更内容**:
- ✅ 所有微服务监控目标改为 `host.docker.internal:40000-40009`
- ✅ 添加 Kafka Exporter 监控配置
- ✅ 保持基础设施组件使用 Docker 内部服务名

**示例配置**:
```yaml
# 微服务（本地运行）
- job_name: 'payment-gateway'
  static_configs:
    - targets: ['host.docker.internal:40002']

# 基础设施（Docker 运行）
- job_name: 'postgres'
  static_configs:
    - targets: ['postgres-exporter:9187']
```

#### 3. 开发脚本更新

**文件**: `scripts/dev-with-air.sh`

**变更内容**:
- ✅ 更新为包含所有 10 个微服务
- ✅ 使用正确的外部端口 (40000-40009)
- ✅ 添加服务访问地址显示

**服务列表**:
```
admin-service        → localhost:40000
merchant-service     → localhost:40001
payment-gateway      → localhost:40002
channel-adapter      → localhost:40003
order-service        → localhost:40004
accounting-service   → localhost:40005
risk-service         → localhost:40006
notification-service → localhost:40007
analytics-service    → localhost:40008
config-service       → localhost:40009
```

#### 4. 新建文档

##### **`DEVELOPMENT_QUICKSTART.md`** - 快速启动指南
- 包含一步步启动说明
- 常用命令速查表
- 常见问题解决方案
- 监控面板访问地址

##### **`docs/LOCAL_DEVELOPMENT.md`** - 完整开发指南
- 详细的开发环境架构说明
- 日常开发工作流程
- 调试技巧（Delve、日志、数据库操作）
- 测试指南（单元测试、集成测试、API 测试）
- 安全最佳实践
- VS Code 配置示例

##### 更新 **`docs/MONITORING_SETUP.md`**
- 添加混合环境说明
- 更新微服务监控目标表格
- 更新故障排查指南

---

## 🚀 快速开始

### 1. 启动 Docker 基础设施

```bash
cd /home/eric/payment
docker-compose up -d
```

### 2. 等待服务就绪

```bash
# 查看容器状态
docker-compose ps

# 所有容器应该显示 "Up" 状态
```

### 3. 配置环境变量

```bash
# 复制模板
cp .env.example .env

# 编辑 .env，确保以下配置：
# DATABASE_URL=postgres://postgres:postgres@localhost:40432/payment_platform?sslmode=disable
# REDIS_HOST=localhost
# REDIS_PORT=40379
# KAFKA_BROKERS=localhost:40092
```

### 4. 启动微服务

```bash
# 启动所有 10 个微服务
./scripts/dev-with-air.sh
```

### 5. 验证

```bash
# 测试微服务
for port in {40000..40009}; do
  curl -s http://localhost:$port/health && echo "✅ Port $port OK"
done

# 查看 Prometheus 监控
open http://localhost:40090/targets

# 访问 Grafana
open http://localhost:40300  # admin/admin
```

---

## 📊 端口分配总览

### 基础设施（Docker）

| 服务 | 内部端口 | 外部端口 | 访问地址 |
|------|---------|---------|----------|
| PostgreSQL | 5432 | 40432 | localhost:40432 |
| Redis | 6379 | 40379 | localhost:40379 |
| Kafka | 9092 | 40092 | localhost:40092 |
| Prometheus | 9090 | 40090 | http://localhost:40090 |
| Grafana | 3000 | 40300 | http://localhost:40300 |
| Jaeger UI | 16686 | 40686 | http://localhost:40686 |
| cAdvisor | 8080 | 40180 | http://localhost:40180 |

### 监控 Exporters（Docker）

| Exporter | 内部端口 | 外部端口 | 访问地址 |
|----------|---------|---------|----------|
| PostgreSQL Exporter | 9187 | 40187 | http://localhost:40187/metrics |
| Redis Exporter | 9121 | 40121 | http://localhost:40121/metrics |
| Kafka Exporter | 9308 | 40308 | http://localhost:40308/metrics |
| Node Exporter | 9100 | 40100 | http://localhost:40100/metrics |

### 微服务（本地）

| 服务 | 外部端口 | 访问地址 | 说明 |
|------|---------|----------|------|
| admin-service | 40000 | http://localhost:40000 | 运营管理 |
| merchant-service | 40001 | http://localhost:40001 | 商户管理 |
| payment-gateway | 40002 | http://localhost:40002 | 支付网关 |
| channel-adapter | 40003 | http://localhost:40003 | 渠道适配 |
| order-service | 40004 | http://localhost:40004 | 订单服务 |
| accounting-service | 40005 | http://localhost:40005 | 账务服务 |
| risk-service | 40006 | http://localhost:40006 | 风控服务 |
| notification-service | 40007 | http://localhost:40007 | 通知服务 |
| analytics-service | 40008 | http://localhost:40008 | 分析服务 |
| config-service | 40009 | http://localhost:40009 | 配置中心 |

---

## 🔧 环境变量配置

### 本地微服务连接 Docker

微服务在本地运行时，通过 `localhost` 访问 Docker 容器：

```bash
# .env 文件内容
DATABASE_URL=postgres://postgres:postgres@localhost:40432/payment_platform?sslmode=disable
REDIS_HOST=localhost
REDIS_PORT=40379
KAFKA_BROKERS=localhost:40092
```

### Prometheus 监控本地微服务

Prometheus 在 Docker 中运行，通过 `host.docker.internal` 访问本地服务：

```yaml
# prometheus.yml
- job_name: 'payment-gateway'
  static_configs:
    - targets: ['host.docker.internal:40002']
```

**重要**: `docker-compose.yml` 中已配置 `extra_hosts` 确保这个域名可用。

---

## 📖 文档导航

| 文档 | 用途 | 推荐阅读顺序 |
|------|------|------------|
| **[DEVELOPMENT_QUICKSTART.md](./DEVELOPMENT_QUICKSTART.md)** | 快速启动指南 | ⭐ 首先阅读 |
| **[docs/LOCAL_DEVELOPMENT.md](./docs/LOCAL_DEVELOPMENT.md)** | 完整开发指南 | ⭐ 深入学习 |
| **[docs/MONITORING_SETUP.md](./docs/MONITORING_SETUP.md)** | 监控系统配置 | 设置监控时 |
| **[docs/AIR_DEVELOPMENT.md](./docs/AIR_DEVELOPMENT.md)** | Air 热加载指南 | 需要时参考 |
| **[docs/PORT_MAPPING.md](./docs/PORT_MAPPING.md)** | 端口映射表 | 需要时参考 |
| **[docs/CONFIGURATION_CHECKLIST.md](./docs/CONFIGURATION_CHECKLIST.md)** | 配置检查清单 | 验证配置时 |

---

## 🔍 关键概念

### host.docker.internal

这是 Docker 提供的特殊域名，允许容器访问宿主机（Host）上的服务。

- **Linux 系统**: 需要在 docker-compose.yml 中配置 `extra_hosts`
- **Mac/Windows Docker Desktop**: 自动可用

```yaml
# docker-compose.yml
services:
  prometheus:
    extra_hosts:
      - "host.docker.internal:host-gateway"  # 关键配置
```

### 端口映射

- **内部端口**: 服务在容器/进程内部监听的端口（如 8000-8009）
- **外部端口**: 从外部访问时使用的端口（如 40000-40009）

**示例**:
```yaml
# Docker 端口映射
ports:
  - "40432:5432"  # 外部:内部

# 本地微服务
# 直接监听外部端口 40000-40009
PORT=40000 go run cmd/main.go
```

### Air 热加载

Air 会监控 Go 源代码文件的变化，自动重新编译和运行服务。

**配置文件**: 每个服务的 `.air.toml`

**启动方式**:
```bash
cd backend/services/payment-gateway
air  # 自动检测文件变化
```

---

## 🛠️ 常见任务

### 查看所有服务状态

```bash
# Docker 服务
docker-compose ps

# 本地微服务
ps aux | grep "air\|go run"

# 或使用脚本检查
for port in {40000..40009}; do
  curl -s http://localhost:$port/health && echo "✅ $port"
done
```

### 重启特定服务

```bash
# Docker 服务
docker-compose restart postgres

# 本地微服务
# 1. 找到进程 PID
ps aux | grep payment-gateway

# 2. 终止进程
kill <PID>

# 3. 重新启动
cd backend/services/payment-gateway
air
```

### 查看日志

```bash
# Docker 服务日志
docker-compose logs -f postgres
docker-compose logs -f prometheus

# 本地微服务日志
tail -f backend/logs/payment-gateway.log
tail -f backend/logs/accounting-service.log

# 所有微服务日志
tail -f backend/logs/*.log
```

### 数据库操作

```bash
# 连接数据库
psql postgres://postgres:postgres@localhost:40432/payment_platform

# 执行 SQL 文件
psql postgres://postgres:postgres@localhost:40432/payment_platform < backup.sql

# 备份数据库
docker exec payment-postgres pg_dump -U postgres payment_platform > backup_$(date +%Y%m%d).sql
```

### 停止所有服务

```bash
# 停止本地微服务
./scripts/stop-services.sh

# 停止 Docker 容器
docker-compose down

# 停止 Docker 并删除数据卷（慎用）
docker-compose down -v
```

---

## ⚠️ 注意事项

### 1. 启动顺序

推荐的启动顺序：
```bash
1. docker-compose up -d      # 启动基础设施
2. 等待 30-60 秒             # 让数据库初始化
3. ./scripts/dev-with-air.sh # 启动微服务
```

### 2. 端口冲突

如果遇到端口被占用：
```bash
# 查找占用进程
lsof -i :40000

# 终止进程
kill -9 <PID>
```

### 3. host.docker.internal 不可用

如果 Prometheus 无法访问本地服务：
```bash
# 检查配置
docker exec payment-prometheus cat /etc/hosts | grep host.docker.internal

# 如果没有，重启 docker-compose
docker-compose down
docker-compose up -d
```

### 4. 数据持久化

Docker volumes 会持久化数据：
- `postgres_data`: PostgreSQL 数据库
- `redis_data`: Redis 数据
- `prometheus_data`: Prometheus 指标历史
- `grafana_data`: Grafana 配置和面板

除非使用 `docker-compose down -v`，否则数据会保留。

---

## 🎯 下一步建议

1. **✅ 完成基础环境启动**
   ```bash
   docker-compose up -d
   ./scripts/dev-with-air.sh
   ```

2. **📊 配置监控面板**
   - 访问 Grafana: http://localhost:40300
   - 导入推荐的 Dashboard（参考 MONITORING_SETUP.md）
   - 创建自定义业务监控面板

3. **🔧 开发第一个功能**
   - 选择一个服务进行开发
   - 修改代码，观察 Air 自动重新加载
   - 通过 Prometheus 查看指标变化

4. **🧪 编写测试**
   - 单元测试：`go test ./...`
   - 集成测试：连接实际的 Docker 服务
   - API 测试：使用 curl 或 Postman

5. **📝 熟悉工具**
   - 学习 Prometheus PromQL 查询语言
   - 探索 Grafana 可视化功能
   - 使用 Jaeger 追踪请求链路

---

## 🤝 支持和帮助

如果遇到问题：

1. **查看文档**
   - [DEVELOPMENT_QUICKSTART.md](./DEVELOPMENT_QUICKSTART.md) - 快速解决方案
   - [docs/LOCAL_DEVELOPMENT.md](./docs/LOCAL_DEVELOPMENT.md) - 详细故障排查

2. **检查日志**
   ```bash
   docker-compose logs <service>
   tail -f backend/logs/<service>.log
   ```

3. **验证配置**
   ```bash
   # 测试数据库连接
   psql postgres://postgres:postgres@localhost:40432/payment_platform -c "SELECT 1"

   # 测试 Redis 连接
   docker exec payment-redis redis-cli ping

   # 测试微服务健康
   curl http://localhost:40000/health
   ```

---

**配置完成时间**: 2025-10-23
**版本**: 1.0
**环境**: 本地开发 + Docker 混合架构
**状态**: ✅ 完全就绪

祝开发顺利！🚀

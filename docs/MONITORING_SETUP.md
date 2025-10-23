# 监控系统配置指南

本文档介绍支付平台的完整监控系统配置，包括 Prometheus、Grafana、各种 Exporters 和告警配置。

## ⚠️ 重要说明：混合开发环境

**开发环境架构**：
- **微服务**：在本地开发机器运行（使用 Air 热加载），监听 `localhost:40000-40009`
- **基础设施**：在 Docker 容器运行（PostgreSQL、Redis、Kafka、监控工具等）
- **监控连接**：Prometheus 通过 `host.docker.internal` 访问本地微服务

如需了解完整的本地开发环境配置，请参考 [本地开发指南](./LOCAL_DEVELOPMENT.md)。

## 📊 监控架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         监控系统架构                             │
└─────────────────────────────────────────────────────────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Grafana    │────▶│  Prometheus  │────▶│  Exporters   │
│  (可视化)     │     │  (指标收集)   │     │  (指标暴露)   │
│  Port: 40300 │     │  Port: 40090 │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │  告警规则     │
                    │  (Alerts)    │
                    └──────────────┘

Exporters 包括:
• PostgreSQL Exporter (40187)  - 数据库监控
• Redis Exporter (40121)       - 缓存监控
• Kafka Exporter (40308)       - 消息队列监控
• cAdvisor (40180)             - 容器监控
• Node Exporter (40100)        - 主机监控
```

---

## 🚀 快速开始

### 1. 启动所有监控组件

```bash
# 启动完整的监控栈
docker-compose up -d

# 验证监控组件状态
docker-compose ps | grep -E '(prometheus|grafana|exporter|cadvisor|node)'
```

### 2. 访问监控界面

| 组件 | 访问地址 | 默认用户名/密码 | 说明 |
|------|---------|----------------|------|
| **Grafana** | http://localhost:40300 | admin/admin | 数据可视化 |
| **Prometheus** | http://localhost:40090 | 无需认证 | 指标查询 |
| **cAdvisor** | http://localhost:40180 | 无需认证 | 容器监控 |

---

## 📈 Prometheus 配置

### 配置文件位置

```
backend/deployments/prometheus/
├── prometheus.yml          # 主配置文件
└── alerts/
    └── service_alerts.yml  # 告警规则
```

### 监控目标

Prometheus 已配置监控以下目标：

#### 基础设施监控

| 目标 | 地址 | 说明 |
|------|------|------|
| PostgreSQL | postgres-exporter:9187 | 数据库连接数、慢查询、性能指标 |
| Redis | redis-exporter:9121 | 内存使用、连接数、命中率 |
| Kafka | kafka-exporter:9308 | 消息队列、消费者延迟、分区状态 |

#### 微服务监控（本地运行）

所有微服务在本地开发机器运行，在 `/metrics` 端点暴露指标。Prometheus 通过 `host.docker.internal` 访问：

| 服务 | 本地地址 | Prometheus 抓取地址 | 关键指标 |
|------|---------|-------------------|---------|
| admin-service | localhost:40000/metrics | host.docker.internal:40000 | 用户认证、权限管理 |
| merchant-service | localhost:40001/metrics | host.docker.internal:40001 | 商户注册、审核 |
| payment-gateway | localhost:40002/metrics | host.docker.internal:40002 | 支付成功率、响应时间 |
| channel-adapter | localhost:40003/metrics | host.docker.internal:40003 | 渠道调用、成功率 |
| order-service | localhost:40004/metrics | host.docker.internal:40004 | 订单创建、状态变更 |
| accounting-service | localhost:40005/metrics | host.docker.internal:40005 | 账户操作、交易数量 |
| risk-service | localhost:40006/metrics | host.docker.internal:40006 | 风控检查、拦截率 |
| notification-service | localhost:40007/metrics | host.docker.internal:40007 | 通知发送、投递状态 |
| analytics-service | localhost:40008/metrics | host.docker.internal:40008 | 数据分析查询 |
| config-service | localhost:40009/metrics | host.docker.internal:40009 | 配置读取次数 |

**注意**：微服务必须在本地运行，否则 Prometheus 会显示这些目标为 DOWN 状态。使用 `./scripts/dev-with-air.sh` 启动所有微服务。

#### 容器和主机监控

| 目标 | 地址 | 说明 |
|------|------|------|
| cAdvisor | cadvisor:8080 | CPU、内存、网络、磁盘使用 |
| Node Exporter | node-exporter:9100 | 主机级别的系统指标 |

### 查询示例

```promql
# 服务可用性
up{job="accounting-service"}

# HTTP 请求率
rate(http_requests_total[5m])

# HTTP 错误率
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# 响应时间 P95
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 数据库连接数
pg_stat_database_numbackends

# Redis 内存使用
redis_memory_used_bytes

# Kafka 消费者延迟
kafka_consumergroup_lag

# 容器 CPU 使用率
rate(container_cpu_usage_seconds_total[5m])

# 容器内存使用
container_memory_usage_bytes
```

---

## 🎨 Grafana 配置

### 自动配置

Grafana 已自动配置：

1. **数据源**: Prometheus (http://prometheus:9090)
2. **Dashboard**: 自动加载 `provisioning/dashboards/json/` 目录下的仪表板
3. **默认密码**: admin/admin（首次登录后会要求修改）

### Dashboard 推荐

访问 Grafana 后，可以导入以下官方 Dashboard：

| Dashboard ID | 名称 | 用途 |
|--------------|------|------|
| 1860 | Node Exporter Full | 主机监控 |
| 893 | Docker & System Monitoring | 容器监控 |
| 763 | Redis Dashboard | Redis 监控 |
| 9628 | PostgreSQL Database | PostgreSQL 监控 |
| 7589 | Kafka Exporter Overview | Kafka 监控 |

#### 导入方法

1. 登录 Grafana (http://localhost:40300)
2. 点击 "+" -> "Import"
3. 输入 Dashboard ID
4. 选择 Prometheus 数据源
5. 点击 "Import"

### 自定义 Dashboard

在 `backend/deployments/grafana/provisioning/dashboards/json/` 目录下创建 JSON 文件即可自动加载。

---

## 🔔 告警配置

### 已配置的告警规则

#### 1. 服务可用性告警

- **ServiceDown**: 服务下线超过 1 分钟
- **CriticalServiceDown**: 关键服务（支付网关、风控等）下线超过 30 秒

#### 2. 数据库告警

- **PostgreSQLConnectionsHigh**: 连接数超过 80
- **PostgreSQLSlowQueries**: 查询性能下降

#### 3. Redis 告警

- **RedisMemoryHigh**: 内存使用率超过 90%
- **RedisConnectionsHigh**: 连接数超过 1000

#### 4. 应用性能告警

- **HighErrorRate**: HTTP 错误率超过 5%
- **HighResponseTime**: P95 响应时间超过 1 秒
- **HighCPUUsage**: CPU 使用率超过 80%
- **HighMemoryUsage**: 内存使用率超过 90%

#### 5. 业务指标告警

- **HighPaymentFailureRate**: 支付失败率超过 10%
- **HighOrderCancellationRate**: 订单取消率超过 20%
- **AbnormalRiskBlockRate**: 风控拦截率超过 30%

### 查看告警

访问 Prometheus 告警页面：
```
http://localhost:40090/alerts
```

### 配置告警通知

编辑 `prometheus.yml` 添加 Alertmanager 配置（可选）：

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

---

## 📦 监控组件详情

### 1. PostgreSQL Exporter

**端口**: 40187
**Docker 容器**: payment-postgres-exporter
**访问**: http://localhost:40187/metrics

**关键指标**:
- `pg_stat_database_numbackends` - 活跃连接数
- `pg_stat_database_tup_returned` - 返回行数
- `pg_stat_database_tup_fetched` - 获取行数
- `pg_stat_database_conflicts` - 冲突数

### 2. Redis Exporter

**端口**: 40121
**Docker 容器**: payment-redis-exporter
**访问**: http://localhost:40121/metrics

**关键指标**:
- `redis_memory_used_bytes` - 内存使用
- `redis_connected_clients` - 连接客户端数
- `redis_keyspace_hits_total` - 缓存命中
- `redis_keyspace_misses_total` - 缓存未命中

### 3. Kafka Exporter

**端口**: 40308
**Docker 容器**: payment-kafka-exporter
**访问**: http://localhost:40308/metrics

**关键指标**:
- `kafka_brokers` - Broker 数量
- `kafka_topic_partitions` - 分区数
- `kafka_consumergroup_lag` - 消费者延迟
- `kafka_topic_partition_current_offset` - 当前偏移量

### 4. cAdvisor

**端口**: 40180
**Docker 容器**: payment-cadvisor
**访问**: http://localhost:40180

**关键指标**:
- `container_cpu_usage_seconds_total` - CPU 使用
- `container_memory_usage_bytes` - 内存使用
- `container_network_receive_bytes_total` - 网络接收
- `container_network_transmit_bytes_total` - 网络发送

### 5. Node Exporter

**端口**: 40100
**Docker 容器**: payment-node-exporter
**访问**: http://localhost:40100/metrics

**关键指标**:
- `node_cpu_seconds_total` - CPU 时间
- `node_memory_MemAvailable_bytes` - 可用内存
- `node_disk_read_bytes_total` - 磁盘读取
- `node_disk_write_bytes_total` - 磁盘写入

---

## 🔧 故障排查

### 问题 1: Prometheus 无法抓取本地微服务指标

**症状**: Prometheus Targets 页面显示微服务 DOWN，错误信息：`dial tcp: lookup host.docker.internal`

**解决方案**:
```bash
# 1. 确认微服务在本地运行
curl http://localhost:40000/health  # Admin Service
curl http://localhost:40002/health  # Payment Gateway
# ... 测试其他服务

# 2. 检查 Prometheus 容器是否能解析 host.docker.internal
docker exec payment-prometheus ping -c 1 host.docker.internal

# 3. 如果 host.docker.internal 不可用（Linux 原生 Docker）
# 检查 docker-compose.yml 中的 extra_hosts 配置
docker exec payment-prometheus cat /etc/hosts | grep host.docker.internal

# 4. 重启 Prometheus 以应用配置
docker-compose restart prometheus

# 5. 测试从容器内访问本地服务
docker exec payment-prometheus wget -O- http://host.docker.internal:40000/metrics
```

### 问题 2: Grafana 无法连接 Prometheus

**症状**: Grafana 数据源测试失败

**解决方案**:
```bash
# 检查 Prometheus 是否运行
docker-compose ps prometheus

# 从 Grafana 容器测试连接
docker exec payment-grafana curl http://prometheus:9090/api/v1/status/config
```

### 问题 3: Exporters 无法连接到目标服务

**症状**: Exporter 日志显示连接错误

**解决方案**:
```bash
# PostgreSQL Exporter
docker-compose logs postgres-exporter

# Redis Exporter
docker-compose logs redis-exporter

# Kafka Exporter
docker-compose logs kafka-exporter

# 检查网络连接
docker exec payment-postgres-exporter ping postgres
```

### 问题 4: cAdvisor 无法启动

**症状**: cAdvisor 容器一直重启

**解决方案**:
```bash
# 检查 Docker socket 权限
ls -l /var/run/docker.sock

# 查看 cAdvisor 日志
docker-compose logs cadvisor

# 可能需要 SELinux 配置（CentOS/RHEL）
sudo setenforce 0
```

---

## 📊 监控最佳实践

### 1. 指标命名规范

所有微服务应遵循 Prometheus 命名约定：

```
<metric_name>_<unit>_<type>

示例:
http_requests_total              # 计数器
http_request_duration_seconds    # 直方图
payment_amount_total             # 计数器
risk_check_score                 # 仪表
```

### 2. 标签使用

合理使用标签进行指标分组：

```promql
http_requests_total{
  service="payment-gateway",
  method="POST",
  status="200",
  endpoint="/api/v1/payments"
}
```

### 3. 告警阈值调整

根据实际业务情况调整告警阈值：

- **开发环境**: 适当放宽阈值
- **生产环境**: 严格监控关键指标
- **定期review**: 根据历史数据调整

### 4. Dashboard 组织

建议按照以下维度组织 Dashboard：

- **Overview Dashboard**: 整体服务状态
- **Service Dashboard**: 单个服务详情
- **Infrastructure Dashboard**: 基础设施监控
- **Business Dashboard**: 业务指标监控

---

## 🎯 下一步

1. **配置告警通知**: 集成 Slack、Email 等通知渠道
2. **添加业务指标**: 在代码中添加自定义业务指标
3. **优化 Dashboard**: 创建符合团队需求的可视化面板
4. **设置 SLO/SLI**: 定义服务级别目标和指标
5. **日志聚合**: 集成 ELK/Loki 进行日志分析

---

## 📚 参考资料

- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
- [PromQL 查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Exporters 列表](https://prometheus.io/docs/instrumenting/exporters/)

---

## 📝 监控组件端口总览

| 组件 | 内部端口 | 外部端口 | 访问地址 |
|------|---------|---------|----------|
| Prometheus | 9090 | 40090 | http://localhost:40090 |
| Grafana | 3000 | 40300 | http://localhost:40300 |
| PostgreSQL Exporter | 9187 | 40187 | http://localhost:40187/metrics |
| Redis Exporter | 9121 | 40121 | http://localhost:40121/metrics |
| Kafka Exporter | 9308 | 40308 | http://localhost:40308/metrics |
| cAdvisor | 8080 | 40180 | http://localhost:40180 |
| Node Exporter | 9100 | 40100 | http://localhost:40100/metrics |

---

**更新时间**: 2025-10-23
**版本**: 1.0
**状态**: ✅ 监控系统已配置完成

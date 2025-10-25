# 支付平台基础设施参考

## 快速命令

```bash
# 查看所有基础设施状态
./scripts/manage-services.sh infra status

# 启动所有基础设施
./scripts/manage-services.sh infra start

# 停止所有基础设施
./scripts/manage-services.sh infra stop

# 重启所有基础设施
./scripts/manage-services.sh infra restart
```

## 基础设施组件清单 (17个)

### 核心数据存储 (2个)

| 组件 | 容器名 | 端口 | 用途 |
|------|--------|------|------|
| PostgreSQL | payment-postgres | 40432 | 主数据库（16个业务数据库） |
| Redis | payment-redis | 40379 | 缓存和会话存储 |

**连接测试**:
```bash
# PostgreSQL
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -c "SELECT version();"

# Redis
redis-cli -h localhost -p 40379 ping
```

### 消息队列 (3个)

| 组件 | 容器名 | 端口 | 用途 |
|------|--------|------|------|
| Zookeeper | payment-zookeeper | 42181 | Kafka 协调服务 |
| Kafka | payment-kafka | 40092, 40093 | 消息队列 |
| Kafka UI | payment-kafka-ui | 40084 | Kafka 可视化管理 |

**访问地址**:
- Kafka UI: http://localhost:40084

**连接配置**:
```bash
# Kafka Brokers
KAFKA_BROKERS=localhost:40092

# 内部访问
kafka:9092

# 外部访问
localhost:40093
```

### API 网关 (3个)

| 组件 | 容器名 | 端口 | 用途 |
|------|--------|------|------|
| Kong PostgreSQL | kong-postgres | 40433 | Kong 配置数据库 |
| Kong Gateway | kong-gateway | 40080 (Proxy)<br>40081 (Admin) | API 网关 |
| Konga UI | konga-ui | 50001 | Kong 管理界面 |

**访问地址**:
- Kong Proxy (对外 API): http://localhost:40080
- Kong Admin API: http://localhost:40081
- Konga UI: http://localhost:50001

**配置 Kong**:
```bash
# 查看所有路由
curl http://localhost:40081/routes

# 查看所有服务
curl http://localhost:40081/services

# 测试 API
curl http://localhost:40080/api/v1/orders
```

### 监控系统 (3个)

| 组件 | 容器名 | 端口 | 用途 |
|------|--------|------|------|
| Prometheus | payment-prometheus | 40090 | 指标收集和存储 |
| Grafana | payment-grafana | 40300 | 可视化仪表板 |
| Jaeger | payment-jaeger | 50686 (UI)<br>50268 (Collector) | 分布式追踪 |

**访问地址**:
- Prometheus: http://localhost:40090
- Grafana: http://localhost:40300 (账号: admin / admin)
- Jaeger UI: http://localhost:50686

**Prometheus 指标端点**:
```bash
# 微服务指标
http://localhost:40001/metrics  # admin-service
http://localhost:40002/metrics  # merchant-service
http://localhost:40003/metrics  # payment-gateway
# ... (所有服务都有 /metrics 端点)

# 基础设施指标
http://localhost:40187/metrics  # PostgreSQL Exporter
http://localhost:40121/metrics  # Redis Exporter
http://localhost:40308/metrics  # Kafka Exporter
http://localhost:40180/metrics  # cAdvisor
http://localhost:40100/metrics  # Node Exporter
```

**Grafana 配置**:
1. 登录: http://localhost:40300 (admin/admin)
2. 添加 Prometheus 数据源: http://prometheus:9090
3. 导入预配置的仪表板 (在 backend/deployments/grafana/provisioning/)

**Jaeger 追踪**:
- 查看服务依赖图
- 搜索 Trace ID
- 分析服务调用链和性能瓶颈

### 监控导出器 (5个)

| 组件 | 容器名 | 端口 | 监控目标 |
|------|--------|------|----------|
| PostgreSQL Exporter | payment-postgres-exporter | 40187 | 数据库性能指标 |
| Redis Exporter | payment-redis-exporter | 40121 | Redis 性能指标 |
| Kafka Exporter | payment-kafka-exporter | 40308 | Kafka 性能指标 |
| cAdvisor | payment-cadvisor | 40180 | Docker 容器资源使用 |
| Node Exporter | payment-node-exporter | 40100 | 主机系统指标 |

**cAdvisor UI**: http://localhost:40180 (容器资源监控)

**监控指标示例**:
```promql
# PostgreSQL
pg_stat_database_tup_inserted{datname="payment_gateway"}

# Redis
redis_connected_clients

# Kafka
kafka_topic_partitions{topic="payment.events"}

# 容器 CPU 使用率
rate(container_cpu_usage_seconds_total[5m])

# 主机内存使用
node_memory_MemAvailable_bytes
```

## 服务启动顺序

管理脚本会自动处理依赖关系，但了解启动顺序有助于故障排查：

1. **Zookeeper** (Kafka 依赖)
2. **PostgreSQL** (Kong 和服务依赖)
3. **Redis** (服务依赖)
4. **Kafka** (依赖 Zookeeper)
5. **Kong Database** (Kong 依赖)
6. **Kong Bootstrap** (数据库初始化)
7. **Kong Gateway** (依赖 Kong Database)
8. **所有其他组件** (并行启动)

## 端口总览

### 数据存储
- 40432: PostgreSQL
- 40379: Redis

### 消息队列
- 42181: Zookeeper
- 40092: Kafka (内部 9092)
- 40093: Kafka (外部)
- 40084: Kafka UI

### API 网关
- 40433: Kong PostgreSQL
- 40080: Kong Proxy
- 40081: Kong Admin
- 50001: Konga UI

### 监控
- 40090: Prometheus
- 40300: Grafana
- 50686: Jaeger UI
- 50268: Jaeger Collector

### 导出器
- 40187: PostgreSQL Exporter
- 40121: Redis Exporter
- 40308: Kafka Exporter
- 40180: cAdvisor
- 40100: Node Exporter

## 数据持久化

所有数据都存储在 Docker volumes 中：

```bash
# 查看所有 volumes
docker volume ls | grep payment

# Volume 列表:
# - postgres_data        - PostgreSQL 数据
# - redis_data           - Redis 数据
# - kafka_data           - Kafka 消息
# - zookeeper_data       - Zookeeper 配置
# - prometheus_data      - Prometheus 时序数据
# - grafana_data         - Grafana 仪表板配置
# - kong_postgres_data   - Kong 配置
```

**备份数据**:
```bash
# PostgreSQL
docker exec payment-postgres pg_dumpall -U postgres > backup.sql

# Redis
docker exec payment-redis redis-cli SAVE
docker cp payment-redis:/data/dump.rdb backup.rdb
```

**清理所有数据** (危险操作):
```bash
# 停止所有容器
docker compose down

# 删除所有 volumes (会丢失所有数据!)
docker volume rm $(docker volume ls -q | grep payment)

# 重新启动将创建空数据库
docker compose up -d
```

## 常见问题

### 1. 容器无法启动

```bash
# 查看容器日志
docker logs payment-postgres
docker logs payment-kafka
docker logs kong-gateway

# 查看容器状态
docker ps -a | grep payment
```

### 2. Kafka 启动失败

**原因**: Zookeeper 未就绪

**解决**:
```bash
# 检查 Zookeeper
docker exec payment-zookeeper nc -z localhost 2181

# 重启 Kafka
docker compose restart kafka
```

### 3. Kong 无法连接后端服务

**原因**: 服务未使用 HTTPS 或 mTLS 证书配置问题

**解决**:
1. 确认后端服务使用 HTTPS (40001-40016)
2. 检查 Kong mTLS 配置 (docker-compose.yml)
3. 验证证书挂载: `docker exec kong-gateway ls /kong/certs`

### 4. 监控指标缺失

**原因**: Exporter 未运行或配置错误

**解决**:
```bash
# 检查 Exporter
docker ps | grep exporter

# 测试 Exporter 连接
curl http://localhost:40187/metrics  # PostgreSQL
curl http://localhost:40121/metrics  # Redis
curl http://localhost:40308/metrics  # Kafka
```

### 5. 磁盘空间不足

**原因**: 日志或数据累积

**解决**:
```bash
# 查看 Docker 磁盘使用
docker system df

# 清理未使用的容器、镜像、网络
docker system prune -a

# 清理日志
docker compose logs > /dev/null
```

## 健康检查

所有容器都配置了健康检查：

```bash
# 查看容器健康状态
docker ps --format "table {{.Names}}\t{{.Status}}"

# 只看不健康的容器
docker ps --filter health=unhealthy

# 查看健康检查详情
docker inspect --format='{{json .State.Health}}' payment-postgres | jq
```

## 性能调优

### PostgreSQL

```bash
# 进入容器
docker exec -it payment-postgres bash

# 查看连接数
psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# 查看慢查询
psql -U postgres -c "SELECT query, calls, total_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
```

### Redis

```bash
# Redis 信息
docker exec payment-redis redis-cli INFO

# 查看内存使用
docker exec payment-redis redis-cli INFO memory

# 查看慢日志
docker exec payment-redis redis-cli SLOWLOG GET 10
```

### Kafka

```bash
# 查看主题列表
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --list

# 查看消费者组
docker exec payment-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list

# 查看主题详情
docker exec payment-kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic payment.events
```

## 升级和维护

### 升级镜像

```bash
# 拉取最新镜像
docker compose pull

# 重新创建容器
docker compose up -d

# 查看镜像版本
docker images | grep -E "postgres|redis|kafka|kong|prometheus|grafana|jaeger"
```

### 定期维护

1. **每周**: 检查日志，清理不需要的数据
2. **每月**: 备份数据库，更新镜像
3. **每季度**: 审查监控告警，优化性能

## 总结

- **17 个基础设施组件** 全部使用 Docker 容器化
- **统一管理脚本** `manage-services.sh` 一键操作
- **完整监控栈** Prometheus + Grafana + Jaeger
- **数据持久化** Docker Volumes 保证数据安全
- **健康检查** 自动检测容器状态

使用 `./scripts/manage-services.sh infra status` 随时查看基础设施状态！

# 端口映射文档

本文档记录了所有服务的端口映射关系，所有对外端口都配置为 40000+ 以避免端口冲突。

## 端口映射规则

**外部端口格式**: `40XXX` (宿主机端口)
**内部端口格式**: `原始端口` (容器内端口)

---

## 基础设施服务

| 服务 | 内部端口 | 外部端口 | 访问地址 | 说明 |
|------|---------|---------|----------|------|
| PostgreSQL | 5432 | 40432 | localhost:40432 | 数据库 |
| Redis | 6379 | 40379 | localhost:40379 | 缓存 |
| Kafka | 9092 | 40092 | localhost:40092 | 消息队列 |
| Zookeeper | 2181 | - | 内部使用 | Kafka 依赖 |

### 连接字符串示例

```bash
# PostgreSQL
postgresql://postgres:postgres@localhost:40432/payment_platform

# Redis
redis://localhost:40379

# Kafka
localhost:40092
```

---

## 微服务端口

| 服务名称 | 内部端口 | 外部端口 | 访问地址 | 健康检查 |
|---------|---------|---------|----------|---------|
| Admin Service | 8000 | 40000 | http://localhost:40000 | http://localhost:40000/health |
| Merchant Service | 8001 | 40001 | http://localhost:40001 | http://localhost:40001/health |
| Payment Gateway | 8002 | 40002 | http://localhost:40002 | http://localhost:40002/health |
| Channel Adapter | 8003 | 40003 | http://localhost:40003 | http://localhost:40003/health |
| Order Service | 8004 | 40004 | http://localhost:40004 | http://localhost:40004/health |
| Accounting Service | 8005 | 40005 | http://localhost:40005 | http://localhost:40005/health |
| Risk Service | 8006 | 40006 | http://localhost:40006 | http://localhost:40006/health |
| Notification Service | 8007 | 40007 | http://localhost:40007 | http://localhost:40007/health |
| Analytics Service | 8008 | 40008 | http://localhost:40008 | http://localhost:40008/health |
| Config Service | 8009 | 40009 | http://localhost:40009 | http://localhost:40009/health |

### API 端点示例

```bash
# 创建账户
curl -X POST http://localhost:40005/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"merchant_id": "...", "account_type": "operating", "currency": "CNY"}'

# 风险检查
curl -X POST http://localhost:40006/api/v1/checks/payment \
  -H "Content-Type: application/json" \
  -d '{"merchant_id": "...", "amount": 10000, "currency": "CNY"}'

# 获取分析数据
curl "http://localhost:40008/api/v1/analytics/payments/summary?merchant_id=...&start_date=2024-01-01"
```

---

## 监控与管理工具

| 工具 | 内部端口 | 外部端口 | 访问地址 | 说明 |
|------|---------|---------|----------|------|
| API Gateway | 80 | 40080 | http://localhost:40080 | Traefik 入口 |
| Traefik Dashboard | 8080 | 40081 | http://localhost:40081 | Traefik 控制台 |
| Prometheus | 9090 | 40090 | http://localhost:40090 | 指标收集 |
| Grafana | 3000 | 40300 | http://localhost:40300 | 数据可视化 |
| Jaeger UI | 16686 | 40686 | http://localhost:40686 | 分布式追踪 |

### Jaeger 完整端口映射

| 功能 | 内部端口 | 外部端口 | 协议 | 说明 |
|------|---------|---------|------|------|
| Compact Thrift | 5775 | 40775 | UDP | Agent 接收 |
| Binary Thrift | 6831 | 40831 | UDP | Agent 接收 |
| Binary Thrift | 6832 | 40832 | UDP | Agent 接收 |
| HTTP Config | 5778 | 40778 | TCP | Agent 配置 |
| UI/Query | 16686 | 40686 | TCP | Web UI |
| HTTP Collector | 14268 | 40268 | TCP | Collector |
| gRPC Collector | 14250 | 40250 | TCP | Collector |
| Zipkin | 9411 | 40411 | TCP | Zipkin 兼容 |

---

## Docker 网络配置

```yaml
networks:
  payment-network:
    driver: bridge
    ipam:
      driver: default
      config:
        - subnet: 172.28.0.0/16
          gateway: 172.28.0.1
    driver_opts:
      com.docker.network.bridge.name: br-payment
```

**网络名称**: `payment_payment-network` (Docker Compose 会自动加前缀)
**子网**: `172.28.0.0/16`
**网关**: `172.28.0.1`
**网桥名称**: `br-payment`

### 查看网络信息

```bash
# 查看网络详情
docker network inspect payment_payment-network

# 查看网络列表
docker network ls | grep payment

# 查看容器 IP
docker inspect <container_name> | grep IPAddress
```

---

## 快速启动

### 1. 启动所有服务

```bash
docker-compose up -d
```

### 2. 查看服务状态

```bash
docker-compose ps
```

### 3. 健康检查

```bash
# 检查所有服务
for port in 40000 40001 40002 40003 40004 40005 40006 40007 40008 40009; do
  echo "Checking port $port..."
  curl -s http://localhost:$port/health
done
```

### 4. 查看日志

```bash
# 查看特定服务日志
docker-compose logs -f accounting-service

# 查看所有服务日志
docker-compose logs -f
```

---

## 本地开发配置

### 环境变量 (.env)

```env
# 数据库配置
DATABASE_URL=postgres://postgres:postgres@localhost:40432/payment_platform?sslmode=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=40379

# Kafka 配置
KAFKA_BROKERS=localhost:40092

# 服务端口
ADMIN_PORT=40000
MERCHANT_PORT=40001
PAYMENT_GATEWAY_PORT=40002
CHANNEL_ADAPTER_PORT=40003
ORDER_SERVICE_PORT=40004
ACCOUNTING_SERVICE_PORT=40005
RISK_SERVICE_PORT=40006
NOTIFICATION_SERVICE_PORT=40007
ANALYTICS_SERVICE_PORT=40008
CONFIG_SERVICE_PORT=40009
```

### 连接数据库

```bash
# 使用 psql
psql -h localhost -p 40432 -U postgres -d payment_platform

# 使用 Docker 执行
docker exec -it payment-postgres psql -U postgres -d payment_platform
```

### 连接 Redis

```bash
# 使用 redis-cli
redis-cli -h localhost -p 40379

# 使用 Docker 执行
docker exec -it payment-redis redis-cli
```

---

## 防火墙配置

如果需要从外部访问，请开放以下端口：

```bash
# 微服务端口
sudo ufw allow 40000:40009/tcp

# 基础设施
sudo ufw allow 40432/tcp  # PostgreSQL
sudo ufw allow 40379/tcp  # Redis
sudo ufw allow 40092/tcp  # Kafka

# 监控工具
sudo ufw allow 40080/tcp  # API Gateway
sudo ufw allow 40081/tcp  # Traefik Dashboard
sudo ufw allow 40090/tcp  # Prometheus
sudo ufw allow 40300/tcp  # Grafana
sudo ufw allow 40686/tcp  # Jaeger UI
```

---

## 端口冲突检查

### 检查端口占用

```bash
# 检查端口是否被占用
lsof -i :40000
lsof -i :40432

# 批量检查所有端口
for port in {40000..40009} 40432 40379 40092 40080 40081 40090 40300 40686; do
  if lsof -i :$port > /dev/null 2>&1; then
    echo "Port $port is in use"
  fi
done
```

### 清理占用的端口

```bash
# 杀死占用端口的进程
lsof -ti:40000 | xargs kill -9

# 停止所有 Docker 容器
docker-compose down
```

---

## 故障排查

### 问题：无法访问服务

1. 检查容器是否运行
   ```bash
   docker-compose ps
   ```

2. 检查端口映射
   ```bash
   docker port <container_name>
   ```

3. 检查防火墙
   ```bash
   sudo ufw status
   ```

4. 检查日志
   ```bash
   docker-compose logs <service_name>
   ```

### 问题：数据库连接失败

1. 确认 PostgreSQL 容器运行
   ```bash
   docker-compose ps postgres
   ```

2. 测试连接
   ```bash
   nc -zv localhost 40432
   ```

3. 检查数据库日志
   ```bash
   docker-compose logs postgres
   ```

### 问题：服务间通信失败

1. 检查网络
   ```bash
   docker network inspect payment_payment-network
   ```

2. 测试容器间连接
   ```bash
   docker exec <container1> ping <container2>
   ```

3. 确认服务发现
   ```bash
   docker exec payment-admin-service nslookup postgres
   ```

---

## 性能优化建议

1. **端口范围**: 使用 40000+ 端口避免与系统常用端口冲突
2. **网络隔离**: 使用独立的 Docker 网络提高安全性
3. **健康检查**: 定期检查服务健康状态
4. **日志管理**: 配置日志轮转避免磁盘占满
5. **资源限制**: 在生产环境配置容器资源限制

---

## 生产环境注意事项

⚠️ **重要提示**：

1. **不要在生产环境暴露所有端口** - 只暴露 API Gateway 端口
2. **使用 HTTPS** - 配置 SSL/TLS 证书
3. **配置防火墙** - 限制访问来源
4. **使用环境变量** - 不要在配置文件中硬编码敏感信息
5. **启用认证** - 所有服务都应该配置认证
6. **监控告警** - 配置 Prometheus 告警规则
7. **备份数据** - 定期备份数据库和重要数据

---

## 参考链接

- [Docker Compose 网络文档](https://docs.docker.com/compose/networking/)
- [端口映射最佳实践](https://docs.docker.com/config/containers/container-networking/)
- [Linux 端口范围说明](https://www.iana.org/assignments/service-names-port-numbers/)

---

**更新时间**: 2025-10-23
**版本**: 1.0
**维护者**: Payment Platform Team

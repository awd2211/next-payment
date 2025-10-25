# 运维指南 (Operations Guide)

## 目录

1. [快速启动](#快速启动)
2. [系统监控](#系统监控)
3. [日志管理](#日志管理)
4. [故障排查](#故障排查)
5. [性能优化](#性能优化)
6. [备份恢复](#备份恢复)
7. [扩容指南](#扩容指南)
8. [安全加固](#安全加固)

---

## 快速启动

### 一键部署

```bash
# 从项目根目录执行
chmod +x deploy.sh
./deploy.sh
```

部署脚本会自动完成:
- ✅ 环境检查 (Docker, Go, Node.js)
- ✅ 启动基础设施 (PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger)
- ✅ 初始化数据库 (19个数据库)
- ✅ 编译后端服务 (19个微服务)
- ✅ 启动后端服务
- ✅ 构建前端应用 (Admin Portal, Merchant Portal, Website)
- ✅ 健康检查
- ✅ 显示访问信息

**预计耗时**: 5-10分钟 (取决于网络速度和硬件性能)

### 手动启动步骤

**1. 启动基础设施**
```bash
docker-compose up -d
```

**2. 初始化数据库**
```bash
cd backend
./scripts/init-db.sh
```

**3. 启动后端服务**
```bash
cd backend
./scripts/start-all-services.sh

# 查看服务状态
./scripts/status-all-services.sh
```

**4. 启动前端应用**
```bash
# Terminal 1: Admin Portal
cd frontend/admin-portal
npm install
npm run dev  # http://localhost:5173

# Terminal 2: Merchant Portal
cd frontend/merchant-portal
npm install
npm run dev  # http://localhost:5174

# Terminal 3: Website (可选)
cd frontend/website
npm install
npm run dev  # http://localhost:5175
```

---

## 系统监控

### 系统状态仪表板

查看完整的系统健康状态:

```bash
cd backend
./scripts/system-status-dashboard.sh
```

**显示内容**:
- 基础设施状态 (PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger)
- 19个后端服务状态及健康检查
- 3个前端应用状态
- 19个数据库状态
- 系统资源使用 (CPU, 内存, 磁盘)
- 快速访问链接

### 服务依赖关系图

查看微服务间的依赖关系:

```bash
cd backend
./scripts/service-dependency-map.sh
```

**显示内容**:
- 基础设施依赖层
- 核心支付流程 (Critical Path)
- 管理平台依赖
- 商户平台依赖
- 财务流程
- 风控与合规流程
- 支撑服务
- 8层服务架构总结

### Grafana 监控仪表板

访问 Grafana: http://localhost:40300
- **用户名**: admin
- **密码**: admin

**可用仪表板**:
- Payment Gateway Metrics - 支付网关核心指标
- Service Performance - 服务性能监控
- Infrastructure - 基础设施监控 (PostgreSQL, Redis, Kafka)
- Business Metrics - 业务指标 (支付成功率, 交易量, GMV)

**关键指标**:
```promql
# 支付成功率 (过去5分钟)
sum(rate(payment_gateway_payment_total{status="success"}[5m]))
/ sum(rate(payment_gateway_payment_total[5m]))

# P95支付延迟
histogram_quantile(0.95,
  rate(payment_gateway_payment_duration_seconds_bucket[5m])
)

# 每秒请求数
sum(rate(http_requests_total[1m])) by (service)

# 错误率
sum(rate(http_requests_total{status=~"5.."}[5m]))
/ sum(rate(http_requests_total[5m]))
```

### Jaeger 分布式追踪

访问 Jaeger UI: http://localhost:40686

**使用场景**:
- 查看完整的支付请求流程
- 识别性能瓶颈 (慢查询, 慢接口)
- 排查跨服务调用问题
- 分析服务依赖关系

**搜索方式**:
- 按服务名: `payment-gateway`, `order-service`
- 按操作: `CreatePayment`, `CheckRisk`, `CreateOrder`
- 按标签: `merchant_id`, `order_no`, `payment_no`
- 按时长: 查找超过1秒的请求

---

## 日志管理

### 后端服务日志

所有服务日志存储在 `backend/logs/` 目录:

```bash
# 查看支付网关日志
tail -f backend/logs/payment-gateway.log

# 查看最近100行
tail -n 100 backend/logs/order-service.log

# 搜索错误日志
grep "ERROR" backend/logs/*.log

# 搜索特定订单
grep "ORDER-12345" backend/logs/payment-gateway.log

# 查看所有服务的最新日志
tail -f backend/logs/*.log
```

### Docker容器日志

查看基础设施日志:

```bash
# PostgreSQL
docker logs payment-postgres
docker logs -f payment-postgres  # 实时跟踪

# Redis
docker logs payment-redis

# Kafka
docker logs payment-kafka

# 查看所有容器日志
docker-compose logs -f
```

### 日志格式

所有服务使用结构化日志 (Zap):

```json
{
  "level": "info",
  "ts": "2025-10-25T10:30:45.123Z",
  "caller": "service/payment.go:123",
  "msg": "Payment created successfully",
  "payment_no": "PAY-202501251030-ABC123",
  "merchant_id": "merchant-001",
  "amount": 10000,
  "currency": "USD",
  "trace_id": "abc123def456"
}
```

### 日志级别

- **DEBUG**: 详细调试信息 (开发环境)
- **INFO**: 一般信息 (默认)
- **WARN**: 警告信息 (需要关注但不影响服务)
- **ERROR**: 错误信息 (需要立即处理)
- **FATAL**: 致命错误 (服务停止)

**生产环境建议**: 使用 `INFO` 级别，避免过多日志影响性能

---

## 故障排查

### 常见问题及解决方案

#### 1. 服务无法启动

**症状**: 运行 `start-all-services.sh` 后服务未启动

**排查步骤**:
```bash
# 检查端口占用
lsof -i :40003

# 查看错误日志
cat backend/logs/payment-gateway.log | grep ERROR

# 检查数据库连接
psql -h localhost -p 40432 -U postgres -d payment_gateway -c "SELECT 1"

# 检查Redis连接
redis-cli -h localhost -p 40379 ping
```

**常见原因**:
- 端口被占用 → 停止占用进程或修改端口配置
- 数据库连接失败 → 检查 `docker-compose` 是否启动
- 依赖服务未就绪 → 等待基础设施启动完成 (30-60秒)

#### 2. 支付创建失败

**症状**: 调用 `/api/v1/payments` 返回500错误

**排查步骤**:
```bash
# 1. 查看payment-gateway日志
tail -n 200 backend/logs/payment-gateway.log | grep ERROR

# 2. 检查依赖服务状态
curl http://localhost:40004/health  # order-service
curl http://localhost:40005/health  # channel-adapter
curl http://localhost:40006/health  # risk-service

# 3. 检查数据库
psql -h localhost -p 40432 -U postgres -d payment_gateway \
  -c "SELECT * FROM payments ORDER BY created_at DESC LIMIT 5;"

# 4. 查看Jaeger追踪
# 访问 http://localhost:40686 搜索最近的 CreatePayment 操作
```

**常见原因**:
- Order Service不可用 → 重启服务
- Channel Adapter配置错误 → 检查Stripe API密钥
- Risk Service拒绝 → 查看风控规则配置
- 数据库写入失败 → 检查数据库连接和表结构

#### 3. 前端无法连接后端

**症状**: Admin Portal显示网络错误

**排查步骤**:
```bash
# 1. 检查后端服务状态
curl http://localhost:40001/health

# 2. 检查CORS配置
curl -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Authorization" \
  -X OPTIONS \
  http://localhost:40001/api/v1/merchants

# 3. 检查JWT Token (使用浏览器开发者工具)
# - 查看 Network 标签
# - 检查 Authorization header
# - 验证 token 是否过期
```

**常见原因**:
- 后端服务未启动 → 运行 `start-all-services.sh`
- CORS配置错误 → 检查 `middleware.CORSMiddleware()` 配置
- Token过期 → 重新登录获取新token
- 端口配置错误 → 检查前端 `services/api.ts` 的 baseURL

#### 4. 数据库连接池耗尽

**症状**: 日志显示 "too many connections"

**排查步骤**:
```bash
# 查看当前连接数
psql -h localhost -p 40432 -U postgres -c \
  "SELECT count(*) FROM pg_stat_activity;"

# 查看每个数据库的连接数
psql -h localhost -p 40432 -U postgres -c \
  "SELECT datname, count(*) FROM pg_stat_activity GROUP BY datname;"

# 终止空闲连接
psql -h localhost -p 40432 -U postgres -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity
   WHERE state = 'idle' AND state_change < now() - interval '5 minutes';"
```

**解决方案**:
- 增加PostgreSQL最大连接数 (修改 `docker-compose.yml`)
- 减少服务的连接池大小 (修改 `db.Config`)
- 修复连接泄漏 (确保正确关闭数据库连接)

#### 5. Redis内存不足

**症状**: 日志显示 "OOM command not allowed"

**排查步骤**:
```bash
# 查看Redis内存使用
redis-cli -h localhost -p 40379 INFO memory

# 查看Redis键数量
redis-cli -h localhost -p 40379 DBSIZE

# 查看占用内存最大的键
redis-cli -h localhost -p 40379 --bigkeys
```

**解决方案**:
- 增加Redis内存限制 (修改 `docker-compose.yml`)
- 设置过期时间 (所有缓存键应设置TTL)
- 清理不必要的键

---

## 性能优化

### 数据库优化

#### 1. 添加索引

```sql
-- 查看慢查询
SELECT * FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- 为常用查询添加索引
CREATE INDEX idx_payments_merchant_created
  ON payments(merchant_id, created_at DESC);

CREATE INDEX idx_orders_order_no
  ON orders(order_no);

CREATE INDEX idx_payments_payment_no
  ON payments(payment_no);
```

#### 2. 连接池配置

修改 `pkg/db/postgres.go`:

```go
// 生产环境建议配置
MaxOpenConns: 100,          // 最大打开连接数
MaxIdleConns: 10,           // 最大空闲连接数
ConnMaxLifetime: time.Hour, // 连接最大生命周期
ConnMaxIdleTime: 10 * time.Minute, // 空闲连接最大生命周期
```

#### 3. 查询优化

```go
// ❌ 错误: N+1查询
for _, payment := range payments {
    order, _ := s.orderRepo.GetByOrderNo(payment.OrderNo)
}

// ✅ 正确: 预加载关联数据
db.Preload("Orders").Find(&payments)
```

### Redis优化

#### 1. 设置合理的过期时间

```go
// 幂等性键: 24小时
redis.Set(ctx, idempotencyKey, paymentNo, 24*time.Hour)

// 会话缓存: 1小时
redis.Set(ctx, sessionKey, userData, 1*time.Hour)

// 验证码: 5分钟
redis.Set(ctx, verifyCodeKey, code, 5*time.Minute)
```

#### 2. 使用Pipeline批量操作

```go
pipe := redis.Pipeline()
for _, key := range keys {
    pipe.Get(ctx, key)
}
results, _ := pipe.Exec(ctx)
```

### 服务优化

#### 1. 启用HTTP/2

```go
// 在 main.go 中启用HTTP/2
srv := &http.Server{
    Addr:    ":40003",
    Handler: router,
}
srv.ListenAndServeTLS("cert.pem", "key.pem")
```

#### 2. 添加响应缓存

```go
// 缓存不常变化的数据
func (s *Service) GetMerchant(ctx context.Context, id string) (*model.Merchant, error) {
    // 先查缓存
    cacheKey := fmt.Sprintf("merchant:%s", id)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached.(*model.Merchant), nil
    }

    // 查数据库
    merchant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 写入缓存
    s.cache.Set(ctx, cacheKey, merchant, 5*time.Minute)
    return merchant, nil
}
```

#### 3. 限流优化

调整限流阈值 (根据实际负载):

```go
// pkg/middleware/ratelimit.go
RateLimitRequests: 1000,          // 每分钟1000请求 (原100)
RateLimitWindow:   time.Minute,
```

---

## 备份恢复

### 数据库备份

#### 自动备份脚本

创建 `backend/scripts/backup-db.sh`:

```bash
#!/bin/bash

BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份所有payment_*数据库
for db in $(psql -h localhost -p 40432 -U postgres -t -c \
  "SELECT datname FROM pg_database WHERE datname LIKE 'payment_%'"); do
    echo "Backing up $db..."
    pg_dump -h localhost -p 40432 -U postgres -d $db \
      > "$BACKUP_DIR/${db}_${TIMESTAMP}.sql"
done

# 压缩备份
tar -czf "$BACKUP_DIR/backup_${TIMESTAMP}.tar.gz" $BACKUP_DIR/*.sql
rm $BACKUP_DIR/*.sql

# 删除7天前的备份
find $BACKUP_DIR -name "backup_*.tar.gz" -mtime +7 -delete

echo "Backup completed: backup_${TIMESTAMP}.tar.gz"
```

#### 设置定时任务

```bash
# 每天凌晨2点自动备份
crontab -e

# 添加以下行
0 2 * * * /home/eric/payment/backend/scripts/backup-db.sh
```

### 数据恢复

```bash
# 1. 解压备份
tar -xzf backup_20250125_020000.tar.gz

# 2. 恢复数据库
psql -h localhost -p 40432 -U postgres -d payment_gateway \
  < payment_gateway_20250125_020000.sql

# 3. 验证数据
psql -h localhost -p 40432 -U postgres -d payment_gateway \
  -c "SELECT count(*) FROM payments;"
```

### Redis备份

```bash
# 手动触发RDB快照
redis-cli -h localhost -p 40379 SAVE

# 复制RDB文件
docker cp payment-redis:/data/dump.rdb ./backup/redis_$(date +%Y%m%d).rdb
```

---

## 扩容指南

### 水平扩容 (Scale Out)

#### 1. 无状态服务扩容

大部分微服务都是无状态的,可以直接启动多个实例:

```bash
# 启动第二个payment-gateway实例 (不同端口)
PORT=40103 \
GRPC_PORT=50103 \
DB_NAME=payment_gateway \
./bin/payment-gateway &

# 使用Nginx/Kong做负载均衡
upstream payment_gateway {
    server localhost:40003;
    server localhost:40103;
}
```

#### 2. 数据库读写分离

配置PostgreSQL主从复制:

```yaml
# docker-compose.yml
services:
  postgres-master:
    image: postgres:15
    ports:
      - "40432:5432"

  postgres-slave:
    image: postgres:15
    environment:
      POSTGRES_MASTER_HOST: postgres-master
    ports:
      - "40433:5432"
```

修改服务配置:

```go
// 写操作使用master
masterDB := db.Connect(db.Config{Host: "localhost", Port: 40432})

// 读操作使用slave
slaveDB := db.Connect(db.Config{Host: "localhost", Port: 40433})
```

#### 3. Redis Cluster

从单节点升级到Redis Cluster:

```bash
# 创建6节点集群 (3主3从)
docker-compose -f docker-compose-redis-cluster.yml up -d

# 创建集群
redis-cli --cluster create \
  localhost:7001 localhost:7002 localhost:7003 \
  localhost:7004 localhost:7005 localhost:7006 \
  --cluster-replicas 1
```

### 垂直扩容 (Scale Up)

#### 增加资源限制

修改 `docker-compose.yml`:

```yaml
services:
  postgres:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
```

#### PostgreSQL性能调优

```sql
-- 修改postgresql.conf
shared_buffers = 2GB              -- 25% of RAM
effective_cache_size = 6GB        -- 75% of RAM
maintenance_work_mem = 512MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1            -- For SSD
effective_io_concurrency = 200    -- For SSD
work_mem = 10MB
min_wal_size = 1GB
max_wal_size = 4GB
max_connections = 200
```

---

## 安全加固

### 1. 网络隔离

使用Docker网络隔离:

```yaml
# docker-compose.yml
networks:
  backend:
    driver: bridge
  frontend:
    driver: bridge

services:
  postgres:
    networks:
      - backend  # 仅backend网络可访问

  payment-gateway:
    networks:
      - backend
      - frontend  # 可被frontend访问
```

### 2. 密钥管理

使用环境变量或密钥管理服务:

```bash
# 生产环境使用.env文件
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
STRIPE_API_KEY=sk_live_xxx
DB_PASSWORD=$(openssl rand -base64 24)
EOF

# 设置文件权限
chmod 600 .env
```

### 3. API限流

```go
// 按IP限流
middleware.RateLimitByIP(100, time.Minute)

// 按用户限流
middleware.RateLimitByUser(1000, time.Minute)

// 按API Key限流
middleware.RateLimitByAPIKey(10000, time.Minute)
```

### 4. SQL注入防护

```go
// ✅ 正确: 使用参数化查询
db.Where("merchant_id = ? AND status = ?", merchantID, status).Find(&payments)

// ❌ 错误: 字符串拼接
db.Where(fmt.Sprintf("merchant_id = '%s'", merchantID)).Find(&payments)
```

### 5. 日志脱敏

```go
// 脱敏敏感信息
logger.Info("Payment created",
    zap.String("payment_no", paymentNo),
    zap.String("card_last4", maskCard(cardNo)), // 仅显示后4位
    zap.Int64("amount", amount),
)

func maskCard(cardNo string) string {
    if len(cardNo) < 4 {
        return "****"
    }
    return "****" + cardNo[len(cardNo)-4:]
}
```

---

## 附录

### 快速命令参考

```bash
# 启动系统
./deploy.sh

# 查看状态
cd backend && ./scripts/system-status-dashboard.sh

# 查看依赖
cd backend && ./scripts/service-dependency-map.sh

# 重启服务
cd backend && ./scripts/stop-all-services.sh && ./scripts/start-all-services.sh

# 查看日志
tail -f backend/logs/payment-gateway.log

# 备份数据库
./backend/scripts/backup-db.sh

# 健康检查
curl http://localhost:40003/health
```

### 监控告警阈值建议

| 指标 | 警告阈值 | 严重阈值 |
|------|---------|---------|
| CPU使用率 | 70% | 90% |
| 内存使用率 | 80% | 95% |
| 磁盘使用率 | 75% | 90% |
| 支付成功率 | <95% | <90% |
| API响应时间 (P95) | >1s | >3s |
| 错误率 | >1% | >5% |
| 数据库连接数 | >80 | >95 |

### 联系方式

- **技术支持**: support@payment-platform.com
- **紧急事件**: oncall@payment-platform.com
- **文档**: https://docs.payment-platform.com

---

**最后更新**: 2025-10-25
**版本**: 1.0.0

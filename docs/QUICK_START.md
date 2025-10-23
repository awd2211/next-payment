# 快速开始指南

## 前置要求

- Docker & Docker Compose
- Go 1.21+
- PostgreSQL 15+
- Redis 7+

## 本地开发环境

### 1. 克隆项目

```bash
cd /home/eric/payment
```

### 2. 配置环境变量

创建 `.env` 文件：

```bash
# 数据库配置
DATABASE_URL=postgres://postgres:postgres@localhost:5432/payment_platform?sslmode=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production

# Stripe 配置
STRIPE_API_KEY=sk_test_your_stripe_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
STRIPE_PUBLISHABLE_KEY=pk_test_your_publishable_key

# 邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASSWORD=your-password

# Kafka 配置
KAFKA_BROKERS=localhost:9092
```

### 3. 启动基础设施

使用 Docker Compose 启动 PostgreSQL、Redis、Kafka：

```bash
docker-compose up -d postgres redis kafka zookeeper
```

等待服务启动（约30秒）：

```bash
# 检查服务状态
docker-compose ps

# 查看日志
docker-compose logs -f postgres redis kafka
```

### 4. 初始化数据库

```bash
# 连接数据库
psql -h localhost -U postgres -d payment_platform

# 执行迁移脚本
\i backend/services/payment-gateway/migrations/001_create_payment_tables.sql
\i backend/services/order-service/migrations/001_create_order_tables.sql
\i backend/services/channel-adapter/migrations/001_create_channel_tables.sql

# 退出
\q
```

或者使用命令行：

```bash
psql -h localhost -U postgres -d payment_platform \
  -f backend/services/payment-gateway/migrations/001_create_payment_tables.sql

psql -h localhost -U postgres -d payment_platform \
  -f backend/services/order-service/migrations/001_create_order_tables.sql

psql -h localhost -U postgres -d payment_platform \
  -f backend/services/channel-adapter/migrations/001_create_channel_tables.sql
```

### 5. 启动服务

#### 方式一：使用 Go 直接运行

在不同终端窗口中运行：

```bash
# 终端1 - Payment Gateway
cd backend/services/payment-gateway
go run cmd/main.go

# 终端2 - Order Service
cd backend/services/order-service
go run cmd/main.go

# 终端3 - Channel Adapter
cd backend/services/channel-adapter
go run cmd/main.go

# 终端4 - Merchant Service
cd backend/services/merchant-service
go run cmd/main.go

# 终端5 - Admin Service
cd backend/services/admin-service
go run cmd/main.go
```

#### 方式二：使用 Docker Compose

```bash
docker-compose up -d
```

### 6. 验证服务

检查服务健康状态：

```bash
# Payment Gateway
curl http://localhost:8002/health

# Order Service
curl http://localhost:8004/health

# Channel Adapter
curl http://localhost:8003/health

# Merchant Service
curl http://localhost:8001/health

# Admin Service
curl http://localhost:8000/health
```

## 测试 API

### 1. 创建支付

```bash
curl -X POST http://localhost:8002/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "00000000-0000-0000-0000-000000000001",
    "order_no": "ORD20240115001",
    "amount": 10000,
    "currency": "USD",
    "customer_email": "test@example.com",
    "customer_name": "John Doe",
    "customer_phone": "+1234567890",
    "customer_ip": "192.168.1.100",
    "description": "测试支付",
    "notify_url": "https://example.com/webhook",
    "return_url": "https://example.com/success",
    "expire_minutes": 30
  }'
```

**成功响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "payment_no": "PY20240115120000abcd",
    "order_no": "ORD20240115001",
    "amount": 10000,
    "currency": "USD",
    "status": "pending",
    "channel": "stripe",
    "created_at": "2024-01-15T12:00:00Z"
  }
}
```

### 2. 查询支付

```bash
curl http://localhost:8002/api/v1/payments/PY20240115120000abcd
```

### 3. 创建订单

```bash
curl -X POST http://localhost:8004/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "00000000-0000-0000-0000-000000000001",
    "customer_email": "test@example.com",
    "customer_name": "John Doe",
    "currency": "USD",
    "items": [
      {
        "product_id": "prod_123",
        "product_name": "商品A",
        "unit_price": 5000,
        "quantity": 2
      }
    ],
    "shipping_fee": 500,
    "discount_amount": 0
  }'
```

### 4. 查询订单

```bash
curl http://localhost:8004/api/v1/orders/OD20240115...
```

### 5. 创建退款

```bash
curl -X POST http://localhost:8002/api/v1/refunds \
  -H "Content-Type: application/json" \
  -d '{
    "payment_no": "PY20240115120000abcd",
    "amount": 5000,
    "reason": "商品质量问题",
    "description": "部分退款",
    "operator_type": "merchant"
  }'
```

## 访问管理界面

### Traefik Dashboard
```
http://localhost:8080
```

### Grafana
```
http://localhost:3000
用户名: admin
密码: admin
```

### Prometheus
```
http://localhost:9090
```

### Jaeger UI
```
http://localhost:16686
```

## 常见问题

### 1. 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
docker-compose ps postgres

# 查看日志
docker-compose logs postgres

# 重启服务
docker-compose restart postgres
```

### 2. 端口冲突

如果端口被占用，修改 `docker-compose.yml` 中的端口映射：

```yaml
ports:
  - "8002:8002"  # 改为 "8012:8002"
```

### 3. Stripe 配置问题

确保在 `.env` 文件中配置了正确的 Stripe 密钥：

```bash
# 获取测试密钥
# https://dashboard.stripe.com/test/apikeys

STRIPE_API_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
```

### 4. 清理和重启

```bash
# 停止所有服务
docker-compose down

# 清理数据（谨慎！会删除所有数据）
docker-compose down -v

# 重新启动
docker-compose up -d
```

## 开发流程

### 1. 添加新功能

```bash
# 创建功能分支
git checkout -b feature/new-feature

# 开发代码...

# 运行测试
go test ./...

# 提交代码
git add .
git commit -m "feat: 添加新功能"
git push origin feature/new-feature
```

### 2. 数据库迁移

在对应服务的 `migrations/` 目录下创建新的 SQL 文件：

```sql
-- 002_add_new_table.sql
CREATE TABLE new_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 3. 添加新的 API 端点

1. 在 `internal/model/` 定义数据模型
2. 在 `internal/repository/` 实现数据库操作
3. 在 `internal/service/` 实现业务逻辑
4. 在 `internal/handler/` 添加 HTTP 处理器
5. 在 `cmd/main.go` 注册路由

## 性能测试

使用 Apache Bench 进行简单压测：

```bash
# 安装 ab
sudo apt-get install apache2-utils  # Ubuntu
brew install apache2-utils           # macOS

# 创建支付压测
ab -n 1000 -c 10 -p payment.json -T application/json \
  http://localhost:8002/api/v1/payments

# 查询支付压测
ab -n 10000 -c 100 \
  http://localhost:8002/api/v1/payments/PY20240115120000abcd
```

## 日志查看

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f payment-gateway

# 查看最近100行日志
docker-compose logs --tail=100 payment-gateway
```

## 停止服务

```bash
# 停止所有服务
docker-compose stop

# 停止并删除容器
docker-compose down

# 停止并删除容器和数据卷
docker-compose down -v
```

## 下一步

- 查看 [API 文档](./API.md)
- 阅读 [架构文档](./ARCHITECTURE.md)
- 了解 [开发指南](./DEVELOPMENT.md)
- 参考 [项目进度](./PROJECT_PROGRESS.md)

## 获取帮助

- GitHub Issues: https://github.com/your-repo/issues
- 文档: https://docs.payment-platform.com
- Email: support@payment-platform.com

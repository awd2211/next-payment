# 🚀 本地开发环境配置指南

本文档说明如何在本地开发环境启动整个支付平台系统。

## 📋 目录

- [系统要求](#系统要求)
- [环境配置](#环境配置)
- [启动后端服务](#启动后端服务)
- [启动前端应用](#启动前端应用)
- [访问地址](#访问地址)
- [常见问题](#常见问题)

---

## 系统要求

- **Go**: 1.21+
- **Node.js**: 18+
- **PostgreSQL**: 15+
- **Redis**: 7+
- **Docker**: 20+ (用于运行 PostgreSQL 和 Redis)

---

## 环境配置

### 1. 启动基础设施

```bash
# 启动 PostgreSQL 和 Redis (如果使用 Docker)
docker ps | grep payment-postgres  # 检查是否已启动
docker ps | grep redis
```

### 2. 验证数据库

所有数据库应该已创建：

```bash
docker exec payment-postgres psql -U postgres -c "\l" | grep payment_
```

应该看到：
- payment_admin
- payment_merchant
- payment_gateway
- payment_order
- payment_channel
- payment_risk
- payment_accounting
- payment_notification
- payment_analytics
- payment_config

---

## 启动后端服务

### 方式 1: 手动启动每个服务

```bash
# 进入后端目录
cd /home/eric/payment/backend

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSL_MODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379

# 启动 Admin Service (8001)
cd services/admin-service
export DB_NAME=payment_admin PORT=8001
go run cmd/main.go &

# 启动 Merchant Service (8002)
cd ../merchant-service
export DB_NAME=payment_merchant PORT=8002
go run cmd/main.go &

# 启动 Payment Gateway (8003)
cd ../payment-gateway
export DB_NAME=payment_gateway PORT=8003
go run cmd/main.go &

# 启动 Order Service (8004)
cd ../order-service
export DB_NAME=payment_order PORT=8004
go run cmd/main.go &

# 启动 Channel Adapter (8005)
cd ../channel-adapter
export DB_NAME=payment_channel PORT=8005
go run cmd/main.go &

# 启动 Risk Service (8006)
cd ../risk-service
export DB_NAME=payment_risk PORT=8006
go run cmd/main.go &

# 启动 Accounting Service (8007)
cd ../accounting-service
export DB_NAME=payment_accounting PORT=8007
go run cmd/main.go &

# 启动 Notification Service (8008)
cd ../notification-service
export DB_NAME=payment_notification PORT=8008
go run cmd/main.go &

# 启动 Analytics Service (8009)
cd ../analytics-service
export DB_NAME=payment_analytics PORT=8009
go run cmd/main.go &

# 启动 Config Service (8010)
cd ../config-service
export DB_NAME=payment_config PORT=8010
go run cmd/main.go &
```

### 方式 2: 使用启动脚本 (推荐)

创建一个启动脚本：

```bash
# 创建启动脚本
cat > /home/eric/payment/backend/start-local.sh << 'EOF'
#!/bin/bash

# 公共环境变量
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSL_MODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export GOWORK=/home/eric/payment/backend/go.work

BASE_DIR="/home/eric/payment/backend/services"

# 定义所有服务
declare -A SERVICES=(
    ["admin-service"]="8001:payment_admin"
    ["merchant-service"]="8002:payment_merchant"
    ["payment-gateway"]="8003:payment_gateway"
    ["order-service"]="8004:payment_order"
    ["channel-adapter"]="8005:payment_channel"
    ["risk-service"]="8006:payment_risk"
    ["accounting-service"]="8007:payment_accounting"
    ["notification-service"]="8008:payment_notification"
    ["analytics-service"]="8009:payment_analytics"
    ["config-service"]="8010:payment_config"
)

echo "========================================"
echo "启动所有后端服务"
echo "========================================"

for service in "${!SERVICES[@]}"; do
    IFS=':' read -r port db_name <<< "${SERVICES[$service]}"

    echo ""
    echo "启动 $service (端口: $port, 数据库: $db_name)"

    cd "$BASE_DIR/$service"

    # 设置服务特定的环境变量并后台启动
    (
        export PORT=$port
        export DB_NAME=$db_name
        go run cmd/main.go > "/tmp/$service.log" 2>&1
    ) &

    echo "✓ $service 已启动 (PID: $!)"
    sleep 2
done

echo ""
echo "========================================"
echo "所有服务已启动！"
echo "========================================"
echo ""
echo "查看日志："
echo "  tail -f /tmp/admin-service.log"
echo ""
echo "停止所有服务："
echo "  pkill -f 'go run cmd/main.go'"
EOF

chmod +x /home/eric/payment/backend/start-local.sh
```

启动所有服务：

```bash
/home/eric/payment/backend/start-local.sh
```

### 方式 3: 使用 Air 热重载 (开发推荐)

每个服务单独启动：

```bash
cd /home/eric/payment/backend/services/admin-service
air  # 自动读取 .air.toml 配置
```

---

## 启动前端应用

### Admin Portal (管理后台)

```bash
cd /home/eric/payment/frontend/admin-portal

# 安装依赖 (首次运行)
npm install

# 启动开发服务器
npm run dev

# 访问: http://localhost:40101
```

### Merchant Portal (商户门户)

```bash
cd /home/eric/payment/frontend/merchant-portal

# 安装依赖 (首次运行)
npm install

# 启动开发服务器
npm run dev

# 访问: http://localhost:40200
```

---

## 访问地址

### 前端应用

| 应用 | 地址 |
|-----|------|
| Admin Portal | http://localhost:40101 |
| Merchant Portal | http://localhost:40200 |

### 后端服务

| 服务 | 端口 | API 文档 | 健康检查 |
|-----|------|---------|---------|
| Admin Service | 8001 | http://localhost:8001/swagger/index.html | http://localhost:8001/health |
| Merchant Service | 8002 | http://localhost:8002/swagger/index.html | http://localhost:8002/health |
| Payment Gateway | 8003 | http://localhost:8003/swagger/index.html | http://localhost:8003/health |
| Order Service | 8004 | http://localhost:8004/swagger/index.html | http://localhost:8004/health |
| Channel Adapter | 8005 | http://localhost:8005/swagger/index.html | http://localhost:8005/health |
| Risk Service | 8006 | http://localhost:8006/swagger/index.html | http://localhost:8006/health |
| Accounting Service | 8007 | http://localhost:8007/swagger/index.html | http://localhost:8007/health |
| Notification Service | 8008 | http://localhost:8008/swagger/index.html | http://localhost:8008/health |
| Analytics Service | 8009 | http://localhost:8009/swagger/index.html | http://localhost:8009/health |
| Config Service | 8010 | http://localhost:8010/swagger/index.html | http://localhost:8010/health |

### 请求流程

```
前端应用 (40101/40200)
    ↓
Vite 开发服务器 (内置代理)
    ↓
根据 URL 路径自动路由到对应后端服务
    ↓
后端微服务 (8001-8010)
```

**示例：**
```
前端请求: http://localhost:40101/api/v1/admins
         ↓
Vite代理: /api/v1/admins → http://localhost:8001/api/v1/admins
         ↓
后端服务: Admin Service (8001)
```

---

## 代理配置说明

### Admin Portal 代理规则

前端配置文件：`frontend/admin-portal/vite.config.ts`

| 前端路径 | 代理到 | 后端服务 |
|---------|--------|---------|
| /api/v1/admins | localhost:8001 | Admin Service |
| /api/v1/roles | localhost:8001 | Admin Service |
| /api/v1/permissions | localhost:8001 | Admin Service |
| /api/v1/merchants | localhost:8002 | Merchant Service |
| /api/v1/payments | localhost:8003 | Payment Gateway |
| /api/v1/orders | localhost:8004 | Order Service |
| /api/v1/analytics | localhost:8009 | Analytics Service |

### Merchant Portal 代理规则

前端配置文件：`frontend/merchant-portal/vite.config.ts`

| 前端路径 | 代理到 | 后端服务 |
|---------|--------|---------|
| /api/v1/merchants | localhost:8002 | Merchant Service |
| /api/v1/api-keys | localhost:8002 | Merchant Service |
| /api/v1/payments | localhost:8003 | Payment Gateway |
| /api/v1/orders | localhost:8004 | Order Service |
| /api/v1/accounts | localhost:8007 | Accounting Service |

---

## 常见问题

### 1. 端口被占用

```bash
# 检查端口占用
lsof -i :8001

# 杀死进程
kill -9 <PID>
```

### 2. 数据库连接失败

检查环境变量配置：
```bash
echo $DB_HOST
echo $DB_PORT
echo $DB_NAME
```

### 3. 前端无法连接后端

1. 确认后端服务已启动
2. 检查 vite.config.ts 代理配置
3. 查看浏览器控制台错误信息
4. 确认后端健康检查接口正常：`curl http://localhost:8001/health`

### 4. Swagger 文档无法访问

确认服务已启动，并访问正确的路径：
```bash
curl http://localhost:8001/swagger/index.html
```

### 5. 停止所有后端服务

```bash
# 停止所有 go run 进程
pkill -f 'go run cmd/main.go'

# 或停止 air 进程
pkill air
```

### 6. 查看服务日志

```bash
# 如果使用启动脚本
tail -f /tmp/admin-service.log

# 如果手动启动，查看终端输出
```

---

## 开发建议

### 推荐的开发流程

1. **启动基础设施** (PostgreSQL, Redis)
2. **启动需要的后端服务** (不需要全部启动)
3. **启动前端应用**
4. **通过 Swagger UI 测试 API**
5. **通过前端应用测试完整流程**

### 高效开发技巧

1. **使用 Air 热重载**：修改代码自动重启服务
2. **只启动需要的服务**：不需要同时启动全部10个服务
3. **使用 Swagger UI**：快速测试 API 而不需要前端
4. **查看日志**：及时发现错误

---

## 生产环境部署

本文档仅适用于本地开发环境。

生产环境建议使用：
- **Docker Compose** + **Traefik**
- **Kubernetes** + **Ingress**

后续我们会创建生产环境的部署配置。

---

## 总结

✅ **本地开发优势：**
- 无需 Docker，直接运行
- 支持断点调试
- 快速迭代开发
- 灵活控制启停

🎯 **适用场景：**
- 日常开发调试
- 功能开发测试
- API 接口测试
- 前后端联调

🚀 **下一步：**
- 完善前端页面开发
- 编写集成测试
- 创建 Docker 部署配置

---

**祝开发愉快！** 🎉

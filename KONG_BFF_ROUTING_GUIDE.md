# Kong API Gateway + BFF 路由配置指南

**更新日期**: 2025-10-27
**架构**: 前端 → Kong Gateway → BFF Services → 微服务

---

## 🏗️ 架构概览

```
┌─────────────────┐       ┌──────────────────┐       ┌─────────────────────┐
│  Admin Portal   │       │                  │       │  admin-bff-service  │
│  (5173)         │──────▶│   Kong Gateway   │──────▶│  (40001)            │
│                 │       │   (40080)        │       │                     │
└─────────────────┘       │                  │       └─────────────────────┘
                          │                  │              │
┌─────────────────┐       │                  │              ├─▶ KYC Service
│ Merchant Portal │       │                  │              ├─▶ Order Service
│  (5174)         │──────▶│   JWT + CORS     │              ├─▶ Settlement
│                 │       │   Rate Limiting  │              ├─▶ Withdrawal
└─────────────────┘       │   Monitoring     │              └─▶ 18 services
                          │                  │
                          │                  │       ┌──────────────────────┐
                          └──────────────────┘──────▶│ merchant-bff-service │
                                                     │  (40023)             │
                                                     │                      │
                                                     └──────────────────────┘
                                                            │
                                                            ├─▶ Payment Gateway
                                                            ├─▶ Order Service
                                                            ├─▶ Settlement
                                                            └─▶ 15 services
```

---

## 📋 Kong 路由配置

### 核心路由

| 前端应用 | Kong路由 | 后端服务 | 端口 |
|---------|---------|---------|------|
| Admin Portal | `/api/v1/admin/*` | admin-bff-service | 40001 |
| Merchant Portal | `/api/v1/merchant/*` | merchant-bff-service | 40023 |

---

## 🔧 配置步骤

### 1. 启动基础设施

```bash
# 启动 Kong + PostgreSQL + Redis + Kafka
cd /home/eric/payment
docker-compose up -d kong-database kong-bootstrap kong konga
```

### 2. 启动 BFF 服务

```bash
cd /home/eric/payment/backend

# 启动 admin-bff-service
cd services/admin-bff-service
PORT=40001 go run cmd/main.go

# 启动 merchant-bff-service (另一个终端)
cd services/merchant-bff-service
PORT=40023 go run cmd/main.go
```

### 3. 配置 Kong 路由

```bash
# 执行 BFF 路由配置脚本
cd /home/eric/payment/backend/scripts
chmod +x kong-setup-bff.sh
./kong-setup-bff.sh
```

### 4. 启动前端

```bash
# Admin Portal
cd /home/eric/payment/frontend/admin-portal
npm run dev  # http://localhost:5173

# Merchant Portal (另一个终端)
cd /home/eric/payment/frontend/merchant-portal
npm run dev  # http://localhost:5174
```

---

## 🌐 访问地址

### Kong Services
- **Kong Proxy (API Gateway)**: http://localhost:40080
- **Kong Admin API**: http://localhost:40081
- **Konga Admin UI**: http://localhost:50001

### Frontend Applications
- **Admin Portal**: http://localhost:5173
- **Merchant Portal**: http://localhost:5174
- **Website**: http://localhost:5175

### BFF Services (Direct Access - 仅测试用)
- **Admin BFF**: http://localhost:40001
- **Merchant BFF**: http://localhost:40023

---

## 🔍 路由规则详解

### Admin BFF 路由

**Kong 配置**:
```bash
Service: admin-bff-service
URL: http://host.docker.internal:40001
Route: /api/v1/admin/*
```

**前端调用示例**:
```typescript
// Admin Portal (src/services/kycService.ts)
request.get('/api/v1/admin/kyc/documents')

// 实际请求流程:
// 1. Frontend → http://localhost:40080/api/v1/admin/kyc/documents
// 2. Kong → http://host.docker.internal:40001/api/v1/admin/kyc/documents
// 3. admin-bff-service → 处理请求并调用 kyc-service
```

**支持的接口** (70+):
- KYC管理: `/api/v1/admin/kyc/*`
- 订单管理: `/api/v1/admin/orders/*`
- 结算管理: `/api/v1/admin/settlements/*`
- 提现管理: `/api/v1/admin/withdrawals/*`
- 争议管理: `/api/v1/admin/disputes/*`
- 对账管理: `/api/v1/admin/reconciliation/*`
- 商户认证: `/api/v1/admin/merchant-auth/*`

---

### Merchant BFF 路由

**Kong 配置**:
```bash
Service: merchant-bff-service
URL: http://host.docker.internal:40023
Route: /api/v1/merchant/*
```

**前端调用示例**:
```typescript
// Merchant Portal
request.get('/api/v1/merchant/orders')

// 实际请求流程:
// 1. Frontend → http://localhost:40080/api/v1/merchant/orders
// 2. Kong → http://host.docker.internal:40023/api/v1/merchant/orders
// 3. merchant-bff-service → 自动注入merchant_id并调用 order-service
```

**支持的接口** (50+):
- 支付查询: `/api/v1/merchant/payments/*`
- 订单查询: `/api/v1/merchant/orders/*`
- 结算查询: `/api/v1/merchant/settlements/*`
- 提现申请: `/api/v1/merchant/withdrawals/*`
- API密钥: `/api/v1/merchant/merchant-auth/api-keys/*`

---

## 🔐 安全配置

### Kong 插件

| 插件 | 作用范围 | 配置 |
|-----|---------|------|
| CORS | Global | Origins: localhost:5173,5174,5175 |
| JWT | admin-bff-routes | Key claim: iss, Verify: exp |
| JWT | merchant-bff-routes | Key claim: iss, Verify: exp |
| Rate Limiting | admin-bff-routes | 60 req/min |
| Rate Limiting | merchant-bff-routes | 300 req/min |
| Request ID | Global | Header: X-Request-ID |
| Prometheus | Global | Metrics export |

### 认证流程

**Admin Portal**:
1. 用户登录 → `POST /api/v1/admin/login`
2. 获取 JWT Token
3. 后续请求带 `Authorization: Bearer {token}` header
4. Kong 验证 JWT → 转发到 admin-bff-service
5. BFF 验证权限 + RBAC → 调用微服务

**Merchant Portal**:
1. 商户登录 → `POST /api/v1/merchant/login`
2. 获取 JWT Token (包含 merchant_id)
3. 后续请求带 `Authorization: Bearer {token}` header
4. Kong 验证 JWT → 转发到 merchant-bff-service
5. BFF 提取 merchant_id + 租户隔离 → 调用微服务

---

## 📊 健康检查

### 检查 Kong 状态
```bash
curl http://localhost:40081/status
```

### 检查 BFF 服务
```bash
# Admin BFF
curl http://localhost:40001/health

# Merchant BFF
curl http://localhost:40023/health
```

### 检查 Kong 路由
```bash
# 列出所有路由
curl http://localhost:40081/routes

# 查看 admin-bff 路由
curl http://localhost:40081/routes/admin-bff-routes

# 查看 merchant-bff 路由
curl http://localhost:40081/routes/merchant-bff-routes
```

---

## 🧪 测试 API 调用

### 通过 Kong 测试 (推荐)

```bash
# Admin Login (不需要JWT)
curl -X POST http://localhost:40080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# KYC 文档列表 (需要JWT)
TOKEN="your-jwt-token-here"
curl -X GET "http://localhost:40080/api/v1/admin/kyc/documents?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# Merchant 订单查询 (需要JWT)
MERCHANT_TOKEN="your-merchant-jwt-token"
curl -X GET "http://localhost:40080/api/v1/merchant/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $MERCHANT_TOKEN"
```

### 直接访问 BFF (仅测试)

```bash
# 直接访问 admin-bff-service
curl -X GET "http://localhost:40001/api/v1/admin/kyc/documents?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# 直接访问 merchant-bff-service
curl -X GET "http://localhost:40023/api/v1/merchant/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $MERCHANT_TOKEN"
```

---

## 🐛 故障排查

### 问题1: Kong 无法连接到 BFF 服务

**症状**: 502 Bad Gateway

**解决方案**:
```bash
# 1. 检查 BFF 服务是否运行
lsof -i :40001  # admin-bff-service
lsof -i :40023  # merchant-bff-service

# 2. 检查 Docker 网络
docker network inspect bridge | grep host.docker.internal

# 3. 使用 docker host 模式 (macOS/Windows)
# 或改用 172.17.0.1 (Linux)
```

### 问题2: CORS 错误

**症状**: Access-Control-Allow-Origin 错误

**解决方案**:
```bash
# 检查 Kong CORS 插件
curl http://localhost:40081/plugins | jq '.data[] | select(.name=="cors")'

# 重新配置 CORS
./backend/scripts/kong-setup-bff.sh
```

### 问题3: JWT 验证失败

**症状**: 401 Unauthorized

**解决方案**:
```bash
# 1. 检查 JWT token 是否过期
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq

# 2. 检查 Kong JWT 插件配置
curl http://localhost:40081/plugins | jq '.data[] | select(.name=="jwt")'

# 3. 检查 JWT secret 是否一致
# 后端服务 JWT_SECRET 必须与 Kong consumer 的 secret 一致
```

### 问题4: 速率限制触发

**症状**: 429 Too Many Requests

**解决方案**:
```bash
# 检查速率限制配置
curl http://localhost:40081/plugins | jq '.data[] | select(.name=="rate-limiting")'

# 临时禁用速率限制 (测试用)
curl -X DELETE http://localhost:40081/plugins/{plugin-id}

# 调整速率限制 (生产环境)
# 修改 kong-setup-bff.sh 中的 config.minute 值
```

---

## 🎯 后续优化

### 短期 (1-2周)
- [ ] 配置 Kong 插件:
  - [ ] Request Transformer (头部转换)
  - [ ] Response Transformer (响应格式化)
  - [ ] IP Restriction (IP白名单)
- [ ] 配置日志聚合 (Loki/ELK)
- [ ] 配置告警规则 (Prometheus Alertmanager)

### 中期 (1-2月)
- [ ] 启用 mTLS (微服务间双向认证)
- [ ] 实现动态路由 (基于数据库配置)
- [ ] 集成 OAuth2 Provider
- [ ] 添加 API 版本管理

### 长期 (3-6月)
- [ ] Kong 集群部署 (高可用)
- [ ] 服务网格迁移 (Istio/Linkerd)
- [ ] API 网关性能优化
- [ ] 全链路追踪增强

---

## 📚 参考文档

- [Kong 官方文档](https://docs.konghq.com/)
- [Kong Admin API](https://docs.konghq.com/gateway/latest/admin-api/)
- [Admin BFF 安全文档](backend/services/admin-bff-service/ADVANCED_SECURITY_COMPLETE.md)
- [Merchant BFF 安全文档](backend/services/merchant-bff-service/MERCHANT_BFF_SECURITY.md)
- [前端 API 修复报告](frontend/admin-portal/ADMIN_API_FIX_REPORT.md)

---

**配置完成**: ✅
**前端对齐**: ✅
**Kong路由**: ⏳ 待配置
**测试验证**: ⏳ 待执行

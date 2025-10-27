# 前后端对齐测试检查清单

**完成状态**: ✅ 代码修复完成 → ⏳ 等待启动测试

---

## 📋 已完成的工作 (100%)

### 1. 前端 API 路径修复 ✅

修复的服务文件:
- ✅ [frontend/admin-portal/src/services/kycService.ts](frontend/admin-portal/src/services/kycService.ts) - 14个接口
- ✅ [frontend/admin-portal/src/services/orderService.ts](frontend/admin-portal/src/services/orderService.ts) - 5个接口
- ✅ [frontend/admin-portal/src/services/settlementService.ts](frontend/admin-portal/src/services/settlementService.ts) - 7个接口
- ✅ [frontend/admin-portal/src/services/withdrawalService.ts](frontend/admin-portal/src/services/withdrawalService.ts) - 8个接口
- ✅ [frontend/admin-portal/src/services/disputeService.ts](frontend/admin-portal/src/services/disputeService.ts) - 7个接口
- ✅ [frontend/admin-portal/src/services/reconciliationService.ts](frontend/admin-portal/src/services/reconciliationService.ts) - 9个接口
- ✅ [frontend/admin-portal/src/services/merchantAuthService.ts](frontend/admin-portal/src/services/merchantAuthService.ts) - 10个接口

**总计**: 70+ API 端点已更新

### 2. Kong 配置脚本 ✅

- ✅ [backend/scripts/kong-setup-bff.sh](backend/scripts/kong-setup-bff.sh) - Kong BFF 路由配置脚本
  - Admin BFF 路由: `/api/v1/admin/*` → `http://host.docker.internal:40001`
  - Merchant BFF 路由: `/api/v1/merchant/*` → `http://host.docker.internal:40023`
  - JWT 认证插件
  - 速率限制插件 (Admin: 60/min, Merchant: 300/min)

### 3. 文档产出 ✅

- ✅ [frontend/admin-portal/ADMIN_API_FIX_REPORT.md](frontend/admin-portal/ADMIN_API_FIX_REPORT.md) - 前端 API 修复详细报告
- ✅ [frontend/admin-portal/API_MISMATCH_ANALYSIS.md](frontend/admin-portal/API_MISMATCH_ANALYSIS.md) - 不匹配问题分析
- ✅ [KONG_BFF_ROUTING_GUIDE.md](KONG_BFF_ROUTING_GUIDE.md) - Kong 配置完整指南
- ✅ [FRONTEND_BACKEND_ALIGNMENT_SUMMARY.md](FRONTEND_BACKEND_ALIGNMENT_SUMMARY.md) - 前后端对齐完成总结

### 4. Git 提交 ✅

所有修改已提交到 Git:
```bash
git log --oneline -3
# 应该看到类似提交:
# - docs: 添加前后端接口对齐完成总结报告
# - docs: 添加Kong BFF路由配置指南和API修复报告
# - fix(frontend): 修复Admin Portal所有API路径以匹配admin-bff-service
```

---

## 🚀 测试步骤 (待执行)

### Step 1: 启动基础设施 (5 分钟)

```bash
cd /home/eric/payment

# 启动 Kong + PostgreSQL + Redis + Kafka
docker-compose up -d kong-database kong-bootstrap kong

# 等待 Kong 启动完成 (~30秒)
# 检查状态
docker-compose ps | grep kong

# 预期输出: kong 容器状态为 Up
```

**验证 Kong**:
```bash
curl http://localhost:40081/status
# 预期: {"database":{"reachable":true},"server":{"connections_active":1,...}}
```

---

### Step 2: 配置 Kong 路由 (2 分钟)

```bash
cd /home/eric/payment/backend/scripts

# 赋予执行权限
chmod +x kong-setup-bff.sh

# 执行配置
./kong-setup-bff.sh
```

**预期输出**:
```
==========================================
  Kong API Gateway BFF 配置工具
==========================================

ℹ 等待 Kong Gateway 启动...
✓ Kong Gateway 已就绪

ℹ 开始配置 BFF 服务...

ℹ 配置服务: admin-bff-service
✓ 服务 admin-bff-service 已创建
ℹ 配置服务: merchant-bff-service
✓ 服务 merchant-bff-service 已创建

ℹ 开始配置 BFF 路由...

ℹ 配置路由: admin-bff-routes
✓ 路由 admin-bff-routes 已创建
ℹ 配置路由: merchant-bff-routes
✓ 路由 merchant-bff-routes 已创建

ℹ 开始配置 BFF 插件...

ℹ 启用插件: jwt (route: admin-bff-routes)
✓ 插件 jwt 已启用
ℹ 启用插件: jwt (route: merchant-bff-routes)
✓ 插件 jwt 已启用
ℹ 启用插件: rate-limiting (route: admin-bff-routes)
✓ 插件 rate-limiting 已启用
ℹ 启用插件: rate-limiting (route: merchant-bff-routes)
✓ 插件 rate-limiting 已启用

✓ Kong BFF 配置完成!
```

**验证路由**:
```bash
curl http://localhost:40081/routes | jq '.data[] | select(.name == "admin-bff-routes")'
# 预期: 返回路由配置,包含 paths: ["/api/v1/admin"]
```

---

### Step 3: 启动 BFF 服务 (2 个终端)

**Terminal 1 - Admin BFF**:
```bash
cd /home/eric/payment/backend/services/admin-bff-service

# 设置环境变量
export PORT=40001
export DB_HOST=localhost
export DB_PORT=40432
export DB_NAME=payment_admin
export REDIS_HOST=localhost
export REDIS_PORT=40379
export JWT_SECRET=your-secret-key-min-32-characters-long

# 启动服务
go run cmd/main.go
```

**预期输出**:
```
[INFO] admin-bff-service starting on port 40001
[INFO] Database connected: payment_admin
[INFO] Redis connected
[INFO] Health check enabled on /health
[INFO] Swagger UI enabled on /swagger/index.html
[INFO] Server listening on :40001
```

**Terminal 2 - Merchant BFF** (可选,仅当测试商户门户时):
```bash
cd /home/eric/payment/backend/services/merchant-bff-service

export PORT=40023
export DB_HOST=localhost
export DB_PORT=40432
export DB_NAME=payment_merchant
export REDIS_HOST=localhost
export REDIS_PORT=40379
export JWT_SECRET=your-secret-key-min-32-characters-long

go run cmd/main.go
```

**验证 BFF 服务**:
```bash
# Admin BFF 健康检查
curl http://localhost:40001/health
# 预期: {"status":"healthy","dependencies":{...}}

# Merchant BFF 健康检查
curl http://localhost:40023/health
# 预期: {"status":"healthy","dependencies":{...}}
```

---

### Step 4: 启动依赖的微服务 (根据需要)

Admin BFF 依赖的核心服务:

**KYC Service** (如果测试 KYC 功能):
```bash
cd /home/eric/payment/backend/services/kyc-service
export PORT=40015
export DB_NAME=payment_kyc
go run cmd/main.go
```

**Order Service** (如果测试订单功能):
```bash
cd /home/eric/payment/backend/services/order-service
export PORT=40004
export DB_NAME=payment_order
go run cmd/main.go
```

**Settlement Service** (如果测试结算功能):
```bash
cd /home/eric/payment/backend/services/settlement-service
export PORT=40013
export DB_NAME=payment_settlement
go run cmd/main.go
```

**Withdrawal Service** (如果测试提现功能):
```bash
cd /home/eric/payment/backend/services/withdrawal-service
export PORT=40014
export DB_NAME=payment_withdrawal
go run cmd/main.go
```

**提示**: 可以使用 `backend/scripts/start-all-services.sh` 一键启动所有 19 个服务

---

### Step 5: 启动前端 (1 分钟)

```bash
cd /home/eric/payment/frontend/admin-portal

# 安装依赖 (首次)
npm install

# 启动开发服务器
npm run dev
```

**预期输出**:
```
VITE v5.x.x  ready in xxx ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
➜  press h + enter to show help
```

**访问**: 打开浏览器 http://localhost:5173

---

### Step 6: 手动功能测试 (15-30 分钟)

#### 6.1 登录测试

1. 打开浏览器 http://localhost:5173
2. 输入管理员账号登录
3. **检查浏览器 Network 标签**:
   - Request URL: `http://localhost:40080/api/v1/admin/login`
   - Status: `200`
   - Response 包含 `token` 字段

**cURL 测试**:
```bash
curl -X POST http://localhost:40080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -v
```

**预期响应**:
```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-28T12:00:00Z",
    "user": {...}
  }
}
```

#### 6.2 KYC 文档列表测试

1. 登录后,访问 KYC 管理页面
2. **检查浏览器 Network 标签**:
   - Request URL: `http://localhost:40080/api/v1/admin/kyc/documents?page=1&page_size=10`
   - Request Headers 包含: `Authorization: Bearer {token}`
   - Status: `200`
   - Response Headers 包含: `X-Request-ID` (Kong 添加)

**cURL 测试**:
```bash
# 先获取 token
TOKEN=$(curl -s -X POST http://localhost:40080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

# 调用 KYC 接口
curl -X GET "http://localhost:40080/api/v1/admin/kyc/documents?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" \
  -v
```

**预期响应**:
```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "documents": [
      {
        "id": "uuid-...",
        "merchant_id": "uuid-...",
        "document_type": "id_card",
        "status": "pending",
        ...
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

#### 6.3 其他功能测试

按照相同步骤测试:
- ✅ 订单列表: `GET /api/v1/admin/orders`
- ✅ 结算列表: `GET /api/v1/admin/settlements`
- ✅ 提现列表: `GET /api/v1/admin/withdrawals`
- ✅ 争议列表: `GET /api/v1/admin/disputes`
- ✅ 对账任务: `GET /api/v1/admin/reconciliation/tasks`

---

### Step 7: 验证安全特性 (5 分钟)

#### 7.1 CORS 验证

浏览器 Network 标签应该看到:
```
Response Headers:
  Access-Control-Allow-Origin: http://localhost:5173
  Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
  Access-Control-Allow-Headers: Content-Type, Authorization, X-Request-ID
```

#### 7.2 JWT 验证

**测试无 token 访问** (应该被拒绝):
```bash
curl -X GET http://localhost:40080/api/v1/admin/kyc/documents
# 预期: 401 Unauthorized
```

**测试过期 token** (应该被拒绝):
```bash
curl -X GET http://localhost:40080/api/v1/admin/kyc/documents \
  -H "Authorization: Bearer invalid-or-expired-token"
# 预期: 401 Unauthorized
```

#### 7.3 速率限制验证

**发送 61 个请求** (应该触发限流):
```bash
TOKEN="your-valid-token"
for i in {1..61}; do
  echo "Request $i"
  curl -X GET "http://localhost:40080/api/v1/admin/kyc/documents" \
    -H "Authorization: Bearer $TOKEN" \
    -w "\nStatus: %{http_code}\n"
  sleep 0.1
done

# 预期: 前 60 个返回 200, 第 61 个返回 429 Too Many Requests
```

**Response Headers** (第 61 个请求):
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit-Minute: 60
X-RateLimit-Remaining-Minute: 0
Retry-After: 60
```

#### 7.4 Request ID 传播

每个响应都应该包含唯一的请求 ID:
```bash
curl -X GET http://localhost:40080/api/v1/admin/kyc/documents \
  -H "Authorization: Bearer $TOKEN" \
  -v | grep -i "x-request-id"

# 预期: X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

---

## ✅ 验收标准

完成测试后,确认以下项目:

### 功能验收

- [ ] 登录功能正常,返回 JWT token
- [ ] KYC 文档列表可正常加载 (通过 Kong → admin-bff → kyc-service)
- [ ] 订单列表可正常加载
- [ ] 结算列表可正常加载
- [ ] 提现列表可正常加载
- [ ] 争议列表可正常加载
- [ ] 对账任务列表可正常加载
- [ ] 所有 API 调用通过 Kong Gateway (端口 40080)
- [ ] 所有 API 路径包含 `/admin/` 前缀

### 安全验收

- [ ] CORS 正常工作 (允许 localhost:5173)
- [ ] JWT 认证正常工作 (无 token 返回 401)
- [ ] JWT 过期检查正常 (过期 token 返回 401)
- [ ] 速率限制正常工作 (超过 60 req/min 返回 429)
- [ ] Request ID 正常生成和传播
- [ ] 所有敏感操作有审计日志 (检查 admin-bff-service 日志)

### 性能验收

- [ ] API 响应时间 < 500ms (P95)
- [ ] Kong 转发延迟 < 50ms
- [ ] BFF 聚合延迟 < 100ms
- [ ] 前端页面加载 < 2s

---

## 🐛 常见问题排查

### 问题 1: Kong 502 Bad Gateway

**症状**: 请求返回 502

**排查步骤**:
```bash
# 1. 检查 BFF 服务是否运行
lsof -i :40001  # admin-bff-service
lsof -i :40023  # merchant-bff-service

# 2. 检查 Kong 服务配置
curl http://localhost:40081/services/admin-bff-service

# 3. 检查 Docker 网络
docker network inspect bridge | grep host.docker.internal

# 4. 尝试直接访问 BFF (绕过 Kong)
curl http://localhost:40001/health
```

**解决方案**:
- 确保 BFF 服务正在运行
- Linux 系统可能需要改用 `172.17.0.1` 代替 `host.docker.internal`

### 问题 2: CORS 错误

**症状**: 浏览器控制台显示 CORS 错误

**排查步骤**:
```bash
# 检查 Kong CORS 插件
curl http://localhost:40081/plugins | jq '.data[] | select(.name=="cors")'
```

**解决方案**:
```bash
# 重新配置 CORS
cd backend/scripts
./kong-setup-bff.sh
```

### 问题 3: JWT 验证失败

**症状**: 所有请求返回 401

**排查步骤**:
```bash
# 1. 检查 token 内容
echo "eyJhbGc..." | cut -d'.' -f2 | base64 -d | jq

# 2. 检查 Kong JWT 插件
curl http://localhost:40081/plugins | jq '.data[] | select(.name=="jwt")'

# 3. 检查 BFF 服务 JWT_SECRET
# 确保 admin-bff-service 和 Kong consumer 的 secret 一致
```

### 问题 4: 速率限制触发太快

**症状**: 发送少量请求就返回 429

**排查步骤**:
```bash
# 检查速率限制配置
curl http://localhost:40081/plugins | jq '.data[] | select(.name=="rate-limiting")'
```

**临时解决方案**:
```bash
# 禁用速率限制 (仅测试用)
PLUGIN_ID=$(curl -s http://localhost:40081/plugins | jq -r '.data[] | select(.name=="rate-limiting" and .route.name=="admin-bff-routes") | .id')
curl -X DELETE http://localhost:40081/plugins/$PLUGIN_ID
```

---

## 📊 测试报告模板

测试完成后,请记录以下信息:

```markdown
# 前后端对齐测试报告

**测试日期**: YYYY-MM-DD
**测试人员**: [姓名]
**测试环境**: Development

## 测试结果

### 1. 功能测试
- [ ] 登录: ✅ 通过 / ❌ 失败 - [错误描述]
- [ ] KYC 文档列表: ✅ 通过 / ❌ 失败
- [ ] 订单列表: ✅ 通过 / ❌ 失败
- [ ] 结算列表: ✅ 通过 / ❌ 失败
- [ ] 提现列表: ✅ 通过 / ❌ 失败
- [ ] 争议列表: ✅ 通过 / ❌ 失败
- [ ] 对账任务: ✅ 通过 / ❌ 失败

### 2. 安全测试
- [ ] CORS: ✅ 通过 / ❌ 失败
- [ ] JWT 认证: ✅ 通过 / ❌ 失败
- [ ] 速率限制: ✅ 通过 / ❌ 失败
- [ ] Request ID: ✅ 通过 / ❌ 失败

### 3. 性能测试
- API 平均响应时间: [xxx ms]
- Kong 转发延迟: [xxx ms]
- 前端页面加载时间: [xxx s]

## 发现的问题

1. [问题描述]
   - 严重程度: 高/中/低
   - 复现步骤: [...]
   - 预期结果: [...]
   - 实际结果: [...]

## 待修复的后端接口

根据测试发现,以下接口需要在 admin-bff-service 中补充:
- [ ] `GET /api/v1/admin/withdrawals/statistics`
- [ ] `GET /api/v1/admin/disputes/export`
- [ ] `GET /api/v1/admin/reconciliation/statistics`
- [ ] `GET /api/v1/admin/merchant-auth/security`

## 总结

[测试总体评价和建议]
```

---

## 📈 下一步工作

### 短期 (测试完成后)

1. **修复发现的问题**
2. **补充缺失的后端接口** (根据测试报告)
3. **优化性能瓶颈** (如果发现)

### 中期 (本周内)

1. **对齐 Merchant Portal** (同样的流程)
2. **添加集成测试** (自动化测试脚本)
3. **配置 Kong 生产环境**

### 长期 (本月内)

1. **实现 API 版本管理** (v1, v2)
2. **添加 GraphQL 网关** (可选)
3. **启用 mTLS** (微服务间双向认证)

---

**准备就绪**: ✅ 所有代码和配置已完成,可以开始测试!

**预计测试时间**: 1-2 小时 (包括问题排查)
**预计修复时间**: 2-3 小时 (补充缺失接口)
**预计全部完成**: 今天内

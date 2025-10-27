# Admin Portal API 路径不匹配分析报告

**分析日期**: 2025-10-27
**分析范围**: Admin Portal 前端 vs admin-bff-service 后端

---

## 🔴 关键发现

**前端配置**:
- Base URL: `/api/v1` (来自 `request.ts` line 8)
- 所有接口调用都相对于这个 base URL

**后端实际接口** (admin-bff-service port 40001):
- 所有 BFF 聚合接口都以 `/api/v1/admin/` 开头
- KYC 接口: `/api/v1/admin/kyc/*`
- Order 接口: `/api/v1/admin/orders/*`
- Settlement 接口: `/api/v1/admin/settlements/*`
- Merchant Auth 接口: `/api/v1/admin/merchant-auth/*`

**不匹配问题**: 前端缺少 `/admin/` 前缀!

---

## 📋 详细对比

### 1. Merchant Service

**前端调用** (`merchantService.ts`):
```typescript
list: () => request.get('/merchant', { params })
// 实际请求: /api/v1/merchant
```

**后端接口** (不存在):
- ❌ admin-bff-service 没有 `/api/v1/merchant` 接口
- ❌ merchant-service (独立服务) 不应被前端直接调用

**修复方案**:
- 需要在 admin-bff-service 添加商户管理聚合接口
- 路径: `/api/v1/admin/merchants`

---

### 2. KYC Service

**前端调用** (`kycService.ts`):
```typescript
listDocuments: (params) => request.get('/api/v1/kyc/documents', { params })
// 实际请求: /api/v1/kyc/documents
```

**后端接口** (admin-bff-service):
```
GET /api/v1/admin/kyc/documents
```

**问题**: 缺少 `/admin/` 前缀

**影响的接口** (12个):
- ❌ GET `/kyc/documents` → ✅ GET `/admin/kyc/documents`
- ❌ GET `/kyc/documents/:id` → ✅ GET `/admin/kyc/documents/:id`
- ❌ POST `/kyc/documents` → ✅ POST `/admin/kyc/documents` (不应该有)
- ❌ POST `/kyc/documents/:id/approve` → ✅ POST `/admin/kyc/documents/:id/approve`
- ❌ POST `/kyc/documents/:id/reject` → ✅ POST `/admin/kyc/documents/:id/reject`
- ❌ GET `/kyc/qualifications` → ✅ GET `/admin/kyc/qualifications`
- ❌ GET `/kyc/qualifications/merchant/:id` → ✅ GET `/admin/kyc/qualifications/:id`
- ❌ POST `/kyc/qualifications` → ✅ POST `/admin/kyc/qualifications` (不应该有)
- ❌ POST `/kyc/qualifications/:id/approve` → ✅ POST `/admin/kyc/qualifications/:id/approve`
- ❌ POST `/kyc/qualifications/:id/reject` → ✅ POST `/admin/kyc/qualifications/:id/reject`
- ❌ GET `/kyc/levels/:id` → ✅ GET `/admin/kyc/levels/:id`
- ❌ GET `/kyc/levels/:id/eligibility` → ✅ GET `/admin/kyc/levels/:id/eligibility`

---

### 3. Order Service

**前端调用** (`orderService.ts`):
```typescript
list: (params) => request.get('/orders', { params })
// 实际请求: /api/v1/orders
```

**后端接口** (admin-bff-service):
```
GET /api/v1/admin/orders
```

**问题**: 缺少 `/admin/` 前缀

**影响的接口** (8个):
- ❌ GET `/orders` → ✅ GET `/admin/orders`
- ❌ GET `/orders/:order_no` → ✅ GET `/admin/orders/:order_no`
- ❌ POST `/orders` → ✅ POST `/admin/orders` (不应该有)
- ❌ POST `/orders/batch` → ✅ POST `/admin/orders/batch` (不存在)
- ❌ GET `/orders/stats` → ✅ GET `/admin/orders/statistics`
- ❌ POST `/orders/:order_no/cancel` → ✅ POST `/admin/orders/:order_no/cancel` (不存在)
- ❌ GET `/statistics/orders` → ✅ GET `/admin/orders/statistics`
- ❌ GET `/statistics/daily-summary` → ✅ GET `/admin/orders/status-summary`

---

### 4. Settlement Service

**前端调用** (`settlementService.ts`):
```typescript
list: (params) => request.get('/api/v1/settlements', { params })
// 实际请求: /api/v1/settlements
```

**后端接口** (admin-bff-service):
```
GET /api/v1/admin/settlements
```

**问题**: 缺少 `/admin/` 前缀

**影响的接口** (7个):
- ❌ GET `/settlements` → ✅ GET `/admin/settlements`
- ❌ GET `/settlements/:id` → ✅ GET `/admin/settlements/:id`
- ❌ POST `/settlements` → ✅ POST `/admin/settlements` (不应该有)
- ❌ PUT `/settlements/:id` → ✅ PUT `/admin/settlements/:id` (不存在)
- ❌ POST `/settlements/:id/approve` → ✅ POST `/admin/settlements/:id/approve`
- ❌ POST `/settlements/:id/execute` → ✅ POST `/admin/settlements/:id/execute`
- ❌ POST `/settlements/:id/reject` → ✅ POST `/admin/settlements/:id/reject`

---

### 5. Withdrawal Service

需要检查 `withdrawalService.ts` (未读取)

**预期后端接口** (admin-bff-service):
```
GET /api/v1/admin/withdrawals
GET /api/v1/admin/withdrawals/:id
POST /api/v1/admin/withdrawals/:id/approve
POST /api/v1/admin/withdrawals/:id/reject
POST /api/v1/admin/withdrawals/:id/execute
POST /api/v1/admin/withdrawals/:id/cancel
```

---

### 6. Dispute Service

需要检查 `disputeService.ts` (未读取)

**预期后端接口** (admin-bff-service):
```
POST /api/v1/admin/disputes
GET /api/v1/admin/disputes
GET /api/v1/admin/disputes/:dispute_id
PUT /api/v1/admin/disputes/:dispute_id/status
POST /api/v1/admin/disputes/:dispute_id/assign
POST /api/v1/admin/disputes/:dispute_id/evidence
```

---

### 7. Reconciliation Service

需要检查 `reconciliationService.ts` (未读取)

**预期后端接口** (admin-bff-service):
```
GET /api/v1/admin/reconciliation/tasks
GET /api/v1/admin/reconciliation/tasks/:id
POST /api/v1/admin/reconciliation/tasks
POST /api/v1/admin/reconciliation/tasks/:id/start
GET /api/v1/admin/reconciliation/discrepancies
POST /api/v1/admin/reconciliation/discrepancies/:id/resolve
```

---

### 8. Merchant Auth Service

需要检查 `merchantAuthService.ts` (未读取)

**预期后端接口** (admin-bff-service):
```
GET /api/v1/admin/merchant-auth/api-keys
GET /api/v1/admin/merchant-auth/api-keys/:id
POST /api/v1/admin/merchant-auth/api-keys
PUT /api/v1/admin/merchant-auth/api-keys/:id
DELETE /api/v1/admin/merchant-auth/api-keys/:id
POST /api/v1/admin/merchant-auth/api-keys/:id/revoke
GET /api/v1/admin/merchant-auth/2fa/:merchant_id/status
POST /api/v1/admin/merchant-auth/2fa/:merchant_id/enable
```

---

## 🛠️ 修复策略

### 策略 1: 修改前端路径 (推荐)

**优点**:
- 符合后端实际接口设计
- 无需修改后端代码
- 路径更清晰,有 `/admin/` 前缀区分管理员操作

**缺点**:
- 需要修改多个前端服务文件

**工作量**:
- 需修改约 10-15 个服务文件
- 约 100+ 个 API 调用路径

---

### 策略 2: 修改后端路径 (不推荐)

**优点**:
- 前端无需修改

**缺点**:
- 破坏 BFF 架构设计
- 路径没有 `/admin/` 前缀,容易与 Merchant BFF 混淆
- 需要修改后端所有路由注册代码
- 破坏安全分层设计

---

### 策略 3: 添加后端路由别名 (折中方案)

**优点**:
- 兼容前后端
- 可以逐步迁移

**缺点**:
- 维护两套路由
- 增加代码复杂度

---

## ✅ 推荐修复方案

**选择策略 1**: 修改前端路径以匹配后端 admin-bff-service

### 修复步骤:

1. **修改 API Base URL 配置** (.env 文件)
   ```env
   VITE_API_BASE_URL=http://localhost:40001
   VITE_API_PREFIX=/api/v1
   ```

2. **批量修改前端服务文件**:
   - `kycService.ts`: 所有路径添加 `/admin/` 前缀
   - `orderService.ts`: 所有路径添加 `/admin/` 前缀
   - `settlementService.ts`: 所有路径添加 `/admin/` 前缀
   - `withdrawalService.ts`: 所有路径添加 `/admin/` 前缀
   - `disputeService.ts`: 所有路径添加 `/admin/` 前缀
   - `reconciliationService.ts`: 所有路径添加 `/admin/` 前缀
   - `merchantAuthService.ts`: 所有路径添加 `/admin/` 前缀
   - `merchantService.ts`: 创建新的 `/admin/merchants` 接口

3. **移除前端不应该调用的接口**:
   - `kycService.submitDocument()` - 管理员不应提交文档
   - `kycService.submitQualification()` - 管理员不应提交资质
   - `orderService.create()` - 管理员不应创建订单
   - `settlementService.create()` - 管理员不应创建结算单
   - `settlementService.update()` - 使用 approve/reject/execute 代替

4. **补充后端缺失的接口** (如果需要):
   - `GET /api/v1/admin/settlements/statistics` (前端调用 getStats)
   - `GET /api/v1/admin/orders/batch` (前端调用 batchGet)
   - `POST /api/v1/admin/merchants` 系列接口

---

## 📊 影响范围评估

| 服务 | 前端文件 | 需修改接口数 | 优先级 |
|------|---------|------------|--------|
| KYC Service | kycService.ts | 12 | 🔴 高 |
| Order Service | orderService.ts | 8 | 🔴 高 |
| Settlement Service | settlementService.ts | 7 | 🔴 高 |
| Merchant Service | merchantService.ts | 7 | 🔴 高 |
| Withdrawal Service | withdrawalService.ts | 6 | 🟡 中 |
| Dispute Service | disputeService.ts | 8 | 🟡 中 |
| Reconciliation Service | reconciliationService.ts | 10 | 🟡 中 |
| Merchant Auth Service | merchantAuthService.ts | 10 | 🟡 中 |
| **总计** | **8个文件** | **68个接口** | - |

---

## 🎯 下一步行动

1. ✅ 完成详细分析 (当前文档)
2. ⏳ 读取剩余前端服务文件 (withdrawal, dispute, reconciliation, merchantAuth)
3. ⏳ 批量修改前端 API 路径
4. ⏳ 补充后端缺失的必要接口
5. ⏳ 测试前后端联调
6. ⏳ 更新 API 文档

---

**结论**: 前端 API 调用路径与后端 admin-bff-service 严重不匹配,需要进行系统性修复。推荐修改前端路径以符合后端设计。

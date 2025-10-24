# 商户后台服务架构总结

## 三个Merchant微服务的职责划分

### 1. **merchant-service** (端口 40002)
**核心商户管理服务**

#### 功能：
- ✅ 商户注册/登录认证
- ✅ 商户基本信息管理 (profile)
- ✅ Dashboard数据聚合
  - 交易汇总统计
  - 余额信息
  - 商户概览数据
- ✅ 商户用户管理

#### Kong路由：
```
POST /api/v1/merchant/register  → merchant-service
POST /api/v1/merchant/login     → merchant-service
GET  /api/v1/merchant/profile   → merchant-service
GET  /api/v1/dashboard          → merchant-service
```

---

### 2. **merchant-auth-service** (端口 40011) ✅ 新增
**商户认证和安全管理服务**

#### 功能：
- ✅ **API Key 管理**
  - 创建 API Key (test/production)
  - 列出所有 API Keys
  - 删除/撤销 API Key
  - 验证 API 签名 (供 payment-gateway 调用)

- ✅ **安全设置**
  - 修改密码
  - 双因素认证 (2FA)
    - 启用 2FA
    - 验证 2FA
    - 禁用 2FA
  - 安全设置管理
    - IP白名单
    - 会话超时
    - 密码策略

#### Kong路由：
```
GET    /api/v1/api-keys                  → merchant-auth-service
POST   /api/v1/api-keys                  → merchant-auth-service
DELETE /api/v1/api-keys/:id              → merchant-auth-service

PUT    /api/v1/security/password         → merchant-auth-service
POST   /api/v1/security/2fa/enable       → merchant-auth-service
POST   /api/v1/security/2fa/verify       → merchant-auth-service
POST   /api/v1/security/2fa/disable      → merchant-auth-service
GET    /api/v1/security/settings         → merchant-auth-service
PUT    /api/v1/security/settings         → merchant-auth-service

POST   /api/v1/auth/validate-signature   → merchant-auth-service (内部调用)
```

---

### 3. **merchant-config-service** (端口 40012) ✅ 新增
**商户业务配置管理服务**

#### 功能：
- ✅ **费率配置**
  - 配置支付手续费率 (百分比/固定/混合)
  - 按渠道和支付方式配置
  - 费率审批流程
  - 手续费计算

- ✅ **交易限额**
  - 单笔交易限额
  - 日交易限额
  - 月交易限额
  - 限额检查

- ✅ **渠道配置**
  - 支付渠道启用/禁用 (Stripe, PayPal等)
  - 渠道优先级设置
  - 渠道API密钥配置
  - 测试模式/生产模式切换

#### Kong路由：
```
GET    /api/v1/fee-configs                          → merchant-config-service
POST   /api/v1/fee-configs                          → merchant-config-service
GET    /api/v1/fee-configs/merchant/:merchant_id    → merchant-config-service
PUT    /api/v1/fee-configs/:id                      → merchant-config-service
DELETE /api/v1/fee-configs/:id                      → merchant-config-service
POST   /api/v1/fee-configs/calculate-fee            → merchant-config-service

GET    /api/v1/transaction-limits                   → merchant-config-service
POST   /api/v1/transaction-limits                   → merchant-config-service
GET    /api/v1/transaction-limits/merchant/:merchant_id → merchant-config-service
PUT    /api/v1/transaction-limits/:id               → merchant-config-service
DELETE /api/v1/transaction-limits/:id               → merchant-config-service
POST   /api/v1/transaction-limits/check-limit       → merchant-config-service

GET    /api/v1/channel-configs                      → merchant-config-service
POST   /api/v1/channel-configs                      → merchant-config-service
GET    /api/v1/channel-configs/merchant/:merchant_id → merchant-config-service
PUT    /api/v1/channel-configs/:id                  → merchant-config-service
DELETE /api/v1/channel-configs/:id                  → merchant-config-service
POST   /api/v1/channel-configs/:id/enable           → merchant-config-service
POST   /api/v1/channel-configs/:id/disable          → merchant-config-service
```

---

## 前端Service封装

### 已创建的前端服务：

1. **merchantService.ts** - 商户基本信息管理
   - 登录/注册
   - 获取个人信息
   - 更新个人信息

2. **dashboardService.ts** ✅ - Dashboard数据
   - 获取Dashboard概览
   - 交易汇总
   - 余额信息

3. **apiKeyService.ts** ✅ - API Key和安全管理
   - API Key CRUD
   - 密码管理
   - 2FA管理
   - 安全设置

4. **configService.ts** ✅ - 商户配置管理
   - 费率配置 CRUD
   - 交易限额 CRUD
   - 渠道配置 CRUD

---

## Kong网关配置状态

### 已注册服务：
- ✅ merchant-service (40002)
- ✅ merchant-auth-service (40011)
- ✅ merchant-config-service (40012)

### 路由配置：
- ✅ 所有merchant相关路由已配置
- ✅ 通过Kong统一入口 (localhost:40080)
- ✅ 支持JWT认证中间件

---

## 商户后台完整功能清单

### 认证与安全
- ✅ 商户注册
- ✅ 商户登录 (JWT)
- ✅ 修改密码
- ✅ 双因素认证 (2FA)
- ✅ IP白名单
- ✅ 会话管理

### API Key 管理
- ✅ 创建测试/生产环境API Key
- ✅ 查看API Key列表
- ✅ 撤销API Key
- ✅ API签名验证

### Dashboard
- ✅ 交易概览统计
- ✅ 余额信息
- ✅ 实时数据展示

### 配置管理
- ✅ 费率配置 (按渠道/支付方式)
- ✅ 交易限额设置
- ✅ 支付渠道管理
- ✅ 手续费计算

### 支付业务
- ⏳ 支付查询 (需调用payment-gateway)
- ⏳ 订单管理 (需调用order-service)
- ⏳ 退款管理 (需调用payment-gateway)
- ⏳ 结算报表 (需调用settlement-service)

---

## 测试信息

### 测试账号：
- 邮箱：`test@test.com` 或 `merchant@example.com`
- 密码：`password123`
- 状态：active

### 访问地址：
- 商户后台：http://localhost:5174
- Kong网关：http://localhost:40080

### 示例API调用：
```bash
# 1. 登录获取token
curl -X POST http://localhost:40080/api/v1/merchant/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123"}'

# 2. 获取Dashboard数据
curl http://localhost:40080/api/v1/dashboard \
  -H "Authorization: Bearer <TOKEN>"

# 3. 创建API Key
curl -X POST http://localhost:40080/api/v1/api-keys \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"environment":"test","description":"Test API Key"}'

# 4. 获取费率配置
curl http://localhost:40080/api/v1/fee-configs/merchant/<MERCHANT_ID> \
  -H "Authorization: Bearer <TOKEN>"
```

---

## 总结

商户后台现在有**完整的三层服务架构**：

1. **merchant-service** - 核心商户管理和Dashboard
2. **merchant-auth-service** - 认证、API Key和安全
3. **merchant-config-service** - 业务配置管理

所有服务都已：
- ✅ 注册到Kong网关
- ✅ 配置正确的路由
- ✅ 前端Service封装完成
- ✅ 支持JWT认证
- ✅ 可通过merchant-portal访问

商户登录后可以：
- 查看Dashboard数据
- 管理API Keys
- 配置费率和限额
- 管理支付渠道
- 修改安全设置

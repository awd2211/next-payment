# Backend API Reference for Admin Portal

本文档记录了Admin Portal前端应该调用的所有后端API路由。

## 路由分类

### 1. Admin Service (Port 40001)

#### Admin Management (`/api/v1/admin`)
- `POST /api/v1/admin/login` - 管理员登录 (无需认证)
- `POST /api/v1/admin` - 创建管理员 (需认证)
- `GET /api/v1/admin/:id` - 获取管理员详情
- `GET /api/v1/admin` - 获取管理员列表
  - Query params: `page`, `page_size`, `status`, `keyword`
- `PUT /api/v1/admin/:id` - 更新管理员
- `DELETE /api/v1/admin/:id` - 删除管理员
- `POST /api/v1/admin/change-password` - 修改密码
- `POST /api/v1/admin/:id/reset-password` - 重置密码

#### Role Management (`/api/v1/roles`)
- `POST /api/v1/roles` - 创建角色
- `GET /api/v1/roles/:id` - 获取角色详情
- `GET /api/v1/roles` - 获取角色列表
  - Query params: `page`, `page_size`
- `PUT /api/v1/roles/:id` - 更新角色
- `DELETE /api/v1/roles/:id` - 删除角色
- `POST /api/v1/roles/:id/permissions` - 为角色分配权限
- `POST /api/v1/roles/assign` - 为管理员分配角色

#### Audit Logs (`/api/v1/audit-logs`)
- `GET /api/v1/audit-logs/:id` - 获取审计日志详情
- `GET /api/v1/audit-logs` - 获取审计日志列表
  - Query params: `page`, `page_size`, `admin_id`, `action`, `resource`, `method`, `ip`, `response_code`, `start_time`, `end_time`
- `GET /api/v1/audit-logs/stats` - 获取统计信息
  - Query params: `start_time`, `end_time`

---

### 2. Merchant Service (Port 40002)

#### Dashboard - Merchant Operations (`/api/v1/dashboard`)
- `GET /api/v1/dashboard` - 获取商户Dashboard数据 (商户端,需JWT认证)
  - 自动从JWT token获取merchant_id
- `GET /api/v1/dashboard/transaction-summary` - 获取交易汇总
  - Query params: `start_date`, `end_date`
- `GET /api/v1/dashboard/balance` - 获取余额信息

**注意**: 这些Dashboard API是为**Merchant Portal**设计的,不是Admin Portal用的!

#### Merchant Management - Admin Operations (`/api/v1/merchant`)
- `POST /api/v1/merchant/register` - 商户注册 (无需认证)
- `POST /api/v1/merchant/login` - 商户登录 (无需认证)
- `POST /api/v1/merchant` - 创建商户 (管理员操作)
- `GET /api/v1/merchant/:id` - 获取商户详情
- `GET /api/v1/merchant` - 获取商户列表
  - Query params: `page`, `page_size`, `status`, `kyc_status`, `keyword`
- `PUT /api/v1/merchant/:id` - 更新商户
- `PUT /api/v1/merchant/:id/status` - 更新商户状态
- `PUT /api/v1/merchant/:id/kyc-status` - 更新KYC状态
- `DELETE /api/v1/merchant/:id` - 删除商户

#### Merchant Profile - Merchant Operations (`/api/v1/merchant`)
- `GET /api/v1/merchant/profile` - 获取当前商户信息 (商户端)
- `PUT /api/v1/merchant/profile` - 更新当前商户信息 (商户端)
- `GET /api/v1/merchant/balance` - 获取商户余额 (临时占位API)
- `GET /api/v1/merchant/stats` - 获取商户统计 (临时占位API)

#### Internal APIs (`/api/v1/merchants`)
- `GET /api/v1/merchants/:id/with-password` - 获取带密码的商户信息 (内部接口)
- `PUT /api/v1/merchants/:id/password` - 更新商户密码 (内部接口)

---

### 3. Payment Gateway (Port 40003)

#### Payment Management (`/api/v1/payments`)
- `POST /api/v1/payments` - 创建支付 (需API Key)
- `GET /api/v1/payments/:paymentNo` - 获取支付详情
- `GET /api/v1/payments` - 查询支付列表
  - Query params: `merchant_id`, `channel`, `status`, `currency`, `customer_email`, `keyword`, `start_time`, `end_time`, `min_amount`, `max_amount`, `page`, `page_size`
- `POST /api/v1/payments/batch` - 批量查询支付
- `POST /api/v1/payments/:paymentNo/cancel` - 取消支付

#### Refund Management (`/api/v1/refunds`)
- `POST /api/v1/refunds` - 创建退款
- `GET /api/v1/refunds/:refundNo` - 获取退款详情
- `GET /api/v1/refunds` - 查询退款列表
  - Query params: `merchant_id`, `payment_id`, `status`, `start_time`, `end_time`, `page`, `page_size`
- `POST /api/v1/refunds/batch` - 批量查询退款

#### Merchant Portal Routes (`/api/v1/merchant`)
- `GET /api/v1/merchant/payments` - 查询支付列表 (商户端,使用JWT)
- `GET /api/v1/merchant/payments/:paymentNo` - 获取支付详情 (商户端)
- `POST /api/v1/merchant/payments/batch` - 批量查询支付 (商户端)
- `GET /api/v1/merchant/refunds` - 查询退款列表 (商户端)
- `GET /api/v1/merchant/refunds/:refundNo` - 获取退款详情 (商户端)
- `POST /api/v1/merchant/refunds/batch` - 批量查询退款 (商户端)

#### Webhooks (`/api/v1/webhooks`)
- `POST /api/v1/webhooks/stripe` - 处理Stripe回调
- `POST /api/v1/webhooks/paypal` - 处理PayPal回调

---

### 4. Order Service (Port 40004)

#### Order Management (`/api/v1/orders`)
- `POST /api/v1/orders` - 创建订单
- `GET /api/v1/orders/:orderNo` - 获取订单详情
- `GET /api/v1/orders` - 查询订单列表
  - Query params: `merchant_id`, `customer_id`, `status`, `pay_status`, `shipping_status`, `currency`, `customer_email`, `keyword`, `start_time`, `end_time`, `min_amount`, `max_amount`, `page`, `page_size`
- `POST /api/v1/orders/batch` - 批量查询订单
- `GET /api/v1/orders/stats` - 获取订单统计 (临时占位API)
- `POST /api/v1/orders/:orderNo/cancel` - 取消订单
- `POST /api/v1/orders/:orderNo/pay` - 支付订单
- `POST /api/v1/orders/:orderNo/refund` - 退款订单
- `POST /api/v1/orders/:orderNo/ship` - 订单发货
- `POST /api/v1/orders/:orderNo/complete` - 完成订单
- `PUT /api/v1/orders/:orderNo/status` - 更新订单状态 (支付网关回调使用)

#### Statistics (`/api/v1/statistics`)
- `GET /api/v1/statistics/orders` - 获取订单统计
  - Query params: `merchant_id` (required), `start_time` (required), `end_time` (required), `currency`
- `GET /api/v1/statistics/daily-summary` - 获取每日汇总
  - Query params: `merchant_id` (required), `date`, `currency`

---

### 5. Channel Adapter (Port 40005)

#### Admin Channel Management (`/api/v1/admin/channels`)
- `GET /api/v1/admin/channels` - 获取渠道列表 (管理员端)
  - Query params: `page`, `page_size`, `channel_type`, `is_enabled`, `is_test_mode`
- `GET /api/v1/admin/channels/:code` - 获取渠道详情
- `POST /api/v1/admin/channels` - 创建渠道
- `PUT /api/v1/admin/channels/:code` - 更新渠道
- `DELETE /api/v1/admin/channels/:code` - 删除渠道

**注意**: 当前返回硬编码数据 (Stripe, PayPal, Alipay, WeChat Pay)

---

### 6. Analytics Service (Port 40009)

#### Payment Analytics (`/api/v1/analytics/payments`)
- `GET /api/v1/analytics/payments/metrics` - 获取支付指标
  - Query params: `merchant_id` (required), `start_date`, `end_date`
- `GET /api/v1/analytics/payments/summary` - 获取支付汇总
  - Query params: `merchant_id` (required), `start_date`, `end_date`

#### Merchant Analytics (`/api/v1/analytics/merchants`)
- `GET /api/v1/analytics/merchants/metrics` - 获取商户指标
  - Query params: `merchant_id` (required), `start_date`, `end_date`
- `GET /api/v1/analytics/merchants/summary` - 获取商户汇总
  - Query params: `merchant_id` (required), `start_date`, `end_date`

#### Channel Analytics (`/api/v1/analytics/channels`)
- `GET /api/v1/analytics/channels/metrics` - 获取渠道指标
  - Query params: `channel_code` (required), `start_date`, `end_date`
- `GET /api/v1/analytics/channels/summary` - 获取渠道汇总
  - Query params: `channel_code` (required), `start_date`, `end_date`

#### Realtime Stats (`/api/v1/analytics/realtime`)
- `GET /api/v1/analytics/realtime/stats` - 获取实时统计
  - Query params: `merchant_id`, `stat_type`, `stat_key`, `period`

---

### 7. Accounting Service (Port 40007)

#### Account Management (`/api/v1/accounts`)
- `POST /api/v1/accounts` - 创建账户
- `GET /api/v1/accounts/:id` - 获取账户
- `GET /api/v1/accounts` - 获取账户列表
  - Query params: `page`, `page_size`, `merchant_id`
- `POST /api/v1/accounts/:id/freeze` - 冻结账户
- `POST /api/v1/accounts/:id/unfreeze` - 解冻账户

#### Transaction Management (`/api/v1/transactions`)
- `POST /api/v1/transactions` - 创建交易 (复式记账)
- `GET /api/v1/transactions/:transactionNo` - 获取交易
- `GET /api/v1/transactions` - 获取交易列表
  - Query params: `page`, `page_size`, `merchant_id`, `transaction_type`
- `POST /api/v1/transactions/:transactionNo/reverse` - 冲正交易

#### Settlement Management (`/api/v1/settlements`)
- `POST /api/v1/settlements` - 创建结算
- `GET /api/v1/settlements/:settlementNo` - 获取结算
- `GET /api/v1/settlements` - 获取结算列表
- `POST /api/v1/settlements/:settlementNo/process` - 处理结算

#### Withdrawal Management (`/api/v1/withdrawals`)
- `POST /api/v1/withdrawals` - 创建提现
- `GET /api/v1/withdrawals/:withdrawalNo` - 获取提现
- `GET /api/v1/withdrawals` - 获取提现列表
- `POST /api/v1/withdrawals/:withdrawalNo/approve` - 批准提现
- `POST /api/v1/withdrawals/:withdrawalNo/reject` - 拒绝提现
- `POST /api/v1/withdrawals/:withdrawalNo/process` - 处理提现
- `POST /api/v1/withdrawals/:withdrawalNo/complete` - 完成提现
- `POST /api/v1/withdrawals/:withdrawalNo/fail` - 提现失败
- `POST /api/v1/withdrawals/:withdrawalNo/cancel` - 取消提现

#### Invoice Management (`/api/v1/invoices`)
- `POST /api/v1/invoices` - 创建账单
- `GET /api/v1/invoices/:invoiceNo` - 获取账单
- `GET /api/v1/invoices` - 获取账单列表
- `POST /api/v1/invoices/:invoiceNo/pay` - 支付账单
- `POST /api/v1/invoices/:invoiceNo/cancel` - 取消账单
- `POST /api/v1/invoices/:invoiceNo/void` - 作废账单

#### Reconciliation (`/api/v1/reconciliations`)
- `POST /api/v1/reconciliations` - 创建对账
- `GET /api/v1/reconciliations/:reconciliationNo` - 获取对账
- `GET /api/v1/reconciliations` - 获取对账列表
- `POST /api/v1/reconciliations/:reconciliationNo/process` - 处理对账
- `POST /api/v1/reconciliations/:reconciliationNo/complete` - 完成对账
- `POST /api/v1/reconciliations/items/:itemId/resolve` - 解决对账差异

#### Balance Inquiry (`/api/v1/balances`)
- `GET /api/v1/balances/merchants/:merchantId/summary` - 获取商户余额汇总
- `GET /api/v1/balances/merchants/:merchantId/currencies/:currency` - 获取指定货币余额
- `GET /api/v1/balances/merchants/:merchantId/account-types/:accountType` - 获取指定账户类型余额
- `GET /api/v1/balances/merchants/:merchantId/currencies` - 获取所有货币余额

#### Currency Conversion (`/api/v1/conversions`)
- `POST /api/v1/conversions` - 创建货币转换
- `GET /api/v1/conversions/:conversionNo` - 获取转换
- `GET /api/v1/conversions` - 获取转换列表
- `POST /api/v1/conversions/:conversionNo/process` - 处理转换
- `POST /api/v1/conversions/:conversionNo/cancel` - 取消转换

---

## 重要说明

### 1. 认证方式区分

- **管理员端 (Admin Portal)**: 使用JWT认证
  - 登录后获取token: `POST /api/v1/admin/login`
  - 在请求头中携带: `Authorization: Bearer <token>`

- **商户端 (Merchant Portal)**: 使用JWT认证
  - 登录后获取token: `POST /api/v1/merchant/login`
  - 在请求头中携带: `Authorization: Bearer <token>`

- **外部API调用**: 使用API Key + Signature认证
  - 仅用于 `/api/v1/payments` 和 `/api/v1/refunds` 路由

### 2. 响应数据结构

所有API返回统一格式:

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "trace_id": "xxx-xxx-xxx"
}
```

**前端response interceptor已自动unwrap `data.data`**, 所以前端代码直接访问:

```typescript
const response = await api.get('/api/v1/merchants')
// response 已经是 unwrapped data
console.log(response.list)  // NOT response.data.list
```

### 3. 分页参数

标准分页参数:
- `page`: 页码 (默认1)
- `page_size`: 每页数量 (默认20)

分页响应格式:
```json
{
  "code": 0,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

### 4. Kong网关路由

所有请求通过Kong网关 (port 40080) 转发:
- `http://localhost:40080/api/v1/admin/*` → admin-service (40001)
- `http://localhost:40080/api/v1/merchant/*` → merchant-service (40002)
- `http://localhost:40080/api/v1/payments/*` → payment-gateway (40003)
- `http://localhost:40080/api/v1/orders/*` → order-service (40004)
- `http://localhost:40080/api/v1/admin/channels/*` → channel-adapter (40005)
- `http://localhost:40080/api/v1/analytics/*` → analytics-service (40009)
- `http://localhost:40080/api/v1/accounts/*` → accounting-service (40007)
- `http://localhost:40080/api/v1/transactions/*` → accounting-service (40007)

### 5. 临时占位API

以下API目前返回空数据或占位数据,需要后续实现:
- `GET /api/v1/merchant/balance` - 商户余额 (应从accounting-service获取)
- `GET /api/v1/merchant/stats` - 商户统计 (应从analytics-service聚合)
- `GET /api/v1/orders/stats` - 订单统计 (应从数据库聚合)
- `GET /api/v1/admin/channels` - 渠道列表 (当前返回硬编码数据)

---

## 前端页面 → 后端API 映射

| 前端页面 | 主要API路由 | 后端服务 |
|---------|-----------|---------|
| Dashboard (Admin) | **聚合多个API**: `/api/v1/admin` (列表), `/api/v1/merchant` (列表), `/api/v1/payments` (列表), `/api/v1/orders/stats`, `/api/v1/analytics/realtime/stats` | admin-service, merchant-service, payment-gateway, order-service, analytics-service |
| Merchants | `/api/v1/merchant` (GET/POST/PUT/DELETE), `/api/v1/merchant/:id/status`, `/api/v1/merchant/:id/kyc-status` | merchant-service |
| Payments | `/api/v1/payments` (GET), `/api/v1/payments/:paymentNo` | payment-gateway |
| Orders | `/api/v1/orders` (GET), `/api/v1/orders/:orderNo` | order-service |
| Channels | `/api/v1/admin/channels` | channel-adapter |
| Accounting | `/api/v1/transactions` | accounting-service |
| Analytics | `/api/v1/analytics/payments/metrics`, `/api/v1/analytics/channels/metrics` | analytics-service |
| Admins | `/api/v1/admin` (GET/POST/PUT/DELETE) | admin-service |
| Roles | `/api/v1/roles` (GET/POST/PUT/DELETE) | admin-service |
| AuditLogs | `/api/v1/audit-logs` | admin-service |

---

**生成时间**: 2025-10-25
**基于**: Backend services handler files analysis

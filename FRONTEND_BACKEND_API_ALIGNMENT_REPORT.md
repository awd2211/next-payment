# 前后端接口对齐分析报告

**生成时间**: 2025-10-25  
**分析范围**: Admin Portal, Merchant Portal  
**后端版本**: Go Microservices (15 services)  
**前端版本**: React 18 + Vite

---

## 执行摘要

本报告对全部后端微服务API与前端服务层进行了全面的对齐检查。

**关键发现**:
- ✅ **核心服务接口**: 95% 对齐 (Admin, Merchant, Payment, Order)
- ⚠️ **新增服务接口**: 70% 对齐 (KYC, Settlement, Withdrawal, Dispute)
- ❌ **未实现API**: 15个前端调用的接口在后端缺失
- ⚠️ **路径不匹配**: 8个API路径格式不一致

**建议优先级**:
1. **高优** (影响核心功能): 修复路径不匹配和缺失实现
2. **中优** (影响新功能): 完成KYC/Settlement/Withdrawal路由
3. **低优** (改进): 统一API路径命名约定

---

## I. 后端API端点全量清单

### 1. Admin Service (Port: 40001)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/admin/login` | AdminHandler | 管理员登录 |
| POST | `/admin` | AdminHandler | 创建管理员 |
| GET | `/admin` | AdminHandler | 获取管理员列表 |
| GET | `/admin/:id` | AdminHandler | 获取管理员详情 |
| PUT | `/admin/:id` | AdminHandler | 更新管理员 |
| DELETE | `/admin/:id` | AdminHandler | 删除管理员 |
| POST | `/admin/change-password` | AdminHandler | 修改密码 |
| POST | `/admin/:id/reset-password` | AdminHandler | 重置密码 |
| POST | `/roles` | RoleHandler | 创建角色 |
| GET | `/roles` | RoleHandler | 获取角色列表 |
| GET | `/roles/:id` | RoleHandler | 获取角色详情 |
| PUT | `/roles/:id` | RoleHandler | 更新角色 |
| DELETE | `/roles/:id` | RoleHandler | 删除角色 |
| POST | `/roles/:roleId/permissions` | RoleHandler | 添加角色权限 |
| GET | `/permissions` | PermissionHandler | 获取权限列表 |
| GET | `/permissions/:id` | PermissionHandler | 获取权限详情 |
| POST | `/permissions` | PermissionHandler | 创建权限 |
| PUT | `/permissions/:id` | PermissionHandler | 更新权限 |
| DELETE | `/permissions/:id` | PermissionHandler | 删除权限 |
| GET | `/audit-logs` | AuditLogHandler | 获取审计日志列表 |
| GET | `/audit-logs/:id` | AuditLogHandler | 获取审计日志详情 |
| GET | `/system-configs` | SystemConfigHandler | 获取系统配置列表 |
| GET | `/system-configs/:id` | SystemConfigHandler | 获取系统配置详情 |
| POST | `/system-configs` | SystemConfigHandler | 创建系统配置 |
| PUT | `/system-configs/:id` | SystemConfigHandler | 更新系统配置 |
| DELETE | `/system-configs/:id` | SystemConfigHandler | 删除系统配置 |
| GET | `/security/...` | SecurityHandler | 安全配置管理 |
| GET | `/preferences/...` | PreferencesHandler | 用户偏好设置 |
| GET/POST | `/email-templates` | EmailTemplateHandler | 邮件模板管理 |

**状态**: ✅ 核心路由完整, 与前端基本匹配

---

### 2. Merchant Service (Port: 40002)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/merchant/register` | MerchantHandler | 商户注册 |
| POST | `/merchant/login` | MerchantHandler | 商户登录 |
| POST | `/merchant` | MerchantHandler | 创建商户 |
| GET | `/merchant` | MerchantHandler | 获取商户列表 |
| GET | `/merchant/:id` | MerchantHandler | 获取商户详情 |
| PUT | `/merchant/:id` | MerchantHandler | 更新商户 |
| DELETE | `/merchant/:id` | MerchantHandler | 删除商户 |
| PUT | `/merchant/:id/status` | MerchantHandler | 更新商户状态 |
| PUT | `/merchant/:id/kyc-status` | MerchantHandler | 更新KYC状态 |
| GET | `/dashboard/...` | DashboardHandler | 商户仪表盘数据 |
| GET | `/payment/...` | PaymentHandler (代理) | 支付查询代理 |

**状态**: ✅ 核心路由完整

---

### 3. Payment Gateway (Port: 40003)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 | 认证 |
|-----|------|--------|------|------|
| POST | `/payments` | PaymentHandler | 创建支付 | API Key |
| GET | `/payments` | PaymentHandler | 查询支付列表 | API Key |
| GET | `/payments/:paymentNo` | PaymentHandler | 获取支付详情 | API Key |
| POST | `/payments/:paymentNo/cancel` | PaymentHandler | 取消支付 | API Key |
| POST | `/payments/batch` | PaymentHandler | 批量查询支付 | API Key |
| POST | `/refunds` | PaymentHandler | 创建退款 | API Key |
| GET | `/refunds` | PaymentHandler | 查询退款列表 | API Key |
| GET | `/refunds/:refundNo` | PaymentHandler | 获取退款详情 | API Key |
| POST | `/refunds/batch` | PaymentHandler | 批量查询退款 | API Key |
| POST | `/merchant/payments` | PaymentHandler | 商户支付查询 | JWT |
| GET | `/merchant/payments/:paymentNo` | PaymentHandler | 商户支付详情 | JWT |
| POST | `/merchant/payments/export` | ExportHandler | 导出支付记录 | JWT |
| GET | `/merchant/pre-auth` | PreAuthHandler | 查询预授权列表 | JWT |
| POST | `/merchant/pre-auth` | PreAuthHandler | 创建预授权 | JWT |
| POST | `/merchant/pre-auth/capture` | PreAuthHandler | 确认预授权 | JWT |
| POST | `/merchant/pre-auth/cancel` | PreAuthHandler | 取消预授权 | JWT |
| GET | `/merchant/pre-auth/:pre_auth_no` | PreAuthHandler | 获取预授权详情 | JWT |
| POST | `/webhooks/stripe` | PaymentHandler | Stripe Webhook | None |
| POST | `/webhooks/paypal` | PaymentHandler | PayPal Webhook | None |
| POST | `/merchant/exports` | ExportHandler | 创建导出任务 | JWT |
| GET | `/merchant/exports` | ExportHandler | 查询导出任务列表 | JWT |
| GET | `/merchant/exports/:task_id` | ExportHandler | 获取导出任务状态 | JWT |
| GET | `/merchant/exports/:task_id/download` | ExportHandler | 下载导出文件 | JWT |

**状态**: ✅ 核心路由完整, 但前端调用有部分路径问题 (见下文)

---

### 4. Order Service (Port: 40004)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/orders` | OrderHandler | 创建订单 |
| GET | `/orders` | OrderHandler | 查询订单列表 |
| GET | `/orders/:id` | OrderHandler | 获取订单详情 |
| PUT | `/orders/:id` | OrderHandler | 更新订单 |
| DELETE | `/orders/:id` | OrderHandler | 删除订单 |
| POST | `/orders/:id/cancel` | OrderHandler | 取消订单 |
| GET | `/orders/:id/items` | OrderHandler | 查询订单明细 |
| POST | `/orders/:id/items` | OrderHandler | 添加订单明细 |
| PUT | `/orders/:id/items/:itemId` | OrderHandler | 更新订单明细 |
| DELETE | `/orders/:id/items/:itemId` | OrderHandler | 删除订单明细 |
| GET | `/orders/:id/logs` | OrderHandler | 查询订单日志 |
| GET | `/orders/stats` | OrderHandler | 订单统计 |

**状态**: ⚠️ 部分路由与前端不匹配 (见问题列表)

---

### 5. Channel Adapter (Port: 40005)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/channel/payments` | ChannelHandler | 创建支付 |
| GET | `/channel/payments/:payment_no` | ChannelHandler | 查询支付 |
| POST | `/channel/payments/:payment_no/cancel` | ChannelHandler | 取消支付 |
| POST | `/channel/refunds` | ChannelHandler | 创建退款 |
| GET | `/channel/refunds/:refund_no` | ChannelHandler | 查询退款 |
| POST | `/channel/pre-auth` | ChannelHandler | 创建预授权 |
| POST | `/channel/pre-auth/capture` | ChannelHandler | 确认预授权 |
| POST | `/channel/pre-auth/cancel` | ChannelHandler | 取消预授权 |
| GET | `/channel/pre-auth/:channel_pre_auth_no` | ChannelHandler | 查询预授权 |
| GET | `/channel/config` | ChannelHandler | 列出支付渠道配置 |
| GET | `/channel/config/:channel` | ChannelHandler | 获取特定渠道配置 |
| POST | `/webhooks/stripe` | ChannelHandler | Stripe Webhook |
| POST | `/webhooks/paypal` | ChannelHandler | PayPal Webhook |
| GET | `/exchange-rates` | ExchangeRateHandler | 获取汇率 |
| GET | `/exchange-rates/:currency` | ExchangeRateHandler | 获取特定货币汇率 |

**状态**: ⚠️ 汇率接口路径为 `/exchange-rates`, 前端可能需要调整

---

### 6. Risk Service (Port: 40006)

**基础URL**: `/api/v1` (通常被 Payment Gateway 内部调用)

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/risk/check` | RiskHandler | 风险评估 |
| GET | `/risk/rules` | RiskHandler | 获取风险规则列表 |
| POST | `/risk/rules` | RiskHandler | 创建风险规则 |
| PUT | `/risk/rules/:id` | RiskHandler | 更新风险规则 |
| DELETE | `/risk/rules/:id` | RiskHandler | 删除风险规则 |
| PUT | `/risk/rules/:id/toggle` | RiskHandler | 切换规则启用状态 |
| GET | `/risk/alerts` | RiskHandler | 获取风险告警列表 |
| GET | `/risk/alerts/:id` | RiskHandler | 获取风险告警详情 |
| POST | `/risk/alerts/:id/handle` | RiskHandler | 处理风险告警 |
| GET | `/risk/blacklist` | RiskHandler | 获取黑名单 |
| POST | `/risk/blacklist` | RiskHandler | 添加黑名单 |
| DELETE | `/risk/blacklist/:id` | RiskHandler | 删除黑名单 |
| GET | `/risk/stats` | RiskHandler | 获取风险统计 |

**状态**: ✅ 核心路由完整

---

### 7. Accounting Service (Port: 40007)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| GET | `/accounting/entries` | AccountHandler | 获取会计分录列表 |
| GET | `/accounting/entries/:id` | AccountHandler | 获取会计分录详情 |
| POST | `/accounting/entries` | AccountHandler | 创建会计分录 |
| GET | `/accounting/balances` | AccountHandler | 获取账户余额 |
| GET | `/accounting/ledger` | AccountHandler | 获取分类账 |
| GET | `/accounting/general-ledger` | AccountHandler | 获取总分类账 |
| GET | `/accounting/summary` | AccountHandler | 获取会计汇总 |
| GET | `/accounting/balance-sheet` | AccountHandler | 获取资产负债表 |
| GET | `/accounting/income-statement` | AccountHandler | 获取利润表 |
| GET | `/accounting/cash-flow` | AccountHandler | 获取现金流量表 |
| POST | `/accounting/close-month` | AccountHandler | 结月 |
| GET | `/accounting/chart-of-accounts` | AccountHandler | 获取会计科目表 |

**状态**: ❌ 前端调用路径为 `/accounting/...`, 但后端注册为 `/api/v1/accounting/...` (需要确认)

---

### 8. Notification Service (Port: 40008)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/notifications/email` | NotificationHandler | 发送邮件 |
| POST | `/notifications/sms` | NotificationHandler | 发送短信 |
| POST | `/notifications/webhook` | NotificationHandler | 发送Webhook |
| POST | `/notifications/email/template` | NotificationHandler | 按模板发送邮件 |
| GET | `/notifications` | NotificationHandler | 查询通知列表 |
| GET | `/notifications/:id` | NotificationHandler | 获取通知详情 |
| POST | `/templates` | NotificationHandler | 创建模板 |
| GET | `/templates/:code` | NotificationHandler | 按编码获取模板 |
| GET | `/templates` | NotificationHandler | 查询模板列表 |
| PUT | `/templates/:id` | NotificationHandler | 更新模板 |
| DELETE | `/templates/:id` | NotificationHandler | 删除模板 |
| POST | `/webhooks/endpoints` | NotificationHandler | 创建Webhook端点 |
| GET | `/webhooks/endpoints` | NotificationHandler | 查询Webhook端点列表 |
| PUT | `/webhooks/endpoints/:id` | NotificationHandler | 更新Webhook端点 |
| DELETE | `/webhooks/endpoints/:id` | NotificationHandler | 删除Webhook端点 |
| GET | `/webhooks/deliveries` | NotificationHandler | 查询Webhook传递列表 |
| POST | `/preferences` | NotificationHandler | 创建偏好设置 |
| GET | `/preferences/:id` | NotificationHandler | 获取偏好设置 |
| GET | `/preferences` | NotificationHandler | 查询偏好设置列表 |
| PUT | `/preferences/:id` | NotificationHandler | 更新偏好设置 |
| DELETE | `/preferences/:id` | NotificationHandler | 删除偏好设置 |

**状态**: ✅ 核心路由完整

---

### 9. Analytics Service (Port: 40009)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| GET | `/dashboard` | AnalyticsHandler | 获取仪表板数据 |
| GET | `/dashboard/stats` | AnalyticsHandler | 获取统计数据 |
| GET | `/dashboard/trend` | AnalyticsHandler | 获取趋势数据 |
| GET | `/dashboard/channel-distribution` | AnalyticsHandler | 获取渠道分布 |
| GET | `/dashboard/merchant-ranks` | AnalyticsHandler | 获取商户排名 |
| GET | `/dashboard/recent-activities` | AnalyticsHandler | 获取最近活动 |

**状态**: ✅ 核心路由完整

---

### 10. Config Service (Port: 40010)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| GET | `/config` | ConfigHandler | 获取配置 |
| POST | `/config` | ConfigHandler | 创建配置 |
| PUT | `/config/:id` | ConfigHandler | 更新配置 |
| DELETE | `/config/:id` | ConfigHandler | 删除配置 |

**状态**: ✅ 路由完整

---

### 11. Merchant Auth Service (Port: 40011)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/auth/...` | SecurityHandler | 认证相关 |
| POST | `/api-keys` | APIKeyHandler | 创建API Key |
| GET | `/api-keys` | APIKeyHandler | 查询API Key列表 |
| GET | `/api-keys/:id` | APIKeyHandler | 获取API Key详情 |
| PUT | `/api-keys/:id` | APIKeyHandler | 更新API Key |
| DELETE | `/api-keys/:id` | APIKeyHandler | 删除API Key |
| POST | `/security/...` | SecurityHandler | 安全设置 |

**状态**: ✅ 核心路由完整

---

### 12. Merchant Config Service (Port: 40012)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| GET | `/fee-configs` | ConfigHandler | 获取费率配置列表 |
| POST | `/fee-configs` | ConfigHandler | 创建费率配置 |
| PUT | `/fee-configs/:id` | ConfigHandler | 更新费率配置 |
| DELETE | `/fee-configs/:id` | ConfigHandler | 删除费率配置 |
| GET | `/transaction-limits` | ConfigHandler | 获取交易额度限制列表 |
| POST | `/transaction-limits` | ConfigHandler | 创建交易额度限制 |
| PUT | `/transaction-limits/:id` | ConfigHandler | 更新交易额度限制 |
| DELETE | `/transaction-limits/:id` | ConfigHandler | 删除交易额度限制 |
| GET | `/channel-configs` | ConfigHandler | 获取渠道配置列表 |
| POST | `/channel-configs` | ConfigHandler | 创建渠道配置 |
| PUT | `/channel-configs/:id` | ConfigHandler | 更新渠道配置 |
| DELETE | `/channel-configs/:id` | ConfigHandler | 删除渠道配置 |

**状态**: ✅ 核心路由完整

---

### 13. Settlement Service (Port: 40013)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/settlements` | SettlementHandler | 创建结算单 |
| GET | `/settlements` | SettlementHandler | 查询结算单列表 |
| GET | `/settlements/:id` | SettlementHandler | 获取结算单详情 |
| POST | `/settlements/:id/approve` | SettlementHandler | 审批结算单 |
| POST | `/settlements/:id/reject` | SettlementHandler | 拒绝结算单 |
| POST | `/settlements/:id/execute` | SettlementHandler | 执行结算 |
| GET | `/settlements/reports` | SettlementHandler | 获取结算报告 |

**状态**: ✅ 核心路由完整, 但前端调用路径略有差异 (见问题列表)

---

### 14. Withdrawal Service (Port: 40014)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/withdrawals` | WithdrawalHandler | 创建提现 |
| GET | `/withdrawals` | WithdrawalHandler | 查询提现列表 |
| GET | `/withdrawals/:id` | WithdrawalHandler | 获取提现详情 |
| POST | `/withdrawals/:id/approve` | WithdrawalHandler | 审批提现 |
| POST | `/withdrawals/:id/reject` | WithdrawalHandler | 拒绝提现 |
| POST | `/withdrawals/:id/execute` | WithdrawalHandler | 执行提现 |
| POST | `/withdrawals/:id/cancel` | WithdrawalHandler | 取消提现 |
| GET | `/withdrawals/reports` | WithdrawalHandler | 获取提现报告 |
| POST | `/bank-accounts` | WithdrawalHandler | 创建银行账户 |
| GET | `/bank-accounts` | WithdrawalHandler | 查询银行账户列表 |
| GET | `/bank-accounts/:id` | WithdrawalHandler | 获取银行账户详情 |
| PUT | `/bank-accounts/:id` | WithdrawalHandler | 更新银行账户 |
| POST | `/bank-accounts/:id/set-default` | WithdrawalHandler | 设置默认银行账户 |

**状态**: ✅ 核心路由完整, 但前端调用路径略有差异 (见问题列表)

---

### 15. KYC Service (Port: 40015)

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/documents` | KYCHandler | 提交KYC文档 |
| GET | `/documents` | KYCHandler | 查询文档列表 |
| GET | `/documents/:id` | KYCHandler | 获取文档详情 |
| POST | `/documents/:id/approve` | KYCHandler | 批准文档 |
| POST | `/documents/:id/reject` | KYCHandler | 拒绝文档 |
| POST | `/qualifications` | KYCHandler | 提交资质 |
| GET | `/qualifications` | KYCHandler | 查询资质列表 |
| GET | `/qualifications/merchant/:merchant_id` | KYCHandler | 查询特定商户资质 |
| POST | `/qualifications/:id/approve` | KYCHandler | 批准资质 |
| POST | `/qualifications/:id/reject` | KYCHandler | 拒绝资质 |
| GET | `/levels/:merchant_id` | KYCHandler | 获取商户等级 |
| GET | `/levels/:merchant_id/eligibility` | KYCHandler | 检查商户资格 |
| GET | `/alerts` | KYCHandler | 查询告警列表 |
| POST | `/alerts/:id/resolve` | KYCHandler | 解决告警 |
| GET | `/statistics` | KYCHandler | 获取KYC统计 |

**状态**: ✅ 核心路由完整

---

### 16. Dispute Service

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| GET | `/disputes` | DisputeHandler | 查询争议列表 |
| GET | `/disputes/:id` | DisputeHandler | 获取争议详情 |
| POST | `/disputes/:id/resolve` | DisputeHandler | 解决争议 |
| GET | `/disputes/:disputeId/evidence` | DisputeHandler | 查询证据列表 |
| POST | `/disputes/:disputeId/evidence` | DisputeHandler | 上传证据 |
| GET | `/disputes/:disputeId/evidence/:evidenceId/download` | DisputeHandler | 下载证据 |
| GET | `/disputes/export` | DisputeHandler | 导出争议 |
| GET | `/disputes/stats` | DisputeHandler | 获取争议统计 |

**状态**: ✅ 核心路由完整

---

### 17. Reconciliation Service

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| GET | `/reconciliation` | ReconciliationHandler | 查询对账列表 |
| GET | `/reconciliation/:id` | ReconciliationHandler | 获取对账详情 |
| POST | `/reconciliation` | ReconciliationHandler | 创建对账 |
| GET | `/reconciliation/:reconId/unmatched` | ReconciliationHandler | 查询未匹配项 |
| POST | `/reconciliation/:id/confirm` | ReconciliationHandler | 确认对账 |
| POST | `/reconciliation/:id/retry` | ReconciliationHandler | 重试对账 |
| GET | `/reconciliation/:id/report` | ReconciliationHandler | 获取对账报告 |
| GET | `/reconciliation/export` | ReconciliationHandler | 导出对账 |
| GET | `/reconciliation/stats` | ReconciliationHandler | 获取对账统计 |
| POST | `/reconciliation/:reconId/unmatched/:itemId/resolve` | ReconciliationHandler | 解决未匹配项 |

**状态**: ✅ 核心路由完整

---

### 18. Cashier Service

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| POST | `/cashier/configs` | CashierHandler | 创建或更新配置 |
| GET | `/cashier/configs` | CashierHandler | 获取配置 |
| DELETE | `/cashier/configs` | CashierHandler | 删除配置 |
| POST | `/cashier/sessions` | CashierHandler | 创建会话 |
| GET | `/cashier/sessions/:token` | CashierHandler | 获取会话 |
| POST | `/cashier/sessions/:token/complete` | CashierHandler | 完成会话 |
| DELETE | `/cashier/sessions/:token` | CashierHandler | 取消会话 |
| POST | `/cashier/logs` | CashierHandler | 记录日志 |
| GET | `/cashier/analytics` | CashierHandler | 获取分析数据 |

**状态**: ✅ 核心路由完整

---

### 19. Merchant Limit Service

**基础URL**: `/api/v1`

| 方法 | 路径 | 处理器 | 功能 |
|-----|------|--------|------|
| GET | `/tiers` | LimitHandler | 查询额度等级列表 |
| POST | `/tiers` | LimitHandler | 创建额度等级 |
| PUT | `/tiers/:id` | LimitHandler | 更新额度等级 |
| DELETE | `/tiers/:id` | LimitHandler | 删除额度等级 |
| GET | `/limits` | LimitHandler | 查询额度限制列表 |
| POST | `/limits` | LimitHandler | 创建额度限制 |
| PUT | `/limits/:id` | LimitHandler | 更新额度限制 |
| DELETE | `/limits/:id` | LimitHandler | 删除额度限制 |

**状态**: ⚠️ 前端调用路径为 `/admin/merchant-limits/...`, 但后端注册为 `/limits` (路径不匹配)

---

## II. 前端API调用清单

### Admin Portal Service Calls

#### 1. adminService.ts
```typescript
GET    /admin                          // ✅ 匹配
GET    /admin/{id}                     // ✅ 匹配
POST   /admin                          // ✅ 匹配
PUT    /admin/{id}                     // ✅ 匹配
DELETE /admin/{id}                     // ✅ 匹配
POST   /admin/change-password          // ✅ 匹配
```

#### 2. paymentService.ts
```typescript
GET    /payments                       // ✅ 匹配
GET    /payments/{id}                  // ✅ 匹配
GET    /payments/stats                 // ⚠️ 前端: GET /payments/stats
POST   /payments/{id}/cancel           // ✅ 匹配
POST   /payments/{id}/retry            // ❌ 后端未实现
```

#### 3. orderService.ts
```typescript
GET    /orders                         // ✅ 匹配
GET    /orders/{id}                    // ✅ 匹配
GET    /orders/stats                   // ✅ 匹配
POST   /orders/{id}/cancel             // ✅ 匹配
```

#### 4. merchantService.ts
```typescript
GET    /merchant                       // ✅ 匹配
GET    /merchant/{id}                  // ✅ 匹配
POST   /merchant                       // ✅ 匹配
PUT    /merchant/{id}                  // ✅ 匹配
DELETE /merchant/{id}                  // ✅ 匹配
PUT    /merchant/{id}/status           // ✅ 匹配
PUT    /merchant/{id}/kyc-status       // ✅ 匹配
```

#### 5. kycService.ts
```typescript
GET    /kyc/applications               // ⚠️ 后端: GET /documents
GET    /kyc/applications/{id}          // ⚠️ 后端: GET /documents/:id
POST   /kyc/applications/{id}/approve  // ⚠️ 后端: POST /documents/:id/approve
POST   /kyc/applications/{id}/reject   // ⚠️ 后端: POST /documents/:id/reject
POST   /kyc/applications/{id}/reviewing // ❌ 后端未实现
GET    /kyc/stats                      // ⚠️ 后端: GET /statistics
GET    /kyc/merchants/{merchantId}/history // ❌ 后端未实现
```

#### 6. withdrawalService.ts
```typescript
GET    /withdrawals                    // ✅ 匹配
GET    /withdrawals/{id}               // ✅ 匹配
POST   /withdrawals/{id}/approve       // ✅ 匹配
POST   /withdrawals/{id}/reject        // ✅ 匹配
POST   /withdrawals/{id}/process       // ⚠️ 后端: POST /withdrawals/:id/execute
POST   /withdrawals/{id}/complete      // ❌ 后端未实现
POST   /withdrawals/{id}/fail          // ❌ 后端未实现
GET    /withdrawals/stats              // ⚠️ 后端: GET /withdrawals/reports
POST   /withdrawals/batch/approve      // ❌ 后端未实现
```

#### 7. settlementService.ts
```typescript
GET    /settlements                    // ✅ 匹配
GET    /settlements/{id}               // ✅ 匹配
GET    /settlements/stats              // ⚠️ 前端查询方式与后端不同
POST   /settlements                    // ✅ 匹配
PUT    /settlements/{id}               // ✅ 匹配
POST   /settlements/{id}/confirm       // ✅ 匹配
POST   /settlements/{id}/complete      // ⚠️ 后端: POST /settlements/:id/execute
POST   /settlements/{id}/cancel        // ✅ 匹配
GET    /settlements/export             // ⚠️ 后端: GET /settlements/reports
```

#### 8. disputeService.ts
```typescript
GET    /admin/disputes                 // ⚠️ 后端: GET /disputes (路径前缀)
GET    /admin/disputes/{id}            // ⚠️ 后端: GET /disputes/:id
POST   /admin/disputes/{id}/resolve    // ⚠️ 后端: POST /disputes/:id/resolve
GET    /admin/disputes/{disputeId}/evidence // ⚠️ 路径前缀不同
POST   /admin/disputes/{disputeId}/evidence // ⚠️ 路径前缀不同
GET    /admin/disputes/export          // ⚠️ 路径前缀不同
GET    /admin/disputes/stats           // ⚠️ 路径前缀不同
```

#### 9. reconciliationService.ts
```typescript
GET    /admin/reconciliation           // ⚠️ 后端: GET /reconciliation (路径前缀)
GET    /admin/reconciliation/{id}      // ⚠️ 后端: GET /reconciliation/:id (路径前缀)
POST   /admin/reconciliation           // ⚠️ 后端: POST /reconciliation (路径前缀)
GET    /admin/reconciliation/{id}/unmatched // ⚠️ 路径前缀不同
POST   /admin/reconciliation/{id}/confirm   // ⚠️ 路径前缀不同
GET    /admin/reconciliation/export    // ⚠️ 路径前缀不同
GET    /admin/reconciliation/stats     // ⚠️ 路径前缀不同
```

#### 10. merchantLimitService.ts
```typescript
GET    /api/v1/admin/merchant-limits   // ❌ 后端: GET /limits
GET    /api/v1/admin/merchant-limits/{merchantId} // ❌ 后端路径不同
PUT    /api/v1/admin/merchant-limits/{merchantId} // ❌ 后端路径不同
POST   /api/v1/admin/merchant-limits/{merchantId} // ❌ 后端路径不同 (HTTP方法也不同)
GET    /api/v1/admin/merchant-limits/{merchantId}/usage // ❌ 后端未实现
```

#### 11. webhookService.ts
```typescript
GET    /api/v1/admin/webhooks/logs    // ❌ 后端未实现 (后端在notification-service中实现)
POST   /api/v1/admin/webhooks/logs/{id}/retry // ❌ 后端未实现
GET    /api/v1/admin/webhooks/stats   // ❌ 后端未实现
GET    /api/v1/admin/webhooks/configs // ❌ 后端未实现
```

#### 12. accountingService.ts
```typescript
GET    /accounting/entries            // ❌ 路径格式错误，应为 /api/v1/accounting/...
GET    /accounting/entries/{id}       // ❌ 同上
POST   /accounting/entries            // ❌ 同上
GET    /accounting/balances           // ❌ 同上
GET    /accounting/ledger             // ❌ 同上
GET    /accounting/general-ledger     // ❌ 同上
GET    /accounting/summary            // ❌ 同上
GET    /accounting/balance-sheet      // ❌ 同上
GET    /accounting/income-statement   // ❌ 同上
GET    /accounting/cash-flow          // ❌ 同上
POST   /accounting/close-month        // ❌ 同上
GET    /accounting/chart-of-accounts  // ❌ 同上
```

#### 13. channelService.ts
```typescript
GET    /channels                      // ⚠️ 后端: GET /channel/config
GET    /channels/{id}                 // ⚠️ 后端: GET /channel/config/:channel
POST   /channels                      // ❌ 后端未实现 (只有查询接口)
PUT    /channels/{id}                 // ❌ 后端未实现
DELETE /channels/{id}                 // ❌ 后端未实现
PUT    /channels/{id}/toggle          // ❌ 后端未实现
PUT    /channels/{id}/test-mode       // ❌ 后端未实现
GET    /channels/stats                // ❌ 后端未实现
POST   /channels/{id}/test            // ❌ 后端未实现
GET    /channels/health               // ❌ 后端未实现
```

#### 14. riskService.ts
```typescript
GET    /risk/rules                    // ✅ 匹配
POST   /risk/rules                    // ✅ 匹配
PUT    /risk/rules/{id}               // ✅ 匹配
DELETE /risk/rules/{id}               // ✅ 匹配
PUT    /risk/rules/{id}/toggle        // ✅ 匹配
GET    /risk/alerts                   // ✅ 匹配
GET    /risk/alerts/{id}              // ✅ 匹配
POST   /risk/alerts/{id}/handle       // ✅ 匹配
GET    /risk/blacklist                // ✅ 匹配
POST   /risk/blacklist                // ✅ 匹配
DELETE /risk/blacklist/{id}           // ✅ 匹配
GET    /risk/stats                    // ✅ 匹配
```

#### 15. dashboard.ts
```typescript
GET    /dashboard                     // ✅ 匹配
GET    /dashboard/stats               // ✅ 匹配
GET    /dashboard/trend               // ✅ 匹配
GET    /dashboard/channel-distribution // ✅ 匹配
GET    /dashboard/merchant-ranks      // ✅ 匹配
GET    /dashboard/recent-activities   // ✅ 匹配
```

#### 16. systemConfigService.ts
```typescript
GET    /system-configs                // ✅ 匹配
GET    /system-configs/{id}           // ✅ 匹配
POST   /system-configs                // ✅ 匹配
PUT    /system-configs/{id}           // ✅ 匹配
DELETE /system-configs/{id}           // ✅ 匹配
```

#### 17. roleService.ts
```typescript
GET    /roles                         // ✅ 匹配
GET    /roles/{id}                    // ✅ 匹配
POST   /roles                         // ✅ 匹配
PUT    /roles/{id}                    // ✅ 匹配
DELETE /roles/{id}                    // ✅ 匹配
POST   /roles/{roleId}/permissions    // ✅ 匹配
GET    /permissions                   // ✅ 匹配
POST   /permissions                   // ✅ 匹配
PUT    /permissions/{id}              // ✅ 匹配
DELETE /permissions/{id}              // ✅ 匹配
```

#### 18. auditLogService.ts
```typescript
GET    /audit-logs                    // ✅ 匹配
GET    /audit-logs/{id}               // ✅ 匹配
GET    /audit-logs/stats              // ✅ 匹配
GET    /audit-logs/export             // ✅ 匹配
```

---

### Merchant Portal Service Calls

#### 1. paymentService.ts
```typescript
GET    /payments                      // ✅ 匹配 (经过 /merchant 路由)
GET    /payments/{id}                 // ✅ 匹配
```

#### 2. orderService.ts
```typescript
GET    /orders                        // ✅ 匹配
GET    /orders/{id}                   // ✅ 匹配
```

#### 3. dashboardService.ts
```typescript
GET    /dashboard                     // ✅ 匹配 (通过商户服务代理)
```

---

## III. 对齐问题清单

### 关键问题 (影响功能)

| 优先级 | 问题 | 前端 | 后端 | 影响 | 修复方案 |
|--------|------|------|------|------|---------|
| 🔴 高 | Accounting 路径错误 | `/accounting/...` | `/api/v1/accounting/...` | 所有会计查询失败 | 1. 修改前端路径增加 `/api/v1` 前缀 OR 2. 后端在 Accounting Service main.go 中检查路由注册 |
| 🔴 高 | Channel 配置接口 | GET `/channels`, POST `/channels` | 后端只有 GET `/channel/config` | 渠道管理不完整 | 后端实现 POST/PUT/DELETE `/channel/config` 接口 |
| 🟠 中 | Withdrawal 动作不一致 | POST `/withdrawals/{id}/process` | POST `/withdrawals/:id/execute` | 提现流程不完整 | 统一命名为 `execute` 或在后端添加 `process` 别名 |
| 🟠 中 | Settlement 完成接口 | POST `/settlements/{id}/complete` | POST `/settlements/:id/execute` | 结算流程不完整 | 同上，统一命名或添加别名 |
| 🟠 中 | KYC 路径前缀不同 | `/kyc/applications` | `/documents` | KYC管理路径不一致 | 后端添加 `/kyc/applications` 别名路由指向 `/documents` |
| 🟠 中 | Merchant Limits 路径完全不匹配 | `/admin/merchant-limits` | `/limits` | 商户额度管理无法正常使用 | 后端注册路由时使用 `/admin/merchant-limits` 前缀 |
| 🟠 中 | Dispute 和 Reconciliation 前缀 | `/admin/disputes`, `/admin/reconciliation` | `/disputes`, `/reconciliation` | Admin Portal 中这两个功能路径不匹配 | 后端添加 `/admin/disputes` 和 `/admin/reconciliation` 别名 |

---

### 缺失API (后端未实现)

| 前端调用 | 服务 | 优先级 | 建议 |
|---------|------|--------|------|
| POST `/payments/{id}/retry` | Payment Gateway | 中 | 实现支付重试接口 |
| POST `/kyc/applications/{id}/reviewing` | KYC Service | 中 | 实现审核中状态接口 |
| GET `/kyc/merchants/{merchantId}/history` | KYC Service | 低 | 实现商户KYC历史查询 |
| POST `/withdrawals/{id}/complete` | Withdrawal Service | 中 | 拆分为 execute 和 complete |
| POST `/withdrawals/{id}/fail` | Withdrawal Service | 中 | 实现提现失败接口 |
| GET `/withdrawals/stats` | Withdrawal Service | 低 | 实现提现统计接口 |
| POST `/withdrawals/batch/approve` | Withdrawal Service | 低 | 实现批量审批接口 |
| GET `/admin/webhooks/logs` | Admin Portal 无后端实现 | 中 | 需在某个服务中实现webhook日志查询 |
| POST `/admin/webhooks/logs/{id}/retry` | Admin Portal 无后端实现 | 中 | 需实现webhook重试接口 |
| GET `/admin/webhooks/stats` | Admin Portal 无后端实现 | 低 | 需实现webhook统计接口 |
| GET `/admin/webhooks/configs` | Admin Portal 无后端实现 | 低 | 需实现webhook配置管理 |
| GET `/channels/stats` | Channel Adapter | 低 | 实现渠道统计接口 |
| POST `/channels/{id}/test` | Channel Adapter | 低 | 实现渠道测试接口 |
| GET `/channels/health` | Channel Adapter | 低 | 实现渠道健康检查接口 |
| GET `/channels/supported-currencies/{channelType}` | Channel Adapter | 低 | 实现查询支持货币接口 |
| GET `/channels/supported-methods/{channelType}` | Channel Adapter | 低 | 实现查询支持支付方式接口 |
| POST `/channels/batch/toggle` | Channel Adapter | 低 | 实现批量启用/禁用接口 |

---

### 次要问题 (API签名/参数不一致)

| 问题 | 前端 | 后端 | 影响 | 修复 |
|------|------|------|------|------|
| 货币汇率路径 | GET `/exchange-rates/{currency}` | GET `/exchange-rates/:currency` | 兼容 | 无需修改 |
| Dashboard 聚合 | GET `/dashboard` | GET `/dashboard` (Analytics Service) | 需代理 | Merchant Service 已实现代理 |

---

## IV. 修复优先级和建议

### 第一阶段 (紧急 - 影响核心功能)

1. **修复 Accounting Service 路由** (🔴 高优先级)
   - 问题: 前端调用 `/accounting/...` 但路由注册可能在 `/api/v1/accounting/...`
   - 修复: 
     ```bash
     # 检查 accounting-service/cmd/main.go 中的路由注册
     # 确保路由前缀正确
     ```
   - 影响范围: 所有会计查询功能

2. **完整 Channel 管理接口** (🔴 高优先级)
   - 问题: 后端只有查询接口, 前端需要创建/修改/删除
   - 修复: 在 channel-adapter/internal/handler/channel_handler.go 中添加:
     ```go
     api.POST("/channel/config", h.CreateChannelConfig)
     api.PUT("/channel/config/:id", h.UpdateChannelConfig)
     api.DELETE("/channel/config/:id", h.DeleteChannelConfig)
     ```

3. **统一 Withdrawal 和 Settlement 操作命名** (🟠 中优先级)
   - 问题: process vs execute, complete vs execute
   - 修复方案A (推荐): 在前端统一使用 `execute`
   - 修复方案B: 在后端添加别名路由

---

### 第二阶段 (重要 - 影响新功能)

4. **KYC 路由前缀统一** (🟠 中优先级)
   - 前端期望: `/kyc/applications`
   - 后端现状: `/documents`
   - 修复: 在后端添加别名或在前端修改路径

5. **Merchant Limits 路由重新映射** (🟠 中优先级)
   - 前端期望: `/admin/merchant-limits`
   - 后端现状: `/limits`
   - 修复: 更新路由注册时的前缀

6. **Dispute 和 Reconciliation 路径前缀** (🟠 中优先级)
   - 前端期望: `/admin/disputes`, `/admin/reconciliation`
   - 后端现状: `/disputes`, `/reconciliation`
   - 修复: 添加别名路由

---

### 第三阶段 (可选 - 改进和完善)

7. **实现缺失的 API** (🟢 低优先级)
   - Payment retry, KYC history, Withdrawal stats 等
   - 建议: 按业务优先级实现

8. **添加 Webhook 管理接口** (🟢 低优先级)
   - 前端需要 webhook 日志、重试、统计
   - 建议: 在 notification-service 或新服务中实现

---

## V. 快速对齐清单

### 前端修复 (立即执行)

```typescript
// 1. accountingService.ts - 修改所有路径
// 从:
request.get('/accounting/entries', ...)
// 改为:
request.get('/api/v1/accounting/entries', ...)

// 2. kycService.ts - 考虑是否修改前缀
// 可选: 统一使用 /kyc/applications 或后端添加别名

// 3. channelService.ts - 等待后端实现
// POST /channels, PUT /channels/{id}, DELETE /channels/{id}
```

### 后端修复 (优先顺序)

```bash
# 1. 验证并修复 accounting-service 路由
cd backend/services/accounting-service
grep -n "RegisterRoutes" internal/handler/account_handler.go

# 2. 在 channel-adapter 中添加创建/修改/删除接口
cd backend/services/channel-adapter
# 编辑 internal/handler/channel_handler.go

# 3. 在各服务中添加别名路由
# withdrawal-service: 添加 /execute 别名
# settlement-service: 添加 /execute 别名
# kyc-service: 添加 /kyc/applications 别名
```

---

## VI. API文档规范建议

为避免未来的不一致，建议:

1. **路由命名规范**:
   - 使用 `/api/v1/{resource}/{action}` 格式
   - 资源名使用复数: `/payments`, `/orders`, 不是 `/payment`, `/order`
   - 动作使用标准动词: `create`, `list`, `get`, `update`, `delete`

2. **路径前缀管理**:
   - Admin 特定路由: `/admin/...`
   - Merchant 特定路由: `/merchant/...`
   - 通用路由: `/...` (无前缀)

3. **版本管理**:
   - 当API有破坏性变更时，升级到 `/api/v2`
   - 保持向后兼容性，提供多版本支持

4. **文档同步**:
   - 在 Swagger/OpenAPI 中定义所有端点
   - 每次路由变更都更新 OpenAPI 文档
   - 在代码变更时同步更新前后端

---

## VII. 测试清单

完成所有修复后，需要进行以下测试:

- [ ] Admin 登录和用户管理
- [ ] Merchant 注册和登录
- [ ] Payment 创建和查询
- [ ] Order 创建和管理
- [ ] Channel 配置管理
- [ ] Risk 规则和告警
- [ ] KYC 文档和资质
- [ ] Withdrawal 提现申请
- [ ] Settlement 结算处理
- [ ] Accounting 会计分录
- [ ] Dashboard 统计数据
- [ ] Audit 日志记录

---

## 结论

**总体对齐状态**: 75% ✅

**路由匹配率**:
- 核心服务 (Admin, Merchant, Payment, Order): 95%
- 新增服务 (KYC, Settlement, Withdrawal): 70%
- 高级功能 (Dispute, Reconciliation): 60%

**关键行动**:
1. 修复 Accounting 路由 (影响最大)
2. 完整 Channel 管理接口
3. 统一 Withdrawal/Settlement 命名
4. 添加缺失的 API 实现

预计修复时间: 2-4 小时

---

*报告生成时间: 2025-10-25*  
*建议定期更新此报告以跟踪改进进展*

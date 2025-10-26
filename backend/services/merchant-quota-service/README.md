# Merchant Quota Service

## 概述

商户配额服务 - 实时追踪商户的配额消耗,提供配额检查、消耗、释放和预警功能。

## 职责

本服务专注于**配额追踪和监控**(动态消耗管理),不涉及策略配置(由merchant-policy-service负责)。

### 核心功能

1. **配额追踪** (Quota Tracking)
   - 实时追踪商户配额使用情况
   - 支持多币种独立追踪
   - 按日/月/年统计
   - 笔数统计

2. **配额操作** (Quota Operations)
   - 消耗配额 (Consume) - 创建订单时
   - 释放配额 (Release) - 订单取消/退款时
   - 重置配额 (Reset) - 定时任务
   - 手动调整 (Adjust) - 管理员操作

3. **配额检查** (Quota Check)
   - 检查是否超限
   - 返回剩余配额
   - 预测是否可用

4. **使用日志** (Usage Log)
   - 完整的审计日志
   - 操作前后快照
   - 关联订单号/支付号

5. **配额预警** (Quota Alert)
   - 达到80%时发送警告
   - 达到100%时触发限流
   - 通知机制集成

## 与其他服务的关系

### 与 merchant-policy-service 的区别

| 维度 | merchant-policy-service | merchant-quota-service |
|------|------------------------|------------------------|
| **职责** | 策略配置(静态规则) | 配额追踪(动态消耗) |
| **数据特征** | 低频变更 | 高频读写 |
| **核心模型** | FeePolicy, LimitPolicy | MerchantQuota, UsageLog |
| **调用场景** | 策略变更、查询规则 | 每笔交易消耗配额 |
| **示例API** | GET /limit-policies/effective | POST /quotas/consume |

### 服务交互

```
payment-gateway 创建订单流程:

1. 调用 policy-service
   GET /api/v1/limit-policies/effective?merchant_id=xxx
   返回: { daily_limit: 5000000, single_max: 1000000 }

2. 调用 quota-service (检查)
   POST /api/v1/quotas/check
   {
     "merchant_id": "xxx",
     "amount": 50000,
     "currency": "USD"
   }
   返回: { "allowed": true, "remaining_daily": 450000 }

3. 创建订单成功

4. 调用 quota-service (消耗)
   POST /api/v1/quotas/consume
   {
     "merchant_id": "xxx",
     "order_no": "ORDER-001",
     "amount": 50000,
     "currency": "USD"
   }
   返回: { "success": true, "daily_used": 50000 }
```

```
退款流程:

1. 调用 quota-service (释放)
   POST /api/v1/quotas/release
   {
     "merchant_id": "xxx",
     "order_no": "ORDER-001",
     "amount": 50000,
     "currency": "USD"
   }
   返回: { "success": true, "daily_used": 0 }
```

## 数据模型

### 1. MerchantQuota (商户配额)
- 按商户和币种追踪配额使用
- 日/月/年使用量统计
- 待结算金额追踪
- 退款金额统计
- 笔数统计
- 版本号(乐观锁)

### 2. QuotaUsageLog (使用日志)
- 完整的操作记录
- 前后状态快照
- 关联订单信息
- 操作人追溯

### 3. QuotaAlert (配额预警)
- 预警类型 (80%, 100%)
- 预警级别 (warning, critical)
- 处理状态
- 通知记录

## API设计

### 配额操作

```
POST   /api/v1/quotas/check                  # 检查配额
POST   /api/v1/quotas/consume                # 消耗配额
POST   /api/v1/quotas/release                # 释放配额
POST   /api/v1/quotas/reset                  # 重置配额 (定时任务)
POST   /api/v1/quotas/adjust                 # 手动调整 (管理员)
```

### 配额查询

```
GET    /api/v1/quotas/merchant/:id           # 获取商户配额
GET    /api/v1/quotas/merchant/:id/currency/:currency  # 按币种查询
GET    /api/v1/quotas/stats                  # 配额统计
```

### 使用日志

```
GET    /api/v1/usage-logs/merchant/:id       # 获取使用日志
GET    /api/v1/usage-logs/order/:order_no    # 按订单号查询
GET    /api/v1/usage-logs/stats              # 使用统计
```

### 配额预警

```
GET    /api/v1/alerts/merchant/:id           # 获取预警列表
POST   /api/v1/alerts/:id/resolve            # 标记预警已处理
GET    /api/v1/alerts/unresolved             # 未处理的预警
```

## 配置

### 环境变量

```bash
# 服务配置
PORT=40022
DB_NAME=payment_merchant_quota

# 数据库配置
DB_HOST=localhost
DB_PORT=40432
DB_USER=postgres
DB_PASSWORD=postgres

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=40379

# JWT配置
JWT_SECRET=payment-platform-secret-key-2024

# 配置中心
ENABLE_CONFIG_CLIENT=true
CONFIG_SERVICE_URL=http://localhost:40010

# merchant-policy-service地址
POLICY_SERVICE_URL=http://localhost:40012
```

## 特性

### 高性能

- **Redis缓存**: 配额数据缓存
- **批量操作**: 支持批量查询
- **异步日志**: 审计日志异步写入

### 高可用

- **乐观锁**: 防止并发冲突
- **幂等性**: 支持订单号去重
- **降级策略**: policy-service不可用时使用默认限额

### 可观测

- **Metrics**: 配额使用率指标
- **Tracing**: 分布式追踪
- **Logging**: 结构化日志

## 部署

### Docker

```bash
docker build -t merchant-quota-service .
docker run -p 40022:40022 merchant-quota-service
```

### Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## 监控

- **Metrics**: http://localhost:40022/metrics
  - `quota_usage_rate{merchant_id,currency}` - 配额使用率
  - `quota_operations_total{action}` - 操作计数
  - `quota_alerts_total{type,level}` - 预警计数

- **Health**: http://localhost:40022/health
- **Swagger**: http://localhost:40022/swagger/index.html

## 定时任务

### 1. 配额重置任务

每天00:00重置日配额:
```go
// 每日零点重置
func ResetDailyQuotas() {
    // 重置所有商户的 DailyUsed, RefundedToday, TransactionsToday
    // 更新 DailyResetAt 为明天00:00
}
```

每月1日00:00重置月配额:
```go
// 每月第一天重置
func ResetMonthlyQuotas() {
    // 重置所有商户的 MonthlyUsed, RefundedMonth, TransactionsMonth
    // 更新 MonthlyResetAt 为下月1日
}
```

### 2. 配额预警任务

每5分钟检查:
```go
// 检查配额使用率
func CheckQuotaAlerts() {
    // 查询使用率超过80%的商户 -> 发送warning预警
    // 查询使用率达到100%的商户 -> 发送critical预警
}
```

## 开发

### 本地运行

```bash
cd backend/services/merchant-quota-service
go run cmd/main.go
```

### 热重载

```bash
air
```

### 测试

```bash
go test ./...
```

## 迁移说明

本服务由原 merchant-limit-service 重构而来:

### 主要变化

1. **名称变更**: merchant-limit-service → merchant-quota-service
2. **职责聚焦**: 专注配额追踪,策略配置移到policy-service
3. **模型优化**:
   - 移除 MerchantTier 模型 (迁移到policy-service)
   - 移除费率配置字段 (迁移到policy-service)
   - 移除限额配置字段 (迁移到policy-service)
   - 保留 MerchantQuota 和 UsageLog
   - 新增 QuotaAlert 预警模型

### 数据迁移

详见: `scripts/migrate-to-quota-service.sql`

## 性能优化建议

1. **Redis缓存**: 缓存配额数据,减少DB查询
2. **读写分离**: 读从库,写主库
3. **分库分表**: 按商户ID分片
4. **批量操作**: 使用批量API减少网络开销

## 版本

- **v1.0.0**: 初始版本 (重构自merchant-limit-service)
- **端口**: 40022
- **数据库**: payment_merchant_quota

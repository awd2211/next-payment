# 商户分级制度使用指南

## 快速开始

### 1. 四个等级概览

| 等级 | 适用对象 | 日限额 | 月限额 | 费率 | 结算周期 |
|------|---------|--------|--------|------|---------|
| **Starter** | 个人/小微商户 | 10万 | 30万 | 0.8% | T+1 |
| **Business** | 中小企业 | 50万 | 150万 | 0.6% | T+1 |
| **Enterprise** | 大型企业 | 200万 | 600万 | 0.45% | T+0 |
| **Premium** | 超大型客户 | 1000万 | 3000万 | 0.3% | D+0 |

### 2. 初始化默认配置

服务启动时自动初始化4个等级的默认配置：

```bash
# merchant-service 启动日志
INFO  默认等级配置已初始化（Starter/Business/Enterprise/Premium）
```

数据库自动创建4条记录到 `merchant_tier_configs` 表。

---

## API使用示例

### 1. 获取所有等级配置

```bash
GET /api/v1/merchant/tiers
```

**响应示例**:
```json
{
  "code": 0,
  "message": "成功",
  "data": [
    {
      "tier": "starter",
      "name": "入门版",
      "name_en": "Starter",
      "daily_limit": 1000000,
      "monthly_limit": 3000000,
      "single_limit": 100000,
      "fee_rate": 0.008,
      "settlement_cycle": "T+1",
      "enable_multi_currency": false,
      "enable_pre_auth": false,
      "api_rate_limit": 100,
      "support_level": "standard"
    },
    // ... Business, Enterprise, Premium
  ]
}
```

### 2. 获取商户当前等级

```bash
GET /api/v1/merchant/profile
```

**响应包含**:
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "name": "测试商户",
    "tier": "business",  // 当前等级
    "status": "active"
  }
}
```

### 3. 升级商户等级

```bash
POST /api/v1/admin/merchants/{merchant_id}/upgrade
Content-Type: application/json

{
  "new_tier": "enterprise",
  "operator": "admin@example.com",
  "reason": "交易量持续增长，主动升级"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "商户等级升级成功",
  "data": {
    "merchant_id": "uuid",
    "from_tier": "business",
    "to_tier": "enterprise",
    "upgraded_at": "2025-01-24T10:30:00Z"
  }
}
```

### 4. 检查功能权限

```bash
GET /api/v1/merchant/permissions/pre_auth
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "feature": "pre_auth",
    "enabled": false,  // Starter/Business 不支持
    "required_tier": "enterprise"
  }
}
```

### 5. 获取升级推荐

```bash
GET /api/v1/merchant/tier/recommendation
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "current_tier": "business",
    "recommended_tier": "enterprise",
    "reason": "月交易量已达限额的85.3%，建议升级以获得更高限额和更低费率",
    "benefits": [
      "交易费率从 0.6% 降至 0.45%",
      "日限额从 50万 提升至 200万",
      "结算周期从 T+1 变为 T+0",
      "支持预授权、循环扣款、分账功能",
      "专属VIP客服，4小时SLA"
    ]
  }
}
```

---

## 编程接口（内部服务调用）

### 1. 获取等级配置

```go
import (
    "payment-platform/merchant-service/internal/model"
    "payment-platform/merchant-service/internal/service"
)

// 获取等级配置
config, err := tierService.GetTierConfig(ctx, model.TierBusiness)
if err != nil {
    return err
}

fmt.Printf("Business版日限额: %d\n", config.DailyLimit) // 5000000 (50万)
fmt.Printf("Business版费率: %.2f%%\n", config.FeeRate*100) // 0.60%
```

### 2. 计算手续费

```go
// 方法1: 通过配置直接计算
config, _ := tierService.GetTierConfig(ctx, model.TierBusiness)
fee := config.CalculateFee(100000) // 100000分 = 1000元
// fee = 600分 = 6元 (0.6%)

// 方法2: 通过服务层计算（自动查询商户等级）
fee, err := tierService.CalculateMerchantFee(ctx, merchantID, 100000)
```

### 3. 检查限额

```go
// 检查是否可以处理该交易
canProcess, reason := config.CanProcess(
    amount,      // 请求金额
    dailyUsed,   // 今日已用
    monthlyUsed, // 本月已用
)

if !canProcess {
    return fmt.Errorf("交易被拒绝: %s", reason)
}
```

### 4. 升级商户等级

```go
err := tierService.UpgradeMerchantTier(
    ctx,
    merchantID,
    model.TierEnterprise,        // 目标等级
    "admin@example.com",         // 操作员
    "交易量增长，主动升级",      // 原因
)

if err != nil {
    log.Printf("升级失败: %v", err)
}

// 自动执行:
// 1. 更新 merchants 表的 tier 字段
// 2. 更新 merchant_limits 表的限额配置
// 3. 记录操作日志
```

### 5. 降级商户等级

```go
err := tierService.DowngradeMerchantTier(
    ctx,
    merchantID,
    model.TierStarter,           // 目标等级
    "risk@example.com",          // 操作员
    "触发风控规则，强制降级",   // 原因
)
```

### 6. 权限检查

```go
// 检查商户是否有权限使用某功能
features := []string{
    "multi_currency",   // 多币种
    "refund",          // 退款
    "partial_refund",  // 部分退款
    "pre_auth",        // 预授权
    "recurring",       // 循环扣款
    "split",           // 分账
    "webhook",         // Webhook
    "custom_branding", // 自定义品牌
}

for _, feature := range features {
    hasPermission, err := tierService.CheckTierPermission(ctx, merchantID, feature)
    if err != nil {
        continue
    }
    fmt.Printf("%s: %v\n", feature, hasPermission)
}
```

### 7. 智能升级推荐

```go
recommendedTier, reason, err := tierService.RecommendTierUpgrade(ctx, merchantID)
if err != nil {
    return err
}

if recommendedTier != nil {
    // 有推荐升级
    fmt.Printf("建议升级到 %s\n", *recommendedTier)
    fmt.Printf("原因: %s\n", reason)

    // 可以触发自动通知
    notificationService.SendUpgradeRecommendation(merchantID, *recommendedTier, reason)
} else {
    // 无需升级
    fmt.Println(reason) // "当前等级适合您的业务规模"
}
```

---

## 等级详细对比

### Starter (入门版)

**适用对象**: 个人商户、小微企业、初创公司

**交易限制**:
- 日交易限额: 10万元
- 月交易限额: 30万元
- 单笔限额: 1万元

**费率**:
- 交易手续费: 0.8%
- 最低手续费: 1元/笔
- 提现费用: 2元/笔 + 0.1%

**结算**:
- 结算周期: T+1 (次日结算)
- 自动结算: 否
- 最低结算金额: 100元

**功能权限**:
- ✅ 基础支付
- ✅ 退款
- ✅ Webhook通知
- ❌ 多币种
- ❌ 部分退款
- ❌ 预授权
- ❌ 循环扣款
- ❌ 分账

**技术支持**:
- API限额: 100次/分钟
- 最大API密钥: 2个
- 支持级别: 标准（24小时响应）
- 专属客服: 否
- 子账户数: 1个
- 数据保留: 90天

---

### Business (商业版)

**适用对象**: 中小企业、成长型公司

**交易限制**:
- 日交易限额: 50万元
- 月交易限额: 150万元
- 单笔限额: 5万元

**费率**:
- 交易手续费: 0.6% ⬇️ 节省25%
- 最低手续费: 0.5元/笔
- 提现费用: 1元/笔 + 0.05%

**结算**:
- 结算周期: T+1
- 自动结算: 是 ✅
- 最低结算金额: 50元

**功能权限**:
- ✅ 基础支付
- ✅ 多币种支持 🌍
- ✅ 完整退款
- ✅ 部分退款 ✨
- ✅ Webhook（最多5次重试）
- ❌ 预授权
- ❌ 循环扣款
- ❌ 分账

**技术支持**:
- API限额: 500次/分钟 ⬆️
- 最大API密钥: 5个
- 支持级别: 优先（12小时响应）
- 专属客服: 否
- 子账户数: 5个
- 数据保留: 180天

---

### Enterprise (企业版)

**适用对象**: 大型企业、集团公司

**交易限制**:
- 日交易限额: 200万元
- 月交易限额: 600万元
- 单笔限额: 20万元

**费率**:
- 交易手续费: 0.45% ⬇️ 节省44%
- 最低手续费: 0.2元/笔
- 提现费用: 免费 🎉

**结算**:
- 结算周期: T+0 (当日结算) ⚡
- 自动结算: 是
- 最低结算金额: 10元

**功能权限**:
- ✅ 所有基础功能
- ✅ 多币种支持
- ✅ 预授权 💳
- ✅ 循环扣款 🔄
- ✅ 分账功能 💰
- ✅ Webhook（最多10次重试）
- ✅ 自定义品牌 🎨

**技术支持**:
- API限额: 2000次/分钟 ⬆️
- 最大API密钥: 10个
- 支持级别: VIP（4小时响应）
- 专属客服: 是 👨‍💼
- 子账户数: 20个
- 数据保留: 365天

---

### Premium (尊享版)

**适用对象**: 超大型企业、战略合作伙伴

**交易限制**:
- 日交易限额: 1000万元
- 月交易限额: 3000万元
- 单笔限额: 100万元

**费率**:
- 交易手续费: 0.3% ⬇️ 节省62.5%
- 最低手续费: 无
- 提现费用: 免费

**结算**:
- 结算周期: D+0 (日内结算，最快2小时) 🚀
- 自动结算: 是
- 最低结算金额: 无限制

**功能权限**:
- ✅ 所有功能
- ✅ 最高优先级处理
- ✅ Webhook（最多20次重试）
- ✅ 定制化开发支持

**技术支持**:
- API限额: 10000次/分钟 ⬆️
- 最大API密钥: 50个
- 支持级别: VIP（1小时响应）
- 专属客服: 是 + 7x24服务
- 子账户数: 100个
- 数据保留: 730天（2年）

**专属特权**:
- 专属客户经理
- 优先功能开发
- 定制化解决方案
- 年度业务回顾会议

---

## 升级决策树

```
当前等级: Starter
├─ 月交易量 > 24万（80%） → 推荐升级到 Business
├─ 日交易量 > 7万（70%） → 推荐升级到 Business
└─ 需要多币种支持 → 必须升级到 Business

当前等级: Business
├─ 月交易量 > 120万（80%） → 推荐升级到 Enterprise
├─ 需要T+0结算 → 必须升级到 Enterprise
├─ 需要预授权/分账 → 必须升级到 Enterprise
└─ 需要自定义品牌 → 必须升级到 Enterprise

当前等级: Enterprise
├─ 月交易量 > 480万（80%） → 推荐升级到 Premium
├─ 需要D+0极速结算 → 推荐升级到 Premium
└─ 需要定制化服务 → 推荐升级到 Premium
```

---

## 降级触发条件

### 自动降级（系统触发）

1. **风控规则触发**:
   - 连续3次支付失败率 > 50%
   - 触发反洗钱规则
   - 黑名单命中

2. **合规问题**:
   - KYC审核失败
   - 业务资质过期
   - 监管要求

3. **欠费问题**:
   - 账户余额为负超过30天
   - 欠费金额 > 10000元

### 手动降级（管理员操作）

需要提供降级原因，系统记录操作日志。

```go
// 管理员手动降级
err := tierService.DowngradeMerchantTier(
    ctx,
    merchantID,
    model.TierBusiness,
    "admin@example.com",
    "触发风控规则: 高风险交易占比过高",
)
```

---

## 数据库结构

### merchant_tier_configs 表

```sql
CREATE TABLE merchant_tier_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),

    -- 交易限额
    daily_limit BIGINT NOT NULL,
    monthly_limit BIGINT NOT NULL,
    single_limit BIGINT NOT NULL,

    -- 费率配置
    fee_rate DECIMAL(5,4) NOT NULL,
    min_fee BIGINT DEFAULT 0,
    refund_fee_rate DECIMAL(5,4) DEFAULT 0,
    withdrawal_fee BIGINT DEFAULT 0,
    withdrawal_fee_rate DECIMAL(5,4) DEFAULT 0,

    -- 结算配置
    settlement_cycle VARCHAR(20) NOT NULL,
    auto_settlement BOOLEAN DEFAULT false,
    min_settlement_amount BIGINT DEFAULT 0,

    -- 功能权限
    enable_multi_currency BOOLEAN DEFAULT false,
    enable_refund BOOLEAN DEFAULT true,
    enable_partial_refund BOOLEAN DEFAULT false,
    enable_pre_auth BOOLEAN DEFAULT false,
    enable_recurring BOOLEAN DEFAULT false,
    enable_split BOOLEAN DEFAULT false,
    enable_webhook BOOLEAN DEFAULT true,
    max_webhook_retry INT DEFAULT 3,

    -- API限制
    api_rate_limit INT DEFAULT 100,
    max_api_keys INT DEFAULT 2,
    enable_api_callback BOOLEAN DEFAULT true,

    -- 风控配置
    risk_level VARCHAR(20) DEFAULT 'medium',
    enable_risk_control BOOLEAN DEFAULT true,
    max_daily_failures INT DEFAULT 100,

    -- 技术支持
    support_level VARCHAR(20) DEFAULT 'standard',
    sla_response_time INT DEFAULT 24,
    dedicated_support BOOLEAN DEFAULT false,

    -- 其他限制
    max_sub_accounts INT DEFAULT 1,
    data_retention INT DEFAULT 90,
    custom_branding BOOLEAN DEFAULT false,
    priority INT DEFAULT 0,
    description TEXT,
    description_en TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_tier_configs_tier ON merchant_tier_configs(tier);
CREATE INDEX idx_tier_configs_priority ON merchant_tier_configs(priority);
```

### merchants 表增加字段

```sql
ALTER TABLE merchants
ADD COLUMN tier VARCHAR(20) DEFAULT 'starter';

CREATE INDEX idx_merchants_tier ON merchants(tier);
```

---

## 监控指标

### 等级分布

```sql
-- 查询各等级商户数量
SELECT tier, COUNT(*) as count
FROM merchants
WHERE status = 'active'
GROUP BY tier
ORDER BY
    CASE tier
        WHEN 'starter' THEN 1
        WHEN 'business' THEN 2
        WHEN 'enterprise' THEN 3
        WHEN 'premium' THEN 4
    END;
```

**期望分布**:
- Starter: 60-70%
- Business: 20-25%
- Enterprise: 8-12%
- Premium: 2-5%

### 升级转化率

```sql
-- 查询最近30天的升级记录
SELECT
    from_tier,
    to_tier,
    COUNT(*) as upgrade_count,
    AVG(EXTRACT(EPOCH FROM (upgraded_at - created_at))/86400) as avg_days_to_upgrade
FROM tier_upgrade_logs
WHERE upgraded_at > NOW() - INTERVAL '30 days'
GROUP BY from_tier, to_tier;
```

### 费率收入分析

```sql
-- 各等级的手续费收入对比
SELECT
    m.tier,
    COUNT(p.id) as payment_count,
    SUM(p.amount) as total_amount,
    SUM(p.fee) as total_fee,
    AVG(p.fee * 100.0 / p.amount) as avg_fee_rate
FROM payments p
JOIN merchants m ON p.merchant_id = m.id
WHERE p.created_at > NOW() - INTERVAL '30 days'
  AND p.status = 'success'
GROUP BY m.tier;
```

---

## 常见问题 (FAQ)

### Q1: 如何为新注册商户设置默认等级？

A: 新商户默认为 `starter` 等级，这在 `merchants` 表的 `tier` 字段默认值中定义。

```sql
-- 查看默认值
SELECT column_default
FROM information_schema.columns
WHERE table_name = 'merchants' AND column_name = 'tier';
-- 结果: 'starter'
```

### Q2: 升级是否立即生效？

A: 是的，升级/降级操作立即生效：
1. 更新 `merchants.tier` 字段
2. 更新 `merchant_limits` 表的限额配置
3. 下一笔交易立即使用新费率

### Q3: 降级后，已经使用的额度怎么办？

A: 降级不影响已使用额度，但新的限额立即生效：
- 如果 `已用额度 > 新限额`，则无法发起新交易，直到次日/次月重置
- 建议在降级前检查额度使用情况

### Q4: 可以跨级升级吗？

A: 可以，但建议逐级升级：
```go
// 允许：Starter → Premium
err := tierService.UpgradeMerchantTier(ctx, merchantID, model.TierPremium, ...)

// 建议：Starter → Business → Enterprise → Premium
// 这样可以让商户逐步适应高等级的功能
```

### Q5: 如何计算升级后的费用节省？

A: 使用服务层提供的计算方法：

```go
// 当前等级费用
currentConfig, _ := tierService.GetTierConfig(ctx, model.TierStarter)
currentFee := currentConfig.CalculateFee(1000000) // 8000分 = 80元

// 升级后费用
businessConfig, _ := tierService.GetTierConfig(ctx, model.TierBusiness)
businessFee := businessConfig.CalculateFee(1000000) // 6000分 = 60元

savings := currentFee - businessFee // 2000分 = 20元 (节省25%)
```

### Q6: Premium等级的D+0结算是如何实现的？

A: D+0结算表示日内结算，最快2小时：
1. 交易成功后，系统立即生成结算单
2. 审核通过后，2-4小时内到账
3. 需要配置快速审核通道（自动审核）

### Q7: 自定义品牌功能包括什么？

A: Enterprise和Premium等级支持：
- 自定义支付页面Logo
- 自定义域名（需备案）
- 自定义邮件模板
- 自定义Webhook Header

---

## 测试脚本

### 1. 测试等级配置初始化

```bash
# 启动merchant-service
cd /home/eric/payment/backend/services/merchant-service
go run cmd/main.go

# 查看日志
# 应该看到: INFO  默认等级配置已初始化（Starter/Business/Enterprise/Premium）

# 查询数据库
psql -h localhost -p 40432 -U postgres -d payment_merchant -c "
SELECT tier, name, daily_limit, fee_rate, settlement_cycle
FROM merchant_tier_configs
ORDER BY priority;
"
```

### 2. 测试商户升级

```bash
# 创建测试商户（默认Starter等级）
curl -X POST http://localhost:40002/api/v1/merchants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试商户",
    "email": "test@example.com",
    "password": "password123"
  }'

# 升级到Business
curl -X POST http://localhost:40002/api/v1/admin/merchants/{merchant_id}/upgrade \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {admin_token}" \
  -d '{
    "new_tier": "business",
    "operator": "admin@example.com",
    "reason": "测试升级"
  }'
```

### 3. 测试手续费计算

```go
package main

import (
    "context"
    "fmt"
    "payment-platform/merchant-service/internal/model"
)

func main() {
    tiers := []model.MerchantTier{
        model.TierStarter,
        model.TierBusiness,
        model.TierEnterprise,
        model.TierPremium,
    }

    amount := int64(1000000) // 10000元

    for _, tier := range tiers {
        config := model.GetDefaultTierConfig(tier)
        fee := config.CalculateFee(amount)
        fmt.Printf("%s: 交易额=%d, 手续费=%d (%.2f%%)\n",
            tier, amount, fee, float64(fee)/float64(amount)*100)
    }
}

// 输出:
// starter: 交易额=1000000, 手续费=8000 (0.80%)
// business: 交易额=1000000, 手续费=6000 (0.60%)
// enterprise: 交易额=1000000, 手续费=4500 (0.45%)
// premium: 交易额=1000000, 手续费=3000 (0.30%)
```

---

## 总结

商户分级制度提供了灵活的商户管理能力，通过差异化的费率、限额和功能权限，满足不同规模商户的需求。

**关键优势**:
- 🎯 精准定价：4个等级覆盖所有商户类型
- 💰 费率优惠：最高节省62.5%手续费
- ⚡ 极速结算：Premium等级支持D+0
- 🔧 功能丰富：预授权、分账、循环扣款等高级功能
- 📊 数据驱动：智能推荐升级，优化商户体验

**生产就绪**:
- ✅ 编译通过
- ✅ 数据库自动迁移
- ✅ 默认配置自动初始化
- ✅ 完整的API和服务层接口
- ✅ 详细的日志记录

建议在正式上线前进行完整的集成测试，确保升级/降级流程、手续费计算、权限检查等功能正常工作。

---

**文档版本**: v1.0
**最后更新**: 2025-01-24
**维护者**: Payment Platform Team

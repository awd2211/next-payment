# 全球支付平台开发计划 - 全球化方案

> **方案名称**: 方案2 - 全球化 (Globalization)
>
> **目标**: 覆盖全球主流支付市场，支持中国、欧美、东南亚
>
> **选定模块**: A(对账系统) + B(拒付管理) + C(商户额度) + D(PayPal) + E(支付宝/微信) + H(支付超时) + M(灾备高可用)
>
> **预估工作量**: 17周 → **并行开发 10-12周**
>
> **建议团队**: 3后端 + 1前端 + 1DevOps
>
> **生成时间**: 2025-10-25

---

## 📋 目录

1. [总览](#总览)
2. [开发时间线](#开发时间线)
3. [模块详细设计](#模块详细设计)
4. [数据库设计](#数据库设计)
5. [API接口设计](#api接口设计)
6. [部署架构](#部署架构)
7. [测试计划](#测试计划)
8. [风险与应对](#风险与应对)

---

## 📊 总览

### 核心目标

- ✅ **财务安全**: 通过对账系统确保资金准确性
- ✅ **风险控制**: 拒付管理 + 商户额度管理降低损失
- ✅ **全球覆盖**: Stripe + PayPal + 支付宝 + 微信支付覆盖90%+全球市场
- ✅ **运营效率**: 自动化超时处理释放人力
- ✅ **高可用性**: 灾备系统保障99.99%可用性

### 覆盖市场

| 地区 | 渠道 | 市场覆盖率 |
|------|------|-----------|
| 🇨🇳 中国 | 支付宝 + 微信支付 | 90%+ |
| 🇺🇸 北美 | Stripe + PayPal | 85%+ |
| 🇪🇺 欧洲 | Stripe + PayPal | 80%+ |
| 🌏 东南亚 | Stripe + 支付宝 | 70%+ |
| 🌍 其他 | Stripe + PayPal | 60%+ |

### 关键指标提升

| 指标 | 现状 | 目标 | 提升 |
|------|------|------|------|
| 市场覆盖 | 仅Stripe (30%) | 90%+ | +60% |
| 拒付损失率 | 不可控 | <0.5% | -70% |
| 对账准确性 | 人工对账 | 99.9% | 自动化 |
| 订单积压 | 长期pending | 0 | 100% |
| 系统可用性 | 单点故障 | 99.99% | 高可用 |

---

## 🗓️ 开发时间线

### 总览 (12周 = 3个月)

```
Week 1-2   : Sprint 1 - 基础设施 + 框架搭建
Week 3-5   : Sprint 2 - 核心功能开发 (对账 + 拒付 + 额度)
Week 6-8   : Sprint 3 - 渠道集成 (PayPal + 支付宝/微信)
Week 9-10  : Sprint 4 - 超时处理 + 灾备部署
Week 11    : Sprint 5 - 集成测试 + 压力测试
Week 12    : Sprint 6 - 上线准备 + 灰度发布
```

### 详细甘特图

#### **Sprint 1: Week 1-2 (基础设施)**

| 任务 | 负责人 | 天数 | 依赖 |
|------|--------|------|------|
| 数据库Schema设计 | 后端1 | 2 | - |
| API接口设计评审 | 团队 | 1 | Schema |
| PostgreSQL主从搭建 | DevOps | 3 | - |
| Redis Cluster搭建 | DevOps | 2 | - |
| Kafka多副本配置 | DevOps | 2 | - |
| CI/CD Pipeline | DevOps | 2 | - |
| 监控Dashboard | DevOps | 1 | - |

**交付物**:
- ✅ 完整的数据库Schema
- ✅ API接口文档 v1.0
- ✅ 高可用基础设施
- ✅ CI/CD自动化流程

---

#### **Sprint 2: Week 3-5 (核心功能)**

**模块A: 对账系统** (后端1, 5天)

| 任务 | 天数 | 说明 |
|------|------|------|
| Reconciliation Service实现 | 2 | 核心对账逻辑 |
| 渠道账单下载器 (Stripe) | 1 | 定时下载Stripe账单 |
| 差异检测算法 | 1 | 三方对账 + 差异标记 |
| 对账报表生成 | 1 | PDF报表 + 邮件通知 |

**模块B: 拒付管理** (后端2, 4天)

| 任务 | 天数 | 说明 |
|------|------|------|
| Dispute Model + Repository | 1 | 数据模型 |
| Webhook接收器 | 1 | stripe.dispute.* 事件 |
| 证据上传API | 1 | 商户上传证据 |
| 自动提交到Stripe | 1 | 调用Dispute Evidence API |

**模块C: 商户额度管理** (后端3, 5天)

| 任务 | 天数 | 说明 |
|------|------|------|
| 商户分级系统 | 1 | Tier配置表 + 升降级逻辑 |
| 额度计算服务 | 1 | 实时额度检查 |
| Redis计数器 | 1 | 日/月累计统计 |
| 超限拦截中间件 | 1 | 支付前检查 |
| 额度统计Dashboard | 1 | Admin后台页面 |

**前端开发** (前端, 5天)

| 任务 | 天数 | 说明 |
|------|------|------|
| 对账报表页面 | 2 | 对账差异列表 + 详情 |
| 拒付管理页面 | 2 | 拒付工单 + 证据上传 |
| 商户额度页面 | 1 | 额度使用情况 + 调整 |

**交付物**:
- ✅ 对账系统每日自动运行
- ✅ 拒付自动检测 + 通知
- ✅ 商户额度实时拦截

---

#### **Sprint 3: Week 6-8 (渠道集成)**

**模块D: PayPal集成** (后端1, 8天)

| 任务 | 天数 | 说明 |
|------|------|------|
| PayPal Adapter实现 | 2 | 实现PaymentAdapter接口 |
| Checkout集成 | 2 | Orders API v2 |
| 退款集成 | 1 | Captures API |
| Webhook验证 | 1 | PAYPAL-TRANSMISSION-SIG |
| 对账文件下载 | 1 | Settlement Report API |
| 单元测试 | 1 | Mock测试 |

**模块E: 支付宝/微信支付** (后端2+3, 10天)

| 任务 | 负责人 | 天数 | 说明 |
|------|--------|------|------|
| 支付宝Adapter | 后端2 | 3 | APP/网页/扫码支付 |
| 支付宝退款 | 后端2 | 1 | 退款API |
| 支付宝对账 | 后端2 | 1 | 账单下载 |
| 微信Adapter | 后端3 | 3 | APP/网页/小程序 |
| 微信退款 | 后端3 | 1 | 退款API |
| 微信对账 | 后端3 | 1 | 账单下载 |

**前端开发** (前端, 5天)

| 任务 | 天数 | 说明 |
|------|------|------|
| PayPal支付页面 | 2 | Checkout集成 |
| 支付宝支付页面 | 2 | 二维码展示 + 跳转 |
| 微信支付页面 | 1 | 小程序/H5跳转 |

**交付物**:
- ✅ PayPal全流程 (支付+退款+对账)
- ✅ 支付宝全流程
- ✅ 微信支付全流程
- ✅ 4大渠道完整覆盖

---

#### **Sprint 4: Week 9-10 (超时处理 + 灾备)**

**模块H: 支付超时处理** (后端1, 3天)

| 任务 | 天数 | 说明 |
|------|------|------|
| TimeoutService实现 | 1 | 定时扫描服务 |
| 渠道取消集成 | 1 | 调用各渠道Cancel API |
| 超时通知 | 1 | 邮件/Webhook通知 |

**模块M: 灾备高可用** (DevOps, 7天)

| 任务 | 天数 | 说明 |
|------|------|------|
| PostgreSQL Failover测试 | 2 | 主从切换演练 |
| Redis Sentinel配置 | 1 | 哨兵模式 |
| Kafka副本验证 | 1 | 副本同步测试 |
| 服务多实例部署 | 2 | Kubernetes HPA |
| 备份恢复演练 | 1 | 数据恢复测试 |

**交付物**:
- ✅ 超时订单自动清理
- ✅ 高可用架构部署完成
- ✅ 99.99%可用性达成

---

#### **Sprint 5: Week 11 (测试)**

| 测试类型 | 负责人 | 天数 | 覆盖范围 |
|---------|--------|------|----------|
| 单元测试 | 后端团队 | 2 | 所有新增Service |
| 集成测试 | 后端团队 | 2 | 端到端流程 |
| 压力测试 | DevOps | 1 | 1000 TPS目标 |
| 对账验证 | 后端1 | 1 | 模拟真实对账 |
| 灾备演练 | DevOps | 1 | 主库故障切换 |

**测试用例数**: 200+ (详见测试计划章节)

---

#### **Sprint 6: Week 12 (上线)**

| 任务 | 负责人 | 天数 | 说明 |
|------|--------|------|------|
| 生产环境部署 | DevOps | 1 | Kubernetes部署 |
| 灰度发布 (10%) | 团队 | 2 | 选定测试商户 |
| 监控告警配置 | DevOps | 1 | Prometheus + PagerDuty |
| 文档完善 | 团队 | 1 | 运维手册 + API文档 |
| 全量发布 | 团队 | 1 | 100%流量切换 |

**上线检查清单**: 50项 (详见部署章节)

---

## 🔧 模块详细设计

### 模块A: 对账系统

#### 业务流程

```
每日凌晨2:00自动触发:
┌─────────────────────────────────────────────────────┐
│ 1. 下载渠道账单                                      │
│    - Stripe: /v1/balance_transactions               │
│    - PayPal: /v1/reporting/transactions             │
│    - 支付宝: dataservice.bill.downloadurl.query      │
│    - 微信: POST /pay/downloadbill                    │
├─────────────────────────────────────────────────────┤
│ 2. 查询内部账务记录                                  │
│    SELECT * FROM payments WHERE paid_at::date = ?   │
├─────────────────────────────────────────────────────┤
│ 3. 三方对账匹配                                      │
│    - 按 payment_no/channel_order_no 匹配            │
│    - 比对金额、状态、手续费                          │
├─────────────────────────────────────────────────────┤
│ 4. 生成对账差异                                      │
│    - missing: 内部有，渠道无 (可能丢单)             │
│    - duplicate: 渠道有，内部无 (可能重复)            │
│    - amount_mismatch: 金额不一致                     │
├─────────────────────────────────────────────────────┤
│ 5. 生成对账报表                                      │
│    - PDF报表邮件发送财务团队                         │
│    - 差异工单自动创建                                │
└─────────────────────────────────────────────────────┘
```

#### 核心代码结构

```go
// backend/services/accounting-service/internal/service/reconciliation_service.go

package service

import (
    "context"
    "time"
)

type ReconciliationService interface {
    // 每日对账（定时任务调用）
    DailyReconcile(ctx context.Context, date time.Time) (*ReconcileReport, error)

    // 下载渠道账单
    DownloadChannelStatements(ctx context.Context, channel string, date time.Time) ([]Transaction, error)

    // 查询内部账务
    GetInternalTransactions(ctx context.Context, date time.Time) ([]Transaction, error)

    // 三方对账匹配
    Match(internal, channel []Transaction) (*MatchResult, error)

    // 处理差异
    HandleDiscrepancies(ctx context.Context, discrepancies []Discrepancy) error

    // 生成对账报表
    GenerateReport(ctx context.Context, result *MatchResult) (*ReconcileReport, error)
}

// 对账结果
type ReconcileReport struct {
    Date                time.Time
    TotalInternal       int64  // 内部总笔数
    TotalChannel        int64  // 渠道总笔数
    Matched             int64  // 匹配成功
    Missing             int64  // 内部有，渠道无
    Duplicate           int64  // 渠道有，内部无
    AmountMismatch      int64  // 金额不一致
    TotalDiscrepancies  int64  // 总差异数
    MatchRate           float64 // 匹配率 (%)
    Discrepancies       []Discrepancy
}

// 差异类型
type Discrepancy struct {
    ID              uuid.UUID
    Type            string  // "missing", "duplicate", "amount_mismatch"
    InternalOrderNo string
    ChannelOrderNo  string
    InternalAmount  int64
    ChannelAmount   int64
    AmountDiff      int64
    Status          string  // "pending", "resolved", "written_off"
    Resolution      string  // 处理说明
    ResolvedAt      *time.Time
}
```

#### 数据库表设计

```sql
-- 对账记录表
CREATE TABLE reconciliation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconcile_date DATE NOT NULL,
    channel VARCHAR(50) NOT NULL,
    total_internal BIGINT NOT NULL,
    total_channel BIGINT NOT NULL,
    matched BIGINT NOT NULL,
    missing BIGINT NOT NULL,
    duplicate BIGINT NOT NULL,
    amount_mismatch BIGINT NOT NULL,
    match_rate DECIMAL(5,2) NOT NULL,
    status VARCHAR(20) NOT NULL, -- "pending", "completed", "reviewed"
    report_file_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    reviewed_by UUID,
    reviewed_at TIMESTAMP,
    INDEX idx_reconcile_date (reconcile_date),
    INDEX idx_status (status)
);

-- 对账差异表
CREATE TABLE reconciliation_discrepancies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reconciliation_reports(id),
    type VARCHAR(20) NOT NULL, -- "missing", "duplicate", "amount_mismatch"
    internal_order_no VARCHAR(100),
    channel_order_no VARCHAR(100),
    internal_amount BIGINT,
    channel_amount BIGINT,
    amount_diff BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    resolution TEXT,
    resolved_by UUID,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_report_id (report_id),
    INDEX idx_status (status),
    INDEX idx_type (type)
);
```

---

### 模块B: 拒付管理

#### 业务流程

```
Stripe Webhook: dispute.created
┌─────────────────────────────────────────────────────┐
│ 1. Webhook接收                                       │
│    POST /webhooks/stripe                            │
│    验证签名 → 解析事件                               │
├─────────────────────────────────────────────────────┤
│ 2. 创建拒付记录                                      │
│    - 提取dispute_id, payment_intent, amount        │
│    - 查询关联的Payment记录                           │
│    - 创建Dispute记录 (status=warning)               │
├─────────────────────────────────────────────────────┤
│ 3. 通知商户                                          │
│    - 邮件通知 (附证据要求)                           │
│    - Webhook通知商户系统                             │
│    - Admin后台创建工单                               │
├─────────────────────────────────────────────────────┤
│ 4. 商户上传证据                                      │
│    POST /api/v1/disputes/{id}/evidence              │
│    - 上传文件到S3/OSS                                │
│    - 保存证据URL到数据库                             │
├─────────────────────────────────────────────────────┤
│ 5. 自动提交到Stripe                                  │
│    POST /v1/disputes/{id}                           │
│    - evidence[customer_name]                        │
│    - evidence[receipt]                              │
│    - submit=true                                    │
├─────────────────────────────────────────────────────┤
│ 6. 跟踪结果                                          │
│    Webhook: dispute.won / dispute.lost              │
│    - 更新status                                      │
│    - 通知商户                                         │
└─────────────────────────────────────────────────────┘
```

#### 核心代码

```go
// backend/services/payment-gateway/internal/service/dispute_service.go

package service

type DisputeService interface {
    // Webhook接收拒付通知
    HandleDisputeWebhook(ctx context.Context, event *stripe.Event) error

    // 获取拒付详情
    GetDispute(ctx context.Context, id uuid.UUID) (*Dispute, error)

    // 商户上传证据
    UploadEvidence(ctx context.Context, disputeID uuid.UUID, evidence *EvidenceInput) error

    // 自动提交证据到Stripe
    SubmitEvidenceToStripe(ctx context.Context, disputeID uuid.UUID) error

    // 拒付列表
    ListDisputes(ctx context.Context, query *DisputeQuery) ([]*Dispute, int64, error)

    // 拒付统计
    GetDisputeStatistics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (*DisputeStats, error)
}

type Dispute struct {
    ID              uuid.UUID
    DisputeNo       string
    PaymentID       uuid.UUID
    PaymentNo       string
    MerchantID      uuid.UUID
    Channel         string  // "stripe", "paypal"
    ChannelDisputeID string
    Amount          int64
    Currency        string
    Reason          string  // "fraud", "duplicate", "product_not_received", "product_unacceptable"
    Status          string  // "warning", "needs_response", "under_review", "won", "lost", "expired"
    DueDate         *time.Time
    EvidenceFiles   []EvidenceFile
    SubmittedAt     *time.Time
    ResolvedAt      *time.Time
    Resolution      string
    CreatedAt       time.Time
}

type EvidenceFile struct {
    Type        string  // "receipt", "tracking", "customer_communication"
    FileName    string
    FileURL     string
    UploadedAt  time.Time
}
```

#### 数据库表设计

```sql
-- 拒付记录表
CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_no VARCHAR(100) UNIQUE NOT NULL,
    payment_id UUID NOT NULL REFERENCES payments(id),
    payment_no VARCHAR(100) NOT NULL,
    merchant_id UUID NOT NULL,
    channel VARCHAR(50) NOT NULL,
    channel_dispute_id VARCHAR(200) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'warning',
    due_date TIMESTAMP,
    evidence_details JSONB, -- 证据详情 (JSON格式)
    submitted_at TIMESTAMP,
    resolved_at TIMESTAMP,
    resolution TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_payment_id (payment_id),
    INDEX idx_merchant_id (merchant_id),
    INDEX idx_status (status),
    INDEX idx_due_date (due_date),
    UNIQUE INDEX idx_channel_dispute (channel, channel_dispute_id)
);

-- 拒付证据文件表
CREATE TABLE dispute_evidence_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id),
    file_type VARCHAR(50) NOT NULL, -- "receipt", "tracking", "communication"
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT,
    uploaded_by UUID,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_dispute_id (dispute_id)
);
```

---

### 模块C: 商户额度管理

#### 商户分级体系

```go
type MerchantTier string

const (
    TierStarter    MerchantTier = "starter"    // 起步级
    TierBusiness   MerchantTier = "business"   // 商务级
    TierEnterprise MerchantTier = "enterprise" // 企业级
    TierPremium    MerchantTier = "premium"    // 顶级
)

type TierConfig struct {
    Tier              MerchantTier
    MonthlyLimit      int64   // 月交易限额 (分)
    DailyLimit        int64   // 日交易限额
    SingleLimit       int64   // 单笔限额
    FeeRate           int     // 手续费率 (基点, 1基点=0.01%)
    SettlementCycle   string  // 结算周期 ("T+1", "T+0")
    SupportLevel      string  // 客服等级
    APIRateLimit      int     // API调用频率 (次/分钟)
}

// 预设分级配置
var DefaultTierConfigs = map[MerchantTier]*TierConfig{
    TierStarter: {
        MonthlyLimit:    10000000,  // 10万元
        DailyLimit:      500000,    // 5千元
        SingleLimit:     100000,    // 1千元
        FeeRate:         60,        // 0.6%
        SettlementCycle: "T+1",
        SupportLevel:    "standard",
        APIRateLimit:    100,
    },
    TierBusiness: {
        MonthlyLimit:    100000000, // 100万元
        DailyLimit:      5000000,   // 5万元
        SingleLimit:     1000000,   // 1万元
        FeeRate:         50,        // 0.5%
        SettlementCycle: "T+1",
        SupportLevel:    "priority",
        APIRateLimit:    500,
    },
    TierEnterprise: {
        MonthlyLimit:    1000000000, // 1000万元
        DailyLimit:      50000000,   // 50万元
        SingleLimit:     10000000,   // 10万元
        FeeRate:         40,         // 0.4%
        SettlementCycle: "T+0",
        SupportLevel:    "premium",
        APIRateLimit:    2000,
    },
    TierPremium: {
        MonthlyLimit:    -1,         // 无限制
        DailyLimit:      -1,
        SingleLimit:     -1,
        FeeRate:         30,         // 0.3% (协商)
        SettlementCycle: "T+0",
        SupportLevel:    "dedicated",
        APIRateLimit:    10000,
    },
}
```

#### 额度检查服务

```go
// backend/services/merchant-service/internal/service/limit_service.go

package service

type LimitService interface {
    // 检查额度 (支付前调用)
    CheckLimit(ctx context.Context, merchantID uuid.UUID, amount int64) (*LimitCheckResult, error)

    // 扣减额度 (支付成功后调用)
    DeductLimit(ctx context.Context, merchantID uuid.UUID, amount int64) error

    // 恢复额度 (退款后调用)
    RestoreLimit(ctx context.Context, merchantID uuid.UUID, amount int64) error

    // 获取额度使用情况
    GetLimitUsage(ctx context.Context, merchantID uuid.UUID) (*LimitUsage, error)

    // 调整额度 (运营手动调整)
    AdjustLimit(ctx context.Context, merchantID uuid.UUID, newTier MerchantTier, customLimits *CustomLimits) error
}

type LimitCheckResult struct {
    Allowed         bool
    RemainingDaily  int64
    RemainingMonthly int64
    Reason          string  // 超限原因
}

type LimitUsage struct {
    Tier              MerchantTier
    DailyLimit        int64
    DailyUsed         int64
    DailyRemaining    int64
    MonthlyLimit      int64
    MonthlyUsed       int64
    MonthlyRemaining  int64
    SingleLimit       int64
}

// Redis Key设计
const (
    KeyDailyLimit  = "merchant:limit:daily:{merchant_id}:{date}"    // TTL: 24小时
    KeyMonthlyLimit = "merchant:limit:monthly:{merchant_id}:{month}" // TTL: 31天
)

// 实现示例
func (s *limitService) CheckLimit(ctx context.Context, merchantID uuid.UUID, amount int64) (*LimitCheckResult, error) {
    // 1. 获取商户分级配置
    merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
    if err != nil {
        return nil, err
    }

    config := DefaultTierConfigs[merchant.Tier]

    // 2. 检查单笔限额
    if config.SingleLimit > 0 && amount > config.SingleLimit {
        return &LimitCheckResult{
            Allowed: false,
            Reason:  fmt.Sprintf("超过单笔限额: %.2f元", float64(config.SingleLimit)/100),
        }, nil
    }

    // 3. 检查日限额
    today := time.Now().Format("20060102")
    dailyKey := fmt.Sprintf("merchant:limit:daily:%s:%s", merchantID, today)
    dailyUsed, _ := s.redisClient.Get(ctx, dailyKey).Int64()

    if config.DailyLimit > 0 && (dailyUsed + amount) > config.DailyLimit {
        return &LimitCheckResult{
            Allowed:        false,
            RemainingDaily: max(0, config.DailyLimit - dailyUsed),
            Reason:         "超过日限额",
        }, nil
    }

    // 4. 检查月限额
    month := time.Now().Format("200601")
    monthlyKey := fmt.Sprintf("merchant:limit:monthly:%s:%s", merchantID, month)
    monthlyUsed, _ := s.redisClient.Get(ctx, monthlyKey).Int64()

    if config.MonthlyLimit > 0 && (monthlyUsed + amount) > config.MonthlyLimit {
        return &LimitCheckResult{
            Allowed:          false,
            RemainingMonthly: max(0, config.MonthlyLimit - monthlyUsed),
            Reason:           "超过月限额",
        }, nil
    }

    return &LimitCheckResult{
        Allowed:          true,
        RemainingDaily:   config.DailyLimit - dailyUsed - amount,
        RemainingMonthly: config.MonthlyLimit - monthlyUsed - amount,
    }, nil
}
```

#### 数据库表设计

```sql
-- 扩展商户表
ALTER TABLE merchants
ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'starter',
ADD COLUMN custom_daily_limit BIGINT,
ADD COLUMN custom_monthly_limit BIGINT,
ADD COLUMN custom_single_limit BIGINT,
ADD COLUMN custom_fee_rate INT,
ADD INDEX idx_tier (tier);

-- 额度使用历史表 (用于对账和审计)
CREATE TABLE merchant_limit_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL,
    usage_date DATE NOT NULL,
    tier VARCHAR(20) NOT NULL,
    daily_limit BIGINT NOT NULL,
    daily_used BIGINT NOT NULL DEFAULT 0,
    monthly_limit BIGINT NOT NULL,
    monthly_used BIGINT NOT NULL DEFAULT 0,
    transaction_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE INDEX idx_merchant_date (merchant_id, usage_date)
);
```

---

### 模块D: PayPal集成

#### 接入方案

PayPal提供两种集成方式，我们选择 **Orders API v2** (推荐)：

```
Orders API v2:
- 支持多种支付方式 (信用卡, PayPal余额, BNPL)
- 统一的Webhook事件
- 更好的错误处理
```

#### 核心流程

```go
// backend/services/channel-adapter/internal/adapter/paypal_adapter.go

package adapter

import (
    "github.com/plutov/paypal/v4"
)

type PayPalAdapter struct {
    client *paypal.Client
    config *PayPalConfig
}

type PayPalConfig struct {
    ClientID     string
    ClientSecret string
    Mode         string // "sandbox" or "live"
    WebhookID    string
}

func NewPayPalAdapter(config *PayPalConfig) *PayPalAdapter {
    client, _ := paypal.NewClient(config.ClientID, config.ClientSecret, config.Mode)
    return &PayPalAdapter{
        client: client,
        config: config,
    }
}

// 实现PaymentAdapter接口
func (a *PayPalAdapter) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    // 1. 创建PayPal订单
    order, err := a.client.CreateOrder(ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{
        {
            Amount: &paypal.PurchaseUnitAmount{
                Currency: req.Currency,
                Value:    fmt.Sprintf("%.2f", float64(req.Amount)/100),
            },
            ReferenceID: req.PaymentNo,
            CustomID:    req.MerchantID,
        },
    }, &paypal.CreateOrderPayer{}, &paypal.ApplicationContext{
        BrandName:  "Your Payment Platform",
        ReturnURL:  req.ReturnURL,
        CancelURL:  req.ReturnURL + "?cancelled=true",
    })

    if err != nil {
        return nil, fmt.Errorf("PayPal CreateOrder failed: %w", err)
    }

    // 2. 提取Approve链接
    var approveURL string
    for _, link := range order.Links {
        if link.Rel == "approve" {
            approveURL = link.Href
            break
        }
    }

    return &CreatePaymentResponse{
        ChannelOrderNo: order.ID,
        PaymentURL:     approveURL,
        Status:         "pending_approval",
    }, nil
}

func (a *PayPalAdapter) CapturePayment(ctx context.Context, orderID string) error {
    // 捕获授权的订单 (用户完成授权后调用)
    _, err := a.client.CaptureOrder(ctx, orderID, paypal.CaptureOrderRequest{})
    return err
}

func (a *PayPalAdapter) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
    // PayPal退款
    refund, err := a.client.RefundCapture(ctx, req.ChannelOrderNo, &paypal.RefundCaptureRequest{
        Amount: &paypal.Money{
            Currency: req.Currency,
            Value:    fmt.Sprintf("%.2f", float64(req.Amount)/100),
        },
    })

    if err != nil {
        return nil, err
    }

    return &RefundResponse{
        ChannelRefundNo: refund.ID,
        Status:          refund.Status,
    }, nil
}
```

#### Webhook验证

```go
func (a *PayPalAdapter) VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) (*paypal.WebhookEvent, error) {
    // PayPal Webhook签名验证
    event, err := paypal.VerifyWebhookSignature(ctx, a.client,
        headers["PAYPAL-TRANSMISSION-ID"],
        headers["PAYPAL-TRANSMISSION-TIME"],
        a.config.WebhookID,
        body,
        headers["PAYPAL-TRANSMISSION-SIG"],
        headers["PAYPAL-CERT-URL"],
        headers["PAYPAL-AUTH-ALGO"],
    )

    return event, err
}
```

---

### 模块E: 支付宝/微信支付集成

#### 支付宝集成

```go
// backend/services/channel-adapter/internal/adapter/alipay_adapter.go

package adapter

import (
    "github.com/smartwalle/alipay/v3"
)

type AlipayAdapter struct {
    client *alipay.Client
    config *AlipayConfig
}

type AlipayConfig struct {
    AppID          string
    PrivateKey     string  // 应用私钥
    AlipayPublicKey string // 支付宝公钥
    NotifyURL      string
}

func NewAlipayAdapter(config *AlipayConfig) *AlipayAdapter {
    client, _ := alipay.New(config.AppID, config.PrivateKey, false)
    client.LoadAliPayPublicKey(config.AlipayPublicKey)

    return &AlipayAdapter{
        client: client,
        config: config,
    }
}

// APP支付
func (a *AlipayAdapter) CreateAppPayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    pay := alipay.TradeAppPay{
        Trade: alipay.Trade{
            Subject:     req.Description,
            OutTradeNo:  req.PaymentNo,
            TotalAmount: fmt.Sprintf("%.2f", float64(req.Amount)/100),
            NotifyURL:   a.config.NotifyURL,
        },
    }

    payParam, err := a.client.TradeAppPay(pay)
    if err != nil {
        return nil, err
    }

    return &CreatePaymentResponse{
        ChannelOrderNo: req.PaymentNo,
        PaymentURL:     payParam.String(), // APP调起参数
        Status:         "pending",
    }, nil
}

// 网页支付
func (a *AlipayAdapter) CreatePagePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    pay := alipay.TradePagePay{
        Trade: alipay.Trade{
            Subject:     req.Description,
            OutTradeNo:  req.PaymentNo,
            TotalAmount: fmt.Sprintf("%.2f", float64(req.Amount)/100),
            NotifyURL:   a.config.NotifyURL,
        },
        ReturnURL: req.ReturnURL,
    }

    payURL, err := a.client.TradePagePay(pay)
    if err != nil {
        return nil, err
    }

    return &CreatePaymentResponse{
        ChannelOrderNo: req.PaymentNo,
        PaymentURL:     payURL.String(),
        Status:         "pending",
    }, nil
}

// 扫码支付
func (a *AlipayAdapter) CreateQRPayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    pay := alipay.TradePrecreate{
        Trade: alipay.Trade{
            Subject:     req.Description,
            OutTradeNo:  req.PaymentNo,
            TotalAmount: fmt.Sprintf("%.2f", float64(req.Amount)/100),
            NotifyURL:   a.config.NotifyURL,
        },
    }

    rsp, err := a.client.TradePrecreate(pay)
    if err != nil {
        return nil, err
    }

    return &CreatePaymentResponse{
        ChannelOrderNo: req.PaymentNo,
        QRCodeURL:      rsp.Content.QRCode, // 二维码URL
        Status:         "pending",
    }, nil
}

// 退款
func (a *AlipayAdapter) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
    refund := alipay.TradeRefund{
        OutTradeNo:   req.PaymentNo,
        RefundAmount: fmt.Sprintf("%.2f", float64(req.Amount)/100),
        RefundReason: req.Reason,
    }

    rsp, err := a.client.TradeRefund(refund)
    if err != nil {
        return nil, err
    }

    return &RefundResponse{
        ChannelRefundNo: rsp.Content.TradeNo,
        Status:          "success",
    }, nil
}
```

#### 微信支付集成

```go
// backend/services/channel-adapter/internal/adapter/wechat_adapter.go

package adapter

import (
    "github.com/wechatpay-apiv3/wechatpay-go/core"
    "github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
    "github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
    "github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

type WechatAdapter struct {
    client *core.Client
    config *WechatConfig
}

type WechatConfig struct {
    MchID       string // 商户号
    AppID       string
    APIv3Key    string // APIv3密钥
    SerialNo    string // 证书序列号
    PrivateKey  string // 商户私钥
    NotifyURL   string
}

// APP支付
func (a *WechatAdapter) CreateAppPayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    svc := app.AppApiService{Client: a.client}

    resp, _, err := svc.Prepay(ctx, app.PrepayRequest{
        Appid:       core.String(a.config.AppID),
        Mchid:       core.String(a.config.MchID),
        Description: core.String(req.Description),
        OutTradeNo:  core.String(req.PaymentNo),
        NotifyUrl:   core.String(a.config.NotifyURL),
        Amount: &app.Amount{
            Total:    core.Int64(req.Amount),
            Currency: core.String(req.Currency),
        },
    })

    if err != nil {
        return nil, err
    }

    return &CreatePaymentResponse{
        ChannelOrderNo: req.PaymentNo,
        PaymentURL:     *resp.PrepayId, // APP调起参数
        Status:         "pending",
    }, nil
}

// JSAPI支付 (小程序/公众号)
func (a *WechatAdapter) CreateJSAPIPayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    svc := jsapi.JsapiApiService{Client: a.client}

    resp, _, err := svc.Prepay(ctx, jsapi.PrepayRequest{
        Appid:       core.String(a.config.AppID),
        Mchid:       core.String(a.config.MchID),
        Description: core.String(req.Description),
        OutTradeNo:  core.String(req.PaymentNo),
        NotifyUrl:   core.String(a.config.NotifyURL),
        Amount: &jsapi.Amount{
            Total: core.Int64(req.Amount),
        },
        Payer: &jsapi.Payer{
            Openid: core.String(req.Extra["openid"].(string)),
        },
    })

    if err != nil {
        return nil, err
    }

    return &CreatePaymentResponse{
        ChannelOrderNo: req.PaymentNo,
        PaymentURL:     *resp.PrepayId,
        Status:         "pending",
    }, nil
}

// Native支付 (扫码)
func (a *WechatAdapter) CreateNativePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
    svc := native.NativeApiService{Client: a.client}

    resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
        Appid:       core.String(a.config.AppID),
        Mchid:       core.String(a.config.MchID),
        Description: core.String(req.Description),
        OutTradeNo:  core.String(req.PaymentNo),
        NotifyUrl:   core.String(a.config.NotifyURL),
        Amount: &native.Amount{
            Total: core.Int64(req.Amount),
        },
    })

    if err != nil {
        return nil, err
    }

    return &CreatePaymentResponse{
        ChannelOrderNo: req.PaymentNo,
        QRCodeURL:      *resp.CodeUrl, // 二维码URL
        Status:         "pending",
    }, nil
}
```

---

### 模块H: 支付超时处理

#### 定时任务设计

```go
// backend/services/payment-gateway/internal/service/timeout_service.go

package service

import (
    "context"
    "time"
)

type TimeoutService interface {
    // 扫描超时订单
    ScanExpiredPayments(ctx context.Context) error

    // 处理单个超时订单
    HandleExpiredPayment(ctx context.Context, payment *model.Payment) error
}

type timeoutService struct {
    paymentRepo   repository.PaymentRepository
    orderClient   *client.OrderClient
    channelClient *client.ChannelClient
    notificationClient *client.NotificationClient
}

func NewTimeoutService(...) TimeoutService {
    return &timeoutService{...}
}

// 定时任务: 每5分钟执行一次
func (s *timeoutService) ScanExpiredPayments(ctx context.Context) error {
    logger.Info("开始扫描超时订单")

    // 查询超时的pending订单
    query := &repository.PaymentQuery{
        Status:     model.PaymentStatusPending,
        ExpiredBefore: time.Now(),
        PageSize:   100,
    }

    payments, _, err := s.paymentRepo.List(ctx, query)
    if err != nil {
        return fmt.Errorf("查询超时订单失败: %w", err)
    }

    logger.Info("发现超时订单", zap.Int("count", len(payments)))

    // 并发处理
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // 限制并发数

    for _, payment := range payments {
        wg.Add(1)
        semaphore <- struct{}{}

        go func(p *model.Payment) {
            defer wg.Done()
            defer func() { <-semaphore }()

            if err := s.HandleExpiredPayment(ctx, p); err != nil {
                logger.Error("处理超时订单失败",
                    zap.String("payment_no", p.PaymentNo),
                    zap.Error(err))
            }
        }(payment)
    }

    wg.Wait()
    logger.Info("超时订单处理完成")
    return nil
}

func (s *timeoutService) HandleExpiredPayment(ctx context.Context, payment *model.Payment) error {
    logger.Info("处理超时订单", zap.String("payment_no", payment.PaymentNo))

    // 1. 调用渠道取消订单
    if s.channelClient != nil {
        if err := s.channelClient.CancelPayment(ctx, payment.Channel, payment.ChannelOrderNo); err != nil {
            logger.Warn("渠道取消失败,继续更新本地状态",
                zap.Error(err),
                zap.String("payment_no", payment.PaymentNo))
        }
    }

    // 2. 更新支付状态为已取消
    payment.Status = model.PaymentStatusCancelled
    payment.ErrorMsg = "支付超时自动取消"
    payment.ErrorCode = "TIMEOUT"

    if err := s.paymentRepo.Update(ctx, payment); err != nil {
        return fmt.Errorf("更新支付状态失败: %w", err)
    }

    // 3. 通知Order Service
    if s.orderClient != nil {
        s.orderClient.CancelOrder(ctx, payment.OrderNo, "支付超时自动取消")
    }

    // 4. 发送通知到商户
    if s.notificationClient != nil {
        go s.sendTimeoutNotification(payment)
    }

    logger.Info("超时订单处理成功", zap.String("payment_no", payment.PaymentNo))
    return nil
}

func (s *timeoutService) sendTimeoutNotification(payment *model.Payment) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    s.notificationClient.SendPaymentNotification(ctx, &client.SendNotificationRequest{
        MerchantID: payment.MerchantID,
        Type:       "payment_timeout",
        Title:      "支付超时",
        Content:    fmt.Sprintf("订单 %s 已超时自动取消", payment.OrderNo),
        Priority:   "normal",
        Data: map[string]interface{}{
            "payment_no": payment.PaymentNo,
            "order_no":   payment.OrderNo,
            "amount":     payment.Amount,
        },
    })
}
```

#### Cron任务配置

```go
// backend/services/payment-gateway/cmd/main.go

import "github.com/robfig/cron/v3"

func setupCronJobs(timeoutService service.TimeoutService) {
    c := cron.New()

    // 每5分钟执行一次
    c.AddFunc("*/5 * * * *", func() {
        ctx := context.Background()
        if err := timeoutService.ScanExpiredPayments(ctx); err != nil {
            logger.Error("超时订单扫描失败", zap.Error(err))
        }
    })

    c.Start()
    logger.Info("定时任务已启动")
}
```

---

### 模块M: 灾备与高可用

#### 架构设计

```
                         ┌─────────────┐
                         │  Load Balancer  │
                         │   (Kong/Nginx)  │
                         └────────┬────────┘
                                  │
                  ┌───────────────┼───────────────┐
                  │               │               │
              ┌───▼────┐     ┌───▼────┐     ┌───▼────┐
              │ Service │     │ Service │     │ Service │
              │ Pod 1   │     │ Pod 2   │     │ Pod 3   │
              └───┬────┘     └───┬────┘     └───┬────┘
                  │               │               │
                  └───────────────┼───────────────┘
                                  │
                  ┌───────────────┴───────────────┐
                  │                               │
          ┌───────▼────────┐              ┌──────▼─────────┐
          │  PostgreSQL    │              │  Redis Cluster │
          │  Primary       │◄──Repl───────┤  Sentinel      │
          │                │              │                │
          └───────┬────────┘              └────────────────┘
                  │
          ┌───────▼────────┐
          │  PostgreSQL    │
          │  Standby       │
          │  (Failover)    │
          └────────────────┘
```

#### PostgreSQL高可用 (Patroni)

```yaml
# docker-compose-ha.yml

services:
  etcd:
    image: quay.io/coreos/etcd:v3.5.0
    environment:
      ETCD_LISTEN_CLIENT_URLS: http://0.0.0.0:2379
      ETCD_ADVERTISE_CLIENT_URLS: http://etcd:2379
    ports:
      - "2379:2379"

  postgres-primary:
    image: patroni/patroni:latest
    environment:
      PATRONI_NAME: postgres1
      PATRONI_POSTGRESQL_DATA_DIR: /var/lib/postgresql/data
      PATRONI_ETCD_HOSTS: etcd:2379
      PATRONI_SCOPE: payment-cluster
      PATRONI_RESTAPI_LISTEN: 0.0.0.0:8008
      PATRONI_POSTGRESQL_LISTEN: 0.0.0.0:5432
      PATRONI_SUPERUSER_PASSWORD: postgres
      PATRONI_REPLICATION_USERNAME: replicator
      PATRONI_REPLICATION_PASSWORD: replicator
    ports:
      - "40432:5432"
      - "40433:8008"
    volumes:
      - postgres-primary-data:/var/lib/postgresql/data

  postgres-standby:
    image: patroni/patroni:latest
    environment:
      PATRONI_NAME: postgres2
      PATRONI_POSTGRESQL_DATA_DIR: /var/lib/postgresql/data
      PATRONI_ETCD_HOSTS: etcd:2379
      PATRONI_SCOPE: payment-cluster
      PATRONI_RESTAPI_LISTEN: 0.0.0.0:8008
      PATRONI_POSTGRESQL_LISTEN: 0.0.0.0:5432
      PATRONI_SUPERUSER_PASSWORD: postgres
      PATRONI_REPLICATION_USERNAME: replicator
      PATRONI_REPLICATION_PASSWORD: replicator
    ports:
      - "40434:5432"
      - "40435:8008"
    volumes:
      - postgres-standby-data:/var/lib/postgresql/data
```

#### Redis Sentinel

```yaml
redis-master:
  image: redis:7-alpine
  command: redis-server --appendonly yes
  ports:
    - "40379:6379"
  volumes:
    - redis-master-data:/data

redis-slave:
  image: redis:7-alpine
  command: redis-server --appendonly yes --slaveof redis-master 6379
  ports:
    - "40380:6379"
  volumes:
    - redis-slave-data:/data

redis-sentinel:
  image: redis:7-alpine
  command: >
    sh -c "echo 'sentinel monitor mymaster redis-master 6379 2' > /etc/redis/sentinel.conf &&
           echo 'sentinel down-after-milliseconds mymaster 5000' >> /etc/redis/sentinel.conf &&
           echo 'sentinel parallel-syncs mymaster 1' >> /etc/redis/sentinel.conf &&
           echo 'sentinel failover-timeout mymaster 10000' >> /etc/redis/sentinel.conf &&
           redis-sentinel /etc/redis/sentinel.conf"
  ports:
    - "26379:26379"
```

#### 备份策略

```bash
#!/bin/bash
# backend/scripts/backup-postgres.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"

# 全量备份
pg_dump -h localhost -p 40432 -U postgres \
    -F custom -f "$BACKUP_DIR/full_backup_$DATE.dump" \
    payment_platform

# 上传到S3/OSS
aws s3 cp "$BACKUP_DIR/full_backup_$DATE.dump" \
    s3://payment-backups/postgres/

# 删除7天前的备份
find $BACKUP_DIR -name "*.dump" -mtime +7 -delete

# Cron配置: 每日凌晨3点执行
# 0 3 * * * /app/scripts/backup-postgres.sh
```

---

## 📊 数据库设计

### 新增表汇总

```sql
-- ========== 模块A: 对账系统 ==========
CREATE TABLE reconciliation_reports (...);
CREATE TABLE reconciliation_discrepancies (...);

-- ========== 模块B: 拒付管理 ==========
CREATE TABLE disputes (...);
CREATE TABLE dispute_evidence_files (...);

-- ========== 模块C: 商户额度 ==========
ALTER TABLE merchants ADD COLUMN tier VARCHAR(20);
CREATE TABLE merchant_limit_usage (...);

-- ========== 模块D/E: 渠道扩展 (无新表,复用payment_callbacks) ==========

-- ========== 模块H: 超时处理 (无新表,使用existing payment表) ==========
```

### Schema迁移脚本

```sql
-- backend/migrations/V1.1__globalization_modules.sql

-- 1. 对账系统
CREATE TABLE IF NOT EXISTS reconciliation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconcile_date DATE NOT NULL,
    channel VARCHAR(50) NOT NULL,
    total_internal BIGINT NOT NULL,
    total_channel BIGINT NOT NULL,
    matched BIGINT NOT NULL,
    missing BIGINT NOT NULL,
    duplicate BIGINT NOT NULL,
    amount_mismatch BIGINT NOT NULL,
    match_rate DECIMAL(5,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    report_file_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    reviewed_by UUID,
    reviewed_at TIMESTAMP
);

CREATE INDEX idx_reconcile_date ON reconciliation_reports(reconcile_date);
CREATE INDEX idx_reconcile_status ON reconciliation_reports(status);

CREATE TABLE IF NOT EXISTS reconciliation_discrepancies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reconciliation_reports(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    internal_order_no VARCHAR(100),
    channel_order_no VARCHAR(100),
    internal_amount BIGINT,
    channel_amount BIGINT,
    amount_diff BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    resolution TEXT,
    resolved_by UUID,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_discrepancy_report ON reconciliation_discrepancies(report_id);
CREATE INDEX idx_discrepancy_status ON reconciliation_discrepancies(status);

-- 2. 拒付管理
CREATE TABLE IF NOT EXISTS disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_no VARCHAR(100) UNIQUE NOT NULL,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    payment_no VARCHAR(100) NOT NULL,
    merchant_id UUID NOT NULL,
    channel VARCHAR(50) NOT NULL,
    channel_dispute_id VARCHAR(200) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'warning',
    due_date TIMESTAMP,
    evidence_details JSONB,
    submitted_at TIMESTAMP,
    resolved_at TIMESTAMP,
    resolution TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dispute_payment ON disputes(payment_id);
CREATE INDEX idx_dispute_merchant ON disputes(merchant_id);
CREATE INDEX idx_dispute_status ON disputes(status);
CREATE UNIQUE INDEX idx_dispute_channel ON disputes(channel, channel_dispute_id);

CREATE TABLE IF NOT EXISTS dispute_evidence_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    file_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT,
    uploaded_by UUID,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evidence_dispute ON dispute_evidence_files(dispute_id);

-- 3. 商户额度
ALTER TABLE merchants
ADD COLUMN IF NOT EXISTS tier VARCHAR(20) NOT NULL DEFAULT 'starter',
ADD COLUMN IF NOT EXISTS custom_daily_limit BIGINT,
ADD COLUMN IF NOT EXISTS custom_monthly_limit BIGINT,
ADD COLUMN IF NOT EXISTS custom_single_limit BIGINT,
ADD COLUMN IF NOT EXISTS custom_fee_rate INT;

CREATE INDEX IF NOT EXISTS idx_merchant_tier ON merchants(tier);

CREATE TABLE IF NOT EXISTS merchant_limit_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    usage_date DATE NOT NULL,
    tier VARCHAR(20) NOT NULL,
    daily_limit BIGINT NOT NULL,
    daily_used BIGINT NOT NULL DEFAULT 0,
    monthly_limit BIGINT NOT NULL,
    monthly_used BIGINT NOT NULL DEFAULT 0,
    transaction_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (merchant_id, usage_date)
);

-- 4. 支付超时处理 (无新表,确保expired_at字段存在)
ALTER TABLE payments
ADD COLUMN IF NOT EXISTS expired_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_payment_expired ON payments(expired_at)
WHERE status = 'pending' AND expired_at IS NOT NULL;
```

---

## 🔌 API接口设计

### 模块A: 对账系统 API

```
POST   /api/v1/reconciliation/run              # 手动触发对账
GET    /api/v1/reconciliation/reports          # 对账报表列表
GET    /api/v1/reconciliation/reports/:id      # 对账报表详情
GET    /api/v1/reconciliation/discrepancies    # 差异列表
PUT    /api/v1/reconciliation/discrepancies/:id/resolve  # 处理差异
GET    /api/v1/reconciliation/statistics       # 对账统计
```

### 模块B: 拒付管理 API

```
GET    /api/v1/disputes                        # 拒付列表
GET    /api/v1/disputes/:id                    # 拒付详情
POST   /api/v1/disputes/:id/evidence           # 上传证据
POST   /api/v1/disputes/:id/submit             # 提交证据到渠道
GET    /api/v1/disputes/statistics             # 拒付统计
POST   /webhooks/stripe/disputes               # Stripe拒付Webhook
```

### 模块C: 商户额度 API

```
GET    /api/v1/merchants/:id/limits            # 查询商户额度
GET    /api/v1/merchants/:id/limits/usage      # 额度使用情况
PUT    /api/v1/merchants/:id/limits            # 调整额度
GET    /api/v1/merchants/:id/tier              # 查询商户分级
PUT    /api/v1/merchants/:id/tier              # 调整分级
GET    /api/v1/tiers                           # 分级配置列表
```

### 模块D: PayPal API

```
POST   /api/v1/payments/paypal                 # 创建PayPal支付
GET    /api/v1/payments/paypal/:id/capture     # 捕获授权
POST   /api/v1/refunds/paypal                  # PayPal退款
POST   /webhooks/paypal                        # PayPal Webhook
```

### 模块E: 支付宝/微信 API

```
POST   /api/v1/payments/alipay/app             # 支付宝APP支付
POST   /api/v1/payments/alipay/page            # 支付宝网页支付
POST   /api/v1/payments/alipay/qr              # 支付宝扫码支付
POST   /api/v1/payments/wechat/app             # 微信APP支付
POST   /api/v1/payments/wechat/jsapi           # 微信JSAPI支付
POST   /api/v1/payments/wechat/native          # 微信Native支付
POST   /webhooks/alipay                        # 支付宝通知
POST   /webhooks/wechat                        # 微信通知
```

---

## 🚀 部署架构

### Kubernetes部署方案

```yaml
# k8s/payment-gateway-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: payment-gateway
  namespace: payment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: payment-gateway
  template:
    metadata:
      labels:
        app: payment-gateway
    spec:
      containers:
      - name: payment-gateway
        image: payment-platform/payment-gateway:v1.1.0
        ports:
        - containerPort: 40003
        env:
        - name: DB_HOST
          value: "postgres-primary.payment.svc.cluster.local"
        - name: REDIS_HOST
          value: "redis-sentinel.payment.svc.cluster.local"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 40003
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 40003
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: payment-gateway
  namespace: payment
spec:
  selector:
    app: payment-gateway
  ports:
  - protocol: TCP
    port: 40003
    targetPort: 40003
  type: ClusterIP

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: payment-gateway-hpa
  namespace: payment
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: payment-gateway
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 🧪 测试计划

### 单元测试 (200+用例)

| 模块 | 测试用例数 | 覆盖率目标 |
|------|-----------|-----------|
| 对账系统 | 40 | 90% |
| 拒付管理 | 30 | 85% |
| 商户额度 | 35 | 90% |
| PayPal集成 | 25 | 80% |
| 支付宝/微信 | 40 | 80% |
| 超时处理 | 20 | 95% |
| 其他 | 10 | 80% |

### 集成测试场景

```
场景1: 完整支付流程 (Stripe → PayPal → 支付宝 → 微信)
场景2: 退款流程 (部分退款 + 全额退款)
场景3: 拒付流程 (创建 → 上传证据 → 提交 → 结果通知)
场景4: 对账流程 (下载账单 → 匹配 → 生成报表)
场景5: 额度控制 (超日限 → 超月限 → 超单笔限)
场景6: 超时处理 (订单超时 → 自动取消 → 通知)
场景7: 灾备切换 (主库故障 → 自动切换 → 服务恢复)
```

### 压力测试指标

```
目标TPS: 1000 (并发支付请求)
目标延迟: P95 < 500ms, P99 < 1000ms
成功率: > 99.9%
并发商户: 1000+
测试时长: 1小时持续压力
```

---

## ⚠️ 风险与应对

### 技术风险

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| PayPal API变更 | 高 | 中 | 使用稳定版本,定期检查changelog |
| 支付宝/微信限流 | 中 | 高 | 实现请求队列,控制频率 |
| 数据库迁移失败 | 高 | 低 | 详细测试,备份数据,可回滚 |
| Redis集群故障 | 中 | 低 | 降级到单机模式,监控告警 |
| 性能瓶颈 | 中 | 中 | 压测验证,优化慢查询 |

### 业务风险

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| 对账差异过多 | 高 | 中 | 人工介入机制,差异分析工具 |
| 拒付率上升 | 高 | 低 | 风控前置,证据自动收集 |
| 商户超限抱怨 | 中 | 中 | 提前通知,自助申请提额 |
| 新渠道稳定性 | 中 | 中 | 灰度发布,流量控制 |

---

## 📅 里程碑

- ✅ **Week 2**: 基础设施就绪
- ✅ **Week 5**: 核心功能完成 (对账+拒付+额度)
- ✅ **Week 8**: 渠道集成完成 (4大渠道)
- ✅ **Week 10**: 超时处理+灾备部署
- ✅ **Week 11**: 测试完成
- 🎯 **Week 12**: 正式上线

---

## 📞 联系方式

**项目负责人**: [您的名字]
**技术支持**: development@payment-platform.com
**紧急联系**: [电话号码]

---

**文档版本**: v1.0
**最后更新**: 2025-10-25
**审核状态**: 待审核

---

> **下一步行动**:
> 1. 团队评审本计划
> 2. 确认资源分配
> 3. 启动Sprint 1
> 4. 每周进度同步会议

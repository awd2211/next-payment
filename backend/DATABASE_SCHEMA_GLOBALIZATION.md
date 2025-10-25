# 全球化支付平台数据库 Schema 设计

本文档定义全球化方案所需的所有数据库表结构，包括对账系统、拒付管理、商户额度管理等模块。

## 概览

**设计原则**:
- 多租户隔离（每个服务独立数据库）
- ACID 事务保证
- 索引优化查询性能
- JSONB 存储扩展字段
- 时间戳审计

**新增服务**:
1. **reconciliation-service** - 对账系统（新增）
2. **dispute-service** - 拒付管理（新增）
3. **merchant-limit-service** - 商户额度管理（新增）

**扩展现有服务**:
- payment-gateway - 新增超时处理字段
- merchant-service - 新增商户分级字段

---

## 一、对账系统 (reconciliation-service)

### 数据库名称

`payment_reconciliation`

### 1.1 reconciliation_tasks (对账任务表)

**描述**: 每日对账任务记录，记录对账执行状态和结果

```sql
CREATE TABLE reconciliation_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 任务信息
    task_no VARCHAR(64) UNIQUE NOT NULL,           -- 任务编号: REC20250124001
    task_date DATE NOT NULL,                       -- 对账日期
    channel VARCHAR(50) NOT NULL,                  -- 渠道: stripe, paypal, alipay, wechat
    task_type VARCHAR(20) NOT NULL,                -- 类型: daily, manual, reconcile

    -- 统计信息
    platform_count INTEGER DEFAULT 0,              -- 平台订单数
    platform_amount BIGINT DEFAULT 0,              -- 平台总金额（分）
    channel_count INTEGER DEFAULT 0,               -- 渠道订单数
    channel_amount BIGINT DEFAULT 0,               -- 渠道总金额（分）
    matched_count INTEGER DEFAULT 0,               -- 匹配成功数
    matched_amount BIGINT DEFAULT 0,               -- 匹配金额（分）
    diff_count INTEGER DEFAULT 0,                  -- 差异笔数
    diff_amount BIGINT DEFAULT 0,                  -- 差异金额（分）

    -- 状态信息
    status VARCHAR(20) NOT NULL,                   -- 状态: pending, processing, completed, failed
    progress INTEGER DEFAULT 0,                    -- 进度百分比: 0-100
    error_message TEXT,                            -- 错误信息

    -- 文件信息
    channel_file_url VARCHAR(500),                 -- 渠道账单文件URL
    report_file_url VARCHAR(500),                  -- 对账报表URL

    -- 时间戳
    started_at TIMESTAMPTZ,                        -- 开始时间
    completed_at TIMESTAMPTZ,                      -- 完成时间
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_task_date (task_date),
    INDEX idx_channel (channel),
    INDEX idx_status (status),
    INDEX idx_task_date_channel (task_date, channel)
);

-- 任务状态常量
-- pending - 待执行
-- processing - 执行中
-- completed - 已完成
-- failed - 失败
```

### 1.2 reconciliation_records (对账明细表)

**描述**: 对账差异明细记录，用于追踪每笔差异订单

```sql
CREATE TABLE reconciliation_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 关联任务
    task_id UUID NOT NULL REFERENCES reconciliation_tasks(id),
    task_no VARCHAR(64) NOT NULL,

    -- 订单信息
    payment_no VARCHAR(64),                        -- 平台支付单号
    channel_trade_no VARCHAR(128),                 -- 渠道交易号
    order_no VARCHAR(64),                          -- 商户订单号
    merchant_id UUID,                              -- 商户ID

    -- 金额信息
    platform_amount BIGINT,                        -- 平台金额（分）
    channel_amount BIGINT,                         -- 渠道金额（分）
    diff_amount BIGINT,                            -- 差异金额（分）
    currency VARCHAR(10),                          -- 货币类型

    -- 状态信息
    platform_status VARCHAR(20),                   -- 平台状态
    channel_status VARCHAR(20),                    -- 渠道状态
    diff_type VARCHAR(20) NOT NULL,                -- 差异类型: matched, platform_only, channel_only, amount_diff, status_diff
    diff_reason TEXT,                              -- 差异原因

    -- 处理信息
    is_resolved BOOLEAN DEFAULT FALSE,             -- 是否已解决
    resolved_by UUID,                              -- 解决人ID
    resolved_at TIMESTAMPTZ,                       -- 解决时间
    resolution_note TEXT,                          -- 解决说明

    -- 扩展信息
    extra JSONB,                                   -- 扩展字段

    -- 时间戳
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_task_id (task_id),
    INDEX idx_payment_no (payment_no),
    INDEX idx_channel_trade_no (channel_trade_no),
    INDEX idx_diff_type (diff_type),
    INDEX idx_is_resolved (is_resolved),
    INDEX idx_merchant_id (merchant_id)
);

-- 差异类型常量
-- matched - 完全匹配
-- platform_only - 仅平台有记录
-- channel_only - 仅渠道有记录
-- amount_diff - 金额不一致
-- status_diff - 状态不一致
```

### 1.3 channel_settlement_files (渠道账单文件表)

**描述**: 存储从各渠道下载的账单文件信息

```sql
CREATE TABLE channel_settlement_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 文件信息
    file_no VARCHAR(64) UNIQUE NOT NULL,           -- 文件编号
    channel VARCHAR(50) NOT NULL,                  -- 渠道
    settlement_date DATE NOT NULL,                 -- 账单日期
    file_url VARCHAR(500) NOT NULL,                -- 文件URL
    file_size BIGINT,                              -- 文件大小（字节）
    file_hash VARCHAR(64),                         -- 文件哈希（SHA256）

    -- 统计信息
    record_count INTEGER DEFAULT 0,                -- 记录数
    total_amount BIGINT DEFAULT 0,                 -- 总金额（分）
    currency VARCHAR(10),                          -- 货币类型

    -- 状态信息
    status VARCHAR(20) NOT NULL,                   -- 状态: pending, downloaded, parsed, imported

    -- 时间戳
    downloaded_at TIMESTAMPTZ,                     -- 下载时间
    parsed_at TIMESTAMPTZ,                         -- 解析时间
    imported_at TIMESTAMPTZ,                       -- 导入时间
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_channel (channel),
    INDEX idx_settlement_date (settlement_date),
    INDEX idx_status (status),
    INDEX idx_channel_date (channel, settlement_date)
);
```

---

## 二、拒付管理系统 (dispute-service)

### 数据库名称

`payment_dispute`

### 2.1 disputes (拒付工单表)

**描述**: Stripe/PayPal 拒付工单记录

```sql
CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 工单信息
    dispute_no VARCHAR(64) UNIQUE NOT NULL,        -- 拒付单号: DIS20250124001
    channel VARCHAR(50) NOT NULL,                  -- 渠道: stripe, paypal
    channel_dispute_id VARCHAR(128) UNIQUE NOT NULL, -- 渠道拒付ID

    -- 关联支付
    payment_id UUID,                               -- 关联支付ID
    payment_no VARCHAR(64),                        -- 支付单号
    merchant_id UUID NOT NULL,                     -- 商户ID
    order_no VARCHAR(64),                          -- 订单号

    -- 金额信息
    amount BIGINT NOT NULL,                        -- 拒付金额（分）
    currency VARCHAR(10) NOT NULL,                 -- 货币类型

    -- 拒付信息
    reason VARCHAR(100),                           -- 拒付原因: fraudulent, duplicate, product_not_received 等
    status VARCHAR(20) NOT NULL,                   -- 状态: warning_needs_response, needs_response, under_review, won, lost, closed
    evidence_due_by TIMESTAMPTZ,                   -- 证据提交截止时间

    -- 证据信息
    has_evidence BOOLEAN DEFAULT FALSE,            -- 是否有证据
    evidence_submitted_at TIMESTAMPTZ,             -- 证据提交时间
    evidence_details JSONB,                        -- 证据详情

    -- 处理信息
    assigned_to UUID,                              -- 处理人ID
    internal_notes TEXT,                           -- 内部备注
    merchant_notes TEXT,                           -- 商户备注

    -- 结果信息
    resolution VARCHAR(20),                        -- 结果: won, lost, withdrawn
    resolved_at TIMESTAMPTZ,                       -- 解决时间
    resolution_note TEXT,                          -- 结果说明

    -- 扩展信息
    extra JSONB,                                   -- 扩展字段（存储渠道原始数据）

    -- 时间戳
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_channel_dispute_id (channel_dispute_id),
    INDEX idx_payment_no (payment_no),
    INDEX idx_merchant_id (merchant_id),
    INDEX idx_status (status),
    INDEX idx_evidence_due_by (evidence_due_by),
    INDEX idx_created_at (created_at)
);

-- 拒付状态常量
-- warning_needs_response - 预警（需要响应）
-- needs_response - 需要响应
-- under_review - 审核中
-- won - 赢得争议
-- lost - 失去争议
-- closed - 已关闭
```

### 2.2 dispute_evidence (拒付证据表)

**描述**: 商户上传的拒付证据文件

```sql
CREATE TABLE dispute_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 关联拒付
    dispute_id UUID NOT NULL REFERENCES disputes(id),
    dispute_no VARCHAR(64) NOT NULL,

    -- 证据信息
    evidence_type VARCHAR(50) NOT NULL,            -- 证据类型: invoice, receipt, shipping_doc, customer_communication, other
    file_name VARCHAR(255) NOT NULL,               -- 文件名
    file_url VARCHAR(500) NOT NULL,                -- 文件URL
    file_size BIGINT,                              -- 文件大小（字节）
    file_type VARCHAR(50),                         -- 文件类型: pdf, jpg, png 等

    -- 描述信息
    description TEXT,                              -- 证据描述

    -- 上传信息
    uploaded_by UUID,                              -- 上传人ID
    uploaded_by_type VARCHAR(20),                  -- 上传人类型: admin, merchant

    -- 时间戳
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_dispute_id (dispute_id),
    INDEX idx_evidence_type (evidence_type),
    INDEX idx_uploaded_at (uploaded_at)
);
```

### 2.3 dispute_timeline (拒付时间线表)

**描述**: 拒付工单操作历史记录

```sql
CREATE TABLE dispute_timeline (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 关联拒付
    dispute_id UUID NOT NULL REFERENCES disputes(id),
    dispute_no VARCHAR(64) NOT NULL,

    -- 事件信息
    event_type VARCHAR(50) NOT NULL,               -- 事件类型: created, updated, evidence_submitted, resolved 等
    event_description TEXT NOT NULL,               -- 事件描述

    -- 操作人信息
    operator_id UUID,                              -- 操作人ID
    operator_type VARCHAR(20),                     -- 操作人类型: system, admin, merchant
    operator_name VARCHAR(100),                    -- 操作人姓名

    -- 变更信息
    old_value JSONB,                               -- 变更前值
    new_value JSONB,                               -- 变更后值

    -- 时间戳
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_dispute_id (dispute_id),
    INDEX idx_event_type (event_type),
    INDEX idx_created_at (created_at)
);
```

---

## 三、商户额度管理系统 (merchant-limit-service)

### 数据库名称

`payment_merchant_limit`

### 3.1 merchant_tiers (商户等级配置表)

**描述**: 商户等级配置，定义每个等级的额度和费率

```sql
CREATE TABLE merchant_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 等级信息
    tier_code VARCHAR(50) UNIQUE NOT NULL,         -- 等级代码: starter, basic, standard, premium, enterprise
    tier_name VARCHAR(100) NOT NULL,               -- 等级名称
    tier_level INTEGER NOT NULL,                   -- 等级级别: 1-5（数字越大等级越高）

    -- 额度配置
    daily_limit BIGINT NOT NULL,                   -- 日限额（分）
    monthly_limit BIGINT NOT NULL,                 -- 月限额（分）
    single_transaction_limit BIGINT NOT NULL,      -- 单笔限额（分）

    -- 费率配置
    transaction_fee_rate DECIMAL(5,4) NOT NULL,    -- 交易费率: 0.0280 表示 2.80%
    fixed_fee BIGINT DEFAULT 0,                    -- 固定手续费（分）

    -- 功能权限
    features JSONB,                                -- 功能权限: {"pre_auth": true, "batch_payment": true}

    -- 升级条件
    upgrade_conditions JSONB,                      -- 升级条件: {"min_volume": 1000000, "min_transactions": 100}

    -- 状态信息
    is_active BOOLEAN DEFAULT TRUE,                -- 是否启用
    description TEXT,                              -- 等级描述

    -- 时间戳
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_tier_code (tier_code),
    INDEX idx_tier_level (tier_level),
    INDEX idx_is_active (is_active)
);

-- 插入默认等级配置
INSERT INTO merchant_tiers (tier_code, tier_name, tier_level, daily_limit, monthly_limit, single_transaction_limit, transaction_fee_rate, fixed_fee) VALUES
('starter', '入门版', 1, 5000000, 100000000, 100000, 0.0320, 30),      -- $50K/日, $1M/月, $1K/笔, 3.20% + $0.30
('basic', '基础版', 2, 20000000, 500000000, 500000, 0.0290, 30),       -- $200K/日, $5M/月, $5K/笔, 2.90% + $0.30
('standard', '标准版', 3, 100000000, 2000000000, 2000000, 0.0260, 30), -- $1M/日, $20M/月, $20K/笔, 2.60% + $0.30
('premium', '高级版', 4, 500000000, 10000000000, 5000000, 0.0230, 30), -- $5M/日, $100M/月, $50K/笔, 2.30% + $0.30
('enterprise', '企业版', 5, 999999999999, 999999999999, 999999999999, 0.0200, 30); -- 无限额, 2.00% + $0.30
```

### 3.2 merchant_limits (商户额度表)

**描述**: 每个商户的当前额度配置和使用情况

```sql
CREATE TABLE merchant_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 商户信息
    merchant_id UUID UNIQUE NOT NULL,              -- 商户ID
    tier_id UUID NOT NULL REFERENCES merchant_tiers(id), -- 等级ID
    tier_code VARCHAR(50) NOT NULL,                -- 等级代码（冗余字段，便于查询）

    -- 额度配置（可自定义，覆盖等级默认值）
    daily_limit BIGINT NOT NULL,                   -- 日限额（分）
    monthly_limit BIGINT NOT NULL,                 -- 月限额（分）
    single_transaction_limit BIGINT NOT NULL,      -- 单笔限额（分）
    is_custom BOOLEAN DEFAULT FALSE,               -- 是否自定义额度

    -- 当前使用量（由定时任务更新）
    daily_used BIGINT DEFAULT 0,                   -- 今日已用（分）
    monthly_used BIGINT DEFAULT 0,                 -- 本月已用（分）
    last_reset_date DATE,                          -- 上次重置日期

    -- 风控信息
    is_suspended BOOLEAN DEFAULT FALSE,            -- 是否暂停（风控冻结）
    suspend_reason TEXT,                           -- 暂停原因
    suspended_at TIMESTAMPTZ,                      -- 暂停时间
    suspended_by UUID,                             -- 暂停操作人

    -- 时间戳
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_merchant_id (merchant_id),
    INDEX idx_tier_id (tier_id),
    INDEX idx_is_suspended (is_suspended)
);
```

### 3.3 limit_usage_logs (额度使用日志表)

**描述**: 额度使用日志，用于审计和统计

```sql
CREATE TABLE limit_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 商户信息
    merchant_id UUID NOT NULL,

    -- 交易信息
    payment_no VARCHAR(64) NOT NULL,               -- 支付单号
    amount BIGINT NOT NULL,                        -- 交易金额（分）
    currency VARCHAR(10) NOT NULL,                 -- 货币类型

    -- 额度信息（记录当时的额度状态）
    before_daily_used BIGINT NOT NULL,             -- 操作前日使用量
    after_daily_used BIGINT NOT NULL,              -- 操作后日使用量
    before_monthly_used BIGINT NOT NULL,           -- 操作前月使用量
    after_monthly_used BIGINT NOT NULL,            -- 操作后月使用量
    daily_limit BIGINT NOT NULL,                   -- 日限额
    monthly_limit BIGINT NOT NULL,                 -- 月限额

    -- 操作信息
    operation VARCHAR(20) NOT NULL,                -- 操作类型: increase, decrease, reset
    operator_id UUID,                              -- 操作人ID
    operator_type VARCHAR(20),                     -- 操作人类型: system, admin

    -- 时间戳
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- 索引
    INDEX idx_merchant_id (merchant_id),
    INDEX idx_payment_no (payment_no),
    INDEX idx_created_at (created_at),
    INDEX idx_merchant_created (merchant_id, created_at DESC)
);

-- 操作类型常量
-- increase - 增加使用量（创建支付）
-- decrease - 减少使用量（退款、取消）
-- reset - 重置使用量（定时任务）
```

---

## 四、扩展现有服务

### 4.1 merchant-service 扩展

**数据库**: `payment_merchant`

**扩展 merchants 表**:

```sql
-- 新增字段
ALTER TABLE merchants ADD COLUMN tier_code VARCHAR(50) DEFAULT 'starter';
ALTER TABLE merchants ADD COLUMN tier_updated_at TIMESTAMPTZ;
ALTER TABLE merchants ADD COLUMN auto_upgrade BOOLEAN DEFAULT TRUE; -- 是否自动升级

-- 添加索引
CREATE INDEX idx_merchants_tier_code ON merchants(tier_code);
```

### 4.2 payment-gateway 扩展

**数据库**: `payment_gateway`

**扩展 payments 表**:

```sql
-- 新增字段（超时处理相关）
ALTER TABLE payments ADD COLUMN timeout_at TIMESTAMPTZ;           -- 超时时间
ALTER TABLE payments ADD COLUMN timeout_handled BOOLEAN DEFAULT FALSE; -- 是否已处理超时
ALTER TABLE payments ADD COLUMN timeout_handled_at TIMESTAMPTZ;   -- 超时处理时间
ALTER TABLE payments ADD COLUMN timeout_action VARCHAR(20);       -- 超时动作: auto_cancel, notify_only

-- 添加索引
CREATE INDEX idx_payments_timeout_at ON payments(timeout_at) WHERE timeout_handled = FALSE;
CREATE INDEX idx_payments_timeout_handled ON payments(timeout_handled);
```

---

## 五、数据库初始化脚本

### 5.1 创建所有数据库

```bash
#!/bin/bash
# scripts/init-globalization-dbs.sh

PGHOST=${DB_HOST:-localhost}
PGPORT=${DB_PORT:-40432}
PGUSER=${DB_USER:-postgres}
PGPASSWORD=${DB_PASSWORD:-postgres}

echo "Creating globalization databases..."

# 对账系统
psql -h $PGHOST -p $PGPORT -U $PGUSER -c "CREATE DATABASE payment_reconciliation;"

# 拒付管理
psql -h $PGHOST -p $PGPORT -U $PGUSER -c "CREATE DATABASE payment_dispute;"

# 商户额度
psql -h $PGHOST -p $PGPORT -U $PGUSER -c "CREATE DATABASE payment_merchant_limit;"

echo "Databases created successfully!"
```

### 5.2 执行迁移

```bash
# 对账系统
psql -h localhost -p 40432 -U postgres -d payment_reconciliation -f schemas/reconciliation.sql

# 拒付管理
psql -h localhost -p 40432 -U postgres -d payment_dispute -f schemas/dispute.sql

# 商户额度
psql -h localhost -p 40432 -U postgres -d payment_merchant_limit -f schemas/merchant_limit.sql
```

---

## 六、性能优化建议

### 6.1 索引策略

- ✅ 复合索引用于高频查询组合（如 task_date + channel）
- ✅ 部分索引用于条件过滤（如 WHERE timeout_handled = FALSE）
- ✅ 时间戳倒序索引用于最新记录查询

### 6.2 分区策略

对于高频写入表，建议按时间分区：

```sql
-- reconciliation_records 按月分区
CREATE TABLE reconciliation_records_2025_01 PARTITION OF reconciliation_records
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE reconciliation_records_2025_02 PARTITION OF reconciliation_records
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
```

### 6.3 归档策略

- **对账记录**: 保留 1 年，1 年后归档到冷存储
- **拒付工单**: 永久保留（法律要求）
- **额度日志**: 保留 6 个月

---

## 七、总结

### 7.1 新增表统计

| 服务 | 表数量 | 核心表 |
|------|-------|--------|
| reconciliation-service | 3 | reconciliation_tasks, reconciliation_records |
| dispute-service | 3 | disputes, dispute_evidence, dispute_timeline |
| merchant-limit-service | 3 | merchant_tiers, merchant_limits, limit_usage_logs |
| **总计** | **9** | **9** |

### 7.2 扩展表统计

| 服务 | 扩展表 | 新增字段 |
|------|-------|---------|
| merchant-service | merchants | 3 |
| payment-gateway | payments | 4 |

### 7.3 Schema 文件清单

创建以下 SQL 文件用于迁移：

```
backend/schemas/
├── reconciliation.sql        # 对账系统
├── dispute.sql               # 拒付管理
├── merchant_limit.sql        # 商户额度
├── merchant_extension.sql    # merchant-service 扩展
└── payment_extension.sql     # payment-gateway 扩展
```

---

**文档版本**: 1.0.0
**最后更新**: 2025-01-24
**状态**: ✅ Schema 设计完成，待评审

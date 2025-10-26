# Merchant Services Deprecation Strategy

## 执行摘要

**问题**: 是否需要删除旧的 `merchant-config-service` 和 `merchant-limit-service`?

**答案**: **不要立即删除,采用分阶段迁移策略**

---

## 一、当前状态分析

### 1.1 服务清单

| 服务类型 | 服务名 | 端口 | 数据库 | 状态 |
|---------|--------|------|--------|------|
| **旧服务** | merchant-config-service | 40012 | payment_merchant_config | ⚠️ 运行中 |
| **旧服务** | merchant-limit-service | 40022 | payment_merchant_limit | ⚠️ 运行中 |
| **新服务** | merchant-policy-service | 40012 | payment_merchant_policy | ✅ 已完成 |
| **新服务** | merchant-quota-service | 40022 | payment_merchant_quota | ✅ 已完成 |

**端口冲突**: 新旧服务使用相同端口,需要协调!

### 1.2 依赖关系发现

**直接依赖者**:
1. **admin-bff-service** (line 208 in main.go)
   ```go
   merchantConfigBFFHandler := handler.NewMerchantConfigBFFHandler(
       getConfig("MERCHANT_CONFIG_SERVICE_URL", "http://localhost:40012")
   )
   ```
   - 调用 merchant-config-service 的费率和限额配置接口
   - 用于管理员后台的商户配置管理

2. **Kong API Gateway** (scripts/kong-setup.sh)
   - 路由配置指向 merchant-config-service:40012
   - 路由: `/merchant-config-fee`, `/merchant-config-limits`, `/merchant-config-channels`

3. **Service Management Scripts**
   - `scripts/manage-services.sh` - 服务启动/停止脚本
   - `scripts/status-all-services.sh` - 状态检查脚本
   - `scripts/start-all-services.sh` - 批量启动脚本

**间接依赖者** (可能存在):
- payment-gateway (可能调用限额检查)
- merchant-bff-service (商户自服务门户)
- 前端应用 (admin-portal, merchant-portal)

### 1.3 数据分析

**旧服务数据表** (需要迁移):
- `merchant_config_service`:
  - fee_configs (费率配置)
  - transaction_limits (交易限额)
  - channel_configs (渠道配置)

- `merchant_limit_service`:
  - merchant_limits (商户限额)
  - limit_usage_records (使用记录)

**新服务数据表** (目标结构):
- `merchant_policy_service`:
  - merchant_tiers (5个默认层级)
  - merchant_fee_policies (费率策略,按tier或merchant)
  - merchant_limit_policies (限额策略,按tier或merchant)
  - merchant_policy_bindings (商户绑定关系)

- `merchant_quota_service`:
  - merchant_quotas (配额使用情况)
  - quota_usage_logs (操作审计)
  - quota_alerts (预警记录)

---

## 二、迁移策略 (推荐)

### ✅ 策略: 蓝绿部署 + 灰度迁移

采用**零停机**迁移方案,分4个阶段执行:

### Phase 1: 并行运行期 (1-2周)

**目标**: 新旧服务同时运行,验证新服务功能完整性

**操作步骤**:
```bash
# 1. 调整端口配置,避免冲突
# 旧服务继续使用原端口
merchant-config-service: 40012 (保持)
merchant-limit-service: 40022 (保持)

# 新服务使用临时端口
merchant-policy-service: 40112 (临时)
merchant-quota-service: 40122 (临时)

# 2. 启动新服务(临时端口)
cd /home/eric/payment/backend/services/merchant-policy-service
PORT=40112 DB_NAME=payment_merchant_policy go run cmd/main.go

cd /home/eric/payment/backend/services/merchant-quota-service
PORT=40122 DB_NAME=payment_merchant_quota go run cmd/main.go

# 3. 执行数据迁移脚本(后续创建)
./scripts/migrate-merchant-data.sh

# 4. 验证新服务API完整性
./scripts/test-merchant-services.sh
```

**验证清单**:
- [ ] 新服务所有API正常响应
- [ ] 数据迁移脚本完成,数据一致性100%
- [ ] 性能测试通过(响应时间 < 100ms)
- [ ] 新服务监控指标正常(Prometheus + Jaeger)

### Phase 2: 灰度切流期 (1周)

**目标**: 逐步将流量从旧服务切换到新服务

**切流策略**:
```
Week 1: 10% 流量 → 新服务 (测试商户)
Week 2: 50% 流量 → 新服务 (部分生产商户)
Week 3: 100% 流量 → 新服务 (全部商户)
```

**实施方法**:

**方案A: Kong API Gateway 金丝雀路由**
```bash
# Kong 配置权重路由
curl -X POST http://localhost:8001/upstreams \
  -d name=merchant-policy-upstream

# 添加旧服务目标 (90% 权重)
curl -X POST http://localhost:8001/upstreams/merchant-policy-upstream/targets \
  -d target=localhost:40012 \
  -d weight=90

# 添加新服务目标 (10% 权重)
curl -X POST http://localhost:8001/upstreams/merchant-policy-upstream/targets \
  -d target=localhost:40112 \
  -d weight=10
```

**方案B: admin-bff-service 代码级灰度**
```go
// admin-bff-service/cmd/main.go
var merchantConfigServiceURL string
if isGrayTraffic(merchantID) {
    // 10% 流量到新服务
    merchantConfigServiceURL = "http://localhost:40112"
} else {
    // 90% 流量到旧服务
    merchantConfigServiceURL = "http://localhost:40012"
}
```

**监控指标**:
- 错误率对比 (新 vs 旧)
- 响应时间对比
- 数据一致性检查 (双写对比)

### Phase 3: 完全切换期 (1天)

**目标**: 100%流量切换到新服务,旧服务只读模式

**操作步骤**:
```bash
# 1. 更新所有依赖方配置
# admin-bff-service/cmd/main.go
MERCHANT_CONFIG_SERVICE_URL=http://localhost:40112

# Kong路由
curl -X PATCH http://localhost:8001/upstreams/merchant-policy-upstream/targets/{old-target-id} \
  -d weight=0

# 2. 旧服务设为只读模式(可选)
# 修改旧服务handler,拦截所有POST/PUT/DELETE请求

# 3. 监控24小时,确认无问题
```

**回滚预案**:
```bash
# 如果新服务出现问题,立即回滚
MERCHANT_CONFIG_SERVICE_URL=http://localhost:40012  # 恢复旧服务
```

### Phase 4: 下线清理期 (1天)

**目标**: 正式下线旧服务,释放资源

**操作步骤**:
```bash
# 1. 停止旧服务进程
pkill -f merchant-config-service
pkill -f merchant-limit-service

# 2. 新服务切换到正式端口
# 修改新服务配置
PORT=40012  # merchant-policy-service
PORT=40022  # merchant-quota-service

# 重启新服务
systemctl restart merchant-policy-service
systemctl restart merchant-quota-service

# 3. 更新文档和脚本
# 修改 scripts/start-all-services.sh
# 修改 scripts/status-all-services.sh
# 修改 Kong 配置

# 4. 归档旧服务代码
mv services/merchant-config-service services/archive/merchant-config-service-deprecated
mv services/merchant-limit-service services/archive/merchant-limit-service-deprecated

# 5. 保留旧数据库3个月(备份)
# 不要立即删除 payment_merchant_config 和 payment_merchant_limit
# 等待3个月观察期后再删除
```

---

## 三、数据迁移脚本 (待实现)

### 3.1 迁移范围

**merchant-config-service → merchant-policy-service**:
```sql
-- 1. 迁移费率配置到 fee_policies (tier_id 为 NULL,merchant_id 不为空)
INSERT INTO payment_merchant_policy.merchant_fee_policies (
    merchant_id, channel, payment_method, currency,
    fee_type, fee_percentage, fee_fixed, min_fee, max_fee,
    priority, status, effective_date, expiry_date, created_at, updated_at
)
SELECT
    merchant_id,
    channel,
    payment_method,
    currency,
    fee_type,
    fee_percentage,
    fee_fixed,
    min_fee,
    max_fee,
    100 AS priority,  -- 商户级策略高优先级
    status,
    effective_date,
    expiry_date,
    created_at,
    updated_at
FROM payment_merchant_config.fee_configs
WHERE merchant_id IS NOT NULL;

-- 2. 迁移限额配置到 limit_policies
INSERT INTO payment_merchant_policy.merchant_limit_policies (
    merchant_id, channel, currency,
    single_trans_min, single_trans_max,
    daily_limit, monthly_limit,
    priority, status, effective_date, created_at, updated_at
)
SELECT
    merchant_id,
    channel,
    currency,
    single_trans_min,
    single_trans_max,
    daily_limit,
    monthly_limit,
    100 AS priority,
    status,
    effective_date,
    created_at,
    updated_at
FROM payment_merchant_config.transaction_limits
WHERE merchant_id IS NOT NULL;
```

**merchant-limit-service → merchant-quota-service**:
```sql
-- 迁移配额使用情况
INSERT INTO payment_merchant_quota.merchant_quotas (
    merchant_id, currency,
    daily_limit, monthly_limit,
    daily_used, monthly_used,
    last_reset_daily, last_reset_monthly,
    status, version, created_at, updated_at
)
SELECT
    merchant_id,
    currency,
    daily_limit,
    monthly_limit,
    daily_used,
    monthly_used,
    last_reset_daily,
    last_reset_monthly,
    status,
    0 AS version,  -- 初始版本号
    created_at,
    updated_at
FROM payment_merchant_limit.merchant_limits;
```

### 3.2 迁移脚本

创建 `scripts/migrate-merchant-data.sh`:
```bash
#!/bin/bash
# 执行数据迁移
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres < scripts/migrate-merchant-data.sql

# 验证数据一致性
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -c "
SELECT
    'merchant_fee_policies' AS table_name,
    COUNT(*) AS migrated_count
FROM payment_merchant_policy.merchant_fee_policies
WHERE merchant_id IS NOT NULL
UNION ALL
SELECT
    'merchant_limit_policies',
    COUNT(*)
FROM payment_merchant_policy.merchant_limit_policies
WHERE merchant_id IS NOT NULL
UNION ALL
SELECT
    'merchant_quotas',
    COUNT(*)
FROM payment_merchant_quota.merchant_quotas;
"
```

---

## 四、风险评估

### 4.1 风险矩阵

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 新服务API不兼容 | 中 | 高 | Phase 1并行验证,API对比测试 |
| 数据迁移丢失 | 低 | 高 | 迁移前全量备份,双写验证 |
| 性能下降 | 低 | 中 | 负载测试,灰度观察 |
| 依赖方未及时更新 | 中 | 中 | 梳理依赖清单,逐个确认 |
| 回滚失败 | 低 | 高 | 保留旧服务数据库,快速切换 |

### 4.2 应急预案

**场景1: 新服务线上故障**
```bash
# 立即回滚到旧服务
export MERCHANT_CONFIG_SERVICE_URL=http://localhost:40012
systemctl restart admin-bff-service

# Kong流量切回
curl -X PATCH .../targets/{new-service} -d weight=0
curl -X PATCH .../targets/{old-service} -d weight=100
```

**场景2: 数据不一致**
```bash
# 停止新服务写操作
# 重新执行数据迁移脚本
# 对比校验数据
```

---

## 五、时间线 (推荐)

| 阶段 | 时间 | 里程碑 |
|------|------|--------|
| **Week 2 Day 8-10** | 3天 | Phase 1: 数据迁移脚本 + 并行运行 |
| **Week 2 Day 11-14** | 4天 | Phase 2: 灰度切流 10% → 50% → 100% |
| **Week 3 Day 15** | 1天 | Phase 3: 全量切换,旧服务只读 |
| **Week 3 Day 16** | 1天 | Phase 4: 正式下线,归档旧服务 |
| **Month 4** | 3个月后 | 删除旧数据库 (观察期结束) |

---

## 六、决策建议

### ❌ 不推荐: 立即删除旧服务

**原因**:
1. **依赖方未更新**: admin-bff-service 仍在调用旧服务
2. **数据未迁移**: 现有商户配置数据会丢失
3. **无回滚能力**: 新服务出问题无法快速恢复
4. **端口冲突**: 新旧服务争抢相同端口

### ✅ 推荐: 分阶段迁移

**理由**:
1. **零停机**: 业务不受影响
2. **可回滚**: 任何阶段都可以回退
3. **可验证**: 每个阶段有明确验证标准
4. **低风险**: 灰度切流逐步验证

---

## 七、下一步行动 (Week 2)

### 立即执行 (Day 8):
```bash
# 1. 修改新服务端口配置(避免冲突)
# merchant-policy-service: 40012 → 40112
# merchant-quota-service: 40022 → 40122

# 2. 启动新服务(临时端口)
PORT=40112 DB_NAME=payment_merchant_policy go run merchant-policy-service/cmd/main.go
PORT=40122 DB_NAME=payment_merchant_quota go run merchant-quota-service/cmd/main.go

# 3. 插入默认种子数据
docker exec -i payment-postgres psql -U postgres < scripts/seed-merchant-tiers.sql
docker exec -i payment-postgres psql -U postgres < scripts/seed-default-policies.sql

# 4. 测试新服务API
./scripts/test-merchant-services.sh

# 5. 创建数据迁移脚本
vi scripts/migrate-merchant-data.sql
vi scripts/migrate-merchant-data.sh
```

### 本周完成 (Day 9-14):
- [ ] 数据迁移脚本编写 + 测试
- [ ] API兼容性测试 (新旧服务对比)
- [ ] 性能测试 (压测新服务)
- [ ] 依赖方梳理 (找出所有调用方)
- [ ] 灰度方案实施 (Kong或代码级)

---

## 八、总结

**回答用户问题: "之前的微服务不需要删除吗?"**

**答案**:
- ❌ **不要现在删除**
- ✅ **采用4阶段迁移策略**
- ⏰ **预计2-3周完成迁移**
- 🗄️ **旧数据库保留3个月观察期**

**关键原则**:
1. **先迁移,后下线**
2. **灰度切流,逐步验证**
3. **保留回滚能力**
4. **数据安全第一**

---

**文档版本**: v1.0
**创建时间**: 2025-10-26
**下次更新**: Phase 1 完成后更新实际执行情况

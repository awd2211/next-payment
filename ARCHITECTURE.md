# Payment Platform Architecture

> **商业支付平台完整架构** - 30个微服务设计
> 最后更新：2025-10-23

---

## 🎯 架构概览

### 总体规模
- **服务总数**：30个微服务
- **当前已实现**：10个服务
- **即将拆分**：5个服务（从现有服务中拆分）
- **未来扩展**：15个服务（Tier 1-3 新增）
- **数据库**：31个独立数据库
- **端口范围**：8001-8040

### 架构原则
✅ **单一职责**：每个服务专注一个业务领域
✅ **数据库独立**：Database per Service
✅ **领域驱动**：基于DDD划分服务边界
✅ **松耦合**：通过API/事件通信
✅ **可独立部署**：每个服务独立发布

---

## 📦 服务清单（按领域分组）

### 1️⃣ 核心支付域 (Payment Domain)
**职责**：处理支付交易的核心流程

| # | 服务名 | 端口 | 数据库 | 状态 | 优先级 |
|---|--------|------|--------|------|--------|
| 1 | payment-gateway | 8003 | payment_gateway | ✅ 已实现 | P0 |
| 2 | order-service | 8004 | payment_order | ✅ 已实现 | P0 |
| 3 | channel-adapter | 8005 | payment_channel | ✅ 已实现 | P0 |
| 4 | risk-service | 8006 | payment_risk | ✅ 已实现 | P0 |
| 5 | routing-service | 8029 | payment_routing | 🔮 Tier 2 | P2 |

**核心流程**：
```
Merchant → payment-gateway → risk-service → order-service → channel-adapter → Stripe/PayPal
```

---

### 2️⃣ 商户管理域 (Merchant Domain)
**职责**：商户生命周期管理

| # | 服务名 | 端口 | 数据库 | 状态 | 优先级 |
|---|--------|------|--------|------|--------|
| 6 | merchant-service | 8002 | payment_merchant | ✅ 已实现（需拆分） | P0 |
| 7 | merchant-auth-service | 8011 | payment_merchant_auth | 📋 待拆分 | P1 |
| 8 | kyc-service | 8014 | payment_kyc | 📋 待拆分 | P2 |
| 9 | merchant-config-service | 8015 | payment_merchant_config | 📋 待拆分 | P2 |

**拆分说明**：
- **merchant-service**：保留核心商户信息（Merchant, MerchantUser, SettlementAccount）
- **merchant-auth-service**：拆分安全认证（2FA, Login, Session, PasswordHistory）
- **kyc-service**：拆分KYC审核（KYCDocument, BusinessQualification）
- **merchant-config-service**：拆分配置管理（APIKey, ChannelConfig, FeeConfig, TransactionLimit）

---

### 3️⃣ 财务结算域 (Finance Domain)
**职责**：资金流转和结算管理

| # | 服务名 | 端口 | 数据库 | 状态 | 优先级 |
|---|--------|------|--------|------|--------|
| 10 | accounting-service | 8008 | payment_accounting | ✅ 已实现（需拆分） | P0 |
| 11 | settlement-service | 8012 | payment_settlement | 📋 待拆分 | P1 |
| 12 | withdrawal-service | 8013 | payment_withdrawal | 📋 待拆分 | P1 |
| 13 | billing-service | 8023 | payment_billing | 🔮 Tier 1 | P1 |
| 14 | payout-service | 8028 | payment_payout | 🔮 Tier 2 | P2 |
| 15 | reconciliation-service | 8021 | payment_reconciliation | 🔮 Tier 1 | P1 |

**拆分说明**：
- **accounting-service**：保留账务记账（Account, AccountTransaction, DoubleEntry）
- **settlement-service**：拆分结算处理（Settlement，批量结算）
- **withdrawal-service**：拆分提现管理（Withdrawal，审批流程）

---

### 4️⃣ 争议处理域 (Dispute Domain)
**职责**：拒付和争议管理

| # | 服务名 | 端口 | 数据库 | 状态 | 优先级 |
|---|--------|------|--------|------|--------|
| 16 | dispute-service | 8020 | payment_dispute | 🔮 Tier 1 | P1 |

**核心功能**：
- Chargeback处理
- 争议证据上传
- 自动冻结争议金额
- 与accounting-service联动

---

### 5️⃣ 合规监管域 (Compliance Domain)
**职责**：合规和审计

| # | 服务名 | 端口 | 数据库 | 状态 | 优先级 |
|---|--------|------|--------|------|--------|
| 17 | compliance-service | 8022 | payment_compliance | 🔮 Tier 1 | P1 |
| 18 | audit-service | 8025 | payment_audit | 🔮 Tier 1 | P1 |

**核心功能**：
- AML（反洗钱）检查
- KYT（了解你的交易）
- 大额交易自动报告
- 操作审计日志

---

### 6️⃣ 平台管理域 (Platform Domain)
**职责**：平台运营和管理

| # | 服务名 | 端口 | 数据库 | 状态 | 优先级 |
|---|--------|------|--------|------|--------|
| 19 | admin-service | 8001 | payment_admin | ✅ 已实现 | P0 |
| 20 | config-service | 8010 | payment_config | ✅ 已实现 | P0 |
| 21 | notification-service | 8007 | payment_notification | ✅ 已实现 | P0 |
| 22 | analytics-service | 8009 | payment_analytics | ✅ 已实现 | P0 |
| 23 | report-service | 8024 | payment_report | 🔮 Tier 1 | P1 |
| 24 | webhook-service | 8026 | payment_webhook | 🔮 Tier 2 | P2 |

---

### 7️⃣ 高级功能域 (Advanced Features Domain)
**职责**：高级商业功能

| # | 服务名 | 端口 | 数据库 | 状态 | 优先级 |
|---|--------|------|--------|------|--------|
| 25 | subscription-service | 8027 | payment_subscription | 🔮 Tier 2 | P2 |
| 26 | fraud-detection-service | 8030 | payment_fraud | 🔮 Tier 2 | P2 |
| 27 | identity-service | 8031 | payment_identity | 🔮 Tier 2 | P3 |
| 28 | document-service | 8032 | payment_document | 🔮 Tier 2 | P3 |
| 29 | marketplace-service | 8033 | payment_marketplace | 🔮 Tier 3 | P3 |
| 30 | currency-service | 8034 | payment_currency | 🔮 Tier 3 | P3 |

---

## 🗺️ 服务依赖关系图

```
                    ┌─────────────────┐
                    │  Merchant App   │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │ payment-gateway │
                    └────┬────┬───┬───┘
                         │    │   │
           ┌─────────────┤    │   └──────────────┐
           │             │    │                  │
    ┌──────▼──────┐ ┌───▼────▼───┐     ┌───────▼────────┐
    │ risk-service│ │order-service│     │channel-adapter │
    └─────────────┘ └──────┬──────┘     └───────┬────────┘
                           │                    │
              ┌────────────┼────────────────────┤
              │            │                    │
     ┌────────▼───────┐ ┌─▼──────────────┐ ┌──▼─────────┐
     │accounting-service│ │settlement-service│ │ Stripe API │
     └──────────────────┘ └──────────────────┘ └────────────┘
```

---

## 📊 优先级定义

| 优先级 | 说明 | 时间线 |
|--------|------|--------|
| P0 | 核心功能，已实现 | 当前 |
| P1 | 必需功能，立即启动 | 0-3个月 |
| P2 | 重要功能，计划中 | 3-6个月 |
| P3 | 高级功能，长期规划 | 6-12个月 |

---

## 🔧 技术栈

### 服务层
- **语言**：Go 1.21+
- **框架**：Gin (HTTP), gRPC (服务间通信)
- **依赖管理**：Go Workspace

### 数据层
- **关系数据库**：PostgreSQL 15
- **缓存**：Redis 7
- **消息队列**：Kafka (事件驱动)

### 基础设施
- **服务发现**：Consul / Nacos（待引入）
- **API网关**：Kong / APISIX（待引入）
- **容器编排**：Kubernetes（待引入）

### 可观测性
- **指标监控**：Prometheus + Grafana
- **分布式追踪**：Jaeger
- **日志聚合**：待引入（ELK/Loki）
- **告警**：Alertmanager

---

## 🚀 实施路线图

### Q1（0-3个月）：核心拆分
1. merchant-auth-service（2周）
2. settlement-service（3周）
3. withdrawal-service（4周）
4. dispute-service（3周）

### Q2（3-6个月）：业务完善
5. kyc-service（3周）
6. merchant-config-service（3周）
7. reconciliation-service（4周）
8. billing-service（4周）

### Q3（6-9个月）：高级功能
9. subscription-service（3周）
10. routing-service（3周）
11. fraud-detection-service（4周）
12. webhook-service（2周）

### Q4（9-12个月）：基础设施
13. 引入服务发现（Consul）
14. 引入API网关（Kong）
15. Kubernetes部署
16. 完善监控告警

---

## 📝 命名规范

### 服务命名
```
{domain}-service
例如：payment-gateway, merchant-auth-service
```

### 数据库命名
```
payment_{domain}
例如：payment_merchant_auth, payment_settlement
```

### 端口分配
```
8001-8010: 当前已实现服务
8011-8020: 拆分服务
8021-8030: Tier 1 必需服务
8031-8040: Tier 2-3 高级服务
```

---

## 🔗 相关文档

- [SERVICE_PORTS.md](./backend/docs/SERVICE_PORTS.md) - 端口分配明细
- [ROADMAP.md](./ROADMAP.md) - 详细实施路线图
- [CLAUDE.md](./CLAUDE.md) - 开发指南（Claude Code专用）

---

## 📧 联系方式

如有架构问题，请联系架构团队。

---

**文档版本**：v1.0
**维护人**：架构团队
**审核状态**：待审核

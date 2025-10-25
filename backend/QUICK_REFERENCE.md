# Payment Platform Quick Reference

**快速参考 | 开发者速查表**

---

## 🚀 快速启动

### 启动基础设施

```bash
# 1. 启动 Docker 容器（PostgreSQL, Redis, Kafka等）
cd /home/eric/payment
docker-compose up -d

# 2. 验证基础设施
docker-compose ps
```

### 启动所有微服务

```bash
# 方式1: 启动所有19个服务
cd /home/eric/payment/backend
./scripts/start-all-services.sh

# 方式2: 只启动 Sprint 2 服务
./scripts/manage-sprint2-services.sh start

# 方式3: 单独启动某个服务（开发模式）
cd services/payment-gateway
air -c .air.toml
```

---

## 📋 服务清单（19个）

| 服务 | 端口 | 数据库 | 功能 |
|------|------|--------|------|
| admin-service | 40001 | payment_admin | 平台管理 |
| merchant-service | 40002 | payment_merchant | 商户管理 |
| payment-gateway | 40003 | payment_gateway | 支付网关 |
| order-service | 40004 | payment_order | 订单管理 |
| channel-adapter | 40005 | payment_channel | 支付通道 |
| risk-service | 40006 | payment_risk | 风控评估 |
| accounting-service | 40007 | payment_accounting | 复式记账 |
| notification-service | 40008 | payment_notify | 通知推送 |
| analytics-service | 40009 | payment_analytics | 数据分析 |
| config-service | 40010 | payment_config | 配置管理 |
| merchant-auth-service | 40011 | payment_merchant_auth | 商户认证 |
| merchant-config-service | 40012 | payment_merchant_config | 商户配置 |
| settlement-service | 40013 | payment_settlement | 结算处理 |
| withdrawal-service | 40014 | payment_withdrawal | 提现管理 |
| kyc-service | 40015 | payment_kyc | KYC验证 |
| cashier-service | 40016 | payment_cashier | 收银台 |
| **reconciliation-service** | **40020** | **payment_reconciliation** | **对账系统** |
| **dispute-service** | **40021** | **payment_dispute** | **拒付管理** |
| **merchant-limit-service** | **40022** | **payment_merchant_limit** | **额度管理** |

---

## 🔧 常用命令速查

| 任务 | 命令 |
|------|------|
| 启动所有服务 | `./scripts/start-all-services.sh` |
| 查看服务状态 | `./scripts/status-all-services.sh` |
| 停止所有服务 | `./scripts/stop-all-services.sh` |
| 检查一致性 | `./scripts/check-consistency.sh` |
| 初始化数据库 | `./scripts/init-db.sh` |
| Sprint 2 管理 | `./scripts/manage-sprint2-services.sh {start\|stop\|status\|logs}` |

---

## 📚 主要文档

1. **[MICROSERVICE_UNIFIED_PATTERNS.md](MICROSERVICE_UNIFIED_PATTERNS.md)** - 统一架构模式（必读）
2. **[SERVICE_PORTS.md](SERVICE_PORTS.md)** - 端口分配表
3. **[SPRINT2_BACKEND_COMPLETE.md](SPRINT2_BACKEND_COMPLETE.md)** - Sprint 2 技术文档
4. **[CONSISTENCY_FINAL_REPORT.md](CONSISTENCY_FINAL_REPORT.md)** - 一致性报告

---

## 🌐 监控端点

- **Prometheus**: http://localhost:40090
- **Grafana**: http://localhost:40300 (admin/admin)
- **Jaeger**: http://localhost:40686
- **Health Check**: http://localhost:PORT/health
- **Metrics**: http://localhost:PORT/metrics

---

**最后更新**: 2025-01-20 | **服务数量**: 19

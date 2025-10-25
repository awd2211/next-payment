# Service Ports Allocation

> **支付平台服务端口分配表**
> 最后更新：2025-10-23

---

## 📋 端口分配总表

| 端口 | 服务名 | 数据库 | 状态 | 启动命令 |
|------|--------|--------|------|----------|
| **8001** | admin-service | payment_admin | ✅ 运行中 | `PORT=8001 go run ./services/admin-service/cmd` |
| **8002** | merchant-service | payment_merchant | ✅ 运行中 | `PORT=8002 go run ./services/merchant-service/cmd` |
| **8003** | payment-gateway | payment_gateway | ✅ 运行中 | `PORT=8003 go run ./services/payment-gateway/cmd` |
| **8004** | order-service | payment_order | ✅ 运行中 | `PORT=8004 go run ./services/order-service/cmd` |
| **8005** | channel-adapter | payment_channel | ✅ 运行中 | `PORT=8005 go run ./services/channel-adapter/cmd` |
| **8006** | risk-service | payment_risk | ✅ 运行中 | `PORT=8006 go run ./services/risk-service/cmd` |
| **8007** | notification-service | payment_notification | ✅ 运行中 | `PORT=8007 go run ./services/notification-service/cmd` |
| **8008** | accounting-service | payment_accounting | ✅ 运行中 | `PORT=8008 go run ./services/accounting-service/cmd` |
| **8009** | analytics-service | payment_analytics | ✅ 运行中 | `PORT=8009 go run ./services/analytics-service/cmd` |
| **8010** | config-service | payment_config | ✅ 运行中 | `PORT=8010 go run ./services/config-service/cmd` |
| **8011** | merchant-auth-service | payment_merchant_auth | 📋 预留（待拆分） | `PORT=8011 go run ./services/merchant-auth-service/cmd` |
| **8012** | settlement-service | payment_settlement | 📋 预留（待拆分） | `PORT=8012 go run ./services/settlement-service/cmd` |
| **8013** | withdrawal-service | payment_withdrawal | 📋 预留（待拆分） | `PORT=8013 go run ./services/withdrawal-service/cmd` |
| **8014** | kyc-service | payment_kyc | 📋 预留（待拆分） | `PORT=8014 go run ./services/kyc-service/cmd` |
| **8015** | merchant-config-service | payment_merchant_config | 📋 预留（待拆分） | `PORT=8015 go run ./services/merchant-config-service/cmd` |
| **8020** | dispute-service | payment_dispute | 🔮 预留（Tier 1） | `PORT=8020 go run ./services/dispute-service/cmd` |
| **8021** | reconciliation-service | payment_reconciliation | 🔮 预留（Tier 1） | `PORT=8021 go run ./services/reconciliation-service/cmd` |
| **8022** | compliance-service | payment_compliance | 🔮 预留（Tier 1） | `PORT=8022 go run ./services/compliance-service/cmd` |
| **8023** | billing-service | payment_billing | 🔮 预留（Tier 1） | `PORT=8023 go run ./services/billing-service/cmd` |
| **8024** | report-service | payment_report | 🔮 预留（Tier 1） | `PORT=8024 go run ./services/report-service/cmd` |
| **8025** | audit-service | payment_audit | 🔮 预留（Tier 1） | `PORT=8025 go run ./services/audit-service/cmd` |
| **8026** | webhook-service | payment_webhook | 🔮 预留（Tier 2） | `PORT=8026 go run ./services/webhook-service/cmd` |
| **8027** | subscription-service | payment_subscription | 🔮 预留（Tier 2） | `PORT=8027 go run ./services/subscription-service/cmd` |
| **8028** | payout-service | payment_payout | 🔮 预留（Tier 2） | `PORT=8028 go run ./services/payout-service/cmd` |
| **8029** | routing-service | payment_routing | 🔮 预留（Tier 2） | `PORT=8029 go run ./services/routing-service/cmd` |
| **8030** | fraud-detection-service | payment_fraud | 🔮 预留（Tier 2） | `PORT=8030 go run ./services/fraud-detection-service/cmd` |
| **8031** | identity-service | payment_identity | 🔮 预留（Tier 2） | `PORT=8031 go run ./services/identity-service/cmd` |
| **8032** | document-service | payment_document | 🔮 预留（Tier 2） | `PORT=8032 go run ./services/document-service/cmd` |
| **8033** | marketplace-service | payment_marketplace | 🔮 预留（Tier 3） | `PORT=8033 go run ./services/marketplace-service/cmd` |
| **8034** | currency-service | payment_currency | 🔮 预留（Tier 3） | `PORT=8034 go run ./services/currency-service/cmd` |

---

## 🔧 基础设施端口

| 端口 | 服务 | 说明 |
|------|------|------|
| **40432** | PostgreSQL | 数据库（docker） |
| **40379** | Redis | 缓存（docker） |
| **40092** | Kafka | 消息队列（docker） |
| **40090** | Prometheus | 指标监控 |
| **40300** | Grafana | 可视化仪表盘（admin/admin） |
| **50686** | Jaeger UI | 分布式追踪 |

---

## 🌐 环境变量配置

### 本地开发环境（Development）
```bash
# 服务端口
export PORT=8001  # 根据服务修改

# 数据库配置
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=payment_admin  # 根据服务修改
export DB_SSL_MODE=disable
export DB_TIMEZONE=UTC

# Redis配置
export REDIS_HOST=localhost
export REDIS_PORT=40379
export REDIS_PASSWORD=
export REDIS_DB=0

# 服务间调用（示例）
export MERCHANT_AUTH_SERVICE_URL=http://localhost:8011
export SETTLEMENT_SERVICE_URL=http://localhost:8012
export WITHDRAWAL_SERVICE_URL=http://localhost:8013
```

### Docker环境
```bash
# 数据库配置（通过docker network）
export DB_HOST=postgres
export DB_PORT=5432

# Redis配置
export REDIS_HOST=redis
export REDIS_PORT=6379
```

---

## 📝 端口使用规范

### 1. 端口范围分配
- **8001-8010**：当前已实现的10个服务
- **8011-8015**：从现有服务拆分的5个服务
- **8016-8019**：预留（未来拆分）
- **8020-8025**：Tier 1 必需服务（6个）
- **8026-8032**：Tier 2 重要服务（7个）
- **8033-8040**：Tier 3 高级服务（8个）

### 2. 端口冲突检查
```bash
# 检查端口是否被占用
lsof -i :8001

# 查看所有服务进程
ps aux | grep "go run"

# 停止特定端口的服务
kill $(lsof -t -i:8001)
```

### 3. Health Check端点
所有服务统一使用以下健康检查端点：
```
GET http://localhost:{PORT}/health

Response:
{
  "status": "ok",
  "service": "service-name",
  "time": 1729728000
}
```

---

## 🔗 服务发现配置（未来）

### Consul配置示例
```json
{
  "service": {
    "name": "merchant-auth-service",
    "port": 8011,
    "tags": ["auth", "merchant"],
    "check": {
      "http": "http://localhost:8011/health",
      "interval": "10s",
      "timeout": "2s"
    }
  }
}
```

---

## 📊 端口监控

### Prometheus抓取配置
```yaml
scrape_configs:
  - job_name: 'payment-services'
    static_configs:
      - targets:
        - 'localhost:8001'  # admin-service
        - 'localhost:8002'  # merchant-service
        - 'localhost:8003'  # payment-gateway
        # ... 其他服务
```

---

## 🚨 注意事项

1. ⚠️ **端口冲突**：启动新服务前，确认端口未被占用
2. ⚠️ **防火墙**：生产环境需要配置防火墙规则
3. ⚠️ **端口转发**：Docker容器需要正确映射端口
4. ⚠️ **负载均衡**：生产环境建议使用负载均衡器（Nginx/HAProxy）
5. ⚠️ **端口预留**：不要随意修改已预留的端口号

---

## 🔄 端口变更流程

如需修改端口分配，请遵循以下流程：

1. 在本文档中标记变更（附带原因）
2. 更新相关服务的环境变量配置
3. 更新docker-compose.yml
4. 更新Prometheus配置
5. 通知团队成员
6. 提交PR并审核

---

## 📞 联系方式

端口分配问题请联系：架构团队

---

**文档版本**：v1.0
**维护人**：架构团队
**最后更新**：2025-10-23

# Service Port Allocation

**Last Updated**: 2025-01-20
**Total Services**: 19

---

## Port Assignment Table

| Port | Service Name | Database | Phase |
|------|-------------|----------|-------|
| 40001 | admin-service | payment_admin | Phase 1 |
| 40002 | merchant-service | payment_merchant | Phase 1 |
| 40003 | payment-gateway | payment_gateway | Phase 1 |
| 40004 | order-service | payment_order | Phase 1 |
| 40005 | channel-adapter | payment_channel | Phase 1 |
| 40006 | risk-service | payment_risk | Phase 1 |
| 40007 | accounting-service | payment_accounting | Phase 2 |
| 40008 | notification-service | payment_notify | Phase 2 |
| 40009 | analytics-service | payment_analytics | Phase 2 |
| 40010 | config-service | payment_config | Phase 1 |
| 40011 | merchant-auth-service | payment_merchant_auth | Phase 3 |
| 40012 | merchant-config-service | payment_merchant_config | Phase 3 |
| 40013 | settlement-service | payment_settlement | Phase 3 |
| 40014 | withdrawal-service | payment_withdrawal | Phase 3 |
| 40015 | kyc-service | payment_kyc | Phase 3 |
| 40016 | cashier-service | payment_cashier | Phase 3 |
| **40020** | **reconciliation-service** | **payment_reconciliation** | **Sprint 2** |
| **40021** | **dispute-service** | **payment_dispute** | **Sprint 2** |
| **40022** | **merchant-limit-service** | **payment_merchant_limit** | **Sprint 2** |

---

## Port Ranges

- **40001-40015**: Core services (Phase 1-3)
- **40016**: Cashier service (Phase 3)
- **40020-40029**: Sprint 2 globalization services (future expansion)

---

## Infrastructure Ports

| Port | Service | Usage |
|------|---------|-------|
| 40090 | Prometheus | Metrics collection |
| 40300 | Grafana | Monitoring dashboard (admin/admin) |
| 40379 | Redis | Cache and session store |
| 40432 | PostgreSQL | Main database (19 databases) |
| 40092 | Kafka | Message broker |
| 50686 | Jaeger UI | Distributed tracing |
| **40561** | **Kibana** | **Log analysis and visualization** |
| **40920** | **Elasticsearch** | **Log storage and search (HTTP)** |
| **40930** | **Elasticsearch** | **TCP transport** |
| **40514** | **Logstash** | **TCP log input** |
| **40515** | **Logstash** | **UDP log input** |
| **40944** | **Logstash** | **Monitoring API** |

---

## Frontend Ports

| Port | Application | Tech Stack |
|------|-------------|-----------|
| 5173 | admin-portal | React + Vite + Ant Design |
| 5174 | merchant-portal | React + Vite + Ant Design |
| 5175 | website | React + Vite + Ant Design |

---

## Reserved Ports

- **40017-40019**: Reserved (previously planned for Sprint 2, adjusted to 40020-40022)
- **40023-40029**: Reserved for future services
- **40030-40039**: Reserved for additional features

---

## Port Conflict Resolution

**Issue**: Original Sprint 2 ports (40016-40018) conflicted with cashier-service (40016)

**Resolution**: Moved Sprint 2 services to 40020-40022 range

**Changed Services**:
- reconciliation-service: 40016 → **40020** ✅
- dispute-service: 40017 → **40021** ✅
- merchant-limit-service: 40018 → **40022** ✅

---

## Verification

Check all ports are listening:

```bash
# All microservices
for port in 40001 40002 40003 40004 40005 40006 40007 40008 40009 40010 40011 40012 40013 40014 40015 40016 40020 40021 40022; do
  echo -n "Port $port: "
  if lsof -i:$port -sTCP:LISTEN >/dev/null 2>&1; then
    echo "✅ LISTENING"
  else
    echo "❌ NOT LISTENING"
  fi
done

# Infrastructure
for port in 40090 40300 40379 40432 40092 50686; do
  echo -n "Port $port: "
  if lsof -i:$port -sTCP:LISTEN >/dev/null 2>&1; then
    echo "✅ LISTENING"
  else
    echo "❌ NOT LISTENING"
  fi
done
```

---

## API Endpoints

All services expose standard endpoints:

- `http://localhost:PORT/health` - Health check
- `http://localhost:PORT/metrics` - Prometheus metrics
- `http://localhost:PORT/api/v1/*` - RESTful API

**Examples**:
- Reconciliation Service: http://localhost:40020/api/v1/reconciliation/tasks
- Dispute Service: http://localhost:40021/api/v1/disputes
- Merchant Limit Service: http://localhost:40022/api/v1/limits

---

## mTLS Configuration

All services support mTLS when `ENABLE_MTLS=true`:

```bash
# Test with mTLS
curl -v https://localhost:40020/health \
  --cacert certs/ca/ca-cert.pem \
  --cert certs/services/reconciliation-service/cert.pem \
  --key certs/services/reconciliation-service/key.pem
```

---

## Notes

1. **No port conflicts** - All 19 services have unique ports
2. **Sequential numbering** - Ports 40001-40016 are sequential (except 40016 for cashier)
3. **Sprint 2 range** - Ports 40020+ reserved for new features
4. **Infrastructure isolated** - Infrastructure uses different port ranges (40090+, 40300+, etc.)
5. **Frontends isolated** - Frontend apps use 5xxx range

---

## Change History

### 2025-01-20
- ✅ Added Sprint 2 services (reconciliation, dispute, merchant-limit)
- ✅ Resolved port conflict with cashier-service
- ✅ Updated all scripts and documentation
- ✅ Verified no duplicates across 19 services

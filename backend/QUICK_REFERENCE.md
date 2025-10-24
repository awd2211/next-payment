# Quick Reference - API Gateway & Service Discovery

## 📋 Executive Summary

**Decision**: APISIX + Consul
**Timeline**: 5 weeks (25 working days)
**Team**: 2-3 developers
**Status**: ✅ Planning Complete, Ready for Implementation

---

## 🎯 Selected Solutions

### API Gateway: Apache APISIX
- **Why**: Cloud-native, high performance (45k+ req/s), etcd-based configuration
- **Port**: 40080 (HTTP), 40443 (HTTPS), 40091 (Admin API)
- **Dashboard**: http://localhost:40000 (admin/admin)

### Service Discovery: Consul
- **Why**: Mature, excellent Go SDK, production-proven
- **Port**: 40500 (HTTP), 40600 (DNS), 48300-48302 (Gossip)
- **UI**: http://localhost:40500/ui

---

## 🚀 Immediate Next Steps

### Step 1: Team Review Meeting (This Week)
- [ ] Review [API_GATEWAY_AND_SERVICE_DISCOVERY_PLAN.md](./API_GATEWAY_AND_SERVICE_DISCOVERY_PLAN.md)
- [ ] Confirm solution selection (APISIX + Consul)
- [ ] Assign 2-3 developers
- [ ] Request test environment resources

### Step 2: Deploy Consul (Week 1, Days 1-2)
```bash
# Add to docker-compose.yml (already provided in plan)
docker-compose up -d consul
docker-compose ps consul

# Verify Consul UI
curl http://localhost:40500/v1/status/leader
```

### Step 3: First Service Integration (Week 1, Days 3-5)
```bash
# Modify payment-gateway to register with Consul
cd /home/eric/payment/backend/services/payment-gateway
# Follow code examples in API_GATEWAY_AND_SERVICE_DISCOVERY_PLAN.md
```

---

## 📊 Architecture Changes

### Before (Current)
```
Client → payment-gateway:40003 → order-service:40004
Client → merchant-service:40002 → risk-service:40006
(15 separate endpoints, hardcoded URLs)
```

### After (Target)
```
Client → APISIX:40080 → Consul DNS → service instances
       ↓
     Route rules, rate limiting, authentication
       ↓
     Dynamic service discovery
```

---

## 🔧 Key Configuration Files

### 1. Consul Registration (Bootstrap Framework)
**File**: `backend/pkg/app/bootstrap.go`
```go
// Add Consul registration capability
consulConfig := consul.DefaultConfig()
consulConfig.Address = "localhost:40500"
consulClient, _ := consul.NewClient(consulConfig)

registration := &consul.AgentServiceRegistration{
    ID:      fmt.Sprintf("%s-%d", cfg.ServiceName, cfg.Port),
    Name:    cfg.ServiceName,
    Port:    cfg.Port,
    Address: "localhost",
    Check: &consul.AgentServiceCheck{
        HTTP:     fmt.Sprintf("http://localhost:%d/health", cfg.Port),
        Interval: "10s",
        Timeout:  "3s",
    },
}
consulClient.Agent().ServiceRegister(registration)
```

### 2. Service Discovery Client
**File**: `backend/pkg/discovery/consul_client.go`
```go
func (c *ConsulClient) Discover(serviceName string) ([]string, error) {
    services, _, err := c.client.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return nil, err
    }

    var addresses []string
    for _, service := range services {
        addr := fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port)
        addresses = append(addresses, addr)
    }
    return addresses, nil
}
```

### 3. APISIX Route Configuration
**File**: Route for payment-gateway
```bash
curl http://localhost:40091/apisix/admin/routes/1 \
-H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' \
-X PUT -d '{
  "uri": "/api/v1/payments/*",
  "upstream": {
    "type": "roundrobin",
    "discovery_type": "consul",
    "service_name": "payment-gateway"
  },
  "plugins": {
    "limit-req": {
      "rate": 100,
      "burst": 50
    }
  }
}'
```

---

## 📈 Success Metrics

### Phase 1 (Week 2)
- ✅ All 15 services registered in Consul
- ✅ Health checks showing green in Consul UI
- ✅ Zero manual configuration changes needed

### Phase 2 (Week 4)
- ✅ 100% traffic through APISIX
- ✅ API latency < 50ms (P95)
- ✅ Rate limiting active on all routes

### Phase 3 (Week 5)
- ✅ Frontend using APISIX endpoints
- ✅ Service discovery working (scale test)
- ✅ Monitoring dashboards complete

---

## 💰 Cost Estimation

- **Development**: 25 days × 2-3 developers = 50-75 person-days
- **Infrastructure**:
  - Consul: 1 node = 2GB RAM, 1 CPU
  - APISIX: 1 node = 2GB RAM, 2 CPU
  - etcd: 1 node = 2GB RAM, 1 CPU
  - **Total**: ~6GB RAM, 4 CPU (staging)
- **Production**: 3x replicas = 18GB RAM, 12 CPU

---

## ⚠️ Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| APISIX learning curve | Start with simple routes, gradual migration |
| Consul network issues | Deploy single-node first, test health checks |
| Service discovery latency | Cache lookups, TTL=30s |
| Frontend compatibility | Test each portal separately |
| Production rollout | Blue-green deployment, gradual traffic shift |

---

## 📚 Documentation References

1. **Full Implementation Plan**: [API_GATEWAY_AND_SERVICE_DISCOVERY_PLAN.md](./API_GATEWAY_AND_SERVICE_DISCOVERY_PLAN.md)
2. **APISIX Docs**: https://apisix.apache.org/docs/apisix/getting-started/
3. **Consul Docs**: https://developer.hashicorp.com/consul/docs
4. **Current Architecture**: [CLAUDE.md](../CLAUDE.md) - Service Ports section

---

## 📞 Contact

**Plan Author**: Claude Code
**Date**: 2025-10-24
**Version**: 1.0
**Status**: ✅ Ready for Team Review

---

## 🎬 Quick Start Commands (After Team Approval)

```bash
# 1. Deploy infrastructure
cd /home/eric/payment/backend
docker-compose up -d consul etcd apisix apisix-dashboard

# 2. Verify services
docker-compose ps | grep -E "consul|apisix|etcd"

# 3. Access UIs
# Consul: http://localhost:40500/ui
# APISIX Dashboard: http://localhost:40000 (admin/admin)

# 4. Test health check
curl http://localhost:40500/v1/status/leader

# 5. Proceed with Week 1 tasks in main plan
```

---

**Next Action**: Schedule team review meeting to discuss this plan and get approval to start Week 1 implementation. 🚀

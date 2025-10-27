# Authentication Architecture Guide

## Overview

The payment platform uses a **two-tier authentication model**:

1. **BFF Layer**: Admin BFF and Merchant BFF handle user authentication via JWT
2. **Internal Services**: Most internal services do NOT require authentication as they are called through the BFF

---

## Authentication Tiers

### Tier 1: Frontend-Facing Services (Require Authentication)

These services handle direct requests from frontend applications and implement full authentication:

| Service | Port | Authentication Method | Authorization |
|---------|------|----------------------|---------------|
| **admin-bff-service** | 40001 | JWT + 2FA | RBAC (6 roles) |
| **merchant-bff-service** | 40023 | JWT | Tenant Isolation |
| **payment-gateway** | 40003 | JWT + Signature | Merchant API Keys |
| **cashier-service** | 40016 | JWT | Basic Auth |
| **merchant-auth-service** | 40011 | JWT | Self (auth service) |

**Authentication Flow**:
```
Frontend → BFF Service → JWT Validation → RBAC Check → Internal Service Call
```

### Tier 2: Internal Services (No Authentication)

These services are called ONLY by other backend services (especially through BFF) and do NOT require authentication at the HTTP route level:

| Service | Port | Security Model | Called By |
|---------|------|----------------|-----------|
| accounting-service | 40007 | Trusted network | BFF, payment-gateway |
| analytics-service | 40009 | Trusted network | BFF, Kafka consumers |
| channel-adapter | 40005 | Trusted network | payment-gateway |
| config-service | 40010 | Trusted network | All services |
| dispute-service | 40021 | Trusted network | BFF |
| kyc-service | 40015 | Trusted network | BFF |
| notification-service | 40008 | Trusted network | BFF, all services |
| order-service | 40004 | Trusted network | payment-gateway, BFF |
| reconciliation-service | 40020 | Trusted network | BFF |
| risk-service | 40006 | Trusted network | payment-gateway |
| settlement-service | 40013 | Trusted network | BFF |
| withdrawal-service | 40014 | Trusted network | BFF, settlement-service |
| merchant-policy-service | 40012 | Trusted network | BFF |
| merchant-quota-service | 40022 | Trusted network | BFF |

**Security Assumptions**:
- These services are deployed in a private VPC or Kubernetes cluster
- Network policies restrict access to these services from outside the cluster
- Service-to-service communication happens over private network
- Optionally can add mTLS for service-to-service authentication in production

---

## Issue: Unused JWT Managers in Internal Services

### Current Situation (HIGH-003)

Many internal services (11 services) create JWT manager instances but never use them:

```go
// In cmd/main.go
jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
_ = jwtManager // 预留给需要认证的路由使用
```

**Services Affected**:
- accounting-service
- analytics-service
- channel-adapter
- config-service
- dispute-service
- kyc-service
- notification-service
- order-service
- reconciliation-service
- risk-service
- withdrawal-service

### Why This Happens

**Root Cause**: These services were scaffolded with JWT manager code in anticipation of direct frontend access, but the BFF pattern was adopted instead.

### Impact

1. **Memory Waste**: ~100KB per service for unused JWT manager
2. **Code Confusion**: Developers may think routes are protected when they're not
3. **False Security**: Creates impression of authentication without actual enforcement

### Recommendation

Choose one of the following approaches:

#### Option A: Remove Unused JWT Managers (Recommended)

**For internal services that will NEVER have direct frontend access:**

```go
// Remove these lines:
// jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
// jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
// _ = jwtManager
```

**Rationale**: Clean code, no unused resources, clear intent

#### Option B: Document for Future Use

**If there's a chance the service will need authentication in the future:**

```go
// JWT authentication is currently handled by BFF layer.
// Uncomment below if this service needs direct frontend access:
// jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key")
// jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
```

**Rationale**: Preserves capability without wasting resources

#### Option C: Apply Authentication (Not Recommended for Internal Services)

**Only if you want to add defense-in-depth:**

```go
jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
authMiddleware := middleware.AuthMiddleware(jwtManager)

// Apply to all routes
api := router.Group("/api/v1")
api.Use(authMiddleware) // Now required!
```

**Rationale**: Defense-in-depth, but adds latency and complexity

---

## Production Security Recommendations

### For VPC/Private Network Deployment:
1. ✅ Use network policies to restrict access to internal services
2. ✅ Only expose BFF services to public internet
3. ✅ Use service mesh (Istio) for mTLS between services
4. ✅ Monitor service-to-service traffic

### For Public Cloud Deployment:
1. ✅ Deploy internal services in private subnets
2. ✅ Use API Gateway for BFF services
3. ✅ Implement IP whitelisting on internal services
4. ✅ Add mTLS for service-to-service communication

### For Kubernetes Deployment:
1. ✅ Use NetworkPolicies to restrict pod-to-pod communication
2. ✅ Use ServiceAccounts for service identity
3. ✅ Deploy service mesh (Istio/Linkerd) for automatic mTLS
4. ✅ Use RBAC for cluster resource access

---

## Authentication Decision Tree

```
Is this service called directly by frontend?
├─ YES → Implement full JWT authentication + RBAC
│         Examples: admin-bff-service, merchant-bff-service
│
└─ NO → Is it in a trusted network?
    ├─ YES → No HTTP-level auth needed (trust network policies)
    │         Examples: order-service, channel-adapter
    │
    └─ NO → Add service-to-service authentication
              Methods: mTLS, API keys, JWT with service accounts
```

---

## Migration Path (If Needed)

If your deployment requires authentication on internal services:

### Step 1: Enable JWT Validation
```go
jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
authMiddleware := middleware.AuthMiddleware(jwtManager)
api.Use(authMiddleware)
```

### Step 2: Update Calling Services
All services calling this internal service must now pass JWT:
```go
req.Header.Set("Authorization", "Bearer "+token)
```

### Step 3: Update BFF Services
BFF services must propagate their JWT to internal services:
```go
// In BFF service
internalServiceToken := c.GetHeader("Authorization")
internalReq.Header.Set("Authorization", internalServiceToken)
```

---

## Service-to-Service Authentication (Advanced)

For production environments requiring service-to-service auth:

### Option 1: mTLS via Service Mesh

**Best for**: Kubernetes deployments

```yaml
# Istio example
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
spec:
  mtls:
    mode: STRICT
```

### Option 2: API Keys

**Best for**: Simple deployments

```go
// In internal service
apiKey := c.GetHeader("X-Internal-API-Key")
if apiKey != config.GetEnv("INTERNAL_API_KEY", "") {
    c.JSON(403, gin.H{"error": "Invalid API key"})
    return
}
```

### Option 3: Service Account JWTs

**Best for**: Complex microservices

```go
// Each service gets its own JWT signed with service account
serviceToken := auth.GenerateServiceToken(serviceName)
req.Header.Set("X-Service-Token", serviceToken)
```

---

## Summary

**Current Architecture** (Recommended for most deployments):
- ✅ BFF services handle authentication
- ✅ Internal services trust the network
- ✅ Network policies provide isolation

**Enhanced Architecture** (For high-security environments):
- ✅ BFF services handle authentication
- ✅ mTLS for all service-to-service communication
- ✅ Network policies + service mesh
- ✅ Zero-trust networking

**Not Recommended**:
- ❌ Unused JWT managers sitting in code
- ❌ Inconsistent authentication patterns
- ❌ Authentication on some internal services but not others

---

## Next Actions

### Immediate (This Week):
1. ✅ Document authentication architecture (this file)
2. ⏳ Decide on Option A or B for unused JWT managers
3. ⏳ Update CLAUDE.md with authentication decision tree

### Short-term (This Month):
4. Add service-to-service authentication if deploying to untrusted network
5. Implement mTLS via service mesh for production
6. Add network policies for Kubernetes deployment

### Long-term (This Quarter):
7. Regular security audits
8. Penetration testing
9. Zero-trust architecture implementation

---

**Document Version**: 1.0
**Last Updated**: 2025-10-26
**Owner**: Backend Team
**Status**: Active


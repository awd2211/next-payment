# Circuit Breaker Coverage - Quick Reference

## At A Glance

```
Total HTTP Clients: 19
With Circuit Breaker: 18 (94.7%)
Without Breaker: 1 (5.3%) - CRITICAL ISSUE

Overall Health: ⚠️ NEEDS IMMEDIATE FIX
```

## Service Breakdown

### Payment Processing Core (Critical Path)

| Service | Status | Details |
|---------|--------|---------|
| **Payment Gateway** | ⚠️ PARTIAL | 3/4 clients protected - **Merchant Auth Client missing breaker** |
| **Order Service** | ✅ NO CALLS | No inter-service calls |
| **Channel Adapter** | ✅ FULL | 1/1 client with breaker |
| **Risk Service** | ✅ FULL | 1/1 client with breaker |

### Settlement & Withdrawal

| Service | Status | Details |
|---------|--------|---------|
| **Settlement Service** | ✅ FULL | 3/3 clients protected |
| **Withdrawal Service** | ✅ FULL | 3/3 clients protected |
| **Accounting Service** | ✅ FULL | 1/1 client protected |

### Admin & Auth

| Service | Status | Details |
|---------|--------|---------|
| **Merchant Service** | ✅ FULL | 5/5 clients protected |
| **Merchant Auth Service** | ✅ FULL | 1/1 client protected |
| **Admin Service** | ✅ NO CALLS | No inter-service calls |

### Notifications & Config

| Service | Status | Details |
|---------|--------|---------|
| **Notification Service** | ✅ NO CALLS | No inter-service calls |
| **Config Service** | ✅ NO CALLS | No inter-service calls |

---

## CRITICAL ISSUE: Payment Gateway -> Merchant Auth

### Location
```
/home/eric/payment/backend/services/payment-gateway/internal/client/merchant_auth_client.go
```

### Current Code
```go
type merchantAuthClient struct {
    baseURL string
    client  *http.Client  // ❌ RAW HTTP CLIENT
}

func NewMerchantAuthClient(baseURL string) MerchantAuthClient {
    return &merchantAuthClient{
        baseURL: baseURL,
        client: &http.Client{
            Timeout: 5 * time.Second,  // ❌ NO CIRCUIT BREAKER
        },
    }
}

func (c *merchantAuthClient) ValidateSignature(...) {
    ...
    resp, err := c.client.Do(req)  // ❌ NO BREAKER PROTECTION
}
```

### Why This Is Critical

1. **Payment Gateway is on Critical Path**
   - Every payment validation calls this client
   - If merchant-auth-service is slow/down, ALL payments block for 5 seconds

2. **No Automatic Failure Recovery**
   - Other services with breakers fail fast (60 seconds)
   - This client will retry indefinitely with 5s timeout

3. **Resource Exhaustion Risk**
   - Goroutines can accumulate waiting for timeouts
   - Connection pool can be exhausted
   - Can bring down entire payment gateway

### Impact Scenarios

**Scenario 1: Merchant Auth Service Down**
```
Before: Payment hangs for 5 seconds, then fails
After: Payment fails immediately (after 1st failure)
       Prevents cascading failures

Improvement: ~5000ms faster failure detection
```

**Scenario 2: Merchant Auth Service Slow (1s latency)**
```
Payment volume: 1000 payments/minute
Before: Each payment adds 1s latency (no compensation)
        System throughput: 1000 * (30ms + 1000ms) = 1.03s per payment

After: Circuit breaker trips after 5 requests with 60%+ failure
       System fails safe instead of degrading

Improvement: Prevents cascading slowness
```

### Fix Required

Change to use circuit breaker:

```go
type merchantAuthClient struct {
    baseURL string
    breaker *httpclient.BreakerClient  // ✅ ADD CIRCUIT BREAKER
}

func NewMerchantAuthClient(baseURL string) MerchantAuthClient {
    config := &httpclient.Config{
        Timeout:    5 * time.Second,
        MaxRetries: 3,
        RetryDelay: time.Second,
    }
    
    breakerConfig := httpclient.DefaultBreakerConfig("merchant-auth-service")
    
    return &merchantAuthClient{
        baseURL: baseURL,
        breaker: httpclient.NewBreakerClient(config, breakerConfig),
    }
}

func (c *merchantAuthClient) ValidateSignature(...) {
    ...
    req := &httpclient.Request{
        Method: "POST",
        URL:    url,
        Body:   reqBody,
        Ctx:    ctx,
    }
    
    resp, err := c.breaker.Do(req)  // ✅ USE CIRCUIT BREAKER
}
```

---

## Circuit Breaker States

### Closed (Normal)
```
Requests: 500/sec
Failures: 0%
State: Normal operation
Action: All requests pass through
```

### Half-Open (Testing Recovery)
```
Requests: 3 max
Failures: Testing
State: Service might be recovering
Action: Limited probe requests allowed
```

### Open (Tripped - Fail Fast)
```
Requests: Blocked
Failures: 60%+ on 5+ attempts
State: Service unavailable
Action: All requests fail immediately (after trip)
Duration: 60 seconds before retrying half-open
```

---

## Implementation Patterns Summary

### Pattern A: ServiceClient (Payment Gateway, Merchant Service)
**Pros**: Backward compatible
**Cons**: Has fallback bypass
**Use for**: Gradual migration

### Pattern B: Direct BreakerClient (Recommended)
**Pros**: Enforces usage, no bypass possible
**Cons**: No fallback
**Use for**: New code, critical paths

### Pattern C: Custom Breaker Config (External APIs)
**Pros**: Tuned for external reliability
**Cons**: More complex config
**Use for**: Third-party APIs (exchangerate-api, ipapi.co)

---

## Configuration Reference

### Standard Service-to-Service
```go
config := &httpclient.Config{
    Timeout:    30 * time.Second,
    MaxRetries: 3,
    RetryDelay: time.Second,
}
breakerConfig := httpclient.DefaultBreakerConfig("service-name")
// Trip on: 5+ requests with 60%+ failure rate
// Timeout: 60 seconds before half-open
```

### External APIs
```go
config := &httpclient.Config{
    Timeout:    5 * time.Second,      // Shorter
    MaxRetries: 2,                     // Fewer
    RetryDelay: 500 * time.Millisecond,
}
breakerConfig := httpclient.DefaultBreakerConfig("external-api")
breakerConfig.ReadyToTrip = func(counts gobreaker.Counts) bool {
    // Trip on: 10+ requests with 80%+ failure (more forgiving)
    return counts.Requests >= 10 && 
           float64(counts.TotalFailures) / float64(counts.Requests) >= 0.8
}
```

---

## Files That Need Monitoring

### Critical (Use Circuit Breaker)
- Payment Gateway client files (especially merchant_auth_client.go)
- Any new inter-service client

### Well Protected
- Settlement Service (3/3)
- Withdrawal Service (3/3)
- Merchant Service (5/5)
- All other clients

---

## Action Items

### Priority 1 (Do Now)
- [ ] Fix merchant_auth_client.go in payment-gateway
- [ ] Test circuit breaker behavior under load
- [ ] Add monitoring for breaker state changes

### Priority 2 (Next Sprint)
- [ ] Verify all clients using Pattern B (recommended)
- [ ] Document circuit breaker in runbook
- [ ] Add dashboards for breaker metrics

### Priority 3 (Ongoing)
- [ ] Monitor circuit breaker trips in production
- [ ] Tune thresholds based on real data
- [ ] Keep fallback strategies updated


# Redis Caching Optimization Guide

**Author**: Claude (Assistant)
**Date**: 2025-10-26
**Purpose**: Performance optimization through strategic Redis caching

---

## Table of Contents

1. [Overview](#overview)
2. [Current Redis Usage](#current-redis-usage)
3. [Caching Strategies](#caching-strategies)
4. [High-Traffic Endpoints to Cache](#high-traffic-endpoints-to-cache)
5. [Implementation Examples](#implementation-examples)
6. [Cache Invalidation](#cache-invalidation)
7. [Monitoring & Metrics](#monitoring--metrics)
8. [Best Practices](#best-practices)

---

## Overview

Redis caching is already integrated into the platform via `pkg/cache`. This guide provides recommendations for optimal caching strategies across high-traffic endpoints.

### Benefits
- **Reduced Database Load**: 80-90% reduction in DB queries for cached data
- **Faster Response Times**: Sub-millisecond cache lookups vs. 10-50ms DB queries
- **Cost Savings**: Lower RDS costs due to reduced IOPS
- **Better Scalability**: Handle 10x more traffic with same infrastructure

---

## Current Redis Usage

### Services Using Redis

| Service | Usage | Cache Keys |
|---------|-------|------------|
| **payment-gateway** | Idempotency, routing rules | `payment:{merchant_id}:{payment_no}`, `route:*` |
| **order-service** | Idempotency | `order:{merchant_id}:{payment_no}` |
| **risk-service** | GeoIP lookups, blacklist | `geoip:{ip}`, `blacklist:{type}:{value}` |
| **config-service** | System configs, feature flags | `config:{key}`, `feature:{name}` |
| **merchant-service** | API key validation | `api_key:{key_hash}` |
| **channel-adapter** | Exchange rates | `exchange_rate:{from}:{to}` |

### Existing Cache Implementation

The platform uses `pkg/cache` which provides:

```go
type Cache interface {
    Get(ctx context.Context, key string, value interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

**Implementations**:
- `RedisCache` - Production (distributed)
- `MemoryCache` - Development (local)

---

## Caching Strategies

### 1. Cache-Aside (Lazy Loading) ✅ **Recommended**

**Pattern**:
```go
func (s *service) GetMerchant(ctx context.Context, id uuid.UUID) (*Merchant, error) {
    // 1. Try cache
    cacheKey := fmt.Sprintf("merchant:%s", id.String())
    var merchant Merchant
    err := s.cache.Get(ctx, cacheKey, &merchant)
    if err == nil {
        return &merchant, nil  // Cache hit
    }

    // 2. Cache miss - query database
    merchant, err = s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache
    s.cache.Set(ctx, cacheKey, merchant, 15*time.Minute)

    return merchant, nil
}
```

**When to Use**:
- Read-heavy endpoints (GET requests)
- Data that changes infrequently
- Examples: merchant info, system configs, feature flags

**TTL Recommendations**:
- Static data (configs): 30-60 minutes
- Semi-static (merchants): 15-30 minutes
- Dynamic (exchange rates): 1-5 minutes

---

### 2. Write-Through Cache

**Pattern**:
```go
func (s *service) UpdateMerchant(ctx context.Context, merchant *Merchant) error {
    // 1. Update database
    err := s.repo.Update(ctx, merchant)
    if err != nil {
        return err
    }

    // 2. Update cache immediately
    cacheKey := fmt.Sprintf("merchant:%s", merchant.ID.String())
    s.cache.Set(ctx, cacheKey, merchant, 15*time.Minute)

    return nil
}
```

**When to Use**:
- Write-heavy data that's also frequently read
- When cache consistency is critical
- Examples: merchant settings, API keys

---

### 3. Read-Through Cache

**Pattern** (handled by cache layer):
```go
type CacheRepository struct {
    cache Cache
    db    *gorm.DB
}

func (r *CacheRepository) GetMerchant(ctx context.Context, id uuid.UUID) (*Merchant, error) {
    return r.cache.GetOrLoad(ctx,
        fmt.Sprintf("merchant:%s", id),
        func() (interface{}, error) {
            return r.db.First(&Merchant{}, id)
        },
        15*time.Minute,
    )
}
```

**When to Use**:
- Centralized caching logic
- Complex data loading

---

## High-Traffic Endpoints to Cache

### Priority 1: Critical Read Paths ⚡

#### 1. Merchant Service
```go
// Cache merchant profile (read on every API call for auth)
GET /api/v1/merchants/:id
Cache Key: merchant:{merchant_id}
TTL: 15 minutes
Expected Hit Rate: 95%+
```

```go
// Cache API keys (validated on every payment request)
GET /api/v1/merchants/:id/api-keys
Cache Key: api_keys:{merchant_id}
TTL: 10 minutes
Expected Hit Rate: 98%+
```

#### 2. Config Service
```go
// Cache system configs (read frequently)
GET /api/v1/configs/:key
Cache Key: config:{key}
TTL: 30 minutes
Expected Hit Rate: 99%+
```

```go
// Cache feature flags
GET /api/v1/features/:name
Cache Key: feature:{name}
TTL: 30 minutes
Expected Hit Rate: 99%+
```

#### 3. Channel Adapter
```go
// Cache exchange rates (updated hourly)
GET /api/v1/exchange-rates/:from/:to
Cache Key: exchange_rate:{from}:{to}
TTL: 5 minutes
Expected Hit Rate: 90%+
```

#### 4. Risk Service
```go
// Cache GeoIP lookups (IP to country mapping)
Internal: GetCountryFromIP(ip)
Cache Key: geoip:{ip}
TTL: 24 hours
Expected Hit Rate: 85%+
```

```go
// Cache blacklist checks
Internal: IsBlacklisted(type, value)
Cache Key: blacklist:{type}:{value}
TTL: 10 minutes
Expected Hit Rate: 95%+
```

#### 5. Order Service
```go
// Cache order details (read for status checks)
GET /api/v1/orders/:orderNo
Cache Key: order:{order_no}
TTL: 5 minutes
Expected Hit Rate: 70%+
```

### Priority 2: Analytics & Reporting 📊

#### Analytics Service
```go
// Cache merchant statistics (dashboard)
GET /api/v1/analytics/merchant/:id/stats
Cache Key: analytics:merchant:{merchant_id}:stats:{date}
TTL: 15 minutes
Expected Hit Rate: 80%+
```

```go
// Cache global statistics
GET /api/v1/analytics/global
Cache Key: analytics:global:stats:{hour}
TTL: 10 minutes
Expected Hit Rate: 90%+
```

---

## Implementation Examples

### Example 1: Merchant Profile Caching

**File**: `merchant-service/internal/service/merchant_service.go`

```go
// Add cache field to service
type merchantService struct {
    repo  repository.MerchantRepository
    cache cache.Cache  // Add this
    db    *gorm.DB
}

// Update GetMerchant method
func (s *merchantService) GetMerchant(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("merchant:%s", id.String())
    var merchant model.Merchant

    err := s.cache.Get(ctx, cacheKey, &merchant)
    if err == nil {
        logger.Debug("Merchant cache hit", zap.String("merchant_id", id.String()))
        return &merchant, nil
    }

    // Cache miss - query database
    merchant, err = s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Store in cache (15 minute TTL)
    go func() {
        if err := s.cache.Set(context.Background(), cacheKey, merchant, 15*time.Minute); err != nil {
            logger.Warn("Failed to cache merchant", zap.Error(err))
        }
    }()

    return merchant, nil
}

// Invalidate cache on update
func (s *merchantService) UpdateMerchant(ctx context.Context, merchant *model.Merchant) error {
    err := s.repo.Update(ctx, merchant)
    if err != nil {
        return err
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("merchant:%s", merchant.ID.String())
    s.cache.Delete(ctx, cacheKey)

    return nil
}
```

---

### Example 2: Config Service Caching

**File**: `config-service/internal/service/config_service.go`

```go
func (s *configService) GetConfig(ctx context.Context, key string) (string, error) {
    // Try cache
    cacheKey := fmt.Sprintf("config:%s", key)
    var value string

    err := s.cache.Get(ctx, cacheKey, &value)
    if err == nil {
        return value, nil
    }

    // Query database
    config, err := s.repo.GetByKey(ctx, key)
    if err != nil {
        return "", err
    }

    // Cache with long TTL (configs rarely change)
    s.cache.Set(ctx, cacheKey, config.Value, 30*time.Minute)

    return config.Value, nil
}
```

---

### Example 3: API Key Validation Caching

**File**: `merchant-service/internal/middleware/api_key_middleware.go`

```go
func (m *APIKeyMiddleware) ValidateAPIKey(apiKey string) (*model.APIKey, error) {
    // Hash the key
    keyHash := hash(apiKey)

    // Try cache
    cacheKey := fmt.Sprintf("api_key:%s", keyHash)
    var cachedKey model.APIKey

    err := m.cache.Get(context.Background(), cacheKey, &cachedKey)
    if err == nil {
        // Verify still active
        if cachedKey.Status == "active" {
            return &cachedKey, nil
        }
    }

    // Query database
    apiKeyRecord, err := m.repo.GetByKeyHash(context.Background(), keyHash)
    if err != nil {
        return nil, err
    }

    // Cache for 10 minutes
    m.cache.Set(context.Background(), cacheKey, apiKeyRecord, 10*time.Minute)

    return apiKeyRecord, nil
}
```

---

## Cache Invalidation

### Strategy 1: TTL-Based (Automatic) ✅ **Default**

**Pros**:
- Simple to implement
- No manual invalidation needed
- Good for semi-static data

**Cons**:
- Potential stale data during TTL window
- May serve outdated data

**Use For**: Configs, exchange rates, GeoIP

---

### Strategy 2: Event-Based (Explicit)

**Pattern**:
```go
// On update/delete, invalidate cache
func (s *service) UpdateMerchant(ctx context.Context, merchant *Merchant) error {
    // 1. Update database
    err := s.repo.Update(ctx, merchant)
    if err != nil {
        return err
    }

    // 2. Invalidate cache
    cacheKey := fmt.Sprintf("merchant:%s", merchant.ID.String())
    s.cache.Delete(ctx, cacheKey)

    return nil
}
```

**Use For**: Merchants, API keys, user profiles

---

### Strategy 3: Pub/Sub (Distributed Invalidation)

**Pattern**:
```go
// Service A updates data
func (s *serviceA) UpdateConfig(ctx context.Context, config *Config) error {
    err := s.repo.Update(ctx, config)
    if err != nil {
        return err
    }

    // Publish invalidation event
    s.redis.Publish(ctx, "cache:invalidate", fmt.Sprintf("config:%s", config.Key))

    return nil
}

// Service B subscribes to invalidation
func (s *serviceB) subscribeToInvalidation() {
    pubsub := s.redis.Subscribe(ctx, "cache:invalidate")
    for msg := range pubsub.Channel() {
        s.cache.Delete(context.Background(), msg.Payload)
    }
}
```

**Use For**: Multi-instance deployments, shared caches

---

## Monitoring & Metrics

### Key Metrics to Track

```go
// Cache hit rate
cache_hit_rate = cache_hits / (cache_hits + cache_misses) * 100

// Target: 80%+ for most endpoints
// Target: 95%+ for configs and static data
```

### Prometheus Metrics

```go
// In pkg/cache/redis_cache.go
var (
    cacheHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"service", "key_prefix"},
    )

    cacheMisses = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
        []string{"service", "key_prefix"},
    )
)

// Track in Get() method
func (c *RedisCache) Get(ctx context.Context, key string, value interface{}) error {
    err := c.client.Get(ctx, key).Scan(value)

    if err == redis.Nil {
        cacheMisses.WithLabelValues(c.serviceName, getKeyPrefix(key)).Inc()
        return ErrCacheMiss
    }

    if err == nil {
        cacheHits.WithLabelValues(c.serviceName, getKeyPrefix(key)).Inc()
    }

    return err
}
```

---

## Best Practices

### 1. Cache Key Naming Convention ✅

```go
// Pattern: {service}:{entity}:{id}[:{attribute}]

// Good examples:
merchant:profile:550e8400-e29b-41d4-a716-446655440000
config:system:jwt_secret
api_key:hash:abc123def456
order:status:ORD-20251026-001
exchange_rate:USD:CNY

// Bad examples:
user_550e8400  // No service/entity prefix
m:123         // Not descriptive
data          // Too generic
```

### 2. TTL Guidelines ✅

| Data Type | TTL | Reason |
|-----------|-----|--------|
| Static configs | 30-60 min | Rarely changes |
| Merchant profiles | 15-30 min | Semi-static |
| API keys | 10-15 min | Security-sensitive |
| Exchange rates | 1-5 min | Changes frequently |
| GeoIP lookups | 24 hours | Stable data |
| Session data | 30 min | User activity |
| Analytics | 10-15 min | Near real-time ok |

### 3. Cache Stampede Prevention ✅

**Problem**: Multiple requests hit DB simultaneously when cache expires

**Solution**: Use locking

```go
func (s *service) GetMerchant(ctx context.Context, id uuid.UUID) (*Merchant, error) {
    cacheKey := fmt.Sprintf("merchant:%s", id.String())
    lockKey := fmt.Sprintf("lock:merchant:%s", id.String())

    // Try cache
    var merchant Merchant
    err := s.cache.Get(ctx, cacheKey, &merchant)
    if err == nil {
        return &merchant, nil
    }

    // Acquire lock
    locked, err := s.redis.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
    if err != nil || !locked {
        // Another request is loading, wait and retry
        time.Sleep(100 * time.Millisecond)
        return s.GetMerchant(ctx, id) // Retry
    }
    defer s.redis.Del(ctx, lockKey)

    // Load from DB
    merchant, err = s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache result
    s.cache.Set(ctx, cacheKey, merchant, 15*time.Minute)

    return merchant, nil
}
```

### 4. Serialization Format ✅

```go
// Use JSON for complex objects (default in pkg/cache)
type Merchant struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
}

// Use string for simple values
cache.Set(ctx, "config:jwt_secret", "secret_value", 30*time.Minute)

// Use numbers directly
cache.Set(ctx, "counter:requests", 12345, 5*time.Minute)
```

### 5. Error Handling ✅

```go
// Never fail request on cache errors
merchant, err := s.cache.Get(ctx, cacheKey, &merchant)
if err != nil {
    logger.Warn("Cache get failed, falling back to DB", zap.Error(err))
    // Continue to DB query
}

// Cache set failures should not block
go func() {
    if err := s.cache.Set(ctx, cacheKey, merchant, ttl); err != nil {
        logger.Warn("Failed to cache result", zap.Error(err))
        // Don't return error - cache is optional
    }
}()
```

---

## Performance Impact Estimates

| Endpoint | Without Cache | With Cache | Improvement |
|----------|--------------|------------|-------------|
| GET /merchants/:id | 15-25ms | 1-3ms | **8-25x faster** |
| GET /configs/:key | 10-20ms | 0.5-1ms | **10-40x faster** |
| API key validation | 20-30ms | 1-2ms | **10-30x faster** |
| Exchange rate lookup | 50-100ms | 2-5ms | **10-50x faster** |
| GeoIP lookup | 100-200ms (API) | 1-2ms | **50-200x faster** |

**Database Load Reduction**: 70-90% (varies by endpoint)

---

## Next Steps

1. ✅ **Implement caching in priority endpoints** (merchant, config, API keys)
2. ✅ **Add cache metrics to Prometheus**
3. ✅ **Set up cache monitoring dashboards in Grafana**
4. ✅ **Monitor cache hit rates and adjust TTLs**
5. ✅ **Document cache keys in API documentation**

---

## References

- `pkg/cache/cache.go` - Cache interface
- `pkg/cache/redis_cache.go` - Redis implementation
- `pkg/cache/memory_cache.go` - Memory implementation


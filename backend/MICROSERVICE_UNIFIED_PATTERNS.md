# Microservice Unified Patterns Guide

## Overview

This document codifies the **unified patterns** that ALL microservices in the payment platform must follow to ensure consistency, maintainability, and operational excellence.

**Last Updated**: 2025-01-20
**Status**: ✅ 100% Compliance (19/19 services)

---

## Table of Contents

1. [Directory Structure](#directory-structure)
2. [Bootstrap Initialization](#bootstrap-initialization)
3. [4-Layer Architecture](#4-layer-architecture)
4. [Hot Reload Configuration](#hot-reload-configuration)
5. [Observability Requirements](#observability-requirements)
6. [API Design Standards](#api-design-standards)
7. [Error Handling Patterns](#error-handling-patterns)
8. [Database Patterns](#database-patterns)
9. [Testing Standards](#testing-standards)
10. [Compliance Checklist](#compliance-checklist)

---

## Directory Structure

Every microservice MUST follow this exact structure:

```
service-name/
├── .air.toml                    # Hot reload configuration (REQUIRED)
├── go.mod                       # Service module definition
├── cmd/
│   └── main.go                 # Entry point (Bootstrap initialization)
├── internal/
│   ├── model/                  # Data models (GORM structs)
│   ├── repository/             # Data access layer (interface + impl)
│   ├── service/                # Business logic layer
│   ├── handler/                # HTTP handlers (Gin)
│   └── client/                 # External service clients (optional)
└── tmp/                        # Air build artifacts (gitignored)
```

**Verification**:
```bash
# Check compliance
ls -la service-name/
# Should show: .air.toml, cmd/, internal/model, internal/repository, internal/service, internal/handler
```

---

## Bootstrap Initialization

### Pattern

ALL services MUST use the `pkg/app.Bootstrap` framework in `cmd/main.go`:

```go
package main

import (
    "log"
    "time"

    "github.com/payment-platform/pkg/app"
    "github.com/payment-platform/pkg/config"

    "payment-platform/service-name/internal/handler"
    "payment-platform/service-name/internal/model"
    "payment-platform/service-name/internal/repository"
    "payment-platform/service-name/internal/service"
)

func main() {
    // 1. Bootstrap framework initialization
    application, err := app.Bootstrap(app.ServiceConfig{
        ServiceName: "service-name",
        DBName:      "payment_service_name",
        Port:        config.GetEnvInt("PORT", 40XXX),
        AutoMigrate: []any{
            &model.Entity1{},
            &model.Entity2{},
        },

        // 2. Standard feature flags (DO NOT CHANGE unless you have a reason)
        EnableTracing:     true,   // Jaeger tracing
        EnableMetrics:     true,   // Prometheus metrics
        EnableRedis:       true,   // Redis connection
        EnableGRPC:        false,  // HTTP-only (default)
        EnableHealthCheck: true,   // /health endpoint
        EnableRateLimit:   true,   // Rate limiting

        RateLimitRequests: 100,
        RateLimitWindow:   time.Minute,
    })
    if err != nil {
        log.Fatalf("Failed to bootstrap service: %v", err)
    }

    // 3. Initialize repository
    repo := repository.NewRepository(application.DB)

    // 4. Initialize external clients (if needed)
    // externalClient := client.NewExternalClient(...)

    // 5. Create service
    svc := service.NewService(repo, application.DB /* , externalClient */)

    // 6. Create handler
    h := handler.NewHandler(svc)

    // 7. Register routes
    api := application.Router.Group("/api/v1")
    h.RegisterRoutes(api)

    // 8. Start service with graceful shutdown
    application.RunWithGracefulShutdown()
}
```

### Benefits of Bootstrap

- ✅ **Auto-configures**: Database, Redis, Logger, Gin router, Middleware stack
- ✅ **Auto-enables**: Tracing, Metrics, Health checks, Rate limiting
- ✅ **HTTP-first**: Default protocol (gRPC optional)
- ✅ **Graceful shutdown**: Handles SIGINT/SIGTERM, closes all resources
- ✅ **Reduces boilerplate**: 26-80% less code vs manual initialization
- ✅ **Consistent configuration**: All services use same setup pattern

### What Bootstrap Provides

After `app.Bootstrap()` returns, you have access to:

```go
application.DB          // *gorm.DB - PostgreSQL connection
application.RedisClient // *redis.Client - Redis connection
application.Logger      // *zap.Logger - Structured logger
application.Router      // *gin.Engine - HTTP router with middleware
application.Config      // app.Config - Service configuration
```

**Middleware stack (automatically applied)**:
1. RequestID middleware (X-Request-ID header)
2. Logger middleware (structured logging)
3. CORS middleware
4. Metrics middleware (Prometheus)
5. Tracing middleware (Jaeger)
6. RateLimit middleware (if EnableRateLimit: true)

---

## 4-Layer Architecture

Every service follows this layered architecture:

```
┌─────────────────────────┐
│  Handler Layer (HTTP)   │  ← Gin handlers, route registration
│  internal/handler/      │
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│  Service Layer          │  ← Business logic, orchestration
│  internal/service/      │
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│  Repository Layer       │  ← Data access, GORM queries
│  internal/repository/   │
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│  Model Layer            │  ← GORM structs, database schema
│  internal/model/        │
└─────────────────────────┘
```

### Handler Layer

**Responsibility**: HTTP request/response handling, input validation, authentication

**Example**:
```go
// internal/handler/handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

type Handler struct {
    service service.Service
}

func NewHandler(service service.Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
    router.POST("/entities", h.CreateEntity)
    router.GET("/entities", h.ListEntities)
    router.GET("/entities/:id", h.GetEntity)
}

func (h *Handler) CreateEntity(c *gin.Context) {
    var req CreateEntityRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := h.service.CreateEntity(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}
```

### Service Layer

**Responsibility**: Business logic, transaction management, external service calls

**Example**:
```go
// internal/service/service.go
package service

type Service interface {
    CreateEntity(ctx context.Context, input *CreateEntityInput) (*Entity, error)
    GetEntity(ctx context.Context, id uuid.UUID) (*Entity, error)
}

type service struct {
    repo repository.Repository
    db   *gorm.DB
}

func NewService(repo repository.Repository, db *gorm.DB) Service {
    return &service{repo: repo, db: db}
}

func (s *service) CreateEntity(ctx context.Context, input *CreateEntityInput) (*Entity, error) {
    // 1. Validation
    if err := validateInput(input); err != nil {
        return nil, err
    }

    // 2. Business logic
    entity := &model.Entity{
        ID:   uuid.New(),
        Name: input.Name,
        // ...
    }

    // 3. Database transaction (if needed)
    err := s.db.Transaction(func(tx *gorm.DB) error {
        if err := s.repo.Create(ctx, entity); err != nil {
            return err
        }
        // ... other operations
        return nil
    })

    return entity, err
}
```

### Repository Layer

**Responsibility**: Database access, GORM queries, no business logic

**Example**:
```go
// internal/repository/repository.go
package repository

type Repository interface {
    Create(ctx context.Context, entity *model.Entity) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.Entity, error)
    List(ctx context.Context, filters Filters, page, pageSize int) ([]*model.Entity, int64, error)
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, entity *model.Entity) error {
    return r.db.WithContext(ctx).Create(entity).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Entity, error) {
    var entity model.Entity
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
        return nil, err
    }
    return &entity, nil
}
```

### Model Layer

**Responsibility**: Data structures, GORM tags, database schema

**Example**:
```go
// internal/model/entity.go
package model

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Entity struct {
    ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
    Name        string         `gorm:"type:varchar(255);not null" json:"name"`
    Description string         `gorm:"type:text" json:"description"`
    Status      string         `gorm:"type:varchar(50);not null;index" json:"status"`
    CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Entity) TableName() string {
    return "entities"
}

// Lifecycle hooks (if needed)
func (e *Entity) BeforeCreate(tx *gorm.DB) error {
    if e.ID == uuid.Nil {
        e.ID = uuid.New()
    }
    return nil
}
```

---

## Hot Reload Configuration

### .air.toml (REQUIRED)

ALL services MUST have `.air.toml` in the root directory:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "GOWORK=/home/eric/payment/backend/go.work go build -o ./tmp/main ./cmd/main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

**Why Air?**
- ✅ Instant hot reload during development
- ✅ Automatic recompilation on file changes
- ✅ Colored output for better debugging
- ✅ Works with Go Workspace (GOWORK)
- ✅ Used by all 19 services

**Start with Air**:
```bash
cd services/service-name
air -c .air.toml
```

---

## Observability Requirements

### 1. Tracing (Jaeger)

ALL services MUST enable tracing via Bootstrap:

```go
EnableTracing: true,
```

**What you get automatically**:
- HTTP request spans (via middleware)
- W3C Trace Context propagation
- X-Trace-ID response header
- Jaeger UI integration (http://localhost:50686)

**Manual spans for business logic**:
```go
import "github.com/payment-platform/pkg/tracing"

ctx, span := tracing.StartSpan(ctx, "service-name", "OperationName")
defer span.End()

tracing.AddSpanTags(ctx, map[string]interface{}{
    "entity_id": entityID.String(),
    "amount": amount,
})

if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

### 2. Metrics (Prometheus)

ALL services MUST enable metrics via Bootstrap:

```go
EnableMetrics: true,
```

**What you get automatically**:
- HTTP request metrics (rate, duration, status)
- `/metrics` endpoint for Prometheus scraping

**Exposed metrics**:
```promql
http_requests_total{service="service-name",method="POST",path="/api/v1/entities",status="200"}
http_request_duration_seconds{method="POST",path="/api/v1/entities",status="200"}
http_request_size_bytes{method="POST",path="/api/v1/entities"}
http_response_size_bytes{method="POST",path="/api/v1/entities"}
```

**Custom business metrics** (if needed):
```go
import "github.com/payment-platform/pkg/metrics"

// In service layer
entityMetrics := metrics.NewCustomMetrics("entity_operations")
entityMetrics.RecordOperation("create", "success", duration)
```

### 3. Logging (Zap)

Use structured logging from Bootstrap:

```go
application.Logger.Info("Entity created",
    zap.String("entity_id", entityID.String()),
    zap.String("name", name),
)

application.Logger.Error("Failed to create entity",
    zap.Error(err),
    zap.String("entity_id", entityID.String()),
)
```

### 4. Health Checks

ALL services MUST enable health checks:

```go
EnableHealthCheck: true,
```

**Endpoints**:
- `GET /health` - Basic health check
- Returns: `{"status": "ok", "service": "service-name", "dependencies": {...}}`

---

## API Design Standards

### 1. Standard Response Format

ALL API responses MUST follow this format:

**Success**:
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**Error**:
```json
{
  "code": 1001,
  "message": "Entity not found",
  "error": "record not found"
}
```

**List with pagination**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 2. Route Naming Convention

```
POST   /api/v1/entities           # Create
GET    /api/v1/entities           # List (with pagination)
GET    /api/v1/entities/:id       # Get by ID
PUT    /api/v1/entities/:id       # Update
DELETE /api/v1/entities/:id       # Delete (soft delete preferred)
```

### 3. Common Query Parameters

```
?page=1                  # Page number (default: 1)
?page_size=10           # Items per page (default: 10, max: 100)
?status=active          # Filter by status
?start_date=2024-01-01  # Date range filter
?end_date=2024-12-31
?sort_by=created_at     # Sorting field
?sort_order=desc        # asc or desc
```

---

## Error Handling Patterns

### 1. Repository Layer

```go
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Entity, error) {
    var entity model.Entity
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrEntityNotFound
        }
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }
    return &entity, nil
}
```

### 2. Service Layer

```go
var (
    ErrEntityNotFound = errors.New("entity not found")
    ErrInvalidInput   = errors.New("invalid input")
)

func (s *service) GetEntity(ctx context.Context, id uuid.UUID) (*Entity, error) {
    entity, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, ErrEntityNotFound) {
            return nil, ErrEntityNotFound
        }
        return nil, fmt.Errorf("service: failed to get entity: %w", err)
    }
    return entity, nil
}
```

### 3. Handler Layer

```go
func (h *Handler) GetEntity(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": 400,
            "error": "invalid entity ID format",
        })
        return
    }

    entity, err := h.service.GetEntity(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, service.ErrEntityNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "code": 404,
                "error": "entity not found",
            })
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "error": "internal server error",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "data": entity,
    })
}
```

---

## Database Patterns

### 1. UUID Primary Keys

ALL tables MUST use UUID primary keys:

```go
type Entity struct {
    ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    // ...
}

func (e *Entity) BeforeCreate(tx *gorm.DB) error {
    if e.ID == uuid.Nil {
        e.ID = uuid.New()
    }
    return nil
}
```

### 2. Soft Deletes

Prefer soft deletes over hard deletes:

```go
type Entity struct {
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

### 3. Timestamps

ALL tables MUST have created_at and updated_at:

```go
type Entity struct {
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

### 4. Money Handling

ALL monetary amounts MUST be stored as integers (cents):

```go
type Payment struct {
    Amount   int64  `gorm:"not null" json:"amount"`        // In cents ($10.00 = 1000)
    Currency string `gorm:"type:varchar(3);not null" json:"currency"` // USD, EUR, etc.
}
```

### 5. Transaction Management

Use service layer for transaction management:

```go
func (s *service) CreateEntityWithRelations(ctx context.Context, input *Input) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Create entity
        if err := s.repo.Create(ctx, entity); err != nil {
            return err
        }

        // Create related records
        if err := s.repo.CreateRelations(ctx, relations); err != nil {
            return err
        }

        // All or nothing
        return nil
    })
}
```

---

## Testing Standards

### 1. Unit Tests

```go
// internal/service/service_test.go
package service_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCreateEntity_Success(t *testing.T) {
    // Setup
    mockRepo := new(mocks.MockRepository)
    service := NewService(mockRepo, nil)

    // Mock expectations
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Entity")).
        Return(nil)

    // Execute
    result, err := service.CreateEntity(context.Background(), &input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

### 2. Run Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific test
go test -run TestCreateEntity_Success ./internal/service
```

---

## Compliance Checklist

Use this checklist when creating a new microservice:

### Structure
- [ ] Has `.air.toml` configuration
- [ ] Has `cmd/main.go` with Bootstrap initialization
- [ ] Has `internal/model/` directory
- [ ] Has `internal/repository/` directory
- [ ] Has `internal/service/` directory
- [ ] Has `internal/handler/` directory
- [ ] Has `go.mod` with correct module name

### Bootstrap Configuration
- [ ] Uses `app.Bootstrap()` in main.go
- [ ] Sets correct ServiceName
- [ ] Sets correct DBName (payment_*)
- [ ] Sets correct Port (40XXX range)
- [ ] Lists all models in AutoMigrate
- [ ] EnableTracing: true
- [ ] EnableMetrics: true
- [ ] EnableRedis: true
- [ ] EnableGRPC: false (unless needed)
- [ ] EnableHealthCheck: true
- [ ] EnableRateLimit: true
- [ ] Calls `application.RunWithGracefulShutdown()`

### 4-Layer Architecture
- [ ] Handler layer only handles HTTP
- [ ] Service layer contains business logic
- [ ] Repository layer only accesses database
- [ ] Model layer only defines structs

### API Standards
- [ ] Uses standard response format (code, message, data)
- [ ] Follows RESTful route naming
- [ ] Implements pagination for list endpoints
- [ ] Returns proper HTTP status codes

### Database
- [ ] All tables use UUID primary keys
- [ ] All tables have created_at, updated_at
- [ ] Uses soft deletes (DeletedAt)
- [ ] Money stored as integers (cents)
- [ ] Transactions managed in service layer

### Observability
- [ ] `/metrics` endpoint accessible
- [ ] `/health` endpoint accessible
- [ ] Structured logging used
- [ ] Tracing enabled
- [ ] Service shows up in Jaeger UI

### Development
- [ ] Service starts with `air -c .air.toml`
- [ ] Hot reload works
- [ ] Compiles successfully
- [ ] No Go warnings or errors

---

## Quick Reference: Creating a New Service

```bash
# 1. Create directory structure
mkdir -p services/new-service/{cmd,internal/{model,repository,service,handler}}

# 2. Initialize Go module
cd services/new-service
go mod init payment-platform/new-service

# 3. Add to go.work
echo "use ./services/new-service" >> ../../go.work

# 4. Copy .air.toml from another service
cp ../notification-service/.air.toml .

# 5. Create cmd/main.go (use template above)

# 6. Create internal layers (model, repository, service, handler)

# 7. Test compilation
GOWORK=/home/eric/payment/backend/go.work go build ./cmd/main.go

# 8. Start with Air
air -c .air.toml
```

---

## Compliance Status

**Last Audit**: 2025-01-20

| Service | Bootstrap | 4-Layer | Air Config | Status |
|---------|-----------|---------|------------|--------|
| accounting-service | ✅ | ✅ | ✅ | ✅ Full |
| admin-service | ✅ | ✅ | ✅ | ✅ Full |
| analytics-service | ✅ | ✅ | ✅ | ✅ Full |
| cashier-service | ✅ | ✅ | ✅ | ✅ Full |
| channel-adapter | ✅ | ✅ | ✅ | ✅ Full |
| config-service | ✅ | ✅ | ✅ | ✅ Full |
| dispute-service | ✅ | ✅ | ✅ | ✅ Full |
| kyc-service | ✅ | ✅ | ✅ | ✅ Full |
| merchant-auth-service | ✅ | ✅ | ✅ | ✅ Full |
| merchant-config-service | ✅ | ✅ | ✅ | ✅ Full |
| merchant-limit-service | ✅ | ✅ | ✅ | ✅ Full |
| merchant-service | ✅ | ✅ | ✅ | ✅ Full |
| notification-service | ✅ | ✅ | ✅ | ✅ Full |
| order-service | ✅ | ✅ | ✅ | ✅ Full |
| payment-gateway | ✅ | ✅ | ✅ | ✅ Full |
| reconciliation-service | ✅ | ✅ | ✅ | ✅ Full |
| risk-service | ✅ | ✅ | ✅ | ✅ Full |
| settlement-service | ✅ | ✅ | ✅ | ✅ Full |
| withdrawal-service | ✅ | ✅ | ✅ | ✅ Full |

**Total**: 19/19 services (100% compliance)

---

## Related Documentation

- [BOOTSTRAP_MIGRATION_FINAL_100PERCENT.md](BOOTSTRAP_MIGRATION_FINAL_100PERCENT.md) - Bootstrap migration guide
- [SPRINT2_BACKEND_COMPLETE.md](SPRINT2_BACKEND_COMPLETE.md) - Sprint 2 implementation details
- [API_DOCUMENTATION_GUIDE.md](API_DOCUMENTATION_GUIDE.md) - API documentation standards

---

## Changelog

### 2025-01-20
- ✅ Initial version created
- ✅ All 19 services audited and compliant
- ✅ Added .air.toml to Sprint 2 services (reconciliation, dispute, merchant-limit)
- ✅ Updated management scripts to use Air

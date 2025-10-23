# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Global Payment Platform** - an enterprise-grade, multi-tenant payment gateway built with Go microservices architecture. The system supports multiple payment channels (Stripe, PayPal, cryptocurrency) and provides a complete payment processing solution.

**Core Architecture**: 10 independent microservices using Go 1.21+, PostgreSQL, Redis, and Kafka.

## Development Commands

### Building and Running Services

All services use Go Workspace (`go.work`) for dependency management. Work from the `backend/` directory:

```bash
# Build a specific service
cd backend/services/payment-gateway
go build -o /tmp/payment-gateway ./cmd/main.go

# Build all services (verify compilation)
cd backend
for service in services/*/; do
  cd "$service"
  go build -o /tmp/$(basename "$service") ./cmd/main.go 2>&1
  cd ../..
done

# Run with hot reload (requires Air)
./scripts/dev-with-air.sh

# Stop all services
./scripts/stop-services.sh
```

### Testing

```bash
# Run tests for a specific service
cd backend/services/payment-gateway
go test ./...

# Run tests for shared pkg
cd backend/pkg
go test ./...

# Clean build cache
go clean -cache
```

### Database Operations

```bash
# Initialize databases
./scripts/init-db.sh

# Run migrations
./scripts/migrate.sh
```

## Architecture

### Microservices Communication Pattern

The system uses **HTTP-based inter-service communication** (not gRPC, despite what the README says). Services communicate via HTTP clients defined in the `internal/client/` directory.

**Payment Gateway** is the orchestrator and calls:
- **Order Service** (port 8004) - Order creation and status updates
- **Channel Adapter** (port 8005) - Payment channel processing
- **Risk Service** (port 8006) - Risk assessment

```go
// Example from payment-gateway/cmd/main.go
orderClient := client.NewOrderClient(orderServiceURL)
channelClient := client.NewChannelClient(channelServiceURL)
riskClient := client.NewRiskClient(riskServiceURL)

paymentService := service.NewPaymentService(
    paymentRepo,
    orderClient,
    channelClient,
    riskClient,
    redisClient,
)
```

### Service Structure (Standard Pattern)

Each microservice follows this structure:
```
service-name/
├── cmd/
│   └── main.go           # Entry point with pkg imports
├── internal/
│   ├── model/            # Data models (GORM)
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic
│   ├── handler/          # HTTP handlers (Gin)
│   ├── client/           # HTTP clients for other services (if needed)
│   └── middleware/       # Service-specific middleware (if needed)
└── go.mod
```

**Key Pattern**: `main.go` imports shared functionality from `pkg/` and wires dependencies:
1. Initialize logger, database, Redis
2. Create repositories
3. Create service clients (if needed)
4. Create services with dependency injection
5. Create handlers
6. Register routes with middleware
7. Start HTTP server

### Shared Libraries (pkg/)

The `backend/pkg/` directory contains reusable components:

- **auth/** - JWT token generation/validation, Claims struct
- **cache/** - Cache interface with Redis and in-memory implementations
- **config/** - Environment variable loading (`GetEnv`, `GetEnvInt`)
- **db/** - PostgreSQL and Redis connection pooling
- **email/** - SMTP and Mailgun email sending
- **httpclient/** - HTTP client with retry logic
- **kafka/** - Kafka producer/consumer
- **logger/** - Zap-based structured logging
- **middleware/** - Gin middleware (CORS, Auth, RateLimit, RequestID, Logger)
- **retry/** - Exponential backoff retry mechanism
- **validator/** - Amount, currency, and string validation (including Luhn for credit cards)

**Important**: All services use these shared packages via the Go Workspace `replace` directive.

### Authentication and Authorization

**Two-tier authentication**:

1. **JWT Authentication** (Admin/Merchant users):
   ```go
   // In main.go
   jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
   authMiddleware := middleware.AuthMiddleware(jwtManager)

   // Apply to routes
   api.Use(authMiddleware)
   ```

2. **Signature Verification** (API clients - used by Payment Gateway):
   ```go
   // Payment Gateway has custom signature middleware
   signatureMiddleware := localMiddleware.NewSignatureMiddleware(secretFetcher)
   api.Use(signatureMiddleware.Verify())
   ```

### Service Ports and Databases

| Service | Port | Database |
|---------|------|----------|
| admin-service | 8001 | payment_admin |
| merchant-service | 8002 | payment_merchant |
| payment-gateway | 8003 | payment_gateway |
| order-service | 8004 | payment_order |
| channel-adapter | 8005 | payment_channel |
| risk-service | 8006 | payment_risk |
| notification-service | 8007 | payment_notify |
| accounting-service | 8008 | payment_accounting |
| analytics-service | 8009 | payment_analytics |
| config-service | 8010 | payment_config |

Each service has its own isolated PostgreSQL database for multi-tenancy.

### Payment Flow Architecture

Complete payment processing flow:

1. **Merchant** calls Payment Gateway API with signature
2. **Payment Gateway** (`CreatePayment`):
   - Validates request and checks idempotency (Redis)
   - Calls **Risk Service** for assessment
   - Generates unique `payment_no`
   - Calls **Order Service** to create order
   - Selects payment channel (via routing rules or manual selection)
   - Calls **Channel Adapter** with payment details
3. **Channel Adapter**:
   - Routes to appropriate adapter (Stripe, PayPal, etc.)
   - Calls external payment provider API
   - Returns payment URL or client secret
4. **Payment Gateway**:
   - Saves payment record with status "pending"
   - Returns payment URL to merchant
5. **Webhook Callback** (async):
   - Payment provider sends webhook to `/webhooks/stripe`
   - Payment Gateway validates signature
   - Updates payment status to "success"
   - Calls Order Service to update order status
   - Triggers notification (future)

### Channel Adapter Pattern

The Channel Adapter uses the **Adapter Pattern** for payment providers:

```go
// internal/adapter/adapter.go defines interface
type PaymentAdapter interface {
    GetChannel() string
    CreatePayment(ctx, *CreatePaymentRequest) (*CreatePaymentResponse, error)
    QueryPayment(ctx, channelTradeNo string) (*QueryPaymentResponse, error)
    CancelPayment(ctx, channelTradeNo string) error
    CreateRefund(ctx, *CreateRefundRequest) (*CreateRefundResponse, error)
    // ...
}

// internal/adapter/stripe_adapter.go implements the interface
type StripeAdapter struct {
    config *model.StripeConfig
}

// Registered in main.go
adapterFactory := adapter.NewAdapterFactory()
stripeAdapter := adapter.NewStripeAdapter(stripeConfig)
adapterFactory.Register(model.ChannelStripe, stripeAdapter)
```

**Currently implemented**: Stripe (using stripe-go v76)
**Planned**: PayPal, cryptocurrency adapters

## Important Patterns and Conventions

### Error Handling

Services return errors up the stack. HTTP handlers convert errors to JSON responses:

```go
// In handlers
if err != nil {
    c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
    return
}
```

### Database Transactions

Repository layer uses GORM. Transactions should be handled in the service layer:

```go
err := s.db.Transaction(func(tx *gorm.DB) error {
    // Multiple operations
    return nil
})
```

### Logging

Use structured logging from `pkg/logger`:

```go
logger.Info("Starting service...")
logger.Error("Failed to connect", zap.Error(err))
logger.Fatal("Critical error") // Calls os.Exit(1)
```

### Configuration

All services read configuration from environment variables with defaults:

```go
dbConfig := db.Config{
    Host:     config.GetEnv("DB_HOST", "localhost"),
    Port:     config.GetEnvInt("DB_PORT", 5432),
    // ...
}
```

### Money Handling

**Critical**: All monetary amounts are stored as **integers in cents/smallest currency unit** to avoid floating-point precision errors:

```go
// Amount in cents (100 = $1.00)
Amount int64 `json:"amount"`
```

Use `pkg/validator` to validate amounts and currencies (supports 32+ currencies including crypto).

## Code Modification Guidelines

### Adding a New Service

1. Create directory under `backend/services/new-service/`
2. Initialize go module: `go mod init payment-platform/new-service`
3. Add to `backend/go.work`: `use ./services/new-service`
4. Create standard structure: `cmd/main.go`, `internal/{model,repository,service,handler}`
5. Import from `github.com/payment-platform/pkg/*` for shared functionality
6. Follow existing service patterns (see order-service or payment-gateway)

### Adding a New Payment Channel

1. Create adapter: `channel-adapter/internal/adapter/newchannel_adapter.go`
2. Implement `PaymentAdapter` interface
3. Add configuration model: `channel-adapter/internal/model/channel_config.go`
4. Register in `channel-adapter/cmd/main.go`:
   ```go
   newAdapter := adapter.NewChannelAdapter(config)
   adapterFactory.Register(model.ChannelNew, newAdapter)
   ```

### Modifying pkg/ (Shared Library)

**Warning**: Changes to `pkg/` affect ALL services. Always:
1. Run `go mod tidy` in the service after pkg changes
2. Test compilation of all services
3. Maintain backward compatibility

### Database Schema Changes

1. Add/modify model structs with GORM tags
2. GORM AutoMigrate handles schema creation (see `main.go`)
3. For complex migrations, use `scripts/migrate.sh`

## Common Issues

### Import Path Resolution

Services import pkg using the module name:
```go
import "github.com/payment-platform/pkg/logger"
```

This works because of the `replace` directive in each service's `go.mod`:
```go
replace github.com/payment-platform/pkg => ../../pkg
```

### Order Service Compilation

If order-service fails to compile with "missing method" errors, run:
```bash
cd backend/services/order-service
go clean -cache
go mod tidy
go build ./cmd/main.go
```

This is usually a Go build cache issue.

### Stripe API Version

The project uses `stripe-go v76`. Key differences from earlier versions:
- `PaymentIntent.Charges` is removed - use `LatestCharge` instead
- Must import `github.com/stripe/stripe-go/v76/charge` explicitly
- Webhook signature verification uses `webhook.ConstructEvent()`

## Project Status

**Completed** (85% overall):
- ✅ All 10 microservices compile successfully
- ✅ Core payment flow (gateway → channel adapter → Stripe)
- ✅ Shared pkg library (15 packages)
- ✅ JWT authentication and RBAC
- ✅ Stripe payment integration (create, query, refund, webhooks)

**In Progress**:
- ⏳ PayPal and cryptocurrency adapters
- ⏳ Complete risk assessment logic
- ⏳ Notification templates and delivery
- ⏳ Accounting reconciliation

**Not Started**:
- ❌ Docker Compose configuration
- ❌ Environment configuration templates (.env)
- ❌ Integration tests
- ❌ Frontend applications (mentioned in README but not present)

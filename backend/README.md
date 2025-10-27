# Payment Platform Backend

Enterprise-grade payment gateway backend built with Go microservices architecture. This directory contains 19 independent microservices, shared libraries, automation scripts, and comprehensive documentation.

## 🚀 Quick Start (Choose Your Method)

### Option 1: Docker Deployment (Recommended for Production) 🐳

**One-command deployment:**
```bash
cd /home/eric/payment
./scripts/deploy-all.sh
```

This automated script will:
1. ✅ Check system requirements
2. ✅ Generate mTLS certificates
3. ✅ Start infrastructure (PostgreSQL, Redis, Kafka)
4. ✅ Initialize 19 databases
5. ✅ Build all Docker images
6. ✅ Start all 19 services
7. ✅ Run health checks

**Access services:**
- Admin BFF: http://localhost:40001/swagger/index.html
- Merchant BFF: http://localhost:40023/swagger/index.html
- Prometheus: http://localhost:40090
- Grafana: http://localhost:40300 (admin/admin)
- Jaeger: http://localhost:50686

**Stop all services:**
```bash
./scripts/stop-all.sh
```

📖 **Complete Docker Guide:** See [../DOCKER_DEPLOYMENT_GUIDE.md](../DOCKER_DEPLOYMENT_GUIDE.md)

---

### Option 2: Local Development (Hot Reload) 🔥

**Prerequisites:**
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Kafka 3.5+

**Start infrastructure:**
```bash
# Start PostgreSQL, Redis, Kafka
docker-compose up -d postgres redis kafka
```

**Initialize databases:**
```bash
cd backend
./scripts/init-db.sh
```

**Start all services with hot reload:**
```bash
./scripts/start-all-services.sh
```

**Check status:**
```bash
./scripts/status-all-services.sh
```

**Stop services:**
```bash
./scripts/stop-all-services.sh
```

---

## 📋 Architecture Overview

```
Backend Architecture (19 Services + 2 BFF)
├── BFF Layer (API Gateways)
│   ├── admin-bff-service (40001)      - Admin portal gateway (8-layer security)
│   └── merchant-bff-service (40023)   - Merchant portal gateway (tenant isolation)
│
├── Core Payment Flow
│   ├── payment-gateway (40003)        - Payment orchestration, Saga, Kafka
│   ├── order-service (40004)          - Order lifecycle, Event publishing
│   ├── channel-adapter (40005)        - 4 payment channels (Stripe/PayPal/Alipay/Crypto)
│   ├── risk-service (40006)           - Risk scoring, GeoIP, Rules engine
│   ├── accounting-service (40007)     - Double-entry bookkeeping, Kafka consumer
│   └── analytics-service (40009)      - Real-time analytics, Event consumer
│
└── Business Support Services
    ├── notification-service (40008)   - Email, SMS, Webhook
    ├── config-service (40010)         - System config, Feature flags
    ├── merchant-auth-service (40011)  - 2FA, API keys, Sessions
    ├── settlement-service (40013)     - Auto settlement, Saga
    ├── withdrawal-service (40014)     - Withdrawal processing, Bank integration
    ├── kyc-service (40015)            - KYC verification
    ├── cashier-service (40016)        - Checkout UI configuration
    ├── reconciliation-service (40020) - Auto reconciliation
    ├── dispute-service (40021)        - Dispute handling, Stripe sync
    ├── merchant-policy-service (40022)- Merchant fee & limit config
    └── merchant-quota-service (40024) - Tiered limits, Quota tracking
```

---

## 🐳 Docker Deployment

### Build All Docker Images

```bash
# Automated build script (recommended)
cd backend
./scripts/build-all-docker-images.sh

# Or use docker-compose
cd ..
docker-compose -f docker-compose.services.yml build
docker-compose -f docker-compose.bff.yml build
```

### Generate Dockerfiles

```bash
# Auto-generate Dockerfiles for all 19 services
cd backend
./scripts/generate-dockerfiles.sh
```

### Docker Compose Files

Located in project root:

| File | Purpose | Services |
|------|---------|----------|
| `docker-compose.yml` | Infrastructure | PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger, Kong |
| `docker-compose.services.yml` | Core Services | 17 microservices |
| `docker-compose.bff.yml` | BFF Services | Admin BFF + Merchant BFF |

### mTLS Configuration

All Docker services use mTLS for inter-service communication:

```bash
# Generate mTLS certificates
cd backend/certs

# Generate CA certificate
./generate-ca-cert.sh

# Generate service certificates (all 19 services)
for service in payment-gateway order-service channel-adapter risk-service \
               accounting-service notification-service analytics-service \
               config-service merchant-auth-service settlement-service \
               withdrawal-service kyc-service cashier-service \
               reconciliation-service dispute-service merchant-policy-service \
               merchant-quota-service admin-bff-service merchant-bff-service; do
    ./generate-service-cert.sh $service
done
```

### Docker Network

Services communicate via internal domain names:

```
Network: payment-network (172.28.0.0/16)
Domain format: <service-name>.payment-network

Examples:
- payment-gateway.payment-network:40003
- order-service.payment-network:40004
- postgres.payment-network:5432
- redis.payment-network:6379
```

### Resource Limits

Each service has resource quotas:

```yaml
deploy:
  resources:
    limits:
      cpus: '1.0'          # Max 1 CPU core
      memory: 512M         # Max 512MB RAM
    reservations:
      cpus: '0.5'          # Reserve 0.5 CPU
      memory: 256M         # Reserve 256MB RAM
```

### Docker Commands

```bash
# View all containers
docker ps

# View service logs
docker-compose -f docker-compose.services.yml logs -f payment-gateway

# Restart service
docker-compose -f docker-compose.services.yml restart payment-gateway

# Scale service
docker-compose -f docker-compose.services.yml up -d --scale payment-gateway=3

# Verify deployment
./scripts/verify-deployment.sh
```

---

## 🛠️ Development

### Building Services

```bash
# Build all services
cd backend
make build

# Build specific service
cd services/payment-gateway
go build -o ../../bin/payment-gateway ./cmd/main.go

# Clean build cache
make clean
```

### Testing

```bash
# Run all tests
make test

# Test specific service
cd services/payment-gateway
go test ./...

# Test with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# End-to-end payment flow test
./scripts/test-payment-flow.sh

# System health check
./scripts/system-health-check.sh
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Install linter (first time only)
make install-lint
```

### API Documentation (Swagger)

```bash
# Generate Swagger docs for all services
make swagger-docs

# Install swag CLI (first time only)
make install-swagger

# Generate docs for specific service
cd services/payment-gateway
swag init -g cmd/main.go -o docs
```

**Access Swagger UI:**
- Admin BFF: http://localhost:40001/swagger/index.html
- Merchant BFF: http://localhost:40023/swagger/index.html
- Payment Gateway: http://localhost:40003/swagger/index.html
- Order Service: http://localhost:40004/swagger/index.html

📖 See [API_DOCUMENTATION_GUIDE.md](API_DOCUMENTATION_GUIDE.md) for complete documentation.

---

## 🏗️ Service Architecture

### Standard Service Structure

```
service-name/
├── cmd/
│   └── main.go              # Entry point with Bootstrap
├── internal/
│   ├── model/               # GORM data models
│   ├── repository/          # Data access layer
│   ├── service/             # Business logic
│   ├── handler/             # HTTP handlers (Gin)
│   ├── client/              # HTTP clients (optional)
│   └── middleware/          # Service middleware (optional)
├── docs/                    # Swagger documentation
├── Dockerfile               # Docker build configuration
├── .dockerignore            # Docker ignore rules
├── go.mod                   # Module definition
└── README.md               # Service documentation
```

### Bootstrap Framework (100% Adoption ✅)

All 19 services use `pkg/app` Bootstrap framework:

```go
import "github.com/payment-platform/pkg/app"

application, err := app.Bootstrap(app.ServiceConfig{
    ServiceName: "payment-gateway",
    DBName:      "payment_gateway",
    Port:        40003,
    AutoMigrate: []any{&model.Payment{}, &model.Refund{}},

    // Feature flags (optional, sensible defaults)
    EnableTracing:     true,   // Jaeger tracing
    EnableMetrics:     true,   // Prometheus metrics
    EnableRedis:       true,   // Redis connection
    EnableGRPC:        false,  // gRPC disabled (HTTP/REST primary)
    EnableHealthCheck: true,   // Enhanced health checks
    EnableRateLimit:   true,   // Rate limiting

    RateLimitRequests: 100,
    RateLimitWindow:   time.Minute,
})

// Register HTTP routes
handler.RegisterRoutes(application.Router, authMiddleware)

// Start with graceful shutdown
application.RunWithGracefulShutdown()
```

**Benefits:**
- ✅ 38.7% average code reduction
- ✅ Auto-configures DB, Redis, Logger, Router, Middleware
- ✅ Built-in tracing, metrics, health checks
- ✅ HTTP-first communication (gRPC optional)

📖 See [BOOTSTRAP_MIGRATION_FINAL_100PERCENT.md](BOOTSTRAP_MIGRATION_FINAL_100PERCENT.md)

### Inter-Service Communication

Services use **HTTP/REST** (not gRPC):

```go
// Payment Gateway → Order Service
orderClient := client.NewOrderClient("http://localhost:40004")
response, err := orderClient.CreateOrder(ctx, &client.CreateOrderRequest{
    MerchantID: merchantID,
    Amount:     amount,
    Currency:   currency,
})
```

**Docker deployment uses HTTPS + mTLS:**
```bash
ORDER_SERVICE_URL=https://order-service.payment-network:40004
```

---

## 📦 Shared Libraries (pkg/)

20 reusable packages used by all services:

### Core Infrastructure
- **app/** - Bootstrap framework
- **auth/** - JWT, password hashing
- **cache/** - Redis/in-memory cache
- **config/** - Environment variables
- **db/** - PostgreSQL/Redis pooling
- **logger/** - Zap structured logging
- **validator/** - Amount, currency validation

### Communication
- **email/** - SMTP/Mailgun
- **httpclient/** - Retry, circuit breaker
- **kafka/** - Producer/consumer
- **grpc/** - gRPC utilities (optional)

### Observability
- **metrics/** - Prometheus metrics
- **tracing/** - Jaeger tracing (W3C)
- **health/** - Health endpoints

### Middleware
- **middleware/** - CORS, Auth, RateLimit, RequestID, Logger, Metrics, Tracing

### Utilities
- **crypto/** - Encryption
- **currency/** - Multi-currency
- **retry/** - Exponential backoff
- **migration/** - Database migrations

---

## 🌐 Service Ports

| Service | HTTP | Database | Docker | Status |
|---------|------|----------|--------|--------|
| admin-bff-service | 40001 | payment_admin | ✅ | Production Ready |
| merchant-bff-service | 40023 | payment_merchant | ✅ | Production Ready |
| payment-gateway | 40003 | payment_gateway | ✅ | Production Ready |
| order-service | 40004 | payment_order | ✅ | Production Ready |
| channel-adapter | 40005 | payment_channel | ✅ | Production Ready |
| risk-service | 40006 | payment_risk | ✅ | Production Ready |
| accounting-service | 40007 | payment_accounting | ✅ | Production Ready |
| notification-service | 40008 | payment_notification | ✅ | Production Ready |
| analytics-service | 40009 | payment_analytics | ✅ | Production Ready |
| config-service | 40010 | payment_config | ✅ | Production Ready |
| merchant-auth-service | 40011 | payment_merchant_auth | ✅ | Production Ready |
| settlement-service | 40013 | payment_settlement | ✅ | Production Ready |
| withdrawal-service | 40014 | payment_withdrawal | ✅ | Production Ready |
| kyc-service | 40015 | payment_kyc | ✅ | Production Ready |
| cashier-service | 40016 | payment_cashier | ✅ | Production Ready |
| reconciliation-service | 40020 | payment_reconciliation | ✅ | Production Ready |
| dispute-service | 40021 | payment_dispute | ✅ | Production Ready |
| merchant-policy-service | 40022 | payment_merchant_policy | ✅ | Production Ready |
| merchant-quota-service | 40024 | payment_merchant_quota | ✅ | Production Ready |

**Infrastructure:**
- PostgreSQL: 40432 (docker) / 5432 (local)
- Redis: 40379 (docker) / 6379 (local)
- Kafka: 40092 (docker) / 9092 (local)
- Prometheus: 40090
- Grafana: 40300 (admin/admin)
- Jaeger UI: 50686
- Kong Gateway: 40080

---

## ⚙️ Environment Variables

Common environment variables:

```bash
# Environment
ENV=development               # development | production

# Database (PostgreSQL)
DB_HOST=localhost            # postgres.payment-network in Docker
DB_PORT=40432                # 5432 for local
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_<service>

# Redis
REDIS_HOST=localhost         # redis.payment-network in Docker
REDIS_PORT=40379             # 6379 for local
REDIS_PASSWORD=

# Kafka
KAFKA_BROKERS=localhost:40092  # kafka.payment-network:9092 in Docker

# Service Configuration
PORT=40001
JWT_SECRET=your-secret-key-change-in-production

# mTLS (Docker only)
ENABLE_MTLS=true
ENABLE_HTTPS=true
TLS_CERT_FILE=/app/certs/services/<service>/<service>.crt
TLS_KEY_FILE=/app/certs/services/<service>/<service>.key
TLS_CA_FILE=/app/certs/ca/ca-cert.pem

# Observability
JAEGER_ENDPOINT=http://localhost:14268/api/traces
JAEGER_SAMPLING_RATE=100     # Use 10-20 for production

# Payment Channels
STRIPE_API_KEY=sk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-password
```

---

## 📊 Monitoring

### Health Checks

```bash
# Individual service
curl http://localhost:40003/health

# Kubernetes probes
curl http://localhost:40003/health/ready
curl http://localhost:40003/health/live

# Full system health
./scripts/system-health-check.sh

# Docker deployment verification
./scripts/verify-deployment.sh
```

### Prometheus Metrics

```bash
# Access metrics
curl http://localhost:40003/metrics

# Prometheus UI
open http://localhost:40090

# Common queries
rate(http_requests_total{service="payment-gateway"}[5m])
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

### Jaeger Tracing

```bash
# Jaeger UI
open http://localhost:50686

# Trace complete payment flow:
# Payment Gateway → Order Service → Channel Adapter → Risk Service
```

### Grafana Dashboards

```bash
# Grafana UI
open http://localhost:40300  # admin/admin

# Pre-configured dashboards:
# - Service Health Overview
# - Payment Flow Monitoring
# - Database Performance
# - Kafka Message Queue
# - Container Resource Usage
```

---

## 🔧 Useful Scripts

Located in `scripts/`:

| Script | Description | Usage |
|--------|-------------|-------|
| `generate-dockerfiles.sh` | Generate all Dockerfiles | Backend scripts |
| `generate-docker-compose-services.sh` | Generate docker-compose.services.yml | Backend scripts |
| `build-all-docker-images.sh` | Build all Docker images | Backend scripts |
| `start-all-services.sh` | Start with hot reload | Local dev |
| `stop-all-services.sh` | Stop all services | Local dev |
| `status-all-services.sh` | Check service status | Local dev |
| `system-health-check.sh` | Full health check | Both |
| `test-payment-flow.sh` | E2E payment test | Both |
| `init-db.sh` | Initialize 19 databases | Both |
| `verify-all-services.sh` | Verify compilation | Local dev |
| `deploy-all.sh` | One-click Docker deployment | Root scripts ⭐ |
| `stop-all.sh` | Stop Docker services | Root scripts |
| `verify-deployment.sh` | Verify Docker deployment | Root scripts |

---

## 🐛 Troubleshooting

### Docker Issues

```bash
# Service won't start
docker logs payment-payment-gateway

# Database connection failed
docker exec payment-payment-gateway ping postgres.payment-network

# mTLS certificate error
docker exec payment-payment-gateway ls -la /app/certs/services/payment-gateway/

# Restart service
docker-compose -f docker-compose.services.yml restart payment-gateway
```

### Local Development Issues

```bash
# Service won't start
tail -f logs/payment-gateway.log

# Port conflict
lsof -i :40003
kill -9 <pid>

# Database connection
psql -h localhost -p 40432 -U postgres -l

# Clean and rebuild
go clean -cache
cd services/payment-gateway
go mod tidy
go build ./cmd/main.go
```

### Import Path Resolution

```go
// Services import pkg using:
import "github.com/payment-platform/pkg/logger"

// Works via replace directive in go.mod:
replace github.com/payment-platform/pkg => ../../pkg
```

---

## 📚 Documentation

### Complete Documentation

- **[../DOCKER_DEPLOYMENT_GUIDE.md](../DOCKER_DEPLOYMENT_GUIDE.md)** - ⭐ Complete Docker deployment guide
- **[../DOCKER_PACKAGE_SUMMARY.md](../DOCKER_PACKAGE_SUMMARY.md)** - Docker packaging summary
- **[../DOCKER_README.md](../DOCKER_README.md)** - Docker quick start
- **[API_DOCUMENTATION_GUIDE.md](API_DOCUMENTATION_GUIDE.md)** - API documentation
- **[QUICK_START.md](QUICK_START.md)** - Quick start guide
- **[BOOTSTRAP_MIGRATION_FINAL_100PERCENT.md](BOOTSTRAP_MIGRATION_FINAL_100PERCENT.md)** - Bootstrap migration
- **[BFF_SECURITY_COMPLETE_SUMMARY.md](BFF_SECURITY_COMPLETE_SUMMARY.md)** - BFF security
- **[SWAGGER_QUICK_REFERENCE.md](SWAGGER_QUICK_REFERENCE.md)** - Swagger syntax

### Individual Service Documentation

Each service has its own README in `services/<service-name>/README.md`

---

## 🔨 Contributing

### Adding a New Service

1. **Create structure:**
```bash
mkdir -p services/new-service/{cmd,internal/{model,repository,service,handler}}
```

2. **Initialize module:**
```bash
cd services/new-service
go mod init payment-platform/new-service
```

3. **Add to workspace:**
```bash
# Edit backend/go.work
use ./services/new-service
```

4. **Use Bootstrap framework in cmd/main.go**

5. **Generate Dockerfile:**
```bash
cd ../..
./scripts/generate-dockerfiles.sh
```

### Modifying Shared Libraries

⚠️ **Warning:** Changes to `pkg/` affect ALL 19 services.

Always:
1. Run `go mod tidy` in affected services
2. Test compilation of all services
3. Maintain backward compatibility
4. Update documentation

---

## 🚢 Production Deployment

### Docker Production Configuration

```yaml
# Use production environment
ENV=production

# Reduce Jaeger sampling
JAEGER_SAMPLING_RATE=10  # 10% instead of 100%

# Configure SSL/TLS
ENABLE_MTLS=true
ENABLE_HTTPS=true

# Set strong secrets
JWT_SECRET=<strong-secret-256-bits>
DB_PASSWORD=<strong-password>
REDIS_PASSWORD=<strong-password>

# Configure backups
DB_BACKUP_ENABLED=true
DB_BACKUP_SCHEDULE="0 2 * * *"
```

### Resource Requirements

**Development:**
- CPU: 4 cores
- Memory: 8 GB
- Disk: 50 GB

**Production:**
- CPU: 16 cores
- Memory: 32 GB
- Disk: 500 GB SSD

### Health Endpoints

All services expose:
- `/health` - Basic health
- `/health/ready` - Readiness probe (K8s)
- `/health/live` - Liveness probe (K8s)

### Graceful Shutdown

```bash
# Services handle SIGINT/SIGTERM
kill -TERM <pid>

# Docker stops gracefully
docker-compose -f docker-compose.services.yml down
```

---

## 📝 License

Commercial License

## 🆘 Support

- **Docker Guide**: [../DOCKER_DEPLOYMENT_GUIDE.md](../DOCKER_DEPLOYMENT_GUIDE.md)
- **Project Docs**: [../CLAUDE.md](../CLAUDE.md)
- **Issues**: GitHub issue tracker
- **Email**: support@payment-platform.com

---

**🎉 Ready to deploy? Run `./scripts/deploy-all.sh` from project root!**

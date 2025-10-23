# Payment Platform - Environment Setup Guide

## Overview

This document describes the environment configuration and management system for the payment platform's 14 microservices.

## What's New

### 1. Environment Configuration Template

**File:** `backend/.env.example`

A comprehensive configuration template covering:
- Database settings (14 separate databases)
- Redis configuration
- Kafka settings
- Service ports (8001-8015)
- Inter-service communication URLs
- Authentication secrets
- Payment channel credentials (Stripe, PayPal, Crypto)
- Notification settings (Email, SMS)
- Observability configuration (Prometheus, Jaeger)
- Feature flags and business rules

**Usage:**
```bash
cp backend/.env.example backend/.env
# Edit .env with your actual values
```

### 2. Unified Management Scripts

Three new comprehensive scripts in `backend/scripts/`:

#### start-all.sh
- **Purpose:** Start all 14 services with dependency management
- **Features:**
  - Infrastructure validation (PostgreSQL, Redis)
  - Automatic service building if binaries missing
  - Dependency-aware startup order
  - Health check verification
  - Detailed progress reporting
- **Usage:** `./scripts/start-all.sh`

#### health-check.sh
- **Purpose:** Monitor service health status
- **Features:**
  - Infrastructure connectivity checks
  - Service health endpoint validation
  - Process ID display
  - Watch mode for continuous monitoring
  - Detailed metrics display
- **Usage:**
  ```bash
  ./scripts/health-check.sh              # One-time check
  ./scripts/health-check.sh --watch      # Continuous monitoring
  ./scripts/health-check.sh --detailed   # With Prometheus metrics
  ```

#### stop-all.sh
- **Purpose:** Gracefully shutdown all services
- **Features:**
  - Reverse dependency order shutdown
  - Graceful termination (SIGTERM)
  - Force kill option (SIGKILL)
  - Log cleanup option
  - Comprehensive status reporting
- **Usage:**
  ```bash
  ./scripts/stop-all.sh               # Graceful shutdown
  ./scripts/stop-all.sh --force       # Force kill
  ./scripts/stop-all.sh --cleanup     # With log cleanup
  ```

### 3. Documentation

**File:** `backend/scripts/README.md`

Complete documentation covering:
- Quick start guide
- Script usage and options
- Environment configuration
- Log management
- Troubleshooting guide
- Development workflow
- Advanced usage patterns

## Service Architecture

### All 14 Microservices

| Service | Port | Database | Status |
|---------|------|----------|--------|
| admin-service | 8001 | payment_admin | ✅ Production |
| merchant-service | 8002 | payment_merchant | ✅ Production (gRPC) |
| payment-gateway | 8003 | payment_gateway | ✅ Production |
| order-service | 8004 | payment_order | ✅ Production |
| channel-adapter | 8005 | payment_channel | ✅ Production |
| risk-service | 8006 | payment_risk | ⏳ Basic |
| notification-service | 8007 | payment_notify | ⏳ Basic |
| accounting-service | 8008 | payment_accounting | ⏳ Basic |
| analytics-service | 8009 | payment_analytics | ⏳ Basic |
| config-service | 8010 | payment_config | ⏳ Basic |
| settlement-service | 8012 | payment_settlement | ✅ Production |
| withdrawal-service | 8013 | payment_withdrawal | ✅ Production |
| kyc-service | 8014 | payment_kyc | ✅ Production |
| fee-service | 8015 | payment_fee | ⏳ Placeholder |

### Dependency Graph

```
config-service (8010)
├── admin-service (8001)
├── merchant-service (8002)
└── channel-adapter (8005)

risk-service (8006) [independent]
kyc-service (8014) [independent]
fee-service (8015) [independent]

order-service (8004) [independent]
└── payment-gateway (8003)
    └── depends on: order-service, channel-adapter, risk-service

accounting-service (8008) [independent]
├── settlement-service (8012)
└── withdrawal-service (8013)

notification-service (8007) [independent]
analytics-service (8009) [independent]
```

## Quick Start Guide

### 1. Infrastructure Setup

```bash
# Start Docker services
cd /home/eric/payment
docker-compose up -d

# Verify infrastructure
docker-compose ps
```

### 2. Configure Environment

```bash
cd /home/eric/payment/backend

# Create .env from template
cp .env.example .env

# Edit configuration
vim .env
```

**Minimum required changes:**
- `JWT_SECRET` - Change from default
- `SIGNATURE_SECRET` - Change from default
- `STRIPE_API_KEY` - Add your Stripe key
- `STRIPE_WEBHOOK_SECRET` - Add your webhook secret

### 3. Start Services

```bash
# Start all services
./scripts/start-all.sh

# Monitor startup
watch ./scripts/health-check.sh
```

### 4. Verify Health

```bash
# Check all services
./scripts/health-check.sh

# Expected output: 14 services healthy
```

## Infrastructure Details

### Docker Services

All accessible via docker-compose:

**Data Stores:**
- PostgreSQL: `localhost:40432` (14 databases)
- Redis: `localhost:40379`
- Kafka: `localhost:40092`

**Monitoring:**
- Prometheus: `http://localhost:40090`
- Grafana: `http://localhost:40300` (admin/admin)
- Jaeger UI: `http://localhost:40686`

**Exporters:**
- postgres-exporter: Port 40432 metrics
- redis-exporter: Port 40379 metrics
- kafka-exporter: Port 40092 metrics
- cadvisor: Container metrics
- node-exporter: System metrics

### Log Management

All services log to `/tmp/<service-name>.log`:

```bash
# View all logs
tail -f /tmp/*-service.log

# View specific service
tail -f /tmp/payment-gateway.log

# Search for errors
grep -i error /tmp/*.log

# Clean logs
rm /tmp/*-service.log
```

## Development Workflow

### Standard Workflow

```bash
# 1. Start infrastructure
docker-compose up -d

# 2. Start all services
cd backend
./scripts/start-all.sh

# 3. Make code changes
vim services/payment-gateway/internal/handler/payment_handler.go

# 4. Rebuild and restart
./scripts/stop-all.sh
cd services/payment-gateway
go build -o /tmp/payment-gateway ./cmd/main.go
cd ../..
./scripts/start-all.sh

# 5. Test changes
curl -X POST http://localhost:8003/api/v1/payments ...

# 6. Check logs
tail -f /tmp/payment-gateway.log
```

### Hot Reload Development

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run service with hot reload
cd services/payment-gateway
air

# Or use dev script
cd backend
./scripts/dev-with-air.sh
```

## Configuration Management

### Environment Variables

**Priority order:**
1. Command-line environment variables
2. `.env` file in `backend/`
3. Script defaults

**Example:**
```bash
# Override with environment variable
DB_PORT=5432 ./scripts/start-all.sh

# Or set in .env file
echo "DB_PORT=5432" >> backend/.env
./scripts/start-all.sh
```

### Service-Specific Configuration

Each service reads from:
- Shared variables (DB_HOST, REDIS_HOST, etc.)
- Service-specific variables (PORT, DB_NAME)
- Feature flags (RATE_LIMIT_ENABLED, etc.)

See `.env.example` for complete list.

## Monitoring and Observability

### Metrics

**Prometheus metrics available at each service:**
```bash
# View metrics
curl http://localhost:8003/metrics

# Key metrics:
# - http_requests_total - Total HTTP requests
# - http_request_duration_seconds - Request latency
# - http_requests_in_flight - Current active requests
```

### Tracing

**Jaeger distributed tracing:**
1. Access UI: http://localhost:40686
2. Search by service name
3. View trace details and spans

### Health Checks

**Each service provides:**
```bash
curl http://localhost:8003/health
# {"service":"payment-gateway","status":"ok","time":1234567890}
```

## Troubleshooting

### Common Issues

**1. Service fails to start**
```bash
# Check logs
tail -50 /tmp/<service-name>.log

# Verify port availability
lsof -i :<port>

# Check infrastructure
docker-compose ps
```

**2. Database connection errors**
```bash
# Test PostgreSQL
pg_isready -h localhost -p 40432 -U postgres

# List databases
docker exec payment-postgres psql -U postgres -l

# Connect to database
docker exec -it payment-postgres psql -U postgres -d payment_gateway
```

**3. Service shows unhealthy**
```bash
# Check dependencies
./scripts/health-check.sh

# View service logs
tail -f /tmp/<service-name>.log

# Restart service
pkill <service-name>
# Then restart with ./scripts/start-all.sh
```

**4. Cannot stop service**
```bash
# Force stop
./scripts/stop-all.sh --force

# Or manually
kill -9 $(lsof -t -i:<port>)
```

## Migration from Old Scripts

### Old vs New Scripts

| Old Script | New Script | Improvements |
|------------|------------|--------------|
| start-all-services.sh | start-all.sh | - Dependency management<br>- Health verification<br>- .env support<br>- Better error handling |
| status-all-services.sh | health-check.sh | - Watch mode<br>- Metrics display<br>- Infrastructure checks<br>- Color-coded output |
| stop-all-services.sh | stop-all.sh | - Graceful shutdown<br>- Reverse order<br>- Log cleanup option<br>- Better error handling |

### Migrating

**Option 1: Keep both**
```bash
# Old scripts still work
./scripts/start-all-services.sh

# New scripts have more features
./scripts/start-all.sh
```

**Option 2: Replace old scripts**
```bash
cd backend/scripts
rm start-all-services.sh status-all-services.sh stop-all-services.sh
ln -s start-all.sh start.sh
ln -s health-check.sh status.sh
ln -s stop-all.sh stop.sh
```

## Next Steps

### Immediate (Recommended)

1. **Copy .env template:**
   ```bash
   cp backend/.env.example backend/.env
   vim backend/.env
   ```

2. **Test startup:**
   ```bash
   ./scripts/start-all.sh
   ./scripts/health-check.sh --watch
   ```

3. **Verify monitoring:**
   - Grafana: http://localhost:40300
   - Jaeger: http://localhost:40686

### Short-term

1. **Write integration tests** - Test complete payment flows
2. **Enhance basic services** - Complete business logic for 5 basic services
3. **Add payment channels** - Implement PayPal and crypto adapters

### Long-term

1. **API Gateway** - Add Kong or similar
2. **Service mesh** - Consider Istio for advanced traffic management
3. **Frontend** - Build admin and merchant portals

## Support

For issues or questions:
1. Check `backend/scripts/README.md` for detailed documentation
2. Review service logs in `/tmp/*.log`
3. See `LOCAL_DEVELOPMENT.md` for development setup
4. Check `PROJECT_STATUS.md` for current implementation status

#!/bin/bash

# Initialize Sprint 2 Services (Reconciliation, Dispute, Merchant Limit)
# This script creates databases, initializes seed data, and starts services

set -e  # Exit on error

echo "========================================"
echo "Sprint 2 Services Initialization"
echo "========================================"

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-40432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Step 1: Check prerequisites
log_info "Checking prerequisites..."

if ! command -v psql &> /dev/null; then
    log_error "psql not found. Please install PostgreSQL client."
    exit 1
fi

if ! command -v go &> /dev/null; then
    log_error "go not found. Please install Go 1.21+."
    exit 1
fi

log_info "✓ Prerequisites OK"

# Step 2: Check database connection
log_info "Testing database connection..."

if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1" &> /dev/null; then
    log_info "✓ Database connection OK"
else
    log_error "Cannot connect to PostgreSQL at $DB_HOST:$DB_PORT"
    log_error "Make sure docker-compose is running: docker-compose up -d"
    exit 1
fi

# Step 3: Create databases
log_info "Creating databases..."

DATABASES=(
    "payment_reconciliation"
    "payment_dispute"
    "payment_merchant_limit"
)

for DB in "${DATABASES[@]}"; do
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB; then
        log_warn "Database $DB already exists, skipping creation"
    else
        PGPASSWORD=$DB_PASSWORD createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB
        log_info "✓ Created database: $DB"
    fi
done

# Step 4: Build services
log_info "Building services..."

cd /home/eric/payment/backend/services

SERVICES=(
    "reconciliation-service"
    "dispute-service"
    "merchant-limit-service"
)

for SERVICE in "${SERVICES[@]}"; do
    log_info "Building $SERVICE..."
    cd $SERVICE
    if GOWORK=/home/eric/payment/backend/go.work go build -o /tmp/$SERVICE ./cmd/main.go; then
        log_info "✓ Built $SERVICE (binary: /tmp/$SERVICE)"
    else
        log_error "Failed to build $SERVICE"
        exit 1
    fi
    cd ..
done

# Step 5: Initialize merchant tiers
log_info "Initializing merchant tiers..."

PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d payment_merchant_limit -f /home/eric/payment/backend/scripts/init-merchant-tiers.sql

log_info "✓ Merchant tiers initialized"

# Step 6: Start services (optional)
read -p "Do you want to start the services now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Starting services..."

    # Create log directory
    mkdir -p /home/eric/payment/backend/logs

    # Start reconciliation-service
    log_info "Starting reconciliation-service on port 40020..."
    cd /home/eric/payment/backend/services/reconciliation-service
    export GOWORK=/home/eric/payment/backend/go.work
    export DB_HOST=$DB_HOST
    export DB_PORT=$DB_PORT
    export DB_USER=$DB_USER
    export DB_PASSWORD=$DB_PASSWORD
    export STRIPE_API_KEY="${STRIPE_API_KEY:-sk_test_placeholder}"
    nohup air -c .air.toml > /home/eric/payment/backend/logs/reconciliation-service.log 2>&1 &
    echo $! > /home/eric/payment/backend/logs/reconciliation-service.pid
    log_info "✓ reconciliation-service started (PID: $!)"

    # Start dispute-service
    log_info "Starting dispute-service on port 40021..."
    cd /home/eric/payment/backend/services/dispute-service
    nohup air -c .air.toml > /home/eric/payment/backend/logs/dispute-service.log 2>&1 &
    echo $! > /home/eric/payment/backend/logs/dispute-service.pid
    log_info "✓ dispute-service started (PID: $!)"

    # Start merchant-limit-service
    log_info "Starting merchant-limit-service on port 40022..."
    cd /home/eric/payment/backend/services/merchant-limit-service
    nohup air -c .air.toml > /home/eric/payment/backend/logs/merchant-limit-service.log 2>&1 &
    echo $! > /home/eric/payment/backend/logs/merchant-limit-service.pid
    log_info "✓ merchant-limit-service started (PID: $!)"

    # Wait for services to start
    sleep 3

    # Health checks
    log_info "Running health checks..."

    PORTS=(40020 40021 40022)
    SERVICE_NAMES=("reconciliation-service" "dispute-service" "merchant-limit-service")

    for i in "${!PORTS[@]}"; do
        PORT=${PORTS[$i]}
        SERVICE=${SERVICE_NAMES[$i]}

        if curl -s http://localhost:$PORT/health > /dev/null; then
            log_info "✓ $SERVICE is healthy (port $PORT)"
        else
            log_warn "$SERVICE may not be ready yet (port $PORT)"
            log_info "  Check logs: tail -f /home/eric/payment/backend/logs/$SERVICE.log"
        fi
    done

    echo ""
    log_info "All services started!"
    log_info ""
    log_info "Service URLs:"
    log_info "  - Reconciliation: http://localhost:40016"
    log_info "  - Dispute:        http://localhost:40017"
    log_info "  - Merchant Limit: http://localhost:40018"
    log_info ""
    log_info "To stop services, run:"
    log_info "  kill \$(cat /home/eric/payment/backend/logs/*.pid)"
    log_info ""
    log_info "To view logs:"
    log_info "  tail -f /home/eric/payment/backend/logs/<service-name>.log"

else
    log_info "Services not started. You can start them manually:"
    log_info "  cd services/<service-name>"
    log_info "  go run cmd/main.go"
fi

echo ""
echo "========================================"
echo "Initialization Complete!"
echo "========================================"
echo ""
echo "Summary:"
echo "  ✓ 3 databases created"
echo "  ✓ 3 services built"
echo "  ✓ 5 merchant tiers initialized"
echo ""
echo "Next steps:"
echo "  1. Test APIs with curl or Postman"
echo "  2. Run integration tests"
echo "  3. Start frontend development"
echo ""

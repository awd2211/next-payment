#!/bin/bash

# ============================================
# Payment Platform - Start All Services
# ============================================
# This script starts all 14 microservices

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="/tmp"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Load environment variables from .env if it exists
if [ -f "$BACKEND_DIR/.env" ]; then
    echo -e "${BLUE}Loading environment from .env${NC}"
    export $(cat "$BACKEND_DIR/.env" | grep -v '^#' | xargs)
else
    echo -e "${YELLOW}No .env file found, using defaults${NC}"
fi

# Default configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-40432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_SSL_MODE=${DB_SSL_MODE:-disable}
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-40379}
JAEGER_ENDPOINT=${JAEGER_ENDPOINT:-http://localhost:40268/api/traces}

# Service definitions: name|port|database|dependencies
SERVICES=(
    "config-service|8010|payment_config|"
    "admin-service|8001|payment_admin|config-service"
    "merchant-service|8002|payment_merchant|config-service"
    "risk-service|8006|payment_risk|"
    "kyc-service|8014|payment_kyc|"
    "fee-service|8015|payment_fee|"
    "channel-adapter|8005|payment_channel|config-service"
    "order-service|8004|payment_order|"
    "payment-gateway|8003|payment_gateway|order-service,channel-adapter,risk-service"
    "accounting-service|8008|payment_accounting|"
    "settlement-service|8012|payment_settlement|accounting-service"
    "withdrawal-service|8013|payment_withdrawal|accounting-service"
    "notification-service|8007|payment_notify|"
    "analytics-service|8009|payment_analytics|"
)

# Check if infrastructure is running
check_infrastructure() {
    echo -e "${BLUE}Checking infrastructure...${NC}"

    # Check PostgreSQL
    if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PostgreSQL is running${NC}"
    else
        echo -e "${RED}✗ PostgreSQL is not running on $DB_HOST:$DB_PORT${NC}"
        echo -e "${YELLOW}Start it with: docker-compose up -d postgres${NC}"
        exit 1
    fi

    # Check Redis
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Redis is running${NC}"
    else
        echo -e "${RED}✗ Redis is not running on $REDIS_HOST:$REDIS_PORT${NC}"
        echo -e "${YELLOW}Start it with: docker-compose up -d redis${NC}"
        exit 1
    fi

    echo ""
}

# Build service if binary doesn't exist
build_service() {
    local service_name=$1
    local service_dir="$BACKEND_DIR/services/$service_name"
    local binary_path="/tmp/$service_name"

    if [ ! -f "$binary_path" ]; then
        echo -e "${YELLOW}Building $service_name...${NC}"
        cd "$service_dir"
        if GOWORK="$BACKEND_DIR/go.work" go build -o "$binary_path" ./cmd/main.go 2>&1 | grep -v "^#"; then
            echo -e "${GREEN}✓ Built $service_name${NC}"
        else
            echo -e "${RED}✗ Failed to build $service_name${NC}"
            return 1
        fi
    fi
}

# Check if service is already running
is_service_running() {
    local service_name=$1
    local port=$2

    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Start a service
start_service() {
    local service_name=$1
    local port=$2
    local database=$3

    # Check if already running
    if is_service_running "$service_name" "$port"; then
        echo -e "${YELLOW}⚠ $service_name is already running on port $port${NC}"
        return 0
    fi

    # Build if needed
    build_service "$service_name" || return 1

    local binary_path="/tmp/$service_name"
    local log_path="$LOG_DIR/$service_name.log"

    # Start service
    echo -e "${BLUE}Starting $service_name on port $port...${NC}"

    DB_HOST=$DB_HOST \
    DB_PORT=$DB_PORT \
    DB_USER=$DB_USER \
    DB_PASSWORD=$DB_PASSWORD \
    DB_NAME=$database \
    DB_SSL_MODE=$DB_SSL_MODE \
    REDIS_HOST=$REDIS_HOST \
    REDIS_PORT=$REDIS_PORT \
    PORT=$port \
    JAEGER_ENDPOINT=$JAEGER_ENDPOINT \
    ORDER_SERVICE_URL=${ORDER_SERVICE_URL:-http://localhost:8004} \
    CHANNEL_SERVICE_URL=${CHANNEL_SERVICE_URL:-http://localhost:8005} \
    RISK_SERVICE_URL=${RISK_SERVICE_URL:-http://localhost:8006} \
    MERCHANT_SERVICE_URL=${MERCHANT_SERVICE_URL:-http://localhost:8002} \
    CONFIG_SERVICE_URL=${CONFIG_SERVICE_URL:-http://localhost:8010} \
    KYC_SERVICE_URL=${KYC_SERVICE_URL:-http://localhost:8014} \
    FEE_SERVICE_URL=${FEE_SERVICE_URL:-http://localhost:8015} \
    JWT_SECRET=${JWT_SECRET:-default-jwt-secret-change-in-production} \
    $binary_path > "$log_path" 2>&1 &

    local pid=$!

    # Wait for service to start
    echo -n "Waiting for $service_name to be ready"
    for i in {1..30}; do
        if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            echo ""
            echo -e "${GREEN}✓ $service_name started (PID: $pid)${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
    done

    echo ""
    echo -e "${RED}✗ $service_name failed to start (check $log_path)${NC}"
    return 1
}

# Wait for dependent services
wait_for_dependencies() {
    local deps=$1

    if [ -z "$deps" ]; then
        return 0
    fi

    IFS=',' read -ra DEP_ARRAY <<< "$deps"
    for dep in "${DEP_ARRAY[@]}"; do
        local dep_port=$(echo "${SERVICES[@]}" | tr ' ' '\n' | grep "^$dep|" | cut -d'|' -f2)
        if [ -n "$dep_port" ]; then
            echo -n "Waiting for dependency $dep on port $dep_port..."
            for i in {1..30}; do
                if curl -s "http://localhost:$dep_port/health" > /dev/null 2>&1; then
                    echo " ready"
                    break
                fi
                sleep 1
            done
        fi
    done
}

# Main execution
main() {
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}Payment Platform - Starting Services${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""

    # Check infrastructure
    check_infrastructure

    # Start services in order
    local failed_services=()
    local started_services=()

    for service_def in "${SERVICES[@]}"; do
        IFS='|' read -r service_name port database deps <<< "$service_def"

        echo ""
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

        # Wait for dependencies
        if [ -n "$deps" ]; then
            wait_for_dependencies "$deps"
        fi

        # Start service
        if start_service "$service_name" "$port" "$database"; then
            started_services+=("$service_name")
        else
            failed_services+=("$service_name")
        fi

        sleep 1
    done

    # Summary
    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}Startup Summary${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""

    if [ ${#started_services[@]} -gt 0 ]; then
        echo -e "${GREEN}Started services (${#started_services[@]}):${NC}"
        for service in "${started_services[@]}"; do
            echo -e "  ${GREEN}✓${NC} $service"
        done
    fi

    if [ ${#failed_services[@]} -gt 0 ]; then
        echo ""
        echo -e "${RED}Failed services (${#failed_services[@]}):${NC}"
        for service in "${failed_services[@]}"; do
            echo -e "  ${RED}✗${NC} $service"
        done
    fi

    echo ""
    echo -e "${BLUE}View logs: tail -f $LOG_DIR/<service-name>.log${NC}"
    echo -e "${BLUE}Check health: ./scripts/health-check.sh${NC}"
    echo -e "${BLUE}Stop all: ./scripts/stop-all.sh${NC}"
    echo ""

    if [ ${#failed_services[@]} -gt 0 ]; then
        exit 1
    fi
}

main "$@"

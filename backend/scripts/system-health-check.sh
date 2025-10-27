#!/bin/bash

# System Health Check Script
# Comprehensive health check for all 33 services + infrastructure
# Author: Claude (Assistant)
# Date: 2025-10-26

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Results arrays
declare -a FAILED_ITEMS
declare -a WARNING_ITEMS

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED_CHECKS++))
    ((TOTAL_CHECKS++))
}

print_error() {
    echo -e "${RED}✗${NC} $1"
    FAILED_ITEMS+=("$1")
    ((FAILED_CHECKS++))
    ((TOTAL_CHECKS++))
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    WARNING_ITEMS+=("$1")
    ((WARNING_CHECKS++))
    ((TOTAL_CHECKS++))
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# ============================================
# 1. Prerequisites Check
# ============================================
print_header "1. Prerequisites Check"

if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go installed: $GO_VERSION"
else
    print_error "Go not installed"
fi

if command_exists psql; then
    PSQL_VERSION=$(psql --version | awk '{print $3}')
    print_success "PostgreSQL client installed: $PSQL_VERSION"
else
    print_error "PostgreSQL client not installed"
fi

if command_exists redis-cli; then
    REDIS_VERSION=$(redis-cli --version | awk '{print $2}')
    print_success "Redis client installed: $REDIS_VERSION"
else
    print_error "Redis client not installed"
fi

if command_exists docker; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    print_success "Docker installed: $DOCKER_VERSION"
else
    print_warning "Docker not installed (optional for local dev)"
fi

# ============================================
# 2. Database Connectivity Check
# ============================================
print_header "2. Database Connectivity Check"

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-40432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}

# Test PostgreSQL connection
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" >/dev/null 2>&1; then
    print_success "PostgreSQL connection: $DB_HOST:$DB_PORT"

    # Check database count
    DB_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT COUNT(*) FROM pg_database WHERE datname LIKE 'payment_%';")
    if [ "$DB_COUNT" -ge 12 ]; then
        print_success "Database count: $DB_COUNT databases found"
    else
        print_warning "Database count: Only $DB_COUNT databases found (expected 12+)"
    fi
else
    print_error "PostgreSQL connection failed: $DB_HOST:$DB_PORT"
fi

# ============================================
# 3. Redis Connectivity Check
# ============================================
print_header "3. Redis Connectivity Check"

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-40379}

if redis-cli -h $REDIS_HOST -p $REDIS_PORT PING >/dev/null 2>&1; then
    print_success "Redis connection: $REDIS_HOST:$REDIS_PORT"

    # Check Redis version
    REDIS_INFO=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT INFO server | grep redis_version | cut -d: -f2 | tr -d '\r')
    print_info "Redis version: $REDIS_INFO"
else
    print_error "Redis connection failed: $REDIS_HOST:$REDIS_PORT"
fi

# ============================================
# 4. Go Workspace Verification
# ============================================
print_header "4. Go Workspace Verification"

if [ -f "go.work" ]; then
    print_success "go.work file exists"

    # Count services in workspace
    SERVICE_COUNT=$(grep -c "use ./services" go.work || echo "0")
    print_info "Services in workspace: $SERVICE_COUNT"

    # Verify workspace is valid
    if GOWORK=$(pwd)/go.work go work sync >/dev/null 2>&1; then
        print_success "go.work is valid and synced"
    else
        print_warning "go.work sync had warnings"
    fi
else
    print_error "go.work file not found"
fi

# ============================================
# 5. Service Compilation Check
# ============================================
print_header "5. Service Compilation Check (19 Microservices)"

SERVICES=(
    "admin-bff-service:40001"
    "merchant-bff-service:40023"
    "payment-gateway:40003"
    "order-service:40004"
    "channel-adapter:40005"
    "risk-service:40006"
    "accounting-service:40007"
    "notification-service:40008"
    "analytics-service:40009"
    "config-service:40010"
    "merchant-auth-service:40011"
    "merchant-policy-service:40012"
    "settlement-service:40013"
    "withdrawal-service:40014"
    "kyc-service:40015"
    "cashier-service:40016"
    "reconciliation-service:40020"
    "dispute-service:40021"
    "merchant-quota-service:40022"
)

COMPILED_COUNT=0
FAILED_SERVICES=()

for service_info in "${SERVICES[@]}"; do
    service_name=$(echo $service_info | cut -d: -f1)
    service_port=$(echo $service_info | cut -d: -f2)

    if [ -d "services/$service_name" ]; then
        # Try to compile
        if GOWORK=$(pwd)/go.work timeout 30 go build -o /tmp/${service_name}-test services/$service_name/cmd/main.go 2>/dev/null; then
            binary_size=$(ls -lh /tmp/${service_name}-test 2>/dev/null | awk '{print $5}')
            print_success "$service_name (port $service_port) - compiled ($binary_size)"
            ((COMPILED_COUNT++))
            rm -f /tmp/${service_name}-test
        else
            print_error "$service_name (port $service_port) - compilation failed"
            FAILED_SERVICES+=("$service_name")
        fi
    else
        print_error "$service_name - directory not found"
    fi
done

print_info "Compilation summary: $COMPILED_COUNT/${#SERVICES[@]} services compiled successfully"

# ============================================
# 6. Shared Package (pkg/) Check
# ============================================
print_header "6. Shared Package (pkg/) Check"

PKG_MODULES=(
    "app"
    "auth"
    "cache"
    "config"
    "crypto"
    "currency"
    "db"
    "email"
    "grpc"
    "health"
    "httpclient"
    "kafka"
    "logger"
    "metrics"
    "middleware"
    "migration"
    "retry"
    "tracing"
    "validator"
)

PKG_COUNT=0

for module in "${PKG_MODULES[@]}"; do
    if [ -d "pkg/$module" ]; then
        # Check if has .go files
        if ls pkg/$module/*.go >/dev/null 2>&1; then
            file_count=$(ls pkg/$module/*.go 2>/dev/null | wc -l)
            print_success "pkg/$module - $file_count files"
            ((PKG_COUNT++))
        else
            print_warning "pkg/$module - no .go files found"
        fi
    else
        print_error "pkg/$module - directory not found"
    fi
done

print_info "Shared packages: $PKG_COUNT/${#PKG_MODULES[@]} modules found"

# ============================================
# 7. Docker Infrastructure Check (Optional)
# ============================================
print_header "7. Docker Infrastructure Check (Optional)"

if command_exists docker; then
    # Check if docker-compose is running
    if docker ps >/dev/null 2>&1; then
        RUNNING_CONTAINERS=$(docker ps --format "{{.Names}}" | wc -l)
        print_info "Running containers: $RUNNING_CONTAINERS"

        # Check specific services
        if docker ps | grep -q postgres; then
            print_success "PostgreSQL container running"
        else
            print_warning "PostgreSQL container not running (may be using local install)"
        fi

        if docker ps | grep -q redis; then
            print_success "Redis container running"
        else
            print_warning "Redis container not running (may be using local install)"
        fi

        if docker ps | grep -q kafka; then
            print_success "Kafka container running"
        else
            print_warning "Kafka container not running (optional)"
        fi
    else
        print_warning "Docker daemon not running or not accessible"
    fi
else
    print_info "Docker not installed - skipping container checks"
fi

# ============================================
# 8. Configuration Files Check
# ============================================
print_header "8. Configuration Files Check"

CONFIG_FILES=(
    ".env.example"
    "Makefile"
    "docker-compose.yml"
    "scripts/init-db.sh"
    "scripts/start-all-services.sh"
    "scripts/optimize-database-indexes.sql"
)

for config_file in "${CONFIG_FILES[@]}"; do
    if [ -f "$config_file" ]; then
        print_success "$config_file exists"
    else
        print_warning "$config_file not found"
    fi
done

# ============================================
# 9. Frontend Applications Check
# ============================================
print_header "9. Frontend Applications Check (Optional)"

FRONTENDS=(
    "../frontend/admin-portal"
    "../frontend/merchant-portal"
    "../frontend/website"
)

for frontend in "${FRONTENDS[@]}"; do
    frontend_name=$(basename $frontend)
    if [ -d "$frontend" ]; then
        if [ -f "$frontend/package.json" ]; then
            print_success "$frontend_name - package.json exists"
        else
            print_warning "$frontend_name - package.json not found"
        fi
    else
        print_info "$frontend_name - directory not found (optional)"
    fi
done

# ============================================
# 10. Documentation Check
# ============================================
print_header "10. Documentation Check"

DOCS=(
    "README.md"
    "CLAUDE.md"
    "API_DOCUMENTATION_GUIDE.md"
    "BOOTSTRAP_MIGRATION_FINAL_100PERCENT.md"
    "EXECUTION_PLAN_PROGRESS.md"
    "docs/REDIS_CACHING_OPTIMIZATION_GUIDE.md"
    "docs/CONNECTION_POOLING_OPTIMIZATION_GUIDE.md"
)

DOC_COUNT=0

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        print_success "$doc exists"
        ((DOC_COUNT++))
    else
        print_warning "$doc not found"
    fi
done

# ============================================
# Final Summary
# ============================================
print_header "Health Check Summary"

echo "Total checks: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $PASSED_CHECKS${NC}"
echo -e "${YELLOW}Warnings: $WARNING_CHECKS${NC}"
echo -e "${RED}Failed: $FAILED_CHECKS${NC}"

# Calculate success rate
if [ $TOTAL_CHECKS -gt 0 ]; then
    SUCCESS_RATE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
    echo ""
    echo "Success rate: ${SUCCESS_RATE}%"
fi

# Print failed items
if [ ${#FAILED_ITEMS[@]} -gt 0 ]; then
    echo ""
    echo -e "${RED}Failed items:${NC}"
    for item in "${FAILED_ITEMS[@]}"; do
        echo -e "  ${RED}✗${NC} $item"
    done
fi

# Print warning items
if [ ${#WARNING_ITEMS[@]} -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}Warning items:${NC}"
    for item in "${WARNING_ITEMS[@]}"; do
        echo -e "  ${YELLOW}⚠${NC} $item"
    done
fi

# Exit code
if [ $FAILED_CHECKS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ All critical checks passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Some checks failed. Please review the errors above.${NC}"
    exit 1
fi

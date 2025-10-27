#!/bin/bash

# Quick Service Verification Script
# Verifies compilation and basic health of all 19 microservices
# Author: Claude (Assistant)
# Date: 2025-10-26

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TOTAL_SERVICES=0
COMPILED_SERVICES=0
FAILED_SERVICES=0

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Service list (name:port)
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

print_header "Global Payment Platform - Service Verification"

print_info "Verifying 19 microservices..."
echo ""

# Compilation check
print_header "1. Compilation Check"

for service_info in "${SERVICES[@]}"; do
    service_name=$(echo $service_info | cut -d: -f1)
    service_port=$(echo $service_info | cut -d: -f2)

    ((TOTAL_SERVICES++))

    if [ -d "services/$service_name" ]; then
        # Try to compile
        if GOWORK=$(pwd)/go.work timeout 30 go build -o /tmp/${service_name}-test services/$service_name/cmd/main.go 2>/dev/null; then
            binary_size=$(ls -lh /tmp/${service_name}-test 2>/dev/null | awk '{print $5}')
            print_success "$service_name (port $service_port) - $binary_size"
            ((COMPILED_SERVICES++))
            rm -f /tmp/${service_name}-test
        else
            print_error "$service_name (port $service_port) - compilation failed"
            ((FAILED_SERVICES++))
        fi
    else
        print_error "$service_name - directory not found"
        ((FAILED_SERVICES++))
    fi
done

# Summary
echo ""
print_header "Verification Summary"

echo "Total services: $TOTAL_SERVICES"
echo -e "${GREEN}Compiled: $COMPILED_SERVICES${NC}"
echo -e "${RED}Failed: $FAILED_SERVICES${NC}"

if [ $TOTAL_SERVICES -gt 0 ]; then
    SUCCESS_RATE=$((COMPILED_SERVICES * 100 / TOTAL_SERVICES))
    echo ""
    echo "Success rate: ${SUCCESS_RATE}%"
fi

# Status
echo ""
if [ $FAILED_SERVICES -eq 0 ]; then
    print_success "All services compiled successfully! ✓"
    echo ""
    print_info "Next steps:"
    echo "1. Apply performance optimizations: ./scripts/apply-performance-optimizations.sh"
    echo "2. Run system health check: ./scripts/system-health-check.sh"
    echo "3. Start all services: make run-all"
    echo "4. Deploy to production (follow NEXT_STEPS_GUIDE.md)"
    exit 0
else
    print_error "Some services failed to compile"
    echo ""
    print_info "Troubleshooting:"
    echo "1. Check go.work file is present"
    echo "2. Run: go mod tidy in each failed service"
    echo "3. Check for missing dependencies"
    echo "4. Review batch reports for service-specific fixes"
    exit 1
fi

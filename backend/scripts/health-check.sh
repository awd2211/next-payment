#!/bin/bash

# ============================================
# Payment Platform - Health Check
# ============================================
# This script checks the health status of all services

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Service definitions: name|port
SERVICES=(
    "admin-service|8001"
    "merchant-service|8002"
    "payment-gateway|8003"
    "order-service|8004"
    "channel-adapter|8005"
    "risk-service|8006"
    "notification-service|8007"
    "accounting-service|8008"
    "analytics-service|8009"
    "config-service|8010"
    "settlement-service|8012"
    "withdrawal-service|8013"
    "kyc-service|8014"
    "fee-service|8015"
)

# Infrastructure
POSTGRES_HOST=${DB_HOST:-localhost}
POSTGRES_PORT=${DB_PORT:-40432}
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-40379}

# Check infrastructure
check_infrastructure() {
    echo -e "${CYAN}Infrastructure Status:${NC}"
    echo ""

    # PostgreSQL
    if pg_isready -h $POSTGRES_HOST -p $POSTGRES_PORT > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} PostgreSQL (${POSTGRES_HOST}:${POSTGRES_PORT})"
    else
        echo -e "  ${RED}✗${NC} PostgreSQL (${POSTGRES_HOST}:${POSTGRES_PORT})"
    fi

    # Redis
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} Redis (${REDIS_HOST}:${REDIS_PORT})"
    else
        echo -e "  ${RED}✗${NC} Redis (${REDIS_HOST}:${REDIS_PORT})"
    fi

    echo ""
}

# Check service health
check_service() {
    local service_name=$1
    local port=$2
    local timeout=${3:-2}

    local url="http://localhost:$port/health"
    local response=$(curl -s -m $timeout "$url" 2>/dev/null)

    if [ $? -eq 0 ] && [ -n "$response" ]; then
        # Parse response
        local status=$(echo "$response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

        if [ "$status" = "ok" ]; then
            # Get PID
            local pid=$(lsof -t -i:$port 2>/dev/null)
            if [ -n "$pid" ]; then
                echo -e "  ${GREEN}✓${NC} $service_name (port $port, PID $pid)"
            else
                echo -e "  ${GREEN}✓${NC} $service_name (port $port)"
            fi
            return 0
        else
            echo -e "  ${YELLOW}⚠${NC} $service_name (port $port) - status: $status"
            return 1
        fi
    else
        # Check if port is in use
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo -e "  ${YELLOW}⚠${NC} $service_name (port $port) - no health endpoint"
            return 1
        else
            echo -e "  ${RED}✗${NC} $service_name (port $port) - not running"
            return 2
        fi
    fi
}

# Get service metrics
get_metrics() {
    local port=$1
    local url="http://localhost:$port/metrics"

    curl -s -m 2 "$url" 2>/dev/null | grep -E '^(http_requests_total|http_request_duration)' | head -5
}

# Main execution
main() {
    local watch_mode=false
    local detailed=false
    local interval=5

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -w|--watch)
                watch_mode=true
                shift
                ;;
            -d|--detailed)
                detailed=true
                shift
                ;;
            -i|--interval)
                interval=$2
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done

    # Watch mode
    if [ "$watch_mode" = true ]; then
        while true; do
            clear
            main_check "$detailed"
            echo ""
            echo -e "${BLUE}Refreshing in ${interval}s... (Ctrl+C to exit)${NC}"
            sleep $interval
        done
    else
        main_check "$detailed"
    fi
}

main_check() {
    local detailed=$1

    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}Payment Platform - Health Check${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""

    # Check infrastructure
    check_infrastructure

    # Check services
    echo -e "${CYAN}Service Status:${NC}"
    echo ""

    local total=0
    local healthy=0
    local unhealthy=0
    local down=0

    for service_def in "${SERVICES[@]}"; do
        IFS='|' read -r service_name port <<< "$service_def"
        check_service "$service_name" "$port"
        local status=$?

        total=$((total + 1))
        if [ $status -eq 0 ]; then
            healthy=$((healthy + 1))

            # Show metrics in detailed mode
            if [ "$detailed" = true ]; then
                local metrics=$(get_metrics "$port")
                if [ -n "$metrics" ]; then
                    echo -e "${BLUE}    Metrics:${NC}"
                    echo "$metrics" | sed 's/^/      /'
                fi
            fi
        elif [ $status -eq 1 ]; then
            unhealthy=$((unhealthy + 1))
        else
            down=$((down + 1))
        fi
    done

    # Summary
    echo ""
    echo -e "${CYAN}Summary:${NC}"
    echo -e "  Total: $total services"
    echo -e "  ${GREEN}Healthy: $healthy${NC}"
    if [ $unhealthy -gt 0 ]; then
        echo -e "  ${YELLOW}Unhealthy: $unhealthy${NC}"
    fi
    if [ $down -gt 0 ]; then
        echo -e "  ${RED}Down: $down${NC}"
    fi

    # Overall status
    echo ""
    if [ $healthy -eq $total ]; then
        echo -e "${GREEN}✓ All services are healthy${NC}"
        return 0
    elif [ $down -gt 0 ]; then
        echo -e "${RED}✗ Some services are down${NC}"
        return 2
    else
        echo -e "${YELLOW}⚠ Some services are unhealthy${NC}"
        return 1
    fi
}

# Help message
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -w, --watch          Watch mode (refresh continuously)"
    echo "  -d, --detailed       Show detailed metrics"
    echo "  -i, --interval SEC   Refresh interval in watch mode (default: 5)"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                  # Check once"
    echo "  $0 -w               # Watch mode"
    echo "  $0 -w -d -i 3       # Watch with details every 3 seconds"
    exit 0
fi

main "$@"

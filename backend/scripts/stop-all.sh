#!/bin/bash

# ============================================
# Payment Platform - Stop All Services
# ============================================
# This script stops all running microservices

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Stop service by port
stop_service_by_port() {
    local service_name=$1
    local port=$2

    local pid=$(lsof -t -i:$port 2>/dev/null)

    if [ -n "$pid" ]; then
        echo -e "${BLUE}Stopping $service_name (PID: $pid)...${NC}"
        kill $pid 2>/dev/null

        # Wait for graceful shutdown
        local count=0
        while kill -0 $pid 2>/dev/null && [ $count -lt 10 ]; do
            sleep 1
            count=$((count + 1))
        done

        # Force kill if still running
        if kill -0 $pid 2>/dev/null; then
            echo -e "${YELLOW}Force killing $service_name...${NC}"
            kill -9 $pid 2>/dev/null
        fi

        if ! kill -0 $pid 2>/dev/null; then
            echo -e "${GREEN}✓ $service_name stopped${NC}"
            return 0
        else
            echo -e "${RED}✗ Failed to stop $service_name${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠ $service_name is not running${NC}"
        return 0
    fi
}

# Stop service by name
stop_service_by_name() {
    local service_name=$1

    local pids=$(ps aux | grep "$service_name" | grep -v grep | grep -v "stop-all" | awk '{print $2}')

    if [ -n "$pids" ]; then
        echo -e "${BLUE}Stopping $service_name by process name...${NC}"
        echo "$pids" | xargs kill 2>/dev/null

        # Wait for graceful shutdown
        sleep 2

        # Force kill if still running
        local remaining=$(ps aux | grep "$service_name" | grep -v grep | grep -v "stop-all" | awk '{print $2}')
        if [ -n "$remaining" ]; then
            echo -e "${YELLOW}Force killing $service_name...${NC}"
            echo "$remaining" | xargs kill -9 2>/dev/null
        fi

        echo -e "${GREEN}✓ $service_name stopped${NC}"
        return 0
    fi

    return 1
}

# Main execution
main() {
    local force=false
    local cleanup=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                force=true
                shift
                ;;
            -c|--cleanup)
                cleanup=true
                shift
                ;;
            *)
                shift
                ;;
        esac
    done

    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}Payment Platform - Stopping Services${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""

    local stopped=0
    local failed=0
    local not_running=0

    # Stop services in reverse order (to respect dependencies)
    for ((i=${#SERVICES[@]}-1; i>=0; i--)); do
        local service_def="${SERVICES[$i]}"
        IFS='|' read -r service_name port <<< "$service_def"

        # Try stopping by port first
        if stop_service_by_port "$service_name" "$port"; then
            if lsof -t -i:$port 2>/dev/null >/dev/null; then
                not_running=$((not_running + 1))
            else
                stopped=$((stopped + 1))
            fi
        else
            # Try by process name
            if stop_service_by_name "$service_name"; then
                stopped=$((stopped + 1))
            else
                # Check if it was not running
                if [ -z "$(lsof -t -i:$port 2>/dev/null)" ]; then
                    not_running=$((not_running + 1))
                else
                    failed=$((failed + 1))
                fi
            fi
        fi

        sleep 0.5
    done

    # Cleanup logs if requested
    if [ "$cleanup" = true ]; then
        echo ""
        echo -e "${BLUE}Cleaning up log files...${NC}"
        rm -f /tmp/*-service.log
        echo -e "${GREEN}✓ Logs cleaned${NC}"
    fi

    # Summary
    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}Shutdown Summary${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo ""
    echo -e "  ${GREEN}Stopped: $stopped${NC}"
    if [ $not_running -gt 0 ]; then
        echo -e "  ${YELLOW}Already stopped: $not_running${NC}"
    fi
    if [ $failed -gt 0 ]; then
        echo -e "  ${RED}Failed: $failed${NC}"
    fi

    echo ""
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}✓ All services stopped successfully${NC}"
        return 0
    else
        echo -e "${RED}✗ Some services failed to stop${NC}"
        echo -e "${YELLOW}Try running with -f flag for force shutdown${NC}"
        return 1
    fi
}

# Help message
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -f, --force     Force kill services (SIGKILL)"
    echo "  -c, --cleanup   Remove log files after stopping"
    echo "  -h, --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                # Graceful shutdown"
    echo "  $0 -f             # Force shutdown"
    echo "  $0 -c             # Shutdown and cleanup logs"
    echo "  $0 -f -c          # Force shutdown and cleanup"
    exit 0
fi

main "$@"

#!/bin/bash

# Manage Sprint 2 Services (start, stop, restart, status)
# Usage: ./manage-sprint2-services.sh {start|stop|restart|status|logs}

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
BACKEND_DIR="/home/eric/payment/backend"
LOG_DIR="$BACKEND_DIR/logs"
SERVICES_DIR="$BACKEND_DIR/services"

SERVICES=(
    "reconciliation-service:40020"
    "dispute-service:40021"
    "merchant-limit-service:40022"
)

# Environment
export GOWORK="$BACKEND_DIR/go.work"
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-40432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export STRIPE_API_KEY="${STRIPE_API_KEY:-sk_test_placeholder}"

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

# Create log directory
mkdir -p "$LOG_DIR"

# Start a single service
start_service() {
    local service_name=$1
    local port=$2

    log_info "Starting $service_name..."

    cd "$SERVICES_DIR/$service_name"

    # Check if already running
    if [ -f "$LOG_DIR/$service_name.pid" ]; then
        local pid=$(cat "$LOG_DIR/$service_name.pid")
        if ps -p $pid > /dev/null 2>&1; then
            log_warn "$service_name is already running (PID: $pid)"
            return 0
        fi
    fi

    # Start service with Air (hot reload)
    nohup air -c .air.toml > "$LOG_DIR/$service_name.log" 2>&1 &
    local pid=$!
    echo $pid > "$LOG_DIR/$service_name.pid"

    # Wait a bit for service to start
    sleep 2

    # Health check
    if curl -s http://localhost:$port/health > /dev/null 2>&1; then
        log_info "✓ $service_name started successfully (PID: $pid, Port: $port)"
        return 0
    else
        log_warn "$service_name started but health check failed (PID: $pid, Port: $port)"
        log_info "  Check logs: tail -f $LOG_DIR/$service_name.log"
        return 1
    fi
}

# Stop a single service
stop_service() {
    local service_name=$1

    log_info "Stopping $service_name..."

    if [ -f "$LOG_DIR/$service_name.pid" ]; then
        local pid=$(cat "$LOG_DIR/$service_name.pid")
        if ps -p $pid > /dev/null 2>&1; then
            kill $pid
            sleep 1
            if ps -p $pid > /dev/null 2>&1; then
                log_warn "$service_name did not stop gracefully, force killing..."
                kill -9 $pid
            fi
            rm "$LOG_DIR/$service_name.pid"
            log_info "✓ $service_name stopped (was PID: $pid)"
        else
            log_warn "$service_name is not running (stale PID file)"
            rm "$LOG_DIR/$service_name.pid"
        fi
    else
        log_warn "$service_name is not running (no PID file)"
    fi
}

# Get service status
service_status() {
    local service_name=$1
    local port=$2

    if [ -f "$LOG_DIR/$service_name.pid" ]; then
        local pid=$(cat "$LOG_DIR/$service_name.pid")
        if ps -p $pid > /dev/null 2>&1; then
            # Check health endpoint
            if curl -s http://localhost:$port/health > /dev/null 2>&1; then
                echo -e "${GREEN}✓ RUNNING${NC} (PID: $pid, Port: $port, Health: OK)"
            else
                echo -e "${YELLOW}⚠ RUNNING${NC} (PID: $pid, Port: $port, Health: FAIL)"
            fi
            return 0
        else
            echo -e "${RED}✗ STOPPED${NC} (stale PID file)"
            return 1
        fi
    else
        echo -e "${RED}✗ STOPPED${NC}"
        return 1
    fi
}

# Command handlers
cmd_start() {
    log_info "Starting all Sprint 2 services..."
    echo ""

    for SERVICE_INFO in "${SERVICES[@]}"; do
        IFS=':' read -r SERVICE PORT <<< "$SERVICE_INFO"
        start_service "$SERVICE" "$PORT"
        echo ""
    done

    log_info "All services start command completed!"
    log_info "Run './manage-sprint2-services.sh status' to check status"
}

cmd_stop() {
    log_info "Stopping all Sprint 2 services..."
    echo ""

    for SERVICE_INFO in "${SERVICES[@]}"; do
        IFS=':' read -r SERVICE PORT <<< "$SERVICE_INFO"
        stop_service "$SERVICE"
    done

    echo ""
    log_info "All services stopped!"
}

cmd_restart() {
    log_info "Restarting all Sprint 2 services..."
    echo ""
    cmd_stop
    echo ""
    sleep 2
    cmd_start
}

cmd_status() {
    echo "=========================================="
    echo "Sprint 2 Services Status"
    echo "=========================================="
    echo ""

    for SERVICE_INFO in "${SERVICES[@]}"; do
        IFS=':' read -r SERVICE PORT <<< "$SERVICE_INFO"
        printf "%-30s: " "$SERVICE"
        service_status "$SERVICE" "$PORT"
    done

    echo ""
    echo "To view logs: ./manage-sprint2-services.sh logs <service-name>"
    echo "Available services: reconciliation-service, dispute-service, merchant-limit-service"
}

cmd_logs() {
    local service_name=$1

    if [ -z "$service_name" ]; then
        log_error "Please specify a service name"
        echo "Usage: $0 logs <service-name>"
        echo "Available services: reconciliation-service, dispute-service, merchant-limit-service"
        exit 1
    fi

    local log_file="$LOG_DIR/$service_name.log"

    if [ -f "$log_file" ]; then
        log_info "Showing logs for $service_name (Ctrl+C to exit)"
        tail -f "$log_file"
    else
        log_error "Log file not found: $log_file"
        exit 1
    fi
}

# Main command handler
case "${1:-}" in
    start)
        cmd_start
        ;;
    stop)
        cmd_stop
        ;;
    restart)
        cmd_restart
        ;;
    status)
        cmd_status
        ;;
    logs)
        cmd_logs "${2:-}"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs}"
        echo ""
        echo "Commands:"
        echo "  start    - Start all Sprint 2 services"
        echo "  stop     - Stop all Sprint 2 services"
        echo "  restart  - Restart all Sprint 2 services"
        echo "  status   - Show status of all services"
        echo "  logs     - Tail logs for a specific service"
        echo ""
        echo "Examples:"
        echo "  $0 start"
        echo "  $0 status"
        echo "  $0 logs reconciliation-service"
        exit 1
        ;;
esac

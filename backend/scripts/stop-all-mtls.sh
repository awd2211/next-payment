#!/bin/bash
# 停止所有服务
# 用途: 停止所有后端微服务 + Kong + 基础设施

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_DIR="$SCRIPT_DIR/../logs"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

echo "=========================================="
echo "  停止所有服务"
echo "=========================================="
echo ""

# ============================================
# 1. 停止后端服务
# ============================================
log_info "[1/2] 停止后端微服务..."

if [ -d "$LOG_DIR" ]; then
    for pidfile in "$LOG_DIR"/*.pid; do
        if [ -f "$pidfile" ]; then
            SERVICE_NAME=$(basename "$pidfile" .pid)
            PID=$(cat "$pidfile")

            if ps -p $PID > /dev/null 2>&1; then
                log_info "停止 $SERVICE_NAME (PID: $PID)..."
                kill $PID
                sleep 1

                # 如果进程仍在运行，强制杀死
                if ps -p $PID > /dev/null 2>&1; then
                    log_info "强制停止 $SERVICE_NAME..."
                    kill -9 $PID
                fi

                log_success "$SERVICE_NAME 已停止"
            else
                log_info "$SERVICE_NAME 已经停止"
            fi

            rm -f "$pidfile"
        fi
    done
else
    log_info "没有运行中的后端服务"
fi

echo ""

# ============================================
# 2. 停止基础设施
# ============================================
log_info "[2/2] 停止基础设施 (Kong, Kafka, Redis, PostgreSQL)..."

cd "$PROJECT_ROOT"

if docker compose version &> /dev/null; then
    docker compose down
    log_success "基础设施已停止"
else
    log_error "docker compose 未安装，无法停止基础设施"
fi

echo ""
echo "=========================================="
echo "  所有服务已停止"
echo "=========================================="
echo ""

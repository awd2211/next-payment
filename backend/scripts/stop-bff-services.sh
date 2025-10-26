#!/bin/bash

# ============================================================================
# BFF Services Stop Script
# 停止 Admin BFF 和 Merchant BFF 两个服务
# ============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
LOG_DIR="$PROJECT_ROOT/logs/bff"

echo ""
log_info "========================================"
log_info "Stopping BFF Services"
log_info "========================================"
echo ""

# 停止 Admin BFF
if [ -f "$LOG_DIR/admin-bff.pid" ]; then
    ADMIN_PID=$(cat "$LOG_DIR/admin-bff.pid")
    if ps -p $ADMIN_PID > /dev/null; then
        log_info "停止 Admin BFF Service (PID: $ADMIN_PID)..."
        kill $ADMIN_PID
        sleep 2

        # 强制杀死
        if ps -p $ADMIN_PID > /dev/null; then
            log_warning "强制停止 Admin BFF Service..."
            kill -9 $ADMIN_PID
        fi

        rm -f "$LOG_DIR/admin-bff.pid"
        log_success "Admin BFF Service 已停止"
    else
        log_warning "Admin BFF Service 未运行"
        rm -f "$LOG_DIR/admin-bff.pid"
    fi
else
    log_warning "未找到 Admin BFF PID 文件"
fi

# 停止 Merchant BFF
if [ -f "$LOG_DIR/merchant-bff.pid" ]; then
    MERCHANT_PID=$(cat "$LOG_DIR/merchant-bff.pid")
    if ps -p $MERCHANT_PID > /dev/null; then
        log_info "停止 Merchant BFF Service (PID: $MERCHANT_PID)..."
        kill $MERCHANT_PID
        sleep 2

        # 强制杀死
        if ps -p $MERCHANT_PID > /dev/null; then
            log_warning "强制停止 Merchant BFF Service..."
            kill -9 $MERCHANT_PID
        fi

        rm -f "$LOG_DIR/merchant-bff.pid"
        log_success "Merchant BFF Service 已停止"
    else
        log_warning "Merchant BFF Service 未运行"
        rm -f "$LOG_DIR/merchant-bff.pid"
    fi
else
    log_warning "未找到 Merchant BFF PID 文件"
fi

# 清理残留进程
log_info "检查残留进程..."
if pgrep -f "admin-bff-service" > /dev/null; then
    log_warning "发现 Admin BFF 残留进程，清理中..."
    pkill -9 -f "admin-bff-service"
fi

if pgrep -f "merchant-bff-service" > /dev/null; then
    log_warning "发现 Merchant BFF 残留进程，清理中..."
    pkill -9 -f "merchant-bff-service"
fi

echo ""
log_success "所有 BFF 服务已停止"
echo ""

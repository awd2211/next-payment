#!/bin/bash

# ============================================================================
# BFF Services Startup Script
# 启动 Admin BFF 和 Merchant BFF 两个服务
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
cd "$PROJECT_ROOT"

# 日志目录
LOG_DIR="$PROJECT_ROOT/logs/bff"
mkdir -p "$LOG_DIR"

# 检查环境变量
check_env() {
    log_info "检查环境变量..."

    # 必需的环境变量
    REQUIRED_VARS=(
        "JWT_SECRET"
    )

    # 可选的环境变量（使用默认值）
    export PORT_ADMIN_BFF="${PORT_ADMIN_BFF:-40001}"
    export PORT_MERCHANT_BFF="${PORT_MERCHANT_BFF:-40023}"
    export DB_HOST="${DB_HOST:-localhost}"
    export DB_PORT="${DB_PORT:-40432}"
    export DB_USER="${DB_USER:-postgres}"
    export DB_PASSWORD="${DB_PASSWORD:-postgres}"
    export REDIS_HOST="${REDIS_HOST:-localhost}"
    export REDIS_PORT="${REDIS_PORT:-40379}"
    export JAEGER_ENDPOINT="${JAEGER_ENDPOINT:-http://localhost:14268/api/traces}"
    export JAEGER_SAMPLING_RATE="${JAEGER_SAMPLING_RATE:-10}"
    export LOG_LEVEL="${LOG_LEVEL:-info}"
    export ENV="${ENV:-development}"

    # 检查必需变量
    for var in "${REQUIRED_VARS[@]}"; do
        if [ -z "${!var}" ]; then
            log_error "环境变量 $var 未设置"
            log_info "请在 .env 文件中设置或执行: export $var=your_value"
            exit 1
        fi
    done

    log_success "环境变量检查通过"
}

# 检查依赖服务
check_dependencies() {
    log_info "检查依赖服务..."

    # 检查 PostgreSQL（仅 Admin BFF 需要）
    if ! nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; then
        log_warning "PostgreSQL 未运行 ($DB_HOST:$DB_PORT)"
        log_warning "Admin BFF 需要数据库支持（审计日志）"
        read -p "是否继续？(y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        log_success "PostgreSQL 已运行"
    fi

    # 检查 Redis（可选）
    if ! nc -z "$REDIS_HOST" "$REDIS_PORT" 2>/dev/null; then
        log_warning "Redis 未运行 ($REDIS_HOST:$REDIS_PORT)"
        log_warning "速率限制将使用内存存储"
    else
        log_success "Redis 已运行"
    fi
}

# 编译服务
build_services() {
    log_info "编译 BFF 服务..."

    # 编译 Admin BFF
    log_info "编译 Admin BFF Service..."
    cd "$PROJECT_ROOT/services/admin-bff-service"
    GOWORK="$PROJECT_ROOT/go.work" go build -o "$PROJECT_ROOT/bin/admin-bff-service" ./cmd/main.go
    if [ $? -eq 0 ]; then
        log_success "Admin BFF Service 编译成功"
    else
        log_error "Admin BFF Service 编译失败"
        exit 1
    fi

    # 编译 Merchant BFF
    log_info "编译 Merchant BFF Service..."
    cd "$PROJECT_ROOT/services/merchant-bff-service"
    GOWORK="$PROJECT_ROOT/go.work" go build -o "$PROJECT_ROOT/bin/merchant-bff-service" ./cmd/main.go
    if [ $? -eq 0 ]; then
        log_success "Merchant BFF Service 编译成功"
    else
        log_error "Merchant BFF Service 编译失败"
        exit 1
    fi

    cd "$PROJECT_ROOT"
}

# 停止已有服务
stop_existing_services() {
    log_info "停止已有的 BFF 服务..."

    # 停止 Admin BFF
    if pgrep -f "admin-bff-service" > /dev/null; then
        pkill -f "admin-bff-service"
        log_success "已停止 Admin BFF Service"
        sleep 2
    fi

    # 停止 Merchant BFF
    if pgrep -f "merchant-bff-service" > /dev/null; then
        pkill -f "merchant-bff-service"
        log_success "已停止 Merchant BFF Service"
        sleep 2
    fi
}

# 启动 Admin BFF
start_admin_bff() {
    log_info "启动 Admin BFF Service (port $PORT_ADMIN_BFF)..."

    # 设置环境变量
    export PORT="$PORT_ADMIN_BFF"
    export DB_NAME="payment_admin"
    export SERVICE_NAME="admin-bff-service"

    # 启动服务
    nohup "$PROJECT_ROOT/bin/admin-bff-service" \
        > "$LOG_DIR/admin-bff.log" 2>&1 &

    ADMIN_PID=$!
    echo $ADMIN_PID > "$LOG_DIR/admin-bff.pid"

    # 等待服务启动
    sleep 3

    # 检查服务状态
    if ps -p $ADMIN_PID > /dev/null; then
        log_success "Admin BFF Service 已启动 (PID: $ADMIN_PID)"
        log_info "日志文件: $LOG_DIR/admin-bff.log"
        log_info "Swagger UI: http://localhost:$PORT_ADMIN_BFF/swagger/index.html"
        log_info "Health Check: http://localhost:$PORT_ADMIN_BFF/health"
        log_info "Metrics: http://localhost:$PORT_ADMIN_BFF/metrics"
    else
        log_error "Admin BFF Service 启动失败"
        log_info "查看日志: tail -f $LOG_DIR/admin-bff.log"
        exit 1
    fi
}

# 启动 Merchant BFF
start_merchant_bff() {
    log_info "启动 Merchant BFF Service (port $PORT_MERCHANT_BFF)..."

    # 设置环境变量
    export PORT="$PORT_MERCHANT_BFF"
    export SERVICE_NAME="merchant-bff-service"

    # 启动服务
    nohup "$PROJECT_ROOT/bin/merchant-bff-service" \
        > "$LOG_DIR/merchant-bff.log" 2>&1 &

    MERCHANT_PID=$!
    echo $MERCHANT_PID > "$LOG_DIR/merchant-bff.pid"

    # 等待服务启动
    sleep 3

    # 检查服务状态
    if ps -p $MERCHANT_PID > /dev/null; then
        log_success "Merchant BFF Service 已启动 (PID: $MERCHANT_PID)"
        log_info "日志文件: $LOG_DIR/merchant-bff.log"
        log_info "Swagger UI: http://localhost:$PORT_MERCHANT_BFF/swagger/index.html"
        log_info "Health Check: http://localhost:$PORT_MERCHANT_BFF/health"
        log_info "Metrics: http://localhost:$PORT_MERCHANT_BFF/metrics"
    else
        log_error "Merchant BFF Service 启动失败"
        log_info "查看日志: tail -f $LOG_DIR/merchant-bff.log"
        exit 1
    fi
}

# 显示服务状态
show_status() {
    echo ""
    log_info "========================================"
    log_info "BFF Services Status"
    log_info "========================================"

    # Admin BFF
    if [ -f "$LOG_DIR/admin-bff.pid" ]; then
        ADMIN_PID=$(cat "$LOG_DIR/admin-bff.pid")
        if ps -p $ADMIN_PID > /dev/null; then
            echo -e "${GREEN}✓${NC} Admin BFF Service    : Running (PID: $ADMIN_PID, Port: $PORT_ADMIN_BFF)"
        else
            echo -e "${RED}✗${NC} Admin BFF Service    : Stopped"
        fi
    else
        echo -e "${RED}✗${NC} Admin BFF Service    : Not started"
    fi

    # Merchant BFF
    if [ -f "$LOG_DIR/merchant-bff.pid" ]; then
        MERCHANT_PID=$(cat "$LOG_DIR/merchant-bff.pid")
        if ps -p $MERCHANT_PID > /dev/null; then
            echo -e "${GREEN}✓${NC} Merchant BFF Service : Running (PID: $MERCHANT_PID, Port: $PORT_MERCHANT_BFF)"
        else
            echo -e "${RED}✗${NC} Merchant BFF Service : Stopped"
        fi
    else
        echo -e "${RED}✗${NC} Merchant BFF Service : Not started"
    fi

    echo ""
    log_info "查看实时日志:"
    echo "  Admin BFF   : tail -f $LOG_DIR/admin-bff.log"
    echo "  Merchant BFF: tail -f $LOG_DIR/merchant-bff.log"
    echo ""
    log_info "停止服务:"
    echo "  ./scripts/stop-bff-services.sh"
    echo ""
}

# 主函数
main() {
    echo ""
    log_info "========================================"
    log_info "Starting BFF Services"
    log_info "========================================"
    echo ""

    # 检查环境
    check_env
    check_dependencies

    # 编译服务
    build_services

    # 停止已有服务
    stop_existing_services

    # 启动服务
    start_admin_bff
    start_merchant_bff

    # 显示状态
    show_status

    log_success "所有 BFF 服务已启动！"
}

# 处理 Ctrl+C
trap 'log_warning "收到中断信号，退出..."; exit 130' INT

# 执行主函数
main "$@"

#!/bin/bash
# 一键启动所有服务（mTLS 模式）
# 用途: 启动基础设施 + Kong + 所有后端微服务

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_DIR="$SCRIPT_DIR/../logs"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# 创建日志目录
mkdir -p "$LOG_DIR"

echo "=========================================="
echo "  Payment Platform - mTLS 全栈启动"
echo "=========================================="
echo ""
log_info "启动模式: mTLS (双向认证)"
log_info "日志目录: $LOG_DIR"
echo ""

# ============================================
# 1. 检查证书
# ============================================
log_info "[1/6] 检查 mTLS 证书..."

if [ ! -f "$SCRIPT_DIR/../certs/ca/ca-cert.pem" ]; then
    log_error "CA 证书不存在"
    log_info "正在生成证书..."
    cd "$SCRIPT_DIR/.."
    ./scripts/generate-mtls-certs.sh
    log_success "证书生成完成"
fi

if [ ! -f "$SCRIPT_DIR/../certs/services/kong-gateway/cert.pem" ]; then
    log_warning "Kong 证书不存在"
    log_info "正在生成 Kong 证书..."
    cd "$SCRIPT_DIR/.."
    ./scripts/setup-kong-mtls-cert.sh
    log_success "Kong 证书生成完成"
fi

log_success "证书检查完成"
echo ""

# ============================================
# 2. 启动基础设施
# ============================================
log_info "[2/6] 启动基础设施 (PostgreSQL, Redis, Kafka, Kong)..."

cd "$PROJECT_ROOT"

# 检查 docker compose 是否可用
if ! docker compose version &> /dev/null; then
    log_error "docker compose 未安装"
    exit 1
fi

# 启动基础设施
docker compose up -d postgres redis zookeeper kafka kong-database kong-bootstrap kong konga

log_info "等待基础设施就绪..."
sleep 10

# 等待 PostgreSQL
log_info "等待 PostgreSQL..."
for i in {1..30}; do
    if docker exec payment-postgres pg_isready -U postgres > /dev/null 2>&1; then
        log_success "PostgreSQL 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "PostgreSQL 启动超时"
        exit 1
    fi
    sleep 2
done

# 等待 Kong
log_info "等待 Kong API Gateway..."
for i in {1..30}; do
    if curl -s -f http://localhost:40081/ > /dev/null 2>&1; then
        log_success "Kong 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "Kong 启动超时"
        exit 1
    fi
    sleep 2
done

log_success "基础设施启动完成"
echo ""

# ============================================
# 3. 配置 Kong (mTLS 模式)
# ============================================
log_info "[3/6] 配置 Kong API Gateway (mTLS 模式)..."

cd "$SCRIPT_DIR/.."
ENABLE_MTLS=true ./scripts/kong-setup.sh

log_success "Kong 配置完成"
echo ""

# ============================================
# 4. 启动后端服务
# ============================================
log_info "[4/6] 启动后端微服务 (mTLS 模式)..."

# 服务列表（按依赖顺序）
SERVICES=(
    "config-service:40010"
    "admin-service:40001"
    "merchant-auth-service:40011"
    "merchant-service:40002"
    "risk-service:40006"
    "channel-adapter:40005"
    "order-service:40004"
    "payment-gateway:40003"
    "accounting-service:40007"
    "analytics-service:40009"
    "notification-service:40008"
    "settlement-service:40013"
    "withdrawal-service:40014"
    "kyc-service:40015"
    "cashier-service:40016"
)

# 导出 mTLS 环境变量
export ENABLE_MTLS=true
export TLS_CA_FILE="$(pwd)/certs/ca/ca-cert.pem"

# 启动所有服务
for service_info in "${SERVICES[@]}"; do
    SERVICE_NAME="${service_info%:*}"
    PORT="${service_info#*:}"

    # 检查端口是否已被占用
    if lsof -i :$PORT > /dev/null 2>&1; then
        log_warning "$SERVICE_NAME 端口 $PORT 已被占用，跳过启动"
        continue
    fi

    log_info "启动 $SERVICE_NAME (端口 $PORT)..."

    # 设置服务专用证书
    export TLS_CERT_FILE="$(pwd)/certs/services/$SERVICE_NAME/cert.pem"
    export TLS_KEY_FILE="$(pwd)/certs/services/$SERVICE_NAME/key.pem"

    # 如果是 payment-gateway，设置客户端证书
    if [ "$SERVICE_NAME" == "payment-gateway" ]; then
        export TLS_CLIENT_CERT="$TLS_CERT_FILE"
        export TLS_CLIENT_KEY="$TLS_KEY_FILE"
        export ORDER_SERVICE_URL="https://localhost:40004"
        export RISK_SERVICE_URL="https://localhost:40006"
        export CHANNEL_SERVICE_URL="https://localhost:40005"
        export MERCHANT_AUTH_SERVICE_URL="https://localhost:40011"
        export NOTIFICATION_SERVICE_URL="https://localhost:40008"
        export ANALYTICS_SERVICE_URL="https://localhost:40009"
    fi

    # 启动服务（后台运行）
    cd "services/$SERVICE_NAME"
    nohup go run cmd/main.go > "$LOG_DIR/$SERVICE_NAME.log" 2>&1 &
    echo $! > "$LOG_DIR/$SERVICE_NAME.pid"
    cd ../..

    # 等待服务启动
    sleep 3

    # 验证服务是否启动成功
    if lsof -i :$PORT > /dev/null 2>&1; then
        log_success "$SERVICE_NAME 启动成功 (PID: $(cat $LOG_DIR/$SERVICE_NAME.pid))"
    else
        log_error "$SERVICE_NAME 启动失败，查看日志: $LOG_DIR/$SERVICE_NAME.log"
    fi
done

echo ""
log_success "所有后端服务启动完成"
echo ""

# ============================================
# 5. 验证配置
# ============================================
log_info "[5/6] 验证 mTLS 配置..."

./scripts/verify-kong-mtls.sh > /dev/null 2>&1 && log_success "Kong mTLS 配置正常" || log_warning "Kong mTLS 配置可能有问题"

echo ""

# ============================================
# 6. 显示状态
# ============================================
log_info "[6/6] 服务状态总览..."
echo ""

echo "=========================================="
echo "  基础设施"
echo "=========================================="
echo "  PostgreSQL:  localhost:40432"
echo "  Redis:       localhost:40379"
echo "  Kafka:       localhost:40092"
echo "  Kong Proxy:  http://localhost:40080"
echo "  Kong Admin:  http://localhost:40081"
echo "  Konga UI:    http://localhost:40082"
echo ""

echo "=========================================="
echo "  后端服务 (mTLS 模式)"
echo "=========================================="
for service_info in "${SERVICES[@]}"; do
    SERVICE_NAME="${service_info%:*}"
    PORT="${service_info#*:}"

    if [ -f "$LOG_DIR/$SERVICE_NAME.pid" ]; then
        PID=$(cat "$LOG_DIR/$SERVICE_NAME.pid")
        if ps -p $PID > /dev/null 2>&1; then
            echo "  ✅ $SERVICE_NAME (PID: $PID, Port: $PORT)"
        else
            echo "  ❌ $SERVICE_NAME (已停止)"
        fi
    else
        echo "  ⊙ $SERVICE_NAME (未启动)"
    fi
done

echo ""
echo "=========================================="
echo "  日志文件"
echo "=========================================="
echo "  日志目录: $LOG_DIR"
echo "  查看日志: tail -f $LOG_DIR/<service-name>.log"
echo ""

echo "=========================================="
echo "  测试命令"
echo "=========================================="
echo "  # 测试 Kong → Order Service (mTLS)"
echo "  curl http://localhost:40080/api/v1/orders"
echo ""
echo "  # 直接测试 Order Service (mTLS，需要证书)"
echo "  curl -v https://localhost:40004/health \\"
echo "    --cacert certs/ca/ca-cert.pem \\"
echo "    --cert certs/services/payment-gateway/cert.pem \\"
echo "    --key certs/services/payment-gateway/key.pem"
echo ""

echo "=========================================="
echo "  管理命令"
echo "=========================================="
echo "  停止所有服务:  ./scripts/stop-all-mtls.sh"
echo "  查看服务状态:  ./scripts/status-all-mtls.sh"
echo "  重启服务:      ./scripts/restart-all-mtls.sh"
echo ""

log_success "mTLS 全栈启动完成！"

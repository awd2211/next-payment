#!/bin/bash

# 启动所有微服务的脚本（mTLS 模式）
# 使用air进行热重载

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}启动所有支付平台微服务 (mTLS)${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 切换到 backend 目录
cd /home/eric/payment/backend

# ==========================================
# 前置检查
# ==========================================
echo -e "${BLUE}[1/5] 前置检查...${NC}"

# 检查 Docker 基础设施
if ! docker ps | grep -q "payment-postgres"; then
    echo -e "${RED}✗ PostgreSQL 容器未运行${NC}"
    echo -e "${YELLOW}请先运行: docker-compose up -d${NC}"
    exit 1
fi

if ! docker ps | grep -q "payment-redis"; then
    echo -e "${RED}✗ Redis 容器未运行${NC}"
    echo -e "${YELLOW}请先运行: docker-compose up -d${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker 基础设施运行正常${NC}"

# 检查 mTLS 证书
if [ ! -f "certs/ca/ca-cert.pem" ]; then
    echo -e "${RED}✗ CA 证书不存在${NC}"
    echo -e "${YELLOW}请先运行: ./scripts/generate-mtls-certs.sh${NC}"
    exit 1
fi

echo -e "${GREEN}✓ mTLS 证书存在${NC}"

# 检查 air
if ! command -v ~/go/bin/air &> /dev/null; then
    echo -e "${RED}✗ air 未安装${NC}"
    echo -e "${YELLOW}请运行: go install github.com/cosmtrek/air@v1.49.0${NC}"
    exit 1
fi

echo -e "${GREEN}✓ air 已安装${NC}"
echo ""

# ==========================================
# 加载环境变量
# ==========================================
echo -e "${BLUE}[2/5] 加载环境变量...${NC}"

if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
    echo -e "${GREEN}✓ 已加载 .env 文件${NC}"
fi

# 设置工作区
export GOWORK=/home/eric/payment/backend/go.work

# ==========================================
# mTLS 配置（使用绝对路径）
# ==========================================
export ENABLE_MTLS=true
export TLS_CA_FILE="/home/eric/payment/backend/certs/ca/ca-cert.pem"

# ==========================================
# 数据库配置 (PostgreSQL Docker 端口: 40432)
# ==========================================
export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres

# ==========================================
# Redis 配置 (Docker 端口: 40379)
# ==========================================
export REDIS_HOST=localhost
export REDIS_PORT=40379
export REDIS_PASSWORD=""

# ==========================================
# Kafka 配置 (Docker 端口: 40092)
# ==========================================
export KAFKA_BROKERS=localhost:40092

echo -e "${GREEN}✓ 环境变量配置完成 (DB_PORT=40432, mTLS=enabled)${NC}"
echo ""

# 服务列表和端口（19个微服务）
declare -A SERVICES=(
    ["admin-bff-service"]=40001
    ["merchant-bff-service"]=40023
    ["payment-gateway"]=40003
    ["order-service"]=40004
    ["channel-adapter"]=40005
    ["risk-service"]=40006
    ["accounting-service"]=40007
    ["notification-service"]=40008
    ["analytics-service"]=40009
    ["config-service"]=40010
    ["merchant-auth-service"]=40011
    ["merchant-policy-service"]=40012
    ["settlement-service"]=40013
    ["withdrawal-service"]=40014
    ["kyc-service"]=40015
    ["cashier-service"]=40016
    ["reconciliation-service"]=40020
    ["dispute-service"]=40021
    ["merchant-quota-service"]=40022
)

# 日志目录
LOG_DIR="/home/eric/payment/backend/logs"
mkdir -p "$LOG_DIR"

# 检查air是否安装
if ! command -v ~/go/bin/air &> /dev/null; then
    echo -e "${RED}错误: air未安装${NC}"
    echo "请运行: go install github.com/cosmtrek/air@v1.49.0"
    exit 1
fi

# 启动函数
start_service() {
    local service=$1
    local port=$2
    local service_dir="/home/eric/payment/backend/services/$service"

    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}错误: 服务目录不存在: $service_dir${NC}"
        return 1
    fi

    # 设置服务专用的 mTLS 证书（使用绝对路径和新的命名格式）
    export TLS_CERT_FILE="/home/eric/payment/backend/certs/services/$service/$service.crt"
    export TLS_KEY_FILE="/home/eric/payment/backend/certs/services/$service/$service.key"
    export TLS_CLIENT_CERT="$TLS_CERT_FILE"
    export TLS_CLIENT_KEY="$TLS_KEY_FILE"

    # 设置数据库名称（19个服务）
    case "$service" in
        "admin-bff-service") export DB_NAME=payment_admin ;;
        "merchant-bff-service") export DB_NAME=payment_merchant ;;
        "payment-gateway") export DB_NAME=payment_gateway ;;
        "order-service") export DB_NAME=payment_order ;;
        "channel-adapter") export DB_NAME=payment_channel ;;
        "risk-service") export DB_NAME=payment_risk ;;
        "accounting-service") export DB_NAME=payment_accounting ;;
        "notification-service") export DB_NAME=payment_notification ;;
        "analytics-service") export DB_NAME=payment_analytics ;;
        "config-service") export DB_NAME=payment_config ;;
        "merchant-auth-service") export DB_NAME=payment_merchant_auth ;;
        "merchant-policy-service") export DB_NAME=payment_merchant_config ;;
        "settlement-service") export DB_NAME=payment_settlement ;;
        "withdrawal-service") export DB_NAME=payment_withdrawal ;;
        "kyc-service") export DB_NAME=payment_kyc ;;
        "cashier-service") export DB_NAME=payment_cashier ;;
        "reconciliation-service") export DB_NAME=payment_reconciliation ;;
        "dispute-service") export DB_NAME=payment_dispute ;;
        "merchant-quota-service") export DB_NAME=payment_merchant_limit ;;
    esac

    # payment-gateway 需要配置下游服务的 HTTPS 地址
    if [ "$service" == "payment-gateway" ]; then
        export ORDER_SERVICE_URL="https://localhost:40004"
        export RISK_SERVICE_URL="https://localhost:40006"
        export CHANNEL_SERVICE_URL="https://localhost:40005"
    fi

    # admin-bff-service 需要知道所有18个微服务的URL（使用HTTPS进行mTLS通信）
    if [ "$service" == "admin-bff-service" ]; then
        export CONFIG_SERVICE_URL="https://localhost:40010"
        export RISK_SERVICE_URL="https://localhost:40006"
        export KYC_SERVICE_URL="https://localhost:40015"
        export MERCHANT_SERVICE_URL="https://localhost:40002"
        export ANALYTICS_SERVICE_URL="https://localhost:40009"
        export LIMIT_SERVICE_URL="https://localhost:40022"
        export CHANNEL_SERVICE_URL="https://localhost:40005"
        export CASHIER_SERVICE_URL="https://localhost:40016"
        export ORDER_SERVICE_URL="https://localhost:40004"
        export ACCOUNTING_SERVICE_URL="https://localhost:40007"
        export DISPUTE_SERVICE_URL="https://localhost:40021"
        export MERCHANT_AUTH_SERVICE_URL="https://localhost:40011"
        export MERCHANT_CONFIG_SERVICE_URL="https://localhost:40012"
        export NOTIFICATION_SERVICE_URL="https://localhost:40008"
        export PAYMENT_SERVICE_URL="https://localhost:40003"
        export RECONCILIATION_SERVICE_URL="https://localhost:40020"
        export SETTLEMENT_SERVICE_URL="https://localhost:40013"
        export WITHDRAWAL_SERVICE_URL="https://localhost:40014"
    fi

    # merchant-bff-service 也需要知道所有15个微服务的URL（使用HTTPS进行mTLS通信）
    if [ "$service" == "merchant-bff-service" ]; then
        export PAYMENT_SERVICE_URL="https://localhost:40003"
        export ORDER_SERVICE_URL="https://localhost:40004"
        export SETTLEMENT_SERVICE_URL="https://localhost:40013"
        export WITHDRAWAL_SERVICE_URL="https://localhost:40014"
        export ACCOUNTING_SERVICE_URL="https://localhost:40007"
        export ANALYTICS_SERVICE_URL="https://localhost:40009"
        export KYC_SERVICE_URL="https://localhost:40015"
        export MERCHANT_AUTH_SERVICE_URL="https://localhost:40011"
        export MERCHANT_CONFIG_SERVICE_URL="https://localhost:40012"
        export LIMIT_SERVICE_URL="https://localhost:40022"
        export NOTIFICATION_SERVICE_URL="https://localhost:40008"
        export RISK_SERVICE_URL="https://localhost:40006"
        export DISPUTE_SERVICE_URL="https://localhost:40021"
        export RECONCILIATION_SERVICE_URL="https://localhost:40020"
        export CASHIER_SERVICE_URL="https://localhost:40016"
    fi

    echo -e "${YELLOW}启动 $service (端口: $port, DB: $DB_NAME, mTLS: $ENABLE_MTLS)...${NC}"

    cd "$service_dir"

    # 启动 air 并传递所有环境变量
    # 注意：使用 env 命令确保环境变量被正确传递
    nohup env \
        PORT=$port \
        DB_NAME=$DB_NAME \
        DB_HOST=$DB_HOST \
        DB_PORT=$DB_PORT \
        DB_USER=$DB_USER \
        DB_PASSWORD=$DB_PASSWORD \
        REDIS_HOST=$REDIS_HOST \
        REDIS_PORT=$REDIS_PORT \
        JWT_SECRET=$JWT_SECRET \
        ENABLE_MTLS=$ENABLE_MTLS \
        TLS_CERT_FILE=$TLS_CERT_FILE \
        TLS_KEY_FILE=$TLS_KEY_FILE \
        TLS_CLIENT_CERT=$TLS_CLIENT_CERT \
        TLS_CLIENT_KEY=$TLS_CLIENT_KEY \
        TLS_CA_FILE=$TLS_CA_FILE \
        ORDER_SERVICE_URL=${ORDER_SERVICE_URL:-} \
        RISK_SERVICE_URL=${RISK_SERVICE_URL:-} \
        CHANNEL_SERVICE_URL=${CHANNEL_SERVICE_URL:-} \
        CONFIG_SERVICE_URL=${CONFIG_SERVICE_URL:-} \
        KYC_SERVICE_URL=${KYC_SERVICE_URL:-} \
        MERCHANT_SERVICE_URL=${MERCHANT_SERVICE_URL:-} \
        ANALYTICS_SERVICE_URL=${ANALYTICS_SERVICE_URL:-} \
        LIMIT_SERVICE_URL=${LIMIT_SERVICE_URL:-} \
        CASHIER_SERVICE_URL=${CASHIER_SERVICE_URL:-} \
        ACCOUNTING_SERVICE_URL=${ACCOUNTING_SERVICE_URL:-} \
        DISPUTE_SERVICE_URL=${DISPUTE_SERVICE_URL:-} \
        MERCHANT_AUTH_SERVICE_URL=${MERCHANT_AUTH_SERVICE_URL:-} \
        MERCHANT_CONFIG_SERVICE_URL=${MERCHANT_CONFIG_SERVICE_URL:-} \
        NOTIFICATION_SERVICE_URL=${NOTIFICATION_SERVICE_URL:-} \
        PAYMENT_SERVICE_URL=${PAYMENT_SERVICE_URL:-} \
        RECONCILIATION_SERVICE_URL=${RECONCILIATION_SERVICE_URL:-} \
        SETTLEMENT_SERVICE_URL=${SETTLEMENT_SERVICE_URL:-} \
        WITHDRAWAL_SERVICE_URL=${WITHDRAWAL_SERVICE_URL:-} \
        ~/go/bin/air -c .air.toml > "$LOG_DIR/$service.log" 2>&1 &

    local pid=$!
    echo $pid > "$LOG_DIR/$service.pid"

    echo -e "${GREEN}✓ $service 已启动 (PID: $pid)${NC}"
}

# ==========================================
# 停止已运行的服务
# ==========================================
echo -e "${BLUE}[3/5] 停止已运行的服务...${NC}"

stopped_count=0
for service in "${!SERVICES[@]}"; do
    if [ -f "$LOG_DIR/$service.pid" ]; then
        old_pid=$(cat "$LOG_DIR/$service.pid")
        if ps -p $old_pid > /dev/null 2>&1; then
            echo -e "${YELLOW}  停止 $service (PID: $old_pid)${NC}"
            kill $old_pid 2>/dev/null || true
            ((stopped_count++))
        fi
        rm -f "$LOG_DIR/$service.pid"
    fi
done

# 清理旧的air进程
pkill -f "air.*payment" 2>/dev/null || true
sleep 2

if [ $stopped_count -gt 0 ]; then
    echo -e "${GREEN}✓ 已停止 $stopped_count 个旧服务${NC}"
else
    echo -e "${GREEN}✓ 没有运行中的服务${NC}"
fi
echo ""

# ==========================================
# 启动所有服务
# ==========================================
echo -e "${BLUE}[4/5] 启动所有微服务...${NC}"
echo ""

# 先启动config-service（其他服务可能依赖它）
start_service "config-service" ${SERVICES["config-service"]}
sleep 3

# 启动其他服务
for service in "${!SERVICES[@]}"; do
    if [ "$service" != "config-service" ]; then
        start_service "$service" ${SERVICES[$service]}
        sleep 1
    fi
done

echo ""

# ==========================================
# 验证服务启动
# ==========================================
echo -e "${BLUE}[5/5] 验证服务启动状态...${NC}"
echo ""

sleep 5  # 等待服务启动

running=0
failed=0
for service in "${!SERVICES[@]}"; do
    port=${SERVICES[$service]}
    if lsof -i:$port -sTCP:LISTEN -t > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $service (端口: $port)"
        ((running++))
    else
        echo -e "${RED}✗${NC} $service (端口: $port) - 未监听"
        ((failed++))
    fi
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}启动完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "运行中: ${GREEN}$running${NC} 个 | 失败: ${RED}$failed${NC} 个"
echo ""

if [ $failed -gt 0 ]; then
    echo -e "${YELLOW}提示: 某些服务未成功启动，请查看日志排查${NC}"
    echo "查看失败服务日志："
    echo "  tail -50 $LOG_DIR/<service-name>.log"
    echo ""
fi

echo -e "${BLUE}数据库配置:${NC}"
echo "  - PostgreSQL: localhost:40432"
echo "  - Redis: localhost:40379"
echo "  - Kafka: localhost:40092"
echo ""

echo -e "${BLUE}服务访问 (HTTPS/mTLS):${NC}"
echo "  - 所有服务使用 HTTPS，端口范围: 40001-40016"
echo "  - 访问需要客户端证书认证"
echo ""

echo -e "${BLUE}常用命令:${NC}"
echo "  查看状态: ./scripts/status-all-services.sh"
echo "  停止服务: ./scripts/stop-all-services.sh"
echo "  查看日志: tail -f $LOG_DIR/<service-name>.log"
echo ""

echo -e "${BLUE}测试 mTLS 访问:${NC}"
echo "  curl https://localhost:40004/health \\"
echo "    --cacert certs/ca/ca-cert.pem \\"
echo "    --cert certs/services/payment-gateway/cert.pem \\"
echo "    --key certs/services/payment-gateway/key.pem"
echo ""

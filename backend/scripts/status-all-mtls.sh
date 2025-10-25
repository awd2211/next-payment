#!/bin/bash
# 查看所有服务状态
# 用途: 检查基础设施和后端微服务运行状态

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_DIR="$SCRIPT_DIR/../logs"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "  Payment Platform - 服务状态"
echo "=========================================="
echo ""

# ============================================
# 1. 基础设施状态
# ============================================
echo "【基础设施】"
echo "-------------------------------------------"

# PostgreSQL
if docker ps | grep -q payment-postgres; then
    echo -e "  ${GREEN}✅${NC} PostgreSQL       (localhost:40432)"
else
    echo -e "  ${RED}❌${NC} PostgreSQL       (未运行)"
fi

# Redis
if docker ps | grep -q payment-redis; then
    echo -e "  ${GREEN}✅${NC} Redis            (localhost:40379)"
else
    echo -e "  ${RED}❌${NC} Redis            (未运行)"
fi

# Kafka
if docker ps | grep -q payment-kafka; then
    echo -e "  ${GREEN}✅${NC} Kafka            (localhost:40092)"
else
    echo -e "  ${RED}❌${NC} Kafka            (未运行)"
fi

# Kong
if docker ps | grep -q kong-gateway; then
    echo -e "  ${GREEN}✅${NC} Kong Gateway     (localhost:40080)"
else
    echo -e "  ${RED}❌${NC} Kong Gateway     (未运行)"
fi

# Konga
if docker ps | grep -q konga-ui; then
    echo -e "  ${GREEN}✅${NC} Konga UI         (localhost:40082)"
else
    echo -e "  ${RED}❌${NC} Konga UI         (未运行)"
fi

echo ""

# ============================================
# 2. 后端服务状态
# ============================================
echo "【后端微服务】"
echo "-------------------------------------------"

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

RUNNING=0
STOPPED=0

for service_info in "${SERVICES[@]}"; do
    SERVICE_NAME="${service_info%:*}"
    PORT="${service_info#*:}"

    # 格式化服务名（对齐）
    PADDED_NAME=$(printf "%-25s" "$SERVICE_NAME")

    if [ -f "$LOG_DIR/$SERVICE_NAME.pid" ]; then
        PID=$(cat "$LOG_DIR/$SERVICE_NAME.pid")
        if ps -p $PID > /dev/null 2>&1; then
            # 检查端口是否监听
            if lsof -i :$PORT > /dev/null 2>&1; then
                echo -e "  ${GREEN}✅${NC} $PADDED_NAME (PID: $PID, Port: $PORT)"
                ((RUNNING++))
            else
                echo -e "  ${YELLOW}⚠${NC}  $PADDED_NAME (PID: $PID, 端口未监听)"
                ((STOPPED++))
            fi
        else
            echo -e "  ${RED}❌${NC} $PADDED_NAME (进程已停止)"
            ((STOPPED++))
        fi
    else
        echo -e "  ${RED}❌${NC} $PADDED_NAME (未启动)"
        ((STOPPED++))
    fi
done

echo ""
echo "-------------------------------------------"
echo "  运行中: $RUNNING   已停止: $STOPPED"
echo ""

# ============================================
# 3. mTLS 配置状态
# ============================================
echo "【mTLS 配置】"
echo "-------------------------------------------"

# 检查证书
if [ -f "$SCRIPT_DIR/../certs/ca/ca-cert.pem" ]; then
    echo -e "  ${GREEN}✅${NC} CA 证书已生成"
else
    echo -e "  ${RED}❌${NC} CA 证书不存在"
fi

if [ -f "$SCRIPT_DIR/../certs/services/kong-gateway/cert.pem" ]; then
    echo -e "  ${GREEN}✅${NC} Kong 证书已生成"
else
    echo -e "  ${YELLOW}⚠${NC}  Kong 证书不存在"
fi

# 检查 Kong mTLS 配置
if docker ps | grep -q kong-gateway; then
    if docker exec kong-gateway env 2>/dev/null | grep -q "KONG_CLIENT_SSL=on"; then
        echo -e "  ${GREEN}✅${NC} Kong mTLS 已启用"
    else
        echo -e "  ${YELLOW}⚠${NC}  Kong mTLS 未启用"
    fi
fi

echo ""

# ============================================
# 4. 健康检查
# ============================================
echo "【健康检查】"
echo "-------------------------------------------"

# Kong Admin API
if curl -s -f http://localhost:40081/ > /dev/null 2>&1; then
    echo -e "  ${GREEN}✅${NC} Kong Admin API 正常"
else
    echo -e "  ${RED}❌${NC} Kong Admin API 不可访问"
fi

# Kong Proxy
if curl -s -f http://localhost:40080/ > /dev/null 2>&1; then
    echo -e "  ${GREEN}✅${NC} Kong Proxy 正常"
else
    echo -e "  ${RED}❌${NC} Kong Proxy 不可访问"
fi

# PostgreSQL
if docker exec payment-postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo -e "  ${GREEN}✅${NC} PostgreSQL 健康"
else
    echo -e "  ${RED}❌${NC} PostgreSQL 不健康"
fi

# Redis
if docker exec payment-redis redis-cli ping > /dev/null 2>&1; then
    echo -e "  ${GREEN}✅${NC} Redis 健康"
else
    echo -e "  ${RED}❌${NC} Redis 不健康"
fi

echo ""

# ============================================
# 5. 快捷操作
# ============================================
echo "=========================================="
echo "  快捷操作"
echo "=========================================="
echo "  启动所有服务:  ./scripts/start-all-mtls.sh"
echo "  停止所有服务:  ./scripts/stop-all-mtls.sh"
echo "  重启所有服务:  ./scripts/restart-all-mtls.sh"
echo ""
echo "  查看日志:"
echo "    tail -f logs/<service-name>.log"
echo ""
echo "  查看 Kong 日志:"
echo "    docker-compose logs -f kong"
echo ""

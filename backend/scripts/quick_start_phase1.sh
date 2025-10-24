#!/bin/bash
# =============================================================================
# 快速启动脚本 - Phase 1 测试
# =============================================================================

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Phase 1 快速启动${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查 Docker Compose 服务
echo -e "${YELLOW}[1/5] 检查 Docker Compose 服务...${NC}"
if ! docker ps | grep -q payment-postgres; then
    echo -e "${YELLOW}PostgreSQL 未运行，正在启动...${NC}"
    cd /home/eric/payment && docker-compose up -d postgres redis
    sleep 5
fi
echo -e "${GREEN}✓ Docker 服务正常${NC}"
echo ""

# 启动 merchant-auth-service
echo -e "${YELLOW}[2/5] 启动 merchant-auth-service...${NC}"
cd /home/eric/payment/backend/services/merchant-auth-service

export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=payment_merchant_auth
export REDIS_HOST=localhost
export REDIS_PORT=40379
export PORT=40011

# 杀掉旧进程
pkill -f "merchant-auth-service" || true
sleep 1

# 后台启动
nohup go run cmd/main.go > /tmp/merchant-auth-service.log 2>&1 &
MERCHANT_AUTH_PID=$!
echo -e "${GREEN}✓ merchant-auth-service 已启动 (PID: $MERCHANT_AUTH_PID)${NC}"
echo ""

# 等待服务就绪
echo -e "${YELLOW}[3/5] 等待服务就绪...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:40011/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ merchant-auth-service 已就绪${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗ merchant-auth-service 启动超时${NC}"
        cat /tmp/merchant-auth-service.log
        exit 1
    fi
    sleep 1
done
echo ""

# 启动 payment-gateway (旧方案)
echo -e "${YELLOW}[4/5] 启动 payment-gateway (旧方案)...${NC}"
cd /home/eric/payment/backend/services/payment-gateway

export DB_HOST=localhost
export DB_PORT=40432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=payment_gateway
export REDIS_HOST=localhost
export REDIS_PORT=40379
export PORT=40003
export USE_AUTH_SERVICE=false  # 旧方案

pkill -f "payment-gateway" || true
sleep 1

nohup go run cmd/main.go > /tmp/payment-gateway-old.log 2>&1 &
GATEWAY_OLD_PID=$!
echo -e "${GREEN}✓ payment-gateway (旧方案) 已启动 (PID: $GATEWAY_OLD_PID)${NC}"
echo ""

# 等待服务就绪
echo -e "${YELLOW}等待 payment-gateway 就绪...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:40003/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ payment-gateway 已就绪${NC}"
        break
    fi
    sleep 1
done
echo ""

# 显示服务状态
echo -e "${YELLOW}[5/5] 服务状态总览${NC}"
echo ""
echo -e "${BLUE}运行中的服务:${NC}"
echo "  - merchant-auth-service: http://localhost:40011 (PID: $MERCHANT_AUTH_PID)"
echo "  - payment-gateway:       http://localhost:40003 (PID: $GATEWAY_OLD_PID)"
echo ""

echo -e "${BLUE}日志文件:${NC}"
echo "  - /tmp/merchant-auth-service.log"
echo "  - /tmp/payment-gateway-old.log"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}服务启动完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

echo -e "${YELLOW}下一步操作:${NC}"
echo ""
echo "1. 执行数据迁移:"
echo "   cd /home/eric/payment/backend"
echo "   ./scripts/migrate_api_keys_to_auth_service.sh"
echo ""
echo "2. 运行集成测试:"
echo "   ./scripts/test_api_key_migration.sh"
echo ""
echo "3. 测试新方案 (重启 payment-gateway):"
echo "   pkill -f payment-gateway"
echo "   cd services/payment-gateway"
echo "   export USE_AUTH_SERVICE=true"
echo "   export MERCHANT_AUTH_SERVICE_URL=http://localhost:40011"
echo "   go run cmd/main.go"
echo ""
echo "4. 查看日志:"
echo "   tail -f /tmp/merchant-auth-service.log"
echo "   tail -f /tmp/payment-gateway-old.log"
echo ""
echo "5. 停止所有服务:"
echo "   pkill -f merchant-auth-service"
echo "   pkill -f payment-gateway"
echo ""

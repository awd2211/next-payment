#!/bin/bash

# 停止所有微服务的脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}停止所有支付平台微服务 (mTLS)${NC}"
echo -e "${YELLOW}========================================${NC}"

# 日志目录
LOG_DIR="/home/eric/payment/backend/logs"

# 服务列表（所有19个微服务）
SERVICES=(
    "admin-bff-service"
    "merchant-bff-service"
    "payment-gateway"
    "order-service"
    "channel-adapter"
    "risk-service"
    "accounting-service"
    "notification-service"
    "analytics-service"
    "config-service"
    "merchant-auth-service"
    "merchant-policy-service"
    "settlement-service"
    "withdrawal-service"
    "kyc-service"
    "cashier-service"
    "reconciliation-service"
    "dispute-service"
    "merchant-quota-service"
)

# 停止函数
stop_service() {
    local service=$1
    local pid_file="$LOG_DIR/$service.pid"

    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            echo -e "${YELLOW}停止 $service (PID: $pid)...${NC}"
            kill $pid 2>/dev/null || true
            sleep 1

            # 如果还在运行，强制杀掉
            if ps -p $pid > /dev/null 2>&1; then
                kill -9 $pid 2>/dev/null || true
            fi
            echo -e "${GREEN}✓ $service 已停止${NC}"
        else
            echo -e "${YELLOW}$service 未运行${NC}"
        fi
        rm -f "$pid_file"
    else
        echo -e "${YELLOW}$service PID文件不存在${NC}"
    fi
}

# 停止所有服务
for service in "${SERVICES[@]}"; do
    stop_service "$service"
done

# 清理所有air相关进程
echo ""
echo -e "${YELLOW}清理所有air进程...${NC}"
pkill -f "air.*payment" 2>/dev/null && echo -e "${GREEN}✓ air进程已清理${NC}" || echo -e "${YELLOW}没有运行的air进程${NC}"

# 清理tmp目录
echo ""
echo -e "${YELLOW}清理临时文件...${NC}"
for service in "${SERVICES[@]}"; do
    tmp_dir="/home/eric/payment/backend/services/$service/tmp"
    if [ -d "$tmp_dir" ]; then
        rm -rf "$tmp_dir"
        echo -e "${GREEN}✓ 已清理 $service/tmp${NC}"
    fi
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}所有服务已停止！${NC}"
echo -e "${GREEN}========================================${NC}"

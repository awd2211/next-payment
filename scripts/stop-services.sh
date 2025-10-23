#!/bin/bash

# 停止所有微服务

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}停止所有微服务${NC}"
echo -e "${BLUE}========================================${NC}\n"

# 服务列表
services=(
    "accounting-service"
    "risk-service"
    "notification-service"
    "analytics-service"
    "config-service"
    "payment-gateway"
    "order-service"
    "channel-adapter"
    "admin-service"
    "merchant-service"
)

for service in "${services[@]}"; do
    pid_file="/tmp/${service}.pid"

    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            echo -e "${YELLOW}停止 $service (PID: $pid)...${NC}"
            kill "$pid" 2>/dev/null
            rm -f "$pid_file"
            echo -e "${GREEN}✓ $service 已停止${NC}"
        else
            echo -e "${YELLOW}$service 进程不存在，清理 PID 文件${NC}"
            rm -f "$pid_file"
        fi
    else
        echo -e "${YELLOW}未找到 $service 的 PID 文件${NC}"
    fi
done

# 清理可能的孤立进程
echo -e "\n${YELLOW}清理孤立进程...${NC}"
pkill -f "air" 2>/dev/null

echo -e "\n${GREEN}所有服务已停止！${NC}"

#!/bin/bash

# 查看所有微服务状态的脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}支付平台微服务状态 (mTLS 模式)${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 日志目录
LOG_DIR="/home/eric/payment/backend/logs"

# 服务列表和端口（按端口排序）
declare -A SERVICES=(
    ["admin-service"]=40001
    ["merchant-service"]=40002
    ["payment-gateway"]=40003
    ["order-service"]=40004
    ["channel-adapter"]=40005
    ["risk-service"]=40006
    ["accounting-service"]=40007
    ["notification-service"]=40008
    ["analytics-service"]=40009
    ["config-service"]=40010
    ["merchant-auth-service"]=40011
    ["merchant-config-service"]=40012
    ["settlement-service"]=40013
    ["withdrawal-service"]=40014
    ["kyc-service"]=40015
    ["reconciliation-service"]=40020
    ["dispute-service"]=40021
    ["merchant-limit-service"]=40022
)

# 统计
running_count=0
stopped_count=0

# 检查函数
check_service() {
    local service=$1
    local port=$2
    local pid_file="$LOG_DIR/$service.pid"
    local status=""
    local pid=""

    printf "%-25s" "$service"

    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            # 检查端口是否监听
            if lsof -i:$port -sTCP:LISTEN -t > /dev/null 2>&1; then
                echo -e "${GREEN}运行中${NC}  PID: $pid  端口: $port"
                ((running_count++))
            else
                echo -e "${YELLOW}启动中${NC}  PID: $pid  端口: $port (等待监听)"
                ((running_count++))
            fi
        else
            echo -e "${RED}已停止${NC}  (PID文件存在但进程不存在)"
            ((stopped_count++))
        fi
    else
        echo -e "${RED}已停止${NC}  (无PID文件)"
        ((stopped_count++))
    fi
}

# 检查所有服务
for service in "${!SERVICES[@]}"; do
    check_service "$service" ${SERVICES[$service]}
done

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "总计: ${GREEN}$running_count${NC} 个服务运行中, ${RED}$stopped_count${NC} 个服务已停止"
echo -e "${BLUE}========================================${NC}"
echo ""

# 显示日志位置
if [ $running_count -gt 0 ]; then
    echo "查看日志："
    echo "  tail -f $LOG_DIR/<service-name>.log"
    echo ""
    echo "查看最近的日志："
    echo "  tail -20 $LOG_DIR/<service-name>.log"
    echo ""
fi

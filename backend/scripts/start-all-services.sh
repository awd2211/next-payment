#!/bin/bash

# 启动所有微服务的脚本
# 使用air进行热重载

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}启动所有支付平台微服务${NC}"
echo -e "${GREEN}========================================${NC}"

# 加载环境变量
if [ -f "/home/eric/payment/backend/.env" ]; then
    export $(grep -v '^#' /home/eric/payment/backend/.env | xargs)
    echo -e "${GREEN}已加载环境变量${NC}"
fi

# 设置工作区
export GOWORK=/home/eric/payment/backend/go.work

# 服务列表和端口
declare -A SERVICES=(
    ["config-service"]=40010
    ["admin-service"]=40001
    ["merchant-service"]=40002
    ["payment-gateway"]=40003
    ["order-service"]=40004
    ["channel-adapter"]=40005
    ["risk-service"]=40006
    ["accounting-service"]=40007
    ["notification-service"]=40008
    ["analytics-service"]=40009
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

    echo -e "${YELLOW}启动 $service (端口: $port)...${NC}"

    cd "$service_dir"
    nohup ~/go/bin/air -c .air.toml > "$LOG_DIR/$service.log" 2>&1 &
    local pid=$!
    echo $pid > "$LOG_DIR/$service.pid"

    echo -e "${GREEN}✓ $service 已启动 (PID: $pid)${NC}"
}

# 停止已运行的服务
echo -e "${YELLOW}检查并停止已运行的服务...${NC}"
for service in "${!SERVICES[@]}"; do
    if [ -f "$LOG_DIR/$service.pid" ]; then
        old_pid=$(cat "$LOG_DIR/$service.pid")
        if ps -p $old_pid > /dev/null 2>&1; then
            echo "停止旧的 $service 进程 (PID: $old_pid)"
            kill $old_pid 2>/dev/null || true
            sleep 1
        fi
        rm -f "$LOG_DIR/$service.pid"
    fi
done

# 清理旧的air进程
pkill -f "air.*payment" 2>/dev/null || true
sleep 2

echo ""
echo -e "${GREEN}开始启动所有服务...${NC}"
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
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}所有服务已启动完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "服务访问地址："
for service in "${!SERVICES[@]}"; do
    port=${SERVICES[$service]}
    echo "  - $service: http://localhost:$port"
done
echo ""
echo "查看日志："
echo "  tail -f $LOG_DIR/<service-name>.log"
echo ""
echo "停止所有服务："
echo "  ./scripts/stop-all-services.sh"
echo ""

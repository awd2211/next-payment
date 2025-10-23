#!/bin/bash

# 开发模式下使用 Air 热加载运行所有微服务
# 需要先安装 Air: go install github.com/cosmtrek/air@latest

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}启动支付平台微服务 (Air 热加载模式)${NC}"
echo -e "${BLUE}========================================${NC}\n"

# 检查 Air 是否已安装
if ! command -v air &> /dev/null; then
    echo -e "${RED}错误: Air 未安装${NC}"
    echo -e "${YELLOW}请先安装 Air: go install github.com/cosmtrek/air@latest${NC}"
    exit 1
fi

# 服务列表（服务名:内部端口:外部端口）
services=(
    "admin-service:8000:40000"
    "merchant-service:8001:40001"
    "payment-gateway:8002:40002"
    "channel-adapter:8003:40003"
    "order-service:8004:40004"
    "accounting-service:8005:40005"
    "risk-service:8006:40006"
    "notification-service:8007:40007"
    "analytics-service:8008:40008"
    "config-service:8009:40009"
)

# 进入 backend/services 目录
cd "$(dirname "$0")/../backend/services" || exit

# 启动每个服务
for service_info in "${services[@]}"; do
    IFS=':' read -r service internal_port external_port <<< "$service_info"

    if [ -d "$service" ]; then
        echo -e "${GREEN}启动 $service (端口: $external_port)...${NC}"
        cd "$service" || continue

        # 设置端口环境变量并在后台启动 air
        PORT=$external_port air > "../../logs/${service}.log" 2>&1 &
        echo $! > "/tmp/${service}.pid"

        cd ..
        sleep 1
    else
        echo -e "${YELLOW}警告: $service 目录不存在${NC}"
    fi
done

echo -e "\n${GREEN}所有服务已启动！${NC}"
echo -e "${YELLOW}日志位置: backend/logs/${NC}"
echo -e "${YELLOW}停止所有服务: ./scripts/stop-services.sh${NC}\n"

# 查看日志提示
echo -e "${BLUE}查看特定服务日志:${NC}"
for service_info in "${services[@]}"; do
    IFS=':' read -r service internal_port external_port <<< "$service_info"
    echo -e "  tail -f backend/logs/${service}.log"
done

# 显示服务访问地址
echo -e "\n${BLUE}服务访问地址:${NC}"
for service_info in "${services[@]}"; do
    IFS=':' read -r service internal_port external_port <<< "$service_info"
    echo -e "  ${service}: http://localhost:${external_port}"
done

#!/bin/bash

#######################################
# System Status Dashboard
# 显示完整的系统健康状态
#######################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# 清屏
clear

echo -e "${BOLD}${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║        Global Payment Platform - System Dashboard           ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

#######################################
# 1. 基础设施状态
#######################################
echo -e "${BOLD}${BLUE}[1] Infrastructure Status${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# PostgreSQL
echo -n "PostgreSQL (40432):       "
if docker ps | grep -q payment-postgres; then
    if psql -h localhost -p 40432 -U postgres -d postgres -c "SELECT 1" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running & Accessible${NC}"
    else
        echo -e "${YELLOW}⚠️  Running but not accessible${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

# Redis
echo -n "Redis (40379):            "
if docker ps | grep -q payment-redis; then
    if redis-cli -h localhost -p 40379 ping >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running & Accessible${NC}"
    else
        echo -e "${YELLOW}⚠️  Running but not accessible${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

# Kafka
echo -n "Kafka (40092):            "
if docker ps | grep -q payment-kafka; then
    echo -e "${GREEN}✅ Running${NC}"
else
    echo -e "${RED}❌ Not Running${NC}"
fi

# Prometheus
echo -n "Prometheus (40090):       "
if docker ps | grep -q payment-prometheus; then
    if curl -s http://localhost:40090/-/healthy >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running & Healthy${NC}"
    else
        echo -e "${YELLOW}⚠️  Running but unhealthy${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

# Grafana
echo -n "Grafana (40300):          "
if docker ps | grep -q payment-grafana; then
    if curl -s http://localhost:40300/api/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running & Healthy${NC}"
    else
        echo -e "${YELLOW}⚠️  Running but unhealthy${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

# Jaeger
echo -n "Jaeger (40686):           "
if docker ps | grep -q payment-jaeger; then
    if curl -s http://localhost:40686/ >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running & Accessible${NC}"
    else
        echo -e "${YELLOW}⚠️  Running but not accessible${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

echo ""

#######################################
# 2. 后端服务状态
#######################################
echo -e "${BOLD}${BLUE}[2] Backend Services Status (19 Services)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 服务端口映射
declare -A SERVICE_PORTS=(
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
    ["merchant-auth-service"]=40011
    ["merchant-config-service"]=40012
    ["settlement-service"]=40013
    ["withdrawal-service"]=40014
    ["kyc-service"]=40015
    ["cashier-service"]=40016
    ["dispute-service"]=40017
    ["reconciliation-service"]=40018
    ["merchant-limit-service"]=40022
)

running_count=0
total_count=${#SERVICE_PORTS[@]}

for service in "${!SERVICE_PORTS[@]}"; do
    port=${SERVICE_PORTS[$service]}

    printf "%-30s" "$service:"

    # 检查端口是否监听
    if lsof -i :$port -sTCP:LISTEN >/dev/null 2>&1; then
        running_count=$((running_count + 1))
        # 尝试访问健康检查端点
        if curl -s -f http://localhost:$port/health >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Running & Healthy (${port})${NC}"
        else
            echo -e "${YELLOW}⚠️  Running (${port}) - Health check failed${NC}"
        fi
    else
        echo -e "${RED}❌ Not Running (${port})${NC}"
    fi
done | sort

echo ""
echo -e "${BOLD}Services Summary: ${GREEN}${running_count}/${total_count} Running${NC}"

#######################################
# 3. 前端应用状态
#######################################
echo ""
echo -e "${BOLD}${BLUE}[3] Frontend Applications Status${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Admin Portal
echo -n "Admin Portal (5173):      "
if lsof -i :5173 >/dev/null 2>&1; then
    if curl -s http://localhost:5173 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running${NC}"
    else
        echo -e "${YELLOW}⚠️  Port open but not accessible${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

# Merchant Portal
echo -n "Merchant Portal (5174):   "
if lsof -i :5174 >/dev/null 2>&1; then
    if curl -s http://localhost:5174 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running${NC}"
    else
        echo -e "${YELLOW}⚠️  Port open but not accessible${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

# Website
echo -n "Website (5175):           "
if lsof -i :5175 >/dev/null 2>&1; then
    if curl -s http://localhost:5175 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Running${NC}"
    else
        echo -e "${YELLOW}⚠️  Port open but not accessible${NC}"
    fi
else
    echo -e "${RED}❌ Not Running${NC}"
fi

#######################################
# 4. 数据库状态
#######################################
echo ""
echo -e "${BOLD}${BLUE}[4] Database Status (19 Databases)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if psql -h localhost -p 40432 -U postgres -d postgres -c "SELECT 1" >/dev/null 2>&1; then
    db_count=$(psql -h localhost -p 40432 -U postgres -t -c "SELECT count(*) FROM pg_database WHERE datname LIKE 'payment_%';" 2>/dev/null | tr -d ' ')
    echo -e "Databases created:        ${GREEN}${db_count}/19${NC}"

    # 显示数据库列表
    echo ""
    echo "Database List:"
    psql -h localhost -p 40432 -U postgres -t -c "SELECT datname FROM pg_database WHERE datname LIKE 'payment_%' ORDER BY datname;" 2>/dev/null | while read db; do
        if [ -n "$db" ]; then
            echo -e "  • ${CYAN}$(echo $db | tr -d ' ')${NC}"
        fi
    done
else
    echo -e "${RED}❌ Cannot connect to PostgreSQL${NC}"
fi

#######################################
# 5. 系统资源
#######################################
echo ""
echo -e "${BOLD}${BLUE}[5] System Resources${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# CPU
cpu_usage=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1}')
echo -e "CPU Usage:                ${CYAN}${cpu_usage}%${NC}"

# Memory
mem_info=$(free -m | awk 'NR==2{printf "%.1f%% (%dMB / %dMB)", $3*100/$2, $3, $2}')
echo -e "Memory Usage:             ${CYAN}${mem_info}${NC}"

# Disk
disk_usage=$(df -h / | awk 'NR==2{printf "%s (%s / %s)", $5, $3, $2}')
echo -e "Disk Usage:               ${CYAN}${disk_usage}${NC}"

# Docker Stats
if command -v docker >/dev/null 2>&1; then
    docker_containers=$(docker ps | wc -l)
    docker_containers=$((docker_containers - 1))
    echo -e "Docker Containers:        ${CYAN}${docker_containers} running${NC}"
fi

#######################################
# 6. 快速链接
#######################################
echo ""
echo -e "${BOLD}${BLUE}[6] Quick Links${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Admin Portal:             ${CYAN}http://localhost:5173${NC}"
echo -e "Merchant Portal:          ${CYAN}http://localhost:5174${NC}"
echo -e "Grafana Dashboard:        ${CYAN}http://localhost:40300${NC} (admin/admin)"
echo -e "Prometheus:               ${CYAN}http://localhost:40090${NC}"
echo -e "Jaeger UI:                ${CYAN}http://localhost:40686${NC}"
echo -e "Swagger (Admin):          ${CYAN}http://localhost:40001/swagger/index.html${NC}"
echo -e "Swagger (Merchant):       ${CYAN}http://localhost:40002/swagger/index.html${NC}"
echo -e "Swagger (Gateway):        ${CYAN}http://localhost:40003/swagger/index.html${NC}"

#######################################
# Footer
#######################################
echo ""
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}Last updated: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

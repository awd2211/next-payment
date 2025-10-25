#!/bin/bash

# ==============================================
# 支付平台微服务管理脚本 (mTLS 模式)
# ==============================================
# 使用方法:
#   ./manage-services.sh start   - 启动所有服务
#   ./manage-services.sh stop    - 停止所有服务
#   ./manage-services.sh restart - 重启所有服务
#   ./manage-services.sh status  - 查看服务状态
#   ./manage-services.sh logs <service-name> - 查看服务日志
# ==============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 切换到 backend 目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
cd "$BACKEND_DIR"

# 日志目录
LOG_DIR="$BACKEND_DIR/logs"
mkdir -p "$LOG_DIR"

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
    ["merchant-auth-service"]=40011
    ["merchant-config-service"]=40012
    ["settlement-service"]=40013
    ["withdrawal-service"]=40014
    ["kyc-service"]=40015
    ["cashier-service"]=40016
)

# ==============================================
# 环境变量配置
# ==============================================
setup_environment() {
    # 加载 .env 文件
    if [ -f "$BACKEND_DIR/.env" ]; then
        export $(grep -v '^#' "$BACKEND_DIR/.env" | xargs)
    fi

    # 设置工作区
    export GOWORK="$BACKEND_DIR/go.work"

    # mTLS 配置（使用绝对路径）
    export ENABLE_MTLS=true
    export TLS_CA_FILE="/home/eric/payment/backend/certs/ca/ca-cert.pem"

    # 数据库配置 (PostgreSQL Docker 端口: 40432)
    export DB_HOST=localhost
    export DB_PORT=40432
    export DB_USER=postgres
    export DB_PASSWORD=postgres

    # Redis 配置 (Docker 端口: 40379)
    export REDIS_HOST=localhost
    export REDIS_PORT=40379
    export REDIS_PASSWORD=""

    # Kafka 配置 (Docker 端口: 40092)
    export KAFKA_BROKERS=localhost:40092
}

# ==============================================
# 检查 Docker 是否安装
# ==============================================
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}✗ Docker 未安装${NC}"
        echo -e "${YELLOW}  请安装 Docker: https://docs.docker.com/get-docker/${NC}"
        return 1
    fi

    if ! docker info &> /dev/null; then
        echo -e "${RED}✗ Docker 服务未运行${NC}"
        echo -e "${YELLOW}  请启动 Docker 服务: sudo systemctl start docker${NC}"
        return 1
    fi

    echo -e "${GREEN}✓ Docker 运行正常${NC}"
    return 0
}

# ==============================================
# 启动 Docker 基础设施
# ==============================================
start_infrastructure() {
    echo -e "${BLUE}启动 Docker 基础设施...${NC}"
    echo ""

    cd "$BACKEND_DIR/.." || exit 1

    # 检查 docker-compose.yml 是否存在
    if [ ! -f "docker-compose.yml" ]; then
        echo -e "${RED}✗ docker-compose.yml 不存在${NC}"
        exit 1
    fi

    # 定义所有基础设施组件
    declare -A INFRA_COMPONENTS=(
        ["payment-postgres"]="postgres"
        ["payment-redis"]="redis"
        ["payment-zookeeper"]="zookeeper"
        ["payment-kafka"]="kafka"
        ["payment-kafka-ui"]="kafka-ui"
        ["kong-postgres"]="kong-database"
        ["kong-bootstrap"]="kong-bootstrap"
        ["kong-gateway"]="kong"
        ["konga-ui"]="konga"
        ["payment-prometheus"]="prometheus"
        ["payment-grafana"]="grafana"
        ["payment-jaeger"]="jaeger"
        ["payment-postgres-exporter"]="postgres-exporter"
        ["payment-redis-exporter"]="redis-exporter"
        ["payment-kafka-exporter"]="kafka-exporter"
        ["payment-cadvisor"]="cadvisor"
        ["payment-node-exporter"]="node-exporter"
    )

    # 检查哪些服务已经在运行
    local already_running=0
    local need_start=()

    echo -e "${CYAN}检查现有容器...${NC}"
    for container in "${!INFRA_COMPONENTS[@]}"; do
        service=${INFRA_COMPONENTS[$container]}
        if docker ps --format '{{.Names}}' | grep -q "$container"; then
            echo -e "${GREEN}  ✓ $service${NC}"
            ((already_running++))
        else
            need_start+=("$service")
        fi
    done

    # 如果全部都在运行，则跳过启动
    if [ $already_running -eq ${#INFRA_COMPONENTS[@]} ]; then
        echo ""
        echo -e "${GREEN}所有基础设施已运行，无需启动${NC}"
        cd "$BACKEND_DIR" || exit 1
        return 0
    fi

    # 启动需要启动的服务
    if [ ${#need_start[@]} -gt 0 ]; then
        echo ""
        echo -e "${CYAN}需要启动 ${#need_start[@]} 个组件...${NC}"
        docker compose up -d "${need_start[@]}"

        echo ""
        echo -e "${CYAN}等待基础设施启动...${NC}"
        sleep 10
    fi

    echo ""
    echo -e "${CYAN}验证基础设施状态...${NC}"
    echo ""

    # 核心数据存储
    echo -e "${YELLOW}核心数据存储:${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-postgres" && echo -e "${GREEN}  ✓ PostgreSQL${NC} (40432)" || echo -e "${RED}  ✗ PostgreSQL${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-redis" && echo -e "${GREEN}  ✓ Redis${NC} (40379)" || echo -e "${RED}  ✗ Redis${NC}"

    # 消息队列
    echo ""
    echo -e "${YELLOW}消息队列:${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-zookeeper" && echo -e "${GREEN}  ✓ Zookeeper${NC} (42181)" || echo -e "${RED}  ✗ Zookeeper${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-kafka" && echo -e "${GREEN}  ✓ Kafka${NC} (40092)" || echo -e "${RED}  ✗ Kafka${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-kafka-ui" && echo -e "${GREEN}  ✓ Kafka UI${NC} (40084)" || echo -e "${YELLOW}  ⚠ Kafka UI${NC}"

    # API 网关
    echo ""
    echo -e "${YELLOW}API 网关:${NC}"
    docker ps --format '{{.Names}}' | grep -q "kong-postgres" && echo -e "${GREEN}  ✓ Kong PostgreSQL${NC} (40433)" || echo -e "${RED}  ✗ Kong PostgreSQL${NC}"
    docker ps --format '{{.Names}}' | grep -q "kong-gateway" && echo -e "${GREEN}  ✓ Kong Gateway${NC} (40080, 40081)" || echo -e "${RED}  ✗ Kong Gateway${NC}"
    docker ps --format '{{.Names}}' | grep -q "konga-ui" && echo -e "${GREEN}  ✓ Konga UI${NC} (50001)" || echo -e "${YELLOW}  ⚠ Konga UI${NC}"

    # 监控系统
    echo ""
    echo -e "${YELLOW}监控系统:${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-prometheus" && echo -e "${GREEN}  ✓ Prometheus${NC} (40090)" || echo -e "${YELLOW}  ⚠ Prometheus${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-grafana" && echo -e "${GREEN}  ✓ Grafana${NC} (40300)" || echo -e "${YELLOW}  ⚠ Grafana${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-jaeger" && echo -e "${GREEN}  ✓ Jaeger${NC} (50686)" || echo -e "${YELLOW}  ⚠ Jaeger${NC}"

    # 监控导出器
    echo ""
    echo -e "${YELLOW}监控导出器:${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-postgres-exporter" && echo -e "${GREEN}  ✓ PostgreSQL Exporter${NC} (40187)" || echo -e "${YELLOW}  ⚠ PostgreSQL Exporter${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-redis-exporter" && echo -e "${GREEN}  ✓ Redis Exporter${NC} (40121)" || echo -e "${YELLOW}  ⚠ Redis Exporter${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-kafka-exporter" && echo -e "${GREEN}  ✓ Kafka Exporter${NC} (40308)" || echo -e "${YELLOW}  ⚠ Kafka Exporter${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-cadvisor" && echo -e "${GREEN}  ✓ cAdvisor${NC} (40180)" || echo -e "${YELLOW}  ⚠ cAdvisor${NC}"
    docker ps --format '{{.Names}}' | grep -q "payment-node-exporter" && echo -e "${GREEN}  ✓ Node Exporter${NC} (40100)" || echo -e "${YELLOW}  ⚠ Node Exporter${NC}"

    cd "$BACKEND_DIR" || exit 1

    echo ""
    echo -e "${GREEN}✓ 基础设施启动完成${NC}"
    echo ""
    return 0
}

# ==============================================
# 停止 Docker 基础设施
# ==============================================
stop_infrastructure() {
    echo -e "${BLUE}停止 Docker 基础设施...${NC}"
    echo ""

    cd "$BACKEND_DIR/.." || exit 1

    # 停止所有基础设施组件
    echo -e "${CYAN}停止所有容器...${NC}"
    docker compose stop \
        postgres redis \
        kafka zookeeper kafka-ui \
        kong-database kong kong konga \
        prometheus grafana jaeger \
        postgres-exporter redis-exporter kafka-exporter \
        cadvisor node-exporter

    echo ""
    echo -e "${GREEN}✓ Docker 基础设施已停止${NC}"

    cd "$BACKEND_DIR" || exit 1
}

# ==============================================
# 检查数据库连接
# ==============================================
check_database_connection() {
    echo -e "${CYAN}检查数据库连接...${NC}"

    # 检查 PostgreSQL 连接
    if PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d postgres -c "SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✓ PostgreSQL 连接成功${NC}"
    else
        echo -e "${RED}✗ PostgreSQL 连接失败${NC}"
        return 1
    fi

    # 检查 Redis 连接
    if redis-cli -h localhost -p 40379 ping &> /dev/null; then
        echo -e "${GREEN}✓ Redis 连接成功${NC}"
    else
        echo -e "${RED}✗ Redis 连接失败${NC}"
        return 1
    fi

    return 0
}

# ==============================================
# 初始化数据库
# ==============================================
init_databases() {
    echo -e "${CYAN}检查数据库是否已初始化...${NC}"

    # 检查是否需要初始化
    local need_init=false
    for db_name in payment_config payment_admin payment_merchant_auth payment_merchant payment_risk payment_channel payment_order payment_gateway payment_accounting payment_analytics payment_notify payment_settlement payment_withdrawal payment_kyc payment_cashier payment_merchant_config; do
        if ! PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -lqt | cut -d \| -f 1 | grep -qw "$db_name"; then
            need_init=true
            break
        fi
    done

    if [ "$need_init" = true ]; then
        echo -e "${YELLOW}需要初始化数据库${NC}"
        echo -e "${CYAN}运行初始化脚本...${NC}"

        if [ -f "$BACKEND_DIR/scripts/init-db.sh" ]; then
            "$BACKEND_DIR/scripts/init-db.sh"
            echo -e "${GREEN}✓ 数据库初始化完成${NC}"
        else
            echo -e "${RED}✗ 初始化脚本不存在: $BACKEND_DIR/scripts/init-db.sh${NC}"
            return 1
        fi
    else
        echo -e "${GREEN}✓ 数据库已初始化${NC}"
    fi

    return 0
}

# ==============================================
# 前置检查
# ==============================================
check_prerequisites() {
    local errors=0

    # 检查 Docker
    if ! check_docker; then
        ((errors++))
    fi

    # 检查 Docker 基础设施
    echo ""
    echo -e "${CYAN}检查 Docker 基础设施...${NC}"

    local infra_running=true
    if ! docker ps --format '{{.Names}}' | grep -q "payment-postgres"; then
        echo -e "${YELLOW}⚠ PostgreSQL 容器未运行${NC}"
        infra_running=false
    fi

    if ! docker ps --format '{{.Names}}' | grep -q "payment-redis"; then
        echo -e "${YELLOW}⚠ Redis 容器未运行${NC}"
        infra_running=false
    fi

    if ! docker ps --format '{{.Names}}' | grep -q "payment-kafka"; then
        echo -e "${YELLOW}⚠ Kafka 容器未运行${NC}"
        infra_running=false
    fi

    if [ "$infra_running" = false ]; then
        echo ""
        read -p "是否自动启动 Docker 基础设施? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            start_infrastructure

            # 检查数据库连接
            echo ""
            if check_database_connection; then
                # 初始化数据库
                echo ""
                if ! init_databases; then
                    ((errors++))
                fi
            else
                ((errors++))
            fi
        else
            echo -e "${RED}✗ 基础设施未运行${NC}"
            ((errors++))
        fi
    else
        echo -e "${GREEN}✓ Docker 基础设施运行正常${NC}"
    fi

    # 检查 mTLS 证书
    echo ""
    echo -e "${CYAN}检查 mTLS 证书...${NC}"
    if [ ! -f "$BACKEND_DIR/certs/ca/ca-cert.pem" ]; then
        echo -e "${YELLOW}⚠ CA 证书不存在${NC}"
        read -p "是否自动生成 mTLS 证书? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if [ -f "$BACKEND_DIR/scripts/generate-mtls-certs.sh" ]; then
                "$BACKEND_DIR/scripts/generate-mtls-certs.sh"
                echo -e "${GREEN}✓ mTLS 证书生成完成${NC}"
            else
                echo -e "${RED}✗ 证书生成脚本不存在${NC}"
                ((errors++))
            fi
        else
            echo -e "${RED}✗ mTLS 证书不存在${NC}"
            ((errors++))
        fi
    else
        echo -e "${GREEN}✓ mTLS 证书存在${NC}"

        # 检查各个服务的证书
        local missing_certs=0
        for service in "${!SERVICES[@]}"; do
            if [ ! -f "$BACKEND_DIR/certs/services/$service/cert.pem" ] || [ ! -f "$BACKEND_DIR/certs/services/$service/key.pem" ]; then
                ((missing_certs++))
            fi
        done

        if [ $missing_certs -gt 0 ]; then
            echo -e "${YELLOW}⚠ 缺少 $missing_certs 个服务的证书${NC}"
            echo -e "${YELLOW}  建议重新生成证书: ./scripts/generate-mtls-certs.sh${NC}"
        fi
    fi

    # 检查 air
    echo ""
    echo -e "${CYAN}检查开发工具...${NC}"
    if ! command -v ~/go/bin/air &> /dev/null; then
        echo -e "${YELLOW}⚠ air 未安装${NC}"
        echo -e "${YELLOW}  请运行: go install github.com/cosmtrek/air@v1.49.0${NC}"
        ((errors++))
    else
        echo -e "${GREEN}✓ air 已安装${NC}"
    fi

    # 检查 psql
    if ! command -v psql &> /dev/null; then
        echo -e "${YELLOW}⚠ psql 未安装 (可选)${NC}"
    else
        echo -e "${GREEN}✓ psql 已安装${NC}"
    fi

    # 检查 redis-cli
    if ! command -v redis-cli &> /dev/null; then
        echo -e "${YELLOW}⚠ redis-cli 未安装 (可选)${NC}"
    else
        echo -e "${GREEN}✓ redis-cli 已安装${NC}"
    fi

    if [ $errors -gt 0 ]; then
        echo ""
        echo -e "${RED}前置检查失败，请解决上述问题后再试${NC}"
        exit 1
    fi

    echo ""
    echo -e "${GREEN}✓ 所有前置检查通过${NC}"
}

# ==============================================
# 启动单个服务
# ==============================================
start_service() {
    local service=$1
    local port=$2
    local service_dir="$BACKEND_DIR/services/$service"

    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}  ✗ 服务目录不存在: $service${NC}"
        return 1
    fi

    # 设置服务专用的 mTLS 证书（使用绝对路径）
    export TLS_CERT_FILE="/home/eric/payment/backend/certs/services/$service/cert.pem"
    export TLS_KEY_FILE="/home/eric/payment/backend/certs/services/$service/key.pem"
    export TLS_CLIENT_CERT="$TLS_CERT_FILE"
    export TLS_CLIENT_KEY="$TLS_KEY_FILE"

    # 设置数据库名称
    case "$service" in
        "config-service") export DB_NAME=payment_config ;;
        "admin-service") export DB_NAME=payment_admin ;;
        "merchant-auth-service") export DB_NAME=payment_merchant_auth ;;
        "merchant-service") export DB_NAME=payment_merchant ;;
        "risk-service") export DB_NAME=payment_risk ;;
        "channel-adapter") export DB_NAME=payment_channel ;;
        "order-service") export DB_NAME=payment_order ;;
        "payment-gateway") export DB_NAME=payment_gateway ;;
        "accounting-service") export DB_NAME=payment_accounting ;;
        "analytics-service") export DB_NAME=payment_analytics ;;
        "notification-service") export DB_NAME=payment_notify ;;
        "settlement-service") export DB_NAME=payment_settlement ;;
        "withdrawal-service") export DB_NAME=payment_withdrawal ;;
        "kyc-service") export DB_NAME=payment_kyc ;;
        "cashier-service") export DB_NAME=payment_cashier ;;
        "merchant-config-service") export DB_NAME=payment_merchant_config ;;
    esac

    # payment-gateway 需要配置下游服务的 HTTPS 地址
    if [ "$service" == "payment-gateway" ]; then
        export ORDER_SERVICE_URL="https://localhost:40004"
        export RISK_SERVICE_URL="https://localhost:40006"
        export CHANNEL_SERVICE_URL="https://localhost:40005"
    fi

    echo -e "${CYAN}  启动 $service (端口: $port, DB: $DB_NAME)${NC}"

    cd "$service_dir"
    nohup ~/go/bin/air -c .air.toml > "$LOG_DIR/$service.log" 2>&1 &
    local pid=$!
    echo $pid > "$LOG_DIR/$service.pid"

    echo -e "${GREEN}  ✓ $service 已启动 (PID: $pid)${NC}"
}

# ==============================================
# 停止单个服务
# ==============================================
stop_service() {
    local service=$1
    local pid_file="$LOG_DIR/$service.pid"

    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            echo -e "${CYAN}  停止 $service (PID: $pid)${NC}"
            kill $pid 2>/dev/null || true
            sleep 1

            # 如果还在运行，强制杀掉
            if ps -p $pid > /dev/null 2>&1; then
                kill -9 $pid 2>/dev/null || true
            fi
            echo -e "${GREEN}  ✓ $service 已停止${NC}"
        fi
        rm -f "$pid_file"
    fi
}

# ==============================================
# 启动所有服务
# ==============================================
cmd_start() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}启动所有支付平台微服务 (mTLS)${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""

    echo -e "${BLUE}[1/5] 前置检查${NC}"
    check_prerequisites
    echo ""

    echo -e "${BLUE}[2/5] 加载环境变量${NC}"
    setup_environment
    echo -e "${GREEN}✓ 环境变量配置完成 (DB_PORT=40432, mTLS=enabled)${NC}"
    echo ""

    echo -e "${BLUE}[3/5] 停止已运行的服务${NC}"
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

    echo -e "${BLUE}[4/5] 启动所有微服务${NC}"
    echo ""

    # 先启动 config-service（其他服务可能依赖它）
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
    echo -e "${BLUE}[5/5] 验证服务启动状态${NC}"
    echo ""

    sleep 5  # 等待服务启动

    running=0
    failed=0
    for service in "${!SERVICES[@]}"; do
        port=${SERVICES[$service]}
        if lsof -i:$port -sTCP:LISTEN -t > /dev/null 2>&1; then
            echo -e "${GREEN}  ✓ $service${NC} (端口: $port)"
            ((running++))
        else
            echo -e "${RED}  ✗ $service${NC} (端口: $port) - 未监听"
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
        echo "  ./manage-services.sh logs <service-name>"
        echo ""
    fi

    show_info
}

# ==============================================
# 停止所有服务
# ==============================================
cmd_stop() {
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}停止所有支付平台微服务${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""

    for service in "${!SERVICES[@]}"; do
        stop_service "$service"
    done

    # 清理所有air相关进程
    echo ""
    echo -e "${CYAN}清理所有air进程${NC}"
    pkill -f "air.*payment" 2>/dev/null && echo -e "${GREEN}✓ air进程已清理${NC}" || echo -e "${GREEN}✓ 没有运行的air进程${NC}"

    # 清理tmp目录
    echo ""
    echo -e "${CYAN}清理临时文件${NC}"
    for service in "${!SERVICES[@]}"; do
        tmp_dir="$BACKEND_DIR/services/$service/tmp"
        if [ -d "$tmp_dir" ]; then
            rm -rf "$tmp_dir"
        fi
    done
    echo -e "${GREEN}✓ 临时文件已清理${NC}"

    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}所有服务已停止！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
}

# ==============================================
# 重启所有服务
# ==============================================
cmd_restart() {
    cmd_stop
    echo ""
    sleep 2
    cmd_start
}

# ==============================================
# 查看服务状态
# ==============================================
cmd_status() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}支付平台微服务状态 (mTLS)${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    running_count=0
    stopped_count=0

    for service in "${!SERVICES[@]}"; do
        port=${SERVICES[$service]}
        pid_file="$LOG_DIR/$service.pid"

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
    done

    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "总计: ${GREEN}$running_count${NC} 个服务运行中, ${RED}$stopped_count${NC} 个服务已停止"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    if [ $running_count -gt 0 ]; then
        show_info
    fi
}

# ==============================================
# 查看服务日志
# ==============================================
cmd_logs() {
    local service=$1

    if [ -z "$service" ]; then
        echo -e "${RED}错误: 请指定服务名称${NC}"
        echo "用法: ./manage-services.sh logs <service-name>"
        echo ""
        echo "可用服务:"
        for svc in "${!SERVICES[@]}"; do
            echo "  - $svc"
        done
        exit 1
    fi

    if [ ! -f "$LOG_DIR/$service.log" ]; then
        echo -e "${RED}错误: 日志文件不存在: $LOG_DIR/$service.log${NC}"
        exit 1
    fi

    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}$service 日志 (实时)${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    tail -f "$LOG_DIR/$service.log"
}

# ==============================================
# 显示使用信息
# ==============================================
show_info() {
    echo -e "${BLUE}数据库配置:${NC}"
    echo "  - PostgreSQL: localhost:40432"
    echo "  - Redis: localhost:40379"
    echo "  - Kafka: localhost:40092"
    echo ""

    echo -e "${BLUE}常用命令:${NC}"
    echo "  查看状态: ./manage-services.sh status"
    echo "  停止服务: ./manage-services.sh stop"
    echo "  重启服务: ./manage-services.sh restart"
    echo "  查看日志: ./manage-services.sh logs <service-name>"
    echo ""

    echo -e "${BLUE}测试 mTLS 访问:${NC}"
    echo "  curl https://localhost:40004/health \\"
    echo "    --cacert certs/ca/ca-cert.pem \\"
    echo "    --cert certs/services/payment-gateway/cert.pem \\"
    echo "    --key certs/services/payment-gateway/key.pem"
    echo ""
}

# ==============================================
# 基础设施管理命令
# ==============================================
cmd_infra() {
    local action=$1

    case "$action" in
        start)
            echo -e "${GREEN}========================================${NC}"
            echo -e "${GREEN}启动 Docker 基础设施${NC}"
            echo -e "${GREEN}========================================${NC}"
            echo ""
            start_infrastructure
            ;;
        stop)
            echo -e "${YELLOW}========================================${NC}"
            echo -e "${YELLOW}停止 Docker 基础设施${NC}"
            echo -e "${YELLOW}========================================${NC}"
            echo ""
            stop_infrastructure
            ;;
        status)
            echo -e "${BLUE}========================================${NC}"
            echo -e "${BLUE}Docker 基础设施状态${NC}"
            echo -e "${BLUE}========================================${NC}"
            echo ""

            # 核心数据存储
            echo -e "${YELLOW}核心数据存储:${NC}"
            docker ps --format '{{.Names}}' | grep -q "payment-postgres" && echo -e "${GREEN}  ✓ PostgreSQL${NC} (40432)" || echo -e "${RED}  ✗ PostgreSQL${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-redis" && echo -e "${GREEN}  ✓ Redis${NC} (40379)" || echo -e "${RED}  ✗ Redis${NC} (未运行)"

            # 消息队列
            echo ""
            echo -e "${YELLOW}消息队列:${NC}"
            docker ps --format '{{.Names}}' | grep -q "payment-zookeeper" && echo -e "${GREEN}  ✓ Zookeeper${NC} (42181)" || echo -e "${RED}  ✗ Zookeeper${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-kafka" && echo -e "${GREEN}  ✓ Kafka${NC} (40092)" || echo -e "${RED}  ✗ Kafka${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-kafka-ui" && echo -e "${GREEN}  ✓ Kafka UI${NC} (40084) - http://localhost:40084" || echo -e "${RED}  ✗ Kafka UI${NC} (未运行)"

            # API 网关
            echo ""
            echo -e "${YELLOW}API 网关:${NC}"
            docker ps --format '{{.Names}}' | grep -q "kong-postgres" && echo -e "${GREEN}  ✓ Kong PostgreSQL${NC} (40433)" || echo -e "${RED}  ✗ Kong PostgreSQL${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "kong-gateway" && echo -e "${GREEN}  ✓ Kong Gateway${NC} (40080, 40081) - http://localhost:40080" || echo -e "${RED}  ✗ Kong Gateway${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "konga-ui" && echo -e "${GREEN}  ✓ Konga UI${NC} (50001) - http://localhost:50001" || echo -e "${RED}  ✗ Konga UI${NC} (未运行)"

            # 监控系统
            echo ""
            echo -e "${YELLOW}监控系统:${NC}"
            docker ps --format '{{.Names}}' | grep -q "payment-prometheus" && echo -e "${GREEN}  ✓ Prometheus${NC} (40090) - http://localhost:40090" || echo -e "${RED}  ✗ Prometheus${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-grafana" && echo -e "${GREEN}  ✓ Grafana${NC} (40300) - http://localhost:40300 (admin/admin)" || echo -e "${RED}  ✗ Grafana${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-jaeger" && echo -e "${GREEN}  ✓ Jaeger${NC} (50686) - http://localhost:50686" || echo -e "${RED}  ✗ Jaeger${NC} (未运行)"

            # 监控导出器
            echo ""
            echo -e "${YELLOW}监控导出器:${NC}"
            docker ps --format '{{.Names}}' | grep -q "payment-postgres-exporter" && echo -e "${GREEN}  ✓ PostgreSQL Exporter${NC} (40187)" || echo -e "${RED}  ✗ PostgreSQL Exporter${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-redis-exporter" && echo -e "${GREEN}  ✓ Redis Exporter${NC} (40121)" || echo -e "${RED}  ✗ Redis Exporter${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-kafka-exporter" && echo -e "${GREEN}  ✓ Kafka Exporter${NC} (40308)" || echo -e "${RED}  ✗ Kafka Exporter${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-cadvisor" && echo -e "${GREEN}  ✓ cAdvisor${NC} (40180) - http://localhost:40180" || echo -e "${RED}  ✗ cAdvisor${NC} (未运行)"
            docker ps --format '{{.Names}}' | grep -q "payment-node-exporter" && echo -e "${GREEN}  ✓ Node Exporter${NC} (40100)" || echo -e "${RED}  ✗ Node Exporter${NC} (未运行)"

            echo ""
            ;;
        restart)
            echo -e "${CYAN}重启 Docker 基础设施${NC}"
            stop_infrastructure
            echo ""
            sleep 2
            start_infrastructure
            ;;
        *)
            echo -e "${RED}错误: 未知基础设施命令 '$action'${NC}"
            echo ""
            echo "用法: ./manage-services.sh infra <start|stop|status|restart>"
            exit 1
            ;;
    esac
}

# ==============================================
# 显示帮助信息
# ==============================================
show_help() {
    echo ""
    echo -e "${CYAN}支付平台微服务管理脚本 (mTLS 模式)${NC}"
    echo ""
    echo "用法:"
    echo "  ./manage-services.sh <command> [options]"
    echo ""
    echo -e "${YELLOW}服务管理:${NC}"
    echo -e "  ${GREEN}start${NC}                  - 启动所有微服务"
    echo -e "  ${GREEN}stop${NC}                   - 停止所有微服务"
    echo -e "  ${GREEN}restart${NC}                - 重启所有微服务"
    echo -e "  ${GREEN}status${NC}                 - 查看微服务状态"
    echo -e "  ${GREEN}logs <service>${NC}         - 查看服务日志 (实时)"
    echo ""
    echo -e "${YELLOW}基础设施管理:${NC}"
    echo -e "  ${GREEN}infra start${NC}            - 启动 Docker 基础设施"
    echo -e "  ${GREEN}infra stop${NC}             - 停止 Docker 基础设施"
    echo -e "  ${GREEN}infra status${NC}           - 查看基础设施状态"
    echo -e "  ${GREEN}infra restart${NC}          - 重启 Docker 基础设施"
    echo ""
    echo "示例:"
    echo "  ./manage-services.sh start"
    echo "  ./manage-services.sh status"
    echo "  ./manage-services.sh logs order-service"
    echo "  ./manage-services.sh infra status"
    echo ""
}

# ==============================================
# 主函数
# ==============================================
main() {
    local command=$1
    shift || true

    case "$command" in
        start)
            cmd_start
            ;;
        stop)
            cmd_stop
            ;;
        restart)
            cmd_restart
            ;;
        status)
            cmd_status
            ;;
        logs)
            cmd_logs "$@"
            ;;
        infra)
            cmd_infra "$@"
            ;;
        help|--help|-h|"")
            show_help
            ;;
        *)
            echo -e "${RED}错误: 未知命令 '$command'${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"

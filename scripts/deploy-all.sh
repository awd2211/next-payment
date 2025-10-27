#!/bin/bash

# ============================================================================
# 一键部署完整支付平台
# ============================================================================
# 功能:
# - 检查系统要求
# - 生成 mTLS 证书
# - 启动基础设施
# - 初始化数据库
# - 构建所有镜像
# - 启动所有服务
# - 健康检查
# ============================================================================

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 配置
BASE_DIR="/home/eric/payment"
BACKEND_DIR="$BASE_DIR/backend"
CERT_DIR="$BACKEND_DIR/certs"

# 打印函数
print_header() {
    echo ""
    echo -e "${CYAN}============================================================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}============================================================================${NC}"
    echo ""
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_step() {
    echo ""
    echo -e "${GREEN}▶ $1${NC}"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 未安装，请先安装"
        exit 1
    fi
}

# 检查系统要求
check_requirements() {
    print_header "步骤 1/8: 检查系统要求"

    # Docker
    print_info "检查 Docker..."
    check_command docker
    DOCKER_VERSION=$(docker --version | grep -oP '\d+\.\d+\.\d+')
    print_success "Docker 版本: $DOCKER_VERSION"

    # Docker Compose
    print_info "检查 Docker Compose..."
    check_command docker-compose
    COMPOSE_VERSION=$(docker-compose --version | grep -oP '\d+\.\d+\.\d+')
    print_success "Docker Compose 版本: $COMPOSE_VERSION"

    # Docker 是否运行
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker 守护进程未运行"
        exit 1
    fi
    print_success "Docker 守护进程运行中"

    # 检查磁盘空间（至少20GB）
    AVAILABLE_SPACE=$(df -BG "$BASE_DIR" | awk 'NR==2 {print $4}' | grep -oP '\d+')
    if [ "$AVAILABLE_SPACE" -lt 20 ]; then
        print_warning "磁盘可用空间不足 20GB，当前: ${AVAILABLE_SPACE}GB"
    else
        print_success "磁盘可用空间: ${AVAILABLE_SPACE}GB"
    fi

    # 检查内存（至少4GB）
    TOTAL_MEM=$(free -g | awk 'NR==2 {print $2}')
    if [ "$TOTAL_MEM" -lt 4 ]; then
        print_warning "系统内存不足 4GB，当前: ${TOTAL_MEM}GB"
    else
        print_success "系统内存: ${TOTAL_MEM}GB"
    fi
}

# 生成环境变量文件
generate_env_file() {
    print_header "步骤 2/8: 生成环境变量文件"

    ENV_FILE="$BASE_DIR/.env"

    if [ -f "$ENV_FILE" ]; then
        print_info ".env 文件已存在，跳过生成"
        return
    fi

    print_info "生成 .env 文件..."

    cat > "$ENV_FILE" << 'EOF'
# ============================================================================
# Payment Platform Environment Variables
# ============================================================================

# 数据库配置
DB_PASSWORD=postgres

# Redis 配置
REDIS_PASSWORD=

# JWT 密钥（生产环境必须修改！）
JWT_SECRET=payment-platform-super-secret-jwt-key-change-in-production

# Stripe 配置
STRIPE_API_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# SMTP 配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@payment-platform.com
EOF

    chmod 600 "$ENV_FILE"
    print_success "已生成 .env 文件（请根据需要修改配置）"
}

# 生成 mTLS 证书
generate_certificates() {
    print_header "步骤 3/8: 生成 mTLS 证书"

    if [ -f "$CERT_DIR/ca/ca-cert.pem" ]; then
        print_info "CA 证书已存在，跳过生成"
    else
        print_info "生成 CA 证书..."
        cd "$CERT_DIR"
        ./generate-ca-cert.sh
        print_success "CA 证书生成完成"
    fi

    print_info "为所有服务生成证书..."

    SERVICES=(
        "payment-gateway" "order-service" "channel-adapter" "risk-service"
        "accounting-service" "notification-service" "analytics-service"
        "config-service" "merchant-auth-service" "settlement-service"
        "withdrawal-service" "kyc-service" "cashier-service"
        "reconciliation-service" "dispute-service" "merchant-policy-service"
        "merchant-quota-service" "admin-bff-service" "merchant-bff-service"
    )

    for service in "${SERVICES[@]}"; do
        if [ -f "$CERT_DIR/services/$service/${service}.crt" ]; then
            print_info "  ✓ $service 证书已存在"
        else
            print_info "  → 生成 $service 证书..."
            cd "$CERT_DIR"
            ./generate-service-cert.sh "$service"
        fi
    done

    print_success "所有证书生成完成"
}

# 启动基础设施
start_infrastructure() {
    print_header "步骤 4/8: 启动基础设施"

    cd "$BASE_DIR"

    print_info "启动 PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger..."
    docker-compose up -d

    print_info "等待基础设施就绪..."
    sleep 10

    # 健康检查
    print_info "检查 PostgreSQL..."
    for i in {1..30}; do
        if docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
            print_success "PostgreSQL 已就绪"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "PostgreSQL 启动超时"
            exit 1
        fi
        sleep 2
    done

    print_info "检查 Redis..."
    if docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; then
        print_success "Redis 已就绪"
    else
        print_warning "Redis 可能未完全启动"
    fi

    print_success "基础设施启动完成"
}

# 初始化数据库
initialize_databases() {
    print_header "步骤 5/8: 初始化数据库"

    print_info "创建所有微服务数据库..."
    cd "$BACKEND_DIR"

    if [ -f "./scripts/init-db.sh" ]; then
        ./scripts/init-db.sh
        print_success "数据库初始化完成"
    else
        print_warning "init-db.sh 脚本不存在，跳过数据库初始化"
    fi
}

# 构建所有镜像
build_images() {
    print_header "步骤 6/8: 构建所有镜像"

    read -p "是否构建所有镜像？（首次部署或代码更新后需要）[Y/n] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]] && [[ ! -z $REPLY ]]; then
        print_info "跳过镜像构建"
        return
    fi

    cd "$BACKEND_DIR"

    print_info "使用自动化脚本构建所有镜像..."
    if [ -f "./scripts/build-all-docker-images.sh" ]; then
        ./scripts/build-all-docker-images.sh
    else
        print_info "使用 docker-compose build..."
        cd "$BASE_DIR"
        docker-compose -f docker-compose.services.yml build
        docker-compose -f docker-compose.bff.yml build
    fi

    print_success "镜像构建完成"
}

# 启动所有服务
start_services() {
    print_header "步骤 7/8: 启动所有服务"

    cd "$BASE_DIR"

    print_info "启动 17 个核心微服务..."
    docker-compose -f docker-compose.services.yml up -d

    print_info "等待服务启动..."
    sleep 15

    print_info "启动 2 个 BFF 服务..."
    docker-compose -f docker-compose.bff.yml up -d

    print_success "所有服务启动完成"
}

# 健康检查
health_check() {
    print_header "步骤 8/8: 健康检查"

    SERVICES=(
        "payment-gateway:40003"
        "order-service:40004"
        "channel-adapter:40005"
        "risk-service:40006"
        "accounting-service:40007"
        "notification-service:40008"
        "analytics-service:40009"
        "config-service:40010"
        "merchant-auth-service:40011"
        "settlement-service:40013"
        "withdrawal-service:40014"
        "kyc-service:40015"
        "cashier-service:40016"
        "reconciliation-service:40020"
        "dispute-service:40021"
        "merchant-policy-service:40022"
        "merchant-quota-service:40024"
        "admin-bff-service:40001"
        "merchant-bff-service:40023"
    )

    print_info "检查所有服务健康状态..."
    echo ""

    SUCCESS_COUNT=0
    FAILED_COUNT=0

    for svc in "${SERVICES[@]}"; do
        IFS=':' read -r name port <<< "$svc"

        # 等待服务启动
        sleep 1

        if curl -sf "http://localhost:$port/health" > /dev/null 2>&1; then
            print_success "✅ $name"
            ((SUCCESS_COUNT++))
        else
            print_error "❌ $name (端口: $port)"
            ((FAILED_COUNT++))
        fi
    done

    echo ""
    echo -e "${CYAN}健康检查结果:${NC}"
    echo -e "  成功: ${GREEN}$SUCCESS_COUNT${NC}"
    echo -e "  失败: ${RED}$FAILED_COUNT${NC}"

    if [ $FAILED_COUNT -gt 0 ]; then
        print_warning "部分服务未通过健康检查，请查看日志"
        print_info "查看日志命令: docker-compose -f docker-compose.services.yml logs -f <service-name>"
    fi
}

# 显示访问信息
show_access_info() {
    print_header "部署完成 🎉"

    echo -e "${CYAN}访问地址:${NC}"
    echo ""
    echo -e "  ${GREEN}核心服务:${NC}"
    echo "    Payment Gateway:  http://localhost:40003/health"
    echo "    Order Service:    http://localhost:40004/health"
    echo ""
    echo -e "  ${GREEN}BFF 服务:${NC}"
    echo "    Admin BFF:        http://localhost:40001/swagger/index.html"
    echo "    Merchant BFF:     http://localhost:40023/swagger/index.html"
    echo ""
    echo -e "  ${GREEN}监控仪表板:${NC}"
    echo "    Prometheus:       http://localhost:40090"
    echo "    Grafana:          http://localhost:40300 (admin/admin)"
    echo "    Jaeger:           http://localhost:50686"
    echo "    Kafka UI:         http://localhost:40084"
    echo ""
    echo -e "  ${GREEN}API 网关:${NC}"
    echo "    Kong Gateway:     http://localhost:40080"
    echo "    Konga UI:         http://localhost:50001"
    echo ""
    echo -e "${CYAN}常用命令:${NC}"
    echo "  查看所有容器:     docker ps"
    echo "  查看服务日志:     docker-compose -f docker-compose.services.yml logs -f <service>"
    echo "  停止所有服务:     cd $BASE_DIR && ./scripts/stop-all.sh"
    echo "  重启服务:         docker-compose -f docker-compose.services.yml restart <service>"
    echo ""
    echo -e "${GREEN}部署成功！祝您使用愉快！${NC}"
    echo ""
}

# 主函数
main() {
    clear
    print_header "🚀 支付平台一键部署工具"

    check_requirements
    generate_env_file
    generate_certificates
    start_infrastructure
    initialize_databases
    build_images
    start_services
    health_check
    show_access_info
}

# 错误处理
trap 'print_error "部署过程中发生错误，请检查日志"; exit 1' ERR

# 运行主函数
main "$@"

#!/bin/bash

#######################################
# One-Click Deployment Script
# 全自动部署整个支付平台系统
#######################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 日志函数
log_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

log_success() {
    echo -e "${GREEN}✅${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠️ ${NC} $1"
}

log_error() {
    echo -e "${RED}❌${NC} $1"
}

log_step() {
    echo ""
    echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${PURPLE}[$1/${TOTAL_STEPS}] $2${NC}"
    echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# 总步骤数
TOTAL_STEPS=8

# 清屏并显示标题
clear
echo -e "${BOLD}${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║       Global Payment Platform - One-Click Deployment        ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

#######################################
# Step 1: 环境检查
#######################################
log_step 1 "Environment Check"

log_info "Checking required tools..."

# Check Docker
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed"
    exit 1
fi
log_success "Docker: $(docker --version | cut -d' ' -f3)"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose is not installed"
    exit 1
fi
log_success "Docker Compose: $(docker-compose --version | cut -d' ' -f4)"

# Check Go
if ! command -v go &> /dev/null; then
    log_error "Go is not installed"
    exit 1
fi
log_success "Go: $(go version | cut -d' ' -f3)"

# Check Node.js
if ! command -v node &> /dev/null; then
    log_error "Node.js is not installed"
    exit 1
fi
log_success "Node.js: $(node --version)"

# Check npm
if ! command -v npm &> /dev/null; then
    log_error "npm is not installed"
    exit 1
fi
log_success "npm: $(npm --version)"

#######################################
# Step 2: 启动基础设施
#######################################
log_step 2 "Starting Infrastructure"

cd "$PROJECT_ROOT"

log_info "Starting PostgreSQL, Redis, Kafka, Prometheus, Grafana, Jaeger..."
docker-compose up -d

log_info "Waiting for infrastructure to be ready..."
sleep 10

# Check PostgreSQL
log_info "Checking PostgreSQL..."
for i in {1..30}; do
    if docker exec payment-postgres pg_isready -U postgres > /dev/null 2>&1; then
        log_success "PostgreSQL is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "PostgreSQL failed to start"
        exit 1
    fi
    sleep 2
done

# Check Redis
log_info "Checking Redis..."
for i in {1..30}; do
    if docker exec payment-redis redis-cli ping > /dev/null 2>&1; then
        log_success "Redis is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "Redis failed to start"
        exit 1
    fi
    sleep 2
done

# Check Kafka
log_info "Checking Kafka..."
for i in {1..60}; do
    if docker logs payment-kafka 2>&1 | grep -q "started (kafka.server.KafkaServer)"; then
        log_success "Kafka is ready"
        break
    fi
    if [ $i -eq 60 ]; then
        log_warning "Kafka may not be fully ready, continuing anyway..."
        break
    fi
    sleep 2
done

log_success "Infrastructure started successfully"

#######################################
# Step 3: 初始化数据库
#######################################
log_step 3 "Initializing Databases"

cd "$PROJECT_ROOT/backend"

log_info "Creating 19 databases..."
if [ -f "./scripts/init-db.sh" ]; then
    chmod +x ./scripts/init-db.sh
    ./scripts/init-db.sh
    log_success "Databases initialized"
else
    log_warning "init-db.sh not found, skipping database initialization"
fi

#######################################
# Step 4: 编译后端服务
#######################################
log_step 4 "Building Backend Services"

cd "$PROJECT_ROOT/backend"

log_info "Building all 19 microservices..."

# 创建 bin 目录
mkdir -p bin

# 编译所有服务
service_count=0
failed_services=()

for service_dir in services/*/; do
    service_name=$(basename "$service_dir")

    if [ ! -f "$service_dir/cmd/main.go" ]; then
        log_warning "Skipping $service_name (no main.go)"
        continue
    fi

    log_info "Building $service_name..."

    if GOWORK="$PROJECT_ROOT/backend/go.work" timeout 60 go build -o "bin/$service_name" "./$service_dir/cmd/main.go" 2>&1; then
        log_success "Built $service_name"
        service_count=$((service_count + 1))
    else
        log_error "Failed to build $service_name"
        failed_services+=("$service_name")
    fi
done

if [ ${#failed_services[@]} -gt 0 ]; then
    log_error "Failed to build ${#failed_services[@]} services: ${failed_services[*]}"
    log_warning "Continuing with deployment..."
else
    log_success "All $service_count services built successfully"
fi

#######################################
# Step 5: 启动后端服务
#######################################
log_step 5 "Starting Backend Services"

cd "$PROJECT_ROOT/backend"

if [ -f "./scripts/start-all-services.sh" ]; then
    log_info "Starting all backend services with hot reload..."
    chmod +x ./scripts/start-all-services.sh
    ./scripts/start-all-services.sh

    log_info "Waiting for services to start (30 seconds)..."
    sleep 30

    log_success "Backend services started"
else
    log_warning "start-all-services.sh not found, starting services manually..."

    # 手动启动服务
    for service_dir in services/*/; do
        service_name=$(basename "$service_dir")

        if [ -f "bin/$service_name" ]; then
            log_info "Starting $service_name..."
            nohup "./bin/$service_name" > "logs/$service_name.log" 2>&1 &
            echo $! > "logs/$service_name.pid"
        fi
    done

    sleep 30
fi

#######################################
# Step 6: 构建前端应用
#######################################
log_step 6 "Building Frontend Applications"

# Admin Portal
if [ -d "$PROJECT_ROOT/frontend/admin-portal" ]; then
    log_info "Building Admin Portal..."
    cd "$PROJECT_ROOT/frontend/admin-portal"

    if [ ! -d "node_modules" ]; then
        log_info "Installing Admin Portal dependencies..."
        npm install --quiet
    fi

    log_info "Building production bundle..."
    npm run build
    log_success "Admin Portal built (dist/)"
fi

# Merchant Portal
if [ -d "$PROJECT_ROOT/frontend/merchant-portal" ]; then
    log_info "Building Merchant Portal..."
    cd "$PROJECT_ROOT/frontend/merchant-portal"

    if [ ! -d "node_modules" ]; then
        log_info "Installing Merchant Portal dependencies..."
        npm install --quiet
    fi

    log_info "Building production bundle..."
    npm run build
    log_success "Merchant Portal built (dist/)"
fi

# Website
if [ -d "$PROJECT_ROOT/frontend/website" ]; then
    log_info "Building Website..."
    cd "$PROJECT_ROOT/frontend/website"

    if [ ! -d "node_modules" ]; then
        log_info "Installing Website dependencies..."
        npm install --quiet
    fi

    log_info "Building production bundle..."
    npm run build
    log_success "Website built (dist/)"
fi

#######################################
# Step 7: 健康检查
#######################################
log_step 7 "Health Check"

cd "$PROJECT_ROOT/backend"

log_info "Checking service health..."

# 定义服务端口
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
)

healthy_count=0
unhealthy_services=()

for service in "${!SERVICE_PORTS[@]}"; do
    port=${SERVICE_PORTS[$service]}

    if lsof -i :$port -sTCP:LISTEN >/dev/null 2>&1; then
        if curl -s -f http://localhost:$port/health >/dev/null 2>&1; then
            log_success "$service is healthy (port $port)"
            healthy_count=$((healthy_count + 1))
        else
            log_warning "$service is running but health check failed (port $port)"
            unhealthy_services+=("$service")
        fi
    else
        log_error "$service is not running (port $port)"
        unhealthy_services+=("$service")
    fi
done

if [ $healthy_count -eq ${#SERVICE_PORTS[@]} ]; then
    log_success "All ${#SERVICE_PORTS[@]} core services are healthy"
else
    log_warning "$healthy_count/${#SERVICE_PORTS[@]} services healthy"
    if [ ${#unhealthy_services[@]} -gt 0 ]; then
        log_warning "Unhealthy services: ${unhealthy_services[*]}"
    fi
fi

#######################################
# Step 8: 显示访问信息
#######################################
log_step 8 "Deployment Complete"

echo ""
echo -e "${BOLD}${GREEN}🎉 Deployment completed successfully!${NC}"
echo ""
echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}Access Information:${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BOLD}Frontend Applications:${NC}"
echo -e "  • Admin Portal:      ${CYAN}http://localhost:5173${NC}"
echo -e "  • Merchant Portal:   ${CYAN}http://localhost:5174${NC}"
echo -e "  • Website:           ${CYAN}http://localhost:5175${NC}"
echo ""
echo -e "${BOLD}Backend Services:${NC}"
echo -e "  • Admin Service:     ${CYAN}http://localhost:40001${NC}"
echo -e "  • Merchant Service:  ${CYAN}http://localhost:40002${NC}"
echo -e "  • Payment Gateway:   ${CYAN}http://localhost:40003${NC}"
echo ""
echo -e "${BOLD}API Documentation:${NC}"
echo -e "  • Admin Swagger:     ${CYAN}http://localhost:40001/swagger/index.html${NC}"
echo -e "  • Merchant Swagger:  ${CYAN}http://localhost:40002/swagger/index.html${NC}"
echo -e "  • Gateway Swagger:   ${CYAN}http://localhost:40003/swagger/index.html${NC}"
echo ""
echo -e "${BOLD}Monitoring & Observability:${NC}"
echo -e "  • Grafana:           ${CYAN}http://localhost:40300${NC} ${YELLOW}(admin/admin)${NC}"
echo -e "  • Prometheus:        ${CYAN}http://localhost:40090${NC}"
echo -e "  • Jaeger UI:         ${CYAN}http://localhost:40686${NC}"
echo ""
echo -e "${BOLD}Infrastructure:${NC}"
echo -e "  • PostgreSQL:        ${CYAN}localhost:40432${NC} ${YELLOW}(postgres/postgres)${NC}"
echo -e "  • Redis:             ${CYAN}localhost:40379${NC}"
echo -e "  • Kafka:             ${CYAN}localhost:40092${NC}"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}Useful Commands:${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  ${YELLOW}# View system status${NC}"
echo -e "  cd backend && ./scripts/system-status-dashboard.sh"
echo ""
echo -e "  ${YELLOW}# View service dependencies${NC}"
echo -e "  cd backend && ./scripts/service-dependency-map.sh"
echo ""
echo -e "  ${YELLOW}# View service logs${NC}"
echo -e "  tail -f backend/logs/payment-gateway.log"
echo ""
echo -e "  ${YELLOW}# Stop all services${NC}"
echo -e "  cd backend && ./scripts/stop-all-services.sh"
echo ""
echo -e "  ${YELLOW}# Stop infrastructure${NC}"
echo -e "  docker-compose down"
echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BOLD}${GREEN}Ready for testing! 🚀${NC}"
echo ""

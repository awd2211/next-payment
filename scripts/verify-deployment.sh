#!/bin/bash

# ============================================================================
# 验证 Docker 部署完整性
# ============================================================================

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

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
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# 统计
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

check_pass() {
    ((TOTAL_CHECKS++))
    ((PASSED_CHECKS++))
    print_success "$1"
}

check_fail() {
    ((TOTAL_CHECKS++))
    ((FAILED_CHECKS++))
    print_error "$1"
}

print_header "🔍 Docker 部署验证工具"

# ============================================================================
# 1. 检查 Docker Compose 文件
# ============================================================================
print_header "1/6: 检查 Docker Compose 文件"

if [ -f "docker-compose.yml" ]; then
    check_pass "docker-compose.yml 存在"
else
    check_fail "docker-compose.yml 不存在"
fi

if [ -f "docker-compose.services.yml" ]; then
    check_pass "docker-compose.services.yml 存在"
else
    check_fail "docker-compose.services.yml 不存在"
fi

if [ -f "docker-compose.bff.yml" ]; then
    check_pass "docker-compose.bff.yml 存在"
else
    check_fail "docker-compose.bff.yml 不存在"
fi

# ============================================================================
# 2. 检查 Dockerfile
# ============================================================================
print_header "2/6: 检查 Dockerfile 文件"

SERVICES=(
    "payment-gateway" "order-service" "channel-adapter" "risk-service"
    "accounting-service" "notification-service" "analytics-service"
    "config-service" "merchant-auth-service" "settlement-service"
    "withdrawal-service" "kyc-service" "cashier-service"
    "reconciliation-service" "dispute-service" "merchant-policy-service"
    "merchant-quota-service" "admin-bff-service" "merchant-bff-service"
)

DOCKERFILE_MISSING=0

for service in "${SERVICES[@]}"; do
    if [ -f "backend/services/$service/Dockerfile" ]; then
        print_success "$service/Dockerfile"
    else
        print_error "$service/Dockerfile 缺失"
        ((DOCKERFILE_MISSING++))
    fi
done

if [ $DOCKERFILE_MISSING -eq 0 ]; then
    check_pass "所有 Dockerfile 完整"
else
    check_fail "$DOCKERFILE_MISSING 个 Dockerfile 缺失"
fi

# ============================================================================
# 3. 检查 mTLS 证书
# ============================================================================
print_header "3/6: 检查 mTLS 证书"

if [ -f "backend/certs/ca/ca-cert.pem" ]; then
    check_pass "CA 证书存在"
else
    check_fail "CA 证书缺失"
fi

CERT_MISSING=0

for service in "${SERVICES[@]}"; do
    cert_file="backend/certs/services/$service/${service}.crt"
    key_file="backend/certs/services/$service/${service}.key"

    if [ -f "$cert_file" ] && [ -f "$key_file" ]; then
        print_success "$service 证书"
    else
        print_error "$service 证书缺失"
        ((CERT_MISSING++))
    fi
done

if [ $CERT_MISSING -eq 0 ]; then
    check_pass "所有服务证书完整"
else
    check_fail "$CERT_MISSING 个服务证书缺失"
fi

# ============================================================================
# 4. 检查环境变量文件
# ============================================================================
print_header "4/6: 检查环境变量文件"

if [ -f ".env" ]; then
    check_pass ".env 文件存在"

    # 检查关键变量
    if grep -q "JWT_SECRET" .env; then
        check_pass "JWT_SECRET 已配置"
    else
        check_fail "JWT_SECRET 未配置"
    fi

    if grep -q "DB_PASSWORD" .env; then
        check_pass "DB_PASSWORD 已配置"
    else
        check_fail "DB_PASSWORD 未配置"
    fi
else
    check_fail ".env 文件不存在"
fi

# ============================================================================
# 5. 检查容器运行状态
# ============================================================================
print_header "5/6: 检查容器运行状态"

# 基础设施
INFRA_CONTAINERS=(
    "payment-postgres"
    "payment-redis"
    "payment-kafka"
    "payment-prometheus"
    "payment-grafana"
    "payment-jaeger"
)

print_info "基础设施容器:"
for container in "${INFRA_CONTAINERS[@]}"; do
    if docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
        check_pass "$container 运行中"
    else
        check_fail "$container 未运行"
    fi
done

# 微服务
print_info "微服务容器:"
SERVICE_CONTAINERS=(
    "payment-payment-gateway"
    "payment-order-service"
    "payment-channel-adapter"
    "payment-risk-service"
    "payment-accounting-service"
    "payment-notification-service"
    "payment-analytics-service"
    "payment-config-service"
)

for container in "${SERVICE_CONTAINERS[@]}"; do
    if docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
        check_pass "$container 运行中"
    else
        check_fail "$container 未运行"
    fi
done

# BFF 服务
print_info "BFF 服务:"
BFF_CONTAINERS=(
    "payment-admin-bff"
    "payment-merchant-bff"
)

for container in "${BFF_CONTAINERS[@]}"; do
    if docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
        check_pass "$container 运行中"
    else
        check_fail "$container 未运行"
    fi
done

# ============================================================================
# 6. 健康检查
# ============================================================================
print_header "6/6: 服务健康检查"

HEALTH_CHECKS=(
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

for svc in "${HEALTH_CHECKS[@]}"; do
    IFS=':' read -r name port <<< "$svc"

    if curl -sf "http://localhost:$port/health" >/dev/null 2>&1; then
        check_pass "$name (http://localhost:$port/health)"
    else
        check_fail "$name (http://localhost:$port/health) 不健康"
    fi
done

# ============================================================================
# 总结
# ============================================================================
print_header "验证结果"

echo -e "${CYAN}总检查项:${NC}   $TOTAL_CHECKS"
echo -e "${GREEN}通过:${NC}       $PASSED_CHECKS"
echo -e "${RED}失败:${NC}       $FAILED_CHECKS"
echo -e "${CYAN}成功率:${NC}     $(awk "BEGIN {printf \"%.1f%%\", ($PASSED_CHECKS/$TOTAL_CHECKS)*100}")"
echo ""

if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "${GREEN}🎉 所有检查通过！部署完全成功！${NC}"
    echo ""
    echo -e "${CYAN}访问地址:${NC}"
    echo "  Admin BFF:        http://localhost:40001/swagger/index.html"
    echo "  Merchant BFF:     http://localhost:40023/swagger/index.html"
    echo "  Prometheus:       http://localhost:40090"
    echo "  Grafana:          http://localhost:40300 (admin/admin)"
    echo "  Jaeger:           http://localhost:50686"
    exit 0
else
    echo -e "${RED}❌ 部分检查失败，请检查日志${NC}"
    echo ""
    echo -e "${CYAN}故障排查:${NC}"
    echo "  查看容器日志:     docker logs <container-name>"
    echo "  查看所有容器:     docker ps -a"
    echo "  重启服务:         docker-compose restart <service>"
    echo "  查看部署指南:     cat DOCKER_DEPLOYMENT_GUIDE.md"
    exit 1
fi

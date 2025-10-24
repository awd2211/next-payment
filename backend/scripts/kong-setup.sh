#!/bin/bash
set -e

# Kong API Gateway 配置脚本
# 用途: 自动配置服务、路由和插件

KONG_ADMIN="http://localhost:40081"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# 等待 Kong 启动
wait_for_kong() {
    log_info "等待 Kong Gateway 启动..."
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f $KONG_ADMIN/ > /dev/null 2>&1; then
            log_success "Kong Gateway 已就绪"
            return 0
        fi
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done

    log_error "Kong Gateway 启动超时"
    return 1
}

# 创建或更新服务
create_or_update_service() {
    local name=$1
    local url=$2

    log_info "配置服务: $name"

    # 检查服务是否存在
    if curl -s -f $KONG_ADMIN/services/$name > /dev/null 2>&1; then
        # 更新现有服务
        curl -s -X PATCH $KONG_ADMIN/services/$name \
            --data "url=$url" \
            > /dev/null
        log_success "服务 $name 已更新"
    else
        # 创建新服务
        curl -s -X POST $KONG_ADMIN/services \
            --data "name=$name" \
            --data "url=$url" \
            --data "connect_timeout=60000" \
            --data "write_timeout=60000" \
            --data "read_timeout=60000" \
            --data "retries=5" \
            > /dev/null
        log_success "服务 $name 已创建"
    fi
}

# 创建或更新路由
create_or_update_route() {
    local service_name=$1
    local route_name=$2
    shift 2
    local paths=("$@")

    log_info "配置路由: $route_name"

    # 检查路由是否存在
    if curl -s -f $KONG_ADMIN/routes/$route_name > /dev/null 2>&1; then
        log_warning "路由 $route_name 已存在,跳过"
        return 0
    fi

    # 构建路径参数
    local path_params=""
    for path in "${paths[@]}"; do
        path_params="${path_params} --data paths[]=$path"
    done

    # 创建路由
    curl -s -X POST $KONG_ADMIN/services/$service_name/routes \
        --data "name=$route_name" \
        $path_params \
        --data "strip_path=false" \
        --data "preserve_host=false" \
        > /dev/null

    log_success "路由 $route_name 已创建"
}

# 启用插件
enable_plugin() {
    local plugin_name=$1
    local target_type=$2  # service, route, 或 global
    local target_name=$3
    shift 3
    local config_params="$@"

    log_info "启用插件: $plugin_name ($target_type)"

    local url=""
    if [ "$target_type" == "global" ]; then
        url="$KONG_ADMIN/plugins"
    elif [ "$target_type" == "service" ]; then
        url="$KONG_ADMIN/services/$target_name/plugins"
    elif [ "$target_type" == "route" ]; then
        url="$KONG_ADMIN/routes/$target_name/plugins"
    fi

    curl -s -X POST "$url" \
        --data "name=$plugin_name" \
        $config_params \
        > /dev/null

    log_success "插件 $plugin_name 已启用"
}

echo ""
echo "=========================================="
echo "  Kong API Gateway 配置工具"
echo "=========================================="
echo ""

# 1. 等待 Kong 启动
wait_for_kong || exit 1

echo ""
log_info "开始配置服务..."
echo ""

# 2. 创建服务
create_or_update_service "admin-service" "http://host.docker.internal:40001"
create_or_update_service "merchant-service" "http://host.docker.internal:40002"
create_or_update_service "payment-gateway" "http://host.docker.internal:40003"
create_or_update_service "order-service" "http://host.docker.internal:40004"
create_or_update_service "channel-adapter" "http://host.docker.internal:40005"
create_or_update_service "risk-service" "http://host.docker.internal:40006"
create_or_update_service "accounting-service" "http://host.docker.internal:40007"
create_or_update_service "notification-service" "http://host.docker.internal:40008"
create_or_update_service "analytics-service" "http://host.docker.internal:40009"
create_or_update_service "config-service" "http://host.docker.internal:40010"

echo ""
log_info "开始配置路由..."
echo ""

# 3. 创建路由 - Admin Service (JWT 认证)
create_or_update_route "admin-service" "admin-auth" \
    "/api/v1/admin/login"

create_or_update_route "admin-service" "admin-management" \
    "/api/v1/admin" \
    "/api/v1/roles" \
    "/api/v1/permissions" \
    "/api/v1/audit-logs" \
    "/api/v1/security" \
    "/api/v1/preferences" \
    "/api/v1/email-templates"

# 4. 创建路由 - Merchant Service (JWT + Public)
create_or_update_route "merchant-service" "merchant-public" \
    "/api/v1/merchant/register" \
    "/api/v1/merchant/login"

create_or_update_route "merchant-service" "merchant-dashboard" \
    "/api/v1/merchant/profile" \
    "/api/v1/dashboard"

create_or_update_route "merchant-service" "merchant-admin" \
    "/api/v1/merchant"

# 5. 创建路由 - Payment Gateway (API Key 认证)
create_or_update_route "payment-gateway" "payment-api" \
    "/api/v1/payments" \
    "/api/v1/refunds"

create_or_update_route "payment-gateway" "payment-webhooks" \
    "/api/v1/webhooks"

# 6. 创建路由 - Config Service
create_or_update_route "config-service" "config-api" \
    "/api/v1/config"

echo ""
log_info "开始配置全局插件..."
echo ""

# 7. 启用全局插件

# CORS 插件
enable_plugin "cors" "global" "" \
    --data "config.origins=http://localhost:5173" \
    --data "config.origins=http://localhost:5174" \
    --data "config.origins=http://localhost:5175" \
    --data "config.methods=GET" \
    --data "config.methods=POST" \
    --data "config.methods=PUT" \
    --data "config.methods=DELETE" \
    --data "config.methods=PATCH" \
    --data "config.methods=OPTIONS" \
    --data "config.headers=Authorization" \
    --data "config.headers=X-API-Key" \
    --data "config.headers=X-Signature" \
    --data "config.headers=X-Timestamp" \
    --data "config.headers=X-Nonce" \
    --data "config.headers=Idempotency-Key" \
    --data "config.headers=Content-Type" \
    --data "config.credentials=true" \
    --data "config.max_age=3600"

# Request ID 插件
enable_plugin "correlation-id" "global" "" \
    --data "config.header_name=X-Request-ID" \
    --data "config.generator=uuid" \
    --data "config.echo_downstream=true"

# Prometheus 插件
enable_plugin "prometheus" "global" ""

# Request Size Limiting 插件
enable_plugin "request-size-limiting" "global" "" \
    --data "config.allowed_payload_size=10"

echo ""
log_info "开始配置路由级别插件..."
echo ""

# 8. 为 Payment API 启用 Key Auth 和 Rate Limiting
enable_plugin "key-auth" "route" "payment-api" \
    --data "config.key_names=X-API-Key" \
    --data "config.hide_credentials=true"

enable_plugin "rate-limiting" "route" "payment-api" \
    --data "config.minute=100" \
    --data "config.policy=redis" \
    --data "config.redis.host=redis" \
    --data "config.redis.port=6379" \
    --data "config.redis.timeout=2000" \
    --data "config.hide_client_headers=false"

# 9. 为 Admin Management 启用 JWT
enable_plugin "jwt" "route" "admin-management" \
    --data "config.key_claim_name=iss" \
    --data "config.claims_to_verify=exp"

# 10. 为 Merchant Dashboard 启用 JWT
enable_plugin "jwt" "route" "merchant-dashboard" \
    --data "config.key_claim_name=iss" \
    --data "config.claims_to_verify=exp"

enable_plugin "jwt" "route" "merchant-admin" \
    --data "config.key_claim_name=iss" \
    --data "config.claims_to_verify=exp"

# 11. 为公开路由启用 Rate Limiting (防止暴力攻击)
enable_plugin "rate-limiting" "route" "admin-auth" \
    --data "config.minute=10" \
    --data "config.policy=local"

enable_plugin "rate-limiting" "route" "merchant-public" \
    --data "config.minute=10" \
    --data "config.policy=local"

echo ""
log_success "Kong 配置完成!"
echo ""
echo "=========================================="
echo "  Kong Gateway 访问信息"
echo "=========================================="
echo ""
echo "  🌐 Kong Proxy (API Gateway): http://localhost:40080"
echo "  ⚙️  Kong Admin API:          http://localhost:40081"
echo "  🎨 Konga Admin UI:           http://localhost:40082"
echo ""
echo "  示例 API 调用:"
echo "  - Admin Login:   POST http://localhost:40080/api/v1/admin/login"
echo "  - Merchant Login: POST http://localhost:40080/api/v1/merchant/login"
echo "  - Create Payment: POST http://localhost:40080/api/v1/payments (需要 X-API-Key)"
echo ""
echo "=========================================="
echo ""

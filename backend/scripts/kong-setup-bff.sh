#!/bin/bash
set -e

# Kong API Gateway 配置脚本 (BFF版本)
# 用途: 配置 Admin BFF 和 Merchant BFF 的路由

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

    # 检查路由是否存在,如果存在则删除旧路由
    if curl -s -f $KONG_ADMIN/routes/$route_name > /dev/null 2>&1; then
        curl -s -X DELETE $KONG_ADMIN/routes/$route_name > /dev/null 2>&1
        log_warning "删除旧路由 $route_name"
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

    log_info "启用插件: $plugin_name ($target_type: $target_name)"

    local url=""
    if [ "$target_type" == "global" ]; then
        url="$KONG_ADMIN/plugins"
    elif [ "$target_type" == "service" ]; then
        url="$KONG_ADMIN/services/$target_name/plugins"
    elif [ "$target_type" == "route" ]; then
        url="$KONG_ADMIN/routes/$target_name/plugins"
    fi

    # 检查插件是否已存在
    local existing_plugin=$(curl -s "$url" | grep -o "\"name\":\"$plugin_name\"" || true)
    if [ -n "$existing_plugin" ]; then
        log_warning "插件 $plugin_name 已存在,跳过"
        return 0
    fi

    curl -s -X POST "$url" \
        --data "name=$plugin_name" \
        $config_params \
        > /dev/null

    log_success "插件 $plugin_name 已启用"
}

echo ""
echo "=========================================="
echo "  Kong API Gateway BFF 配置工具"
echo "=========================================="
echo ""

# 1. 等待 Kong 启动
wait_for_kong || exit 1

echo ""
log_info "开始配置 BFF 服务..."
echo ""

# 2. 创建 BFF 服务
create_or_update_service "admin-bff-service" "http://host.docker.internal:40001"
create_or_update_service "merchant-bff-service" "http://host.docker.internal:40023"

echo ""
log_info "开始配置 BFF 路由..."
echo ""

# 3. Admin BFF 路由 - 所有 /api/v1/admin/* 路径
create_or_update_route "admin-bff-service" "admin-bff-routes" \
    "/api/v1/admin"

# 4. Merchant BFF 路由 - 所有 /api/v1/merchant/* 路径
create_or_update_route "merchant-bff-service" "merchant-bff-routes" \
    "/api/v1/merchant"

echo ""
log_info "开始配置 BFF 插件..."
echo ""

# 5. Admin BFF 启用 JWT (除登录外)
enable_plugin "jwt" "route" "admin-bff-routes" \
    --data "config.key_claim_name=iss" \
    --data "config.claims_to_verify=exp"

# 6. Merchant BFF 启用 JWT (除登录/注册外)
enable_plugin "jwt" "route" "merchant-bff-routes" \
    --data "config.key_claim_name=iss" \
    --data "config.claims_to_verify=exp"

# 7. 启用速率限制
enable_plugin "rate-limiting" "route" "admin-bff-routes" \
    --data "config.minute=60" \
    --data "config.policy=local"

enable_plugin "rate-limiting" "route" "merchant-bff-routes" \
    --data "config.minute=300" \
    --data "config.policy=local"

echo ""
log_success "Kong BFF 配置完成!"
echo ""
echo "=========================================="
echo "  Kong BFF 路由信息"
echo "=========================================="
echo ""
echo "  📱 Admin Portal → Kong Proxy → admin-bff-service"
echo "     http://localhost:40080/api/v1/admin/*"
echo ""
echo "  📱 Merchant Portal → Kong Proxy → merchant-bff-service"
echo "     http://localhost:40080/api/v1/merchant/*"
echo ""
echo "  示例 API 调用:"
echo "  - Admin Login:   POST http://localhost:40080/api/v1/admin/login"
echo "  - Admin KYC Docs: GET http://localhost:40080/api/v1/admin/kyc/documents"
echo "  - Merchant Orders: GET http://localhost:40080/api/v1/merchant/orders"
echo ""
echo "=========================================="
echo ""

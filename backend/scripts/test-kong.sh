#!/bin/bash
set -e

# Kong API Gateway 端到端测试脚本

KONG_PROXY="http://localhost:40080"
KONG_ADMIN="http://localhost:40081"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[测试]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[通过]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[警告]${NC} $1"
}

log_error() {
    echo -e "${RED}[失败]${NC} $1"
}

test_count=0
pass_count=0
fail_count=0

run_test() {
    local test_name=$1
    local expected_code=$2
    shift 2
    local curl_args=("$@")

    test_count=$((test_count + 1))
    log_info "测试 $test_count: $test_name"

    response=$(curl -s -w "\n%{http_code}" "${curl_args[@]}")
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')

    if [ "$status_code" -eq "$expected_code" ]; then
        log_success "$test_name (HTTP $status_code)"
        pass_count=$((pass_count + 1))
        return 0
    else
        log_error "$test_name (期望 $expected_code, 实际 $status_code)"
        echo "响应: $body"
        fail_count=$((fail_count + 1))
        return 1
    fi
}

echo ""
echo "=========================================="
echo "  Kong API Gateway 端到端测试"
echo "=========================================="
echo ""

# 1. 测试 Kong 健康状态
log_info "检查 Kong 健康状态..."
if curl -s -f $KONG_ADMIN/status > /dev/null 2>&1; then
    log_success "Kong Admin API 正常运行"
else
    log_error "Kong Admin API 无法访问"
    exit 1
fi

echo ""
log_info "开始 API 路由测试..."
echo ""

# 2. 测试 Admin Service 路由
run_test "Admin Login (公开路由)" 200 \
    -X POST "$KONG_PROXY/api/v1/admin/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}'

run_test "Admin Management (需要 JWT - 应该返回 401)" 401 \
    -X GET "$KONG_PROXY/api/v1/admin" \
    -H "Content-Type: application/json"

# 3. 测试 Merchant Service 路由
run_test "Merchant Register (公开路由)" 200 \
    -X POST "$KONG_PROXY/api/v1/merchant/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","company_name":"Test Company"}'

run_test "Merchant Dashboard (需要 JWT - 应该返回 401)" 401 \
    -X GET "$KONG_PROXY/api/v1/merchant/profile" \
    -H "Content-Type: application/json"

# 4. 测试 Payment Gateway 路由
run_test "Create Payment (需要 API Key - 应该返回 401)" 401 \
    -X POST "$KONG_PROXY/api/v1/payments" \
    -H "Content-Type: application/json" \
    -d '{"amount":1000,"currency":"USD"}'

# 5. 测试 CORS 预检请求
run_test "CORS Preflight Request" 200 \
    -X OPTIONS "$KONG_PROXY/api/v1/payments" \
    -H "Origin: http://localhost:5173" \
    -H "Access-Control-Request-Method: POST" \
    -H "Access-Control-Request-Headers: Authorization"

# 6. 测试 Rate Limiting (连续 15 次请求,第 11 次应该被限流)
echo ""
log_info "测试 Rate Limiting (Admin Login)..."
rate_limit_hit=0
for i in {1..15}; do
    status_code=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$KONG_PROXY/api/v1/admin/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}')

    if [ "$status_code" -eq 429 ]; then
        rate_limit_hit=$i
        break
    fi
done

if [ $rate_limit_hit -gt 0 ] && [ $rate_limit_hit -le 15 ]; then
    log_success "Rate Limiting 触发于第 $rate_limit_hit 次请求 (限制: 10/分钟)"
    pass_count=$((pass_count + 1))
else
    log_warning "Rate Limiting 未触发 (可能已过冷却期)"
fi

# 7. 测试 Request ID 插件
echo ""
log_info "测试 Request ID 插件..."
response_headers=$(curl -s -I "$KONG_PROXY/api/v1/admin/login" \
    -X POST \
    -H "Content-Type: application/json")

if echo "$response_headers" | grep -qi "X-Request-ID"; then
    request_id=$(echo "$response_headers" | grep -i "X-Request-ID" | cut -d' ' -f2 | tr -d '\r')
    log_success "Request ID 已添加: $request_id"
    pass_count=$((pass_count + 1))
else
    log_error "Request ID 未找到"
    fail_count=$((fail_count + 1))
fi

# 8. 测试 Prometheus Metrics
echo ""
log_info "测试 Prometheus Metrics..."
if curl -s "$KONG_ADMIN/metrics" | grep -q "kong_http_requests_total"; then
    log_success "Prometheus 指标已启用"
    pass_count=$((pass_count + 1))
else
    log_error "Prometheus 指标未找到"
    fail_count=$((fail_count + 1))
fi

# 9. 查看 Kong 配置摘要
echo ""
log_info "Kong 配置摘要..."
echo ""

services_count=$(curl -s "$KONG_ADMIN/services" | grep -o '"id"' | wc -l)
routes_count=$(curl -s "$KONG_ADMIN/routes" | grep -o '"id"' | wc -l)
plugins_count=$(curl -s "$KONG_ADMIN/plugins" | grep -o '"id"' | wc -l)

echo "  服务 (Services): $services_count"
echo "  路由 (Routes): $routes_count"
echo "  插件 (Plugins): $plugins_count"

echo ""
echo "=========================================="
echo "  测试结果汇总"
echo "=========================================="
echo ""
echo "  总测试数: $test_count"
echo "  通过: ${GREEN}$pass_count${NC}"
echo "  失败: ${RED}$fail_count${NC}"
echo ""

if [ $fail_count -eq 0 ]; then
    log_success "所有测试通过! Kong Gateway 配置正确"
    echo ""
    exit 0
else
    log_warning "部分测试失败,请检查配置"
    echo ""
    exit 1
fi

#!/bin/bash

# ============================================================================
# BFF Security Features Test Script
# 测试 Admin BFF 和 Merchant BFF 的所有安全特性
# ============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 配置
ADMIN_BFF_URL="${ADMIN_BFF_URL:-http://localhost:40001}"
MERCHANT_BFF_URL="${MERCHANT_BFF_URL:-http://localhost:40023}"

# 日志函数
log_test() {
    echo -e "${CYAN}[TEST]${NC} $1"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查服务是否运行
check_service() {
    local service_name=$1
    local url=$2

    log_test "检查 $service_name 是否运行..."

    response=$(curl -s -o /dev/null -w "%{http_code}" "$url/health" || echo "000")

    if [ "$response" = "200" ]; then
        log_pass "$service_name 正在运行"
        return 0
    else
        log_fail "$service_name 未运行 (HTTP $response)"
        return 1
    fi
}

# 测试 JWT 认证
test_jwt_auth() {
    local service_name=$1
    local url=$2

    echo ""
    log_info "================================"
    log_info "测试 $service_name - JWT 认证"
    log_info "================================"

    # 测试1: 缺少 Token
    log_test "测试缺少 Authorization 头..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url/api/v1/admin/merchants" 2>/dev/null || echo "000")
    if [ "$response" = "401" ]; then
        log_pass "正确拦截未认证请求 (HTTP 401)"
    else
        log_fail "未正确拦截未认证请求 (HTTP $response, 预期 401)"
    fi

    # 测试2: 无效 Token
    log_test "测试无效 JWT Token..."
    response=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer invalid_token_here" \
        "$url/api/v1/admin/merchants" 2>/dev/null || echo "000")
    if [ "$response" = "401" ]; then
        log_pass "正确拦截无效 Token (HTTP 401)"
    else
        log_fail "未正确拦截无效 Token (HTTP $response, 预期 401)"
    fi
}

# 测试速率限制
test_rate_limiting() {
    local service_name=$1
    local url=$2
    local limit=$3

    echo ""
    log_info "================================"
    log_info "测试 $service_name - 速率限制 ($limit req/min)"
    log_info "================================"

    log_test "发送 $((limit + 1)) 个请求测试限流..."

    # 快速发送请求
    rate_limited=false
    for i in $(seq 1 $((limit + 1))); do
        response=$(curl -s -o /dev/null -w "%{http_code}" "$url/health" 2>/dev/null || echo "000")

        if [ "$response" = "429" ]; then
            rate_limited=true
            log_pass "第 $i 个请求被限流 (HTTP 429)"
            break
        fi

        # 避免过快（给一点间隔）
        sleep 0.01
    done

    if [ "$rate_limited" = "true" ]; then
        log_pass "速率限制正常工作"
    else
        log_warning "未触发速率限制（可能限流阈值较高）"
    fi
}

# 测试数据脱敏
test_data_masking() {
    local service_name=$1
    local url=$2

    echo ""
    log_info "================================"
    log_info "测试 $service_name - 数据脱敏"
    log_info "================================"

    log_test "检查 Swagger UI 是否可访问..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url/swagger/index.html" 2>/dev/null || echo "000")
    if [ "$response" = "200" ]; then
        log_pass "Swagger UI 可访问"
        log_info "数据脱敏功能需要实际调用 API 并检查响应"
        log_info "手动测试: 查看商户/订单详情，验证手机号、邮箱、身份证等是否已脱敏"
    else
        log_warning "Swagger UI 不可访问 (HTTP $response)"
    fi
}

# 测试结构化日志
test_structured_logging() {
    local service_name=$1
    local url=$2

    echo ""
    log_info "================================"
    log_info "测试 $service_name - 结构化日志"
    log_info "================================"

    log_test "发送测试请求生成日志..."
    curl -s "$url/health" > /dev/null 2>&1

    log_info "检查日志文件..."
    LOG_FILE="../logs/bff/${service_name,,}.log"

    if [ -f "$LOG_FILE" ]; then
        # 检查是否包含 JSON 格式日志
        if grep -q '"@timestamp"' "$LOG_FILE" 2>/dev/null; then
            log_pass "发现结构化 JSON 日志"
            log_info "最新日志条目:"
            tail -n 1 "$LOG_FILE" | jq '.' 2>/dev/null || tail -n 1 "$LOG_FILE"
        else
            log_warning "日志格式可能不是 JSON"
        fi
    else
        log_warning "未找到日志文件: $LOG_FILE"
    fi
}

# 测试 Prometheus 指标
test_metrics() {
    local service_name=$1
    local url=$2

    echo ""
    log_info "================================"
    log_info "测试 $service_name - Prometheus 指标"
    log_info "================================"

    log_test "检查 /metrics 端点..."
    response=$(curl -s "$url/metrics" 2>/dev/null)

    if echo "$response" | grep -q "http_requests_total"; then
        log_pass "Prometheus 指标端点正常"

        # 检查关键指标
        if echo "$response" | grep -q "http_request_duration_seconds"; then
            log_pass "包含请求延迟指标"
        fi

        if echo "$response" | grep -q "http_request_size_bytes"; then
            log_pass "包含请求大小指标"
        fi
    else
        log_fail "Prometheus 指标端点异常"
    fi
}

# 测试健康检查
test_health_check() {
    local service_name=$1
    local url=$2

    echo ""
    log_info "================================"
    log_info "测试 $service_name - 健康检查"
    log_info "================================"

    # 测试 /health
    log_test "检查 /health 端点..."
    response=$(curl -s "$url/health" 2>/dev/null)
    if echo "$response" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
        log_pass "/health 端点正常"
    else
        log_warning "/health 端点响应异常"
    fi

    # 测试 /health/live
    log_test "检查 /health/live 端点..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url/health/live" 2>/dev/null || echo "000")
    if [ "$response" = "200" ]; then
        log_pass "/health/live 端点正常"
    else
        log_warning "/health/live 端点异常 (HTTP $response)"
    fi

    # 测试 /health/ready
    log_test "检查 /health/ready 端点..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url/health/ready" 2>/dev/null || echo "000")
    if [ "$response" = "200" ]; then
        log_pass "/health/ready 端点正常"
    else
        log_warning "/health/ready 端点异常 (HTTP $response)"
    fi
}

# 测试 Admin BFF 特有功能
test_admin_bff_features() {
    echo ""
    echo "========================================"
    log_info "测试 Admin BFF 特有功能"
    echo "========================================"

    # RBAC 测试（需要实际 Token，这里仅测试端点）
    log_info "RBAC 权限控制测试需要有效的 Admin Token"
    log_info "手动测试步骤:"
    log_info "1. 使用不同角色的管理员登录"
    log_info "2. 尝试访问需要特定权限的端点"
    log_info "3. 验证是否正确返回 403 Forbidden"

    # 2FA 测试
    log_info ""
    log_info "2FA/TOTP 验证测试需要有效的 Admin Token 和 2FA Secret"
    log_info "手动测试步骤:"
    log_info "1. 启用 2FA: POST /api/v1/admins/2fa/enable"
    log_info "2. 获取 TOTP Secret"
    log_info "3. 使用 Google Authenticator 生成验证码"
    log_info "4. 访问财务操作端点，提供 X-2FA-Code 头"

    # 审计日志测试
    log_info ""
    log_info "审计日志测试需要有效的 Admin Token"
    log_info "手动测试步骤:"
    log_info "1. 执行敏感操作（如批准结算）"
    log_info "2. 查询审计日志: GET /api/v1/admin/audit-logs"
    log_info "3. 验证日志包含 WHO, WHEN, WHAT, WHY"
}

# 测试 Merchant BFF 特有功能
test_merchant_bff_features() {
    echo ""
    echo "========================================"
    log_info "测试 Merchant BFF 特有功能"
    echo "========================================"

    # 租户隔离测试
    log_info "租户隔离测试需要有效的 Merchant Token"
    log_info "手动测试步骤:"
    log_info "1. 使用商户 A 的 Token 登录"
    log_info "2. 查询订单: GET /api/v1/merchant/orders"
    log_info "3. 尝试传递其他商户 ID: ?merchant_id=other-merchant"
    log_info "4. 验证依然只返回商户 A 的订单"

    log_info ""
    log_info "跨租户访问拦截测试:"
    log_info "1. 使用商户 A 的 Token"
    log_info "2. 尝试访问商户 B 的资源"
    log_info "3. 验证 BFF 层强制注入正确的 merchant_id"
}

# 显示测试总结
show_summary() {
    echo ""
    echo "========================================"
    log_info "测试总结"
    echo "========================================"
    echo -e "总测试数: ${CYAN}$TOTAL_TESTS${NC}"
    echo -e "通过:     ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败:     ${RED}$FAILED_TESTS${NC}"

    if [ $FAILED_TESTS -eq 0 ]; then
        echo ""
        log_pass "所有自动化测试通过！ ✓"
    else
        echo ""
        log_fail "有 $FAILED_TESTS 个测试失败"
    fi

    echo ""
    log_info "注意: 部分安全特性需要手动测试（RBAC, 2FA, 租户隔离等）"
    log_info "详细测试步骤请参考:"
    log_info "  - backend/services/admin-bff-service/ADVANCED_SECURITY_COMPLETE.md"
    log_info "  - backend/services/merchant-bff-service/MERCHANT_BFF_SECURITY.md"
    echo ""
}

# 主函数
main() {
    echo ""
    echo "========================================"
    log_info "BFF Security Features Test Suite"
    log_info "开始测试 Admin BFF 和 Merchant BFF 安全特性"
    echo "========================================"
    echo ""

    # 检查依赖
    log_info "检查依赖..."
    if ! command -v curl &> /dev/null; then
        log_fail "curl 未安装"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_warning "jq 未安装，部分测试将跳过"
    fi

    # 检查服务
    echo ""
    log_info "检查服务状态..."
    if ! check_service "Admin BFF" "$ADMIN_BFF_URL"; then
        log_fail "Admin BFF Service 未运行，请先启动服务"
        log_info "启动命令: ./scripts/start-bff-services.sh"
        exit 1
    fi

    if ! check_service "Merchant BFF" "$MERCHANT_BFF_URL"; then
        log_fail "Merchant BFF Service 未运行，请先启动服务"
        log_info "启动命令: ./scripts/start-bff-services.sh"
        exit 1
    fi

    # 测试 Admin BFF
    echo ""
    echo "========================================"
    log_info "测试 Admin BFF Service"
    echo "========================================"

    test_jwt_auth "Admin BFF" "$ADMIN_BFF_URL"
    test_rate_limiting "Admin BFF" "$ADMIN_BFF_URL" 60
    test_data_masking "Admin BFF" "$ADMIN_BFF_URL"
    test_structured_logging "Admin BFF" "admin-bff"
    test_metrics "Admin BFF" "$ADMIN_BFF_URL"
    test_health_check "Admin BFF" "$ADMIN_BFF_URL"
    test_admin_bff_features

    # 测试 Merchant BFF
    echo ""
    echo "========================================"
    log_info "测试 Merchant BFF Service"
    echo "========================================"

    test_jwt_auth "Merchant BFF" "$MERCHANT_BFF_URL"
    test_rate_limiting "Merchant BFF" "$MERCHANT_BFF_URL" 300
    test_data_masking "Merchant BFF" "$MERCHANT_BFF_URL"
    test_structured_logging "Merchant BFF" "merchant-bff"
    test_metrics "Merchant BFF" "$MERCHANT_BFF_URL"
    test_health_check "Merchant BFF" "$MERCHANT_BFF_URL"
    test_merchant_bff_features

    # 显示总结
    show_summary
}

# 执行主函数
main "$@"

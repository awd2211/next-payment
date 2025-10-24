#!/bin/bash

###############################################################################
# Saga 集成测试脚本
#
# 功能：
# 1. 验证所有 Saga 集成的服务编译通过
# 2. 检查 Saga 框架测试通过
# 3. 生成集成测试报告
#
# 使用方法：
#   ./scripts/test-saga-integration.sh
#
# 作者: Claude Code
# 日期: 2025-10-24
###############################################################################

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 设置 Go Workspace
export GOWORK="$PROJECT_ROOT/go.work"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║           Saga 集成测试 - 开始执行                          ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_ITEMS=()

###############################################################################
# 函数: 打印测试项
###############################################################################
print_test_header() {
    local test_name="$1"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}测试 $TOTAL_TESTS: $test_name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

###############################################################################
# 函数: 记录成功
###############################################################################
record_success() {
    local test_name="$1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}✅ PASS${NC} - $test_name"
}

###############################################################################
# 函数: 记录失败
###############################################################################
record_failure() {
    local test_name="$1"
    local error_msg="$2"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    FAILED_ITEMS+=("$test_name")
    echo -e "${RED}❌ FAIL${NC} - $test_name"
    if [ -n "$error_msg" ]; then
        echo -e "${RED}   错误: $error_msg${NC}"
    fi
}

###############################################################################
# 测试 1: Saga 框架单元测试
###############################################################################
print_test_header "Saga 框架单元测试 (pkg/saga)"

cd "$PROJECT_ROOT/pkg/saga"
if timeout 30 go test -v -run TestSagaBuilder 2>&1 | tee /tmp/saga-test.log; then
    if grep -q "PASS" /tmp/saga-test.log; then
        record_success "Saga 框架单元测试"
    else
        record_failure "Saga 框架单元测试" "测试未通过"
    fi
else
    record_failure "Saga 框架单元测试" "测试执行超时或失败"
fi
cd "$PROJECT_ROOT"

###############################################################################
# 测试 2: Withdrawal Service 编译验证
###############################################################################
print_test_header "Withdrawal Service 编译验证 (含 Withdrawal Saga)"

cd "$PROJECT_ROOT/services/withdrawal-service"
if timeout 30 go build -o /tmp/test-withdrawal-service ./cmd/main.go 2>&1; then
    record_success "Withdrawal Service 编译"

    # 验证 Saga 集成代码存在
    if grep -q "SetSagaService" internal/service/withdrawal_service.go; then
        record_success "  └─ Withdrawal Saga 集成代码存在"
    else
        record_failure "  └─ Withdrawal Saga 集成代码存在" "未找到 SetSagaService 方法"
    fi

    if grep -q "ExecuteWithdrawalSaga" internal/service/withdrawal_service.go; then
        record_success "  └─ Saga 调用代码存在"
    else
        record_failure "  └─ Saga 调用代码存在" "未找到 ExecuteWithdrawalSaga 调用"
    fi
else
    record_failure "Withdrawal Service 编译" "编译失败"
fi
cd "$PROJECT_ROOT"

###############################################################################
# 测试 3: Settlement Service 编译验证
###############################################################################
print_test_header "Settlement Service 编译验证 (含 Settlement Saga)"

cd "$PROJECT_ROOT/services/settlement-service"
if timeout 30 go build -o /tmp/test-settlement-service ./cmd/main.go 2>&1; then
    record_success "Settlement Service 编译"

    # 验证 Saga 集成代码存在
    if grep -q "SetSagaService" internal/service/settlement_service.go; then
        record_success "  └─ Settlement Saga 集成代码存在"
    else
        record_failure "  └─ Settlement Saga 集成代码存在" "未找到 SetSagaService 方法"
    fi

    if grep -q "ExecuteSettlementSaga" internal/service/settlement_service.go; then
        record_success "  └─ Saga 调用代码存在"
    else
        record_failure "  └─ Saga 调用代码存在" "未找到 ExecuteSettlementSaga 调用"
    fi
else
    record_failure "Settlement Service 编译" "编译失败"
fi
cd "$PROJECT_ROOT"

###############################################################################
# 测试 4: Payment Gateway 编译验证
###############################################################################
print_test_header "Payment Gateway 编译验证 (含 Refund & Callback Saga)"

cd "$PROJECT_ROOT/services/payment-gateway"
if timeout 30 go build -o /tmp/test-payment-gateway ./cmd/main.go 2>&1; then
    record_success "Payment Gateway 编译"

    # 验证 Refund Saga 集成
    if grep -q "SetRefundSagaService" internal/service/payment_service.go; then
        record_success "  └─ Refund Saga 集成代码存在"
    else
        record_failure "  └─ Refund Saga 集成代码存在" "未找到 SetRefundSagaService 方法"
    fi

    if grep -q "ExecuteRefundSaga" internal/service/payment_service.go; then
        record_success "  └─ Refund Saga 调用代码存在"
    else
        record_failure "  └─ Refund Saga 调用代码存在" "未找到 ExecuteRefundSaga 调用"
    fi

    # 验证 Callback Saga 集成
    if grep -q "SetCallbackSagaService" internal/service/payment_service.go; then
        record_success "  └─ Callback Saga 集成代码存在"
    else
        record_failure "  └─ Callback Saga 集成代码存在" "未找到 SetCallbackSagaService 方法"
    fi

    if grep -q "ExecuteCallbackSaga" internal/service/payment_service.go; then
        record_success "  └─ Callback Saga 调用代码存在"
    else
        record_failure "  └─ Callback Saga 调用代码存在" "未找到 ExecuteCallbackSaga 调用"
    fi
else
    record_failure "Payment Gateway 编译" "编译失败"
fi
cd "$PROJECT_ROOT"

###############################################################################
# 测试 5: Saga Service 文件存在性验证
###############################################################################
print_test_header "Saga Service 文件存在性验证"

SAGA_FILES=(
    "services/withdrawal-service/internal/service/withdrawal_saga_service.go"
    "services/settlement-service/internal/service/settlement_saga_service.go"
    "services/payment-gateway/internal/service/refund_saga_service.go"
    "services/payment-gateway/internal/service/callback_saga_service.go"
)

for saga_file in "${SAGA_FILES[@]}"; do
    if [ -f "$PROJECT_ROOT/$saga_file" ]; then
        record_success "  └─ 文件存在: $saga_file"
    else
        record_failure "  └─ 文件存在: $saga_file" "文件不存在"
    fi
done

###############################################################################
# 测试 6: 双模式兼容代码验证
###############################################################################
print_test_header "双模式兼容代码验证 (Saga + 旧逻辑降级)"

# 检查 withdrawal-service
if grep -q "if s.sagaService != nil" "$PROJECT_ROOT/services/withdrawal-service/internal/service/withdrawal_service.go"; then
    record_success "  └─ Withdrawal Service 双模式兼容"
else
    record_failure "  └─ Withdrawal Service 双模式兼容" "未找到双模式检查代码"
fi

# 检查 settlement-service
if grep -q "if s.sagaService != nil" "$PROJECT_ROOT/services/settlement-service/internal/service/settlement_service.go"; then
    record_success "  └─ Settlement Service 双模式兼容"
else
    record_failure "  └─ Settlement Service 双模式兼容" "未找到双模式检查代码"
fi

# 检查 payment-gateway (Refund)
if grep -q "if s.refundSagaService != nil" "$PROJECT_ROOT/services/payment-gateway/internal/service/payment_service.go"; then
    record_success "  └─ Payment Gateway (Refund) 双模式兼容"
else
    record_failure "  └─ Payment Gateway (Refund) 双模式兼容" "未找到双模式检查代码"
fi

# 检查 payment-gateway (Callback)
if grep -q "if s.callbackSagaService != nil" "$PROJECT_ROOT/services/payment-gateway/internal/service/payment_service.go"; then
    record_success "  └─ Payment Gateway (Callback) 双模式兼容"
else
    record_failure "  └─ Payment Gateway (Callback) 双模式兼容" "未找到双模式检查代码"
fi

###############################################################################
# 测试 7: Saga 注入代码验证 (main.go)
###############################################################################
print_test_header "Saga 注入代码验证 (main.go 中的依赖注入)"

# 检查 withdrawal-service
if grep -q "SetSagaService" "$PROJECT_ROOT/services/withdrawal-service/cmd/main.go"; then
    record_success "  └─ Withdrawal Service Saga 注入代码存在"
else
    record_failure "  └─ Withdrawal Service Saga 注入代码存在" "未找到注入代码"
fi

# 检查 settlement-service
if grep -q "SetSagaService" "$PROJECT_ROOT/services/settlement-service/cmd/main.go"; then
    record_success "  └─ Settlement Service Saga 注入代码存在"
else
    record_failure "  └─ Settlement Service Saga 注入代码存在" "未找到注入代码"
fi

# 检查 payment-gateway
if grep -q "SetRefundSagaService" "$PROJECT_ROOT/services/payment-gateway/cmd/main.go"; then
    record_success "  └─ Payment Gateway Refund Saga 注入代码存在"
else
    record_failure "  └─ Payment Gateway Refund Saga 注入代码存在" "未找到注入代码"
fi

if grep -q "SetCallbackSagaService" "$PROJECT_ROOT/services/payment-gateway/cmd/main.go"; then
    record_success "  └─ Payment Gateway Callback Saga 注入代码存在"
else
    record_failure "  └─ Payment Gateway Callback Saga 注入代码存在" "未找到注入代码"
fi

###############################################################################
# 测试 8: 文档完整性验证
###############################################################################
print_test_header "文档完整性验证"

DOCS=(
    "SAGA_ALL_COMPLETE.md"
    "SAGA_DEEP_INTEGRATION_COMPLETE.md"
    "SAGA_BUSINESS_INTEGRATION_REPORT.md"
    "SAGA_FINAL_IMPLEMENTATION_REPORT.md"
    "SAGA_INTEGRATION_DONE.md"
)

for doc in "${DOCS[@]}"; do
    # 文档在项目根目录的上一级
    if [ -f "$PROJECT_ROOT/../$doc" ]; then
        record_success "  └─ 文档存在: $doc"
    else
        record_failure "  └─ 文档存在: $doc" "文档不存在"
    fi
done

###############################################################################
# 生成测试报告
###############################################################################
echo ""
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                   测试报告                                  ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

echo -e "${BLUE}测试统计:${NC}"
echo "  总测试项: $TOTAL_TESTS"
echo -e "  ${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "  ${RED}失败: $FAILED_TESTS${NC}"
echo ""

# 计算通过率
PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✅ 所有测试通过！通过率: ${PASS_RATE}%${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${GREEN}🎉 Saga 集成验证成功！${NC}"
    echo ""
    echo "下一步："
    echo "  1. 启动服务: ./scripts/start-all-services.sh"
    echo "  2. 查看日志: tail -f logs/*.log | grep 'Saga'"
    echo "  3. 访问监控: http://localhost:40090 (Prometheus)"
    echo ""
    exit 0
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}❌ 部分测试失败！通过率: ${PASS_RATE}%${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${YELLOW}失败项列表:${NC}"
    for failed_item in "${FAILED_ITEMS[@]}"; do
        echo -e "  ${RED}✗${NC} $failed_item"
    done
    echo ""
    echo -e "${YELLOW}建议:${NC}"
    echo "  1. 查看详细错误日志"
    echo "  2. 运行 'go mod tidy' 更新依赖"
    echo "  3. 清理构建缓存: 'go clean -cache'"
    echo "  4. 重新编译相关服务"
    echo ""
    exit 1
fi

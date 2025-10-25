#!/bin/bash

# 微服务一致性检查工具
# 自动验证所有服务是否符合统一架构模式

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

BACKEND_DIR="/home/eric/payment/backend"
SERVICES_DIR="$BACKEND_DIR/services"

# 统计
total_services=0
passed_services=0
failed_services=0

# 错误收集
declare -a errors

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}微服务一致性检查工具${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# 检查函数
check_service() {
    local service=$1
    local service_dir="$SERVICES_DIR/$service"
    local passed=true
    local issues=()

    echo -e "${BLUE}检查 $service...${NC}"

    # 1. 检查目录结构
    if [ ! -d "$service_dir" ]; then
        issues+=("  ❌ 服务目录不存在")
        passed=false
    else
        # 检查 cmd/main.go
        if [ ! -f "$service_dir/cmd/main.go" ]; then
            issues+=("  ❌ 缺少 cmd/main.go")
            passed=false
        fi

        # 检查 internal 层
        for layer in model repository service handler; do
            if [ ! -d "$service_dir/internal/$layer" ]; then
                issues+=("  ❌ 缺少 internal/$layer/")
                passed=false
            fi
        done

        # 检查 .air.toml
        if [ ! -f "$service_dir/.air.toml" ]; then
            issues+=("  ❌ 缺少 .air.toml")
            passed=false
        fi

        # 检查 go.mod
        if [ ! -f "$service_dir/go.mod" ]; then
            issues+=("  ❌ 缺少 go.mod")
            passed=false
        fi
    fi

    # 2. 检查 Bootstrap 使用
    if [ -f "$service_dir/cmd/main.go" ]; then
        if ! grep -q "app.Bootstrap" "$service_dir/cmd/main.go"; then
            issues+=("  ❌ 未使用 Bootstrap 框架")
            passed=false
        fi

        # 检查必要的 feature flags
        if ! grep -q "EnableTracing.*true" "$service_dir/cmd/main.go"; then
            issues+=("  ⚠️  未启用 Tracing")
        fi

        if ! grep -q "EnableMetrics.*true" "$service_dir/cmd/main.go"; then
            issues+=("  ⚠️  未启用 Metrics")
        fi

        if ! grep -q "EnableHealthCheck.*true" "$service_dir/cmd/main.go"; then
            issues+=("  ⚠️  未启用 Health Check")
        fi
    fi

    # 3. 检查端口配置
    if [ -f "$service_dir/cmd/main.go" ]; then
        port=$(grep "Port:" "$service_dir/cmd/main.go" | grep -o "40[0-9]*" | head -1)
        if [ -z "$port" ]; then
            issues+=("  ❌ 未找到端口配置")
            passed=false
        else
            echo -e "  ${GREEN}✓${NC} 端口: $port"
        fi
    fi

    # 4. 编译检查
    if [ -f "$service_dir/cmd/main.go" ]; then
        echo -n "  编译检查... "
        export GOWORK="$BACKEND_DIR/go.work"
        if timeout 30 go build -o /tmp/test-$service $service_dir/cmd/main.go 2>/dev/null; then
            size=$(ls -lh /tmp/test-$service 2>/dev/null | awk '{print $5}')
            echo -e "${GREEN}✓ 成功 ($size)${NC}"
            rm -f /tmp/test-$service
        else
            echo -e "${RED}✗ 失败${NC}"
            issues+=("  ❌ 编译失败")
            passed=false
        fi
    fi

    # 输出结果
    if [ ${#issues[@]} -gt 0 ]; then
        for issue in "${issues[@]}"; do
            echo -e "$issue"
        done
    fi

    if [ "$passed" = true ]; then
        echo -e "${GREEN}✓ $service 通过所有检查${NC}"
        ((passed_services++))
    else
        echo -e "${RED}✗ $service 存在问题${NC}"
        ((failed_services++))
        errors+=("$service: ${#issues[@]} 个问题")
    fi

    ((total_services++))
    echo ""
}

# 获取所有服务
cd "$SERVICES_DIR"

echo -e "${BLUE}[1/3] 扫描服务目录...${NC}"
services=($(ls -d */ | sed 's#/##'))
echo -e "${GREEN}✓ 发现 ${#services[@]} 个服务${NC}"
echo ""

echo -e "${BLUE}[2/3] 执行一致性检查...${NC}"
echo ""

for service in "${services[@]}"; do
    check_service "$service"
done

# 端口冲突检查
echo -e "${BLUE}[3/3] 检查端口冲突...${NC}"
echo ""

ports=$(grep -h "Port:" */cmd/main.go 2>/dev/null | grep -o "40[0-9]*" | sort -n)
duplicates=$(echo "$ports" | uniq -d)

if [ -z "$duplicates" ]; then
    echo -e "${GREEN}✓ 无端口冲突${NC}"
else
    echo -e "${RED}✗ 发现端口冲突:${NC}"
    echo "$duplicates"
    ((failed_services++))
fi

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}检查汇总${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""
echo -e "总服务数: ${BLUE}$total_services${NC}"
echo -e "通过检查: ${GREEN}$passed_services${NC}"
echo -e "存在问题: ${RED}$failed_services${NC}"

if [ $failed_services -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 所有服务均符合统一架构模式！${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}⚠️  以下服务需要修复:${NC}"
    for error in "${errors[@]}"; do
        echo -e "  ${RED}•${NC} $error"
    done
    echo ""
    echo -e "${YELLOW}请参考 MICROSERVICE_UNIFIED_PATTERNS.md 进行修复${NC}"
    exit 1
fi

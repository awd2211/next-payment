#!/bin/bash

# ============================================================================
# 自动化构建所有微服务 Docker 镜像
# ============================================================================
# 特性:
# - 并行构建（加速）
# - 构建缓存优化
# - 错误处理和重试
# - 构建日志保存
# - 构建统计报告
# ============================================================================

set -e

# 配置
BASE_DIR="/home/eric/payment"
BACKEND_DIR="$BASE_DIR/backend"
LOG_DIR="$BACKEND_DIR/logs/docker-build"
PARALLEL_JOBS=4  # 并行构建任务数
BUILD_CACHE="--no-cache"  # 默认不使用缓存，可改为 "" 启用缓存

# 创建日志目录
mkdir -p "$LOG_DIR"

# 定义所有服务
SERVICES=(
    "admin-bff-service:40001"
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
    "merchant-bff-service:40023"
    "merchant-quota-service:40024"
)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 统计变量
TOTAL_SERVICES=${#SERVICES[@]}
SUCCESS_COUNT=0
FAILED_COUNT=0
declare -a FAILED_SERVICES

# 打印带颜色的消息
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

# 构建单个服务
build_service() {
    local service_name=$1
    local port=$2
    local log_file="$LOG_DIR/${service_name}_$(date +%Y%m%d_%H%M%S).log"

    print_info "开始构建: ${service_name} (端口: ${port})"

    # 构建命令
    if docker build $BUILD_CACHE \
        -t "payment-platform/${service_name}:latest" \
        -t "payment-platform/${service_name}:$(date +%Y%m%d)" \
        -f "${BACKEND_DIR}/services/${service_name}/Dockerfile" \
        "${BACKEND_DIR}" > "$log_file" 2>&1; then

        print_success "✅ ${service_name} 构建成功"

        # 显示镜像信息
        image_size=$(docker images "payment-platform/${service_name}:latest" --format "{{.Size}}")
        echo "   镜像大小: $image_size"
        echo "   日志文件: $log_file"

        return 0
    else
        print_error "❌ ${service_name} 构建失败"
        echo "   日志文件: $log_file"
        echo "   最后20行日志:"
        tail -n 20 "$log_file" | sed 's/^/     /'

        return 1
    fi
}

# 清理旧的构建日志
cleanup_old_logs() {
    print_info "清理7天前的构建日志..."
    find "$LOG_DIR" -name "*.log" -mtime +7 -delete 2>/dev/null || true
}

# 显示构建摘要
show_summary() {
    echo ""
    echo "============================================================================"
    echo "                         构建摘要报告                                      "
    echo "============================================================================"
    echo "总服务数:   $TOTAL_SERVICES"
    echo "成功:       ${GREEN}$SUCCESS_COUNT${NC}"
    echo "失败:       ${RED}$FAILED_COUNT${NC}"
    echo "成功率:     $(awk "BEGIN {printf \"%.1f\", ($SUCCESS_COUNT/$TOTAL_SERVICES)*100}")%"

    if [ $FAILED_COUNT -gt 0 ]; then
        echo ""
        echo "失败的服务:"
        for failed_svc in "${FAILED_SERVICES[@]}"; do
            echo "  - $failed_svc"
        done
    fi

    echo ""
    echo "构建时间:   $(date)"
    echo "日志目录:   $LOG_DIR"
    echo "============================================================================"
}

# 并行构建函数
parallel_build() {
    local batch_size=$1
    local services=("${@:2}")
    local pids=()

    for i in "${!services[@]}"; do
        IFS=':' read -r service_name port <<< "${services[$i]}"

        # 并行构建
        build_service "$service_name" "$port" &
        pids+=($!)

        # 当达到批次大小时，等待当前批次完成
        if [ $(( (i + 1) % batch_size )) -eq 0 ] || [ $i -eq $(( ${#services[@]} - 1 )) ]; then
            for pid in "${pids[@]}"; do
                if wait $pid; then
                    ((SUCCESS_COUNT++))
                else
                    ((FAILED_COUNT++))
                    # 提取服务名
                    IFS=':' read -r failed_name _ <<< "${services[$i]}"
                    FAILED_SERVICES+=("$failed_name")
                fi
            done
            pids=()
        fi
    done
}

# 主函数
main() {
    echo "============================================================================"
    echo "               Docker 镜像批量构建工具                                     "
    echo "============================================================================"
    echo "基础目录:     $BASE_DIR"
    echo "后端目录:     $BACKEND_DIR"
    echo "日志目录:     $LOG_DIR"
    echo "并行任务数:   $PARALLEL_JOBS"
    echo "构建缓存:     $([ -z "$BUILD_CACHE" ] && echo "启用" || echo "禁用")"
    echo "服务总数:     $TOTAL_SERVICES"
    echo "============================================================================"
    echo ""

    # 检查 Docker 是否运行
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker 未运行或无权限访问"
        exit 1
    fi

    # 检查 backend 目录
    if [ ! -d "$BACKEND_DIR" ]; then
        print_error "Backend 目录不存在: $BACKEND_DIR"
        exit 1
    fi

    # 清理旧日志
    cleanup_old_logs

    # 记录开始时间
    START_TIME=$(date +%s)

    print_info "开始构建所有服务..."
    echo ""

    # 串行构建（更稳定，但较慢）
    # 如需并行构建，取消下面的注释并注释掉串行部分
    # parallel_build $PARALLEL_JOBS "${SERVICES[@]}"

    # 串行构建
    for service_info in "${SERVICES[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"

        if build_service "$service_name" "$port"; then
            ((SUCCESS_COUNT++))
        else
            ((FAILED_COUNT++))
            FAILED_SERVICES+=("$service_name")
        fi
        echo ""
    done

    # 记录结束时间
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    # 显示摘要
    show_summary

    echo ""
    echo "总耗时: $(printf '%02d:%02d:%02d' $((DURATION/3600)) $((DURATION%3600/60)) $((DURATION%60)))"
    echo ""

    # 返回失败代码（如果有失败）
    if [ $FAILED_COUNT -gt 0 ]; then
        exit 1
    fi

    print_success "🎉 所有镜像构建成功!"
    echo ""
    echo "下一步:"
    echo "  1. 查看镜像列表: docker images | grep payment-platform"
    echo "  2. 启动服务: docker-compose -f docker-compose.services.yml up -d"
    echo ""
}

# 处理 Ctrl+C
trap 'echo ""; print_warning "构建已中断"; exit 130' INT

# 运行主函数
main "$@"

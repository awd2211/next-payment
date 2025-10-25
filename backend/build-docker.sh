#!/bin/bash

# ============================================================
# Docker 镜像构建脚本
# 用于构建所有微服务的 Docker 镜像
# ============================================================

set -e  # 遇到错误立即退出

# 配置
REGISTRY="${DOCKER_REGISTRY:-payment-platform}"  # 镜像仓库前缀
VERSION="${VERSION:-latest}"  # 镜像版本
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 服务列表和端口映射
declare -A SERVICES=(
    ["admin-service"]="40001"
    ["merchant-service"]="40002"
    ["payment-gateway"]="40003"
    ["order-service"]="40004"
    ["channel-adapter"]="40005"
    ["risk-service"]="40006"
    ["accounting-service"]="40007"
    ["notification-service"]="40008"
    ["analytics-service"]="40009"
    ["config-service"]="40010"
    ["merchant-auth-service"]="40011"
    ["merchant-config-service"]="40012"
    ["settlement-service"]="40013"
    ["withdrawal-service"]="40014"
    ["kyc-service"]="40015"
    ["cashier-service"]="40016"
    ["reconciliation-service"]="40020"
    ["dispute-service"]="40021"
    ["merchant-limit-service"]="40022"
)

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 使用说明
usage() {
    cat << EOF
用法: $0 [选项] [服务名...]

选项:
    -h, --help              显示帮助信息
    -v, --version VERSION   指定镜像版本 (默认: latest)
    -r, --registry REGISTRY 指定镜像仓库前缀 (默认: payment-platform)
    -p, --push              构建后推送到镜像仓库
    --no-cache              构建时不使用缓存
    --parallel N            并行构建 N 个镜像 (默认: 4)

示例:
    # 构建所有服务
    $0

    # 构建指定服务
    $0 admin-service merchant-service

    # 构建并推送
    $0 --push --version v1.0.0

    # 并行构建（8 个并发）
    $0 --parallel 8

    # 不使用缓存重新构建
    $0 --no-cache admin-service
EOF
}

# 解析命令行参数
PUSH=false
NO_CACHE=""
PARALLEL=4
SPECIFIC_SERVICES=()

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        --no-cache)
            NO_CACHE="--no-cache"
            shift
            ;;
        --parallel)
            PARALLEL="$2"
            shift 2
            ;;
        -*)
            log_error "未知选项: $1"
            usage
            exit 1
            ;;
        *)
            SPECIFIC_SERVICES+=("$1")
            shift
            ;;
    esac
done

# 确定要构建的服务列表
if [ ${#SPECIFIC_SERVICES[@]} -eq 0 ]; then
    BUILD_SERVICES=("${!SERVICES[@]}")
else
    BUILD_SERVICES=("${SPECIFIC_SERVICES[@]}")
fi

# 验证服务是否存在
for service in "${BUILD_SERVICES[@]}"; do
    if [ -z "${SERVICES[$service]}" ]; then
        log_error "服务不存在: $service"
        exit 1
    fi
    if [ ! -f "services/$service/Dockerfile" ]; then
        log_error "Dockerfile 不存在: services/$service/Dockerfile"
        exit 1
    fi
done

# 打印构建信息
echo "============================================================"
echo "  Docker 镜像构建"
echo "============================================================"
log_info "镜像仓库: $REGISTRY"
log_info "版本标签: $VERSION"
log_info "Git 提交: $GIT_COMMIT"
log_info "构建时间: $BUILD_DATE"
log_info "并行数量: $PARALLEL"
log_info "服务数量: ${#BUILD_SERVICES[@]}"
echo ""

# 构建单个服务
build_service() {
    local service=$1
    local port=${SERVICES[$service]}
    local image_name="${REGISTRY}/${service}:${VERSION}"
    local latest_tag="${REGISTRY}/${service}:latest"

    log_info "[$service] 开始构建..."

    # 构建镜像
    if docker build \
        -f "services/$service/Dockerfile" \
        -t "$image_name" \
        -t "$latest_tag" \
        --build-arg SERVICE_NAME="$service" \
        --build-arg PORT="$port" \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_DATE="$BUILD_DATE" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        $NO_CACHE \
        .; then

        log_info "[$service] ✅ 构建成功: $image_name"

        # 推送镜像
        if [ "$PUSH" = true ]; then
            log_info "[$service] 推送镜像..."
            if docker push "$image_name" && docker push "$latest_tag"; then
                log_info "[$service] ✅ 推送成功"
            else
                log_error "[$service] ❌ 推送失败"
                return 1
            fi
        fi

        return 0
    else
        log_error "[$service] ❌ 构建失败"
        return 1
    fi
}

# 导出函数以便 xargs 使用
export -f build_service
export -f log_info
export -f log_error
export REGISTRY VERSION BUILD_DATE GIT_COMMIT PUSH NO_CACHE
export -A SERVICES

# 并行构建
log_info "开始并行构建..."
echo ""

SUCCESS_COUNT=0
FAILED_SERVICES=()

# 使用 GNU parallel 或 xargs 进行并行构建
if command -v parallel &> /dev/null; then
    # 使用 GNU parallel
    printf "%s\n" "${BUILD_SERVICES[@]}" | parallel -j "$PARALLEL" build_service {} || true
else
    # 使用 xargs（fallback）
    printf "%s\n" "${BUILD_SERVICES[@]}" | xargs -P "$PARALLEL" -I {} bash -c 'build_service "{}"' || true
fi

# 统计结果
echo ""
echo "============================================================"
echo "  构建完成"
echo "============================================================"

for service in "${BUILD_SERVICES[@]}"; do
    image_name="${REGISTRY}/${service}:${VERSION}"
    if docker images "$image_name" | grep -q "$VERSION"; then
        log_info "✅ $service"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        log_error "❌ $service"
        FAILED_SERVICES+=("$service")
    fi
done

echo ""
log_info "成功: $SUCCESS_COUNT/${#BUILD_SERVICES[@]}"

if [ ${#FAILED_SERVICES[@]} -gt 0 ]; then
    log_error "失败的服务:"
    for service in "${FAILED_SERVICES[@]}"; do
        echo "  - $service"
    done
    exit 1
fi

echo ""
log_info "所有镜像构建成功！"

# 显示镜像列表
echo ""
echo "============================================================"
echo "  构建的镜像"
echo "============================================================"
docker images | grep "$REGISTRY" | grep -E "$(IFS="|"; echo "${BUILD_SERVICES[*]}")"

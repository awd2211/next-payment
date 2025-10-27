#!/bin/bash

# ============================================================================
# 停止所有支付平台服务
# ============================================================================

set -e

BASE_DIR="/home/eric/payment"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

echo "============================================================================"
echo "停止所有支付平台服务"
echo "============================================================================"
echo ""

cd "$BASE_DIR"

print_info "停止 BFF 服务..."
docker-compose -f docker-compose.bff.yml down

print_info "停止核心微服务..."
docker-compose -f docker-compose.services.yml down

print_info "停止基础设施..."
docker-compose down

print_success "所有服务已停止"
echo ""
echo "提示:"
echo "  - 如需删除数据卷: docker-compose down -v"
echo "  - 如需删除所有镜像: docker rmi \$(docker images -q 'payment-platform/*')"
echo ""

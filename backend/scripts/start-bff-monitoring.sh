#!/bin/bash

# ============================================================================
# BFF Services Monitoring Startup Script
# 启动 Prometheus + Grafana 监控基础设施，用于 BFF 服务监控
# ============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")/../.." && pwd)
cd "$PROJECT_ROOT"

echo ""
log_info "=========================================="
log_info "BFF Services Monitoring Setup"
log_info "=========================================="
echo ""

# ============================================================================
# Step 1: 验证配置文件
# ============================================================================
log_step "1/6 验证 Prometheus 和 Grafana 配置文件..."

REQUIRED_FILES=(
    "backend/deployments/prometheus/prometheus.yml"
    "backend/deployments/prometheus/alerts/bff-alerts.yml"
    "backend/deployments/prometheus/rules/bff-recording-rules.yml"
    "monitoring/grafana/dashboards/bff-services-dashboard.json"
)

MISSING_FILES=0
for file in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "$PROJECT_ROOT/$file" ]; then
        log_error "缺少文件: $file"
        MISSING_FILES=$((MISSING_FILES + 1))
    else
        log_success "✓ $file"
    fi
done

if [ $MISSING_FILES -gt 0 ]; then
    log_error "发现 $MISSING_FILES 个缺失文件，请先创建这些文件"
    exit 1
fi

log_success "所有配置文件验证通过"
echo ""

# ============================================================================
# Step 2: 检查 Docker 和 docker-compose
# ============================================================================
log_step "2/6 检查 Docker 环境..."

if ! command -v docker &> /dev/null; then
    log_error "未安装 Docker，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    log_error "未安装 docker-compose，请先安装 docker-compose"
    exit 1
fi

log_success "Docker 环境检查通过"
echo ""

# ============================================================================
# Step 3: 停止现有的监控容器（如果存在）
# ============================================================================
log_step "3/6 停止现有监控容器..."

if docker ps -a | grep -q "payment-prometheus"; then
    log_info "停止 Prometheus 容器..."
    docker stop payment-prometheus 2>/dev/null || true
    docker rm payment-prometheus 2>/dev/null || true
    log_success "Prometheus 容器已停止"
fi

if docker ps -a | grep -q "payment-grafana"; then
    log_info "停止 Grafana 容器..."
    docker stop payment-grafana 2>/dev/null || true
    docker rm payment-grafana 2>/dev/null || true
    log_success "Grafana 容器已停止"
fi

echo ""

# ============================================================================
# Step 4: 启动监控基础设施
# ============================================================================
log_step "4/6 启动 Prometheus 和 Grafana..."

docker-compose up -d prometheus grafana

if [ $? -eq 0 ]; then
    log_success "监控容器启动成功"
else
    log_error "监控容器启动失败"
    exit 1
fi

echo ""

# ============================================================================
# Step 5: 等待服务就绪
# ============================================================================
log_step "5/6 等待服务就绪..."

# 等待 Prometheus
log_info "等待 Prometheus 启动..."
for i in {1..30}; do
    if curl -s http://localhost:40090/-/healthy > /dev/null 2>&1; then
        log_success "Prometheus 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "Prometheus 启动超时"
        exit 1
    fi
    sleep 1
done

# 等待 Grafana
log_info "等待 Grafana 启动..."
for i in {1..30}; do
    if curl -s http://localhost:40300/api/health > /dev/null 2>&1; then
        log_success "Grafana 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "Grafana 启动超时"
        exit 1
    fi
    sleep 1
done

echo ""

# ============================================================================
# Step 6: 验证监控目标
# ============================================================================
log_step "6/6 验证监控目标..."

log_info "检查 Prometheus 目标状态..."

# 等待 2 秒让 Prometheus 开始抓取
sleep 2

# 检查 BFF 服务是否被监控
BFF_TARGETS=$(curl -s http://localhost:40090/api/v1/targets | jq -r '.data.activeTargets[] | select(.labels.job | contains("bff")) | .labels.job' 2>/dev/null || echo "")

if echo "$BFF_TARGETS" | grep -q "admin-bff"; then
    log_success "✓ Admin BFF 目标已配置"
else
    log_warning "⚠ Admin BFF 目标未找到（服务可能未启动）"
fi

if echo "$BFF_TARGETS" | grep -q "merchant-bff"; then
    log_success "✓ Merchant BFF 目标已配置"
else
    log_warning "⚠ Merchant BFF 目标未找到（服务可能未启动）"
fi

echo ""

# ============================================================================
# 显示访问信息
# ============================================================================
echo ""
log_info "=========================================="
log_info "监控服务访问信息"
log_info "=========================================="
echo ""
echo -e "${GREEN}Prometheus UI:${NC}"
echo "  URL:      http://localhost:40090"
echo "  Targets:  http://localhost:40090/targets"
echo "  Alerts:   http://localhost:40090/alerts"
echo "  Rules:    http://localhost:40090/rules"
echo ""
echo -e "${GREEN}Grafana Dashboard:${NC}"
echo "  URL:      http://localhost:40300"
echo "  用户名:    admin"
echo "  密码:      admin"
echo ""
echo -e "${GREEN}BFF Services Metrics:${NC}"
echo "  Admin BFF:    http://localhost:40001/metrics"
echo "  Merchant BFF: http://localhost:40023/metrics"
echo ""
log_info "=========================================="
log_info "监控配置统计"
log_info "=========================================="
echo ""
echo "✓ Alert Rules:       21 (6 critical, 11 warning, 4 info)"
echo "✓ Recording Rules:   25 (7 HTTP, 5 security, 4 resources, 3 SLI/SLO)"
echo "✓ Grafana Panels:    15 (status, performance, security, resources)"
echo "✓ Scrape Targets:    2 (admin-bff: 10s, merchant-bff: 15s)"
echo ""

# ============================================================================
# 下一步操作提示
# ============================================================================
log_info "=========================================="
log_info "下一步操作"
log_info "=========================================="
echo ""
echo "1. 启动 BFF 服务 (如果尚未启动):"
echo "   ${CYAN}cd backend && ./scripts/start-bff-services.sh${NC}"
echo ""
echo "2. 验证 Prometheus 目标状态:"
echo "   ${CYAN}curl http://localhost:40090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job | contains(\"bff\"))'${NC}"
echo ""
echo "3. 导入 Grafana Dashboard:"
echo "   - 访问: http://localhost:40300"
echo "   - 登录: admin / admin"
echo "   - 导航: Dashboards → Import"
echo "   - 上传: monitoring/grafana/dashboards/bff-services-dashboard.json"
echo ""
echo "4. 查看实时告警:"
echo "   ${CYAN}open http://localhost:40090/alerts${NC}"
echo ""
echo "5. 停止监控服务:"
echo "   ${CYAN}docker-compose stop prometheus grafana${NC}"
echo ""

# ============================================================================
# 健康检查摘要
# ============================================================================
log_info "=========================================="
log_info "服务健康检查"
log_info "=========================================="
echo ""

PROMETHEUS_STATUS=$(curl -s http://localhost:40090/-/healthy 2>/dev/null && echo "✓ Running" || echo "✗ Down")
GRAFANA_STATUS=$(curl -s http://localhost:40300/api/health 2>/dev/null && echo "✓ Running" || echo "✗ Down")

echo -e "Prometheus:   $PROMETHEUS_STATUS"
echo -e "Grafana:      $GRAFANA_STATUS"
echo ""

# 检查 BFF 服务是否运行
if pgrep -f "admin-bff-service" > /dev/null; then
    echo -e "Admin BFF:    ${GREEN}✓ Running${NC}"
else
    echo -e "Admin BFF:    ${YELLOW}✗ Not Running${NC} (请先启动: ./scripts/start-bff-services.sh)"
fi

if pgrep -f "merchant-bff-service" > /dev/null; then
    echo -e "Merchant BFF: ${GREEN}✓ Running${NC}"
else
    echo -e "Merchant BFF: ${YELLOW}✗ Not Running${NC} (请先启动: ./scripts/start-bff-services.sh)"
fi

echo ""
log_success "监控基础设施启动完成！"
echo ""

# ============================================================================
# 可选: 打开浏览器
# ============================================================================
if command -v xdg-open &> /dev/null; then
    read -p "是否打开 Prometheus 和 Grafana UI? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        xdg-open http://localhost:40090 2>/dev/null &
        xdg-open http://localhost:40300 2>/dev/null &
        log_success "已打开浏览器"
    fi
fi

#!/bin/bash

# ============================================================================
# 为所有微服务生成Dockerfile和.dockerignore
# ============================================================================

set -e

# 定义服务配置 (服务名:端口:数据库名)
declare -A SERVICES=(
    ["admin-bff-service"]="40001:payment_admin"
    ["merchant-bff-service"]="40023:payment_merchant"
    ["payment-gateway"]="40003:payment_gateway"
    ["order-service"]="40004:payment_order"
    ["channel-adapter"]="40005:payment_channel"
    ["risk-service"]="40006:payment_risk"
    ["accounting-service"]="40007:payment_accounting"
    ["notification-service"]="40008:payment_notification"
    ["analytics-service"]="40009:payment_analytics"
    ["config-service"]="40010:payment_config"
    ["merchant-auth-service"]="40011:payment_merchant_auth"
    ["settlement-service"]="40013:payment_settlement"
    ["withdrawal-service"]="40014:payment_withdrawal"
    ["kyc-service"]="40015:payment_kyc"
    ["cashier-service"]="40016:payment_cashier"
    ["reconciliation-service"]="40020:payment_reconciliation"
    ["dispute-service"]="40021:payment_dispute"
    ["merchant-policy-service"]="40022:payment_merchant_policy"
    ["merchant-quota-service"]="40024:payment_merchant_quota"
)

BASE_DIR="/home/eric/payment/backend"
TEMPLATE_FILE="$BASE_DIR/Dockerfile.template"

echo "=== 开始生成Dockerfile和.dockerignore ==="
echo ""

# 遍历所有服务
for service in "${!SERVICES[@]}"; do
    IFS=':' read -r port dbname <<< "${SERVICES[$service]}"
    
    service_dir="$BASE_DIR/services/$service"
    
    if [ ! -d "$service_dir" ]; then
        echo "❌ 服务目录不存在: $service_dir"
        continue
    fi
    
    echo "📦 处理服务: $service (端口: $port, 数据库: $dbname)"
    
    # 生成Dockerfile
    cat > "$service_dir/Dockerfile" << DOCKERFILE_CONTENT
# ============================================================================
# Dockerfile for $service
# ============================================================================
# 基于统一模板构建
# ============================================================================

# ============================================================================
# Stage 1: Builder
# ============================================================================
FROM golang:1.24-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制go.work和go.mod (利用Docker层缓存)
COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY proto/go.mod proto/go.sum* ./proto/
COPY services/$service/go.mod services/$service/go.sum* ./services/$service/

# 下载依赖
WORKDIR /build/services/$service
RUN go mod download

# 复制源代码
WORKDIR /build
COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY services/$service/ ./services/$service/

# 编译服务
WORKDIR /build/services/$service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \\
    -trimpath \\
    -ldflags="-s -w" \\
    -o /app/service \\
    ./cmd/main.go

# ============================================================================
# Stage 2: Runtime
# ============================================================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl bash \\
    && addgroup -g 1000 appgroup \\
    && adduser -D -u 1000 -G appgroup appuser

# 设置时区
ENV TZ=Asia/Shanghai
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 创建必要的目录
RUN mkdir -p /app/logs /app/tmp /app/certs \\
    && chown -R appuser:appgroup /app

# 从builder复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/service /app/service

# 设置工作目录
WORKDIR /app

# 切换到非root用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \\
    CMD curl -f http://localhost:$port/health || exit 1

# 暴露端口
EXPOSE $port

# 环境变量
ENV SERVICE_NAME=$service \\
    PORT=$port \\
    DB_NAME=$dbname \\
    GIN_MODE=release

# 启动服务
CMD ["/app/service"]
DOCKERFILE_CONTENT

    # 生成.dockerignore
    cat > "$service_dir/.dockerignore" << 'DOCKERIGNORE_CONTENT'
# IDE files
.idea
.vscode
*.swp
*.swo
*~

# Build artifacts
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
/tmp/
/bin/

# Air hot reload
.air.toml
tmp/

# Test files
*_test.go
testdata/

# Documentation
*.md
docs/

# Git
.git
.gitignore

# CI/CD
.github
.gitlab-ci.yml

# Logs
*.log
logs/

# Environment files
.env
.env.*

# Coverage
coverage.out
*.cover
DOCKERIGNORE_CONTENT

    echo "  ✅ 已生成: Dockerfile 和 .dockerignore"
    echo ""
done

echo "=== 完成! 共生成 ${#SERVICES[@]} 个服务的 Dockerfile ==="

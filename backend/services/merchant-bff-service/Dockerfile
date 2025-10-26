# ============================================================================
# Merchant BFF Service Dockerfile
# Multi-stage build for minimal image size
# ============================================================================

# Stage 1: Builder
FROM golang:1.21-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git make gcc musl-dev

# 设置工作目录
WORKDIR /build

# 复制 go.work 和所有 go.mod/go.sum
COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/merchant-bff-service/go.mod services/merchant-bff-service/go.sum ./services/merchant-bff-service/

# 下载依赖
RUN cd services/merchant-bff-service && go mod download

# 复制源代码
COPY pkg/ ./pkg/
COPY services/merchant-bff-service/ ./services/merchant-bff-service/

# 编译
WORKDIR /build/services/merchant-bff-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/merchant-bff-service \
    ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    && update-ca-certificates

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从 builder 复制二进制文件
COPY --from=builder /app/merchant-bff-service /app/merchant-bff-service

# 修改权限
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 40023

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:40023/health || exit 1

# 启动服务
ENTRYPOINT ["/app/merchant-bff-service"]

# 元数据
LABEL maintainer="Payment Platform Team"
LABEL service="merchant-bff-service"
LABEL version="1.0.0"
LABEL description="Merchant BFF Service with Tenant Isolation"

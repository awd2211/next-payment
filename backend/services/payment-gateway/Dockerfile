# ============================================================
# Dockerfile for payment-gateway
# ============================================================

# ============================================================
# 阶段 1: 构建阶段
# ============================================================
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 复制 Go Workspace 和依赖
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/payment-gateway/go.mod services/payment-gateway/go.sum ./services/payment-gateway/

# 下载依赖
RUN cd services/payment-gateway && go mod download

# 复制源码
COPY pkg/ ./pkg/
COPY services/payment-gateway/ ./services/payment-gateway/

# 构建二进制文件
RUN cd services/payment-gateway && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/payment-gateway \
    ./cmd/main.go

# ============================================================
# 阶段 2: 运行阶段
# ============================================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl && \
    rm -rf /var/cache/apk/*

ENV TZ=UTC

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /app/logs && \
    chown -R appuser:appuser /app

WORKDIR /app

# 复制二进制文件
COPY --from=builder --chown=appuser:appuser /app/payment-gateway /app/payment-gateway

# 确保可执行
RUN chmod +x /app/payment-gateway

# 切换用户
USER appuser

# 暴露端口
EXPOSE 40003

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:40003/health || exit 1

# 镜像元数据
LABEL service.name="payment-gateway" \
      service.port="40003" \
      org.opencontainers.image.title="payment-gateway" \
      org.opencontainers.image.description="Payment Platform payment-gateway" \
      org.opencontainers.image.vendor="Payment Platform"

# 启动服务
ENTRYPOINT ["/app/payment-gateway"]

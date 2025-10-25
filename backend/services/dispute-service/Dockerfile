# ============================================================
# Dockerfile for dispute-service
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
COPY services/dispute-service/go.mod services/dispute-service/go.sum ./services/dispute-service/

# 下载依赖
RUN cd services/dispute-service && go mod download

# 复制源码
COPY pkg/ ./pkg/
COPY services/dispute-service/ ./services/dispute-service/

# 构建二进制文件
RUN cd services/dispute-service && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/dispute-service \
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
COPY --from=builder --chown=appuser:appuser /app/dispute-service /app/dispute-service

# 确保可执行
RUN chmod +x /app/dispute-service

# 切换用户
USER appuser

# 暴露端口
EXPOSE 40021

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:40021/health || exit 1

# 镜像元数据
LABEL service.name="dispute-service" \
      service.port="40021" \
      org.opencontainers.image.title="dispute-service" \
      org.opencontainers.image.description="Payment Platform dispute-service" \
      org.opencontainers.image.vendor="Payment Platform"

# 启动服务
ENTRYPOINT ["/app/dispute-service"]

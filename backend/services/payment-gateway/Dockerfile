# ============================================================================
# Dockerfile for payment-gateway
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
COPY services/payment-gateway/go.mod services/payment-gateway/go.sum* ./services/payment-gateway/

# 下载依赖
WORKDIR /build/services/payment-gateway
RUN go mod download

# 复制源代码
WORKDIR /build
COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY services/payment-gateway/ ./services/payment-gateway/

# 编译服务
WORKDIR /build/services/payment-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/service \
    ./cmd/main.go

# ============================================================================
# Stage 2: Runtime
# ============================================================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl bash \
    && addgroup -g 1000 appgroup \
    && adduser -D -u 1000 -G appgroup appuser

# 设置时区
ENV TZ=Asia/Shanghai
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 创建必要的目录
RUN mkdir -p /app/logs /app/tmp /app/certs \
    && chown -R appuser:appgroup /app

# 从builder复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/service /app/service

# 设置工作目录
WORKDIR /app

# 切换到非root用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:40003/health || exit 1

# 暴露端口
EXPOSE 40003

# 环境变量
ENV SERVICE_NAME=payment-gateway \
    PORT=40003 \
    DB_NAME=payment_gateway \
    GIN_MODE=release

# 启动服务
CMD ["/app/service"]

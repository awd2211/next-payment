#!/bin/bash
# 启动单个服务并配置 mTLS
# 用法: ./scripts/start-service-mtls.sh <service-name>
# 示例: ./scripts/start-service-mtls.sh order-service

set -e

SERVICE_NAME=$1

if [ -z "$SERVICE_NAME" ]; then
  echo "用法: $0 <service-name>"
  echo "示例: $0 order-service"
  exit 1
fi

SERVICE_DIR="services/$SERVICE_NAME"
CERT_DIR="certs/services/$SERVICE_NAME"

if [ ! -d "$SERVICE_DIR" ]; then
  echo "❌ 服务不存在: $SERVICE_NAME"
  exit 1
fi

if [ ! -d "$CERT_DIR" ]; then
  echo "❌ 证书不存在: $CERT_DIR"
  echo "请先运行: ./scripts/generate-mtls-certs.sh"
  exit 1
fi

echo "========================================="
echo "   启动服务: $SERVICE_NAME (mTLS 模式)"
echo "========================================="

# 配置环境变量
export ENABLE_MTLS=true
export TLS_CERT_FILE="$(pwd)/$CERT_DIR/cert.pem"
export TLS_KEY_FILE="$(pwd)/$CERT_DIR/key.pem"
export TLS_CA_FILE="$(pwd)/certs/ca/ca-cert.pem"

# 如果是客户端（payment-gateway），额外配置客户端证书
if [ "$SERVICE_NAME" == "payment-gateway" ]; then
  export TLS_CLIENT_CERT="$(pwd)/$CERT_DIR/cert.pem"
  export TLS_CLIENT_KEY="$(pwd)/$CERT_DIR/key.pem"

  # 配置目标服务 URL（使用 HTTPS）
  export ORDER_SERVICE_URL="https://localhost:40004"
  export RISK_SERVICE_URL="https://localhost:40006"
  export CHANNEL_SERVICE_URL="https://localhost:40005"
  export MERCHANT_AUTH_SERVICE_URL="https://localhost:40011"
  export NOTIFICATION_SERVICE_URL="https://localhost:40008"
  export ANALYTICS_SERVICE_URL="https://localhost:40009"
fi

echo "✓ mTLS 配置完成:"
echo "  - ENABLE_MTLS=$ENABLE_MTLS"
echo "  - TLS_CERT_FILE=$TLS_CERT_FILE"
echo "  - TLS_KEY_FILE=$TLS_KEY_FILE"
echo "  - TLS_CA_FILE=$TLS_CA_FILE"
echo ""
echo "正在启动服务..."
echo ""

# 启动服务
cd "$SERVICE_DIR"
go run cmd/main.go

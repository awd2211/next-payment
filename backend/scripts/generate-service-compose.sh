#!/bin/bash

# ============================================================================
# 为每个服务生成docker-compose.yml文件(放在服务目录下)
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVICES_DIR="$BACKEND_DIR/services"

echo "=== 为所有服务生成docker-compose.yml ==="

# 服务配置: 端口:数据库名
declare -A SERVICE_CONFIG=(
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

# 生成单个服务的docker-compose.yml
generate_service_compose() {
    local service_name=$1
    local port=$2
    local db_name=$3
    local service_dir="$SERVICES_DIR/$service_name"
    local compose_file="$service_dir/docker-compose.yml"

    echo "生成: $service_name/docker-compose.yml"

    cat > "$compose_file" << EOF
# ============================================================================
# Docker Compose配置 - $service_name
# ============================================================================
# 使用方式:
#   cd services/$service_name
#   docker-compose up -d          # 启动服务
#   docker-compose logs -f        # 查看日志
#   docker-compose down           # 停止服务
#   docker-compose restart        # 重启服务
# ============================================================================

version: '3.8'

services:
  $service_name:
    build:
      context: ../..
      dockerfile: services/$service_name/Dockerfile
    image: payment-platform/$service_name:latest
    container_name: $service_name
    hostname: $service_name.payment-network
    restart: unless-stopped

    networks:
      payment-network:
        aliases:
          - $service_name.payment-network

    ports:
      - "$port:$port"

    environment:
      # 基础配置
      - ENV=production
      - SERVICE_NAME=$service_name
      - PORT=$port
      - DB_NAME=$db_name

      # 数据库配置
      - DB_HOST=postgres.payment-network
      - DB_PORT=5432
      - DB_USER=\${POSTGRES_USER:-postgres}
      - DB_PASSWORD=\${POSTGRES_PASSWORD:-postgres}

      # Redis配置
      - REDIS_HOST=redis.payment-network
      - REDIS_PORT=6379
      - REDIS_PASSWORD=\${REDIS_PASSWORD:-}

      # Kafka配置
      - KAFKA_BROKERS=kafka.payment-network:9092

      # mTLS配置
      - ENABLE_MTLS=true
      - TLS_CERT_FILE=/app/certs/services/$service_name/$service_name.crt
      - TLS_KEY_FILE=/app/certs/services/$service_name/$service_name.key
      - TLS_CLIENT_CERT=/app/certs/services/$service_name/$service_name.crt
      - TLS_CLIENT_KEY=/app/certs/services/$service_name/$service_name.key
      - TLS_CA_FILE=/app/certs/ca/ca-cert.pem

      # JWT配置
      - JWT_SECRET=\${JWT_SECRET:-payment-platform-super-secret-jwt-key-change-in-production}

      # 可观测性配置
      - JAEGER_ENDPOINT=http://jaeger.payment-network:14268/api/traces
      - JAEGER_SAMPLING_RATE=10

      # 微服务URLs (通过Docker网络内部域名访问)
      - ADMIN_BFF_SERVICE_URL=https://admin-bff-service.payment-network:40001
      - MERCHANT_BFF_SERVICE_URL=https://merchant-bff-service.payment-network:40023
      - PAYMENT_GATEWAY_URL=https://payment-gateway.payment-network:40003
      - ORDER_SERVICE_URL=https://order-service.payment-network:40004
      - CHANNEL_ADAPTER_URL=https://channel-adapter.payment-network:40005
      - RISK_SERVICE_URL=https://risk-service.payment-network:40006
      - ACCOUNTING_SERVICE_URL=https://accounting-service.payment-network:40007
      - NOTIFICATION_SERVICE_URL=https://notification-service.payment-network:40008
      - ANALYTICS_SERVICE_URL=https://analytics-service.payment-network:40009
      - CONFIG_SERVICE_URL=https://config-service.payment-network:40010
      - MERCHANT_AUTH_SERVICE_URL=https://merchant-auth-service.payment-network:40011
      - SETTLEMENT_SERVICE_URL=https://settlement-service.payment-network:40013
      - WITHDRAWAL_SERVICE_URL=https://withdrawal-service.payment-network:40014
      - KYC_SERVICE_URL=https://kyc-service.payment-network:40015
      - CASHIER_SERVICE_URL=https://cashier-service.payment-network:40016
      - RECONCILIATION_SERVICE_URL=https://reconciliation-service.payment-network:40020
      - DISPUTE_SERVICE_URL=https://dispute-service.payment-network:40021
      - MERCHANT_POLICY_SERVICE_URL=https://merchant-policy-service.payment-network:40022
      - MERCHANT_QUOTA_SERVICE_URL=https://merchant-quota-service.payment-network:40024

    volumes:
      # 日志持久化
      - $service_name-logs:/app/logs

      # mTLS证书 (只读)
      - ../../certs:/app/certs:ro

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:$port/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M

# ============================================================================
# 网络配置
# ============================================================================
networks:
  payment-network:
    external: true
    name: payment_payment-network

# ============================================================================
# 数据卷配置
# ============================================================================
volumes:
  $service_name-logs:
    driver: local
EOF

    echo "✅ 已生成: $service_name/docker-compose.yml"
}

# 遍历所有服务
count=0
for service_name in "${!SERVICE_CONFIG[@]}"; do
    IFS=':' read -r port db_name <<< "${SERVICE_CONFIG[$service_name]}"

    if [ -d "$SERVICES_DIR/$service_name" ]; then
        generate_service_compose "$service_name" "$port" "$db_name"
        ((count++))
    else
        echo "⚠️  跳过不存在的服务: $service_name"
    fi
done

echo ""
echo "=== 完成 ==="
echo "已为 $count 个服务生成docker-compose.yml文件"
echo ""
echo "使用方式:"
echo "  # 进入服务目录启动"
echo "  cd services/payment-gateway"
echo "  docker-compose up -d"
echo ""
echo "  # 查看日志"
echo "  docker-compose logs -f"
echo ""
echo "  # 停止服务"
echo "  docker-compose down"

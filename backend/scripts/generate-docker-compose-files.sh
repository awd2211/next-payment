#!/bin/bash

# ============================================================================
# 生成独立的docker-compose配置文件 (每个服务一个文件)
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_DIR="$PROJECT_ROOT/docker-compose"

echo "=== 生成docker-compose配置文件 ==="
echo "输出目录: $COMPOSE_DIR"

# 确保目录存在
mkdir -p "$COMPOSE_DIR"

# 服务配置: 服务名:端口:数据库:依赖服务列表
declare -A SERVICES=(
    ["admin-bff-service"]="40001:payment_admin:config-service,risk-service,kyc-service,analytics-service,merchant-policy-service,channel-adapter,cashier-service,order-service,accounting-service,dispute-service,merchant-auth-service,merchant-config-service,notification-service,payment-gateway,reconciliation-service,settlement-service,withdrawal-service"
    ["merchant-bff-service"]="40023:payment_merchant:payment-gateway,order-service,settlement-service,withdrawal-service,accounting-service,analytics-service,kyc-service,merchant-auth-service,merchant-config-service,merchant-policy-service,notification-service,risk-service,dispute-service,reconciliation-service,cashier-service"
    ["payment-gateway"]="40003:payment_gateway:order-service,channel-adapter,risk-service"
    ["order-service"]="40004:payment_order:"
    ["channel-adapter"]="40005:payment_channel:"
    ["risk-service"]="40006:payment_risk:"
    ["accounting-service"]="40007:payment_accounting:"
    ["notification-service"]="40008:payment_notification:"
    ["analytics-service"]="40009:payment_analytics:"
    ["config-service"]="40010:payment_config:"
    ["merchant-auth-service"]="40011:payment_merchant_auth:"
    ["settlement-service"]="40013:payment_settlement:accounting-service,merchant-config-service"
    ["withdrawal-service"]="40014:payment_withdrawal:"
    ["kyc-service"]="40015:payment_kyc:"
    ["cashier-service"]="40016:payment_cashier:"
    ["reconciliation-service"]="40020:payment_reconciliation:"
    ["dispute-service"]="40021:payment_dispute:payment-gateway"
    ["merchant-policy-service"]="40022:payment_merchant_policy:"
    ["merchant-quota-service"]="40024:payment_merchant_quota:"
)

# 生成单个服务的docker-compose文件
generate_compose_file() {
    local service_name=$1
    local config=$2

    IFS=':' read -r port db_name dependencies <<< "$config"

    local compose_file="$COMPOSE_DIR/${service_name}.yml"

    echo "生成: ${service_name}.yml (端口: $port, 数据库: $db_name)"

    # 开始写入文件
    cat > "$compose_file" << 'HEADER'
# ============================================================================
# Docker Compose配置 - SERVICE_NAME_PLACEHOLDER
# ============================================================================
# 使用方式:
#   docker-compose -f docker-compose/SERVICE_NAME_PLACEHOLDER.yml up -d
#   docker-compose -f docker-compose/SERVICE_NAME_PLACEHOLDER.yml logs -f
#   docker-compose -f docker-compose/SERVICE_NAME_PLACEHOLDER.yml down
# ============================================================================

version: '3.8'

services:
HEADER

    # 替换服务名
    sed -i "s/SERVICE_NAME_PLACEHOLDER/${service_name}/g" "$compose_file"

    # 写入服务配置
    cat >> "$compose_file" << EOF
  ${service_name}:
    build:
      context: ../backend
      dockerfile: services/${service_name}/Dockerfile
      args:
        SERVICE_NAME: ${service_name}
        SERVICE_PORT: ${port}
    image: payment-platform/${service_name}:latest
    container_name: ${service_name}
    hostname: ${service_name}.payment-network
    restart: unless-stopped
EOF

    # 添加depends_on (如果有依赖)
    if [ -n "$dependencies" ]; then
        echo "    depends_on:" >> "$compose_file"
        IFS=',' read -ra DEPS <<< "$dependencies"
        for dep in "${DEPS[@]}"; do
            echo "      - $dep" >> "$compose_file"
        done
        echo "" >> "$compose_file"
    fi

    # 继续写入网络和端口配置
    cat >> "$compose_file" << EOF
    networks:
      payment-network:
        aliases:
          - ${service_name}.payment-network

    ports:
      - "${port}:${port}"

    environment:
      # 基础配置
      - ENV=production
      - SERVICE_NAME=${service_name}
      - PORT=${port}
      - DB_NAME=${db_name}

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
      - TLS_CERT_FILE=/app/certs/services/${service_name}/${service_name}.crt
      - TLS_KEY_FILE=/app/certs/services/${service_name}/${service_name}.key
      - TLS_CLIENT_CERT=/app/certs/services/${service_name}/${service_name}.crt
      - TLS_CLIENT_KEY=/app/certs/services/${service_name}/${service_name}.key
      - TLS_CA_FILE=/app/certs/ca/ca-cert.pem

      # JWT配置
      - JWT_SECRET=\${JWT_SECRET:-payment-platform-super-secret-jwt-key-change-in-production}

      # 可观测性配置
      - JAEGER_ENDPOINT=http://jaeger.payment-network:14268/api/traces
      - JAEGER_SAMPLING_RATE=10

      # 下游服务URLs (HTTPS + mTLS)
EOF

    # 添加所有服务URL
    for svc in "${!SERVICES[@]}"; do
        IFS=':' read -r svc_port _ _ <<< "${SERVICES[$svc]}"
        local url_var=$(echo "$svc" | tr '[:lower:]' '[:upper:]' | tr '-' '_')_URL
        echo "      - ${url_var}=https://${svc}.payment-network:${svc_port}" >> "$compose_file"
    done

    # 写入volumes和健康检查
    cat >> "$compose_file" << EOF

    volumes:
      # 日志持久化
      - ${service_name}-logs:/app/logs

      # mTLS证书 (只读)
      - ../backend/certs:/app/certs:ro

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:${port}/health"]
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
  ${service_name}-logs:
    driver: local
EOF

    echo "✅ 已生成: $compose_file"
}

# 生成所有服务的docker-compose文件
count=0
for service in "${!SERVICES[@]}"; do
    generate_compose_file "$service" "${SERVICES[$service]}"
    ((count++))
done

echo ""
echo "=== 完成 ==="
echo "已生成 $count 个docker-compose配置文件"
echo ""
echo "使用方式:"
echo "  # 启动单个服务"
echo "  docker-compose -f docker-compose/payment-gateway.yml up -d"
echo ""
echo "  # 启动所有服务"
echo "  ./scripts/docker-deploy-all.sh"
echo ""
echo "  # 停止服务"
echo "  docker-compose -f docker-compose/payment-gateway.yml down"

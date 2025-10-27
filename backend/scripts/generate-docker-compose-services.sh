#!/bin/bash

# ============================================================================
# 生成完整的 docker-compose.services.yml
# 包含所有 19 个微服务的生产级配置
# ============================================================================
# 特性:
# - 内网域名 (service-name.payment-network)
# - mTLS 启用 (HTTPS + 证书挂载)
# - 健康检查
# - 资源限制
# - 日志管理
# - 依赖管理
# ============================================================================

set -e

OUTPUT_FILE="/home/eric/payment/docker-compose.services.yml"

echo "=== 开始生成 docker-compose.services.yml ==="
echo "输出文件: $OUTPUT_FILE"
echo ""

# 生成文件头部
cat > "$OUTPUT_FILE" << 'EOF_HEADER'
# ============================================================================
# Docker Compose - 19个微服务完整配置
# ============================================================================
# 生成时间: GENERATION_TIME
# 特性:
# - 内网域名格式: <service-name>.payment-network
# - mTLS启用: HTTPS + 证书挂载
# - 持久化: logs卷, certs卷(只读)
# - 网络: payment-network (172.28.0.0/16)
# - 健康检查: HTTP /health 端点
# - 资源限制: CPU/内存配额
# - 日志管理: JSON格式, 10MB轮转
# ============================================================================

version: '3.8'

services:
EOF_HEADER

# 替换生成时间
sed -i "s/GENERATION_TIME/$(date '+%Y-%m-%d %H:%M:%S')/" "$OUTPUT_FILE"

# 定义所有服务配置
declare -A SERVICES=(
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

# 服务间依赖关系（用于生成下游服务URL）
declare -A SERVICE_DEPENDENCIES=(
    ["payment-gateway"]="order-service,channel-adapter,risk-service,accounting-service"
    ["order-service"]="notification-service"
    ["settlement-service"]="accounting-service,withdrawal-service,merchant-policy-service"
    ["withdrawal-service"]="accounting-service,notification-service"
    ["accounting-service"]="notification-service"
    ["dispute-service"]="payment-gateway,notification-service"
    ["reconciliation-service"]="payment-gateway,accounting-service,notification-service"
)

# 按服务排序（按端口号）
SORTED_SERVICES=(
    "payment-gateway"
    "order-service"
    "channel-adapter"
    "risk-service"
    "accounting-service"
    "notification-service"
    "analytics-service"
    "config-service"
    "merchant-auth-service"
    "settlement-service"
    "withdrawal-service"
    "kyc-service"
    "cashier-service"
    "reconciliation-service"
    "dispute-service"
    "merchant-policy-service"
    "merchant-quota-service"
)

# 生成每个服务的配置
for service_name in "${SORTED_SERVICES[@]}"; do
    IFS=':' read -r port dbname <<< "${SERVICES[$service_name]}"

    # 服务名称转换（用于容器名和镜像名）
    container_name="payment-${service_name}"
    image_name="payment-platform/${service_name}:latest"
    hostname="${service_name}.payment-network"

    # 生成服务配置
    cat >> "$OUTPUT_FILE" << EOF_SERVICE

  # ==========================================================================
  # ${service_name} - Port ${port}
  # ==========================================================================
  ${service_name}:
    build:
      context: ./backend
      dockerfile: services/${service_name}/Dockerfile
    container_name: ${container_name}
    image: ${image_name}
    hostname: ${hostname}
    ports:
      - "${port}:${port}"
    environment:
      # 基础配置
      - ENV=production
      - SERVICE_NAME=${service_name}
      - PORT=${port}
      - DB_NAME=${dbname}
      - GIN_MODE=release

      # 数据库配置 (使用内网域名)
      - DB_HOST=postgres.payment-network
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=\${DB_PASSWORD:-postgres}
      - DB_SSLMODE=disable
      - DB_MAX_IDLE_CONNS=10
      - DB_MAX_OPEN_CONNS=100
      - DB_CONN_MAX_LIFETIME=3600

      # Redis配置 (使用内网域名)
      - REDIS_HOST=redis.payment-network
      - REDIS_PORT=6379
      - REDIS_PASSWORD=\${REDIS_PASSWORD:-}
      - REDIS_DB=0

      # Kafka配置 (使用内网域名)
      - KAFKA_BROKERS=kafka.payment-network:9092
      - KAFKA_GROUP_ID=${service_name}-group

      # JWT密钥
      - JWT_SECRET=\${JWT_SECRET:-payment-platform-super-secret-jwt-key-change-in-production}

      # mTLS配置 (启用HTTPS)
      - ENABLE_MTLS=true
      - ENABLE_HTTPS=true
      - TLS_CERT_FILE=/app/certs/services/${service_name}/${service_name}.crt
      - TLS_KEY_FILE=/app/certs/services/${service_name}/${service_name}.key
      - TLS_CLIENT_CERT=/app/certs/services/${service_name}/${service_name}.crt
      - TLS_CLIENT_KEY=/app/certs/services/${service_name}/${service_name}.key
      - TLS_CA_FILE=/app/certs/ca/ca-cert.pem
      - TLS_VERIFY=true
EOF_SERVICE

    # 添加特定服务的下游依赖URL（如果有）
    if [ "${service_name}" == "payment-gateway" ]; then
        cat >> "$OUTPUT_FILE" << 'EOF_PG_DEPS'

      # 下游服务URL (内网域名 + HTTPS)
      - ORDER_SERVICE_URL=https://order-service.payment-network:40004
      - CHANNEL_SERVICE_URL=https://channel-adapter.payment-network:40005
      - RISK_SERVICE_URL=https://risk-service.payment-network:40006
      - ACCOUNTING_SERVICE_URL=https://accounting-service.payment-network:40007
      - NOTIFICATION_SERVICE_URL=https://notification-service.payment-network:40008

      # Stripe配置
      - STRIPE_API_KEY=${STRIPE_API_KEY:-sk_test_...}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET:-whsec_...}
EOF_PG_DEPS
    fi

    if [ "${service_name}" == "settlement-service" ]; then
        cat >> "$OUTPUT_FILE" << 'EOF_SETTLEMENT_DEPS'

      # 下游服务URL (内网域名 + HTTPS)
      - ACCOUNTING_SERVICE_URL=https://accounting-service.payment-network:40007
      - WITHDRAWAL_SERVICE_URL=https://withdrawal-service.payment-network:40014
      - MERCHANT_CONFIG_SERVICE_URL=https://merchant-policy-service.payment-network:40022
      - NOTIFICATION_SERVICE_URL=https://notification-service.payment-network:40008
EOF_SETTLEMENT_DEPS
    fi

    if [ "${service_name}" == "reconciliation-service" ]; then
        cat >> "$OUTPUT_FILE" << 'EOF_RECON_DEPS'

      # 下游服务URL (内网域名 + HTTPS)
      - PAYMENT_SERVICE_URL=https://payment-gateway.payment-network:40003
      - ACCOUNTING_SERVICE_URL=https://accounting-service.payment-network:40007
      - NOTIFICATION_SERVICE_URL=https://notification-service.payment-network:40008
EOF_RECON_DEPS
    fi

    # 监控配置
    cat >> "$OUTPUT_FILE" << 'EOF_MONITORING'

      # 监控配置 (使用内网域名)
      - JAEGER_ENDPOINT=http://jaeger.payment-network:14268/api/traces
      - JAEGER_SAMPLING_RATE=10
      - PROMETHEUS_PUSH_GATEWAY=prometheus.payment-network:9091
      - LOG_LEVEL=info
EOF_MONITORING

    # 卷挂载
    cat >> "$OUTPUT_FILE" << EOF_VOLUMES

    volumes:
      - logs:/app/logs
      - ./backend/certs:/app/certs:ro
EOF_VOLUMES

    # 网络配置
    cat >> "$OUTPUT_FILE" << EOF_NETWORK

    networks:
      payment-network:
        aliases:
          - ${hostname}
EOF_NETWORK

    # 依赖配置
    cat >> "$OUTPUT_FILE" << 'EOF_DEPENDS'

    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      kafka:
        condition: service_started
EOF_DEPENDS

    # 资源限制
    cat >> "$OUTPUT_FILE" << 'EOF_RESOURCES'

    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s
EOF_RESOURCES

    # 健康检查
    cat >> "$OUTPUT_FILE" << EOF_HEALTH

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:${port}/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 30s
EOF_HEALTH

    # 重启策略和日志
    cat >> "$OUTPUT_FILE" << 'EOF_RESTART_LOGS'

    restart: unless-stopped

    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
EOF_RESTART_LOGS

    echo "  ✅ 已生成: ${service_name} 配置"
done

# 添加卷和网络定义
cat >> "$OUTPUT_FILE" << 'EOF_FOOTER'

# ============================================================================
# 持久化卷
# ============================================================================
volumes:
  logs:
    driver: local
    name: payment-logs

# ============================================================================
# 网络配置
# ============================================================================
networks:
  payment-network:
    external: true  # 使用主 docker-compose.yml 创建的网络

# ============================================================================
# 使用说明
# ============================================================================
# 1. 确保主 docker-compose.yml 已启动（创建 payment-network 网络）:
#    docker-compose up -d
#
# 2. 启动所有微服务:
#    docker-compose -f docker-compose.services.yml up -d
#
# 3. 查看服务状态:
#    docker-compose -f docker-compose.services.yml ps
#
# 4. 查看特定服务日志:
#    docker-compose -f docker-compose.services.yml logs -f payment-gateway
#
# 5. 停止所有服务:
#    docker-compose -f docker-compose.services.yml down
#
# 6. 重启特定服务:
#    docker-compose -f docker-compose.services.yml restart payment-gateway
#
# 7. 扩展服务实例:
#    docker-compose -f docker-compose.services.yml up -d --scale payment-gateway=3
#
# 8. 访问服务健康检查:
#    curl http://localhost:40003/health  # Payment Gateway
#    curl http://localhost:40004/health  # Order Service
#    ...
# ============================================================================
EOF_FOOTER

echo ""
echo "=== 生成完成! ==="
echo "文件位置: $OUTPUT_FILE"
echo "文件大小: $(du -h "$OUTPUT_FILE" | cut -f1)"
echo "服务数量: ${#SERVICES[@]}"
echo ""
echo "下一步:"
echo "  1. 检查配置: cat $OUTPUT_FILE"
echo "  2. 启动服务: docker-compose -f $OUTPUT_FILE up -d"
echo ""

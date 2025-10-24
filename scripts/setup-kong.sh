#!/bin/bash

# Kong API Gateway 配置脚本
# 为 merchant-portal 和 admin-portal 配置服务和路由

set -e

KONG_ADMIN_URL="http://localhost:40081"

echo "=========================================="
echo "Kong API Gateway 配置开始"
echo "=========================================="

# 等待 Kong Admin API 就绪
echo "等待 Kong Admin API 就绪..."
until curl -s -f "${KONG_ADMIN_URL}/status" > /dev/null 2>&1; do
  echo "等待 Kong 启动..."
  sleep 2
done
echo "✓ Kong Admin API 已就绪"

# 清理现有配置（可选）
echo ""
echo "清理现有配置..."
curl -s -X DELETE "${KONG_ADMIN_URL}/services/merchant-service" 2>/dev/null || true
curl -s -X DELETE "${KONG_ADMIN_URL}/services/admin-service" 2>/dev/null || true
curl -s -X DELETE "${KONG_ADMIN_URL}/services/payment-gateway" 2>/dev/null || true
curl -s -X DELETE "${KONG_ADMIN_URL}/services/order-service" 2>/dev/null || true
curl -s -X DELETE "${KONG_ADMIN_URL}/services/notification-service" 2>/dev/null || true

echo ""
echo "=========================================="
echo "配置 Merchant Service"
echo "=========================================="

# 1. 创建 Merchant Service
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  -d "name=merchant-service" \
  -d "url=http://host.docker.internal:40002" \
  | jq '.'

# 创建路由：/api/v1/merchant/* -> merchant-service
curl -s -X POST "${KONG_ADMIN_URL}/services/merchant-service/routes" \
  -d "name=merchant-api" \
  -d "paths[]=/api/v1/merchant" \
  -d "paths[]=/api/v1/dashboard" \
  -d "strip_path=false" \
  | jq '.'

# 注意：/api/v1/merchant/payments 由 merchant-service 处理（调用 payment-gateway）

# 添加 CORS 插件
curl -s -X POST "${KONG_ADMIN_URL}/services/merchant-service/plugins" \
  -d "name=cors" \
  -d "config.origins=*" \
  -d "config.methods[]=GET" \
  -d "config.methods[]=POST" \
  -d "config.methods[]=PUT" \
  -d "config.methods[]=DELETE" \
  -d "config.methods[]=PATCH" \
  -d "config.methods[]=OPTIONS" \
  -d "config.headers[]=Accept" \
  -d "config.headers[]=Authorization" \
  -d "config.headers[]=Content-Type" \
  -d "config.headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Trace-ID" \
  -d "config.credentials=true" \
  -d "config.max_age=3600" \
  | jq '.'

echo ""
echo "=========================================="
echo "配置 Admin Service"
echo "=========================================="

# 2. 创建 Admin Service
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  -d "name=admin-service" \
  -d "url=http://host.docker.internal:40001" \
  | jq '.'

# 创建路由：/api/v1/admin/* -> admin-service
curl -s -X POST "${KONG_ADMIN_URL}/services/admin-service/routes" \
  -d "name=admin-api" \
  -d "paths[]=/api/v1/admin" \
  -d "paths[]=/api/v1/merchants" \
  -d "paths[]=/api/v1/users" \
  -d "paths[]=/api/v1/roles" \
  -d "paths[]=/api/v1/permissions" \
  -d "strip_path=false" \
  | jq '.'

# 添加 CORS 插件
curl -s -X POST "${KONG_ADMIN_URL}/services/admin-service/plugins" \
  -d "name=cors" \
  -d "config.origins=*" \
  -d "config.methods[]=GET" \
  -d "config.methods[]=POST" \
  -d "config.methods[]=PUT" \
  -d "config.methods[]=DELETE" \
  -d "config.methods[]=PATCH" \
  -d "config.methods[]=OPTIONS" \
  -d "config.headers[]=Accept" \
  -d "config.headers[]=Authorization" \
  -d "config.headers[]=Content-Type" \
  -d "config.headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Trace-ID" \
  -d "config.credentials=true" \
  -d "config.max_age=3600" \
  | jq '.'

echo ""
echo "=========================================="
echo "配置 Payment Gateway"
echo "=========================================="

# 3. 创建 Payment Gateway Service
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  -d "name=payment-gateway" \
  -d "url=http://host.docker.internal:40003" \
  | jq '.'

# 创建路由：/api/v1/payments/* -> payment-gateway
curl -s -X POST "${KONG_ADMIN_URL}/services/payment-gateway/routes" \
  -d "name=payment-api" \
  -d "paths[]=/api/v1/payments" \
  -d "strip_path=false" \
  | jq '.'

# 添加 CORS 插件
curl -s -X POST "${KONG_ADMIN_URL}/services/payment-gateway/plugins" \
  -d "name=cors" \
  -d "config.origins=*" \
  -d "config.methods[]=GET" \
  -d "config.methods[]=POST" \
  -d "config.methods[]=PUT" \
  -d "config.methods[]=DELETE" \
  -d "config.methods[]=PATCH" \
  -d "config.methods[]=OPTIONS" \
  -d "config.headers[]=Accept" \
  -d "config.headers[]=Authorization" \
  -d "config.headers[]=Content-Type" \
  -d "config.headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Trace-ID" \
  -d "config.credentials=true" \
  -d "config.max_age=3600" \
  | jq '.'

echo ""
echo "=========================================="
echo "配置 Order Service"
echo "=========================================="

# 4. 创建 Order Service
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  -d "name=order-service" \
  -d "url=http://host.docker.internal:40004" \
  | jq '.'

# 创建路由：/api/v1/orders/* -> order-service
curl -s -X POST "${KONG_ADMIN_URL}/services/order-service/routes" \
  -d "name=order-api" \
  -d "paths[]=/api/v1/orders" \
  -d "strip_path=false" \
  | jq '.'

# 添加 CORS 插件
curl -s -X POST "${KONG_ADMIN_URL}/services/order-service/plugins" \
  -d "name=cors" \
  -d "config.origins=*" \
  -d "config.methods[]=GET" \
  -d "config.methods[]=POST" \
  -d "config.methods[]=PUT" \
  -d "config.methods[]=DELETE" \
  -d "config.methods[]=PATCH" \
  -d "config.methods[]=OPTIONS" \
  -d "config.headers[]=Accept" \
  -d "config.headers[]=Authorization" \
  -d "config.headers[]=Content-Type" \
  -d "config.headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Trace-ID" \
  -d "config.credentials=true" \
  -d "config.max_age=3600" \
  | jq '.'

echo ""
echo "=========================================="
echo "配置 Notification Service"
echo "=========================================="

# 5. 创建 Notification Service
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  -d "name=notification-service" \
  -d "url=http://host.docker.internal:40008" \
  | jq '.'

# 创建路由：/api/v1/notifications/* -> notification-service
curl -s -X POST "${KONG_ADMIN_URL}/services/notification-service/routes" \
  -d "name=notification-api" \
  -d "paths[]=/api/v1/notifications" \
  -d "strip_path=false" \
  | jq '.'

# 添加 CORS 插件
curl -s -X POST "${KONG_ADMIN_URL}/services/notification-service/plugins" \
  -d "name=cors" \
  -d "config.origins=*" \
  -d "config.methods[]=GET" \
  -d "config.methods[]=POST" \
  -d "config.methods[]=PUT" \
  -d "config.methods[]=DELETE" \
  -d "config.methods[]=PATCH" \
  -d "config.methods[]=OPTIONS" \
  -d "config.headers[]=Accept" \
  -d "config.headers[]=Authorization" \
  -d "config.headers[]=Content-Type" \
  -d "config.headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Request-ID" \
  -d "config.exposed_headers[]=X-Trace-ID" \
  -d "config.credentials=true" \
  -d "config.max_age=3600" \
  | jq '.'

echo ""
echo "=========================================="
echo "配置完成"
echo "=========================================="

# 显示配置摘要
echo ""
echo "已配置的服务："
curl -s "${KONG_ADMIN_URL}/services" | jq '.data[] | {name: .name, url: .url}'

echo ""
echo "已配置的路由："
curl -s "${KONG_ADMIN_URL}/routes" | jq '.data[] | {name: .name, paths: .paths, service: .service.name}'

echo ""
echo "=========================================="
echo "Kong 配置完成！"
echo "=========================================="
echo ""
echo "访问地址："
echo "  - Kong Proxy (API Gateway): http://localhost:40080"
echo "  - Kong Admin API:          http://localhost:40081"
echo "  - Konga Admin UI:          http://localhost:40082"
echo ""
echo "测试命令："
echo "  curl http://localhost:40080/api/v1/merchant/info -H 'Authorization: Bearer YOUR_JWT_TOKEN'"
echo ""


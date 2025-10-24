#!/bin/bash

# Kong API Gateway 重置脚本
# 清理所有配置并重新设置

set -e

KONG_ADMIN_URL="http://localhost:40081"

echo "=========================================="
echo "Kong API Gateway 重置开始"
echo "=========================================="

# 等待 Kong Admin API 就绪
echo "等待 Kong Admin API 就绪..."
until curl -s -f "${KONG_ADMIN_URL}/status" > /dev/null 2>&1; do
  echo "等待 Kong 启动..."
  sleep 2
done
echo "✓ Kong Admin API 已就绪"

echo ""
echo "清理所有插件..."
# 获取所有插件并删除
PLUGINS=$(curl -s "${KONG_ADMIN_URL}/plugins" | jq -r '.data[].id')
for plugin_id in $PLUGINS; do
  echo "删除插件: $plugin_id"
  curl -s -X DELETE "${KONG_ADMIN_URL}/plugins/$plugin_id" || true
done

echo ""
echo "清理所有路由..."
# 获取所有路由并删除
ROUTES=$(curl -s "${KONG_ADMIN_URL}/routes" | jq -r '.data[].id')
for route_id in $ROUTES; do
  echo "删除路由: $route_id"
  curl -s -X DELETE "${KONG_ADMIN_URL}/routes/$route_id" || true
done

echo ""
echo "清理所有服务..."
# 获取所有服务并删除
SERVICES=$(curl -s "${KONG_ADMIN_URL}/services" | jq -r '.data[].id')
for service_id in $SERVICES; do
  echo "删除服务: $service_id"
  curl -s -X DELETE "${KONG_ADMIN_URL}/services/$service_id" || true
done

echo ""
echo "✓ 清理完成"
echo ""
echo "运行配置脚本..."
/home/eric/payment/scripts/setup-kong.sh


#!/bin/bash
# mTLS 功能测试脚本
# 用途: 验证 mTLS 配置是否正确工作

set -e

CERT_DIR="certs"
CA_CERT="$CERT_DIR/ca/ca-cert.pem"
CLIENT_CERT="$CERT_DIR/services/payment-gateway/cert.pem"
CLIENT_KEY="$CERT_DIR/services/payment-gateway/key.pem"

echo "========================================="
echo "   mTLS 功能测试"
echo "========================================="
echo ""

# 检查证书是否存在
if [ ! -f "$CA_CERT" ]; then
  echo "❌ CA 证书不存在: $CA_CERT"
  echo "请先运行: ./scripts/generate-mtls-certs.sh"
  exit 1
fi

if [ ! -f "$CLIENT_CERT" ] || [ ! -f "$CLIENT_KEY" ]; then
  echo "❌ 客户端证书不存在"
  echo "请先运行: ./scripts/generate-mtls-certs.sh"
  exit 1
fi

echo "✓ 证书文件检查通过"
echo ""

# 测试函数
test_service() {
  local SERVICE_NAME=$1
  local PORT=$2
  local PROTOCOL=${3:-https}

  echo "---------------------------------------"
  echo "测试: $SERVICE_NAME (端口 $PORT)"
  echo "---------------------------------------"

  # 测试 1: 有效证书访问
  echo "[测试1] 使用有效证书访问 /health ..."
  if curl -s -f "$PROTOCOL://localhost:$PORT/health" \
    --cacert "$CA_CERT" \
    --cert "$CLIENT_CERT" \
    --key "$CLIENT_KEY" \
    --connect-timeout 5 > /dev/null 2>&1; then
    echo "  ✅ 成功: 有效证书可以访问"
  else
    echo "  ❌ 失败: 有效证书无法访问（可能服务未启动或未启用 mTLS）"
  fi

  # 测试 2: 无证书访问（应该被拒绝）
  echo "[测试2] 不带证书访问 /health ..."
  if curl -s -f "$PROTOCOL://localhost:$PORT/health" \
    --cacert "$CA_CERT" \
    --connect-timeout 5 > /dev/null 2>&1; then
    echo "  ⚠️  警告: 无证书也能访问（mTLS 可能未启用）"
  else
    echo "  ✅ 成功: 无证书被拒绝（符合预期）"
  fi

  echo ""
}

# 主测试流程
echo "========================================="
echo "开始测试各服务 mTLS 配置..."
echo "========================================="
echo ""

# 检查服务是否运行
echo "检查服务运行状态..."
RUNNING_SERVICES=0

for port in 40004 40006 40005; do
  if lsof -i :$port > /dev/null 2>&1; then
    ((RUNNING_SERVICES++))
  fi
done

if [ $RUNNING_SERVICES -eq 0 ]; then
  echo "⚠️  警告: 没有检测到运行中的服务"
  echo ""
  echo "请先启动服务（使用 mTLS）:"
  echo "  Terminal 1: ./scripts/start-service-mtls.sh order-service"
  echo "  Terminal 2: ./scripts/start-service-mtls.sh risk-service"
  echo "  Terminal 3: ./scripts/start-service-mtls.sh channel-adapter"
  echo ""
  exit 0
fi

echo "✓ 检测到 $RUNNING_SERVICES 个运行中的服务"
echo ""

# 测试各个服务
test_service "order-service" 40004
test_service "risk-service" 40006
test_service "channel-adapter" 40005
test_service "payment-gateway" 40003
test_service "merchant-service" 40002

echo "========================================="
echo "   测试完成"
echo "========================================="
echo ""
echo "说明:"
echo "  ✅ = 符合预期"
echo "  ❌ = 测试失败"
echo "  ⚠️  = 需要检查配置"
echo ""
echo "如果看到 ⚠️ 警告，请确保:"
echo "  1. 服务启动时设置了 ENABLE_MTLS=true"
echo "  2. 环境变量配置正确（TLS_CERT_FILE, TLS_KEY_FILE, TLS_CA_FILE）"
echo "  3. 证书路径正确且文件权限为 600"
echo ""

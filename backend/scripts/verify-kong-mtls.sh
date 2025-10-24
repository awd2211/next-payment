#!/bin/bash
# Kong mTLS 配置验证脚本
set -e

KONG_ADMIN="http://localhost:40081"

echo "=========================================="
echo "  Kong mTLS 配置验证"
echo "=========================================="
echo ""

# 检查 Kong 是否运行
echo "[1/5] 检查 Kong 容器状态..."
if ! docker ps | grep -q kong-gateway; then
    echo "  ❌ Kong 容器未运行"
    echo "  请先启动: docker-compose up -d kong"
    exit 1
fi
echo "  ✅ Kong 容器正在运行"
echo ""

# 检查 Kong API 可访问
echo "[2/5] 检查 Kong Admin API..."
if ! curl -s -f $KONG_ADMIN/ > /dev/null 2>&1; then
    echo "  ❌ Kong Admin API 不可访问"
    exit 1
fi
echo "  ✅ Kong Admin API 正常"
echo ""

# 检查证书挂载
echo "[3/5] 检查证书文件挂载..."
if ! docker exec kong-gateway test -f /kong/certs/ca/ca-cert.pem; then
    echo "  ❌ CA 证书未挂载"
    echo "  检查 docker-compose.yml 中的 volumes 配置"
    exit 1
fi

if ! docker exec kong-gateway test -f /kong/certs/kong-gateway/cert.pem; then
    echo "  ❌ Kong 客户端证书未挂载"
    echo "  请先运行: ./scripts/setup-kong-mtls-cert.sh"
    exit 1
fi

if ! docker exec kong-gateway test -f /kong/certs/kong-gateway/key.pem; then
    echo "  ❌ Kong 客户端私钥未挂载"
    exit 1
fi

echo "  ✅ 所有证书文件已正确挂载"
echo ""

# 检查环境变量
echo "[4/5] 检查 Kong 环境变量..."
if ! docker exec kong-gateway env | grep -q "KONG_CLIENT_SSL=on"; then
    echo "  ❌ KONG_CLIENT_SSL 未设置"
    echo "  检查 docker-compose.yml 中的 environment 配置"
    exit 1
fi

if ! docker exec kong-gateway env | grep -q "KONG_CLIENT_SSL_CERT"; then
    echo "  ❌ KONG_CLIENT_SSL_CERT 未设置"
    exit 1
fi

echo "  ✅ Kong mTLS 环境变量已配置"
echo ""

# 检查服务配置
echo "[5/5] 检查 Kong 服务配置..."
ORDER_SERVICE_URL=$(curl -s $KONG_ADMIN/services/order-service 2>/dev/null | grep -o '"url":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ORDER_SERVICE_URL" ]; then
    echo "  ⚠️  order-service 未配置"
    echo "  请运行: cd backend && ./scripts/kong-setup.sh"
else
    echo "  order-service URL: $ORDER_SERVICE_URL"

    if [[ "$ORDER_SERVICE_URL" == https://* ]]; then
        echo "  ✅ 服务已配置为 HTTPS（mTLS 模式）"
    else
        echo "  ⚠️  服务仍使用 HTTP（非 mTLS 模式）"
        echo "  要启用 mTLS，请运行: ENABLE_MTLS=true ./scripts/kong-setup.sh"
    fi
fi

echo ""
echo "=========================================="
echo "  验证完成"
echo "=========================================="
echo ""

# 显示当前状态
echo "📊 当前状态:"
echo "  - Kong 容器: ✅ 运行中"
echo "  - 证书挂载: ✅ 正常"
echo "  - 环境变量: ✅ 已配置"

if [[ "$ORDER_SERVICE_URL" == https://* ]]; then
    echo "  - mTLS 模式: ✅ 已启用"
else
    echo "  - mTLS 模式: ⚠️  未启用"
fi

echo ""
echo "下一步:"
if [[ "$ORDER_SERVICE_URL" != https://* ]]; then
    echo "  1. 启用 mTLS: ENABLE_MTLS=true ./scripts/kong-setup.sh"
    echo "  2. 启动后端服务: ./scripts/start-service-mtls.sh order-service"
    echo "  3. 测试访问: curl http://localhost:40080/api/v1/orders"
else
    echo "  1. 启动后端服务: ./scripts/start-service-mtls.sh order-service"
    echo "  2. 测试访问: curl http://localhost:40080/api/v1/orders"
    echo "  3. 查看 Kong 日志: docker-compose logs -f kong"
fi
echo ""

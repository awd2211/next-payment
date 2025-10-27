#!/bin/bash

# 配置Kong使用mTLS连接admin-bff-service (持久化方案)
# 说明: admin-bff-service运行在HTTPS上并要求客户端提供mTLS证书

set -e

KONG_ADMIN="http://localhost:40081"
CERT_DIR="/home/eric/payment/backend/certs"
SERVICE_NAME="admin-bff-service"

echo "=========================================="
echo "Kong mTLS配置脚本"
echo "=========================================="
echo ""
echo "目标: 配置Kong通过mTLS连接到admin-bff-service"
echo "证书目录: $CERT_DIR"
echo ""

# 1. 检查Kong是否运行
echo "[1/5] 检查Kong状态..."
if ! curl -s "$KONG_ADMIN/status" > /dev/null; then
    echo "错误: Kong Admin API无法访问 ($KONG_ADMIN)"
    echo "请确保Kong容器正在运行: docker-compose ps kong"
    exit 1
fi
echo "✓ Kong正在运行"
echo ""

# 2. 检查证书文件
echo "[2/5] 检查证书文件..."
CA_CERT="$CERT_DIR/ca/ca-cert.pem"
# 使用任意服务的证书作为Kong的客户端证书(CA签发的都可以)
CLIENT_CERT="$CERT_DIR/services/kong-gateway/kong-gateway.crt"
CLIENT_KEY="$CERT_DIR/services/kong-gateway/kong-gateway.key"

# 如果kong-gateway证书不存在,使用admin-bff-service的证书作为临时方案
if [ ! -f "$CLIENT_CERT" ]; then
    echo "警告: kong-gateway证书不存在,使用payment-gateway证书"
    CLIENT_CERT="$CERT_DIR/services/payment-gateway/payment-gateway.crt"
    CLIENT_KEY="$CERT_DIR/services/payment-gateway/payment-gateway.key"
fi

for file in "$CA_CERT" "$CLIENT_CERT" "$CLIENT_KEY"; do
    if [ ! -f "$file" ]; then
        echo "错误: 证书文件不存在: $file"
        exit 1
    fi
done
echo "✓ 所有证书文件存在"
echo ""

# 3. 上传CA证书到Kong (用于验证服务端证书)
echo "[3/5] 上传CA证书到Kong..."
CA_CERT_ID=$(curl -s -X POST "$KONG_ADMIN/ca_certificates" \
    -F "cert=@$CA_CERT" \
    | jq -r '.id // empty')

if [ -z "$CA_CERT_ID" ]; then
    echo "错误: 上传CA证书失败"
    echo "尝试获取现有CA证书..."
    CA_CERT_ID=$(curl -s "$KONG_ADMIN/ca_certificates" | jq -r '.data[0].id // empty')
    if [ -z "$CA_CERT_ID" ]; then
        echo "错误: 无法获取CA证书ID"
        exit 1
    fi
    echo "✓ 使用现有CA证书: $CA_CERT_ID"
else
    echo "✓ CA证书已上传: $CA_CERT_ID"
fi
echo ""

# 4. 上传客户端证书到Kong (用于mTLS认证)
echo "[4/5] 上传客户端证书到Kong..."
# 合并证书和私钥
COMBINED_CERT="/tmp/kong-client-combined.pem"
cat "$CLIENT_CERT" "$CLIENT_KEY" > "$COMBINED_CERT"

CLIENT_CERT_ID=$(curl -s -X POST "$KONG_ADMIN/certificates" \
    -F "cert=@$CLIENT_CERT" \
    -F "key=@$CLIENT_KEY" \
    | jq -r '.id // empty')

if [ -z "$CLIENT_CERT_ID" ]; then
    echo "错误: 上传客户端证书失败"
    echo "尝试获取现有客户端证书..."
    CLIENT_CERT_ID=$(curl -s "$KONG_ADMIN/certificates" | jq -r '.data[0].id // empty')
    if [ -z "$CLIENT_CERT_ID" ]; then
        echo "错误: 无法获取客户端证书ID"
        exit 1
    fi
    echo "✓ 使用现有客户端证书: $CLIENT_CERT_ID"
else
    echo "✓ 客户端证书已上传: $CLIENT_CERT_ID"
fi
echo ""

# 5. 更新admin-bff-service配置为HTTPS + mTLS
echo "[5/5] 更新admin-bff-service配置..."

# 删除旧的service配置
curl -s -X DELETE "$KONG_ADMIN/services/$SERVICE_NAME" > /dev/null 2>&1 || true

# 创建新的HTTPS service配置
SERVICE_RESPONSE=$(curl -s -X POST "$KONG_ADMIN/services" \
    -d "name=$SERVICE_NAME" \
    -d "protocol=https" \
    -d "host=host.docker.internal" \
    -d "port=40001" \
    -d "path=/" \
    -d "client_certificate.id=$CLIENT_CERT_ID" \
    -d "tls_verify=true" \
    -d "ca_certificates[]=$CA_CERT_ID")

SERVICE_ID=$(echo "$SERVICE_RESPONSE" | jq -r '.id // empty')

if [ -z "$SERVICE_ID" ]; then
    echo "错误: 创建service失败"
    echo "$SERVICE_RESPONSE" | jq .
    exit 1
fi

echo "✓ Service已配置: $SERVICE_ID"
echo ""

# 删除旧的route
curl -s -X DELETE "$KONG_ADMIN/routes/$SERVICE_NAME-route" > /dev/null 2>&1 || true

# 创建新的route
ROUTE_RESPONSE=$(curl -s -X POST "$KONG_ADMIN/services/$SERVICE_NAME/routes" \
    -d "name=$SERVICE_NAME-route" \
    -d "paths[]=/api" \
    -d "strip_path=false")

ROUTE_ID=$(echo "$ROUTE_RESPONSE" | jq -r '.id // empty')

if [ -z "$ROUTE_ID" ]; then
    echo "错误: 创建route失败"
    echo "$ROUTE_RESPONSE" | jq .
    exit 1
fi

echo "✓ Route已配置: $ROUTE_ID"
echo ""

# 6. 配置JWT插件
echo "[6/6] 配置JWT插件..."
curl -s -X DELETE "$KONG_ADMIN/routes/$ROUTE_ID/plugins" > /dev/null 2>&1 || true

PLUGIN_RESPONSE=$(curl -s -X POST "$KONG_ADMIN/routes/$ROUTE_ID/plugins" \
    -d "name=jwt" \
    -d "config.key_claim_name=iss")

PLUGIN_ID=$(echo "$PLUGIN_RESPONSE" | jq -r '.id // empty')

if [ -z "$PLUGIN_ID" ]; then
    echo "错误: 配置JWT插件失败"
    echo "$PLUGIN_RESPONSE" | jq .
    exit 1
fi

echo "✓ JWT插件已配置: $PLUGIN_ID"
echo ""

# 7. 显示配置摘要
echo "=========================================="
echo "配置完成!"
echo "=========================================="
echo ""
echo "Kong Service配置:"
echo "  - 名称: $SERVICE_NAME"
echo "  - 协议: HTTPS (mTLS)"
echo "  - 地址: https://host.docker.internal:40001"
echo "  - CA证书: $CA_CERT_ID"
echo "  - 客户端证书: $CLIENT_CERT_ID"
echo ""
echo "Kong Route配置:"
echo "  - 路径: /api/*"
echo "  - JWT插件: 已启用"
echo ""
echo "测试命令:"
echo "  1. 获取JWT token:"
echo "     curl -X POST http://localhost:40080/api/v1/admin/login \\"
echo "       -H 'Content-Type: application/json' \\"
echo "       -d '{\"username\":\"admin\",\"password\":\"admin123\"}'"
echo ""
echo "  2. 测试配置管理API:"
echo "     curl -H 'Authorization: Bearer YOUR_TOKEN' \\"
echo "       http://localhost:40080/api/v1/admin/configs"
echo ""

# 清理临时文件
rm -f "$COMBINED_CERT"

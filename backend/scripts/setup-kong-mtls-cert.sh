#!/bin/bash
# 为 Kong Gateway 生成 mTLS 客户端证书
set -e

CERTS_DIR="certs"
KONG_CERT_DIR="$CERTS_DIR/services/kong-gateway"
CA_CERT="$CERTS_DIR/ca/ca-cert.pem"
CA_KEY="$CERTS_DIR/ca/ca-key.pem"

echo "=========================================="
echo "  Kong Gateway mTLS 证书生成工具"
echo "=========================================="
echo ""

# 检查 CA 证书
if [ ! -f "$CA_CERT" ] || [ ! -f "$CA_KEY" ]; then
    echo "❌ CA 证书不存在"
    echo "请先运行: ./scripts/generate-mtls-certs.sh"
    exit 1
fi

echo "✓ CA 证书已存在"
echo ""

# 创建 Kong 证书目录
mkdir -p "$KONG_CERT_DIR"

if [ -f "$KONG_CERT_DIR/cert.pem" ]; then
    echo "⚠️  Kong 证书已存在，是否覆盖? (y/n)"
    read -r answer
    if [ "$answer" != "y" ]; then
        echo "跳过生成"
        exit 0
    fi
fi

echo "[1/4] 生成 Kong 私钥..."
openssl genrsa -out "$KONG_CERT_DIR/key.pem" 2048
echo "  ✓ 私钥已生成"
echo ""

echo "[2/4] 生成证书签名请求 (CSR)..."
openssl req -new \
    -key "$KONG_CERT_DIR/key.pem" \
    -out "$KONG_CERT_DIR/csr.pem" \
    -subj "/C=US/ST=California/L=San Francisco/O=Payment Platform/OU=Gateway/CN=kong-gateway"
echo "  ✓ CSR 已生成"
echo ""

echo "[3/4] 使用 CA 签名证书..."
cat > "$KONG_CERT_DIR/ext.cnf" <<EOF
subjectAltName = DNS:kong-gateway,DNS:localhost,DNS:kong,IP:127.0.0.1
extendedKeyUsage = serverAuth,clientAuth
EOF

openssl x509 -req -days 3650 \
    -in "$KONG_CERT_DIR/csr.pem" \
    -CA "$CA_CERT" \
    -CAkey "$CA_KEY" \
    -CAcreateserial \
    -out "$KONG_CERT_DIR/cert.pem" \
    -extfile "$KONG_CERT_DIR/ext.cnf"

echo "  ✓ 证书已签名"
echo ""

echo "[4/4] 清理临时文件..."
rm "$KONG_CERT_DIR/csr.pem" "$KONG_CERT_DIR/ext.cnf"
echo "  ✓ 临时文件已清理"
echo ""

# 设置文件权限
chmod 600 "$KONG_CERT_DIR"/*.pem
echo "✓ 文件权限已设置为 600"
echo ""

# 验证证书
echo "验证证书..."
if openssl verify -CAfile "$CA_CERT" "$KONG_CERT_DIR/cert.pem" > /dev/null 2>&1; then
    echo "  ✅ 证书验证成功"
else
    echo "  ❌ 证书验证失败"
    exit 1
fi

echo ""
echo "=========================================="
echo "  Kong 证书生成完成"
echo "=========================================="
echo ""
echo "证书路径:"
echo "  - 证书: $KONG_CERT_DIR/cert.pem"
echo "  - 私钥: $KONG_CERT_DIR/key.pem"
echo "  - CA:   $CA_CERT"
echo ""
echo "下一步:"
echo "  1. 更新 docker-compose.yml（挂载证书）"
echo "  2. 运行: ENABLE_MTLS=true ./scripts/kong-setup.sh"
echo ""
echo "详细说明请参考: ../KONG_MTLS_GUIDE.md"
echo ""

#!/bin/bash
# mTLS 证书生成脚本
# 用途: 为所有微服务生成 mTLS 证书（开发/测试环境）

set -e

CERT_DIR="./certs"
CA_DIR="$CERT_DIR/ca"
SERVICES_DIR="$CERT_DIR/services"

# 证书有效期（天）
DAYS_VALID=3650  # 10年（仅用于开发环境）

# 服务列表
SERVICES=(
  "payment-gateway"
  "order-service"
  "risk-service"
  "channel-adapter"
  "merchant-service"
  "admin-service"
  "accounting-service"
  "analytics-service"
  "notification-service"
  "config-service"
  "settlement-service"
  "withdrawal-service"
  "kyc-service"
  "cashier-service"
  "merchant-auth-service"
)

echo "========================================="
echo "   mTLS 证书生成工具"
echo "========================================="
echo ""

# 创建目录
mkdir -p "$CA_DIR" "$SERVICES_DIR"

# ============================================
# 1. 生成 Root CA (Certificate Authority)
# ============================================
echo "[1/3] 生成 Root CA..."

if [ ! -f "$CA_DIR/ca-key.pem" ]; then
  # 生成 CA 私钥
  openssl genrsa -out "$CA_DIR/ca-key.pem" 4096
  echo "  ✓ CA 私钥已生成: $CA_DIR/ca-key.pem"

  # 生成 CA 证书（自签名）
  openssl req -new -x509 -days $DAYS_VALID \
    -key "$CA_DIR/ca-key.pem" \
    -out "$CA_DIR/ca-cert.pem" \
    -subj "/C=US/ST=California/L=San Francisco/O=Payment Platform/OU=Security/CN=Payment Platform Root CA"

  echo "  ✓ CA 证书已生成: $CA_DIR/ca-cert.pem"
else
  echo "  ⊙ CA 证书已存在，跳过生成"
fi

echo ""

# ============================================
# 2. 为每个服务生成证书
# ============================================
echo "[2/3] 生成服务证书..."

for service in "${SERVICES[@]}"; do
  SERVICE_DIR="$SERVICES_DIR/$service"
  mkdir -p "$SERVICE_DIR"

  if [ -f "$SERVICE_DIR/cert.pem" ]; then
    echo "  ⊙ $service 证书已存在，跳过"
    continue
  fi

  # 生成服务私钥
  openssl genrsa -out "$SERVICE_DIR/key.pem" 2048

  # 生成证书签名请求 (CSR)
  openssl req -new \
    -key "$SERVICE_DIR/key.pem" \
    -out "$SERVICE_DIR/csr.pem" \
    -subj "/C=US/ST=California/L=San Francisco/O=Payment Platform/OU=Services/CN=$service"

  # 创建扩展配置（支持服务名和 localhost）
  cat > "$SERVICE_DIR/ext.cnf" <<EOF
subjectAltName = DNS:$service,DNS:localhost,DNS:$service.default.svc.cluster.local,IP:127.0.0.1
extendedKeyUsage = serverAuth,clientAuth
EOF

  # 使用 CA 签名证书
  openssl x509 -req -days $DAYS_VALID \
    -in "$SERVICE_DIR/csr.pem" \
    -CA "$CA_DIR/ca-cert.pem" \
    -CAkey "$CA_DIR/ca-key.pem" \
    -CAcreateserial \
    -out "$SERVICE_DIR/cert.pem" \
    -extfile "$SERVICE_DIR/ext.cnf"

  # 清理临时文件
  rm "$SERVICE_DIR/csr.pem" "$SERVICE_DIR/ext.cnf"

  echo "  ✓ $service 证书已生成"
done

echo ""

# ============================================
# 3. 生成验证报告
# ============================================
echo "[3/3] 生成验证报告..."

cat > "$CERT_DIR/README.md" <<EOF
# mTLS 证书说明

## 证书结构

\`\`\`
$CERT_DIR/
├── ca/
│   ├── ca-cert.pem       # Root CA 证书（所有服务需要）
│   └── ca-key.pem        # Root CA 私钥（仅生成证书时使用，生产环境需安全保管）
└── services/
    ├── payment-gateway/
    │   ├── cert.pem      # 服务证书
    │   └── key.pem       # 服务私钥
    ├── order-service/
    │   ├── cert.pem
    │   └── key.pem
    └── ...
\`\`\`

## 使用方法

### 服务端配置（以 order-service 为例）

\`\`\`bash
export TLS_CERT_FILE=./certs/services/order-service/cert.pem
export TLS_KEY_FILE=./certs/services/order-service/key.pem
export TLS_CA_FILE=./certs/ca/ca-cert.pem
export ENABLE_MTLS=true
\`\`\`

### 客户端配置（以 payment-gateway 为例）

\`\`\`bash
export TLS_CLIENT_CERT=./certs/services/payment-gateway/cert.pem
export TLS_CLIENT_KEY=./certs/services/payment-gateway/key.pem
export TLS_CA_FILE=./certs/ca/ca-cert.pem
\`\`\`

## 证书验证

\`\`\`bash
# 查看 CA 证书信息
openssl x509 -in ca/ca-cert.pem -noout -text

# 查看服务证书信息
openssl x509 -in services/order-service/cert.pem -noout -text

# 验证证书链
openssl verify -CAfile ca/ca-cert.pem services/order-service/cert.pem
\`\`\`

## 安全建议

- **开发环境**: 使用此脚本生成的证书
- **生产环境**: 使用专业 CA（如 Let's Encrypt, DigiCert）或企业 PKI
- **证书轮换**: 建议每 90 天轮换一次证书
- **私钥保护**: 严格控制 \`.pem\` 文件权限（chmod 600）

## 证书有效期

- Root CA: $DAYS_VALID 天
- 服务证书: $DAYS_VALID 天

生成时间: $(date)
EOF

echo "  ✓ 说明文档已生成: $CERT_DIR/README.md"
echo ""

# ============================================
# 4. 设置文件权限
# ============================================
echo "[安全] 设置证书文件权限..."
chmod 600 "$CA_DIR"/*.pem
find "$SERVICES_DIR" -name "*.pem" -exec chmod 600 {} \;
echo "  ✓ 所有私钥文件权限已设置为 600"
echo ""

# ============================================
# 5. 验证证书
# ============================================
echo "[验证] 验证证书链..."
VERIFIED=0
FAILED=0

for service in "${SERVICES[@]}"; do
  if openssl verify -CAfile "$CA_DIR/ca-cert.pem" "$SERVICES_DIR/$service/cert.pem" > /dev/null 2>&1; then
    ((VERIFIED++))
  else
    echo "  ✗ $service 证书验证失败"
    ((FAILED++))
  fi
done

echo "  ✓ 证书验证完成: $VERIFIED 成功, $FAILED 失败"
echo ""

# ============================================
# 完成
# ============================================
echo "========================================="
echo "   证书生成完成！"
echo "========================================="
echo ""
echo "证书目录: $CERT_DIR"
echo "CA 证书:   $CA_DIR/ca-cert.pem"
echo "服务证书:  $SERVICES_DIR/{service-name}/cert.pem"
echo ""
echo "下一步:"
echo "  1. 查看 $CERT_DIR/README.md 了解使用方法"
echo "  2. 配置环境变量启用 mTLS"
echo "  3. 重启所有服务"
echo ""

# mTLS 证书说明

## 证书结构

```
./certs/
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
```

## 使用方法

### 服务端配置（以 order-service 为例）

```bash
export TLS_CERT_FILE=./certs/services/order-service/cert.pem
export TLS_KEY_FILE=./certs/services/order-service/key.pem
export TLS_CA_FILE=./certs/ca/ca-cert.pem
export ENABLE_MTLS=true
```

### 客户端配置（以 payment-gateway 为例）

```bash
export TLS_CLIENT_CERT=./certs/services/payment-gateway/cert.pem
export TLS_CLIENT_KEY=./certs/services/payment-gateway/key.pem
export TLS_CA_FILE=./certs/ca/ca-cert.pem
```

## 证书验证

```bash
# 查看 CA 证书信息
openssl x509 -in ca/ca-cert.pem -noout -text

# 查看服务证书信息
openssl x509 -in services/order-service/cert.pem -noout -text

# 验证证书链
openssl verify -CAfile ca/ca-cert.pem services/order-service/cert.pem
```

## 安全建议

- **开发环境**: 使用此脚本生成的证书
- **生产环境**: 使用专业 CA（如 Let's Encrypt, DigiCert）或企业 PKI
- **证书轮换**: 建议每 90 天轮换一次证书
- **私钥保护**: 严格控制 `.pem` 文件权限（chmod 600）

## 证书有效期

- Root CA: 3650 天
- 服务证书: 3650 天

生成时间: Fri Oct 24 05:30:29 PM UTC 2025

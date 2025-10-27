# Payment Platform - mTLS 证书配置指南

本目录包含支付平台所有微服务的 mTLS (mutual TLS) 证书，用于服务间的双向认证。

## 📁 证书目录结构

```
./certs/
├── ca/
│   ├── ca-cert.pem          # Root CA 证书（所有服务共用）
│   ├── ca-key.pem           # Root CA 私钥（仅生成证书时使用）
│   └── ca-cert.srl          # CA 证书序列号
│
└── services/                # 19个微服务的证书
    ├── admin-bff-service/
    │   ├── admin-bff-service.crt      # 服务证书
    │   └── admin-bff-service.key      # 服务私钥
    ├── merchant-bff-service/
    │   ├── merchant-bff-service.crt
    │   └── merchant-bff-service.key
    ├── payment-gateway/
    │   ├── payment-gateway.crt
    │   └── payment-gateway.key
    ├── order-service/
    │   ├── order-service.crt
    │   └── order-service.key
    ├── channel-adapter/
    │   ├── channel-adapter.crt
    │   └── channel-adapter.key
    ├── risk-service/
    │   ├── risk-service.crt
    │   └── risk-service.key
    ├── accounting-service/
    │   ├── accounting-service.crt
    │   └── accounting-service.key
    ├── notification-service/
    │   ├── notification-service.crt
    │   └── notification-service.key
    ├── analytics-service/
    │   ├── analytics-service.crt
    │   └── analytics-service.key
    ├── config-service/
    │   ├── config-service.crt
    │   └── config-service.key
    ├── merchant-auth-service/
    │   ├── merchant-auth-service.crt
    │   └── merchant-auth-service.key
    ├── merchant-policy-service/
    │   ├── merchant-policy-service.crt
    │   └── merchant-policy-service.key
    ├── settlement-service/
    │   ├── settlement-service.crt
    │   └── settlement-service.key
    ├── withdrawal-service/
    │   ├── withdrawal-service.crt
    │   └── withdrawal-service.key
    ├── kyc-service/
    │   ├── kyc-service.crt
    │   └── kyc-service.key
    ├── cashier-service/
    │   ├── cashier-service.crt
    │   └── cashier-service.key
    ├── reconciliation-service/
    │   ├── reconciliation-service.crt
    │   └── reconciliation-service.key
    ├── dispute-service/
    │   ├── dispute-service.crt
    │   └── dispute-service.key
    └── merchant-quota-service/
        ├── merchant-quota-service.crt
        └── merchant-quota-service.key
```

## 🎯 19个微服务列表

| 服务名称 | 端口 | 证书状态 | 说明 |
|---------|------|---------|------|
| admin-bff-service | 40001 | ✅ 已生成 | 管理员 BFF 聚合服务 |
| merchant-bff-service | 40023 | ✅ 已生成 | 商户 BFF 聚合服务 |
| payment-gateway | 40003 | ✅ 已生成 | 支付网关（核心编排） |
| order-service | 40004 | ✅ 已生成 | 订单服务 |
| channel-adapter | 40005 | ✅ 已生成 | 支付渠道适配器 |
| risk-service | 40006 | ✅ 已生成 | 风险控制服务 |
| accounting-service | 40007 | ✅ 已生成 | 会计核算服务 |
| notification-service | 40008 | ✅ 已生成 | 通知服务 |
| analytics-service | 40009 | ✅ 已生成 | 数据分析服务 |
| config-service | 40010 | ✅ 已生成 | 配置管理服务 |
| merchant-auth-service | 40011 | ✅ 已生成 | 商户认证服务 |
| merchant-policy-service | 40012 | ✅ 已生成 | 商户策略服务 |
| settlement-service | 40013 | ✅ 已生成 | 结算服务 |
| withdrawal-service | 40014 | ✅ 已生成 | 提现服务 |
| kyc-service | 40015 | ✅ 已生成 | KYC 验证服务 |
| cashier-service | 40016 | ✅ 已生成 | 收银台服务 |
| reconciliation-service | 40020 | ✅ 已生成 | 对账服务 |
| dispute-service | 40021 | ✅ 已生成 | 争议处理服务 |
| merchant-quota-service | 40022 | ✅ 已生成 | 商户配额服务 |

**验证结果**: ✅ 所有19个服务的证书已生成并验证通过

## 🔧 使用方法

### 1. 服务端配置（启用 mTLS 服务器）

每个服务需要配置以下环境变量：

**示例 1: payment-gateway (端口 40003)**

```bash
# 启用 mTLS
export ENABLE_MTLS=true

# 服务端证书（接受客户端连接）
export TLS_CERT_FILE=./certs/services/payment-gateway/payment-gateway.crt
export TLS_KEY_FILE=./certs/services/payment-gateway/payment-gateway.key

# CA 证书（验证客户端）
export TLS_CA_FILE=./certs/ca/ca-cert.pem
```

**示例 2: order-service (端口 40004)**

```bash
export ENABLE_MTLS=true
export TLS_CERT_FILE=./certs/services/order-service/order-service.crt
export TLS_KEY_FILE=./certs/services/order-service/order-service.key
export TLS_CA_FILE=./certs/ca/ca-cert.pem
```

### 2. 客户端配置（调用其他服务）

当服务A需要调用服务B时，服务A作为客户端需要提供自己的证书：

**示例: payment-gateway 调用 order-service**

```bash
# 客户端证书（用于调用其他服务）
export TLS_CLIENT_CERT=./certs/services/payment-gateway/payment-gateway.crt
export TLS_CLIENT_KEY=./certs/services/payment-gateway/payment-gateway.key

# CA 证书（验证服务端）
export TLS_CA_FILE=./certs/ca/ca-cert.pem
```

### 3. 完整的服务配置示例

**在 .env 文件中配置 payment-gateway:**

```bash
# 服务基础配置
SERVICE_NAME=payment-gateway
PORT=40003

# mTLS 配置
ENABLE_MTLS=true

# 服务端证书（接受其他服务调用）
TLS_CERT_FILE=./certs/services/payment-gateway/payment-gateway.crt
TLS_KEY_FILE=./certs/services/payment-gateway/payment-gateway.key

# 客户端证书（调用其他服务时使用）
TLS_CLIENT_CERT=./certs/services/payment-gateway/payment-gateway.crt
TLS_CLIENT_KEY=./certs/services/payment-gateway/payment-gateway.key

# CA 证书（验证对方身份）
TLS_CA_FILE=./certs/ca/ca-cert.pem

# 可选配置
TLS_INSECURE_SKIP_VERIFY=false  # 生产环境必须为 false
```

## 📝 证书验证命令

### 查看 CA 证书信息

```bash
openssl x509 -in ca/ca-cert.pem -noout -text
openssl x509 -in ca/ca-cert.pem -noout -subject -issuer -dates
```

### 查看服务证书信息

```bash
# 查看 payment-gateway 证书
openssl x509 -in services/payment-gateway/payment-gateway.crt -noout -text

# 查看证书主题和有效期
openssl x509 -in services/payment-gateway/payment-gateway.crt -noout -subject -dates

# 查看证书 SAN (Subject Alternative Names)
openssl x509 -in services/payment-gateway/payment-gateway.crt -noout -text | grep -A 5 "Subject Alternative Name"
```

### 验证证书链

```bash
# 验证单个服务证书
openssl verify -CAfile ca/ca-cert.pem services/payment-gateway/payment-gateway.crt

# 批量验证所有服务证书
for cert in services/*/*.crt; do
    echo "验证: $cert"
    openssl verify -CAfile ca/ca-cert.pem "$cert"
done
```

### 测试 mTLS 连接

```bash
# 使用 curl 测试 mTLS 连接
curl -v \
  --cert ./certs/services/payment-gateway/payment-gateway.crt \
  --key ./certs/services/payment-gateway/payment-gateway.key \
  --cacert ./certs/ca/ca-cert.pem \
  https://localhost:40003/health

# 使用 openssl s_client 测试
openssl s_client \
  -connect localhost:40003 \
  -cert ./certs/services/payment-gateway/payment-gateway.crt \
  -key ./certs/services/payment-gateway/payment-gateway.key \
  -CAfile ./certs/ca/ca-cert.pem
```

## 🔐 安全最佳实践

### 文件权限设置

```bash
# CA 私钥（最高权限保护）
chmod 600 ca/ca-key.pem

# 服务私钥
chmod 600 services/*//*.key

# 证书文件（可读）
chmod 644 ca/ca-cert.pem
chmod 644 services/*/*.crt
```

### 证书轮换策略

- **开发环境**: 证书有效期 10 年，无需频繁轮换
- **生产环境**: 建议每 90 天轮换一次证书
- **自动化**: 使用 cert-manager 或类似工具自动管理证书生命周期

### 生产环境建议

1. **使用专业 CA**:
   - Let's Encrypt (免费自动化)
   - DigiCert, Sectigo (商业CA)
   - 企业内部 PKI

2. **证书存储**:
   - 使用 HashiCorp Vault 或 AWS Secrets Manager 存储私钥
   - 不要将私钥提交到 Git 仓库

3. **监控和告警**:
   - 监控证书过期时间
   - 提前 30 天发送告警

4. **审计日志**:
   - 记录所有 mTLS 连接尝试
   - 追踪证书使用情况

## 📊 证书信息

### CA 证书

- **颁发者**: Payment Platform Root CA
- **有效期**: 10 年 (3650 天)
- **密钥长度**: RSA 2048 位
- **签名算法**: SHA256-RSA

### 服务证书

- **颁发者**: Payment Platform Root CA
- **有效期**: 10 年 (3650 天)
- **密钥长度**: RSA 2048 位
- **用途**: serverAuth (服务器认证) + clientAuth (客户端认证)
- **SAN (Subject Alternative Names)**:
  - DNS: {service-name}
  - DNS: localhost
  - IP: 127.0.0.1

## 🛠️ 证书重新生成

如果需要重新生成所有证书：

```bash
# 1. 备份现有证书
cd backend
cp -r certs certs.backup.$(date +%Y%m%d)

# 2. 删除现有证书（保留目录结构）
rm -f certs/ca/ca-cert.pem certs/ca/ca-key.pem
rm -f certs/services/*/*.crt certs/services/*/*.key

# 3. 运行证书生成脚本
./scripts/generate-mtls-certs.sh

# 4. 验证新证书
cd certs
for cert in services/*/*.crt; do
    openssl verify -CAfile ca/ca-cert.pem "$cert"
done
```

## 🔍 故障排查

### 常见错误

**1. "certificate signed by unknown authority"**
- 原因: CA 证书路径不正确或未配置
- 解决: 检查 `TLS_CA_FILE` 环境变量

**2. "tls: failed to verify certificate: x509: certificate has expired"**
- 原因: 证书已过期
- 解决: 重新生成证书

**3. "tls: bad certificate"**
- 原因: 客户端证书无效或路径错误
- 解决: 检查 `TLS_CLIENT_CERT` 和 `TLS_CLIENT_KEY` 配置

**4. "remote error: tls: unknown certificate authority"**
- 原因: 服务端无法验证客户端证书
- 解决: 确保服务端配置了正确的 CA 证书

### 调试技巧

```bash
# 启用 TLS 调试日志（Go 服务）
export GODEBUG=x509roots=1,tls13=1

# 检查证书匹配
openssl x509 -in services/payment-gateway/payment-gateway.crt -noout -modulus | openssl md5
openssl rsa -in services/payment-gateway/payment-gateway.key -noout -modulus | openssl md5
# 两个 MD5 值应该相同
```

## 📚 相关文档

- [pkg/tls/config.go](../pkg/tls/config.go) - TLS 配置加载逻辑
- [pkg/httpclient/client.go](../pkg/httpclient/client.go) - HTTP 客户端 mTLS 支持
- [pkg/app/bootstrap.go](../pkg/app/bootstrap.go) - Bootstrap 框架 mTLS 集成
- [scripts/generate-mtls-certs.sh](../scripts/generate-mtls-certs.sh) - 证书生成脚本

## 📅 更新记录

- **2025-10-27**:
  - ✅ 为19个微服务生成 mTLS 证书
  - ✅ 统一证书命名格式（{service-name}.crt / {service-name}.key）
  - ✅ 新增6个服务证书（admin-bff, merchant-bff, reconciliation, dispute, merchant-policy, merchant-quota）
  - ✅ 清理冗余的旧服务证书
  - ✅ 所有证书验证通过 (19/19)

- **2024-10-24**: 初始生成 CA 证书和13个基础服务证书

---

**生成时间**: 2025-10-27 01:00 UTC
**证书总数**: 19个服务 + 1个CA = 20个证书
**验证状态**: ✅ 全部通过

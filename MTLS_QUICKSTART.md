# mTLS 快速入门指南

5 分钟快速启用 Payment Platform mTLS 服务间认证。

---

## 🚀 快速开始（3 步）

### 步骤 1: 生成证书

```bash
cd backend
./scripts/generate-mtls-certs.sh
```

**输出**:
```
✓ CA 证书已生成: ./certs/ca/ca-cert.pem
✓ payment-gateway 证书已生成
✓ order-service 证书已生成
✓ ... (共 15 个服务)
✓ 证书验证完成: 15 成功
```

---

### 步骤 2: 启动服务（启用 mTLS）

**方式 A: 使用启动脚本（推荐）**

```bash
# Terminal 1: Order Service
./scripts/start-service-mtls.sh order-service

# Terminal 2: Risk Service
./scripts/start-service-mtls.sh risk-service

# Terminal 3: Payment Gateway
./scripts/start-service-mtls.sh payment-gateway
```

**方式 B: 手动配置环境变量**

```bash
export ENABLE_MTLS=true
export TLS_CERT_FILE=$(pwd)/certs/services/order-service/cert.pem
export TLS_KEY_FILE=$(pwd)/certs/services/order-service/key.pem
export TLS_CA_FILE=$(pwd)/certs/ca/ca-cert.pem

cd services/order-service
go run cmd/main.go
```

**日志输出（成功）**:
```
INFO  正在启动 order-service...
INFO  mTLS 服务间认证已启用
INFO  HTTP 服务器已启用 mTLS
INFO  order-service HTTPS服务器(mTLS)正在监听 :40004
```

---

### 步骤 3: 验证 mTLS

```bash
./scripts/test-mtls.sh
```

**预期输出**:
```
========================================
   mTLS 功能测试
========================================

✓ 证书文件检查通过

---------------------------------------
测试: order-service (端口 40004)
---------------------------------------
[测试1] 使用有效证书访问 /health ...
  ✅ 成功: 有效证书可以访问
[测试2] 不带证书访问 /health ...
  ✅ 成功: 无证书被拒绝（符合预期）
```

---

## ✅ 验证清单

- [ ] 证书已生成（`ls certs/ca/ca-cert.pem`）
- [ ] 服务启动日志显示 "mTLS 服务间认证已启用"
- [ ] 服务启动日志显示 "HTTPS服务器(mTLS)正在监听"
- [ ] 测试脚本显示 "✅ 成功: 有效证书可以访问"
- [ ] 测试脚本显示 "✅ 成功: 无证书被拒绝"

---

## 🧪 手动测试示例

### 测试 1: 使用有效证书访问（应该成功）

```bash
curl -v https://localhost:40004/health \
  --cacert certs/ca/ca-cert.pem \
  --cert certs/services/payment-gateway/cert.pem \
  --key certs/services/payment-gateway/key.pem
```

**预期**: HTTP 200 + `{"status":"healthy"}`

---

### 测试 2: 不带证书访问（应该失败）

```bash
curl -v https://localhost:40004/health \
  --cacert certs/ca/ca-cert.pem
```

**预期**: SSL handshake failed

```
curl: (56) OpenSSL SSL_read: error:1409445C:SSL routines:ssl3_read_bytes:tlsv13 alert certificate required
```

---

### 测试 3: 使用无效证书访问（应该失败）

```bash
# 生成假证书
openssl req -x509 -newkey rsa:2048 -keyout /tmp/fake-key.pem -out /tmp/fake-cert.pem -days 1 -nodes -subj "/CN=fake"

# 尝试访问
curl -v https://localhost:40004/health \
  --cacert certs/ca/ca-cert.pem \
  --cert /tmp/fake-cert.pem \
  --key /tmp/fake-key.pem
```

**预期**: Certificate verification failed

---

## 🔧 配置说明

### 环境变量清单

| 变量 | 必填 | 说明 | 示例 |
|-----|------|------|------|
| `ENABLE_MTLS` | ✅ | 启用 mTLS | `true` |
| `TLS_CERT_FILE` | ✅ | 服务端证书 | `./certs/services/order-service/cert.pem` |
| `TLS_KEY_FILE` | ✅ | 服务端私钥 | `./certs/services/order-service/key.pem` |
| `TLS_CA_FILE` | ✅ | CA 证书 | `./certs/ca/ca-cert.pem` |
| `TLS_CLIENT_CERT` | ⚠️ | 客户端证书（仅客户端） | `./certs/services/payment-gateway/cert.pem` |
| `TLS_CLIENT_KEY` | ⚠️ | 客户端私钥（仅客户端） | `./certs/services/payment-gateway/key.pem` |

⚠️ = 仅 Payment Gateway 等发起 HTTP 调用的服务需要

---

### 服务角色说明

| 服务 | 角色 | 需要配置 |
|-----|------|---------|
| Payment Gateway | 客户端 + 服务端 | `TLS_CERT_FILE` + `TLS_CLIENT_CERT` |
| Order Service | 仅服务端 | `TLS_CERT_FILE` |
| Risk Service | 仅服务端 | `TLS_CERT_FILE` |
| Channel Adapter | 仅服务端 | `TLS_CERT_FILE` |

---

## 📂 文件结构

```
backend/
├── certs/                          # 所有证书（已生成）
│   ├── ca/
│   │   ├── ca-cert.pem            # ✅ Root CA 证书
│   │   └── ca-key.pem             # 🔐 Root CA 私钥（保密）
│   └── services/
│       ├── payment-gateway/
│       │   ├── cert.pem           # ✅ 服务证书
│       │   └── key.pem            # 🔐 服务私钥
│       ├── order-service/
│       │   ├── cert.pem
│       │   └── key.pem
│       └── ...
│
├── scripts/
│   ├── generate-mtls-certs.sh    # 🔨 证书生成脚本
│   ├── start-service-mtls.sh     # 🚀 服务启动脚本
│   └── test-mtls.sh               # 🧪 测试验证脚本
│
├── .env.mtls.example              # 📝 环境变量模板
└── services/
    └── */cmd/main.go              # ✅ 已添加 EnableMTLS 配置
```

---

## ❓ 常见问题

### Q1: 启动时报错 "TLS_CERT_FILE 未配置"

**原因**: 环境变量未设置

**解决**:
```bash
export ENABLE_MTLS=true
export TLS_CERT_FILE=$(pwd)/certs/services/order-service/cert.pem
export TLS_KEY_FILE=$(pwd)/certs/services/order-service/key.pem
export TLS_CA_FILE=$(pwd)/certs/ca/ca-cert.pem
```

---

### Q2: curl 提示 "certificate signed by unknown authority"

**原因**: 未指定 CA 证书

**解决**:
```bash
curl https://localhost:40004/health \
  --cacert certs/ca/ca-cert.pem \  # ⬅️ 添加这行
  --cert certs/services/payment-gateway/cert.pem \
  --key certs/services/payment-gateway/key.pem
```

---

### Q3: 服务间调用失败 "x509: certificate signed by unknown authority"

**原因**: Payment Gateway 未配置客户端证书

**解决**:
```bash
# Payment Gateway 需要额外配置
export TLS_CLIENT_CERT=$(pwd)/certs/services/payment-gateway/cert.pem
export TLS_CLIENT_KEY=$(pwd)/certs/services/payment-gateway/key.pem
```

---

### Q4: 如何临时禁用 mTLS（调试）

```bash
export ENABLE_MTLS=false
# 或者直接不设置该环境变量
```

---

### Q5: 证书过期怎么办

```bash
# 检查过期时间
openssl x509 -in certs/services/order-service/cert.pem -noout -dates

# 重新生成所有证书
rm -rf certs/
./scripts/generate-mtls-certs.sh
```

---

## 🎯 下一步

- [ ] 阅读完整部署文档: [MTLS_DEPLOYMENT_GUIDE.md](MTLS_DEPLOYMENT_GUIDE.md)
- [ ] 配置生产环境证书（使用专业 CA）
- [ ] 配置 Kubernetes Secrets（如果部署到 K8s）
- [ ] 设置证书轮换策略（90 天）
- [ ] 配置监控告警（证书过期告警）

---

## 🔗 相关资源

- [证书生成脚本](backend/scripts/generate-mtls-certs.sh)
- [服务启动脚本](backend/scripts/start-service-mtls.sh)
- [测试验证脚本](backend/scripts/test-mtls.sh)
- [环境变量模板](backend/.env.mtls.example)
- [完整部署文档](MTLS_DEPLOYMENT_GUIDE.md)

---

**最后更新**: 2025-01-20
**维护者**: Platform Team

# Kong mTLS 快速参考

---

## ✅ 是的，Kong 需要配置以支持 mTLS

当后端服务启用 mTLS 后，Kong 需要：
1. 🔐 **客户端证书**：Kong 作为 mTLS 客户端连接后端
2. 📝 **服务配置更新**：后端服务 URL 改为 HTTPS
3. 🐳 **Docker 挂载证书**：Kong 容器需要访问证书文件

---

## 🚀 快速配置（3 步）

### 步骤 1: 生成 Kong 证书

```bash
cd backend
./scripts/setup-kong-mtls-cert.sh
```

**输出**:
```
✅ 证书验证成功
证书路径:
  - certs/services/kong-gateway/cert.pem
  - certs/services/kong-gateway/key.pem
```

---

### 步骤 2: 更新 docker-compose.yml

在 `docker-compose.yml` 的 Kong 服务中添加：

```yaml
services:
  kong:
    image: kong:3.4-alpine
    environment:
      # ... 现有配置 ...

      # ⬇️ 新增 mTLS 配置
      KONG_CLIENT_SSL: "on"
      KONG_CLIENT_SSL_CERT: /kong/certs/kong-gateway/cert.pem
      KONG_CLIENT_SSL_CERT_KEY: /kong/certs/kong-gateway/key.pem
      KONG_LUA_SSL_TRUSTED_CERTIFICATE: /kong/certs/ca/ca-cert.pem
      KONG_LUA_SSL_VERIFY_DEPTH: 2

    volumes:
      - ./kong/declarative:/kong/declarative:ro

      # ⬇️ 新增证书挂载
      - ./backend/certs:/kong/certs:ro
```

---

### 步骤 3: 修改 kong-setup.sh（使用 HTTPS URL）

在 `backend/scripts/kong-setup.sh` 中修改服务 URL：

```bash
# 原来（HTTP）:
create_or_update_service "order-service" "http://host.docker.internal:40004"

# ⬇️ 改为（HTTPS）:
create_or_update_service "order-service" "https://host.docker.internal:40004"
```

**或者**使用环境变量切换：

```bash
if [ "${ENABLE_MTLS:-false}" == "true" ]; then
    create_or_update_service "order-service" "https://host.docker.internal:40004"
else
    create_or_update_service "order-service" "http://host.docker.internal:40004"
fi
```

---

## 🧪 验证

### 1. 重启 Kong

```bash
docker-compose restart kong
```

### 2. 检查 Kong 日志

```bash
docker-compose logs kong | grep -i "ssl\|tls\|certificate"
```

**预期**: 无错误

### 3. 测试通过 Kong 访问后端

```bash
# 启动后端服务（mTLS 模式）
cd backend
./scripts/start-service-mtls.sh order-service

# 通过 Kong 访问
curl http://localhost:40080/api/v1/orders
```

**预期**: 返回订单数据（或认证错误，正常）

---

## 🎯 架构说明

```
┌─────────┐        HTTP         ┌──────┐       mTLS      ┌─────────────┐
│ Client  │ ─────────────────> │ Kong │ ───────────────> │ Order Svc   │
│         │                     │      │  (HTTPS + Cert)  │ (Port 40004)│
└─────────┘                     └──────┘                  └─────────────┘
                                    ↓
                            /kong/certs/
                            ├── kong-gateway/
                            │   ├── cert.pem  (客户端证书)
                            │   └── key.pem
                            └── ca/
                                └── ca-cert.pem (验证后端)
```

**说明**:
- 客户端 → Kong: **HTTP**（简单，无需客户端证书）
- Kong → 后端: **mTLS**（双向认证，安全）

---

## ❓ 常见问题

### Q1: 为什么 Kong 需要配置？

A: 当后端服务启用 mTLS 后，它们只接受 HTTPS + 证书连接。Kong 必须作为 mTLS 客户端提供证书，否则后端会拒绝连接。

---

### Q2: 客户端（前端）需要证书吗？

A: **不需要**！在方案 A（推荐）中：
- 客户端 → Kong: 普通 HTTP（或 HTTPS 但无需客户端证书）
- Kong → 后端: mTLS（Kong 提供证书）

---

### Q3: 如何验证 Kong 已配置 mTLS？

```bash
# 方法 1: 检查环境变量
docker exec kong-gateway env | grep SSL

# 方法 2: 检查证书文件
docker exec kong-gateway ls -la /kong/certs/kong-gateway/

# 方法 3: 检查服务配置
curl http://localhost:40081/services/order-service | jq .url
# 应该返回: "https://host.docker.internal:40004"
```

---

### Q4: 错误 "certificate verify failed" 怎么办？

**原因**: Kong 无法验证后端证书

**解决**:
```yaml
# docker-compose.yml
environment:
  KONG_LUA_SSL_TRUSTED_CERTIFICATE: /kong/certs/ca/ca-cert.pem  # ⬅️ 确保配置
```

```bash
# 验证 CA 证书可访问
docker exec kong-gateway cat /kong/certs/ca/ca-cert.pem
```

---

### Q5: 能否只为部分服务启用 mTLS？

A: **可以**！在 `kong-setup.sh` 中选择性配置：

```bash
# mTLS 服务
create_or_update_service "order-service" "https://host.docker.internal:40004"

# 非 mTLS 服务（保持 HTTP）
create_or_update_service "notification-service" "http://host.docker.internal:40008"
```

---

## 📚 完整文档

- 详细配置指南: [KONG_MTLS_GUIDE.md](KONG_MTLS_GUIDE.md)
- 后端 mTLS 部署: [MTLS_DEPLOYMENT_GUIDE.md](MTLS_DEPLOYMENT_GUIDE.md)
- 快速入门: [MTLS_QUICKSTART.md](MTLS_QUICKSTART.md)

---

## 🔗 相关脚本

```bash
# 生成 Kong 证书
./backend/scripts/setup-kong-mtls-cert.sh

# 配置 Kong（mTLS 模式）
ENABLE_MTLS=true ./backend/scripts/kong-setup.sh

# 启动后端服务（mTLS 模式）
./backend/scripts/start-service-mtls.sh order-service

# 测试 mTLS
./backend/scripts/test-mtls.sh
```

---

**总结**: Kong 需要 3 处修改（证书、docker-compose、URL），配置简单，无需客户端证书。

**最后更新**: 2025-01-20

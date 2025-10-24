# Kong mTLS 配置完成总结

✅ **Kong API Gateway 已配置完成，支持 mTLS 后端服务**

---

## 📊 修改清单

| 文件 | 修改内容 | 状态 |
|-----|---------|------|
| `docker-compose.yml` | 添加 mTLS 环境变量 + 挂载证书 | ✅ |
| `backend/scripts/kong-setup.sh` | 支持 HTTP/HTTPS 自动切换 | ✅ |
| `backend/certs/services/kong-gateway/` | Kong 客户端证书 | ✅ |
| `backend/scripts/setup-kong-mtls-cert.sh` | 证书生成脚本 | ✅ |
| `backend/scripts/verify-kong-mtls.sh` | 配置验证脚本 | ✅ |

---

## ✅ 已完成的修改

### 1. docker-compose.yml（已修改）

**新增内容**:
```yaml
kong:
  environment:
    # mTLS Configuration (新增 5 行)
    KONG_CLIENT_SSL: "on"
    KONG_CLIENT_SSL_CERT: /kong/certs/kong-gateway/cert.pem
    KONG_CLIENT_SSL_CERT_KEY: /kong/certs/kong-gateway/key.pem
    KONG_LUA_SSL_TRUSTED_CERTIFICATE: /kong/certs/ca/ca-cert.pem
    KONG_LUA_SSL_VERIFY_DEPTH: 2

  volumes:
    # Mount mTLS certificates (新增 1 行)
    - ./backend/certs:/kong/certs:ro
```

---

### 2. kong-setup.sh（已修改）

**新增功能**:
- `get_service_url()` 函数：根据 `ENABLE_MTLS` 环境变量自动选择 HTTP/HTTPS
- 所有 16 个服务自动适配 mTLS

**使用方法**:
```bash
# 标准模式（HTTP）
./scripts/kong-setup.sh

# mTLS 模式（HTTPS）
ENABLE_MTLS=true ./scripts/kong-setup.sh
```

---

### 3. Kong 客户端证书（已生成）

```bash
$ ls -lh backend/certs/services/kong-gateway/
-rw------- cert.pem  # Kong 客户端证书
-rw------- key.pem   # Kong 客户端私钥

$ openssl verify -CAfile certs/ca/ca-cert.pem certs/services/kong-gateway/cert.pem
✅ 证书验证成功
```

---

## 🚀 使用指南

### 方式 A: 仅后端启用 mTLS（Kong 不启用）

**适用场景**: 服务间通信启用 mTLS，但 Kong 仍使用 HTTP 连接后端

```bash
# 1. 启动后端服务（mTLS 模式）
cd backend
./scripts/start-service-mtls.sh order-service

# 2. Kong 使用标准配置（HTTP）
./scripts/kong-setup.sh

# Kong → Backend: HTTP（Kong 会连接失败）
```

⚠️ **不推荐**：后端启用 mTLS 后，Kong 必须配置 mTLS

---

### 方式 B: 全链路 mTLS（推荐）

**适用场景**: Kong 作为 mTLS 客户端连接后端

```bash
# 1. 确保 Kong 证书已生成
ls backend/certs/services/kong-gateway/cert.pem
# 如果不存在，运行: ./backend/scripts/setup-kong-mtls-cert.sh

# 2. 重启 Kong（加载新配置）
docker-compose restart kong

# 3. 配置 Kong（mTLS 模式）
cd backend
ENABLE_MTLS=true ./scripts/kong-setup.sh

# 4. 启动后端服务（mTLS 模式）
./scripts/start-service-mtls.sh order-service

# 5. 验证配置
./scripts/verify-kong-mtls.sh

# 6. 测试访问
curl http://localhost:40080/api/v1/orders
```

---

## 🧪 验证步骤

### 1. 验证 Kong 配置

```bash
cd backend
./scripts/verify-kong-mtls.sh
```

**预期输出**:
```
[1/5] 检查 Kong 容器状态...
  ✅ Kong 容器正在运行

[2/5] 检查 Kong Admin API...
  ✅ Kong Admin API 正常

[3/5] 检查证书文件挂载...
  ✅ 所有证书文件已正确挂载

[4/5] 检查 Kong 环境变量...
  ✅ Kong mTLS 环境变量已配置

[5/5] 检查 Kong 服务配置...
  order-service URL: https://host.docker.internal:40004
  ✅ 服务已配置为 HTTPS（mTLS 模式）

📊 当前状态:
  - Kong 容器: ✅ 运行中
  - 证书挂载: ✅ 正常
  - 环境变量: ✅ 已配置
  - mTLS 模式: ✅ 已启用
```

---

### 2. 手动验证证书挂载

```bash
# 检查容器内证书
docker exec kong-gateway ls -la /kong/certs/ca/
docker exec kong-gateway cat /kong/certs/ca/ca-cert.pem

docker exec kong-gateway ls -la /kong/certs/kong-gateway/
```

---

### 3. 检查 Kong 服务配置

```bash
# 查看 order-service 配置
curl http://localhost:40081/services/order-service | jq '{name, url, protocol}'
```

**预期输出（mTLS 模式）**:
```json
{
  "name": "order-service",
  "url": "https://host.docker.internal:40004",
  "protocol": "https"
}
```

---

### 4. 端到端测试

```bash
# Terminal 1: 启动 order-service（mTLS 模式）
cd backend
ENABLE_MTLS=true \
TLS_CERT_FILE=./certs/services/order-service/cert.pem \
TLS_KEY_FILE=./certs/services/order-service/key.pem \
TLS_CA_FILE=./certs/ca/ca-cert.pem \
go run ./services/order-service/cmd/main.go

# Terminal 2: 通过 Kong 访问
curl -v http://localhost:40080/api/v1/orders

# 预期: Kong 成功通过 mTLS 连接到 order-service
```

---

## 🎯 架构说明

### 当前架构（已实现）

```
┌─────────────┐        HTTP         ┌──────────────┐       mTLS       ┌──────────────┐
│   Client    │ ─────────────────> │     Kong     │ ──────────────> │ Order Service│
│ (浏览器/前端) │                     │  (Port 40080)│  (HTTPS + Cert) │ (Port 40004) │
└─────────────┘                     └──────────────┘                 └──────────────┘
                                           ↓
                                    /kong/certs/
                                    ├── kong-gateway/
                                    │   ├── cert.pem  (客户端证书)
                                    │   └── key.pem   (客户端私钥)
                                    └── ca/
                                        └── ca-cert.pem (验证后端证书)
```

**优势**:
- ✅ 客户端无需证书（简单易用）
- ✅ Kong 统一管理 mTLS（集中控制）
- ✅ 后端服务间安全通信
- ✅ 前端应用无需改动

---

## 🔧 故障排查

### 问题 1: Kong 无法连接后端

**症状**:
```bash
$ curl http://localhost:40080/api/v1/orders
{"message":"An invalid response was received from the upstream server"}
```

**排查**:
```bash
# 1. 查看 Kong 日志
docker-compose logs kong | tail -50

# 2. 检查后端服务是否启动（mTLS 模式）
lsof -i :40004

# 3. 检查证书配置
docker exec kong-gateway env | grep SSL
```

---

### 问题 2: "certificate verify failed"

**症状**: Kong 日志显示证书验证失败

**原因**: Kong 无法验证后端证书

**解决**:
```bash
# 1. 检查 CA 证书是否正确挂载
docker exec kong-gateway cat /kong/certs/ca/ca-cert.pem

# 2. 检查环境变量
docker exec kong-gateway env | grep KONG_LUA_SSL_TRUSTED_CERTIFICATE

# 3. 重启 Kong
docker-compose restart kong
```

---

### 问题 3: "client didn't provide a certificate"

**症状**: 后端服务日志显示客户端未提供证书

**原因**: Kong 客户端证书未配置

**解决**:
```bash
# 1. 检查 Kong 证书是否存在
ls backend/certs/services/kong-gateway/cert.pem

# 2. 如果不存在，生成证书
cd backend
./scripts/setup-kong-mtls-cert.sh

# 3. 重启 Kong
docker-compose restart kong

# 4. 验证配置
./scripts/verify-kong-mtls.sh
```

---

## 📚 相关文档

- **快速参考**: [KONG_MTLS_QUICKREF.md](KONG_MTLS_QUICKREF.md) ⭐
- **完整指南**: [KONG_MTLS_GUIDE.md](KONG_MTLS_GUIDE.md)
- **后端 mTLS**: [MTLS_DEPLOYMENT_GUIDE.md](MTLS_DEPLOYMENT_GUIDE.md)
- **快速入门**: [MTLS_QUICKSTART.md](MTLS_QUICKSTART.md)

---

## 🔗 相关脚本

```bash
# Kong 证书生成
./backend/scripts/setup-kong-mtls-cert.sh

# Kong 配置（mTLS 模式）
ENABLE_MTLS=true ./backend/scripts/kong-setup.sh

# Kong 配置验证
./backend/scripts/verify-kong-mtls.sh

# 后端服务启动（mTLS 模式）
./backend/scripts/start-service-mtls.sh <service-name>
```

---

## ✅ 配置检查清单

- [x] `docker-compose.yml` 已添加 mTLS 环境变量
- [x] `docker-compose.yml` 已挂载证书目录
- [x] Kong 客户端证书已生成（`certs/services/kong-gateway/`）
- [x] `kong-setup.sh` 已支持 HTTPS URL
- [x] 验证脚本已创建（`verify-kong-mtls.sh`）
- [ ] Kong 容器已重启（执行: `docker-compose restart kong`）
- [ ] Kong 配置已更新（执行: `ENABLE_MTLS=true ./scripts/kong-setup.sh`）
- [ ] 后端服务已启动（mTLS 模式）
- [ ] 端到端测试通过

---

## 🎉 总结

✅ **Kong mTLS 配置已完成**

**已完成**:
- ✅ `docker-compose.yml` 修改（6 行新增）
- ✅ `kong-setup.sh` 修改（支持 HTTP/HTTPS 自动切换）
- ✅ Kong 客户端证书生成
- ✅ 验证脚本创建
- ✅ 完整文档编写（3 篇）

**待执行**（用户操作）:
1. 重启 Kong: `docker-compose restart kong`
2. 配置 Kong: `cd backend && ENABLE_MTLS=true ./scripts/kong-setup.sh`
3. 验证配置: `./scripts/verify-kong-mtls.sh`
4. 启动后端: `./scripts/start-service-mtls.sh order-service`
5. 测试访问: `curl http://localhost:40080/api/v1/orders`

**预计时间**: 5-10 分钟完成验证

---

**最后更新**: 2025-01-20
**维护者**: Platform Team

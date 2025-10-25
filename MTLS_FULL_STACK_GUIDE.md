# mTLS 全栈启动指南

一键启动 Payment Platform 完整服务（mTLS 模式）

---

## 🚀 快速开始（1 条命令）

```bash
cd backend
./scripts/start-all-mtls.sh
```

**就这么简单！** 🎉

脚本会自动完成：
1. ✅ 检查并生成所有证书
2. ✅ 启动基础设施（PostgreSQL, Redis, Kafka, Kong）
3. ✅ 配置 Kong（mTLS 模式）
4. ✅ 启动所有 16 个后端微服务（mTLS 模式）
5. ✅ 验证配置并显示状态

---

## 📋 脚本列表

| 脚本 | 用途 | 命令 |
|-----|------|------|
| `start-all-mtls.sh` | 一键启动所有服务 | `./scripts/start-all-mtls.sh` |
| `stop-all-mtls.sh` | 停止所有服务 | `./scripts/stop-all-mtls.sh` |
| `status-all-mtls.sh` | 查看服务状态 | `./scripts/status-all-mtls.sh` |
| `restart-all-mtls.sh` | 重启所有服务 | `./scripts/restart-all-mtls.sh` |

---

## 🎯 启动流程详解

### 第 1 步：检查证书（自动）

脚本会自动检查并生成：
- ✅ Root CA 证书
- ✅ 15 个后端服务证书
- ✅ Kong 客户端证书

如果证书不存在，会自动调用：
```bash
./scripts/generate-mtls-certs.sh
./scripts/setup-kong-mtls-cert.sh
```

---

### 第 2 步：启动基础设施（自动）

```bash
docker-compose up -d postgres redis zookeeper kafka kong-database kong-bootstrap kong konga
```

**包含服务**:
- PostgreSQL (port 40432)
- Redis (port 40379)
- Zookeeper + Kafka (port 40092)
- Kong Database + Bootstrap
- Kong Gateway (port 40080)
- Konga Admin UI (port 40082)

**等待时间**: ~15 秒

---

### 第 3 步：配置 Kong（自动）

```bash
ENABLE_MTLS=true ./scripts/kong-setup.sh
```

**配置内容**:
- ✅ 所有服务使用 HTTPS URL
- ✅ Kong 配置客户端证书
- ✅ 创建路由和插件
- ✅ 配置 JWT 认证

---

### 第 4 步：启动后端服务（自动）

**启动顺序**（按依赖关系）:
```
1. config-service        (40010)
2. admin-service         (40001)
3. merchant-auth-service (40011)
4. merchant-service      (40002)
5. risk-service          (40006)
6. channel-adapter       (40005)
7. order-service         (40004)
8. payment-gateway       (40003) ← 调用其他服务
9. accounting-service    (40007)
10. analytics-service    (40009)
11. notification-service (40008)
12. settlement-service   (40013)
13. withdrawal-service   (40014)
14. kyc-service          (40015)
15. cashier-service      (40016)
```

**每个服务**:
- ✅ 自动配置 mTLS 证书
- ✅ 后台运行（nohup）
- ✅ 日志输出到 `logs/<service-name>.log`
- ✅ PID 保存到 `logs/<service-name>.pid`

---

## 📊 查看状态

### 1. 使用状态脚本（推荐）

```bash
./scripts/status-all-mtls.sh
```

**输出示例**:
```
【基础设施】
  ✅ PostgreSQL       (localhost:40432)
  ✅ Redis            (localhost:40379)
  ✅ Kafka            (localhost:40092)
  ✅ Kong Gateway     (localhost:40080)
  ✅ Konga UI         (localhost:40082)

【后端微服务】
  ✅ config-service            (PID: 12345, Port: 40010)
  ✅ admin-service             (PID: 12346, Port: 40001)
  ✅ merchant-auth-service     (PID: 12347, Port: 40011)
  ...

  运行中: 15   已停止: 0

【mTLS 配置】
  ✅ CA 证书已生成
  ✅ Kong 证书已生成
  ✅ Kong mTLS 已启用

【健康检查】
  ✅ Kong Admin API 正常
  ✅ Kong Proxy 正常
  ✅ PostgreSQL 健康
  ✅ Redis 健康
```

---

### 2. 手动检查端口

```bash
# 检查所有服务端口
for port in 40001 40002 40003 40004 40005 40006 40007 40008 40009 40010 40011 40013 40014 40015 40016; do
    echo -n "Port $port: "
    if lsof -i :$port > /dev/null 2>&1; then
        echo "✅ ACTIVE"
    else
        echo "❌ NOT RUNNING"
    fi
done
```

---

### 3. 查看服务日志

```bash
# 实时查看单个服务日志
tail -f logs/order-service.log

# 查看最近 50 行
tail -50 logs/payment-gateway.log

# 查看所有服务日志（按时间）
tail -f logs/*.log
```

---

## 🧪 测试验证

### 测试 1: 通过 Kong 访问后端（推荐）

```bash
# 访问 Order Service（通过 Kong）
curl http://localhost:40080/api/v1/orders

# 访问 Payment Gateway
curl http://localhost:40080/api/v1/payments

# 带 JWT 认证
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:40080/api/v1/merchant/profile
```

**优势**: 客户端无需证书

---

### 测试 2: 直接访问后端服务（需要证书）

```bash
# 使用 mTLS 客户端证书访问
curl -v https://localhost:40004/health \
  --cacert certs/ca/ca-cert.pem \
  --cert certs/services/payment-gateway/cert.pem \
  --key certs/services/payment-gateway/key.pem

# 预期输出: {"status":"healthy"}
```

---

### 测试 3: 验证 mTLS 拒绝无证书请求

```bash
# 不带证书访问（应该失败）
curl -v https://localhost:40004/health --cacert certs/ca/ca-cert.pem

# 预期错误: SSL handshake failed
```

---

## 🛠️ 常用操作

### 停止所有服务

```bash
./scripts/stop-all-mtls.sh
```

**会停止**:
- ✅ 所有 15 个后端服务
- ✅ Kong + Konga
- ✅ Kafka + Zookeeper
- ✅ Redis
- ✅ PostgreSQL

---

### 重启所有服务

```bash
./scripts/restart-all-mtls.sh
```

等价于:
```bash
./scripts/stop-all-mtls.sh
sleep 5
./scripts/start-all-mtls.sh
```

---

### 重启单个服务

```bash
# 1. 找到进程 ID
cat logs/order-service.pid

# 2. 杀死进程
kill $(cat logs/order-service.pid)

# 3. 重新启动
cd services/order-service
ENABLE_MTLS=true \
TLS_CERT_FILE=../../certs/services/order-service/cert.pem \
TLS_KEY_FILE=../../certs/services/order-service/key.pem \
TLS_CA_FILE=../../certs/ca/ca-cert.pem \
nohup go run cmd/main.go > ../../logs/order-service.log 2>&1 &

echo $! > ../../logs/order-service.pid
```

---

### 查看 Docker 容器日志

```bash
# Kong 日志
docker-compose logs -f kong

# PostgreSQL 日志
docker-compose logs -f postgres

# Kafka 日志
docker-compose logs -f kafka

# 所有基础设施日志
docker-compose logs -f
```

---

## 🔧 故障排查

### 问题 1: 服务启动失败

**症状**: `status-all-mtls.sh` 显示服务已停止

**排查步骤**:
```bash
# 1. 查看服务日志
tail -50 logs/<service-name>.log

# 2. 检查端口是否被占用
lsof -i :<port>

# 3. 手动启动服务（看错误信息）
cd services/<service-name>
go run cmd/main.go
```

**常见原因**:
- ❌ 证书路径错误
- ❌ 数据库连接失败
- ❌ 端口被占用
- ❌ 依赖服务未启动

---

### 问题 2: Kong 无法连接后端

**症状**: `curl http://localhost:40080/api/v1/orders` 返回 502

**排查步骤**:
```bash
# 1. 检查 Kong 日志
docker-compose logs kong | tail -50

# 2. 验证 Kong mTLS 配置
./scripts/verify-kong-mtls.sh

# 3. 检查后端服务是否运行
lsof -i :40004

# 4. 检查 Kong 服务配置
curl http://localhost:40081/services/order-service | jq .url
```

**预期**: URL 应该是 `https://host.docker.internal:40004`

---

### 问题 3: 证书验证失败

**症状**: 日志显示 "certificate verify failed"

**解决**:
```bash
# 1. 重新生成所有证书
rm -rf certs/
./scripts/generate-mtls-certs.sh
./scripts/setup-kong-mtls-cert.sh

# 2. 重启所有服务
./scripts/restart-all-mtls.sh
```

---

### 问题 4: 数据库连接失败

**症状**: 日志显示 "connection refused"

**解决**:
```bash
# 1. 检查 PostgreSQL 是否运行
docker ps | grep postgres

# 2. 检查 PostgreSQL 健康
docker exec payment-postgres pg_isready -U postgres

# 3. 重启 PostgreSQL
docker-compose restart postgres

# 4. 等待 10 秒后重启服务
sleep 10
./scripts/restart-all-mtls.sh
```

---

## 📂 文件结构

```
backend/
├── scripts/
│   ├── start-all-mtls.sh      # 🚀 一键启动所有服务
│   ├── stop-all-mtls.sh       # 🛑 停止所有服务
│   ├── status-all-mtls.sh     # 📊 查看服务状态
│   ├── restart-all-mtls.sh    # 🔄 重启所有服务
│   ├── generate-mtls-certs.sh # 🔐 生成证书
│   ├── setup-kong-mtls-cert.sh # 🔐 生成 Kong 证书
│   └── verify-kong-mtls.sh    # ✅ 验证 Kong 配置
│
├── certs/                      # 证书目录
│   ├── ca/
│   │   ├── ca-cert.pem        # Root CA 证书
│   │   └── ca-key.pem         # Root CA 私钥
│   └── services/
│       ├── payment-gateway/
│       ├── order-service/
│       └── ...                # 16 个服务证书
│
├── logs/                       # 日志目录
│   ├── order-service.log      # 服务日志
│   ├── order-service.pid      # 进程 ID
│   └── ...                    # 其他服务日志
│
└── services/                   # 服务源码
    ├── payment-gateway/
    ├── order-service/
    └── ...                    # 16 个服务
```

---

## 🎉 总结

### ✅ 已完成

- ✅ 一键启动脚本（`start-all-mtls.sh`）
- ✅ 停止脚本（`stop-all-mtls.sh`）
- ✅ 状态查看脚本（`status-all-mtls.sh`）
- ✅ 重启脚本（`restart-all-mtls.sh`）
- ✅ 自动证书生成
- ✅ 自动 Kong 配置
- ✅ 按依赖顺序启动服务
- ✅ 完整日志记录

---

### 🚀 立即开始

```bash
# 1. 启动所有服务
cd backend
./scripts/start-all-mtls.sh

# 2. 查看状态
./scripts/status-all-mtls.sh

# 3. 测试访问
curl http://localhost:40080/api/v1/orders
```

---

### 📚 相关文档

- [mTLS 快速入门](MTLS_QUICKSTART.md)
- [mTLS 部署指南](MTLS_DEPLOYMENT_GUIDE.md)
- [Kong mTLS 配置](KONG_MTLS_GUIDE.md)
- [Kong mTLS 快速参考](KONG_MTLS_QUICKREF.md)

---

**最后更新**: 2025-01-20
**维护者**: Platform Team

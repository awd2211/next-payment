# mTLS 服务管理指南

本文档介绍如何使用统一的服务管理脚本 `manage-services.sh` 来管理支付平台的所有微服务和 Docker 基础设施。

## 目录

- [快速开始](#快速开始)
- [脚本功能](#脚本功能)
- [命令详解](#命令详解)
- [配置说明](#配置说明)
- [故障排查](#故障排查)

## 快速开始

### 1. 一键启动全部（推荐）

```bash
cd /home/eric/payment/backend
./scripts/manage-services.sh start
```

这个命令会：
1. 自动检查 Docker 是否运行
2. 询问是否启动 Docker 基础设施（PostgreSQL, Redis, Kafka 等）
3. 检查并生成 mTLS 证书（如果不存在）
4. 初始化数据库（如果需要）
5. 启动所有 16 个微服务
6. 验证服务启动状态

### 2. 查看服务状态

```bash
./scripts/manage-services.sh status
```

### 3. 停止所有服务

```bash
./scripts/manage-services.sh stop
```

## 脚本功能

### 核心特性

- ✅ **智能检测**: 自动检测 Docker、证书、数据库等前置条件
- ✅ **交互式配置**: 缺失组件时提示自动安装/配置
- ✅ **mTLS 认证**: 所有服务间通信使用双向 TLS 认证
- ✅ **数据库隔离**: 每个服务独立数据库（16 个数据库）
- ✅ **端口管理**: 固定端口 40001-40016，避免冲突
- ✅ **日志集中**: 所有日志统一存放在 `logs/` 目录
- ✅ **热重载**: 使用 air 实现代码热重载

### 自动化功能

1. **基础设施检测**
   - 检查 Docker 是否安装并运行
   - 检查 PostgreSQL (40432)、Redis (40379)、Kafka (40092)
   - 自动启动未运行的容器

2. **证书管理**
   - 检测 mTLS 证书是否存在
   - 提供一键生成证书选项
   - 验证每个服务的证书文件

3. **数据库管理**
   - 检测数据库连接
   - 自动初始化缺失的数据库
   - 验证所有 16 个服务数据库

4. **环境变量**
   - 自动配置 DB_PORT=40432（Docker 端口）
   - 配置 mTLS 相关环境变量
   - 设置服务间 HTTPS URL

## 命令详解

### 服务管理命令

#### `start` - 启动所有微服务

```bash
./scripts/manage-services.sh start
```

**执行流程**:
1. [1/5] 前置检查 - 验证 Docker、证书、工具
2. [2/5] 加载环境变量 - 读取 .env 文件
3. [3/5] 停止旧服务 - 清理已运行的进程
4. [4/5] 启动所有服务 - 按顺序启动 16 个微服务
5. [5/5] 验证启动状态 - 检查端口监听

**输出示例**:
```
========================================
启动所有支付平台微服务 (mTLS)
========================================

[1/5] 前置检查
✓ Docker 运行正常
✓ Docker 基础设施运行正常
✓ mTLS 证书存在
✓ air 已安装
✓ 所有前置检查通过

[2/5] 加载环境变量
✓ 环境变量配置完成 (DB_PORT=40432, mTLS=enabled)

[3/5] 停止已运行的服务
✓ 没有运行中的服务

[4/5] 启动所有微服务
  启动 config-service (端口: 40010, DB: payment_config)
  ✓ config-service 已启动 (PID: 123456)
  ...

[5/5] 验证服务启动状态
  ✓ config-service (端口: 40010)
  ✓ admin-service (端口: 40001)
  ...

========================================
启动完成！
========================================

运行中: 16 个 | 失败: 0 个
```

#### `stop` - 停止所有微服务

```bash
./scripts/manage-services.sh stop
```

**功能**:
- 停止所有 16 个微服务进程
- 清理 air 进程
- 删除临时文件（tmp/ 目录）

#### `restart` - 重启所有微服务

```bash
./scripts/manage-services.sh restart
```

等价于: `stop` + `sleep 2` + `start`

#### `status` - 查看服务状态

```bash
./scripts/manage-services.sh status
```

**输出示例**:
```
========================================
支付平台微服务状态 (mTLS)
========================================

config-service           运行中  PID: 123456  端口: 40010
admin-service            运行中  PID: 123457  端口: 40001
merchant-service         启动中  PID: 123458  端口: 40002 (等待监听)
...

========================================
总计: 15 个服务运行中, 1 个服务已停止
========================================
```

#### `logs <service>` - 查看服务日志

```bash
./scripts/manage-services.sh logs order-service
```

实时跟踪服务日志（类似 `tail -f`）。

**可用服务名称**:
- config-service
- admin-service
- merchant-service
- payment-gateway
- order-service
- channel-adapter
- risk-service
- accounting-service
- notification-service
- analytics-service
- merchant-auth-service
- merchant-config-service
- settlement-service
- withdrawal-service
- kyc-service
- cashier-service

### 基础设施管理命令

#### `infra start` - 启动 Docker 基础设施

```bash
./scripts/manage-services.sh infra start
```

**启动组件**:
- PostgreSQL (端口: 40432)
- Redis (端口: 40379)
- Kafka (端口: 40092)
- Zookeeper (端口: 2181)
- Kong Gateway (端口: 40080)
- Kong PostgreSQL (端口: 40433)

**智能检测**: 只启动未运行的组件，已运行的会跳过。

#### `infra stop` - 停止 Docker 基础设施

```bash
./scripts/manage-services.sh infra stop
```

停止所有基础设施容器（不删除）。

#### `infra status` - 查看基础设施状态

```bash
./scripts/manage-services.sh infra status
```

**输出示例**:
```
========================================
Docker 基础设施状态
========================================

✓ PostgreSQL (端口: 40432)
✓ Redis (端口: 40379)
✓ Kafka (端口: 40092)
✓ Zookeeper (端口: 2181)
✓ Kong Gateway (端口: 40080)
✓ Kong PostgreSQL (端口: 40433)
```

#### `infra restart` - 重启 Docker 基础设施

```bash
./scripts/manage-services.sh infra restart
```

## 配置说明

### 环境变量 (.env)

脚本会自动读取 `backend/.env` 文件，默认配置：

```bash
# 环境
ENV=development

# mTLS 配置
ENABLE_MTLS=true
TLS_CA_FILE=/home/eric/payment/backend/certs/ca/ca-cert.pem

# 数据库配置 (Docker 端口 40432)
DB_HOST=localhost
DB_PORT=40432
DB_USER=postgres
DB_PASSWORD=postgres

# Redis 配置 (Docker 端口 40379)
REDIS_HOST=localhost
REDIS_PORT=40379
REDIS_PASSWORD=

# Kafka 配置 (Docker 端口 40092)
KAFKA_BROKERS=localhost:40092

# JWT 配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Stripe 配置
STRIPE_API_KEY=sk_test_your_stripe_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# 服务间通信 URL (HTTPS - mTLS)
ORDER_SERVICE_URL=https://localhost:40004
RISK_SERVICE_URL=https://localhost:40006
CHANNEL_SERVICE_URL=https://localhost:40005
```

### 服务端口分配

| 服务 | 端口 | 数据库 |
|------|------|--------|
| admin-service | 40001 | payment_admin |
| merchant-service | 40002 | payment_merchant |
| payment-gateway | 40003 | payment_gateway |
| order-service | 40004 | payment_order |
| channel-adapter | 40005 | payment_channel |
| risk-service | 40006 | payment_risk |
| accounting-service | 40007 | payment_accounting |
| notification-service | 40008 | payment_notify |
| analytics-service | 40009 | payment_analytics |
| config-service | 40010 | payment_config |
| merchant-auth-service | 40011 | payment_merchant_auth |
| merchant-config-service | 40012 | payment_merchant_config |
| settlement-service | 40013 | payment_settlement |
| withdrawal-service | 40014 | payment_withdrawal |
| kyc-service | 40015 | payment_kyc |
| cashier-service | 40016 | payment_cashier |

### mTLS 证书路径

- **CA 证书**: `backend/certs/ca/ca-cert.pem`
- **服务证书**: `backend/certs/services/<service-name>/cert.pem`
- **服务私钥**: `backend/certs/services/<service-name>/key.pem`

## 故障排查

### 问题 1: Docker 服务未运行

**错误信息**:
```
✗ Docker 服务未运行
  请启动 Docker 服务: sudo systemctl start docker
```

**解决方法**:
```bash
sudo systemctl start docker
```

### 问题 2: 数据库端口错误

**症状**: 服务启动失败，日志显示 `connection refused` 到 `5432` 端口

**原因**: 使用了本地 PostgreSQL 端口而不是 Docker 端口

**解决**: 脚本已自动配置 `DB_PORT=40432`，确保没有其他地方覆盖此配置

### 问题 3: 证书缺失

**错误信息**:
```
⚠ CA 证书不存在
是否自动生成 mTLS 证书? (y/n):
```

**解决**: 输入 `y` 自动生成，或手动运行：
```bash
./scripts/generate-mtls-certs.sh
```

### 问题 4: 服务未监听端口

**症状**: `status` 显示服务 PID 存在但端口未监听

**可能原因**:
1. 服务编译失败
2. 数据库连接失败
3. 证书路径错误

**排查步骤**:
```bash
# 1. 查看服务日志
./scripts/manage-services.sh logs <service-name>

# 2. 检查数据库连接
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -d payment_<service> -c "SELECT 1"

# 3. 检查证书文件
ls -lh certs/services/<service-name>/
```

### 问题 5: air 未安装

**错误信息**:
```
⚠ air 未安装
  请运行: go install github.com/cosmtrek/air@v1.49.0
```

**解决**:
```bash
go install github.com/cosmtrek/air@v1.49.0
```

### 问题 6: 基础设施未就绪

**症状**: 服务启动后立即退出

**排查**:
```bash
# 检查基础设施状态
./scripts/manage-services.sh infra status

# 启动缺失的基础设施
./scripts/manage-services.sh infra start

# 测试 PostgreSQL 连接
PGPASSWORD=postgres psql -h localhost -p 40432 -U postgres -c "SELECT version();"

# 测试 Redis 连接
redis-cli -h localhost -p 40379 ping
```

## 测试 mTLS 连接

### 使用 curl 测试

```bash
# 测试 order-service 健康检查
curl https://localhost:40004/health \
  --cacert certs/ca/ca-cert.pem \
  --cert certs/services/payment-gateway/cert.pem \
  --key certs/services/payment-gateway/key.pem

# 预期输出:
# {"status":"healthy","checks":[...]}
```

### 验证 mTLS 强制认证

```bash
# 不带客户端证书访问（应该失败）
curl https://localhost:40004/health --cacert certs/ca/ca-cert.pem

# 预期错误:
# tlsv13 alert certificate required
```

## 日志文件位置

所有服务日志存放在 `backend/logs/` 目录：

```bash
# 查看最近 50 行日志
tail -50 logs/order-service.log

# 实时跟踪日志
./scripts/manage-services.sh logs order-service

# 或使用 tail -f
tail -f logs/order-service.log
```

## 最佳实践

### 开发工作流

1. **首次启动**:
   ```bash
   ./scripts/manage-services.sh start
   # 按提示配置基础设施和证书
   ```

2. **日常开发**:
   ```bash
   # 查看状态
   ./scripts/manage-services.sh status

   # 查看特定服务日志
   ./scripts/manage-services.sh logs payment-gateway

   # 修改代码后 air 会自动重新编译
   ```

3. **完全重启**:
   ```bash
   ./scripts/manage-services.sh restart
   ```

4. **停止所有服务**:
   ```bash
   ./scripts/manage-services.sh stop
   ```

### 生产部署建议

1. **修改 JWT Secret**:
   编辑 `.env` 文件，设置强密码：
   ```bash
   JWT_SECRET=<64位随机字符串>
   ```

2. **配置实际的 Stripe Key**:
   ```bash
   STRIPE_API_KEY=sk_live_your_production_key
   STRIPE_WEBHOOK_SECRET=whsec_your_production_secret
   ```

3. **调整 Jaeger 采样率**:
   ```bash
   JAEGER_SAMPLING_RATE=10  # 生产环境建议 10-20%
   ```

4. **使用正式 TLS 证书**:
   将自签名证书替换为 Let's Encrypt 或商业 CA 颁发的证书

## 相关脚本

- `manage-services.sh` - 统一管理脚本（推荐使用）
- `start-all-services.sh` - 旧版启动脚本（已更新为 mTLS 模式）
- `stop-all-services.sh` - 旧版停止脚本
- `status-all-services.sh` - 旧版状态查看脚本
- `generate-mtls-certs.sh` - 生成 mTLS 证书
- `init-db.sh` - 初始化数据库

## 总结

使用 `manage-services.sh` 统一管理脚本可以：

- ✅ 一键启动/停止所有服务
- ✅ 自动检测并配置所有依赖
- ✅ 智能跳过已运行的基础设施
- ✅ 提供清晰的状态反馈
- ✅ 集中管理日志查看
- ✅ 确保 mTLS 认证正确配置

享受开发！🚀

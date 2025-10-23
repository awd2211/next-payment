# Merchant Auth Service

> 商户认证和安全管理服务 - 从 merchant-service 拆分的独立微服务

## 📋 概述

Merchant Auth Service 是从 merchant-service 拆分出来的独立微服务，专注于处理商户的安全认证相关功能。遵循单一职责原则（SRP），使得安全功能可以独立演进、部署和扩展。

**状态**: ✅ **代码生成完成** - 待编译测试和数据迁移

## 🎯 核心功能

### 1. 密码管理
- ✅ 修改密码
- ✅ 密码历史记录（防止重复使用最近的密码）
- ✅ 密码强度验证
- ✅ 密码过期提醒

### 2. 双因素认证（2FA）
- ✅ TOTP (Time-based One-Time Password) 支持
- ✅ QR码生成（用于扫描绑定）
- ✅ 备用恢复代码
- ✅ 验证和启用/禁用

### 3. 登录活动记录
- ✅ 记录所有登录尝试（成功/失败）
- ✅ IP地址和地理位置追踪
- ✅ 设备类型和User-Agent记录
- ✅ 异常登录检测（新IP、新国家）

### 4. 安全设置管理
- ✅ 密码过期天数配置
- ✅ 会话超时时间配置
- ✅ 最大并发会话数限制
- ✅ IP白名单
- ✅ 国家白名单/黑名单
- ✅ 登录通知开关
- ✅ 异常行为通知开关

### 5. 会话管理
- ✅ 创建和验证会话
- ✅ 查看活跃会话
- ✅ 撤销单个会话
- ✅ 撤销所有会话（强制退出）
- ✅ 自动清理过期会话（定时任务）

## 🏗️ 架构设计

### 技术栈
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL (payment_merchant_auth)
- **缓存**: Redis
- **认证**: JWT (从 pkg/auth 导入)
- **2FA**: TOTP (github.com/pquerna/otp)
- **日志**: Zap (从 pkg/logger 导入)

### 目录结构
```
merchant-auth-service/
├── cmd/
│   └── main.go                 # 服务入口
├── internal/
│   ├── model/
│   │   └── security.go         # 5个数据模型
│   ├── repository/
│   │   └── security_repository.go  # 数据访问层
│   ├── service/
│   │   └── security_service.go     # 业务逻辑层
│   ├── handler/
│   │   └── security_handler.go     # HTTP处理器
│   └── client/
│       └── merchant_client.go      # Merchant Service客户端
├── migrations/
│   └── 001_migrate_from_merchant_service.sql  # 数据迁移脚本
├── go.mod
└── README.md
```

## ⚙️ 配置

### 端口
- **默认端口**: 8011
- **环境变量**: `PORT`

### 数据库
- **数据库名**: payment_merchant_auth
- **环境变量**:
  ```bash
  DB_HOST=localhost       # 默认: localhost
  DB_PORT=5432           # 默认: 5432
  DB_USER=postgres       # 默认: postgres
  DB_PASSWORD=postgres   # 默认: postgres
  DB_NAME=payment_merchant_auth
  DB_SSL_MODE=disable
  DB_TIMEZONE=UTC
  ```

### Redis
- **环境变量**:
  ```bash
  REDIS_HOST=localhost   # 默认: localhost
  REDIS_PORT=6379       # 默认: 6379
  REDIS_PASSWORD=       # 默认: ""
  REDIS_DB=0            # 默认: 0
  ```

### Merchant Service 客户端
- **环境变量**:
  ```bash
  MERCHANT_SERVICE_URL=http://localhost:8002
  ```

### JWT
- **环境变量**:
  ```bash
  JWT_SECRET=your-secret-key  # 生产环境必须更改
  ```

## 🚀 启动步骤

### 1. 确保依赖服务运行
```bash
# 启动 PostgreSQL 和 Redis
cd /home/eric/payment
docker-compose up -d postgres redis
```

### 2. 启动 Merchant Service（依赖）
```bash
cd backend/services/merchant-service
go run cmd/main.go
```

### 3. 首次启动（自动创建表结构）
```bash
cd backend/services/merchant-auth-service

# 安装依赖
go mod tidy

# 启动服务
go run cmd/main.go
```

### 4. 执行数据迁移（如果需要从旧数据库迁移）
```bash
# 方式1: 使用 docker exec
docker exec -i payment-postgres psql -U postgres < migrations/001_migrate_from_merchant_service.sql

# 方式2: 使用 psql 直接连接
psql -h localhost -p 40432 -U postgres -f migrations/001_migrate_from_merchant_service.sql
```

### 5. 验证服务运行
```bash
curl http://localhost:8011/health
```

预期响应：
```json
{
  "status": "ok",
  "service": "merchant-auth-service",
  "time": 1729689600
}
```

## 📡 API 端点

### 密码管理
```
PUT /api/v1/security/password
请求体: {"old_password": "xxx", "new_password": "xxx"}
```

### 双因素认证
```
POST /api/v1/security/2fa/enable      # 启用2FA，返回QR码
POST /api/v1/security/2fa/verify      # 验证2FA代码
POST /api/v1/security/2fa/disable     # 禁用2FA
```

### 安全设置
```
GET  /api/v1/security/settings        # 获取安全设置
PUT  /api/v1/security/settings        # 更新安全设置
```

### 登录活动
```
GET  /api/v1/security/login-activities?page=1&page_size=20
```

### 会话管理
```
GET    /api/v1/security/sessions                # 获取活跃会话
DELETE /api/v1/security/sessions/:session_id    # 撤销指定会话
DELETE /api/v1/security/sessions                # 撤销所有会话
```

### 系统端点
```
GET /health              # 健康检查
GET /swagger/*any        # API文档
```

## 🔄 与 Merchant Service 的交互

Merchant Auth Service 通过 HTTP 客户端调用 merchant-service 的以下接口：

### 1. 获取商户信息
```
GET /api/v1/merchants/{merchant_id}
响应: {"id", "merchant_no", "merchant_name", "status", "email", "phone"}
```

### 2. 获取带密码的商户信息（内部接口）
```
GET /api/v1/merchants/{merchant_id}/with-password
响应: {..., "password_hash": "xxx"}
```

### 3. 更新商户密码（内部接口）
```
PUT /api/v1/merchants/{merchant_id}/password
请求体: {"password_hash": "xxx"}
```

**注意**: 这些内部接口需要在 merchant-service 中实现。

## 🗄️ 数据模型

### TwoFactorAuth（双因素认证）
```sql
CREATE TABLE two_factor_auths (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    secret VARCHAR(255) NOT NULL,
    is_enabled BOOLEAN DEFAULT false,
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    backup_codes TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### LoginActivity（登录活动）
```sql
CREATE TABLE login_activities (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    ip VARCHAR(50),
    user_agent TEXT,
    device_type VARCHAR(50),
    country VARCHAR(50),
    city VARCHAR(100),
    status VARCHAR(20),
    failure_reason TEXT,
    is_abnormal BOOLEAN DEFAULT false,
    abnormal_reason TEXT,
    login_at TIMESTAMP,
    created_at TIMESTAMP
);
```

### SecuritySettings（安全设置）
```sql
CREATE TABLE security_settings (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL UNIQUE,
    password_expiry_days INT DEFAULT 90,
    password_changed_at TIMESTAMP,
    session_timeout_minutes INT DEFAULT 60,
    max_concurrent_sessions INT DEFAULT 5,
    ip_whitelist TEXT,
    allowed_countries TEXT,
    blocked_countries TEXT,
    login_notification BOOLEAN DEFAULT true,
    abnormal_notification BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### PasswordHistory（密码历史）
```sql
CREATE TABLE password_histories (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP
);
```

### Session（会话）
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL UNIQUE,
    merchant_id UUID NOT NULL,
    ip VARCHAR(50),
    user_agent TEXT,
    device_type VARCHAR(50),
    expires_at TIMESTAMP NOT NULL,
    last_seen_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## 🔧 编译和部署

### 编译
```bash
cd backend/services/merchant-auth-service
go build -o /tmp/merchant-auth-service ./cmd/main.go
```

### 运行
```bash
PORT=8011 \
DB_NAME=payment_merchant_auth \
MERCHANT_SERVICE_URL=http://localhost:8002 \
/tmp/merchant-auth-service
```

### 生成 Swagger 文档
```bash
# 安装 swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
cd backend/services/merchant-auth-service
~/go/bin/swag init -g cmd/main.go -o ./api-docs
```

## 📊 迁移路线图

### ✅ Phase 1: 代码生成（已完成）
- ✅ go.mod 创建
- ✅ 模型文件复制
- ✅ Repository层实现
- ✅ Merchant HTTP客户端
- ✅ Service层实现
- ✅ Handler层实现
- ✅ main.go入口
- ✅ 数据迁移脚本

### ⏳ Phase 2: 编译测试（下一步）
1. 取消注释 `go.work` 中的 merchant-auth-service
2. 运行 `go mod tidy`
3. 编译服务: `go build ./cmd/main.go`
4. 启动服务并测试健康检查
5. 执行数据迁移脚本

### ⏳ Phase 3: Merchant Service改造
1. 在 merchant-service 中添加3个内部接口
2. 实现双写逻辑（同时写入两个数据库）
3. 添加 Feature Flag 控制读取来源

### ⏳ Phase 4: 灰度切换
1. 部分流量切换到 merchant-auth-service
2. 监控性能和错误率
3. 逐步增加切换比例到100%

### ⏳ Phase 5: 完全切换
1. 100%流量切换到 merchant-auth-service
2. 下线 merchant-service 中的安全功能代码
3. 删除旧表数据（保留备份）

## 🐛 故障排查

### 问题1: 编译失败 - "cannot find package"
**解决**:
```bash
# 1. 确保在 go.work 中启用了该服务
cd /home/eric/payment/backend
cat go.work  # 检查是否包含 ./services/merchant-auth-service

# 2. 运行 go mod tidy
cd services/merchant-auth-service
go mod tidy

# 3. 清理缓存
go clean -cache
```

### 问题2: 无法连接 merchant-service
**症状**: 日志显示 "failed to call merchant service"
**解决**:
```bash
# 检查 merchant-service 是否运行
curl http://localhost:8002/health

# 检查环境变量
echo $MERCHANT_SERVICE_URL
```

### 问题3: 数据库连接失败
**解决**:
```bash
# 检查数据库是否存在
docker exec payment-postgres psql -U postgres -c "\l" | grep payment_merchant_auth

# 如果不存在，创建数据库
docker exec payment-postgres psql -U postgres -c "CREATE DATABASE payment_merchant_auth;"
```

### 问题4: Redis连接失败
**解决**:
```bash
# 检查 Redis 是否运行
docker ps | grep redis

# 测试连接
redis-cli -h localhost -p 6379 ping
```

## 🔒 安全建议

1. ✅ 使用强JWT密钥（至少32字符，生产环境必须更改）
2. ✅ 启用HTTPS（生产环境）
3. ✅ 配置IP白名单（如有需要）
4. ✅ 定期审查登录活动日志
5. ✅ 监控异常登录行为
6. ✅ 定期备份数据库
7. ✅ 限制失败登录次数（可在未来版本实现）

## 📚 相关文档

- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - 完整30服务架构设计
- [ROADMAP.md](../../../ROADMAP.md) - 12个月实施路线图
- [SERVICE_PORTS.md](../../docs/SERVICE_PORTS.md) - 端口分配表
- [SETUP_COMPLETE.md](../../../SETUP_COMPLETE.md) - 预留工作完成报告

## 📞 下一步行动

1. **本周**:
   - [ ] 取消注释 `go.work` 中的服务路径
   - [ ] 编译测试 `go build ./cmd/main.go`
   - [ ] 启动服务并验证健康检查
   - [ ] 执行数据迁移脚本

2. **下周**:
   - [ ] 在 merchant-service 中实现3个内部接口
   - [ ] 实现双写逻辑
   - [ ] 集成测试

3. **两周后**:
   - [ ] 灰度切换10%流量
   - [ ] 性能测试
   - [ ] 监控配置

**预计完成时间**: 2周

---

**文档版本**: v1.0
**最后更新**: 2025-10-23
**负责人**: 待定

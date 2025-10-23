# 开发文档

## 项目概述

Global Payment Platform 是一个基于 Go + gRPC 的企业级多租户支付中台系统，支持 Stripe、PayPal、加密货币等多种海外支付渠道。

### 技术架构

**后端技术栈：**
- **语言**：Go 1.21+
- **Web框架**：Gin（HTTP/REST API）
- **RPC框架**：gRPC + Protocol Buffers（微服务通信）
- **数据库**：PostgreSQL 15+（主数据库）、Redis 7+（缓存）
- **消息队列**：Kafka 3.5+（事件驱动）
- **ORM**：GORM
- **认证**：JWT
- **日志**：Zap
- **监控**：Prometheus + Grafana
- **追踪**：Jaeger

**前端技术栈：**
- **框架**：React 18 + TypeScript
- **UI库**：Ant Design / Ant Design Pro
- **状态管理**：Zustand
- **请求**：Axios

---

## 快速开始

### 1. 环境要求

- Go 1.21+
- Docker & Docker Compose
- Node.js 18+（前端开发）
- Make（可选）
- protoc（Protocol Buffers编译器）

### 2. 克隆项目

```bash
git clone <your-repo>
cd payment-platform
```

### 3. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库、Redis等
```

### 4. 启动基础设施（推荐使用Docker）

```bash
# 启动所有服务（PostgreSQL、Redis、Kafka、微服务、监控）
docker-compose up -d

# 或者只启动基础设施
docker-compose up -d postgres redis kafka
```

### 5. 初始化数据库

数据库会在首次启动时自动执行 `scripts/init-db.sql` 进行初始化。

**默认管理员账号：**
- 用户名：`admin`
- 密码：`Admin@123`
- 邮箱：`admin@payment-platform.com`

⚠️ **请登录后立即修改密码！**

### 6. 本地开发（不使用Docker）

#### 生成Proto代码

```bash
cd backend
make proto
```

#### 启动单个服务

```bash
# 启动 Admin Service
cd backend/services/admin-service
go run cmd/main.go

# 启动 Merchant Service
cd backend/services/merchant-service
go run cmd/main.go
```

#### 启动所有服务

```bash
cd backend
make run-all
```

---

## 项目结构详解

```
payment-platform/
├── backend/                          # 后端服务
│   ├── services/                     # 微服务
│   │   ├── admin-service/            # 运营管理服务（管理员、审核、配置）
│   │   │   ├── cmd/                  # 启动入口
│   │   │   │   └── main.go           # main函数
│   │   │   ├── internal/             # 内部代码（不对外暴露）
│   │   │   │   ├── handler/          # HTTP处理器（Gin）
│   │   │   │   ├── service/          # 业务逻辑层
│   │   │   │   ├── repository/       # 数据访问层（GORM）
│   │   │   │   └── model/            # 数据模型
│   │   │   ├── proto/                # gRPC接口定义
│   │   │   ├── go.mod                # Go模块依赖
│   │   │   └── Dockerfile            # Docker镜像
│   │   ├── merchant-service/         # 商户管理服务
│   │   ├── payment-gateway/          # 支付网关服务
│   │   ├── channel-adapter/          # 渠道适配服务
│   │   ├── order-service/            # 订单服务
│   │   ├── accounting-service/       # 账务服务
│   │   ├── risk-service/             # 风控服务
│   │   ├── notification-service/     # 通知服务
│   │   ├── analytics-service/        # 分析服务
│   │   └── config-service/           # 配置中心
│   │
│   ├── pkg/                          # 共享库（所有服务共用）
│   │   ├── auth/                     # JWT认证、密码加密
│   │   ├── crypto/                   # 加密解密工具
│   │   ├── db/                       # 数据库连接（PostgreSQL、Redis）
│   │   ├── kafka/                    # Kafka客户端
│   │   ├── config/                   # 配置加载
│   │   ├── logger/                   # 日志（Zap）
│   │   └── middleware/               # 中间件（认证、限流、日志等）
│   │
│   ├── proto/                        # gRPC Proto定义（所有服务）
│   │   ├── admin/                    # Admin Service Proto
│   │   ├── merchant/                 # Merchant Service Proto
│   │   ├── payment/                  # Payment Service Proto
│   │   └── order/                    # Order Service Proto
│   │
│   ├── deployments/                  # 部署配置
│   │   ├── kubernetes/               # K8s YAML
│   │   └── docker-compose/           # Docker Compose
│   │
│   └── Makefile                      # 构建脚本
│
├── frontend/                         # 前端应用
│   ├── admin-portal/                 # 运营管理后台（React + Ant Design Pro）
│   └── merchant-portal/              # 商户自助后台（React + Ant Design）
│
├── scripts/                          # 脚本工具
│   ├── init-db.sql                   # 数据库初始化脚本
│   └── init-db.sh                    # 数据库初始化Shell脚本
│
├── docs/                             # 文档
│   ├── DEVELOPMENT.md                # 开发文档
│   ├── API.md                        # API文档
│   └── ARCHITECTURE.md               # 架构文档
│
├── docker-compose.yml                # Docker Compose配置
├── .env.example                      # 环境变量模板
├── go.work                           # Go Workspace配置
└── README.md                         # 项目说明
```

---

## 微服务详解

### 1. Admin Service（运营管理服务）`:8001`

**职责：**
- 管理员账号管理（CRUD、登录、权限）
- 角色和权限管理（RBAC）
- 商户审核（KYC审核、资质审核）
- 系统配置管理
- 审批流程（提现审批、退款审批）
- 审计日志

**API示例：**
```bash
# 管理员登录
curl -X POST http://localhost:8001/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}'

# 创建管理员
curl -X POST http://localhost:8001/api/v1/admin \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username":"operator",
    "email":"operator@example.com",
    "password":"Password123",
    "full_name":"运营人员",
    "role_ids":["<role_id>"]
  }'

# 获取管理员列表
curl -X GET "http://localhost:8001/api/v1/admin?page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

### 2. Merchant Service（商户管理服务）`:8002`

**职责：**
- 商户注册、登录
- 商户信息管理
- API密钥管理
- Webhook配置
- 渠道配置（Stripe、PayPal等）

### 3. Payment Gateway（支付网关）`:8003`

**职责：**
- 统一支付入口
- 渠道路由选择
- 幂等性控制
- 支付状态管理

### 4. Order Service（订单服务）`:8004`

**职责：**
- 订单创建、查询
- 订单状态管理
- 订单统计

### 5. Channel Adapter（渠道适配）`:8005`

**职责：**
- Stripe集成
- PayPal集成
- 加密货币集成
- Webhook接收处理

---

## 数据库设计

### 核心表结构

#### admins（管理员表）
```sql
id            UUID PRIMARY KEY
username      VARCHAR(50) UNIQUE
email         VARCHAR(255) UNIQUE
password_hash VARCHAR(255)
full_name     VARCHAR(100)
status        VARCHAR(20)  -- active, disabled, locked
is_super      BOOLEAN
created_at    TIMESTAMPTZ
```

#### roles（角色表）
```sql
id           UUID PRIMARY KEY
name         VARCHAR(50) UNIQUE
display_name VARCHAR(100)
description  TEXT
is_system    BOOLEAN  -- 系统内置角色不可删除
```

#### permissions（权限表）
```sql
id          UUID PRIMARY KEY
code        VARCHAR(100) UNIQUE  -- merchant.view, payment.refund
resource    VARCHAR(50)          -- merchant, payment, order
action      VARCHAR(50)          -- view, create, edit, delete
```

#### system_configs（系统配置表）
```sql
id          UUID PRIMARY KEY
key         VARCHAR(100) UNIQUE
value       TEXT
type        VARCHAR(20)  -- string, number, boolean, json
category    VARCHAR(50)  -- payment, notification, risk
is_public   BOOLEAN
```

---

## 开发规范

### 1. 代码规范

- **命名**：使用驼峰命名（camelCase），包名使用小写
- **注释**：所有注释使用中文
- **错误处理**：统一使用`errors.New()`或自定义错误类型
- **日志**：使用结构化日志（Zap）

### 2. Git提交规范

```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
test: 测试
chore: 构建/工具链更新
```

示例：
```bash
git commit -m "feat(admin): 添加管理员批量导入功能"
git commit -m "fix(payment): 修复Stripe webhook签名验证错误"
```

### 3. API设计规范

- **RESTful风格**
- **统一响应格式**：
```json
{
  "data": {},
  "error": "",
  "request_id": "uuid"
}
```

- **HTTP状态码**：
  - `200` - 成功
  - `201` - 创建成功
  - `400` - 请求参数错误
  - `401` - 未认证
  - `403` - 权限不足
  - `404` - 资源不存在
  - `500` - 服务器错误

---

## 测试

### 单元测试

```bash
cd backend
make test
```

### 集成测试

```bash
cd backend
make test-integration
```

### 压力测试

```bash
cd backend
make load-test
```

---

## 监控与调试

### 1. 健康检查

```bash
curl http://localhost:8001/health
```

### 2. Prometheus指标

访问：http://localhost:9090

### 3. Grafana可视化

访问：http://localhost:3000
- 用户名：`admin`
- 密码：`admin`

### 4. Jaeger追踪

访问：http://localhost:16686

### 5. Traefik Dashboard

访问：http://localhost:8080

---

## 常见问题

### Q1: 数据库连接失败？

检查`.env`中的数据库配置，确保PostgreSQL已启动：
```bash
docker-compose ps postgres
docker-compose logs postgres
```

### Q2: Proto文件修改后如何重新生成？

```bash
cd backend
make clean
make proto
```

### Q3: 如何添加新的权限？

在`scripts/init-db.sql`中添加，或通过API创建：
```sql
INSERT INTO permissions (code, name, resource, action, description)
VALUES ('order.cancel', '取消订单', 'order', 'cancel', '取消未支付订单');
```

### Q4: 如何重置数据库？

```bash
docker-compose down -v
docker-compose up -d postgres
# 数据库会自动执行init-db.sql
```

---

## 下一步

1. ✅ 完成核心微服务开发
2. ⏳ 集成Stripe/PayPal支付渠道
3. ⏳ 开发Admin Portal前端
4. ⏳ 开发Merchant Portal前端
5. ⏳ 编写单元测试和集成测试
6. ⏳ 部署到Kubernetes
7. ⏳ 安全审计和性能优化

---

## 联系方式

- 技术支持：support@payment-platform.com
- 文档：https://docs.payment-platform.com

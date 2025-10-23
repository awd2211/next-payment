# 🚀 Global Payment Platform - 全球支付平台

基于 Go + gRPC 的企业级多租户支付中台系统，支持 Stripe、PayPal、加密货币等多种支付渠道。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/license-Commercial-blue.svg)](LICENSE)

## ✨ 核心特性

✅ **微服务架构** - 10个核心服务，独立部署和扩展
✅ **多租户SaaS** - 完整的商户隔离和管理
✅ **全球化支持** - 12种语言 + 32种货币 + 实时汇率
✅ **企业级安全** - 2FA、登录追踪、会话管理、RBAC
✅ **统一支付网关** - 智能路由、回调处理、退款管理
✅ **Stripe 渠道** - 完整的 Stripe 支付集成（支付、退款、Webhook）
✅ **通知服务** - SMTP/Mailgun 邮件、Twilio 短信、Webhook 回调
✅ **实时汇率** - 支持32种货币的实时转换
✅ **完整文档** - 10个详细的技术文档

## 📦 服务列表

### 核心支付服务
- **merchant-service** (`:8001`) - 商户管理、多租户、API认证
- **payment-gateway** (`:8002`) - 支付网关、路由、幂等性
- **channel-adapter** (`:8003`) - 渠道适配（Stripe/PayPal/Crypto）
- **order-service** (`:8004`) - 订单管理、状态机
- **accounting-service** (`:8005`) - 账务、清结算、复式记账

### 业务支撑服务
- **risk-service** (`:8006`) - 风控、反欺诈、限额
- **notification-service** (`:8007`) - 邮件、短信、Webhook、通知模板
- **analytics-service** (`:8008`) - 数据统计、报表
- **config-service** (`:8009`) - 配置中心、费率管理

### 前端应用
- **admin-portal** (`:3000`) - 运营管理后台（React + Ant Design Pro）
- **merchant-portal** (`:3001`) - 商户自助后台（React + Ant Design）

## 🚀 快速开始

### 前置要求
- Go 1.21+
- Docker & Docker Compose
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Kafka 3.5+

### 本地开发环境

1. 克隆项目
```bash
git clone <your-repo>
cd payment-platform
```

2. 启动基础设施（PostgreSQL、Redis、Kafka）
```bash
docker-compose up -d postgres redis kafka
```

3. 初始化数据库
```bash
./scripts/init-db.sh
```

4. 启动所有微服务
```bash
./scripts/start-services.sh
```

5. 启动前端应用
```bash
cd web/admin-portal && npm install && npm start
cd web/merchant-portal && npm install && npm start
```

### 使用 Docker Compose 一键启动
```bash
docker-compose up -d
```

访问：
- Admin Portal: http://localhost:3000
- Merchant Portal: http://localhost:3001
- API Gateway: http://localhost:8080

## 📚 技术栈

### 后端
- **语言**: Go 1.21
- **框架**: gRPC, gRPC-Gateway, Gin
- **ORM**: GORM
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **消息队列**: Kafka 3.5
- **服务发现**: Consul
- **监控**: Prometheus + Grafana
- **追踪**: Jaeger
- **日志**: Zap

### 前端
- **框架**: React 18 + TypeScript
- **UI库**: Ant Design / Ant Design Pro
- **状态管理**: Zustand / Redux Toolkit
- **请求**: Axios
- **路由**: React Router v6
- **图表**: ECharts / Recharts

### 支付渠道SDK
- stripe-go v76 (已集成)
- paypal-go-sdk (待开发)
- go-ethereum (加密货币，待开发)

## 🗂️ 项目结构

```
payment-platform/
├── backend/                         # 后端服务
│   ├── services/                    # 微服务
│   │   ├── merchant-service/        # 商户管理服务
│   │   ├── admin-service/           # 运营管理服务（管理员、审核、配置）
│   │   ├── payment-gateway/         # 支付网关服务
│   │   ├── channel-adapter/         # 渠道适配服务
│   │   ├── order-service/           # 订单服务
│   │   ├── accounting-service/      # 账务服务
│   │   ├── risk-service/            # 风控服务
│   │   ├── notification-service/    # 通知服务
│   │   ├── analytics-service/       # 分析服务
│   │   └── config-service/          # 配置中心
│   ├── pkg/                         # 共享库
│   │   ├── auth/                    # JWT认证、密码加密
│   │   ├── crypto/                  # 加密解密工具
│   │   ├── db/                      # 数据库连接、多租户
│   │   ├── kafka/                   # Kafka客户端
│   │   ├── grpc/                    # gRPC工具
│   │   ├── config/                  # 配置加载
│   │   ├── logger/                  # 日志
│   │   └── middleware/              # 中间件
│   ├── proto/                       # gRPC Proto定义
│   └── deployments/                 # 部署配置
│       ├── kubernetes/              # K8s YAML
│       └── docker-compose/          # Docker Compose
│
├── frontend/                        # 前端应用
│   ├── admin-portal/                # 运营管理后台（React + Ant Design Pro）
│   └── merchant-portal/             # 商户自助后台（React + Ant Design）
│
├── scripts/                         # 脚本工具
├── docs/                            # 文档
└── go.work                          # Go Workspace配置
```

## 🔑 核心功能

### 运营管理后台 (Admin Portal)
- ✅ 商户管理（注册审核、KYC、状态控制）
- ✅ 交易监控（实时交易流、异常告警）
- ✅ 风控管理（规则配置、黑名单、风险评分）
- ✅ 财务管理（清结算、对账、提现审核）
- ✅ 渠道配置（费率、路由规则）
- ✅ 数据看板（GMV、成功率、渠道分布）
- ✅ 系统设置（角色权限、操作日志）

### 商户自助后台 (Merchant Portal)
- ✅ 商户入驻（自助注册、KYC提交）
- ✅ 渠道接入（Stripe/PayPal配置）
- ✅ API管理（密钥生成、Webhook配置）
- ✅ 交易查询（订单列表、详情、导出）
- ✅ 对账中心（日对账、月对账、差异处理）
- ✅ 财务报表（收入统计、手续费、提现）
- ✅ 数据分析（交易趋势、成功率、用户画像）
- ✅ 开发者工具（API文档、SDK下载、测试环境）

### 支付能力
- ✅ 支付方式：信用卡、借记卡、数字钱包、加密货币
- ✅ 支付场景：一次性支付、订阅支付、分期支付
- ✅ 币种支持：140+ 币种
- ✅ 退款：全额退款、部分退款
- ✅ 对账：T+1自动对账
- ✅ 清结算：自动/手动清结算

## 🔒 安全与合规

- **PCI DSS Level 1**: 不存储敏感卡信息，使用Token化
- **数据加密**: TLS 1.3传输加密 + AES-256数据加密
- **认证授权**: JWT + OAuth2/OIDC
- **密钥管理**: HashiCorp Vault
- **审计日志**: 所有操作可追溯
- **GDPR合规**: 数据导出、删除、匿名化

## 📊 监控与运维

- **健康检查**: `/health` 端点
- **指标监控**: Prometheus + Grafana
- **分布式追踪**: Jaeger
- **日志聚合**: ELK / Loki
- **告警**: AlertManager + 钉钉/Slack

## 🧪 测试

```bash
# 单元测试
make test

# 集成测试
make test-integration

# 压力测试
make load-test
```

## 📖 文档

### API 文档
- Swagger UI: http://localhost:8080/swagger
- gRPC文档: http://localhost:8080/grpc-docs

### 技术文档
- [安全功能文档](docs/SECURITY_FEATURES.md) - 2FA、登录追踪、会话管理、RBAC
- [支付网关文档](docs/PAYMENT_GATEWAY.md) - 支付网关设计、多货币支持、智能路由
- [订单服务文档](docs/ORDER_SERVICE.md) - 订单管理、状态流转、统计分析
- [渠道适配文档](docs/CHANNEL_ADAPTER.md) - Stripe 集成、Webhook 处理、适配器模式
- [项目进度文档](docs/PROJECT_PROGRESS.md) - 开发进度、代码统计、路线图

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 📞 联系方式

- Email: support@payment-platform.com
- Documentation: https://docs.payment-platform.com

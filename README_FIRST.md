# 🚀 Global Payment Platform - 从这里开始

## 👋 欢迎

欢迎使用**全球支付平台** (Global Payment Platform) - 一个企业级、多租户的支付网关系统。

**如果这是你第一次接触本项目,请从这里开始!**

---

## ⚡ 5分钟快速启动

### 前提条件

确保已安装以下工具:
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- Git

### 一键部署

```bash
# 克隆项目 (如果还没有)
git clone <repository-url>
cd payment

# 一键部署整个系统
chmod +x deploy.sh
./deploy.sh
```

**预计耗时**: 5-10分钟

部署完成后,你将看到所有访问链接,包括:
- Admin Portal: http://localhost:5173
- Merchant Portal: http://localhost:5174
- Grafana监控: http://localhost:40300

---

## 📊 系统概览

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      前端应用层                              │
│  Admin Portal  │  Merchant Portal  │  Website  │  Cashier  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    API网关 & 核心服务                        │
│     Payment Gateway  →  Order  →  Channel  →  External      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      业务服务层                              │
│  Risk │ Accounting │ Merchant │ Settlement │ KYC │ Dispute  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      支撑服务层                              │
│  Notification │ Analytics │ Config │ Auth │ Reconciliation  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      基础设施层                              │
│  PostgreSQL │ Redis │ Kafka │ Prometheus │ Grafana │ Jaeger │
└─────────────────────────────────────────────────────────────┘
```

### 核心数据

| 维度 | 数量 |
|------|------|
| 后端微服务 | 19 |
| 前端应用 | 3 |
| 数据库 | 19 |
| 代码行数 | ~117,500 |
| 技术文档 | 100+ |

---

## 🎯 核心功能

### 支付处理
- ✅ 多渠道支付 (Stripe集成,PayPal/加密货币规划中)
- ✅ 实时支付处理
- ✅ 支付查询与退款
- ✅ Webhook回调处理
- ✅ 幂等性保证

### 风控管理
- ✅ 实时风控评估
- ✅ 规则引擎
- ✅ GeoIP检测
- ✅ 黑名单管理
- ✅ 交易限额控制

### 财务管理
- ✅ 双账本记账
- ✅ 自动结算
- ✅ 提现管理
- ✅ 对账管理

### 合规管理
- ✅ KYC验证
- ✅ 争议处理
- ✅ 审计日志

### 可观测性
- ✅ Prometheus指标收集
- ✅ Grafana监控仪表板
- ✅ Jaeger分布式追踪
- ✅ 结构化日志

---

## 📚 必读文档

### 新手入门 (按顺序阅读)

1. **[QUICK_START.md](QUICK_START.md)** ⭐⭐⭐
   - 5分钟快速启动指南
   - 最小化命令集

2. **[deploy.sh](deploy.sh)** ⭐⭐⭐
   - 一键部署脚本
   - 自动化部署所有组件

3. **[PROJECT_STATUS_REPORT.md](PROJECT_STATUS_REPORT.md)** ⭐⭐⭐
   - 完整项目状态报告
   - 26KB详细文档
   - 所有服务清单

4. **[SYSTEM_READY_FOR_PRODUCTION.md](SYSTEM_READY_FOR_PRODUCTION.md)** ⭐⭐
   - 生产就绪检查清单
   - 部署建议
   - 已知限制

### 开发者文档

5. **[CLAUDE.md](CLAUDE.md)** ⭐⭐⭐
   - AI开发助手指南
   - 完整技术栈说明
   - 架构模式详解

6. **[backend/MICROSERVICE_UNIFIED_PATTERNS.md](backend/MICROSERVICE_UNIFIED_PATTERNS.md)** ⭐⭐
   - 微服务统一模式
   - Bootstrap框架使用
   - 代码示例

7. **[backend/API_DOCUMENTATION_GUIDE.md](backend/API_DOCUMENTATION_GUIDE.md)** ⭐⭐
   - Swagger/OpenAPI文档
   - API使用示例

### 运维文档 (NEW)

8. **[OPERATIONS_GUIDE.md](OPERATIONS_GUIDE.md)** ⭐⭐⭐
   - 完整运维手册
   - 故障排查指南
   - 性能优化建议
   - 备份恢复流程

9. **[backend/scripts/system-status-dashboard.sh](backend/scripts/system-status-dashboard.sh)** ⭐⭐⭐
   - 系统状态可视化
   - 一键查看所有服务健康状态

10. **[backend/scripts/service-dependency-map.sh](backend/scripts/service-dependency-map.sh)** ⭐⭐
    - 服务依赖关系图
    - 可视化服务架构

### 前端文档

11. **[FRONTEND_COMPLETE_SUMMARY.md](FRONTEND_COMPLETE_SUMMARY.md)** ⭐⭐
    - 前端完成总结
    - 3个应用, 46个页面
    - 性能优化报告

### 完整文档索引

12. **[DOCUMENTATION_INDEX.md](DOCUMENTATION_INDEX.md)** ⭐⭐
    - 100+ 文档索引
    - 分类导航
    - 推荐阅读顺序

---

## 🔧 常用命令

### 部署与启动

```bash
# 一键部署
./deploy.sh

# 启动基础设施
docker-compose up -d

# 启动后端服务
cd backend && ./scripts/start-all-services.sh

# 启动前端 (开发模式)
cd frontend/admin-portal && npm run dev
```

### 监控与诊断

```bash
# 查看系统状态 (推荐!)
cd backend && ./scripts/system-status-dashboard.sh

# 查看服务依赖关系
cd backend && ./scripts/service-dependency-map.sh

# 查看服务日志
tail -f backend/logs/payment-gateway.log

# 查看所有服务日志
tail -f backend/logs/*.log
```

### 停止服务

```bash
# 停止后端服务
cd backend && ./scripts/stop-all-services.sh

# 停止基础设施
docker-compose down
```

---

## 🌐 访问地址

### 前端应用

| 应用 | 地址 | 说明 |
|------|------|------|
| Admin Portal | http://localhost:5173 | 平台管理后台 |
| Merchant Portal | http://localhost:5174 | 商户自助平台 |
| Website | http://localhost:5175 | 官网 (可选) |

### API文档

| 服务 | Swagger地址 |
|------|-------------|
| Admin Service | http://localhost:40001/swagger/index.html |
| Merchant Service | http://localhost:40002/swagger/index.html |
| Payment Gateway | http://localhost:40003/swagger/index.html |

### 监控工具

| 工具 | 地址 | 凭据 |
|------|------|------|
| Grafana | http://localhost:40300 | admin/admin |
| Prometheus | http://localhost:40090 | - |
| Jaeger | http://localhost:40686 | - |

---

## 🏗️ 项目结构

```
payment/
├── deploy.sh                    # 一键部署脚本 ⭐
├── docker-compose.yml           # 基础设施编排
│
├── backend/                     # 后端服务
│   ├── services/               # 19个微服务
│   │   ├── payment-gateway/   # 支付网关 (核心)
│   │   ├── order-service/     # 订单服务
│   │   ├── channel-adapter/   # 渠道适配器
│   │   ├── risk-service/      # 风控服务
│   │   ├── accounting-service/ # 记账服务
│   │   └── ...                # 其他14个服务
│   │
│   ├── pkg/                    # 共享库 (20个包)
│   │   ├── app/               # Bootstrap框架
│   │   ├── auth/              # JWT认证
│   │   ├── db/                # 数据库连接
│   │   ├── metrics/           # Prometheus指标
│   │   ├── tracing/           # Jaeger追踪
│   │   └── ...
│   │
│   └── scripts/                # 运维脚本
│       ├── system-status-dashboard.sh  ⭐
│       ├── service-dependency-map.sh   ⭐
│       ├── start-all-services.sh
│       └── init-db.sh
│
├── frontend/                   # 前端应用
│   ├── admin-portal/          # 管理后台 (22页面)
│   ├── merchant-portal/       # 商户平台 (20页面)
│   └── website/               # 官网 (4页面)
│
└── docs/                       # 文档目录
    ├── QUICK_START.md         ⭐
    ├── OPERATIONS_GUIDE.md    ⭐
    ├── PROJECT_STATUS_REPORT.md ⭐
    └── ... (100+ 文档)
```

---

## 🎓 学习路径

### Level 1: 入门 (1-2天)

1. 运行 `./deploy.sh` 完成部署
2. 访问 Admin Portal 和 Merchant Portal
3. 查看 Grafana 监控仪表板
4. 阅读 [QUICK_START.md](QUICK_START.md)
5. 浏览 [PROJECT_STATUS_REPORT.md](PROJECT_STATUS_REPORT.md)

### Level 2: 理解架构 (3-5天)

1. 运行 `service-dependency-map.sh` 查看依赖关系
2. 阅读 [CLAUDE.md](CLAUDE.md) 了解技术栈
3. 查看核心服务代码:
   - `payment-gateway/cmd/main.go`
   - `order-service/internal/service/`
4. 理解支付流程
5. 查看 Swagger API 文档

### Level 3: 深入开发 (1-2周)

1. 阅读 [MICROSERVICE_UNIFIED_PATTERNS.md](backend/MICROSERVICE_UNIFIED_PATTERNS.md)
2. 学习 Bootstrap 框架使用
3. 尝试修改现有服务
4. 添加新的 API 端点
5. 编写单元测试

### Level 4: 运维掌握 (1周)

1. 阅读 [OPERATIONS_GUIDE.md](OPERATIONS_GUIDE.md)
2. 实践故障排查流程
3. 进行性能优化
4. 配置监控告警
5. 制定备份策略

---

## 💡 快速问答

### Q: 我应该从哪个文档开始?
**A**: 从 [QUICK_START.md](QUICK_START.md) 开始,然后运行 `./deploy.sh`

### Q: 如何查看系统是否正常运行?
**A**: 运行 `cd backend && ./scripts/system-status-dashboard.sh`

### Q: 服务启动失败怎么办?
**A**: 查看 [OPERATIONS_GUIDE.md](OPERATIONS_GUIDE.md) 的故障排查章节

### Q: 如何添加新的支付渠道?
**A**: 查看 [CLAUDE.md](CLAUDE.md) 的 "Adding a New Payment Channel" 章节

### Q: 前端如何连接后端?
**A**: 查看 `frontend/admin-portal/src/services/api.ts` 的 baseURL 配置

### Q: 如何查看服务日志?
**A**: 日志在 `backend/logs/` 目录,使用 `tail -f backend/logs/*.log`

### Q: 数据库在哪里?
**A**: PostgreSQL在Docker容器中,端口40432,有19个独立数据库

### Q: 如何停止所有服务?
**A**: 运行 `cd backend && ./scripts/stop-all-services.sh` 和 `docker-compose down`

---

## 🚀 生产部署建议

### 部署前检查清单

- [ ] 完成测试环境验证
- [ ] 进行性能测试 (目标: 10,000 req/s)
- [ ] 完成安全审计
- [ ] 配置监控告警
- [ ] 设置数据库备份
- [ ] 准备应急预案
- [ ] 团队培训完成

### 推荐部署流程

1. **测试环境** (1周) - 功能验证
2. **预生产环境** (2周) - 压力测试
3. **生产环境灰度** (1周) - 逐步放量

详见: [SYSTEM_READY_FOR_PRODUCTION.md](SYSTEM_READY_FOR_PRODUCTION.md)

---

## 📈 项目状态

| 维度 | 完成度 | 说明 |
|------|--------|------|
| **后端服务** | 100% | 19个服务全部实现 |
| **前端应用** | 100% | 3个应用, 46个页面 |
| **基础设施** | 100% | Docker Compose编排 |
| **可观测性** | 95% | Metrics/Tracing/Logging |
| **API文档** | 90% | Swagger覆盖主要服务 |
| **运维工具** | 100% | 一键部署 + 监控脚本 |
| **项目文档** | 95% | 100+ 文档 |

**总体完成度**: **95% - 生产就绪** ✅

---

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

---

## 📄 许可证

[待添加许可证信息]

---

## 📞 联系方式

- **技术支持**: support@payment-platform.com
- **紧急事件**: oncall@payment-platform.com
- **文档**: https://docs.payment-platform.com

---

## 🎉 开始你的支付平台之旅!

```bash
# 现在就开始!
./deploy.sh
```

部署完成后,访问 http://localhost:5173 查看 Admin Portal 🚀

---

**最后更新**: 2025-10-25
**项目版本**: 1.0.0
**文档版本**: 1.0.0

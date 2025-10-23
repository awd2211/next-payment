# Payment Platform - Frontend

全球支付平台前端项目，包含管理后台、商户门户和官网三个应用。

## 📁 项目结构

```
frontend/
├── admin-portal/          # 管理后台 (端口: 5173)
├── merchant-portal/       # 商户门户 (端口: 5174)
├── website/               # 官方网站 (端口: 5175)
├── shared/                # 共享代码 (计划中)
├── ecosystem.config.js    # PM2配置文件
├── pnpm-workspace.yaml    # pnpm工作区配置
└── OPTIMIZATION_SUMMARY.md # 优化总结文档
```

## 🚀 快速开始

### 前置要求

- Node.js >= 18.0
- pnpm >= 8.0
- PM2 (可选，用于进程管理)

### 安装依赖

```bash
# 安装pnpm (如果还没安装)
npm install -g pnpm

# 安装所有项目依赖 (pnpm workspace会自动处理)
pnpm install
```

### 开发模式

#### 方式1: 使用pnpm (单个项目)

```bash
# 启动管理后台
cd admin-portal
pnpm dev

# 启动商户门户
cd merchant-portal
pnpm dev

# 启动官网
cd website
pnpm dev
```

#### 方式2: 使用PM2 (所有项目)

```bash
# 启动所有前端项目
pm2 start ecosystem.config.js

# 查看状态
pm2 status

# 查看日志
pm2 logs

# 停止所有
pm2 stop all

# 重启所有
pm2 restart all

# 删除所有进程
pm2 delete all
```

### 生产构建

```bash
# 构建单个项目
cd admin-portal
pnpm build

# 或使用pnpm workspace命令构建所有项目
pnpm -r build
```

### 预览生产构建

```bash
cd admin-portal
pnpm preview
```

## 🛠️ 技术栈

### 核心框架
- **React 18** - UI框架
- **TypeScript** - 类型安全
- **Vite 5** - 构建工具

### UI组件库
- **Ant Design 5.15** - 企业级UI组件
- **@ant-design/icons** - 图标库
- **@ant-design/charts** - 图表库

### 状态管理
- **Zustand 4.5** - 轻量级状态管理

### 路由
- **React Router v6** - 客户端路由

### HTTP客户端
- **Axios** - HTTP请求

### 国际化
- **react-i18next** - 国际化支持 (12种语言)

### 日期处理
- **dayjs** - 轻量级日期库

### PWA
- **vite-plugin-pwa** - PWA支持 (仅admin-portal和merchant-portal)

## 📝 开发规范

### 代码格式化

```bash
# 使用ESLint检查
pnpm lint

# 使用ESLint自动修复
pnpm lint:fix

# 使用Prettier格式化
pnpm format
```

### Git提交规范

建议使用Conventional Commits规范：

```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
test: 测试相关
chore: 构建/工具相关
```

示例：
```bash
git commit -m "feat: 添加支付统计图表"
git commit -m "fix: 修复登录token刷新问题"
```

## 📚 项目特性

### Admin Portal (管理后台)
- ✅ 完整的RBAC权限管理
- ✅ 商户管理和审核
- ✅ 支付和订单监控
- ✅ 风控规则配置
- ✅ 数据分析仪表盘
- ✅ 系统配置管理
- ✅ 审计日志查询
- ✅ 12种语言支持
- ✅ 深色/浅色主题切换
- ✅ PWA离线支持

### Merchant Portal (商户门户)
- ✅ 商户注册和KYC
- ✅ API密钥管理
- ✅ Webhook配置
- ✅ 支付订单查询
- ✅ 交易统计分析
- ✅ 结算报表
- ✅ 多语言支持
- ✅ PWA离线支持

### Website (官网)
- ✅ 产品介绍
- ✅ API文档中心
- ✅ 定价方案
- ✅ 双语支持 (中英文)
- ✅ 响应式设计

## 🔧 配置说明

### 环境变量

每个项目需要创建 `.env.development` 和 `.env.production` 文件：

```env
# .env.development
VITE_APP_TITLE=支付平台管理后台
VITE_PORT=5173
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=10000
```

### 代理配置

开发环境使用Vite代理转发API请求到后端服务 (端口40001-40010)。

配置位置：`vite.config.ts`

```typescript
server: {
  port: 5173,
  proxy: {
    '/api/v1/admins': {
      target: 'http://localhost:40001',
      changeOrigin: true,
    },
    // ...
  }
}
```

## 📦 依赖管理

### 添加依赖

```bash
# 为特定项目添加依赖
cd admin-portal
pnpm add <package>

# 为所有项目添加依赖 (在根目录)
pnpm add <package> -w

# 添加开发依赖
pnpm add -D <package>
```

### 更新依赖

```bash
# 查看可更新的依赖
pnpm outdated

# 更新所有依赖到最新版本
pnpm update

# 更新特定依赖
pnpm update <package>
```

## 🐛 常见问题

### 1. 端口被占用

```bash
# 杀死占用端口的进程
lsof -ti:5173 | xargs kill -9
```

### 2. 依赖安装失败

```bash
# 清理缓存重新安装
pnpm store prune
rm -rf node_modules
pnpm install
```

### 3. TypeScript类型错误

```bash
# 类型检查
pnpm type-check

# 重启TS服务器 (VSCode)
Cmd+Shift+P -> TypeScript: Restart TS Server
```

### 4. 构建失败

```bash
# 清理构建缓存
rm -rf dist
rm -rf .vite

# 重新构建
pnpm build
```

## 📊 性能优化

### 已实施的优化

1. ✅ **代码分割** - React、Ant Design、图表库、工具库分离
2. ✅ **懒加载** - 路由组件按需加载
3. ✅ **PWA缓存** - 静态资源和API响应缓存
4. ✅ **Tree Shaking** - 移除未使用代码
5. ✅ **类型安全** - 完整的TypeScript类型定义

### 性能目标

- ⏱️ 首次加载时间 < 3秒
- 📈 Lighthouse性能评分 > 90
- 📦 单个chunk大小 < 500KB

## 🔐 安全特性

- ✅ JWT Token自动刷新
- ✅ RBAC权限控制
- ✅ XSS防护 (React默认)
- ✅ CSRF防护 (Token验证)
- ✅ 请求签名验证
- ✅ 敏感数据脱敏

## 📖 相关文档

- [优化总结](./OPTIMIZATION_SUMMARY.md) - 详细的优化记录和待办事项
- [Admin Portal文档](./admin-portal/README.md)
- [Merchant Portal文档](./merchant-portal/README.md)
- [Website文档](./website/README.md)

## 🤝 贡献指南

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 📄 许可证

MIT License

## 📞 联系方式

如有问题，请提交Issue或联系开发团队。

---

**最后更新**: 2025-10-23

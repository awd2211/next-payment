# 支付平台前端项目

本目录包含支付平台的两个前端应用：

## 1. Admin Dashboard (管理员后台)

### 功能特性
- 🔐 管理员登录认证
- 📊 系统概览仪表板
- ⚙️ 系统配置管理（支持CRUD操作、分类查看）
- 👥 管理员管理
- 🔑 角色权限管理
- 📝 审计日志查询

### 技术栈
- React 18
- TypeScript
- Ant Design 5
- Vite
- React Router v6
- Zustand (状态管理)
- Axios

### 目录结构
```
admin-portal/
├── src/
│   ├── components/     # 通用组件
│   │   └── Layout.tsx  # 主布局组件
│   ├── pages/          # 页面组件
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── SystemConfigs.tsx
│   │   ├── Admins.tsx
│   │   ├── Roles.tsx
│   │   └── AuditLogs.tsx
│   ├── services/       # API服务
│   │   ├── api.ts
│   │   ├── authService.ts
│   │   └── systemConfigService.ts
│   ├── stores/         # 状态管理
│   │   └── authStore.ts
│   ├── types/          # TypeScript类型定义
│   ├── utils/          # 工具函数
│   ├── App.tsx         # 应用根组件
│   └── main.tsx        # 入口文件
├── package.json
├── tsconfig.json
├── vite.config.ts
└── index.html
```

### 本地开发

```bash
# 进入项目目录
cd admin-portal

# 安装依赖
npm install

# 启动开发服务器（默认端口: 40101）
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview
```

### 默认账号
- 用户名：`admin`
- 密码：`admin123456`

### API代理配置
开发环境下，所有 `/api` 请求将被代理到 `http://localhost:40001`（Admin Service）

---

## 2. Merchant Dashboard (商户中心)

### 功能特性
- 🔐 商户登录认证
- 📈 交易数据概览
- 💰 账户余额查询
- 📋 交易记录查询
- 🛒 订单管理
- 👤 商户信息管理

### 技术栈
- React 18
- TypeScript
- Ant Design 5
- Ant Design Charts (数据可视化)
- Vite
- React Router v6
- Zustand (状态管理)
- Axios

### 目录结构
```
merchant-portal/
├── src/
│   ├── components/     # 通用组件
│   │   └── Layout.tsx  # 主布局组件
│   ├── pages/          # 页面组件
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Transactions.tsx
│   │   ├── Orders.tsx
│   │   └── Account.tsx
│   ├── services/       # API服务
│   │   └── api.ts
│   ├── stores/         # 状态管理
│   │   └── authStore.ts
│   ├── types/          # TypeScript类型定义
│   ├── utils/          # 工具函数
│   ├── App.tsx         # 应用根组件
│   └── main.tsx        # 入口文件
├── package.json
├── tsconfig.json
├── vite.config.ts
└── index.html
```

### 本地开发

```bash
# 进入项目目录
cd merchant-portal

# 安装依赖
npm install

# 启动开发服务器（默认端口: 40200）
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview
```

### API代理配置
开发环境下，所有 `/api` 请求将被代理到 `http://localhost:40002`（Merchant Service）

---

## 开发指南

### 环境要求
- Node.js >= 18.0.0
- npm >= 9.0.0

### 同时运行两个前端

可以使用两个终端分别启动：

```bash
# 终端1: Admin Dashboard
cd admin-portal && npm run dev

# 终端2: Merchant Dashboard
cd merchant-portal && npm run dev
```

访问地址：
- Admin Dashboard: http://localhost:40101
- Merchant Dashboard: http://localhost:40200

### 代码规范
- 使用 ESLint 进行代码检查
- 使用 TypeScript 确保类型安全
- 遵循 React Hooks 最佳实践
- 组件采用函数式编程

### 状态管理
两个应用都使用 Zustand 进行状态管理，主要管理：
- 用户认证状态（token, user info）
- 刷新token
- 权限验证

### API请求
- 所有API请求统一通过 `services/api.ts` 进行
- 自动添加 Authorization header
- 统一的错误处理和提示
- 401 自动跳转登录
- 支持请求/响应拦截器

### 主题定制
可以在 `main.tsx` 中通过 ConfigProvider 定制 Ant Design 主题：

```typescript
<ConfigProvider
  locale={zhCN}
  theme={{
    token: {
      colorPrimary: '#1890ff',
      borderRadius: 4,
    },
  }}
>
  <App />
</ConfigProvider>
```

---

## 生产部署

### 构建

```bash
# Admin Dashboard
cd admin-portal && npm run build

# Merchant Dashboard
cd merchant-portal && npm run build
```

构建产物将生成在各自的 `dist` 目录下。

### Nginx 配置示例

```nginx
# Admin Dashboard
server {
    listen 80;
    server_name admin.example.com;

    root /path/to/admin-portal/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:40001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Merchant Dashboard
server {
    listen 80;
    server_name merchant.example.com;

    root /path/to/merchant-portal/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:40002;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 后续开发计划

### Admin Dashboard
- [ ] 完善管理员管理页面
- [ ] 完善角色权限管理页面
- [ ] 完善审计日志查询页面
- [ ] 添加商户管理功能
- [ ] 添加数据统计和报表
- [ ] 添加系统监控功能

### Merchant Dashboard
- [ ] 实现交易记录查询功能
- [ ] 实现订单管理功能
- [ ] 完善账户信息页面
- [ ] 添加数据可视化图表
- [ ] 添加账单和结算功能
- [ ] 添加API密钥管理

---

## 常见问题

### 1. 端口被占用
修改 `vite.config.ts` 中的 `server.port` 配置。

### 2. API请求失败
确保后端服务已启动：
- Admin Service: http://localhost:40001
- Merchant Service: http://localhost:40002

### 3. 构建错误
清除缓存并重新安装依赖：
```bash
rm -rf node_modules package-lock.json
npm install
```

---

## 技术支持

如有问题，请联系技术支持团队。

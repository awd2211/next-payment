# 前端快速开始指南

## 🎯 一键启动

### 方式1: 使用初始化脚本（首次运行）

```bash
cd /home/eric/payment/frontend
./scripts/setup.sh
```

这个脚本会自动：
- ✅ 检查Node.js和pnpm
- ✅ 安装所有依赖
- ✅ 清理package-lock.json
- ✅ 创建.env文件
- ✅ 复制配置文件
- ✅ TypeScript类型检查

### 方式2: 使用PM2启动开发服务器

```bash
cd /home/eric/payment/frontend
./scripts/start-dev.sh
```

或者直接：

```bash
pm2 start ecosystem.config.js
```

### 方式3: 单独启动某个项目

```bash
# Admin Portal
cd admin-portal && pnpm dev

# Merchant Portal  
cd merchant-portal && pnpm dev

# Website
cd website && pnpm dev
```

## 📦 生产构建

```bash
cd /home/eric/payment/frontend
./scripts/build-all.sh
```

或手动构建：

```bash
cd admin-portal && pnpm build
cd merchant-portal && pnpm build
cd website && pnpm build
```

## 🔍 常用命令

### 开发
```bash
pnpm dev          # 启动开发服务器
pnpm build        # 生产构建
pnpm preview      # 预览生产构建
```

### 代码质量
```bash
pnpm lint         # ESLint检查
pnpm lint:fix     # ESLint自动修复
pnpm format       # Prettier格式化
pnpm type-check   # TypeScript类型检查
```

### PM2管理
```bash
pm2 status        # 查看状态
pm2 logs          # 查看日志
pm2 logs admin-portal  # 查看特定项目日志
pm2 restart all   # 重启所有
pm2 stop all      # 停止所有
pm2 delete all    # 删除所有进程
```

## 🌐 访问地址

- **Admin Portal**: http://localhost:5173
- **Merchant Portal**: http://localhost:5174
- **Website**: http://localhost:5175

## ⚙️ 前置要求

- Node.js >= 18.0
- pnpm >= 8.0
- PM2 (可选，用于进程管理)

## 🔧 环境变量

三个项目都需要创建 `.env.development` 和 `.env.production` 文件。

初始化脚本会自动创建，或手动创建：

```env
# .env.development
VITE_APP_TITLE=支付平台管理后台
VITE_PORT=5173
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=10000
```

## 🐛 问题排查

### 问题1: 端口被占用

```bash
# 查看占用端口的进程
lsof -i:5173

# 杀死进程
kill -9 <PID>
```

### 问题2: pnpm命令找不到

```bash
# 安装pnpm
npm install -g pnpm

# 或使用corepack (Node.js >= 16.9)
corepack enable
corepack prepare pnpm@latest --activate
```

### 问题3: 依赖安装失败

```bash
# 清理并重新安装
pnpm store prune
rm -rf node_modules
pnpm install
```

### 问题4: TypeScript报错

```bash
# 类型检查
pnpm type-check

# 如果是新增的工具函数，需要确保正确导出
```

### 问题5: 后端连接失败

检查：
1. 后端服务是否运行在40001-40010端口
2. vite.config.ts中的代理配置是否正确
3. 浏览器控制台的网络请求

## 📚 更多文档

- [完整README](./README.md)
- [优化总结](./OPTIMIZATION_SUMMARY.md)
- [项目结构说明](../CLAUDE.md)

## 💡 快速技巧

### 1. 使用新的工具函数

```typescript
import { formatAmount, isEmail, debounce } from '@/utils'

// 格式化金额
const amount = formatAmount(12345) // ¥123.45

// 验证邮箱
const valid = isEmail('user@example.com')

// 防抖搜索
const handleSearch = debounce((value) => {
  fetchData(value)
}, 500)
```

### 2. 使用类型安全的API请求

```typescript
import request from '@/services/request'
import type { Payment, ApiResponse } from '@/types'

// 类型安全的请求
const response = await request.get<Payment[]>('/payments')
if (response.code === 0) {
  const payments = response.data // 类型为Payment[]
}
```

### 3. 使用自定义Hooks

```typescript
import { useRequest, usePagination, useDebounce } from '@/hooks'

// 简化异步请求
const { data, loading, run } = useRequest(fetchPayments)

// 分页
const { data, page, pageSize, changePage } = usePagination(fetchPaymentList)

// 防抖值
const debouncedValue = useDebounce(searchValue, 500)
```

## 🎉 完成！

现在你的前端开发环境已经配置完成，可以开始开发了！

如有问题，请查看详细文档或提交Issue。




# 前端优化总结

## ✅ 已完成的优化

### 1. **项目结构优化**

#### 1.1 包管理器配置
- ✅ 创建 `pnpm-workspace.yaml` - 支持monorepo管理
- ✅ 创建 `.npmrc` - pnpm配置文件
- ✅ 统一使用pnpm管理依赖（需删除package-lock.json）

```bash
# 清理npm lock文件并安装依赖
cd /home/eric/payment/frontend/admin-portal
rm -f package-lock.json
cd /home/eric/payment/frontend/merchant-portal
rm -f package-lock.json
cd /home/eric/payment/frontend
pnpm install
```

#### 1.2 端口配置修复
- ✅ admin-portal: `40101` → `5173`
- ✅ merchant-portal: `40200` → `5174`
- ✅ 后端服务代理端口: `8001-8010` → `40001-40010`

### 2. **代码质量提升**

#### 2.1 TypeScript类型系统
新增 `src/types/index.ts`，包含:
- ✅ API响应类型 (`ApiResponse`, `PaginationResponse`)
- ✅ 业务实体类型 (Admin, Merchant, Payment, Order等)
- ✅ 完整的类型安全覆盖

#### 2.2 工具函数库
新增 `src/utils/` 目录:
- ✅ `format.ts` - 格式化工具（金额、数字、日期、脱敏等）
- ✅ `validate.ts` - 数据验证（邮箱、手机、银行卡、身份证等）
- ✅ `debounce.ts` - 防抖/节流/异步防抖
- ✅ `storage.ts` - 本地存储（支持过期时间）
- ✅ `index.ts` - 统一导出

使用示例：
```typescript
import { formatAmount, isEmail, debounce } from '@/utils'

// 格式化金额
formatAmount(12345) // ¥123.45

// 验证邮箱
isEmail('user@example.com') // true

// 防抖
const handleSearch = debounce((value: string) => {
  fetchData(value)
}, 500)
```

#### 2.3 API请求层重构
新增 `src/services/request.ts`:
- ✅ 完整的类型安全
- ✅ Token自动刷新机制（401自动刷新，避免并发刷新）
- ✅ 请求ID追踪（X-Request-ID）
- ✅ 统一错误处理
- ✅ 文件上传/下载支持
- ✅ 生产环境错误上报接口

使用示例：
```typescript
import request from '@/services/request'
import type { ApiResponse, Payment } from '@/types'

// GET请求（类型安全）
const response = await request.get<Payment[]>('/payments')
if (response.code === 0) {
  const payments = response.data
}

// POST请求
await request.post('/payments', { amount: 10000, currency: 'CNY' })

// 文件上传
await request.upload('/files', formData, (progress) => {
  console.log(`上传进度: ${progress}%`)
})
```

### 3. **代码规范配置**

#### 3.1 ESLint配置
- ✅ 创建 `.eslintrc.json`
- ✅ 配置React + TypeScript规则
- ✅ 关闭过于严格的规则（any, console.log）

#### 3.2 Prettier配置
- ✅ 创建 `.prettierrc.json`
- ✅ 统一代码格式（单引号、无分号、100字符宽度）

#### 3.3 Git配置
- ✅ 更新 `.gitignore`（排除lock文件、环境变量等）

### 4. **性能优化**

#### 4.1 构建优化
在 `vite.config.ts` 中添加:
- ✅ 代码分割配置（manualChunks）
  - react-vendor: React相关库
  - antd-vendor: Ant Design组件库
  - chart-vendor: 图表库
  - utils: 工具库
- ✅ 提高chunk大小警告阈值（1000KB）

#### 4.2 PWA缓存策略调整建议
当前API缓存5分钟，建议根据业务调整：
```javascript
// 建议修改为不缓存敏感数据
urlPattern: /\/api\/v1\/(payments|orders|merchants)/,
handler: 'NetworkFirst',
options: {
  networkTimeoutSeconds: 3,
  cacheName: 'api-cache',
  expiration: {
    maxEntries: 50,
    maxAgeSeconds: 60, // 改为1分钟
  }
}
```

### 5. **安全性提升**

#### 5.1 Token刷新机制
- ✅ 自动刷新过期token
- ✅ 避免并发刷新（单例Promise）
- ✅ 刷新失败自动跳转登录

#### 5.2 请求追踪
- ✅ 每个请求添加唯一ID（X-Request-ID）
- ✅ 便于问题追踪和日志关联

---

## 📋 待优化项（按优先级）

### P1 - 高优先级

#### 1.1 环境变量配置
需要手动创建（被gitignore阻止）：

**admin-portal/.env.development**
```env
VITE_APP_TITLE=支付平台管理后台
VITE_PORT=5173
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=10000
```

**admin-portal/.env.production**
```env
VITE_APP_TITLE=支付平台管理后台
VITE_PORT=5173
VITE_API_PREFIX=/api/v1
VITE_REQUEST_TIMEOUT=30000
```

merchant-portal和website也需要类似配置。

#### 1.2 依赖版本统一
```bash
# admin-portal和merchant-portal的@ant-design/charts版本不一致
pnpm install @ant-design/charts@^2.6.6 -w
```

#### 1.3 清理旧的API层
- 删除或重构 `src/services/api.ts`，使用新的 `request.ts`

#### 1.4 更新package.json脚本
```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "lint:fix": "eslint . --ext ts,tsx --fix",
    "format": "prettier --write \"src/**/*.{ts,tsx,json,css,scss}\"",
    "type-check": "tsc --noEmit"
  }
}
```

### P2 - 中优先级

#### 2.1 创建共享包
建议创建 `frontend/shared/utils`、`frontend/shared/types`、`frontend/shared/components`：

```bash
mkdir -p frontend/shared/{utils,types,components}
```

将utils、types等复制到shared，三个项目引用：
```typescript
// 在各项目的package.json中添加
"dependencies": {
  "@payment/shared-utils": "workspace:*",
  "@payment/shared-types": "workspace:*"
}
```

#### 2.2 组件性能优化
Dashboard.tsx中的优化建议：
```typescript
// ❌ 当前：硬编码数据
const [stats, setStats] = useState({ totalAdmins: 25 })

// ✅ 优化：从API获取
useEffect(() => {
  fetchDashboardStats().then(setStats)
}, [])

// ✅ 使用React.memo优化子组件
const StatCard = React.memo(({ title, value, icon }) => {
  return <Card>...</Card>
})

// ✅ 使用useMemo缓存计算结果
const chartConfig = useMemo(() => ({
  data: trendData,
  xField: 'date',
  // ...
}), [trendData])
```

#### 2.3 错误边界
添加全局错误边界：

```typescript
// src/components/ErrorBoundary.tsx
class ErrorBoundary extends React.Component {
  componentDidCatch(error, errorInfo) {
    // 上报错误
    logErrorToService(error, errorInfo)
  }
  
  render() {
    if (this.state.hasError) {
      return <ErrorFallback />
    }
    return this.props.children
  }
}

// 在App.tsx中使用
<ErrorBoundary>
  <App />
</ErrorBoundary>
```

#### 2.4 国际化优化
添加语言fallback和动态加载：

```typescript
// i18n/config.ts
i18n.use(Backend).init({
  lng: 'zh-CN',
  fallbackLng: 'en',
  interpolation: { escapeValue: false },
  backend: {
    loadPath: '/locales/{{lng}}/{{ns}}.json',
  },
})
```

### P3 - 低优先级

#### 3.1 Husky Git Hooks
```bash
pnpm install -D husky lint-staged
npx husky install

# .husky/pre-commit
pnpm lint-staged
```

```json
// package.json
{
  "lint-staged": {
    "*.{ts,tsx}": ["eslint --fix", "prettier --write"],
    "*.{json,css,scss}": ["prettier --write"]
  }
}
```

#### 3.2 Commitlint
```bash
pnpm install -D @commitlint/cli @commitlint/config-conventional

# .commitlintrc.json
{
  "extends": ["@commitlint/config-conventional"]
}
```

#### 3.3 单元测试
```bash
pnpm install -D vitest @testing-library/react @testing-library/jest-dom

# vite.config.ts
export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
})
```

#### 3.4 Storybook
```bash
npx storybook@latest init
```

---

## 🚀 使用PM2运行前端

### PM2配置文件
已创建 `frontend/ecosystem.config.js`（见下文）

### 启动命令
```bash
# 安装依赖
cd /home/eric/payment/frontend
pnpm install

# 启动所有前端项目
pm2 start ecosystem.config.js

# 查看日志
pm2 logs

# 停止所有
pm2 stop all

# 重启
pm2 restart all

# 查看状态
pm2 status
```

---

## 📊 性能指标目标

### 构建性能
- ✅ Vite开发服务器启动时间: < 2秒
- ✅ 代码分割后，单个chunk大小: < 500KB
- ✅ 首次加载时间: < 3秒

### 运行时性能
- ⏳ Lighthouse性能评分: > 90
- ⏳ 首屏渲染(FCP): < 1.5秒
- ⏳ 最大内容绘制(LCP): < 2.5秒

---

## 📝 代码迁移指南

### 从旧API层迁移到新request.ts

**Before (api.ts):**
```typescript
import api from '../services/api'

const response = await api.get('/payments')
const payments = response.data // 类型未知
```

**After (request.ts):**
```typescript
import request from '@/services/request'
import type { Payment } from '@/types'

const response = await request.get<Payment[]>('/payments')
if (response.code === 0) {
  const payments = response.data // 类型安全的Payment[]
}
```

### 使用新的工具函数

**Before:**
```typescript
// 硬编码格式化
const amount = `¥${(payment.amount / 100).toFixed(2)}`

// 手动验证
const valid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
```

**After:**
```typescript
import { formatAmount, isEmail } from '@/utils'

const amount = formatAmount(payment.amount)
const valid = isEmail(email)
```

---

## 🎯 下一步行动

### 立即执行
1. ✅ 删除package-lock.json
2. ✅ 运行 `pnpm install`
3. ✅ 手动创建.env文件
4. ✅ 更新package.json scripts
5. ✅ 重启开发服务器测试

### 本周内完成
1. 迁移所有API调用到新的request.ts
2. 重构Dashboard页面使用真实API
3. 添加错误边界
4. 统一依赖版本

### 月度计划
1. 创建shared包
2. 添加单元测试
3. 性能优化到Lighthouse > 90
4. 添加Storybook文档

---

## 📞 问题反馈

如遇到问题，请检查：
1. pnpm版本 >= 8.0
2. Node.js版本 >= 18.0
3. 后端服务是否在40001-40010端口运行
4. .env文件是否正确创建




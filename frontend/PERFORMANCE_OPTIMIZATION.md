# 前端性能优化指南

## 📊 性能优化总览

本文档提供支付平台前端性能优化的最佳实践和工具。

## 🎯 性能目标

| 指标 | 目标值 | 当前值 |
|------|--------|--------|
| Lighthouse 性能评分 | > 90 | 待测试 |
| 首次内容绘制 (FCP) | < 1.5s | 待测试 |
| 最大内容绘制 (LCP) | < 2.5s | 待测试 |
| 首次输入延迟 (FID) | < 100ms | 待测试 |
| 累积布局偏移 (CLS) | < 0.1 | 待测试 |
| 总阻塞时间 (TBT) | < 300ms | 待测试 |

## 🚀 已实施的优化

### 1. 构建优化 ✅

**代码分割** (vite.config.ts):
```typescript
build: {
  rollupOptions: {
    output: {
      manualChunks: {
        'react-vendor': ['react', 'react-dom', 'react-router-dom'],
        'antd-vendor': ['antd', '@ant-design/icons'],
        'chart-vendor': ['@ant-design/charts'],
        'utils': ['axios', 'dayjs', 'zustand'],
      }
    }
  }
}
```

**优点**：
- ✅ 第三方库单独打包，充分利用浏览器缓存
- ✅ 首次加载只需要核心 chunks
- ✅ 减少主 bundle 大小

### 2. PWA 缓存策略 ✅

**字体缓存** - CacheFirst (365天):
```typescript
{
  urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
  handler: 'CacheFirst',
  options: {
    cacheName: 'google-fonts-cache',
    expiration: { maxAgeSeconds: 60 * 60 * 24 * 365 }
  }
}
```

**API 缓存** - NetworkFirst (5分钟):
```typescript
{
  urlPattern: /\/api\/v1\/.*/i,
  handler: 'NetworkFirst',
  options: {
    cacheName: 'api-cache',
    expiration: { maxAgeSeconds: 60 * 5 },
    networkTimeoutSeconds: 10
  }
}
```

### 3. 共享包架构 ✅

**@payment/shared** 避免代码重复：
- ✅ 三个项目共享 utils、types、hooks
- ✅ 减少重复代码，降低总包大小
- ✅ 统一维护，减少 bug

## 📝 React 性能优化最佳实践

### 1. 使用 React.memo 避免不必要的重渲染

**❌ 不好的做法**:
```typescript
const PaymentCard = ({ payment }) => {
  console.log('PaymentCard rendered')
  return <Card>{payment.amount}</Card>
}
```

**✅ 好的做法**:
```typescript
import { memo } from 'react'

const PaymentCard = memo(({ payment }) => {
  console.log('PaymentCard rendered')
  return <Card>{payment.amount}</Card>
})
```

**使用场景**：
- 大列表中的卡片组件
- 不频繁更新的组件
- 接收复杂 props 但很少变化的组件

### 2. 使用 useMemo 缓存计算结果

**❌ 不好的做法**:
```typescript
function PaymentList({ payments }) {
  // 每次渲染都会重新排序
  const sortedPayments = payments.sort((a, b) => b.created_at - a.created_at)
  return <>{sortedPayments.map(...)}</>
}
```

**✅ 好的做法**:
```typescript
import { useMemo } from 'react'

function PaymentList({ payments }) {
  const sortedPayments = useMemo(() => {
    return payments.sort((a, b) => b.created_at - a.created_at)
  }, [payments])

  return <>{sortedPayments.map(...)}</>
}
```

**使用场景**：
- 复杂的数据转换
- 过滤和排序操作
- 格式化大量数据

### 3. 使用 useCallback 缓存函数引用

**❌ 不好的做法**:
```typescript
function PaymentTable({ payments }) {
  // 每次渲染都创建新函数，导致子组件重渲染
  const handleDelete = (id) => {
    deletePayment(id)
  }

  return <>{payments.map(p => <Row onDelete={handleDelete} />)}</>
}
```

**✅ 好的做法**:
```typescript
import { useCallback } from 'react'

function PaymentTable({ payments }) {
  const handleDelete = useCallback((id) => {
    deletePayment(id)
  }, []) // 空依赖数组，函数引用永不变化

  return <>{payments.map(p => <Row onDelete={handleDelete} />)}</>
}
```

### 4. 列表虚拟化

对于超长列表（1000+ 项），使用虚拟滚动：

```bash
pnpm add react-window
```

```typescript
import { FixedSizeList } from 'react-window'

const PaymentList = ({ payments }) => {
  const Row = ({ index, style }) => (
    <div style={style}>
      <PaymentCard payment={payments[index]} />
    </div>
  )

  return (
    <FixedSizeList
      height={600}
      itemCount={payments.length}
      itemSize={120}
      width="100%"
    >
      {Row}
    </FixedSizeList>
  )
}
```

### 5. 懒加载路由

```typescript
import { lazy, Suspense } from 'react'

// ❌ 不好
import Dashboard from './pages/Dashboard'
import Payments from './pages/Payments'

// ✅ 好
const Dashboard = lazy(() => import('./pages/Dashboard'))
const Payments = lazy(() => import('./pages/Payments'))

function App() {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/payments" element={<Payments />} />
      </Routes>
    </Suspense>
  )
}
```

### 6. 图片优化

```typescript
// 使用 WebP 格式 + 懒加载
<img
  src="payment-icon.webp"
  alt="Payment"
  loading="lazy"
  width={100}
  height={100}
/>

// 响应式图片
<picture>
  <source srcset="payment-icon-large.webp" media="(min-width: 800px)" />
  <source srcset="payment-icon-medium.webp" media="(min-width: 400px)" />
  <img src="payment-icon-small.webp" alt="Payment" loading="lazy" />
</picture>
```

## 🔧 性能监控工具

### 1. React DevTools Profiler

```bash
# 开发模式下使用
# 1. 打开 React DevTools
# 2. 切换到 Profiler 标签
# 3. 点击录制，执行操作
# 4. 查看组件渲染时间和次数
```

### 2. Chrome DevTools Performance

```bash
# 1. 打开 Chrome DevTools
# 2. 切换到 Performance 标签
# 3. 点击录制，执行操作
# 4. 分析火焰图
```

### 3. Lighthouse

```bash
# 命令行运行
pnpm add -D lighthouse

# 或使用 Chrome DevTools Lighthouse 标签
```

## 📊 性能检查清单

### 构建时检查

- [x] 代码分割配置
- [x] Tree shaking 启用
- [x] 压缩 JS/CSS
- [ ] 使用 CDN（可选）
- [x] Source map 仅开发环境

### 运行时检查

- [x] 避免不必要的重渲染（React.memo）
- [x] 缓存昂贵的计算（useMemo）
- [x] 缓存事件处理函数（useCallback）
- [ ] 虚拟化长列表
- [ ] 懒加载路由和组件
- [ ] 图片懒加载

### 网络优化

- [x] PWA 缓存策略
- [x] HTTP/2 服务器推送（Vite 自动支持）
- [ ] 图片使用 WebP 格式
- [ ] 启用 Gzip/Brotli 压缩
- [x] 合理的缓存策略

## 🎯 优化优先级

### P0 - 立即执行
- [x] 代码分割
- [x] 移除未使用的依赖
- [x] 优化 Bundle 大小

### P1 - 本周完成
- [ ] 关键组件使用 React.memo
- [ ] Dashboard 数据使用 useMemo
- [ ] 事件处理器使用 useCallback

### P2 - 本月完成
- [ ] 长列表虚拟化
- [ ] 路由懒加载
- [ ] 图片优化（WebP + 懒加载）

### P3 - 长期优化
- [ ] 服务端渲染（SSR）
- [ ] 静态站点生成（SSG）
- [ ] 边缘计算（Edge Functions）

## 📈 性能测试

### 本地测试

```bash
# 1. 生产构建
pnpm build

# 2. 预览构建
pnpm preview

# 3. 运行 Lighthouse
lighthouse http://localhost:4173 --view

# 4. 查看 Bundle 分析
pnpm add -D rollup-plugin-visualizer
# 在 vite.config.ts 中添加 visualizer()
```

### CI/CD 集成

```yaml
# .github/workflows/performance.yml
name: Performance

on: [pull_request]

jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: pnpm install
      - run: pnpm build
      - uses: treosh/lighthouse-ci-action@v9
        with:
          urls: http://localhost:4173
          uploadArtifacts: true
```

## 🔍 性能问题排查

### 问题1: 首屏加载慢

**症状**: FCP > 3秒

**排查**:
1. 检查 Bundle 大小 (`pnpm build` 查看输出)
2. 检查是否有大的第三方库
3. 查看网络请求瀑布图

**解决**:
- 使用代码分割
- 懒加载非关键路由
- 使用 CDN 加载大库

### 问题2: 滚动卡顿

**症状**: 滚动长列表时掉帧

**排查**:
1. React DevTools Profiler 查看渲染次数
2. 检查是否每次滚动都重渲染

**解决**:
- 使用 React.memo
- 使用虚拟滚动（react-window）
- 避免在滚动时进行复杂计算

### 问题3: 点击响应慢

**症状**: FID > 300ms

**排查**:
1. Chrome Performance 查看主线程阻塞
2. 检查是否有同步的大量计算

**解决**:
- 将计算移到 Web Worker
- 使用 requestIdleCallback
- 分批处理数据

## 📚 参考资源

- [Web.dev Performance](https://web.dev/performance/)
- [React Performance Optimization](https://react.dev/learn/render-and-commit)
- [Vite Build Optimization](https://vitejs.dev/guide/build.html)
- [Lighthouse Scoring](https://web.dev/performance-scoring/)

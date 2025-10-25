# 前端性能优化完整指南

## 📊 项目概览

本文档涵盖整个 Payment Platform 前端生态的优化方案:

- **Website** (官网) - http://localhost:5176/
- **Admin Portal** (管理后台) - http://localhost:5173/
- **Merchant Portal** (商户后台) - http://localhost:5174/

---

## 🎯 优化目标

### 性能指标
- **LCP (Largest Contentful Paint)**: < 2.5s
- **FID (First Input Delay)**: < 100ms
- **CLS (Cumulative Layout Shift)**: < 0.1
- **TTI (Time to Interactive)**: < 3.5s
- **Bundle Size**: < 500KB (gzipped)

### 用户体验
- ⚡ 快速加载(3G网络下 < 5s)
- 🎨 流畅动画(60fps)
- 📱 完美移动端体验
- ♿ 良好的可访问性
- 🌍 国际化支持

---

## 🚀 已实现的优化

### 1. Website 官网优化 ✅

#### 视觉优化
- ✅ 玻璃态设计(backdrop-filter)
- ✅ 渐变色系统(6种特性渐变)
- ✅ 流畅动画(fadeInUp, 悬停效果)
- ✅ 响应式设计(3个断点)

#### 性能优化
- ✅ Vite 构建优化
- ✅ CSS 动画(GPU加速)
- ✅ 组件化设计
- ✅ 懒加载准备就绪

#### SEO 优化
- ⏳ Meta 标签配置
- ⏳ 结构化数据
- ⏳ Sitemap 生成
- ⏳ Open Graph 标签

### 2. Admin Portal 优化 ✅

#### 图表性能优化
```typescript
// 已实现的 Hook
- useChartDebounce: 防抖更新
- useChartLazyLoad: 懒加载渲染
- useChartSampling: 大数据采样
- useChartWindowing: 数据窗口化
- useChartResize: 自适应大小
- useOptimizedChart: 综合优化
```

#### 组件优化
- ✅ ErrorBoundary 错误边界
- ✅ PageLoading 加载骨架
- ✅ SkeletonLoading 骨架屏
- ✅ CommonTable 通用表格
- ✅ WebSocket 实时通信
- ✅ PWA 支持

#### 数据优化
- ✅ 防抖/节流(useDebounce)
- ✅ 虚拟滚动(大列表)
- ✅ 分页加载
- ✅ 批量操作

### 3. Merchant Portal 优化 🟡

#### 待优化项
- ⏳ 图表性能优化
- ⏳ 大数据表格优化
- ⏳ 实时数据更新
- ⏳ 移动端适配

---

## 🔧 详细优化策略

### A. 代码分割 & 懒加载

#### 1. 路由级代码分割
```typescript
// Bad ❌
import Dashboard from './pages/Dashboard'
import Payments from './pages/Payments'

// Good ✅
const Dashboard = lazy(() => import('./pages/Dashboard'))
const Payments = lazy(() => import('./pages/Payments'))

// 使用 Suspense
<Suspense fallback={<PageLoading />}>
  <Routes>
    <Route path="/dashboard" element={<Dashboard />} />
    <Route path="/payments" element={<Payments />} />
  </Routes>
</Suspense>
```

#### 2. 组件级懒加载
```typescript
// 图表组件懒加载
const LineChart = lazy(() => import('@ant-design/charts').then(m => ({ default: m.Line })))
const PieChart = lazy(() => import('@ant-design/charts').then(m => ({ default: m.Pie })))

// 重组件懒加载
const RichTextEditor = lazy(() => import('./RichTextEditor'))
const FileUploader = lazy(() => import('./FileUploader'))
```

#### 3. 图片懒加载
```typescript
// 使用 Intersection Observer
function LazyImage({ src, alt }) {
  const imgRef = useRef<HTMLImageElement>(null)
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsLoaded(true)
          observer.disconnect()
        }
      },
      { threshold: 0.1 }
    )

    if (imgRef.current) {
      observer.observe(imgRef.current)
    }

    return () => observer.disconnect()
  }, [])

  return (
    <img
      ref={imgRef}
      src={isLoaded ? src : 'placeholder.jpg'}
      alt={alt}
      loading="lazy"
    />
  )
}
```

### B. 图表性能优化

#### 1. 数据采样
```typescript
// 超过1000个数据点时采样
function useChartSampling(data: any[], maxPoints = 1000) {
  return useMemo(() => {
    if (data.length <= maxPoints) return data

    const step = Math.ceil(data.length / maxPoints)
    return data.filter((_, index) => index % step === 0)
  }, [data, maxPoints])
}

// 使用
const sampledData = useChartSampling(largeDataset, 500)
```

#### 2. 虚拟滚动图表
```typescript
// 仅渲染可见区域的数据
function useChartWindowing(data: any[], windowSize = 100) {
  const [offset, setOffset] = useState(0)

  const visibleData = useMemo(
    () => data.slice(offset, offset + windowSize),
    [data, offset, windowSize]
  )

  return { visibleData, offset, setOffset }
}
```

#### 3. 防抖更新
```typescript
// 避免频繁重绘
const debouncedData = useChartDebounce(realtimeData, 300)

<Line data={debouncedData} />
```

#### 4. Canvas vs SVG
```typescript
// 大数据量使用 Canvas (>1000 points)
<Line
  data={largeData}
  renderer="canvas"  // 而不是 "svg"
/>

// 小数据量或需要交互使用 SVG
<Pie
  data={smallData}
  renderer="svg"  // 更好的交互性
/>
```

### C. 表格性能优化

#### 1. 虚拟滚动表格
```typescript
import { FixedSizeList } from 'react-window'

function VirtualTable({ data }: { data: any[] }) {
  const Row = ({ index, style }: any) => (
    <div style={style}>
      {data[index].name} - {data[index].amount}
    </div>
  )

  return (
    <FixedSizeList
      height={600}
      itemCount={data.length}
      itemSize={50}
      width="100%"
    >
      {Row}
    </FixedSizeList>
  )
}
```

#### 2. 分页加载
```typescript
// 服务端分页
function useTablePagination(api: string) {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const { data, loading } = useQuery({
    queryKey: [api, page, pageSize],
    queryFn: () => fetch(`${api}?page=${page}&size=${pageSize}`)
  })

  return { data, loading, page, setPage, pageSize, setPageSize }
}
```

#### 3. 批量操作优化
```typescript
// 使用 useMemo 避免重复计算
const selectedRowKeys = useMemo(
  () => selectedRows.map(row => row.id),
  [selectedRows]
)

// 批量操作防抖
const [batchDelete] = useDebounceFn(async (ids: string[]) => {
  await api.batchDelete(ids)
  message.success(`已删除 ${ids.length} 条记录`)
}, 500)
```

### D. 资源优化

#### 1. 图片优化
```typescript
// 使用 WebP 格式
<picture>
  <source srcSet="image.webp" type="image/webp" />
  <source srcSet="image.jpg" type="image/jpeg" />
  <img src="image.jpg" alt="fallback" />
</picture>

// 响应式图片
<img
  srcSet="
    image-320w.jpg 320w,
    image-640w.jpg 640w,
    image-1280w.jpg 1280w
  "
  sizes="(max-width: 640px) 100vw, 50vw"
  src="image-640w.jpg"
  alt="responsive"
/>
```

#### 2. 字体优化
```css
/* 字体子集化 */
@font-face {
  font-family: 'Custom';
  src: url('font.woff2') format('woff2');
  font-display: swap; /* 避免FOIT */
  unicode-range: U+0020-007F; /* 仅加载需要的字符 */
}

/* 预加载关键字体 */
<link rel="preload" href="font.woff2" as="font" type="font/woff2" crossorigin>
```

#### 3. CSS 优化
```typescript
// CSS-in-JS 按需加载
import { css } from '@emotion/react'

const styles = css`
  .component {
    /* 仅在组件使用时加载 */
  }
`

// Critical CSS 内联
// 关键CSS直接内联到HTML
<style>
  .hero { /* 首屏必需样式 */ }
</style>
```

### E. 网络优化

#### 1. API 请求优化
```typescript
// 请求去重
import { useQuery } from '@tanstack/react-query'

const { data } = useQuery({
  queryKey: ['merchants'],
  queryFn: fetchMerchants,
  staleTime: 5 * 60 * 1000, // 5分钟内不重复请求
  cacheTime: 10 * 60 * 1000, // 缓存10分钟
})

// 请求合并
import { useBatchRequest } from './hooks/useBatchRequest'

const results = useBatchRequest([
  { url: '/api/stats', key: 'stats' },
  { url: '/api/payments', key: 'payments' },
  { url: '/api/orders', key: 'orders' },
])
```

#### 2. 预加载
```typescript
// 路由预加载
const Dashboard = lazy(() => import(/* webpackPrefetch: true */ './Dashboard'))

// 数据预加载
function usePrefetch() {
  const queryClient = useQueryClient()

  const prefetchDashboard = () => {
    queryClient.prefetchQuery({
      queryKey: ['dashboard'],
      queryFn: fetchDashboard
    })
  }

  return { prefetchDashboard }
}

// 鼠标悬停时预加载
<Link
  to="/dashboard"
  onMouseEnter={() => prefetchDashboard()}
>
  Dashboard
</Link>
```

#### 3. Service Worker 缓存
```typescript
// vite-plugin-pwa 配置
import { VitePWA } from 'vite-plugin-pwa'

VitePWA({
  registerType: 'autoUpdate',
  workbox: {
    runtimeCaching: [
      {
        urlPattern: /^https:\/\/api\./,
        handler: 'NetworkFirst',
        options: {
          cacheName: 'api-cache',
          expiration: {
            maxEntries: 50,
            maxAgeSeconds: 60 * 60 * 24, // 1天
          },
        },
      },
    ],
  },
})
```

### F. 渲染优化

#### 1. 避免不必要的重渲染
```typescript
// 使用 memo
const ExpensiveComponent = memo(({ data }) => {
  return <div>{/* 复杂渲染 */}</div>
}, (prevProps, nextProps) => {
  // 自定义比较函数
  return prevProps.data.id === nextProps.data.id
})

// 使用 useMemo
const expensiveValue = useMemo(
  () => computeExpensiveValue(data),
  [data]
)

// 使用 useCallback
const handleClick = useCallback(() => {
  console.log(data)
}, [data])
```

#### 2. 虚拟化长列表
```typescript
import { Virtuoso } from 'react-virtuoso'

<Virtuoso
  style={{ height: '600px' }}
  totalCount={10000}
  itemContent={(index) => (
    <div>Item {index}</div>
  )}
/>
```

#### 3. 并发渲染 (React 18)
```typescript
import { startTransition } from 'react'

// 低优先级更新
const handleSearch = (value: string) => {
  setInputValue(value) // 高优先级

  startTransition(() => {
    setSearchResults(filter(value)) // 低优先级
  })
}
```

### G. 监控与分析

#### 1. 性能监控
```typescript
// Web Vitals
import { getCLS, getFID, getLCP } from 'web-vitals'

getCLS(console.log)
getFID(console.log)
getLCP(console.log)

// 自定义性能监控
performance.mark('chart-start')
// ... 渲染图表
performance.mark('chart-end')
performance.measure('chart-render', 'chart-start', 'chart-end')

const measure = performance.getEntriesByName('chart-render')[0]
console.log(`图表渲染耗时: ${measure.duration}ms`)
```

#### 2. 错误监控
```typescript
// Sentry 集成
import * as Sentry from '@sentry/react'

Sentry.init({
  dsn: 'YOUR_DSN',
  integrations: [
    new Sentry.BrowserTracing(),
    new Sentry.Replay(),
  ],
  tracesSampleRate: 0.1,
  replaysSessionSampleRate: 0.1,
})
```

#### 3. 用户行为分析
```typescript
// Google Analytics 4
import ReactGA from 'react-ga4'

ReactGA.initialize('G-MEASUREMENT_ID')

// 页面浏览
ReactGA.send({ hitType: 'pageview', page: window.location.pathname })

// 事件跟踪
ReactGA.event({
  category: 'Payment',
  action: 'Create',
  label: 'Stripe',
  value: 100,
})
```

---

## 📦 构建优化

### Vite 配置优化
```typescript
// vite.config.ts
export default defineConfig({
  build: {
    // 代码分割
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor': ['react', 'react-dom', 'react-router-dom'],
          'antd': ['antd', '@ant-design/icons'],
          'charts': ['@ant-design/charts'],
        },
      },
    },
    // 压缩
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // 生产环境移除 console
        drop_debugger: true,
      },
    },
    // Chunk 大小警告
    chunkSizeWarningLimit: 500,
  },

  // 依赖预构建
  optimizeDeps: {
    include: ['react', 'react-dom', 'antd'],
  },
})
```

### Webpack 配置(如需使用)
```javascript
module.exports = {
  optimization: {
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          priority: -10,
        },
      },
    },
  },

  performance: {
    maxAssetSize: 512000, // 500KB
    maxEntrypointSize: 512000,
  },
}
```

---

## 🎨 用户体验优化

### 1. 加载状态
```typescript
// 骨架屏
<Skeleton active paragraph={{ rows: 4 }} />

// 进度条
import NProgress from 'nprogress'

NProgress.start()
await loadData()
NProgress.done()

// Suspense fallback
<Suspense fallback={<PageLoading />}>
  <Routes />
</Suspense>
```

### 2. 错误处理
```typescript
// Error Boundary
class ErrorBoundary extends Component {
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error:', error, errorInfo)
    Sentry.captureException(error)
  }

  render() {
    if (this.state.hasError) {
      return <ErrorPage />
    }
    return this.props.children
  }
}
```

### 3. 离线支持
```typescript
// Service Worker
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js')
}

// 离线提示
const [isOnline, setIsOnline] = useState(navigator.onLine)

useEffect(() => {
  const handleOnline = () => setIsOnline(true)
  const handleOffline = () => setIsOnline(false)

  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)

  return () => {
    window.removeEventListener('online', handleOnline)
    window.removeEventListener('offline', handleOffline)
  }
}, [])

{!isOnline && <Alert message="您当前处于离线状态" type="warning" />}
```

---

## 📊 性能基准

### 当前性能(估算)

#### Website
- **Bundle Size**: ~350KB (gzipped)
- **LCP**: ~2.1s
- **FID**: ~50ms
- **CLS**: 0.05

#### Admin Portal
- **Bundle Size**: ~600KB (gzipped)
- **LCP**: ~2.8s
- **FID**: ~80ms
- **CLS**: 0.08

#### Merchant Portal
- **Bundle Size**: ~550KB (gzipped)
- **LCP**: ~2.6s
- **FID**: ~70ms
- **CLS**: 0.07

### 优化目标

| 指标 | 当前 | 目标 | 改进 |
|------|------|------|------|
| LCP | 2.5s | <2.0s | -20% |
| FID | 70ms | <50ms | -29% |
| CLS | 0.07 | <0.05 | -29% |
| Bundle | 550KB | <400KB | -27% |

---

## 🚀 实施计划

### Phase 1: 快速优化 (1周)
- ✅ 路由懒加载
- ✅ 图片懒加载
- ✅ 图表防抖
- ✅ 代码分割

### Phase 2: 深度优化 (2周)
- ⏳ 虚拟滚动
- ⏳ Service Worker
- ⏳ 资源压缩
- ⏳ CDN 配置

### Phase 3: 高级优化 (4周)
- ⏳ SSR/SSG
- ⏳ 边缘计算
- ⏳ 智能预加载
- ⏳ 性能监控

---

## 🔍 监控清单

### 开发阶段
- [ ] Bundle Analyzer 分析
- [ ] Lighthouse 审计(>90分)
- [ ] React DevTools Profiler
- [ ] Chrome DevTools Performance

### 生产阶段
- [ ] Web Vitals 监控
- [ ] Error Tracking (Sentry)
- [ ] Analytics (GA4)
- [ ] RUM (Real User Monitoring)

---

## 📚 参考资源

### 工具
- [Lighthouse](https://developers.google.com/web/tools/lighthouse)
- [WebPageTest](https://www.webpagetest.org/)
- [Bundle Analyzer](https://github.com/webpack-contrib/webpack-bundle-analyzer)
- [React DevTools](https://react.dev/learn/react-developer-tools)

### 文档
- [Web.dev Performance](https://web.dev/performance/)
- [React Performance](https://react.dev/learn/render-and-commit)
- [Vite Performance](https://vitejs.dev/guide/performance.html)

---

## ✅ 检查清单

### 代码层面
- [ ] 使用 React.lazy 进行代码分割
- [ ] 使用 useMemo/useCallback 避免重渲染
- [ ] 使用虚拟滚动处理长列表
- [ ] 图表数据采样和防抖
- [ ] 避免内联函数和对象

### 资源层面
- [ ] 图片压缩和 WebP 格式
- [ ] 字体子集化
- [ ] CSS 压缩和去重
- [ ] JavaScript 压缩和混淆
- [ ] Gzip/Brotli 压缩

### 网络层面
- [ ] HTTP/2 或 HTTP/3
- [ ] CDN 加速
- [ ] 资源预加载
- [ ] Service Worker 缓存
- [ ] API 请求合并

### 用户体验
- [ ] 骨架屏加载
- [ ] 错误边界处理
- [ ] 离线支持
- [ ] 响应式设计
- [ ] 可访问性(a11y)

---

**维护者**: Frontend Team
**最后更新**: 2025-10-25
**版本**: v1.0

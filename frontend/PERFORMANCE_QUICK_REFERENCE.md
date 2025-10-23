# 性能优化快速参考

## 🎯 一分钟快速检查

```bash
# 1. 检查包大小
pnpm build
# 查看输出，单个 chunk 不应超过 500KB

# 2. 运行 Lighthouse
# 打开 Chrome DevTools → Lighthouse → 分析页面加载
# 目标：性能评分 > 90

# 3. 检查渲染性能
# React DevTools → Profiler → 录制 → 查看渲染时间
# 单次渲染不应超过 16ms (60fps)
```

## 🔧 常用优化技巧

### 1. React.memo - 避免重渲染

```typescript
// ❌ 父组件更新时，子组件总是重渲染
const MyComponent = ({ data }) => <div>{data}</div>

// ✅ 仅在 props 改变时重渲染
const MyComponent = memo(({ data }) => <div>{data}</div>)
```

**何时使用**:
- ✅ 大列表中的项组件
- ✅ 不频繁更新的组件
- ✅ 纯展示组件
- ❌ 非常简单的组件（优化收益小于成本）

### 2. useMemo - 缓存计算结果

```typescript
// ❌ 每次渲染都重新排序
const sorted = data.sort((a, b) => b.time - a.time)

// ✅ 只在 data 改变时重新排序
const sorted = useMemo(() => {
  return data.sort((a, b) => b.time - a.time)
}, [data])
```

**何时使用**:
- ✅ 复杂的过滤/排序操作
- ✅ 大数组的 map/filter/reduce
- ✅ 格式化大量数据
- ❌ 简单的计算（如 `a + b`）

### 3. useCallback - 缓存函数

```typescript
// ❌ 每次渲染创建新函数
const handleClick = (id) => deleteItem(id)

// ✅ 函数引用不变
const handleClick = useCallback((id) => deleteItem(id), [])
```

**何时使用**:
- ✅ 传递给 memo 组件的回调
- ✅ 作为 useEffect 的依赖
- ✅ 事件处理函数
- ❌ 不传递给子组件的函数

## 📦 导入性能工具

```typescript
// 从共享包导入
import { usePerformance, useAPIPerformance } from '@payment/shared/hooks'

// 使用性能监控
function MyComponent() {
  const { renderCount } = usePerformance('MyComponent')
  const { trackAPI } = useAPIPerformance()

  useEffect(() => {
    const start = Date.now()
    fetchData().then(() => {
      trackAPI('fetchData', Date.now() - start)
    })
  }, [])

  return <div>Render count: {renderCount}</div>
}
```

## 🚨 性能反模式（避免）

### ❌ 1. 在渲染中创建新对象/数组

```typescript
// ❌ 每次渲染都创建新对象
<MyComponent style={{ width: 100, height: 100 }} />

// ✅ 提取到常量
const STYLE = { width: 100, height: 100 }
<MyComponent style={STYLE} />

// 或使用 useMemo
const style = useMemo(() => ({ width: 100, height: 100 }), [])
```

### ❌ 2. 在渲染中执行昂贵操作

```typescript
// ❌ 每次渲染都执行
function MyComponent({ data }) {
  const processed = expensiveOperation(data) // 糟糕!
  return <div>{processed}</div>
}

// ✅ 使用 useMemo
function MyComponent({ data }) {
  const processed = useMemo(() => expensiveOperation(data), [data])
  return <div>{processed}</div>
}
```

### ❌ 3. 过度使用 Context

```typescript
// ❌ 导致所有消费者重渲染
<GlobalContext.Provider value={{ user, theme, settings }}>

// ✅ 拆分 Context
<UserContext.Provider value={user}>
  <ThemeContext.Provider value={theme}>
    <SettingsContext.Provider value={settings}>
```

### ❌ 4. 未使用 key 或使用 index 作为 key

```typescript
// ❌ 使用 index 作为 key（如果列表会重新排序）
{items.map((item, index) => <Item key={index} {...item} />)}

// ✅ 使用稳定的唯一 ID
{items.map((item) => <Item key={item.id} {...item} />)}
```

## 🎨 组件优化模板

### 优化前

```typescript
function PaymentList({ payments }) {
  const [search, setSearch] = useState('')

  // ❌ 问题
  const filtered = payments.filter(p => p.no.includes(search))
  const handleDelete = (id) => deletePayment(id)

  return (
    <>
      {filtered.map(p => (
        <PaymentCard payment={p} onDelete={handleDelete} />
      ))}
    </>
  )
}
```

### 优化后

```typescript
import { memo, useMemo, useCallback } from 'react'

const PaymentCard = memo(({ payment, onDelete }) => {
  return <Card onClick={() => onDelete(payment.id)} />
})

function PaymentList({ payments }) {
  const [search, setSearch] = useState('')

  // ✅ 缓存过滤结果
  const filtered = useMemo(() => {
    return payments.filter(p => p.no.includes(search))
  }, [payments, search])

  // ✅ 缓存函数引用
  const handleDelete = useCallback((id) => {
    deletePayment(id)
  }, [])

  return (
    <>
      {filtered.map(p => (
        <PaymentCard key={p.id} payment={p} onDelete={handleDelete} />
      ))}
    </>
  )
}
```

## 📊 性能指标目标

| 指标 | 优秀 | 良好 | 需要改进 |
|------|------|------|----------|
| Lighthouse | > 90 | 75-90 | < 75 |
| FCP | < 1.8s | 1.8-3s | > 3s |
| LCP | < 2.5s | 2.5-4s | > 4s |
| FID | < 100ms | 100-300ms | > 300ms |
| CLS | < 0.1 | 0.1-0.25 | > 0.25 |
| Bundle 大小 | < 200KB | 200-500KB | > 500KB |
| 组件渲染 | < 16ms | 16-50ms | > 50ms |

## 🔍 快速诊断

### 问题: 页面加载慢

```bash
# 1. 检查 Bundle 大小
pnpm build
# 如果主 bundle > 500KB，考虑代码分割

# 2. 检查网络请求
# Chrome DevTools → Network → 查看瀑布图
# 并行加载，避免串行依赖

# 3. 检查首屏资源
# 移除不必要的库，懒加载非关键组件
```

### 问题: 页面卡顿

```bash
# 1. React DevTools Profiler
# 录制交互，查看哪个组件渲染时间长

# 2. Chrome Performance
# 录制，查看主线程是否阻塞

# 3. 检查列表长度
# 如果 > 1000 项，考虑虚拟滚动
```

### 问题: 点击响应慢

```bash
# 1. 检查事件处理函数
# 是否有同步的大量计算？移到 Web Worker

# 2. 检查状态更新
# 是否导致大量组件重渲染？使用 React.memo

# 3. 检查 useEffect
# 是否频繁触发？优化依赖数组
```

## 🛠️ 开发工具

```typescript
// 1. 启用性能监控（开发环境）
import { usePerformance } from '@payment/shared/hooks'

function MyComponent() {
  usePerformance('MyComponent') // 自动在控制台输出性能指标
  // ...
}

// 2. 启用 React Strict Mode（已启用）
// 检测潜在问题

// 3. 使用 React DevTools Profiler
// 可视化组件渲染

// 4. 使用 why-did-you-render
pnpm add -D @welldone-software/why-did-you-render
```

## 📚 更多资源

- 完整文档: [PERFORMANCE_OPTIMIZATION.md](./PERFORMANCE_OPTIMIZATION.md)
- 示例代码: `shared/src/components/PerformanceExample.tsx`
- React 官方文档: https://react.dev/learn/render-and-commit
- Web.dev 性能指南: https://web.dev/performance/

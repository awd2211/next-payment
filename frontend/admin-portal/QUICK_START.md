# Admin Portal 优化功能快速开始

这份指南帮助您快速了解和使用新增的优化功能。

---

## 🚀 5 分钟快速上手

### 1. 启动开发服务器

```bash
cd frontend/admin-portal
npm install  # 安装新依赖 (如果还没装)
npm run dev  # 启动开发服务器
```

访问: http://localhost:5173

**性能监控**: 按 `Ctrl+Shift+P` 显示/隐藏实时性能监控面板

---

### 2. 最常用的组件

#### ✅ CommonTable - 快速创建表格

```typescript
import CommonTable from '@/components/common/CommonTable'
import { exportToExcel } from '@/utils/exportUtils'

<CommonTable
  columns={columns}
  dataSource={data}
  loading={loading}
  pagination={pagination}
  showRefresh
  showExport
  onRefresh={() => refresh()}
  onExport={() => exportToExcel(data, columns, 'data.xlsx')}
/>
```

**收益**: 减少 30 行样板代码

---

#### ✅ BatchActions - 批量操作

```typescript
import BatchActions, { commonBatchActions } from '@/components/common/BatchActions'

const actions = [
  commonBatchActions.approve(async (keys) => await batchApprove(keys)),
  commonBatchActions.delete(async (keys) => await batchDelete(keys)),
]

<BatchActions
  selectedCount={selectedKeys.length}
  selectedRowKeys={selectedKeys}
  actions={actions}
  onClear={() => setSelectedKeys([])}
/>
```

**收益**: 自动确认对话框、错误处理、加载状态

---

#### ✅ useTable - 表格状态管理

```typescript
import { useTable } from '@/hooks/useTable'

const table = useTable({
  initialPageSize: 20,
  onFetchData: async ({ page, pageSize, filters }) => {
    const res = await api.list({ page, pageSize, ...filters })
    return { data: res.data.list, total: res.data.total }
  }
})

// 使用
<CommonTable {...table} />
```

**收益**: 自动管理分页、筛选、排序、选中行

---

### 3. 最实用的工具函数

#### 💰 金额格式化

```typescript
import { formatAmount } from '@/utils/formatUtils'

formatAmount(10000, 'USD')  // "$100.00"
formatAmount(50000, 'CNY')  // "¥500.00"
```

#### 📅 日期格式化

```typescript
import { formatDateTime, formatRelativeTime } from '@/utils/formatUtils'

formatDateTime(new Date())  // "2025-10-25 10:30:00"
formatRelativeTime(someDate)  // "3分钟前"
```

#### 🔒 敏感数据脱敏

```typescript
import { formatPhone, formatEmail } from '@/utils/formatUtils'

formatPhone('13800138000', true)  // "138****8000"
formatEmail('user@example.com', true)  // "u**r@example.com"
```

#### 📤 数据导出

```typescript
import { exportToExcel } from '@/utils/exportUtils'

exportToExcel(
  data,
  [
    { title: '商户名称', dataIndex: 'name' },
    { title: '金额', dataIndex: 'amount', render: (v) => formatAmount(v) }
  ],
  'merchants.xlsx'
)
```

---

### 4. 键盘快捷键

所有页面默认支持:

| 快捷键 | 功能 |
|--------|------|
| `Ctrl + K` | 打开搜索 |
| `Ctrl + R` | 刷新页面 |
| `Ctrl + S` | 保存 |
| `Ctrl + N` | 新建 |
| `Ctrl + Shift + P` | 显示性能监控 |
| `Escape` | 关闭弹窗 |
| `/` | 聚焦搜索框 |

**自定义快捷键**:

```typescript
import { useKeyboardShortcuts } from '@/hooks/useKeyboard'

useKeyboardShortcuts([
  {
    key: 'f',
    modifiers: ['ctrl'],
    callback: () => openFilter(),
    description: '打开筛选器',
  },
])
```

---

### 5. 安全最佳实践

#### XSS 防护

```typescript
import { validateInput } from '@/utils/securityUtils'

const { valid, sanitized, errors } = validateInput(userInput, {
  maxLength: 200,
  checkXSS: true,
  checkSQL: true,
})

if (!valid) {
  message.error(errors.join(', '))
}
```

#### CSRF Token

```typescript
import { getCSRFToken } from '@/utils/securityUtils'

request.post('/api/payments', data, {
  headers: { 'X-CSRF-Token': getCSRFToken() }
})
```

---

## 📋 常见场景

### 场景 1: 创建带筛选和导出的列表页

```typescript
import { useTable } from '@/hooks/useTable'
import CommonTable from '@/components/common/CommonTable'
import AdvancedFilter from '@/components/common/AdvancedFilter'
import { exportToExcel } from '@/utils/exportUtils'

function ListPage() {
  const table = useTable({ onFetchData: api.list })

  return (
    <>
      <AdvancedFilter
        fields={[
          { name: 'keyword', label: '关键词', type: 'input' },
          { name: 'status', label: '状态', type: 'select', options: [...] },
        ]}
        onSearch={table.updateFilter}
      />

      <CommonTable
        {...table}
        onExport={() => exportToExcel(table.data, columns, 'export.xlsx')}
      />
    </>
  )
}
```

---

### 场景 2: 带批量操作的管理页

```typescript
import { useTable } from '@/hooks/useTable'
import BatchActions, { commonBatchActions } from '@/components/common/BatchActions'

function ManagementPage() {
  const table = useTable({ onFetchData: api.list })

  const batchActions = [
    commonBatchActions.approve(async (keys) => {
      await api.batchApprove(keys)
      table.refresh()
    }),
    commonBatchActions.delete(async (keys) => {
      await api.batchDelete(keys)
      table.refresh()
    }),
  ]

  return (
    <>
      <BatchActions
        selectedCount={table.selectedRowKeys.length}
        selectedRowKeys={table.selectedRowKeys}
        actions={batchActions}
        onClear={() => table.setSelectedRowKeys([])}
      />
      <CommonTable {...table} rowSelection={table.rowSelection} />
    </>
  )
}
```

---

### 场景 3: 优化图表性能

```typescript
import { useOptimizedChart } from '@/hooks/useChartOptimization'
import { Line } from '@ant-design/charts'

function ChartPage() {
  const [rawData, setRawData] = useState(largeDataset)

  const optimizedData = useOptimizedChart(rawData, {
    debounceDelay: 300,
    maxPoints: 1000,
    enableSampling: true,
  })

  return <Line data={optimizedData} />
}
```

---

## 🎯 性能优化检查清单

使用这些优化后,你的页面应该满足:

- ✅ 首屏加载 < 3 秒
- ✅ FPS >= 55 (使用虚拟滚动处理大列表)
- ✅ 无重复请求 (使用 React Query)
- ✅ 搜索输入防抖 >= 300ms
- ✅ 滚动事件节流 >= 200ms
- ✅ 图表数据采样 (<= 1000 点)
- ✅ 所有用户输入已验证
- ✅ 敏感数据已脱敏

---

## 💡 开发技巧

### 1. 快速调试性能

按 `Ctrl+Shift+P` 打开性能监控,实时查看:
- FPS (帧率)
- 内存使用
- 页面加载时间

### 2. 快速导出数据

在任何表格组件上添加:

```typescript
<CommonTable
  {...props}
  showExport
  onExport={() => exportToExcel(data, columns, 'export.xlsx')}
/>
```

### 3. 快速添加骨架屏

```typescript
import SkeletonLoading from '@/components/SkeletonLoading'

{loading ? <SkeletonLoading type="table" /> : <Table {...props} />}
```

### 4. 快速添加防抖搜索

```typescript
import { useDebounce } from '@/hooks/useDebounce'

const [keyword, setKeyword] = useState('')
const debouncedKeyword = useDebounce(keyword, 500)

useEffect(() => {
  search(debouncedKeyword)
}, [debouncedKeyword])
```

---

## 📚 深入学习

- **详细示例**: 查看 `USAGE_EXAMPLES.md`
- **优化总结**: 查看 `OPTIMIZATION_SUMMARY.md`
- **完整报告**: 查看 `FINAL_OPTIMIZATION_REPORT.md`

---

## ❓ 常见问题

### Q: React Query DevTools 在哪里?
A: 开发环境下会自动显示在页面右下角

### Q: 如何禁用性能监控?
A: 按 `Ctrl+Shift+P` 或设置 `<PerformanceMonitor enabled={false} />`

### Q: 虚拟表格适用于多少数据?
A: 建议 >1000 行时使用,小于 100 行用普通表格即可

### Q: 缓存策略如何选择?
A:
- 临时数据 → memory
- 持久数据 → localStorage
- 会话数据 → sessionStorage

---

## 🎉 开始使用

现在你已经了解了所有核心功能,开始在项目中使用吧!

如有问题,查看文档或提 Issue: https://github.com/anthropics/claude-code/issues

**祝开发愉快! 🚀**

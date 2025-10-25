# Admin Portal 组件和工具使用示例

本文档提供所有新增组件和工具的实用示例。

---

## 📋 目录

1. [批量操作 (BatchActions)](#1-批量操作)
2. [高级筛选器 (AdvancedFilter)](#2-高级筛选器)
3. [通用表格 (CommonTable)](#3-通用表格)
4. [通用 Modal (CommonModal)](#4-通用-modal)
5. [数据导出 (Export Utils)](#5-数据导出)
6. [格式化工具 (Format Utils)](#6-格式化工具)
7. [安全工具 (Security Utils)](#7-安全工具)
8. [防抖节流 (Debounce/Throttle)](#8-防抖节流)
9. [确认对话框 (useConfirm)](#9-确认对话框)
10. [图表优化 (Chart Optimization)](#10-图表优化)

---

## 1. 批量操作

### 基础使用

```typescript
import BatchActions, { commonBatchActions } from '@/components/common/BatchActions'

function MerchantList() {
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  const batchActions = [
    commonBatchActions.approve(async (keys) => {
      await merchantService.batchApprove(keys)
      refresh()
    }),
    commonBatchActions.reject(async (keys) => {
      await merchantService.batchReject(keys)
      refresh()
    }),
    commonBatchActions.export((keys) => {
      const selectedData = data.filter(item => keys.includes(item.id))
      exportToExcel(selectedData, columns, 'selected-merchants.xlsx')
    }),
  ]

  return (
    <>
      <BatchActions
        selectedCount={selectedRowKeys.length}
        selectedRowKeys={selectedRowKeys}
        actions={batchActions}
        onClear={() => setSelectedRowKeys([])}
      />

      <Table
        rowSelection={{
          selectedRowKeys,
          onChange: setSelectedRowKeys,
        }}
        // ...
      />
    </>
  )
}
```

### 自定义批量操作

```typescript
const customActions: BatchAction[] = [
  {
    key: 'freeze',
    label: '批量冻结',
    icon: <LockOutlined />,
    danger: true,
    needConfirm: true,
    confirmMessage: '冻结后商户将无法进行交易,确定要冻结吗?',
    onClick: async (keys) => {
      await merchantService.batchFreeze(keys)
    },
  },
]
```

---

## 2. 高级筛选器

### 基础使用

```typescript
import AdvancedFilter from '@/components/common/AdvancedFilter'

function PaymentList() {
  const filterFields: FilterField[] = [
    {
      name: 'payment_no',
      label: '支付单号',
      type: 'input',
      placeholder: '请输入支付单号',
    },
    {
      name: 'status',
      label: '支付状态',
      type: 'select',
      options: [
        { label: '待支付', value: 'pending' },
        { label: '成功', value: 'success' },
        { label: '失败', value: 'failed' },
      ],
    },
    {
      name: 'date_range',
      label: '创建时间',
      type: 'dateRange',
    },
    {
      name: 'amount_min',
      label: '最小金额',
      type: 'number',
      span: 6,
    },
    {
      name: 'amount_max',
      label: '最大金额',
      type: 'number',
      span: 6,
    },
  ]

  const handleSearch = (values: Record<string, any>) => {
    console.log('Search values:', values)
    // 调用API进行搜索
  }

  return (
    <AdvancedFilter
      fields={filterFields}
      onSearch={handleSearch}
      onReset={() => console.log('Reset filters')}
      defaultExpanded={false}
      collapsedRowCount={1}
    />
  )
}
```

---

## 3. 通用表格

### 基础使用

```typescript
import CommonTable from '@/components/common/CommonTable'
import { exportToExcel } from '@/utils/exportUtils'

function MerchantTable() {
  const { data, loading, pagination, refresh } = useTable({
    onFetchData: merchantService.list
  })

  const handleExport = () => {
    exportToExcel(data, columns, 'merchants.xlsx')
  }

  return (
    <CommonTable
      columns={columns}
      dataSource={data}
      loading={loading}
      pagination={pagination}
      showRefresh
      showExport
      onRefresh={refresh}
      onExport={handleExport}
      title="商户列表"
    />
  )
}
```

---

## 4. 通用 Modal

### 创建 Modal

```typescript
import CommonModal from '@/components/common/CommonModal'
import { Form, Input } from 'antd'

function CreateMerchantModal() {
  const [form] = Form.useForm()
  const [visible, setVisible] = useState(false)

  const handleSubmit = async (values: any) => {
    await merchantService.create(values)
    setVisible(false)
  }

  return (
    <>
      <Button onClick={() => setVisible(true)}>新建商户</Button>

      <CommonModal
        type="create"
        visible={visible}
        form={form}
        onSubmit={handleSubmit}
        onCancel={() => setVisible(false)}
        successMessage="商户创建成功"
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="商户名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}>
            <Input />
          </Form.Item>
        </Form>
      </CommonModal>
    </>
  )
}
```

---

## 5. 数据导出

### 导出为 Excel

```typescript
import { exportToExcel } from '@/utils/exportUtils'

const handleExport = () => {
  exportToExcel(
    merchants,
    [
      { title: '商户名称', dataIndex: 'name' },
      { title: '邮箱', dataIndex: 'email' },
      {
        title: '金额',
        dataIndex: 'amount',
        render: (val) => formatAmount(val, 'USD')
      }
    ],
    'merchants.xlsx'
  )
}
```

### 多工作表导出

```typescript
import { exportMultipleSheets } from '@/utils/exportUtils'

const handleExportAll = () => {
  exportMultipleSheets(
    [
      {
        name: '商户列表',
        data: merchants,
        columns: merchantColumns
      },
      {
        name: '支付记录',
        data: payments,
        columns: paymentColumns
      }
    ],
    'report.xlsx'
  )
}
```

---

## 6. 格式化工具

### 金额格式化

```typescript
import { formatAmount, formatNumber } from '@/utils/formatUtils'

// 分转元,添加货币符号
formatAmount(10000, 'USD')  // → "$100.00"
formatAmount(50000, 'CNY')  // → "¥500.00"

// 数字千分位
formatNumber(1000000)  // → "1,000,000"
```

### 日期格式化

```typescript
import { formatDateTime, formatRelativeTime } from '@/utils/formatUtils'

formatDateTime(new Date())  // → "2025-10-25 10:30:00"
formatRelativeTime(new Date(Date.now() - 3600000))  // → "1小时前"
```

### 敏感数据脱敏

```typescript
import { formatPhone, formatEmail, formatBankCard } from '@/utils/formatUtils'

formatPhone('13800138000', true)  // → "138****8000"
formatEmail('user@example.com', true)  // → "u**r@example.com"
formatBankCard('6222021234567890', true)  // → "6222 **** **** 7890"
```

---

## 7. 安全工具

### XSS 防护

```typescript
import { escapeHtml, detectXSS, validateInput } from '@/utils/securityUtils'

// HTML 实体编码
const safe = escapeHtml('<script>alert("xss")</script>')

// 检测 XSS
if (detectXSS(userInput)) {
  message.error('输入包含不安全内容')
}

// 输入验证和清理
const { valid, sanitized, errors } = validateInput(userInput, {
  maxLength: 200,
  checkXSS: true,
  checkSQL: true
})
```

### CSRF 防护

```typescript
import { getCSRFToken } from '@/utils/securityUtils'

// 在请求中添加 CSRF Token
const token = getCSRFToken()
request.post('/api/payments', data, {
  headers: { 'X-CSRF-Token': token }
})
```

### 密码强度检查

```typescript
import { checkPasswordStrength } from '@/utils/securityUtils'

const { score, level, feedback } = checkPasswordStrength('MyPassword123!')
// score: 0-4
// level: 'very-weak' | 'weak' | 'medium' | 'strong' | 'very-strong'
// feedback: ['建议包含特殊字符']
```

---

## 8. 防抖节流

### 搜索框防抖

```typescript
import { useDebounce } from '@/hooks/useDebounce'

function SearchInput() {
  const [keyword, setKeyword] = useState('')
  const debouncedKeyword = useDebounce(keyword, 500)

  useEffect(() => {
    if (debouncedKeyword) {
      searchData(debouncedKeyword)
    }
  }, [debouncedKeyword])

  return <Input value={keyword} onChange={e => setKeyword(e.target.value)} />
}
```

### 函数防抖

```typescript
import { useDebounceFn } from '@/hooks/useDebounce'

function FilterPanel() {
  const [search, cancel] = useDebounceFn((value: string) => {
    console.log('Searching:', value)
  }, 500)

  return <Input onChange={e => search(e.target.value)} />
}
```

### 滚动事件节流

```typescript
import { useThrottle } from '@/hooks/useDebounce'

function ScrollListener() {
  const [handleScroll] = useThrottle(() => {
    console.log('Scroll position:', window.scrollY)
  }, 200)

  useEffect(() => {
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [handleScroll])
}
```

---

## 9. 确认对话框

### 基础确认

```typescript
import { useConfirm } from '@/hooks/useConfirm'

function DeleteButton({ id }: { id: string }) {
  const { confirm } = useConfirm()

  const handleDelete = async () => {
    const result = await confirm(
      () => merchantService.delete(id),
      {
        title: '删除确认',
        content: '删除后无法恢复,确定要删除吗?',
        danger: true
      }
    )

    if (result) {
      message.success('删除成功')
      refresh()
    }
  }

  return <Button danger onClick={handleDelete}>删除</Button>
}
```

### 批量操作确认

```typescript
import { useBatchConfirm } from '@/hooks/useConfirm'

function BatchDeleteButton() {
  const { confirm } = useBatchConfirm()

  const handleBatchDelete = async () => {
    const result = await confirm(
      selectedItems,
      (items) => merchantService.batchDelete(items.map(i => i.id)),
      {
        title: '批量删除确认',
        itemName: '条商户',
        warningMessage: '删除后无法恢复!'
      }
    )

    if (result) {
      refresh()
    }
  }

  return <Button onClick={handleBatchDelete}>批量删除</Button>
}
```

### 撤销/重做

```typescript
import { useHistory } from '@/hooks/useConfirm'

function FormWithHistory() {
  const { state, set, undo, redo, canUndo, canRedo } = useHistory({ name: '', email: '' })

  return (
    <>
      <Space>
        <Button onClick={undo} disabled={!canUndo}>撤销</Button>
        <Button onClick={redo} disabled={!canRedo}>重做</Button>
      </Space>

      <Form.Item label="Name">
        <Input
          value={state.name}
          onChange={e => set({ ...state, name: e.target.value })}
        />
      </Form.Item>
    </>
  )
}
```

---

## 10. 图表优化

### 防抖优化

```typescript
import { useChartDebounce } from '@/hooks/useChartOptimization'

function DashboardChart() {
  const [rawData, setRawData] = useState([])
  const debouncedData = useChartDebounce(rawData, 300)

  return <Line data={debouncedData} />
}
```

### 懒加载

```typescript
import { useChartLazyLoad } from '@/hooks/useChartOptimization'

function LazyChart() {
  const chartRef = useRef<HTMLDivElement>(null)
  const isVisible = useChartLazyLoad(chartRef)

  return (
    <div ref={chartRef}>
      {isVisible ? <Line data={data} /> : <Skeleton />}
    </div>
  )
}
```

### 大数据采样

```typescript
import { useChartSampling } from '@/hooks/useChartOptimization'

function BigDataChart() {
  const sampledData = useChartSampling(largeDataset, 1000)

  return <Line data={sampledData} />
}
```

### 综合优化

```typescript
import { useOptimizedChart } from '@/hooks/useChartOptimization'

function OptimizedChart() {
  const optimizedData = useOptimizedChart(rawData, {
    debounceDelay: 300,
    maxPoints: 1000,
    enableSampling: true,
    enableDebounce: true
  })

  return <Line data={optimizedData} />
}
```

---

## 🎯 实战示例:完整的商户管理页面

```typescript
import { useState } from 'react'
import { Button, Space } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import CommonTable from '@/components/common/CommonTable'
import BatchActions, { commonBatchActions } from '@/components/common/BatchActions'
import AdvancedFilter from '@/components/common/AdvancedFilter'
import CommonModal from '@/components/common/CommonModal'
import { useTable } from '@/hooks/useTable'
import { exportToExcel } from '@/utils/exportUtils'
import { formatAmount, formatDateTime } from '@/utils/formatUtils'

function MerchantManagement() {
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  // 表格管理
  const {
    data,
    loading,
    pagination,
    filters,
    updateFilter,
    refresh,
    selectedRowKeys,
    rowSelection
  } = useTable({
    onFetchData: merchantService.list
  })

  // 筛选字段
  const filterFields = [
    { name: 'keyword', label: '商户名称', type: 'input' as const },
    { name: 'status', label: '状态', type: 'select' as const, options: [...] },
    { name: 'date_range', label: '创建时间', type: 'dateRange' as const },
  ]

  // 表格列
  const columns = [
    { title: '商户名称', dataIndex: 'name' },
    { title: '邮箱', dataIndex: 'email' },
    {
      title: '金额',
      dataIndex: 'amount',
      render: (val: number) => formatAmount(val, 'USD')
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      render: (val: string) => formatDateTime(val)
    },
  ]

  // 批量操作
  const batchActions = [
    commonBatchActions.approve(async (keys) => {
      await merchantService.batchApprove(keys)
      refresh()
    }),
    commonBatchActions.export((keys) => {
      const selectedData = data.filter(item => keys.includes(item.id))
      exportToExcel(selectedData, columns, 'merchants.xlsx')
    }),
  ]

  return (
    <div>
      {/* 筛选器 */}
      <AdvancedFilter
        fields={filterFields}
        onSearch={updateFilter}
        onReset={() => refresh()}
      />

      {/* 批量操作 */}
      <BatchActions
        selectedCount={selectedRowKeys.length}
        selectedRowKeys={selectedRowKeys}
        actions={batchActions}
        onClear={() => rowSelection.onChange([], [])}
      />

      {/* 表格 */}
      <CommonTable
        columns={columns}
        dataSource={data}
        loading={loading}
        pagination={pagination}
        rowSelection={rowSelection}
        showRefresh
        showExport
        onRefresh={refresh}
        onExport={() => exportToExcel(data, columns, 'all-merchants.xlsx')}
        toolbarExtra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
            新建商户
          </Button>
        }
      />

      {/* 创建Modal */}
      <CommonModal
        type="create"
        visible={modalVisible}
        form={form}
        onSubmit={async (values) => {
          await merchantService.create(values)
          refresh()
        }}
        onCancel={() => setModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          {/* 表单字段 */}
        </Form>
      </CommonModal>
    </div>
  )
}
```

---

## 📝 总结

所有组件和工具都已经过优化,可以直接在项目中使用。建议:

1. **性能**: 使用防抖节流优化频繁操作
2. **安全**: 始终验证和清理用户输入
3. **体验**: 使用骨架屏和加载状态
4. **复用**: 优先使用通用组件而非重复代码

**下一步**: 可以参考这些示例,逐步重构现有页面。

import React, { useState, useMemo, useCallback, memo } from 'react'
import { Card, Button, Table } from 'antd'
import type { Payment } from '../types'

/**
 * 性能优化示例组件
 * 展示 React.memo、useMemo、useCallback 的正确用法
 */

// ========== 1. 使用 React.memo 避免不必要的重渲染 ==========

interface PaymentCardProps {
  payment: Payment
  onDelete: (id: string) => void
}

// ❌ 不好的做法：每次父组件更新都会重渲染
// const PaymentCard = ({ payment, onDelete }: PaymentCardProps) => { ... }

// ✅ 好的做法：使用 memo 包裹，仅在 props 改变时重渲染
const PaymentCard = memo(({ payment, onDelete }: PaymentCardProps) => {
  console.log('PaymentCard rendered:', payment.id)

  return (
    <Card
      title={`Payment #${payment.payment_no}`}
      extra={<Button danger onClick={() => onDelete(payment.id)}>删除</Button>}
    >
      <p>金额: ¥{(payment.amount / 100).toFixed(2)}</p>
      <p>状态: {payment.status}</p>
    </Card>
  )
})

PaymentCard.displayName = 'PaymentCard'

// ========== 2. 使用 useMemo 缓存复杂计算 ==========

interface PaymentListProps {
  payments: Payment[]
}

export const PaymentList: React.FC<PaymentListProps> = ({ payments }) => {
  const [searchText, setSearchText] = useState('')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')

  // ✅ 使用 useMemo 缓存过滤和排序结果
  const processedPayments = useMemo(() => {
    console.log('Processing payments...')

    // 1. 过滤
    let filtered = payments
    if (searchText) {
      filtered = payments.filter((p) =>
        p.payment_no.toLowerCase().includes(searchText.toLowerCase()),
      )
    }

    // 2. 排序
    const sorted = [...filtered].sort((a, b) => {
      const timeA = new Date(a.created_at).getTime()
      const timeB = new Date(b.created_at).getTime()
      return sortOrder === 'asc' ? timeA - timeB : timeB - timeA
    })

    return sorted
  }, [payments, searchText, sortOrder]) // 仅在依赖改变时重新计算

  // ✅ 使用 useMemo 缓存统计数据
  const statistics = useMemo(() => {
    console.log('Calculating statistics...')

    return {
      total: processedPayments.length,
      totalAmount: processedPayments.reduce((sum, p) => sum + p.amount, 0),
      successCount: processedPayments.filter((p) => p.status === 'success').length,
      pendingCount: processedPayments.filter((p) => p.status === 'pending').length,
    }
  }, [processedPayments])

  // ========== 3. 使用 useCallback 缓存函数引用 ==========

  // ❌ 不好的做法：每次渲染都创建新函数
  // const handleDelete = (id: string) => { ... }

  // ✅ 好的做法：使用 useCallback 缓存函数引用
  const handleDelete = useCallback((id: string) => {
    console.log('Deleting payment:', id)
    // API 调用删除支付
  }, []) // 空依赖数组，函数引用永不变化

  const handleSearch = useCallback((value: string) => {
    setSearchText(value)
  }, [])

  const handleSort = useCallback(() => {
    setSortOrder((prev) => (prev === 'asc' ? 'desc' : 'asc'))
  }, [])

  return (
    <div>
      {/* 统计信息 */}
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', gap: 16 }}>
          <div>总数: {statistics.total}</div>
          <div>总金额: ¥{(statistics.totalAmount / 100).toFixed(2)}</div>
          <div>成功: {statistics.successCount}</div>
          <div>待处理: {statistics.pendingCount}</div>
        </div>
      </Card>

      {/* 搜索和排序 */}
      <div style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
        <input
          type="text"
          placeholder="搜索支付单号"
          value={searchText}
          onChange={(e) => handleSearch(e.target.value)}
          style={{ flex: 1, padding: '8px' }}
        />
        <Button onClick={handleSort}>
          排序: {sortOrder === 'asc' ? '↑ 升序' : '↓ 降序'}
        </Button>
      </div>

      {/* 支付列表 */}
      <div style={{ display: 'grid', gap: 16 }}>
        {processedPayments.map((payment) => (
          <PaymentCard key={payment.id} payment={payment} onDelete={handleDelete} />
        ))}
      </div>
    </div>
  )
}

// ========== 4. 虚拟化长列表示例 ==========

/**
 * 对于超长列表（1000+ 项），使用虚拟滚动
 * 需要安装: pnpm add react-window
 */

/*
import { FixedSizeList } from 'react-window'

export const VirtualizedPaymentList: React.FC<PaymentListProps> = ({ payments }) => {
  const Row = ({ index, style }: any) => (
    <div style={style}>
      <PaymentCard payment={payments[index]} onDelete={() => {}} />
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
*/

// ========== 5. Table 组件性能优化 ==========

export const OptimizedPaymentTable: React.FC<PaymentListProps> = ({ payments }) => {
  // ✅ 使用 useMemo 定义 columns，避免每次渲染都重新创建
  const columns = useMemo(
    () => [
      {
        title: '支付单号',
        dataIndex: 'payment_no',
        key: 'payment_no',
      },
      {
        title: '金额',
        dataIndex: 'amount',
        key: 'amount',
        render: (amount: number) => `¥${(amount / 100).toFixed(2)}`,
      },
      {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
      },
      {
        title: '创建时间',
        dataIndex: 'created_at',
        key: 'created_at',
      },
    ],
    [],
  )

  return (
    <Table
      columns={columns}
      dataSource={payments}
      rowKey="id"
      pagination={{
        pageSize: 20,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 条`,
      }}
    />
  )
}

export default PaymentList

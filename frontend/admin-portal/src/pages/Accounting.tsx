import { useState, useEffect } from 'react'
import { Card, Table, Tabs, Statistic, Row, Col, Select, Space, Tag } from 'antd'
import { DollarOutlined, SwapOutlined, RiseOutlined, FallOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { accountingService, type Transaction } from '../services/accountingService'

export default function Accounting() {
  const [loading, setLoading] = useState(false)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [activeTab, setActiveTab] = useState<string>('all')
  const [totalIn, setTotalIn] = useState(0)
  const [totalOut, setTotalOut] = useState(0)

  useEffect(() => {
    fetchData()
  }, [activeTab])

  const fetchData = async () => {
    setLoading(true)
    try {
      const params: any = {
        page: 1,
        page_size: 50,
      }

      // 根据 tab 筛选交易类型
      if (activeTab !== 'all') {
        params.transaction_type = activeTab
      }

      const response = await accountingService.listTransactions(params)

      // 响应拦截器已解包，直接使用数据
      if (response && response.list) {
        setTransactions(response.list)

        // 计算总入账和总出账
        const totalInAmount = response.list
          .filter((t: Transaction) => t.amount > 0)
          .reduce((sum: number, t: Transaction) => sum + t.amount, 0)
        const totalOutAmount = response.list
          .filter((t: Transaction) => t.amount < 0)
          .reduce((sum: number, t: Transaction) => sum + Math.abs(t.amount), 0)

        setTotalIn(totalInAmount)
        setTotalOut(totalOutAmount)
      }
    } catch (error) {
      console.error('Failed to fetch transactions:', error)
      // 如果API失败，使用空数据
      setTransactions([])
      setTotalIn(0)
      setTotalOut(0)
    } finally {
      setLoading(false)
    }
  }

  const columns: ColumnsType<Transaction> = [
    {
      title: '交易流水号',
      dataIndex: 'transaction_no',
      width: 200,
      fixed: 'left',
    },
    {
      title: '交易类型',
      dataIndex: 'transaction_type',
      width: 120,
      render: (type: string) => {
        const typeMap: Record<string, { text: string; color: string }> = {
          payment_in: { text: '支付入账', color: 'green' },
          refund_out: { text: '退款出账', color: 'orange' },
          withdraw: { text: '提现', color: 'red' },
          fee: { text: '手续费', color: 'purple' },
          adjustment: { text: '调账', color: 'blue' },
        }
        const config = typeMap[type] || { text: type, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '金额',
      dataIndex: 'amount',
      width: 150,
      align: 'right',
      render: (amount: number, record: Transaction) => {
        const value = (amount / 100).toFixed(2)
        const isPositive = amount > 0
        return (
          <span style={{ color: isPositive ? '#52c41a' : '#ff4d4f' }}>
            {isPositive ? '+' : ''}{value} {record.currency}
          </span>
        )
      },
    },
    {
      title: '货币',
      dataIndex: 'currency',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: string) => {
        const statusMap: Record<string, { text: string; color: string }> = {
          pending: { text: '待处理', color: 'default' },
          completed: { text: '已完成', color: 'success' },
          failed: { text: '失败', color: 'error' },
          reversed: { text: '已冲正', color: 'warning' },
        }
        const config = statusMap[status] || { text: status, color: 'default' }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
    },
  ]

  const tabItems = [
    {
      key: 'all',
      label: '全部交易',
    },
    {
      key: 'payment_in',
      label: '支付入账',
    },
    {
      key: 'refund_out',
      label: '退款出账',
    },
    {
      key: 'withdraw',
      label: '提现',
    },
    {
      key: 'fee',
      label: '手续费',
    },
    {
      key: 'adjustment',
      label: '调账',
    },
  ]

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="总入账"
              value={totalIn / 100}
              precision={2}
              valueStyle={{ color: '#3f8600' }}
              prefix={<RiseOutlined />}
              suffix="USD"
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="总出账"
              value={totalOut / 100}
              precision={2}
              valueStyle={{ color: '#cf1322' }}
              prefix={<FallOutlined />}
              suffix="USD"
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="净额"
              value={(totalIn - totalOut) / 100}
              precision={2}
              valueStyle={{ color: totalIn >= totalOut ? '#3f8600' : '#cf1322' }}
              prefix={<DollarOutlined />}
              suffix="USD"
            />
          </Card>
        </Col>
      </Row>

      {/* 交易列表 */}
      <Card
        title={
          <Space>
            <SwapOutlined />
            <span>账户交易记录</span>
          </Space>
        }
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ marginBottom: 16 }}
        />

        <Table
          columns={columns}
          dataSource={transactions}
          rowKey="id"
          loading={loading}
          pagination={{
            total: transactions.length,
            pageSize: 50,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  )
}

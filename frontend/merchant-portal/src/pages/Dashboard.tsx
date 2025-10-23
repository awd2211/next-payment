import { useState, useEffect } from 'react'
import { Row, Col, Card, Statistic, Typography, Table } from 'antd'
import {
  DollarOutlined,
  TransactionOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { Line, Pie, Column } from '@ant-design/charts'
import { paymentService, Payment, PaymentStats } from '../services/paymentService'
import dayjs from 'dayjs'

const { Title } = Typography

interface Transaction {
  id: string
  order_no: string
  amount: number
  status: string
  created_at: string
}

interface TrendData {
  date: string
  value: number
  type: string
}

interface ChannelData {
  channel: string
  value: number
}

const Dashboard = () => {
  const [loading, setLoading] = useState(false)
  const [stats, setStats] = useState<PaymentStats | null>(null)
  const [recentPayments, setRecentPayments] = useState<Payment[]>([])
  const [trendData, setTrendData] = useState<TrendData[]>([])
  const [channelData, setChannelData] = useState<ChannelData[]>([])
  const [methodData, setMethodData] = useState<ChannelData[]>([])

  useEffect(() => {
    loadStats()
    loadRecentPayments()
    loadTrendData()
    loadChannelData()
    loadMethodData()
  }, [])

  const loadStats = async () => {
    try {
      const response = await paymentService.getStats({})
      setStats(response.data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const loadRecentPayments = async () => {
    try {
      const response = await paymentService.list({
        page: 1,
        page_size: 5,
      })
      setRecentPayments(response.data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const loadTrendData = async () => {
    setLoading(true)
    try {
      // Get data for last 7 days
      const data: TrendData[] = []
      for (let i = 6; i >= 0; i--) {
        const date = dayjs().subtract(i, 'day')
        const startTime = date.startOf('day').toISOString()
        const endTime = date.endOf('day').toISOString()

        const response = await paymentService.getStats({
          start_time: startTime,
          end_time: endTime,
        })

        const dateStr = date.format('MM-DD')
        data.push({
          date: dateStr,
          value: response.data.total_amount / 100,
          type: '交易额',
        })
        data.push({
          date: dateStr,
          value: response.data.total_count,
          type: '交易笔数',
        })
      }
      setTrendData(data)
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const loadChannelData = async () => {
    try {
      // Fetch payments grouped by channel
      const channels = ['stripe', 'paypal', 'alipay', 'wechat']
      const data: ChannelData[] = []

      for (const channel of channels) {
        const response = await paymentService.list({
          page: 1,
          page_size: 1,
          channel,
        })

        if (response.pagination.total > 0) {
          data.push({
            channel: channel === 'stripe' ? 'Stripe' :
                     channel === 'paypal' ? 'PayPal' :
                     channel === 'alipay' ? '支付宝' : '微信支付',
            value: response.pagination.total,
          })
        }
      }
      setChannelData(data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const loadMethodData = async () => {
    try {
      // Fetch payments grouped by method
      const methods = ['card', 'bank_transfer', 'e_wallet']
      const data: ChannelData[] = []

      for (const method of methods) {
        const response = await paymentService.list({
          page: 1,
          page_size: 1,
          method,
        })

        if (response.pagination.total > 0) {
          data.push({
            channel: method === 'card' ? '信用卡' :
                     method === 'bank_transfer' ? '银行转账' : '电子钱包',
            value: response.pagination.total,
          })
        }
      }
      setMethodData(data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const recentTransactions: Transaction[] = []

  const columns: ColumnsType<Payment> = [
    {
      title: '订单号',
      dataIndex: 'order_id',
      key: 'order_id',
      ellipsis: true,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number, record) => `${record.currency} ${(amount / 100).toFixed(2)}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusMap: Record<string, string> = {
          pending: '待支付',
          success: '成功',
          failed: '失败',
          cancelled: '已取消',
          refunded: '已退款',
        }
        return statusMap[status] || status
      },
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => dayjs(time).format('MM-DD HH:mm'),
    },
  ]

  const lineConfig = {
    data: trendData,
    xField: 'date',
    yField: 'value',
    seriesField: 'type',
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
    legend: {
      position: 'top' as const,
    },
  }

  const pieConfig = {
    data: channelData,
    angleField: 'value',
    colorField: 'channel',
    radius: 0.8,
    label: {
      type: 'outer' as const,
      content: '{name} {percentage}',
    },
    interactions: [
      {
        type: 'element-active',
      },
    ],
    legend: {
      position: 'bottom' as const,
    },
  }

  const columnConfig = {
    data: methodData,
    xField: 'channel',
    yField: 'value',
    label: {
      position: 'top' as const,
      style: {
        fill: '#000000',
        opacity: 0.6,
      },
    },
    xAxis: {
      label: {
        autoHide: true,
        autoRotate: false,
      },
    },
    meta: {
      channel: {
        alias: '支付方式',
      },
      value: {
        alias: '交易笔数',
      },
    },
  }

  return (
    <div>
      <Title level={2}>概览</Title>

      {/* Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总交易额"
              value={stats ? stats.total_amount / 100 : 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="USD"
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总交易笔数"
              value={stats?.total_count || 0}
              prefix={<TransactionOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="成功交易"
              value={stats?.success_count || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="成功率"
              value={stats ? stats.success_rate * 100 : 0}
              precision={2}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Charts */}
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={16}>
          <Card title="交易趋势（最近7天）" loading={loading}>
            <Line {...lineConfig} />
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="支付渠道分布">
            <Pie {...pieConfig} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="支付方式统计">
            <Column {...columnConfig} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="最近交易">
            <Table
              columns={columns}
              dataSource={recentPayments}
              rowKey="id"
              pagination={false}
              size="small"
              locale={{ emptyText: '暂无交易记录' }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard

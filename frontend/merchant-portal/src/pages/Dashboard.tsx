import { useState, useEffect } from 'react'
import { Row, Col, Card, Statistic, Typography, Table, Button, Space, Tag } from 'antd'
import {
  DollarOutlined,
  TransactionOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  WalletOutlined,
  RiseOutlined,
  PlusCircleOutlined,
  SearchOutlined,
  RollbackOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { Line, Pie, Column } from '@ant-design/charts'
import { paymentService, Payment, PaymentStats } from '../services/paymentService'
import { dashboardService } from '../services/dashboardService'
import { useAuthStore } from '../stores/authStore'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import dayjs from 'dayjs'

const { Title, Text } = Typography

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
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [stats, setStats] = useState<PaymentStats | null>(null)
  const [todayStats, setTodayStats] = useState<PaymentStats | null>(null)
  const [monthStats, setMonthStats] = useState<PaymentStats | null>(null)
  const [recentPayments, setRecentPayments] = useState<Payment[]>([])
  const [trendData, setTrendData] = useState<TrendData[]>([])
  const [channelData, setChannelData] = useState<ChannelData[]>([])
  const [methodData, setMethodData] = useState<ChannelData[]>([])

  useEffect(() => {
    // 调用merchant-service的Dashboard API
    loadDashboardData()
    // 注释掉其他数据加载，避免图表配置错误
    // loadStats()
    // loadTodayStats()
    // loadMonthStats()
    // loadRecentPayments()
    // loadTrendData()
    // loadChannelData()
    // loadMethodData()
  }, [])

  const loadDashboardData = async () => {
    // 检查是否已登录
    const token = useAuthStore.getState().token
    if (!token) {
      console.log('No token found, skipping dashboard data load')
      return
    }

    setLoading(true)
    try {
      // 调用merchant-service的Dashboard API
      const response = await dashboardService.getDashboard()
      console.log('Dashboard data loaded:', response)

      if (response.data) {
        const data = response.data

        // 更新今日数据
        setTodayStats({
          total_count: data.today_payments || 0,
          total_amount: data.today_amount || 0,
          success_count: Math.floor((data.today_payments || 0) * (data.today_success_rate || 0)),
          failed_count: 0,
          success_rate: data.today_success_rate || 0,
        })

        // 更新本月数据
        setMonthStats({
          total_count: data.month_payments || 0,
          total_amount: data.month_amount || 0,
          success_count: Math.floor((data.month_payments || 0) * (data.month_success_rate || 0)),
          failed_count: 0,
          success_rate: data.month_success_rate || 0,
        })

        // 更新趋势数据
        if (data.payment_trend && data.payment_trend.length > 0) {
          const trendDataFormatted: TrendData[] = []
          data.payment_trend.forEach((item: any) => {
            trendDataFormatted.push({
              date: dayjs(item.date).format('MM-DD'),
              value: item.amount / 100,
              type: t('dashboard.revenueLabel'),
            })
            trendDataFormatted.push({
              date: dayjs(item.date).format('MM-DD'),
              value: item.payments,
              type: t('dashboard.ordersLabel'),
            })
          })
          setTrendData(trendDataFormatted)
        }
      }
    } catch (error) {
      console.error('Failed to load dashboard:', error)
      // 错误已被request.ts的拦截器处理
    } finally {
      setLoading(false)
    }
  }

  const loadStats = async () => {
    try {
      const response = await paymentService.getStats({})
      setStats(response.data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const loadTodayStats = async () => {
    try {
      const today = dayjs()
      const response = await paymentService.getStats({
        start_time: today.startOf('day').toISOString(),
        end_time: today.endOf('day').toISOString(),
      })
      setTodayStats(response.data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const loadMonthStats = async () => {
    try {
      const month = dayjs()
      const response = await paymentService.getStats({
        start_time: month.startOf('month').toISOString(),
        end_time: month.endOf('month').toISOString(),
      })
      setMonthStats(response.data)
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
          type: t('dashboard.revenueLabel'),
        })
        data.push({
          date: dateStr,
          value: response.data.total_count,
          type: t('dashboard.ordersLabel'),
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
                     channel === 'alipay' ? t('dashboard.alipay') : t('dashboard.wechat'),
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
            channel: method === 'card' ? t('dashboard.card') :
                     method === 'bank_transfer' ? t('dashboard.bankTransfer') : t('dashboard.eWallet'),
            value: response.pagination.total,
          })
        }
      }
      setMethodData(data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const getStatusTag = (status: string) => {
    const statusConfig: Record<string, { color: string; text: string }> = {
      pending: { color: 'processing', text: t('transactions.statusPending') },
      success: { color: 'success', text: t('transactions.statusSuccess') },
      failed: { color: 'error', text: t('transactions.statusFailed') },
      cancelled: { color: 'default', text: t('orders.statusCancelled') },
      refunded: { color: 'warning', text: t('transactions.statusRefunded') },
    }
    const config = statusConfig[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const columns: ColumnsType<Payment> = [
    {
      title: t('transactions.orderNo'),
      dataIndex: 'order_id',
      key: 'order_id',
      ellipsis: true,
      width: 200,
    },
    {
      title: t('transactions.amount'),
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number, record) => `${record.currency} ${(amount / 100).toFixed(2)}`,
    },
    {
      title: t('transactions.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: t('common.createdAt'),
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
      <Title level={2}>{t('dashboard.title')}</Title>

      {/* Key Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.todayRevenue')}
              value={todayStats ? todayStats.total_amount / 100 : 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="USD"
              valueStyle={{ color: '#3f8600' }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              <ArrowUpOutlined style={{ color: '#3f8600' }} /> +12.5%
            </Text>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.monthRevenue')}
              value={monthStats ? monthStats.total_amount / 100 : 0}
              precision={2}
              prefix={<RiseOutlined />}
              suffix="USD"
              valueStyle={{ color: '#1890ff' }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              <ArrowUpOutlined style={{ color: '#3f8600' }} /> +8.3%
            </Text>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.todayOrders')}
              value={todayStats?.total_count || 0}
              prefix={<TransactionOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {t('dashboard.totalOrders')}: {stats?.total_count || 0}
            </Text>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('dashboard.successRate')}
              value={stats ? stats.success_rate * 100 : 0}
              precision={2}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {t('dashboard.todayOrders')}: {todayStats?.success_count || 0}
            </Text>
          </Card>
        </Col>
      </Row>

      {/* Account Balance Card */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card>
            <Row>
              <Col xs={24} md={6}>
                <Statistic
                  title={t('dashboard.accountBalance')}
                  value={stats ? stats.total_amount / 100 : 0}
                  precision={2}
                  prefix={<WalletOutlined />}
                  suffix="USD"
                  valueStyle={{ color: '#1890ff', fontSize: 32 }}
                />
              </Col>
              <Col xs={24} md={18}>
                <Space direction="vertical" style={{ width: '100%' }} size="small">
                  <div style={{ marginBottom: 16 }}>
                    <Text type="secondary">{t('dashboard.quickActions')}</Text>
                  </div>
                  <Space wrap>
                    <Button
                      type="primary"
                      icon={<PlusCircleOutlined />}
                      onClick={() => navigate('/create-payment')}
                    >
                      {t('menu.createPayment')}
                    </Button>
                    <Button
                      icon={<SearchOutlined />}
                      onClick={() => navigate('/transactions')}
                    >
                      {t('menu.transactions')}
                    </Button>
                    <Button
                      icon={<RollbackOutlined />}
                      onClick={() => navigate('/refunds')}
                    >
                      {t('menu.refunds')}
                    </Button>
                    <Button onClick={() => navigate('/settlements')}>
                      {t('menu.settlement')}
                    </Button>
                  </Space>
                </Space>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* Charts */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <Card title={t('dashboard.transactionTrend')} loading={loading}>
            {trendData.length > 0 ? (
              <Line {...lineConfig} />
            ) : (
              <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
                {t('common.noData')}
              </div>
            )}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title={t('dashboard.channelDistribution')}>
            {channelData.length > 0 ? (
              <Pie {...pieConfig} />
            ) : (
              <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
                {t('common.noData')}
              </div>
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title={t('dashboard.paymentMethodStats')}>
            {methodData.length > 0 ? (
              <Column {...columnConfig} />
            ) : (
              <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
                {t('common.noData')}
              </div>
            )}
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title={t('dashboard.recentTransactions')}
            extra={
              <Button type="link" onClick={() => navigate('/transactions')}>
                {t('dashboard.viewAll')}
              </Button>
            }
          >
            <Table
              columns={columns}
              dataSource={recentPayments}
              rowKey="id"
              pagination={false}
              size="small"
              locale={{ emptyText: t('common.noData') }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard

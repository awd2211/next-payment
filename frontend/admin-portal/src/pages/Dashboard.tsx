import { useState, useEffect } from 'react'
import {
  Row,
  Col,
  Card,
  Statistic,
  Typography,
  Select,
  Space,
  List,
  Button,
  Tag,
  Divider,
  Avatar,
} from 'antd'
import {
  UserOutlined,
  ShoppingOutlined,
  DollarOutlined,
  SafetyOutlined,
  RiseOutlined,
  FallOutlined,
  ShopOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  ArrowRightOutlined,
} from '@ant-design/icons'
import { Line, Pie, Column } from '@ant-design/charts'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import dayjs from 'dayjs'

const { Title } = Typography

interface Activity {
  id: string
  type: 'payment' | 'merchant' | 'order' | 'risk'
  title: string
  description: string
  timestamp: string
  status: 'success' | 'warning' | 'error' | 'info'
}

interface QuickAction {
  key: string
  title: string
  description: string
  icon: React.ReactNode
  path: string
}

const Dashboard = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [timePeriod, setTimePeriod] = useState<'today' | '7days' | '30days'>('today')
  const [loading, setLoading] = useState(false)

  // 统计数据
  const [stats, setStats] = useState({
    totalAdmins: 25,
    totalMerchants: 156,
    activeMerchants: 142,
    totalTransactions: 1234,
    totalAmount: 5678900,
    successRate: 98.5,
    pendingOrders: 23,
    todayGrowth: 12.5,
  })

  // 交易趋势数据
  const [trendData, setTrendData] = useState<Array<{ date: string; value: number }>>([])

  // 支付渠道分布
  const [channelData, setChannelData] = useState<Array<{ type: string; value: number }>>([])

  // 商户排行
  const [merchantRankData, setMerchantRankData] = useState<
    Array<{ merchant: string; amount: number }>
  >([])

  // 近期活动
  const [recentActivities, setRecentActivities] = useState<Activity[]>([])

  useEffect(() => {
    fetchDashboardData()
  }, [timePeriod])

  const fetchDashboardData = async () => {
    setLoading(true)
    try {
      // Mock data - 根据时间段生成不同的数据
      const days = timePeriod === 'today' ? 24 : timePeriod === '7days' ? 7 : 30
      const isHourly = timePeriod === 'today'

      // 交易趋势
      const trend = Array.from({ length: days }, (_, i) => ({
        date: isHourly
          ? `${i}:00`
          : dayjs()
              .subtract(days - i - 1, 'day')
              .format('MM-DD'),
        value: Math.floor(Math.random() * 100000) + 50000,
      }))
      setTrendData(trend)

      // 支付渠道分布
      setChannelData([
        { type: 'Stripe', value: 45 },
        { type: 'PayPal', value: 30 },
        { type: '支付宝', value: 15 },
        { type: '微信支付', value: 10 },
      ])

      // 商户排行
      setMerchantRankData([
        { merchant: '商户A', amount: 125000 },
        { merchant: '商户B', amount: 98000 },
        { merchant: '商户C', amount: 87000 },
        { merchant: '商户D', amount: 76000 },
        { merchant: '商户E', amount: 65000 },
      ])

      // 近期活动
      setRecentActivities([
        {
          id: '1',
          type: 'payment',
          title: '支付成功',
          description: '商户A 完成支付 ¥12,500.00',
          timestamp: dayjs().subtract(5, 'minute').toISOString(),
          status: 'success',
        },
        {
          id: '2',
          type: 'merchant',
          title: '新商户注册',
          description: '商户XYZ 提交注册申请',
          timestamp: dayjs().subtract(15, 'minute').toISOString(),
          status: 'info',
        },
        {
          id: '3',
          type: 'risk',
          title: '风险预警',
          description: '检测到异常交易行为',
          timestamp: dayjs().subtract(30, 'minute').toISOString(),
          status: 'warning',
        },
        {
          id: '4',
          type: 'order',
          title: '订单取消',
          description: '订单 #ORD123456 已取消',
          timestamp: dayjs().subtract(1, 'hour').toISOString(),
          status: 'error',
        },
        {
          id: '5',
          type: 'payment',
          title: '退款完成',
          description: '订单 #ORD123455 退款 ¥5,600.00',
          timestamp: dayjs().subtract(2, 'hour').toISOString(),
          status: 'info',
        },
      ])
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  const quickActions: QuickAction[] = [
    {
      key: 'merchants',
      title: t('dashboard.quickActions.addMerchant'),
      description: t('dashboard.quickActions.addMerchantDesc'),
      icon: <ShopOutlined style={{ fontSize: 24, color: '#1890ff' }} />,
      path: '/merchants',
    },
    {
      key: 'payments',
      title: t('dashboard.quickActions.viewPayments'),
      description: t('dashboard.quickActions.viewPaymentsDesc'),
      icon: <DollarOutlined style={{ fontSize: 24, color: '#52c41a' }} />,
      path: '/payments',
    },
    {
      key: 'risk',
      title: t('dashboard.quickActions.riskManagement'),
      description: t('dashboard.quickActions.riskManagementDesc'),
      icon: <WarningOutlined style={{ fontSize: 24, color: '#faad14' }} />,
      path: '/risk',
    },
    {
      key: 'config',
      title: t('dashboard.quickActions.systemConfig'),
      description: t('dashboard.quickActions.systemConfigDesc'),
      icon: <SafetyOutlined style={{ fontSize: 24, color: '#722ed1' }} />,
      path: '/system-configs',
    },
  ]

  const getActivityIcon = (type: string, status: string) => {
    const iconMap: Record<string, React.ReactNode> = {
      payment: <DollarOutlined />,
      merchant: <ShopOutlined />,
      order: <ShoppingOutlined />,
      risk: <WarningOutlined />,
    }
    const colors: Record<string, string> = {
      success: '#52c41a',
      warning: '#faad14',
      error: '#ff4d4f',
      info: '#1890ff',
    }
    return <Avatar icon={iconMap[type]} style={{ backgroundColor: colors[status] }} />
  }

  const lineConfig = {
    data: trendData,
    xField: 'date',
    yField: 'value',
    smooth: true,
    color: '#1890ff',
    point: {
      size: 3,
      shape: 'circle',
    },
    label: {
      style: {
        fill: '#aaa',
      },
    },
    yAxis: {
      label: {
        formatter: (v: string) => `¥${(Number(v) / 10000).toFixed(1)}万`,
      },
    },
    tooltip: {
      formatter: (datum: { date: string; value: number }) => ({
        name: t('dashboard.transactionAmount'),
        value: `¥${datum.value.toLocaleString()}`,
      }),
    },
  }

  const pieConfig = {
    data: channelData,
    angleField: 'value',
    colorField: 'type',
    radius: 0.8,
    innerRadius: 0.6,
    label: {
      type: 'inner',
      offset: '-30%',
      content: '{percentage}',
      style: {
        textAlign: 'center',
        fontSize: 14,
      },
    },
    statistic: {
      title: {
        content: t('dashboard.total'),
      },
    },
    legend: {
      position: 'bottom',
    },
  }

  const columnConfig = {
    data: merchantRankData,
    xField: 'merchant',
    yField: 'amount',
    color: '#5B8FF9',
    label: {
      position: 'top',
      formatter: (datum: { amount: number }) => `¥${(datum.amount / 10000).toFixed(1)}万`,
    },
    yAxis: {
      label: {
        formatter: (v: string) => `¥${(Number(v) / 10000).toFixed(0)}万`,
      },
    },
    tooltip: {
      formatter: (datum: { merchant: string; amount: number }) => ({
        name: datum.merchant,
        value: `¥${datum.amount.toLocaleString()}`,
      }),
    },
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2}>{t('dashboard.title')}</Title>
        <Select
          value={timePeriod}
          onChange={setTimePeriod}
          style={{ width: 120 }}
        >
          <Select.Option value="today">{t('dashboard.today')}</Select.Option>
          <Select.Option value="7days">{t('dashboard.last7Days')}</Select.Option>
          <Select.Option value="30days">{t('dashboard.last30Days')}</Select.Option>
        </Select>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.totalTransactions')}
              value={stats.totalTransactions}
              prefix={<ShoppingOutlined />}
              valueStyle={{ color: '#3f8600' }}
              suffix={
                <span style={{ fontSize: 14 }}>
                  <RiseOutlined style={{ color: '#3f8600' }} /> {stats.todayGrowth}%
                </span>
              }
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.totalAmount')}
              value={stats.totalAmount / 100}
              precision={2}
              prefix="¥"
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.successRate')}
              value={stats.successRate}
              precision={1}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.pendingOrders')}
              value={stats.pendingOrders}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.totalMerchants')}
              value={stats.totalMerchants}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.activeMerchants')}
              value={stats.activeMerchants}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#13c2c2' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.totalAdmins')}
              value={stats.totalAdmins}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#eb2f96' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.avgTransactionAmount')}
              value={(stats.totalAmount / stats.totalTransactions / 100).toFixed(2)}
              prefix="¥"
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <Card
            title={t('dashboard.transactionTrend')}
            loading={loading}
            extra={
              <Tag color="blue">
                {timePeriod === 'today'
                  ? t('dashboard.hourly')
                  : timePeriod === '7days'
                  ? t('dashboard.daily')
                  : t('dashboard.daily')}
              </Tag>
            }
          >
            <Line {...lineConfig} height={300} />
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title={t('dashboard.channelDistribution')} loading={loading}>
            <Pie {...pieConfig} height={300} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title={t('dashboard.topMerchants')} loading={loading}>
            <Column {...columnConfig} height={300} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title={t('dashboard.recentActivities')}
            loading={loading}
            extra={
              <Button type="link" onClick={() => navigate('/audit-logs')}>
                {t('dashboard.viewAll')} <ArrowRightOutlined />
              </Button>
            }
          >
            <List
              itemLayout="horizontal"
              dataSource={recentActivities}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={getActivityIcon(item.type, item.status)}
                    title={item.title}
                    description={
                      <Space direction="vertical" size={0}>
                        <span>{item.description}</span>
                        <span style={{ fontSize: 12, color: '#999' }}>
                          {dayjs(item.timestamp).fromNow()}
                        </span>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* 快捷操作 */}
      <Card title={t('dashboard.quickActions')} style={{ marginTop: 16 }} loading={loading}>
        <Row gutter={[16, 16]}>
          {quickActions.map((action) => (
            <Col xs={24} sm={12} lg={6} key={action.key}>
              <Card
                hoverable
                onClick={() => navigate(action.path)}
                style={{ textAlign: 'center' }}
              >
                <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                  {action.icon}
                  <div>
                    <div style={{ fontWeight: 'bold', marginBottom: 8 }}>{action.title}</div>
                    <div style={{ fontSize: 12, color: '#999' }}>{action.description}</div>
                  </div>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>
    </div>
  )
}

export default Dashboard

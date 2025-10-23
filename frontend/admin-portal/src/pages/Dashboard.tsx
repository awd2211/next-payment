import { useState } from 'react'
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
  Avatar,
  message,
} from 'antd'
import {
  UserOutlined,
  ShoppingOutlined,
  DollarOutlined,
  SafetyOutlined,
  RiseOutlined,
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
import { useRequest } from '@payment/shared/hooks'
import { getDashboardData } from '../services/dashboard'

const { Title } = Typography

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

  // 使用 useRequest Hook 获取 Dashboard 数据
  const { data: dashboardData, loading } = useRequest(
    () => getDashboardData(timePeriod),
    {
      onError: (err) => {
        console.error('Failed to fetch dashboard data:', err)
        message.error(t('dashboard.fetchError') || '获取Dashboard数据失败')
      },
    },
  )

  // 从API响应中提取数据，如果没有数据则使用默认值
  const stats = dashboardData?.stats || {
    total_merchants: 0,
    active_merchants: 0,
    total_transactions: 0,
    total_amount: 0,
    success_rate: 0,
    pending_orders: 0,
    today_growth: 0,
    total_admins: 0,
  }

  const trendData = dashboardData?.trend_data || []
  const channelData = dashboardData?.channel_distribution || []
  const merchantRankData = dashboardData?.merchant_ranks || []
  const recentActivities = dashboardData?.recent_activities || []

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
              value={stats.total_transactions}
              prefix={<ShoppingOutlined />}
              valueStyle={{ color: '#3f8600' }}
              suffix={
                <span style={{ fontSize: 14 }}>
                  <RiseOutlined style={{ color: '#3f8600' }} /> {stats.today_growth}%
                </span>
              }
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.totalAmount')}
              value={stats.total_amount / 100}
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
              value={stats.success_rate}
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
              value={stats.pending_orders}
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
              value={stats.total_merchants}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.activeMerchants')}
              value={stats.active_merchants}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#13c2c2' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.totalAdmins')}
              value={stats.total_admins}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#eb2f96' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic
              title={t('dashboard.avgTransactionAmount')}
              value={stats.total_transactions > 0
                ? (stats.total_amount / stats.total_transactions / 100).toFixed(2)
                : '0.00'}
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

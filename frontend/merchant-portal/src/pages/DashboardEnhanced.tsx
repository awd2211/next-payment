import { useState, useEffect } from 'react'
import {
  Row,
  Col,
  Card,
  Statistic,
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Spin,
  Alert,
  DatePicker,
  Select,
} from 'antd'
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
  WarningOutlined,
  BellOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { Line, Pie, Column } from '@ant-design/charts'
import { dashboardService } from '../services/dashboardService'
import { useAuthStore } from '../stores/authStore'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import dayjs from 'dayjs'

const { Title, Text } = Typography
const { RangePicker } = DatePicker

// 类型定义
interface DashboardData {
  today_payments: number
  today_amount: number
  today_success_rate: number
  month_payments: number
  month_amount: number
  month_success_rate: number
  available_balance: number
  frozen_balance: number
  pending_settlement: number
  risk_level: string
  pending_reviews: number
  pending_withdrawals: number
  unread_notifications: number
  payment_trend: DailyData[]
}

interface DailyData {
  date: string
  payments: number
  amount: number
  success_rate: number
}

const DashboardEnhanced = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    const token = useAuthStore.getState().token
    if (!token) {
      setError('请先登录')
      setLoading(false)
      return
    }

    try {
      setError(null)
      const response = await dashboardService.getDashboard()
      setDashboardData(response.data)
    } catch (error: any) {
      console.error('Failed to load dashboard:', error)
      setError(error.response?.data?.message || '加载失败')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  const handleRefresh = () => {
    setRefreshing(true)
    loadDashboardData()
  }

  // 格式化金额
  const formatAmount = (amount: number) => {
    return (amount / 100).toLocaleString('en-US', {
      style: 'currency',
      currency: 'USD',
    })
  }

  // 格式化百分比
  const formatPercent = (rate: number) => {
    return `${(rate * 100).toFixed(2)}%`
  }

  // 趋势图数据转换
  const getTrendChartData = () => {
    if (!dashboardData?.payment_trend) return []
    return dashboardData.payment_trend.map(item => ({
      date: item.date,
      金额: item.amount / 100,
      笔数: item.payments,
    }))
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  if (error) {
    return (
      <Alert
        message="加载失败"
        description={error}
        type="error"
        showIcon
        action={
          <Button size="small" onClick={handleRefresh}>
            重试
          </Button>
        }
      />
    )
  }

  const riskLevelColor = (level: string) => {
    switch (level?.toLowerCase()) {
      case 'low':
        return 'success'
      case 'medium':
        return 'warning'
      case 'high':
        return 'error'
      default:
        return 'default'
    }
  }

  return (
    <div style={{ padding: '24px', background: '#f0f2f5' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2} style={{ margin: 0 }}>
          {t('dashboard.title') || '数据概览'}
        </Title>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            loading={refreshing}
            onClick={handleRefresh}
          >
            刷新
          </Button>
        </Space>
      </div>

      {/* 风险提示 */}
      {dashboardData?.risk_level && dashboardData.risk_level !== 'low' && (
        <Alert
          message={`风险等级: ${dashboardData.risk_level.toUpperCase()}`}
          description={`您有 ${dashboardData.pending_reviews} 笔交易待审核`}
          type="warning"
          showIcon
          icon={<WarningOutlined />}
          closable
          style={{ marginBottom: 16 }}
        />
      )}

      {/* 今日数据卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日收入"
              value={dashboardData?.today_amount || 0}
              precision={2}
              prefix={<DollarOutlined />}
              valueStyle={{ color: '#3f8600' }}
              formatter={(value) => formatAmount(Number(value))}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {dashboardData?.today_payments || 0} 笔交易
            </Text>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="本月收入"
              value={dashboardData?.month_amount || 0}
              precision={2}
              prefix={<RiseOutlined />}
              valueStyle={{ color: '#1890ff' }}
              formatter={(value) => formatAmount(Number(value))}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {dashboardData?.month_payments || 0} 笔交易
            </Text>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="成功率"
              value={dashboardData?.today_success_rate || 0}
              precision={2}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{
                color: (dashboardData?.today_success_rate || 0) >= 0.95 ? '#3f8600' : '#cf1322',
              }}
              formatter={(value) => (Number(value) * 100).toFixed(2)}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              今日数据
            </Text>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="可用余额"
              value={dashboardData?.available_balance || 0}
              precision={2}
              prefix={<WalletOutlined />}
              valueStyle={{ color: '#722ed1' }}
              formatter={(value) => formatAmount(Number(value))}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              冻结: {formatAmount(dashboardData?.frozen_balance || 0)}
            </Text>
          </Card>
        </Col>
      </Row>

      {/* 待处理事项 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            onClick={() => navigate('/orders')}
            style={{ cursor: 'pointer' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text type="secondary">待审核交易</Text>
                <WarningOutlined style={{ fontSize: 20, color: '#faad14' }} />
              </div>
              <Title level={2} style={{ margin: 0 }}>
                {dashboardData?.pending_reviews || 0}
              </Title>
            </Space>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            onClick={() => navigate('/settlements')}
            style={{ cursor: 'pointer' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text type="secondary">待结算金额</Text>
                <TransactionOutlined style={{ fontSize: 20, color: '#1890ff' }} />
              </div>
              <Title level={2} style={{ margin: 0 }}>
                {formatAmount(dashboardData?.pending_settlement || 0)}
              </Title>
            </Space>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            onClick={() => navigate('/withdrawals')}
            style={{ cursor: 'pointer' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text type="secondary">待处理提现</Text>
                <RollbackOutlined style={{ fontSize: 20, color: '#52c41a' }} />
              </div>
              <Title level={2} style={{ margin: 0 }}>
                {dashboardData?.pending_withdrawals || 0}
              </Title>
            </Space>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            onClick={() => navigate('/notifications')}
            style={{ cursor: 'pointer' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text type="secondary">未读通知</Text>
                <BellOutlined style={{ fontSize: 20, color: '#722ed1' }} />
              </div>
              <Title level={2} style={{ margin: 0 }}>
                {dashboardData?.unread_notifications || 0}
              </Title>
            </Space>
          </Card>
        </Col>
      </Row>

      {/* 交易趋势图 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="近7天交易趋势" bordered={false}>
            {dashboardData?.payment_trend && dashboardData.payment_trend.length > 0 ? (
              <Line
                data={getTrendChartData()}
                xField="date"
                yField="金额"
                seriesField="type"
                smooth
                animation={{
                  appear: {
                    animation: 'path-in',
                    duration: 1000,
                  },
                }}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
                暂无数据
              </div>
            )}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="风险状态" bordered={false}>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <div>
                <Text type="secondary">风险等级</Text>
                <div style={{ marginTop: 8 }}>
                  <Tag color={riskLevelColor(dashboardData?.risk_level || 'low')} style={{ fontSize: 16, padding: '4px 12px' }}>
                    {dashboardData?.risk_level?.toUpperCase() || 'LOW'}
                  </Tag>
                </div>
              </div>
              <div>
                <Text type="secondary">待审核交易</Text>
                <div style={{ marginTop: 8, fontSize: 24, fontWeight: 500 }}>
                  {dashboardData?.pending_reviews || 0} 笔
                </div>
              </div>
              <Button
                type="primary"
                block
                onClick={() => navigate('/orders?status=pending')}
              >
                查看待审核交易
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>

      {/* 快捷操作 */}
      <Card title="快捷操作" style={{ marginTop: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={12} sm={6}>
            <Button
              type="primary"
              icon={<PlusCircleOutlined />}
              block
              size="large"
              onClick={() => navigate('/create-payment')}
            >
              创建支付
            </Button>
          </Col>
          <Col xs={12} sm={6}>
            <Button
              icon={<SearchOutlined />}
              block
              size="large"
              onClick={() => navigate('/transactions')}
            >
              查看交易
            </Button>
          </Col>
          <Col xs={12} sm={6}>
            <Button
              icon={<RollbackOutlined />}
              block
              size="large"
              onClick={() => navigate('/refunds')}
            >
              退款管理
            </Button>
          </Col>
          <Col xs={12} sm={6}>
            <Button
              icon={<TransactionOutlined />}
              block
              size="large"
              onClick={() => navigate('/settlements')}
            >
              结算记录
            </Button>
          </Col>
        </Row>
      </Card>
    </div>
  )
}

export default DashboardEnhanced

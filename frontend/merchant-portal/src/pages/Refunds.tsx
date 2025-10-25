import { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Select,
  DatePicker,
  Tag,
  Modal,
  Form,
  InputNumber,
  message,
  Descriptions,
  Statistic,
  Row,
  Col,
  Alert,
  Tooltip,
  Badge,
  Skeleton,
  Typography,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  RollbackOutlined,
  FilterOutlined,
  ClearOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker
const { Title } = Typography

interface Refund {
  refund_no: string
  payment_no: string
  order_no: string
  amount: number
  refund_amount: number
  currency: string
  reason: string
  status: string
  channel: string
  refund_time: string
  created_at: string
}

interface RefundStats {
  total_refunds: number
  total_amount: number
  success_count: number
  success_amount: number
  pending_count: number
  failed_count: number
}

const Refunds = () => {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [statsLoading, setStatsLoading] = useState(false)
  const [dataSource, setDataSource] = useState<Refund[]>([])
  const [selectedRefund, setSelectedRefund] = useState<Refund | null>(null)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [form] = Form.useForm()
  const [stats, setStats] = useState<RefundStats>({
    total_refunds: 0,
    total_amount: 0,
    success_count: 0,
    success_amount: 0,
    pending_count: 0,
    failed_count: 0,
  })
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })
  const [searchFilters, setSearchFilters] = useState({
    refund_no: '',
    payment_no: '',
    status: '',
    date_range: null as [dayjs.Dayjs, dayjs.Dayjs] | null,
  })

  // 计算激活的过滤器数量
  const activeFilterCount = [
    searchFilters.refund_no,
    searchFilters.payment_no,
    searchFilters.status,
    searchFilters.date_range,
  ].filter(Boolean).length

  useEffect(() => {
    fetchRefunds()
    fetchStats()
  }, [pagination.current, pagination.pageSize])

  const fetchStats = async () => {
    setStatsLoading(true)
    try {
      // Mock stats data
      setStats({
        total_refunds: 156,
        total_amount: 234567.89,
        success_count: 142,
        success_amount: 218900.45,
        pending_count: 8,
        failed_count: 6,
      })
    } finally {
      setStatsLoading(false)
    }
  }

  const handleClearFilters = () => {
    setSearchFilters({
      refund_no: '',
      payment_no: '',
      status: '',
      date_range: null,
    })
    setPagination((prev) => ({ ...prev, current: 1 }))
    fetchRefunds()
  }

  const fetchRefunds = async () => {
    setLoading(true)
    try {
      // Mock data
      const mockData: Refund[] = Array.from({ length: 10 }, (_, i) => ({
        refund_no: `REF${Date.now() + i}`,
        payment_no: `PAY${Date.now() - i * 1000}`,
        order_no: `ORD${Date.now() - i * 2000}`,
        amount: Math.floor(Math.random() * 10000) + 1000,
        refund_amount: Math.floor(Math.random() * 10000) + 1000,
        currency: 'CNY',
        reason: ['客户要求退款', '商品质量问题', '重复支付', '订单取消'][Math.floor(Math.random() * 4)],
        status: ['pending', 'success', 'failed'][Math.floor(Math.random() * 3)],
        channel: ['stripe', 'paypal', 'alipay'][Math.floor(Math.random() * 3)],
        refund_time: dayjs().subtract(i, 'day').toISOString(),
        created_at: dayjs().subtract(i, 'day').subtract(1, 'hour').toISOString(),
      }))

      setDataSource(mockData)
      setPagination((prev) => ({ ...prev, total: 100 }))
    } catch (error) {
      message.error(t('refunds.fetchFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = () => {
    setPagination((prev) => ({ ...prev, current: 1 }))
    fetchRefunds()
  }

  const handleReset = () => {
    setSearchFilters({
      refund_no: '',
      payment_no: '',
      status: '',
      date_range: null,
    })
    setPagination((prev) => ({ ...prev, current: 1 }))
    fetchRefunds()
  }

  const handleViewDetail = (record: Refund) => {
    setSelectedRefund(record)
    setDetailModalVisible(true)
  }

  const handleCreateRefund = async (values: any) => {
    try {
      setLoading(true)
      // Mock API call
      await new Promise((resolve) => setTimeout(resolve, 1500))

      message.success(t('refunds.createSuccess'))
      setCreateModalVisible(false)
      form.resetFields()
      fetchRefunds()
      fetchStats()
    } catch (error) {
      message.error(t('refunds.createFailed'))
    } finally {
      setLoading(false)
    }
  }

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${(amount / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; icon: React.ReactNode; text: string }> = {
      pending: {
        color: 'processing',
        icon: <ClockCircleOutlined />,
        text: t('refunds.statusPending')
      },
      success: {
        color: 'success',
        icon: <CheckCircleOutlined />,
        text: t('refunds.statusSuccess')
      },
      failed: {
        color: 'error',
        icon: <CloseCircleOutlined />,
        text: t('refunds.statusFailed')
      },
    }
    const config = statusMap[status] || { color: 'default', icon: null, text: status }
    return (
      <Tag color={config.color} icon={config.icon} style={{ borderRadius: 12 }}>
        {config.text}
      </Tag>
    )
  }

  const getChannelTag = (channel: string) => {
    const channelMap: Record<string, { color: string; text: string }> = {
      stripe: { color: 'blue', text: 'Stripe' },
      paypal: { color: 'cyan', text: 'PayPal' },
      alipay: { color: 'green', text: '支付宝' },
      wechat: { color: 'orange', text: '微信支付' },
    }
    const config = channelMap[channel] || { color: 'default', text: channel }
    return <Tag color={config.color} style={{ borderRadius: 12 }}>{config.text}</Tag>
  }

  const columns: ColumnsType<Refund> = [
    {
      title: t('refunds.refundNo'),
      dataIndex: 'refund_no',
      key: 'refund_no',
      fixed: 'left',
      width: 180,
      render: (id: string) => (
        <Tooltip title={id}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {id.slice(0, 12)}...
          </span>
        </Tooltip>
      ),
    },
    {
      title: t('refunds.paymentNo'),
      dataIndex: 'payment_no',
      key: 'payment_no',
      width: 180,
      render: (id: string) => (
        <Tooltip title={id}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {id.slice(0, 12)}...
          </span>
        </Tooltip>
      ),
    },
    {
      title: t('refunds.orderNo'),
      dataIndex: 'order_no',
      key: 'order_no',
      width: 180,
      render: (id: string) => (
        <Tooltip title={id}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {id.slice(0, 12)}...
          </span>
        </Tooltip>
      ),
    },
    {
      title: t('refunds.originalAmount'),
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      align: 'right',
      render: (amount, record) => formatAmount(amount, record.currency),
    },
    {
      title: t('refunds.refundAmount'),
      dataIndex: 'refund_amount',
      key: 'refund_amount',
      width: 120,
      align: 'right',
      render: (amount, record) => (
        <span style={{ color: '#ff4d4f', fontWeight: 600 }}>
          {formatAmount(amount, record.currency)}
        </span>
      ),
    },
    {
      title: t('refunds.channel'),
      dataIndex: 'channel',
      key: 'channel',
      width: 120,
      render: (channel) => getChannelTag(channel),
    },
    {
      title: t('refunds.reason'),
      dataIndex: 'reason',
      key: 'reason',
      width: 150,
      render: (text: string) => (
        <Tooltip title={text}>
          <span style={{
            display: 'inline-block',
            maxWidth: '140px',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap'
          }}>
            {text}
          </span>
        </Tooltip>
      ),
    },
    {
      title: t('refunds.status'),
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => getStatusTag(status),
    },
    {
      title: t('common.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date) => (
        <Tooltip title={dayjs(date).format('YYYY-MM-DD HH:mm:ss')}>
          {dayjs(date).format('MM-DD HH:mm')}
        </Tooltip>
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      fixed: 'right',
      width: 120,
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record)}
        >
          {t('refunds.viewDetail')}
        </Button>
      ),
    },
  ]

  return (
    <div>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>{t('refunds.title')}</Title>
        <Space>
          <Tooltip title={t('common.refresh')}>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                fetchRefunds()
                fetchStats()
              }}
              loading={loading || statsLoading}
              style={{ borderRadius: 8 }}
            >
              {t('common.refresh')}
            </Button>
          </Tooltip>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
            style={{ borderRadius: 8 }}
          >
            {t('refunds.createRefund')}
          </Button>
        </Space>
      </div>

      {/* Statistics Cards with Skeleton */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable style={{ borderRadius: 12, transition: 'all 0.3s ease', cursor: 'default' }}>
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>{t('refunds.totalRefunds')}</span>}
                value={stats.total_refunds}
                prefix={<RollbackOutlined />}
                valueStyle={{ fontSize: 24, fontWeight: 600 }}
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable style={{ borderRadius: 12, transition: 'all 0.3s ease', cursor: 'default' }}>
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>{t('refunds.totalAmount')}</span>}
                value={stats.total_amount}
                precision={2}
                prefix="¥"
                valueStyle={{ fontSize: 24, fontWeight: 600, color: '#ff4d4f' }}
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable style={{ borderRadius: 12, transition: 'all 0.3s ease', cursor: 'default' }}>
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>{t('refunds.successCount')}</span>}
                value={stats.success_count}
                prefix={<CheckCircleOutlined />}
                valueStyle={{ fontSize: 24, fontWeight: 600, color: '#52c41a' }}
                suffix={
                  <span style={{ fontSize: 14 }}>
                    / ¥{(stats.success_amount / 10000).toFixed(1)}万
                  </span>
                }
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable style={{ borderRadius: 12, transition: 'all 0.3s ease', cursor: 'default' }}>
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>{t('refunds.pendingCount')}</span>}
                value={stats.pending_count}
                prefix={<ClockCircleOutlined />}
                valueStyle={{ fontSize: 24, fontWeight: 600, color: '#faad14' }}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* Smart Filters */}
      <Card style={{ marginBottom: 16, borderRadius: 12 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Space>
              <FilterOutlined />
              <span style={{ fontWeight: 500 }}>{t('common.filter')}</span>
              {activeFilterCount > 0 && (
                <Badge count={activeFilterCount} style={{ backgroundColor: '#1890ff' }} />
              )}
            </Space>
            {activeFilterCount > 0 && (
              <Button
                icon={<ClearOutlined />}
                onClick={handleClearFilters}
                style={{ borderRadius: 8 }}
              >
                清除筛选
              </Button>
            )}
          </div>

          <Space wrap>
            <Input
              placeholder={t('refunds.refundNo')}
              value={searchFilters.refund_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, refund_no: e.target.value })
              }
              style={{ width: 200, borderRadius: 8 }}
              prefix={<SearchOutlined />}
            />
            <Input
              placeholder={t('refunds.paymentNo')}
              value={searchFilters.payment_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, payment_no: e.target.value })
              }
              style={{ width: 200, borderRadius: 8 }}
              prefix={<SearchOutlined />}
            />
            <Select
              placeholder={t('refunds.status')}
              value={searchFilters.status}
              onChange={(value) => setSearchFilters({ ...searchFilters, status: value })}
              style={{ width: 150, borderRadius: 8 }}
              allowClear
            >
              <Select.Option value="pending">{t('refunds.statusPending')}</Select.Option>
              <Select.Option value="success">{t('refunds.statusSuccess')}</Select.Option>
              <Select.Option value="failed">{t('refunds.statusFailed')}</Select.Option>
            </Select>
            <RangePicker
              value={searchFilters.date_range}
              onChange={(dates) =>
                setSearchFilters({ ...searchFilters, date_range: dates as [dayjs.Dayjs, dayjs.Dayjs] | null })
              }
              style={{ borderRadius: 8 }}
            />
            <Button
              type="primary"
              icon={<SearchOutlined />}
              onClick={handleSearch}
              style={{ borderRadius: 8 }}
            >
              {t('common.search')}
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={handleReset}
              style={{ borderRadius: 8 }}
            >
              {t('common.reset')}
            </Button>
          </Space>

          <Alert
            message={t('refunds.notice')}
            description={t('refunds.noticeDesc')}
            type="info"
            showIcon
            closable
            style={{ borderRadius: 8 }}
          />
        </Space>
      </Card>

      {/* Table */}
      <Card style={{ borderRadius: 12 }}>
        <Table
          columns={columns}
          dataSource={dataSource}
          rowKey="refund_no"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => t('common.total', { count: total }),
            onChange: (page, pageSize) => {
              setPagination({ ...pagination, current: page, pageSize })
            },
          }}
          scroll={{ x: 1600 }}
        />
      </Card>

      {/* Detail Modal */}
      <Modal
        title={t('refunds.refundDetail')}
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            {t('common.cancel')}
          </Button>,
        ]}
        width={700}
      >
        {selectedRefund && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label={t('refunds.refundNo')} span={2}>
              {selectedRefund.refund_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('refunds.paymentNo')} span={2}>
              {selectedRefund.payment_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('refunds.orderNo')} span={2}>
              {selectedRefund.order_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('refunds.originalAmount')}>
              {formatAmount(selectedRefund.amount, selectedRefund.currency)}
            </Descriptions.Item>
            <Descriptions.Item label={t('refunds.refundAmount')}>
              <span style={{ color: '#ff4d4f', fontWeight: 'bold', fontSize: 16 }}>
                {formatAmount(selectedRefund.refund_amount, selectedRefund.currency)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label={t('refunds.channel')}>
              {getChannelTag(selectedRefund.channel)}
            </Descriptions.Item>
            <Descriptions.Item label={t('refunds.status')}>
              {getStatusTag(selectedRefund.status)}
            </Descriptions.Item>
            <Descriptions.Item label={t('refunds.reason')} span={2}>
              {selectedRefund.reason}
            </Descriptions.Item>
            {selectedRefund.status === 'success' && (
              <Descriptions.Item label={t('refunds.refundTime')} span={2}>
                {dayjs(selectedRefund.refund_time).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
            )}
            <Descriptions.Item label={t('common.createdAt')} span={2}>
              {dayjs(selectedRefund.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* Create Refund Modal */}
      <Modal
        title={t('refunds.createRefund')}
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        confirmLoading={loading}
        width={600}
      >
        <Alert
          message={t('refunds.createNotice')}
          description={t('refunds.createNoticeDesc')}
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateRefund}
        >
          <Form.Item
            label={t('refunds.paymentNo')}
            name="payment_no"
            rules={[{ required: true, message: t('refunds.paymentNoRequired') }]}
          >
            <Input placeholder={t('refunds.paymentNoPlaceholder')} />
          </Form.Item>

          <Form.Item
            label={t('refunds.refundAmount')}
            name="refund_amount"
            rules={[
              { required: true, message: t('refunds.refundAmountRequired') },
              { type: 'number', min: 0.01, message: t('refunds.refundAmountMin') },
            ]}
            tooltip={t('refunds.refundAmountTooltip')}
          >
            <InputNumber
              style={{ width: '100%' }}
              placeholder="0.00"
              precision={2}
              min={0.01}
              prefix={<DollarOutlined />}
            />
          </Form.Item>

          <Form.Item
            label={t('refunds.reason')}
            name="reason"
            rules={[{ required: true, message: t('refunds.reasonRequired') }]}
          >
            <Select placeholder={t('refunds.reasonPlaceholder')}>
              <Select.Option value="customer_request">{t('refunds.reasonCustomerRequest')}</Select.Option>
              <Select.Option value="quality_issue">{t('refunds.reasonQualityIssue')}</Select.Option>
              <Select.Option value="duplicate_payment">{t('refunds.reasonDuplicatePayment')}</Select.Option>
              <Select.Option value="order_cancelled">{t('refunds.reasonOrderCancelled')}</Select.Option>
              <Select.Option value="other">{t('refunds.reasonOther')}</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label={t('refunds.reasonDetail')}
            name="reason_detail"
          >
            <Input.TextArea
              placeholder={t('refunds.reasonDetailPlaceholder')}
              rows={3}
              maxLength={200}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Refunds

import { useState, useEffect } from 'react'
import {
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Card,
  Row,
  Col,
  Statistic,
  Select,
  Input,
  DatePicker,
  Drawer,
  Descriptions,
  Modal,
  Form,
  InputNumber,
  message,
  Popconfirm,
  Tooltip,
  Badge,
  Skeleton,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  EyeOutlined,
  DollarOutlined,
  TransactionOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  UndoOutlined,
  DownloadOutlined,
  ReloadOutlined,
  FilterOutlined,
  ClearOutlined,
} from '@ant-design/icons'
import { paymentService, Payment, PaymentStats } from '../services/paymentService'
import { useAuthStore } from '../stores/authStore'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'

const { Title } = Typography
const { RangePicker } = DatePicker

const Transactions = () => {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [statsLoading, setStatsLoading] = useState(false)
  const [payments, setPayments] = useState<Payment[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [stats, setStats] = useState<PaymentStats | null>(null)
  const [selectedPayment, setSelectedPayment] = useState<Payment | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false)
  const [refundModalVisible, setRefundModalVisible] = useState(false)
  const [refundForm] = Form.useForm()
  const [filterVisible, setFilterVisible] = useState(false)

  // Filter states
  const [orderIdFilter, setOrderIdFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [channelFilter, setChannelFilter] = useState<string | undefined>()
  const [methodFilter, setMethodFilter] = useState<string | undefined>()
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)

  // 计算激活的过滤器数量
  const activeFilterCount = [
    orderIdFilter,
    statusFilter,
    channelFilter,
    methodFilter,
    dateRange,
  ].filter(Boolean).length

  useEffect(() => {
    loadPayments()
  }, [page, pageSize, orderIdFilter, statusFilter, channelFilter, methodFilter, dateRange])

  useEffect(() => {
    loadStats()
  }, [])

  const loadPayments = async () => {
    const token = useAuthStore.getState().token
    if (!token) {
      console.log('No token found, skipping payments load')
      return
    }

    setLoading(true)
    try {
      const response = await paymentService.list({
        page,
        page_size: pageSize,
        order_id: orderIdFilter || undefined,
        status: statusFilter,
        channel: channelFilter,
        method: methodFilter,
        start_time: dateRange?.[0]?.toISOString(),
        end_time: dateRange?.[1]?.toISOString(),
      })
      // 修复：response.data 包含 list 和 total
      setPayments(response.data?.list || [])
      setTotal(response.data?.total || 0)
    } catch (error) {
      // Error handled by interceptor
      console.error('Failed to load payments:', error)
      setPayments([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  const loadStats = async () => {
    const token = useAuthStore.getState().token
    if (!token) {
      console.log('No token found, skipping stats load')
      return
    }

    setStatsLoading(true)
    try {
      const response = await paymentService.getStats({})
      if (response.data) {
        setStats(response.data)
      }
    } catch (error) {
      // Stats API 可能不存在，暂时忽略错误
      console.log('Stats API not available yet')
    } finally {
      setStatsLoading(false)
    }
  }

  const handleClearFilters = () => {
    setOrderIdFilter('')
    setStatusFilter(undefined)
    setChannelFilter(undefined)
    setMethodFilter(undefined)
    setDateRange(null)
    setPage(1)
  }

  const handleViewDetail = (payment: Payment) => {
    setSelectedPayment(payment)
    setDetailDrawerVisible(true)
  }

  const handleRefund = (payment: Payment) => {
    setSelectedPayment(payment)
    refundForm.setFieldsValue({
      amount: payment.amount,
      reason: '',
    })
    setRefundModalVisible(true)
  }

  const handleRefundSubmit = async () => {
    if (!selectedPayment) return

    try {
      const values = await refundForm.validateFields()
      await paymentService.refund(selectedPayment.id, values)
      message.success('退款申请已提交')
      setRefundModalVisible(false)
      loadPayments()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleCancel = async (payment: Payment) => {
    try {
      await paymentService.cancel(payment.id, '商户取消')
      message.success('交易已取消')
      loadPayments()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleExport = async () => {
    try {
      const response = await paymentService.export({
        order_id: orderIdFilter || undefined,
        status: statusFilter,
        channel: channelFilter,
        method: methodFilter,
        start_time: dateRange?.[0]?.toISOString(),
        end_time: dateRange?.[1]?.toISOString(),
      })

      const blob = new Blob([response as any], { type: 'text/csv;charset=utf-8;' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `transactions_${dayjs().format('YYYYMMDD_HHmmss')}.csv`)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)

      message.success('导出成功')
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const resetFilters = () => {
    setOrderIdFilter('')
    setStatusFilter(undefined)
    setChannelFilter(undefined)
    setMethodFilter(undefined)
    setDateRange(null)
    setPage(1)
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'processing',
      success: 'success',
      failed: 'error',
      cancelled: 'default',
      refunded: 'warning',
    }
    return colors[status] || 'default'
  }

  const getStatusText = (status: string) => {
    const texts: Record<string, string> = {
      pending: t('transactions.statusPending'),
      success: t('transactions.statusSuccess'),
      failed: t('transactions.statusFailed'),
      cancelled: t('orders.statusCancelled'),
      refunded: t('transactions.statusRefunded'),
    }
    return texts[status] || status
  }

  const columns: ColumnsType<Payment> = [
    {
      title: t('transactions.transactionNo'),
      dataIndex: 'id',
      key: 'id',
      width: 120,
      ellipsis: true,
      render: (id: string) => (
        <Tooltip title={id}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {id.slice(0, 8)}...
          </span>
        </Tooltip>
      ),
    },
    {
      title: t('transactions.orderNo'),
      dataIndex: 'order_id',
      key: 'order_id',
      width: 120,
      ellipsis: true,
      render: (orderId: string) => (
        <Tooltip title={orderId}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {orderId?.slice(0, 8)}...
          </span>
        </Tooltip>
      ),
    },
    {
      title: t('transactions.amount'),
      dataIndex: 'amount',
      key: 'amount',
      width: 140,
      render: (amount: number, record) => (
        <span style={{ fontWeight: 600, color: '#1890ff', fontSize: 14 }}>
          {record.currency} {(amount / 100).toFixed(2)}
        </span>
      ),
    },
    {
      title: t('transactions.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)} style={{ borderRadius: 12, fontWeight: 500 }}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: t('transactions.channel'),
      dataIndex: 'channel',
      key: 'channel',
      width: 100,
      render: (channel: string) => (
        <Tag style={{ borderRadius: 12 }}>{channel?.toUpperCase()}</Tag>
      ),
    },
    {
      title: '客户邮箱',
      dataIndex: 'customer_email',
      key: 'customer_email',
      ellipsis: true,
      width: 180,
    },
    {
      title: t('common.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => (
        <Tooltip title={dayjs(time).format('YYYY-MM-DD HH:mm:ss')}>
          {dayjs(time).format('MM-DD HH:mm')}
        </Tooltip>
      ),
    },
    {
      title: t('common.actions'),
      key: 'action',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title={t('transactions.viewDetails')}>
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
              style={{ padding: '4px 8px' }}
            >
              {t('common.detail')}
            </Button>
          </Tooltip>
          {record.status === 'success' && (
            <Tooltip title={t('transactions.refund')}>
              <Button
                type="link"
                size="small"
                icon={<UndoOutlined />}
                onClick={() => handleRefund(record)}
                style={{ padding: '4px 8px' }}
              >
                {t('transactions.refund')}
              </Button>
            </Tooltip>
          )}
          {record.status === 'pending' && (
            <Popconfirm
              title={t('common.confirm')}
              description="确定要取消这笔交易吗？"
              onConfirm={() => handleCancel(record)}
              okText={t('common.yes')}
              cancelText={t('common.no')}
            >
              <Button type="link" size="small" danger style={{ padding: '4px 8px' }}>
                {t('common.cancel')}
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>{t('transactions.title')}</Title>
        <Space>
          <Tooltip title={t('common.refresh')}>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                loadPayments()
                loadStats()
              }}
              loading={loading || statsLoading}
              style={{ borderRadius: 8 }}
            >
              {t('common.refresh')}
            </Button>
          </Tooltip>
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            onClick={handleExport}
            loading={loading}
            style={{ borderRadius: 8 }}
          >
            {t('common.export')}
          </Button>
        </Space>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              borderRadius: 12,
              transition: 'all 0.3s ease',
              cursor: 'default',
            }}
          >
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>总交易额</span>}
                value={stats ? (stats.total_amount / 100).toFixed(2) : 0}
                prefix={<DollarOutlined style={{ color: '#1890ff' }} />}
                suffix="USD"
                valueStyle={{ color: '#1890ff', fontSize: 24, fontWeight: 600 }}
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              borderRadius: 12,
              transition: 'all 0.3s ease',
              cursor: 'default',
            }}
          >
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>交易笔数</span>}
                value={stats?.total_count || 0}
                prefix={<TransactionOutlined style={{ color: '#fa8c16' }} />}
                valueStyle={{ color: '#fa8c16', fontSize: 24, fontWeight: 600 }}
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              borderRadius: 12,
              transition: 'all 0.3s ease',
              cursor: 'default',
            }}
          >
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>成功率</span>}
                value={stats ? (stats.success_rate * 100).toFixed(2) : 0}
                prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
                suffix="%"
                valueStyle={{ color: '#52c41a', fontSize: 24, fontWeight: 600 }}
              />
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card
            hoverable
            style={{
              borderRadius: 12,
              transition: 'all 0.3s ease',
              cursor: 'default',
            }}
          >
            {statsLoading ? (
              <Skeleton active paragraph={{ rows: 1 }} />
            ) : (
              <Statistic
                title={<span style={{ fontSize: 14, fontWeight: 500 }}>今日交易</span>}
                value={stats?.today_amount ? (stats.today_amount / 100).toFixed(2) : 0}
                prefix={<ClockCircleOutlined style={{ color: '#722ed1' }} />}
                suffix="USD"
                valueStyle={{ color: '#722ed1', fontSize: 24, fontWeight: 600 }}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* Filters */}
      <Card
        style={{
          marginBottom: 16,
          borderRadius: 12,
        }}
        bodyStyle={{ padding: '20px' }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <Space align="center">
            <FilterOutlined style={{ fontSize: 16 }} />
            <span style={{ fontWeight: 500, fontSize: 14 }}>{t('common.filter')}</span>
            {activeFilterCount > 0 && (
              <Badge count={activeFilterCount} style={{ backgroundColor: '#1890ff' }} />
            )}
          </Space>
          {activeFilterCount > 0 && (
            <Button
              type="link"
              icon={<ClearOutlined />}
              onClick={handleClearFilters}
              size="small"
            >
              清除筛选
            </Button>
          )}
        </div>
        <Space wrap size="middle">
          <Input
            placeholder="搜索订单号"
            prefix={<SearchOutlined />}
            style={{ width: 220, borderRadius: 8 }}
            allowClear
            value={orderIdFilter}
            onChange={(e) => {
              setOrderIdFilter(e.target.value)
              setPage(1)
            }}
          />
          <Select
            placeholder="交易状态"
            style={{ width: 140, borderRadius: 8 }}
            allowClear
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="pending">{t('transactions.statusPending')}</Select.Option>
            <Select.Option value="success">{t('transactions.statusSuccess')}</Select.Option>
            <Select.Option value="failed">{t('transactions.statusFailed')}</Select.Option>
            <Select.Option value="cancelled">{t('orders.statusCancelled')}</Select.Option>
            <Select.Option value="refunded">{t('transactions.statusRefunded')}</Select.Option>
          </Select>
          <Select
            placeholder="支付渠道"
            style={{ width: 140, borderRadius: 8 }}
            allowClear
            value={channelFilter}
            onChange={(value) => {
              setChannelFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="stripe">Stripe</Select.Option>
            <Select.Option value="paypal">PayPal</Select.Option>
            <Select.Option value="alipay">支付宝</Select.Option>
            <Select.Option value="wechat">微信支付</Select.Option>
          </Select>
          <RangePicker
            showTime
            format="YYYY-MM-DD HH:mm"
            placeholder={['开始时间', '结束时间']}
            value={dateRange}
            onChange={(dates) => {
              setDateRange(dates)
              setPage(1)
            }}
            style={{ borderRadius: 8 }}
          />
        </Space>
      </Card>

      {/* Table */}
      <Card
        style={{
          borderRadius: 12,
        }}
        bodyStyle={{ padding: 0 }}
      >
        <Table
          columns={columns}
          dataSource={payments}
          rowKey="id"
          loading={loading}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showTotal: (total) => t('common.total', { count: total }),
            onChange: (page, pageSize) => {
              setPage(page)
              setPageSize(pageSize)
            },
            style: { padding: '16px' },
          }}
          scroll={{ x: 1400 }}
          size="middle"
          locale={{
            emptyText: (
              <div style={{ padding: '40px', textAlign: 'center' }}>
                <TransactionOutlined style={{ fontSize: 48, color: '#d9d9d9', marginBottom: 16, display: 'block' }} />
                <span style={{ color: '#999' }}>{t('common.noData')}</span>
              </div>
            ),
          }}
        />
      </Card>

      {/* Detail Drawer */}
      <Drawer
        title="交易详情"
        placement="right"
        width={720}
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
      >
        {selectedPayment && (
          <div>
            <Descriptions title="基本信息" bordered column={2}>
              <Descriptions.Item label="交易ID">{selectedPayment.id}</Descriptions.Item>
              <Descriptions.Item label="订单ID">{selectedPayment.order_id}</Descriptions.Item>
              <Descriptions.Item label="金额">
                {selectedPayment.currency} {(selectedPayment.amount / 100).toFixed(2)}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedPayment.status)}>
                  {getStatusText(selectedPayment.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="支付渠道">
                <Tag>{selectedPayment.channel}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="支付方式">
                <Tag color="blue">{selectedPayment.method}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="创建时间" span={2}>
                {dayjs(selectedPayment.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              {selectedPayment.paid_at && (
                <Descriptions.Item label="支付时间" span={2}>
                  {dayjs(selectedPayment.paid_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              )}
              {selectedPayment.expires_at && (
                <Descriptions.Item label="过期时间" span={2}>
                  {dayjs(selectedPayment.expires_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              )}
            </Descriptions>

            <Descriptions
              title="客户信息"
              bordered
              column={2}
              style={{ marginTop: 16 }}
            >
              <Descriptions.Item label="客户ID">{selectedPayment.customer_id || '-'}</Descriptions.Item>
              <Descriptions.Item label="客户邮箱">{selectedPayment.customer_email}</Descriptions.Item>
              <Descriptions.Item label="IP地址">{selectedPayment.ip_address || '-'}</Descriptions.Item>
              <Descriptions.Item label="User Agent" span={2}>
                {selectedPayment.user_agent || '-'}
              </Descriptions.Item>
            </Descriptions>

            <Descriptions
              title="其他信息"
              bordered
              column={1}
              style={{ marginTop: 16 }}
            >
              <Descriptions.Item label="描述">{selectedPayment.description || '-'}</Descriptions.Item>
              <Descriptions.Item label="回调地址">{selectedPayment.callback_url || '-'}</Descriptions.Item>
              <Descriptions.Item label="返回地址">{selectedPayment.return_url || '-'}</Descriptions.Item>
            </Descriptions>

            {selectedPayment.metadata && Object.keys(selectedPayment.metadata).length > 0 && (
              <Card title="元数据" style={{ marginTop: 16 }}>
                <pre style={{ maxHeight: 200, overflow: 'auto', background: '#f5f5f5', padding: 12 }}>
                  {JSON.stringify(selectedPayment.metadata, null, 2)}
                </pre>
              </Card>
            )}
          </div>
        )}
      </Drawer>

      {/* Refund Modal */}
      <Modal
        title="申请退款"
        open={refundModalVisible}
        onOk={handleRefundSubmit}
        onCancel={() => setRefundModalVisible(false)}
        width={500}
      >
        {selectedPayment && (
          <div style={{ marginBottom: 16 }}>
            <p>交易ID: {selectedPayment.id}</p>
            <p>原交易金额: {selectedPayment.currency} {(selectedPayment.amount / 100).toFixed(2)}</p>
          </div>
        )}
        <Form form={refundForm} layout="vertical">
          <Form.Item
            name="amount"
            label="退款金额"
            rules={[{ required: true, message: '请输入退款金额' }]}
          >
            <InputNumber
              style={{ width: '100%' }}
              min={0}
              max={selectedPayment ? selectedPayment.amount : 0}
              precision={2}
              addonBefore={selectedPayment?.currency}
            />
          </Form.Item>
          <Form.Item
            name="reason"
            label="退款原因"
            rules={[{ required: true, message: '请输入退款原因' }]}
          >
            <Input.TextArea rows={3} placeholder="请说明退款原因" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Transactions

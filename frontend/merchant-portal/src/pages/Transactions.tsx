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
} from '@ant-design/icons'
import { paymentService, Payment, PaymentStats } from '../services/paymentService'
import { useAuthStore } from '../stores/authStore'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'

const { Title } = Typography
const { RangePicker } = DatePicker

const Transactions = () => {
  const [loading, setLoading] = useState(false)
  const [payments, setPayments] = useState<Payment[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [stats, setStats] = useState<PaymentStats | null>(null)
  const [selectedPayment, setSelectedPayment] = useState<Payment | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false)
  const [refundModalVisible, setRefundModalVisible] = useState(false)
  const [refundForm] = Form.useForm()

  // Filter states
  const [orderIdFilter, setOrderIdFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [channelFilter, setChannelFilter] = useState<string | undefined>()
  const [methodFilter, setMethodFilter] = useState<string | undefined>()
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)

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

    try {
      const response = await paymentService.getStats({})
      setStats(response.data)
    } catch (error) {
      // Stats API 可能不存在，暂时忽略错误
      console.log('Stats API not available yet')
    }
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
      pending: '待支付',
      success: '成功',
      failed: '失败',
      cancelled: '已取消',
      refunded: '已退款',
    }
    return texts[status] || status
  }

  const columns: ColumnsType<Payment> = [
    {
      title: '交易ID',
      dataIndex: 'id',
      key: 'id',
      width: 100,
      ellipsis: true,
    },
    {
      title: '订单号',
      dataIndex: 'order_id',
      key: 'order_id',
      width: 100,
      ellipsis: true,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number, record) => (
        <span style={{ fontWeight: 'bold', color: '#1890ff' }}>
          {record.currency} {(amount / 100).toFixed(2)}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '支付渠道',
      dataIndex: 'channel',
      key: 'channel',
      width: 100,
      render: (channel: string) => <Tag>{channel}</Tag>,
    },
    {
      title: '支付方式',
      dataIndex: 'method',
      key: 'method',
      width: 100,
      render: (method: string) => <Tag color="blue">{method}</Tag>,
    },
    {
      title: '客户邮箱',
      dataIndex: 'customer_email',
      key: 'customer_email',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            详情
          </Button>
          {record.status === 'success' && (
            <Button
              type="link"
              size="small"
              icon={<UndoOutlined />}
              onClick={() => handleRefund(record)}
            >
              退款
            </Button>
          )}
          {record.status === 'pending' && (
            <Popconfirm
              title="确认取消"
              description="确定要取消这笔交易吗？"
              onConfirm={() => handleCancel(record)}
            >
              <Button type="link" size="small" danger>
                取消
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>交易记录</Title>
        <Button
          type="primary"
          icon={<DownloadOutlined />}
          onClick={handleExport}
          loading={loading}
        >
          导出CSV
        </Button>
      </div>

      {/* Statistics Cards */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总交易额"
                value={(stats.total_amount / 100).toFixed(2)}
                prefix={<DollarOutlined />}
                suffix="USD"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="交易笔数"
                value={stats.total_count}
                prefix={<TransactionOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="成功率"
                value={(stats.success_rate * 100).toFixed(2)}
                prefix={<CheckCircleOutlined />}
                suffix="%"
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="今日交易额"
                value={(stats.today_amount / 100).toFixed(2)}
                prefix={<ClockCircleOutlined />}
                suffix="USD"
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* Filters */}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="搜索订单号"
            prefix={<SearchOutlined />}
            style={{ width: 200 }}
            allowClear
            value={orderIdFilter}
            onChange={(e) => {
              setOrderIdFilter(e.target.value)
              setPage(1)
            }}
          />
          <Select
            placeholder="交易状态"
            style={{ width: 120 }}
            allowClear
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="pending">待支付</Select.Option>
            <Select.Option value="success">成功</Select.Option>
            <Select.Option value="failed">失败</Select.Option>
            <Select.Option value="cancelled">已取消</Select.Option>
            <Select.Option value="refunded">已退款</Select.Option>
          </Select>
          <Select
            placeholder="支付渠道"
            style={{ width: 120 }}
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
          <Select
            placeholder="支付方式"
            style={{ width: 120 }}
            allowClear
            value={methodFilter}
            onChange={(value) => {
              setMethodFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="card">信用卡</Select.Option>
            <Select.Option value="bank_transfer">银行转账</Select.Option>
            <Select.Option value="e_wallet">电子钱包</Select.Option>
          </Select>
          <RangePicker
            showTime
            format="YYYY-MM-DD HH:mm:ss"
            placeholder={['开始时间', '结束时间']}
            value={dateRange}
            onChange={(dates) => {
              setDateRange(dates)
              setPage(1)
            }}
          />
          <Button onClick={resetFilters}>重置筛选</Button>
        </Space>
      </Card>

      {/* Table */}
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
          showTotal: (total) => `共 ${total} 条`,
          onChange: (page, pageSize) => {
            setPage(page)
            setPageSize(pageSize)
          },
        }}
        scroll={{ x: 1400 }}
      />

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

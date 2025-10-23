import { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  DatePicker,
  Modal,
  Descriptions,
  message,
  Tooltip,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  DollarOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

interface Payment {
  id: string
  payment_no: string
  merchant_id: string
  merchant_name: string
  order_no: string
  amount: number
  currency: string
  channel: string
  status: string
  payment_method: string
  client_ip: string
  created_at: string
  updated_at: string
}

const Payments = () => {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<Payment[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [selectedPayment, setSelectedPayment] = useState<Payment | null>(null)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [refundModalVisible, setRefundModalVisible] = useState(false)

  // Search filters
  const [searchFilters, setSearchFilters] = useState({
    payment_no: '',
    order_no: '',
    merchant_id: '',
    status: '',
    channel: '',
    date_range: null as [dayjs.Dayjs, dayjs.Dayjs] | null,
  })

  const fetchData = async () => {
    setLoading(true)
    try {
      // TODO: Replace with actual API call
      // const response = await fetch('/api/v1/payments?' + new URLSearchParams({
      //   page: page.toString(),
      //   page_size: pageSize.toString(),
      //   ...searchFilters
      // }))
      // const result = await response.json()

      // Mock data
      const mockData: Payment[] = Array.from({ length: pageSize }, (_, i) => ({
        id: `payment-${page}-${i}`,
        payment_no: `PAY${Date.now()}${i}`,
        merchant_id: `merchant-${i % 3}`,
        merchant_name: `商户${i % 3 + 1}`,
        order_no: `ORD${Date.now()}${i}`,
        amount: Math.floor(Math.random() * 100000),
        currency: 'USD',
        channel: ['stripe', 'paypal', 'alipay'][i % 3],
        status: ['pending', 'success', 'failed', 'refunded'][i % 4],
        payment_method: 'card',
        client_ip: '192.168.1.100',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }))

      setData(mockData)
      setTotal(100)
    } catch (error) {
      message.error(t('common.operationFailed'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [page, pageSize])

  const handleSearch = () => {
    setPage(1)
    fetchData()
  }

  const handleReset = () => {
    setSearchFilters({
      payment_no: '',
      order_no: '',
      merchant_id: '',
      status: '',
      channel: '',
      date_range: null,
    })
    setPage(1)
    fetchData()
  }

  const handleViewDetail = (record: Payment) => {
    setSelectedPayment(record)
    setDetailModalVisible(true)
  }

  const handleRefund = (record: Payment) => {
    setSelectedPayment(record)
    setRefundModalVisible(true)
  }

  const handleConfirmRefund = async () => {
    try {
      // TODO: Replace with actual API call
      // await fetch(`/api/v1/payments/${selectedPayment?.id}/refund`, { method: 'POST' })

      message.success(t('payments.refundSuccess'))
      setRefundModalVisible(false)
      fetchData()
    } catch (error) {
      message.error(t('common.operationFailed'))
    }
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      pending: { color: 'processing', text: t('payments.statusPending') },
      success: { color: 'success', text: t('payments.statusSuccess') },
      failed: { color: 'error', text: t('payments.statusFailed') },
      refunded: { color: 'default', text: t('payments.statusRefunded') },
    }
    const config = statusMap[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const getChannelTag = (channel: string) => {
    const channelMap: Record<string, { color: string }> = {
      stripe: { color: 'blue' },
      paypal: { color: 'cyan' },
      alipay: { color: 'green' },
    }
    const config = channelMap[channel] || { color: 'default' }
    return <Tag color={config.color}>{channel.toUpperCase()}</Tag>
  }

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${(amount / 100).toFixed(2)}`
  }

  const columns: ColumnsType<Payment> = [
    {
      title: t('payments.paymentNo'),
      dataIndex: 'payment_no',
      key: 'payment_no',
      width: 180,
      fixed: 'left',
    },
    {
      title: t('payments.merchantName'),
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
    },
    {
      title: t('payments.orderNo'),
      dataIndex: 'order_no',
      key: 'order_no',
      width: 180,
    },
    {
      title: t('payments.amount'),
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number, record: Payment) => formatAmount(amount, record.currency),
    },
    {
      title: t('payments.channel'),
      dataIndex: 'channel',
      key: 'channel',
      width: 100,
      render: (channel: string) => getChannelTag(channel),
    },
    {
      title: t('payments.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: t('common.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 150,
      fixed: 'right',
      render: (_: unknown, record: Payment) => (
        <Space size="small">
          <Tooltip title={t('payments.viewDetail')}>
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          {record.status === 'success' && (
            <Tooltip title={t('payments.refund')}>
              <Button
                type="link"
                size="small"
                danger
                icon={<DollarOutlined />}
                onClick={() => handleRefund(record)}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Space wrap>
            <Input
              placeholder={t('payments.paymentNo')}
              value={searchFilters.payment_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, payment_no: e.target.value })
              }
              style={{ width: 200 }}
            />
            <Input
              placeholder={t('payments.orderNo')}
              value={searchFilters.order_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, order_no: e.target.value })
              }
              style={{ width: 200 }}
            />
            <Select
              placeholder={t('payments.status')}
              value={searchFilters.status || undefined}
              onChange={(value) => setSearchFilters({ ...searchFilters, status: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Select.Option value="pending">{t('payments.statusPending')}</Select.Option>
              <Select.Option value="success">{t('payments.statusSuccess')}</Select.Option>
              <Select.Option value="failed">{t('payments.statusFailed')}</Select.Option>
              <Select.Option value="refunded">{t('payments.statusRefunded')}</Select.Option>
            </Select>
            <Select
              placeholder={t('payments.channel')}
              value={searchFilters.channel || undefined}
              onChange={(value) => setSearchFilters({ ...searchFilters, channel: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Select.Option value="stripe">Stripe</Select.Option>
              <Select.Option value="paypal">PayPal</Select.Option>
              <Select.Option value="alipay">Alipay</Select.Option>
            </Select>
            <RangePicker
              value={searchFilters.date_range}
              onChange={(dates) =>
                setSearchFilters({ ...searchFilters, date_range: dates as any })
              }
            />
          </Space>
          <Space>
            <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
              {t('common.search')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              {t('common.reset')}
            </Button>
          </Space>
        </Space>
      </Card>

      <Card>
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1300 }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => t('common.total', { count: total }),
            onChange: (page, pageSize) => {
              setPage(page)
              setPageSize(pageSize)
            },
          }}
        />
      </Card>

      {/* Payment Detail Modal */}
      <Modal
        title={t('payments.paymentDetail')}
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            {t('common.cancel')}
          </Button>,
        ]}
        width={800}
      >
        {selectedPayment && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label={t('payments.paymentNo')} span={2}>
              {selectedPayment.payment_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.orderNo')} span={2}>
              {selectedPayment.order_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.merchantName')}>
              {selectedPayment.merchant_name}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.amount')}>
              {formatAmount(selectedPayment.amount, selectedPayment.currency)}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.channel')}>
              {getChannelTag(selectedPayment.channel)}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.status')}>
              {getStatusTag(selectedPayment.status)}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.paymentMethod')}>
              {selectedPayment.payment_method}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.clientIp')}>
              {selectedPayment.client_ip}
            </Descriptions.Item>
            <Descriptions.Item label={t('common.createdAt')}>
              {dayjs(selectedPayment.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label={t('common.updatedAt')}>
              {dayjs(selectedPayment.updated_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* Refund Modal */}
      <Modal
        title={t('payments.refundConfirm')}
        open={refundModalVisible}
        onOk={handleConfirmRefund}
        onCancel={() => setRefundModalVisible(false)}
        okText={t('common.confirm')}
        cancelText={t('common.cancel')}
        okButtonProps={{ danger: true }}
      >
        <p>{t('payments.refundConfirmMessage')}</p>
        {selectedPayment && (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label={t('payments.paymentNo')}>
              {selectedPayment.payment_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('payments.amount')}>
              {formatAmount(selectedPayment.amount, selectedPayment.currency)}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}

export default Payments

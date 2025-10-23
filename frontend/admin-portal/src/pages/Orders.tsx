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
  Popconfirm,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  StopOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

interface Order {
  id: string
  order_no: string
  merchant_order_no: string
  merchant_id: string
  merchant_name: string
  amount: number
  currency: string
  status: string
  payment_no?: string
  product_name: string
  notify_url: string
  return_url: string
  created_at: string
  updated_at: string
  expired_at: string
}

const Orders = () => {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<Order[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [detailModalVisible, setDetailModalVisible] = useState(false)

  const [searchFilters, setSearchFilters] = useState({
    order_no: '',
    merchant_order_no: '',
    merchant_id: '',
    status: '',
    date_range: null as [dayjs.Dayjs, dayjs.Dayjs] | null,
  })

  const fetchData = async () => {
    setLoading(true)
    try {
      // Mock data
      const mockData: Order[] = Array.from({ length: pageSize }, (_, i) => ({
        id: `order-${page}-${i}`,
        order_no: `ORD${Date.now()}${i}`,
        merchant_order_no: `MORD${Date.now()}${i}`,
        merchant_id: `merchant-${i % 3}`,
        merchant_name: `商户${i % 3 + 1}`,
        amount: Math.floor(Math.random() * 100000),
        currency: 'USD',
        status: ['pending', 'paid', 'cancelled', 'expired'][i % 4],
        payment_no: i % 2 === 0 ? `PAY${Date.now()}${i}` : undefined,
        product_name: `商品${i + 1}`,
        notify_url: 'https://example.com/notify',
        return_url: 'https://example.com/return',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        expired_at: new Date(Date.now() + 3600000).toISOString(),
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
      order_no: '',
      merchant_order_no: '',
      merchant_id: '',
      status: '',
      date_range: null,
    })
    setPage(1)
    fetchData()
  }

  const handleViewDetail = (record: Order) => {
    setSelectedOrder(record)
    setDetailModalVisible(true)
  }

  const handleCancelOrder = async (record: Order) => {
    try {
      // TODO: API call
      message.success(t('orders.cancelSuccess'))
      fetchData()
    } catch (error) {
      message.error(t('common.operationFailed'))
    }
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      pending: { color: 'processing', text: t('orders.statusPending') },
      paid: { color: 'success', text: t('orders.statusPaid') },
      cancelled: { color: 'default', text: t('orders.statusCancelled') },
      expired: { color: 'error', text: t('orders.statusExpired') },
    }
    const config = statusMap[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${(amount / 100).toFixed(2)}`
  }

  const columns: ColumnsType<Order> = [
    {
      title: t('orders.orderNo'),
      dataIndex: 'order_no',
      key: 'order_no',
      width: 180,
      fixed: 'left',
    },
    {
      title: t('orders.merchantOrderNo'),
      dataIndex: 'merchant_order_no',
      key: 'merchant_order_no',
      width: 180,
    },
    {
      title: t('orders.merchantName'),
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
    },
    {
      title: t('orders.productName'),
      dataIndex: 'product_name',
      key: 'product_name',
      width: 150,
    },
    {
      title: t('orders.amount'),
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number, record: Order) => formatAmount(amount, record.currency),
    },
    {
      title: t('orders.status'),
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
      render: (_: unknown, record: Order) => (
        <Space size="small">
          <Tooltip title={t('orders.viewDetail')}>
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          {record.status === 'pending' && (
            <Popconfirm
              title={t('orders.cancelConfirm')}
              onConfirm={() => handleCancelOrder(record)}
              okText={t('common.confirm')}
              cancelText={t('common.cancel')}
            >
              <Tooltip title={t('orders.cancel')}>
                <Button type="link" size="small" danger icon={<StopOutlined />} />
              </Tooltip>
            </Popconfirm>
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
              placeholder={t('orders.orderNo')}
              value={searchFilters.order_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, order_no: e.target.value })
              }
              style={{ width: 200 }}
            />
            <Input
              placeholder={t('orders.merchantOrderNo')}
              value={searchFilters.merchant_order_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, merchant_order_no: e.target.value })
              }
              style={{ width: 200 }}
            />
            <Select
              placeholder={t('orders.status')}
              value={searchFilters.status || undefined}
              onChange={(value) => setSearchFilters({ ...searchFilters, status: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Select.Option value="pending">{t('orders.statusPending')}</Select.Option>
              <Select.Option value="paid">{t('orders.statusPaid')}</Select.Option>
              <Select.Option value="cancelled">{t('orders.statusCancelled')}</Select.Option>
              <Select.Option value="expired">{t('orders.statusExpired')}</Select.Option>
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

      <Modal
        title={t('orders.orderDetail')}
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            {t('common.cancel')}
          </Button>,
        ]}
        width={800}
      >
        {selectedOrder && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label={t('orders.orderNo')} span={2}>
              {selectedOrder.order_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('orders.merchantOrderNo')} span={2}>
              {selectedOrder.merchant_order_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('orders.merchantName')}>
              {selectedOrder.merchant_name}
            </Descriptions.Item>
            <Descriptions.Item label={t('orders.amount')}>
              {formatAmount(selectedOrder.amount, selectedOrder.currency)}
            </Descriptions.Item>
            <Descriptions.Item label={t('orders.status')}>
              {getStatusTag(selectedOrder.status)}
            </Descriptions.Item>
            <Descriptions.Item label={t('orders.productName')}>
              {selectedOrder.product_name}
            </Descriptions.Item>
            {selectedOrder.payment_no && (
              <Descriptions.Item label={t('orders.paymentNo')} span={2}>
                {selectedOrder.payment_no}
              </Descriptions.Item>
            )}
            <Descriptions.Item label={t('orders.notifyUrl')} span={2}>
              {selectedOrder.notify_url}
            </Descriptions.Item>
            <Descriptions.Item label={t('orders.returnUrl')} span={2}>
              {selectedOrder.return_url}
            </Descriptions.Item>
            <Descriptions.Item label={t('common.createdAt')}>
              {dayjs(selectedOrder.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label={t('orders.expiredAt')}>
              {dayjs(selectedOrder.expired_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}

export default Orders

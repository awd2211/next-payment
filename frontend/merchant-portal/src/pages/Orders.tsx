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
  Divider,
  Tooltip,
  Badge,
  Skeleton,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  EyeOutlined,
  PlusOutlined,
  ShoppingCartOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  FilterOutlined,
  ClearOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { orderService, Order, OrderStats, CreateOrderRequest, OrderItem } from '../services/orderService'
import { useAuthStore } from '../stores/authStore'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'

const { Title } = Typography
const { RangePicker } = DatePicker

const Orders = () => {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [statsLoading, setStatsLoading] = useState(false)
  const [orders, setOrders] = useState<Order[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [stats, setStats] = useState<OrderStats | null>(null)
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false)
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [createForm] = Form.useForm()

  // Filter states
  const [merchantOrderIdFilter, setMerchantOrderIdFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [customerEmailFilter, setCustomerEmailFilter] = useState('')
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)

  // 计算激活的过滤器数量
  const activeFilterCount = [
    merchantOrderIdFilter,
    statusFilter,
    customerEmailFilter,
    dateRange,
  ].filter(Boolean).length

  useEffect(() => {
    loadOrders()
  }, [page, pageSize, merchantOrderIdFilter, statusFilter, customerEmailFilter, dateRange])

  useEffect(() => {
    loadStats()
  }, [])

  const loadOrders = async () => {
    const token = useAuthStore.getState().token
    if (!token) {
      console.log('No token found, skipping orders load')
      return
    }

    setLoading(true)
    try {
      const response = await orderService.list({
        page,
        page_size: pageSize,
        merchant_order_id: merchantOrderIdFilter || undefined,
        status: statusFilter,
        customer_email: customerEmailFilter || undefined,
        start_time: dateRange?.[0]?.toISOString(),
        end_time: dateRange?.[1]?.toISOString(),
      })
      console.log('Orders response:', response)
      // 安全处理响应数据，兼容不同的数据结构
      const ordersData = response.data
      if (ordersData) {
        // 如果data直接是数组，使用它；否则使用data.list
        const ordersList = Array.isArray(ordersData) ? ordersData : (ordersData.list || [])
        setOrders(ordersList)
        setTotal(ordersData.total || 0)
      } else {
        setOrders([])
        setTotal(0)
      }
    } catch (error) {
      console.error('Load orders error:', error)
      setOrders([])
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
      const response = await orderService.getStats({})
      if (response.data) {
        setStats(response.data)
      }
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setStatsLoading(false)
    }
  }

  const handleClearFilters = () => {
    setMerchantOrderIdFilter('')
    setStatusFilter(undefined)
    setCustomerEmailFilter('')
    setDateRange(null)
    setPage(1)
  }

  const handleViewDetail = (order: Order) => {
    setSelectedOrder(order)
    setDetailDrawerVisible(true)
  }

  const handleCreateOrder = () => {
    createForm.resetFields()
    setCreateModalVisible(true)
  }

  const handleCreateSubmit = async () => {
    try {
      const values = await createForm.validateFields()
      const orderData: CreateOrderRequest = {
        ...values,
        amount: Math.round(values.amount * 100), // Convert to cents
      }
      await orderService.create(orderData)
      message.success('订单创建成功')
      setCreateModalVisible(false)
      loadOrders()
    } catch (error) {
      // Error handled by interceptor or validation
    }
  }

  const handleCancel = async (order: Order) => {
    try {
      await orderService.cancel(order.id, '商户取消')
      message.success('订单已取消')
      loadOrders()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const resetFilters = () => {
    setMerchantOrderIdFilter('')
    setStatusFilter(undefined)
    setCustomerEmailFilter('')
    setDateRange(null)
    setPage(1)
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'processing',
      paid: 'success',
      cancelled: 'default',
      completed: 'cyan',
    }
    return colors[status] || 'default'
  }

  const getStatusText = (status: string) => {
    const texts: Record<string, string> = {
      pending: t('orders.statusPending'),
      paid: t('orders.statusPaid'),
      cancelled: t('orders.statusCancelled'),
      completed: '已完成',
      expired: t('orders.statusExpired'),
    }
    return texts[status] || status
  }

  const columns: ColumnsType<Order> = [
    {
      title: t('orders.orderNo'),
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
      title: t('orders.merchantOrderNo'),
      dataIndex: 'merchant_order_id',
      key: 'merchant_order_id',
      width: 150,
      ellipsis: true,
      render: (merchantOrderId: string) => (
        <Tooltip title={merchantOrderId}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {merchantOrderId?.slice(0, 12)}...
          </span>
        </Tooltip>
      ),
    },
    {
      title: t('orders.amount'),
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
      title: t('orders.status'),
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
      title: '客户姓名',
      dataIndex: 'customer_name',
      key: 'customer_name',
      width: 120,
    },
    {
      title: '客户邮箱',
      dataIndex: 'customer_email',
      key: 'customer_email',
      ellipsis: true,
      width: 180,
    },
    {
      title: '商品数量',
      key: 'items_count',
      width: 100,
      align: 'center',
      render: (_, record) => (
        <Badge count={record.items?.length || 0} showZero color="#1890ff" />
      ),
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
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title={t('orders.viewDetails')}>
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
          {record.status === 'pending' && (
            <Popconfirm
              title={t('common.confirm')}
              description={t('orders.cancel')}
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

  const itemColumns: ColumnsType<OrderItem> = [
    {
      title: '商品名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '单价',
      dataIndex: 'price',
      key: 'price',
      width: 100,
      render: (price: number) => (price / 100).toFixed(2),
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 80,
    },
    {
      title: '小计',
      dataIndex: 'amount',
      key: 'amount',
      width: 100,
      render: (amount: number) => (amount / 100).toFixed(2),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>订单管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateOrder}>
          创建订单
        </Button>
      </div>

      {/* Statistics Cards */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="订单总额"
                value={(stats.total_amount / 100).toFixed(2)}
                prefix={<DollarOutlined />}
                suffix="USD"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="订单总数"
                value={stats.total_count}
                prefix={<ShoppingCartOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已支付"
                value={stats.paid_count}
                prefix={<CheckCircleOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="今日订单额"
                value={(stats.today_amount / 100).toFixed(2)}
                prefix={<DollarOutlined />}
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
            placeholder="搜索商户订单号"
            prefix={<SearchOutlined />}
            style={{ width: 200 }}
            allowClear
            value={merchantOrderIdFilter}
            onChange={(e) => {
              setMerchantOrderIdFilter(e.target.value)
              setPage(1)
            }}
          />
          <Input
            placeholder="搜索客户邮箱"
            prefix={<SearchOutlined />}
            style={{ width: 200 }}
            allowClear
            value={customerEmailFilter}
            onChange={(e) => {
              setCustomerEmailFilter(e.target.value)
              setPage(1)
            }}
          />
          <Select
            placeholder="订单状态"
            style={{ width: 120 }}
            allowClear
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="pending">待支付</Select.Option>
            <Select.Option value="paid">已支付</Select.Option>
            <Select.Option value="cancelled">已取消</Select.Option>
            <Select.Option value="completed">已完成</Select.Option>
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
        dataSource={orders}
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
        title="订单详情"
        placement="right"
        width={900}
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
      >
        {selectedOrder && (
          <div>
            <Descriptions title="基本信息" bordered column={2}>
              <Descriptions.Item label="订单ID">{selectedOrder.id}</Descriptions.Item>
              <Descriptions.Item label="商户订单号">{selectedOrder.merchant_order_id}</Descriptions.Item>
              <Descriptions.Item label="金额">
                {selectedOrder.currency} {(selectedOrder.amount / 100).toFixed(2)}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedOrder.status)}>
                  {getStatusText(selectedOrder.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="创建时间" span={2}>
                {dayjs(selectedOrder.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              {selectedOrder.paid_at && (
                <Descriptions.Item label="支付时间" span={2}>
                  {dayjs(selectedOrder.paid_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              )}
              {selectedOrder.cancelled_at && (
                <Descriptions.Item label="取消时间" span={2}>
                  {dayjs(selectedOrder.cancelled_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
              )}
              <Descriptions.Item label="订单描述" span={2}>
                {selectedOrder.description || '-'}
              </Descriptions.Item>
            </Descriptions>

            <Divider>客户信息</Divider>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="客户ID">{selectedOrder.customer_id || '-'}</Descriptions.Item>
              <Descriptions.Item label="客户姓名">{selectedOrder.customer_name}</Descriptions.Item>
              <Descriptions.Item label="客户邮箱" span={2}>{selectedOrder.customer_email}</Descriptions.Item>
            </Descriptions>

            <Divider>商品清单</Divider>
            <Table
              columns={itemColumns}
              dataSource={selectedOrder.items}
              rowKey="name"
              pagination={false}
              size="small"
            />

            {selectedOrder.shipping_address && (
              <>
                <Divider>收货地址</Divider>
                <Descriptions bordered column={2}>
                  <Descriptions.Item label="收货人">{selectedOrder.shipping_address.name}</Descriptions.Item>
                  <Descriptions.Item label="联系电话">{selectedOrder.shipping_address.phone}</Descriptions.Item>
                  <Descriptions.Item label="国家">{selectedOrder.shipping_address.country}</Descriptions.Item>
                  <Descriptions.Item label="省份">{selectedOrder.shipping_address.province}</Descriptions.Item>
                  <Descriptions.Item label="城市">{selectedOrder.shipping_address.city}</Descriptions.Item>
                  <Descriptions.Item label="区县">{selectedOrder.shipping_address.district}</Descriptions.Item>
                  <Descriptions.Item label="详细地址" span={2}>
                    {selectedOrder.shipping_address.address}
                  </Descriptions.Item>
                  <Descriptions.Item label="邮编">{selectedOrder.shipping_address.postal_code || '-'}</Descriptions.Item>
                </Descriptions>
              </>
            )}

            {selectedOrder.billing_address && (
              <>
                <Divider>账单地址</Divider>
                <Descriptions bordered column={2}>
                  <Descriptions.Item label="账单姓名">{selectedOrder.billing_address.name}</Descriptions.Item>
                  <Descriptions.Item label="联系电话">{selectedOrder.billing_address.phone}</Descriptions.Item>
                  <Descriptions.Item label="国家">{selectedOrder.billing_address.country}</Descriptions.Item>
                  <Descriptions.Item label="省份">{selectedOrder.billing_address.province}</Descriptions.Item>
                  <Descriptions.Item label="城市">{selectedOrder.billing_address.city}</Descriptions.Item>
                  <Descriptions.Item label="区县">{selectedOrder.billing_address.district}</Descriptions.Item>
                  <Descriptions.Item label="详细地址" span={2}>
                    {selectedOrder.billing_address.address}
                  </Descriptions.Item>
                  <Descriptions.Item label="邮编">{selectedOrder.billing_address.postal_code || '-'}</Descriptions.Item>
                </Descriptions>
              </>
            )}

            {selectedOrder.metadata && Object.keys(selectedOrder.metadata).length > 0 && (
              <>
                <Divider>元数据</Divider>
                <Card>
                  <pre style={{ maxHeight: 200, overflow: 'auto', background: '#f5f5f5', padding: 12 }}>
                    {JSON.stringify(selectedOrder.metadata, null, 2)}
                  </pre>
                </Card>
              </>
            )}
          </div>
        )}
      </Drawer>

      {/* Create Order Modal */}
      <Modal
        title="创建订单"
        open={createModalVisible}
        onOk={handleCreateSubmit}
        onCancel={() => setCreateModalVisible(false)}
        width={600}
      >
        <Form form={createForm} layout="vertical">
          <Form.Item
            name="merchant_order_id"
            label="商户订单号"
            rules={[{ required: true, message: '请输入商户订单号' }]}
          >
            <Input placeholder="唯一的商户订单号" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="amount"
                label="订单金额"
                rules={[{ required: true, message: '请输入订单金额' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  precision={2}
                  placeholder="0.00"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="currency"
                label="货币"
                initialValue="USD"
                rules={[{ required: true, message: '请选择货币' }]}
              >
                <Select>
                  <Select.Option value="USD">USD</Select.Option>
                  <Select.Option value="CNY">CNY</Select.Option>
                  <Select.Option value="EUR">EUR</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="customer_email"
            label="客户邮箱"
            rules={[
              { required: true, message: '请输入客户邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder="customer@example.com" />
          </Form.Item>

          <Form.Item
            name="customer_name"
            label="客户姓名"
            rules={[{ required: true, message: '请输入客户姓名' }]}
          >
            <Input placeholder="客户姓名" />
          </Form.Item>

          <Form.Item name="customer_id" label="客户ID">
            <Input placeholder="可选的客户ID" />
          </Form.Item>

          <Form.Item
            name="description"
            label="订单描述"
            rules={[{ required: true, message: '请输入订单描述' }]}
          >
            <Input.TextArea rows={2} placeholder="订单描述" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Orders

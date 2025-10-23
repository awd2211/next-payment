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
  Descriptions,
  Form,
  message,
  Statistic,
  Row,
  Col,
  Popconfirm,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  ReloadOutlined,
  ExportOutlined,
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  DollarOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

interface Settlement {
  settlement_no: string
  merchant_id: string
  merchant_name: string
  settlement_amount: number
  fee_amount: number
  actual_amount: number
  currency: string
  settlement_date: string
  settlement_period_start: string
  settlement_period_end: string
  transaction_count: number
  status: string
  bank_account: string
  bank_name: string
  account_holder: string
  remark: string
  created_at: string
  updated_at: string
}

interface SettlementStats {
  total_settlements: number
  pending_amount: number
  processing_amount: number
  completed_amount: number
  failed_count: number
}

const Settlements = () => {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [dataSource, setDataSource] = useState<Settlement[]>([])
  const [selectedSettlement, setSelectedSettlement] = useState<Settlement | null>(null)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [stats, setStats] = useState<SettlementStats>({
    total_settlements: 0,
    pending_amount: 0,
    processing_amount: 0,
    completed_amount: 0,
    failed_count: 0,
  })
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })
  const [searchFilters, setSearchFilters] = useState({
    settlement_no: '',
    merchant_id: '',
    status: '',
    date_range: null as [dayjs.Dayjs, dayjs.Dayjs] | null,
  })

  useEffect(() => {
    fetchSettlements()
    fetchStats()
  }, [pagination.current, pagination.pageSize])

  const fetchStats = async () => {
    // Mock stats data
    setStats({
      total_settlements: 245,
      pending_amount: 125000.00,
      processing_amount: 89500.00,
      completed_amount: 1234567.50,
      failed_count: 3,
    })
  }

  const fetchSettlements = async () => {
    setLoading(true)
    try {
      // Mock data
      const mockData: Settlement[] = Array.from({ length: 10 }, (_, i) => ({
        settlement_no: `STL${Date.now() + i}`,
        merchant_id: `M${1000 + i}`,
        merchant_name: `商户${i + 1}`,
        settlement_amount: Math.floor(Math.random() * 100000) + 10000,
        fee_amount: Math.floor(Math.random() * 1000) + 100,
        actual_amount: Math.floor(Math.random() * 99000) + 9000,
        currency: 'CNY',
        settlement_date: dayjs().add(i, 'day').format('YYYY-MM-DD'),
        settlement_period_start: dayjs().subtract(7 + i, 'day').format('YYYY-MM-DD'),
        settlement_period_end: dayjs().subtract(i, 'day').format('YYYY-MM-DD'),
        transaction_count: Math.floor(Math.random() * 100) + 10,
        status: ['pending', 'processing', 'completed', 'failed'][Math.floor(Math.random() * 4)],
        bank_account: `**** **** **** ${1000 + i}`,
        bank_name: ['工商银行', '建设银行', '农业银行', '招商银行'][Math.floor(Math.random() * 4)],
        account_holder: `商户${i + 1}`,
        remark: i % 3 === 0 ? '正常结算' : '',
        created_at: dayjs().subtract(i, 'day').toISOString(),
        updated_at: dayjs().subtract(i, 'hour').toISOString(),
      }))

      setDataSource(mockData)
      setPagination((prev) => ({ ...prev, total: 100 }))
    } catch (error) {
      message.error('获取结算记录失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = () => {
    setPagination((prev) => ({ ...prev, current: 1 }))
    fetchSettlements()
  }

  const handleReset = () => {
    setSearchFilters({
      settlement_no: '',
      merchant_id: '',
      status: '',
      date_range: null,
    })
    setPagination((prev) => ({ ...prev, current: 1 }))
    fetchSettlements()
  }

  const handleViewDetail = (record: Settlement) => {
    setSelectedSettlement(record)
    setDetailModalVisible(true)
  }

  const handleApprove = async (record: Settlement) => {
    try {
      message.success(t('settlements.approveSuccess'))
      fetchSettlements()
      fetchStats()
    } catch (error) {
      message.error(t('common.operationFailed'))
    }
  }

  const handleReject = async (record: Settlement) => {
    try {
      message.success(t('settlements.rejectSuccess'))
      fetchSettlements()
      fetchStats()
    } catch (error) {
      message.error(t('common.operationFailed'))
    }
  }

  const handleExport = () => {
    message.success(t('settlements.exportSuccess'))
  }

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      pending: { color: 'default', text: t('settlements.statusPending') },
      processing: { color: 'processing', text: t('settlements.statusProcessing') },
      completed: { color: 'success', text: t('settlements.statusCompleted') },
      failed: { color: 'error', text: t('settlements.statusFailed') },
    }
    const config = statusMap[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const columns: ColumnsType<Settlement> = [
    {
      title: t('settlements.settlementNo'),
      dataIndex: 'settlement_no',
      key: 'settlement_no',
      fixed: 'left',
      width: 180,
    },
    {
      title: t('settlements.merchantName'),
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
    },
    {
      title: t('settlements.settlementPeriod'),
      key: 'settlement_period',
      width: 220,
      render: (_, record) => (
        <div>
          {dayjs(record.settlement_period_start).format('YYYY-MM-DD')}
          <br />
          至 {dayjs(record.settlement_period_end).format('YYYY-MM-DD')}
        </div>
      ),
    },
    {
      title: t('settlements.settlementDate'),
      dataIndex: 'settlement_date',
      key: 'settlement_date',
      width: 120,
      render: (date) => dayjs(date).format('YYYY-MM-DD'),
    },
    {
      title: t('settlements.transactionCount'),
      dataIndex: 'transaction_count',
      key: 'transaction_count',
      width: 100,
      align: 'right',
    },
    {
      title: t('settlements.settlementAmount'),
      dataIndex: 'settlement_amount',
      key: 'settlement_amount',
      width: 150,
      align: 'right',
      render: (amount, record) => formatAmount(amount, record.currency),
    },
    {
      title: t('settlements.feeAmount'),
      dataIndex: 'fee_amount',
      key: 'fee_amount',
      width: 120,
      align: 'right',
      render: (amount, record) => formatAmount(amount, record.currency),
    },
    {
      title: t('settlements.actualAmount'),
      dataIndex: 'actual_amount',
      key: 'actual_amount',
      width: 150,
      align: 'right',
      render: (amount, record) => (
        <span style={{ fontWeight: 'bold', color: '#52c41a' }}>
          {formatAmount(amount, record.currency)}
        </span>
      ),
    },
    {
      title: t('settlements.status'),
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => getStatusTag(status),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      fixed: 'right',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            {t('settlements.viewDetail')}
          </Button>
          {record.status === 'pending' && (
            <>
              <Popconfirm
                title={t('settlements.approveConfirm')}
                onConfirm={() => handleApprove(record)}
              >
                <Button
                  type="link"
                  size="small"
                  icon={<CheckOutlined />}
                  style={{ color: '#52c41a' }}
                >
                  {t('settlements.approve')}
                </Button>
              </Popconfirm>
              <Popconfirm
                title={t('settlements.rejectConfirm')}
                onConfirm={() => handleReject(record)}
              >
                <Button
                  type="link"
                  size="small"
                  danger
                  icon={<CloseOutlined />}
                >
                  {t('settlements.reject')}
                </Button>
              </Popconfirm>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title={t('settlements.totalSettlements')}
              value={stats.total_settlements}
              prefix={<DollarOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title={t('settlements.pendingAmount')}
              value={stats.pending_amount}
              precision={2}
              prefix="¥"
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title={t('settlements.processingAmount')}
              value={stats.processing_amount}
              precision={2}
              prefix="¥"
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title={t('settlements.completedAmount')}
              value={stats.completed_amount}
              precision={2}
              prefix="¥"
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Card>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Space wrap>
            <Input
              placeholder={t('settlements.settlementNo')}
              value={searchFilters.settlement_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, settlement_no: e.target.value })
              }
              style={{ width: 200 }}
            />
            <Input
              placeholder={t('settlements.merchantName')}
              value={searchFilters.merchant_id}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, merchant_id: e.target.value })
              }
              style={{ width: 200 }}
            />
            <Select
              placeholder={t('settlements.status')}
              value={searchFilters.status}
              onChange={(value) => setSearchFilters({ ...searchFilters, status: value })}
              style={{ width: 150 }}
              allowClear
            >
              <Select.Option value="pending">{t('settlements.statusPending')}</Select.Option>
              <Select.Option value="processing">{t('settlements.statusProcessing')}</Select.Option>
              <Select.Option value="completed">{t('settlements.statusCompleted')}</Select.Option>
              <Select.Option value="failed">{t('settlements.statusFailed')}</Select.Option>
            </Select>
            <RangePicker
              value={searchFilters.date_range}
              onChange={(dates) =>
                setSearchFilters({ ...searchFilters, date_range: dates as [dayjs.Dayjs, dayjs.Dayjs] | null })
              }
            />
            <Button
              type="primary"
              icon={<SearchOutlined />}
              onClick={handleSearch}
            >
              {t('common.search')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              {t('common.reset')}
            </Button>
            <Button
              icon={<ExportOutlined />}
              onClick={handleExport}
            >
              {t('common.export')}
            </Button>
          </Space>

          <Table
            columns={columns}
            dataSource={dataSource}
            rowKey="settlement_no"
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
            scroll={{ x: 1800 }}
          />
        </Space>
      </Card>

      <Modal
        title={t('settlements.settlementDetail')}
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            {t('common.cancel')}
          </Button>,
        ]}
        width={800}
      >
        {selectedSettlement && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label={t('settlements.settlementNo')} span={2}>
              {selectedSettlement.settlement_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.merchantName')}>
              {selectedSettlement.merchant_name}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.merchantId')}>
              {selectedSettlement.merchant_id}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.settlementPeriod')} span={2}>
              {dayjs(selectedSettlement.settlement_period_start).format('YYYY-MM-DD')} 至{' '}
              {dayjs(selectedSettlement.settlement_period_end).format('YYYY-MM-DD')}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.settlementDate')}>
              {dayjs(selectedSettlement.settlement_date).format('YYYY-MM-DD')}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.transactionCount')}>
              {selectedSettlement.transaction_count} 笔
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.settlementAmount')}>
              {formatAmount(selectedSettlement.settlement_amount, selectedSettlement.currency)}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.feeAmount')}>
              {formatAmount(selectedSettlement.fee_amount, selectedSettlement.currency)}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.actualAmount')} span={2}>
              <span style={{ fontWeight: 'bold', fontSize: '16px', color: '#52c41a' }}>
                {formatAmount(selectedSettlement.actual_amount, selectedSettlement.currency)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.bankName')}>
              {selectedSettlement.bank_name}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.bankAccount')}>
              {selectedSettlement.bank_account}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.accountHolder')} span={2}>
              {selectedSettlement.account_holder}
            </Descriptions.Item>
            <Descriptions.Item label={t('settlements.status')} span={2}>
              {getStatusTag(selectedSettlement.status)}
            </Descriptions.Item>
            {selectedSettlement.remark && (
              <Descriptions.Item label={t('settlements.remark')} span={2}>
                {selectedSettlement.remark}
              </Descriptions.Item>
            )}
            <Descriptions.Item label={t('common.createdAt')}>
              {dayjs(selectedSettlement.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label={t('common.updatedAt')}>
              {dayjs(selectedSettlement.updated_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}

export default Settlements

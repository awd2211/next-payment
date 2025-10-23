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
  Statistic,
  Row,
  Col,
  Alert,
  Timeline,
  Divider,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  BankOutlined,
  CalendarOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

interface Settlement {
  settlement_no: string
  settlement_date: string
  settlement_period_start: string
  settlement_period_end: string
  transaction_count: number
  settlement_amount: number
  fee_amount: number
  actual_amount: number
  currency: string
  status: string
  bank_account: string
  bank_name: string
  account_holder: string
  remark: string
  created_at: string
  completed_at?: string
}

interface SettlementStats {
  total_settlements: number
  pending_amount: number
  processing_amount: number
  completed_amount: number
  this_month_amount: number
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
    this_month_amount: 0,
  })
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })
  const [searchFilters, setSearchFilters] = useState({
    settlement_no: '',
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
      total_settlements: 48,
      pending_amount: 15680.50,
      processing_amount: 28900.00,
      completed_amount: 456789.25,
      this_month_amount: 89456.78,
    })
  }

  const fetchSettlements = async () => {
    setLoading(true)
    try {
      // Mock data
      const mockData: Settlement[] = Array.from({ length: 10 }, (_, i) => {
        const settlementAmount = Math.floor(Math.random() * 50000) + 10000
        const feeAmount = Math.floor(settlementAmount * 0.02)
        return {
          settlement_no: `STL${Date.now() + i}`,
          settlement_date: dayjs().add(i + 1, 'day').format('YYYY-MM-DD'),
          settlement_period_start: dayjs().subtract(7 + i, 'day').format('YYYY-MM-DD'),
          settlement_period_end: dayjs().subtract(i, 'day').format('YYYY-MM-DD'),
          transaction_count: Math.floor(Math.random() * 50) + 10,
          settlement_amount: settlementAmount,
          fee_amount: feeAmount,
          actual_amount: settlementAmount - feeAmount,
          currency: 'CNY',
          status: ['pending', 'processing', 'completed'][Math.floor(Math.random() * 3)],
          bank_account: `**** **** **** ${1000 + i}`,
          bank_name: ['工商银行', '建设银行', '招商银行'][Math.floor(Math.random() * 3)],
          account_holder: '商户名称',
          remark: i % 3 === 0 ? '正常结算' : '',
          created_at: dayjs().subtract(i + 1, 'day').toISOString(),
          completed_at: i % 2 === 0 ? dayjs().subtract(i, 'day').toISOString() : undefined,
        }
      })

      setDataSource(mockData)
      setPagination((prev) => ({ ...prev, total: 50 }))
    } catch (error) {
      console.error('Failed to fetch settlements:', error)
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

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${(amount / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; icon: React.ReactNode; text: string }> = {
      pending: {
        color: 'default',
        icon: <ClockCircleOutlined />,
        text: t('settlements.statusPending')
      },
      processing: {
        color: 'processing',
        icon: <ClockCircleOutlined />,
        text: t('settlements.statusProcessing')
      },
      completed: {
        color: 'success',
        icon: <CheckCircleOutlined />,
        text: t('settlements.statusCompleted')
      },
    }
    const config = statusMap[status] || { color: 'default', icon: null, text: status }
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    )
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
      render: (count) => `${count} 笔`,
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
      width: 120,
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record)}
        >
          {t('settlements.viewDetail')}
        </Button>
      ),
    },
  ]

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('settlements.totalSettlements')}
              value={stats.total_settlements}
              prefix={<BankOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
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
        <Col xs={24} sm={12} lg={6}>
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
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('settlements.thisMonthAmount')}
              value={stats.this_month_amount}
              precision={2}
              prefix="¥"
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>

      <Card>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Alert
            message={t('settlements.notice')}
            description={t('settlements.noticeDesc')}
            type="info"
            showIcon
            closable
          />

          <Space wrap>
            <Input
              placeholder={t('settlements.settlementNo')}
              value={searchFilters.settlement_no}
              onChange={(e) =>
                setSearchFilters({ ...searchFilters, settlement_no: e.target.value })
              }
              style={{ width: 200 }}
              prefix={<SearchOutlined />}
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
            </Select>
            <RangePicker
              value={searchFilters.date_range}
              onChange={(dates) =>
                setSearchFilters({ ...searchFilters, date_range: dates as [dayjs.Dayjs, dayjs.Dayjs] | null })
              }
              placeholder={[t('settlements.startDate'), t('settlements.endDate')]}
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
            scroll={{ x: 1500 }}
          />
        </Space>
      </Card>

      {/* Detail Modal */}
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
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label={t('settlements.settlementNo')} span={2}>
                {selectedSettlement.settlement_no}
              </Descriptions.Item>
              <Descriptions.Item label={t('settlements.settlementPeriod')} span={2}>
                {dayjs(selectedSettlement.settlement_period_start).format('YYYY-MM-DD')} 至{' '}
                {dayjs(selectedSettlement.settlement_period_end).format('YYYY-MM-DD')}
                <span style={{ marginLeft: 8, color: '#999' }}>
                  ({dayjs(selectedSettlement.settlement_period_end).diff(
                    dayjs(selectedSettlement.settlement_period_start),
                    'day'
                  )}{' '}
                  天)
                </span>
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
                <span style={{ fontWeight: 'bold', fontSize: 16, color: '#52c41a' }}>
                  {formatAmount(selectedSettlement.actual_amount, selectedSettlement.currency)}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label={t('settlements.status')} span={2}>
                {getStatusTag(selectedSettlement.status)}
              </Descriptions.Item>
            </Descriptions>

            <Divider>{t('settlements.bankInfo')}</Divider>

            <Descriptions bordered column={2}>
              <Descriptions.Item label={t('settlements.bankName')}>
                {selectedSettlement.bank_name}
              </Descriptions.Item>
              <Descriptions.Item label={t('settlements.bankAccount')}>
                {selectedSettlement.bank_account}
              </Descriptions.Item>
              <Descriptions.Item label={t('settlements.accountHolder')} span={2}>
                {selectedSettlement.account_holder}
              </Descriptions.Item>
            </Descriptions>

            {selectedSettlement.remark && (
              <>
                <Divider>{t('settlements.remark')}</Divider>
                <Alert message={selectedSettlement.remark} type="info" showIcon />
              </>
            )}

            <Divider>{t('settlements.timeline')}</Divider>

            <Timeline
              items={[
                {
                  color: 'green',
                  children: (
                    <>
                      <p style={{ fontWeight: 'bold' }}>
                        <CalendarOutlined /> {t('settlements.created')}
                      </p>
                      <p style={{ color: '#999' }}>
                        {dayjs(selectedSettlement.created_at).format('YYYY-MM-DD HH:mm:ss')}
                      </p>
                    </>
                  ),
                },
                ...(selectedSettlement.completed_at
                  ? [
                      {
                        color: 'green',
                        children: (
                          <>
                            <p style={{ fontWeight: 'bold' }}>
                              <CheckCircleOutlined /> {t('settlements.completed')}
                            </p>
                            <p style={{ color: '#999' }}>
                              {dayjs(selectedSettlement.completed_at).format('YYYY-MM-DD HH:mm:ss')}
                            </p>
                          </>
                        ),
                      },
                    ]
                  : []),
              ]}
            />
          </div>
        )}
      </Modal>
    </div>
  )
}

export default Settlements

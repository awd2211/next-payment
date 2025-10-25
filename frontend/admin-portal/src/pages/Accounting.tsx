import { useState, useEffect } from 'react'
import { Card, Table, Tabs, Statistic, Row, Col, DatePicker, Select, Space, message } from 'antd'
import { DollarOutlined, LineChartOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { accountingService, type AccountingEntry, type AccountingSummary } from '../services/accountingService'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

export default function Accounting() {
  const [loading, setLoading] = useState(false)
  const [entries, setEntries] = useState<AccountingEntry[]>([])
  const [summary, setSummary] = useState<AccountingSummary | null>(null)
  const [dateRange, setDateRange] = useState<[string, string]>([
    dayjs().startOf('month').format('YYYY-MM-DD'),
    dayjs().endOf('month').format('YYYY-MM-DD'),
  ])
  const [currency, setCurrency] = useState<string>('USD')

  useEffect(() => {
    fetchData()
    fetchSummary()
  }, [dateRange, currency])

  const fetchData = async () => {
    setLoading(true)
    try {
      const response = await accountingService.listEntries({
        page: 1,
        page_size: 50,
        start_date: dateRange[0],
        end_date: dateRange[1],
        currency,
      })
      if (response.code === 0 && response.data) {
        setEntries(response.data.list)
      } else {
        message.error(response.error?.message || '加载失败')
      }
    } catch (error) {
      message.error('加载失败')
      console.error('Failed to fetch accounting entries:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchSummary = async () => {
    try {
      const response = await accountingService.getSummary({
        start_date: dateRange[0],
        end_date: dateRange[1],
        currency,
      })
      if (response.code === 0 && response.data) {
        setSummary(response.data)
      }
    } catch (error) {
      console.error('Failed to fetch accounting summary:', error)
    }
  }

  const columns: ColumnsType<AccountingEntry> = [
    { title: '凭证号', dataIndex: 'entry_no', width: 180 },
    { title: '日期', dataIndex: 'account_date', width: 120 },
    { title: '借方科目', dataIndex: 'debit_account', width: 150 },
    { title: '贷方科目', dataIndex: 'credit_account', width: 150 },
    {
      title: '金额',
      dataIndex: 'amount',
      width: 120,
      render: (amount, record) => `${record.currency} ${(amount / 100).toFixed(2)}`,
    },
    { title: '摘要', dataIndex: 'description', ellipsis: true },
    { title: '参考号', dataIndex: 'reference_no', width: 180 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
  ]

  const handleDateRangeChange = (dates: any) => {
    if (dates && dates[0] && dates[1]) {
      setDateRange([
        dates[0].format('YYYY-MM-DD'),
        dates[1].format('YYYY-MM-DD'),
      ])
    }
  }

  const handleCurrencyChange = (value: string) => {
    setCurrency(value)
  }

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总资产"
              value={summary?.total_assets || 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix={currency}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="总负债"
              value={summary?.total_liabilities || 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix={currency}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="本期收入"
              value={summary?.total_revenue || 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix={currency}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="本期支出"
              value={summary?.total_expense || 0}
              precision={2}
              prefix={<LineChartOutlined />}
              suffix={currency}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="账务管理"
        extra={
          <Space>
            <RangePicker
              defaultValue={[dayjs(dateRange[0]), dayjs(dateRange[1])]}
              onChange={handleDateRangeChange}
            />
            <Select
              placeholder="币种"
              style={{ width: 100 }}
              value={currency}
              onChange={handleCurrencyChange}
              options={[
                { label: 'USD', value: 'USD' },
                { label: 'CNY', value: 'CNY' },
                { label: 'EUR', value: 'EUR' },
                { label: 'GBP', value: 'GBP' },
              ]}
            />
          </Space>
        }
      >
        <Tabs
          items={[
            {
              key: 'entries',
              label: '会计分录',
              children: (
                <Table
                  columns={columns}
                  dataSource={entries}
                  loading={loading}
                  rowKey="id"
                  scroll={{ x: 1200 }}
                />
              ),
            },
            {
              key: 'balance',
              label: '余额表',
              children: <div style={{ padding: 50, textAlign: 'center' }}>余额表功能开发中...</div>,
            },
            {
              key: 'ledger',
              label: '总账',
              children: <div style={{ padding: 50, textAlign: 'center' }}>总账功能开发中...</div>,
            },
          ]}
        />
      </Card>
    </div>
  )
}

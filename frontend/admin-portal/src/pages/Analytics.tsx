import { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, DatePicker, Select, Space, Tabs } from 'antd'
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import {
  DollarOutlined,
  ShoppingOutlined,
  RiseOutlined,
  FallOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

// 颜色配置
const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8']

export default function Analytics() {
  const [loading, setLoading] = useState(false)
  const [dateRange, setDateRange] = useState<[string, string]>([
    dayjs().subtract(30, 'days').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD'),
  ])
  const [currency, setCurrency] = useState<string>('USD')

  // Mock数据 - 支付趋势
  const paymentTrendData = [
    { date: '2025-10-01', amount: 45000, count: 120 },
    { date: '2025-10-05', amount: 52000, count: 145 },
    { date: '2025-10-10', amount: 48000, count: 132 },
    { date: '2025-10-15', amount: 61000, count: 168 },
    { date: '2025-10-20', amount: 58000, count: 155 },
    { date: '2025-10-25', amount: 67000, count: 182 },
  ]

  // Mock数据 - 渠道分布
  const channelDistData = [
    { name: 'Stripe', value: 45000, count: 450 },
    { name: 'PayPal', value: 28000, count: 280 },
    { name: 'Alipay', value: 18000, count: 180 },
    { name: 'WeChat Pay', value: 12000, count: 120 },
  ]

  // Mock数据 - 商户排行
  const merchantRankData = [
    { merchant: '商户A', amount: 85000, transactions: 320 },
    { merchant: '商户B', amount: 72000, transactions: 285 },
    { merchant: '商户C', amount: 61000, transactions: 245 },
    { merchant: '商户D', amount: 53000, transactions: 210 },
    { merchant: '商户E', amount: 48000, transactions: 195 },
  ]

  // Mock数据 - 支付状态分布
  const statusDistData = [
    { name: '成功', value: 8520 },
    { name: '失败', value: 320 },
    { name: '处理中', value: 85 },
    { name: '退款', value: 165 },
  ]

  useEffect(() => {
    fetchData()
  }, [dateRange, currency])

  const fetchData = async () => {
    setLoading(true)
    // TODO: 调用 analyticsService.getPaymentTrend()
    // TODO: 调用 analyticsService.getChannelDistribution()
    // TODO: 调用 analyticsService.getMerchantRanking()
    setTimeout(() => {
      setLoading(false)
    }, 500)
  }

  const handleDateRangeChange = (dates: any) => {
    if (dates && dates[0] && dates[1]) {
      setDateRange([dates[0].format('YYYY-MM-DD'), dates[1].format('YYYY-MM-DD')])
    }
  }

  return (
    <div>
      {/* 筛选条件 */}
      <Card style={{ marginBottom: 16 }}>
        <Space>
          <RangePicker
            defaultValue={[dayjs(dateRange[0]), dayjs(dateRange[1])]}
            onChange={handleDateRangeChange}
          />
          <Select
            value={currency}
            onChange={setCurrency}
            style={{ width: 100 }}
            options={[
              { label: 'USD', value: 'USD' },
              { label: 'CNY', value: 'CNY' },
              { label: 'EUR', value: 'EUR' },
              { label: 'GBP', value: 'GBP' },
            ]}
          />
        </Space>
      </Card>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总交易额"
              value={325600}
              precision={2}
              prefix={<DollarOutlined />}
              suffix={currency}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="交易笔数"
              value={9090}
              prefix={<ShoppingOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功率"
              value={96.4}
              precision={1}
              suffix="%"
              prefix={<RiseOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="退款率"
              value={1.8}
              precision={1}
              suffix="%"
              prefix={<FallOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Tabs
        items={[
          {
            key: 'trend',
            label: '支付趋势',
            children: (
              <Card>
                <ResponsiveContainer width="100%" height={400}>
                  <LineChart data={paymentTrendData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="date" />
                    <YAxis yAxisId="left" />
                    <YAxis yAxisId="right" orientation="right" />
                    <Tooltip />
                    <Legend />
                    <Line
                      yAxisId="left"
                      type="monotone"
                      dataKey="amount"
                      stroke="#8884d8"
                      name="交易金额"
                    />
                    <Line
                      yAxisId="right"
                      type="monotone"
                      dataKey="count"
                      stroke="#82ca9d"
                      name="交易笔数"
                    />
                  </LineChart>
                </ResponsiveContainer>
              </Card>
            ),
          },
          {
            key: 'channel',
            label: '渠道分布',
            children: (
              <Row gutter={16}>
                <Col span={12}>
                  <Card title="按金额分布">
                    <ResponsiveContainer width="100%" height={350}>
                      <PieChart>
                        <Pie
                          data={channelDistData}
                          dataKey="value"
                          nameKey="name"
                          cx="50%"
                          cy="50%"
                          outerRadius={120}
                          label
                        >
                          {channelDistData.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                          ))}
                        </Pie>
                        <Tooltip />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card title="按笔数分布">
                    <ResponsiveContainer width="100%" height={350}>
                      <BarChart data={channelDistData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="name" />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Bar dataKey="count" fill="#8884d8" name="交易笔数" />
                      </BarChart>
                    </ResponsiveContainer>
                  </Card>
                </Col>
              </Row>
            ),
          },
          {
            key: 'merchant',
            label: '商户排行',
            children: (
              <Card>
                <ResponsiveContainer width="100%" height={400}>
                  <BarChart data={merchantRankData} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis type="number" />
                    <YAxis dataKey="merchant" type="category" width={100} />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="amount" fill="#8884d8" name="交易金额" />
                    <Bar dataKey="transactions" fill="#82ca9d" name="交易笔数" />
                  </BarChart>
                </ResponsiveContainer>
              </Card>
            ),
          },
          {
            key: 'status',
            label: '状态分布',
            children: (
              <Card>
                <ResponsiveContainer width="100%" height={400}>
                  <PieChart>
                    <Pie
                      data={statusDistData}
                      dataKey="value"
                      nameKey="name"
                      cx="50%"
                      cy="50%"
                      outerRadius={150}
                      label={(entry: any) => `${entry.name}: ${entry.value}`}
                    >
                      {statusDistData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </Card>
            ),
          },
        ]}
      />
    </div>
  )
}

import { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, DatePicker, Space, Tabs } from 'antd'
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
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
  PercentageOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042']

export default function Analytics() {
  const [_loading, setLoading] = useState(false)
  const [dateRange, setDateRange] = useState<[string, string]>([
    dayjs().subtract(30, 'days').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD'),
  ])

  // Mock数据 - 交易趋势
  const transactionTrendData = [
    { date: '10-01', revenue: 4500, orders: 45, success_rate: 96 },
    { date: '10-05', revenue: 5200, orders: 52, success_rate: 97 },
    { date: '10-10', revenue: 4800, orders: 48, success_rate: 95 },
    { date: '10-15', revenue: 6100, orders: 61, success_rate: 98 },
    { date: '10-20', revenue: 5800, orders: 58, success_rate: 96 },
    { date: '10-25', revenue: 6700, orders: 67, success_rate: 97 },
  ]

  // Mock数据 - 渠道对比
  const channelCompareData = [
    { channel: 'Stripe', revenue: 25000, orders: 250, avg_amount: 100 },
    { channel: 'PayPal', revenue: 18000, orders: 180, avg_amount: 100 },
    { channel: 'Alipay', revenue: 12000, orders: 120, avg_amount: 100 },
  ]

  // Mock数据 - 支付方式分布
  const paymentMethodData = [
    { name: '信用卡', value: 45 },
    { name: '借记卡', value: 28 },
    { name: '数字钱包', value: 18 },
    { name: '其他', value: 9 },
  ]

  // Mock数据 - 时段分布
  const hourlyData = [
    { hour: '00:00', count: 15 },
    { hour: '04:00', count: 8 },
    { hour: '08:00', count: 35 },
    { hour: '12:00', count: 58 },
    { hour: '16:00', count: 45 },
    { hour: '20:00', count: 52 },
  ]

  useEffect(() => {
    fetchData()
  }, [dateRange])

  const fetchData = async () => {
    setLoading(true)
    // TODO: 调用 analyticsService.getMerchantAnalytics()
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
        </Space>
      </Card>

      {/* 核心指标卡片 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总收入"
              value={55000}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="USD"
              valueStyle={{ color: '#3f8600' }}
            />
            <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>
              <RiseOutlined style={{ color: '#52c41a' }} /> 环比上升 12.5%
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="订单数" value={550} prefix={<ShoppingOutlined />} />
            <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>
              <RiseOutlined style={{ color: '#52c41a' }} /> 环比上升 8.2%
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="平均订单金额"
              value={100}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="USD"
            />
            <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>
              <RiseOutlined style={{ color: '#52c41a' }} /> 环比上升 3.8%
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="支付成功率"
              value={96.7}
              precision={1}
              suffix="%"
              prefix={<PercentageOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
            <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>较上期持平</div>
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Tabs
        items={[
          {
            key: 'trend',
            label: '交易趋势',
            children: (
              <Card>
                <ResponsiveContainer width="100%" height={400}>
                  <AreaChart data={transactionTrendData}>
                    <defs>
                      <linearGradient id="colorRevenue" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
                        <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="date" />
                    <YAxis yAxisId="left" />
                    <YAxis yAxisId="right" orientation="right" />
                    <Tooltip />
                    <Legend />
                    <Area
                      yAxisId="left"
                      type="monotone"
                      dataKey="revenue"
                      stroke="#8884d8"
                      fillOpacity={1}
                      fill="url(#colorRevenue)"
                      name="收入"
                    />
                    <Line
                      yAxisId="right"
                      type="monotone"
                      dataKey="orders"
                      stroke="#82ca9d"
                      name="订单数"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </Card>
            ),
          },
          {
            key: 'channel',
            label: '渠道对比',
            children: (
              <Card>
                <ResponsiveContainer width="100%" height={400}>
                  <BarChart data={channelCompareData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="channel" />
                    <YAxis yAxisId="left" />
                    <YAxis yAxisId="right" orientation="right" />
                    <Tooltip />
                    <Legend />
                    <Bar yAxisId="left" dataKey="revenue" fill="#8884d8" name="收入" />
                    <Bar yAxisId="right" dataKey="orders" fill="#82ca9d" name="订单数" />
                  </BarChart>
                </ResponsiveContainer>
              </Card>
            ),
          },
          {
            key: 'method',
            label: '支付方式',
            children: (
              <Row gutter={16}>
                <Col span={12}>
                  <Card title="支付方式分布">
                    <ResponsiveContainer width="100%" height={350}>
                      <PieChart>
                        <Pie
                          data={paymentMethodData}
                          dataKey="value"
                          nameKey="name"
                          cx="50%"
                          cy="50%"
                          outerRadius={120}
                          label={(entry) => `${entry.name}: ${entry.value}%`}
                        >
                          {paymentMethodData.map((entry, index) => (
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
                  <Card title="时段分布">
                    <ResponsiveContainer width="100%" height={350}>
                      <BarChart data={hourlyData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="hour" />
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
            key: 'conversion',
            label: '转化分析',
            children: (
              <Card>
                <Row gutter={16}>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic title="访问量" value={5800} />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic title="下单量" value={620} />
                      <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>
                        转化率: 10.7%
                      </div>
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic title="支付成功" value={550} />
                      <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>
                        支付成功率: 88.7%
                      </div>
                    </Card>
                  </Col>
                </Row>

                <Card title="转化漏斗" style={{ marginTop: 16 }}>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={[
                      { stage: '访问', count: 5800, rate: 100 },
                      { stage: '下单', count: 620, rate: 10.7 },
                      { stage: '支付', count: 550, rate: 9.5 },
                    ]} layout="vertical">
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis type="number" />
                      <YAxis dataKey="stage" type="category" />
                      <Tooltip />
                      <Legend />
                      <Bar dataKey="count" fill="#8884d8" name="用户数" />
                    </BarChart>
                  </ResponsiveContainer>
                </Card>
              </Card>
            ),
          },
        ]}
      />
    </div>
  )
}

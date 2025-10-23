import { Row, Col, Card, Statistic, Typography } from 'antd'
import {
  UserOutlined,
  ShoppingOutlined,
  DollarOutlined,
  SafetyOutlined,
} from '@ant-design/icons'

const { Title } = Typography

const Dashboard = () => {
  // TODO: 从API获取实际数据
  const stats = {
    totalAdmins: 25,
    totalMerchants: 150,
    totalTransactions: 1234,
    totalAmount: 5678900,
  }

  return (
    <div>
      <Title level={2}>仪表板</Title>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="管理员总数"
              value={stats.totalAdmins}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="商户总数"
              value={stats.totalMerchants}
              prefix={<ShoppingOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日交易笔数"
              value={stats.totalTransactions}
              prefix={<SafetyOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日交易额（元）"
              value={stats.totalAmount / 100}
              precision={2}
              prefix={<DollarOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={16}>
          <Card title="近期活动" style={{ height: 400 }}>
            {/* TODO: 添加图表或活动列表 */}
            <p>暂无数据</p>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="快捷操作" style={{ height: 400 }}>
            {/* TODO: 添加常用操作链接 */}
            <p>暂无数据</p>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard

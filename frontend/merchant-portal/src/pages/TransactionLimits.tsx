import { Card, Descriptions, Tag, Row, Col, Statistic } from 'antd'

export default function TransactionLimits() {
  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={8}>
          <Card>
            <Statistic title="单笔限额" value={10000} suffix="USD" />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="日限额" value={100000} suffix="USD" />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="月限额" value={1000000} suffix="USD" />
          </Card>
        </Col>
      </Row>

      <Card title="交易限额详情">
        <Descriptions bordered>
          <Descriptions.Item label="单笔最小金额">USD 0.50</Descriptions.Item>
          <Descriptions.Item label="单笔最大金额">USD 10,000.00</Descriptions.Item>
          <Descriptions.Item label="每日交易次数">1000次</Descriptions.Item>
          <Descriptions.Item label="每日交易金额">USD 100,000.00</Descriptions.Item>
          <Descriptions.Item label="每月交易金额">USD 1,000,000.00</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color="green">正常</Tag>
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  )
}

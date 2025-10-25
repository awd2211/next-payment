import React, { useState } from 'react'
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
  Progress,
  Alert,
  Statistic,
  Row,
  Col,
  Tabs,
} from 'antd'
import {
  SearchOutlined,
  EyeOutlined,
  DownloadOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

interface ReconciliationRecord {
  id: string
  recon_no: string
  channel: string
  recon_date: string
  total_count: number
  matched_count: number
  unmatched_count: number
  platform_total_amount: number
  channel_total_amount: number
  difference_amount: number
  status: 'processing' | 'completed' | 'confirmed'
  created_at: string
  completed_at?: string
}

interface UnmatchedItem {
  id: string
  payment_no: string
  order_no: string
  amount: number
  currency: string
  platform_time: string
  channel_time?: string
  reason: string
}

export default function Reconciliation() {
  const [loading, setLoading] = useState(false)
  const [records, setRecords] = useState<ReconciliationRecord[]>([
    {
      id: '1',
      recon_no: 'REC-2024-02-01',
      channel: 'Stripe',
      recon_date: '2024-02-01',
      total_count: 523,
      matched_count: 521,
      unmatched_count: 2,
      platform_total_amount: 52300.00,
      channel_total_amount: 52290.00,
      difference_amount: 10.00,
      status: 'confirmed',
      created_at: '2024-02-02 02:00:00',
      completed_at: '2024-02-02 02:10:15',
    },
    {
      id: '2',
      recon_no: 'REC-2024-02-02',
      channel: 'PayPal',
      recon_date: '2024-02-02',
      total_count: 687,
      matched_count: 687,
      unmatched_count: 0,
      platform_total_amount: 68700.00,
      channel_total_amount: 68700.00,
      difference_amount: 0,
      status: 'confirmed',
      created_at: '2024-02-03 02:00:00',
      completed_at: '2024-02-03 02:12:30',
    },
  ])

  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedRecord, setSelectedRecord] = useState<ReconciliationRecord | null>(null)

  // Mock unmatched items
  const unmatchedItems: UnmatchedItem[] = [
    {
      id: '1',
      payment_no: 'PAY-2024-123456',
      order_no: 'ORD-2024-123456',
      amount: 5.00,
      currency: 'USD',
      platform_time: '2024-02-01 15:30:00',
      reason: 'Missing in channel records',
    },
    {
      id: '2',
      payment_no: 'PAY-2024-123457',
      order_no: 'ORD-2024-123457',
      amount: 5.00,
      currency: 'USD',
      platform_time: '2024-02-01 16:45:00',
      channel_time: '2024-02-01 16:46:00',
      reason: 'Amount mismatch',
    },
  ]

  const columns: ColumnsType<ReconciliationRecord> = [
    {
      title: '对账单号',
      dataIndex: 'recon_no',
      key: 'recon_no',
      width: 160,
    },
    {
      title: '支付渠道',
      dataIndex: 'channel',
      key: 'channel',
      width: 120,
    },
    {
      title: '对账日期',
      dataIndex: 'recon_date',
      key: 'recon_date',
      width: 120,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD'),
    },
    {
      title: '交易笔数',
      dataIndex: 'total_count',
      key: 'total_count',
      width: 100,
      render: (count: number) => count.toLocaleString(),
    },
    {
      title: '匹配进度',
      key: 'match_progress',
      width: 180,
      render: (_, record) => {
        const percentage = (record.matched_count / record.total_count) * 100
        return (
          <div>
            <div style={{ fontSize: 12, marginBottom: 4 }}>
              {record.matched_count} / {record.total_count}
            </div>
            <Progress
              percent={Number(percentage.toFixed(2))}
              size="small"
              status={percentage === 100 ? 'success' : 'normal'}
            />
          </div>
        )
      },
    },
    {
      title: '平台总金额',
      dataIndex: 'platform_total_amount',
      key: 'platform_total_amount',
      width: 130,
      render: (amount: number) => `$${amount.toLocaleString()}`,
    },
    {
      title: '渠道总金额',
      dataIndex: 'channel_total_amount',
      key: 'channel_total_amount',
      width: 130,
      render: (amount: number) => `$${amount.toLocaleString()}`,
    },
    {
      title: '差异金额',
      dataIndex: 'difference_amount',
      key: 'difference_amount',
      width: 120,
      render: (amount: number) => (
        <span style={{ color: amount === 0 ? '#52c41a' : '#faad14' }}>
          ${amount.toFixed(2)}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusConfig = {
          processing: { color: 'blue', text: '处理中' },
          completed: { color: 'orange', text: '已完成' },
          confirmed: { color: 'green', text: '已确认' },
        }
        const config = statusConfig[status as keyof typeof statusConfig]
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            查看详情
          </Button>
          <Button type="link" size="small" icon={<DownloadOutlined />}>
            下载
          </Button>
        </Space>
      ),
    },
  ]

  const unmatchedColumns: ColumnsType<UnmatchedItem> = [
    {
      title: '支付单号',
      dataIndex: 'payment_no',
      key: 'payment_no',
    },
    {
      title: '订单号',
      dataIndex: 'order_no',
      key: 'order_no',
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number, record) => `${record.currency} ${amount.toFixed(2)}`,
    },
    {
      title: '平台时间',
      dataIndex: 'platform_time',
      key: 'platform_time',
    },
    {
      title: '渠道时间',
      dataIndex: 'channel_time',
      key: 'channel_time',
      render: (time?: string) => time || '-',
    },
    {
      title: '差异原因',
      dataIndex: 'reason',
      key: 'reason',
    },
  ]

  const handleViewDetail = (record: ReconciliationRecord) => {
    setSelectedRecord(record)
    setDetailVisible(true)
  }

  const totalDifferenceAmount = records.reduce((sum, r) => sum + r.difference_amount, 0)
  const averageMatchRate =
    records.length > 0
      ? records.reduce((sum, r) => sum + (r.matched_count / r.total_count) * 100, 0) / records.length
      : 0

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="对账记录数"
              value={records.length}
              prefix={<FileTextOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="平均匹配率"
              value={averageMatchRate.toFixed(2)}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: averageMatchRate >= 99 ? '#52c41a' : '#faad14' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="总差异金额"
              value={totalDifferenceAmount.toFixed(2)}
              prefix="$"
              valueStyle={{ color: totalDifferenceAmount === 0 ? '#52c41a' : '#faad14' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="待确认"
              value={records.filter((r) => r.status === 'completed').length}
              prefix={<WarningOutlined />}
              valueStyle={{
                color: records.filter((r) => r.status === 'completed').length > 0 ? '#faad14' : '#52c41a',
              }}
            />
          </Col>
        </Row>
      </Card>

      <Card
        title="对账记录"
        extra={
          <Space>
            <Select placeholder="支付渠道" style={{ width: 120 }} allowClear>
              <Select.Option value="stripe">Stripe</Select.Option>
              <Select.Option value="paypal">PayPal</Select.Option>
              <Select.Option value="alipay">Alipay</Select.Option>
            </Select>
            <Select placeholder="状态" style={{ width: 120 }} allowClear>
              <Select.Option value="processing">处理中</Select.Option>
              <Select.Option value="completed">已完成</Select.Option>
              <Select.Option value="confirmed">已确认</Select.Option>
            </Select>
            <RangePicker />
            <Button type="primary" icon={<SearchOutlined />}>
              搜索
            </Button>
          </Space>
        }
      >
        <Alert
          message="对账说明"
          description="平台每日凌晨自动对账，比对平台交易记录与渠道账单数据。如有差异，请及时联系客服处理。"
          type="info"
          showIcon
          closable
          style={{ marginBottom: 16 }}
        />

        <Table
          columns={columns}
          dataSource={records}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1800 }}
          pagination={{
            total: records.length,
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      {/* Detail Modal */}
      <Modal
        title="对账详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="download" icon={<DownloadOutlined />}>
            下载对账单
          </Button>,
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={1000}
      >
        {selectedRecord && (
          <Tabs
            items={[
              {
                key: 'summary',
                label: '对账汇总',
                children: (
                  <div>
                    <Descriptions column={2} bordered>
                      <Descriptions.Item label="对账单号">
                        {selectedRecord.recon_no}
                      </Descriptions.Item>
                      <Descriptions.Item label="对账日期">
                        {dayjs(selectedRecord.recon_date).format('YYYY-MM-DD')}
                      </Descriptions.Item>
                      <Descriptions.Item label="支付渠道">
                        {selectedRecord.channel}
                      </Descriptions.Item>
                      <Descriptions.Item label="状态">
                        <Tag
                          color={
                            selectedRecord.status === 'confirmed'
                              ? 'green'
                              : selectedRecord.status === 'completed'
                              ? 'orange'
                              : 'blue'
                          }
                        >
                          {selectedRecord.status === 'confirmed'
                            ? '已确认'
                            : selectedRecord.status === 'completed'
                            ? '已完成'
                            : '处理中'}
                        </Tag>
                      </Descriptions.Item>
                      <Descriptions.Item label="总交易笔数">
                        {selectedRecord.total_count.toLocaleString()}
                      </Descriptions.Item>
                      <Descriptions.Item label="匹配笔数">
                        <span style={{ color: '#52c41a' }}>
                          {selectedRecord.matched_count.toLocaleString()}
                        </span>
                      </Descriptions.Item>
                      <Descriptions.Item label="差异笔数">
                        <span style={{ color: selectedRecord.unmatched_count > 0 ? '#faad14' : '#52c41a' }}>
                          {selectedRecord.unmatched_count}
                        </span>
                      </Descriptions.Item>
                      <Descriptions.Item label="匹配率">
                        {((selectedRecord.matched_count / selectedRecord.total_count) * 100).toFixed(2)}%
                      </Descriptions.Item>
                      <Descriptions.Item label="平台总金额">
                        ${selectedRecord.platform_total_amount.toLocaleString()}
                      </Descriptions.Item>
                      <Descriptions.Item label="渠道总金额">
                        ${selectedRecord.channel_total_amount.toLocaleString()}
                      </Descriptions.Item>
                      <Descriptions.Item label="差异金额">
                        <span style={{ color: selectedRecord.difference_amount !== 0 ? '#faad14' : '#52c41a' }}>
                          ${selectedRecord.difference_amount.toFixed(2)}
                        </span>
                      </Descriptions.Item>
                      <Descriptions.Item label="创建时间">
                        {selectedRecord.created_at}
                      </Descriptions.Item>
                      <Descriptions.Item label="完成时间">
                        {selectedRecord.completed_at || '-'}
                      </Descriptions.Item>
                    </Descriptions>

                    {selectedRecord.unmatched_count > 0 && (
                      <Alert
                        message="差异提醒"
                        description={`发现 ${selectedRecord.unmatched_count} 笔差异记录，差异金额 $${selectedRecord.difference_amount.toFixed(2)}，如有疑问请联系客服`}
                        type="warning"
                        showIcon
                        style={{ marginTop: 16 }}
                      />
                    )}
                  </div>
                ),
              },
              {
                key: 'unmatched',
                label: `差异明细 (${unmatchedItems.length})`,
                children:
                  unmatchedItems.length > 0 ? (
                    <Table
                      columns={unmatchedColumns}
                      dataSource={unmatchedItems}
                      rowKey="id"
                      pagination={false}
                    />
                  ) : (
                    <Alert message="无差异记录" description="本次对账无差异" type="success" showIcon />
                  ),
              },
            ]}
          />
        )}
      </Modal>
    </div>
  )
}

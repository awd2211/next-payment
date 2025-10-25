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
  Form,
  message,
  Descriptions,
  Progress,
  Alert,
  Statistic,
  Row,
  Col,
  Tabs,
  Upload,
} from 'antd'
import {
  SearchOutlined,
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  UploadOutlined,
  DownloadOutlined,
  SyncOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

interface ReconciliationRecord {
  id: string
  recon_no: string
  merchant_name: string
  merchant_id: string
  channel: string
  recon_date: string
  total_count: number
  matched_count: number
  unmatched_count: number
  platform_total_amount: number
  channel_total_amount: number
  difference_amount: number
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'confirmed'
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
  status: 'pending' | 'resolved' | 'ignored'
}

export default function Reconciliation() {
  const [loading, setLoading] = useState(false)
  const [records, setRecords] = useState<ReconciliationRecord[]>([
    {
      id: '1',
      recon_no: 'REC-2024-02-01',
      merchant_name: 'Tech Store',
      merchant_id: 'MCH-001',
      channel: 'Stripe',
      recon_date: '2024-02-01',
      total_count: 1523,
      matched_count: 1520,
      unmatched_count: 3,
      platform_total_amount: 152300.50,
      channel_total_amount: 152280.50,
      difference_amount: 20.00,
      status: 'completed',
      created_at: '2024-02-02 02:00:00',
      completed_at: '2024-02-02 02:15:30',
    },
    {
      id: '2',
      recon_no: 'REC-2024-02-02',
      merchant_name: 'Fashion Store',
      merchant_id: 'MCH-002',
      channel: 'PayPal',
      recon_date: '2024-02-02',
      total_count: 2315,
      matched_count: 2315,
      unmatched_count: 0,
      platform_total_amount: 231500.00,
      channel_total_amount: 231500.00,
      difference_amount: 0,
      status: 'confirmed',
      created_at: '2024-02-03 02:00:00',
      completed_at: '2024-02-03 02:20:15',
    },
  ])

  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedRecord, setSelectedRecord] = useState<ReconciliationRecord | null>(null)
  const [createVisible, setCreateVisible] = useState(false)
  const [form] = Form.useForm()

  // Mock unmatched items
  const unmatchedItems: UnmatchedItem[] = [
    {
      id: '1',
      payment_no: 'PAY-2024-123456',
      order_no: 'ORD-2024-123456',
      amount: 10.00,
      currency: 'USD',
      platform_time: '2024-02-01 15:30:00',
      reason: 'Missing in channel records',
      status: 'pending',
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
      status: 'pending',
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
      title: '商户名称',
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
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
      width: 150,
      render: (_, record) => {
        const percentage = (record.matched_count / record.total_count) * 100
        return (
          <Progress
            percent={Number(percentage.toFixed(2))}
            size="small"
            status={percentage === 100 ? 'success' : percentage > 95 ? 'normal' : 'exception'}
          />
        )
      },
    },
    {
      title: '差异金额',
      dataIndex: 'difference_amount',
      key: 'difference_amount',
      width: 120,
      render: (amount: number) => (
        <span style={{ color: amount === 0 ? '#52c41a' : amount > 0 ? '#faad14' : '#f5222d' }}>
          USD {amount.toFixed(2)}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => {
        const statusConfig = {
          pending: { color: 'default', text: '待处理' },
          processing: { color: 'blue', text: '处理中' },
          completed: { color: 'orange', text: '已完成' },
          confirmed: { color: 'green', text: '已确认' },
          failed: { color: 'red', text: '失败' },
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
          {record.status === 'completed' && (
            <Button
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => handleConfirm(record)}
            >
              确认
            </Button>
          )}
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
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusConfig = {
          pending: { color: 'orange', text: '待处理' },
          resolved: { color: 'green', text: '已处理' },
          ignored: { color: 'default', text: '已忽略' },
        }
        const config = statusConfig[status as keyof typeof statusConfig]
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
  ]

  const handleViewDetail = (record: ReconciliationRecord) => {
    setSelectedRecord(record)
    setDetailVisible(true)
  }

  const handleConfirm = async (record: ReconciliationRecord) => {
    Modal.confirm({
      title: '确认对账结果',
      content: '确认后将无法修改，是否继续？',
      onOk: async () => {
        try {
          setLoading(true)
          // TODO: Call API to confirm reconciliation
          await new Promise((resolve) => setTimeout(resolve, 1000))

          message.success('对账结果已确认')
          setRecords((prev) =>
            prev.map((r) => (r.id === record.id ? { ...r, status: 'confirmed' } : r))
          )
        } catch (error) {
          message.error('确认失败，请重试')
        } finally {
          setLoading(false)
        }
      },
    })
  }

  const handleCreateReconciliation = () => {
    setCreateVisible(true)
    form.resetFields()
  }

  const handleCreateSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      // TODO: Call API to create reconciliation
      await new Promise((resolve) => setTimeout(resolve, 2000))

      message.success('对账任务已创建，正在处理中...')
      setCreateVisible(false)
      form.resetFields()
    } catch (error) {
      console.error('Create reconciliation error:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="今日对账"
              value={records.length}
              prefix={<SyncOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="待确认"
              value={records.filter((r) => r.status === 'completed').length}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="已确认"
              value={records.filter((r) => r.status === 'confirmed').length}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="差异笔数"
              value={records.reduce((sum, r) => sum + r.unmatched_count, 0)}
              valueStyle={{ color: '#f5222d' }}
            />
          </Col>
        </Row>
      </Card>

      <Card
        title="对账管理"
        extra={
          <Space>
            <Input
              placeholder="搜索对账单号/商户"
              prefix={<SearchOutlined />}
              style={{ width: 200 }}
            />
            <Select placeholder="支付渠道" style={{ width: 120 }} allowClear>
              <Select.Option value="stripe">Stripe</Select.Option>
              <Select.Option value="paypal">PayPal</Select.Option>
              <Select.Option value="alipay">Alipay</Select.Option>
            </Select>
            <Select placeholder="状态" style={{ width: 120 }} allowClear>
              <Select.Option value="pending">待处理</Select.Option>
              <Select.Option value="processing">处理中</Select.Option>
              <Select.Option value="completed">已完成</Select.Option>
              <Select.Option value="confirmed">已确认</Select.Option>
            </Select>
            <RangePicker />
            <Button type="primary" icon={<SyncOutlined />} onClick={handleCreateReconciliation}>
              发起对账
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={records}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1600 }}
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
          <Button key="export" icon={<DownloadOutlined />}>
            导出报告
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
                label: '汇总信息',
                children: (
                  <div>
                    <Descriptions column={2} bordered>
                      <Descriptions.Item label="对账单号">
                        {selectedRecord.recon_no}
                      </Descriptions.Item>
                      <Descriptions.Item label="对账日期">
                        {dayjs(selectedRecord.recon_date).format('YYYY-MM-DD')}
                      </Descriptions.Item>
                      <Descriptions.Item label="商户名称">
                        {selectedRecord.merchant_name}
                      </Descriptions.Item>
                      <Descriptions.Item label="支付渠道">
                        {selectedRecord.channel}
                      </Descriptions.Item>
                      <Descriptions.Item label="总交易笔数">
                        {selectedRecord.total_count.toLocaleString()}
                      </Descriptions.Item>
                      <Descriptions.Item label="匹配笔数">
                        {selectedRecord.matched_count.toLocaleString()}
                      </Descriptions.Item>
                      <Descriptions.Item label="差异笔数">
                        <span style={{ color: selectedRecord.unmatched_count > 0 ? '#f5222d' : '#52c41a' }}>
                          {selectedRecord.unmatched_count}
                        </span>
                      </Descriptions.Item>
                      <Descriptions.Item label="匹配率">
                        {((selectedRecord.matched_count / selectedRecord.total_count) * 100).toFixed(2)}%
                      </Descriptions.Item>
                      <Descriptions.Item label="平台总金额">
                        USD {selectedRecord.platform_total_amount.toFixed(2)}
                      </Descriptions.Item>
                      <Descriptions.Item label="渠道总金额">
                        USD {selectedRecord.channel_total_amount.toFixed(2)}
                      </Descriptions.Item>
                      <Descriptions.Item label="差异金额">
                        <span style={{ color: selectedRecord.difference_amount !== 0 ? '#f5222d' : '#52c41a' }}>
                          USD {selectedRecord.difference_amount.toFixed(2)}
                        </span>
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
                          {selectedRecord.status === 'confirmed' ? '已确认' : selectedRecord.status === 'completed' ? '已完成' : '处理中'}
                        </Tag>
                      </Descriptions.Item>
                    </Descriptions>

                    {selectedRecord.unmatched_count > 0 && (
                      <Alert
                        message="差异提醒"
                        description={`发现 ${selectedRecord.unmatched_count} 笔差异记录，差异金额 USD ${selectedRecord.difference_amount.toFixed(2)}，请及时处理`}
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
                children: (
                  <Table
                    columns={unmatchedColumns}
                    dataSource={unmatchedItems}
                    rowKey="id"
                    pagination={false}
                  />
                ),
              },
            ]}
          />
        )}
      </Modal>

      {/* Create Reconciliation Modal */}
      <Modal
        title="发起对账"
        open={createVisible}
        onOk={handleCreateSubmit}
        onCancel={() => setCreateVisible(false)}
        confirmLoading={loading}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="merchant_id"
            label="选择商户"
            rules={[{ required: true, message: '请选择商户' }]}
          >
            <Select placeholder="选择商户（留空表示全部商户）" allowClear>
              <Select.Option value="MCH-001">Tech Store</Select.Option>
              <Select.Option value="MCH-002">Fashion Store</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="channel"
            label="支付渠道"
            rules={[{ required: true, message: '请选择支付渠道' }]}
          >
            <Select placeholder="选择支付渠道">
              <Select.Option value="stripe">Stripe</Select.Option>
              <Select.Option value="paypal">PayPal</Select.Option>
              <Select.Option value="alipay">Alipay</Select.Option>
              <Select.Option value="all">全部渠道</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="recon_date"
            label="对账日期"
            rules={[{ required: true, message: '请选择对账日期' }]}
          >
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="file" label="渠道账单文件（可选）">
            <Upload>
              <Button icon={<UploadOutlined />}>上传渠道账单（CSV/Excel）</Button>
            </Upload>
          </Form.Item>
          <Alert
            message="对账说明"
            description="系统将自动匹配平台交易记录和渠道账单数据，对账过程可能需要几分钟时间"
            type="info"
            showIcon
          />
        </Form>
      </Modal>
    </div>
  )
}

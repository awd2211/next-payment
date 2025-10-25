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
  Alert,
  Statistic,
  Row,
  Col,
  Typography,
  Timeline,
  Tabs,
} from 'antd'
import {
  SearchOutlined,
  EyeOutlined,
  ReloadOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker
const { TextArea } = Input
const { Text } = Typography

interface WebhookLog {
  id: string
  webhook_id: string
  merchant_name: string
  merchant_id: string
  event_type: string
  url: string
  request_body: string
  response_status: number
  response_body: string
  retry_count: number
  status: 'success' | 'failed' | 'pending' | 'retrying'
  created_at: string
  completed_at?: string
}

export default function Webhooks() {
  const [loading, setLoading] = useState(false)
  const [webhookLogs, setWebhookLogs] = useState<WebhookLog[]>([
    {
      id: '1',
      webhook_id: 'WH-001',
      merchant_name: 'Tech Store',
      merchant_id: 'MCH-001',
      event_type: 'payment.success',
      url: 'https://example.com/webhook',
      request_body: JSON.stringify({ event: 'payment.success', payment_no: 'PAY-123' }),
      response_status: 200,
      response_body: JSON.stringify({ code: 0, message: 'success' }),
      retry_count: 0,
      status: 'success',
      created_at: '2024-02-01 10:30:00',
      completed_at: '2024-02-01 10:30:01',
    },
    {
      id: '2',
      webhook_id: 'WH-002',
      merchant_name: 'Fashion Store',
      merchant_id: 'MCH-002',
      event_type: 'payment.failed',
      url: 'https://example.com/webhook',
      request_body: JSON.stringify({ event: 'payment.failed', payment_no: 'PAY-124' }),
      response_status: 500,
      response_body: JSON.stringify({ error: 'Internal Server Error' }),
      retry_count: 3,
      status: 'failed',
      created_at: '2024-02-01 11:15:00',
      completed_at: '2024-02-01 11:20:00',
    },
  ])

  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedLog, setSelectedLog] = useState<WebhookLog | null>(null)
  const [retryVisible, setRetryVisible] = useState(false)

  const columns: ColumnsType<WebhookLog> = [
    {
      title: 'Webhook ID',
      dataIndex: 'webhook_id',
      key: 'webhook_id',
      width: 120,
    },
    {
      title: '商户名称',
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
    },
    {
      title: '事件类型',
      dataIndex: 'event_type',
      key: 'event_type',
      width: 150,
      render: (type: string) => {
        const typeConfig: Record<string, { color: string; text: string }> = {
          'payment.success': { color: 'green', text: '支付成功' },
          'payment.failed': { color: 'red', text: '支付失败' },
          'refund.success': { color: 'blue', text: '退款成功' },
          'settlement.completed': { color: 'purple', text: '结算完成' },
        }
        const config = typeConfig[type] || { color: 'default', text: type }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: 'Webhook URL',
      dataIndex: 'url',
      key: 'url',
      ellipsis: true,
      width: 250,
    },
    {
      title: '响应状态',
      dataIndex: 'response_status',
      key: 'response_status',
      width: 100,
      render: (status: number) => (
        <Tag color={status >= 200 && status < 300 ? 'green' : 'red'}>{status}</Tag>
      ),
    },
    {
      title: '重试次数',
      dataIndex: 'retry_count',
      key: 'retry_count',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusConfig = {
          success: { color: 'green', text: '成功' },
          failed: { color: 'red', text: '失败' },
          pending: { color: 'default', text: '待发送' },
          retrying: { color: 'orange', text: '重试中' },
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
          {record.status === 'failed' && (
            <Button
              type="link"
              size="small"
              icon={<ReloadOutlined />}
              onClick={() => handleRetry(record)}
            >
              重试
            </Button>
          )}
        </Space>
      ),
    },
  ]

  const handleViewDetail = (log: WebhookLog) => {
    setSelectedLog(log)
    setDetailVisible(true)
  }

  const handleRetry = (log: WebhookLog) => {
    setSelectedLog(log)
    setRetryVisible(true)
  }

  const handleRetryConfirm = async () => {
    try {
      setLoading(true)
      // TODO: Call API to retry webhook
      await new Promise((resolve) => setTimeout(resolve, 1000))

      message.success('Webhook 重试请求已提交')
      setRetryVisible(false)

      // Update status
      setWebhookLogs((prev) =>
        prev.map((log) =>
          log.id === selectedLog?.id ? { ...log, status: 'retrying' as const } : log
        )
      )
    } catch (error) {
      message.error('重试失败，请稍后再试')
    } finally {
      setLoading(false)
    }
  }

  const successCount = webhookLogs.filter((log) => log.status === 'success').length
  const failedCount = webhookLogs.filter((log) => log.status === 'failed').length
  const successRate = webhookLogs.length > 0 ? (successCount / webhookLogs.length) * 100 : 0

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="总发送量"
              value={webhookLogs.length}
              prefix={<ApiOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="成功"
              value={successCount}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="失败"
              value={failedCount}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="成功率"
              value={successRate.toFixed(2)}
              suffix="%"
              valueStyle={{ color: successRate > 95 ? '#52c41a' : '#faad14' }}
            />
          </Col>
        </Row>
      </Card>

      <Card
        title="Webhook 日志"
        extra={
          <Space>
            <Input
              placeholder="搜索商户名称/Webhook ID"
              prefix={<SearchOutlined />}
              style={{ width: 220 }}
            />
            <Select placeholder="事件类型" style={{ width: 150 }} allowClear>
              <Select.Option value="payment.success">支付成功</Select.Option>
              <Select.Option value="payment.failed">支付失败</Select.Option>
              <Select.Option value="refund.success">退款成功</Select.Option>
              <Select.Option value="settlement.completed">结算完成</Select.Option>
            </Select>
            <Select placeholder="状态" style={{ width: 120 }} allowClear>
              <Select.Option value="success">成功</Select.Option>
              <Select.Option value="failed">失败</Select.Option>
              <Select.Option value="retrying">重试中</Select.Option>
            </Select>
            <RangePicker />
            <Button type="primary" icon={<SearchOutlined />}>
              搜索
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={webhookLogs}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1600 }}
          pagination={{
            total: webhookLogs.length,
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      {/* Detail Modal */}
      <Modal
        title="Webhook 详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={900}
      >
        {selectedLog && (
          <Tabs
            items={[
              {
                key: 'info',
                label: '基本信息',
                children: (
                  <Descriptions column={2} bordered>
                    <Descriptions.Item label="Webhook ID">
                      {selectedLog.webhook_id}
                    </Descriptions.Item>
                    <Descriptions.Item label="商户名称">
                      {selectedLog.merchant_name}
                    </Descriptions.Item>
                    <Descriptions.Item label="事件类型">
                      <Tag>{selectedLog.event_type}</Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="响应状态">
                      <Tag
                        color={
                          selectedLog.response_status >= 200 &&
                          selectedLog.response_status < 300
                            ? 'green'
                            : 'red'
                        }
                      >
                        {selectedLog.response_status}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="重试次数">
                      {selectedLog.retry_count}
                    </Descriptions.Item>
                    <Descriptions.Item label="状态">
                      <Tag
                        color={
                          selectedLog.status === 'success'
                            ? 'green'
                            : selectedLog.status === 'failed'
                            ? 'red'
                            : 'orange'
                        }
                      >
                        {selectedLog.status === 'success'
                          ? '成功'
                          : selectedLog.status === 'failed'
                          ? '失败'
                          : '重试中'}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="Webhook URL" span={2}>
                      <Text copyable>{selectedLog.url}</Text>
                    </Descriptions.Item>
                    <Descriptions.Item label="创建时间">
                      {selectedLog.created_at}
                    </Descriptions.Item>
                    <Descriptions.Item label="完成时间">
                      {selectedLog.completed_at || '-'}
                    </Descriptions.Item>
                  </Descriptions>
                ),
              },
              {
                key: 'request',
                label: '请求数据',
                children: (
                  <div>
                    <Alert
                      message="请求 Body"
                      type="info"
                      style={{ marginBottom: 16 }}
                    />
                    <TextArea
                      value={JSON.stringify(
                        JSON.parse(selectedLog.request_body),
                        null,
                        2
                      )}
                      rows={15}
                      readOnly
                      style={{ fontFamily: 'monospace' }}
                    />
                  </div>
                ),
              },
              {
                key: 'response',
                label: '响应数据',
                children: (
                  <div>
                    <Alert
                      message={`响应状态: ${selectedLog.response_status}`}
                      type={
                        selectedLog.response_status >= 200 &&
                        selectedLog.response_status < 300
                          ? 'success'
                          : 'error'
                      }
                      style={{ marginBottom: 16 }}
                    />
                    <TextArea
                      value={JSON.stringify(
                        JSON.parse(selectedLog.response_body),
                        null,
                        2
                      )}
                      rows={15}
                      readOnly
                      style={{ fontFamily: 'monospace' }}
                    />
                  </div>
                ),
              },
              {
                key: 'timeline',
                label: '发送历史',
                children: (
                  <Timeline
                    items={[
                      {
                        color: 'blue',
                        children: (
                          <div>
                            <div>创建 Webhook</div>
                            <Text type="secondary">{selectedLog.created_at}</Text>
                          </div>
                        ),
                      },
                      ...(selectedLog.retry_count > 0
                        ? Array.from({ length: selectedLog.retry_count }).map((_, i) => ({
                            color: 'orange',
                            children: (
                              <div>
                                <div>第 {i + 1} 次重试</div>
                                <Text type="secondary">自动重试</Text>
                              </div>
                            ),
                          }))
                        : []),
                      {
                        color: selectedLog.status === 'success' ? 'green' : 'red',
                        children: (
                          <div>
                            <div>{selectedLog.status === 'success' ? '发送成功' : '发送失败'}</div>
                            <Text type="secondary">{selectedLog.completed_at || '-'}</Text>
                          </div>
                        ),
                      },
                    ]}
                  />
                ),
              },
            ]}
          />
        )}
      </Modal>

      {/* Retry Confirmation Modal */}
      <Modal
        title="重试 Webhook"
        open={retryVisible}
        onOk={handleRetryConfirm}
        onCancel={() => setRetryVisible(false)}
        confirmLoading={loading}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Alert
            message="确认重试"
            description="系统将重新发送此 Webhook 到商户的回调地址，确认继续？"
            type="warning"
            showIcon
          />
          {selectedLog && (
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="Webhook ID">{selectedLog.webhook_id}</Descriptions.Item>
              <Descriptions.Item label="事件类型">{selectedLog.event_type}</Descriptions.Item>
              <Descriptions.Item label="目标 URL">{selectedLog.url}</Descriptions.Item>
              <Descriptions.Item label="已重试次数">{selectedLog.retry_count}</Descriptions.Item>
            </Descriptions>
          )}
        </Space>
      </Modal>
    </div>
  )
}

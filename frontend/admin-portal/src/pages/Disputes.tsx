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
  Timeline,
  Upload,
  Tabs,
  Statistic,
  Row,
  Col,
} from 'antd'
import {
  SearchOutlined,
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  UploadOutlined,
  FileTextOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker
const { TextArea } = Input

interface Dispute {
  id: string
  dispute_no: string
  payment_no: string
  order_no: string
  merchant_name: string
  merchant_id: string
  amount: number
  currency: string
  reason: string
  status: 'pending' | 'reviewing' | 'accepted' | 'rejected' | 'withdrawn'
  evidence_deadline: string
  submitted_at: string
  resolved_at?: string
  created_by: string
}

interface Evidence {
  id: string
  file_name: string
  file_type: string
  file_url: string
  uploaded_at: string
  uploaded_by: string
}

export default function Disputes() {
  const [loading, setLoading] = useState(false)
  const [disputes, setDisputes] = useState<Dispute[]>([
    {
      id: '1',
      dispute_no: 'DSP-2024-0001',
      payment_no: 'PAY-2024-123456',
      order_no: 'ORD-2024-123456',
      merchant_name: 'Tech Store',
      merchant_id: 'MCH-001',
      amount: 99.99,
      currency: 'USD',
      reason: 'Product not received',
      status: 'pending',
      evidence_deadline: '2024-02-10',
      submitted_at: '2024-02-01 10:30:00',
      created_by: 'Customer',
    },
    {
      id: '2',
      dispute_no: 'DSP-2024-0002',
      payment_no: 'PAY-2024-123457',
      order_no: 'ORD-2024-123457',
      merchant_name: 'Fashion Store',
      merchant_id: 'MCH-002',
      amount: 199.99,
      currency: 'USD',
      reason: 'Item significantly not as described',
      status: 'reviewing',
      evidence_deadline: '2024-02-12',
      submitted_at: '2024-02-03 14:20:00',
      resolved_at: '2024-02-05 16:00:00',
      created_by: 'Customer',
    },
  ])

  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedDispute, setSelectedDispute] = useState<Dispute | null>(null)
  const [resolveVisible, setResolveVisible] = useState(false)
  const [form] = Form.useForm()

  // Mock evidence data
  const evidenceList: Evidence[] = [
    {
      id: '1',
      file_name: 'shipping_proof.pdf',
      file_type: 'application/pdf',
      file_url: '#',
      uploaded_at: '2024-02-02 10:00:00',
      uploaded_by: 'Merchant',
    },
    {
      id: '2',
      file_name: 'product_photo.jpg',
      file_type: 'image/jpeg',
      file_url: '#',
      uploaded_at: '2024-02-02 11:30:00',
      uploaded_by: 'Customer',
    },
  ]

  const columns: ColumnsType<Dispute> = [
    {
      title: '争议编号',
      dataIndex: 'dispute_no',
      key: 'dispute_no',
      width: 160,
    },
    {
      title: '支付单号',
      dataIndex: 'payment_no',
      key: 'payment_no',
      width: 180,
    },
    {
      title: '商户名称',
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number, record) => `${record.currency} ${amount.toFixed(2)}`,
    },
    {
      title: '争议原因',
      dataIndex: 'reason',
      key: 'reason',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => {
        const statusConfig = {
          pending: { color: 'orange', text: '待处理' },
          reviewing: { color: 'blue', text: '审核中' },
          accepted: { color: 'green', text: '已接受' },
          rejected: { color: 'red', text: '已拒绝' },
          withdrawn: { color: 'default', text: '已撤回' },
        }
        const config = statusConfig[status as keyof typeof statusConfig]
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '证据截止日期',
      dataIndex: 'evidence_deadline',
      key: 'evidence_deadline',
      width: 140,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD'),
    },
    {
      title: '提交时间',
      dataIndex: 'submitted_at',
      key: 'submitted_at',
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
          {record.status === 'reviewing' && (
            <Button
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => handleResolve(record)}
            >
              处理
            </Button>
          )}
        </Space>
      ),
    },
  ]

  const handleViewDetail = (dispute: Dispute) => {
    setSelectedDispute(dispute)
    setDetailVisible(true)
  }

  const handleResolve = (dispute: Dispute) => {
    setSelectedDispute(dispute)
    setResolveVisible(true)
    form.resetFields()
  }

  const handleResolveSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      // TODO: Call API to resolve dispute
      await new Promise((resolve) => setTimeout(resolve, 1000))

      message.success('争议处理成功')
      setResolveVisible(false)
      form.resetFields()

      // Update dispute status in list
      setDisputes((prev) =>
        prev.map((d) =>
          d.id === selectedDispute?.id
            ? { ...d, status: values.decision === 'accept' ? 'accepted' : 'rejected' }
            : d
        )
      )
    } catch (error) {
      console.error('Resolve dispute error:', error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusSteps = (dispute: Dispute) => {
    const steps = [
      {
        label: '争议提交',
        time: dispute.submitted_at,
        status: 'finish',
      },
    ]

    if (dispute.status === 'reviewing' || dispute.status === 'accepted' || dispute.status === 'rejected') {
      steps.push({
        label: '审核中',
        time: dispute.resolved_at || '-',
        status: 'finish',
      })
    }

    if (dispute.status === 'accepted') {
      steps.push({
        label: '已接受',
        time: dispute.resolved_at || '-',
        status: 'finish',
      })
    } else if (dispute.status === 'rejected') {
      steps.push({
        label: '已拒绝',
        time: dispute.resolved_at || '-',
        status: 'finish',
      })
    }

    return steps
  }

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="总争议数"
              value={disputes.length}
              prefix={<FileTextOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="待处理"
              value={disputes.filter((d) => d.status === 'pending').length}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="审核中"
              value={disputes.filter((d) => d.status === 'reviewing').length}
              valueStyle={{ color: '#1890ff' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="已解决"
              value={
                disputes.filter((d) => d.status === 'accepted' || d.status === 'rejected')
                  .length
              }
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
        </Row>
      </Card>

      <Card
        title="争议管理"
        extra={
          <Space>
            <Input
              placeholder="搜索争议编号/支付单号"
              prefix={<SearchOutlined />}
              style={{ width: 240 }}
            />
            <Select placeholder="状态筛选" style={{ width: 120 }} allowClear>
              <Select.Option value="pending">待处理</Select.Option>
              <Select.Option value="reviewing">审核中</Select.Option>
              <Select.Option value="accepted">已接受</Select.Option>
              <Select.Option value="rejected">已拒绝</Select.Option>
              <Select.Option value="withdrawn">已撤回</Select.Option>
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
          dataSource={disputes}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1500 }}
          pagination={{
            total: disputes.length,
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      {/* Detail Modal */}
      <Modal
        title="争议详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={900}
      >
        {selectedDispute && (
          <Tabs
            items={[
              {
                key: 'info',
                label: '基本信息',
                children: (
                  <Descriptions column={2} bordered>
                    <Descriptions.Item label="争议编号">
                      {selectedDispute.dispute_no}
                    </Descriptions.Item>
                    <Descriptions.Item label="支付单号">
                      {selectedDispute.payment_no}
                    </Descriptions.Item>
                    <Descriptions.Item label="订单号">
                      {selectedDispute.order_no}
                    </Descriptions.Item>
                    <Descriptions.Item label="商户名称">
                      {selectedDispute.merchant_name}
                    </Descriptions.Item>
                    <Descriptions.Item label="金额">
                      {selectedDispute.currency} {selectedDispute.amount.toFixed(2)}
                    </Descriptions.Item>
                    <Descriptions.Item label="状态">
                      <Tag
                        color={
                          selectedDispute.status === 'pending'
                            ? 'orange'
                            : selectedDispute.status === 'reviewing'
                            ? 'blue'
                            : selectedDispute.status === 'accepted'
                            ? 'green'
                            : 'red'
                        }
                      >
                        {selectedDispute.status === 'pending'
                          ? '待处理'
                          : selectedDispute.status === 'reviewing'
                          ? '审核中'
                          : selectedDispute.status === 'accepted'
                          ? '已接受'
                          : '已拒绝'}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="争议原因" span={2}>
                      {selectedDispute.reason}
                    </Descriptions.Item>
                    <Descriptions.Item label="证据截止日期">
                      {dayjs(selectedDispute.evidence_deadline).format('YYYY-MM-DD')}
                    </Descriptions.Item>
                    <Descriptions.Item label="提交时间">
                      {selectedDispute.submitted_at}
                    </Descriptions.Item>
                  </Descriptions>
                ),
              },
              {
                key: 'evidence',
                label: '证据材料',
                children: (
                  <div>
                    <Table
                      columns={[
                        {
                          title: '文件名',
                          dataIndex: 'file_name',
                          key: 'file_name',
                        },
                        {
                          title: '上传者',
                          dataIndex: 'uploaded_by',
                          key: 'uploaded_by',
                        },
                        {
                          title: '上传时间',
                          dataIndex: 'uploaded_at',
                          key: 'uploaded_at',
                        },
                        {
                          title: '操作',
                          key: 'actions',
                          render: (_, record) => (
                            <Button type="link" size="small" href={record.file_url}>
                              下载
                            </Button>
                          ),
                        },
                      ]}
                      dataSource={evidenceList}
                      rowKey="id"
                      pagination={false}
                    />
                  </div>
                ),
              },
              {
                key: 'timeline',
                label: '处理记录',
                children: (
                  <Timeline
                    items={getStatusSteps(selectedDispute).map((step) => ({
                      children: (
                        <div>
                          <div style={{ fontWeight: 'bold' }}>{step.label}</div>
                          <div style={{ fontSize: 12, color: '#999' }}>{step.time}</div>
                        </div>
                      ),
                    }))}
                  />
                ),
              },
            ]}
          />
        )}
      </Modal>

      {/* Resolve Modal */}
      <Modal
        title="处理争议"
        open={resolveVisible}
        onOk={handleResolveSubmit}
        onCancel={() => setResolveVisible(false)}
        confirmLoading={loading}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="decision"
            label="处理决定"
            rules={[{ required: true, message: '请选择处理决定' }]}
          >
            <Select placeholder="选择处理决定">
              <Select.Option value="accept">接受争议（退款给客户）</Select.Option>
              <Select.Option value="reject">拒绝争议（维持原交易）</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="reason"
            label="处理说明"
            rules={[{ required: true, message: '请输入处理说明' }]}
          >
            <TextArea
              rows={4}
              placeholder="请详细说明处理理由..."
              maxLength={500}
              showCount
            />
          </Form.Item>
          <Form.Item name="attachments" label="附件（可选）">
            <Upload>
              <Button icon={<UploadOutlined />}>上传附件</Button>
            </Upload>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

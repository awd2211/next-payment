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
  Alert,
  Steps,
} from 'antd'
import {
  SearchOutlined,
  EyeOutlined,
  UploadOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
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
  amount: number
  currency: string
  reason: string
  status: 'pending' | 'waiting_evidence' | 'under_review' | 'won' | 'lost' | 'withdrawn'
  evidence_deadline: string
  submitted_at: string
  resolved_at?: string
  resolution: string
}

interface Evidence {
  id: string
  file_name: string
  file_type: string
  file_url: string
  uploaded_at: string
}

export default function Disputes() {
  const [loading, setLoading] = useState(false)
  const [disputes, setDisputes] = useState<Dispute[]>([
    {
      id: '1',
      dispute_no: 'DSP-2024-0001',
      payment_no: 'PAY-2024-123456',
      order_no: 'ORD-2024-123456',
      amount: 99.99,
      currency: 'USD',
      reason: 'Product not received',
      status: 'waiting_evidence',
      evidence_deadline: '2024-02-10',
      submitted_at: '2024-02-01 10:30:00',
      resolution: '',
    },
    {
      id: '2',
      dispute_no: 'DSP-2024-0002',
      payment_no: 'PAY-2024-123457',
      order_no: 'ORD-2024-123457',
      amount: 199.99,
      currency: 'USD',
      reason: 'Item significantly not as described',
      status: 'won',
      evidence_deadline: '2024-01-25',
      submitted_at: '2024-01-20 14:20:00',
      resolved_at: '2024-01-28 16:00:00',
      resolution: 'Merchant provided sufficient evidence of delivery and product matching description.',
    },
  ])

  const [detailVisible, setDetailVisible] = useState(false)
  const [evidenceVisible, setEvidenceVisible] = useState(false)
  const [selectedDispute, setSelectedDispute] = useState<Dispute | null>(null)
  const [form] = Form.useForm()

  // Mock evidence list
  const evidenceList: Evidence[] = [
    {
      id: '1',
      file_name: 'shipping_proof.pdf',
      file_type: 'application/pdf',
      file_url: '#',
      uploaded_at: '2024-02-02 10:00:00',
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
      title: '订单号',
      dataIndex: 'order_no',
      key: 'order_no',
      width: 180,
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
      width: 130,
      render: (status: string) => {
        const statusConfig = {
          pending: { color: 'default', text: '待处理' },
          waiting_evidence: { color: 'orange', text: '待提交证据' },
          under_review: { color: 'blue', text: '审核中' },
          won: { color: 'green', text: '争议胜诉' },
          lost: { color: 'red', text: '争议败诉' },
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
      render: (date: string, record) => {
        const isOverdue = dayjs(date).isBefore(dayjs()) && record.status === 'waiting_evidence'
        return (
          <span style={{ color: isOverdue ? '#f5222d' : undefined }}>
            {dayjs(date).format('YYYY-MM-DD')}
            {isOverdue && ' (已逾期)'}
          </span>
        )
      },
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
          {record.status === 'waiting_evidence' && (
            <Button
              type="link"
              size="small"
              icon={<UploadOutlined />}
              onClick={() => handleUploadEvidence(record)}
            >
              上传证据
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

  const handleUploadEvidence = (dispute: Dispute) => {
    setSelectedDispute(dispute)
    setEvidenceVisible(true)
    form.resetFields()
  }

  const handleEvidenceSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      // TODO: Call API to upload evidence
      await new Promise((resolve) => setTimeout(resolve, 1000))

      message.success('证据已提交成功')
      setEvidenceVisible(false)
      form.resetFields()

      // Update dispute status
      setDisputes((prev) =>
        prev.map((d) =>
          d.id === selectedDispute?.id ? { ...d, status: 'under_review' } : d
        )
      )
    } catch (error) {
      console.error('Upload evidence error:', error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusStep = (status: string) => {
    const steps = {
      pending: 0,
      waiting_evidence: 1,
      under_review: 2,
      won: 3,
      lost: 3,
      withdrawn: 3,
    }
    return steps[status as keyof typeof steps] || 0
  }

  return (
    <div>
      <Alert
        message="争议处理提示"
        description="当客户对交易提出争议时，请及时上传相关证据材料（如发货凭证、沟通记录等），在截止日期前提交完整证据将有助于提高胜诉率。"
        type="info"
        showIcon
        closable
        style={{ marginBottom: 16 }}
      />

      <Card
        title="争议处理"
        extra={
          <Space>
            <Input
              placeholder="搜索争议编号/支付单号"
              prefix={<SearchOutlined />}
              style={{ width: 240 }}
            />
            <Select placeholder="状态筛选" style={{ width: 150 }} allowClear>
              <Select.Option value="waiting_evidence">待提交证据</Select.Option>
              <Select.Option value="under_review">审核中</Select.Option>
              <Select.Option value="won">争议胜诉</Select.Option>
              <Select.Option value="lost">争议败诉</Select.Option>
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
          scroll={{ x: 1600 }}
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
          <div>
            <Steps
              current={getStatusStep(selectedDispute.status)}
              status={
                selectedDispute.status === 'lost'
                  ? 'error'
                  : selectedDispute.status === 'won'
                  ? 'finish'
                  : 'process'
              }
              items={[
                { title: '争议提交' },
                { title: '提交证据' },
                { title: '平台审核' },
                {
                  title:
                    selectedDispute.status === 'won'
                      ? '胜诉'
                      : selectedDispute.status === 'lost'
                      ? '败诉'
                      : '结果',
                },
              ]}
              style={{ marginBottom: 24 }}
            />

            <Tabs
              items={[
                {
                  key: 'info',
                  label: '基本信息',
                  children: (
                    <div>
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
                        <Descriptions.Item label="金额">
                          {selectedDispute.currency} {selectedDispute.amount.toFixed(2)}
                        </Descriptions.Item>
                        <Descriptions.Item label="状态">
                          <Tag
                            color={
                              selectedDispute.status === 'waiting_evidence'
                                ? 'orange'
                                : selectedDispute.status === 'under_review'
                                ? 'blue'
                                : selectedDispute.status === 'won'
                                ? 'green'
                                : 'red'
                            }
                          >
                            {selectedDispute.status === 'waiting_evidence'
                              ? '待提交证据'
                              : selectedDispute.status === 'under_review'
                              ? '审核中'
                              : selectedDispute.status === 'won'
                              ? '争议胜诉'
                              : '争议败诉'}
                          </Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label="证据截止日期">
                          {dayjs(selectedDispute.evidence_deadline).format('YYYY-MM-DD')}
                        </Descriptions.Item>
                        <Descriptions.Item label="争议原因" span={2}>
                          {selectedDispute.reason}
                        </Descriptions.Item>
                        <Descriptions.Item label="提交时间">
                          {selectedDispute.submitted_at}
                        </Descriptions.Item>
                        <Descriptions.Item label="解决时间">
                          {selectedDispute.resolved_at || '-'}
                        </Descriptions.Item>
                        {selectedDispute.resolution && (
                          <Descriptions.Item label="处理结果" span={2}>
                            {selectedDispute.resolution}
                          </Descriptions.Item>
                        )}
                      </Descriptions>

                      {selectedDispute.status === 'waiting_evidence' && (
                        <Alert
                          message="请及时上传证据"
                          description={`请在 ${dayjs(selectedDispute.evidence_deadline).format('YYYY-MM-DD')} 前上传相关证据材料，逾期将自动判定为败诉`}
                          type="warning"
                          showIcon
                          icon={<ExclamationCircleOutlined />}
                          style={{ marginTop: 16 }}
                        />
                      )}
                    </div>
                  ),
                },
                {
                  key: 'evidence',
                  label: `已提交证据 (${evidenceList.length})`,
                  children: (
                    <div>
                      {evidenceList.length > 0 ? (
                        <Table
                          columns={[
                            {
                              title: '文件名',
                              dataIndex: 'file_name',
                              key: 'file_name',
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
                                  查看
                                </Button>
                              ),
                            },
                          ]}
                          dataSource={evidenceList}
                          rowKey="id"
                          pagination={false}
                        />
                      ) : (
                        <Alert
                          message="暂无证据材料"
                          description="请点击「上传证据」按钮提交相关证明材料"
                          type="info"
                        />
                      )}
                    </div>
                  ),
                },
                {
                  key: 'guide',
                  label: '证据指南',
                  children: (
                    <div>
                      <Alert
                        message="建议提交的证据类型"
                        type="info"
                        showIcon
                        style={{ marginBottom: 16 }}
                      />
                      <Timeline
                        items={[
                          {
                            children: (
                              <div>
                                <div style={{ fontWeight: 'bold' }}>发货凭证</div>
                                <div style={{ fontSize: 12, color: '#999' }}>
                                  物流单号、快递签收记录、发货时间戳
                                </div>
                              </div>
                            ),
                          },
                          {
                            children: (
                              <div>
                                <div style={{ fontWeight: 'bold' }}>产品描述证明</div>
                                <div style={{ fontSize: 12, color: '#999' }}>
                                  产品图片、详情页截图、产品参数说明
                                </div>
                              </div>
                            ),
                          },
                          {
                            children: (
                              <div>
                                <div style={{ fontWeight: 'bold' }}>沟通记录</div>
                                <div style={{ fontSize: 12, color: '#999' }}>
                                  与客户的聊天记录、邮件往来、电话录音
                                </div>
                              </div>
                            ),
                          },
                          {
                            children: (
                              <div>
                                <div style={{ fontWeight: 'bold' }}>其他辅助材料</div>
                                <div style={{ fontSize: 12, color: '#999' }}>
                                  退换货政策、服务条款、订单确认邮件
                                </div>
                              </div>
                            ),
                          },
                        ]}
                      />
                    </div>
                  ),
                },
              ]}
            />
          </div>
        )}
      </Modal>

      {/* Upload Evidence Modal */}
      <Modal
        title="上传证据材料"
        open={evidenceVisible}
        onOk={handleEvidenceSubmit}
        onCancel={() => setEvidenceVisible(false)}
        confirmLoading={loading}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Alert
            message="证据提交说明"
            description="请上传清晰可辨的证据材料，支持图片、PDF等格式，单个文件不超过10MB"
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

          <Form.Item
            name="evidence_description"
            label="证据说明"
            rules={[{ required: true, message: '请输入证据说明' }]}
          >
            <TextArea
              rows={4}
              placeholder="请简要说明证据内容和证明目的..."
              maxLength={500}
              showCount
            />
          </Form.Item>

          <Form.Item
            name="files"
            label="上传文件"
            rules={[{ required: true, message: '请上传至少一个证据文件' }]}
          >
            <Upload.Dragger multiple maxCount={10}>
              <p className="ant-upload-drag-icon">
                <FileTextOutlined />
              </p>
              <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
              <p className="ant-upload-hint">
                支持单个或批量上传，最多10个文件，每个文件不超过10MB
              </p>
            </Upload.Dragger>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Image, Space, message, Form, Input, Select } from 'antd'
import { EyeOutlined, CheckOutlined, CloseOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { kycService, type KYCApplication } from '../services/kycService'

export default function KYC() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<KYCApplication[]>([])
  const [selectedRecord, setSelectedRecord] = useState<KYCApplication | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)
  const [rejectVisible, setRejectVisible] = useState(false)
  const [rejectForm] = Form.useForm()

  // Mock data - 替换为实际API调用
  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    setLoading(true)
    try {
      const response = await kycService.list({ page: 1, page_size: 20 })
      // 响应拦截器已解包，直接使用数据
      if (response && response.list) {
        setData(response.list)
      }
    } catch (error) {
      // 错误已被拦截器处理并显示
      console.error('Failed to fetch KYC applications:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleView = (record: KYCApplication) => {
    setSelectedRecord(record)
    setDetailVisible(true)
  }

  const handleApprove = async (record: KYCApplication) => {
    Modal.confirm({
      title: '确认通过KYC审核?',
      content: `商户: ${record.merchant_name}`,
      onOk: async () => {
        try {
          const response = await kycService.approve(record.id, {})
          if (response.code === 0) {
            message.success('KYC审核通过')
            fetchData()
          } else {
            message.error(response.error?.message || '操作失败')
          }
        } catch (error) {
          message.error('操作失败')
          console.error('Failed to approve KYC:', error)
        }
      },
    })
  }

  const handleReject = (record: KYCApplication) => {
    setSelectedRecord(record)
    setRejectVisible(true)
  }

  const handleRejectSubmit = async (values: any) => {
    try {
      const response = await kycService.reject(selectedRecord!.id, { reason: values.reason })
      if (response.code === 0) {
        message.success('已拒绝KYC申请')
        setRejectVisible(false)
        rejectForm.resetFields()
        fetchData()
      } else {
        message.error(response.error?.message || '操作失败')
      }
    } catch (error) {
      message.error('操作失败')
      console.error('Failed to reject KYC:', error)
    }
  }

  const columns: ColumnsType<KYCApplication> = [
    {
      title: '商户ID',
      dataIndex: 'merchant_id',
      width: 120,
    },
    {
      title: '商户名称',
      dataIndex: 'merchant_name',
      width: 200,
    },
    {
      title: '业务类型',
      dataIndex: 'business_type',
      width: 120,
    },
    {
      title: '法人姓名',
      dataIndex: 'legal_name',
      width: 120,
    },
    {
      title: '注册号',
      dataIndex: 'registration_number',
      width: 200,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (status: string) => {
        const colorMap = {
          pending: 'orange',
          reviewing: 'blue',
          approved: 'green',
          rejected: 'red',
        }
        const textMap = {
          pending: '待审核',
          reviewing: '审核中',
          approved: '已通过',
          rejected: '已拒绝',
        }
        return <Tag color={colorMap[status as keyof typeof colorMap]}>{textMap[status as keyof typeof textMap]}</Tag>
      },
    },
    {
      title: '提交时间',
      dataIndex: 'submitted_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 250,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => handleView(record)}
          >
            查看
          </Button>
          {record.status === 'pending' && (
            <>
              <Button
                type="primary"
                icon={<CheckOutlined />}
                size="small"
                onClick={() => handleApprove(record)}
              >
                通过
              </Button>
              <Button
                danger
                icon={<CloseOutlined />}
                size="small"
                onClick={() => handleReject(record)}
              >
                拒绝
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="KYC审核管理"
        extra={
          <Space>
            <Select
              placeholder="状态筛选"
              style={{ width: 120 }}
              allowClear
              options={[
                { label: '待审核', value: 'pending' },
                { label: '审核中', value: 'reviewing' },
                { label: '已通过', value: 'approved' },
                { label: '已拒绝', value: 'rejected' },
              ]}
            />
            <Button icon={<SearchOutlined />} onClick={fetchData}>
              刷新
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1400 }}
          pagination={{
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      {/* 详情Modal */}
      <Modal
        title="KYC申请详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        width={800}
        footer={null}
      >
        {selectedRecord && (
          <div>
            <Card title="基本信息" size="small" style={{ marginBottom: 16 }}>
              <p><strong>商户名称:</strong> {selectedRecord.merchant_name}</p>
              <p><strong>业务类型:</strong> {selectedRecord.business_type}</p>
              <p><strong>法人姓名:</strong> {selectedRecord.legal_name}</p>
              <p><strong>注册号:</strong> {selectedRecord.registration_number}</p>
              <p><strong>提交时间:</strong> {selectedRecord.submitted_at}</p>
            </Card>

            <Card title="KYC文档" size="small">
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <strong>营业执照:</strong>
                  <br />
                  <Image
                    width={200}
                    src={selectedRecord.documents.business_license}
                    placeholder
                  />
                </div>
                <div>
                  <strong>身份证正面:</strong>
                  <br />
                  <Image
                    width={200}
                    src={selectedRecord.documents.id_card_front}
                    placeholder
                  />
                </div>
                <div>
                  <strong>身份证背面:</strong>
                  <br />
                  <Image
                    width={200}
                    src={selectedRecord.documents.id_card_back}
                    placeholder
                  />
                </div>
                <div>
                  <strong>银行对账单:</strong>
                  <br />
                  <Image
                    width={200}
                    src={selectedRecord.documents.bank_statement}
                    placeholder
                  />
                </div>
              </Space>
            </Card>

            {selectedRecord.status === 'pending' && (
              <div style={{ marginTop: 16, textAlign: 'right' }}>
                <Space>
                  <Button
                    type="primary"
                    icon={<CheckOutlined />}
                    onClick={() => {
                      setDetailVisible(false)
                      handleApprove(selectedRecord)
                    }}
                  >
                    通过审核
                  </Button>
                  <Button
                    danger
                    icon={<CloseOutlined />}
                    onClick={() => {
                      setDetailVisible(false)
                      handleReject(selectedRecord)
                    }}
                  >
                    拒绝申请
                  </Button>
                </Space>
              </div>
            )}
          </div>
        )}
      </Modal>

      {/* 拒绝Modal */}
      <Modal
        title="拒绝KYC申请"
        open={rejectVisible}
        onCancel={() => {
          setRejectVisible(false)
          rejectForm.resetFields()
        }}
        onOk={() => rejectForm.submit()}
      >
        <Form form={rejectForm} onFinish={handleRejectSubmit} layout="vertical">
          <Form.Item
            name="reason"
            label="拒绝原因"
            rules={[{ required: true, message: '请输入拒绝原因' }]}
          >
            <Input.TextArea rows={4} placeholder="请说明拒绝的具体原因..." />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

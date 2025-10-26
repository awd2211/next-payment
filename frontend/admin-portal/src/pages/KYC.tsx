import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Space, message, Form, Input, Tabs } from 'antd'
import { EyeOutlined, CheckOutlined, CloseOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { kycService, type BusinessQualification } from '../services/kycService'

const { TabPane } = Tabs

export default function KYC() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<BusinessQualification[]>([])
  const [selectedRecord, setSelectedRecord] = useState<BusinessQualification | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)
  const [rejectVisible, setRejectVisible] = useState(false)
  const [rejectForm] = Form.useForm()
  const [activeTab, setActiveTab] = useState('pending')

  useEffect(() => {
    fetchData()
  }, [activeTab])

  const fetchData = async () => {
    setLoading(true)
    try {
      const response = await kycService.listQualifications({
        page: 1,
        page_size: 20,
        status: activeTab === 'all' ? undefined : activeTab
      })
      // 响应拦截器已解包，直接使用数据
      if (response && response.qualifications) {
        setData(response.qualifications)
      }
    } catch (error) {
      // 错误已被拦截器处理并显示
      console.error('Failed to fetch KYC qualifications:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleView = (record: BusinessQualification) => {
    setSelectedRecord(record)
    setDetailVisible(true)
  }

  const handleApprove = async (record: BusinessQualification) => {
    Modal.confirm({
      title: '确认通过资质审核?',
      content: `公司: ${record.company_name}`,
      onOk: async () => {
        try {
          await kycService.approveQualification(record.id)
          message.success('资质审核通过')
          fetchData()
        } catch (error) {
          message.error('操作失败')
          console.error('Failed to approve qualification:', error)
        }
      },
    })
  }

  const handleReject = (record: BusinessQualification) => {
    setSelectedRecord(record)
    setRejectVisible(true)
  }

  const handleRejectSubmit = async (values: any) => {
    try {
      await kycService.rejectQualification(selectedRecord!.id, values.reason, values.remark)
      message.success('已拒绝资质申请')
      setRejectVisible(false)
      rejectForm.resetFields()
      fetchData()
    } catch (error) {
      message.error('操作失败')
      console.error('Failed to reject qualification:', error)
    }
  }

  const columns: ColumnsType<BusinessQualification> = [
    {
      title: '商户ID',
      dataIndex: 'merchant_id',
      width: 200,
      ellipsis: true,
    },
    {
      title: '公司名称',
      dataIndex: 'company_name',
      width: 200,
    },
    {
      title: '法人姓名',
      dataIndex: 'legal_person_name',
      width: 120,
    },
    {
      title: '营业执照号',
      dataIndex: 'business_license_no',
      width: 180,
    },
    {
      title: '行业',
      dataIndex: 'industry',
      width: 120,
    },
    {
      title: '注册资本',
      dataIndex: 'registered_capital',
      width: 120,
      render: (value: number) => {
        return value ? `¥${(value / 100).toLocaleString()}` : '-'
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
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
      dataIndex: 'created_at',
      width: 180,
      render: (text: string) => text ? new Date(text).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button icon={<EyeOutlined />} size="small" onClick={() => handleView(record)}>
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
      <Card title="KYC 资质审核管理" extra={
        <Button icon={<SearchOutlined />} onClick={fetchData}>
          刷新
        </Button>
      }>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="待审核" key="pending" />
          <TabPane tab="审核中" key="reviewing" />
          <TabPane tab="已通过" key="approved" />
          <TabPane tab="已拒绝" key="rejected" />
          <TabPane tab="全部" key="all" />
        </Tabs>

        <Table
          columns={columns}
          dataSource={data}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          scroll={{ x: 1500 }}
        />
      </Card>

      {/* 详情模态框 */}
      <Modal
        title="资质详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
      >
        {selectedRecord && (
          <div>
            <h3>基本信息</h3>
            <p><strong>商户ID:</strong> {selectedRecord.merchant_id}</p>
            <p><strong>公司名称:</strong> {selectedRecord.company_name}</p>
            <p><strong>营业执照号:</strong> {selectedRecord.business_license_no}</p>
            <p><strong>法人姓名:</strong> {selectedRecord.legal_person_name}</p>
            <p><strong>法人身份证:</strong> {selectedRecord.legal_person_id_card}</p>
            <p><strong>注册地址:</strong> {selectedRecord.registered_address || '-'}</p>
            <p><strong>注册资本:</strong> ¥{(selectedRecord.registered_capital / 100).toLocaleString()}</p>
            <p><strong>成立日期:</strong> {selectedRecord.established_date || '-'}</p>
            <p><strong>经营范围:</strong> {selectedRecord.business_scope || '-'}</p>
            <p><strong>行业:</strong> {selectedRecord.industry || '-'}</p>
            <p><strong>税务登记号:</strong> {selectedRecord.tax_registration_no || '-'}</p>
            <p><strong>组织机构代码:</strong> {selectedRecord.organization_code || '-'}</p>

            <h3>证件照片</h3>
            {selectedRecord.business_license_url && (
              <div>
                <p><strong>营业执照:</strong></p>
                <img src={selectedRecord.business_license_url} alt="营业执照" style={{ maxWidth: '100%', marginBottom: 16 }} />
              </div>
            )}
            {selectedRecord.legal_person_id_card_front_url && (
              <div>
                <p><strong>法人身份证正面:</strong></p>
                <img src={selectedRecord.legal_person_id_card_front_url} alt="身份证正面" style={{ maxWidth: '100%', marginBottom: 16 }} />
              </div>
            )}
            {selectedRecord.legal_person_id_card_back_url && (
              <div>
                <p><strong>法人身份证背面:</strong></p>
                <img src={selectedRecord.legal_person_id_card_back_url} alt="身份证背面" style={{ maxWidth: '100%', marginBottom: 16 }} />
              </div>
            )}
            {selectedRecord.tax_registration_url && (
              <div>
                <p><strong>税务登记证:</strong></p>
                <img src={selectedRecord.tax_registration_url} alt="税务登记证" style={{ maxWidth: '100%', marginBottom: 16 }} />
              </div>
            )}

            {selectedRecord.status === 'rejected' && (
              <div>
                <h3>拒绝原因</h3>
                <p style={{ color: 'red' }}>{selectedRecord.reject_reason}</p>
                {selectedRecord.remark && <p><strong>备注:</strong> {selectedRecord.remark}</p>}
              </div>
            )}

            {selectedRecord.reviewed_at && (
              <div>
                <h3>审核信息</h3>
                <p><strong>审核人:</strong> {selectedRecord.reviewer_name || '-'}</p>
                <p><strong>审核时间:</strong> {new Date(selectedRecord.reviewed_at).toLocaleString()}</p>
              </div>
            )}
          </div>
        )}
      </Modal>

      {/* 拒绝模态框 */}
      <Modal
        title="拒绝资质申请"
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
            <Input.TextArea rows={4} placeholder="请详细说明拒绝的原因" />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} placeholder="可选:补充说明" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

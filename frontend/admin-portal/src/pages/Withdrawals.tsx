import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Descriptions, Space, message, Form, Input, InputNumber } from 'antd'
import { EyeOutlined, CheckOutlined, CloseOutlined, DollarOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { withdrawalService, type Withdrawal } from '../services/withdrawalService'

export default function Withdrawals() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<Withdrawal[]>([])
  const [selectedRecord, setSelectedRecord] = useState<Withdrawal | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)
  const [approveVisible, setApproveVisible] = useState(false)
  const [rejectVisible, setRejectVisible] = useState(false)
  const [approveForm] = Form.useForm()
  const [rejectForm] = Form.useForm()

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    setLoading(true)
    try {
      const response = await withdrawalService.list({ page: 1, page_size: 20 })
      if (response.code === 0 && response.data) {
        setData(response.data.list)
      } else {
        message.error(response.error?.message || '加载失败')
      }
    } catch (error) {
      message.error('加载失败')
      console.error('Failed to fetch withdrawals:', error)
    } finally {
      setLoading(false)
    }
  }

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${(amount / 100).toFixed(2)}`
  }

  const handleView = (record: Withdrawal) => {
    setSelectedRecord(record)
    setDetailVisible(true)
  }

  const handleApprove = (record: Withdrawal) => {
    setSelectedRecord(record)
    approveForm.setFieldsValue({
      actual_amount: record.actual_amount / 100,
      remark: '',
    })
    setApproveVisible(true)
  }

  const handleReject = (record: Withdrawal) => {
    setSelectedRecord(record)
    setRejectVisible(true)
  }

  const handleApproveSubmit = async (values: any) => {
    try {
      const response = await withdrawalService.approve(selectedRecord!.id, { remark: values.remark })
      if (response.code === 0) {
        message.success('提现申请已批准')
        setApproveVisible(false)
        approveForm.resetFields()
        fetchData()
      } else {
        message.error(response.error?.message || '操作失败')
      }
    } catch (error) {
      message.error('操作失败')
      console.error('Failed to approve withdrawal:', error)
    }
  }

  const handleRejectSubmit = async (values: any) => {
    try {
      const response = await withdrawalService.reject(selectedRecord!.id, { reason: values.reason })
      if (response.code === 0) {
        message.success('已拒绝提现申请')
        setRejectVisible(false)
        rejectForm.resetFields()
        fetchData()
      } else {
        message.error(response.error?.message || '操作失败')
      }
    } catch (error) {
      message.error('操作失败')
      console.error('Failed to reject withdrawal:', error)
    }
  }

  const columns: ColumnsType<Withdrawal> = [
    {
      title: '提现单号',
      dataIndex: 'withdrawal_no',
      width: 180,
      fixed: 'left',
    },
    {
      title: '商户名称',
      dataIndex: 'merchant_name',
      width: 150,
    },
    {
      title: '提现金额',
      dataIndex: 'amount',
      width: 120,
      render: (amount, record) => formatAmount(amount, record.currency),
    },
    {
      title: '手续费',
      dataIndex: 'fee',
      width: 100,
      render: (fee, record) => formatAmount(fee, record.currency),
    },
    {
      title: '实际到账',
      dataIndex: 'actual_amount',
      width: 120,
      render: (amount, record) => (
        <strong>{formatAmount(amount, record.currency)}</strong>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (status: string) => {
        const colorMap = {
          pending: 'orange',
          approved: 'blue',
          processing: 'cyan',
          completed: 'green',
          rejected: 'red',
          failed: 'volcano',
        }
        const textMap = {
          pending: '待审批',
          approved: '已批准',
          processing: '处理中',
          completed: '已完成',
          rejected: '已拒绝',
          failed: '失败',
        }
        return <Tag color={colorMap[status as keyof typeof colorMap]}>{textMap[status as keyof typeof textMap]}</Tag>
      },
    },
    {
      title: '银行',
      dataIndex: 'bank_name',
      width: 150,
    },
    {
      title: '账号',
      dataIndex: 'bank_account',
      width: 180,
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
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
                批准
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
      <Card title="提现管理" extra={<Button icon={<DollarOutlined />} onClick={fetchData}>刷新</Button>}>
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1600 }}
          pagination={{
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      {/* 详情Modal */}
      <Modal
        title="提现申请详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        width={700}
        footer={null}
      >
        {selectedRecord && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="提现单号" span={2}>
              {selectedRecord.withdrawal_no}
            </Descriptions.Item>
            <Descriptions.Item label="商户名称">
              {selectedRecord.merchant_name}
            </Descriptions.Item>
            <Descriptions.Item label="商户ID">
              {selectedRecord.merchant_id}
            </Descriptions.Item>
            <Descriptions.Item label="提现金额">
              {formatAmount(selectedRecord.amount, selectedRecord.currency)}
            </Descriptions.Item>
            <Descriptions.Item label="手续费">
              {formatAmount(selectedRecord.fee, selectedRecord.currency)}
            </Descriptions.Item>
            <Descriptions.Item label="实际到账" span={2}>
              <strong style={{ fontSize: 16 }}>
                {formatAmount(selectedRecord.actual_amount, selectedRecord.currency)}
              </strong>
            </Descriptions.Item>
            <Descriptions.Item label="状态" span={2}>
              <Tag color={selectedRecord.status === 'completed' ? 'green' : 'orange'}>
                {selectedRecord.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="银行名称">
              {selectedRecord.bank_account.bank_name}
            </Descriptions.Item>
            <Descriptions.Item label="银行账号">
              {selectedRecord.bank_account.account_number}
            </Descriptions.Item>
            <Descriptions.Item label="账户名" span={2}>
              {selectedRecord.bank_account.account_name}
            </Descriptions.Item>
            <Descriptions.Item label="申请时间" span={2}>
              {selectedRecord.created_at}
            </Descriptions.Item>
            {selectedRecord.approved_at && (
              <Descriptions.Item label="批准时间" span={2}>
                {selectedRecord.approved_at}
              </Descriptions.Item>
            )}
            {selectedRecord.completed_at && (
              <Descriptions.Item label="完成时间" span={2}>
                {selectedRecord.completed_at}
              </Descriptions.Item>
            )}
            {selectedRecord.remark && (
              <Descriptions.Item label="备注" span={2}>
                {selectedRecord.remark}
              </Descriptions.Item>
            )}
            {selectedRecord.reject_reason && (
              <Descriptions.Item label="拒绝原因" span={2}>
                <span style={{ color: 'red' }}>{selectedRecord.reject_reason}</span>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      {/* 批准Modal */}
      <Modal
        title="批准提现申请"
        open={approveVisible}
        onCancel={() => {
          setApproveVisible(false)
          approveForm.resetFields()
        }}
        onOk={() => approveForm.submit()}
      >
        <Form form={approveForm} onFinish={handleApproveSubmit} layout="vertical">
          <Form.Item label="实际到账金额" name="actual_amount">
            <InputNumber
              style={{ width: '100%' }}
              precision={2}
              addonAfter={selectedRecord?.currency}
              disabled
            />
          </Form.Item>
          <Form.Item label="备注" name="remark">
            <Input.TextArea rows={3} placeholder="可选:添加备注信息..." />
          </Form.Item>
        </Form>
      </Modal>

      {/* 拒绝Modal */}
      <Modal
        title="拒绝提现申请"
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

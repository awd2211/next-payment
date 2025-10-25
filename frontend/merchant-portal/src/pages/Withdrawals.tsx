import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Form, Input, InputNumber, Select, Space, message, Descriptions, Steps } from 'antd'
import { PlusOutlined, EyeOutlined, DollarOutlined, BankOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

interface BankAccount {
  id: string
  bank_name: string
  bank_code: string
  account_name: string
  account_number: string
  swift_code?: string
  is_default: boolean
}

interface Withdrawal {
  id: string
  withdrawal_no: string
  amount: number
  currency: string
  fee: number
  actual_amount: number
  status: 'pending' | 'approved' | 'rejected' | 'processing' | 'completed' | 'failed'
  bank_account: BankAccount
  remark?: string
  reject_reason?: string
  approved_at?: string
  completed_at?: string
  created_at: string
}

export default function Withdrawals() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<Withdrawal[]>([])
  const [bankAccounts, setBankAccounts] = useState<BankAccount[]>([])
  const [applyModalVisible, setApplyModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [selectedRecord, setSelectedRecord] = useState<Withdrawal | null>(null)
  const [form] = Form.useForm()

  // 可提现余额 (Mock)
  const availableBalance = 125800 // 分,1258.00 USD

  useEffect(() => {
    fetchData()
    fetchBankAccounts()
  }, [])

  const fetchData = async () => {
    setLoading(true)
    // TODO: 调用 withdrawalService.list()
    setTimeout(() => {
      setData([
        {
          id: '1',
          withdrawal_no: 'WD202510250001',
          amount: 50000,
          currency: 'USD',
          fee: 100,
          actual_amount: 49900,
          status: 'completed',
          bank_account: {
            id: '1',
            bank_name: '中国工商银行',
            bank_code: 'ICBC',
            account_name: '张三',
            account_number: '6222 **** **** 1234',
            is_default: true,
          },
          approved_at: '2025-10-24 10:00:00',
          completed_at: '2025-10-25 09:00:00',
          created_at: '2025-10-24 09:30:00',
        },
        {
          id: '2',
          withdrawal_no: 'WD202510240001',
          amount: 30000,
          currency: 'USD',
          fee: 60,
          actual_amount: 29940,
          status: 'pending',
          bank_account: {
            id: '1',
            bank_name: '中国工商银行',
            bank_code: 'ICBC',
            account_name: '张三',
            account_number: '6222 **** **** 1234',
            is_default: true,
          },
          created_at: '2025-10-24 15:30:00',
        },
      ])
      setLoading(false)
    }, 500)
  }

  const fetchBankAccounts = async () => {
    // TODO: 调用 withdrawalService.getBankAccounts()
    setBankAccounts([
      {
        id: '1',
        bank_name: '中国工商银行',
        bank_code: 'ICBC',
        account_name: '张三',
        account_number: '6222 0000 0000 1234',
        swift_code: 'ICBKCNBJ',
        is_default: true,
      },
    ])
  }

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${(amount / 100).toFixed(2)}`
  }

  const handleApply = () => {
    if (bankAccounts.length === 0) {
      message.warning('请先添加银行账户')
      return
    }
    form.resetFields()
    form.setFieldsValue({
      bank_account_id: bankAccounts.find((b) => b.is_default)?.id,
      currency: 'USD',
    })
    setApplyModalVisible(true)
  }

  const handleView = (record: Withdrawal) => {
    setSelectedRecord(record)
    setDetailModalVisible(true)
  }

  const handleSubmit = async (_values: any) => {
    try {
      // TODO: 调用 withdrawalService.create(values)
      message.success('提现申请已提交,等待审核')
      setApplyModalVisible(false)
      form.resetFields()
      fetchData()
    } catch (error) {
      message.error('提交失败')
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
        <strong style={{ color: '#52c41a' }}>{formatAmount(amount, record.currency)}</strong>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (status: string) => {
        const colorMap: Record<string, string> = {
          pending: 'orange',
          approved: 'blue',
          processing: 'cyan',
          completed: 'green',
          rejected: 'red',
          failed: 'volcano',
        }
        const textMap: Record<string, string> = {
          pending: '待审核',
          approved: '已批准',
          processing: '处理中',
          completed: '已完成',
          rejected: '已拒绝',
          failed: '失败',
        }
        return <Tag color={colorMap[status]}>{textMap[status]}</Tag>
      },
    },
    {
      title: '收款银行',
      dataIndex: ['bank_account', 'bank_name'],
      width: 150,
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Button type="link" icon={<EyeOutlined />} onClick={() => handleView(record)}>
          查看
        </Button>
      ),
    },
  ]

  return (
    <div>
      {/* 余额卡片 */}
      <Card style={{ marginBottom: 16 }}>
        <Space size="large">
          <div>
            <div style={{ color: '#999', marginBottom: 8 }}>可提现余额</div>
            <div style={{ fontSize: 32, fontWeight: 'bold', color: '#52c41a' }}>
              <DollarOutlined /> {(availableBalance / 100).toFixed(2)} USD
            </div>
          </div>
          <Button type="primary" size="large" icon={<PlusOutlined />} onClick={handleApply}>
            申请提现
          </Button>
        </Space>
      </Card>

      {/* 提现记录 */}
      <Card title="提现记录">
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1200 }}
        />
      </Card>

      {/* 申请提现Modal */}
      <Modal
        title="申请提现"
        open={applyModalVisible}
        onCancel={() => {
          setApplyModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item
            name="bank_account_id"
            label="收款银行账户"
            rules={[{ required: true, message: '请选择收款账户' }]}
          >
            <Select>
              {bankAccounts.map((account) => (
                <Select.Option key={account.id} value={account.id}>
                  <BankOutlined /> {account.bank_name} - {account.account_number}
                  {account.is_default && <Tag color="blue" style={{ marginLeft: 8 }}>默认</Tag>}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="amount"
            label="提现金额"
            rules={[
              { required: true, message: '请输入提现金额' },
              {
                validator: (_, value) => {
                  if (value && value * 100 > availableBalance) {
                    return Promise.reject('提现金额不能超过可用余额')
                  }
                  return Promise.resolve()
                },
              },
            ]}
            tooltip={`可提现余额: ${(availableBalance / 100).toFixed(2)} USD`}
          >
            <InputNumber
              style={{ width: '100%' }}
              min={1}
              precision={2}
              addonAfter="USD"
              placeholder="请输入金额"
            />
          </Form.Item>

          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={3} placeholder="可选" />
          </Form.Item>

          <Card size="small" title="费用说明" style={{ backgroundColor: '#f5f5f5' }}>
            <p>提现手续费: 2 USD/笔</p>
            <p>到账时间: 1-3个工作日</p>
            <p>单笔限额: 10 - 100,000 USD</p>
          </Card>
        </Form>
      </Modal>

      {/* 详情Modal */}
      <Modal
        title="提现详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={700}
      >
        {selectedRecord && (
          <>
            <Descriptions bordered column={2} style={{ marginBottom: 24 }}>
              <Descriptions.Item label="提现单号" span={2}>
                {selectedRecord.withdrawal_no}
              </Descriptions.Item>
              <Descriptions.Item label="提现金额">
                {formatAmount(selectedRecord.amount, selectedRecord.currency)}
              </Descriptions.Item>
              <Descriptions.Item label="手续费">
                {formatAmount(selectedRecord.fee, selectedRecord.currency)}
              </Descriptions.Item>
              <Descriptions.Item label="实际到账" span={2}>
                <strong style={{ fontSize: 16, color: '#52c41a' }}>
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
              <Descriptions.Item label="账户名">
                {selectedRecord.bank_account.account_name}
              </Descriptions.Item>
              <Descriptions.Item label="银行账号" span={2}>
                {selectedRecord.bank_account.account_number}
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
              {selectedRecord.reject_reason && (
                <Descriptions.Item label="拒绝原因" span={2}>
                  <span style={{ color: 'red' }}>{selectedRecord.reject_reason}</span>
                </Descriptions.Item>
              )}
            </Descriptions>

            {/* 进度条 */}
            <Steps
              current={
                selectedRecord.status === 'pending'
                  ? 0
                  : selectedRecord.status === 'approved'
                  ? 1
                  : selectedRecord.status === 'processing'
                  ? 2
                  : selectedRecord.status === 'completed'
                  ? 3
                  : 0
              }
              status={selectedRecord.status === 'rejected' || selectedRecord.status === 'failed' ? 'error' : 'process'}
              items={[
                { title: '申请提交' },
                { title: '等待审核' },
                { title: '处理中' },
                { title: '完成' },
              ]}
            />
          </>
        )}
      </Modal>
    </div>
  )
}

import { useState, useEffect } from 'react'
import {
  Typography,
  Button,
  Form,
  Input,
  Select,
  Space,
  Tag,
  message,
  Dropdown,
  Card,
  Row,
  Col,
  Statistic,
  Modal,
  Table,
  Popconfirm,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckOutlined,
  CloseOutlined,
  LockOutlined,
  UnlockOutlined,
  MoreOutlined,
  SafetyOutlined,
  ShopOutlined,
  UserOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { MenuProps } from 'antd'
import { merchantService, Merchant, CreateMerchantRequest, UpdateMerchantRequest } from '../services/merchantService'
import dayjs from 'dayjs'

const { Title } = Typography

const Merchants = () => {
  const [loading, setLoading] = useState(false)
  const [merchants, setMerchants] = useState<Merchant[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingMerchant, setEditingMerchant] = useState<Merchant | null>(null)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [kycStatusFilter, setKycStatusFilter] = useState<string | undefined>()
  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    active: 0,
    suspended: 0,
  })
  const [form] = Form.useForm()

  useEffect(() => {
    loadMerchants()
  }, [page, pageSize, searchKeyword, statusFilter, kycStatusFilter])

  useEffect(() => {
    calculateStats()
  }, [merchants])

  const loadMerchants = async () => {
    setLoading(true)
    try {
      const response = await merchantService.list({
        page,
        page_size: pageSize,
        keyword: searchKeyword,
        status: statusFilter,
        kyc_status: kycStatusFilter,
      })
      if (response?.data?.data) {
        setMerchants(response.data.data.list || [])
        setTotal(response.data.data.total || 0)
      }
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const calculateStats = () => {
    setStats({
      total: merchants.length,
      pending: merchants.filter(m => m.status === 'pending').length,
      active: merchants.filter(m => m.status === 'active').length,
      suspended: merchants.filter(m => m.status === 'suspended').length,
    })
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'orange',
      active: 'green',
      suspended: 'red',
      rejected: 'default',
    }
    return colors[status] || 'default'
  }

  const getStatusText = (status: string) => {
    const texts: Record<string, string> = {
      pending: '待审核',
      active: '正常',
      suspended: '已冻结',
      rejected: '已拒绝',
    }
    return texts[status] || status
  }

  const getKYCStatusColor = (kycStatus: string) => {
    const colors: Record<string, string> = {
      pending: 'orange',
      verified: 'green',
      rejected: 'red',
    }
    return colors[kycStatus] || 'default'
  }

  const getKYCStatusText = (kycStatus: string) => {
    const texts: Record<string, string> = {
      pending: '待审核',
      verified: '已认证',
      rejected: '已拒绝',
    }
    return texts[kycStatus] || kycStatus
  }

  const handleApprove = async (id: string) => {
    try {
      await merchantService.updateStatus(id, 'active')
      message.success('审批通过')
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleReject = async (id: string) => {
    try {
      await merchantService.updateStatus(id, 'rejected')
      message.success('已拒绝')
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleSuspend = async (id: string) => {
    try {
      await merchantService.updateStatus(id, 'suspended')
      message.success('已冻结')
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleUnsuspend = async (id: string) => {
    try {
      await merchantService.updateStatus(id, 'active')
      message.success('已解冻')
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleVerifyKYC = async (id: string) => {
    try {
      await merchantService.updateKYCStatus(id, 'verified')
      message.success('KYC已认证')
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleRejectKYC = async (id: string) => {
    try {
      await merchantService.updateKYCStatus(id, 'rejected')
      message.success('KYC已拒绝')
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const getActionMenu = (record: Merchant): MenuProps => ({
    items: [
      record.status === 'pending' && {
        key: 'approve',
        icon: <CheckOutlined />,
        label: '审批通过',
        onClick: () => handleApprove(record.id),
      },
      record.status === 'pending' && {
        key: 'reject',
        icon: <CloseOutlined />,
        label: '拒绝申请',
        onClick: () => handleReject(record.id),
        danger: true,
      },
      record.status === 'active' && {
        key: 'suspend',
        icon: <LockOutlined />,
        label: '冻结账户',
        onClick: () => handleSuspend(record.id),
        danger: true,
      },
      record.status === 'suspended' && {
        key: 'unsuspend',
        icon: <UnlockOutlined />,
        label: '解冻账户',
        onClick: () => handleUnsuspend(record.id),
      },
      record.kyc_status === 'pending' && {
        key: 'verify-kyc',
        icon: <SafetyOutlined />,
        label: 'KYC认证通过',
        onClick: () => handleVerifyKYC(record.id),
      },
      record.kyc_status === 'pending' && {
        key: 'reject-kyc',
        icon: <CloseOutlined />,
        label: 'KYC认证拒绝',
        onClick: () => handleRejectKYC(record.id),
        danger: true,
      },
    ].filter(Boolean) as MenuProps['items'],
  })

  const columns: ColumnsType<Merchant> = [
    {
      title: '商户名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      width: 200,
      ellipsis: true,
    },
    {
      title: '公司名称',
      dataIndex: 'company_name',
      key: 'company_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: '业务类型',
      dataIndex: 'business_type',
      key: 'business_type',
      width: 100,
      render: (type: string) => (
        <Tag>{type === 'individual' ? '个人' : '公司'}</Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: 'KYC状态',
      dataIndex: 'kyc_status',
      key: 'kyc_status',
      width: 100,
      render: (kycStatus: string) => (
        <Tag color={getKYCStatusColor(kycStatus)}>{getKYCStatusText(kycStatus)}</Tag>
      ),
    },
    {
      title: '测试模式',
      dataIndex: 'is_test_mode',
      key: 'is_test_mode',
      width: 100,
      render: (isTestMode: boolean) => (
        isTestMode ? <Tag color="blue">测试</Tag> : <Tag color="green">生产</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
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
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Dropdown menu={getActionMenu(record)} trigger={['click']}>
            <Button type="link" size="small" icon={<MoreOutlined />}>
              更多
            </Button>
          </Dropdown>
          {record.status !== 'active' && (
            <Popconfirm
              title="确认删除"
              description={`确定要删除商户 "${record.name}" 吗？`}
              onConfirm={() => handleDelete(record.id)}
            >
              <Button
                type="link"
                size="small"
                danger
                icon={<DeleteOutlined />}
              >
                删除
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  const handleAdd = () => {
    form.resetFields()
    setEditingMerchant(null)
    setModalVisible(true)
  }

  const handleEdit = (merchant: Merchant) => {
    form.setFieldsValue({
      name: merchant.name,
      email: merchant.email,
      phone: merchant.phone,
      company_name: merchant.company_name,
      business_type: merchant.business_type,
      country: merchant.country,
      website: merchant.website,
    })
    setEditingMerchant(merchant)
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      await merchantService.delete(id)
      message.success('删除成功')
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      if (editingMerchant) {
        const updateData: UpdateMerchantRequest = {
          name: values.name,
          phone: values.phone,
          company_name: values.company_name,
          business_type: values.business_type,
          country: values.country,
          website: values.website,
        }
        await merchantService.update(editingMerchant.id, updateData)
        message.success('更新成功')
      } else {
        const createData: CreateMerchantRequest = {
          name: values.name,
          email: values.email,
          password: values.password,
          phone: values.phone,
          company_name: values.company_name,
          business_type: values.business_type,
          country: values.country,
          website: values.website,
        }
        await merchantService.create(createData)
        message.success('创建成功')
      }

      setModalVisible(false)
      loadMerchants()
    } catch (error) {
      // Error handled by interceptor or validation
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>商户管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          创建商户
        </Button>
      </div>

      {/* Statistics Cards */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总商户数"
              value={total}
              prefix={<ShopOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="待审核"
              value={stats.pending}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="正常运营"
              value={stats.active}
              prefix={<CheckOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已冻结"
              value={stats.suspended}
              prefix={<LockOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters */}
      <Space style={{ marginBottom: 16 }}>
        <Input
          placeholder="搜索商户名称、邮箱"
          prefix={<SearchOutlined />}
          style={{ width: 250 }}
          allowClear
          onChange={(e) => {
            setSearchKeyword(e.target.value)
            setPage(1)
          }}
        />
        <Select
          placeholder="状态筛选"
          style={{ width: 120 }}
          allowClear
          onChange={(value) => {
            setStatusFilter(value)
            setPage(1)
          }}
        >
          <Select.Option value="pending">待审核</Select.Option>
          <Select.Option value="active">正常</Select.Option>
          <Select.Option value="suspended">已冻结</Select.Option>
          <Select.Option value="rejected">已拒绝</Select.Option>
        </Select>
        <Select
          placeholder="KYC状态筛选"
          style={{ width: 140 }}
          allowClear
          onChange={(value) => {
            setKycStatusFilter(value)
            setPage(1)
          }}
        >
          <Select.Option value="pending">待审核</Select.Option>
          <Select.Option value="verified">已认证</Select.Option>
          <Select.Option value="rejected">已拒绝</Select.Option>
        </Select>
      </Space>

      <Table
        columns={columns}
        dataSource={merchants}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          showTotal: (total: number) => `共 ${total} 条`,
          onChange: (page: number, pageSize: number) => {
            setPage(page)
            setPageSize(pageSize)
          },
        }}
        scroll={{ x: 1600 }}
      />

      <Modal
        title={editingMerchant ? '编辑商户' : '创建商户'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={700}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="商户名称"
                rules={[{ required: true, message: '请输入商户名称' }]}
              >
                <Input placeholder="商户名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="email"
                label="邮箱"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' },
                ]}
              >
                <Input placeholder="邮箱" disabled={!!editingMerchant} />
              </Form.Item>
            </Col>
          </Row>

          {!editingMerchant && (
            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 8, message: '密码至少8个字符' },
              ]}
            >
              <Input.Password placeholder="密码" />
            </Form.Item>
          )}

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="phone" label="手机号">
                <Input placeholder="手机号" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="business_type"
                label="业务类型"
                rules={[{ required: true, message: '请选择业务类型' }]}
              >
                <Select placeholder="选择业务类型">
                  <Select.Option value="individual">个人</Select.Option>
                  <Select.Option value="company">公司</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="company_name" label="公司名称">
                <Input placeholder="公司名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="country" label="国家">
                <Input placeholder="国家" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="website" label="网站">
            <Input placeholder="https://example.com" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Merchants

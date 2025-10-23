import { useState, useEffect } from 'react'
import {
  Typography,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Space,
  Tag,
  message,
  Popconfirm,
  Avatar,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  UserOutlined,
  SearchOutlined,
  KeyOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { adminService, Admin, CreateAdminRequest, UpdateAdminRequest } from '../services/adminService'
import dayjs from 'dayjs'

const { Title } = Typography

const Admins = () => {
  const [loading, setLoading] = useState(false)
  const [admins, setAdmins] = useState<Admin[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [modalVisible, setModalVisible] = useState(false)
  const [resetPasswordModalVisible, setResetPasswordModalVisible] = useState(false)
  const [editingAdmin, setEditingAdmin] = useState<Admin | null>(null)
  const [resettingAdmin, setResettingAdmin] = useState<Admin | null>(null)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [form] = Form.useForm()
  const [resetPasswordForm] = Form.useForm()

  useEffect(() => {
    loadAdmins()
  }, [page, pageSize, searchKeyword, statusFilter])

  const loadAdmins = async () => {
    setLoading(true)
    try {
      const response = await adminService.list({
        page,
        page_size: pageSize,
        keyword: searchKeyword,
        status: statusFilter,
      })
      setAdmins(response.data)
      setTotal(response.pagination.total)
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const columns: ColumnsType<Admin> = [
    {
      title: '头像',
      dataIndex: 'avatar',
      key: 'avatar',
      width: 80,
      render: (avatar: string) => (
        <Avatar src={avatar} icon={<UserOutlined />} />
      ),
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      width: 120,
    },
    {
      title: '姓名',
      dataIndex: 'full_name',
      key: 'full_name',
      width: 120,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      ellipsis: true,
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      key: 'phone',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const colors: Record<string, string> = {
          active: 'green',
          inactive: 'red',
          locked: 'orange',
        }
        const labels: Record<string, string> = {
          active: '正常',
          inactive: '禁用',
          locked: '锁定',
        }
        return <Tag color={colors[status]}>{labels[status] || status}</Tag>
      },
    },
    {
      title: '超级管理员',
      dataIndex: 'is_super',
      key: 'is_super',
      width: 100,
      render: (is_super: boolean) => (is_super ? '是' : '否'),
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      width: 180,
      render: (time: string) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-',
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
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          {!record.is_super && (
            <>
              <Button
                type="link"
                size="small"
                icon={<KeyOutlined />}
                onClick={() => handleResetPassword(record)}
              >
                重置密码
              </Button>
              <Popconfirm
                title="确认删除"
                description={`确定要删除管理员 "${record.username}" 吗？`}
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
            </>
          )}
        </Space>
      ),
    },
  ]

  const handleAdd = () => {
    form.resetFields()
    setEditingAdmin(null)
    setModalVisible(true)
  }

  const handleEdit = (admin: Admin) => {
    form.setFieldsValue({
      username: admin.username,
      email: admin.email,
      full_name: admin.full_name,
      phone: admin.phone,
      status: admin.status,
    })
    setEditingAdmin(admin)
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      await adminService.delete(id)
      message.success('删除成功')
      loadAdmins()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleResetPassword = (admin: Admin) => {
    resetPasswordForm.resetFields()
    setResettingAdmin(admin)
    setResetPasswordModalVisible(true)
  }

  const handleResetPasswordSubmit = async () => {
    if (!resettingAdmin) return

    try {
      const values = await resetPasswordForm.validateFields()
      await adminService.resetPassword(resettingAdmin.id, values.new_password)
      message.success('密码重置成功')
      setResetPasswordModalVisible(false)
      setResettingAdmin(null)
    } catch (error) {
      // Error handled by interceptor or validation
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      if (editingAdmin) {
        const updateData: UpdateAdminRequest = {
          email: values.email,
          full_name: values.full_name,
          phone: values.phone,
          status: values.status,
        }
        await adminService.update(editingAdmin.id, updateData)
        message.success('更新成功')
      } else {
        const createData: CreateAdminRequest = {
          username: values.username,
          password: values.password,
          email: values.email,
          full_name: values.full_name,
          phone: values.phone,
        }
        await adminService.create(createData)
        message.success('创建成功')
      }

      setModalVisible(false)
      loadAdmins()
    } catch (error) {
      // Error handled by interceptor or validation
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>管理员管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          添加管理员
        </Button>
      </div>

      <Space style={{ marginBottom: 16 }}>
        <Input
          placeholder="搜索用户名、姓名、邮箱"
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
          <Select.Option value="active">正常</Select.Option>
          <Select.Option value="inactive">禁用</Select.Option>
          <Select.Option value="locked">锁定</Select.Option>
        </Select>
      </Space>

      <Table
        columns={columns}
        dataSource={admins}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (page, pageSize) => {
            setPage(page)
            setPageSize(pageSize)
          },
        }}
        scroll={{ x: 1200 }}
      />

      <Modal
        title={editingAdmin ? '编辑管理员' : '添加管理员'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
            ]}
          >
            <Input placeholder="用户名" disabled={!!editingAdmin} />
          </Form.Item>

          {!editingAdmin && (
            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, message: '密码至少6个字符' },
              ]}
            >
              <Input.Password placeholder="密码" />
            </Form.Item>
          )}

          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder="邮箱" />
          </Form.Item>

          <Form.Item
            name="full_name"
            label="姓名"
            rules={[{ required: true, message: '请输入姓名' }]}
          >
            <Input placeholder="姓名" />
          </Form.Item>

          <Form.Item name="phone" label="手机号">
            <Input placeholder="手机号" />
          </Form.Item>

          {editingAdmin && (
            <Form.Item
              name="status"
              label="状态"
              rules={[{ required: true, message: '请选择状态' }]}
            >
              <Select>
                <Select.Option value="active">正常</Select.Option>
                <Select.Option value="inactive">禁用</Select.Option>
                <Select.Option value="locked">锁定</Select.Option>
              </Select>
            </Form.Item>
          )}
        </Form>
      </Modal>

      <Modal
        title={`重置密码 - ${resettingAdmin?.username}`}
        open={resetPasswordModalVisible}
        onOk={handleResetPasswordSubmit}
        onCancel={() => {
          setResetPasswordModalVisible(false)
          setResettingAdmin(null)
        }}
        width={500}
      >
        <Form form={resetPasswordForm} layout="vertical">
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, message: '密码至少8个字符' },
            ]}
          >
            <Input.Password placeholder="请输入新密码（至少8个字符）" />
          </Form.Item>

          <Form.Item
            name="confirm_password"
            label="确认新密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Admins

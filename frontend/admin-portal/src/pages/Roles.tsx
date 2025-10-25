import { useState, useEffect } from 'react'
import {
  Typography,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Space,
  Tag,
  message,
  Popconfirm,
  Tree,
  Card,
  Row,
  Col,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { DataNode } from 'antd/es/tree'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SafetyOutlined,
  KeyOutlined,
} from '@ant-design/icons'
import { roleService, permissionService, Role, Permission } from '../services/roleService'
import dayjs from 'dayjs'

const { Title } = Typography
const { TextArea } = Input

const Roles = () => {
  const [loading, setLoading] = useState(false)
  const [roles, setRoles] = useState<Role[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [modalVisible, setModalVisible] = useState(false)
  const [permissionModalVisible, setPermissionModalVisible] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [selectedRole, setSelectedRole] = useState<Role | null>(null)
  const [permissions, setPermissions] = useState<Record<string, Permission[]>>({})
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([])
  const [form] = Form.useForm()

  useEffect(() => {
    loadRoles()
    loadPermissions()
  }, [page, pageSize])

  const loadRoles = async () => {
    setLoading(true)
    try {
      const response = await roleService.list({ page, page_size: pageSize })
      // 响应拦截器已解包，直接使用数据
      if (response) {
        setRoles(response.data || [])
        setTotal(response.pagination?.total || 0)
      }
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const loadPermissions = async () => {
    try {
      const response = await permissionService.listGrouped()
      setPermissions(response.data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const columns: ColumnsType<Role> = [
    {
      title: '角色代码',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: '显示名称',
      dataIndex: 'display_name',
      key: 'display_name',
      width: 150,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '系统角色',
      dataIndex: 'is_system',
      key: 'is_system',
      width: 100,
      render: (is_system: boolean) => (
        is_system ? <Tag color="blue">是</Tag> : <Tag>否</Tag>
      ),
    },
    {
      title: '权限数',
      key: 'permission_count',
      width: 100,
      render: (_, record) => record.permissions?.length || 0,
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
            icon={<KeyOutlined />}
            onClick={() => handleManagePermissions(record)}
          >
            权限
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          {!record.is_system && (
            <Popconfirm
              title="确认删除"
              description={`确定要删除角色 "${record.display_name}" 吗？`}
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
    setEditingRole(null)
    setModalVisible(true)
  }

  const handleEdit = (role: Role) => {
    form.setFieldsValue(role)
    setEditingRole(role)
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      await roleService.delete(id)
      message.success('删除成功')
      loadRoles()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      if (editingRole) {
        await roleService.update(editingRole.id, {
          display_name: values.display_name,
          description: values.description,
        })
        message.success('更新成功')
      } else {
        await roleService.create(values)
        message.success('创建成功')
      }

      setModalVisible(false)
      loadRoles()
    } catch (error) {
      // Error handled by interceptor or validation
    }
  }

  const handleManagePermissions = async (role: Role) => {
    setSelectedRole(role)

    // 加载角色详情以获取当前权限
    try {
      const response = await roleService.getById(role.id)
      const currentPermissionIds = response.data.permissions?.map((p: Permission) => p.id) || []
      setSelectedPermissions(currentPermissionIds)
      setPermissionModalVisible(true)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleSavePermissions = async () => {
    if (!selectedRole) return

    try {
      await roleService.assignPermissions(selectedRole.id, selectedPermissions)
      message.success('权限分配成功')
      setPermissionModalVisible(false)
      loadRoles()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  // 将权限转换为树形结构
  const convertPermissionsToTree = (): DataNode[] => {
    return Object.entries(permissions).map(([resource, perms]) => ({
      title: resource,
      key: resource,
      icon: <SafetyOutlined />,
      children: perms.map(p => ({
        title: `${p.name} (${p.code})`,
        key: p.id,
      })),
    }))
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>角色权限管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          添加角色
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={roles}
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

      {/* 角色编辑Modal */}
      <Modal
        title={editingRole ? '编辑角色' : '添加角色'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="角色代码"
            rules={[
              { required: true, message: '请输入角色代码' },
              { pattern: /^[a-z_]+$/, message: '只能使用小写字母和下划线' },
            ]}
          >
            <Input placeholder="例如: merchant_admin" disabled={!!editingRole} />
          </Form.Item>

          <Form.Item
            name="display_name"
            label="显示名称"
            rules={[{ required: true, message: '请输入显示名称' }]}
          >
            <Input placeholder="例如: 商户管理员" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <TextArea rows={3} placeholder="角色描述" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 权限分配Modal */}
      <Modal
        title={`为角色 "${selectedRole?.display_name}" 分配权限`}
        open={permissionModalVisible}
        onOk={handleSavePermissions}
        onCancel={() => setPermissionModalVisible(false)}
        width={800}
      >
        <Card>
          <Tree
            checkable
            defaultExpandAll
            treeData={convertPermissionsToTree()}
            checkedKeys={selectedPermissions}
            onCheck={(checkedKeys) => {
              setSelectedPermissions(checkedKeys as string[])
            }}
          />
        </Card>
        <div style={{ marginTop: 16 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Card size="small" title="已选择权限">
                <Tag color="blue">{selectedPermissions.length} 个权限</Tag>
              </Card>
            </Col>
            <Col span={12}>
              <Card size="small" title="总权限数">
                <Tag color="green">
                  {Object.values(permissions).reduce((sum, perms) => sum + perms.length, 0)} 个权限
                </Tag>
              </Card>
            </Col>
          </Row>
        </div>
      </Modal>
    </div>
  )
}

export default Roles

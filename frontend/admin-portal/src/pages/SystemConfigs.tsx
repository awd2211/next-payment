import { useState, useEffect } from 'react'
import {
  Typography,
  Tabs,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  message,
  Space,
  Tag,
} from 'antd'
import { EditOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { systemConfigService, SystemConfig } from '../services/systemConfigService'

const { Title } = Typography
const { TextArea } = Input

const SystemConfigs = () => {
  const [loading, setLoading] = useState(false)
  const [configs, setConfigs] = useState<Record<string, SystemConfig[]>>({})
  const [modalVisible, setModalVisible] = useState(false)
  const [editingConfig, setEditingConfig] = useState<SystemConfig | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    loadConfigs()
  }, [])

  const loadConfigs = async () => {
    setLoading(true)
    try {
      const response = await systemConfigService.listGrouped()
      // 响应拦截器已解包，直接使用数据
      if (response && response.data) {
        setConfigs(response.data.configs || {})
      }
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const columns: ColumnsType<SystemConfig> = [
    {
      title: '配置键',
      dataIndex: 'key',
      key: 'key',
      width: 250,
    },
    {
      title: '配置值',
      dataIndex: 'value',
      key: 'value',
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => {
        const colors: Record<string, string> = {
          string: 'blue',
          number: 'green',
          boolean: 'orange',
          json: 'purple',
        }
        return <Tag color={colors[type]}>{type}</Tag>
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '公开',
      dataIndex: 'is_public',
      key: 'is_public',
      width: 80,
      render: (is_public: boolean) => (is_public ? '是' : '否'),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
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
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  const handleAdd = () => {
    form.resetFields()
    setEditingConfig(null)
    setModalVisible(true)
  }

  const handleEdit = (config: SystemConfig) => {
    form.setFieldsValue(config)
    setEditingConfig(config)
    setModalVisible(true)
  }

  const handleDelete = (config: SystemConfig) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除配置 "${config.key}" 吗？`,
      onOk: async () => {
        try {
          await systemConfigService.delete(config.id)
          message.success('删除成功')
          loadConfigs()
        } catch (error) {
          // Error handled by interceptor
        }
      },
    })
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      if (editingConfig) {
        await systemConfigService.update(editingConfig.id, values)
        message.success('更新成功')
      } else {
        await systemConfigService.create(values)
        message.success('创建成功')
      }

      setModalVisible(false)
      loadConfigs()
    } catch (error) {
      // Error handled by interceptor or validation
    }
  }

  const categoryNames: Record<string, string> = {
    payment: '支付配置',
    notification: '通知配置',
    risk: '风控配置',
    system: '系统配置',
    settlement: '结算配置',
  }

  const tabItems = Object.entries(configs).map(([category, items]) => ({
    key: category,
    label: categoryNames[category] || category,
    children: (
      <Table
        columns={columns}
        dataSource={items}
        rowKey="id"
        loading={loading}
        pagination={false}
      />
    ),
  }))

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>系统配置</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          添加配置
        </Button>
      </div>

      <Tabs items={tabItems} />

      <Modal
        title={editingConfig ? '编辑配置' : '添加配置'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="key"
            label="配置键"
            rules={[{ required: true, message: '请输入配置键' }]}
          >
            <Input placeholder="例如: payment.default_currency" disabled={!!editingConfig} />
          </Form.Item>

          <Form.Item
            name="value"
            label="配置值"
            rules={[{ required: true, message: '请输入配置值' }]}
          >
            <TextArea rows={3} placeholder="配置值" />
          </Form.Item>

          <Form.Item
            name="type"
            label="数据类型"
            rules={[{ required: true, message: '请选择数据类型' }]}
          >
            <Select placeholder="选择数据类型" disabled={!!editingConfig}>
              <Select.Option value="string">string</Select.Option>
              <Select.Option value="number">number</Select.Option>
              <Select.Option value="boolean">boolean</Select.Option>
              <Select.Option value="json">json</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="category"
            label="配置类别"
            rules={[{ required: true, message: '请选择配置类别' }]}
          >
            <Select placeholder="选择配置类别" disabled={!!editingConfig}>
              <Select.Option value="payment">支付配置</Select.Option>
              <Select.Option value="notification">通知配置</Select.Option>
              <Select.Option value="risk">风控配置</Select.Option>
              <Select.Option value="system">系统配置</Select.Option>
              <Select.Option value="settlement">结算配置</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="description" label="描述">
            <Input placeholder="配置描述" />
          </Form.Item>

          <Form.Item name="is_public" label="是否公开" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default SystemConfigs

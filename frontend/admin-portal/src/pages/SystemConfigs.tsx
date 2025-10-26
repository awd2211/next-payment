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
      const response = await systemConfigService.list({ page: 1, page_size: 100 })
      // 响应拦截器已解包，直接使用数据
      if (response && response.list) {
        // 按服务名称分组
        const grouped = (response.list || []).reduce((acc: Record<string, SystemConfig[]>, config: SystemConfig) => {
          const category = config.service_name || 'uncategorized'
          if (!acc[category]) {
            acc[category] = []
          }
          acc[category].push(config)
          return acc
        }, {})
        setConfigs(grouped)
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
      dataIndex: 'config_key',
      key: 'config_key',
      width: 250,
    },
    {
      title: '配置值',
      dataIndex: 'config_value',
      key: 'config_value',
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'value_type',
      key: 'value_type',
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
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      width: 100,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '加密',
      dataIndex: 'is_encrypted',
      key: 'is_encrypted',
      width: 80,
      render: (is_encrypted: boolean) => (is_encrypted ? '是' : '否'),
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
      content: `确定要删除配置 "${config.config_key}" 吗？`,
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
            name="service_name"
            label="服务名称"
            rules={[{ required: true, message: '请输入服务名称' }]}
          >
            <Input placeholder="例如: payment-gateway" disabled={!!editingConfig} />
          </Form.Item>

          <Form.Item
            name="config_key"
            label="配置键"
            rules={[{ required: true, message: '请输入配置键' }]}
          >
            <Input placeholder="例如: payment.default_currency" disabled={!!editingConfig} />
          </Form.Item>

          <Form.Item
            name="config_value"
            label="配置值"
            rules={[{ required: true, message: '请输入配置值' }]}
          >
            <TextArea rows={3} placeholder="配置值" />
          </Form.Item>

          <Form.Item
            name="value_type"
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
            name="environment"
            label="环境"
            rules={[{ required: true, message: '请选择环境' }]}
          >
            <Select placeholder="选择环境">
              <Select.Option value="development">开发环境</Select.Option>
              <Select.Option value="staging">测试环境</Select.Option>
              <Select.Option value="production">生产环境</Select.Option>
              <Select.Option value="all">所有环境</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="description" label="描述">
            <TextArea rows={2} placeholder="配置描述" />
          </Form.Item>

          <Form.Item name="is_encrypted" label="是否加密" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default SystemConfigs

import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Form, Input, Switch, Space, message, Tabs } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ApiOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { channelService, type Channel } from '../services/channelService'

export default function Channels() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<Channel[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRecord, setEditingRecord] = useState<Channel | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    setLoading(true)
    try {
      const response = await channelService.list({ page: 1, page_size: 50 })
      // 响应拦截器已解包，直接使用数据
      if (response && response.list) {
        setData(response.list)
      }
    } catch (error) {
      // 错误已被拦截器处理并显示
      console.error('Failed to fetch channels:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleAdd = () => {
    setEditingRecord(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: Channel) => {
    setEditingRecord(record)
    form.setFieldsValue({
      ...record,
      api_key: '', // 安全考虑,不显示完整密钥
      api_secret: '',
      webhook_secret: '',
    })
    setModalVisible(true)
  }

  const handleDelete = (record: Channel) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除支付渠道 "${record.channel_name}" 吗?`,
      onOk: async () => {
        try {
          await channelService.delete(record.id)
          message.success('删除成功')
          fetchData()
        } catch (error) {
          // 错误已被拦截器处理并显示
          console.error('Failed to delete channel:', error)
        }
      },
    })
  }

  const handleToggleStatus = async (record: Channel, enabled: boolean) => {
    try {
      await channelService.toggleEnable(record.id, enabled)
      message.success(`已${enabled ? '启用' : '禁用'}渠道`)
      fetchData()
    } catch (error) {
      // 错误已被拦截器处理并显示
      console.error('Failed to toggle channel status:', error)
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      let response
      if (editingRecord) {
        // 构建更新数据,只包含修改的字段
        const updateData: any = {
          channel_name: values.channel_name,
          is_enabled: values.is_enabled,
          is_test_mode: values.test_mode,
        }
        // 只有在填写了新值时才包含敏感字段
        if (values.api_key) {
          updateData.config = {
            api_key: values.api_key,
            api_secret: values.api_secret,
            webhook_secret: values.webhook_secret,
          }
        }
        response = await channelService.update(editingRecord.id, updateData)
      } else {
        // 创建新渠道
        const createData = {
          channel_code: values.channel_code,
          channel_name: values.channel_name,
          channel_type: values.channel_type,
          is_enabled: values.is_enabled || false,
          is_test_mode: values.test_mode || false,
          config: {
            api_key: values.api_key,
            api_secret: values.api_secret,
            webhook_secret: values.webhook_secret,
          },
          supported_currencies: ['USD', 'EUR', 'GBP', 'CNY'],
          fee_type: 'percentage' as const,
          fee_percentage: 2.9,
        }
        response = await channelService.create(createData)
      }

      // 响应拦截器已解包，成功则执行
      message.success(editingRecord ? '更新成功' : '创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchData()
    } catch (error) {
      // 错误已被拦截器处理并显示
      console.error('Failed to save channel:', error)
    }
  }

  const columns: ColumnsType<Channel> = [
    {
      title: '渠道名称',
      dataIndex: 'channel_name',
      width: 150,
      render: (name, record) => (
        <Space>
          <ApiOutlined />
          <strong>{name}</strong>
        </Space>
      ),
    },
    {
      title: '渠道代码',
      dataIndex: 'channel_code',
      width: 120,
    },
    {
      title: '类型',
      dataIndex: 'channel_type',
      width: 120,
      render: (type) => {
        const colorMap: Record<string, string> = {
          stripe: 'blue',
          paypal: 'cyan',
          alipay: 'green',
          wechat: 'lime',
          crypto: 'purple',
        }
        return <Tag color={colorMap[type]}>{type.toUpperCase()}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'is_enabled',
      width: 100,
      render: (enabled, record) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggleStatus(record, checked)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
    },
    {
      title: '模式',
      dataIndex: 'test_mode',
      width: 100,
      render: (testMode) => (
        <Tag color={testMode ? 'orange' : 'green'}>
          {testMode ? '测试' : '生产'}
        </Tag>
      ),
    },
    {
      title: 'API Key',
      dataIndex: 'api_key',
      width: 200,
      ellipsis: true,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            type="link"
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

  return (
    <div>
      <Card
        title="支付渠道管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加渠道
          </Button>
        }
      >
        <Tabs
          defaultActiveKey="all"
          items={[
            {
              key: 'all',
              label: '全部渠道',
              children: (
                <Table
                  columns={columns}
                  dataSource={data}
                  loading={loading}
                  rowKey="id"
                  scroll={{ x: 1200 }}
                />
              ),
            },
            {
              key: 'enabled',
              label: '已启用',
              children: (
                <Table
                  columns={columns}
                  dataSource={data.filter((item) => item.is_enabled)}
                  loading={loading}
                  rowKey="id"
                  scroll={{ x: 1200 }}
                />
              ),
            },
            {
              key: 'disabled',
              label: '已禁用',
              children: (
                <Table
                  columns={columns}
                  dataSource={data.filter((item) => !item.is_enabled)}
                  loading={loading}
                  rowKey="id"
                  scroll={{ x: 1200 }}
                />
              ),
            },
          ]}
        />
      </Card>

      {/* 添加/编辑Modal */}
      <Modal
        title={editingRecord ? '编辑支付渠道' : '添加支付渠道'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item
            name="channel_name"
            label="渠道名称"
            rules={[{ required: true, message: '请输入渠道名称' }]}
          >
            <Input placeholder="例如: Stripe" />
          </Form.Item>

          <Form.Item
            name="channel_code"
            label="渠道代码"
            rules={[{ required: true, message: '请输入渠道代码' }]}
          >
            <Input placeholder="例如: stripe" disabled={!!editingRecord} />
          </Form.Item>

          <Form.Item
            name="channel_type"
            label="渠道类型"
            rules={[{ required: true, message: '请选择渠道类型' }]}
          >
            <Input placeholder="stripe/paypal/alipay/wechat/crypto" disabled={!!editingRecord} />
          </Form.Item>

          <Form.Item
            name="api_key"
            label="API Key"
            rules={[{ required: !editingRecord, message: '请输入API Key' }]}
          >
            <Input.Password placeholder={editingRecord ? '留空表示不修改' : '输入API Key'} />
          </Form.Item>

          <Form.Item name="api_secret" label="API Secret">
            <Input.Password placeholder={editingRecord ? '留空表示不修改' : '输入API Secret'} />
          </Form.Item>

          <Form.Item name="webhook_secret" label="Webhook Secret">
            <Input.Password placeholder={editingRecord ? '留空表示不修改' : '输入Webhook Secret'} />
          </Form.Item>

          <Form.Item name="test_mode" label="测试模式" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item name="is_enabled" label="启用状态" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

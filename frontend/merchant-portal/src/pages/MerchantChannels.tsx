import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Form, Input, Switch, Space, message, Alert } from 'antd'
import { EditOutlined, ApiOutlined, CheckCircleOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

interface MerchantChannel {
  id: string
  channel_code: string
  channel_name: string
  channel_type: 'stripe' | 'paypal' | 'alipay' | 'wechat'
  is_enabled: boolean
  is_configured: boolean
  test_mode: boolean
  config: {
    api_key?: string
    api_secret?: string
    merchant_id?: string
    [key: string]: string | undefined
  }
  created_at: string
  updated_at: string
}

export default function MerchantChannels() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<MerchantChannel[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [editingChannel, setEditingChannel] = useState<MerchantChannel | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    setLoading(true)
    // TODO: 调用 merchantChannelService.list()
    setTimeout(() => {
      setData([
        {
          id: '1',
          channel_code: 'stripe',
          channel_name: 'Stripe',
          channel_type: 'stripe',
          is_enabled: true,
          is_configured: true,
          test_mode: false,
          config: {
            api_key: 'sk_live_***************************',
            webhook_secret: 'whsec_***************************',
          },
          created_at: '2025-10-01 10:00:00',
          updated_at: '2025-10-20 15:30:00',
        },
        {
          id: '2',
          channel_code: 'paypal',
          channel_name: 'PayPal',
          channel_type: 'paypal',
          is_enabled: false,
          is_configured: false,
          test_mode: true,
          config: {},
          created_at: '2025-10-05 14:20:00',
          updated_at: '2025-10-05 14:20:00',
        },
      ])
      setLoading(false)
    }, 500)
  }

  const handleConfigure = (record: MerchantChannel) => {
    setEditingChannel(record)
    form.setFieldsValue({
      channel_code: record.channel_code,
      test_mode: record.test_mode,
      ...record.config,
    })
    setModalVisible(true)
  }

  const handleToggleStatus = async (record: MerchantChannel, enabled: boolean) => {
    if (enabled && !record.is_configured) {
      message.warning('请先配置渠道信息')
      return
    }

    try {
      // TODO: 调用 merchantChannelService.toggleStatus(record.id, enabled)
      message.success(`已${enabled ? '启用' : '禁用'}渠道`)
      fetchData()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleTest = (record: MerchantChannel) => {
    if (!record.is_configured) {
      message.warning('请先配置渠道信息')
      return
    }
    setEditingChannel(record)
    setTestModalVisible(true)
  }

  const handleTestConnection = async () => {
    try {
      // TODO: 调用 merchantChannelService.testConnection(editingChannel.id)
      message.success('连接测试成功')
      setTestModalVisible(false)
    } catch (error) {
      message.error('连接测试失败')
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      const configData = {
        test_mode: values.test_mode,
        config: {} as Record<string, string>,
      }

      // 根据渠道类型构建配置
      if (editingChannel?.channel_type === 'stripe') {
        configData.config = {
          api_key: values.api_key,
          webhook_secret: values.webhook_secret,
        }
      } else if (editingChannel?.channel_type === 'paypal') {
        configData.config = {
          client_id: values.client_id,
          client_secret: values.client_secret,
        }
      }

      // TODO: 调用 merchantChannelService.configure(editingChannel.id, configData)
      message.success('配置成功')
      setModalVisible(false)
      form.resetFields()
      fetchData()
    } catch (error) {
      message.error('配置失败')
    }
  }

  const columns: ColumnsType<MerchantChannel> = [
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
      title: '配置状态',
      dataIndex: 'is_configured',
      width: 120,
      render: (isConfigured: boolean) => (
        <Tag color={isConfigured ? 'green' : 'orange'} icon={isConfigured ? <CheckCircleOutlined /> : undefined}>
          {isConfigured ? '已配置' : '未配置'}
        </Tag>
      ),
    },
    {
      title: '启用状态',
      dataIndex: 'is_enabled',
      width: 120,
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
        <Tag color={testMode ? 'orange' : 'green'}>{testMode ? '测试' : '生产'}</Tag>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
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
            icon={<EditOutlined />}
            onClick={() => handleConfigure(record)}
          >
            配置
          </Button>
          <Button type="link" onClick={() => handleTest(record)}>
            测试连接
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Alert
        message="提示"
        description="配置您自己的支付渠道账号,所有交易将直接进入您的账户。请确保配置信息准确无误。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Card title="我的支付渠道">
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1000 }}
          pagination={false}
        />
      </Card>

      {/* 配置Modal */}
      <Modal
        title={`配置 ${editingChannel?.channel_name}`}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          {editingChannel?.channel_type === 'stripe' && (
            <>
              <Form.Item
                name="api_key"
                label="API Key"
                rules={[{ required: true, message: '请输入Stripe API Key' }]}
                tooltip="从Stripe Dashboard获取"
              >
                <Input.Password placeholder="sk_live_..." />
              </Form.Item>

              <Form.Item
                name="webhook_secret"
                label="Webhook Secret"
                rules={[{ required: true, message: '请输入Webhook Secret' }]}
                tooltip="从Stripe Webhook设置页面获取"
              >
                <Input.Password placeholder="whsec_..." />
              </Form.Item>
            </>
          )}

          {editingChannel?.channel_type === 'paypal' && (
            <>
              <Form.Item
                name="client_id"
                label="Client ID"
                rules={[{ required: true, message: '请输入PayPal Client ID' }]}
              >
                <Input placeholder="AW..." />
              </Form.Item>

              <Form.Item
                name="client_secret"
                label="Client Secret"
                rules={[{ required: true, message: '请输入PayPal Client Secret' }]}
              >
                <Input.Password placeholder="EL..." />
              </Form.Item>
            </>
          )}

          <Form.Item name="test_mode" label="使用测试模式" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Alert
            message="安全提示"
            description="您的API密钥将被加密存储,仅用于处理您的交易,平台不会查看或使用您的密钥进行任何其他操作。"
            type="warning"
            showIcon
          />
        </Form>
      </Modal>

      {/* 测试连接Modal */}
      <Modal
        title="测试渠道连接"
        open={testModalVisible}
        onCancel={() => setTestModalVisible(false)}
        onOk={handleTestConnection}
        okText="开始测试"
      >
        <p>即将测试 <strong>{editingChannel?.channel_name}</strong> 的连接状态</p>
        <p>测试将验证:</p>
        <ul>
          <li>API密钥有效性</li>
          <li>网络连接状态</li>
          <li>账户权限配置</li>
        </ul>
      </Modal>
    </div>
  )
}

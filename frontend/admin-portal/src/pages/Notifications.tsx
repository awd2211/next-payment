import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Form, Input, Select, Space, message, Tabs } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined, MailOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

interface Notification {
  id: string
  title: string
  content: string
  type: 'email' | 'sms' | 'webhook' | 'system'
  status: 'pending' | 'sent' | 'failed'
  recipient: string
  sent_at?: string
  created_at: string
}

interface EmailTemplate {
  id: string
  name: string
  subject: string
  content: string
  variables: string[]
  is_active: boolean
  created_at: string
}

export default function Notifications() {
  const [loading, setLoading] = useState(false)
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [templates, setTemplates] = useState<EmailTemplate[]>([])
  const [modalVisible, setModalVisible] = useState(false)
  const [templateModalVisible, setTemplateModalVisible] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<EmailTemplate | null>(null)
  const [form] = Form.useForm()
  const [templateForm] = Form.useForm()

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    setLoading(true)
    // TODO: 调用 notificationService.list()
    // TODO: 调用 notificationService.listTemplates()
    setTimeout(() => {
      setNotifications([
        {
          id: '1',
          title: '支付成功通知',
          content: '您的订单已支付成功',
          type: 'email',
          status: 'sent',
          recipient: 'user@example.com',
          sent_at: '2025-10-25 10:00:00',
          created_at: '2025-10-25 09:55:00',
        },
        {
          id: '2',
          title: 'KYC审核通知',
          content: '您的KYC申请已通过',
          type: 'email',
          status: 'sent',
          recipient: 'merchant@example.com',
          sent_at: '2025-10-24 15:30:00',
          created_at: '2025-10-24 15:28:00',
        },
      ])

      setTemplates([
        {
          id: '1',
          name: '支付成功通知',
          subject: '支付成功 - 订单 {{order_no}}',
          content: '尊敬的用户,您的订单{{order_no}}已支付成功,金额{{amount}}',
          variables: ['order_no', 'amount', 'merchant_name'],
          is_active: true,
          created_at: '2025-10-01 10:00:00',
        },
        {
          id: '2',
          name: 'KYC审核通知',
          subject: 'KYC审核结果',
          content: '尊敬的商户{{merchant_name}},您的KYC审核{{status}}',
          variables: ['merchant_name', 'status', 'reason'],
          is_active: true,
          created_at: '2025-10-01 10:00:00',
        },
      ])
      setLoading(false)
    }, 500)
  }

  const handleSendNotification = () => {
    setModalVisible(true)
  }

  const handleSendSubmit = async (values: any) => {
    try {
      // TODO: 调用 notificationService.send(values)
      message.success('通知发送成功')
      setModalVisible(false)
      form.resetFields()
      fetchData()
    } catch (error) {
      message.error('发送失败')
    }
  }

  const handleAddTemplate = () => {
    setEditingTemplate(null)
    templateForm.resetFields()
    setTemplateModalVisible(true)
  }

  const handleEditTemplate = (record: EmailTemplate) => {
    setEditingTemplate(record)
    templateForm.setFieldsValue(record)
    setTemplateModalVisible(true)
  }

  const handleDeleteTemplate = (record: EmailTemplate) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除模板 "${record.name}" 吗?`,
      onOk: async () => {
        try {
          // TODO: 调用 notificationService.deleteTemplate(record.id)
          message.success('删除成功')
          fetchData()
        } catch (error) {
          message.error('删除失败')
        }
      },
    })
  }

  const handleTemplateSubmit = async (values: any) => {
    try {
      if (editingTemplate) {
        // TODO: 调用 notificationService.updateTemplate(editingTemplate.id, values)
        message.success('更新成功')
      } else {
        // TODO: 调用 notificationService.createTemplate(values)
        message.success('创建成功')
      }
      setTemplateModalVisible(false)
      templateForm.resetFields()
      fetchData()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const notificationColumns: ColumnsType<Notification> = [
    {
      title: '标题',
      dataIndex: 'title',
      width: 200,
    },
    {
      title: '类型',
      dataIndex: 'type',
      width: 100,
      render: (type) => {
        const colorMap: Record<string, string> = {
          email: 'blue',
          sms: 'green',
          webhook: 'purple',
          system: 'orange',
        }
        const textMap: Record<string, string> = {
          email: '邮件',
          sms: '短信',
          webhook: 'Webhook',
          system: '系统',
        }
        return <Tag color={colorMap[type]}>{textMap[type]}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status) => {
        const colorMap: Record<string, string> = {
          pending: 'orange',
          sent: 'green',
          failed: 'red',
        }
        const textMap: Record<string, string> = {
          pending: '待发送',
          sent: '已发送',
          failed: '失败',
        }
        return <Tag color={colorMap[status]}>{textMap[status]}</Tag>
      },
    },
    {
      title: '收件人',
      dataIndex: 'recipient',
      width: 200,
    },
    {
      title: '发送时间',
      dataIndex: 'sent_at',
      width: 180,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
    },
  ]

  const templateColumns: ColumnsType<EmailTemplate> = [
    {
      title: '模板名称',
      dataIndex: 'name',
      width: 200,
    },
    {
      title: '主题',
      dataIndex: 'subject',
      width: 250,
    },
    {
      title: '变量',
      dataIndex: 'variables',
      width: 250,
      render: (variables: string[]) => (
        <Space wrap>
          {variables.map((v) => (
            <Tag key={v}>{`{{${v}}}`}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      width: 100,
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'red'}>{isActive ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEditTemplate(record)}>
            编辑
          </Button>
          <Button
            type="link"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteTemplate(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Tabs
        items={[
          {
            key: 'notifications',
            label: '通知记录',
            children: (
              <Card
                extra={
                  <Button type="primary" icon={<SendOutlined />} onClick={handleSendNotification}>
                    发送通知
                  </Button>
                }
              >
                <Table
                  columns={notificationColumns}
                  dataSource={notifications}
                  loading={loading}
                  rowKey="id"
                  scroll={{ x: 1200 }}
                />
              </Card>
            ),
          },
          {
            key: 'templates',
            label: '邮件模板',
            children: (
              <Card
                extra={
                  <Button type="primary" icon={<PlusOutlined />} onClick={handleAddTemplate}>
                    添加模板
                  </Button>
                }
              >
                <Table
                  columns={templateColumns}
                  dataSource={templates}
                  loading={loading}
                  rowKey="id"
                  scroll={{ x: 1200 }}
                />
              </Card>
            ),
          },
        ]}
      />

      {/* 发送通知Modal */}
      <Modal
        title="发送通知"
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
      >
        <Form form={form} onFinish={handleSendSubmit} layout="vertical">
          <Form.Item
            name="type"
            label="通知类型"
            rules={[{ required: true, message: '请选择通知类型' }]}
          >
            <Select
              options={[
                { label: '邮件', value: 'email' },
                { label: '短信', value: 'sms' },
                { label: 'Webhook', value: 'webhook' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="template_id"
            label="选择模板"
            rules={[{ required: true, message: '请选择模板' }]}
          >
            <Select
              options={templates.map((t) => ({ label: t.name, value: t.id }))}
            />
          </Form.Item>

          <Form.Item
            name="recipient"
            label="收件人"
            rules={[{ required: true, message: '请输入收件人' }]}
          >
            <Input placeholder="邮箱地址或手机号" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 邮件模板Modal */}
      <Modal
        title={editingTemplate ? '编辑模板' : '添加模板'}
        open={templateModalVisible}
        onCancel={() => {
          setTemplateModalVisible(false)
          templateForm.resetFields()
        }}
        onOk={() => templateForm.submit()}
        width={700}
      >
        <Form form={templateForm} onFinish={handleTemplateSubmit} layout="vertical">
          <Form.Item
            name="name"
            label="模板名称"
            rules={[{ required: true, message: '请输入模板名称' }]}
          >
            <Input placeholder="例如: 支付成功通知" />
          </Form.Item>

          <Form.Item
            name="subject"
            label="邮件主题"
            rules={[{ required: true, message: '请输入邮件主题' }]}
          >
            <Input placeholder="支持变量: {{order_no}}" />
          </Form.Item>

          <Form.Item
            name="content"
            label="邮件内容"
            rules={[{ required: true, message: '请输入邮件内容' }]}
          >
            <Input.TextArea rows={6} placeholder="支持变量: {{variable_name}}" />
          </Form.Item>

          <Form.Item
            name="variables"
            label="可用变量"
            tooltip="多个变量用逗号分隔"
          >
            <Select
              mode="tags"
              placeholder="order_no, amount, merchant_name"
            />
          </Form.Item>

          <Form.Item name="is_active" label="启用状态" valuePropName="checked">
            <Select
              options={[
                { label: '启用', value: true },
                { label: '禁用', value: false },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

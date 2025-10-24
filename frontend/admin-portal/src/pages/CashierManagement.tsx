import { useState, useEffect } from 'react'
import {
  Card,
  Tabs,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  Space,
  message,
  Row,
  Col,
  Statistic,
  Tag,
  Upload,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SettingOutlined,
  BarChartOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { adminCashierService, CashierTemplate } from '../services/cashierService'
import { Column, Pie } from '@ant-design/charts'

const { TextArea } = Input
const { TabPane } = Tabs

const CashierManagement = () => {
  const [loading, setLoading] = useState(false)
  const [templates, setTemplates] = useState<CashierTemplate[]>([])
  const [stats, setStats] = useState<any>(null)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<CashierTemplate | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    loadTemplates()
    loadStats()
  }, [])

  const loadTemplates = async () => {
    setLoading(true)
    try {
      const response = await adminCashierService.listTemplates()
      setTemplates(response.data || [])
    } catch (error: any) {
      if (error.response?.status !== 404) {
        message.error('加载模板失败')
      }
    } finally {
      setLoading(false)
    }
  }

  const loadStats = async () => {
    try {
      const response = await adminCashierService.getPlatformStats()
      setStats(response.data)
    } catch (error) {
      console.error('Failed to load stats:', error)
    }
  }

  const handleAdd = () => {
    form.resetFields()
    setEditingTemplate(null)
    setModalVisible(true)
  }

  const handleEdit = (template: CashierTemplate) => {
    form.setFieldsValue(template)
    setEditingTemplate(template)
    setModalVisible(true)
  }

  const handleDelete = (template: CashierTemplate) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除模板 "${template.name}" 吗？`,
      onOk: async () => {
        try {
          await adminCashierService.deleteTemplate(template.id)
          message.success('删除成功')
          loadTemplates()
        } catch (error) {
          message.error('删除失败')
        }
      },
    })
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      if (editingTemplate) {
        await adminCashierService.updateTemplate(editingTemplate.id, values)
        message.success('更新成功')
      } else {
        await adminCashierService.createTemplate(values)
        message.success('创建成功')
      }

      setModalVisible(false)
      loadTemplates()
    } catch (error: any) {
      if (error.errorFields) {
        message.error('请检查表单填写')
      } else {
        message.error('保存失败')
      }
    } finally {
      setLoading(false)
    }
  }

  const templateColumns: ColumnsType<CashierTemplate> = [
    {
      title: '模板名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类型',
      dataIndex: 'template_type',
      key: 'template_type',
      render: (type: string) => {
        const typeMap: Record<string, { label: string; color: string }> = {
          default: { label: '默认', color: 'blue' },
          ecommerce: { label: '电商', color: 'green' },
          subscription: { label: '订阅', color: 'purple' },
          donation: { label: '捐赠', color: 'orange' },
        }
        const config = typeMap[type] || { label: type, color: 'default' }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (is_active: boolean) => (
        <Tag color={is_active ? 'success' : 'default'}>
          {is_active ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => message.info('预览功能开发中')}
          >
            预览
          </Button>
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

  const channelData = [
    { channel: 'Stripe', count: 450, percentage: 52 },
    { channel: 'PayPal', count: 280, percentage: 32 },
    { channel: '支付宝', count: 138, percentage: 16 },
  ]

  const conversionData = [
    { merchant: '商户A', rate: 78.5 },
    { merchant: '商户B', rate: 72.3 },
    { merchant: '商户C', rate: 68.9 },
    { merchant: '商户D', rate: 65.2 },
    { merchant: '商户E', rate: 61.8 },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Card
        title={
          <Space>
            <SettingOutlined />
            收银台管理
          </Space>
        }
      >
        <Tabs defaultActiveKey="templates">
          {/* 模板管理 */}
          <TabPane
            tab={
              <span>
                <FileTextOutlined />
                模板管理
              </span>
            }
            key="templates"
          >
            <div style={{ marginBottom: 16 }}>
              <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
                创建模板
              </Button>
            </div>
            <Table
              columns={templateColumns}
              dataSource={templates}
              rowKey="id"
              loading={loading}
              pagination={{ pageSize: 10 }}
            />
          </TabPane>

          {/* 全局配置 */}
          <TabPane
            tab={
              <span>
                <SettingOutlined />
                全局配置
              </span>
            }
            key="global-config"
          >
            <Form layout="vertical">
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item label="默认会话超时时间">
                    <Input addonAfter="分钟" placeholder="30" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="默认支付渠道顺序">
                    <Select mode="multiple" placeholder="拖拽排序">
                      <Select.Option value="stripe">Stripe</Select.Option>
                      <Select.Option value="paypal">PayPal</Select.Option>
                      <Select.Option value="alipay">支付宝</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item label="强制3D验证" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item label="强制CVV验证" valuePropName="checked">
                    <Switch defaultChecked />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item label="允许渠道切换" valuePropName="checked">
                    <Switch defaultChecked />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item>
                <Button type="primary">保存全局配置</Button>
              </Form.Item>
            </Form>
          </TabPane>

          {/* 监控面板 */}
          <TabPane
            tab={
              <span>
                <BarChartOutlined />
                监控面板
              </span>
            }
            key="monitoring"
          >
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              <Col span={6}>
                <Statistic
                  title="启用收银台的商户数"
                  value={stats?.total_merchants || 0}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="今日支付会话"
                  value={stats?.total_sessions_today || 0}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="平均转化率"
                  value={stats?.avg_conversion_rate || 0}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: '#faad14' }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="今日完成支付"
                  value={stats?.completed_sessions_today || 0}
                  valueStyle={{ color: '#722ed1' }}
                />
              </Col>
            </Row>

            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Card title="平台渠道分布" size="small">
                  <Pie
                    data={channelData}
                    angleField="count"
                    colorField="channel"
                    radius={0.8}
                    label={{
                      type: 'outer',
                      content: '{name} {percentage}%',
                    }}
                    legend={{ position: 'bottom' }}
                  />
                </Card>
              </Col>

              <Col span={12}>
                <Card title="商户转化率排行" size="small">
                  <Column
                    data={conversionData}
                    xField="merchant"
                    yField="rate"
                    label={{
                      position: 'top',
                      formatter: (datum: any) => `${datum.rate}%`,
                    }}
                    yAxis={{
                      label: {
                        formatter: (v: string) => `${v}%`,
                      },
                    }}
                  />
                </Card>
              </Col>
            </Row>
          </TabPane>

          {/* 日志查询 */}
          <TabPane
            tab={
              <span>
                <FileTextOutlined />
                日志查询
              </span>
            }
            key="logs"
          >
            <Card>
              <p>日志查询功能开发中...</p>
            </Card>
          </TabPane>
        </Tabs>
      </Card>

      {/* 模板编辑弹窗 */}
      <Modal
        title={editingTemplate ? '编辑模板' : '创建模板'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={700}
        confirmLoading={loading}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="模板名称"
            rules={[{ required: true, message: '请输入模板名称' }]}
          >
            <Input placeholder="例如: 电商标准模板" />
          </Form.Item>

          <Form.Item
            name="template_type"
            label="模板类型"
            rules={[{ required: true, message: '请选择模板类型' }]}
          >
            <Select placeholder="选择模板类型">
              <Select.Option value="default">默认</Select.Option>
              <Select.Option value="ecommerce">电商</Select.Option>
              <Select.Option value="subscription">订阅</Select.Option>
              <Select.Option value="donation">捐赠</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="description" label="描述">
            <TextArea rows={3} placeholder="模板描述" />
          </Form.Item>

          <Form.Item name="preview_image_url" label="预览图URL">
            <Input placeholder="https://example.com/preview.png" />
          </Form.Item>

          <Form.Item name="is_active" label="启用状态" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>

          <Form.Item
            name="config"
            label="配置JSON"
            rules={[{ required: true, message: '请输入配置JSON' }]}
          >
            <TextArea
              rows={8}
              placeholder={`{
  "theme_color": "#1890ff",
  "logo_url": "",
  "enabled_channels": ["stripe", "paypal"],
  "default_language": "en"
}`}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default CashierManagement

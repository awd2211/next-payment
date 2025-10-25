import { useState } from 'react'
import {
  Card,
  Button,
  Space,
  Input,
  Modal,
  Form,
  message,
  Alert,
  Typography,
  Tag,
  List,
  Popconfirm,
  Row,
  Col,
  Divider,
  Tooltip,
} from 'antd'
import {
  CopyOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  ReloadOutlined,
  PlusOutlined,
  DeleteOutlined,
  KeyOutlined,
  LockOutlined,
  SafetyOutlined,
  ApiOutlined,
  CheckOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

const { Title, Text, Paragraph } = Typography
const { TextArea } = Input

interface IpWhitelist {
  id: string
  ip: string
  description: string
  created_at: string
}

const ApiKeys = () => {
  const { t } = useTranslation()
  const [apiKeyVisible, setApiKeyVisible] = useState(false)
  const [secretVisible, setSecretVisible] = useState(false)
  const [webhookModalVisible, setWebhookModalVisible] = useState(false)
  const [ipModalVisible, setIpModalVisible] = useState(false)
  const [form] = Form.useForm()
  const [ipForm] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [copiedField, setCopiedField] = useState<string | null>(null)

  // Mock data
  const [apiKey] = useState('mpk_test_1234567890abcdefghijklmnopqrstuvwxyz')
  const [apiSecret] = useState('msk_test_0987654321zyxwvutsrqponmlkjihgfedcba')
  const [webhookUrl, setWebhookUrl] = useState('https://api.example.com/webhook/payment')
  const [ipWhitelist, setIpWhitelist] = useState<IpWhitelist[]>([
    {
      id: '1',
      ip: '192.168.1.100',
      description: '生产服务器',
      created_at: new Date().toISOString(),
    },
    {
      id: '2',
      ip: '192.168.1.101',
      description: '测试服务器',
      created_at: new Date().toISOString(),
    },
  ])

  const handleCopy = async (text: string, type: string, fieldId: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedField(fieldId)
      message.success(t('apiKeys.copySuccess', { type }))
      setTimeout(() => setCopiedField(null), 2000)
    } catch (error) {
      message.error(t('apiKeys.copyFailed'))
    }
  }

  const handleRegenerateApiKey = () => {
    Modal.confirm({
      title: t('apiKeys.regenerateConfirm'),
      content: t('apiKeys.regenerateWarning'),
      okText: t('common.confirm'),
      cancelText: t('common.cancel'),
      okButtonProps: { danger: true },
      onOk: async () => {
        setLoading(true)
        try {
          // Mock API call
          await new Promise((resolve) => setTimeout(resolve, 1500))
          message.success(t('apiKeys.regenerateSuccess'))
        } catch (error) {
          message.error(t('apiKeys.regenerateFailed'))
        } finally {
          setLoading(false)
        }
      },
    })
  }

  const handleUpdateWebhook = async (values: any) => {
    setLoading(true)
    try {
      // Mock API call
      await new Promise((resolve) => setTimeout(resolve, 1000))
      setWebhookUrl(values.webhook_url)
      message.success(t('apiKeys.webhookUpdateSuccess'))
      setWebhookModalVisible(false)
      form.resetFields()
    } catch (error) {
      message.error(t('apiKeys.webhookUpdateFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleAddIp = async (values: any) => {
    setLoading(true)
    try {
      // Mock API call
      await new Promise((resolve) => setTimeout(resolve, 1000))
      const newIp: IpWhitelist = {
        id: Date.now().toString(),
        ip: values.ip,
        description: values.description || '',
        created_at: new Date().toISOString(),
      }
      setIpWhitelist([...ipWhitelist, newIp])
      message.success(t('apiKeys.ipAddSuccess'))
      setIpModalVisible(false)
      ipForm.resetFields()
    } catch (error) {
      message.error(t('apiKeys.ipAddFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteIp = async (id: string) => {
    try {
      // Mock API call
      await new Promise((resolve) => setTimeout(resolve, 500))
      setIpWhitelist(ipWhitelist.filter((item) => item.id !== id))
      message.success(t('apiKeys.ipDeleteSuccess'))
    } catch (error) {
      message.error(t('apiKeys.ipDeleteFailed'))
    }
  }

  const maskString = (str: string, visible: boolean) => {
    if (visible) return str
    return '*'.repeat(str.length)
  }

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>{t('apiKeys.title')}</Title>
        <Paragraph type="secondary" style={{ marginBottom: 0 }}>{t('apiKeys.subtitle')}</Paragraph>
      </div>

      <Alert
        message={t('apiKeys.securityNotice')}
        description={t('apiKeys.securityNoticeDesc')}
        type="warning"
        showIcon
        icon={<SafetyOutlined />}
        closable
        style={{ marginBottom: 16, borderRadius: 8 }}
      />

      {/* API Credentials */}
      <Card
        title={
          <Space>
            <KeyOutlined />
            {t('apiKeys.credentials')}
          </Space>
        }
        style={{ marginBottom: 16, borderRadius: 12 }}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          {/* API Key */}
          <div>
            <div style={{ marginBottom: 8 }}>
              <Text strong>{t('apiKeys.apiKey')}</Text>
              <Tag color="blue" style={{ marginLeft: 8, borderRadius: 12 }}>
                {t('apiKeys.public')}
              </Tag>
            </div>
            <Input.Group compact>
              <Input
                style={{ width: 'calc(100% - 120px)', borderRadius: '8px 0 0 8px' }}
                value={maskString(apiKey, apiKeyVisible)}
                readOnly
                prefix={<ApiOutlined />}
              />
              <Tooltip title={apiKeyVisible ? t('apiKeys.hide') : t('apiKeys.show')}>
                <Button
                  icon={apiKeyVisible ? <EyeInvisibleOutlined /> : <EyeOutlined />}
                  onClick={() => setApiKeyVisible(!apiKeyVisible)}
                  style={{ borderRadius: 0 }}
                />
              </Tooltip>
              <Tooltip title={copiedField === 'apiKey' ? t('common.copied') : t('apiKeys.copy')}>
                <Button
                  type={copiedField === 'apiKey' ? 'primary' : 'default'}
                  icon={copiedField === 'apiKey' ? <CheckOutlined /> : <CopyOutlined />}
                  onClick={() => handleCopy(apiKey, t('apiKeys.apiKey'), 'apiKey')}
                  style={{ borderRadius: '0 8px 8px 0' }}
                />
              </Tooltip>
            </Input.Group>
            <Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0, fontSize: 12 }}>
              {t('apiKeys.apiKeyDesc')}
            </Paragraph>
          </div>

          {/* API Secret */}
          <div>
            <div style={{ marginBottom: 8 }}>
              <Text strong>{t('apiKeys.apiSecret')}</Text>
              <Tag color="red" style={{ marginLeft: 8, borderRadius: 12 }}>
                {t('apiKeys.private')}
              </Tag>
            </div>
            <Input.Group compact>
              <Input
                style={{ width: 'calc(100% - 120px)', borderRadius: '8px 0 0 8px' }}
                value={maskString(apiSecret, secretVisible)}
                readOnly
                prefix={<LockOutlined />}
              />
              <Tooltip title={secretVisible ? t('apiKeys.hide') : t('apiKeys.show')}>
                <Button
                  icon={secretVisible ? <EyeInvisibleOutlined /> : <EyeOutlined />}
                  onClick={() => setSecretVisible(!secretVisible)}
                  style={{ borderRadius: 0 }}
                />
              </Tooltip>
              <Tooltip title={copiedField === 'apiSecret' ? t('common.copied') : t('apiKeys.copy')}>
                <Button
                  type={copiedField === 'apiSecret' ? 'primary' : 'default'}
                  icon={copiedField === 'apiSecret' ? <CheckOutlined /> : <CopyOutlined />}
                  onClick={() => handleCopy(apiSecret, t('apiKeys.apiSecret'), 'apiSecret')}
                  style={{ borderRadius: '0 8px 8px 0' }}
                />
              </Tooltip>
            </Input.Group>
            <Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0, fontSize: 12 }}>
              {t('apiKeys.apiSecretDesc')}
            </Paragraph>
          </div>

          <Divider style={{ margin: '16px 0' }} />

          <Button
            type="primary"
            danger
            icon={<ReloadOutlined />}
            onClick={handleRegenerateApiKey}
            loading={loading}
            style={{ borderRadius: 8 }}
          >
            {t('apiKeys.regenerate')}
          </Button>
        </Space>
      </Card>

      {/* Webhook Configuration */}
      <Card
        title={t('apiKeys.webhookConfig')}
        extra={
          <Button onClick={() => setWebhookModalVisible(true)} style={{ borderRadius: 8 }}>
            {t('apiKeys.edit')}
          </Button>
        }
        style={{ marginBottom: 16, borderRadius: 12 }}
      >
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div>
            <Text strong>{t('apiKeys.webhookUrl')}</Text>
            <Input.Group compact style={{ marginTop: 8 }}>
              <Input
                value={webhookUrl}
                readOnly
                style={{ width: 'calc(100% - 40px)', borderRadius: '8px 0 0 8px' }}
              />
              <Tooltip title={copiedField === 'webhook' ? t('common.copied') : t('apiKeys.copy')}>
                <Button
                  type={copiedField === 'webhook' ? 'primary' : 'default'}
                  icon={copiedField === 'webhook' ? <CheckOutlined /> : <CopyOutlined />}
                  onClick={() => handleCopy(webhookUrl, t('apiKeys.webhookUrl'), 'webhook')}
                  style={{ borderRadius: '0 8px 8px 0' }}
                />
              </Tooltip>
            </Input.Group>
          </div>
          <Alert
            message={t('apiKeys.webhookNotice')}
            description={t('apiKeys.webhookNoticeDesc')}
            type="info"
            showIcon
            style={{ borderRadius: 8 }}
          />
        </Space>
      </Card>

      {/* IP Whitelist */}
      <Card
        title={t('apiKeys.ipWhitelist')}
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setIpModalVisible(true)}
            style={{ borderRadius: 8 }}
          >
            {t('apiKeys.addIp')}
          </Button>
        }
        style={{ borderRadius: 12 }}
      >
        <Alert
          message={t('apiKeys.ipNotice')}
          description={t('apiKeys.ipNoticeDesc')}
          type="info"
          showIcon
          closable
          style={{ marginBottom: 16, borderRadius: 8 }}
        />
        <List
          dataSource={ipWhitelist}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Popconfirm
                  title={t('apiKeys.ipDeleteConfirm')}
                  onConfirm={() => handleDeleteIp(item.id)}
                  okText={t('common.confirm')}
                  cancelText={t('common.cancel')}
                >
                  <Button type="link" danger icon={<DeleteOutlined />} style={{ borderRadius: 8 }}>
                    {t('common.delete')}
                  </Button>
                </Popconfirm>,
              ]}
              style={{ borderRadius: 8, padding: '16px', marginBottom: 8, border: '1px solid #f0f0f0' }}
            >
              <List.Item.Meta
                title={
                  <Space>
                    <Tag color="blue" style={{ fontFamily: 'monospace', borderRadius: 12 }}>
                      {item.ip}
                    </Tag>
                    {item.description && <Text type="secondary">- {item.description}</Text>}
                  </Space>
                }
                description={
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {t('apiKeys.addedAt', {
                      date: new Date(item.created_at).toLocaleString(),
                    })}
                  </Text>
                }
              />
            </List.Item>
          )}
        />
      </Card>

      {/* Webhook Modal */}
      <Modal
        title={t('apiKeys.editWebhook')}
        open={webhookModalVisible}
        onCancel={() => {
          setWebhookModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        confirmLoading={loading}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleUpdateWebhook}
          initialValues={{ webhook_url: webhookUrl }}
        >
          <Form.Item
            label={t('apiKeys.webhookUrl')}
            name="webhook_url"
            rules={[
              { required: true, message: t('apiKeys.webhookUrlRequired') },
              { type: 'url', message: t('apiKeys.webhookUrlInvalid') },
            ]}
          >
            <Input placeholder="https://api.example.com/webhook/payment" />
          </Form.Item>
          <Alert
            message={t('apiKeys.webhookTip')}
            description={t('apiKeys.webhookTipDesc')}
            type="info"
            showIcon
          />
        </Form>
      </Modal>

      {/* IP Whitelist Modal */}
      <Modal
        title={t('apiKeys.addIp')}
        open={ipModalVisible}
        onCancel={() => {
          setIpModalVisible(false)
          ipForm.resetFields()
        }}
        onOk={() => ipForm.submit()}
        confirmLoading={loading}
      >
        <Form form={ipForm} layout="vertical" onFinish={handleAddIp}>
          <Form.Item
            label={t('apiKeys.ipAddress')}
            name="ip"
            rules={[
              { required: true, message: t('apiKeys.ipRequired') },
              {
                pattern:
                  /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
                message: t('apiKeys.ipInvalid'),
              },
            ]}
          >
            <Input placeholder="192.168.1.100" />
          </Form.Item>
          <Form.Item label={t('apiKeys.description')} name="description">
            <TextArea
              placeholder={t('apiKeys.descriptionPlaceholder')}
              rows={3}
              maxLength={100}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ApiKeys

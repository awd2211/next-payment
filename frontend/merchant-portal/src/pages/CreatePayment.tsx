import { useState } from 'react'
import {
  Card,
  Form,
  Input,
  InputNumber,
  Select,
  Button,
  Space,
  message,
  Modal,
  QRCode,
  Typography,
  Steps,
  Row,
  Col,
  Descriptions,
  Tag,
  Tooltip,
  Alert,
  Radio,
  Statistic,
} from 'antd'
import {
  CopyOutlined,
  CheckCircleOutlined,
  QrcodeOutlined,
  LinkOutlined,
  CheckOutlined,
  DollarOutlined,
  CreditCardOutlined,
  WalletOutlined,
  PayCircleOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'

const { Title, Text, Paragraph } = Typography
const { TextArea } = Input

interface PaymentResult {
  order_no: string
  payment_no: string
  amount: number
  currency: string
  channel: string
  payment_url: string
  qr_code_url: string
  expires_at: string
}

const CreatePayment = () => {
  const { t } = useTranslation()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [currentStep, setCurrentStep] = useState(0)
  const [paymentResult, setPaymentResult] = useState<PaymentResult | null>(null)
  const [resultModalVisible, setResultModalVisible] = useState(false)
  const [copiedField, setCopiedField] = useState<string | null>(null)
  const [previewAmount, setPreviewAmount] = useState<number>(0)
  const [previewCurrency, setPreviewCurrency] = useState<string>('CNY')

  const handleSubmit = async (values: any) => {
    setLoading(true)
    try {
      // Mock API call
      await new Promise((resolve) => setTimeout(resolve, 1500))

      const mockResult: PaymentResult = {
        order_no: `ORD${Date.now()}`,
        payment_no: `PAY${Date.now()}`,
        amount: values.amount,
        currency: values.currency,
        channel: values.channel,
        payment_url: `https://payment.example.com/pay/${Date.now()}`,
        qr_code_url: `https://payment.example.com/qr/${Date.now()}`,
        expires_at: dayjs().add(30, 'minute').toISOString(),
      }

      setPaymentResult(mockResult)
      setResultModalVisible(true)
      setCurrentStep(2)
      message.success(t('createPayment.createSuccess'))
    } catch (error) {
      message.error(t('createPayment.createFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = async (text: string, type: string, fieldId: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedField(fieldId)
      message.success(t('createPayment.copySuccess', { type }))
      setTimeout(() => setCopiedField(null), 2000)
    } catch (error) {
      message.error(t('createPayment.copyFailed'))
    }
  }

  const handleValuesChange = (changedValues: any, allValues: any) => {
    if (changedValues.amount !== undefined) {
      setPreviewAmount(changedValues.amount || 0)
    }
    if (changedValues.currency !== undefined) {
      setPreviewCurrency(changedValues.currency)
    }
    if (Object.keys(changedValues).length > 0 && currentStep === 0) {
      setCurrentStep(1)
    }
  }

  const handleReset = () => {
    form.resetFields()
    setCurrentStep(0)
    setPaymentResult(null)
  }

  const handleCreateAnother = () => {
    handleReset()
    setResultModalVisible(false)
  }

  const steps = [
    {
      title: t('createPayment.step1Title'),
      description: t('createPayment.step1Desc'),
    },
    {
      title: t('createPayment.step2Title'),
      description: t('createPayment.step2Desc'),
    },
    {
      title: t('createPayment.step3Title'),
      description: t('createPayment.step3Desc'),
    },
  ]

  const channelOptions = [
    {
      value: 'stripe',
      label: 'Stripe',
      color: 'blue',
      icon: <CreditCardOutlined />,
      recommended: true,
    },
    {
      value: 'paypal',
      label: 'PayPal',
      color: 'cyan',
      icon: <PayCircleOutlined />,
      recommended: false,
    },
    {
      value: 'alipay',
      label: '支付宝',
      color: 'green',
      icon: <WalletOutlined />,
      recommended: false,
    },
    {
      value: 'wechat',
      label: '微信支付',
      color: 'orange',
      icon: <PayCircleOutlined />,
      recommended: false,
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>{t('createPayment.title')}</Title>
        <Paragraph type="secondary" style={{ marginBottom: 0 }}>{t('createPayment.subtitle')}</Paragraph>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card style={{ borderRadius: 12 }}>
            <Steps current={currentStep} items={steps} style={{ marginBottom: 32 }} />

            <Form
              form={form}
              layout="vertical"
              onFinish={handleSubmit}
              initialValues={{
                currency: 'CNY',
                channel: 'stripe',
              }}
              onValuesChange={handleValuesChange}
            >
          <Row gutter={16}>
            <Col xs={24} lg={12}>
              <Form.Item
                label={t('createPayment.merchantOrderNo')}
                name="merchant_order_no"
                rules={[
                  { required: true, message: t('createPayment.merchantOrderNoRequired') },
                  {
                    pattern: /^[A-Za-z0-9_-]+$/,
                    message: t('createPayment.merchantOrderNoPattern'),
                  },
                ]}
              >
                <Input
                  placeholder={t('createPayment.merchantOrderNoPlaceholder')}
                  maxLength={64}
                  style={{ borderRadius: 8 }}
                />
              </Form.Item>
            </Col>

            <Col xs={24} lg={12}>
              <Form.Item
                label={t('createPayment.productName')}
                name="product_name"
                rules={[{ required: true, message: t('createPayment.productNameRequired') }]}
              >
                <Input
                  placeholder={t('createPayment.productNamePlaceholder')}
                  maxLength={128}
                  style={{ borderRadius: 8 }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} sm={12} lg={8}>
              <Form.Item
                label={t('createPayment.amount')}
                name="amount"
                rules={[
                  { required: true, message: t('createPayment.amountRequired') },
                  { type: 'number', min: 0.01, message: t('createPayment.amountMin') },
                ]}
              >
                <InputNumber
                  style={{ width: '100%', borderRadius: 8 }}
                  placeholder="0.00"
                  precision={2}
                  min={0.01}
                  max={999999.99}
                  prefix={<DollarOutlined />}
                />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} lg={8}>
              <Form.Item
                label={t('createPayment.currency')}
                name="currency"
                rules={[{ required: true, message: t('createPayment.currencyRequired') }]}
              >
                <Select style={{ borderRadius: 8 }}>
                  <Select.Option value="CNY">CNY - 人民币</Select.Option>
                  <Select.Option value="USD">USD - 美元</Select.Option>
                  <Select.Option value="EUR">EUR - 欧元</Select.Option>
                  <Select.Option value="GBP">GBP - 英镑</Select.Option>
                  <Select.Option value="JPY">JPY - 日元</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label={t('createPayment.channel')}
            name="channel"
            rules={[{ required: true, message: t('createPayment.channelRequired') }]}
          >
            <Radio.Group style={{ width: '100%' }}>
              <Row gutter={[12, 12]}>
                {channelOptions.map((option) => (
                  <Col xs={12} sm={12} md={6} key={option.value}>
                    <Radio.Button
                      value={option.value}
                      style={{
                        width: '100%',
                        height: '80px',
                        borderRadius: 12,
                        textAlign: 'center',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        position: 'relative',
                      }}
                    >
                      <div>
                        <div style={{ fontSize: 24, marginBottom: 8 }}>{option.icon}</div>
                        <div style={{ fontWeight: 500 }}>{option.label}</div>
                        {option.recommended && (
                          <Tag
                            color="green"
                            style={{
                              position: 'absolute',
                              top: 4,
                              right: 4,
                              fontSize: 10,
                              borderRadius: 8,
                            }}
                          >
                            推荐
                          </Tag>
                        )}
                      </div>
                    </Radio.Button>
                  </Col>
                ))}
              </Row>
            </Radio.Group>
          </Form.Item>

          <Form.Item label={t('createPayment.description')} name="description">
            <TextArea
              placeholder={t('createPayment.descriptionPlaceholder')}
              rows={3}
              maxLength={512}
              showCount
              style={{ borderRadius: 8 }}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col xs={24} lg={12}>
              <Form.Item
                label={t('createPayment.notifyUrl')}
                name="notify_url"
                rules={[
                  { required: true, message: t('createPayment.notifyUrlRequired') },
                  { type: 'url', message: t('createPayment.urlInvalid') },
                ]}
                tooltip={t('createPayment.notifyUrlTooltip')}
              >
                <Input
                  placeholder="https://your-domain.com/api/payment/notify"
                  prefix={<LinkOutlined />}
                  style={{ borderRadius: 8 }}
                />
              </Form.Item>
            </Col>

            <Col xs={24} lg={12}>
              <Form.Item
                label={t('createPayment.returnUrl')}
                name="return_url"
                rules={[
                  { required: true, message: t('createPayment.returnUrlRequired') },
                  { type: 'url', message: t('createPayment.urlInvalid') },
                ]}
                tooltip={t('createPayment.returnUrlTooltip')}
              >
                <Input
                  placeholder="https://your-domain.com/order/success"
                  prefix={<LinkOutlined />}
                  style={{ borderRadius: 8 }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                size="large"
                style={{ borderRadius: 8 }}
              >
                {t('createPayment.createButton')}
              </Button>
              <Button onClick={handleReset} style={{ borderRadius: 8 }}>
                {t('common.reset')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </Col>

    <Col xs={24} lg={8}>
      <Card
        title="金额预览"
        style={{ borderRadius: 12, position: 'sticky', top: 24 }}
      >
        <Statistic
          title="待支付金额"
          value={previewAmount}
          precision={2}
          prefix={previewCurrency}
          valueStyle={{
            fontSize: 36,
            fontWeight: 700,
            color: previewAmount > 0 ? '#1890ff' : '#999',
          }}
        />
        <Alert
          message="提示"
          description="输入金额后,这里会实时显示支付金额预览"
          type="info"
          showIcon
          style={{ marginTop: 16, borderRadius: 8 }}
        />
      </Card>
    </Col>
  </Row>

      {/* Result Modal */}
      <Modal
        title={
          <Space>
            <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 24 }} />
            <span>{t('createPayment.resultTitle')}</span>
          </Space>
        }
        open={resultModalVisible}
        onCancel={() => setResultModalVisible(false)}
        width={800}
        footer={[
          <Button
            key="another"
            type="primary"
            onClick={handleCreateAnother}
            style={{ borderRadius: 8 }}
          >
            {t('createPayment.createAnother')}
          </Button>,
          <Button
            key="close"
            onClick={() => setResultModalVisible(false)}
            style={{ borderRadius: 8 }}
          >
            {t('common.cancel')}
          </Button>,
        ]}
      >
        {paymentResult && (
          <div>
            <Descriptions bordered column={2} style={{ marginBottom: 24, borderRadius: 8 }}>
              <Descriptions.Item label={t('createPayment.orderNo')} span={2}>
                <Space>
                  <Text code style={{ fontFamily: 'monospace' }}>
                    {paymentResult.order_no}
                  </Text>
                  <Tooltip title={copiedField === 'orderNo' ? t('common.copied') : t('apiKeys.copy')}>
                    <Button
                      type={copiedField === 'orderNo' ? 'primary' : 'link'}
                      size="small"
                      icon={copiedField === 'orderNo' ? <CheckOutlined /> : <CopyOutlined />}
                      onClick={() => handleCopy(paymentResult.order_no, t('createPayment.orderNo'), 'orderNo')}
                      style={{ borderRadius: 8 }}
                    />
                  </Tooltip>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label={t('createPayment.paymentNo')} span={2}>
                <Space>
                  <Text code style={{ fontFamily: 'monospace' }}>
                    {paymentResult.payment_no}
                  </Text>
                  <Tooltip title={copiedField === 'paymentNo' ? t('common.copied') : t('apiKeys.copy')}>
                    <Button
                      type={copiedField === 'paymentNo' ? 'primary' : 'link'}
                      size="small"
                      icon={copiedField === 'paymentNo' ? <CheckOutlined /> : <CopyOutlined />}
                      onClick={() =>
                        handleCopy(paymentResult.payment_no, t('createPayment.paymentNo'), 'paymentNo')
                      }
                      style={{ borderRadius: 8 }}
                    />
                  </Tooltip>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label={t('createPayment.amount')}>
                <Text strong style={{ fontSize: 16, color: '#1890ff' }}>
                  {paymentResult.currency} {paymentResult.amount.toFixed(2)}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('createPayment.channel')}>
                <Tag color="blue" style={{ borderRadius: 12 }}>
                  {paymentResult.channel.toUpperCase()}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('createPayment.expiresAt')} span={2}>
                {dayjs(paymentResult.expires_at).format('YYYY-MM-DD HH:mm:ss')}
                <Text type="secondary" style={{ marginLeft: 8 }}>
                  ({dayjs(paymentResult.expires_at).fromNow()})
                </Text>
              </Descriptions.Item>
            </Descriptions>

            <Row gutter={16}>
              <Col xs={24} md={12}>
                <Card
                  title={
                    <Space>
                      <LinkOutlined />
                      {t('createPayment.paymentUrl')}
                    </Space>
                  }
                  size="small"
                  style={{ borderRadius: 12 }}
                >
                  <Paragraph
                    ellipsis={{ rows: 2, expandable: true }}
                    style={{ marginBottom: 8, fontSize: 12 }}
                  >
                    {paymentResult.payment_url}
                  </Paragraph>
                  <Tooltip title={copiedField === 'paymentUrl' ? t('common.copied') : t('createPayment.copyUrl')}>
                    <Button
                      type={copiedField === 'paymentUrl' ? 'default' : 'primary'}
                      block
                      icon={copiedField === 'paymentUrl' ? <CheckOutlined /> : <CopyOutlined />}
                      onClick={() =>
                        handleCopy(paymentResult.payment_url, t('createPayment.paymentUrl'), 'paymentUrl')
                      }
                      style={{ borderRadius: 8 }}
                    >
                      {copiedField === 'paymentUrl' ? '已复制' : t('createPayment.copyUrl')}
                    </Button>
                  </Tooltip>
                </Card>
              </Col>

              <Col xs={24} md={12}>
                <Card
                  title={
                    <Space>
                      <QrcodeOutlined />
                      {t('createPayment.qrCode')}
                    </Space>
                  }
                  size="small"
                  style={{ borderRadius: 12 }}
                >
                  <div style={{ textAlign: 'center' }}>
                    <QRCode value={paymentResult.payment_url} size={200} />
                    <Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0, fontSize: 12 }}>
                      {t('createPayment.qrCodeHint')}
                    </Paragraph>
                  </div>
                </Card>
              </Col>
            </Row>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default CreatePayment

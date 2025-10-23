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
} from 'antd'
import {
  CopyOutlined,
  CheckCircleOutlined,
  QrcodeOutlined,
  LinkOutlined,
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

  const handleCopy = async (text: string, type: string) => {
    try {
      await navigator.clipboard.writeText(text)
      message.success(t('createPayment.copySuccess', { type }))
    } catch (error) {
      message.error(t('createPayment.copyFailed'))
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

  return (
    <div>
      <Title level={2}>{t('createPayment.title')}</Title>
      <Paragraph type="secondary">{t('createPayment.subtitle')}</Paragraph>

      <Card style={{ marginTop: 24 }}>
        <Steps current={currentStep} items={steps} style={{ marginBottom: 32 }} />

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            currency: 'CNY',
            channel: 'stripe',
          }}
          onValuesChange={() => setCurrentStep(1)}
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
                />
              </Form.Item>
            </Col>

            <Col xs={24} lg={12}>
              <Form.Item
                label={t('createPayment.productName')}
                name="product_name"
                rules={[{ required: true, message: t('createPayment.productNameRequired') }]}
              >
                <Input placeholder={t('createPayment.productNamePlaceholder')} maxLength={128} />
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
                  style={{ width: '100%' }}
                  placeholder="0.00"
                  precision={2}
                  min={0.01}
                  max={999999.99}
                />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} lg={8}>
              <Form.Item
                label={t('createPayment.currency')}
                name="currency"
                rules={[{ required: true, message: t('createPayment.currencyRequired') }]}
              >
                <Select>
                  <Select.Option value="CNY">CNY - 人民币</Select.Option>
                  <Select.Option value="USD">USD - 美元</Select.Option>
                  <Select.Option value="EUR">EUR - 欧元</Select.Option>
                  <Select.Option value="GBP">GBP - 英镑</Select.Option>
                  <Select.Option value="JPY">JPY - 日元</Select.Option>
                </Select>
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} lg={8}>
              <Form.Item
                label={t('createPayment.channel')}
                name="channel"
                rules={[{ required: true, message: t('createPayment.channelRequired') }]}
              >
                <Select>
                  <Select.Option value="stripe">
                    <Space>
                      <Tag color="blue">Stripe</Tag>
                      {t('createPayment.channelRecommended')}
                    </Space>
                  </Select.Option>
                  <Select.Option value="paypal">
                    <Tag color="cyan">PayPal</Tag>
                  </Select.Option>
                  <Select.Option value="alipay">
                    <Tag color="green">支付宝</Tag>
                  </Select.Option>
                  <Select.Option value="wechat">
                    <Tag color="orange">微信支付</Tag>
                  </Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label={t('createPayment.description')} name="description">
            <TextArea
              placeholder={t('createPayment.descriptionPlaceholder')}
              rows={3}
              maxLength={512}
              showCount
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
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading} size="large">
                {t('createPayment.createButton')}
              </Button>
              <Button onClick={handleReset}>{t('common.reset')}</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

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
          <Button key="another" type="primary" onClick={handleCreateAnother}>
            {t('createPayment.createAnother')}
          </Button>,
          <Button key="close" onClick={() => setResultModalVisible(false)}>
            {t('common.cancel')}
          </Button>,
        ]}
      >
        {paymentResult && (
          <div>
            <Descriptions bordered column={2} style={{ marginBottom: 24 }}>
              <Descriptions.Item label={t('createPayment.orderNo')} span={2}>
                <Space>
                  <Text copyable>{paymentResult.order_no}</Text>
                  <Button
                    type="link"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => handleCopy(paymentResult.order_no, t('createPayment.orderNo'))}
                  />
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label={t('createPayment.paymentNo')} span={2}>
                <Space>
                  <Text copyable>{paymentResult.payment_no}</Text>
                  <Button
                    type="link"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() =>
                      handleCopy(paymentResult.payment_no, t('createPayment.paymentNo'))
                    }
                  />
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label={t('createPayment.amount')}>
                <Text strong style={{ fontSize: 16 }}>
                  {paymentResult.currency} {paymentResult.amount.toFixed(2)}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('createPayment.channel')}>
                <Tag color="blue">{paymentResult.channel.toUpperCase()}</Tag>
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
                >
                  <Paragraph
                    copyable
                    ellipsis={{ rows: 2, expandable: true }}
                    style={{ marginBottom: 8 }}
                  >
                    {paymentResult.payment_url}
                  </Paragraph>
                  <Button
                    type="primary"
                    block
                    icon={<CopyOutlined />}
                    onClick={() =>
                      handleCopy(paymentResult.payment_url, t('createPayment.paymentUrl'))
                    }
                  >
                    {t('createPayment.copyUrl')}
                  </Button>
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
                >
                  <div style={{ textAlign: 'center' }}>
                    <QRCode value={paymentResult.payment_url} size={200} />
                    <Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0 }}>
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

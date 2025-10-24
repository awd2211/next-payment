import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  Card,
  Form,
  Input,
  Button,
  Radio,
  Checkbox,
  Divider,
  Space,
  Typography,
  Alert,
  Spin,
  Row,
  Col,
  message,
} from 'antd'
import {
  CreditCardOutlined,
  SafetyOutlined,
  LockOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { loadStripe } from '@stripe/stripe-js'
import { Elements } from '@stripe/react-stripe-js'
import { cashierApi } from '../services/cashierApi'
import type { CashierSession, CashierConfig } from '../types'
import {
  validateCardNumber,
  validateExpiryDate,
  validateCVV,
  formatCardNumber,
  formatExpiryDate,
  formatAmount,
} from '../utils/validation'
import StripePaymentForm from '../components/StripePaymentForm'

const { Title, Text } = Typography

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLIC_KEY || '')

const Checkout = () => {
  const [searchParams] = useSearchParams()
  const { t, i18n } = useTranslation()
  const [form] = Form.useForm()

  const sessionToken = searchParams.get('token')

  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [session, setSession] = useState<CashierSession | null>(null)
  const [config, setConfig] = useState<CashierConfig | null>(null)
  const [selectedChannel, setSelectedChannel] = useState<string>('')
  const [paymentStatus, setPaymentStatus] = useState<'pending' | 'success' | 'failed'>('pending')
  const [errorMessage, setErrorMessage] = useState<string>('')
  const [pageLoadTime] = useState(Date.now())

  useEffect(() => {
    if (!sessionToken) {
      setErrorMessage(t('cashier.invalid_session'))
      setLoading(false)
      return
    }

    loadSession()
  }, [sessionToken])

  const loadSession = async () => {
    try {
      setLoading(true)
      const sessionData = await cashierApi.getSession(sessionToken!)
      setSession(sessionData)

      // 加载商户配置
      const configData = await cashierApi.getConfig(sessionData.merchant_id)
      setConfig(configData)

      // 设置语言
      if (configData.default_language) {
        i18n.changeLanguage(configData.default_language)
      }

      // 设置默认支付渠道
      if (configData.enabled_channels.length > 0) {
        setSelectedChannel(
          configData.default_channel || configData.enabled_channels[0]
        )
      }

      // 记录页面加载日志
      const loadDuration = Date.now() - pageLoadTime
      await cashierApi.recordLog({
        session_token: sessionToken!,
        user_agent: navigator.userAgent,
        device_type: /mobile/i.test(navigator.userAgent) ? 'mobile' : 'desktop',
        page_load_time: loadDuration,
      })
    } catch (error: any) {
      console.error('Failed to load session:', error)
      if (error.response?.status === 404) {
        setErrorMessage(t('cashier.invalid_session'))
      } else if (error.message?.includes('expired')) {
        setErrorMessage(t('cashier.session_expired_message'))
      } else {
        setErrorMessage(t('errors.network_error'))
      }
    } finally {
      setLoading(false)
    }
  }

  const handleChannelChange = async (channel: string) => {
    setSelectedChannel(channel)
    await cashierApi.recordLog({
      session_token: sessionToken!,
      selected_channel: channel,
    })
  }

  const handleSubmit = async (values: any) => {
    if (!session || !sessionToken) return

    try {
      setSubmitting(true)

      // 记录表单填写完成
      await cashierApi.recordLog({
        session_token: sessionToken,
        form_filled: true,
        selected_channel: selectedChannel,
        time_to_submit: Date.now() - pageLoadTime,
      })

      // 对于Stripe，使用Stripe Elements
      if (selectedChannel === 'stripe') {
        // Stripe Elements会通过子组件处理
        return
      }

      // 其他支付渠道的处理
      const paymentData = {
        session_token: sessionToken,
        channel: selectedChannel,
        payment_method: 'card',
        card_data: {
          number: values.cardNumber.replace(/\s/g, ''),
          exp_month: values.expiryDate.split('/')[0],
          exp_year: '20' + values.expiryDate.split('/')[1],
          cvv: values.cvv,
          holder_name: values.cardholderName,
        },
      }

      const result = await cashierApi.createPayment(paymentData)

      // 记录支付提交
      await cashierApi.recordLog({
        session_token: sessionToken,
        payment_submitted: true,
      })

      if (result.status === 'success') {
        setPaymentStatus('success')
        message.success(t('cashier.payment_success'))

        // 重定向到成功页面
        if (config?.success_redirect_url) {
          setTimeout(() => {
            window.location.href = config.success_redirect_url + '?payment_no=' + result.payment_no
          }, 2000)
        }
      } else {
        throw new Error('Payment failed')
      }
    } catch (error: any) {
      console.error('Payment error:', error)
      setPaymentStatus('failed')
      setErrorMessage(error.response?.data?.message || t('errors.payment_error'))

      // 记录错误
      await cashierApi.recordLog({
        session_token: sessionToken,
        error_message: error.message,
      })

      message.error(t('errors.payment_error'))
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" tip={t('common.loading')} />
      </div>
    )
  }

  if (errorMessage && !session) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', padding: 24 }}>
        <Card style={{ maxWidth: 500 }}>
          <Alert
            message={t('common.error')}
            description={errorMessage}
            type="error"
            showIcon
            icon={<CloseCircleOutlined />}
          />
        </Card>
      </div>
    )
  }

  if (paymentStatus === 'success') {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', padding: 24 }}>
        <Card style={{ maxWidth: 500, textAlign: 'center' }}>
          <CheckCircleOutlined style={{ fontSize: 64, color: '#52c41a', marginBottom: 24 }} />
          <Title level={3}>{t('cashier.payment_success')}</Title>
          <Text type="secondary">{t('cashier.powered_by')} Payment Platform</Text>
        </Card>
      </div>
    )
  }

  const themeColor = config?.theme_color || '#1890ff'

  return (
    <div
      style={{
        minHeight: '100vh',
        background: config?.background_image_url
          ? `url(${config.background_image_url}) center/cover`
          : '#f0f2f5',
        padding: '40px 24px',
      }}
    >
      <Row justify="center">
        <Col xs={24} sm={20} md={16} lg={12} xl={10}>
          <Card
            style={{
              boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
              borderRadius: 8,
            }}
          >
            {/* 头部 */}
            <div style={{ textAlign: 'center', marginBottom: 24 }}>
              {config?.logo_url && (
                <img
                  src={config.logo_url}
                  alt="Merchant Logo"
                  style={{ maxHeight: 60, marginBottom: 16 }}
                />
              )}
              <Title level={3}>{t('cashier.title')}</Title>
              <Space>
                <LockOutlined style={{ color: themeColor }} />
                <Text type="secondary">{t('cashier.ssl_encrypted')}</Text>
              </Space>
            </div>

            {/* 订单摘要 */}
            <Card
              size="small"
              style={{ backgroundColor: '#fafafa', marginBottom: 24 }}
              title={t('cashier.order_summary')}
            >
              <Space direction="vertical" style={{ width: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Text strong>{t('cashier.order_no')}:</Text>
                  <Text>{session?.order_no}</Text>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Text strong>{t('cashier.description')}:</Text>
                  <Text>{session?.description}</Text>
                </div>
                <Divider style={{ margin: '8px 0' }} />
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Title level={4} style={{ margin: 0 }}>{t('cashier.amount')}:</Title>
                  <Title level={4} style={{ margin: 0, color: themeColor }}>
                    {formatAmount(session?.amount || 0, session?.currency || 'USD')}
                  </Title>
                </div>
              </Space>
            </Card>

            {/* 支付方式选择 */}
            {config && config.enabled_channels.length > 1 && config.allow_channel_switch && (
              <div style={{ marginBottom: 24 }}>
                <Text strong>{t('cashier.select_payment_method')}</Text>
                <Radio.Group
                  value={selectedChannel}
                  onChange={(e) => handleChannelChange(e.target.value)}
                  style={{ width: '100%', marginTop: 12 }}
                >
                  <Space direction="vertical" style={{ width: '100%' }}>
                    {config.enabled_channels.map((channel) => (
                      <Radio key={channel} value={channel} style={{ width: '100%' }}>
                        <CreditCardOutlined /> {channel.toUpperCase()}
                      </Radio>
                    ))}
                  </Space>
                </Radio.Group>
              </div>
            )}

            {/* 支付表单 */}
            {selectedChannel === 'stripe' ? (
              <Elements stripe={stripePromise}>
                <StripePaymentForm
                  session={session!}
                  sessionToken={sessionToken!}
                  config={config!}
                  onSuccess={() => setPaymentStatus('success')}
                  onError={(error) => {
                    setPaymentStatus('failed')
                    setErrorMessage(error)
                  }}
                />
              </Elements>
            ) : (
              <Form
                form={form}
                layout="vertical"
                onFinish={handleSubmit}
                autoComplete="off"
              >
                <Form.Item
                  name="cardNumber"
                  label={t('cashier.card_number')}
                  rules={[
                    { required: true, message: t('errors.required_field') },
                    {
                      validator: (_, value) =>
                        validateCardNumber(value)
                          ? Promise.resolve()
                          : Promise.reject(t('errors.invalid_card_number')),
                    },
                  ]}
                >
                  <Input
                    prefix={<CreditCardOutlined />}
                    placeholder="1234 5678 9012 3456"
                    maxLength={19}
                    onChange={(e) => {
                      const formatted = formatCardNumber(e.target.value)
                      form.setFieldValue('cardNumber', formatted)
                    }}
                  />
                </Form.Item>

                <Form.Item
                  name="cardholderName"
                  label={t('cashier.cardholder_name')}
                  rules={[{ required: true, message: t('errors.required_field') }]}
                >
                  <Input placeholder="John Doe" />
                </Form.Item>

                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item
                      name="expiryDate"
                      label={t('cashier.expiry_date')}
                      rules={[
                        { required: true, message: t('errors.required_field') },
                        {
                          validator: (_, value) =>
                            validateExpiryDate(value)
                              ? Promise.resolve()
                              : Promise.reject(t('errors.invalid_expiry')),
                        },
                      ]}
                    >
                      <Input
                        placeholder="MM/YY"
                        maxLength={5}
                        onChange={(e) => {
                          const formatted = formatExpiryDate(e.target.value)
                          form.setFieldValue('expiryDate', formatted)
                        }}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      name="cvv"
                      label={t('cashier.cvv')}
                      rules={[
                        { required: config?.require_cvv, message: t('errors.required_field') },
                        {
                          validator: (_, value) =>
                            !value || validateCVV(value)
                              ? Promise.resolve()
                              : Promise.reject(t('errors.invalid_cvv')),
                        },
                      ]}
                    >
                      <Input
                        prefix={<SafetyOutlined />}
                        placeholder="123"
                        maxLength={4}
                        type="password"
                      />
                    </Form.Item>
                  </Col>
                </Row>

                <Form.Item name="email" label={t('cashier.email')}>
                  <Input type="email" placeholder="john@example.com" />
                </Form.Item>

                <Form.Item name="saveCard" valuePropName="checked">
                  <Checkbox>{t('cashier.save_card')}</Checkbox>
                </Form.Item>

                {errorMessage && paymentStatus === 'failed' && (
                  <Alert
                    message={errorMessage}
                    type="error"
                    closable
                    style={{ marginBottom: 16 }}
                  />
                )}

                <Button
                  type="primary"
                  htmlType="submit"
                  size="large"
                  block
                  loading={submitting}
                  style={{ backgroundColor: themeColor, borderColor: themeColor }}
                >
                  {submitting ? t('common.processing') : t('common.pay_now')}
                </Button>
              </Form>
            )}

            {/* 底部 */}
            <div style={{ textAlign: 'center', marginTop: 24 }}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {t('cashier.powered_by')} Payment Platform
              </Text>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 自定义CSS */}
      {config?.custom_css && <style>{config.custom_css}</style>}
    </div>
  )
}

export default Checkout

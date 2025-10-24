import { useState } from 'react'
import { Button, Form, Input, Checkbox, Alert, message } from 'antd'
import { CardElement, useStripe, useElements } from '@stripe/react-stripe-js'
import type { StripeCardElementOptions } from '@stripe/stripe-js'
import { useTranslation } from 'react-i18next'
import { cashierService, type CashierSession, type CashierConfig } from '../services/cashierService'

interface StripePaymentFormProps {
  session: CashierSession
  sessionToken: string
  config: CashierConfig
  onSuccess: () => void
  onError: (error: string) => void
}

const StripePaymentForm = ({
  session,
  sessionToken,
  config,
  onSuccess,
  onError,
}: StripePaymentFormProps) => {
  const { t } = useTranslation()
  const stripe = useStripe()
  const elements = useElements()
  const [form] = Form.useForm()

  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string>('')

  const cardElementOptions: StripeCardElementOptions = {
    style: {
      base: {
        fontSize: '16px',
        color: '#424770',
        '::placeholder': {
          color: '#aab7c4',
        },
      },
      invalid: {
        color: '#9e2146',
      },
    },
    hidePostalCode: true,
  }

  const handleSubmit = async (values: any) => {
    if (!stripe || !elements) {
      return
    }

    try {
      setSubmitting(true)
      setErrorMessage('')

      const cardElement = elements.getElement(CardElement)
      if (!cardElement) {
        throw new Error('Card element not found')
      }

      // TODO: 创建支付意图 (需要与payment-gateway集成)
      // 这里应该调用payment-gateway的API创建支付
      const paymentData = {
        client_secret: 'pi_xxx_secret_xxx', // 从payment-gateway获取
        payment_no: 'PAY-' + Date.now(),
      }

      // 确认支付
      const { error, paymentIntent } = await stripe.confirmCardPayment(
        paymentData.client_secret!,
        {
          payment_method: {
            card: cardElement,
            billing_details: {
              name: values.cardholderName,
              email: values.email || session.customer_email,
            },
          },
        }
      )

      if (error) {
        throw new Error(error.message)
      }

      if (paymentIntent?.status === 'succeeded') {
        // 完成会话
        await cashierService.completeSession(sessionToken, paymentData.payment_no)

        message.success(t('cashierCheckout.payment_success') || '支付成功')
        onSuccess()

        // 重定向到成功页面
        if (config.success_redirect_url) {
          setTimeout(() => {
            window.location.href =
              config.success_redirect_url + '?payment_no=' + paymentData.payment_no
          }, 2000)
        }
      } else {
        throw new Error('Payment not completed')
      }
    } catch (error: any) {
      console.error('Stripe payment error:', error)
      const errorMsg = error.message || t('errors.payment_error') || '支付失败'
      setErrorMessage(errorMsg)
      onError(errorMsg)
      message.error(errorMsg)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form form={form} layout="vertical" onFinish={handleSubmit} autoComplete="off">
      <Form.Item
        name="cardholderName"
        label={t('cashierCheckout.cardholder_name') || '持卡人姓名'}
        rules={[{ required: true, message: t('errors.required_field') || '必填' }]}
      >
        <Input placeholder="John Doe" size="large" />
      </Form.Item>

      <Form.Item label={t('cashierCheckout.card_payment') || '卡片信息'} required>
        <div
          style={{
            padding: '12px',
            border: '1px solid #d9d9d9',
            borderRadius: '6px',
            backgroundColor: '#fff',
          }}
        >
          <CardElement options={cardElementOptions} />
        </div>
      </Form.Item>

      <Form.Item name="email" label={t('cashierCheckout.email') || '邮箱'}>
        <Input
          type="email"
          placeholder="john@example.com"
          size="large"
          defaultValue={session.customer_email}
        />
      </Form.Item>

      <Form.Item name="saveCard" valuePropName="checked">
        <Checkbox>{t('cashierCheckout.save_card') || '保存卡片信息'}</Checkbox>
      </Form.Item>

      {errorMessage && (
        <Alert
          message={errorMessage}
          type="error"
          closable
          onClose={() => setErrorMessage('')}
          style={{ marginBottom: 16 }}
        />
      )}

      <Button
        type="primary"
        htmlType="submit"
        size="large"
        block
        loading={submitting}
        disabled={!stripe}
        style={{
          backgroundColor: config.theme_color || '#1890ff',
          borderColor: config.theme_color || '#1890ff',
        }}
      >
        {submitting ? t('common.processing') || '处理中...' : t('common.pay_now') || '立即支付'}
      </Button>
    </Form>
  )
}

export default StripePaymentForm

import { useState } from 'react'
import { Button, Form, Input, Checkbox, Alert, message } from 'antd'
import { CardElement, useStripe, useElements } from '@stripe/react-stripe-js'
import type { StripeCardElementOptions } from '@stripe/stripe-js'
import { useTranslation } from 'react-i18next'
import { cashierApi } from '../services/cashierApi'
import type { CashierSession, CashierConfig } from '../types'

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

      // 创建支付意图
      const paymentData = await cashierApi.createPayment({
        session_token: sessionToken,
        channel: 'stripe',
        payment_method: 'card',
      })

      // 记录支付提交
      await cashierApi.recordLog({
        session_token: sessionToken,
        payment_submitted: true,
        selected_channel: 'stripe',
      })

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
        message.success(t('cashier.payment_success'))
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
      const errorMsg = error.message || t('errors.payment_error')
      setErrorMessage(errorMsg)
      onError(errorMsg)

      // 记录错误
      await cashierApi.recordLog({
        session_token: sessionToken,
        error_message: error.message,
      })

      message.error(errorMsg)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form form={form} layout="vertical" onFinish={handleSubmit} autoComplete="off">
      <Form.Item
        name="cardholderName"
        label={t('cashier.cardholder_name')}
        rules={[{ required: true, message: t('errors.required_field') }]}
      >
        <Input placeholder="John Doe" size="large" />
      </Form.Item>

      <Form.Item label={t('cashier.card_payment')} required>
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

      <Form.Item name="email" label={t('cashier.email')}>
        <Input
          type="email"
          placeholder="john@example.com"
          size="large"
          defaultValue={session.customer_email}
        />
      </Form.Item>

      <Form.Item name="saveCard" valuePropName="checked">
        <Checkbox>{t('cashier.save_card')}</Checkbox>
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
        {submitting ? t('common.processing') : t('common.pay_now')}
      </Button>
    </Form>
  )
}

export default StripePaymentForm

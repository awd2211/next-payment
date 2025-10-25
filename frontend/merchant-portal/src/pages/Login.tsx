import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, App, Checkbox } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { merchantService } from '../services/merchantService'
import { useAuthStore } from '../stores/authStore'

const { Title, Text } = Typography

const Login = () => {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { t } = useTranslation()
  const { setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true)
    try {
      const response = await merchantService.login(values)
      console.log('Login response:', response)

      // 响应拦截器已解包，直接获取 LoginResponse 数据
      if (response && response.token) {
        const { token, merchant } = response

        console.log('Saving auth:', { token: token ? 'exists' : 'null', merchant })

        // 保存登录信息到 store
        setAuth(token, '', merchant)

        console.log('Auth saved, checking store...')

        message.success(t('login.loginSuccess'))

        // 延迟跳转，确保 message 显示
        setTimeout(() => {
          navigate('/dashboard')
        }, 500)
      } else {
        message.error(t('login.loginFailed'))
      }
    } catch (error: any) {
      console.error('登录失败:', error)
      message.error(error?.response?.data?.message || error?.message || t('login.loginFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card
        style={{
          width: 420,
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.15)',
          borderRadius: 12,
          overflow: 'hidden',
          border: 'none',
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: 40 }}>
          <Title level={2} style={{ marginBottom: 8, fontWeight: 600 }}>
            {t('login.title')}
          </Title>
          <Text type="secondary" style={{ fontSize: 14 }}>
            {t('layout.logo')}
          </Text>
        </div>

        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
          initialValues={{ remember: true }}
        >
          <Form.Item
            name="email"
            rules={[
              { required: true, message: t('login.usernameRequired') },
              { type: 'email', message: t('login.usernameRequired') }
            ]}
          >
            <Input
              prefix={<UserOutlined style={{ color: 'rgba(0,0,0,.25)' }} />}
              placeholder={t('login.email')}
              autoComplete="email"
              style={{ borderRadius: 8 }}
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: t('login.passwordRequired') }]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: 'rgba(0,0,0,.25)' }} />}
              placeholder={t('login.password')}
              autoComplete="current-password"
              style={{ borderRadius: 8 }}
            />
          </Form.Item>

          <Form.Item name="remember" valuePropName="checked" style={{ marginBottom: 16 }}>
            <Checkbox>{t('login.rememberMe')}</Checkbox>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              type="primary"
              htmlType="submit"
              block
              loading={loading}
              size="large"
              style={{
                borderRadius: 8,
                height: 44,
                fontSize: 16,
                fontWeight: 500,
              }}
            >
              {t('login.login')}
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center', marginTop: 24, paddingTop: 24, borderTop: '1px solid #f0f0f0' }}>
          <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 8 }}>
            测试账户
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            test@example.com / Password123
          </Text>
        </div>
      </Card>
    </div>
  )
}

export default Login

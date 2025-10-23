import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, App } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { merchantService } from '../services/merchantService'
import { useAuthStore } from '../stores/authStore'

const { Title, Text } = Typography

const Login = () => {
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true)
    try {
      const response = await merchantService.login(values)
      console.log('Login response:', response)

      // 后端返回的数据结构是 {data: {data: {token, merchant}}}
      if (response.data?.data) {
        const { token, merchant } = response.data.data

        // 保存登录信息到 store
        setAuth(token, '', merchant)

        message.success('登录成功')

        // 延迟跳转，确保 message 显示
        setTimeout(() => {
          navigate('/dashboard')
        }, 500)
      } else {
        message.error('登录响应数据格式错误')
      }
    } catch (error: any) {
      console.error('登录失败:', error)
      message.error(error?.response?.data?.message || error?.message || '登录失败，请检查邮箱和密码')
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
          width: 400,
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)',
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title level={2} style={{ marginBottom: 8 }}>
            商户中心
          </Title>
          <Text type="secondary">请使用商户账号登录</Text>
        </div>

        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="邮箱地址"
              autoComplete="email"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="密码"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              block
              loading={loading}
            >
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default Login

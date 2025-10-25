import { useState, useEffect } from 'react'
import { Card, Form, Input, Button, Switch, message, Divider, QRCode, Space, List, Tag, Modal } from 'antd'
import { SafetyOutlined, LockOutlined, MobileOutlined, GlobalOutlined } from '@ant-design/icons'

export default function SecuritySettings() {
  const [passwordForm] = Form.useForm()
  const [ipForm] = Form.useForm()
  const [twoFAEnabled, setTwoFAEnabled] = useState(false)
  const [qrCodeVisible, setQrCodeVisible] = useState(false)
  const [qrCodeUrl, setQrCodeUrl] = useState('')
  const [sessions, setSessions] = useState<any[]>([])

  useEffect(() => {
    fetchSecuritySettings()
    fetchSessions()
  }, [])

  const fetchSecuritySettings = async () => {
    // TODO: 调用 apiKeyService.getSecuritySettings()
    setTwoFAEnabled(false)
  }

  const fetchSessions = async () => {
    // TODO: 调用 apiKeyService.getSessions()
    setSessions([
      {
        id: '1',
        ip_address: '192.168.1.100',
        location: '北京市',
        device: 'Chrome on Windows',
        last_active: '2025-10-25 10:30:00',
        is_current: true,
      },
    ])
  }

  const handleChangePassword = async (values: any) => {
    try {
      // TODO: 调用 apiKeyService.changePassword(values)
      message.success('密码修改成功')
      passwordForm.resetFields()
    } catch (error) {
      message.error('密码修改失败')
    }
  }

  const handleEnable2FA = async () => {
    try {
      // TODO: 调用 apiKeyService.enable2FA()
      setQrCodeUrl('otpauth://totp/PaymentPlatform:merchant@example.com?secret=JBSWY3DPEHPK3PXP&issuer=PaymentPlatform')
      setQrCodeVisible(true)
    } catch (error) {
      message.error('启用失败')
    }
  }

  const handleDisable2FA = async () => {
    Modal.confirm({
      title: '确认禁用双因素认证?',
      content: '禁用后账户安全性会降低',
      onOk: async () => {
        try {
          // TODO: 调用 apiKeyService.disable2FA()
          setTwoFAEnabled(false)
          message.success('已禁用双因素认证')
        } catch (error) {
          message.error('操作失败')
        }
      },
    })
  }

  const handleAddIP = async (values: any) => {
    try {
      // TODO: 调用 apiKeyService.addIPWhitelist(values.ip)
      message.success('IP白名单添加成功')
      ipForm.resetFields()
    } catch (error) {
      message.error('添加失败')
    }
  }

  const handleRevokeSession = async (sessionId: string) => {
    try {
      // TODO: 调用 apiKeyService.revokeSession(sessionId)
      message.success('会话已撤销')
      fetchSessions()
    } catch (error) {
      message.error('操作失败')
    }
  }

  return (
    <div>
      {/* 修改密码 */}
      <Card title={<><LockOutlined /> 修改密码</>} style={{ marginBottom: 16 }}>
        <Form form={passwordForm} onFinish={handleChangePassword} layout="vertical" style={{ maxWidth: 500 }}>
          <Form.Item
            name="old_password"
            label="当前密码"
            rules={[{ required: true, message: '请输入当前密码' }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, message: '密码至少8个字符' },
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            label="确认新密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">
              修改密码
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {/* 双因素认证 */}
      <Card title={<><MobileOutlined /> 双因素认证 (2FA)</>} style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <strong>状态:</strong> <Tag color={twoFAEnabled ? 'green' : 'orange'}>{twoFAEnabled ? '已启用' : '未启用'}</Tag>
          </div>
          <div style={{ color: '#666' }}>
            启用双因素认证后,登录时除了密码外还需要提供动态验证码,大大提高账户安全性。
          </div>
          <div>
            {twoFAEnabled ? (
              <Button danger onClick={handleDisable2FA}>禁用双因素认证</Button>
            ) : (
              <Button type="primary" onClick={handleEnable2FA}>启用双因素认证</Button>
            )}
          </div>
        </Space>
      </Card>

      {/* IP白名单 */}
      <Card title={<><GlobalOutlined /> IP白名单</>} style={{ marginBottom: 16 }}>
        <div style={{ marginBottom: 16, color: '#666' }}>
          配置IP白名单后,只有来自白名单IP的API请求才会被接受,提高API安全性。
        </div>
        <Form form={ipForm} onFinish={handleAddIP} layout="inline">
          <Form.Item
            name="ip"
            rules={[{ required: true, message: '请输入IP地址' }]}
          >
            <Input placeholder="例如: 192.168.1.100" style={{ width: 200 }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">添加</Button>
          </Form.Item>
        </Form>
        <Divider />
        <List
          dataSource={['192.168.1.100', '10.0.0.1']}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button type="link" danger onClick={() => message.success('删除成功')}>删除</Button>,
              ]}
            >
              {item}
            </List.Item>
          )}
        />
      </Card>

      {/* 活跃会话 */}
      <Card title={<><SafetyOutlined /> 活跃会话</>}>
        <List
          dataSource={sessions}
          renderItem={(session) => (
            <List.Item
              actions={[
                !session.is_current && (
                  <Button
                    type="link"
                    danger
                    onClick={() => handleRevokeSession(session.id)}
                  >
                    撤销
                  </Button>
                ),
              ]}
            >
              <List.Item.Meta
                title={
                  <Space>
                    {session.device}
                    {session.is_current && <Tag color="blue">当前</Tag>}
                  </Space>
                }
                description={
                  <div>
                    <div>IP: {session.ip_address}</div>
                    <div>位置: {session.location}</div>
                    <div>最后活跃: {session.last_active}</div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Card>

      {/* 2FA QR Code Modal */}
      <Modal
        title="扫描二维码启用双因素认证"
        open={qrCodeVisible}
        onCancel={() => setQrCodeVisible(false)}
        footer={null}
      >
        <div style={{ textAlign: 'center' }}>
          <QRCode value={qrCodeUrl} size={200} />
          <div style={{ marginTop: 16, color: '#666' }}>
            请使用Google Authenticator或其他TOTP应用扫描二维码
          </div>
        </div>
      </Modal>
    </div>
  )
}

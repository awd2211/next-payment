import { useState, useEffect } from 'react'
import {
  Typography,
  Card,
  Tabs,
  Form,
  Input,
  Button,
  Switch,
  Select,
  Table,
  Space,
  Descriptions,
  Modal,
  message,
  Alert,
  Tag,
  QRCode,
  Divider,
  Progress,
  Tooltip,
  Row,
  Col,
} from 'antd'
import {
  UserOutlined,
  LockOutlined,
  SafetyOutlined,
  HistoryOutlined,
  SettingOutlined,
  KeyOutlined,
  GlobalOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useTranslation } from 'react-i18next'
import { validatePasswordStrength } from '../utils/security'
import { merchantService } from '../services/merchantService'
import dayjs from 'dayjs'

const { Title, Text, Paragraph } = Typography
const { Option } = Select

// 活动日志类型
interface ActivityLog {
  id: string
  action: string
  ip_address: string
  user_agent: string
  created_at: string
  location?: string
  status: 'success' | 'failed'
}

// 商户偏好设置
interface MerchantPreferences {
  language: string
  timezone: string
  currency: string
  date_format: string
  time_format: string
  notifications_email: boolean
  notifications_sms: boolean
  notifications_push: boolean
}

const Account = () => {
  const { t, i18n } = useTranslation()
  const [activeTab, setActiveTab] = useState('profile')
  const [loading, setLoading] = useState(false)

  // 密码修改
  const [passwordForm] = Form.useForm()
  const [passwordStrength, setPasswordStrength] = useState<'weak' | 'medium' | 'strong' | ''>('')

  // 2FA设置
  const [twoFactorEnabled, setTwoFactorEnabled] = useState(false)
  const [showQRCode, setShowQRCode] = useState(false)
  const [qrCodeUrl, setQrCodeUrl] = useState('')
  const [verifyCode, setVerifyCode] = useState('')

  // 活动记录
  const [activityLogs, setActivityLogs] = useState<ActivityLog[]>([])
  const [activityLoading, setActivityLoading] = useState(false)

  // 偏好设置
  const [preferencesForm] = Form.useForm()
  const [preferences, setPreferences] = useState<MerchantPreferences>({
    language: 'zh-CN',
    timezone: 'Asia/Shanghai',
    currency: 'USD',
    date_format: 'YYYY-MM-DD',
    time_format: '24h',
    notifications_email: true,
    notifications_sms: false,
    notifications_push: true,
  })

  useEffect(() => {
    loadActivityLogs()
    loadPreferences()
    load2FAStatus()
  }, [])

  // 加载活动记录
  const loadActivityLogs = async () => {
    setActivityLoading(true)
    try {
      // TODO: 调用实际API
      // const response = await merchantService.getActivityLogs()
      // 模拟数据
      const mockLogs: ActivityLog[] = [
        {
          id: '1',
          action: '登录',
          ip_address: '192.168.1.1',
          user_agent: 'Chrome 120.0',
          created_at: new Date().toISOString(),
          location: '中国 上海',
          status: 'success',
        },
        {
          id: '2',
          action: '修改密码',
          ip_address: '192.168.1.1',
          user_agent: 'Chrome 120.0',
          created_at: dayjs().subtract(1, 'day').toISOString(),
          location: '中国 上海',
          status: 'success',
        },
        {
          id: '3',
          action: '登录失败',
          ip_address: '203.0.113.0',
          user_agent: 'Firefox 121.0',
          created_at: dayjs().subtract(2, 'day').toISOString(),
          location: '美国 纽约',
          status: 'failed',
        },
      ]
      setActivityLogs(mockLogs)
    } catch (error) {
      console.error('加载活动记录失败:', error)
    } finally {
      setActivityLoading(false)
    }
  }

  // 加载偏好设置
  const loadPreferences = async () => {
    try {
      // TODO: 调用实际API
      // const response = await merchantService.getPreferences()
      // const prefs = response.data
      // setPreferences(prefs)
      // preferencesForm.setFieldsValue(prefs)
      preferencesForm.setFieldsValue(preferences)
    } catch (error) {
      console.error('加载偏好设置失败:', error)
    }
  }

  // 加载2FA状态
  const load2FAStatus = async () => {
    try {
      // TODO: 调用实际API
      // const response = await merchantService.get2FAStatus()
      // setTwoFactorEnabled(response.data.enabled)
      setTwoFactorEnabled(false)
    } catch (error) {
      console.error('加载2FA状态失败:', error)
    }
  }

  // 修改密码
  const handleChangePassword = async (values: any) => {
    setLoading(true)
    try {
      // TODO: 调用实际API
      // await merchantService.changePassword({
      //   old_password: values.oldPassword,
      //   new_password: values.newPassword,
      // })

      message.success(t('account.passwordChangeSuccess'))
      passwordForm.resetFields()
      setPasswordStrength('')
    } catch (error: any) {
      message.error(error?.message || t('account.passwordChangeFailed'))
    } finally {
      setLoading(false)
    }
  }

  // 检查密码强度
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const password = e.target.value
    if (password) {
      const result = validatePasswordStrength(password)
      setPasswordStrength(result.strength)
    } else {
      setPasswordStrength('')
    }
  }

  // 启用2FA
  const handleEnable2FA = async () => {
    setLoading(true)
    try {
      // TODO: 调用实际API获取二维码
      // const response = await merchantService.enable2FA()
      // setQrCodeUrl(response.data.qr_code_url)

      // 模拟二维码URL
      setQrCodeUrl('otpauth://totp/MerchantPortal:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=MerchantPortal')
      setShowQRCode(true)
    } catch (error: any) {
      message.error(error?.message || t('account.enable2FAFailed'))
    } finally {
      setLoading(false)
    }
  }

  // 验证2FA代码
  const handleVerify2FA = async () => {
    if (!verifyCode || verifyCode.length !== 6) {
      message.error(t('account.invalidVerifyCode'))
      return
    }

    setLoading(true)
    try {
      // TODO: 调用实际API验证
      // await merchantService.verify2FA({ code: verifyCode })

      setTwoFactorEnabled(true)
      setShowQRCode(false)
      setVerifyCode('')
      message.success(t('account.2FAEnabledSuccess'))
    } catch (error: any) {
      message.error(error?.message || t('account.verify2FAFailed'))
    } finally {
      setLoading(false)
    }
  }

  // 禁用2FA
  const handleDisable2FA = async () => {
    Modal.confirm({
      title: t('account.disable2FAConfirm'),
      content: t('account.disable2FAWarning'),
      onOk: async () => {
        setLoading(true)
        try {
          // TODO: 调用实际API
          // await merchantService.disable2FA()

          setTwoFactorEnabled(false)
          message.success(t('account.2FADisabledSuccess'))
        } catch (error: any) {
          message.error(error?.message || t('account.disable2FAFailed'))
        } finally {
          setLoading(false)
        }
      },
    })
  }

  // 保存偏好设置
  const handleSavePreferences = async (values: MerchantPreferences) => {
    setLoading(true)
    try {
      // TODO: 调用实际API
      // await merchantService.updatePreferences(values)

      setPreferences(values)

      // 更新语言
      if (values.language !== i18n.language) {
        i18n.changeLanguage(values.language)
      }

      message.success(t('account.preferencesSaved'))
    } catch (error: any) {
      message.error(error?.message || t('account.preferencesSaveFailed'))
    } finally {
      setLoading(false)
    }
  }

  // 活动记录表格列
  const activityColumns: ColumnsType<ActivityLog> = [
    {
      title: t('account.action'),
      dataIndex: 'action',
      key: 'action',
      render: (text: string, record: ActivityLog) => (
        <Space>
          <Text>{text}</Text>
          {record.status === 'failed' && <Tag color="red">{t('account.failed')}</Tag>}
        </Space>
      ),
    },
    {
      title: t('account.ipAddress'),
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: t('account.location'),
      dataIndex: 'location',
      key: 'location',
      render: (text?: string) => text || '-',
    },
    {
      title: t('account.device'),
      dataIndex: 'user_agent',
      key: 'user_agent',
    },
    {
      title: t('account.time'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  // 密码强度颜色和百分比
  const getPasswordStrengthInfo = () => {
    switch (passwordStrength) {
      case 'weak':
        return { color: '#ff4d4f', percent: 33, text: '弱' }
      case 'medium':
        return { color: '#faad14', percent: 66, text: '中等' }
      case 'strong':
        return { color: '#52c41a', percent: 100, text: '强' }
      default:
        return { color: '#d9d9d9', percent: 0, text: '' }
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>{t('account.title')}</Title>
        <Paragraph type="secondary" style={{ marginBottom: 0 }}>{t('account.subtitle')}</Paragraph>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        {/* 安全设置 */}
        <Tabs.TabPane
          tab={
            <span>
              <LockOutlined />
              {t('account.securitySettings')}
            </span>
          }
          key="security"
        >
          <Card title={t('account.changePassword')} style={{ marginBottom: 24, borderRadius: 12 }}>
            <Form
              form={passwordForm}
              layout="vertical"
              onFinish={handleChangePassword}
              style={{ maxWidth: 600 }}
            >
              <Form.Item
                label={t('account.oldPassword')}
                name="oldPassword"
                rules={[{ required: true, message: t('account.oldPasswordRequired') }]}
              >
                <Input.Password prefix={<LockOutlined />} style={{ borderRadius: 8 }} />
              </Form.Item>

              <Form.Item
                label={t('account.newPassword')}
                name="newPassword"
                rules={[
                  { required: true, message: t('account.newPasswordRequired') },
                  { min: 8, message: t('account.passwordMinLength') },
                  {
                    validator: (_, value) => {
                      if (!value) return Promise.resolve()
                      const result = validatePasswordStrength(value)
                      if (result.strength === 'weak') {
                        return Promise.reject(t('account.passwordTooWeak'))
                      }
                      return Promise.resolve()
                    },
                  },
                ]}
              >
                <Input.Password
                  prefix={<KeyOutlined />}
                  onChange={handlePasswordChange}
                  style={{ borderRadius: 8 }}
                />
              </Form.Item>

              {passwordStrength && (
                <div style={{ marginBottom: 16 }}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Text type="secondary">密码强度:</Text>
                      <Text strong style={{ color: getPasswordStrengthInfo().color }}>
                        {getPasswordStrengthInfo().text}
                      </Text>
                    </div>
                    <Progress
                      percent={getPasswordStrengthInfo().percent}
                      strokeColor={getPasswordStrengthInfo().color}
                      showInfo={false}
                      strokeWidth={8}
                    />
                  </Space>
                </div>
              )}

              <Form.Item
                label={t('account.confirmPassword')}
                name="confirmPassword"
                dependencies={['newPassword']}
                rules={[
                  { required: true, message: t('account.confirmPasswordRequired') },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('newPassword') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error(t('account.passwordMismatch')))
                    },
                  }),
                ]}
              >
                <Input.Password prefix={<KeyOutlined />} style={{ borderRadius: 8 }} />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading} style={{ borderRadius: 8 }}>
                  {t('account.changePasswordButton')}
                </Button>
              </Form.Item>
            </Form>
          </Card>

          {/* 2FA设置 */}
          <Card title={t('account.twoFactorAuth')} style={{ borderRadius: 12 }}>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <Alert
                message={t('account.2FADescription')}
                description={t('account.2FADescriptionDetail')}
                type="info"
                showIcon
                icon={<SafetyOutlined />}
                style={{ borderRadius: 8 }}
              />

              <Row gutter={16}>
                <Col span={12}>
                  <Card style={{ borderRadius: 8, background: twoFactorEnabled ? '#f6ffed' : '#fff2f0', border: `1px solid ${twoFactorEnabled ? '#b7eb8f' : '#ffccc7'}` }}>
                    <Space>
                      {twoFactorEnabled ? (
                        <CheckCircleOutlined style={{ fontSize: 24, color: '#52c41a' }} />
                      ) : (
                        <WarningOutlined style={{ fontSize: 24, color: '#ff4d4f' }} />
                      )}
                      <div>
                        <Text type="secondary">状态</Text>
                        <div>
                          <Text strong style={{ fontSize: 16, color: twoFactorEnabled ? '#52c41a' : '#ff4d4f' }}>
                            {twoFactorEnabled ? '已启用' : '未启用'}
                          </Text>
                        </div>
                      </div>
                    </Space>
                  </Card>
                </Col>
                <Col span={12}>
                  {!twoFactorEnabled ? (
                    <Button
                      type="primary"
                      icon={<SafetyOutlined />}
                      onClick={handleEnable2FA}
                      loading={loading}
                      block
                      size="large"
                      style={{ borderRadius: 8, height: '100%' }}
                    >
                      {t('account.enable2FA')}
                    </Button>
                  ) : (
                    <Button
                      danger
                      onClick={handleDisable2FA}
                      loading={loading}
                      block
                      size="large"
                      style={{ borderRadius: 8, height: '100%' }}
                    >
                      {t('account.disable2FA')}
                    </Button>
                  )}
                </Col>
              </Row>

              {/* 2FA设置Modal */}
              <Modal
                title={t('account.setup2FA')}
                open={showQRCode}
                onCancel={() => {
                  setShowQRCode(false)
                  setVerifyCode('')
                }}
                footer={null}
                width={500}
              >
                <Space direction="vertical" style={{ width: '100%' }} size="large">
                  <Alert
                    message={t('account.scanQRCode')}
                    description={t('account.scanQRCodeDescription')}
                    type="info"
                  />

                  <div style={{ textAlign: 'center' }}>
                    <QRCode value={qrCodeUrl} size={200} />
                  </div>

                  <Divider>{t('account.orEnterManually')}</Divider>

                  <div>
                    <Text copyable>{qrCodeUrl.split('secret=')[1]?.split('&')[0]}</Text>
                  </div>

                  <div>
                    <Text strong>{t('account.enterVerifyCode')}:</Text>
                    <Input
                      placeholder={t('account.verifyCodePlaceholder')}
                      value={verifyCode}
                      onChange={(e) => setVerifyCode(e.target.value)}
                      maxLength={6}
                      style={{ marginTop: 8 }}
                    />
                  </div>

                  <Button
                    type="primary"
                    block
                    onClick={handleVerify2FA}
                    loading={loading}
                    disabled={verifyCode.length !== 6}
                  >
                    {t('account.verify')}
                  </Button>
                </Space>
              </Modal>
            </Space>
          </Card>
        </Tabs.TabPane>

        {/* 活动记录 */}
        <Tabs.TabPane
          tab={
            <span>
              <HistoryOutlined />
              {t('account.activityLog')}
            </span>
          }
          key="activity"
        >
          <Card style={{ borderRadius: 12 }}>
            <Table
              columns={activityColumns}
              dataSource={activityLogs}
              loading={activityLoading}
              rowKey="id"
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => t('common.total', { count: total }),
              }}
            />
          </Card>
        </Tabs.TabPane>

        {/* 偏好设置 */}
        <Tabs.TabPane
          tab={
            <span>
              <SettingOutlined />
              {t('account.preferences')}
            </span>
          }
          key="preferences"
        >
          <Card style={{ borderRadius: 12 }}>
            <Form
              form={preferencesForm}
              layout="vertical"
              onFinish={handleSavePreferences}
              initialValues={preferences}
              style={{ maxWidth: 600 }}
            >
              <Title level={4}>
                <GlobalOutlined /> {t('account.regionSettings')}
              </Title>

              <Form.Item
                label={t('account.language')}
                name="language"
                tooltip={t('account.languageTooltip')}
              >
                <Select style={{ borderRadius: 8 }}>
                  <Option value="zh-CN">简体中文</Option>
                  <Option value="en-US">English</Option>
                  <Option value="zh-TW">繁體中文</Option>
                  <Option value="ja">日本語</Option>
                  <Option value="ko">한국어</Option>
                </Select>
              </Form.Item>

              <Form.Item
                label={t('account.timezone')}
                name="timezone"
                tooltip={t('account.timezoneTooltip')}
              >
                <Select showSearch style={{ borderRadius: 8 }}>
                  <Option value="Asia/Shanghai">Asia/Shanghai (UTC+8)</Option>
                  <Option value="Asia/Tokyo">Asia/Tokyo (UTC+9)</Option>
                  <Option value="Asia/Seoul">Asia/Seoul (UTC+9)</Option>
                  <Option value="America/New_York">America/New_York (UTC-5)</Option>
                  <Option value="America/Los_Angeles">America/Los_Angeles (UTC-8)</Option>
                  <Option value="Europe/London">Europe/London (UTC+0)</Option>
                  <Option value="UTC">UTC (UTC+0)</Option>
                </Select>
              </Form.Item>

              <Form.Item
                label={t('account.defaultCurrency')}
                name="currency"
                tooltip={t('account.currencyTooltip')}
              >
                <Select style={{ borderRadius: 8 }}>
                  <Option value="USD">USD - 美元</Option>
                  <Option value="CNY">CNY - 人民币</Option>
                  <Option value="EUR">EUR - 欧元</Option>
                  <Option value="GBP">GBP - 英镑</Option>
                  <Option value="JPY">JPY - 日元</Option>
                  <Option value="KRW">KRW - 韩元</Option>
                  <Option value="HKD">HKD - 港币</Option>
                </Select>
              </Form.Item>

              <Divider />

              <Title level={4}>
                <ClockCircleOutlined /> {t('account.formatSettings')}
              </Title>

              <Form.Item label={t('account.dateFormat')} name="date_format">
                <Select style={{ borderRadius: 8 }}>
                  <Option value="YYYY-MM-DD">YYYY-MM-DD (2024-01-15)</Option>
                  <Option value="MM/DD/YYYY">MM/DD/YYYY (01/15/2024)</Option>
                  <Option value="DD/MM/YYYY">DD/MM/YYYY (15/01/2024)</Option>
                  <Option value="YYYY年MM月DD日">YYYY年MM月DD日 (2024年01月15日)</Option>
                </Select>
              </Form.Item>

              <Form.Item label={t('account.timeFormat')} name="time_format">
                <Select style={{ borderRadius: 8 }}>
                  <Option value="24h">24小时制 (14:30)</Option>
                  <Option value="12h">12小时制 (2:30 PM)</Option>
                </Select>
              </Form.Item>

              <Divider />

              <Title level={4}>
                <SafetyOutlined /> {t('account.notificationSettings')}
              </Title>

              <Row gutter={[16, 16]}>
                <Col xs={24} md={8}>
                  <Card style={{ borderRadius: 8, height: '100%' }}>
                    <Space direction="vertical" size="small">
                      <Form.Item
                        label={t('account.emailNotifications')}
                        name="notifications_email"
                        valuePropName="checked"
                        style={{ marginBottom: 0 }}
                      >
                        <Switch />
                      </Form.Item>
                      <Text type="secondary" style={{ fontSize: 12 }}>接收重要事件的邮件通知</Text>
                    </Space>
                  </Card>
                </Col>
                <Col xs={24} md={8}>
                  <Card style={{ borderRadius: 8, height: '100%' }}>
                    <Space direction="vertical" size="small">
                      <Form.Item
                        label={t('account.smsNotifications')}
                        name="notifications_sms"
                        valuePropName="checked"
                        style={{ marginBottom: 0 }}
                      >
                        <Switch />
                      </Form.Item>
                      <Text type="secondary" style={{ fontSize: 12 }}>接收紧急事件的短信通知</Text>
                    </Space>
                  </Card>
                </Col>
                <Col xs={24} md={8}>
                  <Card style={{ borderRadius: 8, height: '100%' }}>
                    <Space direction="vertical" size="small">
                      <Form.Item
                        label={t('account.pushNotifications')}
                        name="notifications_push"
                        valuePropName="checked"
                        style={{ marginBottom: 0 }}
                      >
                        <Switch />
                      </Form.Item>
                      <Text type="secondary" style={{ fontSize: 12 }}>接收浏览器推送通知</Text>
                    </Space>
                  </Card>
                </Col>
              </Row>

              <Divider />

              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit" loading={loading} style={{ borderRadius: 8 }}>
                    {t('account.savePreferences')}
                  </Button>
                  <Button onClick={() => preferencesForm.resetFields()} style={{ borderRadius: 8 }}>
                    {t('common.reset')}
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </Tabs.TabPane>
      </Tabs>
    </div>
  )
}

export default Account

import { useState, useEffect } from 'react'
import {
  Card,
  Tabs,
  Form,
  Input,
  Button,
  Switch,
  Select,
  Upload,
  ColorPicker,
  InputNumber,
  message,
  Space,
  Divider,
  Row,
  Col,
  Statistic,
  Tag,
  Table,
  QRCode,
  Modal,
} from 'antd'
import {
  UploadOutlined,
  CopyOutlined,
  EyeOutlined,
  LinkOutlined,
  BarChartOutlined,
  SettingOutlined,
  BgColorsOutlined,
  GlobalOutlined,
  SafetyOutlined,
  ToolOutlined,
} from '@ant-design/icons'
import type { UploadFile } from 'antd/es/upload/interface'
import { useTranslation } from 'react-i18next'
import { Pie, Funnel } from '@ant-design/charts'
import { cashierService, CashierConfig } from '../services/cashierService'
import dayjs from 'dayjs'

const { TextArea } = Input
const { TabPane } = Tabs

const CashierConfigPage = () => {
  const { t } = useTranslation()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [_config, setConfig] = useState<CashierConfig | null>(null)
  const [analytics, setAnalytics] = useState<any>(null)
  const [paymentLinkModalVisible, setPaymentLinkModalVisible] = useState(false)
  const [generatedLink, setGeneratedLink] = useState('')
  const [generatedToken, setGeneratedToken] = useState('')

  useEffect(() => {
    loadConfig()
    loadAnalytics()
  }, [])

  const loadConfig = async () => {
    setLoading(true)
    try {
      const response = await cashierService.getConfig()
      const configData = response.data
      setConfig(configData)

      // 填充表单
      form.setFieldsValue({
        theme_color: configData.theme_color || '#1890ff',
        logo_url: configData.logo_url,
        background_image_url: configData.background_image_url,
        custom_css: configData.custom_css,
        enabled_channels: configData.enabled_channels || [],
        default_channel: configData.default_channel,
        enabled_languages: configData.enabled_languages || ['en'],
        default_language: configData.default_language || 'en',
        auto_submit: configData.auto_submit,
        show_amount_breakdown: configData.show_amount_breakdown !== false,
        allow_channel_switch: configData.allow_channel_switch !== false,
        session_timeout_minutes: configData.session_timeout_minutes || 30,
        require_cvv: configData.require_cvv !== false,
        enable_3d_secure: configData.enable_3d_secure !== false,
        allowed_countries: configData.allowed_countries || [],
        success_redirect_url: configData.success_redirect_url,
        cancel_redirect_url: configData.cancel_redirect_url,
      })
    } catch (error: any) {
      if (error.response?.status !== 404) {
        message.error(t('cashier.loadConfigError') || '加载配置失败')
      }
    } finally {
      setLoading(false)
    }
  }

  const loadAnalytics = async () => {
    try {
      const startTime = dayjs().subtract(7, 'days').toISOString()
      const endTime = dayjs().toISOString()
      const response = await cashierService.getAnalytics(startTime, endTime)
      setAnalytics(response.data)
    } catch (error) {
      console.error('Failed to load analytics:', error)
    }
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      await cashierService.createOrUpdateConfig(values)
      message.success(t('cashier.saveSuccess') || '保存成功')
      loadConfig()
    } catch (error: any) {
      if (error.errorFields) {
        message.error(t('common.validationError') || '请检查表单填写')
      } else {
        message.error(t('cashier.saveError') || '保存失败')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleGeneratePaymentLink = async (values: any) => {
    try {
      const response = await cashierService.createSession({
        order_no: `ORDER-${Date.now()}`,
        amount: values.amount * 100, // 转换为分
        currency: values.currency || 'USD',
        description: values.description,
        customer_email: values.customer_email,
        expires_in_minutes: 30,
      })

      const { session_token, cashier_url } = response.data
      const fullUrl = `${window.location.origin}${cashier_url}`

      setGeneratedToken(session_token)
      setGeneratedLink(fullUrl)
      setPaymentLinkModalVisible(true)
    } catch (error) {
      message.error(t('cashier.generateLinkError') || '生成链接失败')
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    message.success(t('common.copied') || '已复制到剪贴板')
  }

  // 渠道选项
  const channelOptions = [
    { label: 'Stripe', value: 'stripe' },
    { label: 'PayPal', value: 'paypal' },
    { label: t('cashier.alipay') || '支付宝', value: 'alipay' },
    { label: t('cashier.wechat') || '微信支付', value: 'wechat' },
  ]

  // 语言选项
  const languageOptions = [
    { label: 'English', value: 'en' },
    { label: '简体中文', value: 'zh-CN' },
    { label: '繁體中文', value: 'zh-TW' },
    { label: '日本語', value: 'ja' },
    { label: '한국어', value: 'ko' },
    { label: 'Español', value: 'es' },
    { label: 'Français', value: 'fr' },
  ]

  // 渠道统计数据转换
  const channelData = analytics?.channel_stats
    ? Object.entries(analytics.channel_stats).map(([channel, count]) => ({
        channel,
        value: count as number,
      }))
    : []

  // 漏斗数据示例
  const funnelData = [
    { stage: t('cashier.pageView') || '页面访问', value: 1000 },
    { stage: t('cashier.channelSelect') || '选择渠道', value: 850 },
    { stage: t('cashier.formFill') || '填写信息', value: 720 },
    { stage: t('cashier.submit') || '提交支付', value: 680 },
    { stage: t('cashier.success') || '支付成功', value: 650 },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Card
        title={
          <Space>
            <SettingOutlined />
            {t('cashier.title') || '收银台配置'}
          </Space>
        }
        extra={
          <Button type="primary" loading={loading} onClick={handleSave} style={{ borderRadius: 8 }}>
            {t('common.save') || '保存配置'}
          </Button>
        }
        style={{ borderRadius: 12 }}
      >
        <Tabs defaultActiveKey="appearance" tabPosition="left">
          {/* 外观设置 */}
          <TabPane
            tab={
              <span>
                <BgColorsOutlined />
                {t('cashier.appearance') || '外观设置'}
              </span>
            }
            key="appearance"
          >
            <Form form={form} layout="vertical">
              <Form.Item
                label={t('cashier.logo') || 'Logo'}
                name="logo_url"
                extra={t('cashier.logoTip') || '推荐尺寸: 200x60px, PNG格式'}
              >
                <Input placeholder="https://example.com/logo.png" prefix={<LinkOutlined />} />
              </Form.Item>

              <Form.Item
                label={t('cashier.themeColor') || '主题颜色'}
                name="theme_color"
              >
                <ColorPicker showText />
              </Form.Item>

              <Form.Item
                label={t('cashier.backgroundImage') || '背景图片'}
                name="background_image_url"
              >
                <Input placeholder="https://example.com/bg.jpg" prefix={<LinkOutlined />} />
              </Form.Item>

              <Form.Item
                label={t('cashier.customCSS') || '自定义CSS'}
                name="custom_css"
                extra={t('cashier.customCSSTip') || '高级用户可自定义收银台样式'}
              >
                <TextArea rows={6} placeholder=".checkout-page { background: #f0f0f0; }" />
              </Form.Item>

              <Divider />

              <Card title={t('cashier.preview') || '预览效果'} size="small">
                <div
                  style={{
                    background: form.getFieldValue('theme_color') || '#1890ff',
                    padding: '40px',
                    textAlign: 'center',
                    color: 'white',
                    borderRadius: '8px',
                  }}
                >
                  {form.getFieldValue('logo_url') ? (
                    <img
                      src={form.getFieldValue('logo_url')}
                      alt="Logo"
                      style={{ maxHeight: '60px', marginBottom: '20px' }}
                    />
                  ) : (
                    <h2 style={{ color: 'white' }}>{t('cashier.yourLogo') || '您的Logo'}</h2>
                  )}
                  <div style={{ background: 'white', color: '#333', padding: '20px', borderRadius: '4px' }}>
                    <h3>{t('cashier.paymentAmount') || '支付金额'}: $99.99</h3>
                    <Button type="primary" size="large">
                      {t('cashier.pay') || '立即支付'}
                    </Button>
                  </div>
                </div>
              </Card>
            </Form>
          </TabPane>

          {/* 支付方式 */}
          <TabPane
            tab={
              <span>
                <GlobalOutlined />
                {t('cashier.paymentMethods') || '支付方式'}
              </span>
            }
            key="payment-methods"
          >
            <Form form={form} layout="vertical">
              <Form.Item
                label={t('cashier.enabledChannels') || '启用的支付渠道'}
                name="enabled_channels"
              >
                <Select
                  mode="multiple"
                  options={channelOptions}
                  placeholder={t('cashier.selectChannels') || '选择支付渠道'}
                />
              </Form.Item>

              <Form.Item
                label={t('cashier.defaultChannel') || '默认支付渠道'}
                name="default_channel"
              >
                <Select options={channelOptions} placeholder={t('cashier.selectDefault') || '选择默认渠道'} />
              </Form.Item>

              <Form.Item
                label={t('cashier.allowChannelSwitch') || '允许用户切换渠道'}
                name="allow_channel_switch"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Divider />

              <Form.Item
                label={t('cashier.enabledLanguages') || '支持的语言'}
                name="enabled_languages"
              >
                <Select
                  mode="multiple"
                  options={languageOptions}
                  placeholder={t('cashier.selectLanguages') || '选择支持的语言'}
                />
              </Form.Item>

              <Form.Item
                label={t('cashier.defaultLanguage') || '默认语言'}
                name="default_language"
              >
                <Select options={languageOptions} />
              </Form.Item>
            </Form>
          </TabPane>

          {/* 安全设置 */}
          <TabPane
            tab={
              <span>
                <SafetyOutlined />
                {t('cashier.security') || '安全设置'}
              </span>
            }
            key="security"
          >
            <Form form={form} layout="vertical">
              <Form.Item
                label={t('cashier.sessionTimeout') || '会话超时时间'}
                name="session_timeout_minutes"
              >
                <InputNumber min={5} max={120} addonAfter={t('common.minutes') || '分钟'} style={{ width: '200px' }} />
              </Form.Item>

              <Form.Item
                label={t('cashier.requireCVV') || '强制要求CVV验证'}
                name="require_cvv"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label={t('cashier.enable3DSecure') || '启用3D Secure验证'}
                name="enable_3d_secure"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label={t('cashier.allowedCountries') || '允许的国家/地区'}
                name="allowed_countries"
                extra={t('cashier.allowedCountriesTip') || '留空表示允许所有国家'}
              >
                <Select
                  mode="tags"
                  placeholder={t('cashier.enterCountryCode') || '输入国家代码 (如: US, CN, JP)'}
                />
              </Form.Item>

              <Divider />

              <Form.Item
                label={t('cashier.successRedirectURL') || '支付成功跳转URL'}
                name="success_redirect_url"
              >
                <Input placeholder="https://yoursite.com/success" prefix={<LinkOutlined />} />
              </Form.Item>

              <Form.Item
                label={t('cashier.cancelRedirectURL') || '支付取消跳转URL'}
                name="cancel_redirect_url"
              >
                <Input placeholder="https://yoursite.com/cancel" prefix={<LinkOutlined />} />
              </Form.Item>
            </Form>
          </TabPane>

          {/* 数据分析 */}
          <TabPane
            tab={
              <span>
                <BarChartOutlined />
                {t('cashier.analytics') || '数据分析'}
              </span>
            }
            key="analytics"
          >
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              <Col span={8}>
                <Statistic
                  title={t('cashier.conversionRate') || '转化率'}
                  value={analytics?.conversion_rate || 0}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: '#3f8600' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title={t('cashier.totalSessions') || '总会话数'}
                  value={analytics?.total_sessions || 0}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title={t('cashier.avgCompletionTime') || '平均完成时间'}
                  value={92}
                  suffix={t('common.seconds') || '秒'}
                />
              </Col>
            </Row>

            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Card title={t('cashier.channelPreference') || '渠道偏好'} size="small">
                  {channelData.length > 0 ? (
                    <Pie
                      data={channelData}
                      angleField="value"
                      colorField="channel"
                      radius={0.8}
                      label={{
                        type: 'outer',
                        content: '{name} {percentage}',
                      }}
                      legend={{ position: 'bottom' }}
                    />
                  ) : (
                    <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
                      {t('common.noData') || '暂无数据'}
                    </div>
                  )}
                </Card>
              </Col>

              <Col span={12}>
                <Card title={t('cashier.paymentFunnel') || '支付漏斗'} size="small">
                  <Funnel
                    data={funnelData}
                    xField="stage"
                    yField="value"
                    legend={false}
                  />
                </Card>
              </Col>
            </Row>
          </TabPane>

          {/* 快捷工具 */}
          <TabPane
            tab={
              <span>
                <ToolOutlined />
                {t('cashier.tools') || '快捷工具'}
              </span>
            }
            key="tools"
          >
            <Card title={t('cashier.generatePaymentLink') || '生成支付链接'} style={{ marginBottom: 16 }}>
              <Form onFinish={handleGeneratePaymentLink} layout="vertical">
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item
                      label={t('cashier.amount') || '金额'}
                      name="amount"
                      rules={[{ required: true, message: t('common.required') || '必填' }]}
                    >
                      <InputNumber
                        min={0.01}
                        precision={2}
                        prefix="$"
                        style={{ width: '100%' }}
                        placeholder="99.99"
                      />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label={t('cashier.currency') || '货币'} name="currency" initialValue="USD">
                      <Select>
                        <Select.Option value="USD">USD</Select.Option>
                        <Select.Option value="EUR">EUR</Select.Option>
                        <Select.Option value="CNY">CNY</Select.Option>
                        <Select.Option value="JPY">JPY</Select.Option>
                      </Select>
                    </Form.Item>
                  </Col>
                </Row>

                <Form.Item label={t('cashier.description') || '描述'} name="description">
                  <Input placeholder={t('cashier.productName') || '商品名称'} />
                </Form.Item>

                <Form.Item label={t('cashier.customerEmail') || '客户邮箱'} name="customer_email">
                  <Input type="email" placeholder="customer@example.com" />
                </Form.Item>

                <Form.Item>
                  <Button type="primary" htmlType="submit" icon={<LinkOutlined />}>
                    {t('cashier.generateLink') || '生成链接'}
                  </Button>
                </Form.Item>
              </Form>
            </Card>

            <Card title={t('cashier.testCashier') || '测试收银台'}>
              <p>{t('cashier.testCashierDesc') || '在测试环境中预览收银台页面效果'}</p>
              <Button
                type="primary"
                icon={<EyeOutlined />}
                onClick={() => window.open('/cashier/demo', '_blank')}
              >
                {t('cashier.openTestCashier') || '打开测试收银台'}
              </Button>
            </Card>
          </TabPane>
        </Tabs>
      </Card>

      {/* 生成的支付链接弹窗 */}
      <Modal
        title={t('cashier.paymentLinkGenerated') || '支付链接已生成'}
        open={paymentLinkModalVisible}
        onCancel={() => setPaymentLinkModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setPaymentLinkModalVisible(false)}>
            {t('common.close') || '关闭'}
          </Button>,
        ]}
        width={600}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <p style={{ marginBottom: 8, fontWeight: 'bold' }}>
              {t('cashier.paymentLink') || '支付链接'}:
            </p>
            <Input
              value={generatedLink}
              readOnly
              addonAfter={
                <CopyOutlined
                  onClick={() => copyToClipboard(generatedLink)}
                  style={{ cursor: 'pointer' }}
                />
              }
            />
          </div>

          <div>
            <p style={{ marginBottom: 8, fontWeight: 'bold' }}>
              {t('cashier.sessionToken') || '会话Token'}:
            </p>
            <Input
              value={generatedToken}
              readOnly
              addonAfter={
                <CopyOutlined
                  onClick={() => copyToClipboard(generatedToken)}
                  style={{ cursor: 'pointer' }}
                />
              }
            />
          </div>

          <div style={{ textAlign: 'center' }}>
            <p style={{ marginBottom: 8, fontWeight: 'bold' }}>
              {t('cashier.qrCode') || '二维码'}:
            </p>
            <QRCode value={generatedLink} size={200} />
          </div>
        </Space>
      </Modal>
    </div>
  )
}

export default CashierConfigPage

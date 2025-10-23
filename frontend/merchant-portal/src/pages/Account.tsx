import { useState, useEffect } from 'react'
import {
  Typography,
  Card,
  Row,
  Col,
  Statistic,
  Descriptions,
  Button,
  Form,
  Input,
  Modal,
  message,
  Tabs,
  Space,
  Tag,
  Divider,
  Popconfirm,
} from 'antd'
import {
  UserOutlined,
  DollarOutlined,
  TransactionOutlined,
  PercentageOutlined,
  SafetyOutlined,
  EditOutlined,
  KeyOutlined,
  LockOutlined,
  ReloadOutlined,
  CopyOutlined,
} from '@ant-design/icons'
import { merchantService, Merchant, MerchantBalance, MerchantStats } from '../services/merchantService'

const { Title, Paragraph } = Typography
const { TabPane } = Tabs

const Account = () => {
  const [loading, setLoading] = useState(false)
  const [merchant, setMerchant] = useState<Merchant | null>(null)
  const [balance, setBalance] = useState<MerchantBalance | null>(null)
  const [stats, setStats] = useState<MerchantStats | null>(null)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [passwordModalVisible, setPasswordModalVisible] = useState(false)
  const [editForm] = Form.useForm()
  const [passwordForm] = Form.useForm()

  useEffect(() => {
    loadProfile()
    loadBalance()
    loadStats()
  }, [])

  const loadProfile = async () => {
    setLoading(true)
    try {
      const response = await merchantService.getProfile()
      setMerchant(response.data)
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const loadBalance = async () => {
    try {
      const response = await merchantService.getBalance()
      setBalance(response.data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const loadStats = async () => {
    try {
      const response = await merchantService.getStats({})
      setStats(response.data)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleEdit = () => {
    if (merchant) {
      editForm.setFieldsValue({
        contact_name: merchant.contact_name,
        contact_email: merchant.contact_email,
        contact_phone: merchant.contact_phone,
        website: merchant.website,
        description: merchant.description,
        callback_url: merchant.callback_url,
        return_url: merchant.return_url,
      })
      setEditModalVisible(true)
    }
  }

  const handleEditSubmit = async () => {
    try {
      const values = await editForm.validateFields()
      await merchantService.updateProfile(values)
      message.success('个人信息更新成功')
      setEditModalVisible(false)
      loadProfile()
    } catch (error) {
      // Error handled by interceptor or validation
    }
  }

  const handleChangePassword = () => {
    passwordForm.resetFields()
    setPasswordModalVisible(true)
  }

  const handlePasswordSubmit = async () => {
    try {
      const values = await passwordForm.validateFields()
      await merchantService.changePassword(values)
      message.success('密码修改成功')
      setPasswordModalVisible(false)
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleRegenerateApiKey = async () => {
    try {
      const response = await merchantService.regenerateApiKey()
      message.success('API密钥已重新生成')
      Modal.info({
        title: '新的API密钥',
        width: 600,
        content: (
          <div>
            <Paragraph>
              <strong>API Key:</strong>
              <br />
              <code>{response.data.api_key}</code>
            </Paragraph>
            <Paragraph>
              <strong>API Secret:</strong>
              <br />
              <code>{response.data.api_secret}</code>
            </Paragraph>
            <Paragraph type="danger">
              请妥善保管您的API密钥，此窗口关闭后将无法再次查看API Secret。
            </Paragraph>
          </div>
        ),
      })
      loadProfile()
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    message.success('已复制到剪贴板')
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      active: 'success',
      pending: 'processing',
      suspended: 'warning',
      rejected: 'error',
    }
    return colors[status] || 'default'
  }

  const getStatusText = (status: string) => {
    const texts: Record<string, string> = {
      active: '正常',
      pending: '待审核',
      suspended: '已暂停',
      rejected: '已拒绝',
    }
    return texts[status] || status
  }

  if (!merchant) {
    return <div>加载中...</div>
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>账户信息</Title>
        <Space>
          <Button icon={<LockOutlined />} onClick={handleChangePassword}>
            修改密码
          </Button>
          <Button type="primary" icon={<EditOutlined />} onClick={handleEdit}>
            编辑信息
          </Button>
        </Space>
      </div>

      {/* Balance and Stats Cards */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="可用余额"
              value={(balance?.available_balance || 0) / 100}
              precision={2}
              prefix={<DollarOutlined />}
              suffix={balance?.currency || 'USD'}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="冻结金额"
              value={(balance?.frozen_balance || 0) / 100}
              precision={2}
              prefix={<DollarOutlined />}
              suffix={balance?.currency || 'USD'}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="总交易额"
              value={(stats?.total_amount || 0) / 100}
              precision={2}
              prefix={<TransactionOutlined />}
              suffix="USD"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功率"
              value={(stats?.success_rate || 0) * 100}
              precision={2}
              prefix={<PercentageOutlined />}
              suffix="%"
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Tabs for different sections */}
      <Card>
        <Tabs defaultActiveKey="profile">
          <TabPane tab="基本信息" key="profile">
            <Descriptions title="商户信息" bordered column={2}>
              <Descriptions.Item label="商户名称">{merchant.name}</Descriptions.Item>
              <Descriptions.Item label="商户代码">{merchant.code}</Descriptions.Item>
              <Descriptions.Item label="商户类型">{merchant.type}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(merchant.status)}>
                  {getStatusText(merchant.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="联系人">{merchant.contact_name}</Descriptions.Item>
              <Descriptions.Item label="联系邮箱">{merchant.contact_email}</Descriptions.Item>
              <Descriptions.Item label="联系电话">{merchant.contact_phone || '-'}</Descriptions.Item>
              <Descriptions.Item label="营业执照">{merchant.business_license || '-'}</Descriptions.Item>
              <Descriptions.Item label="网站" span={2}>{merchant.website || '-'}</Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>{merchant.description || '-'}</Descriptions.Item>
              <Descriptions.Item label="回调地址" span={2}>
                {merchant.callback_url || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="返回地址" span={2}>
                {merchant.return_url || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="结算周期">{merchant.settlement_cycle} 天</Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(merchant.created_at).toLocaleString()}
              </Descriptions.Item>
              {merchant.approved_at && (
                <>
                  <Descriptions.Item label="审核时间">
                    {new Date(merchant.approved_at).toLocaleString()}
                  </Descriptions.Item>
                  <Descriptions.Item label="审核人">{merchant.approved_by || '-'}</Descriptions.Item>
                </>
              )}
            </Descriptions>
          </TabPane>

          <TabPane tab="API密钥" key="api">
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <Card
                title="API Key"
                extra={
                  <Button
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => copyToClipboard(merchant.api_key)}
                  >
                    复制
                  </Button>
                }
              >
                <code style={{ fontSize: 14 }}>{merchant.api_key}</code>
              </Card>

              <Card
                title="API Secret"
                extra={
                  <Popconfirm
                    title="重新生成API密钥"
                    description="重新生成后，旧的密钥将立即失效。确定要继续吗？"
                    onConfirm={handleRegenerateApiKey}
                  >
                    <Button
                      size="small"
                      icon={<ReloadOutlined />}
                      danger
                    >
                      重新生成
                    </Button>
                  </Popconfirm>
                }
              >
                <code style={{ fontSize: 14 }}>{'*'.repeat(32)}</code>
                <Paragraph type="secondary" style={{ marginTop: 8 }}>
                  出于安全考虑，API Secret不会显示。如需使用，请重新生成。
                </Paragraph>
              </Card>

              <Card title="使用说明">
                <Paragraph>
                  <strong>认证方式：</strong>
                  <br />
                  在API请求头中添加以下内容：
                </Paragraph>
                <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4 }}>
{`X-Merchant-Key: ${merchant.api_key}
X-Merchant-Secret: [您的API Secret]
Content-Type: application/json`}
                </pre>
                <Paragraph>
                  <strong>安全提示：</strong>
                </Paragraph>
                <ul>
                  <li>请妥善保管您的API密钥，不要泄露给他人</li>
                  <li>建议定期更换API密钥以提高安全性</li>
                  <li>生产环境中请使用HTTPS协议</li>
                  <li>可在风险配置中设置IP白名单</li>
                </ul>
              </Card>
            </Space>
          </TabPane>

          <TabPane tab="结算账户" key="settlement">
            {merchant.settlement_account ? (
              <Descriptions title="银行账户信息" bordered column={2}>
                <Descriptions.Item label="开户银行">
                  {merchant.settlement_account.bank_name}
                </Descriptions.Item>
                <Descriptions.Item label="开户支行">
                  {merchant.settlement_account.bank_branch || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="账户名称">
                  {merchant.settlement_account.account_name}
                </Descriptions.Item>
                <Descriptions.Item label="账户类型">
                  {merchant.settlement_account.account_type}
                </Descriptions.Item>
                <Descriptions.Item label="账户号码" span={2}>
                  {merchant.settlement_account.account_number}
                </Descriptions.Item>
              </Descriptions>
            ) : (
              <Card>
                <Paragraph>尚未配置结算账户，请联系客服添加。</Paragraph>
              </Card>
            )}
          </TabPane>

          <TabPane tab="费率配置" key="rate">
            {merchant.rate_config ? (
              <Descriptions title="费率信息" bordered column={2}>
                <Descriptions.Item label="支付渠道">
                  {merchant.rate_config.channel}
                </Descriptions.Item>
                <Descriptions.Item label="支付方式">
                  {merchant.rate_config.payment_method}
                </Descriptions.Item>
                <Descriptions.Item label="费率">
                  {(merchant.rate_config.rate * 100).toFixed(2)}%
                </Descriptions.Item>
                <Descriptions.Item label="固定手续费">
                  ${(merchant.rate_config.fixed_fee / 100).toFixed(2)}
                </Descriptions.Item>
              </Descriptions>
            ) : (
              <Card>
                <Paragraph>尚未配置费率，请联系客服设置。</Paragraph>
              </Card>
            )}
          </TabPane>

          <TabPane tab="风控配置" key="risk">
            {merchant.risk_config ? (
              <Descriptions title="风控规则" bordered column={2}>
                <Descriptions.Item label="单日限额">
                  ${(merchant.risk_config.daily_limit / 100).toFixed(2)}
                </Descriptions.Item>
                <Descriptions.Item label="单月限额">
                  ${(merchant.risk_config.monthly_limit / 100).toFixed(2)}
                </Descriptions.Item>
                <Descriptions.Item label="单笔限额" span={2}>
                  ${(merchant.risk_config.single_limit / 100).toFixed(2)}
                </Descriptions.Item>
                <Descriptions.Item label="回调重试次数">
                  {merchant.risk_config.callback_retry} 次
                </Descriptions.Item>
                <Descriptions.Item label="IP白名单" span={2}>
                  {merchant.risk_config.ip_whitelist?.length > 0 ? (
                    <Space wrap>
                      {merchant.risk_config.ip_whitelist.map((ip) => (
                        <Tag key={ip}>{ip}</Tag>
                      ))}
                    </Space>
                  ) : (
                    '未设置'
                  )}
                </Descriptions.Item>
              </Descriptions>
            ) : (
              <Card>
                <Paragraph>尚未配置风控规则，请联系客服设置。</Paragraph>
              </Card>
            )}
          </TabPane>
        </Tabs>
      </Card>

      {/* Edit Profile Modal */}
      <Modal
        title="编辑商户信息"
        open={editModalVisible}
        onOk={handleEditSubmit}
        onCancel={() => setEditModalVisible(false)}
        width={700}
      >
        <Form form={editForm} layout="vertical">
          <Form.Item
            name="contact_name"
            label="联系人"
            rules={[{ required: true, message: '请输入联系人' }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            name="contact_email"
            label="联系邮箱"
            rules={[
              { required: true, message: '请输入联系邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input />
          </Form.Item>

          <Form.Item name="contact_phone" label="联系电话">
            <Input />
          </Form.Item>

          <Form.Item name="website" label="网站">
            <Input placeholder="https://example.com" />
          </Form.Item>

          <Form.Item name="description" label="商户描述">
            <Input.TextArea rows={3} />
          </Form.Item>

          <Form.Item
            name="callback_url"
            label="回调地址"
            rules={[{ required: true, message: '请输入回调地址' }]}
          >
            <Input placeholder="https://example.com/callback" />
          </Form.Item>

          <Form.Item
            name="return_url"
            label="返回地址"
            rules={[{ required: true, message: '请输入返回地址' }]}
          >
            <Input placeholder="https://example.com/return" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Change Password Modal */}
      <Modal
        title="修改密码"
        open={passwordModalVisible}
        onOk={handlePasswordSubmit}
        onCancel={() => setPasswordModalVisible(false)}
        width={500}
      >
        <Form form={passwordForm} layout="vertical">
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
        </Form>
      </Modal>
    </div>
  )
}

export default Account

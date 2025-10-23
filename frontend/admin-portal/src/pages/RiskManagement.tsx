import { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Tabs,
  Modal,
  Descriptions,
  message,
  Form,
  InputNumber,
  Switch,
  Tooltip,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  SearchOutlined,
  ReloadOutlined,
  EyeOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'

interface RiskAssessment {
  id: string
  payment_no: string
  merchant_id: string
  merchant_name: string
  risk_score: number
  risk_level: string
  factors: string[]
  status: string
  created_at: string
}

interface RiskRule {
  id: string
  rule_name: string
  rule_type: string
  condition: string
  action: string
  priority: number
  enabled: boolean
  created_at: string
}

interface Blacklist {
  id: string
  item_type: string
  item_value: string
  reason: string
  created_by: string
  created_at: string
  expired_at: string | null
}

const RiskManagement = () => {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState('assessments')

  // Risk Assessments
  const [assessments, setAssessments] = useState<RiskAssessment[]>([])
  const [assessmentsLoading, setAssessmentsLoading] = useState(false)
  const [assessmentsTotal, setAssessmentsTotal] = useState(0)
  const [assessmentsPage, setAssessmentsPage] = useState(1)
  const [assessmentsPageSize, setAssessmentsPageSize] = useState(20)
  const [selectedAssessment, setSelectedAssessment] = useState<RiskAssessment | null>(null)
  const [assessmentDetailVisible, setAssessmentDetailVisible] = useState(false)

  // Risk Rules
  const [rules, setRules] = useState<RiskRule[]>([])
  const [rulesLoading, setRulesLoading] = useState(false)
  const [rulesTotal, setRulesTotal] = useState(0)
  const [rulesPage, setRulesPage] = useState(1)
  const [rulesPageSize, setRulesPageSize] = useState(20)
  const [ruleModalVisible, setRuleModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<RiskRule | null>(null)
  const [ruleForm] = Form.useForm()

  // Blacklist
  const [blacklist, setBlacklist] = useState<Blacklist[]>([])
  const [blacklistLoading, setBlacklistLoading] = useState(false)
  const [blacklistTotal, setBlacklistTotal] = useState(0)
  const [blacklistPage, setBlacklistPage] = useState(1)
  const [blacklistPageSize, setBlacklistPageSize] = useState(20)
  const [blacklistModalVisible, setBlacklistModalVisible] = useState(false)
  const [blacklistForm] = Form.useForm()

  const fetchAssessments = async () => {
    setAssessmentsLoading(true)
    try {
      // Mock data
      const mockData: RiskAssessment[] = Array.from({ length: assessmentsPageSize }, (_, i) => ({
        id: `assessment-${assessmentsPage}-${i}`,
        payment_no: `PAY${Date.now()}${i}`,
        merchant_id: `merchant-${i % 3}`,
        merchant_name: `商户${i % 3 + 1}`,
        risk_score: Math.floor(Math.random() * 100),
        risk_level: ['low', 'medium', 'high', 'critical'][i % 4],
        factors: ['IP异常', '金额异常', '频率异常'].slice(0, Math.floor(Math.random() * 3) + 1),
        status: ['approved', 'rejected', 'reviewing'][i % 3],
        created_at: new Date().toISOString(),
      }))
      setAssessments(mockData)
      setAssessmentsTotal(100)
    } catch (error) {
      message.error(t('common.operationFailed'))
    } finally {
      setAssessmentsLoading(false)
    }
  }

  const fetchRules = async () => {
    setRulesLoading(true)
    try {
      // Mock data
      const mockData: RiskRule[] = Array.from({ length: rulesPageSize }, (_, i) => ({
        id: `rule-${rulesPage}-${i}`,
        rule_name: `风控规则${i + 1}`,
        rule_type: ['amount', 'frequency', 'location', 'device'][i % 4],
        condition: '金额 > 10000',
        action: ['reject', 'review', 'alert'][i % 3],
        priority: (i % 5) + 1,
        enabled: i % 2 === 0,
        created_at: new Date().toISOString(),
      }))
      setRules(mockData)
      setRulesTotal(50)
    } catch (error) {
      message.error(t('common.operationFailed'))
    } finally {
      setRulesLoading(false)
    }
  }

  const fetchBlacklist = async () => {
    setBlacklistLoading(true)
    try {
      // Mock data
      const mockData: Blacklist[] = Array.from({ length: blacklistPageSize }, (_, i) => ({
        id: `blacklist-${blacklistPage}-${i}`,
        item_type: ['ip', 'email', 'card', 'device'][i % 4],
        item_value: `192.168.1.${i}`,
        reason: '疑似欺诈',
        created_by: 'admin',
        created_at: new Date().toISOString(),
        expired_at: i % 2 === 0 ? new Date(Date.now() + 86400000 * 30).toISOString() : null,
      }))
      setBlacklist(mockData)
      setBlacklistTotal(80)
    } catch (error) {
      message.error(t('common.operationFailed'))
    } finally {
      setBlacklistLoading(false)
    }
  }

  useEffect(() => {
    if (activeTab === 'assessments') {
      fetchAssessments()
    } else if (activeTab === 'rules') {
      fetchRules()
    } else if (activeTab === 'blacklist') {
      fetchBlacklist()
    }
  }, [activeTab, assessmentsPage, assessmentsPageSize, rulesPage, rulesPageSize, blacklistPage, blacklistPageSize])

  const getRiskLevelTag = (level: string) => {
    const levelMap: Record<string, { color: string; text: string }> = {
      low: { color: 'success', text: t('risk.levelLow') },
      medium: { color: 'warning', text: t('risk.levelMedium') },
      high: { color: 'orange', text: t('risk.levelHigh') },
      critical: { color: 'error', text: t('risk.levelCritical') },
    }
    const config = levelMap[level] || { color: 'default', text: level }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      approved: { color: 'success', text: t('risk.statusApproved') },
      rejected: { color: 'error', text: t('risk.statusRejected') },
      reviewing: { color: 'processing', text: t('risk.statusReviewing') },
    }
    const config = statusMap[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const handleViewAssessment = (record: RiskAssessment) => {
    setSelectedAssessment(record)
    setAssessmentDetailVisible(true)
  }

  const handleAddRule = () => {
    setEditingRule(null)
    ruleForm.resetFields()
    setRuleModalVisible(true)
  }

  const handleEditRule = (record: RiskRule) => {
    setEditingRule(record)
    ruleForm.setFieldsValue(record)
    setRuleModalVisible(true)
  }

  const handleDeleteRule = async (record: RiskRule) => {
    try {
      // TODO: API call
      message.success(t('common.deleteSuccess'))
      fetchRules()
    } catch (error) {
      message.error(t('common.operationFailed'))
    }
  }

  const handleSaveRule = async () => {
    try {
      const values = await ruleForm.validateFields()
      // TODO: API call
      message.success(t('common.saveSuccess'))
      setRuleModalVisible(false)
      fetchRules()
    } catch (error) {
      // Validation failed
    }
  }

  const handleAddBlacklist = () => {
    blacklistForm.resetFields()
    setBlacklistModalVisible(true)
  }

  const handleSaveBlacklist = async () => {
    try {
      const values = await blacklistForm.validateFields()
      // TODO: API call
      message.success(t('common.saveSuccess'))
      setBlacklistModalVisible(false)
      fetchBlacklist()
    } catch (error) {
      // Validation failed
    }
  }

  const handleDeleteBlacklist = async (record: Blacklist) => {
    try {
      // TODO: API call
      message.success(t('common.deleteSuccess'))
      fetchBlacklist()
    } catch (error) {
      message.error(t('common.operationFailed'))
    }
  }

  const assessmentColumns: ColumnsType<RiskAssessment> = [
    {
      title: t('risk.paymentNo'),
      dataIndex: 'payment_no',
      key: 'payment_no',
      width: 180,
    },
    {
      title: t('risk.merchantName'),
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
    },
    {
      title: t('risk.riskScore'),
      dataIndex: 'risk_score',
      key: 'risk_score',
      width: 100,
      render: (score: number) => <span style={{ fontWeight: 'bold' }}>{score}</span>,
    },
    {
      title: t('risk.riskLevel'),
      dataIndex: 'risk_level',
      key: 'risk_level',
      width: 120,
      render: (level: string) => getRiskLevelTag(level),
    },
    {
      title: t('risk.riskFactors'),
      dataIndex: 'factors',
      key: 'factors',
      width: 200,
      render: (factors: string[]) => (
        <>
          {factors.map((factor, idx) => (
            <Tag key={idx} color="orange" style={{ marginBottom: 4 }}>
              {factor}
            </Tag>
          ))}
        </>
      ),
    },
    {
      title: t('risk.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: t('common.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 100,
      fixed: 'right',
      render: (_: unknown, record: RiskAssessment) => (
        <Tooltip title={t('risk.viewDetail')}>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewAssessment(record)}
          />
        </Tooltip>
      ),
    },
  ]

  const ruleColumns: ColumnsType<RiskRule> = [
    {
      title: t('risk.ruleName'),
      dataIndex: 'rule_name',
      key: 'rule_name',
      width: 180,
    },
    {
      title: t('risk.ruleType'),
      dataIndex: 'rule_type',
      key: 'rule_type',
      width: 120,
      render: (type: string) => <Tag>{type}</Tag>,
    },
    {
      title: t('risk.condition'),
      dataIndex: 'condition',
      key: 'condition',
      width: 200,
    },
    {
      title: t('risk.action'),
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: string) => <Tag color="blue">{action}</Tag>,
    },
    {
      title: t('risk.priority'),
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
    },
    {
      title: t('risk.enabled'),
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>
          {enabled ? t('risk.enabledYes') : t('risk.enabledNo')}
        </Tag>
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 150,
      fixed: 'right',
      render: (_: unknown, record: RiskRule) => (
        <Space size="small">
          <Tooltip title={t('common.edit')}>
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEditRule(record)}
            />
          </Tooltip>
          <Tooltip title={t('common.delete')}>
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDeleteRule(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ]

  const blacklistColumns: ColumnsType<Blacklist> = [
    {
      title: t('risk.itemType'),
      dataIndex: 'item_type',
      key: 'item_type',
      width: 100,
      render: (type: string) => <Tag color="red">{type.toUpperCase()}</Tag>,
    },
    {
      title: t('risk.itemValue'),
      dataIndex: 'item_value',
      key: 'item_value',
      width: 200,
    },
    {
      title: t('risk.reason'),
      dataIndex: 'reason',
      key: 'reason',
      width: 200,
    },
    {
      title: t('risk.createdBy'),
      dataIndex: 'created_by',
      key: 'created_by',
      width: 120,
    },
    {
      title: t('common.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: t('risk.expiredAt'),
      dataIndex: 'expired_at',
      key: 'expired_at',
      width: 180,
      render: (date: string | null) =>
        date ? dayjs(date).format('YYYY-MM-DD HH:mm:ss') : t('risk.permanent'),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 100,
      fixed: 'right',
      render: (_: unknown, record: Blacklist) => (
        <Tooltip title={t('common.delete')}>
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteBlacklist(record)}
          />
        </Tooltip>
      ),
    },
  ]

  const tabItems = [
    {
      key: 'assessments',
      label: t('risk.riskAssessments'),
      children: (
        <Card>
          <Table
            columns={assessmentColumns}
            dataSource={assessments}
            loading={assessmentsLoading}
            rowKey="id"
            scroll={{ x: 1200 }}
            pagination={{
              current: assessmentsPage,
              pageSize: assessmentsPageSize,
              total: assessmentsTotal,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => t('common.total', { count: total }),
              onChange: (page, pageSize) => {
                setAssessmentsPage(page)
                setAssessmentsPageSize(pageSize)
              },
            }}
          />
        </Card>
      ),
    },
    {
      key: 'rules',
      label: t('risk.riskRules'),
      children: (
        <Card
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAddRule}>
              {t('risk.addRule')}
            </Button>
          }
        >
          <Table
            columns={ruleColumns}
            dataSource={rules}
            loading={rulesLoading}
            rowKey="id"
            scroll={{ x: 1000 }}
            pagination={{
              current: rulesPage,
              pageSize: rulesPageSize,
              total: rulesTotal,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => t('common.total', { count: total }),
              onChange: (page, pageSize) => {
                setRulesPage(page)
                setRulesPageSize(pageSize)
              },
            }}
          />
        </Card>
      ),
    },
    {
      key: 'blacklist',
      label: t('risk.blacklist'),
      children: (
        <Card
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAddBlacklist}>
              {t('risk.addBlacklist')}
            </Button>
          }
        >
          <Table
            columns={blacklistColumns}
            dataSource={blacklist}
            loading={blacklistLoading}
            rowKey="id"
            scroll={{ x: 1000 }}
            pagination={{
              current: blacklistPage,
              pageSize: blacklistPageSize,
              total: blacklistTotal,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => t('common.total', { count: total }),
              onChange: (page, pageSize) => {
                setBlacklistPage(page)
                setBlacklistPageSize(pageSize)
              },
            }}
          />
        </Card>
      ),
    },
  ]

  return (
    <div>
      <Tabs activeKey={activeTab} items={tabItems} onChange={setActiveTab} />

      {/* Assessment Detail Modal */}
      <Modal
        title={t('risk.assessmentDetail')}
        open={assessmentDetailVisible}
        onCancel={() => setAssessmentDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setAssessmentDetailVisible(false)}>
            {t('common.cancel')}
          </Button>,
        ]}
        width={700}
      >
        {selectedAssessment && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label={t('risk.paymentNo')} span={2}>
              {selectedAssessment.payment_no}
            </Descriptions.Item>
            <Descriptions.Item label={t('risk.merchantName')}>
              {selectedAssessment.merchant_name}
            </Descriptions.Item>
            <Descriptions.Item label={t('risk.riskScore')}>
              <span style={{ fontSize: 18, fontWeight: 'bold' }}>
                {selectedAssessment.risk_score}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label={t('risk.riskLevel')}>
              {getRiskLevelTag(selectedAssessment.risk_level)}
            </Descriptions.Item>
            <Descriptions.Item label={t('risk.status')}>
              {getStatusTag(selectedAssessment.status)}
            </Descriptions.Item>
            <Descriptions.Item label={t('risk.riskFactors')} span={2}>
              {selectedAssessment.factors.map((factor, idx) => (
                <Tag key={idx} color="orange" style={{ marginBottom: 4 }}>
                  {factor}
                </Tag>
              ))}
            </Descriptions.Item>
            <Descriptions.Item label={t('common.createdAt')} span={2}>
              {dayjs(selectedAssessment.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* Rule Modal */}
      <Modal
        title={editingRule ? t('risk.editRule') : t('risk.addRule')}
        open={ruleModalVisible}
        onOk={handleSaveRule}
        onCancel={() => setRuleModalVisible(false)}
        okText={t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Form form={ruleForm} layout="vertical">
          <Form.Item
            name="rule_name"
            label={t('risk.ruleName')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="rule_type"
            label={t('risk.ruleType')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Select>
              <Select.Option value="amount">Amount</Select.Option>
              <Select.Option value="frequency">Frequency</Select.Option>
              <Select.Option value="location">Location</Select.Option>
              <Select.Option value="device">Device</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="condition"
            label={t('risk.condition')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item
            name="action"
            label={t('risk.action')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Select>
              <Select.Option value="reject">Reject</Select.Option>
              <Select.Option value="review">Review</Select.Option>
              <Select.Option value="alert">Alert</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="priority" label={t('risk.priority')} initialValue={1}>
            <InputNumber min={1} max={10} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="enabled" label={t('risk.enabled')} valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* Blacklist Modal */}
      <Modal
        title={t('risk.addBlacklist')}
        open={blacklistModalVisible}
        onOk={handleSaveBlacklist}
        onCancel={() => setBlacklistModalVisible(false)}
        okText={t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Form form={blacklistForm} layout="vertical">
          <Form.Item
            name="item_type"
            label={t('risk.itemType')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Select>
              <Select.Option value="ip">IP</Select.Option>
              <Select.Option value="email">Email</Select.Option>
              <Select.Option value="card">Card</Select.Option>
              <Select.Option value="device">Device</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="item_value"
            label={t('risk.itemValue')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="reason"
            label={t('risk.reason')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default RiskManagement

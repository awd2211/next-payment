import React, { useState } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Modal,
  Form,
  message,
  Descriptions,
  InputNumber,
  Switch,
  Alert,
  Statistic,
  Row,
  Col,
  Progress,
} from 'antd'
import {
  SearchOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
  DollarOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

interface MerchantLimit {
  id: string
  merchant_id: string
  merchant_name: string
  daily_transaction_limit: number
  daily_amount_limit: number
  monthly_transaction_limit: number
  monthly_amount_limit: number
  single_transaction_min: number
  single_transaction_max: number
  current_daily_count: number
  current_daily_amount: number
  current_monthly_count: number
  current_monthly_amount: number
  is_enabled: boolean
  alert_threshold: number
  created_at: string
  updated_at: string
}

export default function MerchantLimits() {
  const [loading, setLoading] = useState(false)
  const [limits, setLimits] = useState<MerchantLimit[]>([
    {
      id: '1',
      merchant_id: 'MCH-001',
      merchant_name: 'Tech Store',
      daily_transaction_limit: 1000,
      daily_amount_limit: 100000,
      monthly_transaction_limit: 30000,
      monthly_amount_limit: 3000000,
      single_transaction_min: 1,
      single_transaction_max: 5000,
      current_daily_count: 523,
      current_daily_amount: 52300,
      current_monthly_count: 15230,
      current_monthly_amount: 1523000,
      is_enabled: true,
      alert_threshold: 80,
      created_at: '2024-01-01 00:00:00',
      updated_at: '2024-02-01 10:00:00',
    },
    {
      id: '2',
      merchant_id: 'MCH-002',
      merchant_name: 'Fashion Store',
      daily_transaction_limit: 2000,
      daily_amount_limit: 200000,
      monthly_transaction_limit: 60000,
      monthly_amount_limit: 6000000,
      single_transaction_min: 1,
      single_transaction_max: 10000,
      current_daily_count: 1850,
      current_daily_amount: 185000,
      current_monthly_count: 52000,
      current_monthly_amount: 5200000,
      is_enabled: true,
      alert_threshold: 90,
      created_at: '2024-01-01 00:00:00',
      updated_at: '2024-02-01 11:30:00',
    },
  ])

  const [detailVisible, setDetailVisible] = useState(false)
  const [editVisible, setEditVisible] = useState(false)
  const [selectedLimit, setSelectedLimit] = useState<MerchantLimit | null>(null)
  const [form] = Form.useForm()

  const columns: ColumnsType<MerchantLimit> = [
    {
      title: '商户ID',
      dataIndex: 'merchant_id',
      key: 'merchant_id',
      width: 120,
    },
    {
      title: '商户名称',
      dataIndex: 'merchant_name',
      key: 'merchant_name',
      width: 150,
    },
    {
      title: '日交易限额',
      key: 'daily_limit',
      width: 180,
      render: (_, record) => (
        <div>
          <div style={{ fontSize: 12 }}>
            笔数: {record.current_daily_count.toLocaleString()} / {record.daily_transaction_limit.toLocaleString()}
          </div>
          <Progress
            percent={(record.current_daily_count / record.daily_transaction_limit) * 100}
            size="small"
            showInfo={false}
            status={
              (record.current_daily_count / record.daily_transaction_limit) * 100 >= record.alert_threshold
                ? 'exception'
                : 'normal'
            }
          />
        </div>
      ),
    },
    {
      title: '日金额限额',
      key: 'daily_amount',
      width: 180,
      render: (_, record) => {
        const percentage = (record.current_daily_amount / record.daily_amount_limit) * 100
        return (
          <div>
            <div style={{ fontSize: 12 }}>
              ${record.current_daily_amount.toLocaleString()} / ${record.daily_amount_limit.toLocaleString()}
            </div>
            <Progress
              percent={percentage}
              size="small"
              showInfo={false}
              status={percentage >= record.alert_threshold ? 'exception' : 'normal'}
            />
          </div>
        )
      },
    },
    {
      title: '单笔限额',
      key: 'single_limit',
      width: 150,
      render: (_, record) => (
        <span>
          ${record.single_transaction_min} - ${record.single_transaction_max.toLocaleString()}
        </span>
      ),
    },
    {
      title: '预警阈值',
      dataIndex: 'alert_threshold',
      key: 'alert_threshold',
      width: 100,
      render: (threshold: number) => `${threshold}%`,
    },
    {
      title: '状态',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      width: 100,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'red'}>{enabled ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
    },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            查看
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
        </Space>
      ),
    },
  ]

  const handleViewDetail = (limit: MerchantLimit) => {
    setSelectedLimit(limit)
    setDetailVisible(true)
  }

  const handleEdit = (limit: MerchantLimit) => {
    setSelectedLimit(limit)
    form.setFieldsValue(limit)
    setEditVisible(true)
  }

  const handleEditSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      // TODO: Call API to update merchant limits
      await new Promise((resolve) => setTimeout(resolve, 1000))

      message.success('商户限额已更新')
      setEditVisible(false)

      // Update limits in list
      setLimits((prev) =>
        prev.map((limit) =>
          limit.id === selectedLimit?.id ? { ...limit, ...values } : limit
        )
      )
    } catch (error) {
      console.error('Update limit error:', error)
    } finally {
      setLoading(false)
    }
  }

  // Calculate statistics
  const warningCount = limits.filter(
    (limit) =>
      (limit.current_daily_count / limit.daily_transaction_limit) * 100 >= limit.alert_threshold ||
      (limit.current_daily_amount / limit.daily_amount_limit) * 100 >= limit.alert_threshold
  ).length

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic title="总商户数" value={limits.length} prefix={<DollarOutlined />} />
          </Col>
          <Col span={6}>
            <Statistic
              title="已启用"
              value={limits.filter((l) => l.is_enabled).length}
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="已禁用"
              value={limits.filter((l) => !l.is_enabled).length}
              valueStyle={{ color: '#999' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="预警数量"
              value={warningCount}
              prefix={<WarningOutlined />}
              valueStyle={{ color: warningCount > 0 ? '#faad14' : '#52c41a' }}
            />
          </Col>
        </Row>
      </Card>

      <Card
        title="商户限额管理"
        extra={
          <Space>
            <Input
              placeholder="搜索商户ID/名称"
              prefix={<SearchOutlined />}
              style={{ width: 200 }}
            />
            <Select placeholder="状态筛选" style={{ width: 120 }} allowClear>
              <Select.Option value="enabled">已启用</Select.Option>
              <Select.Option value="disabled">已禁用</Select.Option>
            </Select>
            <Select placeholder="预警状态" style={{ width: 120 }} allowClear>
              <Select.Option value="normal">正常</Select.Option>
              <Select.Option value="warning">预警中</Select.Option>
            </Select>
            <Button type="primary" icon={<SearchOutlined />}>
              搜索
            </Button>
          </Space>
        }
      >
        {warningCount > 0 && (
          <Alert
            message={`${warningCount} 个商户达到预警阈值`}
            description="部分商户的交易量或交易金额已达到预警阈值，请及时关注"
            type="warning"
            showIcon
            closable
            style={{ marginBottom: 16 }}
          />
        )}

        <Table
          columns={columns}
          dataSource={limits}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1600 }}
          pagination={{
            total: limits.length,
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      {/* Detail Modal */}
      <Modal
        title="商户限额详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
      >
        {selectedLimit && (
          <div>
            <Descriptions title="基本信息" column={2} bordered>
              <Descriptions.Item label="商户ID">{selectedLimit.merchant_id}</Descriptions.Item>
              <Descriptions.Item label="商户名称">{selectedLimit.merchant_name}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={selectedLimit.is_enabled ? 'green' : 'red'}>
                  {selectedLimit.is_enabled ? '启用' : '禁用'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="预警阈值">{selectedLimit.alert_threshold}%</Descriptions.Item>
            </Descriptions>

            <Descriptions title="单笔限额" column={2} bordered style={{ marginTop: 16 }}>
              <Descriptions.Item label="最小金额">
                ${selectedLimit.single_transaction_min}
              </Descriptions.Item>
              <Descriptions.Item label="最大金额">
                ${selectedLimit.single_transaction_max.toLocaleString()}
              </Descriptions.Item>
            </Descriptions>

            <Descriptions title="日限额" column={2} bordered style={{ marginTop: 16 }}>
              <Descriptions.Item label="交易笔数限额">
                {selectedLimit.daily_transaction_limit.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="当前交易笔数">
                {selectedLimit.current_daily_count.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="交易金额限额">
                ${selectedLimit.daily_amount_limit.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="当前交易金额">
                ${selectedLimit.current_daily_amount.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="笔数使用率" span={2}>
                <Progress
                  percent={(selectedLimit.current_daily_count / selectedLimit.daily_transaction_limit) * 100}
                  status={
                    (selectedLimit.current_daily_count / selectedLimit.daily_transaction_limit) * 100 >=
                    selectedLimit.alert_threshold
                      ? 'exception'
                      : 'normal'
                  }
                />
              </Descriptions.Item>
              <Descriptions.Item label="金额使用率" span={2}>
                <Progress
                  percent={(selectedLimit.current_daily_amount / selectedLimit.daily_amount_limit) * 100}
                  status={
                    (selectedLimit.current_daily_amount / selectedLimit.daily_amount_limit) * 100 >=
                    selectedLimit.alert_threshold
                      ? 'exception'
                      : 'normal'
                  }
                />
              </Descriptions.Item>
            </Descriptions>

            <Descriptions title="月限额" column={2} bordered style={{ marginTop: 16 }}>
              <Descriptions.Item label="交易笔数限额">
                {selectedLimit.monthly_transaction_limit.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="当前交易笔数">
                {selectedLimit.current_monthly_count.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="交易金额限额">
                ${selectedLimit.monthly_amount_limit.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="当前交易金额">
                ${selectedLimit.current_monthly_amount.toLocaleString()}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>

      {/* Edit Modal */}
      <Modal
        title="编辑商户限额"
        open={editVisible}
        onOk={handleEditSubmit}
        onCancel={() => setEditVisible(false)}
        confirmLoading={loading}
        width={700}
      >
        <Form form={form} layout="vertical">
          <Alert
            message="限额说明"
            description="修改限额后立即生效，请谨慎操作。建议根据商户业务规模合理设置限额。"
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

          <Form.Item label="启用状态" name="is_enabled" valuePropName="checked">
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="单笔最小金额 (USD)"
                name="single_transaction_min"
                rules={[{ required: true, message: '请输入单笔最小金额' }]}
              >
                <InputNumber min={0.01} precision={2} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="单笔最大金额 (USD)"
                name="single_transaction_max"
                rules={[{ required: true, message: '请输入单笔最大金额' }]}
              >
                <InputNumber min={1} precision={2} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="日交易笔数限额"
                name="daily_transaction_limit"
                rules={[{ required: true, message: '请输入日交易笔数限额' }]}
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="日交易金额限额 (USD)"
                name="daily_amount_limit"
                rules={[{ required: true, message: '请输入日交易金额限额' }]}
              >
                <InputNumber min={1} precision={2} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="月交易笔数限额"
                name="monthly_transaction_limit"
                rules={[{ required: true, message: '请输入月交易笔数限额' }]}
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="月交易金额限额 (USD)"
                name="monthly_amount_limit"
                rules={[{ required: true, message: '请输入月交易金额限额' }]}
              >
                <InputNumber min={1} precision={2} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="预警阈值 (%)"
            name="alert_threshold"
            rules={[{ required: true, message: '请输入预警阈值' }]}
            tooltip="当使用率达到此阈值时，系统将发送预警通知"
          >
            <InputNumber min={50} max={100} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

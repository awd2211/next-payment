import { useState, useEffect } from 'react'
import {
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Card,
  Row,
  Col,
  Statistic,
  Select,
  Input,
  DatePicker,
  Descriptions,
  Drawer,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  EyeOutlined,
  FileTextOutlined,
  UserOutlined,
  ApiOutlined,
  SafetyOutlined,
  DownloadOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { auditLogService, AuditLog, AuditLogStats } from '../services/auditLogService'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'

const { Title } = Typography
const { RangePicker } = DatePicker

const AuditLogs = () => {
  const [loading, setLoading] = useState(false)
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [stats, setStats] = useState<AuditLogStats | null>(null)
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false)

  // Filter states
  const [adminIdFilter, setAdminIdFilter] = useState<string | undefined>()
  const [actionFilter, setActionFilter] = useState<string | undefined>()
  const [resourceFilter, setResourceFilter] = useState<string | undefined>()
  const [methodFilter, setMethodFilter] = useState<string | undefined>()
  const [ipFilter, setIpFilter] = useState('')
  const [responseCodeFilter, setResponseCodeFilter] = useState<number | undefined>()
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)

  useEffect(() => {
    loadLogs()
  }, [page, pageSize, adminIdFilter, actionFilter, resourceFilter, methodFilter, ipFilter, responseCodeFilter, dateRange])

  useEffect(() => {
    loadStats()
  }, [])

  const loadLogs = async () => {
    setLoading(true)
    try {
      const response = await auditLogService.list({
        page,
        page_size: pageSize,
        admin_id: adminIdFilter,
        action: actionFilter,
        resource: resourceFilter,
        method: methodFilter,
        ip: ipFilter || undefined,
        response_code: responseCodeFilter,
        start_time: dateRange?.[0]?.toISOString(),
        end_time: dateRange?.[1]?.toISOString(),
      })
      // 响应拦截器已解包，直接使用数据
      if (response) {
        setLogs(response.list || [])
        setTotal(response.total || 0)
      }
    } catch (error) {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const loadStats = async () => {
    try {
      const response = await auditLogService.getStats({})
      // 响应拦截器已解包，直接使用数据
      if (response) {
        setStats(response)
      }
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const handleViewDetail = async (log: AuditLog) => {
    setSelectedLog(log)
    setDetailDrawerVisible(true)
  }

  const resetFilters = () => {
    setAdminIdFilter(undefined)
    setActionFilter(undefined)
    setResourceFilter(undefined)
    setMethodFilter(undefined)
    setIpFilter('')
    setResponseCodeFilter(undefined)
    setDateRange(null)
    setPage(1)
  }

  const handleExport = async () => {
    try {
      const response = await auditLogService.export({
        admin_id: adminIdFilter,
        action: actionFilter,
        resource: resourceFilter,
        method: methodFilter,
        ip: ipFilter || undefined,
        response_code: responseCodeFilter,
        start_time: dateRange?.[0]?.toISOString(),
        end_time: dateRange?.[1]?.toISOString(),
      })

      // Create a download link
      const blob = new Blob([response as any], { type: 'text/csv;charset=utf-8;' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `audit_logs_${dayjs().format('YYYYMMDD_HHmmss')}.csv`)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)

      message.success('导出成功')
    } catch (error) {
      // Error handled by interceptor
    }
  }

  const getMethodColor = (method: string) => {
    const colors: Record<string, string> = {
      GET: 'blue',
      POST: 'green',
      PUT: 'orange',
      DELETE: 'red',
      PATCH: 'purple',
    }
    return colors[method] || 'default'
  }

  const getResponseCodeColor = (code: number) => {
    if (code >= 200 && code < 300) return 'success'
    if (code >= 300 && code < 400) return 'processing'
    if (code >= 400 && code < 500) return 'warning'
    if (code >= 500) return 'error'
    return 'default'
  }

  const columns: ColumnsType<AuditLog> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作员',
      dataIndex: 'admin_username',
      key: 'admin_username',
      width: 120,
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 150,
      render: (action: string) => <Tag color="blue">{action}</Tag>,
    },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
      width: 120,
      render: (resource: string) => <Tag>{resource}</Tag>,
    },
    {
      title: '资源ID',
      dataIndex: 'resource_id',
      key: 'resource_id',
      width: 100,
      ellipsis: true,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
      render: (method: string) => <Tag color={getMethodColor(method)}>{method}</Tag>,
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
    },
    {
      title: 'IP地址',
      dataIndex: 'ip',
      key: 'ip',
      width: 130,
    },
    {
      title: '响应码',
      dataIndex: 'response_code',
      key: 'response_code',
      width: 100,
      render: (code: number) => (
        <Tag color={getResponseCodeColor(code)}>{code}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right',
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record)}
        >
          详情
        </Button>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2}>审计日志查询</Title>
        <Button
          type="primary"
          icon={<DownloadOutlined />}
          onClick={handleExport}
          loading={loading}
        >
          导出CSV
        </Button>
      </div>

      {/* Statistics Cards */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总日志数"
                value={stats.total_logs || 0}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="操作类型"
                value={stats?.action_distribution ? Object.keys(stats.action_distribution).length : 0}
                prefix={<ApiOutlined />}
                suffix="种"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="资源类型"
                value={stats?.resource_distribution ? Object.keys(stats.resource_distribution).length : 0}
                prefix={<SafetyOutlined />}
                suffix="种"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃管理员"
                value={stats?.top_admins?.length || 0}
                prefix={<UserOutlined />}
                suffix="人"
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* Filters */}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            placeholder="搜索IP地址"
            prefix={<SearchOutlined />}
            style={{ width: 180 }}
            allowClear
            value={ipFilter}
            onChange={(e) => {
              setIpFilter(e.target.value)
              setPage(1)
            }}
          />
          <Select
            placeholder="操作类型"
            style={{ width: 150 }}
            allowClear
            value={actionFilter}
            onChange={(value) => {
              setActionFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="create">创建</Select.Option>
            <Select.Option value="update">更新</Select.Option>
            <Select.Option value="delete">删除</Select.Option>
            <Select.Option value="login">登录</Select.Option>
            <Select.Option value="logout">登出</Select.Option>
            <Select.Option value="assign">分配</Select.Option>
            <Select.Option value="query">查询</Select.Option>
          </Select>
          <Select
            placeholder="资源类型"
            style={{ width: 150 }}
            allowClear
            value={resourceFilter}
            onChange={(value) => {
              setResourceFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="admin">管理员</Select.Option>
            <Select.Option value="role">角色</Select.Option>
            <Select.Option value="permission">权限</Select.Option>
            <Select.Option value="merchant">商户</Select.Option>
            <Select.Option value="config">配置</Select.Option>
            <Select.Option value="payment">支付</Select.Option>
            <Select.Option value="order">订单</Select.Option>
          </Select>
          <Select
            placeholder="请求方法"
            style={{ width: 120 }}
            allowClear
            value={methodFilter}
            onChange={(value) => {
              setMethodFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value="GET">GET</Select.Option>
            <Select.Option value="POST">POST</Select.Option>
            <Select.Option value="PUT">PUT</Select.Option>
            <Select.Option value="DELETE">DELETE</Select.Option>
            <Select.Option value="PATCH">PATCH</Select.Option>
          </Select>
          <Select
            placeholder="响应码"
            style={{ width: 120 }}
            allowClear
            value={responseCodeFilter}
            onChange={(value) => {
              setResponseCodeFilter(value)
              setPage(1)
            }}
          >
            <Select.Option value={200}>200 OK</Select.Option>
            <Select.Option value={201}>201 Created</Select.Option>
            <Select.Option value={400}>400 Bad Request</Select.Option>
            <Select.Option value={401}>401 Unauthorized</Select.Option>
            <Select.Option value={403}>403 Forbidden</Select.Option>
            <Select.Option value={404}>404 Not Found</Select.Option>
            <Select.Option value={500}>500 Server Error</Select.Option>
          </Select>
          <RangePicker
            showTime
            format="YYYY-MM-DD HH:mm:ss"
            placeholder={['开始时间', '结束时间']}
            value={dateRange}
            onChange={(dates) => {
              setDateRange(dates)
              setPage(1)
            }}
          />
          <Button onClick={resetFilters}>重置筛选</Button>
        </Space>
      </Card>

      {/* Table */}
      <Table
        columns={columns}
        dataSource={logs}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (page, pageSize) => {
            setPage(page)
            setPageSize(pageSize)
          },
        }}
        scroll={{ x: 1400 }}
      />

      {/* Detail Drawer */}
      <Drawer
        title="审计日志详情"
        placement="right"
        width={720}
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
      >
        {selectedLog && (
          <div>
            <Descriptions title="基本信息" bordered column={2}>
              <Descriptions.Item label="日志ID">{selectedLog.id}</Descriptions.Item>
              <Descriptions.Item label="操作时间">
                {dayjs(selectedLog.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label="操作员ID">{selectedLog.admin_id}</Descriptions.Item>
              <Descriptions.Item label="操作员用户名">{selectedLog.admin_username}</Descriptions.Item>
              <Descriptions.Item label="操作类型">
                <Tag color="blue">{selectedLog.action}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="资源类型">
                <Tag>{selectedLog.resource}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="资源ID">{selectedLog.resource_id || '-'}</Descriptions.Item>
              <Descriptions.Item label="请求方法">
                <Tag color={getMethodColor(selectedLog.method)}>{selectedLog.method}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="请求路径" span={2}>{selectedLog.path}</Descriptions.Item>
              <Descriptions.Item label="IP地址">{selectedLog.ip}</Descriptions.Item>
              <Descriptions.Item label="响应码">
                <Tag color={getResponseCodeColor(selectedLog.response_code)}>
                  {selectedLog.response_code}
                </Tag>
              </Descriptions.Item>
            </Descriptions>

            <Descriptions
              title="User Agent"
              bordered
              column={1}
              style={{ marginTop: 16 }}
            >
              <Descriptions.Item label="客户端信息">
                {selectedLog.user_agent || '-'}
              </Descriptions.Item>
            </Descriptions>

            {selectedLog.request_body && (
              <Card title="请求体" style={{ marginTop: 16 }}>
                <pre style={{ maxHeight: 200, overflow: 'auto', background: '#f5f5f5', padding: 12 }}>
                  {JSON.stringify(selectedLog.request_body, null, 2)}
                </pre>
              </Card>
            )}

            {selectedLog.response_body && (
              <Card title="响应体" style={{ marginTop: 16 }}>
                <pre style={{ maxHeight: 200, overflow: 'auto', background: '#f5f5f5', padding: 12 }}>
                  {JSON.stringify(selectedLog.response_body, null, 2)}
                </pre>
              </Card>
            )}

            {selectedLog.error_message && (
              <Card title="错误信息" style={{ marginTop: 16 }}>
                <Tag color="error">{selectedLog.error_message}</Tag>
              </Card>
            )}
          </div>
        )}
      </Drawer>
    </div>
  )
}

export default AuditLogs

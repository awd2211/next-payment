import { useState, useEffect } from 'react'
import {
  Card,
  List,
  Tag,
  Button,
  Space,
  Typography,
  Badge,
  Avatar,
  Empty,
  Spin,
  message,
  Tabs,
  Dropdown,
  Menu,
  Tooltip,
  Checkbox,
} from 'antd'
import {
  BellOutlined,
  CheckOutlined,
  DeleteOutlined,
  MoreOutlined,
  InfoCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  ClearOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const { Title, Text } = Typography
const { TabPane } = Tabs

interface Notification {
  id: string
  type: 'info' | 'success' | 'warning' | 'error'
  title: string
  content: string
  read: boolean
  created_at: string
  action_url?: string
}

const Notifications = () => {
  const { t: _t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [activeTab, setActiveTab] = useState('all')
  const [selectedIds, setSelectedIds] = useState<string[]>([])

  useEffect(() => {
    loadNotifications()
  }, [])

  const loadNotifications = async () => {
    try {
      // TODO: 调用notification-service API
      // const response = await notificationService.getNotifications()
      // setNotifications(response.data)

      // Mock data for now
      setNotifications([
        {
          id: '1',
          type: 'success',
          title: '支付成功',
          content: '订单 ORDER-20251024001 支付成功,金额 $99.99',
          read: false,
          created_at: new Date().toISOString(),
          action_url: '/orders/ORDER-20251024001',
        },
        {
          id: '2',
          type: 'warning',
          title: '风险提醒',
          content: '检测到异常交易模式,请及时查看',
          read: false,
          created_at: dayjs().subtract(2, 'hours').toISOString(),
        },
        {
          id: '3',
          type: 'info',
          title: '结算通知',
          content: '本周结算金额 $1,234.56 已到账',
          read: true,
          created_at: dayjs().subtract(1, 'day').toISOString(),
        },
        {
          id: '4',
          type: 'error',
          title: '支付失败',
          content: '订单 ORDER-20251023005 支付失败,原因: 余额不足',
          read: true,
          created_at: dayjs().subtract(2, 'days').toISOString(),
          action_url: '/orders/ORDER-20251023005',
        },
      ])
    } catch (error) {
      message.error('加载通知失败')
    } finally {
      setLoading(false)
    }
  }

  const handleMarkAsRead = (id: string) => {
    setNotifications(
      notifications.map(n => (n.id === id ? { ...n, read: true } : n))
    )
    message.success('已标记为已读')
  }

  const handleMarkAllAsRead = () => {
    setNotifications(notifications.map(n => ({ ...n, read: true })))
    setSelectedIds([])
    message.success('已全部标记为已读')
  }

  const handleDelete = (id: string) => {
    setNotifications(notifications.filter(n => n.id !== id))
    setSelectedIds(selectedIds.filter(sid => sid !== id))
    message.success('已删除')
  }

  const handleClearAll = () => {
    setNotifications([])
    setSelectedIds([])
    message.success('已清空所有通知')
  }

  const handleBatchDelete = () => {
    setNotifications(notifications.filter(n => !selectedIds.includes(n.id)))
    setSelectedIds([])
    message.success(`已删除 ${selectedIds.length} 条通知`)
  }

  const handleBatchMarkAsRead = () => {
    setNotifications(
      notifications.map(n =>
        selectedIds.includes(n.id) ? { ...n, read: true } : n
      )
    )
    setSelectedIds([])
    message.success(`已标记 ${selectedIds.length} 条为已读`)
  }

  const handleSelectAll = () => {
    if (selectedIds.length === filteredNotifications.length) {
      setSelectedIds([])
    } else {
      setSelectedIds(filteredNotifications.map(n => n.id))
    }
  }

  const getIcon = (type: Notification['type']) => {
    switch (type) {
      case 'info':
        return <InfoCircleOutlined style={{ color: '#1890ff' }} />
      case 'success':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />
      case 'warning':
        return <WarningOutlined style={{ color: '#faad14' }} />
      case 'error':
        return <CloseCircleOutlined style={{ color: '#f5222d' }} />
      default:
        return <BellOutlined />
    }
  }

  const getTypeColor = (type: Notification['type']) => {
    switch (type) {
      case 'info':
        return 'blue'
      case 'success':
        return 'green'
      case 'warning':
        return 'orange'
      case 'error':
        return 'red'
      default:
        return 'default'
    }
  }

  const unreadCount = notifications.filter(n => !n.read).length

  const filteredNotifications = notifications.filter(n => {
    if (activeTab === 'unread') return !n.read
    if (activeTab === 'read') return n.read
    return true
  })

  return (
    <div>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2} style={{ margin: 0 }}>
          <Badge count={unreadCount} offset={[10, 0]}>
            <BellOutlined /> 通知中心
          </Badge>
        </Title>
        <Space>
          <Tooltip title="刷新">
            <Button icon={<ReloadOutlined />} onClick={loadNotifications} style={{ borderRadius: 8 }} />
          </Tooltip>
          <Button
            onClick={handleMarkAllAsRead}
            disabled={unreadCount === 0}
            style={{ borderRadius: 8 }}
          >
            全部已读
          </Button>
          <Button
            danger
            onClick={handleClearAll}
            disabled={notifications.length === 0}
            style={{ borderRadius: 8 }}
          >
            清空所有
          </Button>
        </Space>
      </div>

      <Card style={{ borderRadius: 12 }}>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab={`全部 (${notifications.length})`} key="all" />
          <TabPane
            tab={
              <Badge count={unreadCount} offset={[5, 0]}>
                未读
              </Badge>
            }
            key="unread"
          />
          <TabPane tab={`已读 (${notifications.length - unreadCount})`} key="read" />
        </Tabs>

        {selectedIds.length > 0 && (
          <Space style={{ marginBottom: 16 }}>
            <Text>已选择 {selectedIds.length} 项</Text>
            <Button
              size="small"
              icon={<CheckOutlined />}
              onClick={handleBatchMarkAsRead}
              style={{ borderRadius: 8 }}
            >
              批量已读
            </Button>
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={handleBatchDelete}
              style={{ borderRadius: 8 }}
            >
              批量删除
            </Button>
            <Button
              size="small"
              icon={<ClearOutlined />}
              onClick={() => setSelectedIds([])}
              style={{ borderRadius: 8 }}
            >
              取消选择
            </Button>
          </Space>
        )}

        <div style={{ marginBottom: 16 }}>
          <Checkbox
            checked={selectedIds.length === filteredNotifications.length && filteredNotifications.length > 0}
            indeterminate={selectedIds.length > 0 && selectedIds.length < filteredNotifications.length}
            onChange={handleSelectAll}
          >
            全选
          </Checkbox>
        </div>

        {loading ? (
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <Spin size="large" tip="加载中..." />
          </div>
        ) : filteredNotifications.length === 0 ? (
          <Empty description="暂无通知" style={{ padding: '40px 0' }} />
        ) : (
          <List
            itemLayout="horizontal"
            dataSource={filteredNotifications}
            renderItem={item => (
              <List.Item
                style={{
                  background: item.read ? '#fff' : '#f6ffed',
                  padding: '16px',
                  marginBottom: '8px',
                  borderRadius: 12,
                  border: item.read ? '1px solid #f0f0f0' : '1px solid #b7eb8f',
                  transition: 'all 0.3s ease',
                  cursor: 'pointer',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.boxShadow = 'none'
                }}
                actions={[
                  <Checkbox
                    checked={selectedIds.includes(item.id)}
                    onChange={e => {
                      if (e.target.checked) {
                        setSelectedIds([...selectedIds, item.id])
                      } else {
                        setSelectedIds(selectedIds.filter(id => id !== item.id))
                      }
                    }}
                  />,
                  !item.read && (
                    <Tooltip title="标记已读">
                      <Button
                        type="link"
                        icon={<CheckOutlined />}
                        onClick={() => handleMarkAsRead(item.id)}
                        style={{ borderRadius: 8 }}
                      />
                    </Tooltip>
                  ),
                  <Dropdown
                    overlay={
                      <Menu>
                        <Menu.Item
                          key="delete"
                          icon={<DeleteOutlined />}
                          danger
                          onClick={() => handleDelete(item.id)}
                        >
                          删除
                        </Menu.Item>
                      </Menu>
                    }
                  >
                    <Button type="text" icon={<MoreOutlined />} style={{ borderRadius: 8 }} />
                  </Dropdown>,
                ].filter(Boolean)}
              >
                <List.Item.Meta
                  avatar={<Avatar icon={getIcon(item.type)} size={48} />}
                  title={
                    <Space>
                      <Text strong style={{ fontSize: 15 }}>{item.title}</Text>
                      <Tag color={getTypeColor(item.type)} style={{ borderRadius: 12 }}>
                        {item.type.toUpperCase()}
                      </Tag>
                      {!item.read && <Badge status="processing" text="未读" />}
                    </Space>
                  }
                  description={
                    <Space direction="vertical" style={{ width: '100%' }} size="small">
                      <Text>{item.content}</Text>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {dayjs(item.created_at).fromNow()}
                      </Text>
                      {item.action_url && (
                        <Button type="link" size="small" style={{ padding: 0, height: 'auto' }}>
                          查看详情 →
                        </Button>
                      )}
                    </Space>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </div>
  )
}

export default Notifications

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
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [activeTab, setActiveTab] = useState('all')

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
    message.success('已全部标记为已读')
  }

  const handleDelete = (id: string) => {
    setNotifications(notifications.filter(n => n.id !== id))
    message.success('已删除')
  }

  const handleClearAll = () => {
    setNotifications([])
    message.success('已清空所有通知')
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

  const filteredNotifications = notifications.filter(n => {
    if (activeTab === 'unread') return !n.read
    if (activeTab === 'read') return n.read
    return true
  })

  const unreadCount = notifications.filter(n => !n.read).length

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2} style={{ margin: 0 }}>
          <Badge count={unreadCount} offset={[10, 0]}>
            <BellOutlined /> 通知中心
          </Badge>
        </Title>
        <Space>
          <Button onClick={handleMarkAllAsRead} disabled={unreadCount === 0}>
            全部已读
          </Button>
          <Button danger onClick={handleClearAll} disabled={notifications.length === 0}>
            清空所有
          </Button>
        </Space>
      </div>

      <Card>
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

        {loading ? (
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <Spin size="large" tip="加载中..." />
          </div>
        ) : filteredNotifications.length === 0 ? (
          <Empty description="暂无通知" />
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
                  borderRadius: '4px',
                  border: item.read ? '1px solid #f0f0f0' : '1px solid #b7eb8f',
                }}
                actions={[
                  !item.read && (
                    <Button
                      type="link"
                      icon={<CheckOutlined />}
                      onClick={() => handleMarkAsRead(item.id)}
                    >
                      标记已读
                    </Button>
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
                    <Button type="text" icon={<MoreOutlined />} />
                  </Dropdown>,
                ].filter(Boolean)}
              >
                <List.Item.Meta
                  avatar={<Avatar icon={getIcon(item.type)} />}
                  title={
                    <Space>
                      <Text strong>{item.title}</Text>
                      <Tag color={getTypeColor(item.type)}>{item.type.toUpperCase()}</Tag>
                      {!item.read && <Badge status="processing" text="未读" />}
                    </Space>
                  }
                  description={
                    <Space direction="vertical" style={{ width: '100%' }}>
                      <Text>{item.content}</Text>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {dayjs(item.created_at).fromNow()}
                      </Text>
                      {item.action_url && (
                        <Button type="link" size="small" style={{ padding: 0 }}>
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

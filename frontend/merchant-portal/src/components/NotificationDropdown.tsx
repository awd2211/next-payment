import { Badge, Button, Dropdown, Empty, List, Space, Tag, Typography } from 'antd'
import {
  BellOutlined,
  CheckOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useNotificationStore, type Notification } from '../stores/notificationStore'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const { Text } = Typography

const NotificationItem = ({ notification }: { notification: Notification }) => {
  const { markAsRead, removeNotification } = useNotificationStore()

  const getIcon = () => {
    switch (notification.type) {
      case 'success':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />
      case 'warning':
        return <WarningOutlined style={{ color: '#faad14' }} />
      case 'error':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
      default:
        return <InfoCircleOutlined style={{ color: '#1890ff' }} />
    }
  }

  return (
    <List.Item
      style={{
        padding: '12px 16px',
        cursor: 'pointer',
        background: notification.read ? 'transparent' : 'rgba(24, 144, 255, 0.05)',
      }}
      onClick={() => !notification.read && markAsRead(notification.id)}
    >
      <Space direction="vertical" style={{ width: '100%' }} size={4}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            {getIcon()}
            <Text strong={!notification.read}>{notification.title}</Text>
          </Space>
          <Button
            type="text"
            size="small"
            icon={<DeleteOutlined />}
            onClick={(e) => {
              e.stopPropagation()
              removeNotification(notification.id)
            }}
          />
        </Space>
        <Text type="secondary" style={{ fontSize: 12 }}>
          {notification.message}
        </Text>
        <Text type="secondary" style={{ fontSize: 11 }}>
          {dayjs(notification.timestamp).fromNow()}
        </Text>
      </Space>
    </List.Item>
  )
}

const NotificationDropdown = () => {
  const { t } = useTranslation()
  const { notifications, unreadCount, markAllAsRead, clearAll } = useNotificationStore()

  const dropdownContent = (
    <div
      style={{
        width: 360,
        maxHeight: 480,
        background: 'var(--ant-color-bg-container)',
        borderRadius: 8,
        boxShadow: '0 3px 6px -4px rgba(0,0,0,0.12), 0 6px 16px 0 rgba(0,0,0,0.08)',
      }}
    >
      <div
        style={{
          padding: '12px 16px',
          borderBottom: '1px solid var(--ant-color-border)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <Space>
          <Text strong>{t('notifications.title', '通知')}</Text>
          {unreadCount > 0 && (
            <Tag color="blue">{t('notifications.unread', { count: unreadCount }, `${unreadCount} 条未读`)}</Tag>
          )}
        </Space>
        {notifications.length > 0 && (
          <Space size={4}>
            {unreadCount > 0 && (
              <Button
                type="text"
                size="small"
                icon={<CheckOutlined />}
                onClick={markAllAsRead}
              >
                {t('notifications.markAllRead', '全部已读')}
              </Button>
            )}
            <Button
              type="text"
              size="small"
              icon={<DeleteOutlined />}
              onClick={clearAll}
            >
              {t('notifications.clearAll', '清空')}
            </Button>
          </Space>
        )}
      </div>

      <div style={{ maxHeight: 400, overflowY: 'auto' }}>
        {notifications.length > 0 ? (
          <List
            dataSource={notifications}
            renderItem={(notification) => (
              <NotificationItem key={notification.id} notification={notification} />
            )}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={t('notifications.empty', '暂无通知')}
            style={{ padding: '40px 0' }}
          />
        )}
      </div>
    </div>
  )

  return (
    <Dropdown
      popupRender={() => dropdownContent}
      trigger={['click']}
      placement="bottomRight"
    >
      <Badge count={unreadCount} overflowCount={99}>
        <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
      </Badge>
    </Dropdown>
  )
}

export default NotificationDropdown

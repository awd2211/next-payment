import { useEffect } from 'react'
import { notification as antNotification } from 'antd'
import { useWebSocket, ConnectionStatus, type WebSocketMessage } from '../hooks/useWebSocket'
import { useNotificationStore } from '../stores/notificationStore'
import { useAuthStore } from '../stores/authStore'

interface WebSocketProviderProps {
  children: React.ReactNode
}

const WebSocketProvider = ({ children }: WebSocketProviderProps) => {
  const { token } = useAuthStore()
  const { addNotification } = useNotificationStore()

  const handleMessage = (message: WebSocketMessage) => {
    console.log('[WebSocket] Received message:', message)

    // Handle different message types
    switch (message.type) {
      case 'notification': {
        const data = message.data as {
          type: 'info' | 'success' | 'warning' | 'error'
          title: string
          message: string
        }
        addNotification({
          type: data.type || 'info',
          title: data.title,
          message: data.message,
          data: message.data,
        })

        // Show system notification if supported
        if ('Notification' in window && Notification.permission === 'granted') {
          new Notification(data.title, {
            body: data.message,
            icon: '/logo.png',
          })
        }
        break
      }

      case 'payment_update':
      case 'order_update':
      case 'merchant_update':
        // Add as notification
        addNotification({
          type: 'info',
          title: `${message.type.replace('_', ' ').toUpperCase()}`,
          message: JSON.stringify(message.data),
          data: message.data,
        })
        break

      default:
        console.log('[WebSocket] Unknown message type:', message.type)
    }
  }

  const handleConnect = () => {
    console.log('[WebSocket] Connected to server')
    // 静默连接，不打扰用户
    // antNotification.success({
    //   message: '连接成功',
    //   description: '实时通知已启用',
    //   duration: 2,
    // })
  }

  const handleDisconnect = () => {
    console.log('[WebSocket] Disconnected from server')
  }

  const handleError = (error: Event) => {
    console.error('[WebSocket] Connection error:', error)
    // 只在控制台记录错误，不打扰用户
    // WebSocket 会自动重连，不需要每次都提示用户
    // antNotification.error({
    //   message: '连接错误',
    //   description: '无法连接到通知服务器',
    //   duration: 3,
    // })
  }

  // Get WebSocket URL from environment or use default
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8007/ws'
  const wsUrlWithToken = token ? `${wsUrl}?token=${token}` : wsUrl

  const { status, connect, disconnect } = useWebSocket({
    url: wsUrlWithToken,
    reconnectInterval: 3000,
    maxReconnectAttempts: 5,
    heartbeatInterval: 30000,
    onMessage: handleMessage,
    onConnect: handleConnect,
    onDisconnect: handleDisconnect,
    onError: handleError,
  })

  useEffect(() => {
    // Request notification permission
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission()
    }

    // TODO: WebSocket 功能暂时禁用，等待 Kong 配置 WebSocket 路由
    // Connect when token is available
    // if (token) {
    //   connect()
    // }

    // return () => {
    //   disconnect()
    // }
  }, [token, connect, disconnect])

  useEffect(() => {
    // Show connection status - 只在错误时提示，连接成功时静默关闭之前的错误提示
    if (status === ConnectionStatus.ERROR) {
      // 延迟3秒后再显示错误提示，避免短暂断线时的干扰
      const timer = setTimeout(() => {
        antNotification.warning({
          message: '连接断开',
          description: '正在尝试重新连接...',
          duration: 0,
          key: 'ws-reconnecting',
        })
      }, 3000)
      
      return () => clearTimeout(timer)
    } else if (status === ConnectionStatus.CONNECTED) {
      // 静默关闭之前的错误提示
      antNotification.destroy('ws-reconnecting')
    }
  }, [status])

  return <>{children}</>
}

export default WebSocketProvider

import { useEffect, useRef, useState, useCallback } from 'react'

export interface WebSocketMessage {
  type: string
  data: unknown
  timestamp: string
}

export interface WebSocketConfig {
  url: string
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeatInterval?: number
  onMessage?: (message: WebSocketMessage) => void
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
}

export enum ConnectionStatus {
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  ERROR = 'error',
}

export const useWebSocket = (config: WebSocketConfig) => {
  const [status, setStatus] = useState<ConnectionStatus>(ConnectionStatus.DISCONNECTED)
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const heartbeatIntervalRef = useRef<NodeJS.Timeout>()

  const {
    url,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
    heartbeatInterval = 30000,
    onMessage,
    onConnect,
    onDisconnect,
    onError,
  } = config

  const clearTimers = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current)
    }
  }, [])

  const startHeartbeat = useCallback(() => {
    clearTimers()
    heartbeatIntervalRef.current = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, heartbeatInterval)
  }, [heartbeatInterval, clearTimers])

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    try {
      setStatus(ConnectionStatus.CONNECTING)
      wsRef.current = new WebSocket(url)

      wsRef.current.onopen = () => {
        console.log('[WebSocket] Connected')
        setStatus(ConnectionStatus.CONNECTED)
        reconnectAttemptsRef.current = 0
        startHeartbeat()
        onConnect?.()
      }

      wsRef.current.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WebSocketMessage

          // Ignore pong responses
          if (message.type === 'pong') {
            return
          }

          setLastMessage(message)
          onMessage?.(message)
        } catch (error) {
          console.error('[WebSocket] Failed to parse message:', error)
        }
      }

      wsRef.current.onerror = (error) => {
        console.error('[WebSocket] Error:', error)
        setStatus(ConnectionStatus.ERROR)
        onError?.(error)
      }

      wsRef.current.onclose = () => {
        console.log('[WebSocket] Disconnected')
        setStatus(ConnectionStatus.DISCONNECTED)
        clearTimers()
        onDisconnect?.()

        // Attempt to reconnect
        if (reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++
          console.log(
            `[WebSocket] Reconnecting... (${reconnectAttemptsRef.current}/${maxReconnectAttempts})`
          )
          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, reconnectInterval)
        } else {
          console.log('[WebSocket] Max reconnect attempts reached')
        }
      }
    } catch (error) {
      console.error('[WebSocket] Connection failed:', error)
      setStatus(ConnectionStatus.ERROR)
    }
  }, [
    url,
    reconnectInterval,
    maxReconnectAttempts,
    onMessage,
    onConnect,
    onDisconnect,
    onError,
    startHeartbeat,
    clearTimers,
  ])

  const disconnect = useCallback(() => {
    clearTimers()
    reconnectAttemptsRef.current = maxReconnectAttempts // Prevent auto-reconnect

    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }

    setStatus(ConnectionStatus.DISCONNECTED)
  }, [maxReconnectAttempts, clearTimers])

  const sendMessage = useCallback((message: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message))
      return true
    }
    console.warn('[WebSocket] Cannot send message: connection not open')
    return false
  }, [])

  useEffect(() => {
    return () => {
      disconnect()
    }
  }, [disconnect])

  return {
    status,
    lastMessage,
    connect,
    disconnect,
    sendMessage,
    isConnected: status === ConnectionStatus.CONNECTED,
  }
}

import { useState, useEffect } from 'react'

interface NetworkState {
  online: boolean
  effectiveType?: string // 网络类型: slow-2g, 2g, 3g, 4g
  downlink?: number // 下行速度 (Mbps)
  rtt?: number // 往返时间 (ms)
  saveData?: boolean // 是否开启省流量模式
}

/**
 * 网络状态Hook - 监控网络连接状态
 *
 * @example
 * const network = useNetwork()
 *
 * if (!network.online) {
 *   return <Alert message="网络已断开" type="error" />
 * }
 *
 * if (network.effectiveType === 'slow-2g') {
 *   // 提供低流量模式
 * }
 */
function useNetwork(): NetworkState {
  const [state, setState] = useState<NetworkState>(() => {
    if (typeof window === 'undefined' || !navigator) {
      return { online: true }
    }

    return {
      online: navigator.onLine,
      effectiveType: (navigator as any).connection?.effectiveType,
      downlink: (navigator as any).connection?.downlink,
      rtt: (navigator as any).connection?.rtt,
      saveData: (navigator as any).connection?.saveData,
    }
  })

  useEffect(() => {
    if (typeof window === 'undefined') return

    const updateNetworkState = () => {
      setState({
        online: navigator.onLine,
        effectiveType: (navigator as any).connection?.effectiveType,
        downlink: (navigator as any).connection?.downlink,
        rtt: (navigator as any).connection?.rtt,
        saveData: (navigator as any).connection?.saveData,
      })
    }

    // 监听网络状态变化
    window.addEventListener('online', updateNetworkState)
    window.addEventListener('offline', updateNetworkState)

    // 监听网络信息变化 (如果支持)
    const connection = (navigator as any).connection
    if (connection) {
      connection.addEventListener('change', updateNetworkState)
    }

    return () => {
      window.removeEventListener('online', updateNetworkState)
      window.removeEventListener('offline', updateNetworkState)
      if (connection) {
        connection.removeEventListener('change', updateNetworkState)
      }
    }
  }, [])

  return state
}

export default useNetwork

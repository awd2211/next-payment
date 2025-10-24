import { Alert } from 'antd'
import { WifiOutlined } from '@ant-design/icons'
import { useNetwork } from '../hooks'

/**
 * 网络状态提示组件
 *
 * @example
 * // 在Layout中使用
 * <NetworkStatus />
 */
const NetworkStatus = () => {
  const network = useNetwork()

  if (network.online) {
    return null // 在线时不显示
  }

  return (
    <Alert
      message="网络已断开"
      description="请检查您的网络连接,部分功能可能无法使用"
      type="error"
      icon={<WifiOutlined />}
      banner
      closable={false}
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        zIndex: 10000,
      }}
    />
  )
}

export default NetworkStatus

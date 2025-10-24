import { Tag } from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  StopOutlined,
} from '@ant-design/icons'

type StatusType = 'success' | 'pending' | 'failed' | 'processing' | 'cancelled' | 'refunded'

interface StatusTagProps {
  status: StatusType | string
  text?: string
}

const StatusTag = ({ status, text }: StatusTagProps) => {
  const getConfig = (status: string) => {
    const lowerStatus = status.toLowerCase()

    switch (lowerStatus) {
      case 'success':
      case 'completed':
      case 'approved':
        return {
          color: 'success',
          icon: <CheckCircleOutlined />,
          text: text || '成功',
        }

      case 'pending':
      case 'waiting':
        return {
          color: 'warning',
          icon: <ClockCircleOutlined />,
          text: text || '待处理',
        }

      case 'failed':
      case 'rejected':
        return {
          color: 'error',
          icon: <CloseCircleOutlined />,
          text: text || '失败',
        }

      case 'processing':
      case 'in_progress':
        return {
          color: 'processing',
          icon: <SyncOutlined spin />,
          text: text || '处理中',
        }

      case 'cancelled':
      case 'canceled':
        return {
          color: 'default',
          icon: <StopOutlined />,
          text: text || '已取消',
        }

      case 'refunded':
        return {
          color: 'magenta',
          icon: <CheckCircleOutlined />,
          text: text || '已退款',
        }

      default:
        return {
          color: 'default',
          icon: null,
          text: text || status,
        }
    }
  }

  const config = getConfig(status)

  return (
    <Tag color={config.color} icon={config.icon}>
      {config.text}
    </Tag>
  )
}

export default StatusTag

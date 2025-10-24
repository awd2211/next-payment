import { Empty, Button } from 'antd'
import type { ReactNode } from 'react'

interface EmptyStateProps {
  image?: ReactNode
  title?: string
  description?: string
  actionText?: string
  onAction?: () => void
  extra?: ReactNode
}

const EmptyState = ({
  image,
  title = '暂无数据',
  description,
  actionText,
  onAction,
  extra,
}: EmptyStateProps) => {
  return (
    <div style={{ padding: '40px 0', textAlign: 'center' }}>
      <Empty
        image={image || Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <div>
            <div style={{ fontSize: 16, marginBottom: 8 }}>{title}</div>
            {description && (
              <div style={{ fontSize: 14, color: '#999' }}>{description}</div>
            )}
          </div>
        }
      >
        {actionText && onAction && (
          <Button type="primary" onClick={onAction}>
            {actionText}
          </Button>
        )}
        {extra}
      </Empty>
    </div>
  )
}

export default EmptyState

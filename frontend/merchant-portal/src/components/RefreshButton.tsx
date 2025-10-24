import { Button, Tooltip } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import { useState, useEffect } from 'react'

interface RefreshButtonProps {
  onRefresh: () => Promise<void> | void
  autoRefresh?: boolean
  interval?: number // 自动刷新间隔(秒)
  tooltip?: string
}

const RefreshButton = ({
  onRefresh,
  autoRefresh = false,
  interval = 30,
  tooltip = '刷新',
}: RefreshButtonProps) => {
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState(interval)

  useEffect(() => {
    if (!autoRefresh) return

    const timer = setInterval(() => {
      setCountdown(prev => {
        if (prev <= 1) {
          handleRefresh()
          return interval
        }
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(timer)
  }, [autoRefresh, interval])

  const handleRefresh = async () => {
    setLoading(true)
    try {
      await onRefresh()
      setCountdown(interval)
    } finally {
      setLoading(false)
    }
  }

  const tooltipTitle = autoRefresh
    ? `${tooltip} (${countdown}秒后自动刷新)`
    : tooltip

  return (
    <Tooltip title={tooltipTitle}>
      <Button
        icon={<ReloadOutlined spin={loading} />}
        loading={loading}
        onClick={handleRefresh}
      >
        {autoRefresh && `${countdown}s`}
      </Button>
    </Tooltip>
  )
}

export default RefreshButton

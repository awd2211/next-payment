/**
 * 性能监控组件
 * 实时监控应用性能指标
 */
import { useState, useEffect } from 'react'
import { Card, Statistic, Row, Col, Progress, Badge, Tooltip, Space } from 'antd'
import {
  ThunderboltOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  WifiOutlined
} from '@ant-design/icons'

interface PerformanceMetrics {
  fps: number
  memory: number
  loadTime: number
  apiLatency: number
}

const PerformanceMonitor: React.FC<{ enabled?: boolean }> = ({ enabled = false }) => {
  const [metrics, setMetrics] = useState<PerformanceMetrics>({
    fps: 60,
    memory: 0,
    loadTime: 0,
    apiLatency: 0,
  })
  const [visible, setVisible] = useState(enabled)

  useEffect(() => {
    if (!visible) return

    // FPS 监控
    let lastTime = performance.now()
    let frames = 0

    const measureFPS = () => {
      frames++
      const currentTime = performance.now()

      if (currentTime >= lastTime + 1000) {
        setMetrics((prev) => ({ ...prev, fps: frames }))
        frames = 0
        lastTime = currentTime
      }

      if (visible) {
        requestAnimationFrame(measureFPS)
      }
    }

    requestAnimationFrame(measureFPS)

    // 内存监控
    const measureMemory = () => {
      if ('memory' in performance) {
        const memory = (performance as any).memory
        const usedMB = memory.usedJSHeapSize / 1024 / 1024
        setMetrics((prev) => ({ ...prev, memory: Math.round(usedMB) }))
      }
    }

    const memoryInterval = setInterval(measureMemory, 2000)

    // 页面加载时间
    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
    if (navigation) {
      const loadTime = navigation.loadEventEnd - navigation.fetchStart
      setMetrics((prev) => ({ ...prev, loadTime: Math.round(loadTime) }))
    }

    return () => {
      clearInterval(memoryInterval)
    }
  }, [visible])

  // 切换显示 (Ctrl + Shift + P)
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.shiftKey && e.key === 'P') {
        e.preventDefault()
        setVisible((v) => !v)
      }
    }

    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [])

  if (!visible) return null

  const getFPSColor = () => {
    if (metrics.fps >= 55) return '#52c41a'
    if (metrics.fps >= 30) return '#faad14'
    return '#ff4d4f'
  }

  const getMemoryColor = () => {
    if (metrics.memory < 50) return '#52c41a'
    if (metrics.memory < 100) return '#faad14'
    return '#ff4d4f'
  }

  return (
    <div
      style={{
        position: 'fixed',
        bottom: 16,
        right: 16,
        zIndex: 9999,
        width: 320,
      }}
    >
      <Card
        size="small"
        title={
          <Space>
            <Badge status="processing" />
            性能监控
            <Tooltip title="按 Ctrl+Shift+P 隐藏">
              <ThunderboltOutlined style={{ fontSize: 12, color: '#999' }} />
            </Tooltip>
          </Space>
        }
        style={{ boxShadow: '0 4px 12px rgba(0,0,0,0.15)' }}
      >
        <Row gutter={[8, 8]}>
          <Col span={12}>
            <Statistic
              title="FPS"
              value={metrics.fps}
              suffix="帧/秒"
              valueStyle={{ color: getFPSColor(), fontSize: 18 }}
              prefix={<ThunderboltOutlined />}
            />
            <Progress
              percent={Math.min(100, (metrics.fps / 60) * 100)}
              strokeColor={getFPSColor()}
              showInfo={false}
              size="small"
            />
          </Col>
          <Col span={12}>
            <Statistic
              title="内存"
              value={metrics.memory}
              suffix="MB"
              valueStyle={{ color: getMemoryColor(), fontSize: 18 }}
              prefix={<DatabaseOutlined />}
            />
            <Progress
              percent={Math.min(100, (metrics.memory / 200) * 100)}
              strokeColor={getMemoryColor()}
              showInfo={false}
              size="small"
            />
          </Col>
          <Col span={12}>
            <Statistic
              title="加载时间"
              value={metrics.loadTime}
              suffix="ms"
              valueStyle={{ fontSize: 16 }}
              prefix={<ClockCircleOutlined />}
            />
          </Col>
          <Col span={12}>
            <Statistic
              title="API 延迟"
              value={metrics.apiLatency}
              suffix="ms"
              valueStyle={{ fontSize: 16 }}
              prefix={<WifiOutlined />}
            />
          </Col>
        </Row>
      </Card>
    </div>
  )
}

export default PerformanceMonitor

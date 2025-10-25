// @ts-nocheck
import { useEffect, useRef } from 'react'

/**
 * 性能监控 Hook
 * 监控组件渲染性能
 */
export function usePerformance(componentName: string, enabled = process.env.NODE_ENV === 'development') {
  const renderCount = useRef(0)
  const renderTimes = useRef<number[]>([])
  const startTime = useRef<number>(0)

  useEffect(() => {
    if (!enabled) return

    renderCount.current += 1
    const endTime = performance.now()
    const renderTime = endTime - startTime.current

    renderTimes.current.push(renderTime)

    // 保留最近 10 次渲染时间
    if (renderTimes.current.length > 10) {
      renderTimes.current.shift()
    }

    // 计算平均渲染时间
    const avgRenderTime =
      renderTimes.current.reduce((sum, time) => sum + time, 0) / renderTimes.current.length

    console.log(`[Performance] ${componentName}`, {
      renderCount: renderCount.current,
      lastRenderTime: `${renderTime.toFixed(2)}ms`,
      avgRenderTime: `${avgRenderTime.toFixed(2)}ms`,
    })

    // 如果渲染时间超过 16ms (60fps)，发出警告
    if (renderTime > 16) {
      console.warn(
        `[Performance Warning] ${componentName} render time (${renderTime.toFixed(2)}ms) exceeds 16ms`,
      )
    }
  })

  // 记录渲染开始时间
  startTime.current = performance.now()

  return {
    renderCount: renderCount.current,
    avgRenderTime:
      renderTimes.current.reduce((sum, time) => sum + time, 0) / renderTimes.current.length || 0,
  }
}

/**
 * 监控 API 请求性能
 */
export function useAPIPerformance() {
  const apiCalls = useRef<Map<string, { count: number; totalTime: number }>>(new Map())

  const trackAPI = (apiName: string, duration: number) => {
    const current = apiCalls.current.get(apiName) || { count: 0, totalTime: 0 }
    apiCalls.current.set(apiName, {
      count: current.count + 1,
      totalTime: current.totalTime + duration,
    })

    const avgTime = (current.totalTime + duration) / (current.count + 1)

    if (process.env.NODE_ENV === 'development') {
      console.log(`[API Performance] ${apiName}`, {
        duration: `${duration.toFixed(2)}ms`,
        avgTime: `${avgTime.toFixed(2)}ms`,
        calls: current.count + 1,
      })
    }

    // 如果 API 调用超过 3 秒，发出警告
    if (duration > 3000) {
      console.warn(`[API Warning] ${apiName} took ${duration.toFixed(2)}ms`)
    }
  }

  const getStats = () => {
    const stats: Record<string, { count: number; avgTime: number }> = {}
    apiCalls.current.forEach((value, key) => {
      stats[key] = {
        count: value.count,
        avgTime: value.totalTime / value.count,
      }
    })
    return stats
  }

  return { trackAPI, getStats }
}

/**
 * Web Vitals 监控
 * 监控核心 Web 指标（LCP, FID, CLS）
 */
export function useWebVitals(callback?: (metric: any) => void) {
  useEffect(() => {
    if (typeof window === 'undefined') return

    // 动态导入 web-vitals
    import('web-vitals').then(({ onCLS, onFID, onLCP, onFCP, onTTFB }) => {
      const reportMetric = (metric: any) => {
        console.log(`[Web Vitals] ${metric.name}:`, metric.value)
        callback?.(metric)

        // 可以上报到监控服务
        if (process.env.NODE_ENV === 'production') {
          // navigator.sendBeacon('/api/v1/metrics', JSON.stringify(metric))
        }
      }

      onCLS(reportMetric)
      onFID(reportMetric)
      onLCP(reportMetric)
      onFCP(reportMetric)
      onTTFB(reportMetric)
    })
  }, [callback])
}

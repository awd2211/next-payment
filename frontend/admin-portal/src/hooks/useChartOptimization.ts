/**
 * 图表性能优化 Hook
 * 优化 Ant Design Charts 的渲染性能
 */
import { useState, useEffect, useMemo, useRef, useCallback } from 'react'
import { useDebounceFn, useThrottle } from './useDebounce'

export interface ChartData {
  [key: string]: any
}

/**
 * 图表数据防抖 Hook
 * 避免频繁更新图表
 */
export function useChartDebounce<T extends ChartData[]>(
  data: T,
  delay: number = 300
): T {
  const [debouncedData, setDebouncedData] = useState<T>(data)

  const [updateData] = useDebounceFn((newData: T) => {
    setDebouncedData(newData)
  }, delay)

  useEffect(() => {
    updateData(data)
  }, [data, updateData])

  return debouncedData
}

/**
 * 图表懒加载 Hook
 * 仅在图表进入视口时才渲染
 */
export function useChartLazyLoad(
  ref: React.RefObject<HTMLElement>,
  options: IntersectionObserverInit = {}
): boolean {
  const [isVisible, setIsVisible] = useState(false)
  const [hasLoaded, setHasLoaded] = useState(false)

  useEffect(() => {
    const element = ref.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !hasLoaded) {
          setIsVisible(true)
          setHasLoaded(true)
          observer.disconnect()
        }
      },
      {
        threshold: 0.1,
        ...options,
      }
    )

    observer.observe(element)

    return () => {
      observer.disconnect()
    }
  }, [ref, hasLoaded, options])

  return isVisible
}

/**
 * 图表数据采样 Hook
 * 大数据量时进行采样,减少渲染负担
 */
export function useChartSampling<T extends ChartData[]>(
  data: T,
  maxPoints: number = 1000
): T {
  return useMemo(() => {
    if (data.length <= maxPoints) {
      return data
    }

    // 等间隔采样
    const step = Math.ceil(data.length / maxPoints)
    return data.filter((_, index) => index % step === 0) as T
  }, [data, maxPoints])
}

/**
 * 图表窗口化 Hook
 * 仅显示部分数据,支持滚动查看
 */
export function useChartWindowing<T extends ChartData[]>(
  data: T,
  windowSize: number = 100,
  initialOffset: number = 0
): {
  windowedData: T
  offset: number
  setOffset: (offset: number) => void
  canScrollLeft: boolean
  canScrollRight: boolean
  scrollLeft: () => void
  scrollRight: () => void
} {
  const [offset, setOffset] = useState(initialOffset)

  const windowedData = useMemo(() => {
    return data.slice(offset, offset + windowSize) as T
  }, [data, offset, windowSize])

  const canScrollLeft = offset > 0
  const canScrollRight = offset + windowSize < data.length

  const scrollLeft = useCallback(() => {
    setOffset((prev) => Math.max(0, prev - windowSize))
  }, [windowSize])

  const scrollRight = useCallback(() => {
    setOffset((prev) => Math.min(data.length - windowSize, prev + windowSize))
  }, [data.length, windowSize])

  return {
    windowedData,
    offset,
    setOffset,
    canScrollLeft,
    canScrollRight,
    scrollLeft,
    scrollRight,
  }
}

/**
 * 图表自适应大小 Hook
 * 监听容器大小变化,自动调整图表
 */
export function useChartResize(
  containerRef: React.RefObject<HTMLElement>
): {
  width: number
  height: number
} {
  const [size, setSize] = useState({ width: 0, height: 0 })

  useEffect(() => {
    const element = containerRef.current
    if (!element) return

    const updateSize = () => {
      setSize({
        width: element.clientWidth,
        height: element.clientHeight,
      })
    }

    // 初始化大小
    updateSize()

    // 监听窗口大小变化
    const resizeObserver = new ResizeObserver(updateSize)
    resizeObserver.observe(element)

    return () => {
      resizeObserver.disconnect()
    }
  }, [containerRef])

  return size
}

/**
 * 图表交互节流 Hook
 * 优化图表交互事件(如 tooltip、legend 点击)
 */
export function useChartInteractionThrottle<T extends (...args: any[]) => any>(
  handler: T,
  delay: number = 100
): T {
  const [throttledHandler] = useThrottle(handler, delay)
  return throttledHandler as T
}

/**
 * 图表数据缓存 Hook
 * 缓存计算结果,避免重复计算
 */
export function useChartMemoization<T>(
  computeFn: () => T,
  deps: React.DependencyList
): T {
  return useMemo(computeFn, deps)
}

/**
 * 图表渲染优化 Hook
 * 综合多种优化策略
 */
export function useOptimizedChart<T extends ChartData[]>(
  data: T,
  options: {
    debounceDelay?: number
    maxPoints?: number
    enableSampling?: boolean
    enableDebounce?: boolean
  } = {}
): T {
  const {
    debounceDelay = 300,
    maxPoints = 1000,
    enableSampling = true,
    enableDebounce = true,
  } = options

  // 采样
  const sampledData = useMemo(() => {
    if (!enableSampling) return data
    return data.length > maxPoints
      ? (data.filter((_, index) => index % Math.ceil(data.length / maxPoints) === 0) as T)
      : data
  }, [data, maxPoints, enableSampling])

  // 防抖
  const [debouncedData, setDebouncedData] = useState(sampledData)
  const [updateData] = useDebounceFn((newData: T) => {
    setDebouncedData(newData)
  }, debounceDelay)

  useEffect(() => {
    if (enableDebounce) {
      updateData(sampledData)
    } else {
      setDebouncedData(sampledData)
    }
  }, [sampledData, enableDebounce, updateData])

  return debouncedData
}

/**
 * 图表加载状态 Hook
 * 管理图表加载、错误、空状态
 */
export function useChartState<T>(
  initialData: T | null = null
): {
  data: T | null
  loading: boolean
  error: Error | null
  setData: (data: T) => void
  setLoading: (loading: boolean) => void
  setError: (error: Error | null) => void
  isEmpty: boolean
} {
  const [data, setData] = useState<T | null>(initialData)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const isEmpty = useMemo(() => {
    if (!data) return true
    if (Array.isArray(data)) return data.length === 0
    if (typeof data === 'object') return Object.keys(data).length === 0
    return false
  }, [data])

  return {
    data,
    loading,
    error,
    setData,
    setLoading,
    setError,
    isEmpty,
  }
}

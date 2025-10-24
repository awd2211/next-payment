import { useState, useEffect, useRef } from 'react'

/**
 * 节流 Hook - 性能优化
 * 用于滚动事件、按钮点击等需要限制频率的场景
 *
 * @param value 需要节流的值
 * @param interval 时间间隔(毫秒)
 * @returns 节流后的值
 *
 * @example
 * const [scrollY, setScrollY] = useState(0)
 * const throttledScrollY = useThrottle(scrollY, 200)
 *
 * useEffect(() => {
 *   // 每200ms最多执行一次
 *   console.log('Scroll position:', throttledScrollY)
 * }, [throttledScrollY])
 */
function useThrottle<T>(value: T, interval: number = 200): T {
  const [throttledValue, setThrottledValue] = useState<T>(value)
  const lastExecuted = useRef<number>(Date.now())

  useEffect(() => {
    const now = Date.now()
    const timeSinceLastExecution = now - lastExecuted.current

    if (timeSinceLastExecution >= interval) {
      lastExecuted.current = now
      setThrottledValue(value)
    } else {
      const timer = setTimeout(() => {
        lastExecuted.current = Date.now()
        setThrottledValue(value)
      }, interval - timeSinceLastExecution)

      return () => clearTimeout(timer)
    }
  }, [value, interval])

  return throttledValue
}

export default useThrottle

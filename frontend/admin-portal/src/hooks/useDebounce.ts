import { useEffect, useRef, useState, useCallback } from 'react'

/**
 * 防抖 Hook - 值防抖
 * 用于输入框搜索等场景
 */
export function useDebounce<T>(value: T, delay: number = 500): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

/**
 * 防抖 Hook - 函数防抖
 * 用于频繁触发的函数调用
 */
export function useDebounceFn<T extends (...args: any[]) => any>(
  fn: T,
  delay: number = 500
): [(...args: Parameters<T>) => void, () => void] {
  const timeoutRef = useRef<NodeJS.Timeout>()

  const debouncedFn = useCallback(
    (...args: Parameters<T>) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }

      timeoutRef.current = setTimeout(() => {
        fn(...args)
      }, delay)
    },
    [fn, delay]
  )

  const cancel = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
  }, [])

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  return [debouncedFn, cancel]
}

/**
 * 节流 Hook
 * 限制函数调用频率
 */
export function useThrottle<T extends (...args: any[]) => any>(
  fn: T,
  delay: number = 500
): [(...args: Parameters<T>) => void, () => void] {
  const timeoutRef = useRef<NodeJS.Timeout>()
  const lastRunRef = useRef<number>(0)

  const throttledFn = useCallback(
    (...args: Parameters<T>) => {
      const now = Date.now()

      if (now - lastRunRef.current >= delay) {
        fn(...args)
        lastRunRef.current = now
      } else {
        if (timeoutRef.current) {
          clearTimeout(timeoutRef.current)
        }

        timeoutRef.current = setTimeout(() => {
          fn(...args)
          lastRunRef.current = Date.now()
        }, delay - (now - lastRunRef.current))
      }
    },
    [fn, delay]
  )

  const cancel = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
  }, [])

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  return [throttledFn, cancel]
}






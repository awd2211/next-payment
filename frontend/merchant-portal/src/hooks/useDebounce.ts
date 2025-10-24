import { useState, useEffect } from 'react'

/**
 * 防抖 Hook - 性能优化
 * 用于搜索输入、窗口resize等高频事件
 *
 * @param value 需要防抖的值
 * @param delay 延迟时间(毫秒)
 * @returns 防抖后的值
 *
 * @example
 * const [searchTerm, setSearchTerm] = useState('')
 * const debouncedSearchTerm = useDebounce(searchTerm, 500)
 *
 * useEffect(() => {
 *   // 只在用户停止输入500ms后才搜索
 *   if (debouncedSearchTerm) {
 *     searchAPI(debouncedSearchTerm)
 *   }
 * }, [debouncedSearchTerm])
 */
function useDebounce<T>(value: T, delay: number = 500): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    // 设置定时器
    const timer = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    // 清理函数:在value变化或组件卸载时清除定时器
    return () => {
      clearTimeout(timer)
    }
  }, [value, delay])

  return debouncedValue
}

export default useDebounce

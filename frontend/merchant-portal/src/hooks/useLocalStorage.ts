import { useState, useEffect, useCallback } from 'react'

/**
 * LocalStorage Hook - 数据持久化
 * 自动序列化/反序列化,支持TypeScript类型
 *
 * @param key localStorage的key
 * @param initialValue 初始值
 * @returns [value, setValue, removeValue]
 *
 * @example
 * const [user, setUser, removeUser] = useLocalStorage('user', null)
 *
 * // 保存数据
 * setUser({ id: 1, name: 'John' })
 *
 * // 删除数据
 * removeUser()
 */
function useLocalStorage<T>(
  key: string,
  initialValue: T
): [T, (value: T | ((val: T) => T)) => void, () => void] {
  // 获取初始值
  const readValue = useCallback((): T => {
    // SSR环境检查
    if (typeof window === 'undefined') {
      return initialValue
    }

    try {
      const item = window.localStorage.getItem(key)
      return item ? (JSON.parse(item) as T) : initialValue
    } catch (error) {
      console.warn(`Error reading localStorage key "${key}":`, error)
      return initialValue
    }
  }, [initialValue, key])

  const [storedValue, setStoredValue] = useState<T>(readValue)

  // 保存到localStorage
  const setValue = useCallback(
    (value: T | ((val: T) => T)) => {
      // SSR环境检查
      if (typeof window === 'undefined') {
        console.warn(`localStorage is not available in this environment`)
        return
      }

      try {
        // 支持函数式更新
        const newValue = value instanceof Function ? value(storedValue) : value

        // 保存到state
        setStoredValue(newValue)

        // 保存到localStorage
        window.localStorage.setItem(key, JSON.stringify(newValue))

        // 触发自定义事件,通知其他tabs/windows
        window.dispatchEvent(
          new CustomEvent('local-storage', {
            detail: { key, newValue },
          })
        )
      } catch (error) {
        console.warn(`Error setting localStorage key "${key}":`, error)
      }
    },
    [key, storedValue]
  )

  // 删除localStorage
  const removeValue = useCallback(() => {
    // SSR环境检查
    if (typeof window === 'undefined') {
      return
    }

    try {
      window.localStorage.removeItem(key)
      setStoredValue(initialValue)

      // 触发自定义事件
      window.dispatchEvent(
        new CustomEvent('local-storage', {
          detail: { key, newValue: null },
        })
      )
    } catch (error) {
      console.warn(`Error removing localStorage key "${key}":`, error)
    }
  }, [key, initialValue])

  // 监听storage事件(跨tab同步)
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent | CustomEvent) => {
      if ('key' in e && e.key !== key) {
        return
      }

      const detail = 'detail' in e ? e.detail : { key: e.key, newValue: e.newValue }

      if (detail.key === key) {
        try {
          const newValue = detail.newValue ? JSON.parse(detail.newValue) : initialValue
          setStoredValue(newValue)
        } catch (error) {
          console.warn(`Error syncing localStorage key "${key}":`, error)
        }
      }
    }

    // 监听storage事件和自定义事件
    window.addEventListener('storage', handleStorageChange as EventListener)
    window.addEventListener('local-storage', handleStorageChange as EventListener)

    return () => {
      window.removeEventListener('storage', handleStorageChange as EventListener)
      window.removeEventListener('local-storage', handleStorageChange as EventListener)
    }
  }, [key, initialValue])

  return [storedValue, setValue, removeValue]
}

export default useLocalStorage

// @ts-nocheck
/**
 * 缓存策略 Hook
 * 支持内存缓存、LocalStorage 缓存、SessionStorage 缓存
 */
import { useState, useCallback, useEffect, useRef } from 'react'

export type CacheStrategy = 'memory' | 'localStorage' | 'sessionStorage'

export interface CacheOptions {
  /**
   * 缓存策略
   */
  strategy?: CacheStrategy

  /**
   * 缓存过期时间 (毫秒)
   */
  ttl?: number

  /**
   * 缓存键前缀
   */
  prefix?: string

  /**
   * 是否在窗口关闭时清除缓存
   */
  clearOnUnload?: boolean
}

interface CacheItem<T> {
  value: T
  timestamp: number
  ttl?: number
}

/**
 * 内存缓存
 */
class MemoryCache {
  private cache = new Map<string, CacheItem<any>>()

  set<T>(key: string, value: T, ttl?: number): void {
    this.cache.set(key, {
      value,
      timestamp: Date.now(),
      ttl,
    })
  }

  get<T>(key: string): T | null {
    const item = this.cache.get(key)

    if (!item) {
      return null
    }

    // 检查是否过期
    if (item.ttl && Date.now() - item.timestamp > item.ttl) {
      this.cache.delete(key)
      return null
    }

    return item.value
  }

  has(key: string): boolean {
    return this.cache.has(key)
  }

  delete(key: string): void {
    this.cache.delete(key)
  }

  clear(): void {
    this.cache.clear()
  }

  size(): number {
    return this.cache.size
  }
}

/**
 * Storage 缓存 (LocalStorage/SessionStorage)
 */
class StorageCache {
  constructor(private storage: Storage) {}

  set<T>(key: string, value: T, ttl?: number): void {
    const item: CacheItem<T> = {
      value,
      timestamp: Date.now(),
      ttl,
    }

    try {
      this.storage.setItem(key, JSON.stringify(item))
    } catch (error) {
      console.error('Storage cache set error:', error)
    }
  }

  get<T>(key: string): T | null {
    try {
      const data = this.storage.getItem(key)

      if (!data) {
        return null
      }

      const item: CacheItem<T> = JSON.parse(data)

      // 检查是否过期
      if (item.ttl && Date.now() - item.timestamp > item.ttl) {
        this.storage.removeItem(key)
        return null
      }

      return item.value
    } catch (error) {
      console.error('Storage cache get error:', error)
      return null
    }
  }

  has(key: string): boolean {
    return this.storage.getItem(key) !== null
  }

  delete(key: string): void {
    this.storage.removeItem(key)
  }

  clear(): void {
    this.storage.clear()
  }

  size(): number {
    return this.storage.length
  }
}

// 单例缓存实例
const memoryCache = new MemoryCache()
const localStorageCache = new StorageCache(localStorage)
const sessionStorageCache = new StorageCache(sessionStorage)

/**
 * 缓存 Hook
 */
export function useCache<T = any>(
  key: string,
  options: CacheOptions = {}
): {
  value: T | null
  setValue: (value: T) => void
  clearValue: () => void
  hasValue: boolean
  refresh: () => void
} {
  const {
    strategy = 'memory',
    ttl,
    prefix = 'cache',
    clearOnUnload = false,
  } = options

  const cacheKey = `${prefix}:${key}`

  // 获取缓存实例
  const getCache = useCallback(() => {
    switch (strategy) {
      case 'localStorage':
        return localStorageCache
      case 'sessionStorage':
        return sessionStorageCache
      default:
        return memoryCache
    }
  }, [strategy])

  const cache = getCache()

  // 初始化值
  const [value, setValueState] = useState<T | null>(() => cache.get<T>(cacheKey))
  const [, setRefreshTrigger] = useState(0)

  // 设置缓存
  const setValue = useCallback(
    (newValue: T) => {
      cache.set(cacheKey, newValue, ttl)
      setValueState(newValue)
    },
    [cache, cacheKey, ttl]
  )

  // 清除缓存
  const clearValue = useCallback(() => {
    cache.delete(cacheKey)
    setValueState(null)
  }, [cache, cacheKey])

  // 检查是否有缓存
  const hasValue = cache.has(cacheKey)

  // 刷新缓存
  const refresh = useCallback(() => {
    const cachedValue = cache.get<T>(cacheKey)
    setValueState(cachedValue)
    setRefreshTrigger((prev) => prev + 1)
  }, [cache, cacheKey])

  // 窗口卸载时清除缓存
  useEffect(() => {
    if (clearOnUnload) {
      const handleUnload = () => {
        cache.delete(cacheKey)
      }

      window.addEventListener('beforeunload', handleUnload)
      return () => window.removeEventListener('beforeunload', handleUnload)
    }
  }, [clearOnUnload, cache, cacheKey])

  return {
    value,
    setValue,
    clearValue,
    hasValue,
    refresh,
  }
}

/**
 * 函数结果缓存 Hook
 */
export function useMemoCache<T = any, Args extends any[] = any[]>(
  fn: (...args: Args) => T,
  options: CacheOptions & {
    /**
     * 生成缓存 key 的函数
     */
    keyGenerator?: (...args: Args) => string
  } = {}
): (...args: Args) => T {
  const { keyGenerator, ...cacheOptions } = options
  const cacheRef = useRef(new Map<string, T>())

  return useCallback(
    (...args: Args): T => {
      const key = keyGenerator ? keyGenerator(...args) : JSON.stringify(args)

      if (cacheRef.current.has(key)) {
        return cacheRef.current.get(key)!
      }

      const result = fn(...args)
      cacheRef.current.set(key, result)

      return result
    },
    [fn, keyGenerator]
  )
}

/**
 * 异步函数缓存 Hook
 */
export function useAsyncCache<T = any>(
  key: string,
  fetcher: () => Promise<T>,
  options: CacheOptions & {
    /**
     * 是否自动获取
     */
    autoFetch?: boolean

    /**
     * 重试次数
     */
    retries?: number
  } = {}
): {
  data: T | null
  loading: boolean
  error: Error | null
  fetch: () => Promise<void>
  clearCache: () => void
} {
  const { autoFetch = true, retries = 0, ...cacheOptions } = options

  const cache = useCache<T>(key, cacheOptions)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const isMountedRef = useRef(true)

  const fetchData = useCallback(async () => {
    // 如果有缓存,直接返回
    if (cache.hasValue && cache.value) {
      return
    }

    setLoading(true)
    setError(null)

    let attempt = 0
    while (attempt <= retries) {
      try {
        const data = await fetcher()

        if (isMountedRef.current) {
          cache.setValue(data)
          setLoading(false)
        }

        return
      } catch (err) {
        attempt++

        if (attempt > retries) {
          if (isMountedRef.current) {
            setError(err as Error)
            setLoading(false)
          }
        }
      }
    }
  }, [fetcher, cache, retries])

  useEffect(() => {
    if (autoFetch) {
      fetchData()
    }

    return () => {
      isMountedRef.current = false
    }
  }, [autoFetch, fetchData])

  return {
    data: cache.value,
    loading,
    error,
    fetch: fetchData,
    clearCache: cache.clearValue,
  }
}

/**
 * LRU 缓存 Hook
 */
export function useLRUCache<K = string, V = any>(maxSize: number = 100) {
  const cacheRef = useRef<Map<K, V>>(new Map())

  const get = useCallback((key: K): V | undefined => {
    const cache = cacheRef.current

    if (!cache.has(key)) {
      return undefined
    }

    // LRU: 将访问的项移到最后
    const value = cache.get(key)!
    cache.delete(key)
    cache.set(key, value)

    return value
  }, [])

  const set = useCallback(
    (key: K, value: V): void => {
      const cache = cacheRef.current

      // 如果已存在,先删除
      if (cache.has(key)) {
        cache.delete(key)
      }

      // 添加到末尾
      cache.set(key, value)

      // 检查大小限制
      if (cache.size > maxSize) {
        // 删除最早的项 (第一个)
        const firstKey = cache.keys().next().value
        cache.delete(firstKey)
      }
    },
    [maxSize]
  )

  const clear = useCallback(() => {
    cacheRef.current.clear()
  }, [])

  const has = useCallback((key: K): boolean => {
    return cacheRef.current.has(key)
  }, [])

  const size = useCallback((): number => {
    return cacheRef.current.size
  }, [])

  return {
    get,
    set,
    clear,
    has,
    size,
  }
}

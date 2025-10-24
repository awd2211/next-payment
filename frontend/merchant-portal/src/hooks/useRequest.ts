import { useState, useCallback, useRef, useEffect } from 'react'

/**
 * 请求配置
 */
interface RequestConfig<T> {
  manual?: boolean // 是否手动触发
  defaultParams?: any[] // 默认参数
  onSuccess?: (data: T, params: any[]) => void
  onError?: (error: Error, params: any[]) => void
  onFinally?: (params: any[]) => void
  retryCount?: number // 重试次数
  retryInterval?: number // 重试间隔(ms)
  debounceWait?: number // 防抖延迟(ms)
  throttleWait?: number // 节流延迟(ms)
  cacheKey?: string // 缓存key
  cacheTime?: number // 缓存时间(ms)
}

/**
 * 请求状态
 */
interface RequestState<T> {
  data: T | undefined
  error: Error | undefined
  loading: boolean
}

/**
 * 请求操作
 */
interface RequestActions<T> {
  run: (...params: any[]) => Promise<T | undefined>
  refresh: () => Promise<T | undefined>
  cancel: () => void
  mutate: (data: T | ((oldData: T | undefined) => T)) => void
}

/**
 * 缓存存储
 */
const cache = new Map<string, { data: any; timestamp: number }>()

/**
 * 请求Hook - 简化API调用
 *
 * @example
 * const [state, actions] = useRequest(
 *   async (id: string) => {
 *     return await api.get(`/merchant/${id}`)
 *   },
 *   {
 *     manual: false, // 自动执行
 *     defaultParams: ['123'],
 *     onSuccess: (data) => {
 *       message.success('加载成功')
 *     },
 *     retryCount: 3,
 *     cacheKey: 'merchant-detail',
 *     cacheTime: 60000
 *   }
 * )
 */
function useRequest<T>(
  service: (...params: any[]) => Promise<T>,
  config: RequestConfig<T> = {}
): [RequestState<T>, RequestActions<T>] {
  const [data, setData] = useState<T | undefined>(undefined)
  const [error, setError] = useState<Error | undefined>(undefined)
  const [loading, setLoading] = useState<boolean>(!config.manual)

  const debounceTimerRef = useRef<NodeJS.Timeout>()
  const throttleTimerRef = useRef<NodeJS.Timeout>()
  const lastCallTimeRef = useRef<number>(0)
  const canceledRef = useRef(false)
  const latestParamsRef = useRef<any[]>(config.defaultParams || [])

  /**
   * 获取缓存数据
   */
  const getCachedData = useCallback((): T | undefined => {
    if (!config.cacheKey) return undefined

    const cached = cache.get(config.cacheKey)
    if (!cached) return undefined

    const isExpired =
      config.cacheTime && Date.now() - cached.timestamp > config.cacheTime
    if (isExpired) {
      cache.delete(config.cacheKey)
      return undefined
    }

    return cached.data
  }, [config.cacheKey, config.cacheTime])

  /**
   * 设置缓存数据
   */
  const setCachedData = useCallback(
    (data: T) => {
      if (config.cacheKey) {
        cache.set(config.cacheKey, { data, timestamp: Date.now() })
      }
    },
    [config.cacheKey]
  )

  /**
   * 执行请求(带重试)
   */
  const executeRequest = useCallback(
    async (params: any[], retryCount = 0): Promise<T | undefined> => {
      if (canceledRef.current) return undefined

      // 检查缓存
      const cachedData = getCachedData()
      if (cachedData) {
        setData(cachedData)
        setLoading(false)
        return cachedData
      }

      setLoading(true)
      setError(undefined)

      try {
        const result = await service(...params)

        if (canceledRef.current) return undefined

        setData(result)
        setCachedData(result)
        config.onSuccess?.(result, params)

        return result
      } catch (err) {
        if (canceledRef.current) return undefined

        const error = err instanceof Error ? err : new Error(String(err))

        // 重试逻辑
        const shouldRetry =
          config.retryCount && retryCount < config.retryCount
        if (shouldRetry) {
          await new Promise((resolve) =>
            setTimeout(resolve, config.retryInterval || 1000)
          )
          return executeRequest(params, retryCount + 1)
        }

        setError(error)
        config.onError?.(error, params)
        return undefined
      } finally {
        if (!canceledRef.current) {
          setLoading(false)
          config.onFinally?.(params)
        }
      }
    },
    [service, config, getCachedData, setCachedData]
  )

  /**
   * 运行请求(支持防抖和节流)
   */
  const run = useCallback(
    (...params: any[]): Promise<T | undefined> => {
      latestParamsRef.current = params

      return new Promise((resolve) => {
        // 防抖处理
        if (config.debounceWait) {
          if (debounceTimerRef.current) {
            clearTimeout(debounceTimerRef.current)
          }
          debounceTimerRef.current = setTimeout(() => {
            executeRequest(params).then(resolve)
          }, config.debounceWait)
          return
        }

        // 节流处理
        if (config.throttleWait) {
          const now = Date.now()
          const timeSinceLastCall = now - lastCallTimeRef.current

          if (timeSinceLastCall >= config.throttleWait) {
            lastCallTimeRef.current = now
            executeRequest(params).then(resolve)
          } else {
            if (throttleTimerRef.current) {
              clearTimeout(throttleTimerRef.current)
            }
            throttleTimerRef.current = setTimeout(() => {
              lastCallTimeRef.current = Date.now()
              executeRequest(params).then(resolve)
            }, config.throttleWait - timeSinceLastCall)
          }
          return
        }

        // 直接执行
        executeRequest(params).then(resolve)
      })
    },
    [config.debounceWait, config.throttleWait, executeRequest]
  )

  /**
   * 刷新(使用上次参数)
   */
  const refresh = useCallback(() => {
    return run(...latestParamsRef.current)
  }, [run])

  /**
   * 取消请求
   */
  const cancel = useCallback(() => {
    canceledRef.current = true
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
    }
    if (throttleTimerRef.current) {
      clearTimeout(throttleTimerRef.current)
    }
  }, [])

  /**
   * 手动修改数据
   */
  const mutate = useCallback((newData: T | ((oldData: T | undefined) => T)) => {
    setData((prevData) => {
      const nextData =
        typeof newData === 'function'
          ? (newData as (oldData: T | undefined) => T)(prevData)
          : newData
      return nextData
    })
  }, [])

  /**
   * 自动执行
   */
  useEffect(() => {
    if (!config.manual) {
      run(...(config.defaultParams || []))
    }

    return () => {
      cancel()
    }
  }, []) // 只在mount时执行

  return [
    { data, error, loading },
    { run, refresh, cancel, mutate },
  ]
}

export default useRequest

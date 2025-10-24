/**
 * 请求重试策略 - 智能重试机制
 */

interface RetryConfig {
  maxRetries?: number // 最大重试次数
  baseDelay?: number // 基础延迟(ms)
  maxDelay?: number // 最大延迟(ms)
  backoffFactor?: number // 退避因子
  retryableErrors?: number[] // 可重试的HTTP状态码
  onRetry?: (attempt: number, error: Error) => void
}

/**
 * 指数退避算法
 */
function exponentialBackoff(
  attempt: number,
  baseDelay: number,
  maxDelay: number,
  backoffFactor: number
): number {
  const delay = Math.min(baseDelay * Math.pow(backoffFactor, attempt), maxDelay)
  // 添加随机抖动,避免惊群效应
  const jitter = delay * 0.1 * Math.random()
  return delay + jitter
}

/**
 * 判断错误是否可重试
 */
function isRetryableError(error: any, retryableErrors: number[]): boolean {
  // 网络错误
  if (!error.response) {
    return true
  }

  // 特定的HTTP状态码
  const status = error.response?.status
  return retryableErrors.includes(status)
}

/**
 * 请求重试装饰器
 *
 * @example
 * const fetchWithRetry = withRetry(
 *   () => api.get('/merchant/profile'),
 *   {
 *     maxRetries: 3,
 *     baseDelay: 1000,
 *     retryableErrors: [408, 429, 500, 502, 503, 504]
 *   }
 * )
 */
export async function withRetry<T>(
  fn: () => Promise<T>,
  config: RetryConfig = {}
): Promise<T> {
  const {
    maxRetries = 3,
    baseDelay = 1000,
    maxDelay = 30000,
    backoffFactor = 2,
    retryableErrors = [408, 429, 500, 502, 503, 504],
    onRetry,
  } = config

  let lastError: Error

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      return await fn()
    } catch (error: any) {
      lastError = error

      // 最后一次尝试或不可重试的错误
      if (attempt === maxRetries || !isRetryableError(error, retryableErrors)) {
        throw error
      }

      // 计算延迟时间
      const delay = exponentialBackoff(attempt, baseDelay, maxDelay, backoffFactor)

      console.warn(
        `[Retry] Attempt ${attempt + 1}/${maxRetries} failed, retrying in ${delay}ms...`,
        error
      )

      onRetry?.(attempt + 1, error)

      // 等待后重试
      await new Promise((resolve) => setTimeout(resolve, delay))
    }
  }

  throw lastError!
}

/**
 * 创建带重试的函数
 *
 * @example
 * const retryableFetch = createRetryable(
 *   (url: string) => fetch(url).then(r => r.json()),
 *   { maxRetries: 3 }
 * )
 *
 * const data = await retryableFetch('/api/data')
 */
export function createRetryable<T extends (...args: any[]) => Promise<any>>(
  fn: T,
  config: RetryConfig = {}
): T {
  return ((...args: Parameters<T>) => {
    return withRetry(() => fn(...args), config)
  }) as T
}

/**
 * 批量请求重试
 *
 * @example
 * const results = await retryBatch([
 *   () => api.get('/endpoint1'),
 *   () => api.get('/endpoint2'),
 *   () => api.get('/endpoint3'),
 * ], { maxRetries: 2 })
 */
export async function retryBatch<T>(
  requests: (() => Promise<T>)[],
  config: RetryConfig = {}
): Promise<T[]> {
  return Promise.all(requests.map((req) => withRetry(req, config)))
}

/**
 * 带超时的重试
 *
 * @example
 * const data = await withRetryAndTimeout(
 *   () => api.get('/slow-endpoint'),
 *   { maxRetries: 3, timeout: 5000 }
 * )
 */
export async function withRetryAndTimeout<T>(
  fn: () => Promise<T>,
  config: RetryConfig & { timeout?: number } = {}
): Promise<T> {
  const { timeout = 10000, ...retryConfig } = config

  return withRetry(async () => {
    return Promise.race([
      fn(),
      new Promise<never>((_, reject) =>
        setTimeout(() => reject(new Error('Request timeout')), timeout)
      ),
    ])
  }, retryConfig)
}

export default withRetry

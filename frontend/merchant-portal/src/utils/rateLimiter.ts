/**
 * 请求限流工具 - 防止API滥用
 */

interface RateLimitConfig {
  maxRequests: number // 最大请求数
  timeWindow: number // 时间窗口(毫秒)
  blockDuration?: number // 阻止时长(毫秒)
}

interface RequestRecord {
  timestamps: number[]
  blockedUntil?: number
}

class RateLimiter {
  private records: Map<string, RequestRecord> = new Map()
  private config: Required<RateLimitConfig>

  constructor(config: RateLimitConfig) {
    this.config = {
      blockDuration: 60000, // 默认阻止1分钟
      ...config,
    }
  }

  /**
   * 检查是否允许请求
   * @param key 限流标识(如API端点、用户ID)
   * @returns 是否允许请求
   */
  isAllowed(key: string): boolean {
    const now = Date.now()
    const record = this.records.get(key) || { timestamps: [] }

    // 检查是否在阻止期内
    if (record.blockedUntil && now < record.blockedUntil) {
      return false
    }

    // 移除过期的时间戳
    record.timestamps = record.timestamps.filter(
      (timestamp) => now - timestamp < this.config.timeWindow
    )

    // 检查是否超过限制
    if (record.timestamps.length >= this.config.maxRequests) {
      record.blockedUntil = now + this.config.blockDuration
      this.records.set(key, record)
      return false
    }

    // 记录新请求
    record.timestamps.push(now)
    this.records.set(key, record)
    return true
  }

  /**
   * 获取剩余请求次数
   */
  getRemainingRequests(key: string): number {
    const now = Date.now()
    const record = this.records.get(key)

    if (!record) {
      return this.config.maxRequests
    }

    // 移除过期的时间戳
    const validTimestamps = record.timestamps.filter(
      (timestamp) => now - timestamp < this.config.timeWindow
    )

    return Math.max(0, this.config.maxRequests - validTimestamps.length)
  }

  /**
   * 获取重置时间(秒)
   */
  getResetTime(key: string): number {
    const record = this.records.get(key)
    if (!record || record.timestamps.length === 0) {
      return 0
    }

    const oldestTimestamp = Math.min(...record.timestamps)
    const resetTime = oldestTimestamp + this.config.timeWindow
    return Math.max(0, Math.ceil((resetTime - Date.now()) / 1000))
  }

  /**
   * 清除指定key的记录
   */
  reset(key: string): void {
    this.records.delete(key)
  }

  /**
   * 清除所有记录
   */
  resetAll(): void {
    this.records.clear()
  }
}

/**
 * 全局限流器实例
 */
export const globalRateLimiter = new RateLimiter({
  maxRequests: 100, // 每分钟100次
  timeWindow: 60000, // 1分钟
  blockDuration: 60000, // 阻止1分钟
})

/**
 * API端点限流器(更严格)
 */
export const apiRateLimiter = new RateLimiter({
  maxRequests: 30, // 每分钟30次
  timeWindow: 60000,
  blockDuration: 120000, // 阻止2分钟
})

/**
 * 登录限流器(最严格)
 */
export const loginRateLimiter = new RateLimiter({
  maxRequests: 5, // 每10分钟5次
  timeWindow: 600000, // 10分钟
  blockDuration: 600000, // 阻止10分钟
})

export default RateLimiter

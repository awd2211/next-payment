/**
 * 本地存储工具（支持过期时间）
 */

interface StorageData<T> {
  value: T
  expire?: number
}

class Storage {
  private storage: globalThis.Storage

  constructor(storage: globalThis.Storage) {
    this.storage = storage
  }

  /**
   * 设置存储项
   * @param key 键
   * @param value 值
   * @param expire 过期时间（毫秒），不传则永久有效
   */
  set<T>(key: string, value: T, expire?: number): void {
    const data: StorageData<T> = {
      value,
      expire: expire ? Date.now() + expire : undefined,
    }
    this.storage.setItem(key, JSON.stringify(data))
  }

  /**
   * 获取存储项
   * @param key 键
   * @returns 值，如果不存在或已过期则返回null
   */
  get<T>(key: string): T | null {
    const item = this.storage.getItem(key)
    if (!item) return null

    try {
      const data: StorageData<T> = JSON.parse(item)

      // 检查是否过期
      if (data.expire && Date.now() > data.expire) {
        this.remove(key)
        return null
      }

      return data.value
    } catch {
      return null
    }
  }

  /**
   * 移除存储项
   * @param key 键
   */
  remove(key: string): void {
    this.storage.removeItem(key)
  }

  /**
   * 清空所有存储
   */
  clear(): void {
    this.storage.clear()
  }

  /**
   * 获取所有键
   */
  keys(): string[] {
    const keys: string[] = []
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key) keys.push(key)
    }
    return keys
  }

  /**
   * 检查键是否存在
   * @param key 键
   */
  has(key: string): boolean {
    return this.get(key) !== null
  }
}

// 导出实例
export const localStorage = new Storage(window.localStorage)
export const sessionStorage = new Storage(window.sessionStorage)

// 默认导出localStorage
export default localStorage






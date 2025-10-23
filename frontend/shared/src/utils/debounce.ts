/**
 * 防抖函数
 * @param func 要执行的函数
 * @param wait 等待时间（毫秒）
 * @param immediate 是否立即执行
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number,
  immediate = false
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null

  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      timeout = null
      if (!immediate) func(...args)
    }

    const callNow = immediate && !timeout

    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(later, wait)

    if (callNow) func(...args)
  }
}

/**
 * 节流函数
 * @param func 要执行的函数
 * @param wait 等待时间（毫秒）
 * @param options 选项
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  wait: number,
  options: { leading?: boolean; trailing?: boolean } = {}
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null
  let previous = 0
  const { leading = true, trailing = true } = options

  return function executedFunction(...args: Parameters<T>) {
    const now = Date.now()

    if (!previous && leading === false) previous = now

    const remaining = wait - (now - previous)

    if (remaining <= 0 || remaining > wait) {
      if (timeout) {
        clearTimeout(timeout)
        timeout = null
      }
      previous = now
      func(...args)
    } else if (!timeout && trailing) {
      timeout = setTimeout(() => {
        previous = leading === false ? 0 : Date.now()
        timeout = null
        func(...args)
      }, remaining)
    }
  }
}

/**
 * 异步防抖（返回Promise）
 */
export function asyncDebounce<T extends (...args: any[]) => Promise<any>>(
  func: T,
  wait: number
): (...args: Parameters<T>) => Promise<ReturnType<T>> {
  let timeout: NodeJS.Timeout | null = null
  let resolveList: Array<(value: any) => void> = []
  let rejectList: Array<(reason?: any) => void> = []

  return function executedFunction(...args: Parameters<T>): Promise<ReturnType<T>> {
    return new Promise((resolve, reject) => {
      resolveList.push(resolve)
      rejectList.push(reject)

      if (timeout) clearTimeout(timeout)

      timeout = setTimeout(async () => {
        timeout = null
        const currentResolveList = resolveList
        const currentRejectList = rejectList
        resolveList = []
        rejectList = []

        try {
          const result = await func(...args)
          currentResolveList.forEach(r => r(result))
        } catch (error) {
          currentRejectList.forEach(r => r(error))
        }
      }, wait)
    })
  }
}


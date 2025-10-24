/**
 * 性能监控工具 - 监控页面性能指标
 */

interface PerformanceMetrics {
  // 页面加载性能
  FCP?: number // First Contentful Paint
  LCP?: number // Largest Contentful Paint
  FID?: number // First Input Delay
  CLS?: number // Cumulative Layout Shift
  TTFB?: number // Time to First Byte

  // 导航性能
  dns?: number // DNS解析时间
  tcp?: number // TCP连接时间
  request?: number // 请求时间
  response?: number // 响应时间
  domParse?: number // DOM解析时间
  domReady?: number // DOM Ready时间
  loadComplete?: number // 页面完全加载时间

  // 资源性能
  resourceCount?: number // 资源数量
  resourceSize?: number // 资源总大小
}

/**
 * 性能监控类
 */
class PerformanceMonitor {
  private metrics: PerformanceMetrics = {}
  private observers: PerformanceObserver[] = []

  /**
   * 初始化性能监控
   */
  init(): void {
    if (typeof window === 'undefined' || !window.performance) {
      return
    }

    this.observeWebVitals()
    this.observeNavigation()
    this.observeResources()
  }

  /**
   * 监控Web Vitals (核心性能指标)
   */
  private observeWebVitals(): void {
    // FCP - First Contentful Paint
    this.observePaint('first-contentful-paint', (value) => {
      this.metrics.FCP = value
      this.reportMetric('FCP', value)
    })

    // LCP - Largest Contentful Paint
    this.observeLCP()

    // FID - First Input Delay
    this.observeFID()

    // CLS - Cumulative Layout Shift
    this.observeCLS()

    // TTFB - Time to First Byte
    this.observeTTFB()
  }

  /**
   * 监控Paint事件
   */
  private observePaint(name: string, callback: (value: number) => void): void {
    if (!('PerformanceObserver' in window)) return

    try {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.name === name) {
            callback(entry.startTime)
          }
        }
      })
      observer.observe({ type: 'paint', buffered: true })
      this.observers.push(observer)
    } catch (e) {
      console.warn('PerformanceObserver not supported:', e)
    }
  }

  /**
   * 监控LCP
   */
  private observeLCP(): void {
    if (!('PerformanceObserver' in window)) return

    try {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        const lastEntry = entries[entries.length - 1] as any
        this.metrics.LCP = lastEntry.renderTime || lastEntry.loadTime
        this.reportMetric('LCP', this.metrics.LCP)
      })
      observer.observe({ type: 'largest-contentful-paint', buffered: true })
      this.observers.push(observer)
    } catch (e) {
      console.warn('LCP observer not supported:', e)
    }
  }

  /**
   * 监控FID
   */
  private observeFID(): void {
    if (!('PerformanceObserver' in window)) return

    try {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          const fidEntry = entry as any
          this.metrics.FID = fidEntry.processingStart - fidEntry.startTime
          this.reportMetric('FID', this.metrics.FID)
        }
      })
      observer.observe({ type: 'first-input', buffered: true })
      this.observers.push(observer)
    } catch (e) {
      console.warn('FID observer not supported:', e)
    }
  }

  /**
   * 监控CLS
   */
  private observeCLS(): void {
    if (!('PerformanceObserver' in window)) return

    try {
      let clsValue = 0
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          const layoutShift = entry as any
          if (!layoutShift.hadRecentInput) {
            clsValue += layoutShift.value
            this.metrics.CLS = clsValue
            this.reportMetric('CLS', clsValue)
          }
        }
      })
      observer.observe({ type: 'layout-shift', buffered: true })
      this.observers.push(observer)
    } catch (e) {
      console.warn('CLS observer not supported:', e)
    }
  }

  /**
   * 监控TTFB
   */
  private observeTTFB(): void {
    if (!window.performance?.timing) return

    window.addEventListener('load', () => {
      const { responseStart, requestStart } = window.performance.timing
      this.metrics.TTFB = responseStart - requestStart
      this.reportMetric('TTFB', this.metrics.TTFB)
    })
  }

  /**
   * 监控导航性能
   */
  private observeNavigation(): void {
    window.addEventListener('load', () => {
      const timing = window.performance.timing

      this.metrics.dns = timing.domainLookupEnd - timing.domainLookupStart
      this.metrics.tcp = timing.connectEnd - timing.connectStart
      this.metrics.request = timing.responseStart - timing.requestStart
      this.metrics.response = timing.responseEnd - timing.responseStart
      this.metrics.domParse = timing.domInteractive - timing.domLoading
      this.metrics.domReady = timing.domContentLoadedEventEnd - timing.navigationStart
      this.metrics.loadComplete = timing.loadEventEnd - timing.navigationStart

      this.reportMetric('Navigation', this.metrics)
    })
  }

  /**
   * 监控资源加载
   */
  private observeResources(): void {
    window.addEventListener('load', () => {
      const resources = window.performance.getEntriesByType('resource')
      this.metrics.resourceCount = resources.length

      let totalSize = 0
      resources.forEach((resource: any) => {
        totalSize += resource.transferSize || 0
      })
      this.metrics.resourceSize = totalSize

      this.reportMetric('Resources', {
        count: this.metrics.resourceCount,
        size: `${(totalSize / 1024).toFixed(2)} KB`,
      })
    })
  }

  /**
   * 上报性能指标
   */
  private reportMetric(name: string, value: any): void {
    // 开发环境输出到控制台
    if (process.env.NODE_ENV === 'development') {
      console.log(`[Performance] ${name}:`, value)
    }

    // 生产环境可以上报到监控系统
    if (process.env.NODE_ENV === 'production') {
      // TODO: 上报到监控系统
      // fetch('/api/v1/metrics', {
      //   method: 'POST',
      //   body: JSON.stringify({ metric: name, value })
      // })
    }
  }

  /**
   * 获取所有指标
   */
  getMetrics(): PerformanceMetrics {
    return { ...this.metrics }
  }

  /**
   * 获取性能评分 (0-100)
   */
  getScore(): number {
    const { FCP = 0, LCP = 0, FID = 0, CLS = 0 } = this.metrics

    let score = 100

    // FCP评分 (< 1.8s: 好, < 3s: 中, >= 3s: 差)
    if (FCP >= 3000) score -= 25
    else if (FCP >= 1800) score -= 10

    // LCP评分 (< 2.5s: 好, < 4s: 中, >= 4s: 差)
    if (LCP >= 4000) score -= 25
    else if (LCP >= 2500) score -= 10

    // FID评分 (< 100ms: 好, < 300ms: 中, >= 300ms: 差)
    if (FID >= 300) score -= 25
    else if (FID >= 100) score -= 10

    // CLS评分 (< 0.1: 好, < 0.25: 中, >= 0.25: 差)
    if (CLS >= 0.25) score -= 25
    else if (CLS >= 0.1) score -= 10

    return Math.max(0, score)
  }

  /**
   * 测量函数执行时间
   */
  measureFunction<T>(name: string, fn: () => T): T {
    const start = performance.now()
    const result = fn()
    const duration = performance.now() - start

    this.reportMetric(`Function:${name}`, `${duration.toFixed(2)}ms`)
    return result
  }

  /**
   * 测量异步函数执行时间
   */
  async measureAsync<T>(name: string, fn: () => Promise<T>): Promise<T> {
    const start = performance.now()
    const result = await fn()
    const duration = performance.now() - start

    this.reportMetric(`AsyncFunction:${name}`, `${duration.toFixed(2)}ms`)
    return result
  }

  /**
   * 清理observers
   */
  cleanup(): void {
    this.observers.forEach((observer) => observer.disconnect())
    this.observers = []
  }
}

// 单例实例
export const performanceMonitor = new PerformanceMonitor()

// 自动初始化
if (typeof window !== 'undefined') {
  performanceMonitor.init()
}

export default performanceMonitor

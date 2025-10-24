import { useRef, useState, useEffect, CSSProperties, ReactNode } from 'react'

interface VirtualListProps<T> {
  data: T[]
  itemHeight: number
  containerHeight: number
  renderItem: (item: T, index: number) => ReactNode
  overscan?: number // 额外渲染的项数
  className?: string
  style?: CSSProperties
}

/**
 * 虚拟滚动列表组件 - 优化大列表渲染性能
 *
 * @example
 * <VirtualList
 *   data={transactions}
 *   itemHeight={60}
 *   containerHeight={600}
 *   renderItem={(item, index) => (
 *     <div key={item.id}>
 *       {item.name} - {item.amount}
 *     </div>
 *   )}
 *   overscan={5}
 * />
 */
function VirtualList<T>({
  data,
  itemHeight,
  containerHeight,
  renderItem,
  overscan = 3,
  className,
  style,
}: VirtualListProps<T>) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [scrollTop, setScrollTop] = useState(0)

  // 计算可见区域
  const totalHeight = data.length * itemHeight
  const visibleCount = Math.ceil(containerHeight / itemHeight)
  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan)
  const endIndex = Math.min(
    data.length - 1,
    startIndex + visibleCount + 2 * overscan
  )

  // 可见项
  const visibleItems = data.slice(startIndex, endIndex + 1)

  // 滚动事件处理
  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.target as HTMLDivElement
    setScrollTop(target.scrollTop)
  }

  // 优化: 使用requestAnimationFrame节流滚动事件
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    let rafId: number

    const throttledScroll = (e: Event) => {
      if (rafId) {
        cancelAnimationFrame(rafId)
      }
      rafId = requestAnimationFrame(() => {
        const target = e.target as HTMLDivElement
        setScrollTop(target.scrollTop)
      })
    }

    container.addEventListener('scroll', throttledScroll, { passive: true })

    return () => {
      container.removeEventListener('scroll', throttledScroll)
      if (rafId) {
        cancelAnimationFrame(rafId)
      }
    }
  }, [])

  return (
    <div
      ref={containerRef}
      className={className}
      style={{
        height: containerHeight,
        overflow: 'auto',
        position: 'relative',
        ...style,
      }}
      onScroll={handleScroll}
    >
      {/* 占位容器,撑起总高度 */}
      <div style={{ height: totalHeight, position: 'relative' }}>
        {/* 可见项容器 */}
        <div
          style={{
            position: 'absolute',
            top: startIndex * itemHeight,
            left: 0,
            right: 0,
          }}
        >
          {visibleItems.map((item, i) => (
            <div
              key={startIndex + i}
              style={{
                height: itemHeight,
                overflow: 'hidden',
              }}
            >
              {renderItem(item, startIndex + i)}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default VirtualList

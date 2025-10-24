import { useState, useEffect, useRef, CSSProperties } from 'react'
import { Skeleton } from 'antd'

interface LazyImageProps {
  src: string
  alt?: string
  placeholder?: string
  width?: number | string
  height?: number | string
  className?: string
  style?: CSSProperties
  onLoad?: () => void
  onError?: () => void
  threshold?: number // Intersection Observer阈值
}

/**
 * 懒加载图片组件 - 优化页面性能
 *
 * @example
 * <LazyImage
 *   src="https://example.com/large-image.jpg"
 *   alt="Product Image"
 *   width={300}
 *   height={200}
 *   placeholder="/placeholder.png"
 * />
 */
const LazyImage = ({
  src,
  alt = '',
  placeholder,
  width,
  height,
  className,
  style,
  onLoad,
  onError,
  threshold = 0.01,
}: LazyImageProps) => {
  const [imageSrc, setImageSrc] = useState<string | undefined>(placeholder)
  const [isLoading, setIsLoading] = useState(true)
  const [isError, setIsError] = useState(false)
  const imgRef = useRef<HTMLImageElement>(null)
  const observerRef = useRef<IntersectionObserver | null>(null)

  useEffect(() => {
    if (!imgRef.current) return

    // 创建Intersection Observer
    observerRef.current = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            // 图片进入视口,开始加载
            const img = new Image()
            img.src = src

            img.onload = () => {
              setImageSrc(src)
              setIsLoading(false)
              onLoad?.()
              // 停止观察
              if (observerRef.current && imgRef.current) {
                observerRef.current.unobserve(imgRef.current)
              }
            }

            img.onerror = () => {
              setIsError(true)
              setIsLoading(false)
              onError?.()
              // 停止观察
              if (observerRef.current && imgRef.current) {
                observerRef.current.unobserve(imgRef.current)
              }
            }
          }
        })
      },
      {
        threshold,
        rootMargin: '50px', // 提前50px开始加载
      }
    )

    observerRef.current.observe(imgRef.current)

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect()
      }
    }
  }, [src, threshold, onLoad, onError])

  // 显示骨架屏
  if (isLoading && !placeholder) {
    return (
      <div style={{ width, height, ...style }} className={className}>
        <Skeleton.Image active style={{ width: '100%', height: '100%' }} />
      </div>
    )
  }

  // 显示错误占位符
  if (isError) {
    return (
      <div
        style={{
          width,
          height,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: '#f5f5f5',
          color: '#999',
          ...style,
        }}
        className={className}
      >
        加载失败
      </div>
    )
  }

  return (
    <img
      ref={imgRef}
      src={imageSrc}
      alt={alt}
      width={width}
      height={height}
      className={className}
      style={{
        ...style,
        opacity: isLoading ? 0.5 : 1,
        transition: 'opacity 0.3s ease-in-out',
      }}
    />
  )
}

export default LazyImage

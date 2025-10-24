import { useEffect, useState, RefObject } from 'react'

interface UseIntersectionObserverOptions {
  threshold?: number | number[]
  root?: Element | null
  rootMargin?: string
  freezeOnceVisible?: boolean
}

/**
 * Intersection Observer Hook - 监控元素是否进入视口
 *
 * @example
 * const ref = useRef<HTMLDivElement>(null)
 * const isVisible = useIntersectionObserver(ref, { threshold: 0.5 })
 *
 * return (
 *   <div ref={ref}>
 *     {isVisible ? '元素可见' : '元素不可见'}
 *   </div>
 * )
 */
function useIntersectionObserver(
  elementRef: RefObject<Element>,
  options: UseIntersectionObserverOptions = {}
): boolean {
  const {
    threshold = 0,
    root = null,
    rootMargin = '0px',
    freezeOnceVisible = false,
  } = options

  const [isIntersecting, setIsIntersecting] = useState(false)

  useEffect(() => {
    const element = elementRef.current

    if (!element) return

    // 如果已经可见且设置了freezeOnceVisible,不再更新
    if (freezeOnceVisible && isIntersecting) {
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsIntersecting(entry.isIntersecting)
      },
      { threshold, root, rootMargin }
    )

    observer.observe(element)

    return () => {
      observer.unobserve(element)
    }
  }, [elementRef, threshold, root, rootMargin, freezeOnceVisible, isIntersecting])

  return isIntersecting
}

export default useIntersectionObserver

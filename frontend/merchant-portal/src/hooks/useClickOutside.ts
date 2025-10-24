import { useEffect, RefObject } from 'react'

/**
 * 点击外部区域Hook - 检测点击是否发生在元素外部
 *
 * @example
 * const ref = useRef<HTMLDivElement>(null)
 * useClickOutside(ref, () => {
 *   console.log('点击了外部区域')
 *   setIsOpen(false)
 * })
 *
 * return <div ref={ref}>内容</div>
 */
function useClickOutside(
  ref: RefObject<HTMLElement>,
  handler: (event: MouseEvent | TouchEvent) => void
): void {
  useEffect(() => {
    const listener = (event: MouseEvent | TouchEvent) => {
      const el = ref.current

      // 如果点击的是元素内部或元素不存在,不执行handler
      if (!el || el.contains(event.target as Node)) {
        return
      }

      handler(event)
    }

    document.addEventListener('mousedown', listener)
    document.addEventListener('touchstart', listener)

    return () => {
      document.removeEventListener('mousedown', listener)
      document.removeEventListener('touchstart', listener)
    }
  }, [ref, handler])
}

export default useClickOutside

import { useState, useEffect } from 'react'

/**
 * 媒体查询Hook - 响应式设计
 *
 * @example
 * const isMobile = useMediaQuery('(max-width: 768px)')
 * const isTablet = useMediaQuery('(min-width: 769px) and (max-width: 1024px)')
 * const isDarkMode = useMediaQuery('(prefers-color-scheme: dark)')
 *
 * return (
 *   <div>
 *     {isMobile ? <MobileView /> : <DesktopView />}
 *   </div>
 * )
 */
function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() => {
    if (typeof window === 'undefined') return false
    return window.matchMedia(query).matches
  })

  useEffect(() => {
    if (typeof window === 'undefined') return

    const mediaQuery = window.matchMedia(query)
    const handleChange = (e: MediaQueryListEvent) => {
      setMatches(e.matches)
    }

    // 现代浏览器
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange)
    } else {
      // 旧版浏览器降级
      mediaQuery.addListener(handleChange)
    }

    return () => {
      if (mediaQuery.removeEventListener) {
        mediaQuery.removeEventListener('change', handleChange)
      } else {
        mediaQuery.removeListener(handleChange)
      }
    }
  }, [query])

  return matches
}

/**
 * 预设的媒体查询Hooks
 */
export const useIsMobile = () => useMediaQuery('(max-width: 768px)')
export const useIsTablet = () => useMediaQuery('(min-width: 769px) and (max-width: 1024px)')
export const useIsDesktop = () => useMediaQuery('(min-width: 1025px)')
export const useIsDarkMode = () => useMediaQuery('(prefers-color-scheme: dark)')
export const usePrefersReducedMotion = () => useMediaQuery('(prefers-reduced-motion: reduce)')

export default useMediaQuery

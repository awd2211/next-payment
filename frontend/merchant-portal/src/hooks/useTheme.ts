import { useState, useEffect } from 'react'

export type ThemeMode = 'light' | 'dark'

const THEME_STORAGE_KEY = 'theme-mode'

export const useTheme = () => {
  const [theme, setTheme] = useState<ThemeMode>(() => {
    const savedTheme = localStorage.getItem(THEME_STORAGE_KEY)
    return (savedTheme as ThemeMode) || 'light'
  })

  useEffect(() => {
    localStorage.setItem(THEME_STORAGE_KEY, theme)
    // 触发自定义事件通知主题变化
    window.dispatchEvent(new CustomEvent('themechange', { detail: theme }))
  }, [theme])

  const toggleTheme = () => {
    setTheme((prevTheme) => (prevTheme === 'light' ? 'dark' : 'light'))
  }

  const setLightTheme = () => setTheme('light')
  const setDarkTheme = () => setTheme('dark')

  return {
    theme,
    isDark: theme === 'dark',
    isLight: theme === 'light',
    toggleTheme,
    setLightTheme,
    setDarkTheme,
  }
}

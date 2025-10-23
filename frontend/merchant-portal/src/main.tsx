import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider, theme as antdTheme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import 'dayjs/locale/en'
import App from './App'
import './index.css'
import './i18n/config'
import { getThemeConfig } from './theme/config'
import type { ThemeMode } from './hooks/useTheme'
import ErrorBoundary from './components/ErrorBoundary'

// 动态设置 dayjs 语言
const getAntdLocale = (lang: string) => {
  return lang === 'zh-CN' ? zhCN : enUS
}

const getDayjsLocale = (lang: string) => {
  return lang === 'zh-CN' ? 'zh-cn' : 'en'
}

dayjs.locale(getDayjsLocale(localStorage.getItem('i18nextLng') || 'zh-CN'))

// 创建一个动态语言和主题的包装器
const AppWithI18n = () => {
  const [locale, setLocale] = React.useState(() =>
    getAntdLocale(localStorage.getItem('i18nextLng') || 'zh-CN')
  )
  const [themeMode, setThemeMode] = React.useState<ThemeMode>(() => {
    return (localStorage.getItem('theme-mode') as ThemeMode) || 'light'
  })

  React.useEffect(() => {
    const handleLanguageChange = (lng: string) => {
      setLocale(getAntdLocale(lng))
      dayjs.locale(getDayjsLocale(lng))
    }

    const handleThemeChange = (theme: ThemeMode) => {
      setThemeMode(theme)
    }

    // 监听语言变化
    window.addEventListener('languagechange', ((e: CustomEvent) => {
      handleLanguageChange(e.detail)
    }) as EventListener)

    // 监听主题变化
    window.addEventListener('themechange', ((e: CustomEvent) => {
      handleThemeChange(e.detail)
    }) as EventListener)

    return () => {
      window.removeEventListener('languagechange', () => {})
      window.removeEventListener('themechange', () => {})
    }
  }, [])

  const isDark = themeMode === 'dark'
  const themeConfig = getThemeConfig(isDark)

  return (
    <ConfigProvider
      locale={locale}
      theme={{
        ...themeConfig,
        algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
      }}
    >
      <App />
    </ConfigProvider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <AppWithI18n />
    </ErrorBoundary>
  </React.StrictMode>,
)

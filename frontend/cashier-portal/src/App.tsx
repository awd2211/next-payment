import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import enUS from 'antd/locale/en_US'
import zhCN from 'antd/locale/zh_CN'
import { useTranslation } from 'react-i18next'
import Checkout from './pages/Checkout'
import './i18n'

function App() {
  const { i18n } = useTranslation()

  const antdLocale = i18n.language === 'zh-CN' ? zhCN : enUS

  return (
    <ConfigProvider locale={antdLocale}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Checkout />} />
          <Route path="/checkout" element={<Checkout />} />
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App

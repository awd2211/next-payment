import { Dropdown } from 'antd'
import { GlobalOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { MenuProps } from 'antd'

const LanguageSwitcher = () => {
  const { i18n } = useTranslation()

  const currentLanguage = i18n.language

  const items: MenuProps['items'] = [
    {
      key: 'zh-CN',
      label: (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span>🇨🇳</span>
          <span>简体中文</span>
          {currentLanguage === 'zh-CN' && <span style={{ color: '#1890ff' }}>✓</span>}
        </div>
      ),
    },
    {
      key: 'en-US',
      label: (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span>🇺🇸</span>
          <span>English</span>
          {currentLanguage === 'en-US' && <span style={{ color: '#1890ff' }}>✓</span>}
        </div>
      ),
    },
  ]

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    i18n.changeLanguage(key)
    // 触发自定义事件通知 Ant Design 语言变化
    window.dispatchEvent(new CustomEvent('languagechange', { detail: key }))
  }

  return (
    <Dropdown menu={{ items, onClick: handleMenuClick }} placement="bottomRight">
      <div
        style={{
          cursor: 'pointer',
          padding: '0 12px',
          height: '100%',
          display: 'flex',
          alignItems: 'center',
        }}
      >
        <GlobalOutlined style={{ fontSize: '18px' }} />
      </div>
    </Dropdown>
  )
}

export default LanguageSwitcher

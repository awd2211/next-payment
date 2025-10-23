import { Switch } from 'antd'
import { BulbOutlined, BulbFilled } from '@ant-design/icons'
import { useTheme } from '../hooks/useTheme'
import { useTranslation } from 'react-i18next'

const ThemeSwitcher = () => {
  const { isDark, toggleTheme } = useTheme()
  const { t } = useTranslation()

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        padding: '0 12px',
      }}
    >
      {isDark ? (
        <BulbFilled style={{ fontSize: '18px', color: '#faad14' }} />
      ) : (
        <BulbOutlined style={{ fontSize: '18px' }} />
      )}
      <Switch
        checked={isDark}
        onChange={toggleTheme}
        checkedChildren={t('layout.darkMode')}
        unCheckedChildren={t('layout.lightMode')}
      />
    </div>
  )
}

export default ThemeSwitcher

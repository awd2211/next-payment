import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Layout as AntLayout,
  Menu,
  Avatar,
  Dropdown,
  theme,
  Space,
  Typography,
} from 'antd'
import type { MenuProps } from 'antd'
import {
  DashboardOutlined,
  TransactionOutlined,
  ShoppingOutlined,
  UserOutlined,
  LogoutOutlined,
  WalletOutlined,
  PlusCircleOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../stores/authStore'
import LanguageSwitcher from './LanguageSwitcher'
import ThemeSwitcher from './ThemeSwitcher'
import NotificationDropdown from './NotificationDropdown'

const { Header, Sider, Content } = AntLayout
const { Text } = Typography

const Layout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { t } = useTranslation()
  const { merchant, clearAuth } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken()

  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: t('menu.dashboard'),
    },
    {
      key: '/create-payment',
      icon: <PlusCircleOutlined />,
      label: t('menu.createPayment'),
    },
    {
      key: '/transactions',
      icon: <TransactionOutlined />,
      label: t('menu.transactions'),
    },
    {
      key: '/orders',
      icon: <ShoppingOutlined />,
      label: t('menu.orders'),
    },
    {
      key: '/account',
      icon: <WalletOutlined />,
      label: t('menu.account'),
    },
  ]

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: t('layout.profile'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('layout.logout'),
      danger: true,
    },
  ]

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleUserMenuClick = ({ key }: { key: string }) => {
    if (key === 'logout') {
      clearAuth()
      navigate('/login')
    }
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme="dark"
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            fontSize: collapsed ? 16 : 20,
            fontWeight: 'bold',
          }}
        >
          {collapsed ? t('layout.logoShort') : t('layout.logo')}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <AntLayout>
        <Header
          style={{
            padding: '0 24px',
            background: colorBgContainer,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <div />
          <Space size="large">
            <ThemeSwitcher />
            <LanguageSwitcher />
            <NotificationDropdown />
            <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }}>
              <Space style={{ cursor: 'pointer' }}>
                <Avatar icon={<UserOutlined />} />
                <Text>{merchant?.name}</Text>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content
          style={{
            margin: '16px',
            padding: 24,
            minHeight: 280,
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
          }}
        >
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  )
}

export default Layout

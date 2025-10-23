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
  SettingOutlined,
  TeamOutlined,
  SafetyOutlined,
  FileTextOutlined,
  UserOutlined,
  LogoutOutlined,
  KeyOutlined,
  ShopOutlined,
  DollarOutlined,
  ShoppingOutlined,
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
  const { admin, clearAuth, hasPermission } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken()

  // 菜单项
  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: t('menu.dashboard'),
    },
    hasPermission('config.view') && {
      key: '/system-configs',
      icon: <SettingOutlined />,
      label: t('menu.systemConfigs'),
    },
    hasPermission('admin.view') && {
      key: '/admins',
      icon: <TeamOutlined />,
      label: t('menu.admins'),
    },
    hasPermission('role.view') && {
      key: '/roles',
      icon: <SafetyOutlined />,
      label: t('menu.roles'),
    },
    hasPermission('merchant.view') && {
      key: '/merchants',
      icon: <ShopOutlined />,
      label: t('menu.merchants'),
    },
    hasPermission('payment.view') && {
      key: '/payments',
      icon: <DollarOutlined />,
      label: t('menu.payments'),
    },
    hasPermission('order.view') && {
      key: '/orders',
      icon: <ShoppingOutlined />,
      label: t('menu.orders'),
    },
    hasPermission('audit.view') && {
      key: '/audit-logs',
      icon: <FileTextOutlined />,
      label: t('menu.auditLogs'),
    },
  ].filter(Boolean) as MenuProps['items']

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: t('layout.profile'),
    },
    {
      key: 'change-password',
      icon: <KeyOutlined />,
      label: '修改密码',
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
          {collapsed ? '支付' : '支付平台管理'}
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
                <Avatar src={admin?.avatar} icon={<UserOutlined />} />
                <Text>{admin?.full_name || admin?.username}</Text>
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

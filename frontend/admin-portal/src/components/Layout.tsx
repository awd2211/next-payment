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
} from '@ant-design/icons'
import { useAuthStore } from '../stores/authStore'

const { Header, Sider, Content } = AntLayout
const { Text } = Typography

const Layout = () => {
  const navigate = useNavigate()
  const location = useLocation()
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
      label: '仪表板',
    },
    hasPermission('config.view') && {
      key: '/system-configs',
      icon: <SettingOutlined />,
      label: '系统配置',
    },
    hasPermission('admin.view') && {
      key: '/admins',
      icon: <TeamOutlined />,
      label: '管理员管理',
    },
    hasPermission('role.view') && {
      key: '/roles',
      icon: <SafetyOutlined />,
      label: '角色权限',
    },
    hasPermission('merchant.view') && {
      key: '/merchants',
      icon: <ShopOutlined />,
      label: '商户管理',
    },
    hasPermission('audit.view') && {
      key: '/audit-logs',
      icon: <FileTextOutlined />,
      label: '审计日志',
    },
  ].filter(Boolean) as MenuProps['items']

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
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
      label: '退出登录',
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
          <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }}>
            <Space style={{ cursor: 'pointer' }}>
              <Avatar src={admin?.avatar} icon={<UserOutlined />} />
              <Text>{admin?.full_name || admin?.username}</Text>
            </Space>
          </Dropdown>
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

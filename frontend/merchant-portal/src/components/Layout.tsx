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
  RollbackOutlined,
  BankOutlined,
  KeyOutlined,
  SettingOutlined,
  ApiOutlined,
  MoneyCollectOutlined,
  BarChartOutlined,
  ExclamationCircleOutlined,
  ReconciliationOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../stores/authStore'
import LanguageSwitcher from './LanguageSwitcher'
import ThemeSwitcher from './ThemeSwitcher'
import NotificationDropdown from './NotificationDropdown'
import NetworkStatus from './NetworkStatus'

const { Header, Sider, Content } = AntLayout
const { Text } = Typography

const Layout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { t } = useTranslation()
  const { merchant, clearAuth } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)

  // 记住侧边栏状态
  const handleCollapse = (value: boolean) => {
    setCollapsed(value)
    localStorage.setItem('sidebarCollapsed', String(value))
  }

  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken()

  const menuItems: MenuProps['items'] = [
    // Dashboard - standalone
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: t('menu.dashboard') || '仪表板',
    },

    // Payment Operations - 3 items
    {
      key: 'payment-group',
      icon: <MoneyCollectOutlined />,
      label: t('menu.paymentOperations') || '支付业务',
      children: [
        {
          key: '/create-payment',
          icon: <PlusCircleOutlined />,
          label: t('menu.createPayment') || '发起支付',
        },
        {
          key: '/transactions',
          icon: <TransactionOutlined />,
          label: t('menu.transactions') || '交易记录',
        },
        {
          key: '/orders',
          icon: <ShoppingOutlined />,
          label: t('menu.orders') || '订单管理',
        },
      ],
    },

    // Finance Management - 4 items
    {
      key: 'finance-group',
      icon: <BankOutlined />,
      label: t('menu.financeManagement') || '财务管理',
      children: [
        {
          key: '/refunds',
          icon: <RollbackOutlined />,
          label: t('menu.refunds') || '退款管理',
        },
        {
          key: '/settlements',
          icon: <BankOutlined />,
          label: t('menu.settlement') || '结算账户',
        },
        {
          key: '/withdrawals',
          icon: <MoneyCollectOutlined />,
          label: t('menu.withdrawals') || '提现管理',
        },
        {
          key: '/reconciliation',
          icon: <ReconciliationOutlined />,
          label: t('menu.reconciliation') || '对账记录',
        },
      ],
    },

    // Service Management - 3 items
    {
      key: 'service-group',
      icon: <ApiOutlined />,
      label: t('menu.serviceManagement') || '服务管理',
      children: [
        {
          key: '/channels',
          icon: <ApiOutlined />,
          label: t('menu.channels') || '支付渠道',
        },
        {
          key: '/cashier-config',
          icon: <SettingOutlined />,
          label: t('menu.cashierConfig') || '收银台配置',
        },
        {
          key: '/disputes',
          icon: <ExclamationCircleOutlined />,
          label: t('menu.disputes') || '争议处理',
        },
      ],
    },

    // Data & Settings - 3 items
    {
      key: 'data-group',
      icon: <BarChartOutlined />,
      label: t('menu.dataAndSettings') || '数据与设置',
      children: [
        {
          key: '/analytics',
          icon: <BarChartOutlined />,
          label: t('menu.analytics') || '数据分析',
        },
        {
          key: '/api-keys',
          icon: <KeyOutlined />,
          label: t('menu.apiKeys') || 'API密钥',
        },
        {
          key: '/account',
          icon: <WalletOutlined />,
          label: t('menu.account') || '账户设置',
        },
      ],
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
    <>
      <NetworkStatus />
      <AntLayout style={{ minHeight: '100vh' }}>
        {/* 固定侧边栏 */}
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={handleCollapse}
          theme="dark"
          width={240}
          collapsedWidth={80}
          style={{
            overflow: 'auto',
            height: '100vh',
            position: 'fixed',
            left: 0,
            top: 0,
            bottom: 0,
            zIndex: 10,
            boxShadow: '2px 0 8px rgba(0,0,0,0.15)',
          }}
          trigger={null}
        >
          {/* Logo区域 */}
          <div
            style={{
              height: 64,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'white',
              fontSize: collapsed ? 18 : 22,
              fontWeight: 'bold',
              background: 'rgba(255, 255, 255, 0.1)',
              borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
              transition: 'all 0.2s',
            }}
          >
            {collapsed ? t('layout.logoShort') : t('layout.logo')}
          </div>

          {/* 菜单 */}
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={handleMenuClick}
            style={{
              borderRight: 0,
              paddingTop: 8,
            }}
          />

          {/* 折叠按钮 */}
          <div
            style={{
              position: 'absolute',
              bottom: 20,
              left: 0,
              right: 0,
              display: 'flex',
              justifyContent: 'center',
            }}
          >
            <div
              onClick={() => handleCollapse(!collapsed)}
              style={{
                width: 40,
                height: 40,
                borderRadius: '50%',
                background: 'rgba(255, 255, 255, 0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                color: 'white',
                fontSize: 16,
                transition: 'all 0.3s',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(255, 255, 255, 0.2)'
                e.currentTarget.style.transform = 'scale(1.1)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'rgba(255, 255, 255, 0.1)'
                e.currentTarget.style.transform = 'scale(1)'
              }}
            >
              {collapsed ? '»' : '«'}
            </div>
          </div>
        </Sider>

        {/* 主内容区域 */}
        <AntLayout style={{ marginLeft: collapsed ? 80 : 240, transition: 'margin-left 0.2s' }}>
          {/* 固定顶部导航栏 */}
          <Header
            style={{
              padding: '0 24px',
              background: colorBgContainer,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              position: 'sticky',
              top: 0,
              zIndex: 9,
              boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
              borderBottom: '1px solid #f0f0f0',
            }}
          >
            <div style={{ fontSize: 16, fontWeight: 500 }}>
              {/* 可以在这里显示当前页面标题 */}
            </div>
            <Space size="large">
              <ThemeSwitcher />
              <LanguageSwitcher />
              <NotificationDropdown />
              <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }}>
                <Space style={{ cursor: 'pointer', padding: '8px 12px', borderRadius: 8 }}>
                  <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1890ff' }} />
                  <Text strong>{merchant?.name}</Text>
                </Space>
              </Dropdown>
            </Space>
          </Header>

          {/* 内容区域 */}
          <Content
            style={{
              margin: '24px 16px 16px',
              padding: 24,
              minHeight: 'calc(100vh - 64px - 40px)',
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
            }}
          >
            <Outlet />
          </Content>
        </AntLayout>
      </AntLayout>
    </>
  )
}

export default Layout

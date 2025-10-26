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
  WarningOutlined,
  AccountBookOutlined,
  CreditCardOutlined,
  BankOutlined,
  WalletOutlined,
  ApiOutlined,
  CalculatorOutlined,
  BarChartOutlined,
  BellOutlined,
  ExclamationCircleOutlined,
  ReconciliationOutlined,
  SendOutlined,
  ControlOutlined,
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

  // 菜单项 - 分类结构
  const menuItems: MenuProps['items'] = [
    // 仪表板
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: t('menu.dashboard') || '仪表板',
    },

    // 商户管理
    hasPermission('merchant.view') && {
      key: 'merchant-group',
      icon: <ShopOutlined />,
      label: t('menu.merchantManagement') || '商户管理',
      children: [
        {
          key: '/merchants',
          label: t('menu.merchants') || '商户列表',
        },
        {
          key: '/kyc',
          label: t('menu.kyc') || 'KYC审核',
        },
        {
          key: '/merchant-limits',
          label: t('menu.merchantLimits') || '商户限额',
        },
      ],
    },

    // 交易管理
    hasPermission('payment.view') && {
      key: 'transaction-group',
      icon: <DollarOutlined />,
      label: t('menu.transactionManagement') || '交易管理',
      children: [
        {
          key: '/payments',
          label: t('menu.payments') || '支付记录',
        },
        {
          key: '/orders',
          label: t('menu.orders') || '订单管理',
        },
        {
          key: '/disputes',
          label: t('menu.disputes') || '争议处理',
        },
        {
          key: '/risk',
          label: t('menu.riskManagement') || '风险管理',
        },
      ],
    },

    // 财务管理
    hasPermission('accounting.view') && {
      key: 'finance-group',
      icon: <AccountBookOutlined />,
      label: t('menu.financeManagement') || '财务管理',
      children: [
        {
          key: '/accounting',
          label: t('menu.accounting') || '账务管理',
        },
        {
          key: '/settlements',
          label: t('menu.settlements') || '结算管理',
        },
        {
          key: '/withdrawals',
          label: t('menu.withdrawals') || '提现管理',
        },
        {
          key: '/reconciliation',
          label: t('menu.reconciliation') || '对账管理',
        },
      ],
    },

    // 渠道配置
    hasPermission('config.view') && {
      key: 'channel-group',
      icon: <ApiOutlined />,
      label: t('menu.channelConfig') || '渠道配置',
      children: [
        {
          key: '/channels',
          label: t('menu.channels') || '支付渠道',
        },
        {
          key: '/cashier',
          label: t('menu.cashier') || '收银台管理',
        },
        {
          key: '/webhooks',
          label: t('menu.webhooks') || 'Webhook管理',
        },
      ],
    },

    // 数据分析
    hasPermission('config.view') && {
      key: 'analytics-group',
      icon: <BarChartOutlined />,
      label: t('menu.analyticsCenter') || '数据中心',
      children: [
        {
          key: '/analytics',
          label: t('menu.analytics') || '数据分析',
        },
        {
          key: '/notifications',
          label: t('menu.notifications') || '通知管理',
        },
      ],
    },

    // 系统管理
    hasPermission('admin.view') && {
      key: 'system-group',
      icon: <SettingOutlined />,
      label: t('menu.systemManagement') || '系统管理',
      children: [
        hasPermission('config.view') && {
          key: '/system-configs',
          label: t('menu.systemConfigs') || '系统配置',
        },
        hasPermission('config.view') && {
          key: '/config-management',
          label: t('menu.configManagement') || '配置中心',
        },
        hasPermission('admin.view') && {
          key: '/admins',
          label: t('menu.admins') || '管理员',
        },
        hasPermission('role.view') && {
          key: '/roles',
          label: t('menu.roles') || '角色权限',
        },
        hasPermission('audit.view') && {
          key: '/audit-logs',
          label: t('menu.auditLogs') || '审计日志',
        },
      ].filter(Boolean),
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
      label: t('layout.changePassword'),
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

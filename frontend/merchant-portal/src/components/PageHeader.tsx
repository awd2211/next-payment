import { Space, Typography, Breadcrumb } from 'antd'
import { HomeOutlined } from '@ant-design/icons'
import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'

const { Title } = Typography

interface BreadcrumbItem {
  title: string
  path?: string
}

interface PageHeaderProps {
  title: string
  subtitle?: string
  breadcrumbs?: BreadcrumbItem[]
  extra?: ReactNode
}

const PageHeader = ({
  title,
  subtitle,
  breadcrumbs,
  extra,
}: PageHeaderProps) => {
  const navigate = useNavigate()

  const defaultBreadcrumbs: BreadcrumbItem[] = [
    { title: '首页', path: '/dashboard' },
  ]

  const allBreadcrumbs = breadcrumbs
    ? [...defaultBreadcrumbs, ...breadcrumbs]
    : defaultBreadcrumbs

  return (
    <div style={{ marginBottom: 24 }}>
      {/* 面包屑 */}
      {allBreadcrumbs.length > 1 && (
        <Breadcrumb style={{ marginBottom: 16 }}>
          {allBreadcrumbs.map((item, index) => (
            <Breadcrumb.Item
              key={index}
              onClick={() => item.path && navigate(item.path)}
            >
              {index === 0 && <HomeOutlined style={{ marginRight: 4 }} />}
              {item.title}
            </Breadcrumb.Item>
          ))}
        </Breadcrumb>
      )}

      {/* 标题和操作 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space direction="vertical" size={4}>
          <Title level={2} style={{ margin: 0 }}>
            {title}
          </Title>
          {subtitle && (
            <span style={{ color: '#999', fontSize: 14 }}>{subtitle}</span>
          )}
        </Space>
        {extra && <Space>{extra}</Space>}
      </div>
    </div>
  )
}

export default PageHeader

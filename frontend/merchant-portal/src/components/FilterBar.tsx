import { Card, Space, Button } from 'antd'
import { FilterOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ReactNode } from 'react'

interface FilterBarProps {
  children: ReactNode
  onReset?: () => void
  onFilter?: () => void
  showResetButton?: boolean
  showFilterButton?: boolean
  extra?: ReactNode
}

const FilterBar = ({
  children,
  onReset,
  onFilter,
  showResetButton = true,
  showFilterButton = true,
  extra,
}: FilterBarProps) => {
  return (
    <Card size="small" style={{ marginBottom: 16 }}>
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 16 }}>
          {children}
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            {showFilterButton && onFilter && (
              <Button type="primary" icon={<FilterOutlined />} onClick={onFilter}>
                筛选
              </Button>
            )}
            {showResetButton && onReset && (
              <Button icon={<ReloadOutlined />} onClick={onReset}>
                重置
              </Button>
            )}
          </Space>
          {extra && <Space>{extra}</Space>}
        </div>
      </Space>
    </Card>
  )
}

export default FilterBar

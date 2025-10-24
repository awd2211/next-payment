import { Card, Statistic, Typography } from 'antd'
import type { ReactNode } from 'react'

const { Text } = Typography

interface StatCardProps {
  title: string
  value: number | string
  prefix?: ReactNode
  suffix?: string
  precision?: number
  valueStyle?: React.CSSProperties
  formatter?: (value: number | string) => string
  extra?: string
  loading?: boolean
  onClick?: () => void
}

const StatCard = ({
  title,
  value,
  prefix,
  suffix,
  precision,
  valueStyle,
  formatter,
  extra,
  loading,
  onClick,
}: StatCardProps) => {
  return (
    <Card
      loading={loading}
      hoverable={!!onClick}
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
    >
      <Statistic
        title={title}
        value={value}
        prefix={prefix}
        suffix={suffix}
        precision={precision}
        valueStyle={valueStyle}
        formatter={formatter}
      />
      {extra && (
        <Text type="secondary" style={{ fontSize: 12, marginTop: 8, display: 'block' }}>
          {extra}
        </Text>
      )}
    </Card>
  )
}

export default StatCard

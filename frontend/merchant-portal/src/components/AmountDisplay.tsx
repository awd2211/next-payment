import { Typography } from 'antd'
import { DollarOutlined } from '@ant-design/icons'

const { Text } = Typography

interface AmountDisplayProps {
  amount: number // 金额(分)
  currency?: string
  showCurrency?: boolean
  showIcon?: boolean
  type?: 'success' | 'warning' | 'danger' | 'secondary'
  size?: 'small' | 'default' | 'large'
  strong?: boolean
}

const AmountDisplay = ({
  amount,
  currency = 'USD',
  showCurrency = true,
  showIcon = false,
  type,
  size = 'default',
  strong,
}: AmountDisplayProps) => {
  // 转换分为主货币单位
  const mainAmount = amount / 100

  // 格式化金额
  const formattedAmount = new Intl.NumberFormat('en-US', {
    style: showCurrency ? 'currency' : 'decimal',
    currency: currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(mainAmount)

  // 根据size设置字体大小
  const fontSize = {
    small: 12,
    default: 14,
    large: 18,
  }[size]

  return (
    <Text
      type={type}
      strong={strong}
      style={{ fontSize }}
    >
      {showIcon && <DollarOutlined style={{ marginRight: 4 }} />}
      {formattedAmount}
    </Text>
  )
}

export default AmountDisplay

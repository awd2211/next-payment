import { useState, useEffect } from 'react'
import { Card, Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'

interface FeeConfig {
  id: string
  channel: string
  payment_method: string
  percentage_fee: number
  fixed_fee: number
  currency: string
  status: string
}

export default function FeeConfigs() {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<FeeConfig[]>([])

  useEffect(() => {
    setData([
      {
        id: '1',
        channel: 'Stripe',
        payment_method: 'card',
        percentage_fee: 2.9,
        fixed_fee: 30,
        currency: 'USD',
        status: 'active',
      },
    ])
  }, [])

  const columns: ColumnsType<FeeConfig> = [
    { title: '支付渠道', dataIndex: 'channel' },
    { title: '支付方式', dataIndex: 'payment_method' },
    {
      title: '费率',
      render: (_, record) => `${record.percentage_fee}% + ${(record.fixed_fee / 100).toFixed(2)} ${record.currency}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status) => <Tag color="green">{status}</Tag>,
    },
  ]

  return (
    <Card title="费率配置">
      <Table columns={columns} dataSource={data} loading={loading} rowKey="id" />
    </Card>
  )
}

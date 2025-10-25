/**
 * 骨架屏组件
 * 提供不同类型的加载骨架屏
 */
import { Card, Skeleton, Space } from 'antd'

interface SkeletonLoadingProps {
  type?: 'table' | 'card' | 'form' | 'dashboard' | 'detail'
  rows?: number
}

const SkeletonLoading: React.FC<SkeletonLoadingProps> = ({ type = 'card', rows = 5 }) => {
  switch (type) {
    case 'table':
      return (
        <Card>
          <Skeleton active paragraph={{ rows: rows }} />
        </Card>
      )

    case 'card':
      return (
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          {Array.from({ length: rows }).map((_, index) => (
            <Card key={index}>
              <Skeleton active />
            </Card>
          ))}
        </Space>
      )

    case 'form':
      return (
        <Card>
          <Skeleton.Input active style={{ width: '100%', marginBottom: 16 }} />
          <Skeleton.Input active style={{ width: '100%', marginBottom: 16 }} />
          <Skeleton.Input active style={{ width: '100%', marginBottom: 16 }} />
          <Skeleton.Button active style={{ width: 100 }} />
        </Card>
      )

    case 'dashboard':
      return (
        <div>
          {/* 统计卡片骨架 */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16, marginBottom: 24 }}>
            {Array.from({ length: 4 }).map((_, index) => (
              <Card key={index}>
                <Skeleton active paragraph={{ rows: 1 }} />
              </Card>
            ))}
          </div>

          {/* 图表骨架 */}
          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 16 }}>
            <Card>
              <Skeleton.Node active style={{ width: '100%', height: 300 }}>
                <div />
              </Skeleton.Node>
            </Card>
            <Card>
              <Skeleton.Node active style={{ width: '100%', height: 300 }}>
                <div />
              </Skeleton.Node>
            </Card>
          </div>
        </div>
      )

    case 'detail':
      return (
        <Card>
          <Skeleton active paragraph={{ rows: 8 }} />
        </Card>
      )

    default:
      return <Skeleton active />
  }
}

export default SkeletonLoading

/**
 * 页面加载组件
 * 用于 React.lazy 的 Suspense fallback
 */
import { Spin } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'

interface PageLoadingProps {
  tip?: string
  size?: 'small' | 'default' | 'large'
}

const PageLoading: React.FC<PageLoadingProps> = ({ tip = '加载中...', size = 'large' }) => {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '400px',
        width: '100%',
      }}
    >
      <Spin
        indicator={<LoadingOutlined style={{ fontSize: size === 'large' ? 48 : 24 }} spin />}
        tip={tip}
        size={size}
      />
    </div>
  )
}

export default PageLoading

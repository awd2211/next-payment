import { Spin } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import type { ReactNode } from 'react'

interface LoadingProps {
  /** 是否全屏加载 */
  fullscreen?: boolean
  /** 提示文本 */
  tip?: string
  /** 背景遮罩 */
  overlay?: boolean
  /** 加载中状态 */
  spinning?: boolean
  /** 子内容 */
  children?: ReactNode
  /** 大小 */
  size?: 'small' | 'default' | 'large'
}

/**
 * 加载组件 - 统一的加载状态展示
 *
 * @example
 * // 基础用法
 * <Loading />
 *
 * // 带提示文本
 * <Loading tip="加载中..." />
 *
 * // 全屏加载
 * <Loading fullscreen tip="处理中..." />
 *
 * // 局部加载(包裹内容)
 * <Loading spinning={loading}>
 *   <YourContent />
 * </Loading>
 */
const Loading = ({
  fullscreen = false,
  tip = '加载中...',
  overlay = true,
  size = 'large',
  spinning = true,
  children,
}: LoadingProps) => {
  const defaultIndicator = <LoadingOutlined style={{ fontSize: 48 }} spin />

  if (fullscreen) {
    return (
      <div
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: overlay ? 'rgba(255, 255, 255, 0.8)' : 'transparent',
          zIndex: 9999,
        }}
      >
        <Spin
          indicator={defaultIndicator}
          tip={tip}
          size={size}
          spinning={spinning}
        />
      </div>
    )
  }

  // 如果有children,使用Spin包裹内容
  if (children) {
    return (
      <Spin indicator={defaultIndicator} tip={tip} size={size} spinning={spinning}>
        {children}
      </Spin>
    )
  }

  return (
    <div
      style={{
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '50px 0',
      }}
    >
      <Spin
        indicator={defaultIndicator}
        tip={tip}
        size={size}
        spinning={spinning}
      />
    </div>
  )
}

export default Loading

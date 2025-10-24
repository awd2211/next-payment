import { Component, ReactNode, ErrorInfo } from 'react'
import { Button, Result } from 'antd'

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

interface ErrorBoundaryState {
  hasError: boolean
  error?: Error
  errorInfo?: ErrorInfo
}

/**
 * 错误边界组件 - 捕获子组件错误
 *
 * @example
 * <ErrorBoundary
 *   onError={(error, errorInfo) => {
 *     // 上报错误到监控系统
 *     console.error('Component error:', error, errorInfo)
 *   }}
 * >
 *   <YourComponent />
 * </ErrorBoundary>
 */
class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = {
      hasError: false,
    }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    // 更新 state 使下一次渲染能够显示降级后的 UI
    return {
      hasError: true,
      error,
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // 记录错误到错误报告服务
    console.error('ErrorBoundary caught an error:', error, errorInfo)

    this.setState({
      error,
      errorInfo,
    })

    // 调用自定义错误处理
    this.props.onError?.(error, errorInfo)

    // 可以在这里上报到监控系统
    // reportErrorToService(error, errorInfo)
  }

  handleReset = (): void => {
    this.setState({
      hasError: false,
      error: undefined,
      errorInfo: undefined,
    })
  }

  render(): ReactNode {
    if (this.state.hasError) {
      // 如果有自定义降级 UI,渲染它
      if (this.props.fallback) {
        return this.props.fallback
      }

      // 默认错误 UI
      return (
        <div style={{ padding: '50px', textAlign: 'center' }}>
          <Result
            status="error"
            title="页面出错了"
            subTitle="抱歉，页面遇到了一些问题。您可以尝试刷新页面或返回首页。"
            extra={[
              <Button type="primary" key="refresh" onClick={this.handleReset}>
                重新加载
              </Button>,
              <Button key="home" onClick={() => (window.location.href = '/dashboard')}>
                返回首页
              </Button>,
            ]}
          >
            {process.env.NODE_ENV === 'development' && this.state.error && (
              <div
                style={{
                  textAlign: 'left',
                  padding: '20px',
                  background: '#f5f5f5',
                  borderRadius: '4px',
                  marginTop: '20px',
                }}
              >
                <h3>错误详情(开发模式):</h3>
                <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                  {this.state.error.toString()}
                  {this.state.errorInfo?.componentStack}
                </pre>
              </div>
            )}
          </Result>
        </div>
      )
    }

    return this.props.children
  }
}

export default ErrorBoundary

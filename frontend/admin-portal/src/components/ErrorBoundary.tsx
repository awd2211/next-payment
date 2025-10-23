import React from 'react'
import { Result, Button } from 'antd'

interface Props {
  children: React.ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: React.ErrorInfo | null
}

/**
 * 错误边界组件
 * 捕获子组件树中的 JavaScript 错误，记录错误并展示降级 UI
 */
export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    // 更新 state，下次渲染将显示降级 UI
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // 记录错误信息
    console.error('Error caught by ErrorBoundary:', error, errorInfo)

    this.setState({
      error,
      errorInfo,
    })

    // 生产环境：上报错误到监控系统
    if (import.meta.env.PROD) {
      this.reportErrorToService(error, errorInfo)
    }
  }

  reportErrorToService(error: Error, errorInfo: React.ErrorInfo) {
    // TODO: 集成错误监控服务（如 Sentry、Bugsnag 等）
    try {
      const errorData = {
        message: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
        timestamp: new Date().toISOString(),
        url: window.location.href,
        userAgent: navigator.userAgent,
      }

      // 发送到错误监控服务
      fetch('/api/v1/errors/report', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(errorData),
      }).catch((err) => {
        console.error('Failed to report error:', err)
      })
    } catch (reportError) {
      console.error('Error in error reporting:', reportError)
    }
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    })
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: '50px', maxWidth: '600px', margin: '0 auto' }}>
          <Result
            status="error"
            title="页面出错了"
            subTitle="抱歉，页面发生了错误。您可以刷新页面或返回首页。"
            extra={[
              <Button type="primary" key="refresh" onClick={() => window.location.reload()}>
                刷新页面
              </Button>,
              <Button key="home" onClick={() => (window.location.href = '/')}>
                返回首页
              </Button>,
              <Button key="retry" onClick={this.handleReset}>
                重试
              </Button>,
            ]}
          >
            {/* 开发环境显示详细错误信息 */}
            {import.meta.env.DEV && this.state.error && (
              <div style={{ textAlign: 'left', marginTop: '20px' }}>
                <details style={{ whiteSpace: 'pre-wrap' }}>
                  <summary style={{ cursor: 'pointer', fontWeight: 'bold', marginBottom: '10px' }}>
                    错误详情 (仅开发环境显示)
                  </summary>
                  <div style={{ padding: '10px', background: '#f5f5f5', borderRadius: '4px' }}>
                    <p>
                      <strong>错误信息:</strong> {this.state.error.message}
                    </p>
                    {this.state.error.stack && (
                      <p>
                        <strong>错误堆栈:</strong>
                        <br />
                        {this.state.error.stack}
                      </p>
                    )}
                    {this.state.errorInfo?.componentStack && (
                      <p>
                        <strong>组件堆栈:</strong>
                        <br />
                        {this.state.errorInfo.componentStack}
                      </p>
                    )}
                  </div>
                </details>
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

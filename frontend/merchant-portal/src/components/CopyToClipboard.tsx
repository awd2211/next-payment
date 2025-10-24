import { message, Tooltip } from 'antd'
import { CopyOutlined, CheckOutlined } from '@ant-design/icons'
import { useState } from 'react'

interface CopyToClipboardProps {
  text: string
  successMessage?: string
  children?: React.ReactNode
  onSuccess?: () => void
  onError?: (error: Error) => void
}

/**
 * 复制到剪贴板组件
 *
 * @example
 * // 基础用法
 * <CopyToClipboard text="要复制的内容" />
 *
 * // 自定义触发元素
 * <CopyToClipboard text={apiKey}>
 *   <Button icon={<CopyOutlined />}>复制API Key</Button>
 * </CopyToClipboard>
 *
 * // 自定义成功消息
 * <CopyToClipboard
 *   text={merchantId}
 *   successMessage="商户ID已复制"
 *   onSuccess={() => console.log('复制成功')}
 * />
 */
const CopyToClipboard = ({
  text,
  successMessage = '复制成功',
  children,
  onSuccess,
  onError,
}: CopyToClipboardProps) => {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      // 使用现代 Clipboard API
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text)
      } else {
        // 降级方案:使用 document.execCommand
        const textArea = document.createElement('textarea')
        textArea.value = text
        textArea.style.position = 'fixed'
        textArea.style.left = '-999999px'
        textArea.style.top = '-999999px'
        document.body.appendChild(textArea)
        textArea.focus()
        textArea.select()

        const successful = document.execCommand('copy')
        textArea.remove()

        if (!successful) {
          throw new Error('复制失败')
        }
      }

      setCopied(true)
      message.success(successMessage)
      onSuccess?.()

      // 2秒后重置状态
      setTimeout(() => {
        setCopied(false)
      }, 2000)
    } catch (error) {
      const err = error instanceof Error ? error : new Error('复制失败')
      message.error(err.message)
      onError?.(err)
    }
  }

  // 如果有自定义子元素,包裹点击事件
  if (children) {
    return (
      <span onClick={handleCopy} style={{ cursor: 'pointer' }}>
        {children}
      </span>
    )
  }

  // 默认渲染复制图标
  return (
    <Tooltip title={copied ? '已复制' : '点击复制'}>
      <span
        onClick={handleCopy}
        style={{
          cursor: 'pointer',
          color: copied ? '#52c41a' : '#1890ff',
          transition: 'color 0.3s',
        }}
      >
        {copied ? <CheckOutlined /> : <CopyOutlined />}
      </span>
    </Tooltip>
  )
}

export default CopyToClipboard

/**
 * 通用 Modal 组件
 * 统一 Modal 的样式和行为
 */
import { Modal, Form, message } from 'antd'
import type { ModalProps, FormInstance } from 'antd'
import { useState, useEffect } from 'react'

export interface CommonModalProps extends Omit<ModalProps, 'onOk' | 'onCancel'> {
  /**
   * Modal 类型
   */
  type?: 'create' | 'edit' | 'view' | 'custom'

  /**
   * 表单实例
   */
  form?: FormInstance

  /**
   * 初始值
   */
  initialValues?: Record<string, any>

  /**
   * 提交回调
   */
  onSubmit?: (values: any) => Promise<void> | void

  /**
   * 取消回调
   */
  onCancel?: () => void

  /**
   * 提交成功后是否自动关闭
   */
  autoClose?: boolean

  /**
   * 提交成功提示
   */
  successMessage?: string

  /**
   * 提交失败提示
   */
  errorMessage?: string

  /**
   * 是否显示
   */
  visible: boolean

  /**
   * 宽度
   */
  width?: number | string

  /**
   * 子组件
   */
  children?: React.ReactNode
}

const CommonModal: React.FC<CommonModalProps> = ({
  type = 'custom',
  form: externalForm,
  initialValues,
  onSubmit,
  onCancel,
  autoClose = true,
  successMessage,
  errorMessage,
  visible,
  width = 600,
  children,
  title,
  ...restProps
}) => {
  const [internalForm] = Form.useForm()
  const form = externalForm || internalForm
  const [loading, setLoading] = useState(false)
  const [internalVisible, setInternalVisible] = useState(visible)

  useEffect(() => {
    setInternalVisible(visible)
  }, [visible])

  useEffect(() => {
    if (visible && initialValues) {
      form.setFieldsValue(initialValues)
    }
  }, [visible, initialValues, form])

  /**
   * 获取默认标题
   */
  const getDefaultTitle = () => {
    const titles = {
      create: '新建',
      edit: '编辑',
      view: '查看详情',
      custom: '',
    }
    return title || titles[type]
  }

  /**
   * 处理确认
   */
  const handleOk = async () => {
    if (!onSubmit) {
      handleClose()
      return
    }

    if (type === 'view') {
      handleClose()
      return
    }

    setLoading(true)
    try {
      const values = await form.validateFields()
      await onSubmit(values)

      message.success(successMessage || '操作成功')

      if (autoClose) {
        handleClose()
        form.resetFields()
      }
    } catch (error: any) {
      // 如果是表单验证错误,不显示错误消息
      if (error.errorFields) {
        return
      }

      const errMsg = error?.message || error?.response?.data?.message || errorMessage || '操作失败'
      message.error(errMsg)
    } finally {
      setLoading(false)
    }
  }

  /**
   * 处理取消
   */
  const handleClose = () => {
    setInternalVisible(false)
    form.resetFields()
    onCancel?.()
  }

  /**
   * Modal 关闭后的回调
   */
  const handleAfterClose = () => {
    form.resetFields()
  }

  return (
    <Modal
      title={getDefaultTitle()}
      open={internalVisible}
      onOk={handleOk}
      onCancel={handleClose}
      confirmLoading={loading}
      width={width}
      destroyOnClose
      afterClose={handleAfterClose}
      okText={type === 'view' ? '关闭' : '确定'}
      cancelButtonProps={{ style: type === 'view' ? { display: 'none' } : {} }}
      {...restProps}
    >
      {children}
    </Modal>
  )
}

export default CommonModal

import { Modal, ModalProps } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'

interface ConfirmModalConfig extends Omit<ModalProps, 'open' | 'width'> {
  title?: string
  content?: string
  onOk?: () => void | Promise<void>
  onCancel?: () => void
  okText?: string
  cancelText?: string
  danger?: boolean
  width?: string | number
}

/**
 * 确认对话框工具函数
 *
 * @example
 * // 删除确认
 * confirmModal({
 *   title: '确认删除',
 *   content: '删除后无法恢复，确定要删除吗?',
 *   danger: true,
 *   onOk: async () => {
 *     await deleteItem(id)
 *     message.success('删除成功')
 *   }
 * })
 *
 * // 提交确认
 * confirmModal({
 *   title: '确认提交',
 *   content: '确定要提交该表单吗?',
 *   onOk: async () => {
 *     await submitForm(data)
 *   }
 * })
 */
export const confirmModal = (config: ConfirmModalConfig): void => {
  Modal.confirm({
    title: config.title || '确认操作',
    icon: <ExclamationCircleOutlined />,
    content: config.content,
    okText: config.okText || '确定',
    cancelText: config.cancelText || '取消',
    okButtonProps: {
      danger: config.danger,
      ...config.okButtonProps,
    },
    onOk: config.onOk,
    onCancel: config.onCancel,
    ...config,
  })
}

/**
 * 删除确认(预设)
 */
export const confirmDelete = (
  onOk: () => void | Promise<void>,
  content = '删除后无法恢复，确定要删除吗?'
): void => {
  confirmModal({
    title: '确认删除',
    content,
    danger: true,
    okText: '删除',
    onOk,
  })
}

/**
 * 批量删除确认(预设)
 */
export const confirmBatchDelete = (
  count: number,
  onOk: () => void | Promise<void>
): void => {
  confirmModal({
    title: '批量删除确认',
    content: `即将删除 ${count} 条记录，删除后无法恢复，确定要继续吗?`,
    danger: true,
    okText: '批量删除',
    onOk,
  })
}

/**
 * 提交确认(预设)
 */
export const confirmSubmit = (
  onOk: () => void | Promise<void>,
  content = '确定要提交吗?'
): void => {
  confirmModal({
    title: '确认提交',
    content,
    onOk,
  })
}

/**
 * 离开确认(预设)
 */
export const confirmLeave = (
  onOk: () => void | Promise<void>,
  content = '有未保存的更改，确定要离开吗?'
): void => {
  confirmModal({
    title: '确认离开',
    content,
    okText: '离开',
    danger: true,
    onOk,
  })
}

/**
 * 操作确认(预设)
 */
export const confirmAction = (
  title: string,
  content: string,
  onOk: () => void | Promise<void>,
  danger = false
): void => {
  confirmModal({
    title,
    content,
    danger,
    onOk,
  })
}

export default confirmModal

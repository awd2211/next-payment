/**
 * 批量操作组件
 * 支持批量审核、删除、导出等操作
 */
import { useState } from 'react'
import { Space, Button, Dropdown, message, Modal, Popconfirm } from 'antd'
import type { MenuProps } from 'antd'
import {
  CheckOutlined,
  CloseOutlined,
  DeleteOutlined,
  DownloadOutlined,
  MoreOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

export interface BatchAction {
  key: string
  label: string
  icon?: React.ReactNode
  danger?: boolean
  needConfirm?: boolean
  confirmMessage?: string
  disabled?: boolean
  onClick: (selectedKeys: React.Key[]) => Promise<void> | void
}

export interface BatchActionsProps {
  /**
   * 选中的行数
   */
  selectedCount: number

  /**
   * 选中的行 keys
   */
  selectedRowKeys: React.Key[]

  /**
   * 批量操作配置
   */
  actions: BatchAction[]

  /**
   * 最大显示按钮数,超过则放入下拉菜单
   */
  maxVisibleActions?: number

  /**
   * 是否显示清空按钮
   */
  showClear?: boolean

  /**
   * 清空选择回调
   */
  onClear?: () => void
}

const BatchActions: React.FC<BatchActionsProps> = ({
  selectedCount,
  selectedRowKeys,
  actions,
  maxVisibleActions = 3,
  showClear = true,
  onClear,
}) => {
  const { t } = useTranslation()
  const [loading, setLoading] = useState<string | null>(null)

  /**
   * 执行操作
   */
  const handleAction = async (action: BatchAction) => {
    if (selectedCount === 0) {
      message.warning('请先选择数据')
      return
    }

    if (action.disabled) {
      return
    }

    setLoading(action.key)
    try {
      await action.onClick(selectedRowKeys)
      message.success('操作成功')
    } catch (error: any) {
      message.error(error.message || '操作失败')
    } finally {
      setLoading(null)
    }
  }

  /**
   * 带确认的操作
   */
  const handleActionWithConfirm = (action: BatchAction) => {
    if (selectedCount === 0) {
      message.warning('请先选择数据')
      return
    }

    Modal.confirm({
      title: '确认操作',
      content: action.confirmMessage || `确定要对选中的 ${selectedCount} 条数据执行"${action.label}"操作吗?`,
      onOk: () => handleAction(action),
      okText: '确认',
      cancelText: '取消',
      okButtonProps: action.danger ? { danger: true } : undefined,
    })
  }

  /**
   * 渲染操作按钮
   */
  const renderActionButton = (action: BatchAction) => {
    const button = (
      <Button
        key={action.key}
        icon={action.icon}
        danger={action.danger}
        disabled={action.disabled || selectedCount === 0}
        loading={loading === action.key}
        onClick={() => {
          if (action.needConfirm) {
            handleActionWithConfirm(action)
          } else {
            handleAction(action)
          }
        }}
      >
        {action.label}
      </Button>
    )

    // 如果需要确认且不用 Modal,使用 Popconfirm
    if (action.needConfirm && !action.confirmMessage) {
      return (
        <Popconfirm
          key={action.key}
          title={`确定要${action.label}吗?`}
          onConfirm={() => handleAction(action)}
          disabled={selectedCount === 0}
        >
          {button}
        </Popconfirm>
      )
    }

    return button
  }

  // 分离主要操作和次要操作
  const visibleActions = actions.slice(0, maxVisibleActions)
  const moreActions = actions.slice(maxVisibleActions)

  // 构建下拉菜单
  const menuItems: MenuProps['items'] = moreActions.map((action) => ({
    key: action.key,
    label: action.label,
    icon: action.icon,
    danger: action.danger,
    disabled: action.disabled || selectedCount === 0,
    onClick: () => {
      if (action.needConfirm) {
        handleActionWithConfirm(action)
      } else {
        handleAction(action)
      }
    },
  }))

  if (!actions || actions.length === 0) {
    return null
  }

  return (
    <div style={{ marginBottom: 16, display: 'flex', alignItems: 'center', gap: 16 }}>
      <span>
        已选择 <strong style={{ color: '#1890ff' }}>{selectedCount}</strong> 项
      </span>

      <Space>
        {/* 主要操作按钮 */}
        {visibleActions.map(renderActionButton)}

        {/* 更多操作下拉菜单 */}
        {moreActions.length > 0 && (
          <Dropdown menu={{ items: menuItems }} disabled={selectedCount === 0}>
            <Button icon={<MoreOutlined />}>更多操作</Button>
          </Dropdown>
        )}

        {/* 清空选择按钮 */}
        {showClear && selectedCount > 0 && (
          <Button type="link" onClick={onClear}>
            清空选择
          </Button>
        )}
      </Space>
    </div>
  )
}

export default BatchActions

/**
 * 常用批量操作配置
 */
export const commonBatchActions = {
  /**
   * 批量审核通过
   */
  approve: (onApprove: (keys: React.Key[]) => Promise<void>): BatchAction => ({
    key: 'approve',
    label: '批量审核通过',
    icon: <CheckOutlined />,
    needConfirm: true,
    confirmMessage: '确定要批量审核通过选中的数据吗?',
    onClick: onApprove,
  }),

  /**
   * 批量拒绝
   */
  reject: (onReject: (keys: React.Key[]) => Promise<void>): BatchAction => ({
    key: 'reject',
    label: '批量拒绝',
    icon: <CloseOutlined />,
    danger: true,
    needConfirm: true,
    onClick: onReject,
  }),

  /**
   * 批量删除
   */
  delete: (onDelete: (keys: React.Key[]) => Promise<void>): BatchAction => ({
    key: 'delete',
    label: '批量删除',
    icon: <DeleteOutlined />,
    danger: true,
    needConfirm: true,
    confirmMessage: '删除后无法恢复,确定要批量删除选中的数据吗?',
    onClick: onDelete,
  }),

  /**
   * 批量导出
   */
  export: (onExport: (keys: React.Key[]) => void): BatchAction => ({
    key: 'export',
    label: '导出选中',
    icon: <DownloadOutlined />,
    onClick: onExport,
  }),
}

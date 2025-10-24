import { Space, Button, Dropdown, Popconfirm } from 'antd'
import type { MenuProps } from 'antd'
import {
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  MoreOutlined,
  CopyOutlined,
  DownloadOutlined,
} from '@ant-design/icons'

export interface Action {
  key: string
  label: string
  icon?: React.ReactNode
  danger?: boolean
  disabled?: boolean
  confirm?: {
    title: string
    description?: string
  }
  onClick: () => void
}

interface ActionButtonsProps {
  actions: Action[]
  maxVisible?: number
  size?: 'small' | 'middle' | 'large'
}

const ActionButtons = ({ actions, maxVisible = 3, size = 'small' }: ActionButtonsProps) => {
  if (actions.length === 0) return null

  // 显示的按钮
  const visibleActions = actions.slice(0, maxVisible)
  // 更多菜单中的操作
  const moreActions = actions.slice(maxVisible)

  const renderButton = (action: Action) => {
    const button = (
      <Button
        key={action.key}
        type={action.key === 'delete' ? 'primary' : 'default'}
        danger={action.danger}
        disabled={action.disabled}
        icon={action.icon}
        size={size}
        onClick={action.onClick}
      >
        {action.label}
      </Button>
    )

    if (action.confirm) {
      return (
        <Popconfirm
          key={action.key}
          title={action.confirm.title}
          description={action.confirm.description}
          onConfirm={action.onClick}
          okText="确认"
          cancelText="取消"
        >
          {button}
        </Popconfirm>
      )
    }

    return button
  }

  const moreMenuItems: MenuProps['items'] = moreActions.map(action => ({
    key: action.key,
    label: action.label,
    icon: action.icon,
    danger: action.danger,
    disabled: action.disabled,
    onClick: action.onClick,
  }))

  return (
    <Space size="small">
      {visibleActions.map(renderButton)}
      {moreActions.length > 0 && (
        <Dropdown menu={{ items: moreMenuItems }} placement="bottomRight">
          <Button icon={<MoreOutlined />} size={size} />
        </Dropdown>
      )}
    </Space>
  )
}

// 预定义的常用操作
export const commonActions = {
  view: (onClick: () => void): Action => ({
    key: 'view',
    label: '查看',
    icon: <EyeOutlined />,
    onClick,
  }),

  edit: (onClick: () => void): Action => ({
    key: 'edit',
    label: '编辑',
    icon: <EditOutlined />,
    onClick,
  }),

  delete: (onClick: () => void, confirmTitle = '确认删除?'): Action => ({
    key: 'delete',
    label: '删除',
    icon: <DeleteOutlined />,
    danger: true,
    confirm: {
      title: confirmTitle,
      description: '删除后无法恢复',
    },
    onClick,
  }),

  copy: (onClick: () => void): Action => ({
    key: 'copy',
    label: '复制',
    icon: <CopyOutlined />,
    onClick,
  }),

  download: (onClick: () => void): Action => ({
    key: 'download',
    label: '下载',
    icon: <DownloadOutlined />,
    onClick,
  }),
}

export default ActionButtons

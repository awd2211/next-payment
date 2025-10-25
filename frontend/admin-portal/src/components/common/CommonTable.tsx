/**
 * 通用表格组件
 * 集成分页、筛选、排序、导出等功能
 */
import { Table, Button, Space, Tooltip } from 'antd'
import { ReloadOutlined, DownloadOutlined } from '@ant-design/icons'
import type { TableProps, TablePaginationConfig } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useTranslation } from 'react-i18next'

export interface CommonTableProps<T = any> extends Omit<TableProps<T>, 'title'> {
  /**
   * 表格列配置
   */
  columns: ColumnsType<T>

  /**
   * 表格数据
   */
  dataSource: T[]

  /**
   * 加载状态
   */
  loading?: boolean

  /**
   * 分页配置
   */
  pagination?: TablePaginationConfig | false

  /**
   * 是否显示刷新按钮
   */
  showRefresh?: boolean

  /**
   * 是否显示导出按钮
   */
  showExport?: boolean

  /**
   * 刷新回调
   */
  onRefresh?: () => void

  /**
   * 导出回调
   */
  onExport?: () => void

  /**
   * 表格标题
   */
  title?: string

  /**
   * 自定义工具栏
   */
  toolbarExtra?: React.ReactNode
}

function CommonTable<T extends Record<string, any> = any>(
  props: CommonTableProps<T>
) {
  const {
    columns,
    dataSource,
    loading,
    pagination,
    showRefresh = true,
    showExport = true,
    onRefresh,
    onExport,
    title,
    toolbarExtra,
    ...restProps
  } = props

  const { t } = useTranslation()

  /**
   * 工具栏
   */
  const renderToolbar = () => {
    if (!showRefresh && !showExport && !toolbarExtra) {
      return null
    }

    return (
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        {title && <h3 style={{ margin: 0 }}>{title}</h3>}
        <Space>
          {toolbarExtra}
          {showRefresh && (
            <Tooltip title={t('common.refresh') || '刷新'}>
              <Button
                icon={<ReloadOutlined />}
                onClick={onRefresh}
                loading={loading}
              />
            </Tooltip>
          )}
          {showExport && (
            <Tooltip title={t('common.export') || '导出'}>
              <Button
                icon={<DownloadOutlined />}
                onClick={onExport}
                disabled={!dataSource || dataSource.length === 0}
              />
            </Tooltip>
          )}
        </Space>
      </div>
    )
  }

  /**
   * 默认分页配置
   */
  const defaultPagination: TablePaginationConfig = {
    showSizeChanger: true,
    showQuickJumper: true,
    showTotal: (total) => `${t('common.total') || '共'} ${total} ${t('common.items') || '条'}`,
    pageSizeOptions: ['10', '20', '50', '100'],
    ...pagination,
  }

  return (
    <>
      {renderToolbar()}
      <Table<T>
        columns={columns}
        dataSource={dataSource}
        loading={loading}
        pagination={pagination === false ? false : defaultPagination}
        scroll={{ x: 'max-content' }}
        bordered
        {...restProps}
      />
    </>
  )
}

export default CommonTable

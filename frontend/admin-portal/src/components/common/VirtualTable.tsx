/**
 * 虚拟滚动表格组件
 * 优化大数据量表格的渲染性能
 */
import { useMemo, useRef, useState, useEffect } from 'react'
import { Table } from 'antd'
import type { TableProps } from 'antd'
// @ts-ignore - react-window types compatibility
import { FixedSizeGrid as Grid } from 'react-window'

export interface VirtualTableProps<T = any> extends TableProps<T> {
  /**
   * 行高
   */
  rowHeight?: number

  /**
   * 可视区域高度
   */
  height?: number

  /**
   * 启用虚拟滚动
   */
  virtual?: boolean
}

/**
 * 虚拟滚动表格
 * 适用于大数据量场景 (>1000 行)
 */
function VirtualTable<T extends Record<string, any> = any>(props: VirtualTableProps<T>) {
  const { rowHeight = 54, height = 600, virtual = true, dataSource = [], columns = [], ...restProps } = props

  const [tableWidth, setTableWidth] = useState(0)
  const tableRef = useRef<HTMLDivElement>(null)

  // 计算列宽
  const columnWidths = useMemo(() => {
    return columns.map((col: any) => {
      if (typeof col.width === 'number') {
        return col.width
      }
      return 150 // 默认宽度
    })
  }, [columns])

  // 更新表格宽度
  useEffect(() => {
    if (tableRef.current) {
      setTableWidth(tableRef.current.offsetWidth)
    }

    const handleResize = () => {
      if (tableRef.current) {
        setTableWidth(tableRef.current.offsetWidth)
      }
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  // 如果数据量不大或不启用虚拟滚动,使用普通表格
  if (!virtual || dataSource.length < 100) {
    return <Table<T> {...restProps} dataSource={dataSource} columns={columns} scroll={{ y: height }} />
  }

  // 使用虚拟滚动
  const gridRef = useRef<any>()

  const VirtualBody = (props: any) => {
    const { children } = props
    const [thead] = children

    return (
      <div ref={tableRef}>
        {thead}
        <Grid
          ref={gridRef}
          className="virtual-grid"
          columnCount={columns.length}
          columnWidth={(index: number) => columnWidths[index] || 150}
          height={height}
          rowCount={dataSource.length}
          rowHeight={() => rowHeight}
          width={tableWidth}
        >
          {({ columnIndex, rowIndex, style }: { columnIndex: number; rowIndex: number; style: React.CSSProperties }) => {
            const record = dataSource[rowIndex]
            const column = columns[columnIndex] as any
            const value = record[column.dataIndex]

            return (
              <div
                style={{
                  ...style,
                  display: 'flex',
                  alignItems: 'center',
                  padding: '0 16px',
                  borderBottom: '1px solid #f0f0f0',
                  borderRight: '1px solid #f0f0f0',
                }}
              >
                {column.render ? column.render(value, record, rowIndex) : value}
              </div>
            )
          }}
        </Grid>
      </div>
    )
  }

  return (
    <Table<T>
      {...restProps}
      dataSource={dataSource}
      columns={columns}
      pagination={false}
      components={{
        body: VirtualBody,
      }}
    />
  )
}

export default VirtualTable

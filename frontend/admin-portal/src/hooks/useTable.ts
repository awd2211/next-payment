/**
 * 通用表格 Hook
 * 统一管理表格的分页、筛选、排序等状态
 */
import { useState, useCallback, useMemo } from 'react'
import type { TablePaginationConfig } from 'antd'
import type { FilterValue, SorterResult } from 'antd/es/table/interface'

export interface TableFilters {
  [key: string]: any
}

export interface UseTableOptions<T = any> {
  initialPageSize?: number
  initialFilters?: TableFilters
  onFetchData?: (params: {
    page: number
    pageSize: number
    filters: TableFilters
    sorter?: SorterResult<T>
  }) => Promise<{ data: T[]; total: number }>
}

export interface UseTableReturn<T = any> {
  // 数据状态
  data: T[]
  total: number
  loading: boolean

  // 分页状态
  pagination: TablePaginationConfig
  page: number
  pageSize: number

  // 筛选状态
  filters: TableFilters
  setFilters: (filters: TableFilters) => void
  updateFilter: (key: string, value: any) => void
  resetFilters: () => void

  // 排序状态
  sorter: SorterResult<T> | undefined
  setSorter: (sorter: SorterResult<T>) => void

  // 表格变化处理
  handleTableChange: (
    pagination: TablePaginationConfig,
    filters: Record<string, FilterValue | null>,
    sorter: SorterResult<T> | SorterResult<T>[]
  ) => void

  // 数据刷新
  refresh: () => void
  setData: (data: T[]) => void
  setTotal: (total: number) => void
  setLoading: (loading: boolean) => void

  // 选中行
  selectedRowKeys: React.Key[]
  setSelectedRowKeys: (keys: React.Key[]) => void
  rowSelection: {
    selectedRowKeys: React.Key[]
    onChange: (selectedRowKeys: React.Key[], selectedRows: T[]) => void
  }
}

export function useTable<T = any>(
  options: UseTableOptions<T> = {}
): UseTableReturn<T> {
  const { initialPageSize = 20, initialFilters = {}, onFetchData } = options

  // 数据状态
  const [data, setData] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)

  // 分页状态
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(initialPageSize)

  // 筛选状态
  const [filters, setFilters] = useState<TableFilters>(initialFilters)

  // 排序状态
  const [sorter, setSorter] = useState<SorterResult<T> | undefined>()

  // 选中行状态
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  /**
   * 获取数据
   */
  const fetchData = useCallback(async () => {
    if (!onFetchData) return

    setLoading(true)
    try {
      const result = await onFetchData({
        page,
        pageSize,
        filters,
        sorter,
      })
      setData(result.data)
      setTotal(result.total)
    } catch (error) {
      console.error('Failed to fetch table data:', error)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, filters, sorter, onFetchData])

  /**
   * 更新单个筛选条件
   */
  const updateFilter = useCallback((key: string, value: any) => {
    setFilters((prev) => ({
      ...prev,
      [key]: value,
    }))
    setPage(1) // 重置到第一页
  }, [])

  /**
   * 重置筛选条件
   */
  const resetFilters = useCallback(() => {
    setFilters(initialFilters)
    setPage(1)
  }, [initialFilters])

  /**
   * 处理表格变化 (分页、筛选、排序)
   */
  const handleTableChange = useCallback(
    (
      pagination: TablePaginationConfig,
      tableFilters: Record<string, FilterValue | null>,
      tableSorter: SorterResult<T> | SorterResult<T>[]
    ) => {
      // 更新分页
      if (pagination.current !== page) {
        setPage(pagination.current || 1)
      }
      if (pagination.pageSize !== pageSize) {
        setPageSize(pagination.pageSize || initialPageSize)
        setPage(1) // 改变每页条数时重置到第一页
      }

      // 更新筛选
      const newFilters: TableFilters = {}
      Object.entries(tableFilters).forEach(([key, value]) => {
        if (value !== null && value !== undefined) {
          newFilters[key] = value
        }
      })
      if (JSON.stringify(newFilters) !== JSON.stringify(filters)) {
        setFilters(newFilters)
        setPage(1) // 筛选变化时重置到第一页
      }

      // 更新排序
      const newSorter = Array.isArray(tableSorter) ? tableSorter[0] : tableSorter
      if (JSON.stringify(newSorter) !== JSON.stringify(sorter)) {
        setSorter(newSorter)
      }
    },
    [page, pageSize, filters, sorter, initialPageSize]
  )

  /**
   * 刷新数据
   */
  const refresh = useCallback(() => {
    fetchData()
  }, [fetchData])

  /**
   * 分页配置
   */
  const pagination = useMemo<TablePaginationConfig>(
    () => ({
      current: page,
      pageSize,
      total,
      showSizeChanger: true,
      showQuickJumper: true,
      showTotal: (total) => `共 ${total} 条`,
      pageSizeOptions: ['10', '20', '50', '100'],
    }),
    [page, pageSize, total]
  )

  /**
   * 行选择配置
   */
  const rowSelection = useMemo(
    () => ({
      selectedRowKeys,
      onChange: (selectedRowKeys: React.Key[], selectedRows: T[]) => {
        setSelectedRowKeys(selectedRowKeys)
      },
    }),
    [selectedRowKeys]
  )

  return {
    // 数据状态
    data,
    total,
    loading,

    // 分页状态
    pagination,
    page,
    pageSize,

    // 筛选状态
    filters,
    setFilters,
    updateFilter,
    resetFilters,

    // 排序状态
    sorter,
    setSorter,

    // 表格变化处理
    handleTableChange,

    // 数据刷新
    refresh,
    setData,
    setTotal,
    setLoading,

    // 选中行
    selectedRowKeys,
    setSelectedRowKeys,
    rowSelection,
  }
}

/**
 * 简化版分页 Hook
 */
export function usePagination(initialPageSize: number = 20) {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(initialPageSize)

  const pagination = useMemo<TablePaginationConfig>(
    () => ({
      current: page,
      pageSize,
      showSizeChanger: true,
      showQuickJumper: true,
      showTotal: (total) => `共 ${total} 条`,
      onChange: (page, pageSize) => {
        setPage(page)
        setPageSize(pageSize)
      },
    }),
    [page, pageSize]
  )

  const reset = useCallback(() => {
    setPage(1)
  }, [])

  return {
    page,
    pageSize,
    setPage,
    setPageSize,
    pagination,
    reset,
  }
}

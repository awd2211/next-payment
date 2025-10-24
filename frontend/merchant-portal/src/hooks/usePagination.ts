import { useState, useCallback } from 'react'

interface PaginationState {
  current: number
  pageSize: number
  total: number
}

interface PaginationActions {
  setPage: (page: number) => void
  setPageSize: (size: number) => void
  setTotal: (total: number) => void
  reset: () => void
  nextPage: () => void
  prevPage: () => void
}

/**
 * 分页 Hook - 简化分页逻辑
 *
 * @param initialPage 初始页码
 * @param initialPageSize 初始每页条数
 * @returns [state, actions]
 *
 * @example
 * const [pagination, paginationActions] = usePagination(1, 10)
 *
 * // 获取当前页码
 * console.log(pagination.current)
 *
 * // 设置页码
 * paginationActions.setPage(2)
 *
 * // 下一页
 * paginationActions.nextPage()
 */
function usePagination(
  initialPage: number = 1,
  initialPageSize: number = 10
): [PaginationState, PaginationActions] {
  const [state, setState] = useState<PaginationState>({
    current: initialPage,
    pageSize: initialPageSize,
    total: 0,
  })

  const setPage = useCallback((page: number) => {
    setState(prev => ({ ...prev, current: page }))
  }, [])

  const setPageSize = useCallback((size: number) => {
    setState(prev => ({
      ...prev,
      pageSize: size,
      current: 1, // 重置到第一页
    }))
  }, [])

  const setTotal = useCallback((total: number) => {
    setState(prev => ({ ...prev, total }))
  }, [])

  const reset = useCallback(() => {
    setState({
      current: initialPage,
      pageSize: initialPageSize,
      total: 0,
    })
  }, [initialPage, initialPageSize])

  const nextPage = useCallback(() => {
    setState(prev => {
      const maxPage = Math.ceil(prev.total / prev.pageSize)
      if (prev.current < maxPage) {
        return { ...prev, current: prev.current + 1 }
      }
      return prev
    })
  }, [])

  const prevPage = useCallback(() => {
    setState(prev => {
      if (prev.current > 1) {
        return { ...prev, current: prev.current - 1 }
      }
      return prev
    })
  }, [])

  return [
    state,
    {
      setPage,
      setPageSize,
      setTotal,
      reset,
      nextPage,
      prevPage,
    },
  ]
}

export default usePagination

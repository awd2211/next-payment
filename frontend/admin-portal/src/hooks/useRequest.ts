import { useState, useCallback } from 'react'

interface RequestState<T> {
  data: T | null
  loading: boolean
  error: Error | null
}

interface UseRequestOptions<T> {
  manual?: boolean
  onSuccess?: (data: T) => void
  onError?: (error: Error) => void
}

/**
 * 自定义Hook: 处理异步请求的状态
 */
export function useRequest<T = any, P extends any[] = any[]>(
  requestFn: (...args: P) => Promise<T>,
  options: UseRequestOptions<T> = {}
) {
  const { manual = false, onSuccess, onError } = options

  const [state, setState] = useState<RequestState<T>>({
    data: null,
    loading: !manual,
    error: null,
  })

  const run = useCallback(
    async (...args: P) => {
      setState((prev) => ({ ...prev, loading: true, error: null }))

      try {
        const data = await requestFn(...args)
        setState({ data, loading: false, error: null })
        onSuccess?.(data)
        return data
      } catch (error) {
        const err = error as Error
        setState((prev) => ({ ...prev, loading: false, error: err }))
        onError?.(err)
        throw error
      }
    },
    [requestFn, onSuccess, onError]
  )

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null })
  }, [])

  const mutate = useCallback((newData: T | ((prevData: T | null) => T)) => {
    setState((prev) => ({
      ...prev,
      data: typeof newData === 'function' 
        ? (newData as (prevData: T | null) => T)(prev.data)
        : newData,
    }))
  }, [])

  return {
    ...state,
    run,
    reset,
    mutate,
  }
}

/**
 * 自定义Hook: 处理分页请求
 */
export interface PaginationParams {
  page: number
  page_size: number
}

export interface PaginationResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export function usePagination<T = any>(
  requestFn: (params: PaginationParams) => Promise<PaginationResult<T>>,
  options: { initialPage?: number; initialPageSize?: number } = {}
) {
  const { initialPage = 1, initialPageSize = 10 } = options

  const [pagination, setPagination] = useState({
    page: initialPage,
    page_size: initialPageSize,
  })

  const { data, loading, error, run } = useRequest(requestFn)

  const loadData = useCallback(
    (page?: number, page_size?: number) => {
      const params = {
        page: page ?? pagination.page,
        page_size: page_size ?? pagination.page_size,
      }
      setPagination(params)
      return run(params)
    },
    [pagination, run]
  )

  const refresh = useCallback(() => {
    return loadData(pagination.page, pagination.page_size)
  }, [loadData, pagination])

  const changePage = useCallback(
    (page: number) => {
      return loadData(page, pagination.page_size)
    },
    [loadData, pagination.page_size]
  )

  const changePageSize = useCallback(
    (page_size: number) => {
      return loadData(1, page_size)
    },
    [loadData]
  )

  return {
    data: data?.items ?? [],
    total: data?.total ?? 0,
    page: pagination.page,
    pageSize: pagination.page_size,
    totalPages: data?.total_pages ?? 0,
    loading,
    error,
    loadData,
    refresh,
    changePage,
    changePageSize,
  }
}




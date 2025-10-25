// @ts-nocheck
/**
 * 自定义 React Query Hooks
 * 封装常用的数据请求逻辑
 */
import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
  type UseMutationOptions,
} from '@tanstack/react-query'
import { message } from 'antd'

/**
 * 通用列表数据 Hook
 */
export function useListQuery<TData = any, TError = Error>(
  queryKey: readonly unknown[],
  queryFn: () => Promise<TData>,
  options?: Omit<UseQueryOptions<TData, TError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey,
    queryFn,
    ...options,
  })
}

/**
 * 通用详情数据 Hook
 */
export function useDetailQuery<TData = any, TError = Error>(
  queryKey: readonly unknown[],
  queryFn: () => Promise<TData>,
  options?: Omit<UseQueryOptions<TData, TError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey,
    queryFn,
    // 详情数据缓存时间更长
    staleTime: 1000 * 60 * 10,
    ...options,
  })
}

/**
 * 通用创建/更新 Mutation Hook
 */
export function useCreateMutation<TData = any, TVariables = any, TError = Error>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options?: UseMutationOptions<TData, TError, TVariables>
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onSuccess: (data, variables, context) => {
      // 默认成功提示
      if (options?.onSuccess) {
        options.onSuccess(data, variables, context)
      } else {
        message.success('操作成功')
      }
    },
    onError: (error: any, variables, context) => {
      // 默认错误提示
      if (options?.onError) {
        options.onError(error, variables, context)
      } else {
        const errorMessage = error?.response?.data?.message || error?.message || '操作失败'
        message.error(errorMessage)
      }
    },
    ...options,
  })
}

/**
 * 通用删除 Mutation Hook
 */
export function useDeleteMutation<TData = any, TVariables = any, TError = Error>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options?: UseMutationOptions<TData, TError, TVariables>
) {
  return useCreateMutation(mutationFn, {
    onSuccess: (data, variables, context) => {
      message.success('删除成功')
      options?.onSuccess?.(data, variables, context)
    },
    ...options,
  })
}

/**
 * 乐观更新 Hook
 * 在请求完成前先更新UI,提升用户体验
 */
export function useOptimisticMutation<TData = any, TVariables = any, TError = Error>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  queryKey: readonly unknown[],
  updateFn: (oldData: any, variables: TVariables) => any,
  options?: UseMutationOptions<TData, TError, TVariables>
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn,
    onMutate: async (variables) => {
      // 取消所有相关的查询,避免覆盖乐观更新
      await queryClient.cancelQueries({ queryKey })

      // 保存当前数据快照
      const previousData = queryClient.getQueryData(queryKey)

      // 乐观更新
      queryClient.setQueryData(queryKey, (old: any) => updateFn(old, variables))

      // 返回上下文,包含快照数据
      return { previousData }
    },
    onError: (error, variables, context: any) => {
      // 回滚到之前的数据
      if (context?.previousData) {
        queryClient.setQueryData(queryKey, context.previousData)
      }

      const errorMessage = error?.response?.data?.message || error?.message || '操作失败'
      message.error(errorMessage)
      options?.onError?.(error, variables, context)
    },
    onSettled: () => {
      // 无论成功或失败,都重新获取数据
      queryClient.invalidateQueries({ queryKey })
    },
    ...options,
  })
}

/**
 * 无限滚动 Hook
 */
export function useInfiniteListQuery<TData = any, TError = Error>(
  queryKey: readonly unknown[],
  queryFn: ({ pageParam }: { pageParam: number }) => Promise<TData>,
  options?: any
) {
  return useQuery({
    queryKey,
    queryFn: () => queryFn({ pageParam: 1 }),
    ...options,
  })
}

/**
 * 轮询查询 Hook
 * 定期刷新数据
 */
export function usePollingQuery<TData = any, TError = Error>(
  queryKey: readonly unknown[],
  queryFn: () => Promise<TData>,
  interval: number = 5000, // 默认 5 秒
  options?: Omit<UseQueryOptions<TData, TError>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey,
    queryFn,
    refetchInterval: interval,
    refetchIntervalInBackground: false, // 页面不可见时停止轮询
    ...options,
  })
}

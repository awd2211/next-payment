/**
 * React Query 配置
 * 统一管理数据请求、缓存和状态
 */
import { QueryClient } from '@tanstack/react-query'
import { message } from 'antd'

/**
 * 默认查询配置
 */
const defaultQueryConfig = {
  queries: {
    // 数据默认缓存 5 分钟
    staleTime: 1000 * 60 * 5,
    // 缓存数据保留 10 分钟
    gcTime: 1000 * 60 * 10,
    // 窗口获得焦点时重新获取数据
    refetchOnWindowFocus: false,
    // 网络重连时重新获取数据
    refetchOnReconnect: true,
    // 组件挂载时不自动重新获取
    refetchOnMount: false,
    // 失败重试次数
    retry: 1,
    // 重试延迟
    retryDelay: (attemptIndex: number) => Math.min(1000 * 2 ** attemptIndex, 30000),
  },
  mutations: {
    // mutation 失败时的错误处理
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.message || error?.message || '操作失败'
      message.error(errorMessage)
    },
  },
}

/**
 * 创建 QueryClient 实例
 */
export const queryClient = new QueryClient({
  defaultOptions: defaultQueryConfig,
})

/**
 * Query Keys 工厂函数
 * 统一管理所有的 query key,避免重复和冲突
 */
export const queryKeys = {
  // Dashboard
  dashboard: {
    all: ['dashboard'] as const,
    data: (timePeriod: string) => ['dashboard', 'data', timePeriod] as const,
  },

  // Admins
  admins: {
    all: ['admins'] as const,
    list: (filters?: Record<string, any>) => ['admins', 'list', filters] as const,
    detail: (id: string) => ['admins', 'detail', id] as const,
  },

  // Roles
  roles: {
    all: ['roles'] as const,
    list: (filters?: Record<string, any>) => ['roles', 'list', filters] as const,
    detail: (id: string) => ['roles', 'detail', id] as const,
    permissions: (id: string) => ['roles', 'permissions', id] as const,
  },

  // Merchants
  merchants: {
    all: ['merchants'] as const,
    list: (filters?: Record<string, any>) => ['merchants', 'list', filters] as const,
    detail: (id: string) => ['merchants', 'detail', id] as const,
    stats: ['merchants', 'stats'] as const,
  },

  // Payments
  payments: {
    all: ['payments'] as const,
    list: (filters?: Record<string, any>) => ['payments', 'list', filters] as const,
    detail: (id: string) => ['payments', 'detail', id] as const,
    stats: (filters?: Record<string, any>) => ['payments', 'stats', filters] as const,
  },

  // Orders
  orders: {
    all: ['orders'] as const,
    list: (filters?: Record<string, any>) => ['orders', 'list', filters] as const,
    detail: (id: string) => ['orders', 'detail', id] as const,
    stats: (filters?: Record<string, any>) => ['orders', 'stats', filters] as const,
  },

  // Risk
  risk: {
    all: ['risk'] as const,
    rules: ['risk', 'rules'] as const,
    blacklist: ['risk', 'blacklist'] as const,
    logs: (filters?: Record<string, any>) => ['risk', 'logs', filters] as const,
  },

  // Settlements
  settlements: {
    all: ['settlements'] as const,
    list: (filters?: Record<string, any>) => ['settlements', 'list', filters] as const,
    detail: (id: string) => ['settlements', 'detail', id] as const,
  },

  // Audit Logs
  auditLogs: {
    all: ['audit-logs'] as const,
    list: (filters?: Record<string, any>) => ['audit-logs', 'list', filters] as const,
  },

  // System Configs
  systemConfigs: {
    all: ['system-configs'] as const,
    list: ['system-configs', 'list'] as const,
    detail: (key: string) => ['system-configs', 'detail', key] as const,
  },

  // Cashier
  cashier: {
    all: ['cashier'] as const,
    themes: ['cashier', 'themes'] as const,
    config: (merchantId: string) => ['cashier', 'config', merchantId] as const,
  },
} as const

/**
 * 预定义的错误处理函数
 */
export const handleQueryError = (error: any, customMessage?: string) => {
  const errorMessage = customMessage || error?.response?.data?.message || error?.message || '数据加载失败'
  console.error('Query Error:', error)
  message.error(errorMessage)
}

/**
 * 预定义的成功处理函数
 */
export const handleMutationSuccess = (successMessage: string) => {
  message.success(successMessage)
}

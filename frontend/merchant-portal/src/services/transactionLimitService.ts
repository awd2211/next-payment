import request from './request'

export interface TransactionLimit {
  id: string
  merchant_id: string
  limit_type: 'single_transaction' | 'daily' | 'monthly' | 'yearly'
  currency: string
  min_amount?: number
  max_amount?: number
  max_count?: number
  current_amount?: number
  current_count?: number
  reset_at?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface LimitStatus {
  limit_type: string
  currency: string
  max_amount?: number
  max_count?: number
  used_amount: number
  used_count: number
  remaining_amount?: number
  remaining_count?: number
  usage_percentage: number
  reset_at?: string
  is_exceeded: boolean
}

export const transactionLimitService = {
  /**
   * 获取商户交易限额列表
   */
  list: (params?: {
    limit_type?: string
    currency?: string
  }) => {
    return request.get<TransactionLimit[]>('/merchant/transaction-limits', { params })
  },

  /**
   * 获取限额状态
   */
  getStatus: (params?: {
    limit_type?: string
    currency?: string
  }) => {
    return request.get<LimitStatus[]>('/merchant/transaction-limits/status', { params })
  },

  /**
   * 检查交易是否超限
   */
  checkLimit: (data: {
    amount: number
    currency: string
    limit_type?: string
  }) => {
    return request.post<{ data: {
      is_allowed: boolean
      exceeded_limits: string[]
      remaining_amount?: number
      remaining_count?: number
    } }>('/merchant/transaction-limits/check', data)
  },

  /**
   * 获取限额使用历史
   */
  getHistory: (params?: {
    limit_type?: string
    currency?: string
    start_date?: string
    end_date?: string
  }) => {
    return request.get('/merchant/transaction-limits/history', { params })
  },

  /**
   * 申请提升限额
   */
  requestIncrease: (data: {
    limit_type: string
    currency: string
    requested_amount?: number
    requested_count?: number
    reason: string
  }) => {
    return request.post('/merchant/transaction-limits/increase-request', data)
  },

  /**
   * 获取限额申请记录
   */
  getRequests: (params?: {
    status?: string
  }) => {
    return request.get('/merchant/transaction-limits/requests', { params })
  },
}

export default transactionLimitService

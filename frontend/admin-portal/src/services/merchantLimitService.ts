import request from './request'

// ===== 层级管理 =====

export interface LimitTier {
  id: string
  tier_name: string
  tier_level: number
  daily_limit: number
  monthly_limit: number
  single_transaction_limit: number
  description?: string
  created_at: string
  updated_at: string
}

export interface ListTiersParams {
  page?: number
  page_size?: number
}

export interface ListTiersResponse {
  data: LimitTier[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

// ===== 商户限额管理 =====

export interface MerchantLimit {
  id: string
  merchant_id: string
  merchant_name?: string
  tier_id: string
  tier_name?: string
  daily_limit: number
  monthly_limit: number
  single_transaction_limit: number
  used_daily_amount: number
  used_monthly_amount: number
  remaining_daily_amount: number
  remaining_monthly_amount: number
  is_suspended: boolean
  suspend_reason?: string
  created_at: string
  updated_at: string
}

export interface UsageHistory {
  id: string
  merchant_id: string
  transaction_type: string
  amount: number
  currency: string
  daily_used: number
  monthly_used: number
  created_at: string
}

export interface UsageStatistics {
  merchant_id: string
  current_tier: string
  daily_usage_rate: number
  monthly_usage_rate: number
  total_transactions_today: number
  total_transactions_month: number
  average_transaction_amount: number
  peak_transaction_amount: number
  last_transaction_at?: string
}

// ===== API Service =====

export const merchantLimitService = {
  // ===== 层级管理 API (管理员) =====

  /**
   * 创建限额层级
   */
  createTier: (data: Partial<LimitTier>) => {
    return request.post<LimitTier>('/api/v1/admin/tiers', data)
  },

  /**
   * 获取层级列表
   */
  listTiers: (params: ListTiersParams) => {
    return request.get<ListTiersResponse>('/api/v1/admin/tiers', { params })
  },

  /**
   * 获取单个层级详情
   */
  getTier: (tierId: string) => {
    return request.get<{ data: LimitTier }>(`/api/v1/admin/tiers/${tierId}`)
  },

  /**
   * 更新层级
   */
  updateTier: (tierId: string, data: Partial<LimitTier>) => {
    return request.put<LimitTier>(`/api/v1/admin/tiers/${tierId}`, data)
  },

  /**
   * 删除层级
   */
  deleteTier: (tierId: string) => {
    return request.delete(`/api/v1/admin/tiers/${tierId}`)
  },

  // ===== 商户限额管理 API =====

  /**
   * 初始化商户限额
   */
  initializeLimit: (merchantId: string, tierId: string) => {
    return request.post('/api/v1/admin/limits/initialize', {
      merchant_id: merchantId,
      tier_id: tierId,
    })
  },

  /**
   * 获取商户限额
   */
  getMerchantLimit: (merchantId: string) => {
    return request.get<{ data: MerchantLimit }>(`/api/v1/admin/limits/${merchantId}`)
  },

  /**
   * 更新商户限额
   */
  updateMerchantLimit: (merchantId: string, data: Partial<MerchantLimit>) => {
    return request.put(`/api/v1/admin/limits/${merchantId}`, data)
  },

  /**
   * 变更商户层级
   */
  changeTier: (merchantId: string, newTierId: string, reason?: string) => {
    return request.post(`/api/v1/admin/limits/${merchantId}/change-tier`, {
      new_tier_id: newTierId,
      reason,
    })
  },

  /**
   * 暂停商户交易
   */
  suspendMerchant: (merchantId: string, reason: string) => {
    return request.post(`/api/v1/admin/limits/${merchantId}/suspend`, { reason })
  },

  /**
   * 恢复商户交易
   */
  unsuspendMerchant: (merchantId: string) => {
    return request.post(`/api/v1/admin/limits/${merchantId}/unsuspend`)
  },

  /**
   * 获取商户使用历史
   */
  getUsageHistory: (
    merchantId: string,
    params?: { page?: number; page_size?: number; start_date?: string; end_date?: string },
  ) => {
    return request.get<{ data: UsageHistory[]; pagination: any }>(`/api/v1/admin/limits/${merchantId}/usage-history`, {
      params,
    })
  },

  /**
   * 获取商户使用统计
   */
  getStatistics: (merchantId: string) => {
    return request.get<{ data: UsageStatistics }>(`/api/v1/admin/limits/${merchantId}/statistics`)
  },
}

export default merchantLimitService

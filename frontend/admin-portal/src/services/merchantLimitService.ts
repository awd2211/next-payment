import request from './request'

// Types
export interface MerchantLimit {
  id: string
  merchant_id: string
  merchant_name: string
  daily_transaction_limit: number
  daily_amount_limit: number
  monthly_transaction_limit: number
  monthly_amount_limit: number
  single_transaction_min: number
  single_transaction_max: number
  current_daily_count: number
  current_daily_amount: number
  current_monthly_count: number
  current_monthly_amount: number
  is_enabled: boolean
  alert_threshold: number
  created_at: string
  updated_at: string
}

export interface LimitUsageHistory {
  id: string
  merchant_id: string
  date: string
  transaction_count: number
  transaction_amount: number
  limit_type: 'daily' | 'monthly'
}

export interface ListMerchantLimitsParams {
  page?: number
  page_size?: number
  merchant_id?: string
  merchant_name?: string
  is_enabled?: boolean
  alert_status?: 'normal' | 'warning' | 'exceeded'
}

export interface ListMerchantLimitsResponse {
  code: number
  message: string
  data: {
    list: MerchantLimit[]
    total: number
    page: number
    page_size: number
  }
}

export interface MerchantLimitDetailResponse {
  code: number
  message: string
  data: MerchantLimit
}

export interface UpdateMerchantLimitRequest {
  daily_transaction_limit?: number
  daily_amount_limit?: number
  monthly_transaction_limit?: number
  monthly_amount_limit?: number
  single_transaction_min?: number
  single_transaction_max?: number
  is_enabled?: boolean
  alert_threshold?: number
}

export interface UpdateMerchantLimitResponse {
  code: number
  message: string
  data: MerchantLimit
}

export interface LimitUsageStatsResponse {
  code: number
  message: string
  data: {
    merchant_id: string
    daily: {
      transaction_count: number
      transaction_amount: number
      usage_rate_count: number
      usage_rate_amount: number
    }
    monthly: {
      transaction_count: number
      transaction_amount: number
      usage_rate_count: number
      usage_rate_amount: number
    }
  }
}

export interface LimitAlertConfig {
  id: string
  merchant_id: string
  alert_type: 'email' | 'sms' | 'webhook'
  recipients: string[]
  is_enabled: boolean
}

// API Methods
export const merchantLimitService = {
  /**
   * Get merchant limits list
   */
  list: (params: ListMerchantLimitsParams) => {
    return request.get<ListMerchantLimitsResponse>('/api/v1/admin/merchant-limits', { params })
  },

  /**
   * Get merchant limit detail by merchant ID
   */
  getDetail: (merchantId: string) => {
    return request.get<MerchantLimitDetailResponse>(`/api/v1/admin/merchant-limits/${merchantId}`)
  },

  /**
   * Update merchant limit
   */
  update: (merchantId: string, data: UpdateMerchantLimitRequest) => {
    return request.put<UpdateMerchantLimitResponse>(`/api/v1/admin/merchant-limits/${merchantId}`, data)
  },

  /**
   * Create merchant limit (for new merchant)
   */
  create: (merchantId: string, data: UpdateMerchantLimitRequest) => {
    return request.post<UpdateMerchantLimitResponse>(`/api/v1/admin/merchant-limits/${merchantId}`, data)
  },

  /**
   * Get current usage statistics for a merchant
   */
  getUsageStats: (merchantId: string) => {
    return request.get<LimitUsageStatsResponse>(`/api/v1/admin/merchant-limits/${merchantId}/usage`)
  },

  /**
   * Get usage history
   */
  getUsageHistory: (merchantId: string, params?: { start_date?: string; end_date?: string; limit_type?: string }) => {
    return request.get(`/api/v1/admin/merchant-limits/${merchantId}/history`, { params })
  },

  /**
   * Reset daily/monthly counters (admin operation)
   */
  resetCounters: (merchantId: string, type: 'daily' | 'monthly') => {
    return request.post(`/api/v1/admin/merchant-limits/${merchantId}/reset`, { type })
  },

  /**
   * Get merchants exceeding limits (alert dashboard)
   */
  getAlertslist: (params?: { threshold?: number }) => {
    return request.get('/api/v1/admin/merchant-limits/alerts', { params })
  },

  /**
   * Get limit alert configuration
   */
  getAlertConfig: (merchantId: string) => {
    return request.get(`/api/v1/admin/merchant-limits/${merchantId}/alert-config`)
  },

  /**
   * Update limit alert configuration
   */
  updateAlertConfig: (merchantId: string, data: Partial<LimitAlertConfig>) => {
    return request.put(`/api/v1/admin/merchant-limits/${merchantId}/alert-config`, data)
  },

  /**
   * Batch update limits (for multiple merchants)
   */
  batchUpdate: (data: { merchant_ids: string[]; limits: UpdateMerchantLimitRequest }) => {
    return request.post('/api/v1/admin/merchant-limits/batch-update', data)
  },

  /**
   * Export merchant limits report
   */
  export: (params: ListMerchantLimitsParams) => {
    return request.get('/api/v1/admin/merchant-limits/export', {
      params,
      responseType: 'blob',
    })
  },

  /**
   * Get system-wide limit statistics
   */
  getSystemStats: () => {
    return request.get('/api/v1/admin/merchant-limits/system-stats')
  },

  /**
   * Get limit templates (preset configurations)
   */
  getTemplates: () => {
    return request.get('/api/v1/admin/merchant-limits/templates')
  },

  /**
   * Apply limit template to merchant
   */
  applyTemplate: (merchantId: string, templateId: string) => {
    return request.post(`/api/v1/admin/merchant-limits/${merchantId}/apply-template`, { template_id: templateId })
  },
}

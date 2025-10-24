import request from './request'

// ========== 费率配置 ==========
export interface FeeConfig {
  id: string
  merchant_id: string
  channel: string // stripe, paypal, alipay, etc.
  payment_method: string // card, bank_transfer, e_wallet
  fee_type: string // percentage, fixed, mixed
  percentage_fee: number // 百分比费率 (如 2.9% 存储为 2.9)
  fixed_fee: number // 固定费用 (分为单位)
  min_fee: number // 最低手续费
  max_fee: number // 最高手续费
  currency: string
  status: string // active, inactive, pending_approval
  effective_date: string
  expiry_date: string
  created_at: string
  updated_at: string
}

export interface CreateFeeConfigInput {
  channel: string
  payment_method: string
  fee_type: string
  percentage_fee: number
  fixed_fee: number
  min_fee?: number
  max_fee?: number
  currency: string
  effective_date?: string
  expiry_date?: string
}

// ========== 交易限额 ==========
export interface TransactionLimit {
  id: string
  merchant_id: string
  limit_type: string // single, daily, monthly
  currency: string
  min_amount: number
  max_amount: number
  daily_count_limit: number
  daily_amount_limit: number
  monthly_count_limit: number
  monthly_amount_limit: number
  status: string
  created_at: string
  updated_at: string
}

export interface CreateTransactionLimitInput {
  limit_type: string
  currency: string
  min_amount?: number
  max_amount?: number
  daily_count_limit?: number
  daily_amount_limit?: number
  monthly_count_limit?: number
  monthly_amount_limit?: number
}

// ========== 渠道配置 ==========
export interface ChannelConfig {
  id: string
  merchant_id: string
  channel: string // stripe, paypal, alipay, wechat
  is_enabled: boolean
  priority: number // 优先级，数字越小优先级越高
  config: Record<string, any> // 渠道特定配置 (JSON)
  api_key: string
  api_secret: string
  webhook_secret: string
  test_mode: boolean
  created_at: string
  updated_at: string
}

export interface CreateChannelConfigInput {
  channel: string
  priority?: number
  config?: Record<string, any>
  api_key: string
  api_secret: string
  webhook_secret?: string
  test_mode?: boolean
}

export const configService = {
  // ========== 费率配置 ==========
  listFeeConfigs: (merchantId: string) => {
    return request.get<{ data: FeeConfig[] }>(`/fee-configs/merchant/${merchantId}`)
  },

  getFeeConfig: (id: string) => {
    return request.get<{ data: FeeConfig }>(`/fee-configs/${id}`)
  },

  createFeeConfig: (data: CreateFeeConfigInput) => {
    return request.post<{ data: FeeConfig }>('/fee-configs', data)
  },

  updateFeeConfig: (id: string, data: Partial<CreateFeeConfigInput>) => {
    return request.put<{ data: FeeConfig }>(`/fee-configs/${id}`, data)
  },

  deleteFeeConfig: (id: string) => {
    return request.delete(`/fee-configs/${id}`)
  },

  calculateFee: (amount: number, channel: string, paymentMethod: string) => {
    return request.post<{ data: { fee: number; total: number } }>('/fee-configs/calculate-fee', {
      amount,
      channel,
      payment_method: paymentMethod,
    })
  },

  // ========== 交易限额 ==========
  listTransactionLimits: (merchantId: string) => {
    return request.get<{ data: TransactionLimit[] }>(`/transaction-limits/merchant/${merchantId}`)
  },

  getTransactionLimit: (id: string) => {
    return request.get<{ data: TransactionLimit }>(`/transaction-limits/${id}`)
  },

  createTransactionLimit: (data: CreateTransactionLimitInput) => {
    return request.post<{ data: TransactionLimit }>('/transaction-limits', data)
  },

  updateTransactionLimit: (id: string, data: Partial<CreateTransactionLimitInput>) => {
    return request.put<{ data: TransactionLimit }>(`/transaction-limits/${id}`, data)
  },

  deleteTransactionLimit: (id: string) => {
    return request.delete(`/transaction-limits/${id}`)
  },

  checkLimit: (amount: number, currency: string) => {
    return request.post<{ data: { allowed: boolean; reason?: string } }>('/transaction-limits/check-limit', {
      amount,
      currency,
    })
  },

  // ========== 渠道配置 ==========
  listChannelConfigs: (merchantId: string) => {
    return request.get<{ data: ChannelConfig[] }>(`/channel-configs/merchant/${merchantId}`)
  },

  getChannelConfig: (id: string) => {
    return request.get<{ data: ChannelConfig }>(`/channel-configs/${id}`)
  },

  getMerchantChannel: (merchantId: string, channel: string) => {
    return request.get<{ data: ChannelConfig }>(`/channel-configs/merchant/${merchantId}/channel/${channel}`)
  },

  createChannelConfig: (data: CreateChannelConfigInput) => {
    return request.post<{ data: ChannelConfig }>('/channel-configs', data)
  },

  updateChannelConfig: (id: string, data: Partial<CreateChannelConfigInput>) => {
    return request.put<{ data: ChannelConfig }>(`/channel-configs/${id}`, data)
  },

  deleteChannelConfig: (id: string) => {
    return request.delete(`/channel-configs/${id}`)
  },

  enableChannel: (id: string) => {
    return request.post<{ data: ChannelConfig }>(`/channel-configs/${id}/enable`)
  },

  disableChannel: (id: string) => {
    return request.post<{ data: ChannelConfig }>(`/channel-configs/${id}/disable`)
  },
}

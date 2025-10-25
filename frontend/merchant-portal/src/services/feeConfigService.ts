import request from './request'

export interface FeeConfig {
  id: string
  merchant_id: string
  channel: string
  payment_method: string
  fee_type: 'percentage' | 'fixed' | 'mixed'
  fee_percentage?: number
  fee_fixed?: number
  min_fee?: number
  max_fee?: number
  currency: string
  is_active: boolean
  effective_from: string
  effective_until?: string
  created_at: string
  updated_at: string
}

export interface FeeCalculation {
  base_amount: number
  fee_amount: number
  total_amount: number
  currency: string
  channel: string
  payment_method: string
  fee_breakdown: {
    percentage_fee?: number
    fixed_fee?: number
    min_fee_applied?: boolean
    max_fee_applied?: boolean
  }
}

export const feeConfigService = {
  /**
   * 获取商户费率配置列表
   */
  list: (params?: {
    channel?: string
    payment_method?: string
    is_active?: boolean
  }) => {
    return request.get<FeeConfig[]>('/merchant/fee-configs', { params })
  },

  /**
   * 获取特定渠道的费率配置
   */
  getByChannel: (channel: string, paymentMethod?: string) => {
    return request.get<FeeConfig>('/merchant/fee-configs/channel', {
      params: { channel, payment_method: paymentMethod }
    })
  },

  /**
   * 计算手续费
   */
  calculateFee: (data: {
    amount: number
    currency: string
    channel: string
    payment_method?: string
  }) => {
    return request.post<FeeCalculation>('/merchant/fee-configs/calculate', data)
  },

  /**
   * 获取费率历史
   */
  getHistory: (params?: {
    channel?: string
    start_date?: string
    end_date?: string
  }) => {
    return request.get('/merchant/fee-configs/history', { params })
  },

  /**
   * 获取所有支持的渠道及费率
   */
  getSupportedChannels: () => {
    return request.get('/merchant/fee-configs/channels')
  },
}

export default feeConfigService

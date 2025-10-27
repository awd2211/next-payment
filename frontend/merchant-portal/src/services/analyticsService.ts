import request from './request'

export interface PaymentMetrics {
  total_count: number
  total_amount: number
  success_count: number
  success_amount: number
  failed_count: number
  failed_amount: number
  pending_count: number
  pending_amount: number
  success_rate: number
  average_amount: number
}

export interface PaymentSummary {
  date: string
  count: number
  amount: number
  success_count: number
  success_amount: number
  failed_count: number
  failed_amount: number
}

export interface MerchantMetrics {
  total_merchants: number
  active_merchants: number
  new_merchants: number
  merchants_with_transactions: number
}

export interface ChannelMetrics {
  channel: string
  count: number
  amount: number
  success_rate: number
  average_amount: number
}

export interface RealtimeStats {
  current_tps: number // transactions per second
  current_online_users: number
  today_total_count: number
  today_total_amount: number
  last_hour_count: number
  last_hour_amount: number
}

export interface AnalyticsParams {
  merchant_id?: string
  start_date?: string
  end_date?: string
  currency?: string
  channel?: string
}

export const analyticsService = {
  // Payment Analytics
  getPaymentMetrics: (params: AnalyticsParams) => {
    return request.get<PaymentMetrics>('/merchant/analytics/payments/metrics', { params })
  },

  getPaymentSummary: (params: AnalyticsParams) => {
    return request.get<PaymentSummary[]>('/merchant/analytics/payments/summary', { params })
  },

  // Merchant Analytics
  getMerchantMetrics: (params: Omit<AnalyticsParams, 'merchant_id'>) => {
    return request.get<MerchantMetrics>('/merchant/analytics/merchants/metrics', { params })
  },

  getMerchantSummary: (params: Omit<AnalyticsParams, 'merchant_id'>) => {
    return request.get('/merchant/analytics/merchants/summary', { params })
  },

  // Channel Analytics
  getChannelMetrics: (params: AnalyticsParams) => {
    return request.get<ChannelMetrics[]>('/merchant/analytics/channels/metrics', { params })
  },

  getChannelSummary: (params: AnalyticsParams) => {
    return request.get('/merchant/analytics/channels/summary', { params })
  },

  // Real-time Statistics
  getRealtimeStats: () => {
    return request.get<RealtimeStats>('/merchant/analytics/realtime/stats')
  },
}

export default analyticsService

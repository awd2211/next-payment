import api from './api'

export interface Payment {
  id: string
  payment_no: string
  merchant_id: string
  merchant_name: string
  order_id: string
  amount: number
  currency: string
  channel: string
  method: string
  status: string
  client_ip: string
  notify_url: string
  return_url: string
  created_at: string
  updated_at: string
}

export interface PaymentStats {
  total_count: number
  total_amount: number
  success_count: number
  failed_count: number
  pending_count: number
  success_rate: number
}

export interface PaymentListParams {
  page: number
  page_size: number
  payment_no?: string
  merchant_id?: string
  status?: string
  channel?: string
  method?: string
  start_time?: string
  end_time?: string
}

export interface PaymentListResponse {
  data: Payment[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface PaymentStatsParams {
  merchant_id?: string
  start_time?: string
  end_time?: string
}

export const paymentService = {
  // 获取支付列表
  list: (params: PaymentListParams) => {
    return api.get<PaymentListResponse>('/payments', { params })
  },

  // 获取支付详情
  get: (id: string) => {
    return api.get<Payment>(`/payments/${id}`)
  },

  // 获取支付统计
  getStats: (params: PaymentStatsParams) => {
    return api.get<{ data: PaymentStats }>('/payments/stats', { params })
  },

  // 取消支付
  cancel: (id: string, reason: string) => {
    return api.post(`/payments/${id}/cancel`, { reason })
  },

  // 重试支付
  retry: (id: string) => {
    return api.post(`/payments/${id}/retry`)
  },
}

export default paymentService

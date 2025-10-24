import request from './request'

export interface Payment {
  id: string
  merchant_id: string
  order_id: string
  channel: string
  method: string
  amount: number
  currency: string
  status: string
  customer_id: string
  customer_email: string
  description: string
  callback_url: string
  return_url: string
  ip_address: string
  user_agent: string
  metadata: Record<string, any>
  paid_at: string
  expires_at: string
  created_at: string
  updated_at: string
}

export interface PaymentStats {
  total_amount: number
  total_count: number
  success_count: number
  failed_count: number
  pending_count: number
  success_rate: number
  today_amount: number
  today_count: number
}

export interface ListPaymentsParams {
  page?: number
  page_size?: number
  order_id?: string
  status?: string
  channel?: string
  method?: string
  start_time?: string
  end_time?: string
}

export interface ListPaymentsResponse {
  data: Payment[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}

export const paymentService = {
  list: (params: ListPaymentsParams) => {
    return request.get<ListPaymentsResponse>('/merchant/payments', { params })
  },

  getById: (id: string) => {
    return request.get(`/merchant/payments/${id}`)
  },

  getStats: (params: { start_time?: string; end_time?: string }) => {
    return request.get<{ data: PaymentStats }>('/merchant/payments/stats', { params })
  },

  refund: (id: string, data: { amount?: number; reason: string }) => {
    return request.post(`/merchant/payments/${id}/refund`, data)
  },

  cancel: (id: string, reason: string) => {
    return request.post(`/merchant/payments/${id}/cancel`, { reason })
  },

  export: (params: ListPaymentsParams) => {
    return request.get('/merchant/payments/export', {
      params,
      responseType: 'blob',
    })
  },
}

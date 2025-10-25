import request from './request'

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

export interface Refund {
  id: string
  refund_no: string
  payment_no: string
  merchant_id: string
  amount: number
  currency: string
  reason?: string
  status: string
  channel_refund_no?: string
  error_message?: string
  created_at: string
  updated_at: string
}

export interface RefundListParams {
  page: number
  page_size: number
  refund_no?: string
  payment_no?: string
  merchant_id?: string
  status?: string
  start_time?: string
  end_time?: string
}

export const paymentService = {
  // Payment Management
  list: (params: PaymentListParams) => {
    return request.get<PaymentListResponse>('/payments', { params })
  },

  get: (paymentNo: string) => {
    return request.get<{ data: Payment }>(`/payments/${paymentNo}`)
  },

  batchGet: (paymentNos: string[]) => {
    return request.post<{ data: Payment[] }>('/payments/batch', { payment_nos: paymentNos })
  },

  cancel: (paymentNo: string) => {
    return request.post(`/payments/${paymentNo}/cancel`)
  },

  // Refund Management
  createRefund: (data: {
    payment_no: string
    amount: number
    reason?: string
  }) => {
    return request.post<{ data: Refund }>('/refunds', data)
  },

  getRefund: (refundNo: string) => {
    return request.get<{ data: Refund }>(`/refunds/${refundNo}`)
  },

  listRefunds: (params: RefundListParams) => {
    return request.get<{ data: Refund[]; pagination: any }>('/refunds', { params })
  },

  batchGetRefunds: (refundNos: string[]) => {
    return request.post<{ data: Refund[] }>('/refunds/batch', { refund_nos: refundNos })
  },

  // Merchant Portal APIs (with merchant auth)
  merchantListPayments: (params: PaymentListParams) => {
    return request.get('/merchant/payments', { params })
  },

  merchantGetPayment: (paymentNo: string) => {
    return request.get(`/merchant/payments/${paymentNo}`)
  },

  merchantBatchGetPayments: (paymentNos: string[]) => {
    return request.post('/merchant/payments/batch', { payment_nos: paymentNos })
  },

  merchantListRefunds: (params: RefundListParams) => {
    return request.get('/merchant/refunds', { params })
  },

  merchantGetRefund: (refundNo: string) => {
    return request.get(`/merchant/refunds/${refundNo}`)
  },

  merchantBatchGetRefunds: (refundNos: string[]) => {
    return request.post('/merchant/refunds/batch', { refund_nos: refundNos })
  },

  // Statistics
  getStats: (params: PaymentStatsParams) => {
    return request.get<{ data: PaymentStats }>('/payments/stats', { params })
  },
}

export default paymentService

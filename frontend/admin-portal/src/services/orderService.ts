import request from './request'

export interface Order {
  id: string
  order_no: string
  merchant_id: string
  merchant_name: string
  merchant_order_no: string
  product_name: string
  amount: number
  currency: string
  status: string
  payment_method: string
  payment_channel: string
  notify_url: string
  return_url: string
  client_ip: string
  expires_at: string
  paid_at?: string
  created_at: string
  updated_at: string
}

export interface OrderStats {
  total_count: number
  total_amount: number
  pending_count: number
  paid_count: number
  cancelled_count: number
  expired_count: number
}

export interface OrderListParams {
  page: number
  page_size: number
  order_no?: string
  merchant_order_no?: string
  merchant_id?: string
  status?: string
  start_time?: string
  end_time?: string
}

export interface OrderListResponse {
  data: Order[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface OrderStatsParams {
  merchant_id?: string
  start_time?: string
  end_time?: string
}

export interface CreateOrderRequest {
  merchant_id: string
  merchant_order_no: string
  product_name: string
  amount: number
  currency: string
  customer_id?: string
  customer_email?: string
  customer_phone?: string
  notify_url?: string
  return_url?: string
  client_ip?: string
  metadata?: any
  expires_in?: number // seconds
}

export interface DailySummary {
  date: string
  total_count: number
  total_amount: number
  paid_count: number
  paid_amount: number
  cancelled_count: number
  cancelled_amount: number
}

export interface OrderStatistics {
  merchant_id?: string
  currency?: string
  total_orders: number
  total_amount: number
  paid_orders: number
  paid_amount: number
  pending_orders: number
  pending_amount: number
  cancelled_orders: number
  cancelled_amount: number
  average_order_value: number
  conversion_rate: number
}

export const orderService = {
  // Order Management
  create: (data: CreateOrderRequest) => {
    return request.post<{ data: Order }>('/orders', data)
  },

  get: (orderNo: string) => {
    return request.get<{ data: Order }>(`/orders/${orderNo}`)
  },

  list: (params: OrderListParams) => {
    return request.get<OrderListResponse>('/orders', { params })
  },

  batchGet: (orderNos: string[]) => {
    return request.post<{ data: Order[] }>('/orders/batch', { order_nos: orderNos })
  },

  // Order Status Operations
  cancel: (orderNo: string, reason?: string) => {
    return request.post(`/orders/${orderNo}/cancel`, { reason })
  },

  markAsPaid: (orderNo: string, data?: { payment_no?: string; paid_at?: string }) => {
    return request.post(`/orders/${orderNo}/pay`, data)
  },

  refund: (orderNo: string, data: { amount?: number; reason?: string }) => {
    return request.post(`/orders/${orderNo}/refund`, data)
  },

  ship: (orderNo: string, data: { tracking_no?: string; carrier?: string }) => {
    return request.post(`/orders/${orderNo}/ship`, data)
  },

  complete: (orderNo: string) => {
    return request.post(`/orders/${orderNo}/complete`)
  },

  updateStatus: (orderNo: string, status: string) => {
    return request.put(`/orders/${orderNo}/status`, { status })
  },

  // Statistics & Reports
  getStats: (params: OrderStatsParams) => {
    return request.get<{ data: OrderStats }>('/orders/stats', { params })
  },

  getStatistics: (params: {
    merchant_id?: string
    start_time?: string
    end_time?: string
    currency?: string
  }) => {
    return request.get<{ data: OrderStatistics }>('/statistics/orders', { params })
  },

  getDailySummary: (params: {
    merchant_id?: string
    date?: string
    currency?: string
  }) => {
    return request.get<{ data: DailySummary }>('/statistics/daily-summary', { params })
  },
}

export default orderService

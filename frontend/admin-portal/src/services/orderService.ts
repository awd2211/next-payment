import api from './api'

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

export const orderService = {
  // 获取订单列表
  list: (params: OrderListParams) => {
    return api.get<OrderListResponse>('/orders', { params })
  },

  // 获取订单详情
  get: (id: string) => {
    return api.get<Order>(`/orders/${id}`)
  },

  // 获取订单统计
  getStats: (params: OrderStatsParams) => {
    return api.get<{ data: OrderStats }>('/orders/stats', { params })
  },

  // 取消订单
  cancel: (id: string, reason: string) => {
    return api.post(`/orders/${id}/cancel`, { reason })
  },
}

export default orderService

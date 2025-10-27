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
  // 获取订单列表 (管理员通过admin-bff-service)
  list: (params: OrderListParams) => {
    return request.get<OrderListResponse>('/api/v1/admin/orders', { params })
  },

  // 获取订单详情
  get: (orderNo: string) => {
    return request.get<{ data: Order }>(\`/api/v1/admin/orders/\${orderNo}\`)
  },

  // 获取指定商户的订单
  getMerchantOrders: (merchantId: string, params?: { page?: number; page_size?: number }) => {
    return request.get<OrderListResponse>(\`/api/v1/admin/orders/merchant/\${merchantId}\`, { params })
  },

  // 获取订单统计信息
  getStatistics: (params?: OrderStatsParams) => {
    return request.get<{ data: OrderStatistics }>('/api/v1/admin/orders/statistics', { params })
  },

  // 获取订单状态摘要
  getStatusSummary: (params?: { merchant_id?: string }) => {
    return request.get('/api/v1/admin/orders/status-summary', { params })
  },
}

export default orderService

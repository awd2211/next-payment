import request from './request'

export interface Order {
  id: string
  merchant_id: string
  merchant_order_id: string
  amount: number
  currency: string
  status: string
  customer_id: string
  customer_email: string
  customer_name: string
  description: string
  items: OrderItem[]
  shipping_address: ShippingAddress
  billing_address: BillingAddress
  metadata: Record<string, any>
  created_at: string
  updated_at: string
  paid_at: string
  cancelled_at: string
}

export interface OrderItem {
  name: string
  description: string
  quantity: number
  price: number
  amount: number
}

export interface ShippingAddress {
  name: string
  phone: string
  country: string
  province: string
  city: string
  district: string
  address: string
  postal_code: string
}

export interface BillingAddress {
  name: string
  phone: string
  country: string
  province: string
  city: string
  district: string
  address: string
  postal_code: string
}

export interface OrderStats {
  total_amount: number
  total_count: number
  paid_count: number
  pending_count: number
  cancelled_count: number
  today_amount: number
  today_count: number
}

export interface ListOrdersParams {
  page?: number
  page_size?: number
  merchant_order_id?: string
  status?: string
  customer_email?: string
  start_time?: string
  end_time?: string
}

export interface ListOrdersResponse {
  data: Order[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}

export interface CreateOrderRequest {
  merchant_order_id: string
  amount: number
  currency: string
  customer_id?: string
  customer_email: string
  customer_name: string
  description: string
  items: OrderItem[]
  shipping_address?: ShippingAddress
  billing_address?: BillingAddress
  metadata?: Record<string, any>
}

export const orderService = {
  list: (params: ListOrdersParams) => {
    return request.get<ListOrdersResponse>('/orders', { params })
  },

  getById: (id: string) => {
    return request.get(`/orders/${id}`)
  },

  create: (data: CreateOrderRequest) => {
    return request.post('/orders', data)
  },

  cancel: (id: string, reason: string) => {
    return request.post(`/orders/${id}/cancel`, { reason })
  },

  getStats: (params: { start_time?: string; end_time?: string }) => {
    return request.get<{ data: OrderStats }>('/orders/stats', { params })
  },
}

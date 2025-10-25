import request from './request'

export interface ChannelConfig {
  api_key?: string
  api_secret?: string
  webhook_secret?: string
  merchant_id?: string
  app_id?: string
  public_key?: string
  private_key?: string
  [key: string]: string | undefined
}

export interface Channel {
  id: string
  channel_code: string
  channel_name: string
  channel_type: 'stripe' | 'paypal' | 'alipay' | 'wechat' | 'crypto' | 'bank'
  is_enabled: boolean
  is_test_mode: boolean
  config: ChannelConfig
  supported_currencies: string[]
  supported_countries: string[]
  supported_payment_methods: string[]
  min_amount: number
  max_amount: number
  fee_type: 'percentage' | 'fixed' | 'mixed'
  fee_percentage?: number
  fee_fixed?: number
  priority: number
  description?: string
  created_at: string
  updated_at: string
}

export interface ListChannelsParams {
  page?: number
  page_size?: number
  channel_type?: string
  is_enabled?: boolean
  is_test_mode?: boolean
}

export interface ListChannelsResponse {
  list: Channel[]
  total: number
  page: number
  page_size: number
}

export interface CreateChannelRequest {
  channel_code: string
  channel_name: string
  channel_type: 'stripe' | 'paypal' | 'alipay' | 'wechat' | 'crypto' | 'bank'
  is_enabled?: boolean
  is_test_mode?: boolean
  config: ChannelConfig
  supported_currencies: string[]
  supported_countries?: string[]
  supported_payment_methods?: string[]
  min_amount?: number
  max_amount?: number
  fee_type: 'percentage' | 'fixed' | 'mixed'
  fee_percentage?: number
  fee_fixed?: number
  priority?: number
  description?: string
}

export interface UpdateChannelRequest {
  channel_name?: string
  is_enabled?: boolean
  is_test_mode?: boolean
  config?: ChannelConfig
  supported_currencies?: string[]
  supported_countries?: string[]
  supported_payment_methods?: string[]
  min_amount?: number
  max_amount?: number
  fee_type?: 'percentage' | 'fixed' | 'mixed'
  fee_percentage?: number
  fee_fixed?: number
  priority?: number
  description?: string
}

export interface ChannelStats {
  total_channels: number
  enabled_channels: number
  disabled_channels: number
  test_mode_channels: number
  channels_by_type: Record<string, number>
}

export interface ChannelHealthStatus {
  channel_code: string
  channel_name: string
  is_healthy: boolean
  last_check_at: string
  error_message?: string
  response_time?: number
}

export interface CreateChannelPaymentRequest {
  merchant_id: string
  order_no: string
  amount: number
  currency: string
  channel: string
  payment_method?: string
  customer_email?: string
  customer_phone?: string
  return_url?: string
  notify_url?: string
}

export interface ChannelPaymentResponse {
  payment_no: string
  channel_trade_no?: string
  payment_url?: string
  client_secret?: string
  qr_code?: string
  status: string
}

export interface PreAuthRequest {
  merchant_id: string
  order_no: string
  amount: number
  currency: string
  channel: string
  customer_email?: string
  return_url?: string
}

export interface PreAuthResponse {
  pre_auth_no: string
  channel_pre_auth_no: string
  status: string
  client_secret?: string
}

export interface ExchangeRate {
  from_currency: string
  to_currency: string
  rate: number
  updated_at: string
}

export const channelService = {
  // Admin Channel Management
  listAdminChannels: (params?: ListChannelsParams) => {
    return request.get<Channel[]>('/admin/channels', { params })
  },

  getAdminChannel: (code: string) => {
    return request.get<Channel>(`/admin/channels/${code}`)
  },

  createAdminChannel: (data: CreateChannelRequest) => {
    return request.post<Channel>('/admin/channels', data)
  },

  updateAdminChannel: (code: string, data: UpdateChannelRequest) => {
    return request.put<Channel>(`/admin/channels/${code}`, data)
  },

  deleteAdminChannel: (code: string) => {
    return request.delete(`/admin/channels/${code}`)
  },

  // Channel Payment Operations
  createChannelPayment: (data: CreateChannelPaymentRequest) => {
    return request.post<ChannelPaymentResponse>('/channel/payments', data)
  },

  queryChannelPayment: (paymentNo: string) => {
    return request.get<ChannelPaymentResponse>(`/channel/payments/${paymentNo}`)
  },

  cancelChannelPayment: (paymentNo: string) => {
    return request.post(`/channel/payments/${paymentNo}/cancel`)
  },

  // Channel Refund Operations
  createChannelRefund: (data: {
    payment_no: string
    refund_no: string
    amount: number
    currency: string
    reason?: string
  }) => {
    return request.post('/channel/refunds', data)
  },

  queryChannelRefund: (refundNo: string) => {
    return request.get(`/channel/refunds/${refundNo}`)
  },

  // Pre-authorization (Stripe specific)
  createPreAuth: (data: PreAuthRequest) => {
    return request.post<PreAuthResponse>('/channel/pre-auth', data)
  },

  capturePreAuth: (data: {
    channel_pre_auth_no: string
    amount?: number
  }) => {
    return request.post('/channel/pre-auth/capture', data)
  },

  cancelPreAuth: (channelPreAuthNo: string) => {
    return request.post('/channel/pre-auth/cancel', {
      channel_pre_auth_no: channelPreAuthNo
    })
  },

  queryPreAuth: (channelPreAuthNo: string) => {
    return request.get(`/channel/pre-auth/${channelPreAuthNo}`)
  },

  // Channel Configuration
  listChannelConfigs: () => {
    return request.get('/channel/config')
  },

  getChannelConfig: (channel: string) => {
    return request.get(`/channel/config/${channel}`)
  },

  // Exchange Rates
  getExchangeRates: (fromCurrency?: string, toCurrency?: string) => {
    return request.get<ExchangeRate[]>('/exchange-rates', {
      params: { from_currency: fromCurrency, to_currency: toCurrency }
    })
  },

  getExchangeRate: (fromCurrency: string, toCurrency: string) => {
    return request.get<ExchangeRate>('/exchange-rates/convert', {
      params: { from: fromCurrency, to: toCurrency }
    })
  },

  // Legacy/Compatibility APIs
  list: (params: ListChannelsParams) => {
    return request.get<ListChannelsResponse>('/channels', { params })
  },

  getById: (id: string) => {
    return request.get<Channel>(`/channels/${id}`)
  },

  create: (data: CreateChannelRequest) => {
    return request.post<Channel>('/channels', data)
  },

  update: (id: string, data: UpdateChannelRequest) => {
    return request.put<Channel>(`/channels/${id}`, data)
  },

  delete: (id: string) => {
    return request.delete(`/channels/${id}`)
  },

  toggleEnable: (id: string, is_enabled: boolean) => {
    return request.put(`/channels/${id}/toggle`, { is_enabled })
  },

  toggleTestMode: (id: string, is_test_mode: boolean) => {
    return request.put(`/channels/${id}/test-mode`, { is_test_mode })
  },

  getStats: () => {
    return request.get<ChannelStats>('/channels/stats')
  },

  testConnection: (id: string) => {
    return request.post<{ success: boolean; message: string }>(`/channels/${id}/test`)
  },

  getHealthStatus: () => {
    return request.get<ChannelHealthStatus[]>('/channels/health')
  },

  getSupportedCurrencies: (channelType: string) => {
    return request.get<string[]>(`/channels/supported-currencies/${channelType}`)
  },

  getSupportedPaymentMethods: (channelType: string) => {
    return request.get<string[]>(`/channels/supported-methods/${channelType}`)
  },

  batchToggle: (ids: string[], is_enabled: boolean) => {
    return request.post('/channels/batch/toggle', { ids, is_enabled })
  },
}

export default channelService

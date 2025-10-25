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

export const channelService = {
  /**
   * 获取支付渠道列表
   */
  list: (params: ListChannelsParams) => {
    return request.get<ListChannelsResponse>('/channels', { params })
  },

  /**
   * 获取单个支付渠道详情
   */
  getById: (id: string) => {
    return request.get<Channel>(`/channels/${id}`)
  },

  /**
   * 创建支付渠道
   */
  create: (data: CreateChannelRequest) => {
    return request.post<Channel>('/channels', data)
  },

  /**
   * 更新支付渠道
   */
  update: (id: string, data: UpdateChannelRequest) => {
    return request.put<Channel>(`/channels/${id}`, data)
  },

  /**
   * 删除支付渠道
   */
  delete: (id: string) => {
    return request.delete(`/channels/${id}`)
  },

  /**
   * 启用/禁用支付渠道
   */
  toggleEnable: (id: string, is_enabled: boolean) => {
    return request.put(`/channels/${id}/toggle`, { is_enabled })
  },

  /**
   * 切换测试/生产模式
   */
  toggleTestMode: (id: string, is_test_mode: boolean) => {
    return request.put(`/channels/${id}/test-mode`, { is_test_mode })
  },

  /**
   * 获取渠道统计信息
   */
  getStats: () => {
    return request.get<ChannelStats>('/channels/stats')
  },

  /**
   * 测试渠道连接
   */
  testConnection: (id: string) => {
    return request.post<{ success: boolean; message: string }>(`/channels/${id}/test`)
  },

  /**
   * 获取所有渠道的健康状态
   */
  getHealthStatus: () => {
    return request.get<ChannelHealthStatus[]>('/channels/health')
  },

  /**
   * 获取支持的货币列表
   */
  getSupportedCurrencies: (channelType: string) => {
    return request.get<string[]>(`/channels/supported-currencies/${channelType}`)
  },

  /**
   * 获取支持的支付方式
   */
  getSupportedPaymentMethods: (channelType: string) => {
    return request.get<string[]>(`/channels/supported-methods/${channelType}`)
  },

  /**
   * 批量启用/禁用渠道
   */
  batchToggle: (ids: string[], is_enabled: boolean) => {
    return request.post('/channels/batch/toggle', { ids, is_enabled })
  },
}

export default channelService

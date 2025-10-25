import request from './request'

export interface MerchantProfile {
  id: string
  name: string
  email: string
  phone: string
  company_name: string
  business_type: string
  country: string
  website: string
  logo?: string
  status: string
  kyc_status: string
  is_test_mode: boolean
  metadata?: any
  created_at: string
  updated_at: string
}

export interface MerchantBalance {
  available_balance: number
  frozen_balance: number
  total_balance: number
  currency: string
  balances_by_currency: Array<{
    currency: string
    available: number
    frozen: number
    total: number
  }>
}

export interface MerchantStats {
  total_payments: number
  total_amount: number
  success_rate: number
  total_refunds: number
  total_refund_amount: number
  pending_settlements: number
  pending_settlement_amount: number
  today_payments: number
  today_amount: number
  this_month_payments: number
  this_month_amount: number
}

export interface UpdateProfileRequest {
  name?: string
  phone?: string
  company_name?: string
  business_type?: string
  country?: string
  website?: string
  metadata?: any
}

export const profileService = {
  /**
   * 获取商户资料
   */
  getProfile: () => {
    return request.get<{ data: MerchantProfile }>('/merchant/profile')
  },

  /**
   * 更新商户资料
   */
  updateProfile: (data: UpdateProfileRequest) => {
    return request.put<{ data: MerchantProfile }>('/merchant/profile', data)
  },

  /**
   * 获取商户余额
   */
  getBalance: () => {
    return request.get<{ data: MerchantBalance }>('/merchant/balance')
  },

  /**
   * 获取商户统计信息
   */
  getStats: () => {
    return request.get<{ data: MerchantStats }>('/merchant/stats')
  },

  /**
   * 上传商户Logo
   */
  uploadLogo: (file: File, onProgress?: (progress: number) => void) => {
    const formData = new FormData()
    formData.append('logo', file)
    return request.upload<{ data: { logo_url: string } }>('/merchant/logo', formData, onProgress)
  },

  /**
   * 删除Logo
   */
  deleteLogo: () => {
    return request.delete('/merchant/logo')
  },

  /**
   * 获取商户通知设置
   */
  getNotificationSettings: () => {
    return request.get('/merchant/notification-settings')
  },

  /**
   * 更新商户通知设置
   */
  updateNotificationSettings: (data: {
    email_notifications?: boolean
    sms_notifications?: boolean
    webhook_notifications?: boolean
    notify_on_payment?: boolean
    notify_on_refund?: boolean
    notify_on_settlement?: boolean
    notify_on_withdrawal?: boolean
  }) => {
    return request.put('/merchant/notification-settings', data)
  },

  /**
   * 获取API使用统计
   */
  getApiUsage: (params?: {
    start_date?: string
    end_date?: string
  }) => {
    return request.get('/merchant/api-usage', { params })
  },

  /**
   * 获取费率配置
   */
  getFeeConfig: () => {
    return request.get('/merchant/fee-config')
  },
}

export default profileService

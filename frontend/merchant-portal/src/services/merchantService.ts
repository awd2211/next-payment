import api from './api'

export interface Merchant {
  id: string
  name: string
  code: string
  type: string
  status: string
  contact_name: string
  contact_email: string
  contact_phone: string
  business_license: string
  website: string
  description: string
  api_key: string
  api_secret: string
  callback_url: string
  return_url: string
  settlement_cycle: number
  settlement_account: SettlementAccount
  rate_config: RateConfig
  risk_config: RiskConfig
  created_at: string
  updated_at: string
  approved_at: string
  approved_by: string
}

export interface SettlementAccount {
  bank_name: string
  bank_branch: string
  account_name: string
  account_number: string
  account_type: string
}

export interface RateConfig {
  channel: string
  payment_method: string
  rate: number
  fixed_fee: number
}

export interface RiskConfig {
  daily_limit: number
  monthly_limit: number
  single_limit: number
  ip_whitelist: string[]
  callback_retry: number
}

export interface MerchantBalance {
  available_balance: number
  frozen_balance: number
  total_balance: number
  currency: string
  updated_at: string
}

export interface MerchantStats {
  total_transactions: number
  total_amount: number
  success_rate: number
  today_transactions: number
  today_amount: number
  this_month_amount: number
}

export interface UpdateMerchantRequest {
  contact_name?: string
  contact_email?: string
  contact_phone?: string
  website?: string
  description?: string
  callback_url?: string
  return_url?: string
}

export interface RegenerateApiKeyResponse {
  api_key: string
  api_secret: string
}

export const merchantService = {
  getProfile: () => {
    return api.get<any, { data: Merchant }>('/merchant/profile')
  },

  updateProfile: (data: UpdateMerchantRequest) => {
    return api.put('/merchant/profile', data)
  },

  getBalance: () => {
    return api.get<any, { data: MerchantBalance }>('/merchant/balance')
  },

  getStats: (params: { start_time?: string; end_time?: string }) => {
    return api.get<any, { data: MerchantStats }>('/merchant/stats', { params })
  },

  regenerateApiKey: () => {
    return api.post<any, { data: RegenerateApiKeyResponse }>('/merchant/regenerate-api-key')
  },

  changePassword: (data: { old_password: string; new_password: string }) => {
    return api.post('/merchant/change-password', data)
  },
}

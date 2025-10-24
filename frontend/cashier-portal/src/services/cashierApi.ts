import axios from 'axios'
import type { CashierSession, CashierConfig } from '../types'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

export const cashierApi = {
  // 获取会话信息（不需要认证）
  getSession: async (sessionToken: string) => {
    const { data } = await api.get<{
      code: number
      data: CashierSession
      message: string
    }>(`/cashier/sessions/${sessionToken}`)
    return data.data
  },

  // 获取商户配置（通过会话获取，不需要认证）
  getConfig: async (merchantId: string) => {
    const { data } = await api.get<{
      code: number
      data: CashierConfig
      message: string
    }>('/cashier/configs', {
      params: { merchant_id: merchantId },
    })
    return data.data
  },

  // 记录用户行为日志
  recordLog: async (logData: {
    session_token: string
    user_ip?: string
    user_agent?: string
    device_type?: string
    browser?: string
    selected_channel?: string
    selected_method?: string
    form_filled?: boolean
    payment_submitted?: boolean
    page_load_time?: number
    time_to_submit?: number
    dropped_at_step?: string
    error_message?: string
  }) => {
    await api.post('/cashier/logs', logData)
  },

  // 创建支付（调用 Payment Gateway）
  createPayment: async (paymentData: {
    session_token: string
    channel: string
    payment_method: string
    card_data?: {
      number: string
      exp_month: string
      exp_year: string
      cvv: string
      holder_name: string
    }
  }) => {
    const { data } = await api.post<{
      code: number
      data: {
        payment_no: string
        payment_url?: string
        client_secret?: string
        status: string
      }
      message: string
    }>('/payments/create', paymentData)
    return data.data
  },
}

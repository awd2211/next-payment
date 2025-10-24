import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:40016/api/v1'

// 获取认证token
const getAuthToken = () => {
  return localStorage.getItem('token') || ''
}

// 创建axios实例
const cashierApi = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 添加token
cashierApi.interceptors.request.use(
  (config) => {
    const token = getAuthToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理错误
cashierApi.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      // Token过期,跳转登录
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// 类型定义
export interface CashierConfig {
  id: string
  merchant_id: string
  tenant_id: string
  theme_color: string
  logo_url: string
  background_image_url: string
  custom_css: string
  enabled_channels: string[]
  default_channel: string
  enabled_languages: string[]
  default_language: string
  auto_submit: boolean
  show_amount_breakdown: boolean
  allow_channel_switch: boolean
  session_timeout_minutes: number
  require_cvv: boolean
  enable_3d_secure: boolean
  allowed_countries: string[]
  success_redirect_url: string
  cancel_redirect_url: string
  created_at: string
  updated_at: string
}

export interface CashierSession {
  id: string
  session_token: string
  merchant_id: string
  order_no: string
  amount: number
  currency: string
  description: string
  customer_email: string
  customer_name: string
  customer_ip: string
  allowed_channels: string[]
  allowed_methods: string[]
  metadata: Record<string, any>
  status: 'pending' | 'active' | 'completed' | 'expired'
  payment_no: string
  created_at: string
  expires_at: string
  completed_at?: string
}

export interface CreateSessionInput {
  order_no: string
  amount: number
  currency: string
  description: string
  customer_email?: string
  customer_name?: string
  allowed_channels?: string[]
  allowed_methods?: string[]
  metadata?: Record<string, any>
  expires_in_minutes?: number
}

export interface AnalyticsData {
  conversion_rate: number
  channel_stats: Record<string, number>
  total_sessions: number
}

export interface CashierLog {
  id: string
  session_id: string
  merchant_id: string
  user_ip: string
  user_agent: string
  device_type: string
  browser: string
  selected_channel: string
  selected_method: string
  form_filled: boolean
  payment_submitted: boolean
  page_load_time: number
  time_to_submit: number
  dropped_at_step: string
  error_message: string
  created_at: string
}

// API服务
export const cashierService = {
  // 配置管理
  async getConfig(): Promise<{ code: number; data: CashierConfig; message: string }> {
    return cashierApi.get('/cashier/configs')
  },

  async createOrUpdateConfig(config: Partial<CashierConfig>): Promise<{ code: number; data: CashierConfig; message: string }> {
    return cashierApi.post('/cashier/configs', config)
  },

  async deleteConfig(): Promise<{ code: number; message: string }> {
    return cashierApi.delete('/cashier/configs')
  },

  // 会话管理
  async createSession(input: CreateSessionInput): Promise<{
    code: number
    data: {
      session_token: string
      session: CashierSession
      cashier_url: string
    }
    message: string
  }> {
    return cashierApi.post('/cashier/sessions', input)
  },

  async getSession(token: string): Promise<{ code: number; data: CashierSession; message: string }> {
    return cashierApi.get(`/cashier/sessions/${token}`)
  },

  async completeSession(token: string, paymentNo: string): Promise<{ code: number; message: string }> {
    return cashierApi.post(`/cashier/sessions/${token}/complete`, { payment_no: paymentNo })
  },

  async cancelSession(token: string): Promise<{ code: number; message: string }> {
    return cashierApi.delete(`/cashier/sessions/${token}`)
  },

  // 统计分析
  async getAnalytics(startTime?: string, endTime?: string): Promise<{ code: number; data: AnalyticsData; message: string }> {
    const params = new URLSearchParams()
    if (startTime) params.append('start_time', startTime)
    if (endTime) params.append('end_time', endTime)

    return cashierApi.get(`/cashier/analytics?${params.toString()}`)
  },
}

export default cashierService

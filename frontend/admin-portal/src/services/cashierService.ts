import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_CASHIER_API_BASE_URL || 'http://localhost:40016/api/v1'

const getAuthToken = () => {
  return localStorage.getItem('token') || ''
}

const cashierApi = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

cashierApi.interceptors.request.use(
  (config) => {
    const token = getAuthToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

cashierApi.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// 类型定义
export interface CashierTemplate {
  id: string
  name: string
  description: string
  config: Record<string, any>
  template_type: string
  preview_image_url: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface PlatformStats {
  total_merchants: number
  active_cashiers: number
  total_sessions: number
  avg_conversion_rate: number
  total_sessions_today: number
  completed_sessions_today: number
}

// API服务
export const adminCashierService = {
  // 模板管理
  async listTemplates(): Promise<{ code: number; data: CashierTemplate[]; message: string }> {
    return cashierApi.get('/admin/cashier/templates')
  },

  async createTemplate(template: Partial<CashierTemplate>): Promise<{ code: number; data: CashierTemplate; message: string }> {
    return cashierApi.post('/admin/cashier/templates', template)
  },

  async updateTemplate(id: string, template: Partial<CashierTemplate>): Promise<{ code: number; data: CashierTemplate; message: string }> {
    return cashierApi.put(`/admin/cashier/templates/${id}`, template)
  },

  async deleteTemplate(id: string): Promise<{ code: number; message: string }> {
    return cashierApi.delete(`/admin/cashier/templates/${id}`)
  },

  // 平台统计
  async getPlatformStats(): Promise<{ code: number; data: PlatformStats; message: string }> {
    return cashierApi.get('/admin/cashier/stats')
  },
}

export default adminCashierService

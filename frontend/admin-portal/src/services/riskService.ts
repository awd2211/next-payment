import request from './request'

export interface RiskRule {
  id: string
  name: string
  type: string
  condition: string
  action: string
  priority: number
  enabled: boolean
  description: string
  created_at: string
  updated_at: string
}

export interface RiskAlert {
  id: string
  rule_id: string
  rule_name: string
  merchant_id: string
  merchant_name: string
  payment_no: string
  order_no: string
  risk_level: string
  risk_score: number
  reason: string
  status: string
  handled_by?: string
  handled_at?: string
  created_at: string
}

export interface BlacklistItem {
  id: string
  type: string
  value: string
  reason: string
  merchant_id?: string
  expires_at?: string
  created_by: string
  created_at: string
}

export interface RiskStats {
  total_alerts: number
  high_risk_count: number
  medium_risk_count: number
  low_risk_count: number
  handled_count: number
  pending_count: number
  blacklist_count: number
}

export interface RiskRuleListParams {
  page: number
  page_size: number
  type?: string
  enabled?: boolean
}

export interface RiskAlertListParams {
  page: number
  page_size: number
  risk_level?: string
  status?: string
  merchant_id?: string
  start_time?: string
  end_time?: string
}

export interface BlacklistListParams {
  page: number
  page_size: number
  type?: string
  value?: string
  merchant_id?: string
}

export interface ListResponse<T> {
  data: T[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export const riskService = {
  // 风险规则管理
  listRules: (params: RiskRuleListParams) => {
    return request.get<ListResponse<RiskRule>>('/risk/rules', { params })
  },

  createRule: (data: Partial<RiskRule>) => {
    return request.post<RiskRule>('/risk/rules', data)
  },

  updateRule: (id: string, data: Partial<RiskRule>) => {
    return request.put<RiskRule>(`/risk/rules/${id}`, data)
  },

  deleteRule: (id: string) => {
    return request.delete(`/risk/rules/${id}`)
  },

  toggleRule: (id: string, enabled: boolean) => {
    return request.put(`/risk/rules/${id}/toggle`, { enabled })
  },

  // 风险告警管理
  listAlerts: (params: RiskAlertListParams) => {
    return request.get<ListResponse<RiskAlert>>('/risk/alerts', { params })
  },

  getAlert: (id: string) => {
    return request.get<RiskAlert>(`/risk/alerts/${id}`)
  },

  handleAlert: (id: string, action: string, remark: string) => {
    return request.post(`/risk/alerts/${id}/handle`, { action, remark })
  },

  // 黑名单管理
  listBlacklist: (params: BlacklistListParams) => {
    return request.get<ListResponse<BlacklistItem>>('/risk/blacklist', { params })
  },

  addBlacklist: (data: Partial<BlacklistItem>) => {
    return request.post<BlacklistItem>('/risk/blacklist', data)
  },

  removeBlacklist: (id: string) => {
    return request.delete(`/risk/blacklist/${id}`)
  },

  // 风险统计
  getStats: () => {
    return request.get<{ data: RiskStats }>('/risk/stats')
  },
}

export default riskService

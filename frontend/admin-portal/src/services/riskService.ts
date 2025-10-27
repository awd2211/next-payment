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
    return request.get<ListResponse<RiskRule>>('/api/v1/rules', { params })
  },

  createRule: (data: Partial<RiskRule>) => {
    return request.post<RiskRule>('/api/v1/rules', data)
  },

  updateRule: (id: string, data: Partial<RiskRule>) => {
    return request.put<RiskRule>(`/api/v1/admin/rules/${id}`, data)
  },

  deleteRule: (id: string) => {
    return request.delete(`/api/v1/admin/rules/${id}`)
  },

  // 切换规则状态 - 对应后端的enable/disable接口
  toggleRule: (id: string, enabled: boolean) => {
    const action = enabled ? 'enable' : 'disable'
    return request.post(`/api/v1/admin/rules/${id}/${action}`)
  },

  // 风险检查记录管理 (后端使用/checks而不是/alerts)
  listAlerts: (params: RiskAlertListParams) => {
    return request.get<ListResponse<RiskAlert>>('/api/v1/checks', { params })
  },

  getAlert: (id: string) => {
    return request.get<RiskAlert>(`/api/v1/admin/checks/${id}`)
  },

  // 处理告警 - 注意: 后端需要实现此接口
  handleAlert: (id: string, action: string, remark: string) => {
    return request.post(`/api/v1/admin/checks/${id}/handle`, { action, remark })
  },

  // 黑名单管理
  listBlacklist: (params: BlacklistListParams) => {
    return request.get<ListResponse<BlacklistItem>>('/api/v1/blacklist', { params })
  },

  addBlacklist: (data: Partial<BlacklistItem>) => {
    return request.post<BlacklistItem>('/api/v1/blacklist', data)
  },

  removeBlacklist: (id: string) => {
    return request.delete(`/api/v1/blacklist/${id}`)
  },

  // 风险统计 - 注意: 后端需要实现此接口
  getStats: () => {
    return request.get<{ data: RiskStats }>('/api/v1/risk/stats')
  },
}

export default riskService

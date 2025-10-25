import request from './request'

// Types
export interface RiskAlert {
  id: string
  rule_id: string
  rule_name: string
  merchant_id: string
  payment_no: string
  order_no: string
  risk_level: 'low' | 'medium' | 'high' | 'critical'
  risk_score: number
  reason: string
  status: 'pending' | 'handled' | 'ignored'
  handled_by?: string
  handled_at?: string
  created_at: string
}

export interface RiskAssessment {
  payment_no: string
  order_no: string
  risk_level: string
  risk_score: number
  factors: Array<{
    factor: string
    score: number
    description: string
  }>
  decision: 'approve' | 'reject' | 'review'
  assessed_at: string
}

export interface RiskStats {
  total_alerts: number
  high_risk_count: number
  medium_risk_count: number
  low_risk_count: number
  recent_alerts: RiskAlert[]
}

export interface ListRiskAlertsParams {
  page?: number
  page_size?: number
  risk_level?: string
  status?: string
  start_date?: string
  end_date?: string
  payment_no?: string
}

export interface ListRiskAlertsResponse {
  code: number
  message: string
  data: {
    list: RiskAlert[]
    total: number
    page: number
    page_size: number
  }
}

export interface RiskStatsResponse {
  code: number
  message: string
  data: RiskStats
}

// API Methods
export const riskService = {
  /**
   * Get merchant's risk alerts
   */
  listAlerts: (params: ListRiskAlertsParams) => {
    return request.get<ListRiskAlertsResponse>('/merchant/risk/alerts', { params })
  },

  /**
   * Get risk alert detail
   */
  getAlert: (id: string) => {
    return request.get<RiskAlert>(`/merchant/risk/alerts/${id}`)
  },

  /**
   * Get risk assessment for a payment
   */
  getAssessment: (paymentNo: string) => {
    return request.get<RiskAssessment>(`/merchant/risk/assessments/${paymentNo}`)
  },

  /**
   * Get merchant risk statistics
   */
  getStats: (params?: { start_date?: string; end_date?: string }) => {
    return request.get<RiskStatsResponse>('/merchant/risk/stats', { params })
  },

  /**
   * Appeal a risk decision
   */
  appeal: (alertId: string, data: { reason: string; attachments?: string[] }) => {
    return request.post(`/merchant/risk/alerts/${alertId}/appeal`, data)
  },

  /**
   * Export risk alerts
   */
  export: (params: ListRiskAlertsParams) => {
    return request.get('/merchant/risk/alerts/export', {
      params,
      responseType: 'blob',
    })
  },

  /**
   * Get risk trend chart data
   */
  getTrend: (params: { start_date: string; end_date: string; interval?: 'day' | 'week' | 'month' }) => {
    return request.get('/merchant/risk/trend', { params })
  },
}

export default riskService

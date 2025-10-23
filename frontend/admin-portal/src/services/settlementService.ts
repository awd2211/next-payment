import api from './api'

export interface Settlement {
  id: string
  settlement_no: string
  merchant_id: string
  merchant_name: string
  settlement_period: string
  settlement_date: string
  transaction_count: number
  settlement_amount: number
  fee_amount: number
  actual_amount: number
  status: string
  bank_name: string
  bank_account: string
  account_holder: string
  remark: string
  created_at: string
  updated_at: string
  completed_at?: string
}

export interface SettlementStats {
  total_settlements: number
  pending_amount: number
  processing_amount: number
  completed_amount: number
  this_month_amount: number
}

export interface SettlementListParams {
  page: number
  page_size: number
  settlement_no?: string
  merchant_id?: string
  status?: string
  start_date?: string
  end_date?: string
}

export interface SettlementListResponse {
  data: Settlement[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface SettlementStatsParams {
  merchant_id?: string
  start_date?: string
  end_date?: string
}

export const settlementService = {
  // 获取结算列表
  list: (params: SettlementListParams) => {
    return api.get<SettlementListResponse>('/settlements', { params })
  },

  // 获取结算详情
  get: (id: string) => {
    return api.get<Settlement>(`/settlements/${id}`)
  },

  // 获取结算统计
  getStats: (params: SettlementStatsParams) => {
    return api.get<{ data: SettlementStats }>('/settlements/stats', { params })
  },

  // 创建结算单
  create: (data: Partial<Settlement>) => {
    return api.post<Settlement>('/settlements', data)
  },

  // 更新结算单
  update: (id: string, data: Partial<Settlement>) => {
    return api.put<Settlement>(`/settlements/${id}`, data)
  },

  // 确认结算
  confirm: (id: string) => {
    return api.post(`/settlements/${id}/confirm`)
  },

  // 完成结算
  complete: (id: string, remark: string) => {
    return api.post(`/settlements/${id}/complete`, { remark })
  },

  // 取消结算
  cancel: (id: string, reason: string) => {
    return api.post(`/settlements/${id}/cancel`, { reason })
  },

  // 导出结算数据
  export: (params: SettlementListParams) => {
    return api.get('/settlements/export', { params, responseType: 'blob' })
  },
}

export default settlementService

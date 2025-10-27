import request from './request'

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
    return request.get<SettlementListResponse>('/merchant/settlements', { params })
  },

  // 获取结算详情
  get: (id: string) => {
    return request.get<Settlement>(`/settlements/${id}`)
  },

  // 获取结算统计
  getStats: (params: SettlementStatsParams) => {
    return request.get<SettlementStats>('/merchant/settlements/stats', { params })
  },

  // 创建结算单
  create: (data: Partial<Settlement>) => {
    return request.post<Settlement>('/merchant/settlements', data)
  },

  // 更新结算单
  update: (id: string, data: Partial<Settlement>) => {
    return request.put<Settlement>(`/settlements/${id}`, data)
  },

  // 确认结算
  confirm: (id: string) => {
    return request.post(`/settlements/${id}/confirm`)
  },

  // 完成结算
  complete: (id: string, remark: string) => {
    return request.post(`/settlements/${id}/complete`, { remark })
  },

  // 取消结算
  cancel: (id: string, reason: string) => {
    return request.post(`/settlements/${id}/cancel`, { reason })
  },

  // 导出结算数据
  export: (params: SettlementListParams) => {
    return request.get('/merchant/settlements/export', { params, responseType: 'blob' })
  },
}

export default settlementService

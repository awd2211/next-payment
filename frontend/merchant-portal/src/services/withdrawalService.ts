import request from './request'

export interface BankAccount {
  bank_name: string
  bank_code: string
  bank_branch?: string
  account_name: string
  account_number: string
  swift_code?: string
  iban?: string
}

export interface Withdrawal {
  id: string
  withdrawal_no: string
  merchant_id: string
  merchant_name: string
  amount: number
  currency: string
  fee: number
  actual_amount: number
  status: 'pending' | 'approved' | 'rejected' | 'processing' | 'completed' | 'failed'
  bank_account: BankAccount
  remark?: string
  reject_reason?: string
  transaction_id?: string
  approved_by?: string
  approved_at?: string
  processed_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export interface ListWithdrawalsParams {
  page?: number
  page_size?: number
  status?: string
  merchant_id?: string
  currency?: string
  start_time?: string
  end_time?: string
  withdrawal_no?: string
}

export interface ListWithdrawalsResponse {
  list: Withdrawal[]
  total: number
  page: number
  page_size: number
}

export interface ApproveWithdrawalRequest {
  remark?: string
}

export interface RejectWithdrawalRequest {
  reason: string
  remark?: string
}

export interface ProcessWithdrawalRequest {
  transaction_id: string
  remark?: string
}

export interface WithdrawalStats {
  total_count: number
  pending_count: number
  approved_count: number
  rejected_count: number
  processing_count: number
  completed_count: number
  failed_count: number
  total_amount: number
  pending_amount: number
  completed_amount: number
}

export const withdrawalService = {
  /**
   * 获取提现申请列表
   */
  list: (params: ListWithdrawalsParams) => {
    return request.get<ListWithdrawalsResponse>('/withdrawals', { params })
  },

  /**
   * 获取单个提现申请详情
   */
  getById: (id: string) => {
    return request.get<Withdrawal>(`/withdrawals/${id}`)
  },

  /**
   * 批准提现申请
   */
  approve: (id: string, data: ApproveWithdrawalRequest) => {
    return request.post(`/withdrawals/${id}/approve`, data)
  },

  /**
   * 拒绝提现申请
   */
  reject: (id: string, data: RejectWithdrawalRequest) => {
    return request.post(`/withdrawals/${id}/reject`, data)
  },

  /**
   * 处理提现（标记为处理中）
   */
  process: (id: string, data: ProcessWithdrawalRequest) => {
    return request.post(`/withdrawals/${id}/process`, data)
  },

  /**
   * 完成提现
   */
  complete: (id: string) => {
    return request.post(`/withdrawals/${id}/complete`)
  },

  /**
   * 标记提现失败
   */
  fail: (id: string, reason: string) => {
    return request.post(`/withdrawals/${id}/fail`, { reason })
  },

  /**
   * 获取提现统计信息
   */
  getStats: (params?: { start_time?: string; end_time?: string; currency?: string }) => {
    return request.get<WithdrawalStats>('/withdrawals/stats', { params })
  },

  /**
   * 批量批准提现申请
   */
  batchApprove: (ids: string[], remark?: string) => {
    return request.post('/withdrawals/batch/approve', { ids, remark })
  },

  /**
   * 导出提现记录
   */
  export: (params: ListWithdrawalsParams) => {
    return request.download('/withdrawals/export', 'withdrawals.xlsx', { params })
  },
}

export default withdrawalService

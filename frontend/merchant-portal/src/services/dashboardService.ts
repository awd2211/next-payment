import request from './request'

export interface DashboardData {
  total_transactions: number
  total_amount: number
  today_transactions: number
  today_amount: number
  pending_withdrawals: number
  available_balance: number
  // 根据后端实际返回结构调整
}

export interface TransactionSummary {
  total_count: number
  total_amount: number
  success_count: number
  failed_count: number
  // 根据后端实际返回结构调整
}

export interface BalanceInfo {
  available_balance: number
  frozen_balance: number
  total_balance: number
  currency: string
}

export const dashboardService = {
  // 获取Dashboard概览数据
  getDashboard: () => {
    return request.get<{ data: DashboardData }>('/dashboard')
  },

  // 获取交易汇总
  getTransactionSummary: (params: { start_date?: string; end_date?: string }) => {
    return request.get<{ data: TransactionSummary }>('/dashboard/transaction-summary', { params })
  },

  // 获取余额信息
  getBalanceInfo: () => {
    return request.get<{ data: BalanceInfo }>('/dashboard/balance')
  },
}

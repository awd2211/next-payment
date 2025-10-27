import request from './request'

export interface DashboardData {
  total_transactions: number
  total_amount: number
  today_transactions: number
  today_amount: number
  today_payments: number
  today_success_rate: number
  month_payments: number
  month_amount: number
  month_success_rate: number
  payment_trend: Array<{ date: string; amount: number; count: number }>
  pending_withdrawals: number
  available_balance: number
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
    return request.get<DashboardData>('/merchant/dashboard')
  },

  // 获取交易汇总
  getTransactionSummary: (params: { start_date?: string; end_date?: string }) => {
    return request.get<TransactionSummary>('/merchant/dashboard/transaction-summary', { params })
  },

  // 获取余额信息
  getBalanceInfo: () => {
    return request.get<BalanceInfo>('/merchant/dashboard/balance')
  },
}

import request from './request'

export interface AccountingEntry {
  id: string
  entry_no: string
  account_date: string
  debit_account: string
  debit_account_name: string
  credit_account: string
  credit_account_name: string
  amount: number
  currency: string
  description: string
  reference_type: 'payment' | 'refund' | 'withdrawal' | 'settlement' | 'adjustment'
  reference_no: string
  merchant_id?: string
  merchant_name?: string
  posted_by: string
  posted_at: string
  created_at: string
}

export interface AccountBalance {
  account_code: string
  account_name: string
  account_type: 'asset' | 'liability' | 'equity' | 'revenue' | 'expense'
  currency: string
  debit_balance: number
  credit_balance: number
  net_balance: number
  updated_at: string
}

export interface Ledger {
  account_code: string
  account_name: string
  entries: LedgerEntry[]
  opening_balance: number
  closing_balance: number
  total_debit: number
  total_credit: number
}

export interface LedgerEntry {
  entry_no: string
  account_date: string
  description: string
  debit_amount: number
  credit_amount: number
  balance: number
  reference_no: string
}

export interface ListEntriesParams {
  page?: number
  page_size?: number
  start_date?: string
  end_date?: string
  account_code?: string
  reference_type?: string
  reference_no?: string
  merchant_id?: string
  currency?: string
}

export interface ListEntriesResponse {
  list: AccountingEntry[]
  total: number
  page: number
  page_size: number
}

export interface ListBalancesParams {
  account_type?: string
  currency?: string
  account_code?: string
}

export interface ListBalancesResponse {
  list: AccountBalance[]
}

export interface GetLedgerParams {
  account_code: string
  start_date: string
  end_date: string
  currency?: string
}

export interface CreateEntryRequest {
  account_date: string
  debit_account: string
  credit_account: string
  amount: number
  currency: string
  description: string
  reference_type: 'payment' | 'refund' | 'withdrawal' | 'settlement' | 'adjustment'
  reference_no: string
  merchant_id?: string
}

export interface AccountingSummary {
  total_revenue: number
  total_expense: number
  net_income: number
  total_assets: number
  total_liabilities: number
  currency: string
  period_start: string
  period_end: string
}

export const accountingService = {
  /**
   * 获取会计分录列表
   */
  listEntries: (params: ListEntriesParams) => {
    return request.get<ListEntriesResponse>('/accounting/entries', { params })
  },

  /**
   * 获取单个会计分录详情
   */
  getEntryById: (id: string) => {
    return request.get<AccountingEntry>(`/accounting/entries/${id}`)
  },

  /**
   * 创建会计分录（手工调整）
   */
  createEntry: (data: CreateEntryRequest) => {
    return request.post<AccountingEntry>('/accounting/entries', data)
  },

  /**
   * 获取账户余额表
   */
  listBalances: (params: ListBalancesParams) => {
    return request.get<ListBalancesResponse>('/accounting/balances', { params })
  },

  /**
   * 获取账户明细账
   */
  getLedger: (params: GetLedgerParams) => {
    return request.get<Ledger>('/accounting/ledger', { params })
  },

  /**
   * 获取总账
   */
  getGeneralLedger: (params: { start_date: string; end_date: string; currency?: string }) => {
    return request.get<Ledger[]>('/accounting/general-ledger', { params })
  },

  /**
   * 获取会计汇总报表
   */
  getSummary: (params: { start_date: string; end_date: string; currency?: string; merchant_id?: string }) => {
    return request.get<AccountingSummary>('/accounting/summary', { params })
  },

  /**
   * 获取资产负债表
   */
  getBalanceSheet: (params: { as_of_date: string; currency?: string }) => {
    return request.get('/accounting/balance-sheet', { params })
  },

  /**
   * 获取损益表
   */
  getIncomeStatement: (params: { start_date: string; end_date: string; currency?: string }) => {
    return request.get('/accounting/income-statement', { params })
  },

  /**
   * 获取现金流量表
   */
  getCashFlowStatement: (params: { start_date: string; end_date: string; currency?: string }) => {
    return request.get('/accounting/cash-flow', { params })
  },

  /**
   * 导出会计分录
   */
  exportEntries: (params: ListEntriesParams) => {
    return request.download('/accounting/entries/export', 'accounting-entries.xlsx', { params })
  },

  /**
   * 导出账户余额表
   */
  exportBalances: (params: ListBalancesParams) => {
    return request.download('/accounting/balances/export', 'account-balances.xlsx', { params })
  },

  /**
   * 关账（月末结账）
   */
  closeMonth: (params: { year: number; month: number; currency: string }) => {
    return request.post('/accounting/close-month', params)
  },

  /**
   * 获取科目表
   */
  getChartOfAccounts: () => {
    return request.get('/accounting/chart-of-accounts')
  },
}

export default accountingService

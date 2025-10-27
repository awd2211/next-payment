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

export interface Account {
  id: string
  account_code: string
  account_name: string
  account_type: string
  currency: string
  merchant_id?: string
  balance: number
  frozen_balance: number
  available_balance: number
  is_frozen: boolean
  created_at: string
  updated_at: string
}

export interface Transaction {
  id: string
  transaction_no: string
  transaction_type: string
  amount: number
  currency: string
  status: string
  merchant_id?: string
  description?: string
  created_at: string
  updated_at: string
}

export interface Settlement {
  id: string
  settlement_no: string
  merchant_id: string
  amount: number
  currency: string
  status: string
  settled_at?: string
  created_at: string
}

export interface Withdrawal {
  id: string
  withdrawal_no: string
  merchant_id: string
  amount: number
  currency: string
  status: string
  bank_account?: string
  processed_at?: string
  created_at: string
}

export interface Invoice {
  id: string
  invoice_no: string
  merchant_id: string
  amount: number
  currency: string
  status: string
  due_date: string
  paid_at?: string
  created_at: string
}

export interface Reconciliation {
  id: string
  reconciliation_no: string
  start_date: string
  end_date: string
  status: string
  total_matched: number
  total_unmatched: number
  created_at: string
}

export interface CurrencyConversion {
  id: string
  conversion_no: string
  from_currency: string
  to_currency: string
  from_amount: number
  to_amount: number
  exchange_rate: number
  status: string
  created_at: string
}

export const accountingService = {
  // Account Management
  createAccount: (data: Partial<Account>) => {
    return request.post<Account>('/merchant/accounts', data)
  },

  getAccount: (id: string) => {
    return request.get<Account>(`/accounts/${id}`)
  },

  listAccounts: (params?: { page?: number; page_size?: number; merchant_id?: string }) => {
    return request.get('/merchant/accounts', { params })
  },

  freezeAccount: (id: string) => {
    return request.post(`/accounts/${id}/freeze`)
  },

  unfreezeAccount: (id: string) => {
    return request.post(`/accounts/${id}/unfreeze`)
  },

  // Transaction Management (Double-entry Accounting)
  createTransaction: (data: {
    debit_account_id: string
    credit_account_id: string
    amount: number
    currency: string
    transaction_type: string
    description?: string
  }) => {
    return request.post<Transaction>('/merchant/transactions', data)
  },

  getTransaction: (transactionNo: string) => {
    return request.get<Transaction>(`/transactions/${transactionNo}`)
  },

  listTransactions: (params?: {
    page?: number
    page_size?: number
    merchant_id?: string
    transaction_type?: string
  }) => {
    return request.get('/merchant/transactions', { params })
  },

  reverseTransaction: (transactionNo: string, reason: string) => {
    return request.post(`/transactions/${transactionNo}/reverse`, { reason })
  },

  // Settlement Management
  createSettlement: (data: {
    merchant_id: string
    amount: number
    currency: string
    settlement_date?: string
  }) => {
    return request.post<Settlement>('/merchant/settlements', data)
  },

  getSettlement: (settlementNo: string) => {
    return request.get<Settlement>(`/settlements/${settlementNo}`)
  },

  listSettlements: (params?: {
    page?: number
    page_size?: number
    merchant_id?: string
    status?: string
  }) => {
    return request.get('/merchant/settlements', { params })
  },

  processSettlement: (settlementNo: string) => {
    return request.post(`/settlements/${settlementNo}/process`)
  },

  // Withdrawal Management
  createWithdrawal: (data: {
    merchant_id: string
    amount: number
    currency: string
    bank_account_id: string
  }) => {
    return request.post<Withdrawal>('/merchant/withdrawals', data)
  },

  getWithdrawal: (withdrawalNo: string) => {
    return request.get<Withdrawal>(`/withdrawals/${withdrawalNo}`)
  },

  listWithdrawals: (params?: {
    page?: number
    page_size?: number
    merchant_id?: string
    status?: string
  }) => {
    return request.get('/merchant/withdrawals', { params })
  },

  approveWithdrawal: (withdrawalNo: string) => {
    return request.post(`/withdrawals/${withdrawalNo}/approve`)
  },

  rejectWithdrawal: (withdrawalNo: string, reason: string) => {
    return request.post(`/withdrawals/${withdrawalNo}/reject`, { reason })
  },

  processWithdrawal: (withdrawalNo: string) => {
    return request.post(`/withdrawals/${withdrawalNo}/process`)
  },

  completeWithdrawal: (withdrawalNo: string) => {
    return request.post(`/withdrawals/${withdrawalNo}/complete`)
  },

  failWithdrawal: (withdrawalNo: string, reason: string) => {
    return request.post(`/withdrawals/${withdrawalNo}/fail`, { reason })
  },

  cancelWithdrawal: (withdrawalNo: string) => {
    return request.post(`/withdrawals/${withdrawalNo}/cancel`)
  },

  // Invoice Management
  createInvoice: (data: {
    merchant_id: string
    amount: number
    currency: string
    due_date: string
    items?: any[]
  }) => {
    return request.post<Invoice>('/merchant/invoices', data)
  },

  getInvoice: (invoiceNo: string) => {
    return request.get<Invoice>(`/invoices/${invoiceNo}`)
  },

  listInvoices: (params?: {
    page?: number
    page_size?: number
    merchant_id?: string
    status?: string
  }) => {
    return request.get('/merchant/invoices', { params })
  },

  payInvoice: (invoiceNo: string, payment_method?: string) => {
    return request.post(`/invoices/${invoiceNo}/pay`, { payment_method })
  },

  cancelInvoice: (invoiceNo: string) => {
    return request.post(`/invoices/${invoiceNo}/cancel`)
  },

  voidInvoice: (invoiceNo: string) => {
    return request.post(`/invoices/${invoiceNo}/void`)
  },

  // Reconciliation
  createReconciliation: (data: {
    start_date: string
    end_date: string
    merchant_id?: string
  }) => {
    return request.post<Reconciliation>('/merchant/reconciliations', data)
  },

  getReconciliation: (reconciliationNo: string) => {
    return request.get<Reconciliation>(`/reconciliations/${reconciliationNo}`)
  },

  listReconciliations: (params?: {
    page?: number
    page_size?: number
    status?: string
  }) => {
    return request.get('/merchant/reconciliations', { params })
  },

  processReconciliation: (reconciliationNo: string) => {
    return request.post(`/reconciliations/${reconciliationNo}/process`)
  },

  completeReconciliation: (reconciliationNo: string) => {
    return request.post(`/reconciliations/${reconciliationNo}/complete`)
  },

  resolveReconciliationItem: (itemId: string, resolution: string) => {
    return request.post(`/reconciliations/items/${itemId}/resolve`, { resolution })
  },

  // Balance Inquiry (Aggregated)
  getMerchantBalanceSummary: (merchantId: string) => {
    return request.get(`/balances/merchants/${merchantId}/summary`)
  },

  getMerchantBalanceByCurrency: (merchantId: string, currency: string) => {
    return request.get(`/balances/merchants/${merchantId}/currencies/${currency}`)
  },

  getMerchantBalanceByAccountType: (merchantId: string, accountType: string) => {
    return request.get(`/balances/merchants/${merchantId}/account-types/${accountType}`)
  },

  getAllCurrencyBalances: (merchantId: string) => {
    return request.get(`/balances/merchants/${merchantId}/currencies`)
  },

  // Currency Conversion
  createConversion: (data: {
    merchant_id: string
    from_currency: string
    to_currency: string
    from_amount: number
  }) => {
    return request.post<CurrencyConversion>('/merchant/conversions', data)
  },

  getConversion: (conversionNo: string) => {
    return request.get<CurrencyConversion>(`/conversions/${conversionNo}`)
  },

  listConversions: (params?: {
    page?: number
    page_size?: number
    merchant_id?: string
  }) => {
    return request.get('/merchant/conversions', { params })
  },

  processConversion: (conversionNo: string) => {
    return request.post(`/conversions/${conversionNo}/process`)
  },

  cancelConversion: (conversionNo: string) => {
    return request.post(`/conversions/${conversionNo}/cancel`)
  },

  // Legacy/Existing APIs
  listEntries: (params: ListEntriesParams) => {
    return request.get<ListEntriesResponse>('/merchant/accounting/entries', { params })
  },

  getEntryById: (id: string) => {
    return request.get<AccountingEntry>(`/accounting/entries/${id}`)
  },

  createEntry: (data: CreateEntryRequest) => {
    return request.post<AccountingEntry>('/merchant/accounting/entries', data)
  },

  listBalances: (params: ListBalancesParams) => {
    return request.get<ListBalancesResponse>('/merchant/accounting/balances', { params })
  },

  getLedger: (params: GetLedgerParams) => {
    return request.get<Ledger>('/merchant/accounting/ledger', { params })
  },

  getGeneralLedger: (params: { start_date: string; end_date: string; currency?: string }) => {
    return request.get<Ledger[]>('/merchant/accounting/general-ledger', { params })
  },

  getSummary: (params: { start_date: string; end_date: string; currency?: string; merchant_id?: string }) => {
    return request.get<AccountingSummary>('/merchant/accounting/summary', { params })
  },

  getBalanceSheet: (params: { as_of_date: string; currency?: string }) => {
    return request.get('/merchant/accounting/balance-sheet', { params })
  },

  getIncomeStatement: (params: { start_date: string; end_date: string; currency?: string }) => {
    return request.get('/merchant/accounting/income-statement', { params })
  },

  getCashFlowStatement: (params: { start_date: string; end_date: string; currency?: string }) => {
    return request.get('/merchant/accounting/cash-flow', { params })
  },

  exportEntries: (params: ListEntriesParams) => {
    return request.download('/merchant/accounting/entries/export', 'accounting-entries.xlsx', { params })
  },

  exportBalances: (params: ListBalancesParams) => {
    return request.download('/merchant/accounting/balances/export', 'account-balances.xlsx', { params })
  },

  closeMonth: (params: { year: number; month: number; currency: string }) => {
    return request.post('/merchant/accounting/close-month', params)
  },

  getChartOfAccounts: () => {
    return request.get('/merchant/accounting/chart-of-accounts')
  },
}

export default accountingService

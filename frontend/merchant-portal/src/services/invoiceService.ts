import request from './request'

// Types
export interface Invoice {
  id: string
  invoice_no: string
  merchant_id: string
  merchant_name: string
  billing_period_start: string
  billing_period_end: string
  total_amount: number
  currency: string
  status: 'pending' | 'paid' | 'overdue' | 'cancelled'
  due_date: string
  paid_at?: string
  payment_method?: string
  items: InvoiceItem[]
  created_at: string
  updated_at: string
}

export interface InvoiceItem {
  id: string
  invoice_id: string
  description: string
  quantity: number
  unit_price: number
  amount: number
  item_type: 'payment_fee' | 'refund_fee' | 'monthly_fee' | 'service_fee' | 'other'
}

export interface InvoiceSummary {
  total_invoices: number
  total_amount: number
  paid_amount: number
  unpaid_amount: number
  overdue_amount: number
  currency: string
}

export interface ListInvoicesParams {
  page?: number
  page_size?: number
  status?: string
  start_date?: string
  end_date?: string
  invoice_no?: string
}

export interface ListInvoicesResponse {
  code: number
  message: string
  data: {
    list: Invoice[]
    total: number
    page: number
    page_size: number
  }
}

export interface InvoiceDetailResponse {
  code: number
  message: string
  data: Invoice
}

// API Methods
export const invoiceService = {
  /**
   * Get merchant invoices list
   */
  list: (params: ListInvoicesParams) => {
    return request.get<ListInvoicesResponse>('/merchant/invoices', { params })
  },

  /**
   * Get invoice detail by invoice number
   */
  getDetail: (invoiceNo: string) => {
    return request.get<InvoiceDetailResponse>(`/merchant/invoices/${invoiceNo}`)
  },

  /**
   * Get invoice summary
   */
  getSummary: (params?: { start_date?: string; end_date?: string }) => {
    return request.get<InvoiceSummary>('/merchant/invoices/summary', { params })
  },

  /**
   * Pay invoice
   */
  pay: (invoiceNo: string, data: { payment_method: string }) => {
    return request.post(`/merchant/invoices/${invoiceNo}/pay`, data)
  },

  /**
   * Download invoice PDF
   */
  download: (invoiceNo: string) => {
    return request.get(`/merchant/invoices/${invoiceNo}/download`, {
      responseType: 'blob',
    })
  },

  /**
   * Export invoices to Excel
   */
  export: (params: ListInvoicesParams) => {
    return request.get('/merchant/invoices/export', {
      params,
      responseType: 'blob',
    })
  },

  /**
   * Get invoice items detail
   */
  getItems: (invoiceNo: string) => {
    return request.get<InvoiceItem[]>(`/merchant/invoices/${invoiceNo}/items`)
  },

  /**
   * Get upcoming invoice preview
   */
  getUpcoming: () => {
    return request.get<Invoice>('/merchant/invoices/upcoming')
  },

  /**
   * Dispute an invoice
   */
  dispute: (invoiceNo: string, data: { reason: string; attachments?: string[] }) => {
    return request.post(`/merchant/invoices/${invoiceNo}/dispute`, data)
  },
}

export default invoiceService

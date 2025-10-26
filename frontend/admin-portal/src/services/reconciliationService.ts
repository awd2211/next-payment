import request from './request'

// Types
export interface ReconciliationRecord {
  id: string
  recon_no: string
  merchant_name: string
  merchant_id: string
  channel: string
  recon_date: string
  total_count: number
  matched_count: number
  unmatched_count: number
  platform_total_amount: number
  channel_total_amount: number
  difference_amount: number
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'confirmed'
  created_at: string
  completed_at?: string
}

export interface UnmatchedItem {
  id: string
  recon_id: string
  payment_no: string
  order_no: string
  amount: number
  currency: string
  platform_time: string
  channel_time?: string
  reason: string
  status: 'pending' | 'resolved' | 'ignored'
}

export interface ListReconciliationParams {
  page?: number
  page_size?: number
  merchant_id?: string
  channel?: string
  status?: string
  start_date?: string
  end_date?: string
}

export interface ListReconciliationResponse {
  code: number
  message: string
  data: {
    list: ReconciliationRecord[]
    total: number
    page: number
    page_size: number
  }
}

export interface ReconciliationDetailResponse {
  code: number
  message: string
  data: ReconciliationRecord
}

export interface UnmatchedItemsResponse {
  code: number
  message: string
  data: {
    list: UnmatchedItem[]
    total: number
  }
}

export interface CreateReconciliationRequest {
  merchant_id?: string
  channel: string
  recon_date: string
  file?: File
}

export interface CreateReconciliationResponse {
  code: number
  message: string
  data: {
    recon_id: string
    recon_no: string
  }
}

export interface ConfirmReconciliationRequest {
  notes?: string
}

export interface ConfirmReconciliationResponse {
  code: number
  message: string
  data: {
    recon_id: string
    status: string
  }
}

// API Methods
export const reconciliationService = {
  /**
   * Get reconciliation records list
   */
  list: (params: ListReconciliationParams) => {
    return request.get<ListReconciliationResponse>('/api/v1/reconciliation/tasks', { params })
  },

  /**
   * Get reconciliation detail by ID
   */
  getDetail: (id: string) => {
    return request.get<ReconciliationDetailResponse>(`/api/v1/reconciliation/tasks/${id}`)
  },

  /**
   * Get unmatched items for a reconciliation - 对应后端的records接口
   */
  getUnmatchedItems: (reconId: string) => {
    return request.get<UnmatchedItemsResponse>(`/api/v1/reconciliation/records`, {
      params: { task_id: reconId }
    })
  },

  /**
   * Create new reconciliation task
   */
  create: (data: CreateReconciliationRequest) => {
    const formData = new FormData()
    if (data.merchant_id) formData.append('merchant_id', data.merchant_id)
    formData.append('channel', data.channel)
    formData.append('recon_date', data.recon_date)
    if (data.file) formData.append('file', data.file)

    return request.post<CreateReconciliationResponse>('/api/v1/reconciliation/tasks', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  /**
   * Confirm reconciliation result - 注意: 后端需要实现此接口
   */
  confirm: (id: string, data: ConfirmReconciliationRequest) => {
    return request.post<ConfirmReconciliationResponse>(`/api/v1/reconciliation/tasks/${id}/confirm`, data)
  },

  /**
   * Retry failed reconciliation
   */
  retry: (id: string) => {
    return request.post(`/api/v1/reconciliation/tasks/${id}/retry`)
  },

  /**
   * Download reconciliation report
   */
  downloadReport: (id: string) => {
    return request.get(`/api/v1/reconciliation/reports/${id}`, {
      responseType: 'blob',
    })
  },

  /**
   * Export reconciliation records - 注意: 后端需要实现此接口
   */
  export: (params: ListReconciliationParams) => {
    return request.get('/api/v1/reconciliation/tasks/export', {
      params,
      responseType: 'blob',
    })
  },

  /**
   * Get reconciliation statistics - 注意: 后端需要实现此接口
   */
  getStats: (params?: { start_date?: string; end_date?: string }) => {
    return request.get('/api/v1/reconciliation/stats', { params })
  },

  /**
   * Resolve unmatched item
   */
  resolveUnmatched: (reconId: string, itemId: string, data: { action: 'resolve' | 'ignore'; notes?: string }) => {
    return request.post(`/api/v1/reconciliation/records/${itemId}/resolve`, data)
  },
}

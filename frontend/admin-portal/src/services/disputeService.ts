import request from './request'

// Types
export interface Dispute {
  id: string
  dispute_no: string
  payment_no: string
  order_no: string
  merchant_name: string
  merchant_id: string
  amount: number
  currency: string
  reason: string
  status: 'pending' | 'reviewing' | 'accepted' | 'rejected' | 'withdrawn'
  evidence_deadline: string
  submitted_at: string
  resolved_at?: string
  created_by: string
}

export interface Evidence {
  id: string
  dispute_id: string
  file_name: string
  file_type: string
  file_url: string
  uploaded_at: string
  uploaded_by: string
}

export interface ListDisputesParams {
  page?: number
  page_size?: number
  status?: string
  merchant_id?: string
  dispute_no?: string
  payment_no?: string
  start_date?: string
  end_date?: string
}

export interface ListDisputesResponse {
  code: number
  message: string
  data: {
    list: Dispute[]
    total: number
    page: number
    page_size: number
  }
}

export interface DisputeDetailResponse {
  code: number
  message: string
  data: Dispute
}

export interface EvidenceListResponse {
  code: number
  message: string
  data: {
    list: Evidence[]
    total: number
  }
}

export interface ResolveDisputeRequest {
  decision: 'accept' | 'reject'
  reason: string
  attachments?: string[]
}

export interface ResolveDisputeResponse {
  code: number
  message: string
  data: {
    dispute_id: string
    status: string
  }
}

// API Methods
export const disputeService = {
  /**
   * Get disputes list with filters
   */
  list: (params: ListDisputesParams) => {
    return request.get<ListDisputesResponse>('/api/v1/admin/disputes', { params })
  },

  /**
   * Get dispute detail by ID
   */
  getDetail: (id: string) => {
    return request.get<DisputeDetailResponse>(`/api/v1/admin/disputes/${id}`)
  },

  /**
   * Get evidence list for a dispute
   */
  getEvidence: (disputeId: string) => {
    return request.get<EvidenceListResponse>(`/api/v1/admin/disputes/${disputeId}/evidence`)
  },

  /**
   * Resolve a dispute (accept or reject) - 对应后端的status update
   */
  resolve: (id: string, data: ResolveDisputeRequest) => {
    return request.put<ResolveDisputeResponse>(`/api/v1/admin/disputes/${id}/status`, {
      status: data.decision === 'accept' ? 'accepted' : 'rejected',
      reason: data.reason,
    })
  },

  /**
   * Upload evidence file
   */
  uploadEvidence: (disputeId: string, formData: FormData) => {
    return request.post(`/api/v1/admin/disputes/${disputeId}/evidence`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  /**
   * Download evidence file
   */
  downloadEvidence: (disputeId: string, evidenceId: string) => {
    return request.get(`/api/v1/admin/disputes/evidence/${evidenceId}`, {
      responseType: 'blob',
    })
  },

  /**
   * Export disputes report - 注意: 后端需要实现此接口
   */
  export: (params: ListDisputesParams) => {
    return request.get('/api/v1/admin/disputes/export', {
      params,
      responseType: 'blob',
    })
  },

  /**
   * Get dispute statistics (对应后端的statistics接口)
   */
  getStats: (params?: { start_date?: string; end_date?: string }) => {
    return request.get('/api/v1/admin/disputes/statistics', { params })
  },
}

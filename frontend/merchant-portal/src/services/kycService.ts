import request from './request'

export interface KYCDocument {
  business_license?: string
  id_card_front?: string
  id_card_back?: string
  bank_statement?: string
  tax_certificate?: string
  [key: string]: string | undefined
}

export interface KYCApplication {
  id: string
  merchant_id: string
  merchant_name: string
  business_type: string
  legal_name: string
  registration_number: string
  tax_id: string
  registered_address: string
  business_address: string
  contact_person: string
  contact_phone: string
  contact_email: string
  documents: KYCDocument
  status: 'pending' | 'approved' | 'rejected' | 'reviewing'
  reject_reason?: string
  reviewed_by?: string
  reviewed_at?: string
  submitted_at: string
  created_at: string
  updated_at: string
}

export interface ListKYCParams {
  page?: number
  page_size?: number
  status?: string
  merchant_id?: string
  business_type?: string
  start_time?: string
  end_time?: string
}

export interface ListKYCResponse {
  list: KYCApplication[]
  total: number
  page: number
  page_size: number
}

export interface ApproveKYCRequest {
  remark?: string
}

export interface RejectKYCRequest {
  reason: string
  remark?: string
}

export interface KYCStats {
  total: number
  pending: number
  approved: number
  rejected: number
  reviewing: number
}

export const kycService = {
  /**
   * 获取KYC申请列表
   */
  list: (params: ListKYCParams) => {
    return request.get<ListKYCResponse>('/kyc/applications', { params })
  },

  /**
   * 获取单个KYC申请详情
   */
  getById: (id: string) => {
    return request.get<KYCApplication>(`/kyc/applications/${id}`)
  },

  /**
   * 批准KYC申请
   */
  approve: (id: string, data: ApproveKYCRequest) => {
    return request.post(`/kyc/applications/${id}/approve`, data)
  },

  /**
   * 拒绝KYC申请
   */
  reject: (id: string, data: RejectKYCRequest) => {
    return request.post(`/kyc/applications/${id}/reject`, data)
  },

  /**
   * 设置KYC申请为审核中
   */
  setReviewing: (id: string) => {
    return request.post(`/kyc/applications/${id}/reviewing`)
  },

  /**
   * 获取KYC统计信息
   */
  getStats: () => {
    return request.get<KYCStats>('/kyc/stats')
  },

  /**
   * 下载KYC文档
   */
  downloadDocument: (id: string, documentType: string) => {
    return request.download(`/kyc/applications/${id}/documents/${documentType}`, `${documentType}.pdf`)
  },

  /**
   * 获取商户的KYC历史记录
   */
  getHistory: (merchantId: string) => {
    return request.get<KYCApplication[]>(`/kyc/merchants/${merchantId}/history`)
  },
}

export default kycService

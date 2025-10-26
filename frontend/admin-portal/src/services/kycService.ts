import request from './request'

// ===== 文档管理 =====

export interface KYCDocument {
  id: string
  merchant_id: string
  document_type: string
  document_url: string
  document_number?: string
  issue_date?: string
  expiry_date?: string
  status: 'pending' | 'approved' | 'rejected'
  reviewer_id?: string
  reviewer_name?: string
  reject_reason?: string
  reviewed_at?: string
  created_at: string
  updated_at: string
}

export interface ListDocumentsParams {
  page?: number
  page_size?: number
  status?: string
  merchant_id?: string
  document_type?: string
}

export interface ListDocumentsResponse {
  documents: KYCDocument[]
  total: number
  page: number
  page_size: number
}

// ===== 资质审核 =====

export interface BusinessQualification {
  id: string
  merchant_id: string
  company_name: string
  business_license_no: string
  business_license_url: string
  legal_person_name: string
  legal_person_id_card: string
  legal_person_id_card_front_url?: string
  legal_person_id_card_back_url?: string
  registered_address?: string
  registered_capital: number
  established_date?: string
  business_scope?: string
  industry?: string
  tax_registration_no?: string
  tax_registration_url?: string
  organization_code?: string
  status: 'pending' | 'approved' | 'rejected' | 'reviewing'
  reviewer_id?: string
  reviewer_name?: string
  reject_reason?: string
  remark?: string
  reviewed_at?: string
  created_at: string
  updated_at: string
}

export interface ListQualificationsParams {
  page?: number
  page_size?: number
  status?: string
  industry?: string
}

export interface ListQualificationsResponse {
  qualifications: BusinessQualification[]
  total: number
  page: number
  page_size: number
}

// ===== 商户等级 =====

export interface MerchantLevel {
  merchant_id: string
  current_level: string
  next_level?: string
  qualification_status: string
  document_completeness: number
  business_volume: number
  compliance_score: number
  can_upgrade: boolean
  upgrade_requirements?: string[]
  created_at: string
  updated_at: string
}

// ===== 风险预警 =====

export interface KYCAlert {
  id: string
  merchant_id: string
  alert_type: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  title: string
  description: string
  status: 'active' | 'resolved' | 'ignored'
  resolved_by?: string
  resolved_at?: string
  resolution_note?: string
  created_at: string
  updated_at: string
}

export interface ListAlertsParams {
  page?: number
  page_size?: number
  status?: string
  severity?: string
  merchant_id?: string
}

export interface ListAlertsResponse {
  alerts: KYCAlert[]
  total: number
  page: number
  page_size: number
}

// ===== 统计信息 =====

export interface KYCStatistics {
  total_documents: number
  pending_documents: number
  approved_documents: number
  rejected_documents: number

  total_qualifications: number
  pending_qualifications: number
  approved_qualifications: number
  rejected_qualifications: number

  active_alerts: number
  resolved_alerts: number

  merchants_by_level: {
    level: string
    count: number
  }[]
}

// ===== API Service =====

export const kycService = {
  // ===== 文档管理 API =====

  /**
   * 获取KYC文档列表
   */
  listDocuments: (params: ListDocumentsParams) => {
    return request.get<ListDocumentsResponse>('/api/v1/kyc/documents', { params })
  },

  /**
   * 获取单个文档详情
   */
  getDocument: (id: string) => {
    return request.get<{ data: KYCDocument }>(`/api/v1/kyc/documents/${id}`)
  },

  /**
   * 提交KYC文档
   */
  submitDocument: (data: Partial<KYCDocument>) => {
    return request.post('/api/v1/kyc/documents', data)
  },

  /**
   * 批准文档
   */
  approveDocument: (id: string, remark?: string) => {
    return request.post(`/api/v1/kyc/documents/${id}/approve`, { remark })
  },

  /**
   * 拒绝文档
   */
  rejectDocument: (id: string, reason: string, remark?: string) => {
    return request.post(`/api/v1/kyc/documents/${id}/reject`, { reason, remark })
  },

  // ===== 资质审核 API =====

  /**
   * 获取资质列表
   */
  listQualifications: (params: ListQualificationsParams) => {
    return request.get<ListQualificationsResponse>('/api/v1/kyc/qualifications', { params })
  },

  /**
   * 获取商户资质
   */
  getQualificationByMerchant: (merchantId: string) => {
    return request.get<{ data: BusinessQualification }>(`/api/v1/kyc/qualifications/merchant/${merchantId}`)
  },

  /**
   * 提交企业资质
   */
  submitQualification: (data: Partial<BusinessQualification>) => {
    return request.post('/api/v1/kyc/qualifications', data)
  },

  /**
   * 批准资质
   */
  approveQualification: (id: string, remark?: string) => {
    return request.post(`/api/v1/kyc/qualifications/${id}/approve`, { remark })
  },

  /**
   * 拒绝资质
   */
  rejectQualification: (id: string, reason: string, remark?: string) => {
    return request.post(`/api/v1/kyc/qualifications/${id}/reject`, { reason, remark })
  },

  // ===== 商户等级 API =====

  /**
   * 获取商户等级信息
   */
  getMerchantLevel: (merchantId: string) => {
    return request.get<{ data: MerchantLevel }>(`/api/v1/kyc/levels/${merchantId}`)
  },

  /**
   * 检查商户升级资格
   */
  checkEligibility: (merchantId: string) => {
    return request.get<{ data: { eligible: boolean; requirements: string[] } }>(
      `/api/v1/kyc/levels/${merchantId}/eligibility`
    )
  },

  // ===== 风险预警 API =====

  /**
   * 获取预警列表
   */
  listAlerts: (params: ListAlertsParams) => {
    return request.get<ListAlertsResponse>('/api/v1/kyc/alerts', { params })
  },

  /**
   * 解决预警 - 注意: 后端需要实现此接口
   */
  resolveAlert: (id: string, resolutionNote: string) => {
    return request.post(`/api/v1/kyc/alerts/${id}/resolve`, { resolution_note: resolutionNote })
  },

  // ===== 统计信息 API =====

  /**
   * 获取KYC统计信息 - 注意: 后端需要实现此接口
   */
  getStatistics: () => {
    return request.get<{ data: KYCStatistics }>('/api/v1/kyc/statistics')
  },
}

export default kycService

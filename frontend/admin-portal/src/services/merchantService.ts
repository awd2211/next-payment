import request from './request'

export interface Merchant {
  id: string
  name: string
  email: string
  phone: string
  company_name: string
  business_type: string
  country: string
  website: string
  status: string
  kyc_status: string
  is_test_mode: boolean
  metadata: any
  created_at: string
  updated_at: string
}

export interface ListMerchantsParams {
  page?: number
  page_size?: number
  status?: string
  kyc_status?: string
  keyword?: string
}

export interface ListMerchantsResponse {
  list: Merchant[]
  total: number
  page: number
  page_size: number
}

export interface CreateMerchantRequest {
  name: string
  email: string
  password: string
  phone?: string
  company_name?: string
  business_type: string
  country?: string
  website?: string
}

export interface UpdateMerchantRequest {
  name?: string
  phone?: string
  company_name?: string
  business_type?: string
  country?: string
  website?: string
}

export const merchantService = {
  list: (params: ListMerchantsParams) => {
    return request.get<ListMerchantsResponse>('/api/v1/admin/merchants', { params })
  },

  getById: (id: string) => {
    return request.get(`/api/v1/admin/merchants/${id}`)
  },

  create: (data: CreateMerchantRequest) => {
    return request.post('/api/v1/admin/merchants', data)
  },

  update: (id: string, data: UpdateMerchantRequest) => {
    return request.put(`/api/v1/admin/merchants/${id}`, data)
  },

  delete: (id: string) => {
    return request.delete(`/api/v1/admin/merchants/${id}`)
  },

  updateStatus: (id: string, status: string) => {
    return request.put(`/api/v1/admin/merchants/${id}/status`, { status })
  },

  updateKYCStatus: (id: string, kyc_status: string) => {
    return request.put(`/api/v1/admin/merchants/${id}/kyc-status`, { kyc_status })
  },
}

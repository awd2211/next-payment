import api from './api'

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
  data: {
    list: Merchant[]
    total: number
    page: number
    page_size: number
  }
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
    return api.get<any, ListMerchantsResponse>('/merchant', { params })
  },

  getById: (id: string) => {
    return api.get(`/merchant/${id}`)
  },

  create: (data: CreateMerchantRequest) => {
    return api.post('/merchant', data)
  },

  update: (id: string, data: UpdateMerchantRequest) => {
    return api.put(`/merchant/${id}`, data)
  },

  delete: (id: string) => {
    return api.delete(`/merchant/${id}`)
  },

  updateStatus: (id: string, status: string) => {
    return api.put(`/merchant/${id}/status`, { status })
  },

  updateKYCStatus: (id: string, kyc_status: string) => {
    return api.put(`/merchant/${id}/kyc-status`, { kyc_status })
  },
}

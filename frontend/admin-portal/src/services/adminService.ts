import api from './api'

export interface Admin {
  id: string
  username: string
  email: string
  full_name: string
  phone: string
  avatar: string
  status: string
  is_super: boolean
  last_login_at: string
  last_login_ip: string
  created_at: string
  updated_at: string
  roles?: Array<{
    id: string
    name: string
    display_name: string
  }>
}

export interface ListAdminsParams {
  page?: number
  page_size?: number
  status?: string
  keyword?: string
}

export interface ListAdminsResponse {
  data: Admin[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}

export interface CreateAdminRequest {
  username: string
  password: string
  email: string
  full_name: string
  phone?: string
  avatar?: string
}

export interface UpdateAdminRequest {
  email?: string
  full_name?: string
  phone?: string
  avatar?: string
  status?: string
}

export const adminService = {
  list: (params: ListAdminsParams) => {
    return api.get<any, ListAdminsResponse>('/admin', { params })
  },

  getById: (id: string) => {
    return api.get(`/admin/${id}`)
  },

  create: (data: CreateAdminRequest) => {
    return api.post('/admin', data)
  },

  update: (id: string, data: UpdateAdminRequest) => {
    return api.put(`/admin/${id}`, data)
  },

  delete: (id: string) => {
    return api.delete(`/admin/${id}`)
  },

  changePassword: (data: { old_password: string; new_password: string }) => {
    return api.post('/admin/change-password', data)
  },

  resetPassword: (id: string, new_password: string) => {
    return api.post(`/admin/${id}/reset-password`, { new_password })
  },
}

import request from './request'

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
    return request.get<ListAdminsResponse>('/api/v1/admin/admins', { params })
  },

  getById: (id: string) => {
    return request.get(`/api/v1/admin/admins/${id}`)
  },

  create: (data: CreateAdminRequest) => {
    return request.post('/api/v1/admin/admins', data)
  },

  update: (id: string, data: UpdateAdminRequest) => {
    return request.put(`/api/v1/admin/admins/${id}`, data)
  },

  delete: (id: string) => {
    return request.delete(`/api/v1/admin/admins/${id}`)
  },

  changePassword: (data: { old_password: string; new_password: string }) => {
    return request.post('/api/v1/admin/change-password', data)
  },

  resetPassword: (id: string, new_password: string) => {
    return request.post(`/api/v1/admin/admins/${id}/reset-password`, { new_password })
  },
}

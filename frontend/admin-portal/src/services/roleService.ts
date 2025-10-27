import request from './request'

export interface Permission {
  id: string
  code: string
  name: string
  resource: string
  action: string
  description: string
  created_at: string
  updated_at: string
}

export interface Role {
  id: string
  name: string
  display_name: string
  description: string
  is_system: boolean
  created_at: string
  updated_at: string
  permissions?: Permission[]
}

export interface ListRolesParams {
  page?: number
  page_size?: number
}

export interface ListRolesResponse {
  data: Role[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}

export interface CreateRoleRequest {
  name: string
  display_name: string
  description?: string
}

export interface UpdateRoleRequest {
  display_name?: string
  description?: string
}

export const roleService = {
  list: (params: ListRolesParams) => {
    return request.get<ListRolesResponse>('/api/v1/admin/roles', { params })
  },

  getById: (id: string) => {
    return request.get(`/api/v1/admin/roles/${id}`)
  },

  create: (data: CreateRoleRequest) => {
    return request.post('/api/v1/admin/roles', data)
  },

  update: (id: string, data: UpdateRoleRequest) => {
    return request.put(`/api/v1/admin/roles/${id}`, data)
  },

  delete: (id: string) => {
    return request.delete(`/api/v1/admin/roles/${id}`)
  },

  assignPermissions: (roleId: string, permissionIds: string[]) => {
    return request.post(`/api/v1/admin/roles/${roleId}/permissions`, { permission_ids: permissionIds })
  },

  assignToAdmin: (adminId: string, roleIds: string[]) => {
    return request.post('/api/v1/admin/roles/assign', { admin_id: adminId, role_ids: roleIds })
  },
}

export const permissionService = {
  list: (resource?: string) => {
    return request.get('/api/v1/admin/permissions', { params: { resource } })
  },

  listGrouped: () => {
    return request.get('/api/v1/admin/permissions/grouped')
  },

  getById: (id: string) => {
    return request.get(`/permissions/${id}`)
  },
}

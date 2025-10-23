import api from './api'

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
    return api.get<any, ListRolesResponse>('/roles', { params })
  },

  getById: (id: string) => {
    return api.get(`/roles/${id}`)
  },

  create: (data: CreateRoleRequest) => {
    return api.post('/roles', data)
  },

  update: (id: string, data: UpdateRoleRequest) => {
    return api.put(`/roles/${id}`, data)
  },

  delete: (id: string) => {
    return api.delete(`/roles/${id}`)
  },

  assignPermissions: (roleId: string, permissionIds: string[]) => {
    return api.post(`/roles/${roleId}/permissions`, { permission_ids: permissionIds })
  },

  assignToAdmin: (adminId: string, roleIds: string[]) => {
    return api.post('/roles/assign', { admin_id: adminId, role_ids: roleIds })
  },
}

export const permissionService = {
  list: (resource?: string) => {
    return api.get('/permissions', { params: { resource } })
  },

  listGrouped: () => {
    return api.get('/permissions/grouped')
  },

  getById: (id: string) => {
    return api.get(`/permissions/${id}`)
  },
}

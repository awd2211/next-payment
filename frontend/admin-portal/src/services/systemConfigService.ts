import api from './api'

export interface SystemConfig {
  id: string
  key: string
  value: string
  type: string
  category: string
  description: string
  is_public: boolean
  updated_by: string
  created_at: string
  updated_at: string
}

export interface ListConfigsParams {
  page?: number
  page_size?: number
  category?: string
}

export interface ListConfigsResponse {
  success: boolean
  data: SystemConfig[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}

export interface GroupedConfigsResponse {
  success: boolean
  data: {
    configs: Record<string, SystemConfig[]>
    total: number
  }
}

export const systemConfigService = {
  list: (params: ListConfigsParams) => {
    return api.get<any, ListConfigsResponse>('/system-configs', { params })
  },

  listGrouped: () => {
    return api.get<any, GroupedConfigsResponse>('/system-configs/grouped')
  },

  getById: (id: string) => {
    return api.get(`/system-configs/${id}`)
  },

  getByKey: (key: string) => {
    return api.get(`/system-configs/key/${key}`)
  },

  create: (data: Partial<SystemConfig>) => {
    return api.post('/system-configs', data)
  },

  update: (id: string, data: Partial<SystemConfig>) => {
    return api.put(`/system-configs/${id}`, data)
  },

  delete: (id: string) => {
    return api.delete(`/system-configs/${id}`)
  },

  batchUpdate: (configs: Array<{ id: string; value: string; description?: string }>) => {
    return api.post('/system-configs/batch', { configs })
  },
}

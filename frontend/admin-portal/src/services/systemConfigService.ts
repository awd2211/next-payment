import request from './request'

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
    return request.get<ListConfigsResponse>('/system-configs', { params })
  },

  listGrouped: () => {
    return request.get<GroupedConfigsResponse>('/system-configs/grouped')
  },

  getById: (id: string) => {
    return request.get(`/system-configs/${id}`)
  },

  getByKey: (key: string) => {
    return request.get(`/system-configs/key/${key}`)
  },

  create: (data: Partial<SystemConfig>) => {
    return request.post('/system-configs', data)
  },

  update: (id: string, data: Partial<SystemConfig>) => {
    return request.put(`/system-configs/${id}`, data)
  },

  delete: (id: string) => {
    return request.delete(`/system-configs/${id}`)
  },

  batchUpdate: (configs: Array<{ id: string; value: string; description?: string }>) => {
    return request.post('/system-configs/batch', { configs })
  },
}

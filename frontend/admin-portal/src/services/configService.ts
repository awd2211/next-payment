import request from './request'

export interface SystemConfig {
  id: string
  key: string
  value: string
  category: string
  description: string
  is_encrypted: boolean
  is_public: boolean
  created_at: string
  updated_at: string
  version: number
}

export interface FeatureFlag {
  id: string
  key: string
  name: string
  description: string
  enabled: boolean
  rollout_percentage: number
  target_users?: string[]
  target_merchants?: string[]
  created_at: string
  updated_at: string
}

export interface ServiceRegistration {
  name: string
  address: string
  port: number
  metadata?: Record<string, any>
  health_check_url?: string
}

export interface ServiceInfo {
  name: string
  address: string
  port: number
  status: string
  last_heartbeat: string
  metadata?: Record<string, any>
}

export interface ConfigHistoryItem {
  id: string
  config_id: string
  old_value: string
  new_value: string
  changed_by: string
  changed_at: string
  reason?: string
}

export interface ListConfigsParams {
  page?: number
  page_size?: number
  category?: string
  keyword?: string
}

export interface ListFeatureFlagsParams {
  page?: number
  page_size?: number
  enabled?: boolean
}

export const configService = {
  // Configuration Management
  createConfig: (data: Partial<SystemConfig>) => {
    return request.post('/api/v1/admin/configs', data)
  },

  listConfigs: (params: ListConfigsParams) => {
    return request.get('/api/v1/admin/configs', { params })
  },

  getConfig: (id: string) => {
    return request.get<{ data: SystemConfig }>(`/api/v1/admin/configs/${id}`)
  },

  updateConfig: (id: string, data: Partial<SystemConfig>) => {
    return request.put(`/api/v1/admin/configs/${id}`, data)
  },

  deleteConfig: (id: string) => {
    return request.delete(`/api/v1/admin/configs/${id}`)
  },

  getConfigHistory: (id: string) => {
    return request.get<{ data: ConfigHistoryItem[] }>(`/api/v1/admin/configs/${id}/history`)
  },

  rollbackConfig: (id: string, version: number) => {
    return request.post(`/api/v1/admin/configs/${id}/rollback`, { version })
  },

  // Feature Flags
  createFeatureFlag: (data: Partial<FeatureFlag>) => {
    return request.post('/api/v1/admin/feature-flags', data)
  },

  listFeatureFlags: (params?: ListFeatureFlagsParams) => {
    return request.get<{ data: FeatureFlag[] }>('/api/v1/admin/feature-flags', { params })
  },

  getFeatureFlag: (key: string) => {
    return request.get<{ data: FeatureFlag }>(`/feature-flags/${key}`)
  },

  checkFeatureEnabled: (key: string, merchantId?: string) => {
    return request.get<{ data: { enabled: boolean } }>(`/feature-flags/${key}/enabled`, {
      params: { merchant_id: merchantId }
    })
  },

  deleteFeatureFlag: (id: string) => {
    return request.delete(`/feature-flags/${id}`)
  },

  // Service Registry (Service Discovery)
  registerService: (data: ServiceRegistration) => {
    return request.post('/services/register', data)
  },

  listServices: () => {
    return request.get<{ data: ServiceInfo[] }>('/services')
  },

  getService: (name: string) => {
    return request.get<{ data: ServiceInfo }>(`/services/${name}`)
  },

  sendHeartbeat: (name: string) => {
    return request.post(`/services/${name}/heartbeat`)
  },

  deregisterService: (name: string) => {
    return request.post(`/services/${name}/deregister`)
  },
}

export default configService

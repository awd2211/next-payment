import request from './request'

/**
 * 系统配置模型 - 完全对齐后端
 */
export interface SystemConfig {
  id: string
  service_name: string
  config_key: string
  config_value: string
  value_type: string
  environment: string
  description: string
  is_encrypted: boolean
  version: number
  created_by: string
  updated_by: string
  created_at: string
  updated_at: string
}

export interface ListConfigsParams {
  page?: number
  page_size?: number
  service_name?: string
  environment?: string
}

export interface ListConfigsResponse {
  list: SystemConfig[]
  page: number
  page_size: number
  total: number
}

export interface ConfigHistoryItem {
  id: string
  config_id: string
  key: string
  value: string
  changed_by: string
  change_reason: string
  created_at: string
}

/**
 * 系统配置服务
 * 对齐后端 /api/v1/configs 接口
 */
export const systemConfigService = {
  /**
   * 获取配置列表
   */
  list: (params: ListConfigsParams) => {
    return request.get<ListConfigsResponse>('/configs', { params })
  },

  /**
   * 获取单个配置
   */
  getById: (id: string) => {
    return request.get<SystemConfig>(`/api/v1/admin/configs/${id}`)
  },

  /**
   * 创建配置
   */
  create: (data: Partial<SystemConfig>) => {
    return request.post<SystemConfig>('/configs', data)
  },

  /**
   * 更新配置
   */
  update: (id: string, data: Partial<SystemConfig>) => {
    return request.put<SystemConfig>(`/api/v1/admin/configs/${id}`, data)
  },

  /**
   * 删除配置
   */
  delete: (id: string) => {
    return request.delete(`/api/v1/admin/configs/${id}`)
  },

  /**
   * 获取配置历史记录
   */
  getHistory: (id: string) => {
    return request.get<{ list: ConfigHistoryItem[] }>(`/api/v1/admin/configs/${id}/history`)
  },

  /**
   * 回滚配置到指定版本
   */
  rollback: (id: string, historyId: string) => {
    return request.post<SystemConfig>(`/api/v1/admin/configs/${id}/rollback`, { history_id: historyId })
  },
}

export default systemConfigService

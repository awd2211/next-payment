import api from './api'

export interface AuditLog {
  id: string
  admin_id: string
  admin_username: string
  action: string
  resource: string
  resource_id: string
  method: string
  path: string
  ip: string
  user_agent: string
  request_body: any
  response_code: number
  response_body: any
  error_message: string
  created_at: string
}

export interface ListAuditLogsParams {
  page?: number
  page_size?: number
  admin_id?: string
  action?: string
  resource?: string
  method?: string
  ip?: string
  response_code?: number
  start_time?: string
  end_time?: string
}

export interface ListAuditLogsResponse {
  data: AuditLog[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_page: number
  }
}

export interface AuditLogStats {
  total_logs: number
  action_distribution: Record<string, number>
  resource_distribution: Record<string, number>
  response_code_distribution: Record<string, number>
  top_admins: Array<{
    admin_id: string
    admin_username: string
    count: number
  }>
}

export const auditLogService = {
  list: (params: ListAuditLogsParams) => {
    return api.get<any, ListAuditLogsResponse>('/audit-logs', { params })
  },

  getById: (id: string) => {
    return api.get(`/audit-logs/${id}`)
  },

  getStats: (params: { start_time?: string; end_time?: string }) => {
    return api.get<any, { data: AuditLogStats }>('/audit-logs/stats', { params })
  },

  export: (params: ListAuditLogsParams) => {
    return api.get('/audit-logs/export', {
      params,
      responseType: 'blob',
    })
  },
}

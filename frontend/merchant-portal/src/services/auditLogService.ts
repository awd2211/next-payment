import request from './request'

// Types
export interface AuditLog {
  id: string
  merchant_id: string
  user_id: string
  username: string
  action: string
  resource: string
  resource_id: string
  method: string
  path: string
  ip: string
  user_agent: string
  request_body?: any
  response_code: number
  response_body?: any
  error_message?: string
  created_at: string
}

export interface ListAuditLogsParams {
  page?: number
  page_size?: number
  user_id?: string
  action?: string
  resource?: string
  method?: string
  ip?: string
  response_code?: number
  start_date?: string
  end_date?: string
}

export interface ListAuditLogsResponse {
  code: number
  message: string
  data: {
    list: AuditLog[]
    total: number
    page: number
    page_size: number
  }
}

export interface AuditLogStats {
  total_logs: number
  action_distribution: Record<string, number>
  resource_distribution: Record<string, number>
  response_code_distribution: Record<string, number>
  top_users: Array<{
    user_id: string
    username: string
    count: number
  }>
  hourly_distribution: Array<{
    hour: number
    count: number
  }>
}

export interface AuditLogDetail {
  log: AuditLog
  related_logs: AuditLog[]
}

// API Methods
export const auditLogService = {
  /**
   * Get audit logs list
   */
  list: (params: ListAuditLogsParams) => {
    return request.get<ListAuditLogsResponse>('/merchant/audit-logs', { params })
  },

  /**
   * Get audit log detail by ID
   */
  getById: (id: string) => {
    return request.get<{ data: AuditLogDetail }>(`/merchant/audit-logs/${id}`)
  },

  /**
   * Get audit log statistics
   */
  getStats: (params?: { start_date?: string; end_date?: string }) => {
    return request.get<{ data: AuditLogStats }>('/merchant/audit-logs/stats', { params })
  },

  /**
   * Export audit logs
   */
  export: (params: ListAuditLogsParams) => {
    return request.get('/merchant/audit-logs/export', {
      params,
      responseType: 'blob',
    })
  },

  /**
   * Get login history
   */
  getLoginHistory: (params?: {
    page?: number
    page_size?: number
    start_date?: string
    end_date?: string
    user_id?: string
  }) => {
    return request.get('/merchant/audit-logs/login-history', { params })
  },

  /**
   * Get API call history
   */
  getApiHistory: (params?: {
    page?: number
    page_size?: number
    start_date?: string
    end_date?: string
    api_key_id?: string
  }) => {
    return request.get('/merchant/audit-logs/api-history', { params })
  },

  /**
   * Get sensitive operation logs
   */
  getSensitiveOps: (params?: {
    page?: number
    page_size?: number
    start_date?: string
    end_date?: string
  }) => {
    return request.get('/merchant/audit-logs/sensitive-ops', { params })
  },

  /**
   * Get failed operations logs
   */
  getFailedOps: (params?: {
    page?: number
    page_size?: number
    start_date?: string
    end_date?: string
  }) => {
    return request.get('/merchant/audit-logs/failed-ops', { params })
  },

  /**
   * Search audit logs by keyword
   */
  search: (keyword: string, params?: {
    page?: number
    page_size?: number
    start_date?: string
    end_date?: string
  }) => {
    return request.get('/merchant/audit-logs/search', {
      params: { keyword, ...params }
    })
  },
}

export default auditLogService

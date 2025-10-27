import request from './request'

// Types
export interface WebhookLog {
  id: string
  webhook_id: string
  merchant_name: string
  merchant_id: string
  event_type: string
  url: string
  request_body: string
  response_status: number
  response_body: string
  retry_count: number
  status: 'success' | 'failed' | 'pending' | 'retrying'
  created_at: string
  completed_at?: string
  error_message?: string
}

export interface WebhookConfig {
  id: string
  merchant_id: string
  merchant_name: string
  url: string
  secret: string
  enabled_events: string[]
  is_active: boolean
  retry_count: number
  retry_interval: number
  timeout: number
  created_at: string
  updated_at: string
}

export interface ListWebhookLogsParams {
  page?: number
  page_size?: number
  merchant_id?: string
  webhook_id?: string
  event_type?: string
  status?: string
  start_date?: string
  end_date?: string
}

export interface ListWebhookLogsResponse {
  code: number
  message: string
  data: {
    list: WebhookLog[]
    total: number
    page: number
    page_size: number
  }
}

export interface WebhookLogDetailResponse {
  code: number
  message: string
  data: WebhookLog
}

export interface RetryWebhookResponse {
  code: number
  message: string
  data: {
    webhook_id: string
    status: string
  }
}

export interface WebhookStatsResponse {
  code: number
  message: string
  data: {
    total_count: number
    success_count: number
    failed_count: number
    pending_count: number
    success_rate: number
    avg_response_time: number
  }
}

export interface ListWebhookConfigsResponse {
  code: number
  message: string
  data: {
    list: WebhookConfig[]
    total: number
  }
}

export interface UpdateWebhookConfigRequest {
  url?: string
  enabled_events?: string[]
  is_active?: boolean
  retry_count?: number
  retry_interval?: number
  timeout?: number
}

// API Methods
export const webhookService = {
  /**
   * Get webhook logs list
   */
  list: (params: ListWebhookLogsParams) => {
    return request.get<ListWebhookLogsResponse>('/merchant/webhooks/logs', { params })
  },

  /**
   * Get webhook log detail by ID
   */
  getDetail: (id: string) => {
    return request.get<WebhookLogDetailResponse>(`/merchant/webhooks/logs/${id}`)
  },

  /**
   * Retry failed webhook
   */
  retry: (id: string) => {
    return request.post<RetryWebhookResponse>(`/merchant/webhooks/logs/${id}/retry`)
  },

  /**
   * Batch retry failed webhooks
   */
  batchRetry: (ids: string[]) => {
    return request.post('/merchant/webhooks/logs/batch-retry', { ids })
  },

  /**
   * Get webhook statistics
   */
  getStats: (params?: { merchant_id?: string; start_date?: string; end_date?: string }) => {
    return request.get<WebhookStatsResponse>('/merchant/webhooks/stats', { params })
  },

  /**
   * Export webhook logs
   */
  export: (params: ListWebhookLogsParams) => {
    return request.get('/merchant/webhooks/logs/export', {
      params,
      responseType: 'blob',
    })
  },

  /**
   * Get webhook configs list
   */
  listConfigs: (params?: { merchant_id?: string }) => {
    return request.get<ListWebhookConfigsResponse>('/merchant/webhooks/configs', { params })
  },

  /**
   * Get webhook config detail
   */
  getConfig: (id: string) => {
    return request.get(`/merchant/webhooks/configs/${id}`)
  },

  /**
   * Update webhook config
   */
  updateConfig: (id: string, data: UpdateWebhookConfigRequest) => {
    return request.put(`/merchant/webhooks/configs/${id}`, data)
  },

  /**
   * Test webhook endpoint
   */
  testWebhook: (merchantId: string, data: { event_type: string; test_data?: any }) => {
    return request.post(`/merchant/webhooks/merchants/${merchantId}/test`, data)
  },

  /**
   * Get retry history for a webhook
   */
  getRetryHistory: (id: string) => {
    return request.get(`/merchant/webhooks/logs/${id}/retry-history`)
  },

  /**
   * Get available event types
   */
  getEventTypes: () => {
    return request.get<{ code: number; message: string; data: string[] }>('/merchant/webhooks/event-types')
  },
}

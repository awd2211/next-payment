import request from './request'

// Types
export interface Report {
  id: string
  report_no: string
  report_type: 'payment' | 'settlement' | 'transaction' | 'reconciliation' | 'fee' | 'custom'
  report_name: string
  merchant_id: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  file_url?: string
  file_size?: number
  params: Record<string, any>
  created_at: string
  completed_at?: string
  expires_at?: string
}

export interface ReportTemplate {
  id: string
  template_name: string
  template_type: string
  description: string
  fields: string[]
  filters: string[]
}

export interface ScheduledReport {
  id: string
  report_name: string
  report_type: string
  schedule: 'daily' | 'weekly' | 'monthly'
  schedule_time: string
  is_enabled: boolean
  recipients: string[]
  last_run_at?: string
  next_run_at: string
  created_at: string
}

export interface CreateReportRequest {
  report_type: string
  report_name: string
  start_date: string
  end_date: string
  currency?: string
  channel?: string
  status?: string
  format?: 'xlsx' | 'csv' | 'pdf'
  fields?: string[]
  filters?: Record<string, any>
}

export interface CreateScheduledReportRequest {
  report_name: string
  report_type: string
  schedule: 'daily' | 'weekly' | 'monthly'
  schedule_time: string
  recipients: string[]
  params: Record<string, any>
}

export interface ListReportsParams {
  page?: number
  page_size?: number
  report_type?: string
  status?: string
  start_date?: string
  end_date?: string
}

export interface ListReportsResponse {
  code: number
  message: string
  data: {
    list: Report[]
    total: number
    page: number
    page_size: number
  }
}

// API Methods
export const reportService = {
  /**
   * Get reports list
   */
  list: (params: ListReportsParams) => {
    return request.get<ListReportsResponse>('/merchant/reports', { params })
  },

  /**
   * Create a new report
   */
  create: (data: CreateReportRequest) => {
    return request.post<{ data: Report }>('/merchant/reports', data)
  },

  /**
   * Get report detail
   */
  getDetail: (reportNo: string) => {
    return request.get<{ data: Report }>(`/merchant/reports/${reportNo}`)
  },

  /**
   * Download report file
   */
  download: (reportNo: string) => {
    return request.get(`/merchant/reports/${reportNo}/download`, {
      responseType: 'blob',
    })
  },

  /**
   * Delete a report
   */
  delete: (reportNo: string) => {
    return request.delete(`/merchant/reports/${reportNo}`)
  },

  /**
   * Get report templates
   */
  getTemplates: () => {
    return request.get<{ data: ReportTemplate[] }>('/merchant/reports/templates')
  },

  /**
   * Get scheduled reports
   */
  listScheduled: (params?: { page?: number; page_size?: number }) => {
    return request.get<{ data: { list: ScheduledReport[]; total: number } }>('/merchant/reports/scheduled', { params })
  },

  /**
   * Create scheduled report
   */
  createScheduled: (data: CreateScheduledReportRequest) => {
    return request.post<{ data: ScheduledReport }>('/merchant/reports/scheduled', data)
  },

  /**
   * Update scheduled report
   */
  updateScheduled: (id: string, data: Partial<CreateScheduledReportRequest>) => {
    return request.put<{ data: ScheduledReport }>(`/merchant/reports/scheduled/${id}`, data)
  },

  /**
   * Delete scheduled report
   */
  deleteScheduled: (id: string) => {
    return request.delete(`/merchant/reports/scheduled/${id}`)
  },

  /**
   * Toggle scheduled report
   */
  toggleScheduled: (id: string, isEnabled: boolean) => {
    return request.put(`/merchant/reports/scheduled/${id}/toggle`, { is_enabled: isEnabled })
  },

  /**
   * Run scheduled report immediately
   */
  runScheduled: (id: string) => {
    return request.post<{ data: Report }>(`/merchant/reports/scheduled/${id}/run`)
  },

  /**
   * Get available report fields
   */
  getFields: (reportType: string) => {
    return request.get<{ data: string[] }>(`/merchant/reports/fields/${reportType}`)
  },

  /**
   * Preview report data
   */
  preview: (data: CreateReportRequest) => {
    return request.post('/merchant/reports/preview', data)
  },
}

export default reportService

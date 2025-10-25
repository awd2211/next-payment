import request from './request'

export interface EmailNotification {
  to: string
  subject: string
  body: string
  template_id?: string
  variables?: Record<string, any>
}

export interface SMSNotification {
  phone: string
  message: string
  template_id?: string
  variables?: Record<string, any>
}

export interface WebhookNotification {
  url: string
  method: string
  headers?: Record<string, string>
  body: any
  retry_count?: number
}

export interface Notification {
  id: string
  type: string // email, sms, webhook
  status: string
  recipient: string
  content: string
  error_message?: string
  sent_at?: string
  created_at: string
  updated_at: string
}

export interface EmailTemplate {
  id: string
  name: string
  subject: string
  body: string
  variables: string[]
  category: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ListTemplatesParams {
  page?: number
  page_size?: number
  category?: string
  keyword?: string
}

export interface ListNotificationsParams {
  page?: number
  page_size?: number
  type?: string
  status?: string
  start_time?: string
  end_time?: string
}

export const notificationService = {
  // Email Notifications
  sendEmail: (data: EmailNotification) => {
    return request.post('/notifications/email', data)
  },

  // SMS Notifications
  sendSMS: (data: SMSNotification) => {
    return request.post('/notifications/sms', data)
  },

  // Webhook Notifications
  sendWebhook: (data: WebhookNotification) => {
    return request.post('/notifications/webhook', data)
  },

  // Email Template Management
  createTemplate: (data: Partial<EmailTemplate>) => {
    return request.post('/email-templates', data)
  },

  getTemplate: (id: string) => {
    return request.get<EmailTemplate>(`/email-templates/${id}`)
  },

  listTemplates: (params: ListTemplatesParams) => {
    return request.get('/email-templates', { params })
  },

  updateTemplate: (id: string, data: Partial<EmailTemplate>) => {
    return request.put(`/email-templates/${id}`, data)
  },

  deleteTemplate: (id: string) => {
    return request.delete(`/email-templates/${id}`)
  },

  // Notification History
  listNotifications: (params: ListNotificationsParams) => {
    return request.get('/notifications/history', { params })
  },

  getNotification: (id: string) => {
    return request.get<Notification>(`/notifications/${id}`)
  },
}

export default notificationService

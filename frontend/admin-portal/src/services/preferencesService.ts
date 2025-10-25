import request from './request'

export interface AdminPreferences {
  id: string
  admin_id: string
  theme: string // light, dark, auto
  language: string // en, zh-CN, etc.
  timezone: string
  date_format: string
  time_format: string
  currency: string
  notifications_enabled: boolean
  email_notifications: boolean
  sms_notifications: boolean
  dashboard_layout?: any
  table_page_size: number
  created_at: string
  updated_at: string
}

export interface UpdatePreferencesRequest {
  theme?: string
  language?: string
  timezone?: string
  date_format?: string
  time_format?: string
  currency?: string
  notifications_enabled?: boolean
  email_notifications?: boolean
  sms_notifications?: boolean
  dashboard_layout?: any
  table_page_size?: number
}

export const preferencesService = {
  // Get current admin's preferences
  getPreferences: () => {
    return request.get<{ data: AdminPreferences }>('/preferences')
  },

  // Update current admin's preferences
  updatePreferences: (data: UpdatePreferencesRequest) => {
    return request.put<{ data: AdminPreferences }>('/preferences', data)
  },

  // Get specific admin's preferences (admin only)
  getAdminPreferences: (adminId: string) => {
    return request.get<{ data: AdminPreferences }>(`/admin/${adminId}/preferences`)
  },

  // Update specific admin's preferences (admin only)
  updateAdminPreferences: (adminId: string, data: UpdatePreferencesRequest) => {
    return request.put<{ data: AdminPreferences }>(`/admin/${adminId}/preferences`, data)
  },

  // Reset to default preferences
  resetPreferences: () => {
    return request.delete('/preferences')
  },
}

export default preferencesService

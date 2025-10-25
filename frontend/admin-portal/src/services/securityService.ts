import request from './request'

export interface SecurityEvent {
  id: string
  admin_id: string
  event_type: string // login, logout, password_change, permission_change, etc.
  ip_address: string
  user_agent: string
  status: string // success, failed, blocked
  risk_score?: number
  metadata?: any
  created_at: string
}

export interface LoginAttempt {
  id: string
  username: string
  ip_address: string
  status: string
  failed_reason?: string
  created_at: string
}

export interface IPWhitelist {
  id: string
  admin_id?: string
  ip_address: string
  description?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface SecuritySettings {
  id: string
  max_login_attempts: number
  lockout_duration_minutes: number
  password_min_length: number
  password_require_uppercase: boolean
  password_require_lowercase: boolean
  password_require_numbers: boolean
  password_require_special_chars: boolean
  password_expiry_days: number
  session_timeout_minutes: number
  two_factor_required: boolean
  ip_whitelist_enabled: boolean
  created_at: string
  updated_at: string
}

export interface ListSecurityEventsParams {
  page?: number
  page_size?: number
  admin_id?: string
  event_type?: string
  status?: string
  start_time?: string
  end_time?: string
}

export interface ListLoginAttemptsParams {
  page?: number
  page_size?: number
  username?: string
  ip_address?: string
  status?: string
  start_time?: string
  end_time?: string
}

export const securityService = {
  // Security Events
  listSecurityEvents: (params: ListSecurityEventsParams) => {
    return request.get('/security/events', { params })
  },

  getSecurityEvent: (id: string) => {
    return request.get<{ data: SecurityEvent }>(`/security/events/${id}`)
  },

  // Login Attempts
  listLoginAttempts: (params: ListLoginAttemptsParams) => {
    return request.get('/security/login-attempts', { params })
  },

  getLoginAttempt: (id: string) => {
    return request.get<{ data: LoginAttempt }>(`/security/login-attempts/${id}`)
  },

  // IP Whitelist
  addIPToWhitelist: (data: { ip_address: string; description?: string }) => {
    return request.post<{ data: IPWhitelist }>('/security/ip-whitelist', data)
  },

  listIPWhitelist: (adminId?: string) => {
    return request.get<{ data: IPWhitelist[] }>('/security/ip-whitelist', {
      params: { admin_id: adminId }
    })
  },

  removeIPFromWhitelist: (id: string) => {
    return request.delete(`/security/ip-whitelist/${id}`)
  },

  checkIPWhitelisted: (ipAddress: string) => {
    return request.get<{ data: { is_whitelisted: boolean } }>('/security/ip-whitelist/check', {
      params: { ip_address: ipAddress }
    })
  },

  // Security Settings
  getSecuritySettings: () => {
    return request.get<{ data: SecuritySettings }>('/security/settings')
  },

  updateSecuritySettings: (data: Partial<SecuritySettings>) => {
    return request.put<{ data: SecuritySettings }>('/security/settings', data)
  },

  // Account Security
  unlockAccount: (adminId: string) => {
    return request.post(`/security/unlock/${adminId}`)
  },

  forcePasswordReset: (adminId: string) => {
    return request.post(`/security/force-password-reset/${adminId}`)
  },

  // Session Management
  listActiveSessions: (adminId?: string) => {
    return request.get('/security/sessions', {
      params: { admin_id: adminId }
    })
  },

  terminateSession: (sessionId: string) => {
    return request.delete(`/security/sessions/${sessionId}`)
  },

  terminateAllSessions: (adminId: string) => {
    return request.delete(`/security/sessions/admin/${adminId}`)
  },
}

export default securityService

import request from './request'

export interface APIKey {
  id: string
  merchant_id: string
  key_id: string
  key_secret: string // Only returned on creation
  name: string
  description?: string
  permissions: string[]
  rate_limit?: number
  ip_whitelist?: string[]
  is_active: boolean
  expires_at?: string
  last_used_at?: string
  created_at: string
  updated_at: string
}

export interface AuthSession {
  id: string
  merchant_id: string
  token: string
  user_agent?: string
  ip_address?: string
  expires_at: string
  created_at: string
}

export interface SecuritySettings {
  merchant_id: string
  two_factor_enabled: boolean
  two_factor_method?: string // totp, sms, email
  ip_whitelist_enabled: boolean
  ip_whitelist?: string[]
  webhook_signature_enabled: boolean
  webhook_signature_algorithm?: string
  api_key_rotation_days?: number
  session_timeout_minutes?: number
}

export interface CreateAPIKeyRequest {
  name: string
  description?: string
  permissions: string[]
  rate_limit?: number
  ip_whitelist?: string[]
  expires_at?: string
}

export interface UpdateSecuritySettingsRequest {
  two_factor_enabled?: boolean
  two_factor_method?: string
  ip_whitelist_enabled?: boolean
  ip_whitelist?: string[]
  webhook_signature_enabled?: boolean
  webhook_signature_algorithm?: string
  api_key_rotation_days?: number
  session_timeout_minutes?: number
}

export const merchantAuthService = {
  // API Key Management
  createAPIKey: (data: CreateAPIKeyRequest) => {
    return request.post<{ data: APIKey }>('/api-keys', data)
  },

  listAPIKeys: (merchantId?: string) => {
    return request.get<{ data: APIKey[] }>('/api-keys', {
      params: { merchant_id: merchantId }
    })
  },

  getAPIKey: (id: string) => {
    return request.get<{ data: APIKey }>(`/api-keys/${id}`)
  },

  rotateAPIKey: (id: string) => {
    return request.post<{ data: APIKey }>(`/api-keys/${id}/rotate`)
  },

  deleteAPIKey: (id: string) => {
    return request.delete(`/api-keys/${id}`)
  },

  // Session Management
  createSession: (merchantId: string, password: string) => {
    return request.post<{ data: AuthSession }>('/auth/sessions', {
      merchant_id: merchantId,
      password
    })
  },

  getSession: (token: string) => {
    return request.get<{ data: AuthSession }>(`/auth/sessions/${token}`)
  },

  logoutSession: (token: string) => {
    return request.delete(`/auth/sessions/${token}`)
  },

  // Security Settings
  getSecuritySettings: (merchantId?: string) => {
    return request.get<{ data: SecuritySettings }>('/security/settings', {
      params: { merchant_id: merchantId }
    })
  },

  updateSecuritySettings: (data: UpdateSecuritySettingsRequest) => {
    return request.put<{ data: SecuritySettings }>('/security/settings', data)
  },

  enable2FA: (method: string) => {
    return request.post('/security/2fa/enable', { method })
  },

  disable2FA: () => {
    return request.post('/security/2fa/disable')
  },

  // Signature Validation (Internal - for testing)
  validateSignature: (data: {
    merchant_id: string
    timestamp: string
    signature: string
    body: string
  }) => {
    return request.post('/auth/validate-signature', data)
  },
}

export default merchantAuthService

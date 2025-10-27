import request from './request'

// ========== API Key 管理 ==========
export interface APIKey {
  id: string
  merchant_id: string
  key_prefix: string // API Key前缀 (如 pk_test_, sk_live_)
  environment: string // test, production
  status: string // active, revoked
  permissions: string[] // 权限列表
  last_used_at: string
  expires_at: string
  created_at: string
  updated_at: string
}

export interface CreateAPIKeyInput {
  environment: 'test' | 'production'
  description?: string
  permissions?: string[]
  expires_in_days?: number
}

export interface CreateAPIKeyResponse {
  api_key: string // 完整的API Key (只在创建时返回一次)
  api_secret: string // API Secret (只在创建时返回一次)
  key_info: APIKey
}

// ========== 安全设置 ==========
export interface SecuritySettings {
  two_factor_enabled: boolean
  ip_whitelist: string[]
  session_timeout: number // 分钟
  password_expires_in_days: number
  require_password_change: boolean
}

export interface ChangePasswordInput {
  old_password: string
  new_password: string
}

export const apiKeyService = {
  // ========== API Key 管理 ==========
  /**
   * 创建新的API Key
   * 注意：API Key和Secret只在创建时返回一次，请妥善保管
   */
  createAPIKey: (data: CreateAPIKeyInput) => {
    return request.post<CreateAPIKeyResponse>('/merchant/api-keys', data)
  },

  /**
   * 列出商户的所有API Keys
   */
  listAPIKeys: () => {
    return request.get<APIKey[]>('/merchant/api-keys')
  },

  /**
   * 删除(撤销) API Key
   */
  deleteAPIKey: (id: string) => {
    return request.delete(`/api-keys/${id}`)
  },

  // ========== 安全设置 ==========
  /**
   * 修改密码
   */
  changePassword: (data: ChangePasswordInput) => {
    return request.put('/merchant/security/password', data)
  },

  /**
   * 启用双因素认证
   */
  enable2FA: () => {
    return request.post<{ qr_code: string; secret: string }>('/merchant/security/2fa/enable')
  },

  /**
   * 验证双因素认证
   */
  verify2FA: (code: string) => {
    return request.post('/merchant/security/2fa/verify', { code })
  },

  /**
   * 禁用双因素认证
   */
  disable2FA: (password: string) => {
    return request.post('/merchant/security/2fa/disable', { password })
  },

  /**
   * 获取安全设置
   */
  getSecuritySettings: () => {
    return request.get<SecuritySettings>('/merchant/security/settings')
  },

  /**
   * 更新安全设置
   */
  updateSecuritySettings: (data: Partial<SecuritySettings>) => {
    return request.put<SecuritySettings>('/merchant/security/settings', data)
  },
}

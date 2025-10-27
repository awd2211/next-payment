import request from './request'

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  refresh_token: string
  merchant: {
    id: string
    name: string
    email: string
    company_name: string
    status: string
    kyc_status: string
  }
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
  phone?: string
  company_name?: string
  business_type: string
  country?: string
  website?: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  new_password: string
}

export const authService = {
  /**
   * 商户登录
   */
  login: (data: LoginRequest) => {
    return request.post<LoginResponse>('/merchant/login', data)
  },

  /**
   * 商户注册
   */
  register: (data: RegisterRequest) => {
    return request.post<LoginResponse>('/merchant/register', data)
  },

  /**
   * 登出
   */
  logout: () => {
    return request.post('/merchant/logout')
  },

  /**
   * 修改密码
   */
  changePassword: (data: ChangePasswordRequest) => {
    return request.post('/merchant/change-password', data)
  },

  /**
   * 忘记密码 - 发送重置链接
   */
  forgotPassword: (data: ForgotPasswordRequest) => {
    return request.post('/merchant/forgot-password', data)
  },

  /**
   * 重置密码
   */
  resetPassword: (data: ResetPasswordRequest) => {
    return request.post('/merchant/reset-password', data)
  },

  /**
   * 验证邮箱
   */
  verifyEmail: (token: string) => {
    return request.post('/merchant/verify-email', { token })
  },

  /**
   * 重新发送验证邮件
   */
  resendVerificationEmail: () => {
    return request.post('/merchant/resend-verification')
  },

  /**
   * 刷新Token
   */
  refreshToken: (refreshToken: string) => {
    return request.post<{ token: string; refresh_token: string }>('/merchant/refresh', {
      refresh_token: refreshToken,
    })
  },
}

export default authService

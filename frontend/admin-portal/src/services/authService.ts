import request from './request'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  admin: any
  token: string
  refresh_token: string
  expires_in: number
}

export const authService = {
  login: (data: LoginRequest) => {
    return request.post<LoginResponse>('/api/v1/admin/login', data)
  },

  logout: () => {
    return request.post('/api/v1/admin/logout')
  },

  changePassword: (data: { old_password: string; new_password: string }) => {
    return request.post('/api/v1/admin/change-password', data)
  },

  getCurrentAdmin: () => {
    return request.get('/api/v1/admin/me')
  },
}

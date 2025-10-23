import api from './api'

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
    return api.post<any, LoginResponse>('/admin/login', data)
  },

  logout: () => {
    return api.post('/admin/logout')
  },

  changePassword: (data: { old_password: string; new_password: string }) => {
    return api.post('/admin/change-password', data)
  },

  getCurrentAdmin: () => {
    return api.get('/admin/me')
  },
}

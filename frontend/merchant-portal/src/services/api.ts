import axios from 'axios'
import { message } from 'antd'
import { useAuthStore } from '../stores/authStore'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 401:
          message.error('未授权，请重新登录')
          useAuthStore.getState().clearAuth()
          window.location.href = '/login'
          break
        case 403:
          message.error('没有权限执行此操作')
          break
        case 404:
          message.error('请求的资源不存在')
          break
        case 500:
          message.error(data?.error?.message || '服务器错误')
          break
        default:
          message.error(data?.error?.message || '请求失败')
      }
    } else if (error.request) {
      message.error('网络错误，请检查您的连接')
    } else {
      message.error('请求配置错误')
    }

    return Promise.reject(error)
  }
)

export default api

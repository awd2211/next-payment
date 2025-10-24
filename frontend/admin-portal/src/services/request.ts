import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import { message } from 'antd'
import { useAuthStore } from '../stores/authStore'
import type { ApiResponse } from '../types'

// 创建axios实例
const instance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_PREFIX || '/api/v1',
  timeout: Number(import.meta.env.VITE_REQUEST_TIMEOUT) || 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 刷新token的Promise（避免并发刷新）
let refreshTokenPromise: Promise<string> | null = null

/**
 * 刷新token
 */
const refreshAccessToken = async (): Promise<string> => {
  if (refreshTokenPromise) {
    return refreshTokenPromise
  }

  refreshTokenPromise = (async () => {
    try {
      const { refreshToken } = useAuthStore.getState()
      if (!refreshToken) {
        throw new Error('No refresh token')
      }

      // 调用刷新token接口
      const response = await axios.post<ApiResponse<{ token: string; refresh_token: string }>>(
        '/api/v1/auth/refresh',
        { refresh_token: refreshToken }
      )

      if (response.data.code === 0 && response.data.data) {
        const { token, refresh_token } = response.data.data
        const { admin } = useAuthStore.getState()
        if (admin) {
          useAuthStore.getState().setAuth(token, refresh_token, admin)
        }
        return token
      } else {
        throw new Error('Refresh token failed')
      }
    } catch (error) {
      // 刷新失败，清除认证信息
      useAuthStore.getState().clearAuth()
      window.location.href = '/login'
      throw error
    } finally {
      refreshTokenPromise = null
    }
  })()

  return refreshTokenPromise
}

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // 添加请求ID（用于链路追踪）
    config.headers['X-Request-ID'] = `${Date.now()}-${Math.random().toString(36).slice(2)}`

    return config
  },
  (error: AxiosError) => {
    console.error('[Request Error]', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
instance.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>): any => {
    // 直接返回data字段
    return response.data
  },
  async (error: AxiosError<ApiResponse>) => {
    if (!error.response) {
      // 网络错误
      message.error('网络错误，请检查您的网络连接')
      return Promise.reject(error)
    }

    const { status, data, config } = error.response

    // 处理401 - 尝试刷新token
    if (status === 401) {
      // 如果是刷新token接口本身失败，直接跳转登录
      if (config.url?.includes('/auth/refresh')) {
        message.error('登录已过期，请重新登录')
        useAuthStore.getState().clearAuth()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        // 尝试刷新token
        const newToken = await refreshAccessToken()

        // 重试原请求
        if (config.headers) {
          config.headers.Authorization = `Bearer ${newToken}`
        }
        return instance.request(config)
      } catch (refreshError) {
        // 刷新失败，已经在refreshAccessToken中处理跳转
        return Promise.reject(refreshError)
      }
    }

    // 处理其他错误
    const errorMessage = data?.error?.message || getErrorMessage(status)
    message.error(errorMessage)

    // 记录错误日志（生产环境可以上报到监控系统）
    if (import.meta.env.MODE === 'production') {
      // TODO: 上报错误到监控系统
      console.error('[API Error]', {
        url: config.url,
        method: config.method,
        status,
        data,
      })
    }

    return Promise.reject(error)
  }
)

/**
 * 获取默认错误消息
 */
function getErrorMessage(status: number): string {
  const errorMessages: Record<number, string> = {
    400: '请求参数错误',
    401: '未授权，请重新登录',
    403: '没有权限执行此操作',
    404: '请求的资源不存在',
    405: '请求方法不允许',
    408: '请求超时',
    409: '数据冲突',
    422: '数据验证失败',
    429: '请求过于频繁，请稍后再试',
    500: '服务器内部错误',
    502: '网关错误',
    503: '服务暂时不可用',
    504: '网关超时',
  }

  return errorMessages[status] || `请求失败 (${status})`
}

/**
 * 封装的请求方法
 */
class Request {
  /**
   * GET请求
   */
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.get(url, config)
  }

  /**
   * POST请求
   */
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.post(url, data, config)
  }

  /**
   * PUT请求
   */
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.put(url, data, config)
  }

  /**
   * DELETE请求
   */
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.delete(url, config)
  }

  /**
   * PATCH请求
   */
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return instance.patch(url, data, config)
  }

  /**
   * 上传文件
   */
  upload<T = any>(url: string, formData: FormData, onProgress?: (progress: number) => void): Promise<ApiResponse<T>> {
    return instance.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      },
    })
  }

  /**
   * 下载文件
   */
  download(url: string, filename?: string, config?: AxiosRequestConfig): Promise<void> {
    return instance
      .get(url, {
        ...config,
        responseType: 'blob',
      })
      .then((response: any) => {
        const blob = new Blob([response])
        const downloadUrl = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = downloadUrl
        link.download = filename || `download-${Date.now()}`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(downloadUrl)
      })
  }
}

export default new Request()



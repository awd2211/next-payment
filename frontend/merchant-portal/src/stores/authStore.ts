import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface Merchant {
  id: string
  name: string
  email: string
  phone?: string
  company_name?: string
  business_type?: string
  country?: string
  website?: string
  status: string
  kyc_status?: string
  is_test_mode?: boolean
}

interface AuthState {
  token: string | null
  refreshToken: string | null
  merchant: Merchant | null
  setAuth: (token: string, refreshToken: string, merchant: Merchant) => void
  clearAuth: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      merchant: null,

      setAuth: (token, refreshToken, merchant) => {
        set({ token, refreshToken, merchant })
      },

      clearAuth: () => {
        set({ token: null, refreshToken: null, merchant: null })
        // 双保险:也清除 localStorage 中的所有认证相关数据
        localStorage.removeItem('token')
        localStorage.removeItem('refreshToken')
        localStorage.removeItem('merchant')
      },
    }),
    {
      name: 'merchant-auth-storage',
    }
  )
)

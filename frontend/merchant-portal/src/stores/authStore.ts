import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface Merchant {
  id: string
  name: string
  code: string
  contact_email: string
  contact_phone: string
  status: string
  balance: number
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
      },
    }),
    {
      name: 'merchant-auth-storage',
    }
  )
)

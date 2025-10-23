import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface Admin {
  id: string
  username: string
  email: string
  full_name: string
  avatar: string
  status: string
  is_super: boolean
  roles: Array<{
    id: string
    name: string
    display_name: string
    description: string
    permissions: Array<{
      id: string
      code: string
      name: string
      resource: string
      action: string
    }>
  }>
}

interface AuthState {
  token: string | null
  refreshToken: string | null
  admin: Admin | null
  setAuth: (token: string, refreshToken: string, admin: Admin) => void
  clearAuth: () => void
  hasPermission: (permission: string) => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      refreshToken: null,
      admin: null,

      setAuth: (token, refreshToken, admin) => {
        set({ token, refreshToken, admin })
      },

      clearAuth: () => {
        set({ token: null, refreshToken: null, admin: null })
      },

      hasPermission: (permission: string) => {
        const { admin } = get()
        if (!admin) return false
        if (admin.is_super) return true

        return admin.roles.some(role =>
          role.permissions.some(p => p.code === permission)
        )
      },
    }),
    {
      name: 'auth-storage',
    }
  )
)

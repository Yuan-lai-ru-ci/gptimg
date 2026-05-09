import { create } from 'zustand'
import type { User } from '@/types/user'
import { authApi } from '@/lib/api/auth'

interface AuthStore {
  user: User | null
  token: string | null
  isLoading: boolean
  isAuthenticated: boolean

  setAuth: (token: string, user: User) => void
  clearAuth: () => void
  loadAuth: () => void
  refreshUser: () => Promise<void>
}

export const useAuthStore = create<AuthStore>((set, get) => ({
  user: null,
  token: null,
  isLoading: true,
  isAuthenticated: false,

  setAuth: (token: string, user: User) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('access_token', token)
      localStorage.setItem('user', JSON.stringify(user))
    }
    set({ token, user, isAuthenticated: true })
  },

  clearAuth: () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token')
      localStorage.removeItem('user')
    }
    set({ token: null, user: null, isAuthenticated: false })
  },

  loadAuth: () => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token')
      const userStr = localStorage.getItem('user')
      if (token && userStr) {
        try {
          const user = JSON.parse(userStr)
          set({ token, user, isAuthenticated: true, isLoading: false })
        } catch {
          set({ isLoading: false })
        }
      } else {
        set({ isLoading: false })
      }
    }
  },

  refreshUser: async () => {
    try {
      const res = await authApi.getMe()
      const user = res.data
      if (typeof window !== 'undefined') {
        localStorage.setItem('user', JSON.stringify(user))
      }
      set({ user })
    } catch (error) {
      get().clearAuth()
    }
  },
}))

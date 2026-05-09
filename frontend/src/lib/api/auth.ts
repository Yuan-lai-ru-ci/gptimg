import apiClient from './client'
import type { AuthResponse, LoginRequest, RegisterRequest, User } from '@/types/user'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export const authApi = {
  register: async (data: RegisterRequest): Promise<ApiResponse<AuthResponse>> => {
    return apiClient.post('/auth/register', data)
  },

  login: async (data: LoginRequest): Promise<ApiResponse<AuthResponse>> => {
    return apiClient.post('/auth/login', data)
  },

  getMe: async (): Promise<ApiResponse<User>> => {
    return apiClient.get('/auth/me')
  },
}

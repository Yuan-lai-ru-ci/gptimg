import apiClient from './client'
import type { Session, CreateSessionRequest } from '@/types/chat'
import type { ImageRecord } from '@/types/image'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export const sessionApi = {
  getList: async (limit = 50, offset = 0): Promise<ApiResponse<Session[]>> => {
    return apiClient.get('/sessions', { params: { limit, offset } })
  },

  create: async (data: CreateSessionRequest): Promise<ApiResponse<Session>> => {
    return apiClient.post('/sessions', data)
  },

  getById: async (id: string): Promise<ApiResponse<Session>> => {
    return apiClient.get(`/sessions/${id}`)
  },

  update: async (id: string, data: Partial<Session>): Promise<ApiResponse<Session>> => {
    return apiClient.put(`/sessions/${id}`, data)
  },

  getMessages: async (id: string, limit = 100, offset = 0): Promise<ApiResponse<ImageRecord[]>> => {
    return apiClient.get(`/sessions/${id}/messages`, { params: { limit, offset } })
  },

  delete: async (id: string): Promise<void> => {
    return apiClient.delete(`/sessions/${id}`)
  },
}
